package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/grpcpulse/internal/acknowledger"
)

func newAcknowledger() *acknowledgerHandler {
	return &acknowledgerHandler{ack: acknowledger.New()}
}

func TestAcknowledgerHandler_ListEmpty(t *testing.T) {
	h := newAcknowledger()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/acks", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var out []interface{}
	json.NewDecoder(rec.Body).Decode(&out)
	if len(out) != 0 {
		t.Errorf("expected empty list, got %d items", len(out))
	}
}

func TestAcknowledgerHandler_AddAndList(t *testing.T) {
	h := newAcknowledger()
	body := `{"target":"svc:50051","reason":"known","acked_by":"alice","ttl":"1h"}`
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/acks", bytes.NewBufferString(body)))
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}

	rec2 := httptest.NewRecorder()
	h.ServeHTTP(rec2, httptest.NewRequest(http.MethodGet, "/acks", nil))
	var out []map[string]interface{}
	json.NewDecoder(rec2.Body).Decode(&out)
	if len(out) != 1 {
		t.Fatalf("expected 1 ack, got %d", len(out))
	}
	if out[0]["target"] != "svc:50051" {
		t.Errorf("unexpected target %v", out[0]["target"])
	}
}

func TestAcknowledgerHandler_InvalidMethod(t *testing.T) {
	h := newAcknowledger()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPut, "/acks", nil))
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

func TestAcknowledgerHandler_DeleteAck(t *testing.T) {
	h := newAcknowledger()
	body := `{"target":"svc:50051","reason":"r","acked_by":"bob","ttl":"1h"}`
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/acks", bytes.NewBufferString(body)))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodDelete, "/acks?target=svc:50051", nil))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
	if h.ack.IsAcknowledged("svc:50051") {
		t.Error("expected ack to be removed")
	}
}

func TestAcknowledgerHandler_MissingTargetOnDelete(t *testing.T) {
	h := newAcknowledger()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodDelete, "/acks", nil))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}
