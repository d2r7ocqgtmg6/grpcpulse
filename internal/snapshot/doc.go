// Package snapshot provides point-in-time capture of all target health states.
//
// A Registry wraps a HealthSource and an optional TagSource to produce
// Snapshot values that can be stored, exported, or served over HTTP.
//
// Snapshots are immutable once captured; call Capture again to refresh.
package snapshot
