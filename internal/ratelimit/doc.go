// Package ratelimit implements a per-target token-bucket rate limiter used by
// the grpcpulse scheduler to prevent hammering unhealthy or slow gRPC endpoints.
//
// Usage:
//
//	limiter := ratelimit.New(ratelimit.DefaultConfig())
//
//	if limiter.Allow(target) {
//		// perform health check
//	}
//
// The MinDelay field in Config controls the minimum interval between successive
// checks for the same target address. Calls to Allow that arrive before the
// delay has elapsed return false without modifying state.
package ratelimit
