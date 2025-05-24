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
	LogLevel     string       `json:"log_level" yaml:"log_level"` // Deprecated: use Logger.Level instead
	Logger       LoggerConfig `json:"logger" yaml:"logger"`
	ScanInterval int          `json:"scan_interval" yaml:"scan_interval"` // in minutes

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

	// Worker pool settings
	WorkerPool WorkerPoolConfig `json:"worker_pool" yaml:"worker_pool"`

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

	// Rate limiting settings
	RateLimiting RateLimitingConfig `json:"rate_limiting" yaml:"rate_limiting"`

	// Cache settings
	Cache CacheConfig `json:"cache" yaml:"cache"`
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

// RateLimitingConfig represents the rate limiting configuration
type RateLimitingConfig struct {
	Enabled bool `json:"enabled" yaml:"enabled"`

	// Rate limits for each provider (requests per second)
	TMDB    float64 `json:"tmdb" yaml:"tmdb"`
	TVDB    float64 `json:"tvdb" yaml:"tvdb"`
	Bangumi float64 `json:"bangumi" yaml:"bangumi"`

	// Burst sizes for each provider
	TMDBBurst    int `json:"tmdb_burst" yaml:"tmdb_burst"`
	TVDBBurst    int `json:"tvdb_burst" yaml:"tvdb_burst"`
	BangumiBurst int `json:"bangumi_burst" yaml:"bangumi_burst"`
}

// CacheConfig represents the cache configuration
type CacheConfig struct {
	Enabled bool `json:"enabled" yaml:"enabled"`

	// TTL (time to live) for each type of cache entry (in hours)
	SearchTTL  int `json:"search_ttl" yaml:"search_ttl"`   // For search results
	DetailsTTL int `json:"details_ttl" yaml:"details_ttl"` // For detailed information
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
	ScanInterval    int      `json:"scan_interval" yaml:"scan_interval"`

	// File system monitoring settings
	UseWatcher      bool            `json:"use_watcher" yaml:"use_watcher"`
	WatcherSettings WatcherSettings `json:"watcher_settings" yaml:"watcher_settings"`
}

// WatcherSettings represents the file system watcher settings
type WatcherSettings struct {
	// Whether to watch subdirectories recursively
	Recursive bool `json:"recursive" yaml:"recursive"`

	// Delay in seconds before processing a new file (to avoid processing files that are still being written)
	ProcessDelay int `json:"process_delay" yaml:"process_delay"`
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

// WorkerPoolConfig represents the worker pool configuration
type WorkerPoolConfig struct {
	Enabled             bool `json:"enabled" yaml:"enabled"`
	WorkerCount         int  `json:"worker_count" yaml:"worker_count"`                     // Number of workers for processing files
	QueueSize           int  `json:"queue_size" yaml:"queue_size"`                         // Size of the task queue
	BatchWorkerCount    int  `json:"batch_worker_count" yaml:"batch_worker_count"`         // Number of workers for batch processing
	MaxConcurrentLLM    int  `json:"max_concurrent_llm" yaml:"max_concurrent_llm"`         // Maximum concurrent LLM requests
	MaxConcurrentAPI    int  `json:"max_concurrent_api" yaml:"max_concurrent_api"`         // Maximum concurrent API requests
	MaxConcurrentFileOp int  `json:"max_concurrent_file_op" yaml:"max_concurrent_file_op"` // Maximum concurrent file operations
}

// NotificationConfig represents the notification configuration
type NotificationConfig struct {
	Enabled        bool   `json:"enabled" yaml:"enabled"`
	Provider       string `json:"provider" yaml:"provider"` // telegram
	TelegramToken  string `json:"telegram_token" yaml:"telegram_token"`
	SuccessChannel string `json:"success_channel" yaml:"success_channel"`
	ErrorGroup     string `json:"error_group" yaml:"error_group"`
}

// LoggerConfig represents the logger configuration
type LoggerConfig struct {
	// Level is the log level (debug, info, warn, error, fatal)
	Level string `json:"level" yaml:"level"`

	// Format is the log format (json, console)
	Format string `json:"format" yaml:"format"`

	// Output is the log output (stdout, stderr, file)
	Output string `json:"output" yaml:"output"`

	// File is the log file path (only used if Output is "file")
	File string `json:"file" yaml:"file"`

	// MaxSize is the maximum size in megabytes of the log file before it gets rotated
	MaxSize int `json:"max_size" yaml:"max_size"`

	// MaxBackups is the maximum number of old log files to retain
	MaxBackups int `json:"max_backups" yaml:"max_backups"`

	// MaxAge is the maximum number of days to retain old log files
	MaxAge int `json:"max_age" yaml:"max_age"`

	// Compress determines if the rotated log files should be compressed
	Compress bool `json:"compress" yaml:"compress"`
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
		LogLevel: "info", // Deprecated: use Logger.Level instead
		Logger: LoggerConfig{
			Level:      "info",
			Format:     "json",
			Output:     "stdout",
			File:       "logs/mediascanner.log",
			MaxSize:    100, // 100 MB
			MaxBackups: 3,
			MaxAge:     28, // 28 days
			Compress:   true,
		},
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
			RateLimiting: RateLimitingConfig{
				Enabled:      true,
				TMDB:         5.0, // 5 requests per second
				TVDB:         2.0, // 2 requests per second
				Bangumi:      1.0, // 1 request per second
				TMDBBurst:    10,  // Burst of 10 requests
				TVDBBurst:    5,   // Burst of 5 requests
				BangumiBurst: 3,   // Burst of 3 requests
			},
			Cache: CacheConfig{
				Enabled:    true,
				SearchTTL:  24,     // Cache search results for 24 hours
				DetailsTTL: 7 * 24, // Cache details for 7 days
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
			ScanInterval:    5, // 5 minutes
			UseWatcher:      true,
			WatcherSettings: WatcherSettings{
				Recursive:    true,
				ProcessDelay: 30, // 30 seconds
			},
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
		WorkerPool: WorkerPoolConfig{
			Enabled:             true,
			WorkerCount:         4,
			QueueSize:           100,
			BatchWorkerCount:    2,
			MaxConcurrentLLM:    2,
			MaxConcurrentAPI:    5,
			MaxConcurrentFileOp: 3,
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
	// Logger settings
	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		config.Logger.Level = logLevel
		// Also set the deprecated LogLevel for backward compatibility
		config.LogLevel = logLevel
	}
	if logFormat := os.Getenv("LOG_FORMAT"); logFormat != "" {
		config.Logger.Format = logFormat
	}
	if logOutput := os.Getenv("LOG_OUTPUT"); logOutput != "" {
		config.Logger.Output = logOutput
	}
	if logFile := os.Getenv("LOG_FILE"); logFile != "" {
		config.Logger.File = logFile
	}
	if logMaxSize := os.Getenv("LOG_MAX_SIZE"); logMaxSize != "" {
		var maxSize int
		fmt.Sscanf(logMaxSize, "%d", &maxSize)
		if maxSize > 0 {
			config.Logger.MaxSize = maxSize
		}
	}
	if logMaxBackups := os.Getenv("LOG_MAX_BACKUPS"); logMaxBackups != "" {
		var maxBackups int
		fmt.Sscanf(logMaxBackups, "%d", &maxBackups)
		if maxBackups > 0 {
			config.Logger.MaxBackups = maxBackups
		}
	}
	if logMaxAge := os.Getenv("LOG_MAX_AGE"); logMaxAge != "" {
		var maxAge int
		fmt.Sscanf(logMaxAge, "%d", &maxAge)
		if maxAge > 0 {
			config.Logger.MaxAge = maxAge
		}
	}
	if logCompress := os.Getenv("LOG_COMPRESS"); logCompress != "" {
		config.Logger.Compress = logCompress == "true"
	}

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

	// Rate limiting settings
	if rateLimitingEnabled := os.Getenv("RATE_LIMITING_ENABLED"); rateLimitingEnabled != "" {
		config.APIs.RateLimiting.Enabled = rateLimitingEnabled == "true"
	}
	if tmdbRateLimit := os.Getenv("TMDB_RATE_LIMIT"); tmdbRateLimit != "" {
		var rate float64
		fmt.Sscanf(tmdbRateLimit, "%f", &rate)
		if rate > 0 {
			config.APIs.RateLimiting.TMDB = rate
		}
	}
	if tvdbRateLimit := os.Getenv("TVDB_RATE_LIMIT"); tvdbRateLimit != "" {
		var rate float64
		fmt.Sscanf(tvdbRateLimit, "%f", &rate)
		if rate > 0 {
			config.APIs.RateLimiting.TVDB = rate
		}
	}
	if bangumiRateLimit := os.Getenv("BANGUMI_RATE_LIMIT"); bangumiRateLimit != "" {
		var rate float64
		fmt.Sscanf(bangumiRateLimit, "%f", &rate)
		if rate > 0 {
			config.APIs.RateLimiting.Bangumi = rate
		}
	}

	// Cache settings
	if cacheEnabled := os.Getenv("CACHE_ENABLED"); cacheEnabled != "" {
		config.APIs.Cache.Enabled = cacheEnabled == "true"
	}
	if searchTTL := os.Getenv("CACHE_SEARCH_TTL"); searchTTL != "" {
		var ttl int
		fmt.Sscanf(searchTTL, "%d", &ttl)
		if ttl > 0 {
			config.APIs.Cache.SearchTTL = ttl
		}
	}
	if detailsTTL := os.Getenv("CACHE_DETAILS_TTL"); detailsTTL != "" {
		var ttl int
		fmt.Sscanf(detailsTTL, "%d", &ttl)
		if ttl > 0 {
			config.APIs.Cache.DetailsTTL = ttl
		}
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

	// Worker pool settings
	if workerCount := os.Getenv("WORKER_COUNT"); workerCount != "" {
		var count int
		fmt.Sscanf(workerCount, "%d", &count)
		if count > 0 {
			config.WorkerPool.WorkerCount = count
		}
	}
	if batchWorkerCount := os.Getenv("BATCH_WORKER_COUNT"); batchWorkerCount != "" {
		var count int
		fmt.Sscanf(batchWorkerCount, "%d", &count)
		if count > 0 {
			config.WorkerPool.BatchWorkerCount = count
		}
	}
	if maxConcurrentLLM := os.Getenv("MAX_CONCURRENT_LLM"); maxConcurrentLLM != "" {
		var count int
		fmt.Sscanf(maxConcurrentLLM, "%d", &count)
		if count > 0 {
			config.WorkerPool.MaxConcurrentLLM = count
		}
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
