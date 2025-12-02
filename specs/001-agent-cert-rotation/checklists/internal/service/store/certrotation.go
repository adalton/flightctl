package store

import (
	"time"
)

// CertificateRenewalRequest represents a certificate renewal request in the database
type CertificateRenewalRequest struct {
	// ID is the auto-incrementing primary key
	ID int `db:"id"`
	// DeviceID is the reference to the device (foreign key)
	DeviceID string `db:"device_id"`
	// RequestID is the UUID from the agent's renewal request (for correlation)
	RequestID string `db:"request_id"`
	// RequestTime is when the request was received by the service
	RequestTime time.Time `db:"request_time"`
	// CompletionTime is when processing finished (success or failure)
	CompletionTime *time.Time `db:"completion_time"`
	// Status is the current processing state
	Status string `db:"status"`
	// SecurityProofType is the authentication method used
	SecurityProofType string `db:"security_proof_type"`
	// OldCertificateSerial is the serial of the certificate being renewed
	OldCertificateSerial string `db:"old_certificate_serial"`
	// NewCertificateSerial is the serial of the newly issued certificate (if successful)
	NewCertificateSerial string `db:"new_certificate_serial"`
	// NewCertificatePEM is the PEM-encoded new certificate (for retrieval)
	NewCertificatePEM string `db:"new_certificate_pem"`
	// ClientIP is the IP address of the requesting device (audit)
	ClientIP string `db:"client_ip"`
	// ErrorMessage contains error details if renewal failed
	ErrorMessage string `db:"error_message"`
	// ProcessingDurationMS is the time taken to process the request (metrics)
	ProcessingDurationMS int `db:"processing_duration_ms"`
}

// CertRotationStore defines the interface for certificate rotation database operations
type CertRotationStore interface {
	// CreateRenewalRequest creates a new renewal request record
	CreateRenewalRequest(req *CertificateRenewalRequest) error
	// UpdateRenewalRequest updates an existing renewal request
	UpdateRenewalRequest(req *CertificateRenewalRequest) error
	// GetRenewalRequest retrieves a renewal request by request ID
	GetRenewalRequest(requestID string) (*CertificateRenewalRequest, error)
	// ListRenewalRequests lists renewal requests for a device
	ListRenewalRequests(deviceID string, limit int) ([]*CertificateRenewalRequest, error)
}
