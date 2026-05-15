package server_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/example/grpcpulse/internal/dependency"
	"github.com/example/grpcpulse/internal/server"
)

func newDepRegistry() *dependency.Registry {
	return dependency.New()
}

func TestDependencyHandler_GetEmpty(t *testing.T) {
	reg := newDepRegistry()
	h := server.ExportedDependencyHandler(reg)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/dependency?target=svc-a", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var body map[string][]string
	_ = json.NewDecoder(rec.Body).Decode(&body)
	if len(body["depends_on"]) != 0 || len(body["dependents"]) != 0 {
		t.Fatalf("expected empty lists, got %v", body)
	}
}

func TestDependencyHandler_AddAndGet(t *testing.T) {
	reg := newDepRegistry()
	h := server.ExportedDependencyHandler(reg)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/dependency?target=svc-a&dep=svc-b", nil))
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}

	rec2 := httptest.NewRecorder()
	h.ServeHTTP(rec2, httptest.NewRequest(http.MethodGet, "/dependency?target=svc-a", nil))
	var body map[string][]string
	_ = json.NewDecoder(rec2.Body).Decode(&body)
	if len(body["depends_on"]) != 1 || body["depends_on"][0] != "svc-b" {
		t.Fatalf("expected svc-b in depends_on, got %v", body)
	}
}

func TestDependencyHandler_CycleReturnsConflict(t *testing.T) {
	reg := newDepRegistry()
	_ = reg.Add("svc-a", "svc-b")
	_ = reg.Add("svc-b", "svc-c")
	h := server.ExportedDependencyHandler(reg)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/dependency?target=svc-c&dep=svc-a", nil))
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409 conflict, got %d", rec.Code)
	}
}

func TestDependencyHandler_Delete(t *testing.T) {
	reg := newDepRegistry()
	_ = reg.Add("svc-a", "svc-b")
	h := server.ExportedDependencyHandler(reg)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodDelete, "/dependency?target=svc-a&dep=svc-b", nil))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
	if deps := reg.DependsOn("svc-a"); len(deps) != 0 {
		t.Fatalf("expected no deps after delete, got %v", deps)
	}
}

func TestDependencyHandler_InvalidMethod(t *testing.T) {
	reg := newDepRegistry()
	h := server.ExportedDependencyHandler(reg)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPatch, "/dependency?target=svc-a", nil))
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}
