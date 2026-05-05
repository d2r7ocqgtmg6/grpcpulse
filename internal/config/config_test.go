package config_test

import (
	"os"
	"testing"
	"time"

	"github.com/yourorg/grpcpulse/internal/config"
)

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "grpcpulse-*.yaml")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	f.Close()
	return f.Name()
}

func TestLoad_ValidConfig(t *testing.T) {
	path := writeTempConfig(t, `
listen_addr: ":8080"
targets:
  - name: auth-service
    address: localhost:50051
    interval: 10s
    timeout: 3s
    tls: false
`)

	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.ListenAddr != ":8080" {
		t.Errorf("listen_addr: got %q, want %q", cfg.ListenAddr, ":8080")
	}
	if len(cfg.Targets) != 1 {
		t.Fatalf("targets: got %d, want 1", len(cfg.Targets))
	}
	if cfg.Targets[0].Interval != 10*time.Second {
		t.Errorf("interval: got %v, want 10s", cfg.Targets[0].Interval)
	}
}

func TestLoad_DefaultsApplied(t *testing.T) {
	path := writeTempConfig(t, `
targets:
  - name: order-service
    address: localhost:50052
`)

	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.ListenAddr != ":9090" {
		t.Errorf("default listen_addr: got %q, want %q", cfg.ListenAddr, ":9090")
	}
	if cfg.Targets[0].Interval != 15*time.Second {
		t.Errorf("default interval: got %v, want 15s", cfg.Targets[0].Interval)
	}
	if cfg.Targets[0].Timeout != 5*time.Second {
		t.Errorf("default timeout: got %v, want 5s", cfg.Targets[0].Timeout)
	}
}

func TestLoad_MissingAddress(t *testing.T) {
	path := writeTempConfig(t, `
targets:
  - name: broken-service
`)
	if _, err := config.Load(path); err == nil {
		t.Fatal("expected error for missing address, got nil")
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	if _, err := config.Load("/nonexistent/path.yaml"); err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}
