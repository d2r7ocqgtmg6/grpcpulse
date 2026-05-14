// Package audit provides a bounded, thread-safe append-only event log that
// records configuration and state change events within grpcpulse.
//
// Events are classified by EventKind (e.g. silence_added, circuit_open) and
// optionally associated with a named target. The log evicts the oldest entries
// once its capacity is exceeded, keeping memory usage predictable.
//
// Typical usage:
//
//	log := audit.New(512)
//	log.Record(audit.EventKindSilenceAdded, "payments-svc", "silenced for 30m")
//	events := log.FilterByTarget("payments-svc")
package audit
