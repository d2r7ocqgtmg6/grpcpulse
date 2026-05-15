// Package dependency tracks directed dependency relationships between gRPC
// service targets monitored by grpcpulse.
//
// A dependency edge A → B means "A depends on B". The registry provides
// cycle detection so that configurations cannot inadvertently create
// infinite propagation loops.
//
// Typical usage:
//
//	reg := dependency.New()
//	_ = reg.Add("payment-svc", "auth-svc")
//	deps := reg.DependsOn("payment-svc") // ["auth-svc"]
package dependency
