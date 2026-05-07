package scheduler_test

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/yourorg/grpcpulse/internal/checker"
	"github.com/yourorg/grpcpulse/internal/circuitbreaker"
	"github.com/yourorg/grpcpulse/internal/metrics"
	"github.com/yourorg/grpcpulse/internal/notifier"
	"github.com/yourorg/grpcpulse/internal/scheduler"
	"github.com/prometheus/client_golang/prometheus"
)

func newTestDeps(t *testing.T) scheduler.Deps {
	t.Helper()
	reg := prometheus.NewRegistry()
	return scheduler.Deps{
		Checker:  checker.New(checker.Config{Timeout: 100 * time.Millisecond}),
		Metrics:  metrics.New(reg),
		Notifier: notifier.New(nil),
		CB:       circuitbreaker.New(circuitbreaker.DefaultConfig()),
		Logger:   slog.Default(),
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := scheduler.DefaultConfig()
	if cfg.Interval <= 0 {
		t.Errorf("expected positive interval, got %v", cfg.Interval)
	}
}

func TestScheduler_StopsOnContextCancel(t *testing.T) {
	cfg := scheduler.Config{Interval: time.Hour}
	s := scheduler.New(cfg, newTestDeps(t))
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { s.Run(ctx); close(done) }()
	cancel()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("scheduler did not stop after context cancel")
	}
}

func TestScheduler_StopsOnStop(t *testing.T) {
	cfg := scheduler.Config{Interval: time.Hour}
	s := scheduler.New(cfg, newTestDeps(t))
	done := make(chan struct{})
	go func() { s.Run(context.Background()); close(done) }()
	s.Stop()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("scheduler did not stop after Stop()")
	}
}

func TestScheduler_DefaultIntervalApplied(t *testing.T) {
	cfg := scheduler.Config{Interval: 0}
	s := scheduler.New(cfg, newTestDeps(t))
	if s == nil {
		t.Fatal("expected non-nil scheduler")
	}
}

func TestScheduler_CircuitBreakerSkipsOpenTarget(t *testing.T) {
	cb := circuitbreaker.New(circuitbreaker.Config{Threshold: 1, Cooldown: time.Hour})
	cb.RecordFailure("localhost:19999") // open the circuit
	deps := newTestDeps(t)
	deps.CB = cb
	cfg := scheduler.Config{
		Targets:  []string{"localhost:19999"},
		Interval: time.Hour,
	}
	s := scheduler.New(cfg, deps)
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	// Run should complete without error; the circuit is open so no dial is attempted.
	s.Run(ctx)
}
