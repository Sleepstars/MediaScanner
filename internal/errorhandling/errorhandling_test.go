package errorhandling

import (
	"context"
	"errors"
	"testing"

	"github.com/sleepstars/mediascanner/internal/api"
	"github.com/sleepstars/mediascanner/internal/config"
	"github.com/sleepstars/mediascanner/internal/fileops"
	"github.com/sleepstars/mediascanner/internal/llm"
	"github.com/sleepstars/mediascanner/internal/models"
	"github.com/sleepstars/mediascanner/internal/notification"
	"github.com/sleepstars/mediascanner/internal/processor"
	"github.com/sleepstars/mediascanner/internal/testutil"
)

// MockLLM is a mock implementation of the LLM client
type MockLLM struct {
	ProcessMediaFileFunc func(ctx context.Context, filename string, directoryStructure map[string][]string) (*llm.MediaFileResult, error)
}

func (m *MockLLM) ProcessMediaFile(ctx context.Context, filename string, directoryStructure map[string][]string) (*llm.MediaFileResult, error) {
	return m.ProcessMediaFileFunc(ctx, filename, directoryStructure)
}

func (m *MockLLM) ProcessBatchFiles(ctx context.Context, filenames []string, directoryStructure map[string][]string) ([]*llm.MediaFileResult, error) {
	return nil, errors.New("not implemented")
}

// MockAPIClient is a mock implementation of the API client
type MockAPIClient struct {
	SearchMovieFunc func(ctx context.Context, query string, year int) (*api.MovieSearchResult, error)
	SearchTVFunc    func(ctx context.Context, query string, year int) (*api.TVSearchResult, error)
}

func (m *MockAPIClient) SearchMovie(ctx context.Context, query string, year int) (*api.MovieSearchResult, error) {
	return m.SearchMovieFunc(ctx, query, year)
}

func (m *MockAPIClient) SearchTV(ctx context.Context, query string, year int) (*api.TVSearchResult, error) {
	return m.SearchTVFunc(ctx, query, year)
}

func (m *MockAPIClient) GetMovieDetails(ctx context.Context, id int) (*api.MovieDetails, error) {
	return nil, errors.New("not implemented")
}

func (m *MockAPIClient) GetTVDetails(ctx context.Context, id int) (*api.TVDetails, error) {
	return nil, errors.New("not implemented")
}

func (m *MockAPIClient) GetSeasonDetails(ctx context.Context, tvID, seasonNumber int) (*api.SeasonDetails, error) {
	return nil, errors.New("not implemented")
}

func (m *MockAPIClient) GetEpisodeDetails(ctx context.Context, tvID, seasonNumber, episodeNumber int) (*api.EpisodeDetails, error) {
	return nil, errors.New("not implemented")
}

// MockDatabase is a mock implementation of the database
type MockDatabase struct {
	UpdateMediaFileFunc func(file *models.MediaFile) error
	CreateMediaInfoFunc func(info *models.MediaInfo) error
	CreateNotificationFunc func(notification *models.Notification) error
}

func (m *MockDatabase) UpdateMediaFile(file *models.MediaFile) error {
	return m.UpdateMediaFileFunc(file)
}

func (m *MockDatabase) CreateMediaInfo(info *models.MediaInfo) error {
	return m.CreateMediaInfoFunc(info)
}

func (m *MockDatabase) CreateNotification(notification *models.Notification) error {
	return m.CreateNotificationFunc(notification)
}

// TestMalformedFilename tests handling of malformed filenames
func TestMalformedFilename(t *testing.T) {
	// Create a mock LLM that returns an error for malformed filenames
	mockLLM := &MockLLM{
		ProcessMediaFileFunc: func(ctx context.Context, filename string, directoryStructure map[string][]string) (*llm.MediaFileResult, error) {
			return nil, errors.New("failed to parse filename")
		},
	}
	
	// Create a mock database
	var updatedFile *models.MediaFile
	var createdNotification *models.Notification
	
	mockDB := &MockDatabase{
		UpdateMediaFileFunc: func(file *models.MediaFile) error {
			updatedFile = file
			return nil
		},
		CreateNotificationFunc: func(notification *models.Notification) error {
			createdNotification = notification
			return nil
		},
	}
	
	// Create configuration
	cfg := &config.Config{
		FileOps: config.FileOpsConfig{
			Mode:    "copy",
			DestDir: "/test/dest",
		},
	}
	
	// Create components
	fileOps := fileops.New(&cfg.FileOps)
	notifier := notification.New(&cfg.Notification, mockDB)
	
	// Create processor
	proc := processor.New(cfg, mockDB, mockLLM, nil, fileOps, notifier)
	
	// Create a media file with a malformed filename
	mediaFile := &models.MediaFile{
		ID:           1,
		OriginalPath: "/test/path/malformed_filename.mp4",
		OriginalName: "malformed_filename.mp4",
		Status:       "pending",
	}
	
	// Process the media file
	err := proc.ProcessMediaFile(context.Background(), mediaFile)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	
	// Verify the media file was updated with error status
	if updatedFile == nil {
		t.Fatal("Expected media file to be updated")
	}
	
	if updatedFile.Status != "failed" {
		t.Errorf("Expected status to be 'failed', got %q", updatedFile.Status)
	}
	
	if updatedFile.ErrorMessage == "" {
		t.Error("Expected error message to be set")
	}
	
	// Verify a notification was created
	if createdNotification == nil {
		t.Fatal("Expected notification to be created")
	}
	
	if createdNotification.Type != "error" {
		t.Errorf("Expected notification type to be 'error', got %q", createdNotification.Type)
	}
	
	if createdNotification.MediaFileID != 1 {
		t.Errorf("Expected notification media file ID to be 1, got %d", createdNotification.MediaFileID)
	}
}

// TestAPIFailure tests handling of API failures
func TestAPIFailure(t *testing.T) {
	// Create a mock LLM that returns a valid result
	mockLLM := &MockLLM{
		ProcessMediaFileFunc: func(ctx context.Context, filename string, directoryStructure map[string][]string) (*llm.MediaFileResult, error) {
			return &llm.MediaFileResult{
				OriginalFilename: "The.Matrix.1999.1080p.BluRay.x264.mp4",
				Title:            "The Matrix",
				OriginalTitle:    "The Matrix",
				Year:             1999,
				MediaType:        "movie",
				TMDBID:           603,
				Category:         "movie",
				Subcategory:      "sci-fi",
				DestinationPath:  "/test/dest/Movies/Sci-Fi/The Matrix (1999)/The.Matrix.1999.1080p.BluRay.x264.mp4",
				Confidence:       0.95,
			}, nil
		},
	}
	
	// Create a mock API client that returns an error
	mockAPIClient := &MockAPIClient{
		SearchMovieFunc: func(ctx context.Context, query string, year int) (*api.MovieSearchResult, error) {
			return nil, errors.New("API error")
		},
		SearchTVFunc: func(ctx context.Context, query string, year int) (*api.TVSearchResult, error) {
			return nil, errors.New("API error")
		},
	}
	
	// Create a mock database
	var updatedFile *models.MediaFile
	var createdMediaInfo *models.MediaInfo
	var createdNotification *models.Notification
	
	mockDB := &MockDatabase{
		UpdateMediaFileFunc: func(file *models.MediaFile) error {
			updatedFile = file
			return nil
		},
		CreateMediaInfoFunc: func(info *models.MediaInfo) error {
			createdMediaInfo = info
			return nil
		},
		CreateNotificationFunc: func(notification *models.Notification) error {
			createdNotification = notification
			return nil
		},
	}
	
	// Create configuration
	cfg := &config.Config{
		FileOps: config.FileOpsConfig{
			Mode:    "copy",
			DestDir: "/test/dest",
		},
	}
	
	// Create components
	fileOps := fileops.New(&cfg.FileOps)
	notifier := notification.New(&cfg.Notification, mockDB)
	
	// Create processor
	proc := processor.New(cfg, mockDB, mockLLM, mockAPIClient, fileOps, notifier)
	
	// Create a media file
	mediaFile := &models.MediaFile{
		ID:           1,
		OriginalPath: "/test/path/The.Matrix.1999.1080p.BluRay.x264.mp4",
		OriginalName: "The.Matrix.1999.1080p.BluRay.x264.mp4",
		Status:       "pending",
	}
	
	// Process the media file
	err := proc.ProcessMediaFile(context.Background(), mediaFile)
	
	// We expect the processor to continue despite API errors
	if err != nil {
		t.Fatalf("ProcessMediaFile failed: %v", err)
	}
	
	// Verify the media file was updated with success status
	if updatedFile == nil {
		t.Fatal("Expected media file to be updated")
	}
	
	if updatedFile.Status != "success" {
		t.Errorf("Expected status to be 'success', got %q", updatedFile.Status)
	}
	
	// Verify media info was created
	if createdMediaInfo == nil {
		t.Fatal("Expected media info to be created")
	}
	
	if createdMediaInfo.MediaFileID != 1 {
		t.Errorf("Expected media file ID to be 1, got %d", createdMediaInfo.MediaFileID)
	}
	
	if createdMediaInfo.Title != "The Matrix" {
		t.Errorf("Expected title to be 'The Matrix', got %q", createdMediaInfo.Title)
	}
	
	// Verify a notification was created
	if createdNotification == nil {
		t.Fatal("Expected notification to be created")
	}
	
	if createdNotification.Type != "success" {
		t.Errorf("Expected notification type to be 'success', got %q", createdNotification.Type)
	}
}

// TestDatabaseFailure tests handling of database failures
func TestDatabaseFailure(t *testing.T) {
	// Create a mock LLM that returns a valid result
	mockLLM := &MockLLM{
		ProcessMediaFileFunc: func(ctx context.Context, filename string, directoryStructure map[string][]string) (*llm.MediaFileResult, error) {
			return &llm.MediaFileResult{
				OriginalFilename: "The.Matrix.1999.1080p.BluRay.x264.mp4",
				Title:            "The Matrix",
				OriginalTitle:    "The Matrix",
				Year:             1999,
				MediaType:        "movie",
				TMDBID:           603,
				Category:         "movie",
				Subcategory:      "sci-fi",
				DestinationPath:  "/test/dest/Movies/Sci-Fi/The Matrix (1999)/The.Matrix.1999.1080p.BluRay.x264.mp4",
				Confidence:       0.95,
			}, nil
		},
	}
	
	// Create a mock database that returns an error
	mockDB := &MockDatabase{
		UpdateMediaFileFunc: func(file *models.MediaFile) error {
			return errors.New("database error")
		},
		CreateMediaInfoFunc: func(info *models.MediaInfo) error {
			return errors.New("database error")
		},
		CreateNotificationFunc: func(notification *models.Notification) error {
			return errors.New("database error")
		},
	}
	
	// Create configuration
	cfg := &config.Config{
		FileOps: config.FileOpsConfig{
			Mode:    "copy",
			DestDir: "/test/dest",
		},
	}
	
	// Create components
	fileOps := fileops.New(&cfg.FileOps)
	notifier := notification.New(&cfg.Notification, mockDB)
	
	// Create processor
	proc := processor.New(cfg, mockDB, mockLLM, nil, fileOps, notifier)
	
	// Create a media file
	mediaFile := &models.MediaFile{
		ID:           1,
		OriginalPath: "/test/path/The.Matrix.1999.1080p.BluRay.x264.mp4",
		OriginalName: "The.Matrix.1999.1080p.BluRay.x264.mp4",
		Status:       "pending",
	}
	
	// Process the media file
	err := proc.ProcessMediaFile(context.Background(), mediaFile)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	
	// Verify the error message
	expectedError := "error updating media file: database error"
	if err.Error() != expectedError {
		t.Errorf("Expected error message %q, got %q", expectedError, err.Error())
	}
}

// TestFileSystemErrors tests handling of file system errors
func TestFileSystemErrors(t *testing.T) {
	// Create temp directories
	sourceDir := testutil.CreateTempDir(t)
	destDir := testutil.CreateTempDir(t)
	
	// Create a test file
	sourcePath := testutil.CreateTestFile(t, sourceDir, "test.mp4", "test content")
	
	// Create a mock LLM that returns a valid result
	mockLLM := &MockLLM{
		ProcessMediaFileFunc: func(ctx context.Context, filename string, directoryStructure map[string][]string) (*llm.MediaFileResult, error) {
			return &llm.MediaFileResult{
				OriginalFilename: "test.mp4",
				Title:            "Test",
				MediaType:        "movie",
				Category:         "movie",
				Subcategory:      "test",
				DestinationPath:  "/non/existent/path/test.mp4", // Non-existent path
				Confidence:       0.95,
			}, nil
		},
	}
	
	// Create a mock database
	var updatedFile *models.MediaFile
	
	mockDB := &MockDatabase{
		UpdateMediaFileFunc: func(file *models.MediaFile) error {
			updatedFile = file
			return nil
		},
		CreateMediaInfoFunc: func(info *models.MediaInfo) error {
			return nil
		},
		CreateNotificationFunc: func(notification *models.Notification) error {
			return nil
		},
	}
	
	// Create configuration with a non-existent destination directory
	cfg := &config.Config{
		FileOps: config.FileOpsConfig{
			Mode:    "copy",
			DestDir: "/non/existent/path",
		},
	}
	
	// Create components
	fileOps := fileops.New(&cfg.FileOps)
	notifier := notification.New(&cfg.Notification, mockDB)
	
	// Create processor
	proc := processor.New(cfg, mockDB, mockLLM, nil, fileOps, notifier)
	
	// Create a media file
	mediaFile := &models.MediaFile{
		ID:           1,
		OriginalPath: sourcePath,
		OriginalName: "test.mp4",
		Status:       "pending",
	}
	
	// Process the media file
	err := proc.ProcessMediaFile(context.Background(), mediaFile)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	
	// Verify the media file was updated with error status
	if updatedFile == nil {
		t.Fatal("Expected media file to be updated")
	}
	
	if updatedFile.Status != "failed" {
		t.Errorf("Expected status to be 'failed', got %q", updatedFile.Status)
	}
	
	if updatedFile.ErrorMessage == "" {
		t.Error("Expected error message to be set")
	}
}
