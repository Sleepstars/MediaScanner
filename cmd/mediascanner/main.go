package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
	"github.com/sleepstars/mediascanner/internal/api"
	"github.com/sleepstars/mediascanner/internal/config"
	"github.com/sleepstars/mediascanner/internal/database"
	"github.com/sleepstars/mediascanner/internal/fileops"
	"github.com/sleepstars/mediascanner/internal/llm"
	"github.com/sleepstars/mediascanner/internal/logger"
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
			log.Fatal().Err(err).Str("config_file", *configFile).Msg("Failed to load configuration")
		}
		log.Info().Str("config_file", *configFile).Msg("Configuration loaded successfully")
	} else {
		// Use default configuration
		cfg = config.DefaultConfig()
		log.Info().Msg("Using default configuration")
	}

	// Initialize logger with configuration
	loggerConfig := &logger.Config{
		Level:      logger.LogLevel(cfg.Logger.Level),
		Format:     cfg.Logger.Format,
		Output:     cfg.Logger.Output,
		File:       cfg.Logger.File,
		MaxSize:    cfg.Logger.MaxSize,
		MaxBackups: cfg.Logger.MaxBackups,
		MaxAge:     cfg.Logger.MaxAge,
		Compress:   cfg.Logger.Compress,
	}
	logger.Init(loggerConfig)
	log.Info().Msg("Logger initialized successfully")

	// Initialize database
	db, err := database.New(&cfg.Database)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize database")
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Error().Err(err).Msg("Error closing database connection")
		}
	}()

	// Migrate database
	if err := db.Migrate(); err != nil {
		log.Fatal().Err(err).Msg("Failed to migrate database")
	}
	log.Info().Msg("Database migration completed successfully")

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
		log.Fatal().Err(err).Msg("Failed to initialize LLM client")
	}
	log.Info().Str("provider", cfg.LLM.Provider).Str("model", cfg.LLM.Model).Msg("LLM client initialized successfully")

	// Initialize API clients
	apiClient, err := api.New(&cfg.APIs, db)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize API clients")
	}
	log.Info().Msg("API clients initialized successfully")

	// Initialize file operations
	fileOps := fileops.New(&cfg.FileOps)
	log.Info().Str("mode", cfg.FileOps.Mode).Str("destination", cfg.FileOps.DestinationRoot).Msg("File operations initialized")

	// Initialize notification system
	notifier := notification.New(&cfg.Notification, db)
	log.Info().Bool("enabled", cfg.Notification.Enabled).Str("provider", cfg.Notification.Provider).Msg("Notification system initialized")

	// Initialize processor
	proc := processor.New(cfg, db, llmClient, apiClient, fileOps, notifier)
	log.Info().Msg("Processor initialized successfully")

	// Initialize scanner
	scan := scanner.New(&cfg.Scanner, db)
	log.Info().Int("media_dirs", len(cfg.Scanner.MediaDirs)).Bool("use_watcher", cfg.Scanner.UseWatcher).Msg("Scanner initialized successfully")

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
		defer func() {
			if r := recover(); r != nil {
				log.Error().Interface("panic", r).Msg("Scanner goroutine panicked")
			}
		}()

		// Check if file system watcher is enabled
		if cfg.Scanner.UseWatcher {
			log.Info().Msg("Starting file system watcher")

			// Start the watcher
			if err := scan.StartWatching(); err != nil {
				log.Error().Err(err).Msg("Failed to start file system watcher, falling back to periodic scanning")
				startPeriodicScanner(ctx, scan, proc, cfg)
				return
			}

			log.Info().Msg("File system watcher started successfully")

			// Wait for context cancellation
			<-ctx.Done()

			// Stop the watcher
			if err := scan.StopWatching(); err != nil {
				log.Error().Err(err).Msg("Error stopping file system watcher")
			} else {
				log.Info().Msg("File system watcher stopped successfully")
			}
		} else {
			// Use periodic scanning
			log.Info().Msg("Using periodic scanning mode")
			startPeriodicScanner(ctx, scan, proc, cfg)
		}
	}()

	// Start notification sender in a goroutine
	if cfg.Notification.Enabled {
		go func() {
			defer func() {
				if r := recover(); r != nil {
					log.Error().Interface("panic", r).Msg("Notification sender goroutine panicked")
				}
			}()

			log.Info().Msg("Starting notification sender")

			// Setup ticker for sending notifications
			ticker := time.NewTicker(1 * time.Minute)
			defer ticker.Stop()

			for {
				select {
				case <-ctx.Done():
					log.Info().Msg("Notification sender stopping")
					return
				case <-ticker.C:
					if err := notifier.SendPendingNotifications(); err != nil {
						log.Error().Err(err).Msg("Error sending notifications")
					}
				}
			}
		}()
	} else {
		log.Info().Msg("Notification system disabled")
	}

	log.Info().Msg("MediaScanner started successfully, waiting for termination signal")

	// Wait for termination signal
	sig := <-sigChan
	log.Info().Str("signal", sig.String()).Msg("Received termination signal, initiating graceful shutdown")

	// Cancel context to stop goroutines
	cancel()

	// Allow some time for goroutines to clean up
	log.Info().Msg("Waiting for goroutines to finish...")
	time.Sleep(2 * time.Second)

	log.Info().Msg("Shutdown completed successfully")
}

// startPeriodicScanner starts a periodic scanner that scans for new files at regular intervals
func startPeriodicScanner(ctx context.Context, scan *scanner.Scanner, proc *processor.Processor, cfg *config.Config) {
	log.Info().Int("interval_minutes", cfg.ScanInterval).Msg("Starting periodic scanner")

	// Initial scan
	scanAndProcess(ctx, scan, proc, cfg.WorkerPool.Enabled)

	// Setup ticker for periodic scanning
	ticker := time.NewTicker(time.Duration(cfg.ScanInterval) * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("Periodic scanner stopping")
			return
		case <-ticker.C:
			scanAndProcess(ctx, scan, proc, cfg.WorkerPool.Enabled)
		}
	}
}

// scanAndProcess scans for new files and processes them
func scanAndProcess(ctx context.Context, scan *scanner.Scanner, proc *processor.Processor, useWorkerPool bool) {
	log.Info().Msg("Starting scan for new files")

	// Scan for new files
	result, err := scan.Scan()
	if err != nil {
		log.Error().Err(err).Msg("Failed to scan for new files")
		return
	}

	log.Info().
		Int("new_files", len(result.NewFiles)).
		Int("batch_dirs", len(result.BatchDirs)).
		Int("excluded_files", len(result.ExcludedFiles)).
		Msg("Scan completed")

	// Process batch directories
	for dir, files := range result.BatchDirs {
		log.Info().Str("directory", dir).Int("file_count", len(files)).Msg("Processing batch directory")

		// Create batch process
		batchProcess, err := scan.CreateBatchProcess(dir, files)
		if err != nil {
			log.Error().Err(err).Str("directory", dir).Msg("Failed to create batch process")
			continue
		}

		if useWorkerPool {
			// Queue batch process for processing
			if err := proc.QueueBatchProcess(batchProcess); err != nil {
				log.Error().Err(err).Str("directory", dir).Msg("Failed to queue batch process")
			} else {
				log.Info().Str("directory", dir).Msg("Batch process queued successfully")
			}
		} else {
			// Process batch directly
			if err := proc.ProcessBatchFiles(ctx, batchProcess); err != nil {
				log.Error().Err(err).Str("directory", dir).Msg("Failed to process batch")
			} else {
				log.Info().Str("directory", dir).Msg("Batch processed successfully")
			}
		}
	}

	// Process individual files (not in batch directories)
	individualFiles := 0
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

		individualFiles++
		log.Info().Str("file", file).Msg("Processing individual file")

		// Create media file
		mediaFile, err := scan.CreateMediaFile(file)
		if err != nil {
			log.Error().Err(err).Str("file", file).Msg("Failed to create media file record")
			continue
		}

		if useWorkerPool {
			// Queue media file for processing
			if err := proc.QueueMediaFile(mediaFile); err != nil {
				log.Error().Err(err).Str("file", file).Msg("Failed to queue media file")
			} else {
				log.Debug().Str("file", file).Msg("Media file queued successfully")
			}
		} else {
			// Process media file directly
			if err := proc.ProcessMediaFile(ctx, mediaFile); err != nil {
				log.Error().Err(err).Str("file", file).Msg("Failed to process media file")
			} else {
				log.Info().Str("file", file).Msg("Media file processed successfully")
			}
		}
	}

	log.Info().Int("individual_files", individualFiles).Msg("Individual file processing completed")
}
