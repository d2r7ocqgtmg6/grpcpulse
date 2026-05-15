// Package grouping provides logical grouping of monitored gRPC targets,
// allowing operators to organise services into named groups for aggregated
// status reporting and bulk operations.
package grouping

import (
	"errors"
	"sync"
)

// ErrGroupNotFound is returned when a requested group does not exist.
var ErrGroupNotFound = errors.New("group not found")

// ErrTargetNotInGroup is returned when a target is not a member of the group.
var ErrTargetNotInGroup = errors.New("target not in group")

// Registry manages named groups of targets.
type Registry struct {
	mu     sync.RWMutex
	groups map[string]map[string]struct{}
}

// New returns an empty Registry.
func New() *Registry {
	return &Registry{
		groups: make(map[string]map[string]struct{}),
	}
}

// Add creates the group if it does not exist and adds target to it.
func (r *Registry) Add(group, target string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.groups[group]; !ok {
		r.groups[group] = make(map[string]struct{})
	}
	r.groups[group][target] = struct{}{}
}

// Remove removes target from group. Returns ErrGroupNotFound or
// ErrTargetNotInGroup when the operation cannot be completed.
func (r *Registry) Remove(group, target string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	members, ok := r.groups[group]
	if !ok {
		return ErrGroupNotFound
	}
	if _, ok := members[target]; !ok {
		return ErrTargetNotInGroup
	}
	delete(members, target)
	if len(members) == 0 {
		delete(r.groups, group)
	}
	return nil
}

// Members returns a sorted copy of all targets in the given group.
// Returns ErrGroupNotFound if the group does not exist.
func (r *Registry) Members(group string) ([]string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	members, ok := r.groups[group]
	if !ok {
		return nil, ErrGroupNotFound
	}
	out := make([]string, 0, len(members))
	for t := range members {
		out = append(out, t)
	}
	return out, nil
}

// Groups returns the names of all currently defined groups.
func (r *Registry) Groups() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]string, 0, len(r.groups))
	for g := range r.groups {
		out = append(out, g)
	}
	return out
}
