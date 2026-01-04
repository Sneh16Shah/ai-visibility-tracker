package ai

import (
	"context"
	"errors"
	"sync"
	"time"
)

// Common errors
var (
	ErrRateLimited      = errors.New("rate limit exceeded, please try again later")
	ErrRequestInFlight  = errors.New("analysis already in progress for this brand")
	ErrProviderNotReady = errors.New("AI provider not configured")
	ErrEmptyResponse    = errors.New("received empty response from AI")
)

// AIRequest represents a request to the AI provider
type AIRequest struct {
	BrandID    int
	PromptID   int
	PromptText string
}

// AIResult represents the result from an AI call
type AIResult struct {
	PromptText   string
	ResponseText string
	ModelName    string
	Error        error
}

// Provider is the interface for AI providers
type Provider interface {
	// Query sends a prompt to the AI and returns the response
	Query(ctx context.Context, prompt string) (string, error)
	// GetModelName returns the name of the AI model being used
	GetModelName() string
	// IsAvailable checks if the provider is properly configured
	IsAvailable() bool
}

// RateLimiter controls the rate of API calls
type RateLimiter struct {
	mu             sync.Mutex
	lastCall       time.Time
	minInterval    time.Duration // Minimum time between calls
	maxCallsPerMin int
	callsThisMin   int
	minuteStart    time.Time
}

// NewRateLimiter creates a new rate limiter
// minInterval: minimum time between individual calls (e.g., 2 seconds)
// maxCallsPerMin: maximum calls allowed per minute (e.g., 3)
func NewRateLimiter(minInterval time.Duration, maxCallsPerMin int) *RateLimiter {
	return &RateLimiter{
		minInterval:    minInterval,
		maxCallsPerMin: maxCallsPerMin,
		minuteStart:    time.Now(),
	}
}

// CanProceed checks if we can make another API call
func (r *RateLimiter) CanProceed() bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()

	// Reset minute counter if a minute has passed
	if now.Sub(r.minuteStart) >= time.Minute {
		r.callsThisMin = 0
		r.minuteStart = now
	}

	// Check if we've exceeded calls per minute
	if r.callsThisMin >= r.maxCallsPerMin {
		return false
	}

	// Check minimum interval between calls
	if now.Sub(r.lastCall) < r.minInterval {
		return false
	}

	return true
}

// RecordCall records that an API call was made
func (r *RateLimiter) RecordCall() {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	r.lastCall = now
	r.callsThisMin++

	// Reset if new minute
	if now.Sub(r.minuteStart) >= time.Minute {
		r.callsThisMin = 1
		r.minuteStart = now
	}
}

// TimeUntilNextAllowed returns the time until the next call is allowed
func (r *RateLimiter) TimeUntilNextAllowed() time.Duration {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()

	// Check if we need to wait for minute reset
	if r.callsThisMin >= r.maxCallsPerMin {
		return r.minuteStart.Add(time.Minute).Sub(now)
	}

	// Check if we need to wait for min interval
	nextAllowed := r.lastCall.Add(r.minInterval)
	if now.Before(nextAllowed) {
		return nextAllowed.Sub(now)
	}

	return 0
}

// GetStatus returns current rate limiter status
func (r *RateLimiter) GetStatus() map[string]interface{} {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()

	return map[string]interface{}{
		"calls_this_minute":    r.callsThisMin,
		"max_calls_per_minute": r.maxCallsPerMin,
		"seconds_until_reset":  int(r.minuteStart.Add(time.Minute).Sub(now).Seconds()),
		"can_proceed":          r.callsThisMin < r.maxCallsPerMin && now.Sub(r.lastCall) >= r.minInterval,
	}
}

// InFlightTracker prevents duplicate concurrent requests
type InFlightTracker struct {
	mu       sync.Mutex
	inFlight map[int]time.Time // brandID -> start time
	timeout  time.Duration
}

// NewInFlightTracker creates a new tracker
func NewInFlightTracker(timeout time.Duration) *InFlightTracker {
	return &InFlightTracker{
		inFlight: make(map[int]time.Time),
		timeout:  timeout,
	}
}

// TryAcquire attempts to acquire a slot for a brand analysis
// Returns true if acquired, false if already in progress
func (t *InFlightTracker) TryAcquire(brandID int) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Check if there's an existing request
	if startTime, exists := t.inFlight[brandID]; exists {
		// If it's been too long, consider it stale and allow new request
		if time.Since(startTime) < t.timeout {
			return false
		}
	}

	t.inFlight[brandID] = time.Now()
	return true
}

// Release releases the slot for a brand
func (t *InFlightTracker) Release(brandID int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.inFlight, brandID)
}

// IsInFlight checks if a brand analysis is in progress
func (t *InFlightTracker) IsInFlight(brandID int) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	if startTime, exists := t.inFlight[brandID]; exists {
		return time.Since(startTime) < t.timeout
	}
	return false
}
