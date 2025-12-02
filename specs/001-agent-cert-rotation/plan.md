# Implementation Plan: Agent Certificate Rotation

**Branch**: `001-agent-cert-rotation` | **Date**: 2025-12-02 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `specs/001-agent-cert-rotation/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

Implement automatic certificate lifecycle management for Flight Control edge devices, enabling:
- **Proactive renewal**: Devices automatically renew management certificates 30 days before expiration
- **Expired certificate recovery**: Devices that have been offline can recover using bootstrap certificates or TPM attestation
- **Atomic rotation**: Certificate swaps are atomic (all-or-nothing) to prevent devices from becoming unreachable

This eliminates manual certificate re-enrollment for fleet administrators and ensures devices can operate continuously without certificate-related service disruptions.

## Technical Context

**Language/Version**: Go 1.24+ (per go.mod)
**Primary Dependencies**:
- `github.com/google/go-tpm` and `github.com/google/go-tpm-tools` for TPM operations
- `github.com/google/renameio` for atomic file operations
- OpenTelemetry for tracing (existing)
- PostgreSQL for certificate request tracking (service-side)

**Storage**:
- Agent-side: Filesystem (`/var/lib/flightctl/certs/` and `/etc/flightctl/certs/`)
  - Management certificate: `agent.crt` and `agent.key`
  - Bootstrap certificate: `client-enrollment.crt` and `client-enrollment.key`
  - CA bundle: `ca.crt`
  - TPM keys: `tpm-blob.yaml`
- Service-side: PostgreSQL for tracking certificate metadata and renewal status

**Testing**: Go standard testing + Ginkgo/Gomega for integration tests
**Target Platform**: Linux edge devices (systemd-based, TPM 2.0 optional)
**Project Type**: Distributed system (agent + API service)

**Performance Goals**:
- Certificate expiration check: < 100ms per agent sync cycle
- Renewal request processing: < 5 seconds end-to-end (agent request → service response)
- Atomic certificate swap: < 500ms (minimize service interruption window)
- Support 10,000+ concurrent renewal requests during fleet-wide renewal events

**Constraints**:
- **Offline-capable**: Devices may be disconnected for extended periods (weeks/months)
- **No external dependencies**: Cannot rely on external PKI or certificate authorities
- **Atomic operations**: Must use atomic file operations to prevent partial certificate states
- **Minimal resource usage**: Edge devices may have limited CPU/memory
- **Backward compatible**: Must work with existing enrollment flow and certificate formats

**Scale/Scope**:
- Target fleet size: 10,000 - 100,000 devices
- Certificate validity: 365 days (existing behavior)
- Renewal window: 30 days before expiration (configurable)
- Maximum retry attempts: Unlimited with exponential backoff (cap at 24 hours between attempts)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### ✅ I. Distributed Systems First

**How this handles network partitioning**:
- Certificate expiration monitoring runs locally on device (no network required for detection)
- Renewal requests use retry with exponential backoff when service unavailable
- Devices retain existing valid certificates during network outages
- Recovery mechanism supports devices that have been offline beyond certificate expiration
- No single point of failure - each device manages its own certificate lifecycle

**Compliance**: PASS - Feature is explicitly designed for edge devices with intermittent connectivity

### ✅ II. Observability is Non-Negotiable

**Tracing and metrics**:
- All certificate operations will use `tracing.StartSpan(ctx, "flightctl/agent", operation)` for:
  - Certificate expiration monitoring
  - Renewal request generation
  - Certificate validation
  - Atomic swap operations
- Metrics to emit (per FR-028 through FR-032):
  - Certificate expiration dates
  - Renewal attempts (count, success/failure rate)
  - Service-side renewal request processing (received, validated, issued, rejected)
- Structured logging for all lifecycle events (initiation, progress, completion, failure)

**Compliance**: PASS - Comprehensive observability planned per spec requirements

### ✅ III. Test-Driven Development (TDD)

**Test coverage planned**:
- **Contract tests**: CSR renewal endpoint API contract (request/response format, authentication)
- **Integration tests**: Full renewal flow (agent → service → certificate issuance)
- **Integration tests**: Expired certificate recovery flow (bootstrap cert fallback, TPM attestation)
- **Integration tests**: Atomic swap operations (failure scenarios, rollback)
- **Unit tests**: Certificate expiration calculation, retry backoff logic, validation functions

**Compliance**: PASS - Tests will be written first per TDD workflow before implementation

### ✅ IV. Security by Design

**Security considerations**:
- **Input validation**: All renewal requests validated on service side (device identity, security proof)
- **Authentication**: TPM attestation and device fingerprint verification for expired certificate recovery
- **Cryptographic operations**: Reusing existing `internal/crypto/` and `internal/tpm/` libraries (no custom crypto)
- **Security assumptions documented**: See spec Assumptions section (TPM hardware, bootstrap cert availability, CA stability)
- **Threat model**: Mitigates certificate expiration attacks; prevents unauthorized certificate issuance

**Compliance**: PASS - Leverages established crypto libraries, validates all inputs, documents security assumptions

### ✅ V. API Stability and Versioning

**API impact assessment**:
- **New endpoint**: POST `/api/v1beta1/devices/{name}/certificaterenewal` (backward-compatible addition)
- **Existing endpoints**: No breaking changes to current enrollment or device APIs
- **Certificate format**: Maintains existing X.509 certificate format (365-day validity)
- **Versioning**: MINOR version bump (new feature, backward-compatible)

**Compliance**: PASS - No breaking changes; follows semantic versioning for new endpoint

### ✅ VI. Database Migration Discipline

**Database changes needed**:
- New table: `certificate_renewal_requests` (tracking renewal state, timestamps, device references)
- Columns: `device_id`, `request_time`, `completion_time`, `status`, `security_proof_type`, `certificate_serial`
- **Migration testing**: Will test with 1k, 10k, 100k device records
- **Rollback procedure**: Migration includes DOWN migration to drop table
- **Non-blocking**: Table creation and index creation will be non-blocking operations

**Compliance**: PASS - Versioned migrations planned with rollback support and performance testing

### ✅ VII. Code Quality and Linting

**Linting strategy**:
- All code will pass `make lint` before commit
- Error handling: All CSR operations, file I/O, and HTTP requests will have explicit error handling
- Import grouping: Standard library → third-party → local (per Go conventions)
- Exported functions: Certificate monitoring, renewal, and atomic swap operations will be documented
- Complexity management: Will extract helper functions if cyclomatic complexity exceeds reasonable limits

**Compliance**: PASS - Will follow established Go best practices and run `make lint` continuously

### Summary

**All Constitution gates: PASSED** ✅

No violations detected. Feature aligns with all seven core principles. No complexity justifications needed.

## Project Structure

### Documentation (this feature)

```text
specs/[###-feature]/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

This feature follows Flight Control's established distributed system architecture:

```text
internal/agent/                           # Agent-side implementation
├── device/
│   └── certrotation/                     # NEW: Certificate rotation manager
│       ├── monitor.go                    # Certificate expiration monitoring
│       ├── renewer.go                    # Renewal request generation
│       ├── atomic.go                     # Atomic certificate swap operations
│       └── recovery.go                   # Expired certificate recovery logic
└── config/
    └── config.go                         # MODIFY: Add renewal threshold configuration

internal/service/                          # Service-side implementation
├── certrotation/                         # NEW: Renewal request processing
│   ├── handler.go                        # HTTP handler for renewal endpoints
│   ├── validator.go                      # Security proof validation
│   └── issuer.go                         # Certificate issuance logic
└── store/
    └── certrotation.go                   # NEW: Database operations for renewal tracking

internal/api/                             # API definitions
└── v1beta1/
    └── openapi.yaml                      # MODIFY: Add certificate renewal endpoint

internal/crypto/                          # Existing crypto library (reuse)
└── cert.go                               # Certificate utilities (may need extensions)

internal/tpm/                             # Existing TPM library (reuse)
├── tpm.go                                # TPM client operations
└── csr.go                                # CSR generation (reuse for renewal)

test/                                     # Integration tests
├── integration/
│   └── certrotation/                     # NEW: Certificate rotation integration tests
│       ├── renewal_test.go               # Proactive renewal flow tests
│       ├── recovery_test.go              # Expired certificate recovery tests
│       └── atomic_test.go                # Atomic swap failure scenarios
└── contract/
    └── certrotation_contract_test.go     # NEW: API contract tests

db/migrations/                            # Database migrations
└── YYYYMMDDHHMMSS_add_cert_renewal.sql   # NEW: Certificate renewal tracking table
```

**Structure Decision**: Distributed system structure with agent and service components. The agent-side code manages certificate lifecycle (monitoring, renewal, atomic swap), while the service-side code processes renewal requests and issues certificates. Follows existing Flight Control patterns: `internal/agent/device/*` for agent logic, `internal/service/*` for service logic, `internal/store/*` for database operations.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| [e.g., 4th project] | [current need] | [why 3 projects insufficient] |
| [e.g., Repository pattern] | [specific problem] | [why direct DB access insufficient] |
