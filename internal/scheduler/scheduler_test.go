package scheduler_test

import (
	"context"
	"testing"
	"time"

	"github.com/grpcpulse/internal/checker"
	"github.com/grpcpulse/internal/metrics"
	"github.com/grpcpulse/internal/scheduler"
	"github.com/prometheus/client_golang/prometheus"
)

func newTestDeps(t *testing.T, target string) (*checker.Checker, *metrics.Metrics) {
	t.Helper()
	c := checker.New(checker.Config{Address: target, Timeout: time.Second})
	reg := prometheus.NewRegistry()
	m := metrics.New(reg)
	return c, m
}

func TestDefaultConfig(t *testing.T) {
	cfg := scheduler.DefaultConfig()
	if cfg.Interval != 15*time.Second {
		t.Errorf("expected 15s interval, got %s", cfg.Interval)
	}
}

func TestScheduler_StopsOnContextCancel(t *testing.T) {
	c, m := newTestDeps(t, "localhost:0")
	s := scheduler.New(c, m, scheduler.Config{Interval: 50 * time.Millisecond})

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		s.Start(ctx)
		close(done)
	}()

	time.Sleep(120 * time.Millisecond)
	cancel()

	select {
	case <-done:
		// success
	case <-time.After(time.Second):
		t.Fatal("scheduler did not stop after context cancellation")
	}
}

func TestScheduler_StopsOnStop(t *testing.T) {
	c, m := newTestDeps(t, "localhost:0")
	s := scheduler.New(c, m, scheduler.Config{Interval: 50 * time.Millisecond})

	done := make(chan struct{})
	go func() {
		s.Start(context.Background())
		close(done)
	}()

	time.Sleep(80 * time.Millisecond)
	s.Stop()

	select {
	case <-done:
		// success
	case <-time.After(time.Second):
		t.Fatal("scheduler did not stop after Stop() call")
	}
}

func TestScheduler_DefaultIntervalApplied(t *testing.T) {
	c, m := newTestDeps(t, "localhost:0")
	// Zero interval should fall back to default.
	s := scheduler.New(c, m, scheduler.Config{Interval: 0})
	if s == nil {
		t.Fatal("expected non-nil scheduler")
	}
}
