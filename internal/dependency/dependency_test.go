package dependency_test

import (
	"testing"

	"github.com/example/grpcpulse/internal/dependency"
)

func TestAdd_And_DependsOn(t *testing.T) {
	r := dependency.New()
	if err := r.Add("a", "b"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	deps := r.DependsOn("a")
	if len(deps) != 1 || deps[0] != "b" {
		t.Fatalf("expected [b], got %v", deps)
	}
}

func TestDependents(t *testing.T) {
	r := dependency.New()
	_ = r.Add("a", "c")
	_ = r.Add("b", "c")
	dependents := r.Dependents("c")
	if len(dependents) != 2 {
		t.Fatalf("expected 2 dependents, got %d", len(dependents))
	}
}

func TestAdd_CycleDetected(t *testing.T) {
	r := dependency.New()
	_ = r.Add("a", "b")
	_ = r.Add("b", "c")
	if err := r.Add("c", "a"); err != dependency.ErrCycle {
		t.Fatalf("expected ErrCycle, got %v", err)
	}
}

func TestAdd_SelfCycle(t *testing.T) {
	r := dependency.New()
	if err := r.Add("a", "a"); err != dependency.ErrCycle {
		t.Fatalf("expected ErrCycle for self-loop, got %v", err)
	}
}

func TestRemove_Edge(t *testing.T) {
	r := dependency.New()
	_ = r.Add("a", "b")
	r.Remove("a", "b")
	if deps := r.DependsOn("a"); len(deps) != 0 {
		t.Fatalf("expected no deps after remove, got %v", deps)
	}
}

func TestDependsOn_UnknownTarget(t *testing.T) {
	r := dependency.New()
	if deps := r.DependsOn("unknown"); len(deps) != 0 {
		t.Fatalf("expected empty slice, got %v", deps)
	}
}

func TestDependents_UnknownTarget(t *testing.T) {
	r := dependency.New()
	if d := r.Dependents("ghost"); len(d) != 0 {
		t.Fatalf("expected empty, got %v", d)
	}
}
