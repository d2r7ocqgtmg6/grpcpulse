// Package sla tracks per-target SLA budgets based on uptime history.
package sla

import (
	"fmt"
	"sync"
	"time"
)

// Budget defines the SLA target for a single gRPC endpoint.
type Budget struct {
	Target    string
	Objective float64 // e.g. 99.9 means 99.9% uptime
	Window    time.Duration
}

// Status summarises the current SLA standing for a target.
type Status struct {
	Target        string
	Objective     float64
	ActualUptime  float64
	ErrorBudget   float64 // minutes of downtime remaining
	Breached      bool
}

// String returns a human-readable representation.
func (s Status) String() string {
	return fmt.Sprintf("%s: %.2f%% (objective %.2f%%, budget %.1fm remaining, breached=%v)",
		s.Target, s.ActualUptime, s.Objective, s.ErrorBudget, s.Breached)
}

// UptimeSource is satisfied by history.Log.
type UptimeSource interface {
	UptimePercent(target string, window time.Duration) float64
}

// Registry holds SLA budgets and evaluates them against live uptime data.
type Registry struct {
	mu      sync.RWMutex
	budgets map[string]Budget
	source  UptimeSource
}

// New creates a Registry backed by the given UptimeSource.
func New(source UptimeSource) *Registry {
	return &Registry{
		budgets: make(map[string]Budget),
		source:  source,
	}
}

// Set registers or replaces the SLA budget for a target.
func (r *Registry) Set(b Budget) error {
	if b.Objective <= 0 || b.Objective > 100 {
		return fmt.Errorf("sla: objective must be in (0, 100], got %.2f", b.Objective)
	}
	if b.Window <= 0 {
		return fmt.Errorf("sla: window must be positive")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.budgets[b.Target] = b
	return nil
}

// Delete removes the SLA budget for a target.
func (r *Registry) Delete(target string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.budgets, target)
}

// Evaluate returns the current SLA Status for a target.
// Returns an error if no budget is registered.
func (r *Registry) Evaluate(target string) (Status, error) {
	r.mu.RLock()
	b, ok := r.budgets[target]
	r.mu.RUnlock()
	if !ok {
		return Status{}, fmt.Errorf("sla: no budget registered for %q", target)
	}

	actual := r.source.UptimePercent(target, b.Window)
	allowedDowntimePct := 100.0 - b.Objective
	actualDowntimePct := 100.0 - actual
	remainingBudgetPct := allowedDowntimePct - actualDowntimePct
	remainingMinutes := remainingBudgetPct / 100.0 * b.Window.Minutes()

	return Status{
		Target:       target,
		Objective:    b.Objective,
		ActualUptime: actual,
		ErrorBudget:  remainingMinutes,
		Breached:     actual < b.Objective,
	}, nil
}

// All returns the status for every registered target.
func (r *Registry) All() []Status {
	r.mu.RLock()
	keys := make([]string, 0, len(r.budgets))
	for k := range r.budgets {
		keys = append(keys, k)
	}
	r.mu.RUnlock()

	out := make([]Status, 0, len(keys))
	for _, k := range keys {
		if s, err := r.Evaluate(k); err == nil {
			out = append(out, s)
		}
	}
	return out
}
