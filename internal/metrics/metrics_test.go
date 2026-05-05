package metrics_test

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"

	"github.com/yourorg/grpcpulse/internal/checker"
	"github.com/yourorg/grpcpulse/internal/metrics"
)

func newTestMetrics(t *testing.T) (*metrics.Metrics, *prometheus.Registry) {
	t.Helper()
	reg := prometheus.NewRegistry()
	m := metrics.New(reg)
	return m, reg
}

func TestRecord_HealthyResult(t *testing.T) {
	m, _ := newTestMetrics(t)

	result := checker.Result{
		Address: "localhost:50051",
		Status:  checker.StatusHealthy,
		Latency: 42 * time.Millisecond,
	}
	m.Record(result)

	if got := testutil.ToFloat64(m.UpGauge.WithLabelValues("localhost:50051")); got != 1.0 {
		t.Errorf("up gauge: want 1.0, got %f", got)
	}
	if got := testutil.ToFloat64(m.ChecksTotal.WithLabelValues("localhost:50051", "healthy")); got != 1.0 {
		t.Errorf("checks_total healthy: want 1.0, got %f", got)
	}
}

func TestRecord_UnhealthyResult(t *testing.T) {
	m, _ := newTestMetrics(t)

	result := checker.Result{
		Address: "localhost:50052",
		Status:  checker.StatusUnhealthy,
		Latency: 500 * time.Millisecond,
	}
	m.Record(result)

	if got := testutil.ToFloat64(m.UpGauge.WithLabelValues("localhost:50052")); got != 0.0 {
		t.Errorf("up gauge: want 0.0, got %f", got)
	}
	if got := testutil.ToFloat64(m.ChecksTotal.WithLabelValues("localhost:50052", "unhealthy")); got != 1.0 {
		t.Errorf("checks_total unhealthy: want 1.0, got %f", got)
	}
}

func TestRecord_MultipleChecks(t *testing.T) {
	m, _ := newTestMetrics(t)
	addr := "localhost:50053"

	for i := 0; i < 3; i++ {
		m.Record(checker.Result{Address: addr, Status: checker.StatusHealthy, Latency: time.Millisecond})
	}
	m.Record(checker.Result{Address: addr, Status: checker.StatusUnhealthy, Latency: time.Millisecond})

	if got := testutil.ToFloat64(m.ChecksTotal.WithLabelValues(addr, "healthy")); got != 3.0 {
		t.Errorf("expected 3 healthy checks, got %f", got)
	}
	if got := testutil.ToFloat64(m.ChecksTotal.WithLabelValues(addr, "unhealthy")); got != 1.0 {
		t.Errorf("expected 1 unhealthy check, got %f", got)
	}
}
