// Package history maintains a rolling window of check results
// for each target, enabling trend analysis and status pages.
package history

import (
	"sync"
	"time"
)

// Entry represents a single recorded check result.
type Entry struct {
	Timestamp time.Time
	Healthy   bool
	Latency   time.Duration
}

// History stores recent check entries per target with a bounded capacity.
type History struct {
	mu       sync.RWMutex
	records  map[string][]Entry
	capacity int
}

// New creates a History that retains at most capacity entries per target.
func New(capacity int) *History {
	if capacity <= 0 {
		capacity = 100
	}
	return &History{
		records:  make(map[string][]Entry),
		capacity: capacity,
	}
}

// Record appends an entry for the given target, evicting the oldest if full.
func (h *History) Record(target string, healthy bool, latency time.Duration) {
	h.mu.Lock()
	defer h.mu.Unlock()

	e := Entry{Timestamp: time.Now(), Healthy: healthy, Latency: latency}
	buf := h.records[target]
	if len(buf) >= h.capacity {
		buf = buf[1:]
	}
	h.records[target] = append(buf, e)
}

// Get returns a copy of all entries for the given target.
func (h *History) Get(target string) []Entry {
	h.mu.RLock()
	defer h.mu.RUnlock()

	src := h.records[target]
	out := make([]Entry, len(src))
	copy(out, src)
	return out
}

// Targets returns all known target names.
func (h *History) Targets() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	keys := make([]string, 0, len(h.records))
	for k := range h.records {
		keys = append(keys, k)
	}
	return keys
}

// UptimePercent returns the percentage of healthy entries for a target.
// Returns 0 if no entries exist.
func (h *History) UptimePercent(target string) float64 {
	h.mu.RLock()
	defer h.mu.RUnlock()

	entries := h.records[target]
	if len(entries) == 0 {
		return 0
	}
	var healthy int
	for _, e := range entries {
		if e.Healthy {
			healthy++
		}
	}
	return float64(healthy) / float64(len(entries)) * 100
}
