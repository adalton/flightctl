---

description: "Task list for package-mode support implementation"
---

# Tasks: Package Mode Support

**Input**: Design documents from `/specs/001-package-mode-support/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/

**Tests**: No test tasks included (not explicitly requested in specification)

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

---

**📋 Constitution Reminders**:
- **Quality Gates** (I): All code MUST pass `make lint` and tests before marking tasks complete
- **Test Coverage** (IV): Interface changes require immediate `go test -v -race ./...` execution
- **Documentation** (VI): Task descriptions should be concise and actionable

---

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

- **Monorepo paths**: `internal/agent/`, `internal/api/`, `test/integration/`
- Paths shown below use absolute references from repository root

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure validation

- [ ] T001 Verify existing agent structure matches plan.md project structure
- [ ] T002 Verify Go 1.24.0 toolchain and dependencies (go.mod matches plan requirements)
- [ ] T003 [P] Verify existing systeminfo collector pattern in internal/agent/device/systeminfo/manager.go

**Checkpoint**: Foundation ready - user story implementation can now begin

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [ ] T004 Create package-mode detection module in internal/agent/device/systeminfo/packagemode.go
- [ ] T005 [P] Add packageMode field to systeminfo manager struct in internal/agent/device/systeminfo/manager.go
- [ ] T006 Implement DetectPackageMode() function using exec.LookPath("bootc") in internal/agent/device/systeminfo/packagemode.go
- [ ] T007 [P] Add IsPackageMode() accessor method to systeminfo manager in internal/agent/device/systeminfo/manager.go
- [ ] T008 Call DetectPackageMode() in manager.Initialize() and cache result in internal/agent/device/systeminfo/manager.go
- [ ] T009 Add INFO logging for package-mode detection in manager.Initialize() in internal/agent/device/systeminfo/manager.go
- [ ] T010 Update systeminfo Status() method to populate additionalProperties["packageMode"] in internal/agent/device/systeminfo/manager.go
- [ ] T011 Run `make lint` to verify packagemode.go and manager.go changes pass linting
- [ ] T012 Run `go test -v -race ./internal/agent/device/systeminfo/...` to verify no regressions

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Agent Installation on Package-Based Systems (Priority: P1) 🎯 MVP

**Goal**: Agent installs successfully on RHEL 9+ and Ubuntu 22.04+ package-managed systems

**Independent Test**: Install agent via RPM on RHEL system and .deb on Ubuntu system. Agent starts successfully and enrolls with Flight Control service.

### Implementation for User Story 1

- [ ] T013 [P] [US1] Document RHEL package-mode installation procedure in docs/user/installing/installing-agent-rhel-package-mode.md
- [ ] T014 [P] [US1] Document Ubuntu package-mode installation procedure in docs/user/installing/installing-agent-ubuntu-package-mode.md
- [ ] T015 [US1] Verify agent systemd service definition supports package-mode deployments (no bootc dependency)
- [ ] T016 [US1] Update docs/user/installing/README.md with links to package-mode installation guides
- [ ] T017 [US1] Run `make lint` to verify documentation formatting

**Checkpoint**: At this point, User Story 1 should be fully functional - agent can be installed and started on package-mode systems

---

## Phase 4: User Story 2 - Automatic Environment Detection (Priority: P1)

**Goal**: Agent automatically detects package-mode vs image-mode and adapts behavior accordingly

**Independent Test**: Deploy agent on RHEL package-mode system. Agent status reports OS type and package-mode indicator. No error logs related to bootc operations.

### Implementation for User Story 2

- [ ] T018 [P] [US2] Add unit test TestDetectPackageMode_BootcPresent in internal/agent/device/systeminfo/packagemode_test.go
- [ ] T019 [P] [US2] Add unit test TestDetectPackageMode_BootcAbsent in internal/agent/device/systeminfo/packagemode_test.go
- [ ] T020 [P] [US2] Add unit test TestSystemInfoManager_PackageModeInStatus in internal/agent/device/systeminfo/packagemode_test.go
- [ ] T021 [US2] Run `go test -v -race ./internal/agent/device/systeminfo/...` to verify all tests pass
- [ ] T022 [US2] Verify agent logs INFO message on package-mode detection during startup (manual test)
- [ ] T023 [US2] Verify device status includes additionalProperties["packageMode"] = "true" for package-mode device (manual test)
- [ ] T024 [US2] Run `make lint` to verify test files pass linting

**Checkpoint**: At this point, User Story 2 should be fully functional - agent correctly detects and reports package-mode

---

## Phase 5: User Story 3 - Selective Update Management (Priority: P1)

**Goal**: Agent applies Flight Control config/app updates while ignoring OS updates in package-mode

**Independent Test**: Push configuration update via Flight Control to package-mode device. Configuration applies successfully. Run OS package update via dnf/apt. OS update completes independently without agent interference.

### Implementation for User Story 3

- [ ] T025 [US3] Modify beforeUpdate() in internal/agent/device/device.go to skip OS prefetch when packageMode detected (line ~340)
- [ ] T026 [US3] Add INFO logging "Package-mode device, skipping OS image prefetch" in beforeUpdate() in internal/agent/device/device.go
- [ ] T027 [US3] Modify afterUpdate() in internal/agent/device/device.go to skip OS switch/reboot when packageMode detected (line ~460)
- [ ] T028 [US3] Add INFO logging "Package-mode device, ignoring OS update to <image>" in afterUpdate() in internal/agent/device/device.go
- [ ] T029 [US3] Add unit test TestBeforeUpdate_PackageMode to verify OS prefetch skipped in internal/agent/device/device_test.go
- [ ] T030 [US3] Add unit test TestAfterUpdate_PackageMode to verify OS operations skipped in internal/agent/device/device_test.go
- [ ] T031 [US3] Run `go test -v -race ./internal/agent/device/...` to verify all tests pass
- [ ] T032 [US3] Run `make lint` to verify device.go changes pass linting
- [ ] T033 [US3] Manual test: Push config update to package-mode device, verify application without OS operations
- [ ] T034 [US3] Manual test: Run `dnf update` concurrently with Flight Control config update, verify no conflicts

**Checkpoint**: At this point, User Story 3 should be fully functional - agent correctly manages updates in package-mode

---

## Phase 6: User Story 4 - Console Visibility (Priority: P2)

**Goal**: Operators can see deployment mode (package-mode vs image-mode) in console UI

**Independent Test**: View device inventory in Flight Control console. Package-mode devices display mode indicator distinct from image-mode devices.

### Implementation for User Story 4

- [ ] T035 [US4] Update device status API handler to include additionalProperties in response (verify no changes needed in internal/api/device.go)
- [ ] T036 [US4] Document OpenAPI schema extension for packageMode field in api/core/v1beta1/openapi.yaml (informational comment)
- [ ] T037 [US4] Create API integration test to verify additionalProperties["packageMode"] returned in GET /api/v1/devices/:id response
- [ ] T038 [US4] Run `go test -v ./internal/api/...` to verify API tests pass
- [ ] T039 [US4] Manual test: Query device status API, verify packageMode field present in response
- [ ] T040 [US4] Update user documentation to explain packageMode indicator in docs/user/references/api-resources.md

**Checkpoint**: All user stories should now be independently functional - console visibility complete

---

## Phase 7: Integration Testing & Validation

**Purpose**: Cross-story validation and platform-specific testing

- [ ] T041 [P] Setup RHEL 9 test runner infrastructure in test/integration/fixtures/rhel9-runner/
- [ ] T042 [P] Setup Ubuntu 22.04 test runner infrastructure in test/integration/fixtures/ubuntu22-runner/
- [ ] T043 Create RHEL integration test in test/integration/agent/packagemode_rhel_test.go
- [ ] T044 Create Ubuntu integration test in test/integration/agent/packagemode_ubuntu_test.go
- [ ] T045 Integration test: Verify agent installation on RHEL 9 package-mode runner
- [ ] T046 Integration test: Verify package-mode detection on RHEL 9
- [ ] T047 Integration test: Verify config updates work on RHEL 9 package-mode
- [ ] T048 Integration test: Verify OS update ignored on RHEL 9 when spec contains OS image
- [ ] T049 Integration test: Verify agent installation on Ubuntu 22.04 package-mode runner
- [ ] T050 Integration test: Verify package-mode detection on Ubuntu 22.04
- [ ] T051 Integration test: Verify config updates work on Ubuntu 22.04 package-mode
- [ ] T052 Integration test: Verify OS update ignored on Ubuntu 22.04 when spec contains OS image
- [ ] T053 Run `make integration-test` to execute all integration tests
- [ ] T054 Run `make lint` on all test files to verify compliance

**Checkpoint**: Integration tests validate package-mode works correctly on both platforms

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [ ] T055 [P] Add Prometheus metric flightctl_agent_deployment_mode in internal/instrumentation/metrics.go
- [ ] T056 [P] Add Prometheus counter flightctl_agent_os_updates_skipped_total in internal/instrumentation/metrics.go
- [ ] T057 Instrument package-mode detection to emit deployment_mode metric in internal/agent/device/systeminfo/manager.go
- [ ] T058 Instrument OS update skipping to increment skipped counter in internal/agent/device/device.go
- [ ] T059 [P] Update architecture documentation with package-mode section in docs/developer/architecture/architecture.md
- [ ] T060 [P] Update troubleshooting guide with package-mode scenarios in docs/user/using/troubleshooting.md
- [ ] T061 Run complete test suite: `make unit-test && make integration-test`
- [ ] T062 Run linting on all modified files: `make lint`
- [ ] T063 Verify all existing image-mode tests still pass (regression check)
- [ ] T064 Review quickstart.md against actual installation experience, update if needed

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3-6)**: All depend on Foundational phase completion
  - US1 (Installation): Can start after Foundational - No dependencies on other stories
  - US2 (Detection): Can start after Foundational - No dependencies on other stories (but logically builds on US1)
  - US3 (Update Management): Depends on US2 (requires detection logic) - Should complete US2 first
  - US4 (Console Visibility): Can start after Foundational - No dependencies on other stories
- **Integration Testing (Phase 7)**: Depends on US1, US2, US3 being complete
- **Polish (Phase 8)**: Depends on all user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2) - No dependencies on other stories
- **User Story 2 (P1)**: Can start after Foundational (Phase 2) - No dependencies on other stories (documentation builds on US1 conceptually)
- **User Story 3 (P1)**: **Depends on User Story 2** (requires IsPackageMode() method from detection implementation)
- **User Story 4 (P2)**: Can start after Foundational (Phase 2) - No dependencies on other stories (but needs US2's status field population)

### Within Each User Story

**User Story 1** (Installation):
- Documentation tasks [US1] can all run in parallel (T013, T014)
- Service verification (T015) independent
- README update (T016) after individual docs complete
- Lint check (T017) after all docs written

**User Story 2** (Detection):
- Unit tests [US2] can all run in parallel (T018, T019, T020)
- Test execution (T021) after tests written
- Manual validation (T022, T023) after test execution passes
- Lint check (T024) after tests written

**User Story 3** (Update Management):
- beforeUpdate modification (T025) and afterUpdate modification (T027) can run in parallel
- Logging additions (T026, T028) with their respective modifications
- Unit tests (T029, T030) can run in parallel after modifications
- Test execution (T031) after tests written
- Lint (T032) and manual tests (T033, T034) sequential after test execution

**User Story 4** (Console Visibility):
- API handler verification (T035) and schema documentation (T036) can run in parallel
- Integration test creation (T037) independent
- Test execution (T038) after test creation
- Manual test (T039) and documentation (T040) can run in parallel after tests pass

**Integration Testing** (Phase 7):
- Test infrastructure setup (T041, T042) can run in parallel
- Test file creation (T043, T044) can run in parallel after infrastructure
- RHEL tests (T045-T048) sequential within RHEL platform
- Ubuntu tests (T049-T052) sequential within Ubuntu platform
- RHEL and Ubuntu test sequences can run in parallel
- Final execution (T053) and lint (T054) sequential after all tests written

**Polish** (Phase 8):
- Metric additions (T055, T056) can run in parallel
- Instrumentation (T057, T058) sequential after metrics defined
- Documentation updates (T059, T060) can run in parallel
- Final validation (T061-T064) sequential

### Parallel Opportunities

```bash
# Setup phase - all sequential (verification tasks)

# Foundational phase - sequential (building detection foundation)

# User Story 1 - Documentation in parallel:
Task T013: "Document RHEL installation in docs/user/installing/installing-agent-rhel-package-mode.md"
Task T014: "Document Ubuntu installation in docs/user/installing/installing-agent-ubuntu-package-mode.md"

# User Story 2 - Tests in parallel:
Task T018: "TestDetectPackageMode_BootcPresent in internal/agent/device/systeminfo/packagemode_test.go"
Task T019: "TestDetectPackageMode_BootcAbsent in internal/agent/device/systeminfo/packagemode_test.go"
Task T020: "TestSystemInfoManager_PackageModeInStatus in internal/agent/device/systeminfo/packagemode_test.go"

# User Story 3 - Modifications in parallel:
Task T025: "Modify beforeUpdate() in internal/agent/device/device.go"
Task T027: "Modify afterUpdate() in internal/agent/device/device.go"
# Then tests in parallel:
Task T029: "TestBeforeUpdate_PackageMode in internal/agent/device/device_test.go"
Task T030: "TestAfterUpdate_PackageMode in internal/agent/device/device_test.go"

# User Story 4 - Verification and docs in parallel:
Task T035: "Verify API handler in internal/api/device.go"
Task T036: "Document schema in api/core/v1beta1/openapi.yaml"

# Integration Testing - Infrastructure in parallel:
Task T041: "Setup RHEL 9 runner in test/integration/fixtures/rhel9-runner/"
Task T042: "Setup Ubuntu 22.04 runner in test/integration/fixtures/ubuntu22-runner/"
# Then test creation in parallel:
Task T043: "Create RHEL test in test/integration/agent/packagemode_rhel_test.go"
Task T044: "Create Ubuntu test in test/integration/agent/packagemode_ubuntu_test.go"

# Polish - Metrics and docs in parallel:
Task T055: "Add deployment_mode metric in internal/instrumentation/metrics.go"
Task T056: "Add os_updates_skipped counter in internal/instrumentation/metrics.go"
Task T059: "Update architecture docs in docs/developer/architecture/architecture.md"
Task T060: "Update troubleshooting in docs/user/using/troubleshooting.md"
```

---

## Implementation Strategy

### MVP First (User Story 1 + User Story 2 + User Story 3)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL - blocks all stories)
3. Complete Phase 3: User Story 1 (Installation documentation)
4. Complete Phase 4: User Story 2 (Detection logic) - **Depends on Foundational**
5. Complete Phase 5: User Story 3 (Update management) - **Depends on User Story 2**
6. **STOP and VALIDATE**: Test package-mode agent on RHEL/Ubuntu
7. Deploy/demo if ready

This MVP delivers core functionality: package-mode detection and selective update management.

### Incremental Delivery

1. Complete Setup + Foundational → Foundation ready
2. Add User Story 1 (Installation docs) → Deliverable: Installation guides
3. Add User Story 2 (Detection) → Test independently → Deliverable: Mode detection working
4. Add User Story 3 (Update mgmt) → Test independently → Deliverable: Updates work correctly (MVP complete!)
5. Add User Story 4 (Console visibility) → Test independently → Deliverable: Full feature complete
6. Add Integration Testing → Validate on real systems
7. Add Polish → Production-ready

Each story adds value without breaking previous stories.

### Parallel Team Strategy

With multiple developers after Foundational phase:

1. Team completes Setup + Foundational together
2. Once Foundational is done:
   - Developer A: User Story 1 (Documentation)
   - Developer B: User Story 2 (Detection)
   - Developer C: User Story 4 (Console visibility - independent)
3. After User Story 2 complete:
   - Developer B: User Story 3 (depends on detection from US2)
4. After all stories complete:
   - Team collaborates on Integration Testing and Polish

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- User Story 3 depends on User Story 2 (detection logic required)
- Most user stories (1, 2, 4) are independently completable after Foundational phase
- No test tasks included (tests not explicitly requested in spec.md)
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- Avoid: vague tasks, same file conflicts, cross-story dependencies that break independence
- **Quality Gate**: Run `make lint` and `make unit-test` before marking implementation complete
- **Constitution Compliance**: All code must pass linting before completion (Principle I)
