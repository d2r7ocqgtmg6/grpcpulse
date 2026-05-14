package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/grpcpulse/internal/notifier"
)

type fakeNotifier struct {
	events []notifier.Event
	current map[string]notifier.State
}

func (f *fakeNotifier) Events() []notifier.Event { return f.events }
func (f *fakeNotifier) Current() map[string]notifier.State { return f.current }

func newFakeNotifier() *fakeNotifier {
	return &fakeNotifier{
		events:  []notifier.Event{},
		current: map[string]notifier.State{},
	}
}

func TestNotifierHandler_CurrentEmpty(t *testing.T) {
	fn := newFakeNotifier()
	h := notifierHandler(fn)

	req := httptest.NewRequest(http.MethodGet, "/notifier", nil)
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, req)

	if rw.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rw.Code)
	}
	var result map[string]string
	if err := json.NewDecoder(rw.Body).Decode(&result); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty map, got %v", result)
	}
}

func TestNotifierHandler_CurrentWithState(t *testing.T) {
	fn := newFakeNotifier()
	fn.current["svc-a:50051"] = notifier.StateHealthy
	fn.current["svc-b:50051"] = notifier.StateUnhealthy
	h := notifierHandler(fn)

	req := httptest.NewRequest(http.MethodGet, "/notifier", nil)
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, req)

	if rw.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rw.Code)
	}
	var result map[string]string
	if err := json.NewDecoder(rw.Body).Decode(&result); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if result["svc-a:50051"] != "healthy" {
		t.Errorf("expected healthy, got %s", result["svc-a:50051"])
	}
	if result["svc-b:50051"] != "unhealthy" {
		t.Errorf("expected unhealthy, got %s", result["svc-b:50051"])
	}
}

func TestNotifierEventsHandler_Empty(t *testing.T) {
	fn := newFakeNotifier()
	h := notifierEventsHandler(fn)

	req := httptest.NewRequest(http.MethodGet, "/notifier/events", nil)
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, req)

	if rw.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rw.Code)
	}
	var events []notifier.Event
	if err := json.NewDecoder(rw.Body).Decode(&events); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(events) != 0 {
		t.Errorf("expected empty events, got %d", len(events))
	}
}

func TestNotifierEventsHandler_WithEvents(t *testing.T) {
	fn := newFakeNotifier()
	fn.events = []notifier.Event{
		{Target: "svc-a:50051", From: notifier.StateHealthy, To: notifier.StateUnhealthy, At: time.Now()},
	}
	h := notifierEventsHandler(fn)

	req := httptest.NewRequest(http.MethodGet, "/notifier/events", nil)
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, req)

	if rw.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rw.Code)
	}
	body := rw.Body.String()
	if !strings.Contains(body, "svc-a:50051") {
		t.Errorf("expected target in response, got: %s", body)
	}
}

func TestNotifierHandler_MethodNotAllowed(t *testing.T) {
	fn := newFakeNotifier()
	h := notifierHandler(fn)

	req := httptest.NewRequest(http.MethodPost, "/notifier", nil)
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, req)

	if rw.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rw.Code)
	}
}
