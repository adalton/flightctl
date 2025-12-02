package certrotation

import (
	"context"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"net/http"
	"time"

	"github.com/flightctl/flightctl/internal/crypto"
	"github.com/flightctl/flightctl/internal/crypto/signer"
	"github.com/flightctl/flightctl/internal/store"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// RenewalRequest represents the request payload for certificate renewal
type RenewalRequest struct {
	RequestID            string                  `json:"requestId"`
	CSR                  string                  `json:"csr"`
	SecurityProofType    string                  `json:"securityProofType"`
	OldCertificateSerial string                  `json:"oldCertificateSerial,omitempty"`
	TPMAttestationProof  *TPMAttestationProofReq `json:"tpmAttestationProof,omitempty"`
}

// TPMAttestationProofReq represents TPM attestation proof (for Phase 5 - User Story 2)
type TPMAttestationProofReq struct {
	AttestationData string `json:"attestationData"`
	Signature       string `json:"signature"`
}

// RenewalResponse represents the response payload for certificate renewal
type RenewalResponse struct {
	RequestID    string    `json:"requestId"`
	Status       string    `json:"status"`
	Certificate  string    `json:"certificate"`
	SerialNumber string    `json:"serialNumber"`
	NotBefore    time.Time `json:"notBefore"`
	NotAfter     time.Time `json:"notAfter"`
	IssuedAt     time.Time `json:"issuedAt"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error     string `json:"error"`
	Details   string `json:"details,omitempty"`
	RequestID string `json:"requestId,omitempty"`
}

// Handler handles certificate renewal requests
type Handler struct {
	store     store.Store
	ca        *crypto.CAClient
	validator *Validator
	log       *logrus.Logger
}

// NewHandler creates a new certificate renewal handler
func NewHandler(store store.Store, ca *crypto.CAClient, validator *Validator, log *logrus.Logger) *Handler {
	return &Handler{
		store:     store,
		ca:        ca,
		validator: validator,
		log:       log,
	}
}

// HandleRenewalRequest processes a certificate renewal request
func (h *Handler) HandleRenewalRequest(ctx context.Context, deviceID string, req *RenewalRequest, clientCert *x509.Certificate) (*RenewalResponse, int, error) {
	startTime := time.Now()

	h.log.WithFields(logrus.Fields{
		"device_id":  deviceID,
		"request_id": req.RequestID,
		"proof_type": req.SecurityProofType,
	}).Info("Processing certificate renewal request")

	// Validate request ID format
	if _, err := uuid.Parse(req.RequestID); err != nil {
		return nil, http.StatusBadRequest, fmt.Errorf("invalid request ID format: %w", err)
	}

	// Validate security proof type
	if req.SecurityProofType != "valid_cert" {
		// For Phase 3 (User Story 1), we only support valid_cert
		// bootstrap_cert and tpm_attestation will be added in Phase 5 (User Story 2)
		return nil, http.StatusBadRequest, fmt.Errorf("unsupported security proof type: %s (only 'valid_cert' supported in this phase)", req.SecurityProofType)
	}

	// Parse CSR from PEM
	csrPEM := []byte(req.CSR)

	// Validate the renewal request (CSR + client certificate)
	csr, err := h.validator.ValidateRenewalRequest(csrPEM, deviceID, clientCert)
	if err != nil {
		h.log.WithError(err).WithFields(logrus.Fields{
			"device_id":  deviceID,
			"request_id": req.RequestID,
		}).Warn("Renewal request validation failed")
		return nil, http.StatusUnauthorized, fmt.Errorf("validation failed: %w", err)
	}

	// Create database record for audit trail
	dbRecord := &store.CertificateRenewalRequest{
		DeviceID:             deviceID,
		RequestID:            req.RequestID,
		RequestTime:          startTime,
		Status:               "processing",
		SecurityProofType:    req.SecurityProofType,
		OldCertificateSerial: req.OldCertificateSerial,
	}

	// TODO: Get orgID from context when multi-tenancy is properly integrated
	// For now, using empty string as orgID
	orgID := ""
	if err := h.store.CertRotation().CreateRenewalRequest(ctx, orgID, dbRecord); err != nil {
		h.log.WithError(err).Error("Failed to create renewal request record")
		// Continue with renewal even if audit logging fails
	}

	// Issue new certificate
	cert, err := h.issueCertificate(ctx, csr, deviceID)
	if err != nil {
		// Update database record with error
		dbRecord.Status = "failed"
		dbRecord.ErrorMessage = err.Error()
		dbRecord.ProcessingDurationMS = int(time.Since(startTime).Milliseconds())
		completionTime := time.Now()
		dbRecord.CompletionTime = &completionTime
		_ = h.store.CertRotation().UpdateRenewalRequest(ctx, orgID, dbRecord)

		h.log.WithError(err).WithField("device_id", deviceID).Error("Failed to issue certificate")
		return nil, http.StatusInternalServerError, fmt.Errorf("certificate issuance failed: %w", err)
	}

	// Update database record with success
	dbRecord.Status = "completed"
	dbRecord.NewCertificateSerial = hex.EncodeToString(cert.Certificate.SerialNumber.Bytes())
	dbRecord.NewCertificatePEM = string(cert.CertificatePEM)
	dbRecord.ProcessingDurationMS = int(time.Since(startTime).Milliseconds())
	completionTime := time.Now()
	dbRecord.CompletionTime = &completionTime
	if err := h.store.CertRotation().UpdateRenewalRequest(ctx, orgID, dbRecord); err != nil {
		h.log.WithError(err).Warn("Failed to update renewal request record")
		// Continue - certificate was issued successfully
	}

	h.log.WithFields(logrus.Fields{
		"device_id":     deviceID,
		"request_id":    req.RequestID,
		"serial_number": dbRecord.NewCertificateSerial,
		"duration_ms":   dbRecord.ProcessingDurationMS,
	}).Info("Certificate renewal completed successfully")

	// Build response
	response := &RenewalResponse{
		RequestID:    req.RequestID,
		Status:       "completed",
		Certificate:  string(cert.CertificatePEM),
		SerialNumber: dbRecord.NewCertificateSerial,
		NotBefore:    cert.Certificate.NotBefore,
		NotAfter:     cert.Certificate.NotAfter,
		IssuedAt:     completionTime,
	}

	return response, http.StatusOK, nil
}

// IssuedCertificate represents a newly issued certificate
type IssuedCertificate struct {
	Certificate    *x509.Certificate
	CertificatePEM []byte
}

// issueCertificate issues a new certificate based on the CSR
func (h *Handler) issueCertificate(ctx context.Context, csr *x509.CertificateRequest, deviceID string) (*IssuedCertificate, error) {
	// Encode CSR to PEM format for the signer
	csrDER := csr.Raw
	csrPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE REQUEST",
		Bytes: csrDER,
	})

	// Create sign request using the signer library
	// Use ClientBootstrapSignerName for certificate renewal (same as bootstrap enrollment)
	signReq, err := signer.NewSignRequestFromBytes(
		h.ca.Cfg.ClientBootstrapSignerName,
		csrPEM,
		signer.WithResourceName(deviceID),
	)
	if err != nil {
		return nil, fmt.Errorf("creating sign request: %w", err)
	}

	// Sign the certificate
	certPEM, err := signer.SignAsPEM(ctx, h.ca, signReq)
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

	h.log.WithFields(logrus.Fields{
		"device_id": deviceID,
		"serial":    hex.EncodeToString(cert.SerialNumber.Bytes()),
		"not_after": cert.NotAfter,
	}).Debug("Certificate issued successfully")

	return &IssuedCertificate{
		Certificate:    cert,
		CertificatePEM: certPEM,
	}, nil
}
