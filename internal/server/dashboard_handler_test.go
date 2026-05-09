package server_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/grpcpulse/internal/checker"
	"github.com/grpcpulse/internal/dashboard"
	"github.com/grpcpulse/internal/history"
	"github.com/grpcpulse/internal/notifier"
)

func newDashboardBuilder(t *testing.T) *dashboard.Builder {
	t.Helper()
	h := history.New(history.DefaultConfig())
	n := notifier.New(nil)
	return dashboard.New(h, n)
}

func newDashboardBuilderWithData(t *testing.T) *dashboard.Builder {
	t.Helper()
	h := history.New(history.DefaultConfig())
	n := notifier.New(nil)

	addr := "svc:9090"
	h.Record(addr, checker.Result{Healthy: true, Latency: 20 * time.Millisecond})
	n.Observe(addr, checker.Result{Healthy: true})

	return dashboard.New(h, n)
}

func TestDashboardEndpoint_Empty(t *testing.T) {
	b := newDashboardBuilder(t)
	srv := newServerWithHistory(t) // reuse helper that wires a real server
	_ = srv

	// Test the handler directly.
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)

	// Inline handler invocation.
	handler := dashboardHandlerExport(b)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d, want %d", rec.Code, http.StatusOK)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("content-type: got %q", ct)
	}

	var snap dashboard.Snapshot
	if err := json.NewDecoder(rec.Body).Decode(&snap); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(snap.Targets) != 0 {
		t.Errorf("expected 0 targets, got %d", len(snap.Targets))
	}
}

func TestDashboardEndpoint_WithData(t *testing.T) {
	b := newDashboardBuilderWithData(t)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)

	handler := dashboardHandlerExport(b)
	handler.ServeHTTP(rec, req)

	var snap dashboard.Snapshot
	if err := json.NewDecoder(rec.Body).Decode(&snap); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(snap.Targets) != 1 {
		t.Fatalf("expected 1 target, got %d", len(snap.Targets))
	}
	if snap.Targets[0].UptimePercent != 100.0 {
		t.Errorf("uptime: got %.2f", snap.Targets[0].UptimePercent)
	}
}
