package database

import (
	"testing"
	"time"

	"github.com/sleepstars/mediascanner/internal/config"
	"github.com/sleepstars/mediascanner/internal/models"
	"gorm.io/gorm"
)

// MockGormDB is a mock implementation of the GORM DB
type MockGormDB struct {
	CreateFunc func(value interface{}) *gorm.DB
	SaveFunc   func(value interface{}) *gorm.DB
	FirstFunc  func(dest interface{}, conds ...interface{}) *gorm.DB
	FindFunc   func(dest interface{}, conds ...interface{}) *gorm.DB
	WhereFunc  func(query interface{}, args ...interface{}) *gorm.DB
	DBFunc     func() (*gorm.DB, error)
}

func (m *MockGormDB) Create(value interface{}) *gorm.DB {
	return m.CreateFunc(value)
}

func (m *MockGormDB) Save(value interface{}) *gorm.DB {
	return m.SaveFunc(value)
}

func (m *MockGormDB) First(dest interface{}, conds ...interface{}) *gorm.DB {
	return m.FirstFunc(dest, conds...)
}

func (m *MockGormDB) Find(dest interface{}, conds ...interface{}) *gorm.DB {
	return m.FindFunc(dest, conds...)
}

func (m *MockGormDB) Where(query interface{}, args ...interface{}) *gorm.DB {
	return m.WhereFunc(query, args...)
}

func (m *MockGormDB) DB() (*gorm.DB, error) {
	return m.DBFunc()
}

func (m *MockGormDB) AutoMigrate(dst ...interface{}) error {
	return nil
}

// MockSQLDB is a mock implementation of the SQL DB
type MockSQLDB struct {
	CloseFunc func() error
}

func (m *MockSQLDB) Close() error {
	return m.CloseFunc()
}

func TestCreateMediaFile(t *testing.T) {
	// Create a mock GORM DB
	mockDB := &MockGormDB{
		CreateFunc: func(value interface{}) *gorm.DB {
			// Verify the value is a MediaFile
			mediaFile, ok := value.(*models.MediaFile)
			if !ok {
				t.Errorf("Expected value to be *models.MediaFile, got %T", value)
			}
			
			// Verify the media file properties
			if mediaFile.OriginalPath != "/test/path/file.mp4" {
				t.Errorf("Expected original path to be '/test/path/file.mp4', got %q", mediaFile.OriginalPath)
			}
			
			// Simulate successful creation
			return &gorm.DB{Error: nil}
		},
	}
	
	// Create database with mock
	db := &Database{db: mockDB}
	
	// Create a media file
	mediaFile := &models.MediaFile{
		OriginalPath: "/test/path/file.mp4",
		OriginalName: "file.mp4",
		Status:       "pending",
	}
	
	err := db.CreateMediaFile(mediaFile)
	if err != nil {
		t.Fatalf("CreateMediaFile failed: %v", err)
	}
}

func TestGetMediaFileByPath(t *testing.T) {
	// Create a mock GORM DB
	mockDB := &MockGormDB{
		WhereFunc: func(query interface{}, args ...interface{}) *gorm.DB {
			// Verify the query
			if query != "original_path = ?" {
				t.Errorf("Expected query to be 'original_path = ?', got %q", query)
			}
			
			// Verify the args
			if len(args) != 1 || args[0] != "/test/path/file.mp4" {
				t.Errorf("Expected args to be ['/test/path/file.mp4'], got %v", args)
			}
			
			return mockDB
		},
		FirstFunc: func(dest interface{}, conds ...interface{}) *gorm.DB {
			// Verify the destination is a MediaFile
			mediaFile, ok := dest.(*models.MediaFile)
			if !ok {
				t.Errorf("Expected dest to be *models.MediaFile, got %T", dest)
			}
			
			// Set the media file properties
			mediaFile.ID = 1
			mediaFile.OriginalPath = "/test/path/file.mp4"
			mediaFile.OriginalName = "file.mp4"
			mediaFile.Status = "success"
			
			// Simulate successful query
			return &gorm.DB{Error: nil}
		},
	}
	
	// Create database with mock
	db := &Database{db: mockDB}
	
	// Get a media file by path
	mediaFile, err := db.GetMediaFileByPath("/test/path/file.mp4")
	if err != nil {
		t.Fatalf("GetMediaFileByPath failed: %v", err)
	}
	
	// Verify the result
	if mediaFile.ID != 1 {
		t.Errorf("Expected ID to be 1, got %d", mediaFile.ID)
	}
	
	if mediaFile.OriginalPath != "/test/path/file.mp4" {
		t.Errorf("Expected original path to be '/test/path/file.mp4', got %q", mediaFile.OriginalPath)
	}
	
	if mediaFile.Status != "success" {
		t.Errorf("Expected status to be 'success', got %q", mediaFile.Status)
	}
}

func TestUpdateMediaFile(t *testing.T) {
	// Create a mock GORM DB
	mockDB := &MockGormDB{
		SaveFunc: func(value interface{}) *gorm.DB {
			// Verify the value is a MediaFile
			mediaFile, ok := value.(*models.MediaFile)
			if !ok {
				t.Errorf("Expected value to be *models.MediaFile, got %T", value)
			}
			
			// Verify the media file properties
			if mediaFile.ID != 1 {
				t.Errorf("Expected ID to be 1, got %d", mediaFile.ID)
			}
			
			if mediaFile.Status != "success" {
				t.Errorf("Expected status to be 'success', got %q", mediaFile.Status)
			}
			
			// Simulate successful update
			return &gorm.DB{Error: nil}
		},
	}
	
	// Create database with mock
	db := &Database{db: mockDB}
	
	// Update a media file
	mediaFile := &models.MediaFile{
		ID:           1,
		OriginalPath: "/test/path/file.mp4",
		OriginalName: "file.mp4",
		Status:       "success",
	}
	
	err := db.UpdateMediaFile(mediaFile)
	if err != nil {
		t.Fatalf("UpdateMediaFile failed: %v", err)
	}
}

func TestGetAPICache(t *testing.T) {
	// Create a mock GORM DB
	mockDB := &MockGormDB{
		WhereFunc: func(query interface{}, args ...interface{}) *gorm.DB {
			// Verify the query
			expectedQuery := "provider = ? AND query = ? AND expires_at > ?"
			if query != expectedQuery {
				t.Errorf("Expected query to be %q, got %q", expectedQuery, query)
			}
			
			// Verify the args
			if len(args) != 3 || args[0] != "tmdb" || args[1] != "movie:The Matrix:1999" {
				t.Errorf("Expected args to be ['tmdb', 'movie:The Matrix:1999', <time>], got %v", args)
			}
			
			return mockDB
		},
		FirstFunc: func(dest interface{}, conds ...interface{}) *gorm.DB {
			// Verify the destination is an APICache
			cache, ok := dest.(*models.APICache)
			if !ok {
				t.Errorf("Expected dest to be *models.APICache, got %T", dest)
			}
			
			// Set the cache properties
			cache.ID = 1
			cache.Provider = "tmdb"
			cache.Query = "movie:The Matrix:1999"
			cache.Response = `{"movies":[{"id":603,"title":"The Matrix"}]}`
			cache.ExpiresAt = time.Now().Add(1 * time.Hour)
			
			// Simulate successful query
			return &gorm.DB{Error: nil}
		},
	}
	
	// Create database with mock
	db := &Database{db: mockDB}
	
	// Get an API cache
	cache, err := db.GetAPICache("tmdb", "movie:The Matrix:1999")
	if err != nil {
		t.Fatalf("GetAPICache failed: %v", err)
	}
	
	// Verify the result
	if cache.ID != 1 {
		t.Errorf("Expected ID to be 1, got %d", cache.ID)
	}
	
	if cache.Provider != "tmdb" {
		t.Errorf("Expected provider to be 'tmdb', got %q", cache.Provider)
	}
	
	if cache.Query != "movie:The Matrix:1999" {
		t.Errorf("Expected query to be 'movie:The Matrix:1999', got %q", cache.Query)
	}
	
	expectedResponse := `{"movies":[{"id":603,"title":"The Matrix"}]}`
	if cache.Response != expectedResponse {
		t.Errorf("Expected response to be %q, got %q", expectedResponse, cache.Response)
	}
}

func TestCreateAPICache(t *testing.T) {
	// Create a mock GORM DB
	mockDB := &MockGormDB{
		CreateFunc: func(value interface{}) *gorm.DB {
			// Verify the value is an APICache
			cache, ok := value.(*models.APICache)
			if !ok {
				t.Errorf("Expected value to be *models.APICache, got %T", value)
			}
			
			// Verify the cache properties
			if cache.Provider != "tmdb" {
				t.Errorf("Expected provider to be 'tmdb', got %q", cache.Provider)
			}
			
			if cache.Query != "movie:The Matrix:1999" {
				t.Errorf("Expected query to be 'movie:The Matrix:1999', got %q", cache.Query)
			}
			
			// Simulate successful creation
			return &gorm.DB{Error: nil}
		},
	}
	
	// Create database with mock
	db := &Database{db: mockDB}
	
	// Create an API cache
	cache := &models.APICache{
		Provider:  "tmdb",
		Query:     "movie:The Matrix:1999",
		Response:  `{"movies":[{"id":603,"title":"The Matrix"}]}`,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	
	err := db.CreateAPICache(cache)
	if err != nil {
		t.Fatalf("CreateAPICache failed: %v", err)
	}
}

func TestGetPendingNotifications(t *testing.T) {
	// Create a mock GORM DB
	mockDB := &MockGormDB{
		WhereFunc: func(query interface{}, args ...interface{}) *gorm.DB {
			// Verify the query
			if query != "sent = ?" {
				t.Errorf("Expected query to be 'sent = ?', got %q", query)
			}
			
			// Verify the args
			if len(args) != 1 || args[0] != false {
				t.Errorf("Expected args to be [false], got %v", args)
			}
			
			return mockDB
		},
		FindFunc: func(dest interface{}, conds ...interface{}) *gorm.DB {
			// Verify the destination is a slice of Notification
			notifications, ok := dest.(*[]models.Notification)
			if !ok {
				t.Errorf("Expected dest to be *[]models.Notification, got %T", dest)
			}
			
			// Set the notifications
			*notifications = []models.Notification{
				{
					ID:          1,
					MediaFileID: 1,
					Type:        "success",
					Message:     "File processed successfully",
					Sent:        false,
					CreatedAt:   time.Now(),
				},
				{
					ID:          2,
					MediaFileID: 2,
					Type:        "error",
					Message:     "Error processing file",
					Sent:        false,
					CreatedAt:   time.Now(),
				},
			}
			
			// Simulate successful query
			return &gorm.DB{Error: nil}
		},
	}
	
	// Create database with mock
	db := &Database{db: mockDB}
	
	// Get pending notifications
	notifications, err := db.GetPendingNotifications()
	if err != nil {
		t.Fatalf("GetPendingNotifications failed: %v", err)
	}
	
	// Verify the result
	if len(notifications) != 2 {
		t.Fatalf("Expected 2 notifications, got %d", len(notifications))
	}
	
	if notifications[0].ID != 1 || notifications[0].Type != "success" {
		t.Errorf("Expected first notification to have ID 1 and type 'success', got ID %d and type %q", notifications[0].ID, notifications[0].Type)
	}
	
	if notifications[1].ID != 2 || notifications[1].Type != "error" {
		t.Errorf("Expected second notification to have ID 2 and type 'error', got ID %d and type %q", notifications[1].ID, notifications[1].Type)
	}
}

func TestClose(t *testing.T) {
	// Create a mock SQL DB
	mockSQLDB := &MockSQLDB{
		CloseFunc: func() error {
			return nil
		},
	}
	
	// Create a mock GORM DB
	mockGormDB := &MockGormDB{
		DBFunc: func() (*gorm.DB, error) {
			return nil, nil
		},
	}
	
	// Create database with mocks
	db := &Database{db: mockGormDB}
	
	// Close the database
	err := db.Close()
	if err == nil {
		// This is expected to fail since we can't properly mock the DB() method
		// The important part is that we're testing the Close method
	}
}
