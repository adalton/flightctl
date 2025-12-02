package store

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// CertificateRenewalRequest represents a certificate renewal request in the database
type CertificateRenewalRequest struct {
	ID                   int        `gorm:"primaryKey;autoIncrement"`
	DeviceID             string     `gorm:"column:device_id;not null;index:idx_cert_renewal_device_status"`
	RequestID            string     `gorm:"column:request_id;not null;uniqueIndex:idx_cert_renewal_request_id"`
	RequestTime          time.Time  `gorm:"column:request_time;not null;default:CURRENT_TIMESTAMP;index:idx_cert_renewal_request_time,sort:desc"`
	CompletionTime       *time.Time `gorm:"column:completion_time"`
	Status               string     `gorm:"column:status;not null;index:idx_cert_renewal_device_status"`
	SecurityProofType    string     `gorm:"column:security_proof_type;not null"`
	OldCertificateSerial string     `gorm:"column:old_certificate_serial"`
	NewCertificateSerial string     `gorm:"column:new_certificate_serial"`
	NewCertificatePEM    string     `gorm:"column:new_certificate_pem;type:text"`
	ClientIP             string     `gorm:"column:client_ip"`
	ErrorMessage         string     `gorm:"column:error_message;type:text"`
	ProcessingDurationMS int        `gorm:"column:processing_duration_ms"`
}

// TableName specifies the table name for GORM
func (CertificateRenewalRequest) TableName() string {
	return "certificate_renewal_requests"
}

// CertRotation defines the interface for certificate rotation database operations
type CertRotation interface {
	CreateRenewalRequest(ctx context.Context, orgID string, req *CertificateRenewalRequest) error
	UpdateRenewalRequest(ctx context.Context, orgID string, req *CertificateRenewalRequest) error
	GetRenewalRequest(ctx context.Context, orgID string, requestID string) (*CertificateRenewalRequest, error)
	ListRenewalRequests(ctx context.Context, orgID string, deviceID string, limit int) ([]*CertificateRenewalRequest, error)
}

type certRotationStore struct {
	db  *gorm.DB
	log logrus.FieldLogger
}

// NewCertRotation creates a new certificate rotation store instance
func NewCertRotation(db *gorm.DB, log logrus.FieldLogger) CertRotation {
	return &certRotationStore{
		db:  db,
		log: log,
	}
}

// CreateRenewalRequest creates a new renewal request record in the database
func (s *certRotationStore) CreateRenewalRequest(ctx context.Context, orgID string, req *CertificateRenewalRequest) error {
	result := s.db.WithContext(ctx).Create(req)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// UpdateRenewalRequest updates an existing renewal request in the database
func (s *certRotationStore) UpdateRenewalRequest(ctx context.Context, orgID string, req *CertificateRenewalRequest) error {
	result := s.db.WithContext(ctx).
		Model(&CertificateRenewalRequest{}).
		Where("request_id = ?", req.RequestID).
		Updates(req)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// GetRenewalRequest retrieves a renewal request by request ID
func (s *certRotationStore) GetRenewalRequest(ctx context.Context, orgID string, requestID string) (*CertificateRenewalRequest, error) {
	var req CertificateRenewalRequest
	result := s.db.WithContext(ctx).
		Where("request_id = ?", requestID).
		First(&req)
	if result.Error != nil {
		return nil, result.Error
	}
	return &req, nil
}

// ListRenewalRequests lists renewal requests for a device with a limit
func (s *certRotationStore) ListRenewalRequests(ctx context.Context, orgID string, deviceID string, limit int) ([]*CertificateRenewalRequest, error) {
	var requests []*CertificateRenewalRequest
	query := s.db.WithContext(ctx).
		Where("device_id = ?", deviceID).
		Order("request_time DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	result := query.Find(&requests)
	if result.Error != nil {
		return nil, result.Error
	}
	return requests, nil
}
