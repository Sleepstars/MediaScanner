package testutil

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/sleepstars/mediascanner/internal/api"
	"github.com/sleepstars/mediascanner/internal/llm"
	"github.com/sleepstars/mediascanner/internal/models"
)

// MockHTTPServer creates a mock HTTP server for testing API clients
func MockHTTPServer(handler http.Handler) *httptest.Server {
	return httptest.NewServer(handler)
}

// MockTMDBServer creates a mock TMDB API server
func MockTMDBServer() *httptest.Server {
	handler := http.NewServeMux()
	
	// Mock search movies endpoint
	handler.HandleFunc("/3/search/movie", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("query")
		year := r.URL.Query().Get("year")
		
		if query == "The Matrix" && (year == "" || year == "1999") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"page": 1,
				"results": [
					{
						"id": 603,
						"title": "The Matrix",
						"original_title": "The Matrix",
						"overview": "Set in the 22nd century, The Matrix tells the story of a computer hacker who joins a group of underground insurgents fighting the vast and powerful computers who now rule the earth.",
						"release_date": "1999-03-30",
						"poster_path": "/path/to/poster.jpg",
						"backdrop_path": "/path/to/backdrop.jpg",
						"popularity": 100.0,
						"vote_average": 8.7
					}
				],
				"total_pages": 1,
				"total_results": 1
			}`))
			return
		}
		
		// Return empty results for other queries
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"page": 1,
			"results": [],
			"total_pages": 0,
			"total_results": 0
		}`))
	})
	
	// Mock search TV shows endpoint
	handler.HandleFunc("/3/search/tv", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("query")
		year := r.URL.Query().Get("first_air_date_year")
		
		if query == "Breaking Bad" && (year == "" || year == "2008") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"page": 1,
				"results": [
					{
						"id": 1396,
						"name": "Breaking Bad",
						"original_name": "Breaking Bad",
						"overview": "When Walter White, a New Mexico chemistry teacher, is diagnosed with Stage III cancer and given a prognosis of only two years left to live.",
						"first_air_date": "2008-01-20",
						"poster_path": "/path/to/poster.jpg",
						"backdrop_path": "/path/to/backdrop.jpg",
						"popularity": 100.0,
						"vote_average": 8.7
					}
				],
				"total_pages": 1,
				"total_results": 1
			}`))
			return
		}
		
		// Return empty results for other queries
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"page": 1,
			"results": [],
			"total_pages": 0,
			"total_results": 0
		}`))
	})
	
	// Mock movie details endpoint
	handler.HandleFunc("/3/movie/603", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"id": 603,
			"title": "The Matrix",
			"original_title": "The Matrix",
			"overview": "Set in the 22nd century, The Matrix tells the story of a computer hacker who joins a group of underground insurgents fighting the vast and powerful computers who now rule the earth.",
			"release_date": "1999-03-30",
			"poster_path": "/path/to/poster.jpg",
			"backdrop_path": "/path/to/backdrop.jpg",
			"imdb_id": "tt0133093",
			"genres": [
				{"id": 28, "name": "Action"},
				{"id": 878, "name": "Science Fiction"}
			],
			"production_countries": [
				{"iso_3166_1": "US", "name": "United States of America"}
			],
			"spoken_languages": [
				{"iso_639_1": "en", "name": "English"}
			],
			"runtime": 136,
			"vote_average": 8.7,
			"vote_count": 10000
		}`))
	})
	
	// Mock TV show details endpoint
	handler.HandleFunc("/3/tv/1396", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"id": 1396,
			"name": "Breaking Bad",
			"original_name": "Breaking Bad",
			"overview": "When Walter White, a New Mexico chemistry teacher, is diagnosed with Stage III cancer and given a prognosis of only two years left to live.",
			"first_air_date": "2008-01-20",
			"poster_path": "/path/to/poster.jpg",
			"backdrop_path": "/path/to/backdrop.jpg",
			"genres": [
				{"id": 18, "name": "Drama"},
				{"id": 80, "name": "Crime"}
			],
			"production_countries": [
				{"iso_3166_1": "US", "name": "United States of America"}
			],
			"number_of_seasons": 5,
			"vote_average": 8.7,
			"vote_count": 10000
		}`))
	})
	
	// Mock error endpoint
	handler.HandleFunc("/3/error", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"status_code": 500, "status_message": "Internal Server Error"}`))
	})
	
	return httptest.NewServer(handler)
}

// MockTVDBServer creates a mock TVDB API server
func MockTVDBServer() *httptest.Server {
	handler := http.NewServeMux()
	
	// Mock login endpoint
	handler.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"token": "mock-token"}`))
	})
	
	// Mock search endpoint
	handler.HandleFunc("/search/series", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("name")
		
		if query == "Breaking Bad" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"data": [
					{
						"id": 81189,
						"seriesName": "Breaking Bad",
						"firstAired": "2008-01-20",
						"overview": "When Walter White, a New Mexico chemistry teacher, is diagnosed with Stage III cancer and given a prognosis of only two years left to live."
					}
				]
			}`))
			return
		}
		
		// Return empty results for other queries
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data": []}`))
	})
	
	return httptest.NewServer(handler)
}

// MockBangumiServer creates a mock Bangumi API server
func MockBangumiServer() *httptest.Server {
	handler := http.NewServeMux()
	
	// Mock search endpoint
	handler.HandleFunc("/search/subject", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("keyword")
		
		if query == "Attack on Titan" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"results": 1,
				"list": [
					{
						"id": 55770,
						"name": "Attack on Titan",
						"name_cn": "进击的巨人",
						"air_date": "2013-04-06",
						"type": 2
					}
				]
			}`))
			return
		}
		
		// Return empty results for other queries
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"results": 0, "list": []}`))
	})
	
	return httptest.NewServer(handler)
}

// MockOpenAIServer creates a mock OpenAI API server
func MockOpenAIServer() *httptest.Server {
	handler := http.NewServeMux()
	
	// Mock chat completions endpoint
	handler.HandleFunc("/v1/chat/completions", func(w http.ResponseWriter, r *http.Request) {
		// Parse the request body
		var request struct {
			Messages []struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"messages"`
		}
		
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&request); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		
		// Check if this is a request for a movie
		isMovie := false
		isTV := false
		filename := ""
		
		for _, message := range request.Messages {
			if message.Role == "user" && message.Content != "" {
				content := message.Content
				if content[0:29] == "Please analyze this filename: " {
					filename = content[29:]
					
					if filename == "The.Matrix.1999.1080p.BluRay.x264.mp4" {
						isMovie = true
					} else if filename == "Breaking.Bad.S01E01.1080p.BluRay.x264.mp4" {
						isTV = true
					}
				}
			}
		}
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		
		if isMovie {
			// Return a movie result
			w.Write([]byte(`{
				"id": "mock-id",
				"object": "chat.completion",
				"created": 1677858242,
				"model": "gpt-3.5-turbo",
				"choices": [
					{
						"message": {
							"role": "assistant",
							"content": "{\"original_filename\":\"The.Matrix.1999.1080p.BluRay.x264.mp4\",\"title\":\"The Matrix\",\"original_title\":\"The Matrix\",\"year\":1999,\"media_type\":\"movie\",\"tmdb_id\":603,\"category\":\"movie\",\"subcategory\":\"sci-fi\",\"destination_path\":\"/movies/Sci-Fi/The Matrix (1999)/The.Matrix.1999.1080p.BluRay.x264.mp4\",\"confidence\":0.95}"
						},
						"finish_reason": "stop",
						"index": 0
					}
				],
				"usage": {
					"prompt_tokens": 10,
					"completion_tokens": 20,
					"total_tokens": 30
				}
			}`))
		} else if isTV {
			// Return a TV show result
			w.Write([]byte(`{
				"id": "mock-id",
				"object": "chat.completion",
				"created": 1677858242,
				"model": "gpt-3.5-turbo",
				"choices": [
					{
						"message": {
							"role": "assistant",
							"content": "{\"original_filename\":\"Breaking.Bad.S01E01.1080p.BluRay.x264.mp4\",\"title\":\"Breaking Bad\",\"original_title\":\"Breaking Bad\",\"year\":2008,\"media_type\":\"tv\",\"season\":1,\"episode\":1,\"episode_title\":\"Pilot\",\"tmdb_id\":1396,\"category\":\"tv\",\"subcategory\":\"drama\",\"destination_path\":\"/tv/Drama/Breaking Bad/Season 1/Breaking.Bad.S01E01.1080p.BluRay.x264.mp4\",\"confidence\":0.95}"
						},
						"finish_reason": "stop",
						"index": 0
					}
				],
				"usage": {
					"prompt_tokens": 10,
					"completion_tokens": 20,
					"total_tokens": 30
				}
			}`))
		} else {
			// Return an error result for unknown filenames
			w.Write([]byte(`{
				"id": "mock-id",
				"object": "chat.completion",
				"created": 1677858242,
				"model": "gpt-3.5-turbo",
				"choices": [
					{
						"message": {
							"role": "assistant",
							"content": "I'm sorry, but I couldn't analyze this filename. It doesn't appear to follow a standard naming convention for movies or TV shows."
						},
						"finish_reason": "stop",
						"index": 0
					}
				],
				"usage": {
					"prompt_tokens": 10,
					"completion_tokens": 20,
					"total_tokens": 30
				}
			}`))
		}
	})
	
	return httptest.NewServer(handler)
}

// MockLLMClient is a mock implementation of the LLM client
type MockLLMClient struct {
	ProcessMediaFileFunc  func(ctx context.Context, filename string, directoryStructure map[string][]string) (*llm.MediaFileResult, error)
	ProcessBatchFilesFunc func(ctx context.Context, filenames []string, directoryStructure map[string][]string) ([]*llm.MediaFileResult, error)
}

func (m *MockLLMClient) ProcessMediaFile(ctx context.Context, filename string, directoryStructure map[string][]string) (*llm.MediaFileResult, error) {
	return m.ProcessMediaFileFunc(ctx, filename, directoryStructure)
}

func (m *MockLLMClient) ProcessBatchFiles(ctx context.Context, filenames []string, directoryStructure map[string][]string) ([]*llm.MediaFileResult, error) {
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

func (m *MockDatabase) UpdateMediaInfo(info *models.MediaInfo) error {
	return m.UpdateMediaInfoFunc(info)
}

func (m *MockDatabase) GetAPICache(provider, query string) (*models.APICache, error) {
	return m.GetAPICacheFunc(provider, query)
}

func (m *MockDatabase) CreateAPICache(cache *models.APICache) error {
	return m.CreateAPICacheFunc(cache)
}

func (m *MockDatabase) CreateLLMRequest(request *models.LLMRequest) error {
	return m.CreateLLMRequestFunc(request)
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

func (m *MockDatabase) GetPendingNotifications() ([]models.Notification, error) {
	return m.GetPendingNotificationsFunc()
}

func (m *MockDatabase) UpdateNotification(notification *models.Notification) error {
	return m.UpdateNotificationFunc(notification)
}

func (m *MockDatabase) GetPendingMediaFiles() ([]models.MediaFile, error) {
	return m.GetPendingMediaFilesFunc()
}

func (m *MockDatabase) GetMediaFilesByDirectory(directory string) ([]models.MediaFile, error) {
	return m.GetMediaFilesByDirectoryFunc(directory)
}

func (m *MockDatabase) GetPendingBatchProcesses() ([]models.BatchProcess, error) {
	return m.GetPendingBatchProcessesFunc()
}

// CreateMockDatabase creates a mock database with default implementations
func CreateMockDatabase() *MockDatabase {
	return &MockDatabase{
		GetMediaFileByPathFunc: func(path string) (*models.MediaFile, error) {
			return nil, nil
		},
		CreateMediaFileFunc: func(file *models.MediaFile) error {
			file.ID = 1
			file.CreatedAt = time.Now()
			file.UpdatedAt = time.Now()
			return nil
		},
		UpdateMediaFileFunc: func(file *models.MediaFile) error {
			file.UpdatedAt = time.Now()
			return nil
		},
		GetMediaFileByIDFunc: func(id int64) (*models.MediaFile, error) {
			return &models.MediaFile{
				ID:           id,
				OriginalPath: "/test/path/file.mp4",
				OriginalName: "file.mp4",
				Status:       "pending",
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			}, nil
		},
		CreateMediaInfoFunc: func(info *models.MediaInfo) error {
			info.ID = 1
			info.CreatedAt = time.Now()
			info.UpdatedAt = time.Now()
			return nil
		},
		GetMediaInfoByMediaFileIDFunc: func(mediaFileID int64) (*models.MediaInfo, error) {
			return &models.MediaInfo{
				ID:          1,
				MediaFileID: mediaFileID,
				Title:       "Test",
				MediaType:   "movie",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}, nil
		},
		UpdateMediaInfoFunc: func(info *models.MediaInfo) error {
			info.UpdatedAt = time.Now()
			return nil
		},
		GetAPICacheFunc: func(provider, query string) (*models.APICache, error) {
			return nil, nil
		},
		CreateAPICacheFunc: func(cache *models.APICache) error {
			cache.ID = 1
			cache.CreatedAt = time.Now()
			return nil
		},
		CreateLLMRequestFunc: func(request *models.LLMRequest) error {
			request.ID = 1
			request.CreatedAt = time.Now()
			return nil
		},
		CreateBatchProcessFunc: func(batch *models.BatchProcess) error {
			batch.ID = 1
			batch.CreatedAt = time.Now()
			batch.UpdatedAt = time.Now()
			return nil
		},
		UpdateBatchProcessFunc: func(batch *models.BatchProcess) error {
			batch.UpdatedAt = time.Now()
			return nil
		},
		CreateBatchProcessFileFunc: func(file *models.BatchProcessFile) error {
			file.ID = 1
			file.CreatedAt = time.Now()
			file.UpdatedAt = time.Now()
			return nil
		},
		UpdateBatchProcessFileFunc: func(file *models.BatchProcessFile) error {
			file.UpdatedAt = time.Now()
			return nil
		},
		GetBatchProcessFilesByBatchIDFunc: func(batchID int64) ([]models.BatchProcessFile, error) {
			return []models.BatchProcessFile{
				{
					ID:             1,
					BatchProcessID: batchID,
					MediaFileID:    1,
					Status:         "pending",
					CreatedAt:      time.Now(),
					UpdatedAt:      time.Now(),
				},
			}, nil
		},
		CreateNotificationFunc: func(notification *models.Notification) error {
			notification.ID = 1
			notification.CreatedAt = time.Now()
			return nil
		},
		GetPendingNotificationsFunc: func() ([]models.Notification, error) {
			return []models.Notification{
				{
					ID:          1,
					MediaFileID: 1,
					Type:        "success",
					Message:     "File processed successfully",
					Sent:        false,
					CreatedAt:   time.Now(),
				},
			}, nil
		},
		UpdateNotificationFunc: func(notification *models.Notification) error {
			notification.SentAt = time.Now()
			return nil
		},
		GetPendingMediaFilesFunc: func() ([]models.MediaFile, error) {
			return []models.MediaFile{
				{
					ID:           1,
					OriginalPath: "/test/path/file.mp4",
					OriginalName: "file.mp4",
					Status:       "pending",
					CreatedAt:    time.Now(),
					UpdatedAt:    time.Now(),
				},
			}, nil
		},
		GetMediaFilesByDirectoryFunc: func(directory string) ([]models.MediaFile, error) {
			return []models.MediaFile{
				{
					ID:           1,
					OriginalPath: directory + "/file.mp4",
					OriginalName: "file.mp4",
					Status:       "pending",
					CreatedAt:    time.Now(),
					UpdatedAt:    time.Now(),
				},
			}, nil
		},
		GetPendingBatchProcessesFunc: func() ([]models.BatchProcess, error) {
			return []models.BatchProcess{
				{
					ID:        1,
					Directory: "/test/path",
					FileCount: 3,
					Status:    "pending",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
			}, nil
		},
	}
}
