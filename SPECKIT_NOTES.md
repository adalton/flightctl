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


