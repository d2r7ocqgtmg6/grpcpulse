// Package dashboard assembles an aggregated, point-in-time snapshot of
// every monitored gRPC target by combining data from the history store
// and the notifier state machine.
//
// Typical usage:
//
//	b := dashboard.New(historyStore, notifierInstance)
//	snap := b.Build() // call on each HTTP request or on a ticker
//
// The resulting [Snapshot] is safe to serialise directly to JSON and is
// served by the /dashboard HTTP endpoint registered in the server package.
package dashboard
