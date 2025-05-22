package testutil

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sleepstars/mediascanner/internal/config"
)

// CreateTempDir creates a temporary directory for testing
func CreateTempDir(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "mediascanner-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		os.RemoveAll(dir)
	})
	return dir
}

// CreateTestFile creates a test file with the given content
func CreateTestFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	err := os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	return path
}

// DefaultTestConfig returns a default configuration for testing
func DefaultTestConfig() *config.Config {
	return &config.Config{
		Database: config.DatabaseConfig{
			Type:     "postgres",
			Host:     "localhost",
			Port:     5432,
			User:     "postgres",
			Password: "postgres",
			Database: "mediascanner_test",
			SSLMode:  "disable",
		},
		LLM: config.LLMConfig{
			Provider:         "openai",
			APIKey:           "test-api-key",
			BaseURL:          "https://api.openai.com/v1",
			Model:            "gpt-3.5-turbo",
			SystemPrompt:     "You are a helpful assistant that analyzes media filenames.",
			BatchSystemPrompt: "You are a helpful assistant that analyzes batches of media filenames.",
			MaxRetries:       3,
			Timeout:          30,
		},
		APIs: config.APIConfig{
			TMDB: config.TMDBConfig{
				APIKey:       "test-tmdb-api-key",
				Language:     "en-US",
				IncludeAdult: false,
			},
			TVDB: config.TVDBConfig{
				APIKey:   "test-tvdb-api-key",
				Language: "en",
			},
			Bangumi: config.BangumiConfig{
				APIKey:    "test-bangumi-api-key",
				Language:  "en",
				UserAgent: "test-user-agent",
			},
		},
		FileOps: config.FileOpsConfig{
			Mode:            "copy",
			DestinationRoot: "/test/dest",
			DirectoryStructure: map[string][]string{
				"movies": {"Action", "Sci-Fi", "Drama"},
				"tv":     {"Comedy", "Drama", "Anime"},
			},
			MovieTemplate:   "{title} ({year})",
			TVShowTemplate:  "{title} ({year})",
			EpisodeTemplate: "{title} - S{season:02d}E{episode:02d} - {episode_title}",
		},
		Scanner: config.ScannerConfig{
			MediaDirs:       []string{"/test/media"},
			VideoExtensions: []string{".mp4", ".mkv", ".avi"},
			ExcludePatterns: []string{"sample", "trailer"},
			ExcludeDirs:     []string{"_UNPACK_", "tmp"},
			BatchThreshold:  3,
		},
		WorkerPool: config.WorkerPoolConfig{
			Enabled:             true,
			WorkerCount:         2,
			QueueSize:           10,
			BatchWorkerCount:    1,
			MaxConcurrentLLM:    2,
			MaxConcurrentAPI:    2,
			MaxConcurrentFileOp: 2,
		},
		Notification: config.NotificationConfig{
			Enabled:        false,
			Provider:       "telegram",
			TelegramToken:  "test-token",
			SuccessChannel: "test-channel",
			ErrorGroup:     "test-group",
		},
		LogLevel:     "info",
		ScanInterval: 5,
	}
}

// MockDB is a simple mock database for testing
type MockDB struct {
	MediaFiles       map[int64]*struct{}
	MediaInfos       map[int64]*struct{}
	APICaches        map[string]*struct{}
	BatchProcesses   map[int64]*struct{}
	Notifications    map[int64]*struct{}
	CreateMediaFileFunc func(interface{}) error
	GetMediaFileByPathFunc func(string) (interface{}, error)
	UpdateMediaFileFunc func(interface{}) error
	GetAPICache func(string, string) (interface{}, error)
	CreateAPICache func(interface{}) error
}

// NewMockDB creates a new mock database
func NewMockDB() *MockDB {
	return &MockDB{
		MediaFiles:     make(map[int64]*struct{}),
		MediaInfos:     make(map[int64]*struct{}),
		APICaches:      make(map[string]*struct{}),
		BatchProcesses: make(map[int64]*struct{}),
		Notifications:  make(map[int64]*struct{}),
	}
}
