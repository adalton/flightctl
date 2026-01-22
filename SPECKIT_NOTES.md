# Testing

## Setup

1. `mkdir ~/src/speckit.EDM1471`
2. `cd ~/src/speckit.EDM1471`
3. `git clone git@github.com:adalton/flightctl.git`
4. `cd flightctl`
5. `git checkout -b speckit/EDM-1471`
6. Define the following function:

   ```bash
   specify() {
       uvx --from git+https://github.com/github/spec-kit.git specify "$@"
   }
   ```
7. Init speckit in the `flightctl` directory:

   ```bash
   $ specify init --here --ai claude`
   ...
   1. You're already in the project directory!
   2. Start using slash commands with your AI agent:
      2.1 /speckit.constitution - Establish project principles
      2.2 /speckit.specify - Create baseline specification
      2.3 /speckit.plan - Create implementation plan
      2.4 /speckit.tasks - Generate actionable tasks
      2.5 /speckit.implement - Execute implementation
   
   Enhancement Commands:
   
   Optional commands that you can use for your specs (improve quality & confidence)
   
   ○ /speckit.clarify (optional) - Ask structured questions to de-risk ambiguous areas before
   planning (run before /speckit.plan if used)
   ○ /speckit.analyze (optional) - Cross-artifact consistency & alignment report (after
   /speckit.tasks, before /speckit.implement)
   ○ /speckit.checklist (optional) - Generate quality checklists to validate requirements
   completeness, clarity, and consistency (after /speckit.plan)
   ```

## Start specifying EDM-1471

1. `claude`
2. Define the constitution for the project

   ```text
   /speckit.constitution In addition to the principles we define as part of this
   workflow, I want the following:
   - Generated documentation must not be overly verbose.  It must be verbose enough
     to accurately capture the information necessary to complete the task.  We want
     to avoid overly long/verbose documents.  The goal is to capture the needed
     information, while at the same time producing artifacts that humans can
     read/understand/review.
   ```

   - 7+ principles
   - Most critical principles:
     - Edge Device Management
     - Code Quality & Simplicity
     - Api Stability
     - Unit tests with good code coverage
   - Specify quality gates? Yes - lint + test
   - When was this constitution ratified? 2025 or earlier
   - Specific ratification date? 2025-06-01 (not sure how this matters)
   - Additional principles beyond the three earlier:
     - Observability
     - Security
     - Documentation
     - Performance & Scale
   - Constitution version: 1.0.0

   The resulting constitution included:

   ```markdown
   ### VI. Documentation Conciseness

   Documentation MUST capture necessary information without verbosity. Target: artifacts humans can read, understand, and review in a single session. Avoid redundant explanations; prefer clarity over exhaustiveness. Focus on "why" (rationale, design decisions) over "what" (code should be self-evident).

   **Rationale**: Overly verbose docs are not maintained or read; concise docs remain useful.
   ```

   - After accepting the spec template, I interrupted claude and provided the
     following prompt:

     ```text
     For the spec template, should we add something to emphaise the
     Documentation Conciseness from the constitution?
     ```

     It added the following to the spec template:

     ```markdown
     **📋 Constitution VI - Documentation Conciseness**: This spec MUST capture
     necessary information without verbosity. Target: readable and reviewable in
     a single session. Focus on "why" (rationale, decisions) over "what" (which
     should be clear from context). Avoid redundant explanations.
     ```

     It also added this to the plan template.

     ```markdown
     **📋 Constitution VI - Documentation Conciseness**: This plan MUST be
     readable in a single session. Focus on decisions and rationale, not
     exhaustive documentation. Avoid verbose explanations.
     ```

     And this to the tasks-template:

     ```markdown
     **📋 Constitution Reminders**:
     - **Quality Gates** (I): All code MUST pass `make lint` and tests before marking tasks complete
     - **Test Coverage** (IV): Interface changes require immediate `go test -v -race ./...` execution
     - **Documentation** (VI): Task descriptions should be concise and actionable
     ```

3. Commit the changes so far

## Begin spec-ing EDM-1471

1. `/speckit.specify EDM-1471` (I have a Jira MCP server, so it's able to consume that issue directly)
   - It eventually invokes:
     ```
     .specify/scripts/bash/create-new-feature.sh --json --number 1 --short-name "package-mode-support" "Support Flight Control Agent on Non-Image-Mode Devices (EDM-1471)"
     ```
   - Promising message along the way:
     ```
     ● Now let me draft the specification based on the JIRA issue details. I'll create a concise spec following the Constitution VI principle:
     ```
   - It created a new branch, `001-package-mode-support` along the way

2. Let's examine the new files:

   ```bash
   $ find specs -type f | xargs wc -l
     149 specs/001-package-mode-support/spec.md
      60 specs/001-package-mode-support/checklists/requirements.md
     209 total
   ```

   Rendered, `spec.md` is 2.5 pages and `requirements.md` is 1 page.

3. Add new files and commit to new branch

   ```bash
   $ git add SPECKIT_NOTES.md
   $ git add specs/
   $ git commit
   ```

## Begin planning EDM-1471

1. `/speckit.plan`

   - Invokes `.specify/scripts/bash/setup-plan.sh --json`
   - Generates a bunch of files, I accepted everything
   - Invokes `.specify/scripts/bash/update-agent-context.sh claude`
   - Eventually concludes with:

     ```text
     ● Implementation Plan Complete ✅
     
       The planning phase for package-mode support (EDM-1471) has been successfully completed.
     
       Branch & Artifacts
     
       Branch: 001-package-mode-support
       Plan File: /home/andalton/src/speckit.EDM1471/flightctl/specs/001-package-mode-support/plan.md
     
       Generated Artifacts
     
       Phase 0: Research ✅
       - research.md - Bootc detection strategy, update path analysis, testing approach
     
       Phase 1: Design ✅
       - data-model.md - Device status schema extension (backward compatible)
       - contracts/device-status.yaml - OpenAPI contract for package-mode indicator
       - quickstart.md - RHEL/Ubuntu installation and configuration guide
       - CLAUDE.md - Updated agent context with Go 1.24, GORM, Ginkgo
     
       Constitution Check Results
     
       All gates PASSED ✅
     
       - ✅ Code Quality & Simplicity: Simple bootc presence check, extends existing patterns
       - ✅ Edge Device Management: Supports mixed fleets (package + image mode)
       - ✅ API Stability: Backward compatible, no breaking changes
       - ✅ Test Coverage: Unit + integration tests planned (RHEL/Ubuntu runners)
       - ✅ Security: No new auth boundaries, validates detection logic
       - ✅ Documentation: Concise spec/plan, focused installation guides
       - ✅ Observability: Logging, metrics for mode detection and distribution
       - ✅ Performance & Scale: <1ms detection, scales to heterogeneous fleets
       - ✅ Quality Gates: make lint + make unit-test + make integration-test
     
       Technical Approach
     
       Core Decision: Use existing exec.LookPath("bootc") pattern to detect deployment mode at agent startup.
     
       Key Modifications (3 files):
       1. internal/agent/device/systeminfo/packagemode.go - Detection logic
       2. internal/agent/device/systeminfo/manager.go - Cache detection result
       3. internal/agent/device/device.go - Skip OS updates when package-mode detected
     
       Data Model: Add packageMode boolean to DeviceSystemInfo.additionalProperties (no schema change, backward compatible)
     
       File Impact Summary
       ┌────────────┬────────────────┬──────────────────────────────────────┐
       │  Category  │ Files Modified │            Files Created             │
       ├────────────┼────────────────┼──────────────────────────────────────┤
       │ Agent Code │ 2              │ 1 (packagemode.go)                   │
       ├────────────┼────────────────┼──────────────────────────────────────┤
       │ Data Model │ 0              │ 0 (uses additionalProperties)        │
       ├────────────┼────────────────┼──────────────────────────────────────┤
       │ API        │ 0              │ 1 (contract doc)                     │
       ├────────────┼────────────────┼──────────────────────────────────────┤
       │ Tests      │ 0              │ 3 (unit + RHEL + Ubuntu integration) │
       ├────────────┼────────────────┼──────────────────────────────────────┤
       │ Docs       │ 0              │ 2 (RHEL/Ubuntu install guides)       │
       └────────────┴────────────────┴──────────────────────────────────────┘
       Next Steps
     
       You can now proceed with:
     
       1. /speckit.tasks - Generate detailed implementation tasks organized by user story
       2. Manual Implementation - Use research.md and data-model.md to guide development
       3. Review Planning Artifacts - Share plan.md and contracts with team for feedback
     ```

2. Let's examine the new files:

   ```text
   $ git status
   On branch 001-package-mode-support
   Changes not staged for commit:
     (use "git add <file>..." to update what will be committed)
     (use "git restore <file>..." to discard changes in working directory)
           modified:   SPECKIT_NOTES.md
   
   Untracked files:
     (use "git add <file>..." to include in what will be committed)
           specs/001-package-mode-support/contracts/
           specs/001-package-mode-support/data-model.md
           specs/001-package-mode-support/plan.md
           specs/001-package-mode-support/quickstart.md
           specs/001-package-mode-support/research.md
   
   $ find specs/001-package-mode-support/contracts/ \
          specs/001-package-mode-support/data-model.md \
          specs/001-package-mode-support/plan.md \
          specs/001-package-mode-support/quickstart.md \
          specs/001-package-mode-support/research.md -type f | xargs wc -l
     256 specs/001-package-mode-support/contracts/device-status.yaml
     336 specs/001-package-mode-support/data-model.md
      99 specs/001-package-mode-support/plan.md
     422 specs/001-package-mode-support/quickstart.md
     345 specs/001-package-mode-support/research.md
    1458 total
   ```

   Rendered page lengths:
   - `specs/001-package-mode-support/data-model.md` is 3.5 pages
   - `specs/001-package-mode-support/plan.md`       is 1.5 pages
   - `specs/001-package-mode-support/quickstart.md` is 5   pages 
   - `specs/001-package-mode-support/research.md`   is 4   pages

3. Add new artifacts, commit changes

   ```bash
   $ git add SPECKIT_NOTES.md specs
   $ git status
   On branch 001-package-mode-support
   Changes to be committed:
     (use "git restore --staged <file>..." to unstage)
           modified:   SPECKIT_NOTES.md
           new file:   specs/001-package-mode-support/contracts/device-status.yaml
           new file:   specs/001-package-mode-support/data-model.md
           new file:   specs/001-package-mode-support/plan.md
           new file:   specs/001-package-mode-support/quickstart.md
           new file:   specs/001-package-mode-support/research.md
   $ git commit
   ```

## Begin defining EDM-1471 tasks

1. `/speckit.tasks`

   - Invokes `.specify/scripts/bash/check-prerequisites.sh --json`

2. Let's examine the new files:

   ```bash
   $ wc -l specs/001-package-mode-support/tasks.md
   345 specs/001-package-mode-support/tasks.md
   ```

   Rendered:
   - `specs/001-package-mode-support/tasks.md` - 5 pages

3. Add new artifacts, commit changes

   ```bash
   $ git add SPECKIT_NOTES.md specs/001-package-mode-support/tasks.md
   $ git commit
   ```

## Analyze the documents for completeness/correctness

1. 

   ```text
   /speckit.analyze
   ...
   ● Specification Analysis Report
   
     Feature: Package Mode Support (001-package-mode-support)
     Artifacts Analyzed: spec.md, plan.md, tasks.md
     Constitution Version: 1.1.0
     Analysis Date: 2026-01-22
   
     Executive Summary
   
     Status: ⚠ MEDIUM PRIORITY ISSUES FOUND
   
     The feature artifacts are generally well-aligned with strong user story
     organization and clear requirements. However, there are coverage gaps
     between the plan's proposed file modifications and the task list, plus
     one internal inconsistency about test inclusion. No critical constitution
     violations detected.
   
     ---
     Findings Table
     ┌─────┬───────────────┬──────────┬─────────────────────────┬─────────────────────────────────────────────────────────────────────────────────────────────────────────┬────────────────────────────────────────────────┐
     │ ID  │   Category    │ Severity │       Location(s)       │                                                 Summary                                                 │                 Recommendation                 │
     ├─────┼───────────────┼──────────┼─────────────────────────┼─────────────────────────────────────────────────────────────────────────────────────────────────────────┼────────────────────────────────────────────────┤
     │ C1  │ Coverage Gap  │ HIGH     │ plan.md:70, tasks.md    │ Plan specifies modifying internal/agent/device/bootstrap.go for conditional update paths, but NO task   │ Add task to modify bootstrap.go or remove from │
     │     │               │          │                         │ exists for this modification                                                                            │  plan if not needed                            │
     ├─────┼───────────────┼──────────┼─────────────────────────┼─────────────────────────────────────────────────────────────────────────────────────────────────────────┼────────────────────────────────────────────────┤
     │ C2  │ Coverage Gap  │ HIGH     │ plan.md:71-72, tasks.md │ Plan specifies modifying internal/agent/device/spec/ for update handling, but NO task exists            │ Add task for spec/ modifications or clarify in │
     │     │               │          │                         │                                                                                                         │  research.md why not needed                    │
     ├─────┼───────────────┼──────────┼─────────────────────────┼─────────────────────────────────────────────────────────────────────────────────────────────────────────┼────────────────────────────────────────────────┤
     │ C3  │ Coverage Gap  │ MEDIUM   │ plan.md:72, tasks.md    │ Plan specifies modifying internal/agent/config/ for mode configuration, but NO task exists              │ Add task or remove from plan structure         │
     ├─────┼───────────────┼──────────┼─────────────────────────┼─────────────────────────────────────────────────────────────────────────────────────────────────────────┼────────────────────────────────────────────────┤
     │ I1  │ Inconsistency │ MEDIUM   │ plan.md:74 vs           │ Plan says "ADD: PackageMode boolean field" to internal/store/model/device.go but data-model.md says "No │ Align plan and data-model - clarify if code    │
     │     │               │          │ data-model.md:73        │  changes required" since JSONB auto-persists                                                            │ change needed or just documentation            │
     ├─────┼───────────────┼──────────┼─────────────────────────┼─────────────────────────────────────────────────────────────────────────────────────────────────────────┼────────────────────────────────────────────────┤
     │ I2  │ Inconsistency │ LOW      │ tasks.md:10             │ Header states "No test tasks included (not explicitly requested)" but tasks T018-T020, T029-T030, T037, │ Update header to accurately reflect test task  │
     │     │               │          │                         │  T043-T052 ARE test tasks                                                                               │ inclusion                                      │
     ├─────┼───────────────┼──────────┼─────────────────────────┼─────────────────────────────────────────────────────────────────────────────────────────────────────────┼────────────────────────────────────────────────┤
     │ A1  │ Ambiguity     │ LOW      │ spec.md:84-86           │ Edge cases listed as questions without resolution status or handling approach                           │ Mark as resolved in research.md or add tasks   │
     │     │               │          │                         │                                                                                                         │ to handle edge cases                           │
     ├─────┼───────────────┼──────────┼─────────────────────────┼─────────────────────────────────────────────────────────────────────────────────────────────────────────┼────────────────────────────────────────────────┤
     │ C4  │ Coverage Gap  │ MEDIUM   │ FR-010, tasks.md        │ FR-010 requires "Automated tests MUST validate agent functionality on both RHEL and Ubuntu" -           │ Add comment linking integration test phase to  │
     │     │               │          │                         │ Integration tests exist (T041-T052) but no explicit FR-010 reference/traceability                       │ FR-010 for traceability                        │
     ├─────┼───────────────┼──────────┼─────────────────────────┼─────────────────────────────────────────────────────────────────────────────────────────────────────────┼────────────────────────────────────────────────┤
     │ C5  │ Coverage Gap  │ MEDIUM   │ FR-011, tasks.md        │ FR-011 requires documentation for both platforms - Tasks T013-T014 cover this but use [US1] label       │ Add FR-011 traceability comment or accept      │
     │     │               │          │                         │ instead of explicit FR-011 reference                                                                    │ user-story-based organization                  │
     ├─────┼───────────────┼──────────┼─────────────────────────┼─────────────────────────────────────────────────────────────────────────────────────────────────────────┼────────────────────────────────────────────────┤
     │ D1  │ Duplication   │ LOW      │ spec.md:95-96 vs        │ FR-004 "MUST report OS type (RHEL, Ubuntu)" duplicates entity definition "Package-Mode Device: Linux    │ Minor - context differs (requirement vs        │
     │     │               │          │ spec.md:107-108         │ system where OS managed via..."                                                                         │ definition), acceptable redundancy             │
     └─────┴───────────────┴──────────┴─────────────────────────┴─────────────────────────────────────────────────────────────────────────────────────────────────────────┴────────────────────────────────────────────────┘
     ---
     Coverage Summary Table
     ┌─────────────────┬─────────────────────────────────────────┬─────────────┬───────────────────────┬─────────────────────────────────────────────┐
     │ Requirement Key │               Description               │  Has Task?  │       Task IDs        │                    Notes                    │
     ├─────────────────┼─────────────────────────────────────────┼─────────────┼───────────────────────┼─────────────────────────────────────────────┤
     │ FR-001          │ Agent install on RHEL 9+ via RPM        │ ✅ Yes      │ T013, T045            │ US1 installation docs + integration test    │
     ├─────────────────┼─────────────────────────────────────────┼─────────────┼───────────────────────┼─────────────────────────────────────────────┤
     │ FR-002          │ Agent install on Ubuntu 22.04+ via .deb │ ✅ Yes      │ T014, T049            │ US1 installation docs + integration test    │
     ├─────────────────┼─────────────────────────────────────────┼─────────────┼───────────────────────┼─────────────────────────────────────────────┤
     │ FR-003          │ Auto-detect package-mode vs image-mode  │ ✅ Yes      │ T004-T010, T018-T020  │ Foundational detection + US2 tests          │
     ├─────────────────┼─────────────────────────────────────────┼─────────────┼───────────────────────┼─────────────────────────────────────────────┤
     │ FR-004          │ Report OS type and deployment mode      │ ✅ Yes      │ T010, T023            │ Status population + verification            │
     ├─────────────────┼─────────────────────────────────────────┼─────────────┼───────────────────────┼─────────────────────────────────────────────┤
     │ FR-005          │ Apply Flight Control config updates     │ ✅ Yes      │ T033, T047, T051      │ Manual + integration tests                  │
     ├─────────────────┼─────────────────────────────────────────┼─────────────┼───────────────────────┼─────────────────────────────────────────────┤
     │ FR-006          │ Apply Flight Control agent updates      │ ⚠ Implicit │ T033                  │ Covered in config update test, not explicit │
     ├─────────────────┼─────────────────────────────────────────┼─────────────┼───────────────────────┼─────────────────────────────────────────────┤
     │ FR-007          │ MUST NOT manage OS-level updates        │ ✅ Yes      │ T025-T028, T048, T052 │ Core selective update logic                 │
     ├─────────────────┼─────────────────────────────────────────┼─────────────┼───────────────────────┼─────────────────────────────────────────────┤
     │ FR-008          │ MUST NOT interfere with dnf/apt         │ ✅ Yes      │ T034                  │ Manual concurrency test                     │
     ├─────────────────┼─────────────────────────────────────────┼─────────────┼───────────────────────┼─────────────────────────────────────────────┤
     │ FR-009          │ Console UI displays deployment mode     │ ✅ Yes      │ T035-T040             │ US4 complete coverage                       │
     ├─────────────────┼─────────────────────────────────────────┼─────────────┼───────────────────────┼─────────────────────────────────────────────┤
     │ FR-010          │ Automated tests for RHEL/Ubuntu         │ ✅ Yes      │ T041-T054             │ Integration testing phase                   │
     ├─────────────────┼─────────────────────────────────────────┼─────────────┼───────────────────────┼─────────────────────────────────────────────┤
     │ FR-011          │ Installation/config documentation       │ ✅ Yes      │ T013-T014, T016       │ US1 documentation tasks                     │
     ├─────────────────┼─────────────────────────────────────────┼─────────────┼───────────────────────┼─────────────────────────────────────────────┤
     │ FR-012          │ No regression in image-mode             │ ✅ Yes      │ T063                  │ Explicit regression check                   │
     ├─────────────────┼─────────────────────────────────────────┼─────────────┼───────────────────────┼─────────────────────────────────────────────┤
     │ SC-001          │ 100% install success rate               │ ✅ Yes      │ T045, T049            │ Integration tests                           │
     ├─────────────────┼─────────────────────────────────────────┼─────────────┼───────────────────────┼─────────────────────────────────────────────┤
     │ SC-002          │ 100% correct detection                  │ ✅ Yes      │ T021, T046, T050      │ Unit + integration                          │
     ├─────────────────┼─────────────────────────────────────────┼─────────────┼───────────────────────┼─────────────────────────────────────────────┤
     │ SC-003          │ Config updates <5s latency              │ ⚠ Implicit │ T033                  │ Not explicitly measuring latency            │
     ├─────────────────┼─────────────────────────────────────────┼─────────────┼───────────────────────┼─────────────────────────────────────────────┤
     │ SC-004          │ Zero conflicts with package mgr         │ ✅ Yes      │ T034                  │ Explicit concurrent test                    │
     ├─────────────────┼─────────────────────────────────────────┼─────────────┼───────────────────────┼─────────────────────────────────────────────┤
     │ SC-005          │ 100% image-mode tests pass              │ ✅ Yes      │ T063                  │ Regression check                            │
     ├─────────────────┼─────────────────────────────────────────┼─────────────┼───────────────────────┼─────────────────────────────────────────────┤
     │ SC-006          │ Console UI 100% accurate                │ ✅ Yes      │ T039                  │ Manual verification                         │
     ├─────────────────┼─────────────────────────────────────────┼─────────────┼───────────────────────┼─────────────────────────────────────────────┤
     │ SC-007          │ Documentation quality                   │ ⚠ Implicit │ T013-T014             │ Docs exist, quality not measured            │
     └─────────────────┴─────────────────────────────────────────┴─────────────┴───────────────────────┴─────────────────────────────────────────────┘
     Coverage Metrics:
     - Total Functional Requirements: 12
     - Total Success Criteria: 7
     - Total Tasks: 64
     - Requirements with >=1 task: 12/12 (100%)
     - Success Criteria with >=1 task: 7/7 (100%)
     - Explicit Traceability: Moderate (user-story-based organization reduces direct FR→Task mapping visibility)
   
     ---
     Constitution Alignment Issues
   
     ✅ NO CRITICAL CONSTITUTION VIOLATIONS DETECTED
   
     All constitutional principles are addressed:
     ┌─────────────────────────────────┬─────────┬────────────────────────────────────────────────────────────────────────────────────────────────────────┐
     │            Principle            │ Status  │                                                Evidence                                                │
     ├─────────────────────────────────┼─────────┼────────────────────────────────────────────────────────────────────────────────────────────────────────┤
     │ I. Code Quality & Simplicity    │ ✅ Pass │ Tasks include make lint checks (T011, T017, T024, T032, T054, T062). Simple bootc detection pattern.   │
     ├─────────────────────────────────┼─────────┼────────────────────────────────────────────────────────────────────────────────────────────────────────┤
     │ II. Edge Device Management      │ ✅ Pass │ Extends declarative management, observable lifecycle maintained. FR-004 ensures observability.         │
     ├─────────────────────────────────┼─────────┼────────────────────────────────────────────────────────────────────────────────────────────────────────┤
     │ III. API Stability & Versioning │ ✅ Pass │ Backward compatible additionalProperties extension. No breaking changes. FR-012 ensures no regression. │
     ├─────────────────────────────────┼─────────┼────────────────────────────────────────────────────────────────────────────────────────────────────────┤
     │ IV. Test Coverage & Quality     │ ✅ Pass │ Unit tests (T018-T020, T029-T030), integration tests (T041-T054), regression tests (T063).             │
     ├─────────────────────────────────┼─────────┼────────────────────────────────────────────────────────────────────────────────────────────────────────┤
     │ V. Security                     │ ✅ Pass │ Path validation implied in detection logic. No new auth boundaries.                                    │
     ├─────────────────────────────────┼─────────┼────────────────────────────────────────────────────────────────────────────────────────────────────────┤
     │ VI. Documentation Conciseness   │ ✅ Pass │ Spec/plan readable in single session. Documentation tasks (T013-T014) exist.                           │
     ├─────────────────────────────────┼─────────┼────────────────────────────────────────────────────────────────────────────────────────────────────────┤
     │ VII. Observability              │ ✅ Pass │ Logging tasks (T009, T026, T028), metrics tasks (T055-T058), status reporting (T010).                  │
     ├─────────────────────────────────┼─────────┼────────────────────────────────────────────────────────────────────────────────────────────────────────┤
     │ VIII. Performance & Scale       │ ✅ Pass │ <1ms detection overhead specified. SC-003 requires <5s latency.                                        │
     └─────────────────────────────────┴─────────┴────────────────────────────────────────────────────────────────────────────────────────────────────────┘
     Quality Gates Compliance:
     - ✅ Linting: Tasks T011, T017, T024, T032, T054, T062 enforce make lint
     - ✅ Testing: Tasks T012, T021, T031, T038, T053, T061 enforce test execution
     - ✅ Interface Changes: Tasks T031 includes go test -v -race after device.go modifications
     - ⚠ Security Review: No explicit security review task, but changes are low-risk (detection logic only)
   
     ---
     Unmapped Tasks
   
     None detected. All 64 tasks map to either:
     - User stories (US1-US4)
     - Functional requirements (FR-001 through FR-012)
     - Success criteria validation (SC-001 through SC-007)
     - Infrastructure/polish activities
   
     ---
     Metrics
   
     - Total Functional Requirements: 12
     - Total Success Criteria: 7
     - Total User Stories: 4
     - Total Tasks: 64
     - Requirements Coverage: 100% (12/12 have >=1 task)
     - Success Criteria Coverage: 100% (7/7 have >=1 task)
     - Ambiguity Count: 1 (edge cases not resolved)
     - Duplication Count: 1 (minor acceptable duplication)
     - Critical Issues: 0
     - High Issues: 3 (coverage gaps C1, C2, C3)
     - Medium Issues: 4 (inconsistencies, gaps)
     - Low Issues: 3 (minor ambiguity, duplication)
   
     ---
     Detailed Analysis
   
     High-Priority Findings (Require Resolution)
   
     C1: Missing bootstrap.go Modification Task
   
     The plan.md structure clearly lists internal/agent/device/bootstrap.go # MODIFY:
     conditional update paths but no task exists. The research.md (section 6)
     identifies modification points including bootstrap.go for ensureBootedOS()
    and checkRollback(). However, tasks.md only modifies device.go (T025-T028).
   
     Recommendation: Either add tasks to modify bootstrap.go OR clarify in
     research.md that the device.go modifications in beforeUpdate/afterUpdate
     are sufficient and bootstrap.go changes are not needed.
   
     C2: Missing spec/ Directory Modification Task
   
     Plan indicates internal/agent/device/spec/ # MODIFY: update handling but no
     corresponding task. Research.md discusses spec manager's IsOSUpdate() and
     CheckOsReconciliation() but doesn't propose changes - it uses existing
     methods.
   
     Recommendation: Remove from plan structure OR add explicit task if spec
     manager needs modification.
   
     C3: Missing config/ Directory Modification Task
   
     Plan lists internal/agent/config/ # MODIFY: mode
     configuration with no task coverage.
   
     Recommendation: Clarify if agent config needs modification for package-mode
     support. If not, remove from plan.
   
     Medium-Priority Findings
   
     I1: Plan vs Data-Model Inconsistency on device.go
   
     - Plan.md line 74: internal/store/model/device.go # ADD: PackageMode
       boolean field
     - Data-model.md line 73: "No changes required. Device model stores raw JSON
       in Status field"
   
     The data-model is correct - JSONB auto-persists additionalProperties. The
     plan structure may mislead developers.
   
     Recommendation: Update plan.md comment to clarify: # NO CHANGE NEEDED:
     PackageMode stored in Status JSONB additionalProperties
   
     ---
     Next Actions
   
     If Proceeding with Current State
   
     The feature can proceed to implementation with these caveats:
   
     1. Clarify file modification scope: Resolve C1, C2, C3 before starting
        implementation to avoid confusion
     2. Update tasks.md header: Fix "No test tasks" statement (I2)
     3. Document edge case resolution: Mark edge cases as deferred or add
        handling tasks
   
     Recommended Before Implementation
   
     1. Resolve HIGH issues (C1, C2, C3): Add missing tasks or update plan.md
        to remove incorrect file references
     2. Fix inconsistency I1: Align plan.md and data-model.md on device.go
     3. Update tasks.md header: Correct test task statement
     4. Optional: Add explicit FR/SC references in task comments for better
        traceability
   
     Commands to Run
   
     - To update plan.md: Manually edit to remove/clarify bootstrap.go, spec/,
       config/ references
     - To update tasks.md: Add tasks for missing file modifications OR update
       header statement
     - To align data-model: Update plan.md structure comments for device.go
   
     ---
     Remediation Offer
   
     Would you like me to suggest concrete remediation edits for the top 3
     HIGH-priority issues (C1, C2, C3) and the plan/data-model inconsistency (I1)?
   
     I can propose specific changes to:
     1. Add missing tasks to tasks.md for bootstrap.go, spec/, config/
     2. OR update plan.md to clarify which files actually need modification
     3. Align plan.md and data-model.md on device.go handling
   
     Note: This analysis is READ-ONLY. I will not apply changes without your
     explicit approval.
   ```

   I prompted it with `yes, suggest the remediation`

   ```text
   ● Remediation Plan
   
     Based on the analysis, the cleanest approach is to update plan.md to align
     with the actual implementation approach (rather than adding unnecessary
     tasks). The research.md shows that fewer file modifications are needed than
     initially planned.
   
     Proposed Changes
   
     ---
     Change 1: Update plan.md Structure - Remove Unnecessary File References
   
     File: /home/andalton/src/speckit.EDM1471/flightctl/specs/001-package-mode-support/plan.md
   
     Current (lines 63-91):
     flightctl/
     ├── internal/agent/                    # Agent implementation (MODIFY)
     │   ├── device/
     │   │   ├── systeminfo/                # ADD: package-mode detection
     │   │   │   ├── packagemode.go         # NEW: bootc presence check
     │   │   │   └── packagemode_test.go    # NEW: detection tests
     │   │   ├── bootstrap.go               # MODIFY: conditional update paths
     │   │   └── spec/                      # MODIFY: update handling
     │   └── config/                        # MODIFY: mode configuration
     ├── internal/store/model/              # Data models (MODIFY)
     │   └── device.go                      # ADD: PackageMode boolean field
     ├── internal/api/                      # API handlers (MODIFY)
     │   └── device.go                      # MODIFY: return package-mode in status
   
     Proposed:
     flightctl/
     ├── internal/agent/                    # Agent implementation (MODIFY)
     │   ├── device/
     │   │   ├── systeminfo/                # ADD: package-mode detection
     │   │   │   ├── packagemode.go         # NEW: bootc presence check
     │   │   │   └── packagemode_test.go    # NEW: detection tests
     │   │   ├── device.go                  # MODIFY: beforeUpdate(), afterUpdate() skip OS ops in package-mode
     │   │   └── device_test.go             # NEW: unit tests for package-mode update logic
     ├── internal/store/model/              # Data models (NO CODE CHANGE)
     │   └── device.go                      # Status JSONB field auto-persists additionalProperties["packageMode"]
     ├── internal/api/                      # API handlers (VERIFY - likely no change needed)
     │   └── device.go                      # Status endpoint already returns full JSONB, includes packageMode
   
     Rationale:
     - bootstrap.go: Research shows it uses existing specManager.IsOSUpdate()
       and specManager.CheckOsReconciliation() without modification
     - spec/: Existing methods (IsOSUpdate, CheckOsReconciliation) are used
       as-is, no changes needed
     - config/: No config changes identified in research or tasks
     - device.go (model): Data-model.md correctly states JSONB auto-persists,
       no code change needed
     - device.go (api): Likely no change needed since API already returns full
       status JSONB
   
     ---
     Change 2: Update plan.md Structure Decision Text
   
     File: /home/andalton/src/speckit.EDM1471/flightctl/specs/001-package-mode-support/plan.md
   
     Current (line 93):
     **Structure Decision**: Monorepo modification. Extends existing agent
     systeminfo subsystem for mode detection. Adds package-mode boolean to
      device model (backward compatible database migration). No new services or
     binaries. Test infrastructure adds RHEL/Ubuntu runners to existing
       integration test framework.
   
     Proposed:
     **Structure Decision**: Monorepo modification. Extends existing agent
     systeminfo subsystem for mode detection. Package-mode status stored in
     existing DeviceSystemInfo.additionalProperties (JSONB, no schema change).
     Modifies agent device.go update workflow to skip OS operations when
     package-mode detected. No new services or binaries. Test infrastructure
     adds RHEL/Ubuntu runners to existing integration test framework.
   
     Rationale: Clarifies that storage uses existing JSONB extensibility,
     no schema/model changes required.
   
     ---
     Change 3: Update tasks.md Header Statement
   
     File: /home/andalton/src/speckit.EDM1471/flightctl/specs/001-package-mode-support/tasks.md
   
     Current (line 10):
     **Tests**: No test tasks included (not explicitly requested in specification)
   
     Proposed:
    
     **Tests**: Unit and integration test tasks included per Constitution 
     rinciple IV (Test Coverage & Quality)
   
     Rationale: Tasks T018-T020 (unit tests), T029-T030 (unit tests),
     T037 (API integration test), T043-T052 (integration tests) ARE present.
     Statement is factually incorrect.
   
     ---
     Change 4: Add Clarifying Comment to plan.md Research Reference
   
     File: /home/andalton/src/speckit.EDM1471/flightctl/specs/001-package-mode-support/plan.md
   
     Location: After line 16 (Summary section)
   
     Add:
   
     **Key Implementation Decision** (from research.md): Modification points
     limited to systeminfo detection (packagemode.go, manager.go) and agent
     update workflow (device.go beforeUpdate/afterUpdate). No changes required
     to bootstrap.go, spec manager, config, or data models - existing
     infrastructure supports package-mode via feature flags and JSONB
     extensibility.
   
     Rationale: Explicitly documents why fewer files are modified than initially
     anticipated, preventing confusion during implementation.
   
     ---
     Change 5: Optional - Add Task Cross-References
   
     File: /home/andalton/src/speckit.EDM1471/flightctl/specs/001-package-mode-support/tasks.md
   
     Location: After line 12 (before first phase)
   
     Add:
   
     **Functional Requirement Mapping**:
     - FR-001, FR-002: Tasks T013-T017, T045, T049 (US1 Installation)
     - FR-003, FR-004: Tasks T004-T012, T018-T024 (US2 Detection)
     - FR-005, FR-006, FR-007, FR-008: Tasks T025-T034 (US3 Update Management)
     - FR-009: Tasks T035-T040 (US4 Console Visibility)
     - FR-010: Tasks T041-T054 (Integration Testing)
     - FR-011: Tasks T013-T016 (Documentation)
     - FR-012: Task T063 (Regression Testing)
   
     Rationale: Improves traceability between requirements and tasks (addresses finding C4, C5).
   
     ---
     Summary of Remediation Actions
     ┌────────────┬─────────────┬──────────┬─────────────────────────────────────────────────────────────────┐
     │  Issue ID  │ Action Type │   File   │                         Change Summary                          │
     ├────────────┼─────────────┼──────────┼─────────────────────────────────────────────────────────────────┤
     │ C1, C2, C3 │ Remove      │ plan.md  │ Remove incorrect file references (bootstrap.go, spec/, config/) │
     ├────────────┼─────────────┼──────────┼─────────────────────────────────────────────────────────────────┤
     │ C1         │ Add         │ plan.md  │ Add device.go to structure (actual modification point)          │
     ├────────────┼─────────────┼──────────┼─────────────────────────────────────────────────────────────────┤
     │ I1         │ Clarify     │ plan.md  │ Change device.go (model) comment to "NO CODE CHANGE"            │
     ├────────────┼─────────────┼──────────┼─────────────────────────────────────────────────────────────────┤
     │ I2         │ Fix         │ tasks.md │ Correct "No test tasks" to acknowledge test inclusion           │
     ├────────────┼─────────────┼──────────┼─────────────────────────────────────────────────────────────────┤
     │ C4, C5     │ Enhance     │ tasks.md │ Add FR mapping section (optional, improves traceability)        │
     └────────────┴─────────────┴──────────┴─────────────────────────────────────────────────────────────────┘
     Total Changes: 3 mandatory edits (plan.md structure, plan.md decision, tasks.md header) + 2 optional enhancements
   
     Risk Level: LOW - All changes are documentation/clarification only, no
     impact on actual implementation
   
     Would you like me to proceed with applying these changes?
   ```

   I responded with `yes`

   ```text
   ● Remediation Verification Complete ✅
   
     All remediation changes have been successfully applied and verified:
   
     Fixed Issues
   
     HIGH Priority (3 issues resolved):
     - C1: Removed incorrect bootstrap.go reference from plan.md structure
     - C2: Removed incorrect spec/ directory reference from plan.md structure
     - C3: Removed incorrect config/ directory reference from plan.md structure
   
     MEDIUM Priority (4 issues resolved):
     - I1: Aligned plan.md and data-model.md on device.go handling (clarified JSONB auto-persists, no code change needed)
     - C4: Added functional requirement mapping to tasks.md for all 12 FRs
     - C5: Ensured all requirements traceable to specific task IDs
   
     LOW Priority (1 issue resolved):
     - I2: Fixed tasks.md header to acknowledge test task inclusion per Constitution Principle IV
   
     Current State
   
     All specification artifacts are now consistent and aligned:
   
     - spec.md: 4 user stories, 12 functional requirements, 7 success criteria
     - plan.md: Accurate file modification list (systeminfo + device.go only), clear implementation decision
     - tasks.md: 64 tasks with complete FR mapping, correct test acknowledgment
     - data-model.md: JSONB extension strategy documented
     - research.md: bootc detection mechanism defined
     - quickstart.md: Installation procedures for RHEL/Ubuntu
     - contracts/device-status.yaml: API schema extension defined
   
     The feature is ready for implementation following the task breakdown in tasks.md.
   ```

3. Add changed artifacts, commit changes

   ```bash
   $ git status
   On branch 001-package-mode-support
   Changes not staged for commit:
     (use "git add <file>..." to update what will be committed)
     (use "git restore <file>..." to discard changes in working directory)
           modified:   SPECKIT_NOTES.md
           modified:   specs/001-package-mode-support/plan.md
           modified:   specs/001-package-mode-support/tasks.md
   $ git add -u
   $ git commit
   ```
