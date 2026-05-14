// Package silencer provides time-based alert suppression for gRPC health checks.
// A silenced target will not trigger notifications during the silence window.
package silencer

import (
	"sync"
	"time"
)

// Silence represents an active suppression window for a target.
type Silence struct {
	Target    string
	Reason    string
	ExpiresAt time.Time
}

// Expired reports whether the silence window has passed.
func (s Silence) Expired() bool {
	return time.Now().After(s.ExpiresAt)
}

// Registry holds active silences keyed by target address.
type Registry struct {
	mu      sync.RWMutex
	silences map[string]Silence
	now      func() time.Time
}

// New returns a new Registry.
func New() *Registry {
	return &Registry{
		silences: make(map[string]Silence),
		now:      time.Now,
	}
}

// Silence adds or replaces a silence for the given target.
func (r *Registry) Silence(target, reason string, duration time.Duration) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.silences[target] = Silence{
		Target:    target,
		Reason:    reason,
		ExpiresAt: r.now().Add(duration),
	}
}

// IsSilenced reports whether the target currently has an active silence.
func (r *Registry) IsSilenced(target string) bool {
	r.mu.RLock()
	s, ok := r.silences[target]
	r.mu.RUnlock()
	if !ok {
		return false
	}
	if s.Expired() {
		r.mu.Lock()
		delete(r.silences, target)
		r.mu.Unlock()
		return false
	}
	return true
}

// Lift removes a silence for the given target before it expires.
func (r *Registry) Lift(target string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.silences, target)
}

// Active returns a snapshot of all non-expired silences.
func (r *Registry) Active() []Silence {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]Silence, 0, len(r.silences))
	for k, s := range r.silences {
		if s.Expired() {
			delete(r.silences, k)
			continue
		}
		out = append(out, s)
	}
	return out
}
