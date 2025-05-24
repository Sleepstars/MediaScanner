package ratelimiter

import (
	"context"
	"sync"
	"time"
)

// RateLimiter defines the interface for a rate limiter
type RateLimiter interface {
	// Wait blocks until the rate limiter allows an event to happen
	Wait(ctx context.Context) error
	
	// TryWait returns true if the rate limiter allows an event to happen immediately
	TryWait(ctx context.Context) bool
}

// TokenBucketRateLimiter implements a token bucket rate limiter
type TokenBucketRateLimiter struct {
	rate       float64 // tokens per second
	bucketSize float64 // maximum burst size
	tokens     float64 // current number of tokens
	lastRefill time.Time
	mu         sync.Mutex
}

// NewTokenBucketRateLimiter creates a new token bucket rate limiter
func NewTokenBucketRateLimiter(rate float64, bucketSize float64) *TokenBucketRateLimiter {
	return &TokenBucketRateLimiter{
		rate:       rate,
		bucketSize: bucketSize,
		tokens:     bucketSize,
		lastRefill: time.Now(),
	}
}

// refill refills the token bucket based on the time elapsed since the last refill
func (tb *TokenBucketRateLimiter) refill() {
	now := time.Now()
	elapsed := now.Sub(tb.lastRefill).Seconds()
	tb.lastRefill = now
	
	// Calculate new tokens to add
	newTokens := elapsed * tb.rate
	
	// Add new tokens, but don't exceed bucket size
	tb.tokens = min(tb.tokens+newTokens, tb.bucketSize)
}

// Wait blocks until the rate limiter allows an event to happen
func (tb *TokenBucketRateLimiter) Wait(ctx context.Context) error {
	tb.mu.Lock()
	
	// Refill the bucket
	tb.refill()
	
	// If we have at least one token, consume it and return immediately
	if tb.tokens >= 1.0 {
		tb.tokens -= 1.0
		tb.mu.Unlock()
		return nil
	}
	
	// Calculate how long to wait for the next token
	waitTime := time.Duration((1.0 - tb.tokens) / tb.rate * float64(time.Second))
	
	// Unlock while waiting
	tb.mu.Unlock()
	
	// Wait for the token or context cancellation
	select {
	case <-time.After(waitTime):
		// Lock again to consume the token
		tb.mu.Lock()
		tb.refill()
		tb.tokens -= 1.0
		tb.mu.Unlock()
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// TryWait returns true if the rate limiter allows an event to happen immediately
func (tb *TokenBucketRateLimiter) TryWait(ctx context.Context) bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	
	// Refill the bucket
	tb.refill()
	
	// If we have at least one token, consume it and return true
	if tb.tokens >= 1.0 {
		tb.tokens -= 1.0
		return true
	}
	
	return false
}

// min returns the minimum of two float64 values
func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// ProviderRateLimiter manages rate limiters for different API providers
type ProviderRateLimiter struct {
	limiters map[string]RateLimiter
	mu       sync.RWMutex
}

// NewProviderRateLimiter creates a new provider rate limiter
func NewProviderRateLimiter() *ProviderRateLimiter {
	return &ProviderRateLimiter{
		limiters: make(map[string]RateLimiter),
	}
}

// RegisterLimiter registers a rate limiter for a provider
func (p *ProviderRateLimiter) RegisterLimiter(provider string, limiter RateLimiter) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.limiters[provider] = limiter
}

// Wait blocks until the rate limiter for the provider allows an event to happen
func (p *ProviderRateLimiter) Wait(ctx context.Context, provider string) error {
	p.mu.RLock()
	limiter, ok := p.limiters[provider]
	p.mu.RUnlock()
	
	if !ok {
		// No rate limiter for this provider, allow immediately
		return nil
	}
	
	return limiter.Wait(ctx)
}

// TryWait returns true if the rate limiter for the provider allows an event to happen immediately
func (p *ProviderRateLimiter) TryWait(ctx context.Context, provider string) bool {
	p.mu.RLock()
	limiter, ok := p.limiters[provider]
	p.mu.RUnlock()
	
	if !ok {
		// No rate limiter for this provider, allow immediately
		return true
	}
	
	return limiter.TryWait(ctx)
}
