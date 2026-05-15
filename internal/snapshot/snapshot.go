// Package snapshot captures a point-in-time view of all target health states,
// useful for exporting a consistent summary to external systems or dashboards.
package snapshot

import (
	"sync"
	"time"
)

// TargetSnapshot holds the health state of a single target at a point in time.
type TargetSnapshot struct {
	Target    string            `json:"target"`
	Healthy   bool              `json:"healthy"`
	UptimePct float64           `json:"uptime_pct"`
	Tags      map[string]string `json:"tags,omitempty"`
	CapturedAt time.Time        `json:"captured_at"`
}

// Snapshot is an immutable collection of target states captured at a moment.
type Snapshot struct {
	Targets    []TargetSnapshot `json:"targets"`
	CapturedAt time.Time        `json:"captured_at"`
}

// HealthSource provides current health state for all known targets.
type HealthSource interface {
	Targets() []string
	UptimePercent(target string) float64
	Healthy(target string) bool
}

// TagSource provides tags for a target.
type TagSource interface {
	Get(target string) map[string]string
}

// Registry stores the most recent snapshot and allows retrieval.
type Registry struct {
	mu       sync.RWMutex
	latest   *Snapshot
	health   HealthSource
	tags     TagSource
}

// New creates a new snapshot Registry backed by the given sources.
func New(health HealthSource, tags TagSource) *Registry {
	return &Registry{health: health, tags: tags}
}

// Capture builds a new snapshot from the current state of all sources.
func (r *Registry) Capture() Snapshot {
	now := time.Now().UTC()
	targets := r.health.Targets()
	snaps := make([]TargetSnapshot, 0, len(targets))

	for _, t := range targets {
		ts := TargetSnapshot{
			Target:     t,
			Healthy:    r.health.Healthy(t),
			UptimePct:  r.health.UptimePercent(t),
			CapturedAt: now,
		}
		if r.tags != nil {
			ts.Tags = r.tags.Get(t)
		}
		snaps = append(snaps, ts)
	}

	snap := Snapshot{Targets: snaps, CapturedAt: now}

	r.mu.Lock()
	r.latest = &snap
	r.mu.Unlock()

	return snap
}

// Latest returns the most recently captured snapshot, or nil if none exists.
func (r *Registry) Latest() *Snapshot {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.latest
}
