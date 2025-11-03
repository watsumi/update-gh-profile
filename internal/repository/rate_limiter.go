package repository

import (
	"context"
	"sync"
	"time"

	"github.com/google/go-github/v76/github"
)

// RateLimiter manages GitHub API rate limiting
type RateLimiter struct {
	mu              sync.Mutex
	client          *github.Client
	remaining       int
	limit           int
	resetTime       time.Time
	requestInterval time.Duration // Request interval (default: 100ms)
	lastRequest     time.Time
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(client *github.Client) *RateLimiter {
	return &RateLimiter{
		client:          client,
		requestInterval: 100 * time.Millisecond, // Default: 100ms interval
	}
}

// SetRequestInterval sets the request interval
func (r *RateLimiter) SetRequestInterval(interval time.Duration) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.requestInterval = interval
}

// WaitIfNeeded waits if needed considering rate limits
func (r *RateLimiter) WaitIfNeeded(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check request interval
	now := time.Now()
	elapsed := now.Sub(r.lastRequest)
	if elapsed < r.requestInterval {
		waitTime := r.requestInterval - elapsed
		r.mu.Unlock()
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(waitTime):
			// Wait completed
		}
		r.mu.Lock()
	}
	r.lastRequest = time.Now()

	// Check rate limit status (if remaining count is low)
	if r.remaining > 0 && r.remaining < 100 {
		// Increase interval if remaining count is low
		waitTime := r.requestInterval * 2
		r.mu.Unlock()
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(waitTime):
			// Wait completed
		}
		return nil
	}

	// Check time until rate limit resets
	if !r.resetTime.IsZero() && r.remaining == 0 {
		waitTime := time.Until(r.resetTime)
		if waitTime > 0 {
			// Wait until reset time (max 10 seconds)
			maxWait := 10 * time.Second
			if waitTime > maxWait {
				waitTime = maxWait
			}
			r.mu.Unlock()
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(waitTime):
				// Wait completed
			}
			return nil
		}
	}

	return nil
}

// UpdateRateLimitInfo updates rate limit information
func (r *RateLimiter) UpdateRateLimitInfo(remaining, limit int, resetTime time.Time) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.remaining = remaining
	r.limit = limit
	r.resetTime = resetTime
}

// GetRemaining gets the remaining request count
func (r *RateLimiter) GetRemaining() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.remaining
}
