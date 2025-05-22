package scanner

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/sleepstars/mediascanner/internal/config"
	"github.com/sleepstars/mediascanner/internal/database"
	"github.com/sleepstars/mediascanner/internal/models"
	"github.com/sleepstars/mediascanner/internal/testutil"
)

// Ensure we're using the database package
var _ database.DatabaseInterface = (*MockDB)(nil)

// MockDB is a mock implementation of the database.DatabaseInterface
type MockDB struct {
	GetMediaFileByPathFunc      func(path string) (*models.MediaFile, error)
	CreateMediaFileFunc         func(file *models.MediaFile) error
	UpdateMediaFileFunc         func(file *models.MediaFile) error
	GetMediaFileByIDFunc        func(id int64) (*models.MediaFile, error)
	CreateMediaInfoFunc         func(info *models.MediaInfo) error
	GetMediaInfoByMediaFileIDFunc func(mediaFileID int64) (*models.MediaInfo, error)
	UpdateMediaInfoFunc         func(info *models.MediaInfo) error
	GetAPICacheFunc             func(provider, query string) (*models.APICache, error)
	CreateAPICacheFunc          func(cache *models.APICache) error
	CreateLLMRequestFunc        func(request *models.LLMRequest) error
	CreateBatchProcessFunc      func(batch *models.BatchProcess) error
	UpdateBatchProcessFunc      func(batch *models.BatchProcess) error
	CreateBatchProcessFileFunc  func(file *models.BatchProcessFile) error
	UpdateBatchProcessFileFunc  func(file *models.BatchProcessFile) error
	GetBatchProcessFilesByBatchIDFunc func(batchID int64) ([]models.BatchProcessFile, error)
	CreateNotificationFunc      func(notification *models.Notification) error
	GetPendingNotificationsFunc func() ([]models.Notification, error)
	UpdateNotificationFunc      func(notification *models.Notification) error
	GetPendingMediaFilesFunc    func() ([]models.MediaFile, error)
	GetMediaFilesByDirectoryFunc func(directory string) ([]models.MediaFile, error)
	GetPendingBatchProcessesFunc func() ([]models.BatchProcess, error)
	CloseFunc                   func() error
}

func (m *MockDB) GetMediaFileByPath(path string) (*models.MediaFile, error) {
	return m.GetMediaFileByPathFunc(path)
}

func (m *MockDB) CreateMediaFile(file *models.MediaFile) error {
	return m.CreateMediaFileFunc(file)
}

func (m *MockDB) UpdateMediaFile(file *models.MediaFile) error {
	if m.UpdateMediaFileFunc != nil {
		return m.UpdateMediaFileFunc(file)
	}
	return nil
}

func (m *MockDB) GetMediaFileByID(id int64) (*models.MediaFile, error) {
	if m.GetMediaFileByIDFunc != nil {
		return m.GetMediaFileByIDFunc(id)
	}
	return nil, errors.New("not implemented")
}

func (m *MockDB) CreateMediaInfo(info *models.MediaInfo) error {
	if m.CreateMediaInfoFunc != nil {
		return m.CreateMediaInfoFunc(info)
	}
	return nil
}

func (m *MockDB) GetMediaInfoByMediaFileID(mediaFileID int64) (*models.MediaInfo, error) {
	if m.GetMediaInfoByMediaFileIDFunc != nil {
		return m.GetMediaInfoByMediaFileIDFunc(mediaFileID)
	}
	return nil, errors.New("not implemented")
}

func (m *MockDB) UpdateMediaInfo(info *models.MediaInfo) error {
	if m.UpdateMediaInfoFunc != nil {
		return m.UpdateMediaInfoFunc(info)
	}
	return nil
}

func (m *MockDB) GetAPICache(provider, query string) (*models.APICache, error) {
	if m.GetAPICacheFunc != nil {
		return m.GetAPICacheFunc(provider, query)
	}
	return nil, errors.New("not implemented")
}

func (m *MockDB) CreateAPICache(cache *models.APICache) error {
	if m.CreateAPICacheFunc != nil {
		return m.CreateAPICacheFunc(cache)
	}
	return nil
}

func (m *MockDB) CreateLLMRequest(request *models.LLMRequest) error {
	if m.CreateLLMRequestFunc != nil {
		return m.CreateLLMRequestFunc(request)
	}
	return nil
}

func (m *MockDB) CreateBatchProcess(batch *models.BatchProcess) error {
	return m.CreateBatchProcessFunc(batch)
}

func (m *MockDB) UpdateBatchProcess(batch *models.BatchProcess) error {
	if m.UpdateBatchProcessFunc != nil {
		return m.UpdateBatchProcessFunc(batch)
	}
	return nil
}

func (m *MockDB) CreateBatchProcessFile(file *models.BatchProcessFile) error {
	return m.CreateBatchProcessFileFunc(file)
}

func (m *MockDB) UpdateBatchProcessFile(file *models.BatchProcessFile) error {
	if m.UpdateBatchProcessFileFunc != nil {
		return m.UpdateBatchProcessFileFunc(file)
	}
	return nil
}

func (m *MockDB) GetBatchProcessFilesByBatchID(batchID int64) ([]models.BatchProcessFile, error) {
	if m.GetBatchProcessFilesByBatchIDFunc != nil {
		return m.GetBatchProcessFilesByBatchIDFunc(batchID)
	}
	return nil, errors.New("not implemented")
}

func (m *MockDB) CreateNotification(notification *models.Notification) error {
	if m.CreateNotificationFunc != nil {
		return m.CreateNotificationFunc(notification)
	}
	return nil
}

func (m *MockDB) GetPendingNotifications() ([]models.Notification, error) {
	if m.GetPendingNotificationsFunc != nil {
		return m.GetPendingNotificationsFunc()
	}
	return nil, errors.New("not implemented")
}

func (m *MockDB) UpdateNotification(notification *models.Notification) error {
	if m.UpdateNotificationFunc != nil {
		return m.UpdateNotificationFunc(notification)
	}
	return nil
}

func (m *MockDB) GetPendingMediaFiles() ([]models.MediaFile, error) {
	if m.GetPendingMediaFilesFunc != nil {
		return m.GetPendingMediaFilesFunc()
	}
	return nil, errors.New("not implemented")
}

func (m *MockDB) GetMediaFilesByDirectory(directory string) ([]models.MediaFile, error) {
	if m.GetMediaFilesByDirectoryFunc != nil {
		return m.GetMediaFilesByDirectoryFunc(directory)
	}
	return nil, errors.New("not implemented")
}

func (m *MockDB) GetPendingBatchProcesses() ([]models.BatchProcess, error) {
	if m.GetPendingBatchProcessesFunc != nil {
		return m.GetPendingBatchProcessesFunc()
	}
	return nil, errors.New("not implemented")
}

func (m *MockDB) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}



func TestNew(t *testing.T) {
	cfg := &config.ScannerConfig{
		MediaDirs:       []string{"/test/media"},
		VideoExtensions: []string{".mp4", ".mkv", ".avi"},
		ExcludePatterns: []string{"sample", "trailer"},
		ExcludeDirs:     []string{"_UNPACK_", "tmp"},
		BatchThreshold:  3,
	}

	// Create a mock database
	db := &MockDB{}

	// Create a new scanner with the mock database
	scanner := New(cfg, db)

	if scanner == nil {
		t.Fatal("Expected non-nil Scanner instance")
	}

	if scanner.config != cfg {
		t.Errorf("Expected config to be %v, got %v", cfg, scanner.config)
	}

	// Compare the database
	if scanner.db != db {
		t.Errorf("Expected db to be %v, got %v", db, scanner.db)
	}
}

func TestScan(t *testing.T) {
	// Create temp directory structure
	baseDir := testutil.CreateTempDir(t)

	// Create media directories
	mediaDir := filepath.Join(baseDir, "media")
	if err := os.MkdirAll(mediaDir, 0755); err != nil {
		t.Fatalf("Failed to create media directory: %v", err)
	}

	// Create subdirectories
	moviesDir := filepath.Join(mediaDir, "movies")
	if err := os.MkdirAll(moviesDir, 0755); err != nil {
		t.Fatalf("Failed to create movies directory: %v", err)
	}

	excludeDir := filepath.Join(mediaDir, "_UNPACK_")
	if err := os.MkdirAll(excludeDir, 0755); err != nil {
		t.Fatalf("Failed to create exclude directory: %v", err)
	}

	// Create test files
	testutil.CreateTestFile(t, moviesDir, "movie1.mp4", "test content")
	testutil.CreateTestFile(t, moviesDir, "movie2.mkv", "test content")
	testutil.CreateTestFile(t, moviesDir, "movie3.avi", "test content")
	testutil.CreateTestFile(t, moviesDir, "sample.mp4", "test content") // Should be excluded
	testutil.CreateTestFile(t, moviesDir, "trailer.mkv", "test content") // Should be excluded
	testutil.CreateTestFile(t, moviesDir, "document.txt", "test content") // Not a video file
	testutil.CreateTestFile(t, excludeDir, "movie4.mp4", "test content") // In excluded directory

	// Create batch directory
	batchDir := filepath.Join(mediaDir, "batch")
	if err := os.MkdirAll(batchDir, 0755); err != nil {
		t.Fatalf("Failed to create batch directory: %v", err)
	}

	// Create batch files (more than batch threshold)
	testutil.CreateTestFile(t, batchDir, "batch1.mp4", "test content")
	testutil.CreateTestFile(t, batchDir, "batch2.mp4", "test content")
	testutil.CreateTestFile(t, batchDir, "batch3.mp4", "test content")
	testutil.CreateTestFile(t, batchDir, "batch4.mp4", "test content")

	// Create mock database
	db := &MockDB{
		GetMediaFileByPathFunc: func(path string) (*models.MediaFile, error) {
			// Simulate that movie1.mp4 is already in the database
			if filepath.Base(path) == "movie1.mp4" {
				return &models.MediaFile{
					ID:           1,
					OriginalPath: path,
					OriginalName: "movie1.mp4",
					Status:       "success",
				}, nil
			}
			return nil, errors.New("not found")
		},
	}

	// Create scanner
	cfg := &config.ScannerConfig{
		MediaDirs:       []string{mediaDir},
		VideoExtensions: []string{".mp4", ".mkv", ".avi"},
		ExcludePatterns: []string{"sample", "trailer"},
		ExcludeDirs:     []string{"_UNPACK_", "tmp"},
		BatchThreshold:  3,
	}

	scanner := New(cfg, db)

	// Scan for files
	result, err := scanner.Scan()
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	// Verify the result
	// Should find movie2.mkv, movie3.avi, and all batch files
	expectedNewFiles := 6 // movie2.mkv, movie3.avi, batch1.mp4, batch2.mp4, batch3.mp4, batch4.mp4
	if len(result.NewFiles) != expectedNewFiles {
		t.Errorf("Expected %d new files, got %d", expectedNewFiles, len(result.NewFiles))
	}

	// Should find 1 batch directory
	if len(result.BatchDirs) != 1 {
		t.Errorf("Expected 1 batch directory, got %d", len(result.BatchDirs))
	}

	// The batch directory should have 4 files
	batchFiles, ok := result.BatchDirs[batchDir]
	if !ok {
		t.Fatalf("Expected batch directory %q in result", batchDir)
	}

	if len(batchFiles) != 4 {
		t.Errorf("Expected 4 files in batch directory, got %d", len(batchFiles))
	}

	// Should find 2 excluded files
	if len(result.ExcludedFiles) != 2 {
		t.Errorf("Expected 2 excluded files, got %d", len(result.ExcludedFiles))
	}
}

func TestCreateMediaFile(t *testing.T) {
	// Create temp directory
	tempDir := testutil.CreateTempDir(t)

	// Create test file
	filePath := testutil.CreateTestFile(t, tempDir, "test.mp4", "test content")

	// Create mock database
	var createdFile *models.MediaFile
	db := &MockDB{
		CreateMediaFileFunc: func(file *models.MediaFile) error {
			createdFile = file
			file.ID = 1 // Simulate database assigning an ID
			return nil
		},
	}

	// Create scanner
	cfg := &config.ScannerConfig{}
	scanner := New(cfg, db)

	// Create media file
	mediaFile, err := scanner.CreateMediaFile(filePath)
	if err != nil {
		t.Fatalf("CreateMediaFile failed: %v", err)
	}

	// Verify the result
	if mediaFile.ID != 1 {
		t.Errorf("Expected ID to be 1, got %d", mediaFile.ID)
	}

	if mediaFile.OriginalPath != filePath {
		t.Errorf("Expected original path to be %q, got %q", filePath, mediaFile.OriginalPath)
	}

	if mediaFile.OriginalName != "test.mp4" {
		t.Errorf("Expected original name to be 'test.mp4', got %q", mediaFile.OriginalName)
	}

	if mediaFile.Status != "pending" {
		t.Errorf("Expected status to be 'pending', got %q", mediaFile.Status)
	}

	// Verify the file was created in the database
	if createdFile == nil {
		t.Fatal("Expected file to be created in database")
	}

	if createdFile.OriginalPath != filePath {
		t.Errorf("Expected created file path to be %q, got %q", filePath, createdFile.OriginalPath)
	}
}

func TestCreateMediaFile_Error(t *testing.T) {
	// Create mock database that returns an error
	db := &MockDB{
		CreateMediaFileFunc: func(file *models.MediaFile) error {
			return errors.New("database error")
		},
	}

	// Create scanner
	cfg := &config.ScannerConfig{}
	scanner := New(cfg, db)

	// Create temp directory
	tempDir := testutil.CreateTempDir(t)

	// Create test file
	filePath := testutil.CreateTestFile(t, tempDir, "test.mp4", "test content")

	// Create media file
	_, err := scanner.CreateMediaFile(filePath)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	expectedError := "error creating media file record: database error"
	if err.Error() != expectedError {
		t.Errorf("Expected error message %q, got %q", expectedError, err.Error())
	}
}

func TestCreateBatchProcess(t *testing.T) {
	// Create temp directory
	tempDir := testutil.CreateTempDir(t)

	// Create test files
	file1 := testutil.CreateTestFile(t, tempDir, "file1.mp4", "test content")
	file2 := testutil.CreateTestFile(t, tempDir, "file2.mp4", "test content")

	// Create mock database
	var createdBatch *models.BatchProcess
	var createdFiles []*models.MediaFile
	var createdBatchFiles []*models.BatchProcessFile

	db := &MockDB{
		CreateBatchProcessFunc: func(batch *models.BatchProcess) error {
			createdBatch = batch
			batch.ID = 1 // Simulate database assigning an ID
			return nil
		},
		CreateMediaFileFunc: func(file *models.MediaFile) error {
			createdFiles = append(createdFiles, file)
			file.ID = int64(len(createdFiles)) // Simulate database assigning an ID
			return nil
		},
		CreateBatchProcessFileFunc: func(file *models.BatchProcessFile) error {
			createdBatchFiles = append(createdBatchFiles, file)
			file.ID = int64(len(createdBatchFiles)) // Simulate database assigning an ID
			return nil
		},
	}

	// Create scanner
	cfg := &config.ScannerConfig{}
	scanner := New(cfg, db)

	// Create batch process
	files := []string{file1, file2}
	batchProcess, err := scanner.CreateBatchProcess(tempDir, files)
	if err != nil {
		t.Fatalf("CreateBatchProcess failed: %v", err)
	}

	// Verify the result
	if batchProcess.ID != 1 {
		t.Errorf("Expected ID to be 1, got %d", batchProcess.ID)
	}

	if batchProcess.Directory != tempDir {
		t.Errorf("Expected directory to be %q, got %q", tempDir, batchProcess.Directory)
	}

	if batchProcess.FileCount != 2 {
		t.Errorf("Expected file count to be 2, got %d", batchProcess.FileCount)
	}

	if batchProcess.Status != "pending" {
		t.Errorf("Expected status to be 'pending', got %q", batchProcess.Status)
	}

	// Verify the batch was created in the database
	if createdBatch == nil {
		t.Fatal("Expected batch to be created in database")
	}

	// Verify the media files were created
	if len(createdFiles) != 2 {
		t.Fatalf("Expected 2 media files to be created, got %d", len(createdFiles))
	}

	// Verify the batch process files were created
	if len(createdBatchFiles) != 2 {
		t.Fatalf("Expected 2 batch process files to be created, got %d", len(createdBatchFiles))
	}

	// Verify the batch process files reference the correct batch and media files
	for i, bpf := range createdBatchFiles {
		if bpf.BatchProcessID != 1 {
			t.Errorf("Expected batch process ID to be 1, got %d", bpf.BatchProcessID)
		}

		if bpf.MediaFileID != int64(i+1) {
			t.Errorf("Expected media file ID to be %d, got %d", i+1, bpf.MediaFileID)
		}

		if bpf.Status != "pending" {
			t.Errorf("Expected status to be 'pending', got %q", bpf.Status)
		}
	}
}

func TestCreateBatchProcess_Error(t *testing.T) {
	// Create mock database that returns an error
	db := &MockDB{
		CreateBatchProcessFunc: func(batch *models.BatchProcess) error {
			return errors.New("database error")
		},
	}

	// Create scanner
	cfg := &config.ScannerConfig{}
	scanner := New(cfg, db)

	// Create batch process
	_, err := scanner.CreateBatchProcess("/test/dir", []string{"/test/dir/file1.mp4"})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	expectedError := "error creating batch process record: database error"
	if err.Error() != expectedError {
		t.Errorf("Expected error message %q, got %q", expectedError, err.Error())
	}
}
