package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type fakeOncall struct {
	rotations map[string]string // target -> responder
}

func (f *fakeOncall) CurrentResponder(target string, _ time.Time) (string, bool) {
	r, ok := f.rotations[target]
	return r, ok
}

func (f *fakeOncall) All() map[string]interface{} {
	out := make(map[string]interface{}, len(f.rotations))
	for k, v := range f.rotations {
		out[k] = v
	}
	return out
}

func newFakeOncall(rotations map[string]string) *fakeOncall {
	return &fakeOncall{rotations: rotations}
}

func TestOncallHandler_ListAll(t *testing.T) {
	oc := newFakeOncall(map[string]string{"svc-a": "alice", "svc-b": "bob"})
	req := httptest.NewRequest(http.MethodGet, "/oncall", nil)
	w := httptest.NewRecorder()
	oncallHandler(oc).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var body map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(body) != 2 {
		t.Errorf("expected 2 entries, got %d", len(body))
	}
}

func TestOncallHandler_GetKnownTarget(t *testing.T) {
	oc := newFakeOncall(map[string]string{"svc-a": "alice"})
	req := httptest.NewRequest(http.MethodGet, "/oncall?target=svc-a", nil)
	w := httptest.NewRecorder()
	oncallHandler(oc).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var body map[string]string
	_ = json.NewDecoder(w.Body).Decode(&body)
	if body["responder"] != "alice" {
		t.Errorf("expected alice, got %q", body["responder"])
	}
}

func TestOncallHandler_GetUnknownTarget(t *testing.T) {
	oc := newFakeOncall(map[string]string{})
	req := httptest.NewRequest(http.MethodGet, "/oncall?target=unknown", nil)
	w := httptest.NewRecorder()
	oncallHandler(oc).ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestOncallHandler_InvalidMethod(t *testing.T) {
	oc := newFakeOncall(map[string]string{})
	req := httptest.NewRequest(http.MethodPost, "/oncall", nil)
	w := httptest.NewRecorder()
	oncallHandler(oc).ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}
