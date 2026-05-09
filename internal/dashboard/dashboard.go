// Package dashboard provides an aggregated status view combining
// metrics, history, and notifier state for all monitored targets.
package dashboard

import (
	"time"

	"github.com/grpcpulse/internal/history"
	"github.com/grpcpulse/internal/notifier"
)

// TargetSummary holds the aggregated status for a single gRPC target.
type TargetSummary struct {
	Address      string         `json:"address"`
	CurrentState notifier.State `json:"current_state"`
	UptimePercent float64        `json:"uptime_percent"`
	LastChecked   time.Time      `json:"last_checked"`
	LastLatencyMs float64        `json:"last_latency_ms"`
	TotalChecks   int            `json:"total_checks"`
}

// Snapshot is the full dashboard payload returned to callers.
type Snapshot struct {
	GeneratedAt time.Time       `json:"generated_at"`
	Targets     []TargetSummary `json:"targets"`
}

// Builder assembles a Snapshot from its dependencies.
type Builder struct {
	history  *history.History
	notifier *notifier.Notifier
}

// New returns a Builder wired to the provided history and notifier.
func New(h *history.History, n *notifier.Notifier) *Builder {
	return &Builder{history: h, notifier: n}
}

// Build produces a fresh Snapshot.
func (b *Builder) Build() Snapshot {
	targets := b.history.Targets()
	summaries := make([]TargetSummary, 0, len(targets))

	for _, addr := range targets {
		entries := b.history.Get(addr)
		if len(entries) == 0 {
			continue
		}

		last := entries[len(entries)-1]
		ts := TargetSummary{
			Address:       addr,
			CurrentState:  b.notifier.Current(addr),
			UptimePercent: b.history.UptimePercent(addr),
			LastChecked:   last.CheckedAt,
			LastLatencyMs: float64(last.Latency.Milliseconds()),
			TotalChecks:   len(entries),
		}
		summaries = append(summaries, ts)
	}

	return Snapshot{
		GeneratedAt: time.Now().UTC(),
		Targets:     summaries,
	}
}
