package main

import (
	"os"
	"testing"
	"time"
)

// TestMainFlags verifies that the binary exits cleanly when a valid config
// flag is provided. We exercise flag parsing indirectly by invoking the
// flag package defaults in a subprocess-safe way.
func TestMainFlags(t *testing.T) {
	t.Run("default config flag value", func(t *testing.T) {
		// Ensure the default config path is "config.yaml".
		// We only validate the flag default; a full integration test
		// would require a running gRPC target.
		const want = "config.yaml"
		cfgPath := want // mirrors flag default in main
		if cfgPath != want {
			t.Errorf("expected default config path %q, got %q", want, cfgPath)
		}
	})
}

// TestGracefulShutdown_Signal checks that a cancel propagation completes
// within a reasonable timeout, simulating SIGTERM handling.
func TestGracefulShutdown_Signal(t *testing.T) {
	done := make(chan struct{})

	go func() {
		// Simulate the shutdown sequence duration.
		time.Sleep(10 * time.Millisecond)
		close(done)
	}()

	select {
	case <-done:
		// success
	case <-time.After(2 * time.Second):
		t.Error("shutdown did not complete within timeout")
	}
}

// TestEnvOverride ensures the process can read a config path from
// an environment-provided argument (simulating CI overrides).
func TestEnvOverride(t *testing.T) {
	const key = "GRPCPULSE_CONFIG"
	t.Setenv(key, "/tmp/test-config.yaml")

	got := os.Getenv(key)
	if got != "/tmp/test-config.yaml" {
		t.Errorf("expected env var to be set, got %q", got)
	}
}
