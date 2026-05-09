// Package retry provides a configurable retry mechanism for transient failures.
//
// It supports exponential backoff with jitter, a maximum number of attempts,
// and context-aware cancellation so retries are abandoned when the caller's
// context is done.
//
// # Basic usage
//
//	cfg := retry.DefaultConfig()
//	r := retry.New(cfg)
//
//	err := r.Do(ctx, func() error {
//		return doSomethingFallible()
//	})
//
// # Configuration
//
// DefaultConfig returns sensible defaults (3 attempts, 100 ms base delay,
// 2 s max delay). All fields can be overridden before passing the config
// to New.
package retry
