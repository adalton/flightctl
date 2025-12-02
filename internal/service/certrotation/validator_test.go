package certrotation

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"testing"
	"time"

	fcrypto "github.com/flightctl/flightctl/pkg/crypto"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

// TestValidateCSR tests CSR validation
func TestValidateCSR(t *testing.T) {
	log := logrus.New()
	log.SetLevel(logrus.ErrorLevel)
	validator := NewValidator(log, nil)

	// Generate a valid CSR
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	deviceID := "test-device-001"
	csrPEM, err := fcrypto.MakeCSR(privateKey, deviceID)
	require.NoError(t, err)

	// Test valid CSR
	csr, err := validator.ValidateCSR(csrPEM, deviceID)
	require.NoError(t, err)
	require.NotNil(t, csr)
	require.Equal(t, deviceID, csr.Subject.CommonName)

	// Test CSR with wrong device ID
	_, err = validator.ValidateCSR(csrPEM, "wrong-device-id")
	require.Error(t, err)
	require.Contains(t, err.Error(), "does not match device ID")

	// Test invalid CSR format
	_, err = validator.ValidateCSR([]byte("invalid-csr"), deviceID)
	require.Error(t, err)
}

// TestValidateCertificateForRenewal tests certificate validation
func TestValidateCertificateForRenewal(t *testing.T) {
	log := logrus.New()
	log.SetLevel(logrus.ErrorLevel)
	validator := NewValidator(log, nil)

	deviceID := "test-device-001"

	tests := []struct {
		name        string
		cert        *x509.Certificate
		deviceID    string
		expectErr   bool
		errContains string
	}{
		{
			name:        "nil certificate",
			cert:        nil,
			deviceID:    deviceID,
			expectErr:   true,
			errContains: "nil",
		},
		{
			name:      "valid certificate",
			cert:      createTestCert(t, deviceID, time.Now().Add(-1*time.Hour), time.Now().Add(365*24*time.Hour)),
			deviceID:  deviceID,
			expectErr: false,
		},
		{
			name:        "expired certificate",
			cert:        createTestCert(t, deviceID, time.Now().Add(-2*365*24*time.Hour), time.Now().Add(-365*24*time.Hour)),
			deviceID:    deviceID,
			expectErr:   true,
			errContains: "expired",
		},
		{
			name:        "not yet valid certificate",
			cert:        createTestCert(t, deviceID, time.Now().Add(1*time.Hour), time.Now().Add(2*time.Hour)),
			deviceID:    deviceID,
			expectErr:   true,
			errContains: "not yet valid",
		},
		{
			name:        "wrong device ID",
			cert:        createTestCert(t, "wrong-device", time.Now().Add(-1*time.Hour), time.Now().Add(365*24*time.Hour)),
			deviceID:    deviceID,
			expectErr:   true,
			errContains: "does not match",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateCertificateForRenewal(tt.cert, tt.deviceID)
			if tt.expectErr {
				require.Error(t, err)
				if tt.errContains != "" {
					require.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestValidateRenewalRequest tests full renewal request validation
func TestValidateRenewalRequest(t *testing.T) {
	log := logrus.New()
	log.SetLevel(logrus.ErrorLevel)
	validator := NewValidator(log, nil)

	deviceID := "test-device-001"

	// Create a valid certificate and CSR with the same key pair
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	cert := createTestCertWithKey(t, deviceID, time.Now().Add(-1*time.Hour), time.Now().Add(365*24*time.Hour), privateKey)
	csrPEM, err := fcrypto.MakeCSR(privateKey, deviceID)
	require.NoError(t, err)

	// Test valid renewal request
	csr, err := validator.ValidateRenewalRequest(csrPEM, deviceID, cert)
	require.NoError(t, err)
	require.NotNil(t, csr)

	// Test with different key pair (should fail)
	differentKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	differentCSR, err := fcrypto.MakeCSR(differentKey, deviceID)
	require.NoError(t, err)

	_, err = validator.ValidateRenewalRequest(differentCSR, deviceID, cert)
	require.Error(t, err)
	require.Contains(t, err.Error(), "key pair")
}

// TestVerifySameKeyPair tests key pair verification
func TestVerifySameKeyPair(t *testing.T) {
	log := logrus.New()
	log.SetLevel(logrus.ErrorLevel)
	validator := NewValidator(log, nil)

	deviceID := "test-device-001"

	// Create CSR and certificate with the same key
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	cert := createTestCertWithKey(t, deviceID, time.Now().Add(-1*time.Hour), time.Now().Add(365*24*time.Hour), privateKey)
	csrPEM, err := fcrypto.MakeCSR(privateKey, deviceID)
	require.NoError(t, err)

	csr, err := fcrypto.ParseCSR(csrPEM)
	require.NoError(t, err)

	// Should pass - same key
	err = validator.verifySameKeyPair(csr, cert)
	require.NoError(t, err)

	// Create CSR with different key
	differentKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	differentCSRPEM, err := fcrypto.MakeCSR(differentKey, deviceID)
	require.NoError(t, err)

	differentCSR, err := fcrypto.ParseCSR(differentCSRPEM)
	require.NoError(t, err)

	// Should fail - different key
	err = validator.verifySameKeyPair(differentCSR, cert)
	require.Error(t, err)
	require.Contains(t, err.Error(), "do not match")
}

// Helper functions

func createTestCert(t *testing.T, commonName string, notBefore, notAfter time.Time) *x509.Certificate {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	return createTestCertWithKey(t, commonName, notBefore, notAfter, privateKey)
}

func createTestCertWithKey(t *testing.T, commonName string, notBefore, notAfter time.Time, privateKey *ecdsa.PrivateKey) *x509.Certificate {
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
