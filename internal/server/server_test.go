package server_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/grpcpulse/internal/server"
)

func freePort() int {
	// Use a high ephemeral port range for tests to avoid conflicts.
	return 19090
}

func TestServer_HealthzEndpoint(t *testing.T) {
	port := freePort()
	cfg := server.Config{
		Port:            port,
		ReadTimeout:     2 * time.Second,
		ShutdownTimeout: 2 * time.Second,
	}

	srv := server.New(cfg)

	go func() {
		_ = srv.Start()
	}()

	// Give the server a moment to start.
	time.Sleep(100 * time.Millisecond)

	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/healthz", port))
	if err != nil {
		t.Fatalf("failed to reach /healthz: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if string(body) != "ok" {
		t.Errorf("expected body 'ok', got %q", string(body))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		t.Errorf("shutdown error: %v", err)
	}
}

func TestServer_MetricsEndpoint(t *testing.T) {
	port := freePort() + 1
	cfg := server.Config{
		Port:            port,
		ReadTimeout:     2 * time.Second,
		ShutdownTimeout: 2 * time.Second,
	}

	srv := server.New(cfg)
	go func() { _ = srv.Start() }()
	time.Sleep(100 * time.Millisecond)

	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/metrics", port))
	if err != nil {
		t.Fatalf("failed to reach /metrics: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
}

func TestDefaultConfig(t *testing.T) {
	cfg := server.DefaultConfig()
	if cfg.Port != 9090 {
		t.Errorf("expected default port 9090, got %d", cfg.Port)
	}
	if cfg.ReadTimeout != 5*time.Second {
		t.Errorf("expected read timeout 5s, got %v", cfg.ReadTimeout)
	}
}

func TestServer_Addr(t *testing.T) {
	cfg := server.Config{Port: 8080, ReadTimeout: time.Second, ShutdownTimeout: time.Second}
	srv := server.New(cfg)
	if srv.Addr() != ":8080" {
		t.Errorf("expected addr ':8080', got %q", srv.Addr())
	}
}
