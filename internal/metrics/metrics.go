package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/yourorg/grpcpulse/internal/checker"
)

const namespace = "grpcpulse"

// Metrics holds all Prometheus metrics for grpcpulse.
type Metrics struct {
	UpGauge       *prometheus.GaugeVec
	LatencyHist   *prometheus.HistogramVec
	ChecksTotal   *prometheus.CounterVec
}

// New registers and returns a new Metrics instance using the provided registerer.
func New(reg prometheus.Registerer) *Metrics {
	f := promauto.With(reg)
	return &Metrics{
		UpGauge: f.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "up",
			Help:      "1 if the gRPC target is reachable, 0 otherwise.",
		}, []string{"address"}),

		LatencyHist: f.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "check_duration_seconds",
			Help:      "Latency of gRPC health checks in seconds.",
			Buckets:   prometheus.DefBuckets,
		}, []string{"address"}),

		ChecksTotal: f.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "checks_total",
			Help:      "Total number of health checks performed.",
		}, []string{"address", "status"}),
	}
}

// Record updates all metrics based on the provided check result.
func (m *Metrics) Record(result checker.Result) {
	addr := result.Address
	status := result.Status.String()

	up := 0.0
	if result.Status == checker.StatusHealthy {
		up = 1.0
	}

	m.UpGauge.WithLabelValues(addr).Set(up)
	m.LatencyHist.WithLabelValues(addr).Observe(result.Latency.Seconds())
	m.ChecksTotal.WithLabelValues(addr, status).Inc()
}
