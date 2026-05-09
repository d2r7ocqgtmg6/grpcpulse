package dashboard_test

import (
	"testing"
	"time"

	"github.com/grpcpulse/internal/checker"
	"github.com/grpcpulse/internal/dashboard"
	"github.com/grpcpulse/internal/history"
	"github.com/grpcpulse/internal/notifier"
)

func newDeps(t *testing.T) (*history.History, *notifier.Notifier) {
	t.Helper()
	h := history.New(history.DefaultConfig())
	n := notifier.New(nil)
	return h, n
}

func TestBuild_EmptyHistory(t *testing.T) {
	h, n := newDeps(t)
	b := dashboard.New(h, n)
	snap := b.Build()
	if len(snap.Targets) != 0 {
		t.Fatalf("expected 0 targets, got %d", len(snap.Targets))
	}
	if snap.GeneratedAt.IsZero() {
		t.Fatal("expected GeneratedAt to be set")
	}
}

func TestBuild_SingleHealthyTarget(t *testing.T) {
	h, n := newDeps(t)
	addr := "localhost:50051"

	h.Record(addr, checker.Result{Healthy: true, Latency: 10 * time.Millisecond})
	n.Observe(addr, checker.Result{Healthy: true})

	snap := dashboard.New(h, n).Build()
	if len(snap.Targets) != 1 {
		t.Fatalf("expected 1 target, got %d", len(snap.Targets))
	}

	s := snap.Targets[0]
	if s.Address != addr {
		t.Errorf("address: got %q, want %q", s.Address, addr)
	}
	if s.UptimePercent != 100.0 {
		t.Errorf("uptime: got %.2f, want 100.00", s.UptimePercent)
	}
	if s.LastLatencyMs != 10 {
		t.Errorf("latency: got %.2f, want 10", s.LastLatencyMs)
	}
	if s.TotalChecks != 1 {
		t.Errorf("total checks: got %d, want 1", s.TotalChecks)
	}
}

func TestBuild_MultipleTargets(t *testing.T) {
	h, n := newDeps(t)
	addrs := []string{"svc-a:50051", "svc-b:50052"}

	for _, a := range addrs {
		h.Record(a, checker.Result{Healthy: true, Latency: 5 * time.Millisecond})
		n.Observe(a, checker.Result{Healthy: true})
	}

	snap := dashboard.New(h, n).Build()
	if len(snap.Targets) != 2 {
		t.Fatalf("expected 2 targets, got %d", len(snap.Targets))
	}
}
