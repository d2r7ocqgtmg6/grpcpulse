// Package circuitbreaker implements a simple circuit breaker for gRPC health checks.
// It tracks consecutive failures per target and opens the circuit after a threshold,
// preventing further checks until a cooldown period elapses.
package circuitbreaker

import (
	"sync"
	"time"
)

// State represents the circuit breaker state.
type State int

const (
	StateClosed   State = iota // normal operation
	StateOpen                  // failing; checks skipped
	StateHalfOpen              // cooldown elapsed; next check allowed
)

func (s State) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// Config holds circuit breaker configuration.
type Config struct {
	Threshold int           // consecutive failures before opening
	Cooldown  time.Duration // duration to wait before half-opening
}

// DefaultConfig returns sensible defaults.
func DefaultConfig() Config {
	return Config{
		Threshold: 3,
		Cooldown:  30 * time.Second,
	}
}

type entry struct {
	failures  int
	openedAt  time.Time
	state     State
}

// CircuitBreaker tracks per-target circuit state.
type CircuitBreaker struct {
	cfg     Config
	mu      sync.Mutex
	entries map[string]*entry
}

// New creates a CircuitBreaker with the given config.
func New(cfg Config) *CircuitBreaker {
	if cfg.Threshold <= 0 {
		cfg.Threshold = DefaultConfig().Threshold
	}
	if cfg.Cooldown <= 0 {
		cfg.Cooldown = DefaultConfig().Cooldown
	}
	return &CircuitBreaker{
		cfg:     cfg,
		entries: make(map[string]*entry),
	}
}

func (cb *CircuitBreaker) get(target string) *entry {
	e, ok := cb.entries[target]
	if !ok {
		e = &entry{state: StateClosed}
		cb.entries[target] = e
	}
	return e
}

// Allow reports whether a check for target should proceed.
func (cb *CircuitBreaker) Allow(target string) bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	e := cb.get(target)
	switch e.state {
	case StateClosed:
		return true
	case StateOpen:
		if time.Since(e.openedAt) >= cb.cfg.Cooldown {
			e.state = StateHalfOpen
			return true
		}
		return false
	case StateHalfOpen:
		return true
	}
	return true
}

// RecordSuccess resets the circuit for target.
func (cb *CircuitBreaker) RecordSuccess(target string) {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	e := cb.get(target)
	e.failures = 0
	e.state = StateClosed
}

// RecordFailure increments the failure count and may open the circuit.
func (cb *CircuitBreaker) RecordFailure(target string) {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	e := cb.get(target)
	e.failures++
	if e.failures >= cb.cfg.Threshold {
		e.state = StateOpen
		e.openedAt = time.Now()
	}
}

// StateOf returns the current state for target.
func (cb *CircuitBreaker) StateOf(target string) State {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.get(target).state
}
