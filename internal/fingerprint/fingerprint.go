// Package fingerprint provides stable hashing for gRPC target identifiers,
// allowing consistent key generation across restarts and components.
package fingerprint

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
)

// Fingerprint is a stable, short identifier derived from a target address and tags.
type Fingerprint string

// String returns the string representation of the fingerprint.
func (f Fingerprint) String() string {
	return string(f)
}

// Of computes a deterministic fingerprint for the given target address and
// optional key=value labels. Label order does not affect the result.
//
//	Of("localhost:50051")                        // address only
//	Of("localhost:50051", "env=prod", "region=us") // address + labels
func Of(address string, labels ...string) Fingerprint {
	normalized := make([]string, len(labels))
	copy(normalized, labels)
	sort.Strings(normalized)

	raw := address
	if len(normalized) > 0 {
		raw = fmt.Sprintf("%s|%s", address, strings.Join(normalized, ","))
	}

	sum := sha256.Sum256([]byte(raw))
	return Fingerprint(hex.EncodeToString(sum[:])[:16])
}

// Registry maps target addresses to their computed fingerprints.
type Registry struct {
	entries map[string]Fingerprint
}

// New returns an empty Registry.
func New() *Registry {
	return &Registry{entries: make(map[string]Fingerprint)}
}

// Register computes and stores the fingerprint for the given address and labels.
func (r *Registry) Register(address string, labels ...string) Fingerprint {
	fp := Of(address, labels...)
	r.entries[address] = fp
	return fp
}

// Get returns the fingerprint previously registered for address, and whether it
// was found.
func (r *Registry) Get(address string) (Fingerprint, bool) {
	fp, ok := r.entries[address]
	return fp, ok
}

// All returns a snapshot of all registered address→fingerprint pairs.
func (r *Registry) All() map[string]Fingerprint {
	out := make(map[string]Fingerprint, len(r.entries))
	for k, v := range r.entries {
		out[k] = v
	}
	return out
}
