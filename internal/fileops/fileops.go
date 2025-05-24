package fileops

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/sleepstars/mediascanner/internal/config"
)

// FileOps represents the file operations
type FileOps struct {
	config *config.FileOpsConfig
}

// New creates a new file operations instance
func New(cfg *config.FileOpsConfig) *FileOps {
	return &FileOps{
		config: cfg,
	}
}

// ProcessFile processes a file (copy, move, or symlink)
func (f *FileOps) ProcessFile(sourcePath, destDir string) (string, error) {
	// Validate input parameters
	if sourcePath == "" {
		return "", fmt.Errorf("source path cannot be empty")
	}
	if destDir == "" {
		return "", fmt.Errorf("destination directory cannot be empty")
	}

	// Check if source file exists
	sourceInfo, err := os.Stat(sourcePath)
	if err != nil {
		return "", fmt.Errorf("error getting file info: %w", err)
	}

	// Ensure source is a regular file
	if !sourceInfo.Mode().IsRegular() {
		return "", fmt.Errorf("source is not a regular file: %s", sourcePath)
	}

	// Create destination directory if it doesn't exist
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return "", fmt.Errorf("error creating destination directory: %w", err)
	}

	// Get file extension
	ext := filepath.Ext(sourcePath)

	// Get file name without extension
	fileName := filepath.Base(sourcePath)
	fileNameWithoutExt := strings.TrimSuffix(fileName, ext)

	// Create destination file path
	destPath := filepath.Join(destDir, fileNameWithoutExt+ext)

	// Check if destination file already exists
	if _, err := os.Stat(destPath); err == nil {
		// File already exists, append a suffix
		for i := 1; ; i++ {
			destPath = filepath.Join(destDir, fmt.Sprintf("%s (%d)%s", fileNameWithoutExt, i, ext))
			if _, err := os.Stat(destPath); os.IsNotExist(err) {
				break
			}
		}
	}

	// Process the file based on the configured mode
	switch f.config.Mode {
	case "copy":
		if err := copyFile(sourcePath, destPath); err != nil {
			return "", fmt.Errorf("error copying file: %w", err)
		}
	case "move":
		if err := moveFile(sourcePath, destPath); err != nil {
			return "", fmt.Errorf("error moving file: %w", err)
		}
	case "symlink":
		if err := symlinkFile(sourcePath, destPath); err != nil {
			return "", fmt.Errorf("error creating symlink: %w", err)
		}
	default:
		return "", fmt.Errorf("unknown file operation mode: %s", f.config.Mode)
	}

	return destPath, nil
}

// copyFile copies a file from source to destination
func copyFile(sourcePath, destPath string) error {
	// Open source file
	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("error opening source file: %w", err)
	}
	defer sourceFile.Close()

	// Create destination file
	destFile, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("error creating destination file: %w", err)
	}
	defer destFile.Close()

	// Copy the file
	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return fmt.Errorf("error copying file: %w", err)
	}

	// Sync the file to ensure it's written to disk
	if err := destFile.Sync(); err != nil {
		return fmt.Errorf("error syncing file: %w", err)
	}

	// Get source file info
	sourceInfo, err := os.Stat(sourcePath)
	if err != nil {
		return fmt.Errorf("error getting source file info: %w", err)
	}

	// Set file permissions
	if err := os.Chmod(destPath, sourceInfo.Mode()); err != nil {
		return fmt.Errorf("error setting file permissions: %w", err)
	}

	return nil
}

// moveFile moves a file from source to destination
func moveFile(sourcePath, destPath string) error {
	// Try to rename the file (this works if source and destination are on the same filesystem)
	err := os.Rename(sourcePath, destPath)
	if err == nil {
		return nil
	}

	// If rename fails, copy the file and then delete the source
	if err := copyFile(sourcePath, destPath); err != nil {
		return fmt.Errorf("error copying file: %w", err)
	}

	// Delete the source file
	if err := os.Remove(sourcePath); err != nil {
		return fmt.Errorf("error removing source file: %w", err)
	}

	return nil
}

// symlinkFile creates a symlink from source to destination
func symlinkFile(sourcePath, destPath string) error {
	// Get absolute path of source file
	absSourcePath, err := filepath.Abs(sourcePath)
	if err != nil {
		return fmt.Errorf("error getting absolute path: %w", err)
	}

	// Create symlink
	if err := os.Symlink(absSourcePath, destPath); err != nil {
		return fmt.Errorf("error creating symlink: %w", err)
	}

	return nil
}

// DownloadImage downloads an image from a URL
func (f *FileOps) DownloadImage(url, destPath string) error {
	// Create destination directory if it doesn't exist
	destDir := filepath.Dir(destPath)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("error creating destination directory: %w", err)
	}

	// TODO: Implement image downloading
	return nil
}

// CreateNFOFile creates an NFO file
func (f *FileOps) CreateNFOFile(destPath string, content string) error {
	// Create destination directory if it doesn't exist
	destDir := filepath.Dir(destPath)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("error creating destination directory: %w", err)
	}

	// Create the file
	file, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("error creating NFO file: %w", err)
	}
	defer file.Close()

	// Write the content
	_, err = file.WriteString(content)
	if err != nil {
		return fmt.Errorf("error writing NFO file: %w", err)
	}

	return nil
}
