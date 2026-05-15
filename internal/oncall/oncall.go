// Package oncall manages on-call rotation schedules for alert routing.
package oncall

import (
	"errors"
	"sync"
	"time"
)

// Rotation represents a named on-call rotation with an ordered list of
// responders and a shift duration.
type Rotation struct {
	Name      string
	Responders []string
	ShiftDuration time.Duration
	StartedAt time.Time
}

// CurrentResponder returns the responder who is currently on-call based on
// elapsed time since the rotation started.
func (r *Rotation) CurrentResponder(now time.Time) string {
	if len(r.Responders) == 0 {
		return ""
	}
	elapsed := now.Sub(r.StartedAt)
	if elapsed < 0 {
		elapsed = 0
	}
	shifts := int(elapsed / r.ShiftDuration)
	return r.Responders[shifts%len(r.Responders)]
}

// Registry holds all named on-call rotations.
type Registry struct {
	mu        sync.RWMutex
	rotations map[string]*Rotation
}

// New returns an initialised Registry.
func New() *Registry {
	return &Registry{rotations: make(map[string]*Rotation)}
}

// Set registers or replaces a rotation by name.
func (reg *Registry) Set(r Rotation) error {
	if r.Name == "" {
		return errors.New("rotation name must not be empty")
	}
	if len(r.Responders) == 0 {
		return errors.New("rotation must have at least one responder")
	}
	if r.ShiftDuration <= 0 {
		return errors.New("shift duration must be positive")
	}
	reg.mu.Lock()
	defer reg.mu.Unlock()
	copy := r
	reg.rotations[r.Name] = &copy
	return nil
}

// Get returns the rotation registered under name, or false if absent.
func (reg *Registry) Get(name string) (Rotation, bool) {
	reg.mu.RLock()
	defer reg.mu.RUnlock()
	r, ok := reg.rotations[name]
	if !ok {
		return Rotation{}, false
	}
	return *r, true
}

// Delete removes a rotation by name.
func (reg *Registry) Delete(name string) {
	reg.mu.Lock()
	defer reg.mu.Unlock()
	delete(reg.rotations, name)
}

// All returns a snapshot of every registered rotation.
func (reg *Registry) All() []Rotation {
	reg.mu.RLock()
	defer reg.mu.RUnlock()
	out := make([]Rotation, 0, len(reg.rotations))
	for _, r := range reg.rotations {
		out = append(out, *r)
	}
	return out
}
