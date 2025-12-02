# Quickstart: Testing Agent Certificate Rotation Locally

**Date**: 2025-12-02
**Feature**: Agent Certificate Rotation
**Purpose**: Guide for developers to test certificate rotation locally

## Overview

This guide walks through setting up a local development environment to test automatic certificate rotation, including proactive renewal, expired certificate recovery, and atomic swap operations.

## Prerequisites

- Go 1.24+ installed
- PostgreSQL 16+ running locally
- TPM simulator (optional, for TPM attestation testing)
- Flight Control repository cloned
- Make and development tools installed

## Setup Steps

### 1. Start Local Development Environment

```bash
# From repository root
make deploy

# This starts:
# - PostgreSQL database
# - Flight Control API service
# - Flight Control worker
# - Redis (for rate limiting)
```

### 2. Apply Database Migration

```bash
# Apply the certificate renewal tracking migration
psql -h localhost -U flightctl -d flightctl < db/migrations/YYYYMMDDHHMMSS_add_cert_renewal.sql

# Verify table created
psql -h localhost -U flightctl -d flightctl -c "\d certificate_renewal_requests"
```

Expected output:
```
                                  Table "public.certificate_renewal_requests"
        Column         |            Type             | Collation | Nullable |                       Default
-----------------------+-----------------------------+-----------+----------+------------------------------------------------------
 id                    | integer                     |           | not null | nextval('certificate_renewal_requests_id_seq'::regclass)
 device_id             | text                        |           | not null |
 request_id            | uuid                        |           | not null |
 request_time          | timestamp without time zone |           | not null | now()
 ...
```

### 3. Configure Agent with Certificate Rotation

Create test agent configuration:

```bash
# Create test agent config directory
mkdir -p /tmp/flightctl-test/etc/flightctl
mkdir -p /tmp/flightctl-test/var/lib/flightctl/certs

# Create agent config with rotation enabled
cat > /tmp/flightctl-test/etc/flightctl/config.yaml <<EOF
management:
  server: https://localhost:7443
  certificateAuthority: /tmp/flightctl-test/etc/flightctl/certs/ca.crt

certRotation:
  enabled: true
  renewalThresholdDays: 30
  retryInitialInterval: 1m
  retryMaxInterval: 24h
  retryBackoffMultiplier: 2.0
  monitorIntervalSeconds: 60

logLevel: debug
EOF
```

### 4. Create Test Certificates

Generate test certificates with short validity periods for testing:

```bash
# Generate CA certificate (if not already present)
openssl req -x509 -newkey rsa:2048 -keyout /tmp/flightctl-test/etc/flightctl/certs/ca.key \
  -out /tmp/flightctl-test/etc/flightctl/certs/ca.crt -days 365 -nodes \
  -subj "/CN=Flight Control Test CA"

# Generate device certificate expiring in 35 days (within renewal window)
openssl req -newkey rsa:2048 -keyout /tmp/flightctl-test/var/lib/flightctl/certs/agent.key \
  -out /tmp/flightctl-test/var/lib/flightctl/certs/agent.csr -nodes \
  -subj "/CN=test-device-001"

openssl x509 -req -in /tmp/flightctl-test/var/lib/flightctl/certs/agent.csr \
  -CA /tmp/flightctl-test/etc/flightctl/certs/ca.crt \
  -CAkey /tmp/flightctl-test/etc/flightctl/certs/ca.key \
  -CAcreateserial -out /tmp/flightctl-test/var/lib/flightctl/certs/agent.crt \
  -days 35 -sha256

# Generate bootstrap certificate with longer validity
openssl req -newkey rsa:2048 -keyout /tmp/flightctl-test/etc/flightctl/certs/client-enrollment.key \
  -out /tmp/flightctl-test/etc/flightctl/certs/client-enrollment.csr -nodes \
  -subj "/CN=test-device-001-bootstrap"

openssl x509 -req -in /tmp/flightctl-test/etc/flightctl/certs/client-enrollment.csr \
  -CA /tmp/flightctl-test/etc/flightctl/certs/ca.crt \
  -CAkey /tmp/flightctl-test/etc/flightctl/certs/ca.key \
  -CAcreateserial -out /tmp/flightctl-test/etc/flightctl/certs/client-enrollment.crt \
  -days 730 -sha256

# Clean up CSR files
rm /tmp/flightctl-test/var/lib/flightctl/certs/agent.csr
rm /tmp/flightctl-test/etc/flightctl/certs/client-enrollment.csr
```

### 5. Start Test Agent

```bash
# Build agent with rotation feature
make build

# Run agent with test configuration
export FLIGHTCTL_TEST_ROOT_DIR=/tmp/flightctl-test
bin/flightctl-agent --config /tmp/flightctl-test/etc/flightctl/config.yaml
```

## Test Scenarios

### Scenario 1: Proactive Certificate Renewal (Happy Path)

**Goal**: Verify device automatically renews certificate when it approaches expiration

**Steps**:

1. Start agent with certificate expiring in 35 days (created in setup)
2. Wait for first sync cycle (~60 seconds)
3. Check logs for renewal initiation:
   ```
   level=info msg="Certificate expiring soon detected" daysRemaining=35 renewalThreshold=30
   level=info msg="Initiating certificate renewal" requestId=<uuid>
   ```
4. Check logs for successful renewal:
   ```
   level=info msg="Renewal request submitted" requestId=<uuid>
   level=info msg="New certificate received" serialNumber=<new-serial>
   level=info msg="Atomic certificate swap initiated"
   level=info msg="Certificate renewed successfully" oldSerial=<old-serial> newSerial=<new-serial>
   ```
5. Verify new certificate installed:
   ```bash
   openssl x509 -in /tmp/flightctl-test/var/lib/flightctl/certs/agent.crt -noout -dates
   ```
   Should show new `notAfter` date (~365 days from now)

6. Query database for renewal record:
   ```sql
   SELECT * FROM certificate_renewal_requests
   WHERE device_id = 'test-device-001' AND status = 'completed'
   ORDER BY request_time DESC LIMIT 1;
   ```

**Expected Outcome**:
- Agent detects expiring certificate
- Generates CSR with `valid_cert` security proof type
- Service issues new certificate
- Agent performs atomic swap
- Old certificate backed up with `.old` suffix (temporarily)
- New certificate operational

---

### Scenario 2: Expired Certificate Recovery with Bootstrap Certificate

**Goal**: Verify device can recover from expired management certificate using bootstrap cert

**Steps**:

1. Create device certificate that is already expired:
   ```bash
   # Generate expired certificate (valid from 400 days ago to 35 days ago)
   faketime '400 days ago' openssl req -newkey rsa:2048 \
     -keyout /tmp/flightctl-test/var/lib/flightctl/certs/agent.key \
     -out /tmp/flightctl-test/var/lib/flightctl/certs/agent.csr -nodes \
     -subj "/CN=test-device-002"

   faketime '400 days ago' openssl x509 -req \
     -in /tmp/flightctl-test/var/lib/flightctl/certs/agent.csr \
     -CA /tmp/flightctl-test/etc/flightctl/certs/ca.crt \
     -CAkey /tmp/flightctl-test/etc/flightctl/certs/ca.key \
     -CAcreateserial -out /tmp/flightctl-test/var/lib/flightctl/certs/agent.crt \
     -days 365 -sha256
   ```

2. Start agent
3. Check logs for recovery initiation:
   ```
   level=warn msg="Expired management certificate detected" expiredSince=<timestamp>
   level=info msg="Initiating certificate recovery" securityProofType=bootstrap_cert
   ```

4. Check logs for successful recovery:
   ```
   level=info msg="Bootstrap certificate authentication successful"
   level=info msg="New certificate received" serialNumber=<new-serial>
   level=info msg="Certificate recovery completed"
   level=info msg="Device status: online"
   ```

5. Verify device can now communicate with service normally

**Expected Outcome**:
- Agent detects expired management certificate on startup
- Falls back to bootstrap certificate authentication
- Service validates bootstrap certificate and issues new management certificate
- Device resumes normal operations

---

### Scenario 3: Atomic Swap Rollback on Validation Failure

**Goal**: Verify atomic swap rollback preserves old certificate if new certificate is invalid

**Setup**: This requires code instrumentation to simulate validation failure. For manual testing:

1. Modify `internal/agent/device/certrotation/atomic.go` to inject validation failure:
   ```go
   func (a *AtomicSwap) validateNewCertificate(certPath string) error {
       // TESTING: Force validation failure
       return errors.New("simulated validation failure for testing")
   }
   ```

2. Rebuild agent: `make build`
3. Start agent with expiring certificate

**Steps**:

1. Wait for renewal to trigger
2. Check logs for rollback:
   ```
   level=error msg="New certificate validation failed" error="simulated validation failure for testing"
   level=warn msg="Rolling back to old certificate"
   level=info msg="Rollback completed successfully" certificatePath=<path>
   ```

3. Verify old certificate still in place:
   ```bash
   openssl x509 -in /tmp/flightctl-test/var/lib/flightctl/certs/agent.crt -noout -serial
   # Should show old serial number
   ```

4. Verify agent continues operating with old certificate

**Expected Outcome**:
- New certificate written to temp location
- Validation fails
- Rollback mechanism restores old certificate
- Agent continues normal operation with old certificate
- Renewal will be retried on next sync cycle

---

### Scenario 4: Retry with Exponential Backoff on Service Unavailable

**Goal**: Verify agent retries renewal with exponential backoff when service is unavailable

**Steps**:

1. Start agent with expiring certificate
2. Stop the API service:
   ```bash
   docker stop flightctl-api
   ```

3. Check logs for retry behavior:
   ```
   level=error msg="Renewal request failed" error="connection refused"
   level=info msg="Scheduling retry" attemptNumber=1 nextRetryAfter=1m
   ```

4. Wait 1 minute, check for second retry:
   ```
   level=error msg="Renewal request failed" error="connection refused"
   level=info msg="Scheduling retry" attemptNumber=2 nextRetryAfter=2m
   ```

5. Wait 2 minutes, check for third retry:
   ```
   level=error msg="Renewal request failed" error="connection refused"
   level=info msg="Scheduling retry" attemptNumber=3 nextRetryAfter=4m
   ```

6. Restart API service:
   ```bash
   docker start flightctl-api
   ```

7. Wait for next retry, verify successful renewal:
   ```
   level=info msg="Renewal request submitted" requestId=<uuid> attemptNumber=4
   level=info msg="Certificate renewed successfully"
   ```

**Expected Outcome**:
- Agent retries with exponential backoff: 1m, 2m, 4m, 8m, ...
- Device maintains old certificate during retries
- Renewal succeeds when service becomes available

---

### Scenario 5: Database Query Performance at Scale

**Goal**: Verify database queries perform well with large number of renewal records

**Steps**:

1. Generate test data (simulate 100,000 devices with renewal history):
   ```sql
   -- Insert 100,000 test renewal records
   INSERT INTO certificate_renewal_requests
     (device_id, request_id, request_time, completion_time, status, security_proof_type)
   SELECT
     'device-' || generate_series,
     gen_random_uuid(),
     NOW() - (random() * interval '30 days'),
     NOW() - (random() * interval '30 days'),
     CASE WHEN random() > 0.01 THEN 'completed' ELSE 'failed' END,
     (ARRAY['valid_cert', 'bootstrap_cert', 'tpm_attestation'])[floor(random() * 3 + 1)]
   FROM generate_series(1, 100000);
   ```

2. Test query performance for common operations:

   ```sql
   -- Query 1: Get renewal history for specific device
   EXPLAIN ANALYZE
   SELECT * FROM certificate_renewal_requests
   WHERE device_id = 'device-50000'
   ORDER BY request_time DESC LIMIT 10;

   -- Query 2: Count failed renewals in last 24 hours
   EXPLAIN ANALYZE
   SELECT COUNT(*) FROM certificate_renewal_requests
   WHERE status = 'failed'
     AND request_time > NOW() - interval '24 hours';

   -- Query 3: Renewal success rate by security proof type
   EXPLAIN ANALYZE
   SELECT
     security_proof_type,
     COUNT(*) as total,
     SUM(CASE WHEN status = 'completed' THEN 1 ELSE 0 END) as successful,
     ROUND(100.0 * SUM(CASE WHEN status = 'completed' THEN 1 ELSE 0 END) / COUNT(*), 2) as success_rate
   FROM certificate_renewal_requests
   WHERE request_time > NOW() - interval '30 days'
   GROUP BY security_proof_type;
   ```

3. Verify all queries use indexes (check EXPLAIN output for "Index Scan")

**Expected Outcome**:
- All queries complete in < 100ms
- Indexes are used effectively
- No table scans on large dataset

---

## Observability and Metrics

### View Prometheus Metrics

```bash
# Start observability stack
make deploy-e2e-extras

# Access Prometheus
open http://localhost:9090

# Query certificate expiration time
flightctl_agent_cert_expiration_time_seconds

# Query renewal success rate
rate(flightctl_agent_cert_renewal_successes_total[5m])
  /
rate(flightctl_agent_cert_renewal_attempts_total[5m])
```

### View OpenTelemetry Traces

If tracing is configured, view traces in Jaeger or other OTLP-compatible backend:

1. Search for spans: `flightctl/agent/certrotation`
2. Filter by operation: `RenewCertificate`, `RecoverFromExpiredCert`, `SwapCertificate`
3. Examine span attributes: `device.id`, `cert.daysUntilExpiry`, `renewal.attemptNumber`

### Check Database Renewal Records

```sql
-- Recent renewals
SELECT
  device_id,
  request_time,
  status,
  security_proof_type,
  processing_duration_ms
FROM certificate_renewal_requests
ORDER BY request_time DESC
LIMIT 20;

-- Success rate by device
SELECT
  device_id,
  COUNT(*) as total_attempts,
  SUM(CASE WHEN status = 'completed' THEN 1 ELSE 0 END) as successful,
  ROUND(100.0 * SUM(CASE WHEN status = 'completed' THEN 1 ELSE 0 END) / COUNT(*), 2) as success_rate
FROM certificate_renewal_requests
GROUP BY device_id
HAVING COUNT(*) > 1
ORDER BY success_rate ASC;
```

## Troubleshooting

### Issue: Agent doesn't detect expiring certificate

**Check**:
- Certificate is actually within renewal threshold: `openssl x509 -in agent.crt -noout -dates`
- Agent config has `certRotation.enabled: true`
- Agent is syncing with service (check status update logs)

### Issue: Renewal request fails with 401 Unauthorized

**Check**:
- Device exists in database: `SELECT * FROM devices WHERE name = 'test-device-001';`
- Certificate subject matches device name
- Bootstrap certificate is valid (not expired)

### Issue: Atomic swap fails

**Check**:
- File permissions: agent must have write access to `/var/lib/flightctl/certs/`
- Disk space available: `df -h /var/lib/flightctl`
- New certificate is valid: `openssl x509 -in agent.crt.tmp -noout -text`

### Issue: Service returns 500 Internal Server Error

**Check**:
- Service logs for CA errors
- Database connection healthy
- Certificate authority keys accessible

## Cleanup

```bash
# Stop agent
pkill -f flightctl-agent

# Stop services
make clean

# Remove test data
rm -rf /tmp/flightctl-test

# Clean database
psql -h localhost -U flightctl -d flightctl -c "TRUNCATE TABLE certificate_renewal_requests;"
```

## Next Steps

After local testing:
1. Review implementation in `internal/agent/device/certrotation/`
2. Review service implementation in `internal/service/certrotation/`
3. Run integration tests: `go test ./test/integration/certrotation/...`
4. Run contract tests: `go test ./test/contract/certrotation_contract_test.go`
5. Test in staging environment with real devices
6. Monitor metrics in production rollout

## References

- [Specification](spec.md)
- [Data Model](data-model.md)
- [API Contract](contracts/renewal-api.yaml)
- [Research](research.md)
