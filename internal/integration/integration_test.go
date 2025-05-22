package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/sleepstars/mediascanner/internal/api"
	"github.com/sleepstars/mediascanner/internal/config"
	"github.com/sleepstars/mediascanner/internal/fileops"
	"github.com/sleepstars/mediascanner/internal/llm"
	"github.com/sleepstars/mediascanner/internal/models"
	"github.com/sleepstars/mediascanner/internal/notification"
	"github.com/sleepstars/mediascanner/internal/processor"
	"github.com/sleepstars/mediascanner/internal/scanner"
	"github.com/sleepstars/mediascanner/internal/testutil"
)

// MockLLM is a mock implementation of the LLM client
type MockLLM struct {
	ProcessMediaFileFunc  func(ctx context.Context, filename string, directoryStructure map[string][]string) (*llm.MediaFileResult, error)
	ProcessBatchFilesFunc func(ctx context.Context, filenames []string, directoryStructure map[string][]string) ([]*llm.MediaFileResult, error)
}

func (m *MockLLM) ProcessMediaFile(ctx context.Context, filename string, directoryStructure map[string][]string) (*llm.MediaFileResult, error) {
	return m.ProcessMediaFileFunc(ctx, filename, directoryStructure)
}

func (m *MockLLM) ProcessBatchFiles(ctx context.Context, filenames []string, directoryStructure map[string][]string) ([]*llm.MediaFileResult, error) {
	return m.ProcessBatchFilesFunc(ctx, filenames, directoryStructure)
}

// MockAPIClient is a mock implementation of the API client
type MockAPIClient struct {
	SearchMovieFunc       func(ctx context.Context, query string, year int) (*api.MovieSearchResult, error)
	SearchTVFunc          func(ctx context.Context, query string, year int) (*api.TVSearchResult, error)
	GetMovieDetailsFunc   func(ctx context.Context, id int) (*api.MovieDetails, error)
	GetTVDetailsFunc      func(ctx context.Context, id int) (*api.TVDetails, error)
	GetSeasonDetailsFunc  func(ctx context.Context, tvID, seasonNumber int) (*api.SeasonDetails, error)
	GetEpisodeDetailsFunc func(ctx context.Context, tvID, seasonNumber, episodeNumber int) (*api.EpisodeDetails, error)
}

func (m *MockAPIClient) SearchMovie(ctx context.Context, query string, year int) (*api.MovieSearchResult, error) {
	return m.SearchMovieFunc(ctx, query, year)
}

func (m *MockAPIClient) SearchTV(ctx context.Context, query string, year int) (*api.TVSearchResult, error) {
	return m.SearchTVFunc(ctx, query, year)
}

func (m *MockAPIClient) GetMovieDetails(ctx context.Context, id int) (*api.MovieDetails, error) {
	return m.GetMovieDetailsFunc(ctx, id)
}

func (m *MockAPIClient) GetTVDetails(ctx context.Context, id int) (*api.TVDetails, error) {
	return m.GetTVDetailsFunc(ctx, id)
}

func (m *MockAPIClient) GetSeasonDetails(ctx context.Context, tvID, seasonNumber int) (*api.SeasonDetails, error) {
	return m.GetSeasonDetailsFunc(ctx, tvID, seasonNumber)
}

func (m *MockAPIClient) GetEpisodeDetails(ctx context.Context, tvID, seasonNumber, episodeNumber int) (*api.EpisodeDetails, error) {
	return m.GetEpisodeDetailsFunc(ctx, tvID, seasonNumber, episodeNumber)
}

// MockDatabase is a mock implementation of the database
type MockDatabase struct {
	GetMediaFileByPathFunc      func(path string) (*models.MediaFile, error)
	CreateMediaFileFunc         func(file *models.MediaFile) error
	UpdateMediaFileFunc         func(file *models.MediaFile) error
	GetMediaFileByIDFunc        func(id int64) (*models.MediaFile, error)
	CreateMediaInfoFunc         func(info *models.MediaInfo) error
	GetMediaInfoByMediaFileIDFunc func(mediaFileID int64) (*models.MediaInfo, error)
	CreateBatchProcessFunc      func(batch *models.BatchProcess) error
	UpdateBatchProcessFunc      func(batch *models.BatchProcess) error
	CreateBatchProcessFileFunc  func(file *models.BatchProcessFile) error
	UpdateBatchProcessFileFunc  func(file *models.BatchProcessFile) error
	GetBatchProcessFilesByBatchIDFunc func(batchID int64) ([]models.BatchProcessFile, error)
	CreateNotificationFunc      func(notification *models.Notification) error
}

func (m *MockDatabase) GetMediaFileByPath(path string) (*models.MediaFile, error) {
	return m.GetMediaFileByPathFunc(path)
}

func (m *MockDatabase) CreateMediaFile(file *models.MediaFile) error {
	return m.CreateMediaFileFunc(file)
}

func (m *MockDatabase) UpdateMediaFile(file *models.MediaFile) error {
	return m.UpdateMediaFileFunc(file)
}

func (m *MockDatabase) GetMediaFileByID(id int64) (*models.MediaFile, error) {
	return m.GetMediaFileByIDFunc(id)
}

func (m *MockDatabase) CreateMediaInfo(info *models.MediaInfo) error {
	return m.CreateMediaInfoFunc(info)
}

func (m *MockDatabase) GetMediaInfoByMediaFileID(mediaFileID int64) (*models.MediaInfo, error) {
	return m.GetMediaInfoByMediaFileIDFunc(mediaFileID)
}

func (m *MockDatabase) CreateBatchProcess(batch *models.BatchProcess) error {
	return m.CreateBatchProcessFunc(batch)
}

func (m *MockDatabase) UpdateBatchProcess(batch *models.BatchProcess) error {
	return m.UpdateBatchProcessFunc(batch)
}

func (m *MockDatabase) CreateBatchProcessFile(file *models.BatchProcessFile) error {
	return m.CreateBatchProcessFileFunc(file)
}

func (m *MockDatabase) UpdateBatchProcessFile(file *models.BatchProcessFile) error {
	return m.UpdateBatchProcessFileFunc(file)
}

func (m *MockDatabase) GetBatchProcessFilesByBatchID(batchID int64) ([]models.BatchProcessFile, error) {
	return m.GetBatchProcessFilesByBatchIDFunc(batchID)
}

func (m *MockDatabase) CreateNotification(notification *models.Notification) error {
	return m.CreateNotificationFunc(notification)
}

// TestEndToEndWorkflow tests the end-to-end workflow from scanning to processing
func TestEndToEndWorkflow(t *testing.T) {
	// Create temp directories
	sourceDir := testutil.CreateTempDir(t)
	destDir := testutil.CreateTempDir(t)
	
	// Create test files
	movieFile := testutil.CreateTestFile(t, sourceDir, "The.Matrix.1999.1080p.BluRay.x264.mp4", "test content")
	tvFile := testutil.CreateTestFile(t, sourceDir, "Breaking.Bad.S01E01.1080p.BluRay.x264.mp4", "test content")
	
	// Create batch directory
	batchDir := filepath.Join(sourceDir, "batch")
	if err := os.MkdirAll(batchDir, 0755); err != nil {
		t.Fatalf("Failed to create batch directory: %v", err)
	}
	
	// Create batch files
	batchFile1 := testutil.CreateTestFile(t, batchDir, "The.Matrix.1999.1080p.BluRay.x264.mp4", "test content")
	batchFile2 := testutil.CreateTestFile(t, batchDir, "The.Matrix.Reloaded.2003.1080p.BluRay.x264.mp4", "test content")
	batchFile3 := testutil.CreateTestFile(t, batchDir, "The.Matrix.Revolutions.2003.1080p.BluRay.x264.mp4", "test content")
	
	// Create mock database
	var createdMediaFiles []*models.MediaFile
	var createdMediaInfos []*models.MediaInfo
	var createdBatchProcesses []*models.BatchProcess
	var createdBatchProcessFiles []*models.BatchProcessFile
	var createdNotifications []*models.Notification
	
	mockDB := &MockDatabase{
		GetMediaFileByPathFunc: func(path string) (*models.MediaFile, error) {
			// Simulate that no files are in the database yet
			return nil, os.ErrNotExist
		},
		CreateMediaFileFunc: func(file *models.MediaFile) error {
			file.ID = int64(len(createdMediaFiles) + 1)
			createdMediaFiles = append(createdMediaFiles, file)
			return nil
		},
		UpdateMediaFileFunc: func(file *models.MediaFile) error {
			// Find and update the file
			for i, f := range createdMediaFiles {
				if f.ID == file.ID {
					createdMediaFiles[i] = file
					return nil
				}
			}
			return os.ErrNotExist
		},
		GetMediaFileByIDFunc: func(id int64) (*models.MediaFile, error) {
			// Find the file by ID
			for _, f := range createdMediaFiles {
				if f.ID == id {
					return f, nil
				}
			}
			return nil, os.ErrNotExist
		},
		CreateMediaInfoFunc: func(info *models.MediaInfo) error {
			info.ID = int64(len(createdMediaInfos) + 1)
			createdMediaInfos = append(createdMediaInfos, info)
			return nil
		},
		GetMediaInfoByMediaFileIDFunc: func(mediaFileID int64) (*models.MediaInfo, error) {
			// Find the media info by media file ID
			for _, info := range createdMediaInfos {
				if info.MediaFileID == mediaFileID {
					return info, nil
				}
			}
			return nil, os.ErrNotExist
		},
		CreateBatchProcessFunc: func(batch *models.BatchProcess) error {
			batch.ID = int64(len(createdBatchProcesses) + 1)
			createdBatchProcesses = append(createdBatchProcesses, batch)
			return nil
		},
		UpdateBatchProcessFunc: func(batch *models.BatchProcess) error {
			// Find and update the batch
			for i, b := range createdBatchProcesses {
				if b.ID == batch.ID {
					createdBatchProcesses[i] = batch
					return nil
				}
			}
			return os.ErrNotExist
		},
		CreateBatchProcessFileFunc: func(file *models.BatchProcessFile) error {
			file.ID = int64(len(createdBatchProcessFiles) + 1)
			createdBatchProcessFiles = append(createdBatchProcessFiles, file)
			return nil
		},
		UpdateBatchProcessFileFunc: func(file *models.BatchProcessFile) error {
			// Find and update the batch process file
			for i, f := range createdBatchProcessFiles {
				if f.ID == file.ID {
					createdBatchProcessFiles[i] = file
					return nil
				}
			}
			return os.ErrNotExist
		},
		GetBatchProcessFilesByBatchIDFunc: func(batchID int64) ([]models.BatchProcessFile, error) {
			// Find batch process files by batch ID
			var files []models.BatchProcessFile
			for _, f := range createdBatchProcessFiles {
				if f.BatchProcessID == batchID {
					files = append(files, *f)
				}
			}
			return files, nil
		},
		CreateNotificationFunc: func(notification *models.Notification) error {
			notification.ID = int64(len(createdNotifications) + 1)
			createdNotifications = append(createdNotifications, notification)
			return nil
		},
	}
	
	// Create mock LLM
	mockLLM := &MockLLM{
		ProcessMediaFileFunc: func(ctx context.Context, filename string, directoryStructure map[string][]string) (*llm.MediaFileResult, error) {
			// Return different results based on the filename
			if filepath.Base(filename) == "The.Matrix.1999.1080p.BluRay.x264.mp4" {
				return &llm.MediaFileResult{
					OriginalFilename: filepath.Base(filename),
					Title:            "The Matrix",
					OriginalTitle:    "The Matrix",
					Year:             1999,
					MediaType:        "movie",
					TMDBID:           603,
					Category:         "movie",
					Subcategory:      "sci-fi",
					DestinationPath:  filepath.Join(destDir, "Movies/Sci-Fi/The Matrix (1999)/The.Matrix.1999.1080p.BluRay.x264.mp4"),
					Confidence:       0.95,
				}, nil
			} else if filepath.Base(filename) == "Breaking.Bad.S01E01.1080p.BluRay.x264.mp4" {
				return &llm.MediaFileResult{
					OriginalFilename: filepath.Base(filename),
					Title:            "Breaking Bad",
					OriginalTitle:    "Breaking Bad",
					Year:             2008,
					MediaType:        "tv",
					Season:           1,
					Episode:          1,
					EpisodeTitle:     "Pilot",
					TMDBID:           1396,
					Category:         "tv",
					Subcategory:      "drama",
					DestinationPath:  filepath.Join(destDir, "TV/Drama/Breaking Bad/Season 1/Breaking.Bad.S01E01.1080p.BluRay.x264.mp4"),
					Confidence:       0.95,
				}, nil
			}
			
			return nil, os.ErrNotExist
		},
		ProcessBatchFilesFunc: func(ctx context.Context, filenames []string, directoryStructure map[string][]string) ([]*llm.MediaFileResult, error) {
			// Return results for the batch files
			results := make([]*llm.MediaFileResult, 0, len(filenames))
			
			for _, filename := range filenames {
				if filepath.Base(filename) == "The.Matrix.1999.1080p.BluRay.x264.mp4" {
					results = append(results, &llm.MediaFileResult{
						OriginalFilename: filepath.Base(filename),
						Title:            "The Matrix",
						OriginalTitle:    "The Matrix",
						Year:             1999,
						MediaType:        "movie",
						TMDBID:           603,
						Category:         "movie",
						Subcategory:      "sci-fi",
						DestinationPath:  filepath.Join(destDir, "Movies/Sci-Fi/The Matrix (1999)/The.Matrix.1999.1080p.BluRay.x264.mp4"),
						Confidence:       0.95,
					})
				} else if filepath.Base(filename) == "The.Matrix.Reloaded.2003.1080p.BluRay.x264.mp4" {
					results = append(results, &llm.MediaFileResult{
						OriginalFilename: filepath.Base(filename),
						Title:            "The Matrix Reloaded",
						OriginalTitle:    "The Matrix Reloaded",
						Year:             2003,
						MediaType:        "movie",
						TMDBID:           604,
						Category:         "movie",
						Subcategory:      "sci-fi",
						DestinationPath:  filepath.Join(destDir, "Movies/Sci-Fi/The Matrix Reloaded (2003)/The.Matrix.Reloaded.2003.1080p.BluRay.x264.mp4"),
						Confidence:       0.95,
					})
				} else if filepath.Base(filename) == "The.Matrix.Revolutions.2003.1080p.BluRay.x264.mp4" {
					results = append(results, &llm.MediaFileResult{
						OriginalFilename: filepath.Base(filename),
						Title:            "The Matrix Revolutions",
						OriginalTitle:    "The Matrix Revolutions",
						Year:             2003,
						MediaType:        "movie",
						TMDBID:           605,
						Category:         "movie",
						Subcategory:      "sci-fi",
						DestinationPath:  filepath.Join(destDir, "Movies/Sci-Fi/The Matrix Revolutions (2003)/The.Matrix.Revolutions.2003.1080p.BluRay.x264.mp4"),
						Confidence:       0.95,
					})
				}
			}
			
			return results, nil
		},
	}
	
	// Create mock API client
	mockAPIClient := &MockAPIClient{
		SearchMovieFunc: func(ctx context.Context, query string, year int) (*api.MovieSearchResult, error) {
			// Return different results based on the query
			if query == "The Matrix" && year == 1999 {
				return &api.MovieSearchResult{
					Query: query,
					Year:  year,
					Movies: []api.Movie{
						{
							ID:            603,
							Title:         "The Matrix",
							OriginalTitle: "The Matrix",
							ReleaseYear:   1999,
							Overview:      "Set in the 22nd century, The Matrix tells the story of a computer hacker who joins a group of underground insurgents fighting the vast and powerful computers who now rule the earth.",
							PosterPath:    "/path/to/poster.jpg",
							BackdropPath:  "/path/to/backdrop.jpg",
						},
					},
				}, nil
			}
			
			return &api.MovieSearchResult{
				Query:  query,
				Year:   year,
				Movies: []api.Movie{},
			}, nil
		},
		SearchTVFunc: func(ctx context.Context, query string, year int) (*api.TVSearchResult, error) {
			// Return different results based on the query
			if query == "Breaking Bad" && year == 2008 {
				return &api.TVSearchResult{
					Query: query,
					Year:  year,
					Shows: []api.TVShow{
						{
							ID:           1396,
							Name:         "Breaking Bad",
							OriginalName: "Breaking Bad",
							FirstAirYear: 2008,
							Overview:     "When Walter White, a New Mexico chemistry teacher, is diagnosed with Stage III cancer and given a prognosis of only two years left to live.",
							PosterPath:   "/path/to/poster.jpg",
							BackdropPath: "/path/to/backdrop.jpg",
						},
					},
				}, nil
			}
			
			return &api.TVSearchResult{
				Query: query,
				Year:  year,
				Shows: []api.TVShow{},
			}, nil
		},
		GetMovieDetailsFunc: func(ctx context.Context, id int) (*api.MovieDetails, error) {
			// Return different results based on the ID
			if id == 603 {
				return &api.MovieDetails{
					ID:            603,
					Title:         "The Matrix",
					OriginalTitle: "The Matrix",
					ReleaseYear:   1999,
					Overview:      "Set in the 22nd century, The Matrix tells the story of a computer hacker who joins a group of underground insurgents fighting the vast and powerful computers who now rule the earth.",
					PosterPath:    "/path/to/poster.jpg",
					BackdropPath:  "/path/to/backdrop.jpg",
					ImdbID:        "tt0133093",
					Genres:        []string{"Action", "Science Fiction"},
					Countries:     []string{"United States of America"},
					Languages:     []string{"English"},
					Runtime:       136,
				}, nil
			}
			
			return nil, os.ErrNotExist
		},
		GetTVDetailsFunc: func(ctx context.Context, id int) (*api.TVDetails, error) {
			// Return different results based on the ID
			if id == 1396 {
				return &api.TVDetails{
					ID:              1396,
					Name:            "Breaking Bad",
					OriginalName:    "Breaking Bad",
					FirstAirYear:    2008,
					Overview:        "When Walter White, a New Mexico chemistry teacher, is diagnosed with Stage III cancer and given a prognosis of only two years left to live.",
					PosterPath:      "/path/to/poster.jpg",
					BackdropPath:    "/path/to/backdrop.jpg",
					ImdbID:          "tt0903747",
					TVDBID:          81189,
					Genres:          []string{"Drama", "Crime"},
					Countries:       []string{"United States of America"},
					Languages:       []string{"English"},
					NumberOfSeasons: 5,
				}, nil
			}
			
			return nil, os.ErrNotExist
		},
		GetSeasonDetailsFunc: func(ctx context.Context, tvID, seasonNumber int) (*api.SeasonDetails, error) {
			// Return different results based on the TV ID and season number
			if tvID == 1396 && seasonNumber == 1 {
				return &api.SeasonDetails{
					ID:           3572,
					Name:         "Season 1",
					Overview:     "The first season of Breaking Bad.",
					SeasonNumber: 1,
					PosterPath:   "/path/to/season_poster.jpg",
					AirDate:      "2008-01-20",
					Episodes: []api.Episode{
						{
							ID:            62085,
							Name:          "Pilot",
							Overview:      "Chemistry teacher Walter White's life changes when he is diagnosed with terminal cancer.",
							EpisodeNumber: 1,
							SeasonNumber:  1,
							StillPath:     "/path/to/still.jpg",
							AirDate:       "2008-01-20",
						},
					},
				}, nil
			}
			
			return nil, os.ErrNotExist
		},
		GetEpisodeDetailsFunc: func(ctx context.Context, tvID, seasonNumber, episodeNumber int) (*api.EpisodeDetails, error) {
			// Return different results based on the TV ID, season number, and episode number
			if tvID == 1396 && seasonNumber == 1 && episodeNumber == 1 {
				return &api.EpisodeDetails{
					ID:            62085,
					Name:          "Pilot",
					Overview:      "Chemistry teacher Walter White's life changes when he is diagnosed with terminal cancer.",
					EpisodeNumber: 1,
					SeasonNumber:  1,
					StillPath:     "/path/to/still.jpg",
					AirDate:       "2008-01-20",
				}, nil
			}
			
			return nil, os.ErrNotExist
		},
	}
	
	// Create configuration
	cfg := &config.Config{
		ScanInterval: 60,
		Scanner: config.ScannerConfig{
			MediaDirs:       []string{sourceDir},
			VideoExtensions: []string{".mp4", ".mkv", ".avi"},
			ExcludePatterns: []string{"sample", "trailer"},
			ExcludeDirs:     []string{"_UNPACK_", "tmp"},
			BatchThreshold:  3,
		},
		FileOps: config.FileOpsConfig{
			Mode:        "copy",
			DestDir:     destDir,
			CreateNFO:   true,
			DownloadArt: false,
		},
		WorkerPool: config.WorkerPoolConfig{
			Enabled:    false,
			NumWorkers: 2,
		},
	}
	
	// Create components
	fileOps := fileops.New(&cfg.FileOps)
	notifier := notification.New(&cfg.Notification, mockDB)
	scan := scanner.New(&cfg.Scanner, mockDB)
	proc := processor.New(cfg, mockDB, mockLLM, mockAPIClient, fileOps, notifier)
	
	// Create context
	ctx := context.Background()
	
	// Run the scan
	result, err := scan.Scan()
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}
	
	// Verify the scan result
	if len(result.NewFiles) != 5 {
		t.Errorf("Expected 5 new files, got %d", len(result.NewFiles))
	}
	
	if len(result.BatchDirs) != 1 {
		t.Errorf("Expected 1 batch directory, got %d", len(result.BatchDirs))
	}
	
	// Process batch directories
	for dir, files := range result.BatchDirs {
		t.Logf("Processing batch directory: %s (%d files)", dir, len(files))
		
		// Create batch process
		batchProcess, err := scan.CreateBatchProcess(dir, files)
		if err != nil {
			t.Fatalf("CreateBatchProcess failed: %v", err)
		}
		
		// Process batch
		err = proc.ProcessBatchFiles(ctx, batchProcess)
		if err != nil {
			t.Fatalf("ProcessBatchFiles failed: %v", err)
		}
	}
	
	// Process individual files
	for _, file := range result.NewFiles {
		// Skip files that are in batch directories
		inBatch := false
		for dir := range result.BatchDirs {
			if filepath.Dir(file) == dir {
				inBatch = true
				break
			}
		}
		
		if inBatch {
			continue
		}
		
		t.Logf("Processing individual file: %s", file)
		
		// Create media file
		mediaFile, err := scan.CreateMediaFile(file)
		if err != nil {
			t.Fatalf("CreateMediaFile failed: %v", err)
		}
		
		// Process media file
		err = proc.ProcessMediaFile(ctx, mediaFile)
		if err != nil {
			t.Fatalf("ProcessMediaFile failed: %v", err)
		}
	}
	
	// Verify the results
	
	// Check that media files were created
	if len(createdMediaFiles) != 5 {
		t.Errorf("Expected 5 media files to be created, got %d", len(createdMediaFiles))
	}
	
	// Check that media infos were created
	if len(createdMediaInfos) != 5 {
		t.Errorf("Expected 5 media infos to be created, got %d", len(createdMediaInfos))
	}
	
	// Check that batch processes were created
	if len(createdBatchProcesses) != 1 {
		t.Errorf("Expected 1 batch process to be created, got %d", len(createdBatchProcesses))
	}
	
	// Check that batch process files were created
	if len(createdBatchProcessFiles) != 3 {
		t.Errorf("Expected 3 batch process files to be created, got %d", len(createdBatchProcessFiles))
	}
	
	// Check that notifications were created
	if len(createdNotifications) != 5 {
		t.Errorf("Expected 5 notifications to be created, got %d", len(createdNotifications))
	}
	
	// Check that files were copied to the destination directory
	destinationFiles := []string{
		filepath.Join(destDir, "Movies/Sci-Fi/The Matrix (1999)/The.Matrix.1999.1080p.BluRay.x264.mp4"),
		filepath.Join(destDir, "TV/Drama/Breaking Bad/Season 1/Breaking.Bad.S01E01.1080p.BluRay.x264.mp4"),
		filepath.Join(destDir, "Movies/Sci-Fi/The Matrix Reloaded (2003)/The.Matrix.Reloaded.2003.1080p.BluRay.x264.mp4"),
		filepath.Join(destDir, "Movies/Sci-Fi/The Matrix Revolutions (2003)/The.Matrix.Revolutions.2003.1080p.BluRay.x264.mp4"),
	}
	
	for _, file := range destinationFiles {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			t.Errorf("Expected file %s to exist in destination directory", file)
		}
	}
}
