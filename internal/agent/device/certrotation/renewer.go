package certrotation

import (
	"context"
	"crypto"
	"encoding/pem"
	"fmt"
	"os"
	"time"

	"github.com/flightctl/flightctl/internal/agent/config"
	"github.com/flightctl/flightctl/internal/instrumentation/tracing"
	fcrypto "github.com/flightctl/flightctl/pkg/crypto"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
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
	// Start tracing span for renewal request generation
	_, span := tracing.StartSpan(ctx, TracingComponent, OpGenerateRenewalCSR)
	defer span.End()

	// Add span attributes
	span.SetAttributes(
		attribute.String("device.id", r.deviceID),
		attribute.String("certificate.serial", certMetadata.SerialNumber),
		attribute.String("certificate.not_after", certMetadata.NotAfter.Format(time.RFC3339)),
	)

	// Structured logging: certificate.renewal.initiated
	r.log.WithFields(logrus.Fields{
		"event":          "certificate.renewal.initiated",
		"device_id":      r.deviceID,
		"cert_serial":    certMetadata.SerialNumber,
		"cert_not_after": certMetadata.NotAfter.Format(time.RFC3339),
		"time_to_expiry": time.Until(certMetadata.NotAfter).String(),
	}).Info("Certificate renewal initiated")

	// Load the private key
	privateKey, err := r.loadPrivateKey()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to load private key")
		// Structured logging: certificate.renewal.failed
		r.log.WithFields(logrus.Fields{
			"event":     "certificate.renewal.failed",
			"device_id": r.deviceID,
			"reason":    "private_key_load_failed",
			"error":     err.Error(),
		}).Error("Certificate renewal failed: could not load private key")
		return nil, fmt.Errorf("loading private key: %w", err)
	}

	// Generate CSR using the existing private key
	csr, err := r.generateCSR(privateKey, r.deviceID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to generate CSR")
		// Structured logging: certificate.renewal.failed
		r.log.WithFields(logrus.Fields{
			"event":     "certificate.renewal.failed",
			"device_id": r.deviceID,
			"reason":    "csr_generation_failed",
			"error":     err.Error(),
		}).Error("Certificate renewal failed: could not generate CSR")
		return nil, fmt.Errorf("generating CSR: %w", err)
	}

	// Determine security proof type based on certificate validity
	proofType := r.determineSecurityProofType(certMetadata)

	// Create renewal request
	requestID := uuid.New().String()
	request := &RenewalRequest{
		RequestID:            requestID,
		DeviceID:             r.deviceID,
		OldCertificateSerial: certMetadata.SerialNumber,
		CSR:                  csr,
		SecurityProofType:    proofType,
		CreatedAt:            time.Now(),
		RetryCount:           0,
		Status:               StatusPending,
	}

	// Update span with request details
	span.SetAttributes(
		attribute.String("renewal.request_id", requestID),
		attribute.String("renewal.security_proof_type", string(proofType)),
	)
	span.SetStatus(codes.Ok, "renewal request generated successfully")

	// Structured logging: renewal request ready (will be submitted later)
	r.log.WithFields(logrus.Fields{
		"event":               "certificate.renewal.request_created",
		"request_id":          request.RequestID,
		"device_id":           r.deviceID,
		"security_proof_type": string(request.SecurityProofType),
		"old_cert_serial":     certMetadata.SerialNumber,
	}).Info("Certificate renewal request created and ready for submission")

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
