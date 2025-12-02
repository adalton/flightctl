package certrotation

import (
	"context"
	"crypto"
	"encoding/pem"
	"fmt"
	"os"
	"time"

	"github.com/flightctl/flightctl/internal/agent/config"
	fcrypto "github.com/flightctl/flightctl/pkg/crypto"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// Renewer handles certificate renewal operations
type Renewer struct {
	config   *config.Config
	log      *logrus.Logger
	deviceID string
	keyPath  string
	certPath string
}

// NewRenewer creates a new certificate renewer
func NewRenewer(cfg *config.Config, log *logrus.Logger, deviceID string, certPath string, keyPath string) *Renewer {
	return &Renewer{
		config:   cfg,
		log:      log,
		deviceID: deviceID,
		certPath: certPath,
		keyPath:  keyPath,
	}
}

// GenerateRenewalRequest creates a renewal request with CSR for the current certificate
func (r *Renewer) GenerateRenewalRequest(ctx context.Context, certMetadata CertificateMetadata) (*RenewalRequest, error) {
	r.log.WithFields(logrus.Fields{
		"device_id":      r.deviceID,
		"cert_serial":    certMetadata.SerialNumber,
		"cert_not_after": certMetadata.NotAfter,
	}).Info("Generating certificate renewal request")

	// Load the private key
	privateKey, err := r.loadPrivateKey()
	if err != nil {
		return nil, fmt.Errorf("loading private key: %w", err)
	}

	// Generate CSR using the existing private key
	csr, err := r.generateCSR(privateKey, r.deviceID)
	if err != nil {
		return nil, fmt.Errorf("generating CSR: %w", err)
	}

	// Determine security proof type based on certificate validity
	proofType := r.determineSecurityProofType(certMetadata)

	// Create renewal request
	request := &RenewalRequest{
		RequestID:            uuid.New().String(),
		DeviceID:             r.deviceID,
		OldCertificateSerial: certMetadata.SerialNumber,
		CSR:                  csr,
		SecurityProofType:    proofType,
		CreatedAt:            time.Now(),
		RetryCount:           0,
		Status:               StatusPending,
	}

	r.log.WithFields(logrus.Fields{
		"request_id":          request.RequestID,
		"security_proof_type": request.SecurityProofType,
	}).Info("Certificate renewal request generated")

	return request, nil
}

// loadPrivateKey reads and parses the device's private key
func (r *Renewer) loadPrivateKey() (crypto.Signer, error) {
	keyPEM, err := os.ReadFile(r.keyPath)
	if err != nil {
		return nil, fmt.Errorf("reading private key file %s: %w", r.keyPath, err)
	}

	block, _ := pem.Decode(keyPEM)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block from %s", r.keyPath)
	}

	// Use the existing crypto library to parse the key
	privKey, err := fcrypto.ParseKeyPEM(keyPEM)
	if err != nil {
		return nil, fmt.Errorf("parsing private key: %w", err)
	}

	signer, ok := privKey.(crypto.Signer)
	if !ok {
		return nil, fmt.Errorf("private key does not implement crypto.Signer")
	}

	return signer, nil
}

// generateCSR creates a Certificate Signing Request
func (r *Renewer) generateCSR(privateKey crypto.Signer, deviceID string) ([]byte, error) {
	csr, err := fcrypto.MakeCSR(privateKey, deviceID)
	if err != nil {
		return nil, fmt.Errorf("creating CSR: %w", err)
	}

	return csr, nil
}

// determineSecurityProofType determines which authentication method to use
func (r *Renewer) determineSecurityProofType(certMetadata CertificateMetadata) SecurityProofType {
	// Check if current certificate is still valid
	now := time.Now()
	if now.Before(certMetadata.NotAfter) && now.After(certMetadata.NotBefore) {
		// Certificate is valid, use it for authentication
		r.log.Debug("Using valid certificate for authentication")
		return ValidCertificate
	}

	// Certificate is expired - need to check for bootstrap cert or TPM
	// For Phase 3 (User Story 1), we only implement proactive renewal with valid certs
	// Bootstrap and TPM recovery will be added in Phase 5 (User Story 2)
	r.log.Warn("Certificate expired - bootstrap/TPM recovery not yet implemented")
	return ValidCertificate // Will fail on service side, but allows us to complete User Story 1
}
