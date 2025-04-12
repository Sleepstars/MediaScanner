package processor

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/sleepstars/mediascanner/internal/api"
	"github.com/sleepstars/mediascanner/internal/config"
	"github.com/sleepstars/mediascanner/internal/database"
	"github.com/sleepstars/mediascanner/internal/fileops"
	"github.com/sleepstars/mediascanner/internal/llm"
	"github.com/sleepstars/mediascanner/internal/models"
	"github.com/sleepstars/mediascanner/internal/notification"
	"github.com/sleepstars/mediascanner/internal/worker"
)

// Processor represents the media processor
type Processor struct {
	config     *config.Config
	db         *database.Database
	llmClient  *llm.LLM
	apiClient  *api.API
	fileOps    *fileops.FileOps
	notifier   *notification.Notifier
	workerPool worker.WorkerPool
}

// New creates a new processor
func New(cfg *config.Config, db *database.Database, llmClient *llm.LLM, apiClient *api.API, fileOps *fileops.FileOps, notifier *notification.Notifier) *Processor {
	p := &Processor{
		config:    cfg,
		db:        db,
		llmClient: llmClient,
		apiClient: apiClient,
		fileOps:   fileOps,
		notifier:  notifier,
	}

	// Create worker pool if enabled
	if cfg.WorkerPool.Enabled {
		// Create a worker pool with the processor as the task processor
		pool := worker.New(&cfg.WorkerPool, p.processTask)
		p.workerPool = pool
	}

	return p
}

// ProcessMediaFile processes a media file
func (p *Processor) ProcessMediaFile(ctx context.Context, mediaFile *models.MediaFile) error {
	log.Printf("Processing media file: %s", mediaFile.OriginalPath)

	// Update status to processing
	mediaFile.Status = "processing"
	mediaFile.UpdatedAt = time.Now()
	if err := p.db.UpdateMediaFile(mediaFile); err != nil {
		return fmt.Errorf("error updating media file status: %w", err)
	}

	// Register API function handlers
	p.registerFunctionHandlers()

	// Process the file with LLM
	result, err := p.llmClient.ProcessMediaFile(ctx, mediaFile.OriginalName, p.config.FileOps.DirectoryStructure)
	if err != nil {
		// Update status to failed
		mediaFile.Status = "failed"
		mediaFile.ErrorMessage = fmt.Sprintf("LLM processing error: %v", err)
		mediaFile.UpdatedAt = time.Now()
		_ = p.db.UpdateMediaFile(mediaFile)

		// Create notification
		_ = p.createErrorNotification(mediaFile, fmt.Sprintf("LLM processing error: %v", err))

		return fmt.Errorf("error processing media file with LLM: %w", err)
	}

	// Create media info record
	mediaInfo := &models.MediaInfo{
		MediaFileID:   mediaFile.ID,
		Title:         result.Title,
		OriginalTitle: result.OriginalTitle,
		Year:          result.Year,
		MediaType:     result.MediaType,
		Season:        result.Season,
		Episode:       result.Episode,
		EpisodeTitle:  result.EpisodeTitle,
		TMDBID:        result.TMDBID,
		TVDBID:        result.TVDBID,
		BangumiID:     result.BangumiID,
		ImdbID:        result.ImdbID,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := p.db.CreateMediaInfo(mediaInfo); err != nil {
		// Update status to failed
		mediaFile.Status = "failed"
		mediaFile.ErrorMessage = fmt.Sprintf("Error creating media info record: %v", err)
		mediaFile.UpdatedAt = time.Now()
		_ = p.db.UpdateMediaFile(mediaFile)

		// Create notification
		_ = p.createErrorNotification(mediaFile, fmt.Sprintf("Error creating media info record: %v", err))

		return fmt.Errorf("error creating media info record: %w", err)
	}

	// Fetch additional metadata
	if err := p.fetchAdditionalMetadata(ctx, mediaInfo); err != nil {
		log.Printf("Warning: Error fetching additional metadata for %s: %v", mediaFile.OriginalPath, err)
	}

	// Generate destination path
	destPath, err := p.generateDestinationPath(result)
	if err != nil {
		// Update status to failed
		mediaFile.Status = "failed"
		mediaFile.ErrorMessage = fmt.Sprintf("Error generating destination path: %v", err)
		mediaFile.UpdatedAt = time.Now()
		_ = p.db.UpdateMediaFile(mediaFile)

		// Create notification
		_ = p.createErrorNotification(mediaFile, fmt.Sprintf("Error generating destination path: %v", err))

		return fmt.Errorf("error generating destination path: %w", err)
	}

	// Process the file
	destFilePath, err := p.fileOps.ProcessFile(mediaFile.OriginalPath, destPath)
	if err != nil {
		// Update status to failed
		mediaFile.Status = "failed"
		mediaFile.ErrorMessage = fmt.Sprintf("Error processing file: %v", err)
		mediaFile.UpdatedAt = time.Now()
		_ = p.db.UpdateMediaFile(mediaFile)

		// Create notification
		_ = p.createErrorNotification(mediaFile, fmt.Sprintf("Error processing file: %v", err))

		return fmt.Errorf("error processing file: %w", err)
	}

	// Update media file record
	mediaFile.DestinationPath = destFilePath
	mediaFile.Status = "success"
	mediaFile.ProcessedAt = time.Now()
	mediaFile.UpdatedAt = time.Now()
	if err := p.db.UpdateMediaFile(mediaFile); err != nil {
		return fmt.Errorf("error updating media file record: %w", err)
	}

	// Create NFO files and download images
	if err := p.createMetadataFiles(ctx, mediaFile, mediaInfo, result); err != nil {
		log.Printf("Warning: Error creating metadata files for %s: %v", mediaFile.OriginalPath, err)
	}

	// Create success notification
	_ = p.createSuccessNotification(mediaFile, mediaInfo)

	log.Printf("Successfully processed media file: %s -> %s", mediaFile.OriginalPath, mediaFile.DestinationPath)
	return nil
}

// ProcessBatchFiles processes a batch of media files
func (p *Processor) ProcessBatchFiles(ctx context.Context, batchProcess *models.BatchProcess) error {
	log.Printf("Processing batch: %s (%d files)", batchProcess.Directory, batchProcess.FileCount)

	// Update status to processing
	batchProcess.Status = "processing"
	batchProcess.UpdatedAt = time.Now()
	if err := p.db.UpdateBatchProcess(batchProcess); err != nil {
		return fmt.Errorf("error updating batch process status: %w", err)
	}

	// Get batch process files
	batchFiles, err := p.db.GetBatchProcessFilesByBatchID(batchProcess.ID)
	if err != nil {
		batchProcess.Status = "failed"
		batchProcess.UpdatedAt = time.Now()
		_ = p.db.UpdateBatchProcess(batchProcess)
		return fmt.Errorf("error getting batch process files: %w", err)
	}

	// Get media files
	mediaFiles := make([]*models.MediaFile, 0, len(batchFiles))
	filenames := make([]string, 0, len(batchFiles))
	for _, batchFile := range batchFiles {
		mediaFile, err := p.db.GetMediaFileByID(batchFile.MediaFileID)
		if err != nil {
			log.Printf("Error getting media file %d: %v", batchFile.MediaFileID, err)
			continue
		}
		mediaFiles = append(mediaFiles, mediaFile)
		filenames = append(filenames, mediaFile.OriginalName)
	}

	// Register API function handlers
	p.registerFunctionHandlers()

	// Process the files with LLM
	results, err := p.llmClient.ProcessBatchFiles(ctx, filenames, p.config.FileOps.DirectoryStructure)
	if err != nil {
		batchProcess.Status = "failed"
		batchProcess.UpdatedAt = time.Now()
		_ = p.db.UpdateBatchProcess(batchProcess)
		return fmt.Errorf("error processing batch files with LLM: %w", err)
	}

	// Create a map of results by filename
	resultMap := make(map[string]*llm.MediaFileResult)
	for _, result := range results {
		resultMap[result.OriginalFilename] = result
	}

	// Process each file
	for i, mediaFile := range mediaFiles {
		batchFile := batchFiles[i]

		// Update status to processing
		mediaFile.Status = "processing"
		mediaFile.UpdatedAt = time.Now()
		if err := p.db.UpdateMediaFile(mediaFile); err != nil {
			log.Printf("Error updating media file status: %v", err)
			continue
		}

		// Get result for this file
		result, ok := resultMap[mediaFile.OriginalName]
		if !ok {
			// Update status to failed
			mediaFile.Status = "failed"
			mediaFile.ErrorMessage = "No result found for this file in batch processing"
			mediaFile.UpdatedAt = time.Now()
			_ = p.db.UpdateMediaFile(mediaFile)

			// Update batch file status
			batchFile.Status = "failed"
			batchFile.UpdatedAt = time.Now()
			_ = p.db.UpdateBatchProcessFile(&batchFile)

			// Create notification
			_ = p.createErrorNotification(mediaFile, "No result found for this file in batch processing")
			continue
		}

		// Create media info record
		mediaInfo := &models.MediaInfo{
			MediaFileID:   mediaFile.ID,
			Title:         result.Title,
			OriginalTitle: result.OriginalTitle,
			Year:          result.Year,
			MediaType:     result.MediaType,
			Season:        result.Season,
			Episode:       result.Episode,
			EpisodeTitle:  result.EpisodeTitle,
			TMDBID:        result.TMDBID,
			TVDBID:        result.TVDBID,
			BangumiID:     result.BangumiID,
			ImdbID:        result.ImdbID,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}

		if err := p.db.CreateMediaInfo(mediaInfo); err != nil {
			// Update status to failed
			mediaFile.Status = "failed"
			mediaFile.ErrorMessage = fmt.Sprintf("Error creating media info record: %v", err)
			mediaFile.UpdatedAt = time.Now()
			_ = p.db.UpdateMediaFile(mediaFile)

			// Update batch file status
			batchFile.Status = "failed"
			batchFile.UpdatedAt = time.Now()
			_ = p.db.UpdateBatchProcessFile(&batchFile)

			// Create notification
			_ = p.createErrorNotification(mediaFile, fmt.Sprintf("Error creating media info record: %v", err))
			continue
		}

		// Fetch additional metadata
		if err := p.fetchAdditionalMetadata(ctx, mediaInfo); err != nil {
			log.Printf("Warning: Error fetching additional metadata for %s: %v", mediaFile.OriginalPath, err)
		}

		// Generate destination path
		destPath, err := p.generateDestinationPath(result)
		if err != nil {
			// Update status to failed
			mediaFile.Status = "failed"
			mediaFile.ErrorMessage = fmt.Sprintf("Error generating destination path: %v", err)
			mediaFile.UpdatedAt = time.Now()
			_ = p.db.UpdateMediaFile(mediaFile)

			// Update batch file status
			batchFile.Status = "failed"
			batchFile.UpdatedAt = time.Now()
			_ = p.db.UpdateBatchProcessFile(&batchFile)

			// Create notification
			_ = p.createErrorNotification(mediaFile, fmt.Sprintf("Error generating destination path: %v", err))
			continue
		}

		// Process the file
		destFilePath, err := p.fileOps.ProcessFile(mediaFile.OriginalPath, destPath)
		if err != nil {
			// Update status to failed
			mediaFile.Status = "failed"
			mediaFile.ErrorMessage = fmt.Sprintf("Error processing file: %v", err)
			mediaFile.UpdatedAt = time.Now()
			_ = p.db.UpdateMediaFile(mediaFile)

			// Update batch file status
			batchFile.Status = "failed"
			batchFile.UpdatedAt = time.Now()
			_ = p.db.UpdateBatchProcessFile(&batchFile)

			// Create notification
			_ = p.createErrorNotification(mediaFile, fmt.Sprintf("Error processing file: %v", err))
			continue
		}

		// Update media file record
		mediaFile.DestinationPath = destFilePath
		mediaFile.Status = "success"
		mediaFile.ProcessedAt = time.Now()
		mediaFile.UpdatedAt = time.Now()
		if err := p.db.UpdateMediaFile(mediaFile); err != nil {
			log.Printf("Error updating media file record: %v", err)
		}

		// Update batch file status
		batchFile.Status = "success"
		batchFile.UpdatedAt = time.Now()
		if err := p.db.UpdateBatchProcessFile(&batchFile); err != nil {
			log.Printf("Error updating batch file status: %v", err)
		}

		// Create NFO files and download images
		if err := p.createMetadataFiles(ctx, mediaFile, mediaInfo, result); err != nil {
			log.Printf("Warning: Error creating metadata files for %s: %v", mediaFile.OriginalPath, err)
		}

		// Create success notification
		_ = p.createSuccessNotification(mediaFile, mediaInfo)

		log.Printf("Successfully processed media file: %s -> %s", mediaFile.OriginalPath, mediaFile.DestinationPath)
	}

	// Update batch process status
	batchProcess.Status = "completed"
	batchProcess.CompletedAt = time.Now()
	batchProcess.UpdatedAt = time.Now()
	if err := p.db.UpdateBatchProcess(batchProcess); err != nil {
		return fmt.Errorf("error updating batch process status: %w", err)
	}

	log.Printf("Successfully processed batch: %s", batchProcess.Directory)
	return nil
}

// registerFunctionHandlers registers the function handlers for the LLM
func (p *Processor) registerFunctionHandlers() {
	// Register TMDB search function
	p.llmClient.RegisterFunction("searchTMDB", func(ctx context.Context, args json.RawMessage) (interface{}, error) {
		var params struct {
			Query     string `json:"query"`
			Year      int    `json:"year,omitempty"`
			MediaType string `json:"mediaType,omitempty"`
		}
		if err := json.Unmarshal(args, &params); err != nil {
			return nil, fmt.Errorf("error parsing arguments: %w", err)
		}

		if params.MediaType == "movie" || params.MediaType == "" {
			result, err := p.apiClient.TMDB.SearchMovie(ctx, params.Query, params.Year)
			if err != nil {
				return nil, fmt.Errorf("error searching TMDB for movie: %w", err)
			}
			return result, nil
		} else if params.MediaType == "tv" {
			result, err := p.apiClient.TMDB.SearchTV(ctx, params.Query, params.Year)
			if err != nil {
				return nil, fmt.Errorf("error searching TMDB for TV show: %w", err)
			}
			return result, nil
		}

		return nil, fmt.Errorf("invalid media type: %s", params.MediaType)
	})

	// Register TVDB search function
	p.llmClient.RegisterFunction("searchTVDB", func(ctx context.Context, args json.RawMessage) (interface{}, error) {
		var params struct {
			Query string `json:"query"`
		}
		if err := json.Unmarshal(args, &params); err != nil {
			return nil, fmt.Errorf("error parsing arguments: %w", err)
		}

		result, err := p.apiClient.TVDB.SearchSeries(ctx, params.Query)
		if err != nil {
			return nil, fmt.Errorf("error searching TVDB: %w", err)
		}
		return result, nil
	})

	// Register Bangumi search function
	p.llmClient.RegisterFunction("searchBangumi", func(ctx context.Context, args json.RawMessage) (interface{}, error) {
		var params struct {
			Query string `json:"query"`
		}
		if err := json.Unmarshal(args, &params); err != nil {
			return nil, fmt.Errorf("error parsing arguments: %w", err)
		}

		result, err := p.apiClient.Bangumi.SearchAnime(ctx, params.Query)
		if err != nil {
			return nil, fmt.Errorf("error searching Bangumi: %w", err)
		}
		return result, nil
	})
}

// fetchAdditionalMetadata fetches additional metadata for a media file
func (p *Processor) fetchAdditionalMetadata(ctx context.Context, mediaInfo *models.MediaInfo) error {
	if mediaInfo.MediaType == "movie" && mediaInfo.TMDBID > 0 {
		// Fetch movie details from TMDB
		movie, err := p.apiClient.TMDB.GetMovieDetails(ctx, int(mediaInfo.TMDBID))
		if err != nil {
			return fmt.Errorf("error fetching movie details from TMDB: %w", err)
		}

		// Update media info
		mediaInfo.Overview = movie.Overview
		mediaInfo.PosterPath = movie.PosterPath
		mediaInfo.BackdropPath = movie.BackdropPath
		mediaInfo.Genres = strings.Join(movie.Genres, ",")
		mediaInfo.Countries = strings.Join(movie.Countries, ",")
		mediaInfo.Languages = strings.Join(movie.Languages, ",")
		mediaInfo.ImdbID = movie.ImdbID

		// Save to database
		if err := p.db.UpdateMediaInfo(mediaInfo); err != nil {
			return fmt.Errorf("error updating media info: %w", err)
		}
	} else if mediaInfo.MediaType == "tv" {
		if mediaInfo.TMDBID > 0 {
			// Fetch TV show details from TMDB
			tv, err := p.apiClient.TMDB.GetTVDetails(ctx, int(mediaInfo.TMDBID))
			if err != nil {
				return fmt.Errorf("error fetching TV show details from TMDB: %w", err)
			}

			// Update media info
			mediaInfo.Overview = tv.Overview
			mediaInfo.PosterPath = tv.PosterPath
			mediaInfo.BackdropPath = tv.BackdropPath
			mediaInfo.Genres = strings.Join(tv.Genres, ",")
			mediaInfo.Countries = strings.Join(tv.Countries, ",")
			mediaInfo.Languages = strings.Join(tv.Languages, ",")
			mediaInfo.ImdbID = tv.ImdbID
			mediaInfo.TVDBID = int64(tv.TVDBID)

			// Save to database
			if err := p.db.UpdateMediaInfo(mediaInfo); err != nil {
				return fmt.Errorf("error updating media info: %w", err)
			}

			// Fetch season details if available
			if mediaInfo.Season > 0 {
				season, err := p.apiClient.TMDB.GetSeasonDetails(ctx, int(mediaInfo.TMDBID), mediaInfo.Season)
				if err != nil {
					log.Printf("Warning: Error fetching season details from TMDB: %v", err)
				} else {
					// Find episode
					for _, episode := range season.Episodes {
						if episode.EpisodeNumber == mediaInfo.Episode {
							mediaInfo.EpisodeTitle = episode.Name
							// Save to database
							if err := p.db.UpdateMediaInfo(mediaInfo); err != nil {
								log.Printf("Warning: Error updating media info with episode title: %v", err)
							}
							break
						}
					}
				}
			}
		} else if mediaInfo.TVDBID > 0 {
			// Fetch TV show details from TVDB
			tv, err := p.apiClient.TVDB.GetSeriesDetails(ctx, int(mediaInfo.TVDBID))
			if err != nil {
				return fmt.Errorf("error fetching TV show details from TVDB: %w", err)
			}

			// Update media info
			mediaInfo.Overview = tv.Overview
			mediaInfo.Genres = strings.Join(tv.Genres, ",")
			mediaInfo.Countries = strings.Join(tv.Countries, ",")
			mediaInfo.Languages = strings.Join(tv.Languages, ",")
			mediaInfo.ImdbID = tv.ImdbID

			// Save to database
			if err := p.db.UpdateMediaInfo(mediaInfo); err != nil {
				return fmt.Errorf("error updating media info: %w", err)
			}

			// Find season
			var seasonID int
			for _, season := range tv.Seasons {
				if season.Number == mediaInfo.Season {
					seasonID = season.ID
					break
				}
			}

			// Fetch season episodes if available
			if seasonID > 0 {
				seasonEpisodes, err := p.apiClient.TVDB.GetSeasonEpisodes(ctx, seasonID)
				if err != nil {
					log.Printf("Warning: Error fetching season episodes from TVDB: %v", err)
				} else {
					// Find episode
					for _, episode := range seasonEpisodes.Episodes {
						if episode.EpisodeNumber == mediaInfo.Episode && episode.SeasonNumber == mediaInfo.Season {
							mediaInfo.EpisodeTitle = episode.Name
							// Save to database
							if err := p.db.UpdateMediaInfo(mediaInfo); err != nil {
								log.Printf("Warning: Error updating media info with episode title: %v", err)
							}
							break
						}
					}
				}
			}
		} else if mediaInfo.BangumiID > 0 {
			// Fetch anime details from Bangumi
			anime, err := p.apiClient.Bangumi.GetAnimeDetails(ctx, int(mediaInfo.BangumiID))
			if err != nil {
				return fmt.Errorf("error fetching anime details from Bangumi: %w", err)
			}

			// Update media info
			mediaInfo.Overview = anime.Summary
			mediaInfo.Genres = strings.Join(anime.Tags, ",")

			// Save to database
			if err := p.db.UpdateMediaInfo(mediaInfo); err != nil {
				return fmt.Errorf("error updating media info: %w", err)
			}

			// Find episode
			if mediaInfo.Episode > 0 && len(anime.Episodes) > 0 {
				for _, episode := range anime.Episodes {
					if episode.Sort == mediaInfo.Episode {
						if episode.NameCN != "" {
							mediaInfo.EpisodeTitle = episode.NameCN
						} else {
							mediaInfo.EpisodeTitle = episode.Name
						}
						// Save to database
						if err := p.db.UpdateMediaInfo(mediaInfo); err != nil {
							log.Printf("Warning: Error updating media info with episode title: %v", err)
						}
						break
					}
				}
			}
		}
	}

	return nil
}

// generateDestinationPath generates the destination path for a media file
func (p *Processor) generateDestinationPath(result *llm.MediaFileResult) (string, error) {
	// Get destination root
	destRoot := p.config.FileOps.DestinationRoot
	if destRoot == "" {
		return "", fmt.Errorf("destination root is not configured")
	}

	// Build path based on media type
	var destPath string
	if result.MediaType == "movie" {
		// Movie path: /DestinationRoot/Category/Subcategory/Title (Year)/Title (Year).ext
		destPath = filepath.Join(destRoot, result.Category, result.Subcategory, fmt.Sprintf("%s (%d)", result.Title, result.Year))
	} else if result.MediaType == "tv" {
		// TV show path: /DestinationRoot/Category/Subcategory/Title (Year)/Season X/Title - SXXEXX - Episode Title.ext
		showDir := filepath.Join(destRoot, result.Category, result.Subcategory, fmt.Sprintf("%s (%d)", result.Title, result.Year))
		seasonDir := filepath.Join(showDir, fmt.Sprintf("Season %d", result.Season))
		destPath = seasonDir
	} else {
		return "", fmt.Errorf("unknown media type: %s", result.MediaType)
	}

	return destPath, nil
}

// createMetadataFiles creates NFO files and downloads images for a media file
func (p *Processor) createMetadataFiles(ctx context.Context, mediaFile *models.MediaFile, mediaInfo *models.MediaInfo, result *llm.MediaFileResult) error {
	// TODO: Implement NFO file generation and image downloading
	return nil
}

// createSuccessNotification creates a success notification
func (p *Processor) createSuccessNotification(mediaFile *models.MediaFile, mediaInfo *models.MediaInfo) error {
	if !p.config.Notification.Enabled {
		return nil
	}

	// Create notification message
	var message string
	if mediaInfo.MediaType == "movie" {
		message = fmt.Sprintf("Successfully processed movie: %s (%d)", mediaInfo.Title, mediaInfo.Year)
	} else if mediaInfo.MediaType == "tv" {
		message = fmt.Sprintf("Successfully processed TV show: %s (%d) - S%02dE%02d - %s",
			mediaInfo.Title, mediaInfo.Year, mediaInfo.Season, mediaInfo.Episode, mediaInfo.EpisodeTitle)
	} else {
		message = fmt.Sprintf("Successfully processed media: %s", mediaInfo.Title)
	}

	// Create notification record
	notification := &models.Notification{
		MediaFileID: mediaFile.ID,
		Type:        "success",
		Message:     message,
		Sent:        false,
		CreatedAt:   time.Now(),
	}

	// Save to database
	return p.db.CreateNotification(notification)
}

// createErrorNotification creates an error notification
func (p *Processor) createErrorNotification(mediaFile *models.MediaFile, errorMessage string) error {
	if !p.config.Notification.Enabled {
		return nil
	}

	// Create notification message
	message := fmt.Sprintf("Error processing file: %s\nError: %s", mediaFile.OriginalName, errorMessage)

	// Create notification record
	notification := &models.Notification{
		MediaFileID: mediaFile.ID,
		Type:        "error",
		Message:     message,
		Sent:        false,
		CreatedAt:   time.Now(),
	}

	// Save to database
	return p.db.CreateNotification(notification)
}
