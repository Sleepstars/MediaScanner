package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/sleepstars/mediascanner/internal/api"
	"github.com/sleepstars/mediascanner/internal/config"
	"github.com/sleepstars/mediascanner/internal/database"
	"github.com/sleepstars/mediascanner/internal/fileops"
	"github.com/sleepstars/mediascanner/internal/llm"
	"github.com/sleepstars/mediascanner/internal/notification"
	"github.com/sleepstars/mediascanner/internal/processor"
	"github.com/sleepstars/mediascanner/internal/scanner"
	"github.com/sleepstars/mediascanner/internal/worker"
)

func main() {
	// Load environment variables from .env file if it exists
	_ = godotenv.Load()

	// Parse command line flags
	configFile := flag.String("config", "", "Path to configuration file")
	flag.Parse()

	// Print banner
	fmt.Println("MediaScanner - LLM-based Media Information Scraper")
	fmt.Println("=================================================")

	// Load configuration
	var cfg *config.Config
	var err error

	if *configFile != "" {
		// Load configuration from file
		cfg, err = config.LoadConfig(*configFile)
		if err != nil {
			log.Fatalf("Error loading configuration: %v", err)
		}
		log.Printf("Loaded configuration from %s", *configFile)
	} else {
		// Use default configuration
		cfg = config.DefaultConfig()
		log.Println("Using default configuration")
	}

	// Initialize database
	db, err := database.New(&cfg.Database)
	if err != nil {
		log.Fatalf("Error initializing database: %v", err)
	}
	defer db.Close()

	// Migrate database
	if err := db.Migrate(); err != nil {
		log.Fatalf("Error migrating database: %v", err)
	}

	// Create worker pool semaphores
	var llmSemaphore worker.Semaphore

	// Initialize worker pool if enabled
	if cfg.WorkerPool.Enabled {
		// Create a worker pool
		pool := worker.New(&cfg.WorkerPool, nil) // We'll set the processor function later

		// Create semaphores
		llmSemaphore = worker.NewLLMSemaphore(pool)
	} else {
		// Use no-op semaphores
		llmSemaphore = worker.NewNoOpSemaphore()
	}

	// Initialize LLM client
	llmClient, err := llm.New(&cfg.LLM, llmSemaphore)
	if err != nil {
		log.Fatalf("Error initializing LLM client: %v", err)
	}

	// Initialize API clients
	apiClient, err := api.New(&cfg.APIs, db)
	if err != nil {
		log.Fatalf("Error initializing API clients: %v", err)
	}

	// Initialize file operations
	fileOps := fileops.New(&cfg.FileOps)

	// Initialize notification system
	notifier := notification.New(&cfg.Notification, db)

	// Initialize processor
	proc := processor.New(cfg, db, llmClient, apiClient, fileOps, notifier)

	// Initialize scanner
	scan := scanner.New(&cfg.Scanner, db)

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Create context for scanner
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start worker pool if enabled
	if cfg.WorkerPool.Enabled {
		proc.StartWorkerPool()
		defer proc.StopWorkerPool()
	}

	// Start scanner in a goroutine
	go func() {
		log.Printf("Starting scanner with interval of %d minutes", cfg.ScanInterval)

		// Initial scan
		scanAndProcess(ctx, scan, proc, cfg.WorkerPool.Enabled)

		// Setup ticker for periodic scanning
		ticker := time.NewTicker(time.Duration(cfg.ScanInterval) * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				log.Println("Scanner stopping")
				return
			case <-ticker.C:
				scanAndProcess(ctx, scan, proc, cfg.WorkerPool.Enabled)
			}
		}
	}()

	// Start notification sender in a goroutine
	if cfg.Notification.Enabled {
		go func() {
			log.Println("Starting notification sender")

			// Setup ticker for sending notifications
			ticker := time.NewTicker(1 * time.Minute)
			defer ticker.Stop()

			for {
				select {
				case <-ctx.Done():
					log.Println("Notification sender stopping")
					return
				case <-ticker.C:
					if err := notifier.SendPendingNotifications(); err != nil {
						log.Printf("Error sending notifications: %v", err)
					}
				}
			}
		}()
	}

	// Wait for termination signal
	sig := <-sigChan
	log.Printf("Received signal %v, shutting down...", sig)

	// Cancel context to stop goroutines
	cancel()

	// Allow some time for goroutines to clean up
	time.Sleep(2 * time.Second)

	log.Println("Shutdown complete")
}

// scanAndProcess scans for new files and processes them
func scanAndProcess(ctx context.Context, scan *scanner.Scanner, proc *processor.Processor, useWorkerPool bool) {
	log.Println("Scanning for new files...")

	// Scan for new files
	result, err := scan.Scan()
	if err != nil {
		log.Printf("Error scanning for new files: %v", err)
		return
	}

	log.Printf("Found %d new files", len(result.NewFiles))

	// Process batch directories
	for dir, files := range result.BatchDirs {
		log.Printf("Processing batch directory: %s (%d files)", dir, len(files))

		// Create batch process
		batchProcess, err := scan.CreateBatchProcess(dir, files)
		if err != nil {
			log.Printf("Error creating batch process: %v", err)
			continue
		}

		if useWorkerPool {
			// Queue batch process for processing
			if err := proc.QueueBatchProcess(batchProcess); err != nil {
				log.Printf("Error queuing batch process: %v", err)
			}
		} else {
			// Process batch directly
			if err := proc.ProcessBatchFiles(ctx, batchProcess); err != nil {
				log.Printf("Error processing batch: %v", err)
			}
		}
	}

	// Process individual files (not in batch directories)
	for _, file := range result.NewFiles {
		// Skip files that are in batch directories
		inBatch := false
		for dir := range result.BatchDirs {
			if strings.HasPrefix(file, dir) {
				inBatch = true
				break
			}
		}

		if inBatch {
			continue
		}

		log.Printf("Processing individual file: %s", file)

		// Create media file
		mediaFile, err := scan.CreateMediaFile(file)
		if err != nil {
			log.Printf("Error creating media file record: %v", err)
			continue
		}

		if useWorkerPool {
			// Queue media file for processing
			if err := proc.QueueMediaFile(mediaFile); err != nil {
				log.Printf("Error queuing media file: %v", err)
			}
		} else {
			// Process media file directly
			if err := proc.ProcessMediaFile(ctx, mediaFile); err != nil {
				log.Printf("Error processing media file: %v", err)
			}
		}
	}
}
