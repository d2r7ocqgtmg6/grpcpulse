package silencer_test

import (
	"testing"
	"time"

	"github.com/yourorg/grpcpulse/internal/silencer"
)

func TestIsSilenced_NotPresent(t *testing.T) {
	r := silencer.New()
	if r.IsSilenced("host:50051") {
		t.Fatal("expected not silenced")
	}
}

func TestSilence_ActiveDuringWindow(t *testing.T) {
	r := silencer.New()
	r.Silence("host:50051", "maintenance", 10*time.Minute)
	if !r.IsSilenced("host:50051") {
		t.Fatal("expected target to be silenced")
	}
}

func TestSilence_ExpiredReturnsNotSilenced(t *testing.T) {
	r := silencer.New()
	r.Silence("host:50051", "test", -1*time.Second) // already expired
	if r.IsSilenced("host:50051") {
		t.Fatal("expected expired silence to be inactive")
	}
}

func TestLift_RemovesSilence(t *testing.T) {
	r := silencer.New()
	r.Silence("host:50051", "manual", 10*time.Minute)
	r.Lift("host:50051")
	if r.IsSilenced("host:50051") {
		t.Fatal("expected silence to be lifted")
	}
}

func TestActive_ReturnsOnlyNonExpired(t *testing.T) {
	r := silencer.New()
	r.Silence("a:1", "reason", 10*time.Minute)
	r.Silence("b:2", "expired", -1*time.Second)
	r.Silence("c:3", "reason", 5*time.Minute)

	active := r.Active()
	if len(active) != 2 {
		t.Fatalf("expected 2 active silences, got %d", len(active))
	}
}

func TestSilence_OverwritesPrevious(t *testing.T) {
	r := silencer.New()
	r.Silence("host:50051", "first", -1*time.Second)
	r.Silence("host:50051", "second", 10*time.Minute)
	if !r.IsSilenced("host:50051") {
		t.Fatal("expected overwritten silence to be active")
	}
}

func TestActive_EmptyRegistry(t *testing.T) {
	r := silencer.New()
	if got := r.Active(); len(got) != 0 {
		t.Fatalf("expected empty, got %d", len(got))
	}
}
