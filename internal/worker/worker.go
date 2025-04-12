package worker

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/sleepstars/mediascanner/internal/config"
	"github.com/sleepstars/mediascanner/internal/models"
)

// TaskType represents the type of task
type TaskType int

const (
	// TaskTypeMediaFile represents a media file processing task
	TaskTypeMediaFile TaskType = iota
	// TaskTypeBatchProcess represents a batch process task
	TaskTypeBatchProcess
)

// Task represents a task to be processed by the worker pool
type Task struct {
	Type         TaskType
	MediaFile    *models.MediaFile
	BatchProcess *models.BatchProcess
}

// ProcessorFunc is a function that processes a task
type ProcessorFunc func(ctx context.Context, task *Task) error

// Pool represents a worker pool
type Pool struct {
	config         *config.WorkerPoolConfig
	taskQueue      chan *Task
	batchTaskQueue chan *Task
	llmSemaphore   chan struct{}
	apiSemaphore   chan struct{}
	fileOpSemaphore chan struct{}
	processor      ProcessorFunc
	wg             sync.WaitGroup
	ctx            context.Context
	cancel         context.CancelFunc
}

// New creates a new worker pool
func New(cfg *config.WorkerPoolConfig, processor ProcessorFunc) *Pool {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &Pool{
		config:         cfg,
		taskQueue:      make(chan *Task, cfg.QueueSize),
		batchTaskQueue: make(chan *Task, cfg.QueueSize),
		llmSemaphore:   make(chan struct{}, cfg.MaxConcurrentLLM),
		apiSemaphore:   make(chan struct{}, cfg.MaxConcurrentAPI),
		fileOpSemaphore: make(chan struct{}, cfg.MaxConcurrentFileOp),
		processor:      processor,
		ctx:            ctx,
		cancel:         cancel,
	}
}

// Start starts the worker pool
func (p *Pool) Start() {
	log.Printf("Starting worker pool with %d workers and %d batch workers", p.config.WorkerCount, p.config.BatchWorkerCount)
	
	// Start regular workers
	for i := 0; i < p.config.WorkerCount; i++ {
		p.wg.Add(1)
		go p.worker(i)
	}
	
	// Start batch workers
	for i := 0; i < p.config.BatchWorkerCount; i++ {
		p.wg.Add(1)
		go p.batchWorker(i)
	}
}

// Stop stops the worker pool
func (p *Pool) Stop() {
	log.Println("Stopping worker pool...")
	p.cancel()
	p.wg.Wait()
	log.Println("Worker pool stopped")
}

// AddTask adds a task to the worker pool
func (p *Pool) AddTask(task *Task) error {
	select {
	case <-p.ctx.Done():
		return fmt.Errorf("worker pool is stopped")
	default:
		if task.Type == TaskTypeBatchProcess {
			select {
			case p.batchTaskQueue <- task:
				return nil
			default:
				return fmt.Errorf("batch task queue is full")
			}
		} else {
			select {
			case p.taskQueue <- task:
				return nil
			default:
				return fmt.Errorf("task queue is full")
			}
		}
	}
}

// worker processes tasks from the task queue
func (p *Pool) worker(id int) {
	defer p.wg.Done()
	
	log.Printf("Worker %d started", id)
	
	for {
		select {
		case <-p.ctx.Done():
			log.Printf("Worker %d stopping", id)
			return
		case task := <-p.taskQueue:
			log.Printf("Worker %d processing task: %v", id, task.Type)
			
			// Create a context with timeout
			ctx, cancel := context.WithTimeout(p.ctx, 30*time.Minute)
			
			// Process the task
			err := p.processor(ctx, task)
			if err != nil {
				log.Printf("Worker %d error processing task: %v", id, err)
			}
			
			cancel()
		}
	}
}

// batchWorker processes batch tasks from the batch task queue
func (p *Pool) batchWorker(id int) {
	defer p.wg.Done()
	
	log.Printf("Batch worker %d started", id)
	
	for {
		select {
		case <-p.ctx.Done():
			log.Printf("Batch worker %d stopping", id)
			return
		case task := <-p.batchTaskQueue:
			log.Printf("Batch worker %d processing batch task", id)
			
			// Create a context with timeout
			ctx, cancel := context.WithTimeout(p.ctx, 2*time.Hour)
			
			// Process the task
			err := p.processor(ctx, task)
			if err != nil {
				log.Printf("Batch worker %d error processing batch task: %v", id, err)
			}
			
			cancel()
		}
	}
}

// AcquireLLMSemaphore acquires the LLM semaphore
func (p *Pool) AcquireLLMSemaphore(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case p.llmSemaphore <- struct{}{}:
		return nil
	}
}

// ReleaseLLMSemaphore releases the LLM semaphore
func (p *Pool) ReleaseLLMSemaphore() {
	select {
	case <-p.llmSemaphore:
	default:
		log.Println("Warning: Attempting to release LLM semaphore that wasn't acquired")
	}
}

// AcquireAPISemaphore acquires the API semaphore
func (p *Pool) AcquireAPISemaphore(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case p.apiSemaphore <- struct{}{}:
		return nil
	}
}

// ReleaseAPISemaphore releases the API semaphore
func (p *Pool) ReleaseAPISemaphore() {
	select {
	case <-p.apiSemaphore:
	default:
		log.Println("Warning: Attempting to release API semaphore that wasn't acquired")
	}
}

// AcquireFileOpSemaphore acquires the file operation semaphore
func (p *Pool) AcquireFileOpSemaphore(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case p.fileOpSemaphore <- struct{}{}:
		return nil
	}
}

// ReleaseFileOpSemaphore releases the file operation semaphore
func (p *Pool) ReleaseFileOpSemaphore() {
	select {
	case <-p.fileOpSemaphore:
	default:
		log.Println("Warning: Attempting to release file operation semaphore that wasn't acquired")
	}
}

// GetQueueStats returns statistics about the task queues
func (p *Pool) GetQueueStats() (int, int) {
	return len(p.taskQueue), len(p.batchTaskQueue)
}
