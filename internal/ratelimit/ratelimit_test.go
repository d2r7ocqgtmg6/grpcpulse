package ratelimit_test

import (
	"testing"
	"time"

	"github.com/yourorg/grpcpulse/internal/ratelimit"
)

func TestAllow_FirstCallPermitted(t *testing.T) {
	l := ratelimit.New(ratelimit.Config{MinDelay: 100 * time.Millisecond})
	if !l.Allow("target:50051") {
		t.Fatal("expected first Allow to return true")
	}
}

func TestAllow_SecondCallBlocked(t *testing.T) {
	l := ratelimit.New(ratelimit.Config{MinDelay: 200 * time.Millisecond})
	l.Allow("target:50051")
	if l.Allow("target:50051") {
		t.Fatal("expected second immediate Allow to return false")
	}
}

func TestAllow_AfterDelayPermitted(t *testing.T) {
	l := ratelimit.New(ratelimit.Config{MinDelay: 50 * time.Millisecond})
	l.Allow("target:50051")
	time.Sleep(60 * time.Millisecond)
	if !l.Allow("target:50051") {
		t.Fatal("expected Allow to return true after delay elapsed")
	}
}

func TestAllow_IndependentTargets(t *testing.T) {
	l := ratelimit.New(ratelimit.Config{MinDelay: 500 * time.Millisecond})
	l.Allow("a:50051")
	if !l.Allow("b:50051") {
		t.Fatal("expected different target to be allowed independently")
	}
}

func TestReset_AllowsImmediateCheck(t *testing.T) {
	l := ratelimit.New(ratelimit.Config{MinDelay: 500 * time.Millisecond})
	l.Allow("target:50051")
	l.Reset("target:50051")
	if !l.Allow("target:50051") {
		t.Fatal("expected Allow to return true after Reset")
	}
}

func TestNextAllowed_ZeroForUnknown(t *testing.T) {
	l := ratelimit.New(ratelimit.DefaultConfig())
	if !l.NextAllowed("unknown").IsZero() {
		t.Fatal("expected zero time for unknown target")
	}
}

func TestNextAllowed_FutureAfterAllow(t *testing.T) {
	l := ratelimit.New(ratelimit.Config{MinDelay: 1 * time.Second})
	before := time.Now()
	l.Allow("target:50051")
	next := l.NextAllowed("target:50051")
	if !next.After(before) {
		t.Fatalf("expected NextAllowed to be in the future, got %v", next)
	}
}

func TestDefaultConfig_PositiveDelay(t *testing.T) {
	cfg := ratelimit.DefaultConfig()
	if cfg.MinDelay <= 0 {
		t.Fatalf("expected positive MinDelay, got %v", cfg.MinDelay)
	}
}
