package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

// CertRotationMetrics holds all Prometheus metrics for certificate rotation operations
type CertRotationMetrics struct {
	// CertExpirationTime is a gauge for certificate expiration timestamp (unix seconds)
	CertExpirationTime prometheus.Gauge

	// RenewalAttempts is a counter for total renewal attempts
	RenewalAttempts prometheus.Counter

	// RenewalSuccesses is a counter for successful renewals
	RenewalSuccesses prometheus.Counter

	// RenewalFailures is a counter for failed renewals
	RenewalFailures *prometheus.CounterVec

	// RotationDuration is a histogram for time to complete rotation (seconds)
	RotationDuration prometheus.Histogram
}

// NewCertRotationMetrics creates and registers certificate rotation metrics
func NewCertRotationMetrics(reg prometheus.Registerer) *CertRotationMetrics {
	m := &CertRotationMetrics{
		CertExpirationTime: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "flightctl_agent_cert_expiration_time_seconds",
			Help: "Certificate expiration timestamp in Unix seconds",
		}),
		RenewalAttempts: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "flightctl_agent_cert_renewal_attempts_total",
			Help: "Total number of certificate renewal attempts",
		}),
		RenewalSuccesses: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "flightctl_agent_cert_renewal_successes_total",
			Help: "Total number of successful certificate renewals",
		}),
		RenewalFailures: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "flightctl_agent_cert_renewal_failures_total",
			Help: "Total number of failed certificate renewals",
		}, []string{"reason"}),
		RotationDuration: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "flightctl_agent_cert_rotation_duration_seconds",
			Help:    "Time taken to complete certificate rotation in seconds",
			Buckets: prometheus.DefBuckets,
		}),
	}

	// Register all metrics
	if reg != nil {
		reg.MustRegister(
			m.CertExpirationTime,
			m.RenewalAttempts,
			m.RenewalSuccesses,
			m.RenewalFailures,
			m.RotationDuration,
		)
	}

	return m
}
