// Package alertmanager implements a lightweight client for forwarding
// grpcpulse health-transition events to an Alertmanager-compatible HTTP
// endpoint.
//
// Usage:
//
//	client := alertmanager.New(alertmanager.Config{
//		URL:     "http://alertmanager:9093",
//		Timeout: 5 * time.Second,
//	})
//
//	err := client.Send(ctx, []alertmanager.Alert{
//		{
//			Labels:      map[string]string{"alertname": "ServiceDown", "target": addr},
//			Annotations: map[string]string{"summary": "gRPC target is unhealthy"},
//			StartsAt:    time.Now(),
//		},
//	})
//
// Alerts are serialised as a JSON array and posted to /api/v2/alerts.
// Any non-2xx response is treated as an error.
package alertmanager
