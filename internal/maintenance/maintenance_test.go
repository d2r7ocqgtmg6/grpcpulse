package maintenance

import (
	"testing"
	"time"
)

func fixedNow(t time.Time) func() time.Time { return func() time.Time { return t } }

var base = time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)

func TestIsUnderMaintenance_NoWindows(t *testing.T) {
	r := New()
	if r.IsUnderMaintenance("svc-a") {
		t.Fatal("expected false for unknown target")
	}
}

func TestIsUnderMaintenance_ActiveWindow(t *testing.T) {
	r := New()
	r.now = fixedNow(base)
	r.Add(Window{
		Target: "svc-a",
		Start:  base.Add(-time.Minute),
		End:    base.Add(time.Hour),
		Reason: "deploy",
	})
	if !r.IsUnderMaintenance("svc-a") {
		t.Fatal("expected true for active window")
	}
}

func TestIsUnderMaintenance_ExpiredWindow(t *testing.T) {
	r := New()
	r.now = fixedNow(base)
	r.Add(Window{
		Target: "svc-a",
		Start:  base.Add(-2 * time.Hour),
		End:    base.Add(-time.Hour),
		Reason: "old deploy",
	})
	if r.IsUnderMaintenance("svc-a") {
		t.Fatal("expected false for expired window")
	}
}

func TestActive_ReturnsOnlyCurrentWindows(t *testing.T) {
	r := New()
	r.now = fixedNow(base)
	r.Add(Window{Target: "svc-a", Start: base.Add(-time.Minute), End: base.Add(time.Hour), Reason: "active"})
	r.Add(Window{Target: "svc-b", Start: base.Add(-2 * time.Hour), End: base.Add(-time.Hour), Reason: "expired"})

	active := r.Active()
	if len(active) != 1 {
		t.Fatalf("expected 1 active window, got %d", len(active))
	}
	if active[0].Target != "svc-a" {
		t.Errorf("unexpected target %s", active[0].Target)
	}
}

func TestRemove_ClearsTarget(t *testing.T) {
	r := New()
	r.now = fixedNow(base)
	r.Add(Window{Target: "svc-a", Start: base.Add(-time.Minute), End: base.Add(time.Hour)})
	r.Remove("svc-a")
	if r.IsUnderMaintenance("svc-a") {
		t.Fatal("expected false after Remove")
	}
}

func TestPurge_RemovesExpiredWindows(t *testing.T) {
	r := New()
	r.now = fixedNow(base)
	r.Add(Window{Target: "svc-a", Start: base.Add(-2 * time.Hour), End: base.Add(-time.Hour)})
	r.Add(Window{Target: "svc-b", Start: base.Add(-time.Minute), End: base.Add(time.Hour)})
	r.Purge()

	if r.IsUnderMaintenance("svc-a") {
		t.Error("svc-a should have been purged")
	}
	if !r.IsUnderMaintenance("svc-b") {
		t.Error("svc-b should still be active")
	}
}

func TestWindow_IsActive(t *testing.T) {
	w := Window{Start: base.Add(-time.Minute), End: base.Add(time.Hour)}
	if !w.IsActive(base) {
		t.Error("expected window to be active at base")
	}
	if w.IsActive(base.Add(2 * time.Hour)) {
		t.Error("expected window to be inactive after end")
	}
	if w.IsActive(base.Add(-2 * time.Minute)) {
		t.Error("expected window to be inactive before start")
	}
}
