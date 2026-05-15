package escalation_test

import (
	"testing"
	"time"

	"github.com/yourorg/grpcpulse/internal/escalation"
)

var (
	t0      = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	policy  = escalation.Policy{
		WarnAfter:     1 * time.Minute,
		CriticalAfter: 3 * time.Minute,
		PageAfter:     10 * time.Minute,
	}
)

func newReg() *escalation.Registry {
	return escalation.New(policy)
}

func TestLevel_String(t *testing.T) {
	cases := []struct {
		level escalation.Level
		want  string
	}{
		{escalation.LevelNone, "none"},
		{escalation.LevelWarn, "warn"},
		{escalation.LevelCritical, "critical"},
		{escalation.LevelPage, "page"},
	}
	for _, c := range cases {
		if got := c.level.String(); got != c.want {
			t.Errorf("Level(%d).String() = %q, want %q", c.level, got, c.want)
		}
	}
}

func TestObserve_NoneBeforeThreshold(t *testing.T) {
	reg := newReg()
	level := reg.Observe("svc-a", false, t0)
	if level != escalation.LevelNone {
		t.Fatalf("expected LevelNone, got %s", level)
	}
}

func TestObserve_WarnAfterThreshold(t *testing.T) {
	reg := newReg()
	reg.Observe("svc-a", false, t0)
	level := reg.Observe("svc-a", false, t0.Add(90*time.Second))
	if level != escalation.LevelWarn {
		t.Fatalf("expected LevelWarn, got %s", level)
	}
}

func TestObserve_CriticalAfterThreshold(t *testing.T) {
	reg := newReg()
	reg.Observe("svc-a", false, t0)
	level := reg.Observe("svc-a", false, t0.Add(4*time.Minute))
	if level != escalation.LevelCritical {
		t.Fatalf("expected LevelCritical, got %s", level)
	}
}

func TestObserve_PageAfterThreshold(t *testing.T) {
	reg := newReg()
	reg.Observe("svc-a", false, t0)
	level := reg.Observe("svc-a", false, t0.Add(11*time.Minute))
	if level != escalation.LevelPage {
		t.Fatalf("expected LevelPage, got %s", level)
	}
}

func TestObserve_HealthyClearsState(t *testing.T) {
	reg := newReg()
	reg.Observe("svc-a", false, t0)
	reg.Observe("svc-a", false, t0.Add(11*time.Minute))
	level := reg.Observe("svc-a", true, t0.Add(12*time.Minute))
	if level != escalation.LevelNone {
		t.Fatalf("expected LevelNone after recovery, got %s", level)
	}
	if got := reg.Current("svc-a"); got != escalation.LevelNone {
		t.Fatalf("Current after recovery: got %s, want none", got)
	}
}

func TestUnhealthySince_NotFound(t *testing.T) {
	reg := newReg()
	_, err := reg.UnhealthySince("unknown")
	if err == nil {
		t.Fatal("expected error for unknown target")
	}
}

func TestUnhealthySince_ReturnsFirstObservation(t *testing.T) {
	reg := newReg()
	reg.Observe("svc-b", false, t0)
	reg.Observe("svc-b", false, t0.Add(5*time.Minute))
	got, err := reg.UnhealthySince("svc-b")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got.Equal(t0) {
		t.Fatalf("UnhealthySince = %v, want %v", got, t0)
	}
}
