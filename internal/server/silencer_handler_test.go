package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/grpcpulse/internal/silencer"
)

func newSilencer() *silencer.Silencer {
	return silencer.New()
}

func TestSilencerHandler_ListEmpty(t *testing.T) {
	s := newSilencer()
	h := silencerHandler(s)

	req := httptest.NewRequest(http.MethodGet, "/silences", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var result []interface{}
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty list, got %d items", len(result))
	}
}

func TestSilencerHandler_AddAndList(t *testing.T) {
	s := newSilencer()
	h := silencerHandler(s)

	body, _ := json.Marshal(map[string]interface{}{
		"target":   "svc-a",
		"duration": "5m",
	})
	req := httptest.NewRequest(http.MethodPost, "/silences", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/silences", nil)
	rec2 := httptest.NewRecorder()
	h.ServeHTTP(rec2, req2)

	var result []map[string]interface{}
	if err := json.NewDecoder(rec2.Body).Decode(&result); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(result) != 1 {
		t.Errorf("expected 1 silence, got %d", len(result))
	}
}

func TestSilencerHandler_InvalidMethod(t *testing.T) {
	s := newSilencer()
	h := silencerHandler(s)

	req := httptest.NewRequest(http.MethodPut, "/silences", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

func TestSilencerHandler_LiftSilence(t *testing.T) {
	s := newSilencer()
	s.Silence("svc-b", time.Minute)
	h := silencerHandler(s)

	req := httptest.NewRequest(http.MethodDelete, "/silences?target=svc-b", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
	if s.IsSilenced("svc-b") {
		t.Error("expected silence to be lifted")
	}
}
