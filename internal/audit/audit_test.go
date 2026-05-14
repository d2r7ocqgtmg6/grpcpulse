package audit_test

import (
	"testing"

	"github.com/yourorg/grpcpulse/internal/audit"
)

func TestRecord_AppendsEvent(t *testing.T) {
	l := audit.New(10)
	l.Record(audit.EventKindTagSet, "svc-a", "set env=prod")
	if l.Len() != 1 {
		t.Fatalf("expected 1 event, got %d", l.Len())
	}
	events := l.All()
	if events[0].Kind != audit.EventKindTagSet {
		t.Errorf("unexpected kind: %s", events[0].Kind)
	}
	if events[0].Target != "svc-a" {
		t.Errorf("unexpected target: %s", events[0].Target)
	}
}

func TestCapacity_Eviction(t *testing.T) {
	l := audit.New(3)
	for i := 0; i < 5; i++ {
		l.Record(audit.EventKindConfigReloaded, "", "reload")
	}
	if l.Len() != 3 {
		t.Fatalf("expected capacity 3, got %d", l.Len())
	}
}

func TestAll_ReturnsCopy(t *testing.T) {
	l := audit.New(10)
	l.Record(audit.EventKindSilenceAdded, "svc-b", "silenced for 1h")
	a := l.All()
	a[0].Target = "mutated"
	original := l.All()
	if original[0].Target == "mutated" {
		t.Error("All() should return an independent copy")
	}
}

func TestFilterByTarget(t *testing.T) {
	l := audit.New(20)
	l.Record(audit.EventKindTagSet, "svc-a", "tag a")
	l.Record(audit.EventKindTagSet, "svc-b", "tag b")
	l.Record(audit.EventKindCircuitOpen, "svc-a", "opened")

	results := l.FilterByTarget("svc-a")
	if len(results) != 2 {
		t.Fatalf("expected 2 events for svc-a, got %d", len(results))
	}
	for _, e := range results {
		if e.Target != "svc-a" {
			t.Errorf("unexpected target in filter result: %s", e.Target)
		}
	}
}

func TestFilterByTarget_Unknown(t *testing.T) {
	l := audit.New(10)
	l.Record(audit.EventKindSilenceLifted, "svc-a", "lifted")
	if got := l.FilterByTarget("unknown"); len(got) != 0 {
		t.Errorf("expected empty slice, got %v", got)
	}
}

func TestNew_DefaultCapacity(t *testing.T) {
	l := audit.New(0)
	for i := 0; i < 300; i++ {
		l.Record(audit.EventKindConfigReloaded, "", "r")
	}
	if l.Len() != 256 {
		t.Errorf("expected default capacity 256, got %d", l.Len())
	}
}
