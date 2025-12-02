-- +goose Up
-- +goose StatementBegin
-- Create certificate_renewal_requests table for tracking device certificate renewal requests
CREATE TABLE IF NOT EXISTS certificate_renewal_requests (
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
    
    -- Foreign key constraint
    CONSTRAINT fk_certificate_renewal_device FOREIGN KEY (device_id) REFERENCES devices(name) ON DELETE CASCADE
);

-- Create indexes for performance
CREATE INDEX idx_cert_renewal_device_status ON certificate_renewal_requests(device_id, status);
CREATE INDEX idx_cert_renewal_request_time ON certificate_renewal_requests(request_time DESC);
CREATE INDEX idx_cert_renewal_request_id ON certificate_renewal_requests(request_id);

-- Add comment for documentation
COMMENT ON TABLE certificate_renewal_requests IS 'Tracks certificate renewal requests from devices for audit and observability';
COMMENT ON COLUMN certificate_renewal_requests.security_proof_type IS 'Authentication method: valid_cert (proactive renewal), bootstrap_cert (expired recovery), tpm_attestation (TPM-based recovery)';
COMMENT ON COLUMN certificate_renewal_requests.status IS 'Renewal status: pending (queued), processing (in progress), completed (success), failed (terminal failure)';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Drop indexes first
DROP INDEX IF EXISTS idx_cert_renewal_request_id;
DROP INDEX IF EXISTS idx_cert_renewal_request_time;
DROP INDEX IF EXISTS idx_cert_renewal_device_status;

-- Drop table
DROP TABLE IF EXISTS certificate_renewal_requests;
-- +goose StatementEnd
