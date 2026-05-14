// Package tags provides a thread-safe registry for attaching arbitrary
// key-value label pairs (tags) to monitored gRPC targets.
//
// Tags can be used to group, filter, or annotate targets in dashboards and
// alert notifications. The Registry is safe for concurrent use and stores
// a defensive copy of each Tags map on write.
//
// Example:
//
//	reg := tags.New()
//	reg.Set("payments.internal:443", tags.Tags{"env": "prod", "team": "payments"})
//
//	// Later, retrieve all production targets:
//	prodTargets := reg.Filter(tags.Tags{"env": "prod"})
package tags
