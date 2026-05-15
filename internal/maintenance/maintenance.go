// Package maintenance provides scheduled maintenance window management.
// During a maintenance window, health-check alerts and notifications are
// suppressed without affecting metric collection.
package maintenance

import (
	"sync"
	"time"
)

// Window represents a scheduled maintenance period for a target.
type Window struct {
	Target    string
	Start     time.Time
	End       time.Time
	Reason    string
}

// IsActive reports whether the window is currently active.
func (w Window) IsActive(now time.Time) bool {
	return !now.Before(w.Start) && now.Before(w.End)
}

// Registry stores and queries maintenance windows.
type Registry struct {
	mu      sync.RWMutex
	windows map[string][]Window
	now     func() time.Time
}

// New returns a new Registry.
func New() *Registry {
	return &Registry{
		windows: make(map[string][]Window),
		now:     time.Now,
	}
}

// Add registers a maintenance window for the given target.
func (r *Registry) Add(w Window) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.windows[w.Target] = append(r.windows[w.Target], w)
}

// IsUnderMaintenance reports whether the target has an active window right now.
func (r *Registry) IsUnderMaintenance(target string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	now := r.now()
	for _, w := range r.windows[target] {
		if w.IsActive(now) {
			return true
		}
	}
	return false
}

// Active returns all currently active windows across all targets.
func (r *Registry) Active() []Window {
	r.mu.RLock()
	defer r.mu.RUnlock()
	now := r.now()
	var out []Window
	for _, ws := range r.windows {
		for _, w := range ws {
			if w.IsActive(now) {
				out = append(out, w)
			}
		}
	}
	return out
}

// Remove cancels all maintenance windows for the given target.
func (r *Registry) Remove(target string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.windows, target)
}

// Purge removes all expired windows from the registry.
func (r *Registry) Purge() {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := r.now()
	for target, ws := range r.windows {
		var kept []Window
		for _, w := range ws {
			if now.Before(w.End) {
				kept = append(kept, w)
			}
		}
		if len(kept) == 0 {
			delete(r.windows, target)
		} else {
			r.windows[target] = kept
		}
	}
}
