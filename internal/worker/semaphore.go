package worker

import (
	"context"
)

// Semaphore defines the interface for a semaphore
type Semaphore interface {
	// Acquire acquires the semaphore
	Acquire(ctx context.Context) error
	
	// Release releases the semaphore
	Release()
}

// NoOpSemaphore is a no-op implementation of Semaphore
type NoOpSemaphore struct{}

// Acquire acquires the semaphore (no-op)
func (s *NoOpSemaphore) Acquire(ctx context.Context) error {
	return nil
}

// Release releases the semaphore (no-op)
func (s *NoOpSemaphore) Release() {
	// No-op
}

// NewNoOpSemaphore creates a new no-op semaphore
func NewNoOpSemaphore() Semaphore {
	return &NoOpSemaphore{}
}

// LLMSemaphore is a wrapper for the LLM semaphore
type LLMSemaphore struct {
	pool *Pool
}

// Acquire acquires the LLM semaphore
func (s *LLMSemaphore) Acquire(ctx context.Context) error {
	return s.pool.AcquireLLMSemaphore(ctx)
}

// Release releases the LLM semaphore
func (s *LLMSemaphore) Release() {
	s.pool.ReleaseLLMSemaphore()
}

// NewLLMSemaphore creates a new LLM semaphore
func NewLLMSemaphore(pool *Pool) Semaphore {
	return &LLMSemaphore{pool: pool}
}

// APISemaphore is a wrapper for the API semaphore
type APISemaphore struct {
	pool *Pool
}

// Acquire acquires the API semaphore
func (s *APISemaphore) Acquire(ctx context.Context) error {
	return s.pool.AcquireAPISemaphore(ctx)
}

// Release releases the API semaphore
func (s *APISemaphore) Release() {
	s.pool.ReleaseAPISemaphore()
}

// NewAPISemaphore creates a new API semaphore
func NewAPISemaphore(pool *Pool) Semaphore {
	return &APISemaphore{pool: pool}
}

// FileOpSemaphore is a wrapper for the file operation semaphore
type FileOpSemaphore struct {
	pool *Pool
}

// Acquire acquires the file operation semaphore
func (s *FileOpSemaphore) Acquire(ctx context.Context) error {
	return s.pool.AcquireFileOpSemaphore(ctx)
}

// Release releases the file operation semaphore
func (s *FileOpSemaphore) Release() {
	s.pool.ReleaseFileOpSemaphore()
}

// NewFileOpSemaphore creates a new file operation semaphore
func NewFileOpSemaphore(pool *Pool) Semaphore {
	return &FileOpSemaphore{pool: pool}
}
