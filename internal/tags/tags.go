// Package tags provides label/tag management for gRPC targets,
// allowing targets to be grouped and filtered by arbitrary key-value pairs.
package tags

import (
	"fmt"
	"sort"
	"strings"
	"sync"
)

// Tags is an immutable set of key-value label pairs attached to a target.
type Tags map[string]string

// String returns a deterministic, human-readable representation of the tags.
func (t Tags) String() string {
	if len(t) == 0 {
		return ""
	}
	keys := make([]string, 0, len(t))
	for k := range t {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%s", k, t[k]))
	}
	return strings.Join(parts, ",")
}

// Registry maps target addresses to their associated Tags.
type Registry struct {
	mu   sync.RWMutex
	data map[string]Tags
}

// New returns an initialised Registry.
func New() *Registry {
	return &Registry{data: make(map[string]Tags)}
}

// Set associates tags with the given target address, replacing any existing tags.
func (r *Registry) Set(target string, tags Tags) {
	copy := make(Tags, len(tags))
	for k, v := range tags {
		copy[k] = v
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[target] = copy
}

// Get returns the tags for a target. The second return value is false when
// the target has no registered tags.
func (r *Registry) Get(target string) (Tags, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	t, ok := r.data[target]
	return t, ok
}

// Filter returns all targets whose tags contain every key-value pair in the
// provided selector. An empty selector matches all registered targets.
func (r *Registry) Filter(selector Tags) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var matches []string
	for target, tags := range r.data {
		if matches_(tags, selector) {
			matches = append(matches, target)
		}
	}
	sort.Strings(matches)
	return matches
}

// Delete removes the tags for a target.
func (r *Registry) Delete(target string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.data, target)
}

func matches_(tags, selector Tags) bool {
	for k, v := range selector {
		if tags[k] != v {
			return false
		}
	}
	return true
}
