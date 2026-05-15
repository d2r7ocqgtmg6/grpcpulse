// Package runbook associates targets with runbook URLs that operators
// can consult when a health-check alert fires.
package runbook

import (
	"errors"
	"net/url"
	"sync"
)

// ErrInvalidURL is returned when the supplied runbook URL cannot be parsed.
var ErrInvalidURL = errors.New("runbook: invalid URL")

// Registry stores a runbook URL per target address.
type Registry struct {
	mu      sync.RWMutex
	entries map[string]string
}

// New returns an empty Registry.
func New() *Registry {
	return &Registry{entries: make(map[string]string)}
}

// Set associates target with the given rawURL.
// Returns ErrInvalidURL if rawURL is not a valid absolute URL.
func (r *Registry) Set(target, rawURL string) error {
	u, err := url.ParseRequestURI(rawURL)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return ErrInvalidURL
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.entries[target] = rawURL
	return nil
}

// Get returns the runbook URL for target and whether it was found.
func (r *Registry) Get(target string) (string, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	v, ok := r.entries[target]
	return v, ok
}

// Delete removes the runbook entry for target.
func (r *Registry) Delete(target string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.entries, target)
}

// All returns a snapshot of all target → URL mappings.
func (r *Registry) All() map[string]string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make(map[string]string, len(r.entries))
	for k, v := range r.entries {
		out[k] = v
	}
	return out
}
