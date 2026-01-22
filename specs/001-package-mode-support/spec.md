# Feature Specification: Package Mode Support

**Feature Branch**: `001-package-mode-support`
**Created**: 2026-01-22
**Status**: Draft
**Input**: EDM-1471 - Support Flight Control Agent on Non-Image-Mode Devices

---

**📋 Constitution VI - Documentation Conciseness**: This spec MUST capture necessary information without verbosity. Target: readable and reviewable in a single session. Focus on "why" (rationale, decisions) over "what" (which should be clear from context). Avoid redundant explanations.

---

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Agent Installation on Package-Based Systems (Priority: P1)

System administrators need to install the Flight Control agent on traditional Linux distributions (RHEL, Ubuntu) where the OS is managed through package managers rather than bootc image updates.

**Why this priority**: Enables Flight Control adoption on existing infrastructure without requiring migration to image-based deployments. Critical for customers with established package-managed environments.

**Independent Test**: Install agent via RPM on RHEL system and via .deb on Ubuntu system. Agent starts successfully and enrolls with Flight Control service.

**Acceptance Scenarios**:

1. **Given** a RHEL 9 system with dnf package manager, **When** administrator installs flightctl-agent RPM, **Then** installation completes without errors and all dependencies are satisfied
2. **Given** an Ubuntu 22.04 system with apt package manager, **When** administrator installs flightctl-agent .deb package, **Then** installation completes without errors and systemd service is enabled
3. **Given** agent installation completed, **When** agent service starts, **Then** agent successfully connects to Flight Control service and reports device status

---

### User Story 2 - Automatic Environment Detection (Priority: P1)

Agent must automatically detect when running in package-mode (non-bootc) versus image-mode environment and adapt behavior accordingly to prevent conflicts with system package managers.

**Why this priority**: Core requirement for correct operation. Without proper detection, agent could interfere with OS package management or generate errors.

**Independent Test**: Deploy agent on RHEL package-mode system. Agent status reports OS type and package-mode indicator. No error logs related to bootc operations.

**Acceptance Scenarios**:

1. **Given** agent running on package-mode RHEL system, **When** agent initializes, **Then** agent detects package-mode environment and sets internal operating mode accordingly
2. **Given** agent running on package-mode Ubuntu system, **When** agent reports status, **Then** status includes OS type (Ubuntu) and package-mode indicator
3. **Given** agent running in package-mode, **When** system administrator runs package manager operations (dnf update, apt upgrade), **Then** agent does not interfere or generate errors

---

### User Story 3 - Selective Update Management (Priority: P1)

Agent must apply Flight Control managed configuration and application updates while ignoring OS-level package updates, maintaining separation of concerns between Flight Control and system package managers.

**Why this priority**: Essential functional requirement. Agent must know what to manage and what to ignore.

**Independent Test**: Push configuration update via Flight Control to package-mode device. Configuration applies successfully. Run OS package update via dnf/apt. OS update completes independently without agent interference.

**Acceptance Scenarios**:

1. **Given** Flight Control configuration update pushed to package-mode device, **When** agent receives update, **Then** agent applies configuration changes successfully
2. **Given** Flight Control agent binary update available, **When** update is pushed to package-mode device, **Then** agent self-updates successfully
3. **Given** OS packages require updates, **When** system administrator runs dnf/apt update, **Then** agent ignores OS-level changes and continues normal operation
4. **Given** simultaneous Flight Control config update and OS package update, **When** both operations execute, **Then** no conflicts occur and both complete successfully

---

### User Story 4 - Console Visibility (Priority: P2)

Flight Control operators need to see which devices are running in package-mode versus image-mode within the console UI for operational awareness and troubleshooting.

**Why this priority**: Important for operations but not blocking core functionality. Enhances user experience.

**Independent Test**: View device inventory in Flight Control console. Package-mode devices display OS type and mode indicator distinct from image-mode devices.

**Acceptance Scenarios**:

1. **Given** multiple devices enrolled (mix of package-mode and image-mode), **When** operator views device inventory, **Then** each device clearly indicates whether it is package-mode or image-mode
2. **Given** device running RHEL package-mode, **When** operator views device details, **Then** OS type shows "RHEL" and mode shows "package-mode"
3. **Given** operator troubleshooting device issue, **When** reviewing device properties, **Then** deployment mode is visible to inform troubleshooting approach

---

### Edge Cases

- What happens when bootc is installed later on a package-mode system?
- How does agent handle transition from package-mode to image-mode?
- What if agent package dependencies conflict with existing system packages?
- How does agent behave if OS package manager is locked during configuration update?

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Agent MUST successfully install on RHEL 9+ systems via RPM package
- **FR-002**: Agent MUST successfully install on Ubuntu 22.04+ systems via .deb package
- **FR-003**: Agent MUST automatically detect package-mode vs image-mode environment at startup
- **FR-004**: Agent MUST report OS type (RHEL, Ubuntu) and deployment mode (package-mode, image-mode) in status
- **FR-005**: Agent MUST apply Flight Control configuration updates on package-mode devices
- **FR-006**: Agent MUST apply Flight Control agent binary updates on package-mode devices
- **FR-007**: Agent MUST NOT attempt to manage OS-level package updates on package-mode devices
- **FR-008**: Agent MUST NOT interfere with system package manager operations (dnf, apt)
- **FR-009**: Console UI MUST display deployment mode (package-mode vs image-mode) for each device
- **FR-010**: Automated tests MUST validate agent functionality on both RHEL and Ubuntu package-mode systems
- **FR-011**: Installation and configuration documentation MUST exist for package-mode deployments on both platforms
- **FR-012**: Existing image-mode functionality MUST continue to work without regression

### Key Entities

- **Package-Mode Device**: Linux system where OS is managed via traditional package managers (dnf/yum, apt) rather than bootc image updates. Agent runs on these devices but only manages Flight Control configurations and agent updates.
- **Image-Mode Device**: Linux system where OS is managed via bootc image updates. Agent manages both OS images and configurations.
- **Environment Detection**: Agent capability to identify which mode it's operating in based on presence/absence of bootc tooling and system configuration.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Agent installs successfully on 100% of RHEL 9+ and Ubuntu 22.04+ package-mode test systems
- **SC-002**: Agent correctly detects and reports deployment mode in 100% of test cases (package-mode vs image-mode)
- **SC-003**: Configuration updates apply successfully on package-mode devices with <5 second latency from push to application
- **SC-004**: Zero conflicts between agent operations and concurrent OS package manager operations across 100 test iterations
- **SC-005**: 100% of existing image-mode automated tests continue to pass (no regression)
- **SC-006**: Console UI displays correct deployment mode for 100% of enrolled devices
- **SC-007**: Documentation enables users to install and configure package-mode agents without support escalation (measured by support ticket reduction)

### Constitution Alignment

Per Flight Control Constitution v1.1.0, this feature MUST address:

- **Edge Device Management** (Principle II): Extends declarative management to package-based deployments. Device lifecycle remains observable and auditable. Handles mixed fleet of package-mode and image-mode devices.
- **API Stability** (Principle III): No breaking API changes. Adds package-mode detection to status reporting. Existing device APIs unchanged.
- **Security** (Principle V): Package-mode detection logic must not introduce vulnerabilities. Agent must validate environment detection to prevent mode confusion attacks.
- **Performance & Scale** (Principle VIII): Environment detection adds negligible overhead (<1ms). Supports heterogeneous fleets mixing thousands of package-mode and image-mode devices.
- **Observability** (Principle VII): Package-mode indicator visible in device status, logs, and console UI. Clear separation of managed vs unmanaged updates logged.

## Assumptions

1. RHEL 9+ and Ubuntu 22.04+ are the target platforms (no support for older versions or other distributions in initial release)
2. Agent packages (.deb for Ubuntu) will be made available through standard distribution channels
3. Package-mode systems have systemd available for service management
4. Network connectivity requirements are identical for package-mode and image-mode devices
5. E2E test infrastructure will include RHEL runners (not currently available, will be added as part of this feature)
6. Agent self-update mechanism in package-mode uses same update distribution as image-mode (Flight Control service pushes updates)
7. "Non-interference" with package managers means agent does not invoke dnf/apt or lock package databases

## Out of Scope

- OS-level update management in package-mode (remains responsibility of system administrators using dnf/apt)
- Support for Linux distributions other than RHEL and Ubuntu
- Automatic migration from package-mode to image-mode
- Creation of .deb packages for agent (assumed to be available)
- Backward compatibility breaking changes to image-mode operations
