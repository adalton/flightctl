package certrotation

// Tracing constants for certificate rotation operations
// Use with existing tracing.StartSpan(ctx, TracingComponent, operation)
const (
	// TracingComponent is the component name for all cert rotation tracing spans
	TracingComponent = "flightctl/agent/certrotation"
)

// Span operation names (will be converted to kebab-case by StartSpan)
const (
	OpMonitorExpiration   = "MonitorCertificateExpiration"
	OpGenerateRenewalCSR  = "GenerateRenewalCSR"
	OpSubmitRenewal       = "SubmitRenewalRequest"
	OpValidateCertificate = "ValidateCertificate"
	OpSwapCertificate     = "SwapCertificate"
	OpRecoverExpired      = "RecoverFromExpiredCert"
	OpRetryRenewal        = "RetryRenewalRequest"
	OpAtomicWrite         = "AtomicFileWrite"
	OpRollbackCert        = "RollbackCertificate"
)
