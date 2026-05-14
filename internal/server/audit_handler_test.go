package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yourorg/grpcpulse/internal/audit"
)

func newAuditLog(t *testing.T) *audit.Log {
	t.Helper()
	return audit.New(64)
}

func TestAuditEndpoint_Empty(t *testing.T) {
	log := newAuditLog(t)
	h := auditHandler(log)

	req := httptest.NewRequest(http.MethodGet, "/audit", nil)
	rec := httptest.NewRecorder()
	h(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var events []audit.Event
	if err := json.NewDecoder(rec.Body).Decode(&events); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(events) != 0 {
		t.Errorf("expected empty list, got %d events", len(events))
	}
}

func TestAuditEndpoint_AllEvents(t *testing.T) {
	log := newAuditLog(t)
	log.Record(audit.EventKindTagSet, "svc-a", "set region=eu")
	log.Record(audit.EventKindCircuitOpen, "svc-b", "threshold exceeded")

	h := auditHandler(log)
	req := httptest.NewRequest(http.MethodGet, "/audit", nil)
	rec := httptest.NewRecorder()
	h(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var events []audit.Event
	if err := json.NewDecoder(rec.Body).Decode(&events); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(events) != 2 {
		t.Errorf("expected 2 events, got %d", len(events))
	}
}

func TestAuditEndpoint_FilterByTarget(t *testing.T) {
	log := newAuditLog(t)
	log.Record(audit.EventKindTagSet, "svc-a", "tag")
	log.Record(audit.EventKindSilenceAdded, "svc-b", "silenced")
	log.Record(audit.EventKindCircuitOpen, "svc-a", "open")

	h := auditHandler(log)
	req := httptest.NewRequest(http.MethodGet, "/audit?target=svc-a", nil)
	rec := httptest.NewRecorder()
	h(rec, req)

	var events []audit.Event
	if err := json.NewDecoder(rec.Body).Decode(&events); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(events) != 2 {
		t.Errorf("expected 2 events for svc-a, got %d", len(events))
	}
}

func TestAuditEndpoint_MethodNotAllowed(t *testing.T) {
	log := newAuditLog(t)
	h := auditHandler(log)

	req := httptest.NewRequest(http.MethodPost, "/audit", nil)
	rec := httptest.NewRecorder()
	h(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}
}
