// Package acknowledger provides a thread-safe store for manual acknowledgements
// of unhealthy gRPC targets.
//
// An acknowledgement suppresses further alert notifications for a given target
// for a configurable duration. Once expired or explicitly cleared, the target
// resumes normal alerting behaviour.
//
// Typical usage:
//
//	ack := acknowledger.New()
//	ack.Acknowledge("payments:50051", "known outage", "ops-team", 2*time.Hour)
//	if ack.IsAcknowledged("payments:50051") {
//	    // skip alert
//	}
package acknowledger
