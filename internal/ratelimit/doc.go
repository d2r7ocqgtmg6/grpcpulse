// Package ratelimit provides per-target rate limiting for health checks.
//
// It ensures that a given target is not checked more frequently than a
// configured minimum interval, preventing thundering-herd scenarios and
// reducing load on downstream services.
//
// Usage:
//
//	rl := ratelimit.New(ratelimit.DefaultConfig())
//	if rl.Allow("my-service:443") {
//	    // perform check
//	}
package ratelimit
