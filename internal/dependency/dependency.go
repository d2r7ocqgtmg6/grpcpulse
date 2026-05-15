// Package dependency tracks inter-service dependency relationships,
// allowing grpcpulse to propagate health status across dependent targets.
package dependency

import (
	"errors"
	"sync"
)

// ErrCycle is returned when adding a dependency would create a cycle.
var ErrCycle = errors.New("dependency: cycle detected")

// Registry holds directed dependency edges between targets.
// If target A depends on target B, an unhealthy B may affect A.
type Registry struct {
	mu   sync.RWMutex
	edges map[string]map[string]struct{} // edges[a] = {b: ...} means a depends on b
}

// New returns an empty Registry.
func New() *Registry {
	return &Registry{
		edges: make(map[string]map[string]struct{}),
	}
}

// Add records that `from` depends on `to`.
// Returns ErrCycle if the edge would introduce a cycle.
func (r *Registry) Add(from, to string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.wouldCycle(from, to) {
		return ErrCycle
	}
	if r.edges[from] == nil {
		r.edges[from] = make(map[string]struct{})
	}
	r.edges[from][to] = struct{}{}
	return nil
}

// Remove deletes the dependency edge from -> to.
func (r *Registry) Remove(from, to string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.edges[from], to)
}

// DependsOn returns all targets that `target` directly depends on.
func (r *Registry) DependsOn(target string) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]string, 0, len(r.edges[target]))
	for dep := range r.edges[target] {
		out = append(out, dep)
	}
	return out
}

// Dependents returns all targets that directly depend on `target`.
func (r *Registry) Dependents(target string) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []string
	for src, deps := range r.edges {
		if _, ok := deps[target]; ok {
			out = append(out, src)
		}
	}
	return out
}

// wouldCycle checks whether adding from->to would create a cycle (caller holds lock).
func (r *Registry) wouldCycle(from, to string) bool {
	// DFS from `to`; if we can reach `from`, it's a cycle.
	visited := make(map[string]bool)
	var dfs func(node string) bool
	dfs = func(node string) bool {
		if node == from {
			return true
		}
		if visited[node] {
			return false
		}
		visited[node] = true
		for dep := range r.edges[node] {
			if dfs(dep) {
				return true
			}
		}
		return false
	}
	return dfs(to)
}
