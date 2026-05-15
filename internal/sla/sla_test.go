package sla_test

import (
	"testing"
	"time"

	"github.com/yourorg/grpcpulse/internal/sla"
)

// stubSource implements UptimeSource for testing.
type stubSource struct {
	values map[string]float64
}

func (s *stubSource) UptimePercent(target string, _ time.Duration) float64 {
	return s.values[target]
}

func newRegistry(values map[string]float64) *sla.Registry {
	return sla.New(&stubSource{values: values})
}

func TestSet_And_Evaluate(t *testing.T) {
	reg := newRegistry(map[string]float64{"svc:50051": 99.5})
	err := reg.Set(sla.Budget{Target: "svc:50051", Objective: 99.9, Window: 30 * 24 * time.Hour})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	status, err := reg.Evaluate("svc:50051")
	if err != nil {
		t.Fatalf("evaluate error: %v", err)
	}
	if !status.Breached {
		t.Error("expected SLA breach but got none")
	}
	if status.ActualUptime != 99.5 {
		t.Errorf("expected 99.5 actual uptime, got %.2f", status.ActualUptime)
	}
}

func TestEvaluate_NotBreached(t *testing.T) {
	reg := newRegistry(map[string]float64{"svc:50051": 99.95})
	_ = reg.Set(sla.Budget{Target: "svc:50051", Objective: 99.9, Window: 30 * 24 * time.Hour})

	status, err := reg.Evaluate("svc:50051")
	if err != nil {
		t.Fatalf("evaluate error: %v", err)
	}
	if status.Breached {
		t.Error("expected no breach but got one")
	}
	if status.ErrorBudget <= 0 {
		t.Errorf("expected positive error budget, got %.2f", status.ErrorBudget)
	}
}

func TestEvaluate_UnknownTarget(t *testing.T) {
	reg := newRegistry(nil)
	_, err := reg.Evaluate("unknown:9000")
	if err == nil {
		t.Fatal("expected error for unknown target")
	}
}

func TestSet_InvalidObjective(t *testing.T) {
	reg := newRegistry(nil)
	if err := reg.Set(sla.Budget{Target: "x", Objective: 0, Window: time.Hour}); err == nil {
		t.Error("expected error for zero objective")
	}
	if err := reg.Set(sla.Budget{Target: "x", Objective: 101, Window: time.Hour}); err == nil {
		t.Error("expected error for objective > 100")
	}
}

func TestDelete_RemovesBudget(t *testing.T) {
	reg := newRegistry(map[string]float64{"svc:50051": 100})
	_ = reg.Set(sla.Budget{Target: "svc:50051", Objective: 99.9, Window: time.Hour})
	reg.Delete("svc:50051")
	_, err := reg.Evaluate("svc:50051")
	if err == nil {
		t.Fatal("expected error after delete")
	}
}

func TestAll_ReturnsAllStatuses(t *testing.T) {
	reg := newRegistry(map[string]float64{"a:1": 100, "b:2": 98})
	_ = reg.Set(sla.Budget{Target: "a:1", Objective: 99.9, Window: time.Hour})
	_ = reg.Set(sla.Budget{Target: "b:2", Objective: 99.9, Window: time.Hour})

	all := reg.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 statuses, got %d", len(all))
	}
}

func TestStatus_String(t *testing.T) {
	s := sla.Status{Target: "svc:50051", Objective: 99.9, ActualUptime: 99.5, ErrorBudget: -2.88, Breached: true}
	got := s.String()
	if got == "" {
		t.Error("expected non-empty string")
	}
}
