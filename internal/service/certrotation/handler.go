package certrotation

import (
	"context"
	"crypto/x509"
	"fmt"
	"net/http"
	"time"

	"github.com/flightctl/flightctl/internal/crypto"
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
	validator *Validator
	issuer    *CertificateIssuer
	log       *logrus.Logger
}

// NewHandler creates a new certificate renewal handler
func NewHandler(store store.Store, ca *crypto.CAClient, validator *Validator, log *logrus.Logger) *Handler {
	return &Handler{
		store:     store,
		validator: validator,
		issuer:    NewCertificateIssuer(ca, log),
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
	cert, err := h.issuer.IssueCertificate(ctx, csr, deviceID)
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
	dbRecord.NewCertificateSerial = cert.SerialNumber
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
		SerialNumber: cert.SerialNumber,
		NotBefore:    cert.Certificate.NotBefore,
		NotAfter:     cert.Certificate.NotAfter,
		IssuedAt:     completionTime,
	}

	return response, http.StatusOK, nil
}
