# Research: Agent Certificate Rotation

**Date**: 2025-12-02
**Feature**: Agent Certificate Rotation
**Purpose**: Research implementation details, identify reusable components, and validate technical approach

## Overview

This document consolidates research findings for implementing automatic certificate rotation in Flight Control. The research confirms that substantial infrastructure already exists for certificate management, CSR processing, and atomic file operations.

## Key Findings

### 1. Existing Certificate Infrastructure

#### Certificate Signing Request (CSR) Mechanisms

**Location**: `internal/tpm/csr.go`

Flight Control already has sophisticated CSR handling with TPM support:

- **TCG CSR-IDEVID Standard**: Implements Trusted Computing Group CSR format for device identity
- **Key Function**: `BuildTCGCSRIDevID()` - Creates CSR with embedded TPM attestation
- **Components**:
  - Product model and serial number
  - TPM manufacturer certificates
  - SHA256 hashing for integrity
  - Digital signature support

**Decision**: Reuse existing CSR infrastructure for renewal requests. The `BuildTCGCSRIDevID()` function can generate renewal CSRs with the same security proofs required for expired certificate recovery.

**Rationale**: Leveraging existing TCG CSR format ensures consistency with enrollment flow and provides built-in TPM attestation support for security validation.

#### Atomic File Operations

**Location**: `internal/agent/device/fileio/writer.go`

Flight Control uses `github.com/google/renameio` for atomic file writes:

- **Key Function**: `writeFileAtomically()`
- **Pattern**:
  1. Create temporary file with `renameio.TempFile()`
  2. Write content to temporary file
  3. Set permissions and ownership
  4. Atomic rename with `CloseAtomicallyReplace()`
  5. Cleanup on error with `defer t.Cleanup()`

```go
// Existing pattern (from writer.go)
t, err := renameio.TempFile(dir, path)
if err != nil {
    return err
}
defer t.Cleanup()

// Write content
_, err = t.Write(content)
if err != nil {
    return err
}

// Atomic replace
return t.CloseAtomicallyReplace()
```

**Decision**: Use existing `fileio.Writer` interface for atomic certificate swaps. This provides battle-tested atomic operations with proper error handling.

**Rationale**: The existing implementation handles edge cases (permissions, cleanup, errors) and is already used for critical agent files. Ensures consistency with existing file operations.

#### Certificate Storage

**Location**: `internal/agent/device/certmanager/provider/storage/fs.go`

Certificate storage is abstracted through a filesystem storage provider:

- **Config**: `FileSystemStorageConfig` with customizable paths
- **Permissions**: Certificates (0644), Private keys (0600)
- **Operations**: `Load()`, `Write()`, `Delete()`
- **Paths**:
  - Management cert: `/var/lib/flightctl/certs/agent.crt`
  - Management key: `/var/lib/flightctl/certs/agent.key`
  - Bootstrap cert: `/etc/flightctl/certs/client-enrollment.crt`
  - Bootstrap key: `/etc/flightctl/certs/client-enrollment.key`

**Decision**: Extend the filesystem storage provider to support certificate rotation operations (swap, backup, validate).

**Rationale**: Maintains separation of concerns and allows for future storage backend flexibility.

### 2. Retry and Error Handling

**Location**: `internal/agent/device/certmanager/retryqueue.go`

A generic retry queue implementation exists with:

- **Exponential backoff support**
- **Configurable retry handler**
- **At-least-once delivery**
- **Graceful shutdown**
- **Thread-safe operations**

```go
// Existing RetryQueue interface pattern
type Item struct {
    Value       interface{}
    RetryCount  int
    NextRetry   time.Time
}

// Handler processes items with retry capability
type Handler func(ctx context.Context, item interface{}) error
```

**Decision**: Use existing `RetryQueue` for renewal request retries. Configure with exponential backoff (initial: 1 minute, max: 24 hours).

**Rationale**: Proven implementation with proper concurrency handling. Matches requirements for retry with exponential backoff (FR-021).

**Alternatives Considered**:
- Custom retry loop: Rejected - would duplicate existing functionality
- Third-party retry library: Rejected - existing implementation is sufficient

### 3. Agent Sync Mechanism

**Location**: `internal/agent/config/config.go`

Agent has configurable sync intervals:

- **SpecFetchInterval**: 60 seconds (default) - deprecated, controlled by server
- **StatusUpdateInterval**: 60 seconds (default) - agent→service status reporting
- **MinSyncInterval**: 2 seconds (minimum allowed)

**Decision**: Certificate expiration monitoring will run as part of the existing status update cycle (every 60 seconds by default). This avoids adding a new background goroutine.

**Rationale**:
- Simpler architecture (one less background process)
- 60-second granularity is sufficient for 30-day renewal window (0.002% margin of error)
- Reduces resource usage on edge devices

**Alternatives Considered**:
- Separate monitoring goroutine: Rejected - unnecessary complexity
- Daily cron job: Rejected - requires external scheduler, less reliable

### 4. Service-Side Certificate Issuance

**Research Question**: How does the service currently issue certificates during enrollment?

**Finding**: The service uses existing CA infrastructure (location TBD during implementation).

**Decision**: Certificate renewal will reuse the same certificate issuance code path as enrollment, with modifications to:
1. Accept renewal requests from devices with expired certificates
2. Validate security proof (bootstrap cert or TPM attestation)
3. Issue new certificate with same device identity
4. Track renewal events in database

**Rationale**: Maximizes code reuse, ensures consistency between enrollment and renewal certificates.

### 5. Database Schema for Renewal Tracking

**Research Question**: What database structure is needed to track renewal requests and status?

**Decision**: Add new table `certificate_renewal_requests`:

```sql
CREATE TABLE certificate_renewal_requests (
    id                  SERIAL PRIMARY KEY,
    device_id           TEXT NOT NULL REFERENCES devices(name),
    request_time        TIMESTAMP NOT NULL DEFAULT NOW(),
    completion_time     TIMESTAMP,
    status              TEXT NOT NULL, -- pending, completed, failed
    security_proof_type TEXT NOT NULL, -- valid_cert, bootstrap_cert, tpm_attestation
    certificate_serial  TEXT,
    error_message       TEXT,
    INDEX idx_device_status (device_id, status),
    INDEX idx_request_time (request_time)
);
```

**Rationale**:
- Enables tracking renewal history per device
- Supports observability (renewal success/failure metrics)
- Allows detecting repeated failures for alerting
- Indexes optimized for common queries (device status lookups, time-based analysis)

**Alternatives Considered**:
- Add columns to existing `devices` table: Rejected - violates single responsibility, harder to track history
- No persistent tracking: Rejected - loses observability and audit trail

### 6. Configuration Schema

**Decision**: Add renewal configuration to agent config (`config.yaml`):

```yaml
certRotation:
  enabled: true                          # Enable/disable automatic rotation
  renewalThresholdDays: 30               # Trigger renewal N days before expiration
  retryInitialInterval: 1m               # Initial retry interval
  retryMaxInterval: 24h                  # Maximum retry interval
  retryBackoffMultiplier: 2.0            # Exponential backoff multiplier
```

**Rationale**:
- Follows existing config pattern in `internal/agent/config/config.go`
- Allows per-device customization if needed
- Sensible defaults match spec requirements (FR-025, FR-026, FR-027)

### 7. OpenTelemetry Tracing Integration

**Research Question**: How to instrument certificate rotation with OpenTelemetry?

**Finding**: Flight Control has existing tracing infrastructure in `internal/tracing/`.

**Decision**: Use existing tracing patterns:

```go
ctx, span := tracing.StartSpan(ctx, "flightctl/agent/certrotation", "RenewCertificate")
defer span.End()

span.SetAttributes(
    attribute.String("device.id", deviceID),
    attribute.Int64("cert.expiration", expirationTime.Unix()),
    attribute.Int("cert.daysUntilExpiry", daysRemaining),
)
```

**Operations to trace**:
- `MonitorCertificateExpiration` - Expiration checking
- `GenerateRenewalRequest` - CSR creation
- `SubmitRenewalRequest` - HTTP request to service
- `ValidateNewCertificate` - Certificate validation
- `SwapCertificate` - Atomic swap operation
- `RecoverFromExpiredCert` - Expired certificate recovery flow

**Rationale**: Consistent with constitution requirement (Observability is Non-Negotiable) and existing codebase patterns.

## Implementation Approach Summary

### Agent-Side Components

1. **Certificate Expiration Monitor** (`internal/agent/device/certrotation/monitor.go`)
   - Runs during status update cycle (every 60s)
   - Parses X.509 certificate `NotAfter` field
   - Compares against configured threshold (default: 30 days)
   - Triggers renewal when threshold reached

2. **Renewal Request Generator** (`internal/agent/device/certrotation/renewer.go`)
   - Reuses existing CSR infrastructure (`internal/tpm/csr.go`)
   - Generates TCG CSR-IDEVID for renewal
   - Includes security proof (valid cert, bootstrap cert, or TPM attestation)
   - Submits to new service endpoint

3. **Atomic Certificate Swap** (`internal/agent/device/certrotation/atomic.go`)
   - Uses existing `fileio.Writer` for atomic operations
   - Pattern:
     1. Write new cert/key to temp location
     2. Validate new certificate (parse, verify signature, check expiration)
     3. Atomic rename to production location
     4. Keep old cert as backup (`.old` suffix) until validation complete
     5. Remove backup only after new cert is confirmed operational

4. **Recovery Handler** (`internal/agent/device/certrotation/recovery.go`)
   - Detects expired management certificate on startup
   - Falls back to bootstrap certificate authentication
   - If bootstrap also expired, uses TPM attestation
   - Generates recovery CSR with security proof
   - Resumes normal operations after new certificate installed

### Service-Side Components

1. **Renewal Endpoint Handler** (`internal/service/certrotation/handler.go`)
   - New HTTP POST endpoint: `/api/v1beta1/devices/{name}/certificaterenewal`
   - Accepts CSR with security proof
   - Validates device identity
   - Routes to certificate issuer

2. **Security Proof Validator** (`internal/service/certrotation/validator.go`)
   - Validates three authentication methods:
     - **Valid certificate**: Device presents current (not expired) management cert
     - **Bootstrap certificate**: Device presents valid bootstrap cert (fallback)
     - **TPM attestation**: Device provides TPM attestation proof
   - Verifies device identity matches CSR subject
   - Checks device exists in database and is not revoked

3. **Certificate Issuer** (`internal/service/certrotation/issuer.go`)
   - Reuses existing CA code from enrollment flow
   - Issues new X.509 certificate with 365-day validity
   - Records renewal event in database
   - Returns signed certificate to device

4. **Database Layer** (`internal/service/store/certrotation.go`)
   - CRUD operations for `certificate_renewal_requests` table
   - Query methods for metrics (success rate, failure tracking)
   - Transaction support for renewal state updates

## Technical Risks and Mitigations

### Risk 1: Certificate Swap Failure Leaves Device Unreachable

**Mitigation**:
- Use atomic file operations (existing `renameio` library)
- Validate new certificate before removing old one
- Keep old certificate as backup (`.old` suffix)
- Implement rollback on validation failure
- Integration tests simulating power loss during swap

### Risk 2: Clock Skew Causes Premature/Late Renewal

**Mitigation**:
- Document dependency on NTP time synchronization (already in spec Assumptions)
- Log warnings if certificate expiration appears abnormal
- Service validates certificate expiration server-side
- Consider clock skew in expiration calculations (e.g., add 1-hour buffer)

### Risk 3: Expired Bootstrap Certificate Prevents Recovery

**Mitigation**:
- Bootstrap certificates have longer validity (documented assumption)
- TPM attestation provides third authentication method
- Document requirement for fleet administrators to monitor bootstrap cert expiration
- Future enhancement: automatic bootstrap cert renewal (out of scope for this feature)

### Risk 4: Fleet-Wide Renewal Storm Overwhelms Service

**Mitigation**:
- Stagger renewals with jitter (randomize renewal time within threshold window)
- Service implements rate limiting (existing capability in Flight Control)
- Database indexes optimized for high-volume renewal queries
- Load test with 10,000 concurrent renewal requests (performance goal)

## Open Questions (Resolved)

### Q1: Should renewal trigger on every sync, or use a separate timer?

**Resolution**: Integrate with existing status update cycle (60-second interval). This avoids additional background goroutines and is sufficient granularity for 30-day renewal window.

### Q2: How to handle devices that have never synced and certificate expires?

**Resolution**: Out of scope. Device must sync at least once before expiration to trigger renewal. If device remains offline for full validity period (365 days), recovery via bootstrap cert or TPM attestation is required.

### Q3: Should old certificates be kept for audit purposes?

**Resolution**: Yes, but with automatic cleanup. Keep old certificate with `.old` suffix until new certificate is validated. Then delete old certificate to conserve disk space. Database tracks renewal history for audit.

## Dependencies

### External Libraries

- `github.com/google/renameio` - Atomic file operations (already in use)
- `github.com/google/go-tpm` - TPM operations (already in use)
- `github.com/google/go-tpm-tools` - TPM utilities (already in use)

### Internal Components

- `internal/tpm/csr.go` - CSR generation (reuse)
- `internal/agent/device/fileio` - Atomic file writes (reuse)
- `internal/agent/device/certmanager/retryqueue.go` - Retry mechanism (reuse)
- `internal/crypto/cert.go` - Certificate utilities (may need extensions)
- `internal/tracing` - OpenTelemetry integration (reuse)

### Database

- PostgreSQL 16+ (existing)
- New migration: `YYYYMMDDHHMMSS_add_cert_renewal.sql`

## Next Steps

Phase 1 will produce:
- `data-model.md` - Entity definitions for renewal requests, certificate metadata
- `contracts/` - OpenAPI specification for renewal endpoint
- `quickstart.md` - Developer guide for testing certificate rotation locally
