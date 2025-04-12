package models

import (
	"time"
)

// MediaFile represents a media file in the database
type MediaFile struct {
	ID              int64     `json:"id" gorm:"primaryKey"`
	OriginalPath    string    `json:"original_path" gorm:"uniqueIndex;not null"`
	OriginalName    string    `json:"original_name" gorm:"not null"`
	DestinationPath string    `json:"destination_path"`
	FileSize        int64     `json:"file_size"`
	MediaType       string    `json:"media_type"` // movie, tv
	Status          string    `json:"status"`     // pending, processing, success, failed, manual
	ErrorMessage    string    `json:"error_message"`
	CreatedAt       time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt       time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	ProcessedAt     time.Time `json:"processed_at"`
}

// MediaInfo represents the media information in the database
type MediaInfo struct {
	ID            int64     `json:"id" gorm:"primaryKey"`
	MediaFileID   int64     `json:"media_file_id" gorm:"uniqueIndex;not null"`
	Title         string    `json:"title"`
	OriginalTitle string    `json:"original_title"`
	Year          int       `json:"year"`
	MediaType     string    `json:"media_type"` // movie, tv
	Season        int       `json:"season"`
	Episode       int       `json:"episode"`
	EpisodeTitle  string    `json:"episode_title"`
	Overview      string    `json:"overview"`
	TMDBID        int64     `json:"tmdb_id"`
	TVDBID        int64     `json:"tvdb_id"`
	BangumiID     int64     `json:"bangumi_id"`
	ImdbID        string    `json:"imdb_id"`
	Genres        string    `json:"genres"`
	Countries     string    `json:"countries"`
	Languages     string    `json:"languages"`
	PosterPath    string    `json:"poster_path"`
	BackdropPath  string    `json:"backdrop_path"`
	CreatedAt     time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// APICache represents a cached API response
type APICache struct {
	ID        int64     `json:"id" gorm:"primaryKey"`
	Provider  string    `json:"provider" gorm:"not null"` // tmdb, tvdb, bangumi
	Query     string    `json:"query" gorm:"not null"`
	Response  string    `json:"response" gorm:"type:text"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	ExpiresAt time.Time `json:"expires_at"`
}

// LLMRequest represents an LLM request in the database
type LLMRequest struct {
	ID           int64     `json:"id" gorm:"primaryKey"`
	MediaFileID  int64     `json:"media_file_id"`
	Prompt       string    `json:"prompt" gorm:"type:text"`
	Response     string    `json:"response" gorm:"type:text"`
	Model        string    `json:"model"`
	Tokens       int       `json:"tokens"`
	Success      bool      `json:"success"`
	ErrorMessage string    `json:"error_message"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// BatchProcess represents a batch processing job
type BatchProcess struct {
	ID          int64     `json:"id" gorm:"primaryKey"`
	Directory   string    `json:"directory" gorm:"not null"`
	FileCount   int       `json:"file_count"`
	Status      string    `json:"status"` // pending, processing, completed, failed
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	CompletedAt time.Time `json:"completed_at"`
}

// BatchProcessFile represents a file in a batch processing job
type BatchProcessFile struct {
	ID             int64     `json:"id" gorm:"primaryKey"`
	BatchProcessID int64     `json:"batch_process_id" gorm:"not null"`
	MediaFileID    int64     `json:"media_file_id" gorm:"not null"`
	Status         string    `json:"status"` // pending, processing, success, failed, manual
	CreatedAt      time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// Notification represents a notification in the database
type Notification struct {
	ID          int64     `json:"id" gorm:"primaryKey"`
	MediaFileID int64     `json:"media_file_id"`
	Type        string    `json:"type"` // success, error
	Message     string    `json:"message" gorm:"type:text"`
	Sent        bool      `json:"sent"`
	SentAt      time.Time `json:"sent_at"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
}
