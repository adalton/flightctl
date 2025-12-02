# Data Model: Agent Certificate Rotation

**Date**: 2025-12-02
**Feature**: Agent Certificate Rotation
**Purpose**: Define data structures, entities, and database schema for certificate rotation

## Overview

This document defines the data model for automatic certificate rotation, including agent-side structures, service-side entities, and database schema.

## Entity Definitions

### 1. Certificate Metadata

**Purpose**: Represents certificate information tracked by the agent

**Attributes**:
- `FilePath` (string): Absolute path to certificate file (e.g., `/var/lib/flightctl/certs/agent.crt`)
- `NotBefore` (time.Time): Certificate validity start time
- `NotAfter` (time.Time): Certificate validity end time
- `Subject` (string): Certificate subject (device identity)
- `Issuer` (string): Certificate issuer (CA)
- `SerialNumber` (string): Certificate serial number (hex-encoded)
- `Fingerprint` (string): SHA256 fingerprint of certificate

**Relationships**:
- One-to-one with device identity
- Referenced by Renewal Request for tracking

**Validation Rules**:
- `NotAfter` must be after `NotBefore`
- `FilePath` must be absolute path
- `SerialNumber` must be unique
- `Fingerprint` calculated as SHA256(DER-encoded certificate)

**State Transitions**:
```
[Valid] → threshold reached → [ExpiringS oon]
[ExpiringSoon] → renewal successful → [Valid]
[ExpiringSoon] → expiration time reached → [Expired]
[Expired] → recovery successful → [Valid]
```

---

### 2. Renewal Request (Agent-Side)

**Purpose**: Represents a certificate renewal request being processed by the agent

**Attributes**:
- `RequestID` (string, UUID): Unique identifier for this renewal attempt
- `DeviceID` (string): Device identifier
- `OldCertificateSerial` (string): Serial number of certificate being renewed
- `CSR` ([]byte): PEM-encoded Certificate Signing Request
- `SecurityProofType` (enum): Authentication method used
  - `ValidCertificate`: Using current valid management certificate
  - `BootstrapCertificate`: Using bootstrap certificate (fallback)
  - `TPMAttestation`: Using TPM attestation (expired cert recovery)
- `CreatedAt` (time.Time): When renewal request was generated
- `RetryCount` (int): Number of retry attempts
- `NextRetryAt` (time.Time): When to retry if previous attempt failed
- `Status` (enum):
  - `Pending`: Waiting to be sent
  - `Submitted`: Sent to service, awaiting response
  - `Completed`: New certificate received and installed
  - `Failed`: Renewal failed (terminal state)

**Relationships**:
- References Certificate Metadata (old certificate)
- May result in new Certificate Metadata (new certificate)

**Validation Rules**:
- `CSR` must be valid PEM-encoded CSR
- `SecurityProofType` must match available credentials:
  - `ValidCertificate` only if management cert not expired
  - `BootstrapCertificate` only if bootstrap cert exists and valid
  - `TPMAttestation` only if TPM available and enabled
- `RetryCount` increments on each failure
- `NextRetryAt` calculated with exponential backoff

---

### 3. Certificate Renewal Request (Service-Side Database Entity)

**Purpose**: Persistent record of certificate renewal requests processed by the service

**Database Table**: `certificate_renewal_requests`

**Schema**:
```sql
CREATE TABLE certificate_renewal_requests (
    id                   SERIAL PRIMARY KEY,
    device_id            TEXT NOT NULL,
    request_id           UUID NOT NULL UNIQUE,
    request_time         TIMESTAMP NOT NULL DEFAULT NOW(),
    completion_time      TIMESTAMP,
    status               TEXT NOT NULL CHECK (status IN ('pending', 'processing', 'completed', 'failed')),
    security_proof_type  TEXT NOT NULL CHECK (security_proof_type IN ('valid_cert', 'bootstrap_cert', 'tpm_attestation')),

    -- Certificate information
    old_certificate_serial TEXT,
    new_certificate_serial TEXT,
    new_certificate_pem    TEXT,

    -- Audit and debugging
    client_ip              TEXT,
    error_message          TEXT,
    processing_duration_ms INTEGER,

    -- Indexes for performance
    INDEX idx_device_status (device_id, status),
    INDEX idx_request_time (request_time DESC),
    INDEX idx_request_id (request_id),

    -- Foreign key
    FOREIGN KEY (device_id) REFERENCES devices(name) ON DELETE CASCADE
);
```

**Attributes**:
- `id`: Auto-incrementing primary key
- `device_id`: Reference to device (foreign key)
- `request_id`: UUID from agent's renewal request (for correlation)
- `request_time`: When request was received by service
- `completion_time`: When processing finished (success or failure)
- `status`: Current processing state
- `security_proof_type`: Authentication method used
- `old_certificate_serial`: Serial of certificate being renewed
- `new_certificate_serial`: Serial of newly issued certificate (if successful)
- `new_certificate_pem`: PEM-encoded new certificate (for retrieval)
- `client_ip`: IP address of requesting device (audit)
- `error_message`: Error details if renewal failed
- `processing_duration_ms`: Time taken to process request (metrics)

**Relationships**:
- Many-to-one with Device (multiple renewal requests per device over time)
- Each request associated with one old certificate and optionally one new certificate

**Validation Rules**:
- `status` must be one of: pending, processing, completed, failed
- `security_proof_type` must be one of: valid_cert, bootstrap_cert, tpm_attestation
- `completion_time` must be after `request_time` if set
- `new_certificate_serial` required if status is 'completed'
- `error_message` required if status is 'failed'

**Indexes**:
- `idx_device_status`: Fast lookup of device's recent renewal requests
- `idx_request_time`: Time-based queries for metrics and reporting
- `idx_request_id`: Agent correlation lookup

---

### 4. Certificate Rotation Configuration

**Purpose**: Agent configuration for certificate rotation behavior

**Structure** (YAML in `config.yaml`):
```yaml
certRotation:
  enabled: true
  renewalThresholdDays: 30
  retryInitialInterval: 1m
  retryMaxInterval: 24h
  retryBackoffMultiplier: 2.0
  monitorIntervalSeconds: 60
```

**Attributes**:
- `enabled` (bool): Enable/disable automatic certificate rotation
- `renewalThresholdDays` (int): Days before expiration to trigger renewal (default: 30)
- `retryInitialInterval` (duration): Initial retry interval (default: 1m)
- `retryMaxInterval` (duration): Maximum retry interval (default: 24h)
- `retryBackoffMultiplier` (float64): Exponential backoff multiplier (default: 2.0)
- `monitorIntervalSeconds` (int): How often to check certificate expiration (default: 60)

**Validation Rules**:
- `renewalThresholdDays` must be > 0 and < certificate validity period (365 days)
- `retryInitialInterval` must be >= 1 second
- `retryMaxInterval` must be >= `retryInitialInterval`
- `retryBackoffMultiplier` must be >= 1.0
- `monitorIntervalSeconds` must be >= 1

---

## Data Flows

### Flow 1: Proactive Certificate Renewal

```
1. Agent monitors certificate expiration (every 60s)
   └─> Reads agent.crt, parses NotAfter field

2. Expiration threshold reached (30 days before expiry)
   └─> Creates RenewalRequest entity (agent-side)
   └─> Generates CSR using existing TPM infrastructure
   └─> SecurityProofType = ValidCertificate

3. Agent submits renewal request to service
   └─> HTTP POST /api/v1beta1/devices/{name}/certificaterenewal
   └─> Body: { requestID, csr, securityProofType }

4. Service creates CertificateRenewalRequest record (database)
   └─> Status = 'processing'
   └─> Validates security proof
   └─> Issues new certificate

5. Service updates database record
   └─> Status = 'completed'
   └─> Stores new_certificate_serial and new_certificate_pem

6. Service returns new certificate to agent
   └─> Response: { certificate, key, serialNumber }

7. Agent performs atomic certificate swap
   └─> Validates new certificate
   └─> Writes to temp file (agent.crt.tmp)
   └─> Atomic rename (agent.crt.tmp → agent.crt)
   └─> Updates RenewalRequest status = 'Completed'
```

### Flow 2: Expired Certificate Recovery

```
1. Device comes online after extended offline period
   └─> Agent startup detects agent.crt is expired
   └─> Checks NotAfter < time.Now()

2. Agent attempts recovery
   └─> Creates RenewalRequest entity (agent-side)
   └─> Checks bootstrap certificate validity

   IF bootstrap cert valid:
     └─> SecurityProofType = BootstrapCertificate
     └─> Uses bootstrap cert for authentication

   ELSE IF TPM available:
     └─> SecurityProofType = TPMAttestation
     └─> Generates TPM attestation proof

   ELSE:
     └─> Terminal failure (manual intervention required)

3. Agent submits recovery request to service
   └─> HTTP POST /api/v1beta1/devices/{name}/certificaterenewal
   └─> Authenticates using bootstrap cert or TPM proof

4. Service validates recovery request
   └─> Verifies device identity
   └─> Validates security proof (bootstrap cert signature or TPM attestation)
   └─> Checks device is not revoked

5. Service issues new certificate (same flow as proactive renewal)
   └─> Creates CertificateRenewalRequest record
   └─> Issues new management certificate
   └─> Returns to agent

6. Agent installs new certificate and resumes normal operations
   └─> Atomic swap to agent.crt
   └─> Reconnect to service with new certificate
   └─> Resume status updates
```

### Flow 3: Renewal Failure and Retry

```
1. Renewal request fails (e.g., service unavailable, network error)
   └─> RenewalRequest.Status = 'Failed'
   └─> RenewalRequest.RetryCount++
   └─> Calculate NextRetryAt = Now() + backoff_delay

2. Backoff calculation:
   delay = min(
     retryInitialInterval * (backoffMultiplier ^ RetryCount),
     retryMaxInterval
   )

   Example:
   Attempt 1: 1 minute
   Attempt 2: 2 minutes
   Attempt 3: 4 minutes
   Attempt 4: 8 minutes
   ...
   Attempt 11+: 24 hours (capped at retryMaxInterval)

3. Retry queue reschedules request
   └─> Wait until NextRetryAt
   └─> Retry submission (goto Flow 1, step 3)

4. Continue retrying until success
   └─> No maximum retry limit (unlimited retries with backoff)
   └─> Device will eventually succeed when service becomes available
```

---

## Database Migrations

### Migration: Add Certificate Renewal Tracking

**File**: `db/migrations/YYYYMMDDHHMMSS_add_cert_renewal.sql`

**UP Migration**:
```sql
-- Create certificate_renewal_requests table
CREATE TABLE IF NOT EXISTS certificate_renewal_requests (
    id                   SERIAL PRIMARY KEY,
    device_id            TEXT NOT NULL,
    request_id           UUID NOT NULL UNIQUE,
    request_time         TIMESTAMP NOT NULL DEFAULT NOW(),
    completion_time      TIMESTAMP,
    status               TEXT NOT NULL CHECK (status IN ('pending', 'processing', 'completed', 'failed')),
    security_proof_type  TEXT NOT NULL CHECK (security_proof_type IN ('valid_cert', 'bootstrap_cert', 'tpm_attestation')),
    old_certificate_serial TEXT,
    new_certificate_serial TEXT,
    new_certificate_pem    TEXT,
    client_ip              TEXT,
    error_message          TEXT,
    processing_duration_ms INTEGER,

    FOREIGN KEY (device_id) REFERENCES devices(name) ON DELETE CASCADE
);

-- Create indexes for performance
CREATE INDEX idx_cert_renewal_device_status ON certificate_renewal_requests(device_id, status);
CREATE INDEX idx_cert_renewal_request_time ON certificate_renewal_requests(request_time DESC);
CREATE INDEX idx_cert_renewal_request_id ON certificate_renewal_requests(request_id);

-- Add comment for documentation
COMMENT ON TABLE certificate_renewal_requests IS 'Tracks certificate renewal requests from devices for audit and observability';
```

**DOWN Migration**:
```sql
-- Drop indexes first
DROP INDEX IF EXISTS idx_cert_renewal_request_id;
DROP INDEX IF EXISTS idx_cert_renewal_request_time;
DROP INDEX IF EXISTS idx_cert_renewal_device_status;

-- Drop table
DROP TABLE IF EXISTS certificate_renewal_requests;
```

**Testing Strategy**:
- Test with 1,000 devices (small fleet)
- Test with 10,000 devices (medium fleet)
- Test with 100,000 devices (large fleet)
- Measure index performance on queries:
  - List renewals for device
  - Count successful renewals in last 30 days
  - Find failed renewals for alerting

---

## Go Type Definitions

### Agent-Side Types

```go
// CertificateMetadata represents certificate information
type CertificateMetadata struct {
    FilePath     string
    NotBefore    time.Time
    NotAfter     time.Time
    Subject      string
    Issuer       string
    SerialNumber string
    Fingerprint  string
}

// RenewalRequest represents an in-flight renewal request
type RenewalRequest struct {
    RequestID           string
    DeviceID            string
    OldCertificateSerial string
    CSR                 []byte
    SecurityProofType   SecurityProofType
    CreatedAt           time.Time
    RetryCount          int
    NextRetryAt         time.Time
    Status              RenewalStatus
}

// SecurityProofType enum
type SecurityProofType string

const (
    ValidCertificate     SecurityProofType = "valid_cert"
    BootstrapCertificate SecurityProofType = "bootstrap_cert"
    TPMAttestation       SecurityProofType = "tpm_attestation"
)

// RenewalStatus enum
type RenewalStatus string

const (
    StatusPending   RenewalStatus = "pending"
    StatusSubmitted RenewalStatus = "submitted"
    StatusCompleted RenewalStatus = "completed"
    StatusFailed    RenewalStatus = "failed"
)

// CertRotationConfig configuration structure
type CertRotationConfig struct {
    Enabled                 bool          `json:"enabled"`
    RenewalThresholdDays    int           `json:"renewalThresholdDays"`
    RetryInitialInterval    time.Duration `json:"retryInitialInterval"`
    RetryMaxInterval        time.Duration `json:"retryMaxInterval"`
    RetryBackoffMultiplier  float64       `json:"retryBackoffMultiplier"`
    MonitorIntervalSeconds  int           `json:"monitorIntervalSeconds"`
}
```

### Service-Side Types

```go
// CertificateRenewalRequest database entity
type CertificateRenewalRequest struct {
    ID                   int       `db:"id"`
    DeviceID             string    `db:"device_id"`
    RequestID            string    `db:"request_id"`
    RequestTime          time.Time `db:"request_time"`
    CompletionTime       *time.Time `db:"completion_time"`
    Status               string    `db:"status"`
    SecurityProofType    string    `db:"security_proof_type"`
    OldCertificateSerial string    `db:"old_certificate_serial"`
    NewCertificateSerial string    `db:"new_certificate_serial"`
    NewCertificatePEM    string    `db:"new_certificate_pem"`
    ClientIP             string    `db:"client_ip"`
    ErrorMessage         string    `db:"error_message"`
    ProcessingDurationMS int       `db:"processing_duration_ms"`
}
```

---

## Metrics and Observability

### Metrics to Emit

**Agent-Side**:
- `flightctl_agent_cert_expiration_time_seconds` (gauge): Certificate expiration timestamp
- `flightctl_agent_cert_renewal_attempts_total` (counter): Total renewal attempts
- `flightctl_agent_cert_renewal_successes_total` (counter): Successful renewals
- `flightctl_agent_cert_renewal_failures_total` (counter): Failed renewals
- `flightctl_agent_cert_rotation_duration_seconds` (histogram): Time to complete rotation

**Service-Side**:
- `flightctl_service_cert_renewal_requests_total` (counter): Total renewal requests received
- `flightctl_service_cert_renewal_issued_total` (counter): Certificates issued
- `flightctl_service_cert_renewal_rejected_total` (counter): Requests rejected (validation failed)
- `flightctl_service_cert_renewal_processing_duration_seconds` (histogram): Request processing time

### Structured Logging Events

**Agent-Side**:
- `certificate.expiration.detected`: Certificate expiring soon detected
- `certificate.renewal.initiated`: Renewal request created
- `certificate.renewal.submitted`: Request sent to service
- `certificate.renewal.completed`: New certificate installed
- `certificate.renewal.failed`: Renewal attempt failed
- `certificate.recovery.initiated`: Expired certificate recovery started
- `certificate.swap.success`: Atomic swap completed successfully
- `certificate.swap.rollback`: Swap failed, rolled back to old certificate

**Service-Side**:
- `certificate.renewal.request.received`: Renewal request received
- `certificate.renewal.security_proof.validated`: Security proof validated
- `certificate.renewal.issued`: New certificate issued
- `certificate.renewal.failed`: Request processing failed

---

## Summary

This data model provides:
- **Agent-side** structures for tracking certificate metadata and renewal requests
- **Service-side** database schema for audit trail and observability
- **Configuration** schema for customizing renewal behavior
- **Data flows** illustrating lifecycle of renewal and recovery
- **Database migrations** with rollback support and performance testing strategy
- **Type definitions** for Go implementation
- **Observability** metrics and logging events per constitution requirements

All entities align with functional requirements (FR-001 through FR-032) and support the three user stories (proactive renewal, expired recovery, atomic rotation).
