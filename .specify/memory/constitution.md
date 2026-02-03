<!--
Sync Impact Report - Constitution Update
========================================
Version Change: Initial → 1.0.0
Rationale: Initial ratification of Flight Control project constitution

Modified Principles: N/A (initial version)
Added Sections:
  - Core Principles (3): Test-First Development, Code Quality & Standards, Observability
  - Development Standards section
  - Governance section

Removed Sections: N/A (initial version)

Templates Status:
  ✅ plan-template.md - Constitution Check section verified and aligned
  ✅ spec-template.md - Requirements alignment verified
  ✅ tasks-template.md - Task categorization aligned with principles
  ✅ agent-file-template.md - Verified compatibility
  ✅ checklist-template.md - Verified compatibility

Follow-up TODOs: None

Notes:
  - CLAUDE.md referenced as runtime agent guidance (separation of concerns)
  - Go 1.24.0 with golangci-lint enforcement
  - Semantic versioning policy established
-->

# Flight Control Constitution

## Core Principles

### I. Test-First Development (NON-NEGOTIABLE)

**TDD is mandatory for all feature development.**

- Tests MUST be written before implementation code
- Test files MUST be reviewed and approved by stakeholders before implementation begins
- Red-Green-Refactor cycle MUST be strictly followed:
  1. Write tests that fail (RED)
  2. Get user/reviewer approval of test coverage
  3. Implement minimum code to pass tests (GREEN)
  4. Refactor while keeping tests green
- Contract tests MUST be written for:
  - New service interfaces
  - API endpoint changes
  - Shared data structures between components
  - Inter-service communication patterns
- Integration tests MUST cover:
  - Critical user workflows (enrollment, fleet management, device updates)
  - Service boundary interactions
  - Database schema changes

**Rationale**: Flight Control manages critical edge device infrastructure. Defects in device enrollment, configuration management, or fleet orchestration can affect thousands of devices. TDD ensures correctness before deployment and serves as executable documentation of system behavior.

### II. Code Quality & Standards

**All code MUST pass project linting and adhere to Go best practices.**

- `make lint` MUST pass before commits
- golangci-lint violations MUST be fixed, not suppressed (exceptions require documented justification)
- Go best practices MUST be followed:
  - Error handling: Always check and handle error returns
  - Import ordering: standard library → third-party → local packages
  - Complexity limits: Extract functions when cyclomatic complexity exceeds project threshold
  - Security: No command injection, XSS, SQL injection, or OWASP Top 10 vulnerabilities
- Code changes MUST preserve existing behavior unless explicitly changing it
- Interfaces changes MUST:
  - Update all implementations immediately
  - Update all mocks to match new signatures
  - Update all test code
  - Pass full test suite (`go test -v -race ./...`) before proceeding
- Documentation:
  - Function comments MUST explain "why" (rationale, design decisions)
  - "What" should be clear from code itself
  - Complex algorithms MUST include explanatory comments

**Rationale**: Flight Control is a production system managing device fleets at scale. Code quality directly impacts system reliability, maintainability, and security. Consistent standards enable team collaboration and reduce technical debt.

### III. Observability

**All services MUST be observable through structured logging and metrics.**

- Structured logging REQUIRED:
  - Use consistent log levels (DEBUG, INFO, WARN, ERROR)
  - Include request IDs for tracing
  - Log key state transitions (enrollment approved, fleet updated, device reconciled)
  - Do NOT log sensitive data (certificates, credentials, PII)
- Metrics collection REQUIRED for:
  - Request rates and latencies (p50, p95, p99)
  - Error rates by type
  - Resource utilization (DB connections, memory, goroutines)
  - Business metrics (devices enrolled, fleets created, deployments in progress)
- Distributed tracing:
  - Propagate trace context across service boundaries
  - Instrument critical paths (enrollment flow, device sync, fleet rollout)

**Rationale**: Flight Control operates in distributed environments managing thousands of devices. Observability is essential for diagnosing issues in production, understanding system behavior under load, and maintaining SLAs. Structured logs and metrics enable proactive monitoring and rapid incident response.

## Development Standards

### Technology Stack

- **Language**: Go 1.24.0 (with FIPS-validated toolset when available)
- **Testing**: Go standard testing, gotestsum for output formatting
- **Mocking**: go.uber.org/mock/mockgen v0.4.0
- **Database**: PostgreSQL (versions per deployment: e2e, small-1k, medium-10k)
- **Linting**: golangci-lint (configured in project)
- **Container Runtime**: Podman with systemd Quadlets support

### Coding Standards

- Follow patterns established in existing codebase
- Use dependency injection for testability
- Prefer composition over inheritance
- Keep functions focused and single-purpose
- Avoid premature abstraction (three instances before extracting pattern)
- Security: Validate inputs at system boundaries, trust internal code and framework guarantees
- Error messages MUST be actionable (include context, not just error text)

### Testing Standards

- Unit tests: Test public interfaces, not internal implementation
- Integration tests: Test service interactions and critical user workflows
- Contract tests: Validate API schemas and service boundaries
- Mock external dependencies (databases, APIs, filesystems), not internal logic
- Test names MUST describe scenario and expected outcome
- Tests MUST be deterministic (no flaky tests tolerated)

## Governance

### Amendment Process

- Constitution changes REQUIRE:
  - Documentation of proposed change with rationale
  - Review and approval from project maintainers
  - Migration plan if changes affect existing code or processes
  - Version increment per semantic versioning rules (see below)
- All PRs and code reviews MUST verify compliance with constitution
- Complexity beyond constitution principles MUST be justified in implementation plan
- Violations require explicit acknowledgment and remediation plan

### Versioning Policy

Constitution follows semantic versioning (MAJOR.MINOR.PATCH):

- **MAJOR**: Backward incompatible changes (principle removal, governance redefinition, non-negotiable additions)
- **MINOR**: Additive changes (new principles, expanded guidance, new sections)
- **PATCH**: Clarifications, wording improvements, typo fixes, non-semantic refinements

### Compliance Review

- Feature plans MUST include "Constitution Check" section (per plan-template.md)
- Plan approval gates on constitution compliance
- Implementation reviews verify adherence to principles
- Retrospectives assess constitution effectiveness and identify needed amendments

### Runtime Guidance

For agent-specific development guidance (communication style, tool usage patterns, file operation preferences), refer to:
- `/home/andalton/.claude/CLAUDE.md` - Agent runtime instructions
- `.specify/templates/` - Workflow-specific templates and patterns

Constitution establishes **project-level governance**; agent guidance provides **implementation-level details**.

**Version**: 1.0.0 | **Ratified**: 2026-02-03 | **Last Amended**: 2026-02-03
