package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/grpcpulse/internal/acknowledger"
)

func newAcknowledger() *acknowledger.Acknowledger {
	return acknowledger.New()
}

func TestAcknowledgerHandler_ListEmpty(t *testing.T) {
	h := acknowledgerHandler(newAcknowledger())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/acks", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var acks []acknowledger.Acknowledgement
	if err := json.NewDecoder(rec.Body).Decode(&acks); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(acks) != 0 {
		t.Fatalf("expected empty list, got %d", len(acks))
	}
}

func TestAcknowledgerHandler_AddAndList(t *testing.T) {
	store := newAcknowledger()
	h := acknowledgerHandler(store)

	body, _ := json.Marshal(map[string]string{"target": "svc-a", "duration": "5m"})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/acks", bytes.NewReader(body)))
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}

	rec2 := httptest.NewRecorder()
	h.ServeHTTP(rec2, httptest.NewRequest(http.MethodGet, "/acks", nil))
	var acks []acknowledger.Acknowledgement
	json.NewDecoder(rec2.Body).Decode(&acks)
	if len(acks) != 1 {
		t.Fatalf("expected 1 ack, got %d", len(acks))
	}
	if acks[0].Target != "svc-a" {
		t.Errorf("unexpected target %q", acks[0].Target)
	}
}

func TestAcknowledgerHandler_InvalidMethod(t *testing.T) {
	h := acknowledgerHandler(newAcknowledger())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPut, "/acks", nil))
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

func TestAcknowledgerHandler_DeleteAck(t *testing.T) {
	store := newAcknowledger()
	_ = store.Acknowledge("svc-b", 10*time.Minute)

	h := acknowledgerHandler(store)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/acks?target=svc-b", nil)
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
	if store.IsAcknowledged("svc-b") {
		t.Error("expected svc-b to be cleared")
	}
}

func TestAcknowledgerHandler_InvalidDuration(t *testing.T) {
	h := acknowledgerHandler(newAcknowledger())
	body, _ := json.Marshal(map[string]string{"target": "svc-c", "duration": "bad"})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/acks", bytes.NewReader(body)))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}
