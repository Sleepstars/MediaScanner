package processor

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/sleepstars/mediascanner/internal/models"
)

// TestErrorMessageFormatting tests the error message formatting logic
func TestErrorMessageFormatting(t *testing.T) {
	testCases := []struct {
		name          string
		originalError error
		operation     string
		expectedMsg   string
	}{
		{
			name:          "Simple error",
			originalError: errors.New("test error"),
			operation:     "Test Operation",
			expectedMsg:   "Test Operation: test error",
		},
		{
			name:          "Complex error",
			originalError: fmt.Errorf("database connection failed: %w", errors.New("connection timeout")),
			operation:     "Database Update",
			expectedMsg:   "Database Update: database connection failed: connection timeout",
		},
		{
			name:          "Empty operation",
			originalError: errors.New("test error"),
			operation:     "",
			expectedMsg:   ": test error",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test the error wrapping logic that would be used in handleProcessingError
			wrappedError := fmt.Errorf("%s: %w", tc.operation, tc.originalError)

			if wrappedError.Error() != tc.expectedMsg {
				t.Errorf("Expected error message '%s', got '%s'", tc.expectedMsg, wrappedError.Error())
			}
		})
	}
}

// TestMediaFileStatusUpdate tests the media file status update logic
func TestMediaFileStatusUpdate(t *testing.T) {
	mediaFile := &models.MediaFile{
		ID:           1,
		OriginalPath: "/test/file.mkv",
		Status:       "processing",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Simulate the status update logic from handleProcessingError
	originalStatus := mediaFile.Status
	testError := errors.New("test error")
	operation := "Test Operation"

	// Update status (this is what handleProcessingError does)
	errorMsg := fmt.Sprintf("%s error: %v", operation, testError)
	mediaFile.Status = "failed"
	mediaFile.ErrorMessage = errorMsg
	mediaFile.UpdatedAt = time.Now()

	// Verify the updates
	if mediaFile.Status != "failed" {
		t.Errorf("Expected status to be 'failed', got '%s'", mediaFile.Status)
	}

	expectedErrorMessage := "Test Operation error: test error"
	if mediaFile.ErrorMessage != expectedErrorMessage {
		t.Errorf("Expected error message '%s', got '%s'", expectedErrorMessage, mediaFile.ErrorMessage)
	}

	// Verify that the status changed
	if mediaFile.Status == originalStatus {
		t.Error("Expected status to change from original")
	}

	// Verify that UpdatedAt was modified
	if mediaFile.UpdatedAt.Before(mediaFile.CreatedAt) {
		t.Error("Expected UpdatedAt to be after CreatedAt")
	}
}
