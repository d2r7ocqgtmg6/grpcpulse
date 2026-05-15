// Package changelog tracks configuration and state change events for auditing
// and observability purposes, providing a time-ordered log of what changed,
// when, and by whom.
package changelog

import (
	"sync"
	"time"
)

// EntryKind describes the category of a changelog entry.
type EntryKind string

const (
	KindConfig      EntryKind = "config"
	KindSilence     EntryKind = "silence"
	KindAck         EntryKind = "ack"
	KindMaintenance EntryKind = "maintenance"
	KindRunbook     EntryKind = "runbook"
	KindOncall      EntryKind = "oncall"
)

// Entry represents a single recorded change.
type Entry struct {
	ID        uint64            `json:"id"`
	Timestamp time.Time         `json:"timestamp"`
	Kind      EntryKind         `json:"kind"`
	Target    string            `json:"target,omitempty"`
	Actor     string            `json:"actor,omitempty"`
	Message   string            `json:"message"`
	Meta      map[string]string `json:"meta,omitempty"`
}

// Log is a bounded, thread-safe changelog.
type Log struct {
	mu       sync.RWMutex
	entries  []Entry
	capacity int
	next     uint64
}

// New creates a Log that retains at most capacity entries.
func New(capacity int) *Log {
	if capacity <= 0 {
		capacity = 500
	}
	return &Log{
		entries:  make([]Entry, 0, capacity),
		capacity: capacity,
	}
}

// Record appends a new entry to the log.
func (l *Log) Record(kind EntryKind, target, actor, message string, meta map[string]string) Entry {
	l.mu.Lock()
	defer l.mu.Unlock()

	e := Entry{
		ID:        l.next,
		Timestamp: time.Now().UTC(),
		Kind:      kind,
		Target:    target,
		Actor:     actor,
		Message:   message,
		Meta:      meta,
	}
	l.next++

	if len(l.entries) >= l.capacity {
		l.entries = l.entries[1:]
	}
	l.entries = append(l.entries, e)
	return e
}

// All returns a copy of all entries in insertion order.
func (l *Log) All() []Entry {
	l.mu.RLock()
	defer l.mu.RUnlock()
	out := make([]Entry, len(l.entries))
	copy(out, l.entries)
	return out
}

// FilterByKind returns entries whose Kind matches the given value.
func (l *Log) FilterByKind(kind EntryKind) []Entry {
	l.mu.RLock()
	defer l.mu.RUnlock()
	var out []Entry
	for _, e := range l.entries {
		if e.Kind == kind {
			out = append(out, e)
		}
	}
	return out
}

// FilterByTarget returns entries whose Target matches the given value.
func (l *Log) FilterByTarget(target string) []Entry {
	l.mu.RLock()
	defer l.mu.RUnlock()
	var out []Entry
	for _, e := range l.entries {
		if e.Target == target {
			out = append(out, e)
		}
	}
	return out
}
