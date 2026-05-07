// Package ratelimit provides a simple token-bucket rate limiter
// to throttle outgoing health-check requests per target.
package ratelimit

import (
	"sync"
	"time"
)

// Limiter controls how frequently checks may be performed for a given target.
type Limiter struct {
	mu       sync.Mutex
	tokens   map[string]time.Time
	minDelay time.Duration
}

// Config holds configuration for the Limiter.
type Config struct {
	// MinDelay is the minimum time between successive checks for the same target.
	MinDelay time.Duration
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() Config {
	return Config{
		MinDelay: 5 * time.Second,
	}
}

// New creates a new Limiter with the given config.
func New(cfg Config) *Limiter {
	if cfg.MinDelay <= 0 {
		cfg.MinDelay = DefaultConfig().MinDelay
	}
	return &Limiter{
		tokens:   make(map[string]time.Time),
		minDelay: cfg.MinDelay,
	}
}

// Allow reports whether a check for target is permitted right now.
// If allowed, it updates the internal timestamp for the target.
func (l *Limiter) Allow(target string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	next, ok := l.tokens[target]
	if ok && time.Now().Before(next) {
		return false
	}
	l.tokens[target] = time.Now().Add(l.minDelay)
	return true
}

// Reset clears the rate-limit state for target, allowing an immediate check.
func (l *Limiter) Reset(target string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.tokens, target)
}

// NextAllowed returns the earliest time at which target may be checked again.
// If the target has no recorded state, it returns the zero time.
func (l *Limiter) NextAllowed(target string) time.Time {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.tokens[target]
}
