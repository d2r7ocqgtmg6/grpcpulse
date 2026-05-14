package alertmanager_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/yourorg/grpcpulse/internal/alertmanager"
)

func TestDefaultConfig_Sane(t *testing.T) {
	cfg := alertmanager.DefaultConfig()
	if cfg.Timeout != 5*time.Second {
		t.Fatalf("expected 5s timeout, got %v", cfg.Timeout)
	}
}

func TestSend_Success(t *testing.T) {
	var received []alertmanager.Alert

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v2/alerts" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Errorf("decode body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client := alertmanager.New(alertmanager.Config{URL: srv.URL})
	alerts := []alertmanager.Alert{
		{
			Labels:      map[string]string{"alertname": "ServiceDown", "target": "svc:50051"},
			Annotations: map[string]string{"summary": "gRPC service is unhealthy"},
			StartsAt:    time.Now(),
		},
	}

	if err := client.Send(context.Background(), alerts); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(received) != 1 {
		t.Fatalf("expected 1 alert, got %d", len(received))
	}
	if received[0].Labels["target"] != "svc:50051" {
		t.Errorf("unexpected target label: %s", received[0].Labels["target"])
	}
}

func TestSend_EmptyAlertsNoRequest(t *testing.T) {
	called := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))
	defer srv.Close()

	client := alertmanager.New(alertmanager.Config{URL: srv.URL})
	if err := client.Send(context.Background(), nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called {
		t.Error("expected no HTTP request for empty alert list")
	}
}

func TestSend_NonSuccessStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	client := alertmanager.New(alertmanager.Config{URL: srv.URL})
	alerts := []alertmanager.Alert{
		{Labels: map[string]string{"alertname": "test"}, StartsAt: time.Now()},
	}

	err := client.Send(context.Background(), alerts)
	if err == nil {
		t.Fatal("expected error for 500 status, got nil")
	}
}

func TestSend_ContextCancelled(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	client := alertmanager.New(alertmanager.Config{URL: srv.URL})
	alerts := []alertmanager.Alert{
		{Labels: map[string]string{"alertname": "test"}, StartsAt: time.Now()},
	}

	if err := client.Send(ctx, alerts); err == nil {
		t.Fatal("expected error for cancelled context, got nil")
	}
}
