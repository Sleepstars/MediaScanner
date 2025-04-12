package processor

import (
	"context"
	"fmt"
	"github.com/sleepstars/mediascanner/internal/models"
	"github.com/sleepstars/mediascanner/internal/worker"
)

// StartWorkerPool starts the worker pool
func (p *Processor) StartWorkerPool() {
	if p.workerPool != nil {
		p.workerPool.Start()
	}
}

// StopWorkerPool stops the worker pool
func (p *Processor) StopWorkerPool() {
	if p.workerPool != nil {
		p.workerPool.Stop()
	}
}

// QueueMediaFile queues a media file for processing
func (p *Processor) QueueMediaFile(mediaFile *models.MediaFile) error {
	if p.workerPool == nil {
		return fmt.Errorf("worker pool is not enabled")
	}

	return p.workerPool.AddMediaFileTask(mediaFile)
}

// QueueBatchProcess queues a batch process for processing
func (p *Processor) QueueBatchProcess(batchProcess *models.BatchProcess) error {
	if p.workerPool == nil {
		return fmt.Errorf("worker pool is not enabled")
	}

	return p.workerPool.AddBatchProcessTask(batchProcess)
}

// processTask processes a task from the worker pool
func (p *Processor) processTask(ctx context.Context, task *worker.Task) error {
	switch task.Type {
	case worker.TaskTypeMediaFile:
		return p.ProcessMediaFile(ctx, task.MediaFile)
	case worker.TaskTypeBatchProcess:
		return p.ProcessBatchFiles(ctx, task.BatchProcess)
	default:
		return fmt.Errorf("unknown task type: %v", task.Type)
	}
}

// GetWorkerPoolStats returns statistics about the worker pool
func (p *Processor) GetWorkerPoolStats() (int, int) {
	if p.workerPool == nil {
		return 0, 0
	}

	return p.workerPool.GetQueueStats()
}

// IsWorkerPoolEnabled returns whether the worker pool is enabled
func (p *Processor) IsWorkerPoolEnabled() bool {
	return p.workerPool != nil
}
