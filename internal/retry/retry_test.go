package retry_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/yourorg/grpcpulse/internal/retry"
)

var errFake = errors.New("fake error")

func TestDo_SuccessOnFirstAttempt(t *testing.T) {
	r := retry.New(retry.DefaultConfig())
	calls := 0
	err := r.Do(context.Background(), func(_ context.Context) error {
		calls++
		return nil
	})
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if calls != 1 {
		t.Fatalf("expected 1 call, got %d", calls)
	}
}

func TestDo_RetriesOnFailure(t *testing.T) {
	cfg := retry.Config{MaxAttempts: 3, Delay: time.Millisecond, Multiplier: 1.0}
	r := retry.New(cfg)
	calls := 0
	err := r.Do(context.Background(), func(_ context.Context) error {
		calls++
		return errFake
	})
	if !errors.Is(err, errFake) {
		t.Fatalf("expected errFake, got %v", err)
	}
	if calls != 3 {
		t.Fatalf("expected 3 calls, got %d", calls)
	}
}

func TestDo_SucceedsOnSecondAttempt(t *testing.T) {
	cfg := retry.Config{MaxAttempts: 3, Delay: time.Millisecond, Multiplier: 1.0}
	r := retry.New(cfg)
	calls := 0
	err := r.Do(context.Background(), func(_ context.Context) error {
		calls++
		if calls < 2 {
			return errFake
		}
		return nil
	})
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if calls != 2 {
		t.Fatalf("expected 2 calls, got %d", calls)
	}
}

func TestDo_ContextCancelled(t *testing.T) {
	cfg := retry.Config{MaxAttempts: 5, Delay: 100 * time.Millisecond, Multiplier: 1.0}
	r := retry.New(cfg)
	ctx, cancel := context.WithCancel(context.Background())
	calls := 0
	err := r.Do(ctx, func(_ context.Context) error {
		calls++
		cancel()
		return errFake
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
	if calls != 1 {
		t.Fatalf("expected 1 call before cancel, got %d", calls)
	}
}

func TestDefaultConfig_Sane(t *testing.T) {
	cfg := retry.DefaultConfig()
	if cfg.MaxAttempts <= 0 {
		t.Error("MaxAttempts must be positive")
	}
	if cfg.Delay <= 0 {
		t.Error("Delay must be positive")
	}
	if cfg.Multiplier <= 0 {
		t.Error("Multiplier must be positive")
	}
}
