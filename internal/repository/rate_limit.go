package repository

import (
	"context"
	"log"
	"time"

	"github.com/google/go-github/v76/github"
)

// HandleRateLimit detects API rate limiting and waits appropriately
//
// Preconditions:
// - resp is a GitHub API response
//
// Postconditions:
// - If rate limit is reached, waits until limit is reset
//
// Invariants:
// - Wait duration is calculated from response headers
func HandleRateLimit(ctx context.Context, resp *github.Response) error {
	// Check rate limit status
	if resp.Rate.Remaining == 0 {
		// Rate limit reached
		// Reset is the time when rate limit will be reset (time.Time type)
		resetTime := resp.Rate.Reset.Time
		waitDuration := time.Until(resetTime)

		// Set wait duration to 0 if negative (already reset)
		if waitDuration < 0 {
			waitDuration = 0
		}

		// Add a small buffer to wait duration (1 second)
		waitDuration += time.Second

		if waitDuration > 0 {
			log.Printf("Rate limit reached. Waiting %v...", waitDuration)
			select {
			case <-ctx.Done():
				return ctx.Err() // Context was cancelled
			case <-time.After(waitDuration):
				// Wait completed
			}
		}
	} else {
		// Log remaining requests if rate limit has room
		log.Printf("Rate limit remaining: %d/%d (reset time: %v)",
			resp.Rate.Remaining,
			resp.Rate.Limit,
			resp.Rate.Reset.Time.Format("2006-01-02 15:04:05"))
	}

	return nil
}
