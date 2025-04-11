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
		cfg, err = config.LoadConfig(*configFile)
		if err != nil {
			log.Fatalf("Error loading configuration: %v", err)
		}
	} else {
		cfg = config.DefaultConfig()
	}

	// Initialize database
	db, err := database.New(cfg)
	if err != nil {
		log.Fatalf("Error initializing database: %v", err)
	}

	// Migrate database
	if err := db.Migrate(); err != nil {
		log.Fatalf("Error migrating database: %v", err)
	}

	// Initialize LLM client
	llmClient, err := llm.New(&cfg.LLM)
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

	// Initialize scanner
	scanner := scanner.New(&cfg.Scanner, db)

	// Initialize processor
	processor := processor.New(cfg, db, llmClient, apiClient, fileOps, notifier)

	// Setup signal handling for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		log.Printf("Received signal %v, shutting down...", sig)
		cancel()
	}()

	// Start the main loop
	log.Println("Starting MediaScanner...")
	mainLoop(ctx, cfg, scanner, processor, notifier)
}

// mainLoop runs the main processing loop
func mainLoop(ctx context.Context, cfg *config.Config, scanner *scanner.Scanner, processor *processor.Processor, notifier *notification.Notifier) {
	ticker := time.NewTicker(time.Duration(cfg.ScanInterval) * time.Minute)
	defer ticker.Stop()

	// Run once immediately
	runScanAndProcess(ctx, scanner, processor, notifier)

	// Then run on the ticker
	for {
		select {
		case <-ctx.Done():
			log.Println("Shutting down...")
			return
		case <-ticker.C:
			runScanAndProcess(ctx, scanner, processor, notifier)
		}
	}
}

// runScanAndProcess runs a scan and processes the results
func runScanAndProcess(ctx context.Context, scanner *scanner.Scanner, processor *processor.Processor, notifier *notification.Notifier) {
	// Send pending notifications
	if err := notifier.SendPendingNotifications(); err != nil {
		log.Printf("Error sending pending notifications: %v", err)
	}

	// Scan for new files
	log.Println("Scanning for new media files...")
	result, err := scanner.Scan()
	if err != nil {
		log.Printf("Error scanning for media files: %v", err)
		return
	}

	log.Printf("Found %d new files, %d batch directories, %d excluded files", len(result.NewFiles), len(result.BatchDirs), len(result.ExcludedFiles))

	// Process batch directories
	for dir, files := range result.BatchDirs {
		log.Printf("Creating batch process for directory: %s (%d files)", dir, len(files))
		batchProcess, err := scanner.CreateBatchProcess(dir, files)
		if err != nil {
			log.Printf("Error creating batch process: %v", err)
			continue
		}

		// Process the batch
		if err := processor.ProcessBatchFiles(ctx, batchProcess); err != nil {
			log.Printf("Error processing batch: %v", err)
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

		// Create media file record
		mediaFile, err := scanner.CreateMediaFile(file)
		if err != nil {
			log.Printf("Error creating media file record: %v", err)
			continue
		}

		// Process the file
		if err := processor.ProcessMediaFile(ctx, mediaFile); err != nil {
			log.Printf("Error processing media file: %v", err)
		}
	}
}
