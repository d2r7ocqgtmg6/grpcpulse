package scheduler

import (
	"context"
	"log/slog"
	"time"

	"github.com/yourorg/grpcpulse/internal/checker"
	"github.com/yourorg/grpcpulse/internal/circuitbreaker"
	"github.com/yourorg/grpcpulse/internal/metrics"
	"github.com/yourorg/grpcpulse/internal/notifier"
)

// Config holds scheduler configuration.
type Config struct {
	Targets  []string
	Interval time.Duration
}

// DefaultConfig returns sensible scheduler defaults.
func DefaultConfig() Config {
	return Config{
		Interval: 15 * time.Second,
	}
}

// Deps groups the dependencies required by the Scheduler.
type Deps struct {
	Checker  *checker.Checker
	Metrics  *metrics.Metrics
	Notifier *notifier.Notifier
	CB       *circuitbreaker.CircuitBreaker
	Logger   *slog.Logger
}

// Scheduler periodically checks gRPC targets.
type Scheduler struct {
	cfg  Config
	deps Deps
	stop chan struct{}
}

// New creates a Scheduler with the provided config and dependencies.
func New(cfg Config, deps Deps) *Scheduler {
	if cfg.Interval <= 0 {
		cfg.Interval = DefaultConfig().Interval
	}
	if deps.Logger == nil {
		deps.Logger = slog.Default()
	}
	return &Scheduler{cfg: cfg, deps: deps, stop: make(chan struct{})}
}

// Run starts the check loop and blocks until ctx is cancelled or Stop is called.
func (s *Scheduler) Run(ctx context.Context) {
	ticker := time.NewTicker(s.cfg.Interval)
	defer ticker.Stop()
	s.checkAll(ctx)
	for {
		select {
		case <-ticker.C:
			s.checkAll(ctx)
		case <-s.stop:
			return
		case <-ctx.Done():
			return
		}
	}
}

// Stop signals the scheduler to cease operation.
func (s *Scheduler) Stop() { close(s.stop) }

func (s *Scheduler) checkAll(ctx context.Context) {
	for _, target := range s.cfg.Targets {
		if s.deps.CB != nil && !s.deps.CB.Allow(target) {
			s.deps.Logger.Warn("circuit open, skipping check", "target", target)
			continue
		}
		result := s.deps.Checker.Check(ctx, target)
		if s.deps.Metrics != nil {
			s.deps.Metrics.Record(result)
		}
		if s.deps.Notifier != nil {
			s.deps.Notifier.Observe(target, result)
		}
		if s.deps.CB != nil {
			if result.Healthy {
				s.deps.CB.RecordSuccess(target)
			} else {
				s.deps.CB.RecordFailure(target)
			}
		}
	}
}
