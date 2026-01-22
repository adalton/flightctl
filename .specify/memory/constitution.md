<!--
  SYNC IMPACT REPORT
  ==================
  Version Change: 1.0.0 → 1.1.0 (MINOR bump)
  Rationale: Added new principle (Documentation Conciseness) and materially expanded governance guidance

  Modified Principles:
  - NEW: I. Code Quality & Simplicity
  - NEW: II. Edge Device Management
  - NEW: III. API Stability & Versioning
  - NEW: IV. Test Coverage & Quality
  - NEW: V. Security
  - NEW: VI. Documentation Conciseness
  - NEW: VII. Observability
  - NEW: VIII. Performance & Scale

  Templates Requiring Updates:
  ✅ .specify/templates/plan-template.md - Updated Constitution Check with v1.1.0 principles + conciseness reminder
  ✅ .specify/templates/spec-template.md - Added Constitution Alignment section + conciseness reminder at top
  ✅ .specify/templates/tasks-template.md - Added Constitution Reminders for quality gates + conciseness
  ✅ .specify/templates/agent-file-template.md - No changes needed (generic template)

  Follow-up TODOs: None
-->

# Flight Control Constitution

## Core Principles

### I. Code Quality & Simplicity

Code MUST be production-ready, idiomatic, and pass all linting checks before completion. YAGNI principle applies—avoid over-engineering, premature abstractions, and hypothetical future requirements. Prefer three similar lines over a premature helper. No backwards-compatibility hacks for unused code; delete completely. All code must pass `make lint` before marking work complete.

**Rationale**: Simple, clean code reduces bugs, improves maintainability, and accelerates development. Linting catches quality issues early.

### II. Edge Device Management

Features MUST support declarative fleet management at scale. Device lifecycle (provisioning, enrollment, updates, decommissioning) must be observable and auditable. Workload orchestration must handle partial fleet updates, rollbacks, and failure scenarios gracefully.

**Rationale**: Flight Control manages potentially thousands of edge devices; robustness and transparency are non-negotiable.

### III. API Stability & Versioning

Public APIs MUST follow semantic versioning (MAJOR.MINOR.PATCH). Breaking changes require MAJOR version bump, new features MINOR, fixes PATCH. Deprecation warnings required one minor version before removal. API contracts (documented in `specs/*/contracts/`) are binding.

**Rationale**: Fleet operators depend on stable APIs; unannounced breaking changes disrupt production deployments.

### IV. Test Coverage & Quality

Unit tests with good coverage are REQUIRED. Tests MUST verify observable behavior, not implementation details. Mock external dependencies (databases, APIs), not internal logic. Interface changes MUST be followed immediately by running `go test -v -race ./...` to catch broken callers.

**Rationale**: Tests are the safety net for refactoring and prevent regressions in production.

### V. Security

Secure defaults, certificate management alignment with cert-manager semantics, authentication/authorization at system boundaries. Validate external inputs; trust internal code. No security vulnerabilities (XSS, SQL injection, command injection, OWASP Top 10). Security-sensitive changes require explicit review.

**Rationale**: Edge devices operate in untrusted environments; security failures have fleet-wide impact.

### VI. Documentation Conciseness

Documentation MUST capture necessary information without verbosity. Target: artifacts humans can read, understand, and review in a single session. Avoid redundant explanations; prefer clarity over exhaustiveness. Focus on "why" (rationale, design decisions) over "what" (code should be self-evident).

**Rationale**: Overly verbose docs are not maintained or read; concise docs remain useful.

### VII. Observability

Structured logging required for all service operations. Metrics for fleet health, device status, API performance. System-info opt-out honored. Post-operation status collection (e.g., certificate renewal). Debuggability through text I/O where feasible.

**Rationale**: Operating fleets at scale requires visibility into system behavior and device state.

### VIII. Performance & Scale

Features MUST scale to thousands of devices. Avoid unnecessary database queries, N+1 patterns, or blocking operations. Resource efficiency matters (memory, CPU, storage). Performance degradation under load is a bug.

**Rationale**: Fleet management systems must handle growth without architectural rewrites.

## Quality Gates

All code MUST pass the following gates before completion:

1. **Linting**: `make lint` (or equivalent) with zero errors
2. **Testing**: All tests pass (`go test -v -race ./...` or equivalent)
3. **Interface Changes**: Immediate test execution after signature changes
4. **Security Review**: For authentication, authorization, certificate, or external input handling

**Non-Negotiable**: Lint errors and failing tests indicate unfinished work.

## Development Workflow

### Code Changes

1. Read existing code before proposing changes
2. Write tests for current behavior if coverage insufficient
3. Make minimal, focused changes (avoid scope creep)
4. Run linters and tests
5. Fix all errors before proceeding
6. Commit after each logical unit of work

### Pull Requests

PRs MUST include:
- Summary of changes (1-3 bullet points)
- Test plan or evidence of testing
- Breaking change notice (if applicable)
- Link to related issue/spec

Reviews verify:
- Constitution compliance (all principles)
- Test coverage and passing status
- Lint compliance
- Security considerations
- Documentation updates (if user-facing)

## Governance

This constitution supersedes all other development practices. Amendments require:

1. Documented rationale for change
2. Version bump per semantic versioning rules
3. Sync Impact Report identifying affected templates/docs
4. Update to dependent artifacts (plan/spec/tasks templates)

**Version**: 1.1.0 | **Ratified**: 2025-01-15 | **Last Amended**: 2026-01-22
