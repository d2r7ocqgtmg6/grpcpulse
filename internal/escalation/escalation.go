// Package escalation manages alert escalation policies for gRPC health targets.
// When a target remains unhealthy beyond a configurable threshold, escalation
// records track which policy tier has been reached.
package escalation

import (
	"errors"
	"sync"
	"time"
)

// Level represents an escalation tier.
type Level int

const (
	LevelNone   Level = 0
	LevelWarn   Level = 1
	LevelCritical Level = 2
	LevelPage   Level = 3
)

func (l Level) String() string {
	switch l {
	case LevelWarn:
		return "warn"
	case LevelCritical:
		return "critical"
	case LevelPage:
		return "page"
	default:
		return "none"
	}
}

// Policy defines thresholds for escalation levels.
type Policy struct {
	WarnAfter     time.Duration
	CriticalAfter time.Duration
	PageAfter     time.Duration
}

var DefaultPolicy = Policy{
	WarnAfter:     2 * time.Minute,
	CriticalAfter: 5 * time.Minute,
	PageAfter:     15 * time.Minute,
}

// entry tracks when a target first became unhealthy.
type entry struct {
	unhealthySince time.Time
	level          Level
}

// Registry tracks escalation state per target.
type Registry struct {
	mu     sync.Mutex
	policy Policy
	state  map[string]*entry
}

// New returns a Registry using the given policy.
func New(p Policy) *Registry {
	return &Registry{
		policy: p,
		state:  make(map[string]*entry),
	}
}

// Observe records that target is unhealthy at now and returns the current Level.
// Call with healthy=true to clear the escalation state.
func (r *Registry) Observe(target string, healthy bool, now time.Time) Level {
	r.mu.Lock()
	defer r.mu.Unlock()

	if healthy {
		delete(r.state, target)
		return LevelNone
	}

	e, ok := r.state[target]
	if !ok {
		e = &entry{unhealthySince: now}
		r.state[target] = e
	}

	dur := now.Sub(e.unhealthySince)
	switch {
	case dur >= r.policy.PageAfter:
		e.level = LevelPage
	case dur >= r.policy.CriticalAfter:
		e.level = LevelCritical
	case dur >= r.policy.WarnAfter:
		e.level = LevelWarn
	default:
		e.level = LevelNone
	}
	return e.level
}

// Current returns the current escalation level for a target, or LevelNone.
func (r *Registry) Current(target string) Level {
	r.mu.Lock()
	defer r.mu.Unlock()
	if e, ok := r.state[target]; ok {
		return e.level
	}
	return LevelNone
}

// UnhealthySince returns when the target first became unhealthy.
func (r *Registry) UnhealthySince(target string) (time.Time, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if e, ok := r.state[target]; ok {
		return e.unhealthySince, nil
	}
	return time.Time{}, errors.New("escalation: target not found")
}
