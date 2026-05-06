// Package notifier provides alerting capabilities when gRPC services
// transition between healthy and unhealthy states.
package notifier

import (
	"fmt"
	"log/slog"
	"sync"
)

// State represents the health state of a target.
type State int

const (
	StateUnknown   State = iota
	StateHealthy
	StateUnhealthy
)

func (s State) String() string {
	switch s {
	case StateHealthy:
		return "healthy"
	case StateUnhealthy:
		return "unhealthy"
	default:
		return "unknown"
	}
}

// AlertFunc is called when a state transition occurs.
type AlertFunc func(target string, previous, current State)

// Notifier tracks per-target health states and fires AlertFunc on transitions.
type Notifier struct {
	mu     sync.Mutex
	states map[string]State
	alert  AlertFunc
	log    *slog.Logger
}

// New creates a Notifier with the given alert callback.
// If alert is nil a log-only handler is used.
func New(alert AlertFunc, log *slog.Logger) *Notifier {
	if log == nil {
		log = slog.Default()
	}
	if alert == nil {
		alert = func(target string, prev, cur State) {
			log.Warn("state transition",
				"target", target,
				"from", prev.String(),
				"to", cur.String())
		}
	}
	return &Notifier{
		states: make(map[string]State),
		alert:  alert,
		log:    log,
	}
}

// Observe records the current health for target and fires the alert on change.
func (n *Notifier) Observe(target string, healthy bool) {
	current := StateUnhealthy
	if healthy {
		current = StateHealthy
	}

	n.mu.Lock()
	prev, exists := n.states[target]
	n.states[target] = current
	n.mu.Unlock()

	if !exists || prev != current {
		n.log.Info("health state changed",
			"target", target,
			"previous", prev.String(),
			"current", current.String())
		n.alert(target, prev, current)
	}
}

// Current returns the last known state for target, or StateUnknown.
func (n *Notifier) Current(target string) State {
	n.mu.Lock()
	defer n.mu.Unlock()
	s, ok := n.states[target]
	if !ok {
		return StateUnknown
	}
	return s
}

// Summary returns a human-readable summary of all tracked targets.
func (n *Notifier) Summary() string {
	n.mu.Lock()
	defer n.mu.Unlock()
	out := ""
	for t, s := range n.states {
		out += fmt.Sprintf("%s=%s ", t, s.String())
	}
	return out
}
