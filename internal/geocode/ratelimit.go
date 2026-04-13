package geocode

import (
	"sync"
	"time"
)

// RateLimiter enforces a minimum delay between requests
type RateLimiter struct {
	lastRequest time.Time
	mu          sync.Mutex
	minDelay    time.Duration
}

// NewRateLimiter creates a new rate limiter with 1 second minimum delay
func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		minDelay: time.Second,
	}
}

// Wait blocks until enough time has passed since the last request
func (r *RateLimiter) Wait() {
	r.mu.Lock()
	defer r.mu.Unlock()

	elapsed := time.Since(r.lastRequest)
	if elapsed < r.minDelay {
		time.Sleep(r.minDelay - elapsed)
	}
	r.lastRequest = time.Now()
}
