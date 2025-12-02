package certrotation

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/flightctl/flightctl/internal/agent/config"
	fcrypto "github.com/flightctl/flightctl/pkg/crypto"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

// TestShouldRenew tests the expiration calculation logic
func TestShouldRenew(t *testing.T) {
	tests := []struct {
		name          string
		notAfter      time.Time
		thresholdDays int
		expected      bool
	}{
		{
			name:          "cert expires in 10 days, threshold 30 days - should renew",
			notAfter:      time.Now().Add(10 * 24 * time.Hour),
			thresholdDays: 30,
			expected:      true,
		},
		{
			name:          "cert expires in 50 days, threshold 30 days - should not renew",
			notAfter:      time.Now().Add(50 * 24 * time.Hour),
			thresholdDays: 30,
			expected:      false,
		},
		{
			name:          "cert expires in exactly 30 days, threshold 30 days - should renew",
			notAfter:      time.Now().Add(30 * 24 * time.Hour),
			thresholdDays: 30,
			expected:      true,
		},
		{
			name:          "cert already expired - should renew",
			notAfter:      time.Now().Add(-1 * time.Hour),
			thresholdDays: 30,
			expected:      true,
		},
		{
			name:          "cert expires in 1 day, threshold 7 days - should renew",
			notAfter:      time.Now().Add(1 * 24 * time.Hour),
			thresholdDays: 7,
			expected:      true,
		},
		{
			name:          "cert expires in 10 days, threshold 7 days - should not renew",
			notAfter:      time.Now().Add(10 * 24 * time.Hour),
			thresholdDays: 7,
			expected:      false,
		},
		{
			name:          "use default threshold (30 days) when zero",
			notAfter:      time.Now().Add(20 * 24 * time.Hour),
			thresholdDays: 0,
			expected:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ShouldRenew(tt.notAfter, tt.thresholdDays)
			require.Equal(t, tt.expected, result)
		})
	}
}

// TestTimeUntilRenewal tests calculation of time until renewal is needed
func TestTimeUntilRenewal(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name          string
		notAfter      time.Time
		thresholdDays int
		expectedMin   time.Duration
		expectedMax   time.Duration
	}{
		{
			name:          "cert expires in 60 days, threshold 30 - renewal in ~30 days",
			notAfter:      now.Add(60 * 24 * time.Hour),
			thresholdDays: 30,
			expectedMin:   29 * 24 * time.Hour,
			expectedMax:   31 * 24 * time.Hour,
		},
		{
			name:          "cert expires in 20 days, threshold 30 - renewal needed now",
			notAfter:      now.Add(20 * 24 * time.Hour),
			thresholdDays: 30,
			expectedMin:   0,
			expectedMax:   0,
		},
		{
			name:          "cert already expired - renewal needed now",
			notAfter:      now.Add(-10 * 24 * time.Hour),
			thresholdDays: 30,
			expectedMin:   0,
			expectedMax:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TimeUntilRenewal(tt.notAfter, tt.thresholdDays)
			require.GreaterOrEqual(t, result, tt.expectedMin)
			require.LessOrEqual(t, result, tt.expectedMax)
		})
	}
}

// TestMonitorLoadCertificateMetadata tests reading and parsing certificate metadata
func TestMonitorLoadCertificateMetadata(t *testing.T) {
	// Create a temporary directory for test certificates
	tmpDir := t.TempDir()
	certPath := filepath.Join(tmpDir, "test-cert.crt")

	// Generate a test certificate
	cert := createTestCertificate(t, "test-device", time.Now(), time.Now().Add(365*24*time.Hour))
	certPEM, err := fcrypto.EncodeCertificatePEM(cert)
	require.NoError(t, err)

	// Write certificate to file
	err = os.WriteFile(certPath, certPEM, 0600)
	require.NoError(t, err)

	// Create monitor
	cfg := config.NewDefault()
	log := logrus.New()
	log.SetLevel(logrus.ErrorLevel) // Reduce noise in tests

	renewalChan := make(chan RenewalTrigger, 1)
	monitor := NewMonitor(cfg, log, certPath, renewalChan, nil)

	// Load certificate metadata
	metadata, err := monitor.loadCertificateMetadata()
	require.NoError(t, err)

	// Verify metadata
	require.Equal(t, certPath, metadata.FilePath)
	require.Equal(t, "CN=test-device", metadata.Subject)
	require.NotEmpty(t, metadata.SerialNumber)
	require.NotEmpty(t, metadata.Fingerprint)
	require.False(t, metadata.NotBefore.IsZero())
	require.False(t, metadata.NotAfter.IsZero())
}

// TestMonitorCheckExpirationTriggersRenewal tests that the monitor triggers renewal when appropriate
func TestMonitorCheckExpirationTriggersRenewal(t *testing.T) {
	tmpDir := t.TempDir()
	certPath := filepath.Join(tmpDir, "expiring-cert.crt")

	// Create a certificate that expires in 10 days (within default 30-day threshold)
	cert := createTestCertificate(t, "test-device", time.Now().Add(-1*time.Hour), time.Now().Add(10*24*time.Hour))
	certPEM, err := fcrypto.EncodeCertificatePEM(cert)
	require.NoError(t, err)

	err = os.WriteFile(certPath, certPEM, 0600)
	require.NoError(t, err)

	// Create monitor with default config (30-day threshold)
	cfg := config.NewDefault()
	cfg.CertRotation.RenewalThresholdDays = 30
	log := logrus.New()
	log.SetLevel(logrus.ErrorLevel)

	renewalChan := make(chan RenewalTrigger, 1)
	monitor := NewMonitor(cfg, log, certPath, renewalChan, nil)

	// Check expiration - should trigger renewal
	ctx := context.Background()
	err = monitor.checkCertificateExpiration(ctx)
	require.NoError(t, err)

	// Verify renewal was triggered
	select {
	case trigger := <-renewalChan:
		require.NotEmpty(t, trigger.CertMetadata.SerialNumber)
		require.Contains(t, trigger.Reason, "expires in")
		require.Greater(t, trigger.TimeToExpiry, time.Duration(0))
	case <-time.After(1 * time.Second):
		t.Fatal("Expected renewal trigger but none received")
	}
}

// TestMonitorCheckExpirationNoRenewalNeeded tests that monitor doesn't trigger renewal unnecessarily
func TestMonitorCheckExpirationNoRenewalNeeded(t *testing.T) {
	tmpDir := t.TempDir()
	certPath := filepath.Join(tmpDir, "valid-cert.crt")

	// Create a certificate that expires in 60 days (beyond 30-day threshold)
	cert := createTestCertificate(t, "test-device", time.Now().Add(-1*time.Hour), time.Now().Add(60*24*time.Hour))
	certPEM, err := fcrypto.EncodeCertificatePEM(cert)
	require.NoError(t, err)

	err = os.WriteFile(certPath, certPEM, 0600)
	require.NoError(t, err)

	// Create monitor with default config (30-day threshold)
	cfg := config.NewDefault()
	cfg.CertRotation.RenewalThresholdDays = 30
	log := logrus.New()
	log.SetLevel(logrus.ErrorLevel)

	renewalChan := make(chan RenewalTrigger, 1)
	monitor := NewMonitor(cfg, log, certPath, renewalChan, nil)

	// Check expiration - should NOT trigger renewal
	ctx := context.Background()
	err = monitor.checkCertificateExpiration(ctx)
	require.NoError(t, err)

	// Verify no renewal was triggered
	select {
	case <-renewalChan:
		t.Fatal("Unexpected renewal trigger for valid certificate")
	case <-time.After(100 * time.Millisecond):
		// Expected - no renewal triggered
	}
}

// createTestCertificate creates a self-signed certificate for testing
func createTestCertificate(t *testing.T, commonName string, notBefore, notAfter time.Time) *x509.Certificate {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	require.NoError(t, err)

	template := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName: commonName,
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &privateKey.PublicKey, privateKey)
	require.NoError(t, err)

	cert, err := x509.ParseCertificate(certDER)
	require.NoError(t, err)

	return cert
}
