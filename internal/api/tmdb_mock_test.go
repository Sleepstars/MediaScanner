package api

import (
	"errors"
	
	tmdb "github.com/cyruzin/golang-tmdb"
	"github.com/sleepstars/mediascanner/internal/models"
)

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

func (m *MockTMDBClient) SetClientAutoRetry() {
	// No-op for mock
}

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

// Implement the rest of the database.DatabaseInterface methods with no-op implementations
func (m *MockDatabase) GetMediaFileByPath(path string) (*models.MediaFile, error) {
	return nil, errors.New("not implemented")
}

func (m *MockDatabase) CreateMediaFile(file *models.MediaFile) error {
	return errors.New("not implemented")
}

func (m *MockDatabase) UpdateMediaFile(file *models.MediaFile) error {
	return errors.New("not implemented")
}

func (m *MockDatabase) GetMediaFileByID(id int64) (*models.MediaFile, error) {
	return nil, errors.New("not implemented")
}

func (m *MockDatabase) CreateMediaInfo(info *models.MediaInfo) error {
	return errors.New("not implemented")
}

func (m *MockDatabase) GetMediaInfoByMediaFileID(mediaFileID int64) (*models.MediaInfo, error) {
	return nil, errors.New("not implemented")
}

func (m *MockDatabase) UpdateMediaInfo(info *models.MediaInfo) error {
	return errors.New("not implemented")
}

func (m *MockDatabase) CreateLLMRequest(request *models.LLMRequest) error {
	return errors.New("not implemented")
}

func (m *MockDatabase) CreateBatchProcess(batch *models.BatchProcess) error {
	return errors.New("not implemented")
}

func (m *MockDatabase) UpdateBatchProcess(batch *models.BatchProcess) error {
	return errors.New("not implemented")
}

func (m *MockDatabase) CreateBatchProcessFile(file *models.BatchProcessFile) error {
	return errors.New("not implemented")
}

func (m *MockDatabase) UpdateBatchProcessFile(file *models.BatchProcessFile) error {
	return errors.New("not implemented")
}

func (m *MockDatabase) GetBatchProcessFilesByBatchID(batchID int64) ([]models.BatchProcessFile, error) {
	return nil, errors.New("not implemented")
}

func (m *MockDatabase) CreateNotification(notification *models.Notification) error {
	return errors.New("not implemented")
}

func (m *MockDatabase) GetPendingNotifications() ([]models.Notification, error) {
	return nil, errors.New("not implemented")
}

func (m *MockDatabase) UpdateNotification(notification *models.Notification) error {
	return errors.New("not implemented")
}

func (m *MockDatabase) GetPendingMediaFiles() ([]models.MediaFile, error) {
	return nil, errors.New("not implemented")
}

func (m *MockDatabase) GetMediaFilesByDirectory(directory string) ([]models.MediaFile, error) {
	return nil, errors.New("not implemented")
}

func (m *MockDatabase) GetPendingBatchProcesses() ([]models.BatchProcess, error) {
	return nil, errors.New("not implemented")
}

func (m *MockDatabase) Close() error {
	return errors.New("not implemented")
}
