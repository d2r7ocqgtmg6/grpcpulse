// Package ratelimit provides per-target rate limiting for health-check
// probes. It ensures that a given target cannot be checked more frequently
// than a configured minimum interval, preventing thundering-herd scenarios
// when many targets share the same upstream gRPC service.
//
// Usage:
//
//	rl := ratelimit.New(ratelimit.DefaultConfig())
//	if rl.Allow("my-service:443") {
//	    // perform health check
//	}
package ratelimit
