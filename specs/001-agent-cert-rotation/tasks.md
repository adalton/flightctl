# Tasks: Agent Certificate Rotation

**Input**: Design documents from `specs/001-agent-cert-rotation/`
**Prerequisites**: plan.md, spec.md, data-model.md, contracts/renewal-api.yaml, research.md

**Tests**: TDD approach - tests are written FIRST before implementation per constitution requirement III

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

Flight Control uses the following structure:
- Agent code: `internal/agent/device/`
- Service code: `internal/service/`
- Database store: `internal/service/store/`
- API definitions: `internal/api/v1beta1/`
- Tests: `test/integration/` and `test/contract/`
- Migrations: `db/migrations/`

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure

- [X] T001 Create agent-side package structure in internal/agent/device/certrotation/
- [X] T002 Create service-side package structure in internal/service/certrotation/
- [X] T003 [P] Add certificate rotation configuration to internal/agent/config/config.go
- [X] T004 [P] Create database migration file db/migrations/20251202133504_add_cert_renewal.sql
- [X] T005 [P] Create integration test directory test/integration/certrotation/
- [X] T006 [P] Create contract test file test/contract/certrotation_contract_test.go

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [ ] T007 Apply database migration to create certificate_renewal_requests table
- [ ] T008 [P] Implement CertificateMetadata type in internal/agent/device/certrotation/types.go
- [ ] T009 [P] Implement RenewalRequest type in internal/agent/device/certrotation/types.go
- [ ] T010 [P] Implement CertRotationConfig type in internal/agent/config/config.go
- [ ] T011 [P] Implement CertificateRenewalRequest database entity in internal/service/store/certrotation.go
- [ ] T012 Implement database store methods (Create, Update, Get, List) in internal/service/store/certrotation.go
- [ ] T013 [P] Add OpenTelemetry tracing utilities for certificate operations in internal/agent/device/certrotation/tracing.go
- [ ] T014 [P] Add metrics definitions for renewal operations in internal/agent/instrumentation/metrics/

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Automatic Certificate Renewal (Priority: P1) 🎯 MVP

**Goal**: Devices automatically renew certificates 30 days before expiration without manual intervention

**Independent Test**: Deploy test device with cert expiring in 31 days, verify automatic renewal at 30-day threshold with no service interruption

### Tests for User Story 1 (TDD - Write FIRST) ⚠️

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [ ] T015 [P] [US1] Write contract test for renewal endpoint in test/contract/certrotation_contract_test.go
- [ ] T016 [P] [US1] Write integration test for proactive renewal flow in test/integration/certrotation/renewal_test.go
- [ ] T017 [P] [US1] Write unit test for expiration calculation in internal/agent/device/certrotation/monitor_test.go
- [ ] T018 [P] [US1] Write integration test for network interruption retry in test/integration/certrotation/retry_test.go

### Implementation for User Story 1

- [ ] T019 [P] [US1] Implement certificate expiration monitor in internal/agent/device/certrotation/monitor.go
- [ ] T020 [P] [US1] Implement CSR generation for renewal in internal/agent/device/certrotation/renewer.go
- [ ] T021 [US1] Integrate monitor with agent status update cycle in internal/agent/device/status/status.go
- [ ] T022 [US1] Implement retry queue for renewal requests in internal/agent/device/certrotation/retry.go
- [ ] T023 [P] [US1] Implement renewal endpoint handler in internal/service/certrotation/handler.go
- [ ] T024 [P] [US1] Implement security proof validator for valid certificates in internal/service/certrotation/validator.go
- [ ] T025 [US1] Implement certificate issuer in internal/service/certrotation/issuer.go
- [ ] T026 [US1] Register renewal endpoint in API router internal/api/v1beta1/router.go
- [ ] T027 [US1] Add renewal request/response models to internal/api/v1beta1/models.go
- [ ] T028 [US1] Update OpenAPI specification in internal/api/v1beta1/openapi.yaml
- [ ] T029 [US1] Add tracing for renewal operations in internal/agent/device/certrotation/renewer.go
- [ ] T030 [US1] Add metrics emission for renewal attempts in internal/agent/device/certrotation/monitor.go
- [ ] T031 [US1] Add structured logging for renewal lifecycle in internal/agent/device/certrotation/renewer.go

**Checkpoint**: At this point, User Story 1 should be fully functional and testable independently. Devices proactively renew certificates before expiration.

---

## Phase 4: User Story 3 - Atomic Certificate Rotation (Priority: P1)

**Goal**: Certificate swaps are atomic (all-or-nothing) to prevent devices from becoming unreachable

**Independent Test**: Simulate power loss, network failure, and disk errors during rotation - verify device always retains at least one valid certificate

**Note**: US3 comes before US2 because atomic swap is needed for both proactive renewal (US1) and recovery (US2)

### Tests for User Story 3 (TDD - Write FIRST) ⚠️

- [ ] T032 [P] [US3] Write integration test for atomic swap success in test/integration/certrotation/atomic_test.go
- [ ] T033 [P] [US3] Write integration test for rollback on validation failure in test/integration/certrotation/atomic_test.go
- [ ] T034 [P] [US3] Write integration test for power loss during swap in test/integration/certrotation/atomic_test.go
- [ ] T035 [P] [US3] Write unit test for idempotent retry in internal/agent/device/certrotation/atomic_test.go

### Implementation for User Story 3

- [ ] T036 [P] [US3] Implement atomic file write operations in internal/agent/device/certrotation/atomic.go
- [ ] T037 [P] [US3] Implement certificate validation before swap in internal/agent/device/certrotation/atomic.go
- [ ] T038 [US3] Implement rollback mechanism for failed swaps in internal/agent/device/certrotation/atomic.go
- [ ] T039 [US3] Integrate atomic swap into renewal flow in internal/agent/device/certrotation/renewer.go
- [ ] T040 [US3] Add backup certificate preservation (.old suffix) in internal/agent/device/certrotation/atomic.go
- [ ] T041 [US3] Add tracing for swap operations in internal/agent/device/certrotation/atomic.go
- [ ] T042 [US3] Add metrics for swap success/failure in internal/agent/device/certrotation/atomic.go

**Checkpoint**: At this point, User Stories 1 AND 3 work together. Certificate renewal is atomic and safe from interruptions.

---

## Phase 5: User Story 2 - Expired Certificate Recovery (Priority: P2)

**Goal**: Devices with expired certificates automatically recover using bootstrap cert or TPM attestation

**Independent Test**: Deploy device with expired management cert and valid bootstrap cert, verify automatic recovery on startup

### Tests for User Story 2 (TDD - Write FIRST) ⚠️

- [ ] T043 [P] [US2] Write integration test for bootstrap cert recovery in test/integration/certrotation/recovery_test.go
- [ ] T044 [P] [US2] Write integration test for TPM attestation recovery in test/integration/certrotation/recovery_test.go
- [ ] T045 [P] [US2] Write unit test for expired certificate detection in internal/agent/device/certrotation/recovery_test.go

### Implementation for User Story 2

- [ ] T046 [P] [US2] Implement expired certificate detection in internal/agent/device/certrotation/recovery.go
- [ ] T047 [P] [US2] Implement bootstrap certificate fallback logic in internal/agent/device/certrotation/recovery.go
- [ ] T048 [P] [US2] Implement TPM attestation proof generation in internal/agent/device/certrotation/recovery.go
- [ ] T049 [US2] Integrate recovery handler into agent bootstrap in internal/agent/device/bootstrap.go
- [ ] T050 [US2] Implement bootstrap cert validation in internal/service/certrotation/validator.go
- [ ] T051 [US2] Implement TPM attestation validation in internal/service/certrotation/validator.go
- [ ] T052 [US2] Add tracing for recovery operations in internal/agent/device/certrotation/recovery.go
- [ ] T053 [US2] Add metrics for recovery attempts in internal/agent/device/certrotation/recovery.go
- [ ] T054 [US2] Add structured logging for recovery flow in internal/agent/device/certrotation/recovery.go

**Checkpoint**: All user stories are now independently functional. Devices can renew proactively (US1), recover from expiration (US2), and do so atomically (US3).

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [ ] T055 [P] Update docs/user/references/security-guidelines.md with certificate rotation security best practices
- [ ] T056 [P] Update docs/developer/architecture/ with certificate rotation architecture documentation
- [ ] T057 [P] Add certificate rotation monitoring to docs/user/using/device-observability.md
- [ ] T058 [P] Create example configurations in examples/certificate-rotation/
- [ ] T059 [P] Add certificate rotation troubleshooting section to docs/user/using/troubleshooting.md
- [ ] T060 Run quickstart.md validation scenarios locally
- [ ] T061 Performance test database queries with 100k renewal records
- [ ] T062 Load test service with 10,000 concurrent renewal requests
- [ ] T063 [P] Code cleanup and refactoring across all packages
- [ ] T064 Run make lint and fix all linting errors
- [ ] T065 Run make unit-test and verify all tests pass
- [ ] T066 Security review of renewal endpoint and validation logic

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3-5)**: All depend on Foundational phase completion
  - **US1 (Phase 3)**: Can start after Foundational - Independent
  - **US3 (Phase 4)**: Can start after US1 (atomic swap needed for renewal) - Depends on US1
  - **US2 (Phase 5)**: Can start after US1 and US3 (uses renewal + atomic swap) - Depends on US1, US3
- **Polish (Phase 6)**: Depends on all user stories being complete

### User Story Dependencies

- **User Story 1 (P1 - Automatic Renewal)**: No dependencies on other stories (can start after Foundational)
- **User Story 3 (P1 - Atomic Rotation)**: Depends on User Story 1 (integrates with renewal flow)
- **User Story 2 (P2 - Recovery)**: Depends on User Story 1 and 3 (reuses renewal and atomic swap)

### Within Each User Story

- Tests MUST be written and FAIL before implementation
- Agent-side and service-side tasks can run in parallel where marked [P]
- Integration points (e.g., T021, T026) must wait for both sides to complete
- Observability (tracing, metrics, logging) added after core functionality works

### Parallel Opportunities

**Setup Phase (All Parallel)**:
- T001 Agent package + T002 Service package + T003 Config + T004 Migration + T005 Test dirs + T006 Contract test

**Foundational Phase (Types Parallel)**:
- T008 CertMetadata + T009 RenewalRequest + T010 Config + T011 DB Entity + T013 Tracing + T014 Metrics
- Then T007 Migration + T012 Store methods (sequential)

**User Story 1 - Tests (All Parallel)**:
- T015 Contract test + T016 Integration test + T017 Unit test + T018 Retry test

**User Story 1 - Implementation (Agent/Service Parallel)**:
```
Agent side (parallel): T019 Monitor + T020 Renewer
Agent integration: T021 Status integration, T022 Retry queue
Service side (parallel): T023 Handler + T024 Validator + T025 Issuer
Service integration: T026 Router + T027 Models + T028 OpenAPI
Observability (parallel): T029 Tracing + T030 Metrics + T031 Logging
```

**User Story 3 - Tests (All Parallel)**:
- T032 + T033 + T034 + T035 (4 atomic swap tests)

**User Story 3 - Implementation (Parallel Where Possible)**:
- T036 File ops + T037 Validation (parallel)
- T038 Rollback → T039 Integration (sequential)
- T040 Backup + T041 Tracing + T042 Metrics (parallel)

**User Story 2 - Tests (All Parallel)**:
- T043 + T044 + T045 (3 recovery tests)

**User Story 2 - Implementation (Agent/Service Parallel)**:
```
Agent side: T046 Detection + T047 Bootstrap + T048 TPM (parallel)
Agent integration: T049 Bootstrap integration
Service side: T050 Bootstrap validation + T051 TPM validation (parallel)
Observability: T052 Tracing + T053 Metrics + T054 Logging (parallel)
```

**Polish Phase (Most Parallel)**:
- T055 Security docs + T056 Architecture docs + T057 Observability docs + T058 Examples + T059 Troubleshooting
- Then T060-T066 sequential (testing and validation)

---

## Parallel Example: User Story 1 Implementation

```bash
# Launch all parallel agent-side tasks together:
Task T019: "Implement certificate expiration monitor in internal/agent/device/certrotation/monitor.go"
Task T020: "Implement CSR generation for renewal in internal/agent/device/certrotation/renewer.go"

# After agent-side basics done, launch parallel service-side tasks:
Task T023: "Implement renewal endpoint handler in internal/service/certrotation/handler.go"
Task T024: "Implement security proof validator in internal/service/certrotation/validator.go"
Task T025: "Implement certificate issuer in internal/service/certrotation/issuer.go"

# After both sides done, launch parallel observability tasks:
Task T029: "Add tracing for renewal operations"
Task T030: "Add metrics emission for renewal attempts"
Task T031: "Add structured logging for renewal lifecycle"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

**Fastest path to value**:

1. Complete Phase 1: Setup (T001-T006)
2. Complete Phase 2: Foundational (T007-T014)
3. Complete Phase 3: User Story 1 (T015-T031)
4. **STOP and VALIDATE**: Test proactive renewal independently
5. Deploy/demo if ready

**Deliverable**: Devices automatically renew certificates before expiration. This alone eliminates 90% of manual certificate management burden.

### Add Critical Reliability (User Story 3)

After US1 validated:

1. Complete Phase 4: User Story 3 (T032-T042)
2. **VALIDATE**: Test atomic operations with failure scenarios
3. Deploy/demo enhanced reliability

**Deliverable**: Certificate renewal is now safe from power loss and network interruptions. Devices cannot become unreachable.

### Add Recovery Capability (User Story 2)

After US1 + US3 validated:

1. Complete Phase 5: User Story 2 (T043-T054)
2. **VALIDATE**: Test recovery scenarios independently
3. Deploy/demo complete feature

**Deliverable**: Devices that have been offline for extended periods can automatically recover. Complete certificate lifecycle management.

### Incremental Delivery Timeline

```
Week 1: Setup + Foundational → Foundation ready
Week 2-3: User Story 1 → Test independently → Deploy/Demo (MVP!)
Week 4: User Story 3 → Test independently → Deploy/Demo (Enhanced)
Week 5: User Story 2 → Test independently → Deploy/Demo (Complete)
Week 6: Polish → Final validation → Production rollout
```

Each phase adds value without breaking previous functionality.

### Parallel Team Strategy

With multiple developers:

1. **Team completes Setup + Foundational together** (T001-T014)
2. Once Foundational is done:
   - **Developer A**: User Story 1 agent-side (T019-T022)
   - **Developer B**: User Story 1 service-side (T023-T028)
   - **Developer C**: User Story 1 tests (T015-T018)
3. Integration meeting to connect components (T021, T026)
4. Team validates US1 together
5. Repeat pattern for US3 and US2

---

## Notes

- [P] tasks = different files, no dependencies - can run in parallel
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- **TDD workflow**: Write tests first (T015-T018), verify they fail, then implement (T019-T031)
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- Run `make lint` and `make unit-test` frequently during implementation
- Avoid: vague tasks, same file conflicts, cross-story dependencies that break independence

## Task Count Summary

- **Total Tasks**: 66
- **Setup Phase**: 6 tasks
- **Foundational Phase**: 8 tasks (BLOCKS all stories)
- **User Story 1 (P1)**: 17 tasks (4 tests + 13 implementation)
- **User Story 3 (P1)**: 11 tasks (4 tests + 7 implementation)
- **User Story 2 (P2)**: 12 tasks (3 tests + 9 implementation)
- **Polish Phase**: 12 tasks

**Parallel Opportunities**: 38 tasks marked [P] can run in parallel (58% of total)

**MVP Scope (Suggested)**: T001-T031 (31 tasks) delivers User Story 1 - automatic certificate renewal
