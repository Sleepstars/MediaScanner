package worker

import (
	"context"

	"github.com/sleepstars/mediascanner/internal/models"
)

// WorkerPool defines the interface for a worker pool
type WorkerPool interface {
	// Start starts the worker pool
	Start()
	
	// Stop stops the worker pool
	Stop()
	
	// AddMediaFileTask adds a media file task to the worker pool
	AddMediaFileTask(mediaFile *models.MediaFile) error
	
	// AddBatchProcessTask adds a batch process task to the worker pool
	AddBatchProcessTask(batchProcess *models.BatchProcess) error
	
	// AcquireLLMSemaphore acquires the LLM semaphore
	AcquireLLMSemaphore(ctx context.Context) error
	
	// ReleaseLLMSemaphore releases the LLM semaphore
	ReleaseLLMSemaphore()
	
	// AcquireAPISemaphore acquires the API semaphore
	AcquireAPISemaphore(ctx context.Context) error
	
	// ReleaseAPISemaphore releases the API semaphore
	ReleaseAPISemaphore()
	
	// AcquireFileOpSemaphore acquires the file operation semaphore
	AcquireFileOpSemaphore(ctx context.Context) error
	
	// ReleaseFileOpSemaphore releases the file operation semaphore
	ReleaseFileOpSemaphore()
	
	// GetQueueStats returns statistics about the task queues
	GetQueueStats() (int, int)
}

// AddMediaFileTask adds a media file task to the worker pool
func (p *Pool) AddMediaFileTask(mediaFile *models.MediaFile) error {
	task := &Task{
		Type:      TaskTypeMediaFile,
		MediaFile: mediaFile,
	}
	return p.AddTask(task)
}

// AddBatchProcessTask adds a batch process task to the worker pool
func (p *Pool) AddBatchProcessTask(batchProcess *models.BatchProcess) error {
	task := &Task{
		Type:         TaskTypeBatchProcess,
		BatchProcess: batchProcess,
	}
	return p.AddTask(task)
}
