package certrotation

import (
	"time"
)

// CertificateMetadata represents certificate information tracked by the agent
type CertificateMetadata struct {
	FilePath     string
	NotBefore    time.Time
	NotAfter     time.Time
	Subject      string
	Issuer       string
	SerialNumber string
	Fingerprint  string
}

// RenewalRequest represents a certificate renewal request being processed by the agent
type RenewalRequest struct {
	RequestID            string
	DeviceID             string
	OldCertificateSerial string
	CSR                  []byte
	SecurityProofType    SecurityProofType
	CreatedAt            time.Time
	RetryCount           int
	NextRetryAt          time.Time
	Status               RenewalStatus
}

// SecurityProofType defines the authentication method for certificate renewal
type SecurityProofType string

const (
	ValidCertificate     SecurityProofType = "valid_cert"
	BootstrapCertificate SecurityProofType = "bootstrap_cert"
	TPMAttestation       SecurityProofType = "tpm_attestation"
)

// RenewalStatus represents the state of a renewal request
type RenewalStatus string

const (
	StatusPending   RenewalStatus = "pending"
	StatusSubmitted RenewalStatus = "submitted"
	StatusCompleted RenewalStatus = "completed"
	StatusFailed    RenewalStatus = "failed"
)
