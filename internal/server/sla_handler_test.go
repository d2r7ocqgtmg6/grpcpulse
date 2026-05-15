package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/yourorg/grpcpulse/internal/sla"
)

type stubUptime struct{}

func (s *stubUptime) UptimePercent(_ string, _ time.Duration) float64 { return 100.0 }

func newSLARegistry() *sla.Registry {
	return sla.New(&stubUptime{})
}

func TestSLAHandler_GetEmpty(t *testing.T) {
	reg := newSLARegistry()
	h := slaHandler(reg)

	req := httptest.NewRequest(http.MethodGet, "/sla", nil)
	rec := httptest.NewRecorder()
	h(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestSLAHandler_PutAndGet(t *testing.T) {
	reg := newSLARegistry()
	h := slaHandler(reg)

	body, _ := json.Marshal(map[string]interface{}{
		"target":         "svc:50051",
		"objective":      99.9,
		"window_seconds": 86400,
	})
	req := httptest.NewRequest(http.MethodPut, "/sla", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	h(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/sla?target=svc:50051", nil)
	rec2 := httptest.NewRecorder()
	h(rec2, req2)
	if rec2.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec2.Code)
	}

	var status sla.Status
	if err := json.NewDecoder(rec2.Body).Decode(&status); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if status.Target != "svc:50051" {
		t.Errorf("unexpected target %q", status.Target)
	}
}

func TestSLAHandler_GetUnknown(t *testing.T) {
	h := slaHandler(newSLARegistry())
	req := httptest.NewRequest(http.MethodGet, "/sla?target=unknown", nil)
	rec := httptest.NewRecorder()
	h(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestSLAHandler_DeleteEntry(t *testing.T) {
	reg := newSLARegistry()
	_ = reg.Set(sla.Budget{Target: "svc:50051", Objective: 99.9, Window: time.Hour})
	h := slaHandler(reg)

	req := httptest.NewRequest(http.MethodDelete, "/sla?target=svc:50051", nil)
	rec := httptest.NewRecorder()
	h(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/sla?target=svc:50051", nil)
	rec2 := httptest.NewRecorder()
	h(rec2, req2)
	if rec2.Code != http.StatusNotFound {
		t.Errorf("expected 404 after delete, got %d", rec2.Code)
	}
}

func TestSLAHandler_InvalidMethod(t *testing.T) {
	h := slaHandler(newSLARegistry())
	req := httptest.NewRequest(http.MethodPatch, "/sla", nil)
	rec := httptest.NewRecorder()
	h(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}
}
