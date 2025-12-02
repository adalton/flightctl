package certrotation

import (
	"context"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"fmt"
	"os"
	"time"

	"github.com/flightctl/flightctl/internal/agent/config"
	"github.com/flightctl/flightctl/internal/agent/instrumentation/metrics"
	"github.com/flightctl/flightctl/pkg/crypto"
	"github.com/sirupsen/logrus"
)

// Monitor checks certificate expiration and triggers renewal when needed
type Monitor struct {
	config      *config.Config
	log         *logrus.Logger
	certPath    string
	renewalChan chan<- RenewalTrigger
	stopChan    chan struct{}
	checkTicker *time.Ticker
	metrics     *metrics.CertRotationMetrics
}

// RenewalTrigger contains information about why renewal was triggered
type RenewalTrigger struct {
	CertMetadata CertificateMetadata
	Reason       string
	TimeToExpiry time.Duration
}

// NewMonitor creates a new certificate expiration monitor
func NewMonitor(cfg *config.Config, log *logrus.Logger, certPath string, renewalChan chan<- RenewalTrigger, metricsCollector *metrics.CertRotationMetrics) *Monitor {
	return &Monitor{
		config:      cfg,
		log:         log,
		certPath:    certPath,
		renewalChan: renewalChan,
		stopChan:    make(chan struct{}),
		metrics:     metricsCollector,
	}
}

// Start begins monitoring certificate expiration
func (m *Monitor) Start(ctx context.Context) error {
	interval := time.Duration(m.config.CertRotation.MonitorIntervalSeconds) * time.Second
	if interval == 0 {
		interval = time.Duration(config.DefaultCertMonitorIntervalSeconds) * time.Second
	}

	m.checkTicker = time.NewTicker(interval)
	defer m.checkTicker.Stop()

	m.log.WithField("interval", interval).Info("Starting certificate expiration monitor")

	// Perform initial check immediately
	if err := m.checkCertificateExpiration(ctx); err != nil {
		m.log.WithError(err).Warn("Initial certificate expiration check failed")
	}

	for {
		select {
		case <-ctx.Done():
			m.log.Info("Certificate monitor stopping due to context cancellation")
			return ctx.Err()
		case <-m.stopChan:
			m.log.Info("Certificate monitor stopped")
			return nil
		case <-m.checkTicker.C:
			if err := m.checkCertificateExpiration(ctx); err != nil {
				m.log.WithError(err).Error("Certificate expiration check failed")
			}
		}
	}
}

// Stop halts the certificate monitor
func (m *Monitor) Stop() {
	close(m.stopChan)
}

// checkCertificateExpiration reads the certificate and checks if renewal is needed
func (m *Monitor) checkCertificateExpiration(ctx context.Context) error {
	metadata, err := m.loadCertificateMetadata()
	if err != nil {
		return fmt.Errorf("loading certificate metadata: %w", err)
	}

	// Emit certificate expiration time metric
	if m.metrics != nil {
		m.metrics.CertExpirationTime.Set(float64(metadata.NotAfter.Unix()))
	}

	timeToExpiry := time.Until(metadata.NotAfter)
	thresholdDays := m.config.CertRotation.RenewalThresholdDays
	if thresholdDays == 0 {
		thresholdDays = config.DefaultCertRenewalThresholdDays
	}
	threshold := time.Duration(thresholdDays) * 24 * time.Hour

	m.log.WithFields(logrus.Fields{
		"serial":         metadata.SerialNumber,
		"not_after":      metadata.NotAfter,
		"time_to_expiry": timeToExpiry,
		"threshold":      threshold,
	}).Debug("Certificate expiration check")

	if ShouldRenew(metadata.NotAfter, thresholdDays) {
		// Increment renewal attempt counter
		if m.metrics != nil {
			m.metrics.RenewalAttempts.Inc()
		}

		trigger := RenewalTrigger{
			CertMetadata: metadata,
			Reason:       fmt.Sprintf("Certificate expires in %v (threshold: %d days)", timeToExpiry, thresholdDays),
			TimeToExpiry: timeToExpiry,
		}

		select {
		case m.renewalChan <- trigger:
			// Structured logging: certificate.expiration.detected
			m.log.WithFields(logrus.Fields{
				"event":          "certificate.expiration.detected",
				"cert_serial":    metadata.SerialNumber,
				"not_after":      metadata.NotAfter.Format(time.RFC3339),
				"time_to_expiry": timeToExpiry.String(),
				"threshold_days": thresholdDays,
			}).Info("Certificate expiring soon - renewal triggered")
		case <-ctx.Done():
			return ctx.Err()
		default:
			m.log.WithFields(logrus.Fields{
				"event":          "certificate.renewal.skipped",
				"cert_serial":    metadata.SerialNumber,
				"reason":         "renewal_channel_full",
				"time_to_expiry": timeToExpiry.String(),
			}).Warn("Renewal channel full, skipping trigger (renewal may already be in progress)")
		}
	}

	return nil
}

// loadCertificateMetadata reads the certificate file and extracts metadata
func (m *Monitor) loadCertificateMetadata() (CertificateMetadata, error) {
	certPEM, err := os.ReadFile(m.certPath)
	if err != nil {
		return CertificateMetadata{}, fmt.Errorf("reading certificate file %s: %w", m.certPath, err)
	}

	cert, err := crypto.ParseCertificatePEM(certPEM)
	if err != nil {
		return CertificateMetadata{}, fmt.Errorf("parsing certificate PEM: %w", err)
	}

	return extractCertificateMetadata(m.certPath, cert), nil
}

// extractCertificateMetadata converts an x509.Certificate to CertificateMetadata
func extractCertificateMetadata(filePath string, cert *x509.Certificate) CertificateMetadata {
	fingerprint := sha256.Sum256(cert.Raw)

	return CertificateMetadata{
		FilePath:     filePath,
		NotBefore:    cert.NotBefore,
		NotAfter:     cert.NotAfter,
		Subject:      cert.Subject.String(),
		Issuer:       cert.Issuer.String(),
		SerialNumber: hex.EncodeToString(cert.SerialNumber.Bytes()),
		Fingerprint:  hex.EncodeToString(fingerprint[:]),
	}
}

// ShouldRenew determines if a certificate should be renewed based on expiration time
// This function is exported for testing
func ShouldRenew(notAfter time.Time, thresholdDays int) bool {
	if thresholdDays == 0 {
		thresholdDays = config.DefaultCertRenewalThresholdDays
	}
	threshold := time.Duration(thresholdDays) * 24 * time.Hour
	timeToExpiry := time.Until(notAfter)
	return timeToExpiry <= threshold
}

// TimeUntilRenewal calculates how long until a certificate should be renewed
// Returns 0 if renewal is already needed
func TimeUntilRenewal(notAfter time.Time, thresholdDays int) time.Duration {
	if thresholdDays == 0 {
		thresholdDays = config.DefaultCertRenewalThresholdDays
	}
	threshold := time.Duration(thresholdDays) * 24 * time.Hour
	renewalTime := notAfter.Add(-threshold)
	timeUntilRenewal := time.Until(renewalTime)

	if timeUntilRenewal < 0 {
		return 0
	}
	return timeUntilRenewal
}
