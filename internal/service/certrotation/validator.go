package certrotation

import (
	"crypto/x509"
	"fmt"
	"time"

	fcrypto "github.com/flightctl/flightctl/pkg/crypto"
	"github.com/sirupsen/logrus"
)

// Validator validates certificate renewal requests and security proofs
type Validator struct {
	log *logrus.Logger
	ca  *x509.CertPool
}

// NewValidator creates a new certificate renewal validator
func NewValidator(log *logrus.Logger, caCertPool *x509.CertPool) *Validator {
	return &Validator{
		log: log,
		ca:  caCertPool,
	}
}

// ValidateCSR validates a Certificate Signing Request
func (v *Validator) ValidateCSR(csrPEM []byte, deviceID string) (*x509.CertificateRequest, error) {
	// Parse the CSR
	csr, err := fcrypto.ParseCSR(csrPEM)
	if err != nil {
		return nil, fmt.Errorf("parsing CSR: %w", err)
	}

	// Validate CSR signature
	if err := fcrypto.ValidateX509CSR(csr); err != nil {
		return nil, fmt.Errorf("invalid CSR signature: %w", err)
	}

	// Verify the CSR subject matches the device ID
	if csr.Subject.CommonName != deviceID {
		return nil, fmt.Errorf("CSR subject (%s) does not match device ID (%s)", csr.Subject.CommonName, deviceID)
	}

	v.log.WithFields(logrus.Fields{
		"device_id": deviceID,
		"subject":   csr.Subject.String(),
	}).Debug("CSR validated successfully")

	return csr, nil
}

// ValidateCertificateForRenewal validates that a certificate can be used for authentication
// For Phase 3 (User Story 1), this validates the management certificate presented via mTLS
func (v *Validator) ValidateCertificateForRenewal(cert *x509.Certificate, deviceID string) error {
	if cert == nil {
		return fmt.Errorf("certificate is nil")
	}

	// Check if certificate has expired
	now := time.Now()
	if now.Before(cert.NotBefore) {
		return fmt.Errorf("certificate not yet valid (NotBefore: %v)", cert.NotBefore)
	}
	if now.After(cert.NotAfter) {
		return fmt.Errorf("certificate has expired (NotAfter: %v)", cert.NotAfter)
	}

	// Verify the certificate subject matches the device ID
	if cert.Subject.CommonName != deviceID {
		return fmt.Errorf("certificate subject (%s) does not match device ID (%s)", cert.Subject.CommonName, deviceID)
	}

	// Verify certificate chain against CA
	if v.ca != nil {
		opts := x509.VerifyOptions{
			Roots:       v.ca,
			CurrentTime: now,
			KeyUsages:   []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		}

		if _, err := cert.Verify(opts); err != nil {
			return fmt.Errorf("certificate chain verification failed: %w", err)
		}
	}

	v.log.WithFields(logrus.Fields{
		"device_id": deviceID,
		"subject":   cert.Subject.String(),
		"serial":    cert.SerialNumber.String(),
		"not_after": cert.NotAfter,
	}).Debug("Certificate validated successfully for renewal")

	return nil
}

// ValidateRenewalRequest validates the overall renewal request
func (v *Validator) ValidateRenewalRequest(csrPEM []byte, deviceID string, cert *x509.Certificate) (*x509.CertificateRequest, error) {
	// Validate the CSR
	csr, err := v.ValidateCSR(csrPEM, deviceID)
	if err != nil {
		return nil, fmt.Errorf("CSR validation failed: %w", err)
	}

	// Validate the client certificate (security proof)
	if err := v.ValidateCertificateForRenewal(cert, deviceID); err != nil {
		return nil, fmt.Errorf("certificate validation failed: %w", err)
	}

	// Verify that the CSR public key matches the certificate public key
	// This ensures the device is renewing with the same key pair
	if err := v.verifySameKeyPair(csr, cert); err != nil {
		return nil, fmt.Errorf("key pair verification failed: %w", err)
	}

	return csr, nil
}

// verifySameKeyPair verifies that the CSR and certificate use the same public key
func (v *Validator) verifySameKeyPair(csr *x509.CertificateRequest, cert *x509.Certificate) error {
	// Compare the public keys
	// This is a security check to ensure the device is using the same key pair
	csrKeyBytes, err := x509.MarshalPKIXPublicKey(csr.PublicKey)
	if err != nil {
		return fmt.Errorf("marshalling CSR public key: %w", err)
	}

	certKeyBytes, err := x509.MarshalPKIXPublicKey(cert.PublicKey)
	if err != nil {
		return fmt.Errorf("marshalling certificate public key: %w", err)
	}

	// Compare byte representations
	if len(csrKeyBytes) != len(certKeyBytes) {
		return fmt.Errorf("public key length mismatch: CSR=%d, cert=%d", len(csrKeyBytes), len(certKeyBytes))
	}

	for i := range csrKeyBytes {
		if csrKeyBytes[i] != certKeyBytes[i] {
			return fmt.Errorf("public keys do not match")
		}
	}

	v.log.Debug("CSR and certificate use the same public key")
	return nil
}
