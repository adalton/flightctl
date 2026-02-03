# [FEATURE]

**Author(s)**: [Author's Name]
**Status**: Draft | In Review | Final
**Jira feature**: [Jira issue link]
**Branch**: `[###-feature-name]`
**Date**: [DATE]
**Spec**: [link to spec.md]

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

# 1. Overview

[Extract from feature spec: primary requirement + technical approach from research]

# 2. Technical Context

<!--
  ACTION REQUIRED: Replace the content in this section with the technical details
  for the project. The structure here is presented in advisory capacity to guide
  the iteration process.
-->

**Language/Version**: [e.g., Go 1.22, Python 3.11 or NEEDS CLARIFICATION]
**Primary Dependencies**: [e.g., Kubernetes client-go, gRPC or NEEDS CLARIFICATION]
**Storage**: [if applicable, e.g., PostgreSQL, SQLite, files or N/A]
**Testing**: [e.g., go test, pytest or NEEDS CLARIFICATION]
**Target Platform**: [e.g., Linux server, RHEL 9+, Ubuntu 22.04+ or NEEDS CLARIFICATION]
**Project Type**: [single/web/mobile - determines source structure]
**Performance Goals**: [domain-specific, e.g., 1000 req/s, <100ms p95 latency or NEEDS CLARIFICATION]
**Constraints**: [domain-specific, e.g., <200ms p95, <100MB memory, offline-capable or NEEDS CLARIFICATION]
**Scale/Scope**: [domain-specific, e.g., 10k devices, 1M requests/day or NEEDS CLARIFICATION]

# 3. Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

Verify compliance with Flight Control Constitution (v1.0.0):

**I. Test-First Development**
- [ ] TDD approach documented: Tests will be written before implementation
- [ ] Test approval process defined: Test files submitted for review before coding begins
- [ ] Contract tests identified for: [list new service interfaces, API changes, shared structures]
- [ ] Integration tests scoped for: [list critical workflows, service boundaries, schema changes]

**II. Code Quality & Standards**
- [ ] Linting verified: `make lint` passes on feature branch
- [ ] Go best practices confirmed: Error handling, import ordering, security considerations documented
- [ ] Interface changes inventoried: [list any interface signature changes with migration plan]
- [ ] Documentation plan: "Why" comments planned for complex logic

**III. Observability**
- [ ] Structured logging plan: Key state transitions and request tracing identified
- [ ] Metrics plan: Request rates, error rates, resource utilization, business metrics defined
- [ ] Tracing plan: Critical paths instrumented (if applicable)

**Complexity Justifications** (complete ONLY if introducing patterns beyond standard practices):

| Potential Violation | Why Needed | Simpler Alternative Rejected Because |
|---------------------|------------|--------------------------------------|
| [e.g., New abstraction layer] | [specific problem] | [why direct approach insufficient] |

# 4. Goals and Non-Goals

## 4.1 Goals

* [List the goals for this feature]

## 4.2 Non-Goals

* [List things this feature explicitly is NOT implementing]

# 5. Motivation / Background

[Describe the existing problem, limitations of the current system, and the rationale for this proposal.]

# 6. Design

## 6.1 Architecture

[High-level overview, diagrams, and component responsibilities.]

## 6.2 Data Model / Schema Changes

[Detail any new tables, fields, constraints, or index changes.]

## 6.3 API Changes

[Specify new endpoints, request/response formats, and versioning impact.]

## 6.4 Scalability

Estimate:

* Memory and CPU usage.

* Expected database load (reads/writes).

* Data retention and cleanup policies.

* Future growth assumptions.

## 6.5 Security Considerations

[Cover potential vulnerabilities, authentication/authorization, data exposure risks.]

## 6.6 Failure Handling and Recovery

[Explain behavior under partial failure, retries, idempotency, and recovery flows.]

## 6.7 RBAC / Tenancy

Describe:

* Role-based access rules.
* Tenancy or org/resource isolation.
* Visibility constraints and edge cases.

## 6.8 Extensibility / Future-Proofing

[Explain how the design accommodates future enhancements.]

# 7. Project Structure

## 7.1 Documentation (this feature)

```text
specs/[###-feature]/
├── spec.md              # Feature specification (/speckit.specify command output)
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
├── checklists/          # Quality validation checklists
│   └── requirements.md  # Spec quality checklist
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

## 7.2 Source Code (repository root)

<!--
  ACTION REQUIRED: Replace the placeholder tree below with the concrete layout
  for this feature. Delete unused options and expand the chosen structure with
  real paths (e.g., internal/agent, pkg/api). The delivered plan must not
  include Option labels.
-->

```text
# [REMOVE IF UNUSED] Option 1: Single project (DEFAULT for Flight Control)
internal/
├── agent/
├── api/
├── store/
└── service/

pkg/
├── api/v1alpha1/
└── client/

test/
├── e2e/
├── integration/
└── unit/

# [REMOVE IF UNUSED] Option 2: Web application (when "frontend" + "backend" detected)
backend/
├── internal/
│   ├── models/
│   ├── service/
│   └── api/
└── test/

frontend/
├── src/
│   ├── components/
│   ├── pages/
│   └── services/
└── tests/

# [REMOVE IF UNUSED] Option 3: Mobile + API (when "iOS/Android" detected)
api/
└── [same as backend above]

ios/ or android/
└── [platform-specific structure: feature modules, UI flows, platform tests]
```

**Structure Decision**: [Document the selected structure and reference the real directories captured above]

# 8. Alternatives Considered

[Briefly explain other approaches evaluated and why they were not selected.]

# 9. Observability and Monitoring

[List new metrics, events, and alerts.]

# 10. Impact and Compatibility

Note any:

* Backward-incompatible changes.
* DB migration impacts.
* Changes to existing APIs or workflows.

# 11. Open Questions

* [List any open questions]

---

## Appendix: Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|--------------------------------------|
| [e.g., 4th project] | [current need] | [why 3 projects insufficient] |
| [e.g., Repository pattern] | [specific problem] | [why direct DB access insufficient] |
