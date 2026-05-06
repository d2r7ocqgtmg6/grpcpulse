// Package history provides a thread-safe, bounded ring-buffer of health-check
// results keyed by target address.
//
// Each target maintains up to a configurable number of [Entry] values.
// Older entries are evicted automatically when the buffer is full, making
// the package suitable for long-running daemons where unbounded growth
// would otherwise be a concern.
//
// Typical usage:
//
//	h := history.New(200)
//	h.Record("localhost:50051", true, 3*time.Millisecond)
//	entries := h.Get("localhost:50051")
//	uptime := h.UptimePercent("localhost:50051")
package history
