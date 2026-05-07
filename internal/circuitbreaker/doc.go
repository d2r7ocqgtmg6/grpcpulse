// Package circuitbreaker provides a per-target circuit breaker used by the
// grpcpulse scheduler to avoid hammering unhealthy gRPC endpoints.
//
// The circuit transitions through three states:
//
//   - Closed: checks proceed normally.
//   - Open: checks are skipped until the cooldown period elapses.
//   - Half-Open: one probe check is allowed; success closes the circuit,
//     failure reopens it.
//
// Configuration (threshold and cooldown) can be set via [Config]; sensible
// defaults are provided by [DefaultConfig].
package circuitbreaker
