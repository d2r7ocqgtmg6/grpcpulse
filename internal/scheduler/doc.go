// Package scheduler provides a periodic health-check runner.
//
// The Scheduler wraps a [checker.Checker] and a [metrics.Metrics] instance,
// executing health checks at a configurable interval and recording each
// result so that Prometheus metrics stay up to date.
//
// Basic usage:
//
//	cfg := scheduler.DefaultConfig()
//	cfg.Interval = 10 * time.Second
//	s := scheduler.New(chk, met, cfg)
//	go s.Start(ctx)
//	// later...
//	s.Stop()
package scheduler
