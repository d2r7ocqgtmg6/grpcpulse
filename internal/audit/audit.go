// Package audit provides an append-only log of configuration and state
// change events for grpcpulse targets.
package audit

import (
	"sync"
	"time"
)

// EventKind classifies the type of audit event.
type EventKind string

const (
	EventKindSilenceAdded   EventKind = "silence_added"
	EventKindSilenceLifted  EventKind = "silence_lifted"
	EventKindTagSet         EventKind = "tag_set"
	EventKindCircuitOpen    EventKind = "circuit_open"
	EventKindCircuitClosed  EventKind = "circuit_closed"
	EventKindConfigReloaded EventKind = "config_reloaded"
)

// Event represents a single auditable action.
type Event struct {
	Timestamp time.Time `json:"timestamp"`
	Target    string    `json:"target,omitempty"`
	Kind      EventKind `json:"kind"`
	Message   string    `json:"message"`
}

// Log is a bounded, thread-safe audit log.
type Log struct {
	mu       sync.RWMutex
	events   []Event
	capacity int
}

// New creates a new Log with the given maximum capacity.
// When capacity is exceeded the oldest entries are evicted.
func New(capacity int) *Log {
	if capacity <= 0 {
		capacity = 256
	}
	return &Log{capacity: capacity}
}

// Record appends a new event to the log.
func (l *Log) Record(kind EventKind, target, message string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	e := Event{
		Timestamp: time.Now().UTC(),
		Target:    target,
		Kind:      kind,
		Message:   message,
	}
	l.events = append(l.events, e)
	if len(l.events) > l.capacity {
		l.events = l.events[len(l.events)-l.capacity:]
	}
}

// All returns a copy of all stored events, oldest first.
func (l *Log) All() []Event {
	l.mu.RLock()
	defer l.mu.RUnlock()
	out := make([]Event, len(l.events))
	copy(out, l.events)
	return out
}

// FilterByTarget returns events for a specific target.
func (l *Log) FilterByTarget(target string) []Event {
	l.mu.RLock()
	defer l.mu.RUnlock()
	var out []Event
	for _, e := range l.events {
		if e.Target == target {
			out = append(out, e)
		}
	}
	return out
}

// Len returns the current number of stored events.
func (l *Log) Len() int {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return len(l.events)
}
