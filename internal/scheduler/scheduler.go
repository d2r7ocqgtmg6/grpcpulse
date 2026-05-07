package scheduler

import (
	"context"
	"log/slog"
	"time"

	"github.com/yourorg/grpcpulse/internal/checker"
	"github.com/yourorg/grpcpulse/internal/metrics"
	"github.com/yourorg/grpcpulse/internal/ratelimit"
)

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() Config {
	return Config{
		Interval: 15 * time.Second,
	}
}

// Config holds configuration for the Scheduler.
type Config struct {
	Interval time.Duration
	Targets  []string
}

// Deps groups the external dependencies required by the Scheduler.
type Deps struct {
	Checker checker.Checker
	Metrics *metrics.Metrics
	Limiter *ratelimit.Limiter
	Logger  *slog.Logger
}

// Scheduler periodically runs health checks against configured gRPC targets.
type Scheduler struct {
	cfg     Config
	deps    Deps
	stopCh  chan struct{}
	done    chan struct{}
}

// New creates a new Scheduler. If cfg.Interval is zero the default is applied.
func New(cfg Config, deps Deps) *Scheduler {
	if cfg.Interval <= 0 {
		cfg.Interval = DefaultConfig().Interval
	}
	return &Scheduler{
		cfg:    cfg,
		deps:   deps,
		stopCh: make(chan struct{}),
		done:   make(chan struct{}),
	}
}

// Run starts the scheduling loop and blocks until ctx is cancelled or Stop is called.
func (s *Scheduler) Run(ctx context.Context) {
	defer close(s.done)
	ticker := time.NewTicker(s.cfg.Interval)
	defer ticker.Stop()

	s.checkAll(ctx)

	for {
		select {
		case <-ticker.C:
			s.checkAll(ctx)
		case <-s.stopCh:
			return
		case <-ctx.Done():
			return
		}
	}
}

// Stop signals the scheduler to stop and waits for it to finish.
func (s *Scheduler) Stop() {
	close(s.stopCh)
	<-s.done
}

func (s *Scheduler) checkAll(ctx context.Context) {
	for _, target := range s.cfg.Targets {
		if s.deps.Limiter != nil && !s.deps.Limiter.Allow(target) {
			s.deps.Logger.Debug("rate limited, skipping", "target", target)
			continue
		}
		result := s.deps.Checker.Check(ctx, target)
		s.deps.Metrics.Record(result)
		s.deps.Logger.Info("health check", "target", target, "status", result.Status)
	}
}
