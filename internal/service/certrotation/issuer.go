package certrotation

import (
	"context"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"

	"github.com/flightctl/flightctl/internal/crypto"
	"github.com/flightctl/flightctl/internal/crypto/signer"
	"github.com/sirupsen/logrus"
)

// CertificateIssuer issues new certificates for device renewal requests
type CertificateIssuer struct {
	ca  *crypto.CAClient
	log *logrus.Logger
}

// NewCertificateIssuer creates a new certificate issuer
func NewCertificateIssuer(ca *crypto.CAClient, log *logrus.Logger) *CertificateIssuer {
	return &CertificateIssuer{
		ca:  ca,
		log: log,
	}
}

// IssuedCertificate represents a newly issued certificate with metadata
type IssuedCertificate struct {
	Certificate    *x509.Certificate
	CertificatePEM []byte
	SerialNumber   string
}

// IssueCertificate issues a new certificate based on a validated CSR
// This is used for both proactive renewal (US1) and expired certificate recovery (US2)
func (ci *CertificateIssuer) IssueCertificate(ctx context.Context, csr *x509.CertificateRequest, deviceID string) (*IssuedCertificate, error) {
	ci.log.WithFields(logrus.Fields{
		"device_id": deviceID,
		"subject":   csr.Subject.String(),
	}).Info("Issuing new certificate")

	// Encode CSR to PEM format for the signer
	csrDER := csr.Raw
	csrPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE REQUEST",
		Bytes: csrDER,
	})

	// Create sign request using the signer library
	// Use ClientBootstrapSignerName for certificate renewal (same as bootstrap enrollment)
	// This ensures the renewed certificate has the same validity period and properties
	signReq, err := signer.NewSignRequestFromBytes(
		ci.ca.Cfg.ClientBootstrapSignerName,
		csrPEM,
		signer.WithResourceName(deviceID),
	)
	if err != nil {
		return nil, fmt.Errorf("creating sign request: %w", err)
	}

	// Sign the certificate using the CA
	certPEM, err := signer.SignAsPEM(ctx, ci.ca, signReq)
	if err != nil {
		return nil, fmt.Errorf("signing certificate: %w", err)
	}

	// Parse the signed certificate to extract metadata
	cert, ok := signReq.IssuedCertificate()
	if !ok {
		// If not available from sign request, parse from PEM
		block, _ := pem.Decode(certPEM)
		if block == nil {
			return nil, fmt.Errorf("failed to decode certificate PEM")
		}
		cert, err = x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("parsing signed certificate: %w", err)
		}
	}

	serialNumber := hex.EncodeToString(cert.SerialNumber.Bytes())

	ci.log.WithFields(logrus.Fields{
		"device_id":     deviceID,
		"serial_number": serialNumber,
		"not_before":    cert.NotBefore,
		"not_after":     cert.NotAfter,
		"validity_days": cert.NotAfter.Sub(cert.NotBefore).Hours() / 24,
	}).Info("Certificate issued successfully")

	return &IssuedCertificate{
		Certificate:    cert,
		CertificatePEM: certPEM,
		SerialNumber:   serialNumber,
	}, nil
}

// ValidateCertificate verifies that a certificate is well-formed and meets requirements
// This can be used before storing or returning certificates to devices
func (ci *CertificateIssuer) ValidateCertificate(cert *x509.Certificate, deviceID string) error {
	if cert == nil {
		return fmt.Errorf("certificate is nil")
	}

	// Verify subject matches device ID
	if cert.Subject.CommonName != deviceID {
		return fmt.Errorf("certificate subject (%s) does not match device ID (%s)", cert.Subject.CommonName, deviceID)
	}

	// Verify validity period is reasonable
	if cert.NotAfter.Before(cert.NotBefore) {
		return fmt.Errorf("certificate NotAfter (%v) is before NotBefore (%v)", cert.NotAfter, cert.NotBefore)
	}

	// Verify certificate is currently valid
	// Note: This is a sanity check - newly issued certs should be valid immediately
	if cert.NotBefore.After(cert.NotAfter) {
		return fmt.Errorf("certificate has invalid validity period")
	}

	ci.log.WithFields(logrus.Fields{
		"device_id": deviceID,
		"serial":    hex.EncodeToString(cert.SerialNumber.Bytes()),
	}).Debug("Certificate validation passed")

	return nil
}
