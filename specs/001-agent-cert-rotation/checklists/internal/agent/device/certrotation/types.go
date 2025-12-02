package certrotation

import (
	"time"
)

// CertificateMetadata represents certificate information tracked by the agent
type CertificateMetadata struct {
	// FilePath is the absolute path to the certificate file
	FilePath string
	// NotBefore is the certificate validity start time
	NotBefore time.Time
	// NotAfter is the certificate validity end time
	NotAfter time.Time
	// Subject is the certificate subject (device identity)
	Subject string
	// Issuer is the certificate issuer (CA)
	Issuer string
	// SerialNumber is the certificate serial number (hex-encoded)
	SerialNumber string
	// Fingerprint is the SHA256 fingerprint of the certificate
	Fingerprint string
}

// RenewalRequest represents a certificate renewal request being processed by the agent
type RenewalRequest struct {
	// RequestID is a unique identifier for this renewal attempt (UUID)
	RequestID string
	// DeviceID is the device identifier
	DeviceID string
	// OldCertificateSerial is the serial number of the certificate being renewed
	OldCertificateSerial string
	// CSR is the PEM-encoded Certificate Signing Request
	CSR []byte
	// SecurityProofType is the authentication method used
	SecurityProofType SecurityProofType
	// CreatedAt is when the renewal request was generated
	CreatedAt time.Time
	// RetryCount is the number of retry attempts
	RetryCount int
	// NextRetryAt is when to retry if the previous attempt failed
	NextRetryAt time.Time
	// Status is the current status of the renewal request
	Status RenewalStatus
}

// SecurityProofType defines the authentication method for certificate renewal
type SecurityProofType string

const (
	// ValidCertificate indicates using current valid management certificate
	ValidCertificate SecurityProofType = "valid_cert"
	// BootstrapCertificate indicates using bootstrap certificate (fallback)
	BootstrapCertificate SecurityProofType = "bootstrap_cert"
	// TPMAttestation indicates using TPM attestation (expired cert recovery)
	TPMAttestation SecurityProofType = "tpm_attestation"
)

// RenewalStatus represents the state of a renewal request
type RenewalStatus string

const (
	// StatusPending indicates the request is waiting to be sent
	StatusPending RenewalStatus = "pending"
	// StatusSubmitted indicates the request has been sent to the service
	StatusSubmitted RenewalStatus = "submitted"
	// StatusCompleted indicates the new certificate has been received and installed
	StatusCompleted RenewalStatus = "completed"
	// StatusFailed indicates the renewal failed (terminal state)
	StatusFailed RenewalStatus = "failed"
)
