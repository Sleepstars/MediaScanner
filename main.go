// Package main provides a simple entry point that redirects to the actual implementation.
// The real implementation is in cmd/mediascanner/main.go
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	// Get the directory of the current executable
	execPath, err := os.Executable()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting executable path: %v\n", err)
		os.Exit(1)
	}

	// Construct path to the actual mediascanner binary
	actualBinary := filepath.Join(filepath.Dir(execPath), "cmd", "mediascanner", "mediascanner")

	// If the actual binary doesn't exist, try building it
	if _, err := os.Stat(actualBinary); os.IsNotExist(err) {
		fmt.Println("Building MediaScanner...")
		buildCmd := exec.Command("go", "build", "-o", actualBinary, "./cmd/mediascanner")
		buildCmd.Stdout = os.Stdout
		buildCmd.Stderr = os.Stderr
		if err := buildCmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error building MediaScanner: %v\n", err)
			os.Exit(1)
		}
	}

	// Execute the actual binary with all arguments
	cmd := exec.Command(actualBinary, os.Args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			os.Exit(exitError.ExitCode())
		}
		fmt.Fprintf(os.Stderr, "Error running MediaScanner: %v\n", err)
		os.Exit(1)
	}
}
