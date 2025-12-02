<!--
SYNC IMPACT REPORT
==================
Version Change: (new) → 1.0.0
Modified Principles: N/A (initial version)
Added Sections:
  - Core Principles (I-VII)
  - Development Standards
  - Quality Gates
  - Governance
Removed Sections: N/A
Templates Requiring Updates:
  ✅ .specify/templates/plan-template.md - Constitution Check section aligns with principles
  ✅ .specify/templates/spec-template.md - User scenarios and requirements structure compatible
  ✅ .specify/templates/tasks-template.md - Task organization supports testing and observability principles
Follow-up TODOs: None - all placeholders filled
-->

# Flight Control Constitution

## Core Principles

### I. Distributed Systems First

Flight Control is a distributed edge device management system. All features MUST:

- Support distributed deployment (API server, worker services, edge agents)
- Handle network partitioning and intermittent connectivity gracefully
- Maintain eventual consistency across system boundaries
- Design for horizontal scalability of service components

**Rationale**: Edge devices operate in unreliable network conditions. The system architecture must reflect this reality from day one, not as an afterthought.

### II. Observability is Non-Negotiable

Every service component MUST be instrumented with OpenTelemetry tracing. All features MUST:

- Create spans for significant operations using `tracing.StartSpan(ctx, component, operation)`
- Propagate context through all service boundaries (HTTP, gRPC, message queues)
- Record meaningful attributes (entity IDs, operation parameters, error details)
- Set appropriate span status (OK/Error) based on operation outcome
- Emit structured logs using the configured logger (logrus)

**Rationale**: Distributed systems are impossible to debug without comprehensive tracing. Observability is a first-class requirement, not a feature to add later.

### III. Test-Driven Development (TDD)

Testing discipline is MANDATORY. The required workflow is:

1. Write tests FIRST based on acceptance criteria
2. Ensure tests FAIL (red)
3. Implement minimum code to pass (green)
4. Refactor while keeping tests green

Test coverage requirements:

- **Contract tests**: REQUIRED for all API endpoints and service interfaces
- **Integration tests**: REQUIRED for cross-service interactions, database operations, message queues
- **Unit tests**: REQUIRED for complex business logic, data transformations, validation rules

**Rationale**: Edge device management is mission-critical infrastructure. Untested code is unshippable code. The cost of bugs in production far exceeds the cost of writing tests first.

### IV. Security by Design

All features touching authentication, authorization, device identity, or sensitive data MUST:

- Follow the principle of least privilege
- Validate all inputs at system boundaries (API, agent communication, database)
- Use established cryptographic libraries (no custom crypto)
- Document security assumptions and threat model changes
- Pass security review before merging

Security-sensitive areas include:

- Certificate signing requests (CSR) and device enrollment
- Authentication provider integration (OAuth2, OIDC, PAM)
- Organization and multi-tenancy isolation
- Device attestation and identity verification

**Rationale**: Flight Control manages fleet infrastructure. Security compromises can affect thousands of edge devices. Security must be designed in, not bolted on.

### V. API Stability and Versioning

API changes MUST follow semantic versioning (MAJOR.MINOR.PATCH):

- **MAJOR**: Breaking changes to public APIs (endpoint removal, required field changes, behavior changes)
- **MINOR**: Backward-compatible additions (new endpoints, optional fields, new features)
- **PATCH**: Bug fixes, documentation, non-behavioral changes

Breaking changes require:

- Deprecation notice in prior MINOR release
- Migration guide for API consumers
- Upgrade compatibility documentation update

**Rationale**: Edge devices update on their own schedules. API stability ensures smooth fleet operations during rolling upgrades.

### VI. Database Migration Discipline

All database schema changes MUST:

- Be expressed as versioned migrations (up and down)
- Be tested with realistic data volumes (1k, 10k, 100k devices)
- Include rollback procedure documentation
- Be validated against supported PostgreSQL versions (16+)
- Minimize blocking operations (prefer non-blocking index creation, column additions)

**Rationale**: Flight Control manages persistent state for thousands of devices. Schema changes that lock tables or lose data are unacceptable.

### VII. Code Quality and Linting

All code MUST pass linting before merge. The workflow is:

1. Write lint-compliant code from the start
2. Run `make lint` after any code changes
3. Fix ALL lint errors before marking work complete
4. NEVER leave lint errors for others to fix

Common Go linting requirements:

- Handle all error returns (or explicitly ignore with `_` and comment why)
- Group imports: standard library, third-party, local
- Keep cyclomatic complexity reasonable (extract helper functions when needed)
- Set `ReadHeaderTimeout` on HTTP servers
- Document exported types and functions

**Rationale**: Lint errors indicate code quality issues. The project uses `golangci-lint` to enforce consistent standards across the codebase.

## Development Standards

### Technology Stack

- **Language**: Go 1.24+ (see go.mod for exact version)
- **Database**: PostgreSQL 16+
- **Caching**: Redis (for rate limiting, TTL caches)
- **Messaging**: Work queues for async operations
- **API Framework**: go-chi for HTTP routing, oapi-codegen for OpenAPI
- **Tracing**: OpenTelemetry with OTLP HTTP exporter
- **Testing**: Go standard testing, Ginkgo/Gomega for integration tests

### Project Structure

```
cmd/           - Service entry points (api, worker, agent, CLI)
internal/      - Internal packages (not importable by other projects)
  api/         - API handlers and OpenAPI definitions
  service/     - Business logic layer
  store/       - Database repository layer
  agent/       - Agent-specific logic
  crypto/      - Cryptographic operations
  tracing/     - OpenTelemetry instrumentation
pkg/           - Public packages (importable)
test/          - Integration and E2E tests
docs/          - User and developer documentation
```

### Simplicity and YAGNI

- Start with the simplest solution that could work
- Avoid premature abstraction (three uses minimum before extracting helpers)
- Don't add configurability or extensibility until needed
- Only handle errors that can actually occur (trust internal contracts)
- Delete unused code completely (no commented-out code, no backwards-compatibility shims for removed features)

## Quality Gates

### Pre-Merge Requirements

Every pull request MUST:

1. ✅ Pass `make lint` with zero errors
2. ✅ Pass all tests (`make unit-test`, integration tests)
3. ✅ Include tests for new functionality (contract/integration tests per TDD principle)
4. ✅ Update relevant documentation (API docs, user guides, architecture docs)
5. ✅ Pass code review from at least one maintainer
6. ✅ Have meaningful commit messages following project conventions

### Constitution Compliance Checklist

During planning and design review, verify:

- [ ] Distributed systems first: How does this handle network partitioning?
- [ ] Observability: Are all service operations traced? Are errors recorded?
- [ ] Testing: Are tests written first? Do they cover contract and integration scenarios?
- [ ] Security: Are inputs validated? Are security assumptions documented?
- [ ] API stability: Does this change affect public APIs? Is versioning correct?
- [ ] Database discipline: Are migrations tested? Is rollback documented?
- [ ] Code quality: Does code follow Go best practices? Will it pass linting?

## Governance

### Amendment Process

1. Propose change via pull request to this file
2. Include rationale for change and impact analysis
3. Update version number according to semantic versioning rules:
   - MAJOR: Removed or redefined core principle
   - MINOR: New principle or section added
   - PATCH: Clarification, wording improvement
4. Update dependent templates (plan, spec, tasks) for consistency
5. Require approval from project maintainers
6. Update LAST_AMENDED_DATE upon merge

### Constitution Authority

This constitution supersedes all other development practices. When in conflict:

1. Constitution principles take precedence
2. Documented architecture patterns second
3. Code review feedback third
4. Individual preferences last

### Complexity Justification

Any violation of constitution principles MUST be documented in the implementation plan's "Complexity Tracking" section with:

- Which principle is violated
- Why the violation is necessary
- What simpler alternative was rejected and why

### Compliance Review

- All code reviews MUST verify constitution compliance
- Quarterly review of constitution adherence across merged PRs
- Annual review of constitution itself for relevance and completeness

**Version**: 1.0.0 | **Ratified**: 2025-12-02 | **Last Amended**: 2025-12-02
