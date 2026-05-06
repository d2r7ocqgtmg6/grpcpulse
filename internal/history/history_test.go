package history_test

import (
	"testing"
	"time"

	"github.com/yourorg/grpcpulse/internal/history"
)

func TestRecord_And_Get(t *testing.T) {
	h := history.New(10)
	h.Record("svc-a", true, 5*time.Millisecond)
	h.Record("svc-a", false, 10*time.Millisecond)

	entries := h.Get("svc-a")
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if !entries[0].Healthy {
		t.Error("first entry should be healthy")
	}
	if entries[1].Healthy {
		t.Error("second entry should be unhealthy")
	}
}

func TestCapacity_Eviction(t *testing.T) {
	h := history.New(3)
	for i := 0; i < 5; i++ {
		h.Record("svc", true, time.Millisecond)
	}
	if got := len(h.Get("svc")); got != 3 {
		t.Fatalf("expected capacity 3, got %d", got)
	}
}

func TestGet_UnknownTarget(t *testing.T) {
	h := history.New(10)
	if entries := h.Get("ghost"); len(entries) != 0 {
		t.Fatalf("expected empty slice, got %d entries", len(entries))
	}
}

func TestTargets(t *testing.T) {
	h := history.New(10)
	h.Record("a", true, time.Millisecond)
	h.Record("b", false, time.Millisecond)

	targets := h.Targets()
	if len(targets) != 2 {
		t.Fatalf("expected 2 targets, got %d", len(targets))
	}
}

func TestUptimePercent(t *testing.T) {
	h := history.New(10)
	h.Record("svc", true, time.Millisecond)
	h.Record("svc", true, time.Millisecond)
	h.Record("svc", false, time.Millisecond)

	got := h.UptimePercent("svc")
	want := 66.66666666666667
	if got < 66.6 || got > 66.7 {
		t.Fatalf("expected ~%.2f%%, got %.2f%%", want, got)
	}
}

func TestUptimePercent_NoEntries(t *testing.T) {
	h := history.New(10)
	if pct := h.UptimePercent("empty"); pct != 0 {
		t.Fatalf("expected 0%%, got %.2f%%", pct)
	}
}

func TestDefaultCapacity(t *testing.T) {
	h := history.New(0) // should default to 100
	for i := 0; i < 105; i++ {
		h.Record("svc", true, time.Millisecond)
	}
	if got := len(h.Get("svc")); got != 100 {
		t.Fatalf("expected default capacity 100, got %d", got)
	}
}
