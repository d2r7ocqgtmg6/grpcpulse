package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/grpcpulse/internal/tags"
)

func newTagsRegistry() *tags.Registry {
	return tags.New()
}

func TestTagsHandler_GetUnknownTarget(t *testing.T) {
	r := newTagsRegistry()
	h := tagsHandler(r)

	req := httptest.NewRequest(http.MethodGet, "/tags?target=unknown", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestTagsHandler_SetAndGet(t *testing.T) {
	r := newTagsRegistry()
	h := tagsHandler(r)

	body, _ := json.Marshal(map[string]interface{}{
		"target": "svc-x",
		"tags":   map[string]string{"env": "prod", "team": "platform"},
	})
	req := httptest.NewRequest(http.MethodPost, "/tags", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/tags?target=svc-x", nil)
	rec2 := httptest.NewRecorder()
	h.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec2.Code)
	}
	var result map[string]string
	if err := json.NewDecoder(rec2.Body).Decode(&result); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if result["env"] != "prod" {
		t.Errorf("expected env=prod, got %q", result["env"])
	}
	if result["team"] != "platform" {
		t.Errorf("expected team=platform, got %q", result["team"])
	}
}

func TestTagsHandler_InvalidMethod(t *testing.T) {
	r := newTagsRegistry()
	h := tagsHandler(r)

	req := httptest.NewRequest(http.MethodPatch, "/tags", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

func TestTagsHandler_MissingTargetParam(t *testing.T) {
	r := newTagsRegistry()
	h := tagsHandler(r)

	req := httptest.NewRequest(http.MethodGet, "/tags", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}
