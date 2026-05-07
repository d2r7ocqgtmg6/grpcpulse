package circuitbreaker_test

import (
	"testing"
	"time"

	"github.com/yourorg/grpcpulse/internal/circuitbreaker"
)

func TestState_String(t *testing.T) {
	cases := []struct {
		s    circuitbreaker.State
		want string
	}{
		{circuitbreaker.StateClosed, "closed"},
		{circuitbreaker.StateOpen, "open"},
		{circuitbreaker.StateHalfOpen, "half-open"},
	}
	for _, c := range cases {
		if got := c.s.String(); got != c.want {
			t.Errorf("State(%d).String() = %q, want %q", c.s, got, c.want)
		}
	}
}

func TestAllow_ClosedByDefault(t *testing.T) {
	cb := circuitbreaker.New(circuitbreaker.DefaultConfig())
	if !cb.Allow("svc") {
		t.Fatal("expected Allow=true for fresh target")
	}
}

func TestCircuit_OpensAfterThreshold(t *testing.T) {
	cb := circuitbreaker.New(circuitbreaker.Config{Threshold: 2, Cooldown: time.Hour})
	cb.RecordFailure("svc")
	if !cb.Allow("svc") {
		t.Fatal("circuit should still be closed after 1 failure")
	}
	cb.RecordFailure("svc")
	if cb.Allow("svc") {
		t.Fatal("circuit should be open after threshold failures")
	}
	if cb.StateOf("svc") != circuitbreaker.StateOpen {
		t.Errorf("expected StateOpen, got %s", cb.StateOf("svc"))
	}
}

func TestCircuit_SuccessResetsClosed(t *testing.T) {
	cb := circuitbreaker.New(circuitbreaker.Config{Threshold: 1, Cooldown: time.Hour})
	cb.RecordFailure("svc")
	cb.RecordSuccess("svc")
	if cb.StateOf("svc") != circuitbreaker.StateClosed {
		t.Errorf("expected StateClosed after success, got %s", cb.StateOf("svc"))
	}
	if !cb.Allow("svc") {
		t.Fatal("expected Allow=true after reset")
	}
}

func TestCircuit_HalfOpenAfterCooldown(t *testing.T) {
	cb := circuitbreaker.New(circuitbreaker.Config{Threshold: 1, Cooldown: 10 * time.Millisecond})
	cb.RecordFailure("svc")
	if cb.Allow("svc") {
		t.Fatal("should be blocked immediately after open")
	}
	time.Sleep(20 * time.Millisecond)
	if !cb.Allow("svc") {
		t.Fatal("should allow after cooldown (half-open)")
	}
	if cb.StateOf("svc") != circuitbreaker.StateHalfOpen {
		t.Errorf("expected StateHalfOpen, got %s", cb.StateOf("svc"))
	}
}

func TestCircuit_IndependentTargets(t *testing.T) {
	cb := circuitbreaker.New(circuitbreaker.Config{Threshold: 1, Cooldown: time.Hour})
	cb.RecordFailure("svc-a")
	if !cb.Allow("svc-b") {
		t.Fatal("svc-b should be unaffected by svc-a failures")
	}
}

func TestDefaultConfig_Sensible(t *testing.T) {
	cfg := circuitbreaker.DefaultConfig()
	if cfg.Threshold <= 0 {
		t.Errorf("expected positive threshold, got %d", cfg.Threshold)
	}
	if cfg.Cooldown <= 0 {
		t.Errorf("expected positive cooldown, got %v", cfg.Cooldown)
	}
}
