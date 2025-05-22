package database

import (
	"fmt"
	"time"

	"github.com/sleepstars/mediascanner/internal/config"
	"github.com/sleepstars/mediascanner/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// GormDBInterface defines the interface for GORM DB operations
type GormDBInterface interface {
	Create(value interface{}) *gorm.DB
	Save(value interface{}) *gorm.DB
	First(dest interface{}, conds ...interface{}) *gorm.DB
	Find(dest interface{}, conds ...interface{}) *gorm.DB
	Where(query interface{}, args ...interface{}) *gorm.DB
	AutoMigrate(dst ...interface{}) error
	GetDB() *gorm.DB
}

// GormDBWrapper wraps a *gorm.DB to implement GormDBInterface
type GormDBWrapper struct {
	db *gorm.DB
}

func (w *GormDBWrapper) Create(value interface{}) *gorm.DB {
	return w.db.Create(value)
}

func (w *GormDBWrapper) Save(value interface{}) *gorm.DB {
	return w.db.Save(value)
}

func (w *GormDBWrapper) First(dest interface{}, conds ...interface{}) *gorm.DB {
	return w.db.First(dest, conds...)
}

func (w *GormDBWrapper) Find(dest interface{}, conds ...interface{}) *gorm.DB {
	return w.db.Find(dest, conds...)
}

func (w *GormDBWrapper) Where(query interface{}, args ...interface{}) *gorm.DB {
	return w.db.Where(query, args...)
}

func (w *GormDBWrapper) AutoMigrate(dst ...interface{}) error {
	return w.db.AutoMigrate(dst...)
}

func (w *GormDBWrapper) GetDB() *gorm.DB {
	return w.db
}

// Database represents the database connection
type Database struct {
	db GormDBInterface
}

// New creates a new database connection
func New(cfg *config.DatabaseConfig) (*Database, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Password,
		cfg.Database,
		cfg.SSLMode,
	)

	// Configure GORM logger
	gormLogger := logger.Default

	// Connect to the database
	gormDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	sqlDB, err := gormDB.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database connection: %w", err)
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Wrap the gorm.DB in our wrapper
	db := &GormDBWrapper{db: gormDB}

	return &Database{db: db}, nil
}

// Migrate performs database migrations
func (d *Database) Migrate() error {
	return d.db.AutoMigrate(
		&models.MediaFile{},
		&models.MediaInfo{},
		&models.APICache{},
		&models.LLMRequest{},
		&models.BatchProcess{},
		&models.BatchProcessFile{},
		&models.Notification{},
	)
}

// GetDB returns the GORM database instance
func (d *Database) GetDB() *gorm.DB {
	return d.db.GetDB()
}

// Close closes the database connection
func (d *Database) Close() error {
	gormDB := d.db.GetDB()
	sqlDB, err := gormDB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// CreateMediaFile creates a new media file record
func (d *Database) CreateMediaFile(file *models.MediaFile) error {
	return d.db.Create(file).Error
}

// GetMediaFileByPath retrieves a media file by its original path
func (d *Database) GetMediaFileByPath(path string) (*models.MediaFile, error) {
	var file models.MediaFile
	err := d.db.Where("original_path = ?", path).First(&file).Error
	if err != nil {
		return nil, err
	}
	return &file, nil
}

// GetMediaFileByID retrieves a media file by its ID
func (d *Database) GetMediaFileByID(id int64) (*models.MediaFile, error) {
	var file models.MediaFile
	err := d.db.Where("id = ?", id).First(&file).Error
	if err != nil {
		return nil, err
	}
	return &file, nil
}

// UpdateMediaFile updates a media file record
func (d *Database) UpdateMediaFile(file *models.MediaFile) error {
	return d.db.Save(file).Error
}

// CreateMediaInfo creates a new media info record
func (d *Database) CreateMediaInfo(info *models.MediaInfo) error {
	return d.db.Create(info).Error
}

// GetMediaInfoByMediaFileID retrieves media info by media file ID
func (d *Database) GetMediaInfoByMediaFileID(mediaFileID int64) (*models.MediaInfo, error) {
	var info models.MediaInfo
	err := d.db.Where("media_file_id = ?", mediaFileID).First(&info).Error
	if err != nil {
		return nil, err
	}
	return &info, nil
}

// UpdateMediaInfo updates a media info record
func (d *Database) UpdateMediaInfo(info *models.MediaInfo) error {
	return d.db.Save(info).Error
}

// GetAPICache retrieves a cached API response
func (d *Database) GetAPICache(provider, query string) (*models.APICache, error) {
	var cache models.APICache
	err := d.db.Where("provider = ? AND query = ? AND expires_at > ?", provider, query, time.Now()).First(&cache).Error
	if err != nil {
		return nil, err
	}
	return &cache, nil
}

// CreateAPICache creates a new API cache record
func (d *Database) CreateAPICache(cache *models.APICache) error {
	return d.db.Create(cache).Error
}

// CreateLLMRequest creates a new LLM request record
func (d *Database) CreateLLMRequest(request *models.LLMRequest) error {
	return d.db.Create(request).Error
}

// CreateBatchProcess creates a new batch process record
func (d *Database) CreateBatchProcess(batch *models.BatchProcess) error {
	return d.db.Create(batch).Error
}

// UpdateBatchProcess updates a batch process record
func (d *Database) UpdateBatchProcess(batch *models.BatchProcess) error {
	return d.db.Save(batch).Error
}

// CreateBatchProcessFile creates a new batch process file record
func (d *Database) CreateBatchProcessFile(file *models.BatchProcessFile) error {
	return d.db.Create(file).Error
}

// UpdateBatchProcessFile updates a batch process file record
func (d *Database) UpdateBatchProcessFile(file *models.BatchProcessFile) error {
	return d.db.Save(file).Error
}

// CreateNotification creates a new notification record
func (d *Database) CreateNotification(notification *models.Notification) error {
	return d.db.Create(notification).Error
}

// GetPendingNotifications retrieves pending notifications
func (d *Database) GetPendingNotifications() ([]models.Notification, error) {
	var notifications []models.Notification
	err := d.db.Where("sent = ?", false).Find(&notifications).Error
	if err != nil {
		return nil, err
	}
	return notifications, nil
}

// UpdateNotification updates a notification record
func (d *Database) UpdateNotification(notification *models.Notification) error {
	return d.db.Save(notification).Error
}

// GetPendingMediaFiles retrieves pending media files
func (d *Database) GetPendingMediaFiles() ([]models.MediaFile, error) {
	var files []models.MediaFile
	err := d.db.Where("status = ?", "pending").Find(&files).Error
	if err != nil {
		return nil, err
	}
	return files, nil
}

// GetMediaFilesByDirectory retrieves media files by directory
func (d *Database) GetMediaFilesByDirectory(directory string) ([]models.MediaFile, error) {
	var files []models.MediaFile
	err := d.db.Where("original_path LIKE ?", directory+"%").Find(&files).Error
	if err != nil {
		return nil, err
	}
	return files, nil
}

// GetPendingBatchProcesses retrieves pending batch processes
func (d *Database) GetPendingBatchProcesses() ([]models.BatchProcess, error) {
	var batches []models.BatchProcess
	err := d.db.Where("status = ?", "pending").Find(&batches).Error
	if err != nil {
		return nil, err
	}
	return batches, nil
}

// GetBatchProcessFilesByBatchID retrieves batch process files by batch ID
func (d *Database) GetBatchProcessFilesByBatchID(batchID int64) ([]models.BatchProcessFile, error) {
	var files []models.BatchProcessFile
	err := d.db.Where("batch_process_id = ?", batchID).Find(&files).Error
	if err != nil {
		return nil, err
	}
	return files, nil
}
