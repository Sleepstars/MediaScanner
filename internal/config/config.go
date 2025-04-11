package config

import (
	"encoding/json"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	// General settings
	LogLevel     string `json:"log_level" yaml:"log_level"`
	ScanInterval int    `json:"scan_interval" yaml:"scan_interval"` // in minutes

	// LLM settings
	LLM LLMConfig `json:"llm" yaml:"llm"`

	// API settings
	APIs APIConfig `json:"apis" yaml:"apis"`

	// Database settings
	Database DatabaseConfig `json:"database" yaml:"database"`

	// Scanner settings
	Scanner ScannerConfig `json:"scanner" yaml:"scanner"`

	// File operations settings
	FileOps FileOpsConfig `json:"file_ops" yaml:"file_ops"`

	// Notification settings
	Notification NotificationConfig `json:"notification" yaml:"notification"`
}

// LLMConfig represents the LLM configuration
type LLMConfig struct {
	Provider          string `json:"provider" yaml:"provider"` // openai, one-api, etc.
	APIKey            string `json:"api_key" yaml:"api_key"`
	BaseURL           string `json:"base_url" yaml:"base_url"`
	Model             string `json:"model" yaml:"model"`
	SystemPrompt      string `json:"system_prompt" yaml:"system_prompt"`             // Custom system prompt for single file processing
	BatchSystemPrompt string `json:"batch_system_prompt" yaml:"batch_system_prompt"` // Custom system prompt for batch processing
	MaxRetries        int    `json:"max_retries" yaml:"max_retries"`
	Timeout           int    `json:"timeout" yaml:"timeout"` // in seconds
}

// APIConfig represents the API configuration
type APIConfig struct {
	TMDB    TMDBConfig    `json:"tmdb" yaml:"tmdb"`
	TVDB    TVDBConfig    `json:"tvdb" yaml:"tvdb"`
	Bangumi BangumiConfig `json:"bangumi" yaml:"bangumi"`
}

// TMDBConfig represents the TMDB API configuration
type TMDBConfig struct {
	APIKey       string `json:"api_key" yaml:"api_key"`
	Language     string `json:"language" yaml:"language"`
	IncludeAdult bool   `json:"include_adult" yaml:"include_adult"`
}

// TVDBConfig represents the TVDB API configuration
type TVDBConfig struct {
	APIKey   string `json:"api_key" yaml:"api_key"`
	Language string `json:"language" yaml:"language"`
}

// BangumiConfig represents the Bangumi API configuration
type BangumiConfig struct {
	APIKey    string `json:"api_key" yaml:"api_key"`
	Language  string `json:"language" yaml:"language"`
	UserAgent string `json:"user_agent" yaml:"user_agent"`
}

// DatabaseConfig represents the database configuration
type DatabaseConfig struct {
	Type     string `json:"type" yaml:"type"` // postgres
	Host     string `json:"host" yaml:"host"`
	Port     int    `json:"port" yaml:"port"`
	User     string `json:"user" yaml:"user"`
	Password string `json:"password" yaml:"password"`
	Database string `json:"database" yaml:"database"`
	SSLMode  string `json:"ssl_mode" yaml:"ssl_mode"`
}

// ScannerConfig represents the scanner configuration
type ScannerConfig struct {
	MediaDirs       []string `json:"media_dirs" yaml:"media_dirs"`
	ExcludeDirs     []string `json:"exclude_dirs" yaml:"exclude_dirs"`
	ExcludePatterns []string `json:"exclude_patterns" yaml:"exclude_patterns"`
	VideoExtensions []string `json:"video_extensions" yaml:"video_extensions"`
	BatchThreshold  int      `json:"batch_threshold" yaml:"batch_threshold"`
}

// FileOpsConfig represents the file operations configuration
type FileOpsConfig struct {
	Mode               string              `json:"mode" yaml:"mode"` // copy, move, symlink
	DestinationRoot    string              `json:"destination_root" yaml:"destination_root"`
	DirectoryStructure map[string][]string `json:"directory_structure" yaml:"directory_structure"`
	MovieTemplate      string              `json:"movie_template" yaml:"movie_template"`
	TVShowTemplate     string              `json:"tv_show_template" yaml:"tv_show_template"`
	EpisodeTemplate    string              `json:"episode_template" yaml:"episode_template"`
}

// NotificationConfig represents the notification configuration
type NotificationConfig struct {
	Enabled        bool   `json:"enabled" yaml:"enabled"`
	Provider       string `json:"provider" yaml:"provider"` // telegram
	TelegramToken  string `json:"telegram_token" yaml:"telegram_token"`
	SuccessChannel string `json:"success_channel" yaml:"success_channel"`
	ErrorGroup     string `json:"error_group" yaml:"error_group"`
}

// LoadConfig loads the configuration from a file
func LoadConfig(filePath string) (*Config, error) {
	file, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	var config Config

	// Determine file type by extension
	switch {
	case len(filePath) > 5 && filePath[len(filePath)-5:] == ".json":
		err = json.Unmarshal(file, &config)
	case len(filePath) > 5 && filePath[len(filePath)-5:] == ".yaml":
		err = yaml.Unmarshal(file, &config)
	case len(filePath) > 4 && filePath[len(filePath)-4:] == ".yml":
		err = yaml.Unmarshal(file, &config)
	default:
		return nil, fmt.Errorf("unsupported config file format: %s", filePath)
	}

	if err != nil {
		return nil, fmt.Errorf("error parsing config file: %w", err)
	}

	// Apply environment variable overrides
	applyEnvironmentOverrides(&config)

	return &config, nil
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		LogLevel:     "info",
		ScanInterval: 5, // 5 minutes
		LLM: LLMConfig{
			Provider: "openai",
			BaseURL:  "https://api.openai.com/v1",
			Model:    "gpt-3.5-turbo",
			SystemPrompt: `You are a media file analyzer that helps identify movies and TV shows from filenames.
Your task is to analyze the given filename and determine:
1. The correct title of the media
2. Whether it's a movie or TV show
3. For TV shows, identify the season and episode numbers
4. The year of release (if available)
5. The appropriate category for the media based on the provided directory structure

You should use the searchTMDB, searchTVDB, and searchBangumi functions to get accurate information.
For anime content, prioritize using searchBangumi after confirming it's anime through TMDB/TVDB.

Respond with a structured JSON containing the media information and the appropriate destination path.`,
			BatchSystemPrompt: `You are a media file analyzer that helps identify movies and TV shows from filenames.
Your task is to analyze the given filenames and determine for each file:
1. The correct title of the media
2. Whether it's a movie or TV show
3. For TV shows, identify the season and episode numbers
4. The year of release (if available)
5. The appropriate category for the media based on the provided directory structure

You should use the searchTMDB, searchTVDB, and searchBangumi functions to get accurate information.
For anime content, prioritize using searchBangumi after confirming it's anime through TMDB/TVDB.

Respond with a structured JSON array containing the media information and the appropriate destination path for each file.`,
			MaxRetries: 3,
			Timeout:    30,
		},
		APIs: APIConfig{
			TMDB: TMDBConfig{
				Language:     "en-US",
				IncludeAdult: false,
			},
			TVDB: TVDBConfig{
				Language: "en",
			},
			Bangumi: BangumiConfig{
				Language:  "zh-CN",
				UserAgent: "sleepstars/MediaScanner (https://github.com/sleepstars/MediaScanner)",
			},
		},
		Database: DatabaseConfig{
			Type:     "postgres",
			Host:     "localhost",
			Port:     5432,
			User:     "postgres",
			Password: "postgres",
			Database: "mediascanner",
			SSLMode:  "disable",
		},
		Scanner: ScannerConfig{
			MediaDirs:       []string{},
			ExcludeDirs:     []string{"sample", "extras", "featurettes"},
			ExcludePatterns: []string{"sample", "trailer", "extra", "featurette"},
			VideoExtensions: []string{".mkv", ".mp4", ".avi", ".ts", ".mov", ".wmv"},
			BatchThreshold:  50,
		},
		FileOps: FileOpsConfig{
			Mode:            "copy",
			DestinationRoot: "",
			DirectoryStructure: map[string][]string{
				"电影":  {"外语电影", "国语电影", "动画电影"},
				"电视剧": {"综艺", "纪录片", "动画", "国产剧", "欧美剧", "日韩剧", "其他"},
			},
			MovieTemplate:   "{title} ({year})",
			TVShowTemplate:  "{title} ({year})",
			EpisodeTemplate: "{title} - S{season:02d}E{episode:02d} - {episode_title}",
		},
		Notification: NotificationConfig{
			Enabled:        false,
			Provider:       "telegram",
			TelegramToken:  "",
			SuccessChannel: "",
			ErrorGroup:     "",
		},
	}
}

// applyEnvironmentOverrides applies environment variable overrides to the configuration
func applyEnvironmentOverrides(config *Config) {
	// LLM settings
	if apiKey := os.Getenv("LLM_API_KEY"); apiKey != "" {
		config.LLM.APIKey = apiKey
	}
	if baseURL := os.Getenv("LLM_BASE_URL"); baseURL != "" {
		config.LLM.BaseURL = baseURL
	}
	if model := os.Getenv("LLM_MODEL"); model != "" {
		config.LLM.Model = model
	}
	if systemPrompt := os.Getenv("LLM_SYSTEM_PROMPT"); systemPrompt != "" {
		config.LLM.SystemPrompt = systemPrompt
	}
	if batchSystemPrompt := os.Getenv("LLM_BATCH_SYSTEM_PROMPT"); batchSystemPrompt != "" {
		config.LLM.BatchSystemPrompt = batchSystemPrompt
	}

	// API settings
	if tmdbAPIKey := os.Getenv("TMDB_API_KEY"); tmdbAPIKey != "" {
		config.APIs.TMDB.APIKey = tmdbAPIKey
	}
	if tvdbAPIKey := os.Getenv("TVDB_API_KEY"); tvdbAPIKey != "" {
		config.APIs.TVDB.APIKey = tvdbAPIKey
	}
	if bangumiAPIKey := os.Getenv("BANGUMI_API_KEY"); bangumiAPIKey != "" {
		config.APIs.Bangumi.APIKey = bangumiAPIKey
	}
	if bangumiUserAgent := os.Getenv("BANGUMI_USER_AGENT"); bangumiUserAgent != "" {
		config.APIs.Bangumi.UserAgent = bangumiUserAgent
	}

	// Database settings
	if dbHost := os.Getenv("DB_HOST"); dbHost != "" {
		config.Database.Host = dbHost
	}
	if dbPort := os.Getenv("DB_PORT"); dbPort != "" {
		var port int
		fmt.Sscanf(dbPort, "%d", &port)
		if port > 0 {
			config.Database.Port = port
		}
	}
	if dbUser := os.Getenv("DB_USER"); dbUser != "" {
		config.Database.User = dbUser
	}
	if dbPassword := os.Getenv("DB_PASSWORD"); dbPassword != "" {
		config.Database.Password = dbPassword
	}
	if dbName := os.Getenv("DB_NAME"); dbName != "" {
		config.Database.Database = dbName
	}

	// Notification settings
	if telegramToken := os.Getenv("TELEGRAM_TOKEN"); telegramToken != "" {
		config.Notification.TelegramToken = telegramToken
	}
	if successChannel := os.Getenv("TELEGRAM_SUCCESS_CHANNEL"); successChannel != "" {
		config.Notification.SuccessChannel = successChannel
	}
	if errorGroup := os.Getenv("TELEGRAM_ERROR_GROUP"); errorGroup != "" {
		config.Notification.ErrorGroup = errorGroup
	}
}
