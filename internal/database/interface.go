package database

import (
	"github.com/sleepstars/mediascanner/internal/models"
)

// DatabaseInterface defines the interface for database operations
type DatabaseInterface interface {
	GetMediaFileByPath(path string) (*models.MediaFile, error)
	CreateMediaFile(file *models.MediaFile) error
	UpdateMediaFile(file *models.MediaFile) error
	GetMediaFileByID(id int64) (*models.MediaFile, error)
	CreateMediaInfo(info *models.MediaInfo) error
	GetMediaInfoByMediaFileID(mediaFileID int64) (*models.MediaInfo, error)
	UpdateMediaInfo(info *models.MediaInfo) error
	GetAPICache(provider, query string) (*models.APICache, error)
	CreateAPICache(cache *models.APICache) error
	CreateLLMRequest(request *models.LLMRequest) error
	CreateBatchProcess(batch *models.BatchProcess) error
	UpdateBatchProcess(batch *models.BatchProcess) error
	CreateBatchProcessFile(file *models.BatchProcessFile) error
	UpdateBatchProcessFile(file *models.BatchProcessFile) error
	GetBatchProcessFilesByBatchID(batchID int64) ([]models.BatchProcessFile, error)
	CreateNotification(notification *models.Notification) error
	GetPendingNotifications() ([]models.Notification, error)
	UpdateNotification(notification *models.Notification) error
	GetPendingMediaFiles() ([]models.MediaFile, error)
	GetMediaFilesByDirectory(directory string) ([]models.MediaFile, error)
	GetPendingBatchProcesses() ([]models.BatchProcess, error)
	Close() error
}
