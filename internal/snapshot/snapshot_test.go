package snapshot_test

import (
	"testing"

	"github.com/user/grpcpulse/internal/snapshot"
)

// fakeHealth implements HealthSource for testing.
type fakeHealth struct {
	targets []string
	healthy map[string]bool
	uptime  map[string]float64
}

func (f *fakeHealth) Targets() []string { return f.targets }
func (f *fakeHealth) Healthy(t string) bool { return f.healthy[t] }
func (f *fakeHealth) UptimePercent(t string) float64 { return f.uptime[t] }

// fakeTags implements TagSource for testing.
type fakeTags struct {
	data map[string]map[string]string
}

func (f *fakeTags) Get(t string) map[string]string { return f.data[t] }

func newDeps() (*fakeHealth, *fakeTags) {
	h := &fakeHealth{
		targets: []string{"svc-a", "svc-b"},
		healthy: map[string]bool{"svc-a": true, "svc-b": false},
		uptime:  map[string]float64{"svc-a": 99.5, "svc-b": 72.0},
	}
	tg := &fakeTags{
		data: map[string]map[string]string{
			"svc-a": {"env": "prod"},
		},
	}
	return h, tg
}

func TestCapture_ReturnsAllTargets(t *testing.T) {
	h, tg := newDeps()
	reg := snapshot.New(h, tg)
	snap := reg.Capture()

	if len(snap.Targets) != 2 {
		t.Fatalf("expected 2 targets, got %d", len(snap.Targets))
	}
}

func TestCapture_HealthState(t *testing.T) {
	h, tg := newDeps()
	reg := snapshot.New(h, tg)
	snap := reg.Capture()

	for _, ts := range snap.Targets {
		switch ts.Target {
		case "svc-a":
			if !ts.Healthy {
				t.Errorf("svc-a should be healthy")
			}
			if ts.UptimePct != 99.5 {
				t.Errorf("unexpected uptime: %v", ts.UptimePct)
			}
		case "svc-b":
			if ts.Healthy {
				t.Errorf("svc-b should be unhealthy")
			}
		}
	}
}

func TestCapture_TagsIncluded(t *testing.T) {
	h, tg := newDeps()
	reg := snapshot.New(h, tg)
	snap := reg.Capture()

	for _, ts := range snap.Targets {
		if ts.Target == "svc-a" {
			if ts.Tags["env"] != "prod" {
				t.Errorf("expected tag env=prod, got %v", ts.Tags)
			}
		}
	}
}

func TestLatest_NilBeforeCapture(t *testing.T) {
	h, tg := newDeps()
	reg := snapshot.New(h, tg)
	if reg.Latest() != nil {
		t.Error("expected nil before first capture")
	}
}

func TestLatest_AfterCapture(t *testing.T) {
	h, tg := newDeps()
	reg := snapshot.New(h, tg)
	reg.Capture()
	if reg.Latest() == nil {
		t.Error("expected non-nil after capture")
	}
}
