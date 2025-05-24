package logger

import (
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/natefinch/lumberjack.v2"
)

// LogLevel represents the log level
type LogLevel string

const (
	// DebugLevel represents debug level
	DebugLevel LogLevel = "debug"
	// InfoLevel represents info level
	InfoLevel LogLevel = "info"
	// WarnLevel represents warn level
	WarnLevel LogLevel = "warn"
	// ErrorLevel represents error level
	ErrorLevel LogLevel = "error"
	// FatalLevel represents fatal level
	FatalLevel LogLevel = "fatal"
)

// Config represents the logger configuration
type Config struct {
	// Level is the log level (debug, info, warn, error, fatal)
	Level LogLevel `json:"level" yaml:"level"`
	
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

// Init initializes the logger
func Init(cfg *Config) {
	// Set default values if not provided
	if cfg.Level == "" {
		cfg.Level = InfoLevel
	}
	if cfg.Format == "" {
		cfg.Format = "json"
	}
	if cfg.Output == "" {
		cfg.Output = "stdout"
	}
	if cfg.MaxSize <= 0 {
		cfg.MaxSize = 100 // 100 MB
	}
	if cfg.MaxBackups <= 0 {
		cfg.MaxBackups = 3
	}
	if cfg.MaxAge <= 0 {
		cfg.MaxAge = 28 // 28 days
	}

	// Set log level
	var level zerolog.Level
	switch cfg.Level {
	case DebugLevel:
		level = zerolog.DebugLevel
	case InfoLevel:
		level = zerolog.InfoLevel
	case WarnLevel:
		level = zerolog.WarnLevel
	case ErrorLevel:
		level = zerolog.ErrorLevel
	case FatalLevel:
		level = zerolog.FatalLevel
	default:
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	// Set time format
	zerolog.TimeFieldFormat = time.RFC3339

	// Set output
	var output io.Writer
	switch cfg.Output {
	case "stdout":
		output = os.Stdout
	case "stderr":
		output = os.Stderr
	case "file":
		if cfg.File == "" {
			// Default to logs directory in current directory
			cfg.File = filepath.Join("logs", "mediascanner.log")
		}
		
		// Create logs directory if it doesn't exist
		logsDir := filepath.Dir(cfg.File)
		if err := os.MkdirAll(logsDir, 0755); err != nil {
			log.Error().Err(err).Str("path", logsDir).Msg("Failed to create logs directory")
			output = os.Stdout // Fallback to stdout
		} else {
			// Use lumberjack for log rotation
			output = &lumberjack.Logger{
				Filename:   cfg.File,
				MaxSize:    cfg.MaxSize,
				MaxBackups: cfg.MaxBackups,
				MaxAge:     cfg.MaxAge,
				Compress:   cfg.Compress,
			}
		}
	default:
		output = os.Stdout
	}

	// Set format
	if cfg.Format == "console" {
		output = zerolog.ConsoleWriter{
			Out:        output,
			TimeFormat: time.RFC3339,
		}
	}

	// Set global logger
	log.Logger = zerolog.New(output).With().Timestamp().Logger()
}

// Debug logs a debug message
func Debug() *zerolog.Event {
	return log.Debug()
}

// Info logs an info message
func Info() *zerolog.Event {
	return log.Info()
}

// Warn logs a warning message
func Warn() *zerolog.Event {
	return log.Warn()
}

// Error logs an error message
func Error() *zerolog.Event {
	return log.Error()
}

// Fatal logs a fatal message and exits
func Fatal() *zerolog.Event {
	return log.Fatal()
}

// With returns a context logger with the given key-value pair
func With() zerolog.Context {
	return log.With()
}

// Logger returns the global logger
func Logger() zerolog.Logger {
	return log.Logger
}
