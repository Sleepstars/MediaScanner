package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file if it exists
	_ = godotenv.Load()

	// Parse command line flags
	// configFile := flag.String("config", "", "Path to configuration file")
	flag.Parse()

	// TODO: Use configFile when implementing configuration loading

	// Print banner
	fmt.Println("MediaScanner - LLM-based Media Information Scraper")
	fmt.Println("=================================================")

	// TODO: Initialize configuration
	// TODO: Initialize database
	// TODO: Initialize LLM client
	// TODO: Initialize API clients
	// TODO: Initialize scanner
	// TODO: Initialize processor
	// TODO: Initialize notification system

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// TODO: Start the scanner

	// Wait for termination signal
	sig := <-sigChan
	log.Printf("Received signal %v, shutting down...", sig)

	// TODO: Perform cleanup
}
