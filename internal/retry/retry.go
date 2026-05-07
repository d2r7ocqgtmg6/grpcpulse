// Package retry provides a configurable retry mechanism for gRPC health checks.
package retry

import (
	"context"
	"time"
)

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() Config {
	return Config{
		MaxAttempts: 3,
		Delay:       500 * time.Millisecond,
		Multiplier:  2.0,
	}
}

// Config holds retry policy settings.
type Config struct {
	// MaxAttempts is the total number of attempts (including the first call).
	MaxAttempts int
	// Delay is the initial wait time between attempts.
	Delay time.Duration
	// Multiplier scales the delay after each failure.
	Multiplier float64
}

// Retryer executes a function with exponential back-off retry logic.
type Retryer struct {
	cfg Config
}

// New creates a Retryer with the given Config.
func New(cfg Config) *Retryer {
	if cfg.MaxAttempts <= 0 {
		cfg.MaxAttempts = DefaultConfig().MaxAttempts
	}
	if cfg.Delay <= 0 {
		cfg.Delay = DefaultConfig().Delay
	}
	if cfg.Multiplier <= 0 {
		cfg.Multiplier = DefaultConfig().Multiplier
	}
	return &Retryer{cfg: cfg}
}

// Do calls fn up to MaxAttempts times. It returns nil on the first success.
// If ctx is cancelled the function returns ctx.Err() immediately.
func (r *Retryer) Do(ctx context.Context, fn func(ctx context.Context) error) error {
	delay := r.cfg.Delay
	var err error
	for attempt := 0; attempt < r.cfg.MaxAttempts; attempt++ {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if err = fn(ctx); err == nil {
			return nil
		}
		if attempt < r.cfg.MaxAttempts-1 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}
			delay = time.Duration(float64(delay) * r.cfg.Multiplier)
		}
	}
	return err
}
