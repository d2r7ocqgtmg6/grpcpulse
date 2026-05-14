package notifier

import (
	"sync"
	"time"
)

// State represents the health state of a target.
type State int

const (
	StateUnknown   State = iota
	StateHealthy         // service is reachable and healthy
	StateUnhealthy       // service is unreachable or unhealthy
)

// String returns a human-readable label for the state.
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

// Event records a state transition for a target.
type Event struct {
	Target string    `json:"target"`
	From   State     `json:"from"`
	To     State     `json:"to"`
	At     time.Time `json:"at"`
}

// Notifier tracks per-target health states and emits transition events.
type Notifier struct {
	mu      sync.RWMutex
	states  map[string]State
	events  []Event
	cap     int
	onAlert func(Event)
}

// New creates a Notifier with the given event capacity and optional alert callback.
func New(capacity int, onAlert func(Event)) *Notifier {
	if capacity <= 0 {
		capacity = 256
	}
	if onAlert == nil {
		onAlert = func(Event) {}
	}
	return &Notifier{
		states:  make(map[string]State),
		events:  make([]Event, 0, capacity),
		cap:     capacity,
		onAlert: onAlert,
	}
}

// Observe records a new observed state for target. If the state differs from
// the previous one a transition Event is appended and onAlert is called.
func (n *Notifier) Observe(target string, s State) {
	n.mu.Lock()
	defer n.mu.Unlock()

	prev, ok := n.states[target]
	n.states[target] = s

	if !ok || prev == s {
		return
	}

	ev := Event{Target: target, From: prev, To: s, At: time.Now()}
	if len(n.events) >= n.cap {
		n.events = n.events[1:]
	}
	n.events = append(n.events, ev)
	n.onAlert(ev)
}

// Current returns a snapshot of the latest state per target.
func (n *Notifier) Current() map[string]State {
	n.mu.RLock()
	defer n.mu.RUnlock()
	out := make(map[string]State, len(n.states))
	for k, v := range n.states {
		out[k] = v
	}
	return out
}

// Events returns a copy of the recorded transition events.
func (n *Notifier) Events() []Event {
	n.mu.RLock()
	defer n.mu.RUnlock()
	out := make([]Event, len(n.events))
	copy(out, n.events)
	return out
}
