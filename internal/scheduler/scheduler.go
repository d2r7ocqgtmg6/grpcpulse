package scheduler

import (
	"context"
	"log"
	"time"

	"github.com/grpcpulse/internal/checker"
	"github.com/grpcpulse/internal/metrics"
)

// Scheduler periodically runs health checks and records results.
type Scheduler struct {
	checker  *checker.Checker
	metrics  *metrics.Metrics
	interval time.Duration
	stopCh   chan struct{}
}

// Config holds configuration for the Scheduler.
type Config struct {
	Interval time.Duration
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() Config {
	return Config{
		Interval: 15 * time.Second,
	}
}

// New creates a new Scheduler.
func New(c *checker.Checker, m *metrics.Metrics, cfg Config) *Scheduler {
	if cfg.Interval <= 0 {
		cfg.Interval = DefaultConfig().Interval
	}
	return &Scheduler{
		checker:  c,
		metrics:  m,
		interval: cfg.Interval,
		stopCh:   make(chan struct{}),
	}
}

// Start begins the scheduling loop, running until ctx is cancelled or Stop is called.
func (s *Scheduler) Start(ctx context.Context) {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	// Run an immediate check on start.
	s.runCheck(ctx)

	for {
		select {
		case <-ticker.C:
			s.runCheck(ctx)
		case <-s.stopCh:
			log.Println("scheduler: stopped")
			return
		case <-ctx.Done():
			log.Println("scheduler: context cancelled")
			return
		}
	}
}

// Stop signals the scheduler to stop after the current check completes.
func (s *Scheduler) Stop() {
	close(s.stopCh)
}

func (s *Scheduler) runCheck(ctx context.Context) {
	result := s.checker.Check(ctx)
	s.metrics.Record(result)
	log.Printf("scheduler: check result status=%s latency=%s", result.Status, result.Latency)
}
