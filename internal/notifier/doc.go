// Package notifier tracks per-target health-state transitions for gRPC
// services monitored by grpcpulse.
//
// It is intentionally side-effect free: callers supply an AlertFunc that
// is invoked exactly once per state transition (unknown → healthy,
// healthy → unhealthy, etc.).  A nil AlertFunc falls back to structured
// log output via [log/slog].
//
// Typical usage inside the scheduler loop:
//
//	notif := notifier.New(myAlertFunc, logger)
//	// ... after each checker result:
//	notif.Observe(target.Address, result.Healthy)
//
// The Notifier is safe for concurrent use.
package notifier
