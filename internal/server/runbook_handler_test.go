package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yourorg/grpcpulse/internal/runbook"
)

func newRunbookRegistry() *runbook.Registry { return runbook.New() }

func TestRunbookHandler_GetEmpty(t *testing.T) {
	reg := newRunbookRegistry()
	h := runbookHandler(reg)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/runbooks", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var out map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&out); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(out) != 0 {
		t.Fatalf("expected empty map, got %v", out)
	}
}

func TestRunbookHandler_PutAndGet(t *testing.T) {
	reg := newRunbookRegistry()
	h := runbookHandler(reg)

	body, _ := json.Marshal(map[string]string{"url": "https://wiki.example.com/rb"})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPut, "/runbooks?target=svc:50051", bytes.NewReader(body)))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}

	rec2 := httptest.NewRecorder()
	h.ServeHTTP(rec2, httptest.NewRequest(http.MethodGet, "/runbooks?target=svc:50051", nil))
	if rec2.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec2.Code)
	}
	var out map[string]string
	_ = json.NewDecoder(rec2.Body).Decode(&out)
	if out["url"] != "https://wiki.example.com/rb" {
		t.Fatalf("unexpected url: %v", out)
	}
}

func TestRunbookHandler_InvalidMethod(t *testing.T) {
	h := runbookHandler(newRunbookRegistry())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/runbooks", nil))
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

func TestRunbookHandler_DeleteEntry(t *testing.T) {
	reg := newRunbookRegistry()
	_ = reg.Set("svc:1", "https://example.com/rb")
	h := runbookHandler(reg)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodDelete, "/runbooks?target=svc:1", nil))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
	_, ok := reg.Get("svc:1")
	if ok {
		t.Fatal("entry should have been removed")
	}
}

func TestRunbookHandler_GetUnknownTarget(t *testing.T) {
	h := runbookHandler(newRunbookRegistry())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/runbooks?target=ghost:9", nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}
