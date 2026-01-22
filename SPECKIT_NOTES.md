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
