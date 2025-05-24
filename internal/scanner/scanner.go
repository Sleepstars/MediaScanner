package scanner

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/sleepstars/mediascanner/internal/config"
	"github.com/sleepstars/mediascanner/internal/database"
	"github.com/sleepstars/mediascanner/internal/fsnotify"
	"github.com/sleepstars/mediascanner/internal/models"
)

// Scanner represents the media scanner
type Scanner struct {
	config   *config.ScannerConfig
	db       *database.Database
	watcher  *fsnotify.Watcher
	watching bool
	mu       sync.RWMutex
}

// New creates a new scanner
func New(cfg *config.ScannerConfig, db *database.Database) *Scanner {
	return &Scanner{
		config:   cfg,
		db:       db,
		watching: false,
	}
}

// ScanResult represents the result of a scan
type ScanResult struct {
	NewFiles      []string
	BatchDirs     map[string][]string
	ExcludedFiles []string
}

// Scan scans the media directories for new files
func (s *Scanner) Scan() (*ScanResult, error) {
	result := &ScanResult{
		NewFiles:      make([]string, 0),
		BatchDirs:     make(map[string][]string),
		ExcludedFiles: make([]string, 0),
	}

	// Compile exclude patterns
	excludePatterns := make([]*regexp.Regexp, 0, len(s.config.ExcludePatterns))
	for _, pattern := range s.config.ExcludePatterns {
		re, err := regexp.Compile("(?i)" + pattern)
		if err != nil {
			log.Printf("Warning: Invalid exclude pattern %q: %v", pattern, err)
			continue
		}
		excludePatterns = append(excludePatterns, re)
	}

	// Create a map of video extensions for faster lookup
	videoExtMap := make(map[string]bool)
	for _, ext := range s.config.VideoExtensions {
		videoExtMap[strings.ToLower(ext)] = true
	}

	// Create a map of exclude directories for faster lookup
	excludeDirMap := make(map[string]bool)
	for _, dir := range s.config.ExcludeDirs {
		excludeDirMap[strings.ToLower(dir)] = true
	}

	// Scan each media directory
	for _, mediaDir := range s.config.MediaDirs {
		err := filepath.Walk(mediaDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				log.Printf("Error accessing path %q: %v", path, err)
				return nil
			}

			// Skip directories that match exclude patterns
			if info.IsDir() {
				dirName := strings.ToLower(filepath.Base(path))
				if excludeDirMap[dirName] {
					return filepath.SkipDir
				}
				return nil
			}

			// Check if the file is a video file
			ext := strings.ToLower(filepath.Ext(path))
			if !videoExtMap[ext] {
				return nil
			}

			// Check if the file matches any exclude patterns
			fileName := filepath.Base(path)
			for _, pattern := range excludePatterns {
				if pattern.MatchString(fileName) {
					result.ExcludedFiles = append(result.ExcludedFiles, path)
					return nil
				}
			}

			// Check if the file is already in the database
			_, err = s.db.GetMediaFileByPath(path)
			if err == nil {
				// File is already in the database
				return nil
			}

			// Add the file to the result
			result.NewFiles = append(result.NewFiles, path)

			// Add the file to the batch directory map
			dir := filepath.Dir(path)
			result.BatchDirs[dir] = append(result.BatchDirs[dir], path)

			return nil
		})

		if err != nil {
			log.Printf("Error scanning directory %q: %v", mediaDir, err)
		}
	}

	// Filter batch directories based on threshold
	for dir, files := range result.BatchDirs {
		if len(files) >= s.config.BatchThreshold {
			log.Printf("Directory %q has %d files, exceeding batch threshold of %d", dir, len(files), s.config.BatchThreshold)
		} else {
			delete(result.BatchDirs, dir)
		}
	}

	return result, nil
}

// StartWatching starts watching the media directories for changes
func (s *Scanner) StartWatching() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.watching {
		return fmt.Errorf("already watching")
	}

	// Create a new watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create watcher: %w", err)
	}
	s.watcher = watcher

	// Add handler for file events
	s.watcher.AddHandler(s.handleFileEvent)

	// Add media directories to the watcher
	for _, mediaDir := range s.config.MediaDirs {
		if err := s.watcher.AddDirectory(mediaDir); err != nil {
			log.Error().Err(err).Str("directory", mediaDir).Msg("Failed to add directory to watcher")
		} else {
			log.Debug().Str("directory", mediaDir).Msg("Directory added to watcher")
		}
	}

	s.watching = true
	log.Info().Int("directory_count", len(s.config.MediaDirs)).Msg("File system watcher started")
	return nil
}

// StopWatching stops watching the media directories for changes
func (s *Scanner) StopWatching() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.watching {
		return fmt.Errorf("not watching")
	}

	if err := s.watcher.Close(); err != nil {
		return fmt.Errorf("failed to close watcher: %w", err)
	}

	s.watching = false
	s.watcher = nil
	log.Printf("Stopped watching media directories")
	return nil
}

// IsWatching returns whether the scanner is watching for changes
func (s *Scanner) IsWatching() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.watching
}

// handleFileEvent handles a file system event
func (s *Scanner) handleFileEvent(event fsnotify.Event) {
	// Only process create and modify events
	if event.Type != fsnotify.EventCreate && event.Type != fsnotify.EventModify {
		return
	}

	// Check if the file is a directory
	info, err := os.Stat(event.Path)
	if err != nil {
		log.Printf("Error getting file info for %q: %v", event.Path, err)
		return
	}
	if info.IsDir() {
		return
	}

	// Check if the file is a video file
	ext := strings.ToLower(filepath.Ext(event.Path))
	isVideoFile := false
	for _, videoExt := range s.config.VideoExtensions {
		if ext == strings.ToLower(videoExt) {
			isVideoFile = true
			break
		}
	}
	if !isVideoFile {
		return
	}

	// Check if the file matches any exclude patterns
	fileName := filepath.Base(event.Path)
	for _, pattern := range s.config.ExcludePatterns {
		re, err := regexp.Compile("(?i)" + pattern)
		if err != nil {
			continue
		}
		if re.MatchString(fileName) {
			return
		}
	}

	// Check if the file is already in the database
	_, err = s.db.GetMediaFileByPath(event.Path)
	if err == nil {
		// File is already in the database
		return
	}

	// Create a media file record
	_, err = s.CreateMediaFile(event.Path)
	if err != nil {
		log.Error().Err(err).Str("file", event.Path).Msg("Failed to create media file record")
		return
	}

	log.Info().Str("file", event.Path).Msg("New file detected and added")
}

// CreateMediaFile creates a new media file record in the database
func (s *Scanner) CreateMediaFile(path string) (*models.MediaFile, error) {
	// Get file info
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("error getting file info: %w", err)
	}

	// Create media file record
	mediaFile := &models.MediaFile{
		OriginalPath: path,
		OriginalName: filepath.Base(path),
		FileSize:     info.Size(),
		Status:       "pending",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Save to database
	err = s.db.CreateMediaFile(mediaFile)
	if err != nil {
		return nil, fmt.Errorf("error creating media file record: %w", err)
	}

	return mediaFile, nil
}

// CreateBatchProcess creates a new batch process record in the database
func (s *Scanner) CreateBatchProcess(dir string, files []string) (*models.BatchProcess, error) {
	// Create batch process record
	batchProcess := &models.BatchProcess{
		Directory: dir,
		FileCount: len(files),
		Status:    "pending",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Save to database
	err := s.db.CreateBatchProcess(batchProcess)
	if err != nil {
		return nil, fmt.Errorf("error creating batch process record: %w", err)
	}

	// Create media file records for each file
	for _, file := range files {
		mediaFile, err := s.CreateMediaFile(file)
		if err != nil {
			log.Printf("Error creating media file record for %q: %v", file, err)
			continue
		}

		// Create batch process file record
		batchProcessFile := &models.BatchProcessFile{
			BatchProcessID: batchProcess.ID,
			MediaFileID:    mediaFile.ID,
			Status:         "pending",
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}

		// Save to database
		err = s.db.CreateBatchProcessFile(batchProcessFile)
		if err != nil {
			log.Printf("Error creating batch process file record for %q: %v", file, err)
		}
	}

	return batchProcess, nil
}
