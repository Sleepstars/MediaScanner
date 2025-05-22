package api

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	tmdb "github.com/cyruzin/golang-tmdb"
	"github.com/sleepstars/mediascanner/internal/config"
	"github.com/sleepstars/mediascanner/internal/models"
)

// TMDBClientInterface defines the interface for TMDB client operations
type TMDBClientInterface interface {
	GetSearchMovies(query string, options map[string]string) (*tmdb.SearchMovies, error)
	GetSearchTVShow(query string, options map[string]string) (*tmdb.SearchTVShows, error)
	GetMovieDetails(id int, options map[string]string) (*tmdb.MovieDetails, error)
	GetTVDetails(id int, options map[string]string) (*tmdb.TVDetails, error)
	GetTVSeasonDetails(id, seasonNumber int, options map[string]string) (*tmdb.TVSeasonDetails, error)
	GetTVEpisodeDetails(id, seasonNumber, episodeNumber int, options map[string]string) (*tmdb.TVEpisodeDetails, error)
	SetClientAutoRetry()
}

// MockTMDBClient is a mock implementation of the TMDB client
type MockTMDBClient struct {
	GetSearchMoviesFunc    func(query string, options map[string]string) (*tmdb.SearchMovies, error)
	GetSearchTVShowFunc    func(query string, options map[string]string) (*tmdb.SearchTVShows, error)
	GetMovieDetailsFunc    func(id int, options map[string]string) (*tmdb.MovieDetails, error)
	GetTVDetailsFunc       func(id int, options map[string]string) (*tmdb.TVDetails, error)
	GetTVSeasonDetailsFunc func(id, seasonNumber int, options map[string]string) (*tmdb.TVSeasonDetails, error)
	GetTVEpisodeDetailsFunc func(id, seasonNumber, episodeNumber int, options map[string]string) (*tmdb.TVEpisodeDetails, error)
}

func (m *MockTMDBClient) GetSearchMovies(query string, options map[string]string) (*tmdb.SearchMovies, error) {
	return m.GetSearchMoviesFunc(query, options)
}

func (m *MockTMDBClient) GetSearchTVShow(query string, options map[string]string) (*tmdb.SearchTVShows, error) {
	return m.GetSearchTVShowFunc(query, options)
}

func (m *MockTMDBClient) GetMovieDetails(id int, options map[string]string) (*tmdb.MovieDetails, error) {
	return m.GetMovieDetailsFunc(id, options)
}

func (m *MockTMDBClient) GetTVDetails(id int, options map[string]string) (*tmdb.TVDetails, error) {
	return m.GetTVDetailsFunc(id, options)
}

func (m *MockTMDBClient) GetTVSeasonDetails(id, seasonNumber int, options map[string]string) (*tmdb.TVSeasonDetails, error) {
	return m.GetTVSeasonDetailsFunc(id, seasonNumber, options)
}

func (m *MockTMDBClient) GetTVEpisodeDetails(id, seasonNumber, episodeNumber int, options map[string]string) (*tmdb.TVEpisodeDetails, error) {
	return m.GetTVEpisodeDetailsFunc(id, seasonNumber, episodeNumber, options)
}

func (m *MockTMDBClient) SetClientAutoRetry() {}

// MockDatabase is a mock implementation of the database
type MockDatabase struct {
	GetAPICacheFunc    func(provider, query string) (*models.APICache, error)
	CreateAPICacheFunc func(cache *models.APICache) error
}

func (m *MockDatabase) GetAPICache(provider, query string) (*models.APICache, error) {
	return m.GetAPICacheFunc(provider, query)
}

func (m *MockDatabase) CreateAPICache(cache *models.APICache) error {
	return m.CreateAPICacheFunc(cache)
}

func TestNewTMDBClient(t *testing.T) {
	// Test with valid config
	cfg := &config.TMDBConfig{
		APIKey:       "test-api-key",
		Language:     "en-US",
		IncludeAdult: false,
	}

	db := &MockDatabase{}

	client, err := NewTMDBClient(cfg, db)
	if err != nil {
		t.Fatalf("NewTMDBClient failed with valid config: %v", err)
	}

	if client == nil {
		t.Fatal("Expected non-nil TMDBClient instance")
	}

	// Test with empty API key
	cfg = &config.TMDBConfig{
		APIKey:       "",
		Language:     "en-US",
		IncludeAdult: false,
	}

	_, err = NewTMDBClient(cfg, db)
	if err == nil {
		t.Fatal("Expected error for empty API key, got nil")
	}

	expectedError := "TMDB API key is required"
	if err.Error() != expectedError {
		t.Errorf("Expected error message %q, got %q", expectedError, err.Error())
	}
}

func TestSearchMovie(t *testing.T) {
	// Create a mock TMDB client
	mockTMDBClient := &MockTMDBClient{
		GetSearchMoviesFunc: func(query string, options map[string]string) (*tmdb.SearchMovies, error) {
			// Verify the query and options
			if query != "The Matrix" {
				t.Errorf("Expected query to be 'The Matrix', got %q", query)
			}

			if options["language"] != "en-US" {
				t.Errorf("Expected language to be 'en-US', got %q", options["language"])
			}

			if options["year"] != "1999" {
				t.Errorf("Expected year to be '1999', got %q", options["year"])
			}

			// Return a mock response
			return &tmdb.SearchMovies{
				Page: 1,
				Results: []struct {
					Adult            bool    `json:"adult"`
					BackdropPath     string  `json:"backdrop_path"`
					GenreIDs         []int64 `json:"genre_ids"`
					ID               int64   `json:"id"`
					OriginalLanguage string  `json:"original_language"`
					OriginalTitle    string  `json:"original_title"`
					Overview         string  `json:"overview"`
					Popularity       float32 `json:"popularity"`
					PosterPath       string  `json:"poster_path"`
					ReleaseDate      string  `json:"release_date"`
					Title            string  `json:"title"`
					Video            bool    `json:"video"`
					VoteAverage      float32 `json:"vote_average"`
					VoteCount        int64   `json:"vote_count"`
				}{
					{
						ID:            603,
						Title:         "The Matrix",
						OriginalTitle: "The Matrix",
						Overview:      "Set in the 22nd century, The Matrix tells the story of a computer hacker who joins a group of underground insurgents fighting the vast and powerful computers who now rule the earth.",
						ReleaseDate:   "1999-03-30",
						PosterPath:    "/path/to/poster.jpg",
						BackdropPath:  "/path/to/backdrop.jpg",
						Popularity:    100.0,
						VoteAverage:   8.7,
					},
				},
				TotalPages:   1,
				TotalResults: 1,
			}, nil
		},
	}

	// Create a mock database that returns a cache miss
	mockDB := &MockDatabase{
		GetAPICacheFunc: func(provider, query string) (*models.APICache, error) {
			return nil, errors.New("cache miss")
		},
		CreateAPICacheFunc: func(cache *models.APICache) error {
			// Verify the cache
			if cache.Provider != "tmdb" {
				t.Errorf("Expected provider to be 'tmdb', got %q", cache.Provider)
			}

			if cache.Query != "movie:The Matrix:1999" {
				t.Errorf("Expected query to be 'movie:The Matrix:1999', got %q", cache.Query)
			}

			return nil
		},
	}

	// Create the TMDB client with mocks
	cfg := &config.TMDBConfig{
		APIKey:       "test-api-key",
		Language:     "en-US",
		IncludeAdult: false,
	}

	client := &TMDBClient{
		client: mockTMDBClient,
		config: cfg,
		db:     mockDB,
	}

	// Search for a movie
	result, err := client.SearchMovie(context.Background(), "The Matrix", 1999)
	if err != nil {
		t.Fatalf("SearchMovie failed: %v", err)
	}

	// Verify the result
	if result.Query != "The Matrix" {
		t.Errorf("Expected query to be 'The Matrix', got %q", result.Query)
	}

	if result.Year != 1999 {
		t.Errorf("Expected year to be 1999, got %d", result.Year)
	}

	if len(result.Movies) != 1 {
		t.Fatalf("Expected 1 movie, got %d", len(result.Movies))
	}

	if result.Movies[0].ID != 603 {
		t.Errorf("Expected movie ID to be 603, got %d", result.Movies[0].ID)
	}

	if result.Movies[0].Title != "The Matrix" {
		t.Errorf("Expected movie title to be 'The Matrix', got %q", result.Movies[0].Title)
	}

	if result.Movies[0].ReleaseYear != 1999 {
		t.Errorf("Expected release year to be 1999, got %d", result.Movies[0].ReleaseYear)
	}
}

func TestSearchMovie_CacheHit(t *testing.T) {
	// Create a mock database that returns a cache hit
	mockDB := &MockDatabase{
		GetAPICacheFunc: func(provider, query string) (*models.APICache, error) {
			// Verify the cache query
			if provider != "tmdb" {
				t.Errorf("Expected provider to be 'tmdb', got %q", provider)
			}

			if query != "movie:The Matrix:1999" {
				t.Errorf("Expected query to be 'movie:The Matrix:1999', got %q", query)
			}

			// Return a cached result
			cachedResult := &MovieSearchResult{
				Query: "The Matrix",
				Year:  1999,
				Movies: []Movie{
					{
						ID:            603,
						Title:         "The Matrix",
						OriginalTitle: "The Matrix",
						ReleaseYear:   1999,
						Overview:      "Cached overview",
					},
				},
			}

			resultJSON, _ := json.Marshal(cachedResult)

			return &models.APICache{
				Provider:  "tmdb",
				Query:     query,
				Response:  string(resultJSON),
				ExpiresAt: time.Now().Add(1 * time.Hour),
			}, nil
		},
	}

	// Create the TMDB client with mocks
	cfg := &config.TMDBConfig{
		APIKey:       "test-api-key",
		Language:     "en-US",
		IncludeAdult: false,
	}

	client := &TMDBClient{
		client: nil, // Not used in this test
		config: cfg,
		db:     mockDB,
	}

	// Search for a movie
	result, err := client.SearchMovie(context.Background(), "The Matrix", 1999)
	if err != nil {
		t.Fatalf("SearchMovie failed: %v", err)
	}

	// Verify the result
	if result.Query != "The Matrix" {
		t.Errorf("Expected query to be 'The Matrix', got %q", result.Query)
	}

	if result.Year != 1999 {
		t.Errorf("Expected year to be 1999, got %d", result.Year)
	}

	if len(result.Movies) != 1 {
		t.Fatalf("Expected 1 movie, got %d", len(result.Movies))
	}

	if result.Movies[0].ID != 603 {
		t.Errorf("Expected movie ID to be 603, got %d", result.Movies[0].ID)
	}

	if result.Movies[0].Title != "The Matrix" {
		t.Errorf("Expected movie title to be 'The Matrix', got %q", result.Movies[0].Title)
	}

	if result.Movies[0].Overview != "Cached overview" {
		t.Errorf("Expected overview to be 'Cached overview', got %q", result.Movies[0].Overview)
	}
}

func TestSearchMovie_APIError(t *testing.T) {
	// Create a mock TMDB client that returns an error
	mockTMDBClient := &MockTMDBClient{
		GetSearchMoviesFunc: func(query string, options map[string]string) (*tmdb.SearchMovies, error) {
			return nil, errors.New("API error")
		},
	}

	// Create a mock database that returns a cache miss
	mockDB := &MockDatabase{
		GetAPICacheFunc: func(provider, query string) (*models.APICache, error) {
			return nil, errors.New("cache miss")
		},
	}

	// Create the TMDB client with mocks
	cfg := &config.TMDBConfig{
		APIKey:       "test-api-key",
		Language:     "en-US",
		IncludeAdult: false,
	}

	client := &TMDBClient{
		client: mockTMDBClient,
		config: cfg,
		db:     mockDB,
	}

	// Search for a movie
	_, err := client.SearchMovie(context.Background(), "The Matrix", 1999)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	expectedError := "TMDB search movie error: API error"
	if err.Error() != expectedError {
		t.Errorf("Expected error message %q, got %q", expectedError, err.Error())
	}
}

func TestSearchTV(t *testing.T) {
	// Create a mock TMDB client
	mockTMDBClient := &MockTMDBClient{
		GetSearchTVShowFunc: func(query string, options map[string]string) (*tmdb.SearchTVShows, error) {
			// Verify the query and options
			if query != "Breaking Bad" {
				t.Errorf("Expected query to be 'Breaking Bad', got %q", query)
			}

			if options["language"] != "en-US" {
				t.Errorf("Expected language to be 'en-US', got %q", options["language"])
			}

			if options["first_air_date_year"] != "2008" {
				t.Errorf("Expected first_air_date_year to be '2008', got %q", options["first_air_date_year"])
			}

			// Return a mock response
			return &tmdb.SearchTVShows{
				Page: 1,
				Results: []struct {
					BackdropPath     string   `json:"backdrop_path"`
					FirstAirDate     string   `json:"first_air_date"`
					GenreIDs         []int64  `json:"genre_ids"`
					ID               int64    `json:"id"`
					Name             string   `json:"name"`
					OriginCountry    []string `json:"origin_country"`
					OriginalLanguage string   `json:"original_language"`
					OriginalName     string   `json:"original_name"`
					Overview         string   `json:"overview"`
					Popularity       float32  `json:"popularity"`
					PosterPath       string   `json:"poster_path"`
					VoteAverage      float32  `json:"vote_average"`
					VoteCount        int64    `json:"vote_count"`
				}{
					{
						ID:           1396,
						Name:         "Breaking Bad",
						OriginalName: "Breaking Bad",
						Overview:     "When Walter White, a New Mexico chemistry teacher, is diagnosed with Stage III cancer and given a prognosis of only two years left to live.",
						FirstAirDate: "2008-01-20",
						PosterPath:   "/path/to/poster.jpg",
						BackdropPath: "/path/to/backdrop.jpg",
						Popularity:   100.0,
						VoteAverage:  8.7,
					},
				},
				TotalPages:   1,
				TotalResults: 1,
			}, nil
		},
	}

	// Create a mock database that returns a cache miss
	mockDB := &MockDatabase{
		GetAPICacheFunc: func(provider, query string) (*models.APICache, error) {
			return nil, errors.New("cache miss")
		},
		CreateAPICacheFunc: func(cache *models.APICache) error {
			return nil
		},
	}

	// Create the TMDB client with mocks
	cfg := &config.TMDBConfig{
		APIKey:       "test-api-key",
		Language:     "en-US",
		IncludeAdult: false,
	}

	client := &TMDBClient{
		client: mockTMDBClient,
		config: cfg,
		db:     mockDB,
	}

	// Search for a TV show
	result, err := client.SearchTV(context.Background(), "Breaking Bad", 2008)
	if err != nil {
		t.Fatalf("SearchTV failed: %v", err)
	}

	// Verify the result
	if result.Query != "Breaking Bad" {
		t.Errorf("Expected query to be 'Breaking Bad', got %q", result.Query)
	}

	if result.Year != 2008 {
		t.Errorf("Expected year to be 2008, got %d", result.Year)
	}

	if len(result.Shows) != 1 {
		t.Fatalf("Expected 1 show, got %d", len(result.Shows))
	}

	if result.Shows[0].ID != 1396 {
		t.Errorf("Expected show ID to be 1396, got %d", result.Shows[0].ID)
	}

	if result.Shows[0].Name != "Breaking Bad" {
		t.Errorf("Expected show name to be 'Breaking Bad', got %q", result.Shows[0].Name)
	}

	if result.Shows[0].FirstAirYear != 2008 {
		t.Errorf("Expected first air year to be 2008, got %d", result.Shows[0].FirstAirYear)
	}
}

func TestGetMovieDetails(t *testing.T) {
	// Create a mock TMDB client
	mockTMDBClient := &MockTMDBClient{
		GetMovieDetailsFunc: func(id int, options map[string]string) (*tmdb.MovieDetails, error) {
			// Verify the ID and options
			if id != 603 {
				t.Errorf("Expected ID to be 603, got %d", id)
			}

			if options["language"] != "en-US" {
				t.Errorf("Expected language to be 'en-US', got %q", options["language"])
			}

			if options["append_to_response"] != "credits,images,external_ids" {
				t.Errorf("Expected append_to_response to be 'credits,images,external_ids', got %q", options["append_to_response"])
			}

			// Return a mock response
			return &tmdb.MovieDetails{
				ID:            603,
				Title:         "The Matrix",
				OriginalTitle: "The Matrix",
				Overview:      "Set in the 22nd century, The Matrix tells the story of a computer hacker who joins a group of underground insurgents fighting the vast and powerful computers who now rule the earth.",
				ReleaseDate:   "1999-03-30",
				PosterPath:    "/path/to/poster.jpg",
				BackdropPath:  "/path/to/backdrop.jpg",
				IMDbID:        "tt0133093",
				Runtime:       136,
				VoteAverage:   8.7,
				VoteCount:     10000,
				Genres: []struct {
					ID   int64  `json:"id"`
					Name string `json:"name"`
				}{
					{ID: 28, Name: "Action"},
					{ID: 878, Name: "Science Fiction"},
				},
				ProductionCountries: []struct {
					Iso31661 string `json:"iso_3166_1"`
					Name     string `json:"name"`
				}{
					{Iso31661: "US", Name: "United States of America"},
				},
				SpokenLanguages: []struct {
					Iso6391 string `json:"iso_639_1"`
					Name    string `json:"name"`
				}{
					{Iso6391: "en", Name: "English"},
				},
			}, nil
		},
	}

	// Create a mock database that returns a cache miss
	mockDB := &MockDatabase{
		GetAPICacheFunc: func(provider, query string) (*models.APICache, error) {
			return nil, errors.New("cache miss")
		},
		CreateAPICacheFunc: func(cache *models.APICache) error {
			return nil
		},
	}

	// Create the TMDB client with mocks
	cfg := &config.TMDBConfig{
		APIKey:       "test-api-key",
		Language:     "en-US",
		IncludeAdult: false,
	}

	client := &TMDBClient{
		client: mockTMDBClient,
		config: cfg,
		db:     mockDB,
	}

	// Get movie details
	result, err := client.GetMovieDetails(context.Background(), 603)
	if err != nil {
		t.Fatalf("GetMovieDetails failed: %v", err)
	}

	// Verify the result
	if result.ID != 603 {
		t.Errorf("Expected ID to be 603, got %d", result.ID)
	}

	if result.Title != "The Matrix" {
		t.Errorf("Expected title to be 'The Matrix', got %q", result.Title)
	}

	if result.ReleaseYear != 1999 {
		t.Errorf("Expected release year to be 1999, got %d", result.ReleaseYear)
	}

	if result.ImdbID != "tt0133093" {
		t.Errorf("Expected IMDb ID to be 'tt0133093', got %q", result.ImdbID)
	}

	if len(result.Genres) != 2 {
		t.Fatalf("Expected 2 genres, got %d", len(result.Genres))
	}

	if result.Genres[0] != "Action" {
		t.Errorf("Expected first genre to be 'Action', got %q", result.Genres[0])
	}

	if result.Genres[1] != "Science Fiction" {
		t.Errorf("Expected second genre to be 'Science Fiction', got %q", result.Genres[1])
	}

	if len(result.Countries) != 1 {
		t.Fatalf("Expected 1 country, got %d", len(result.Countries))
	}

	if result.Countries[0] != "United States of America" {
		t.Errorf("Expected country to be 'United States of America', got %q", result.Countries[0])
	}

	if len(result.Languages) != 1 {
		t.Fatalf("Expected 1 language, got %d", len(result.Languages))
	}

	if result.Languages[0] != "English" {
		t.Errorf("Expected language to be 'English', got %q", result.Languages[0])
	}

	if result.Runtime != 136 {
		t.Errorf("Expected runtime to be 136, got %d", result.Runtime)
	}
}

func TestGetImageURL(t *testing.T) {
	// Create the TMDB client
	cfg := &config.TMDBConfig{
		APIKey:       "test-api-key",
		Language:     "en-US",
		IncludeAdult: false,
	}

	client := &TMDBClient{
		config: cfg,
	}

	// Test with a poster path
	path := "/path/to/poster.jpg"
	size := "w500"

	url := client.GetImageURL(path, size)

	// The actual URL format depends on the tmdb library implementation
	// We're just checking that it's not empty and contains the path
	if url == "" {
		t.Error("Expected non-empty URL")
	}

	if !strings.Contains(url, path) {
		t.Errorf("Expected URL to contain %q, got %q", path, url)
	}
}
