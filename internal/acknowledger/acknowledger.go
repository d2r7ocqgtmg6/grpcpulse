// Package acknowledger tracks manual acknowledgements of unhealthy targets,
// suppressing repeated alerts until the acknowledgement expires or is cleared.
package acknowledger

import (
	"sync"
	"time"
)

// Acknowledgement represents a single ack entry for a target.
type Acknowledgement struct {
	Target    string    `json:"target"`
	Reason    string    `json:"reason"`
	AckedBy   string    `json:"acked_by"`
	AckedAt   time.Time `json:"acked_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// IsExpired reports whether the acknowledgement has passed its expiry time.
func (a Acknowledgement) IsExpired(now time.Time) bool {
	return now.After(a.ExpiresAt)
}

// Acknowledger stores and queries target acknowledgements.
type Acknowledger struct {
	mu   sync.RWMutex
	acks map[string]Acknowledgement
}

// New returns an initialised Acknowledger.
func New() *Acknowledger {
	return &Acknowledger{acks: make(map[string]Acknowledgement)}
}

// Acknowledge records an acknowledgement for target, expiring after ttl.
func (a *Acknowledger) Acknowledge(target, reason, ackedBy string, ttl time.Duration) {
	a.mu.Lock()
	defer a.mu.Unlock()
	now := time.Now()
	a.acks[target] = Acknowledgement{
		Target:    target,
		Reason:    reason,
		AckedBy:   ackedBy,
		AckedAt:   now,
		ExpiresAt: now.Add(ttl),
	}
}

// IsAcknowledged reports whether target has a current (non-expired) ack.
func (a *Acknowledger) IsAcknowledged(target string) bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	ack, ok := a.acks[target]
	return ok && !ack.IsExpired(time.Now())
}

// Clear removes the acknowledgement for target.
func (a *Acknowledger) Clear(target string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	delete(a.acks, target)
}

// Active returns all non-expired acknowledgements.
func (a *Acknowledger) Active() []Acknowledgement {
	a.mu.RLock()
	defer a.mu.RUnlock()
	now := time.Now()
	out := make([]Acknowledgement, 0, len(a.acks))
	for _, ack := range a.acks {
		if !ack.IsExpired(now) {
			out = append(out, ack)
		}
	}
	return out
}
