package changelog_test

import (
	"testing"

	"github.com/yourorg/grpcpulse/internal/changelog"
)

func TestRecord_AppendsEntry(t *testing.T) {
	l := changelog.New(10)
	e := l.Record(changelog.KindConfig, "svc-a", "admin", "updated interval", nil)

	if e.ID != 0 {
		t.Fatalf("expected first ID=0, got %d", e.ID)
	}
	if e.Kind != changelog.KindConfig {
		t.Fatalf("unexpected kind: %s", e.Kind)
	}
	if e.Target != "svc-a" {
		t.Fatalf("unexpected target: %s", e.Target)
	}
	if e.Timestamp.IsZero() {
		t.Fatal("timestamp should not be zero")
	}
}

func TestAll_ReturnsCopy(t *testing.T) {
	l := changelog.New(10)
	l.Record(changelog.KindSilence, "svc-b", "ops", "silenced for deploy", nil)
	l.Record(changelog.KindAck, "svc-c", "dev", "acknowledged alert", nil)

	all := l.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(all))
	}
	// Mutating the returned slice must not affect the log.
	all[0].Message = "tampered"
	if l.All()[0].Message == "tampered" {
		t.Fatal("All() should return a copy, not a reference")
	}
}

func TestCapacity_Eviction(t *testing.T) {
	l := changelog.New(3)
	for i := 0; i < 5; i++ {
		l.Record(changelog.KindRunbook, "", "bot", "change", nil)
	}
	all := l.All()
	if len(all) != 3 {
		t.Fatalf("expected 3 entries after eviction, got %d", len(all))
	}
	// The oldest entries (ID 0,1) should have been evicted.
	if all[0].ID != 2 {
		t.Fatalf("expected oldest remaining ID=2, got %d", all[0].ID)
	}
}

func TestFilterByKind(t *testing.T) {
	l := changelog.New(20)
	l.Record(changelog.KindConfig, "svc-a", "admin", "msg", nil)
	l.Record(changelog.KindSilence, "svc-b", "ops", "msg", nil)
	l.Record(changelog.KindConfig, "svc-c", "admin", "msg", nil)

	results := l.FilterByKind(changelog.KindConfig)
	if len(results) != 2 {
		t.Fatalf("expected 2 config entries, got %d", len(results))
	}
	for _, e := range results {
		if e.Kind != changelog.KindConfig {
			t.Fatalf("unexpected kind in filtered result: %s", e.Kind)
		}
	}
}

func TestFilterByTarget(t *testing.T) {
	l := changelog.New(20)
	l.Record(changelog.KindAck, "svc-x", "dev", "acked", nil)
	l.Record(changelog.KindAck, "svc-y", "dev", "acked", nil)
	l.Record(changelog.KindMaintenance, "svc-x", "ops", "maint", nil)

	results := l.FilterByTarget("svc-x")
	if len(results) != 2 {
		t.Fatalf("expected 2 entries for svc-x, got %d", len(results))
	}
}

func TestFilterByTarget_Unknown(t *testing.T) {
	l := changelog.New(10)
	l.Record(changelog.KindOncall, "svc-a", "bot", "rotation updated", nil)

	results := l.FilterByTarget("nonexistent")
	if results != nil && len(results) != 0 {
		t.Fatalf("expected no results for unknown target, got %d", len(results))
	}
}

func TestRecord_IDMonotonicallyIncreases(t *testing.T) {
	l := changelog.New(50)
	for i := 0; i < 10; i++ {
		l.Record(changelog.KindConfig, "svc", "admin", "change", nil)
	}
	all := l.All()
	for i := 1; i < len(all); i++ {
		if all[i].ID <= all[i-1].ID {
			t.Fatalf("IDs not monotonically increasing at index %d: %d <= %d", i, all[i].ID, all[i-1].ID)
		}
	}
}
