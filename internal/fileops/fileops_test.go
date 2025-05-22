package fileops

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sleepstars/mediascanner/internal/config"
	"github.com/sleepstars/mediascanner/internal/testutil"
)

func TestNew(t *testing.T) {
	cfg := &config.FileOpsConfig{
		Mode: "copy",
	}
	
	fileOps := New(cfg)
	
	if fileOps == nil {
		t.Fatal("Expected non-nil FileOps instance")
	}
	
	if fileOps.config != cfg {
		t.Errorf("Expected config to be %v, got %v", cfg, fileOps.config)
	}
}

func TestProcessFile_Copy(t *testing.T) {
	// Create temp directories for source and destination
	sourceDir := testutil.CreateTempDir(t)
	destDir := testutil.CreateTempDir(t)
	
	// Create a test file
	testContent := "test content"
	sourcePath := testutil.CreateTestFile(t, sourceDir, "test.mp4", testContent)
	
	// Create FileOps with copy mode
	fileOps := New(&config.FileOpsConfig{
		Mode: "copy",
	})
	
	// Process the file
	destPath, err := fileOps.ProcessFile(sourcePath, destDir)
	if err != nil {
		t.Fatalf("ProcessFile failed: %v", err)
	}
	
	// Verify the destination file exists
	if _, err := os.Stat(destPath); os.IsNotExist(err) {
		t.Errorf("Destination file %s does not exist", destPath)
	}
	
	// Verify the source file still exists
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		t.Errorf("Source file %s should still exist after copy", sourcePath)
	}
	
	// Verify the content of the destination file
	destContent, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatalf("Failed to read destination file: %v", err)
	}
	
	if string(destContent) != testContent {
		t.Errorf("Expected destination content to be %q, got %q", testContent, string(destContent))
	}
}

func TestProcessFile_Move(t *testing.T) {
	// Create temp directories for source and destination
	sourceDir := testutil.CreateTempDir(t)
	destDir := testutil.CreateTempDir(t)
	
	// Create a test file
	testContent := "test content"
	sourcePath := testutil.CreateTestFile(t, sourceDir, "test.mp4", testContent)
	
	// Create FileOps with move mode
	fileOps := New(&config.FileOpsConfig{
		Mode: "move",
	})
	
	// Process the file
	destPath, err := fileOps.ProcessFile(sourcePath, destDir)
	if err != nil {
		t.Fatalf("ProcessFile failed: %v", err)
	}
	
	// Verify the destination file exists
	if _, err := os.Stat(destPath); os.IsNotExist(err) {
		t.Errorf("Destination file %s does not exist", destPath)
	}
	
	// Verify the source file no longer exists
	if _, err := os.Stat(sourcePath); !os.IsNotExist(err) {
		t.Errorf("Source file %s should not exist after move", sourcePath)
	}
	
	// Verify the content of the destination file
	destContent, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatalf("Failed to read destination file: %v", err)
	}
	
	if string(destContent) != testContent {
		t.Errorf("Expected destination content to be %q, got %q", testContent, string(destContent))
	}
}

func TestProcessFile_Symlink(t *testing.T) {
	// Skip on Windows as symlinks require admin privileges
	if os.Getenv("OS") == "Windows_NT" {
		t.Skip("Skipping symlink test on Windows")
	}
	
	// Create temp directories for source and destination
	sourceDir := testutil.CreateTempDir(t)
	destDir := testutil.CreateTempDir(t)
	
	// Create a test file
	testContent := "test content"
	sourcePath := testutil.CreateTestFile(t, sourceDir, "test.mp4", testContent)
	
	// Create FileOps with symlink mode
	fileOps := New(&config.FileOpsConfig{
		Mode: "symlink",
	})
	
	// Process the file
	destPath, err := fileOps.ProcessFile(sourcePath, destDir)
	if err != nil {
		t.Fatalf("ProcessFile failed: %v", err)
	}
	
	// Verify the destination symlink exists
	if _, err := os.Stat(destPath); os.IsNotExist(err) {
		t.Errorf("Destination symlink %s does not exist", destPath)
	}
	
	// Verify it's a symlink
	fi, err := os.Lstat(destPath)
	if err != nil {
		t.Fatalf("Failed to get file info: %v", err)
	}
	
	if fi.Mode()&os.ModeSymlink == 0 {
		t.Errorf("Expected %s to be a symlink", destPath)
	}
	
	// Verify the source file still exists
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		t.Errorf("Source file %s should still exist after symlink", sourcePath)
	}
	
	// Verify the content through the symlink
	destContent, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatalf("Failed to read destination file: %v", err)
	}
	
	if string(destContent) != testContent {
		t.Errorf("Expected destination content to be %q, got %q", testContent, string(destContent))
	}
}

func TestProcessFile_FileAlreadyExists(t *testing.T) {
	// Create temp directories for source and destination
	sourceDir := testutil.CreateTempDir(t)
	destDir := testutil.CreateTempDir(t)
	
	// Create a test file
	testContent := "test content"
	sourcePath := testutil.CreateTestFile(t, sourceDir, "test.mp4", testContent)
	
	// Create a file with the same name in the destination directory
	destFilePath := filepath.Join(destDir, "test.mp4")
	err := os.WriteFile(destFilePath, []byte("existing content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create existing file: %v", err)
	}
	
	// Create FileOps with copy mode
	fileOps := New(&config.FileOpsConfig{
		Mode: "copy",
	})
	
	// Process the file
	destPath, err := fileOps.ProcessFile(sourcePath, destDir)
	if err != nil {
		t.Fatalf("ProcessFile failed: %v", err)
	}
	
	// Verify the destination path is different from the original
	expectedPath := filepath.Join(destDir, "test (1).mp4")
	if destPath != expectedPath {
		t.Errorf("Expected destination path to be %q, got %q", expectedPath, destPath)
	}
	
	// Verify both files exist
	if _, err := os.Stat(destFilePath); os.IsNotExist(err) {
		t.Errorf("Original destination file %s does not exist", destFilePath)
	}
	
	if _, err := os.Stat(destPath); os.IsNotExist(err) {
		t.Errorf("New destination file %s does not exist", destPath)
	}
}

func TestProcessFile_InvalidMode(t *testing.T) {
	// Create temp directories for source and destination
	sourceDir := testutil.CreateTempDir(t)
	destDir := testutil.CreateTempDir(t)
	
	// Create a test file
	sourcePath := testutil.CreateTestFile(t, sourceDir, "test.mp4", "test content")
	
	// Create FileOps with invalid mode
	fileOps := New(&config.FileOpsConfig{
		Mode: "invalid",
	})
	
	// Process the file
	_, err := fileOps.ProcessFile(sourcePath, destDir)
	if err == nil {
		t.Fatal("Expected error for invalid mode, got nil")
	}
	
	expectedError := "unknown file operation mode: invalid"
	if err.Error() != expectedError {
		t.Errorf("Expected error message %q, got %q", expectedError, err.Error())
	}
}

func TestProcessFile_SourceNotExists(t *testing.T) {
	// Create temp directory for destination
	destDir := testutil.CreateTempDir(t)
	
	// Non-existent source path
	sourcePath := "/non/existent/path/test.mp4"
	
	// Create FileOps with copy mode
	fileOps := New(&config.FileOpsConfig{
		Mode: "copy",
	})
	
	// Process the file
	_, err := fileOps.ProcessFile(sourcePath, destDir)
	if err == nil {
		t.Fatal("Expected error for non-existent source, got nil")
	}
}

func TestCreateNFOFile(t *testing.T) {
	// Create temp directory for destination
	destDir := testutil.CreateTempDir(t)
	
	// Create FileOps
	fileOps := New(&config.FileOpsConfig{
		Mode: "copy",
	})
	
	// Test content
	content := "<movie>\n<title>Test Movie</title>\n</movie>"
	
	// Create NFO file
	destPath := filepath.Join(destDir, "movie.nfo")
	err := fileOps.CreateNFOFile(destPath, content)
	if err != nil {
		t.Fatalf("CreateNFOFile failed: %v", err)
	}
	
	// Verify the file exists
	if _, err := os.Stat(destPath); os.IsNotExist(err) {
		t.Errorf("NFO file %s does not exist", destPath)
	}
	
	// Verify the content
	fileContent, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatalf("Failed to read NFO file: %v", err)
	}
	
	if string(fileContent) != content {
		t.Errorf("Expected NFO content to be %q, got %q", content, string(fileContent))
	}
}

func TestCreateNFOFile_DirectoryNotExists(t *testing.T) {
	// Create temp directory for base
	baseDir := testutil.CreateTempDir(t)
	
	// Non-existent subdirectory
	destDir := filepath.Join(baseDir, "subdir1", "subdir2")
	destPath := filepath.Join(destDir, "movie.nfo")
	
	// Create FileOps
	fileOps := New(&config.FileOpsConfig{
		Mode: "copy",
	})
	
	// Create NFO file
	content := "<movie>\n<title>Test Movie</title>\n</movie>"
	err := fileOps.CreateNFOFile(destPath, content)
	if err != nil {
		t.Fatalf("CreateNFOFile failed: %v", err)
	}
	
	// Verify the file exists
	if _, err := os.Stat(destPath); os.IsNotExist(err) {
		t.Errorf("NFO file %s does not exist", destPath)
	}
	
	// Verify the content
	fileContent, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatalf("Failed to read NFO file: %v", err)
	}
	
	if string(fileContent) != content {
		t.Errorf("Expected NFO content to be %q, got %q", content, string(fileContent))
	}
}
