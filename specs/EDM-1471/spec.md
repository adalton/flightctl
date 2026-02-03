# Feature Specification: Non-Image-Mode Device Support

* **Feature Branch**: `EDM-1471`
* **Created**: 2026-02-03
* **Status**: Draft
* **Input**: User description: "Jira issue EDM-1471"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Agent Installation on Package-Based Systems (Priority: P1)

As a system administrator, I want to install the Flight Control agent on my RHEL or Ubuntu systems managed by traditional package managers (yum/dnf/apt), so that I can manage device configurations and applications through Flight Control without requiring bootc image-based deployments.

**Why this priority**: This is the foundational capability enabling Flight Control to work on non-image-mode systems. Without this, the feature cannot function at all. It represents the minimum viable product and delivers immediate value by expanding platform support.

**Independent Test**: Can be fully tested by installing the agent package on RHEL and Ubuntu hosts using standard package managers, verifying successful installation, service startup, and agent enrollment with Flight Control.

**Acceptance Scenarios**:

1. **Given** a RHEL 9 system managed by dnf, **When** the Flight Control agent RPM is installed, **Then** all dependencies are resolved, the package installs successfully, and the agent service starts without errors
2. **Given** an Ubuntu 22.04 system managed by apt, **When** the Flight Control agent is installed (via available installation method), **Then** the installation completes successfully and the agent service starts without errors
3. **Given** a successfully installed agent on either platform, **When** the agent is configured with Flight Control server details, **Then** the agent successfully enrolls and appears in the Flight Control inventory

---

### User Story 2 - Environment Detection and Appropriate Behavior (Priority: P2)

As a Flight Control user, I want the agent to automatically detect when it's running in a non-image-mode (package-based) environment and adapt its behavior accordingly, so that it operates correctly without causing errors or conflicts with the system package manager.

**Why this priority**: This ensures the agent doesn't misbehave or attempt image-based operations on package-based systems. It prevents operational issues and system instability, making it critical for production use.

**Independent Test**: Can be tested by deploying agents on both image-mode and non-image-mode systems, then verifying that the agent correctly identifies its environment type and reports this status through the Flight Control API and UI.

**Acceptance Scenarios**:

1. **Given** an agent running on a RHEL system without bootc, **When** the agent starts, **Then** it detects the environment as non-image-mode and reports this status
2. **Given** an agent running on an Ubuntu system, **When** the agent starts, **Then** it detects the environment as non-image-mode and reports the OS type as Ubuntu
3. **Given** an agent running on a bootc-managed system, **When** the agent starts, **Then** it detects the environment as image-mode and continues existing behavior unchanged
4. **Given** an agent in non-image-mode, **When** queried for device status, **Then** the agent reports "non-image-mode" or "package-mode" status through the API

---

### User Story 3 - Configuration and Application Update Management (Priority: P3)

As a system administrator, I want the agent to apply Flight Control-managed configuration and application updates while automatically ignoring OS-level package updates, so that I can manage applications through Flight Control while maintaining control over OS updates through traditional package managers.

**Why this priority**: This enables the core value proposition - managing configurations and applications via Flight Control. While important, it depends on P1 and P2 being complete. It can be tested independently once the agent is installed and environment detection works.

**Independent Test**: Can be tested by creating Flight Control configurations targeting non-image-mode devices and verifying that configuration updates apply successfully while OS package updates are ignored.

**Acceptance Scenarios**:

1. **Given** a non-image-mode device enrolled in Flight Control, **When** a configuration update is deployed, **Then** the agent applies the configuration changes successfully without interfering with system packages
2. **Given** a non-image-mode device with pending OS updates (from yum/apt), **When** the agent checks for updates, **Then** the agent ignores OS-level package updates and only reports Flight Control-managed updates
3. **Given** a non-image-mode device, **When** a Flight Control application update is available, **Then** the agent applies the application update successfully
4. **Given** a non-image-mode device where apt/dnf is actively managing packages, **When** the agent is performing operations, **Then** there are no conflicts or errors from concurrent package management operations

---

### User Story 4 - Console UI Visibility of Deployment Mode (Priority: P4)

As a Flight Control operator, I want to see which devices are running in image-mode vs package-mode in the console UI, so that I can understand the deployment model of each device and troubleshoot accordingly.

**Why this priority**: This is a UI enhancement that improves operational visibility. It's valuable but not critical for core functionality. It can be developed and tested independently once the agent properly reports mode status.

**Independent Test**: Can be tested by viewing the device inventory in the console UI and verifying that the deployment mode (image-mode/package-mode) is clearly displayed for each device.

**Acceptance Scenarios**:

1. **Given** devices enrolled in Flight Control with mixed deployment types, **When** viewing the device inventory, **Then** each device shows its deployment mode (image-mode or package-mode)
2. **Given** a non-image-mode Ubuntu device, **When** viewing device details, **Then** the UI displays the OS type (Ubuntu) and deployment mode (package-mode)
3. **Given** a non-image-mode RHEL device, **When** viewing device details, **Then** the UI displays the OS type (RHEL) and deployment mode (package-mode)

---

### Edge Cases

- What happens when the agent cannot reliably detect whether it's in image-mode or non-image-mode? (e.g., ambiguous system state)
- How does the system handle attempted OS updates through Flight Control when in non-image-mode? (Should gracefully reject or show appropriate error)
- What happens when a device transitions from package-mode to image-mode or vice versa?
- How does the agent behave during package manager lock conditions (yum/apt already running)?
- What happens if Flight Control attempts to deploy an image-mode-only feature to a package-mode device?
- How does the agent handle partial installations or corrupted agent packages?
- What happens when the agent runs on an unsupported Linux distribution (neither RHEL nor Ubuntu)?

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Agent MUST successfully install on RHEL systems using standard RPM package installation methods
- **FR-002**: Agent MUST successfully install on Ubuntu systems using available installation methods
- **FR-003**: Agent MUST automatically detect whether it's running in image-mode (bootc-managed) or non-image-mode (package-managed) environment
- **FR-004**: Agent MUST correctly identify and report the underlying OS type (RHEL or Ubuntu) when in non-image-mode
- **FR-005**: Agent MUST report its deployment mode (image-mode or package-mode) through the Flight Control API
- **FR-006**: Agent MUST apply configuration updates from Flight Control on non-image-mode systems without errors
- **FR-007**: Agent MUST apply Flight Control-managed application updates on non-image-mode systems
- **FR-008**: Agent MUST NOT attempt to manage OS-level package updates (from yum/dnf/apt) when in non-image-mode
- **FR-009**: Agent MUST NOT interfere with system package manager operations (yum/dnf/apt) when performing Flight Control operations
- **FR-010**: Agent MUST continue to function correctly in image-mode environments without regression (existing functionality preserved)
- **FR-011**: Flight Control UI MUST display deployment mode (image-mode or package-mode) for each device in the inventory
- **FR-012**: Flight Control UI MUST display OS type for non-image-mode devices
- **FR-013**: Agent installation MUST satisfy all required dependencies on both RHEL and Ubuntu platforms
- **FR-014**: Agent service MUST start successfully after installation on both RHEL and Ubuntu systems
- **FR-015**: Agent MUST handle scenarios where it cannot determine deployment mode by reporting an error state rather than making incorrect assumptions

### Key Entities

- **Device**: Represents a managed host running the Flight Control agent; has attributes including OS type, deployment mode (image/package), current configuration state, and enrollment status
- **Deployment Mode**: An enumeration indicating whether a device is managed via image-mode (bootc) or package-mode (traditional package managers); determines agent behavior for updates
- **OS Type**: Identifies the Linux distribution (RHEL, Ubuntu, etc.); used for platform-specific operations and display purposes
- **Configuration**: Flight Control-managed settings and parameters that the agent applies to the device; independent of deployment mode
- **Application Update**: Updates to Flight Control-managed applications or the agent itself; applies to package-mode devices but excludes OS-level package updates

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Agent installs successfully on 100% of tested RHEL 9 systems using standard package installation procedures
- **SC-002**: Agent installs successfully on 100% of tested Ubuntu 22.04 LTS systems using available installation procedures
- **SC-003**: Agent correctly identifies deployment mode (image vs package) with 100% accuracy across test environments
- **SC-004**: Configuration updates complete successfully on non-image-mode devices within the same timeframe as image-mode devices (no performance degradation)
- **SC-005**: Zero conflicts or errors occur between agent operations and concurrent system package manager operations
- **SC-006**: 100% of existing image-mode automated tests continue to pass (no regressions)
- **SC-007**: Operators can identify device deployment mode within the console UI in under 5 seconds per device
- **SC-008**: Agent successfully ignores OS-level package updates 100% of the time when in non-image-mode
- **SC-009**: Flight Control-managed application updates complete successfully on 100% of non-image-mode test devices
- **SC-010**: Documentation enables a system administrator to install and configure the agent on RHEL or Ubuntu without engineering support

## Assumptions *(mandatory)*

- RHEL refers to RHEL 9 or later versions; earlier versions are out of scope
- Ubuntu refers to Ubuntu 22.04 LTS or later; earlier versions are out of scope
- Agent packages (RPM for RHEL) are available through standard distribution channels or Flight Control repositories
- Ubuntu installation may use alternative methods if .deb packages are not yet available (documented in Out of Scope)
- "Configuration updates" refer to Flight Control-managed configuration files, not system-level configurations managed by the OS
- "Application updates" refer to Flight Control agent updates and Flight Control-managed applications, not arbitrary system packages
- Package managers (yum/dnf/apt) are functioning correctly and the system is in a healthy state
- Network connectivity exists between devices and the Flight Control server for enrollment and updates
- The bootc command or absence thereof can reliably indicate whether a system is image-mode or package-mode
- Existing RHEL test infrastructure includes or will include non-image-mode (traditional RHEL installation) test runners

## Dependencies *(include if applicable)*

- E2E test infrastructure must support RHEL non-image-mode test runners (traditional RHEL installations, not bootc images)
- Agent packaging must produce installable RPM files compatible with RHEL 9+ package management
- Flight Control API must support reporting and querying deployment mode status
- Console UI must support displaying deployment mode and OS type fields
- Documentation team availability for creating installation and configuration guides

## Security & Privacy Considerations *(include if applicable)*

- Agent installation must follow standard Linux package security practices (signature verification for RPMs)
- Agent must not escalate privileges beyond what is necessary for configuration management
- Configuration updates must be applied with appropriate file permissions to prevent unauthorized access
- Agent logs must not expose sensitive configuration data in plain text
- Communication between agent and Flight Control server must use encrypted channels (existing security model applies)

## Out of Scope

- OS-level update management: OS package updates remain managed exclusively by system package managers (yum/dnf/apt), not by Flight Control
- Creating .deb packages for Ubuntu: Ubuntu installation may use alternative methods in this phase
- Support for Linux distributions beyond RHEL and Ubuntu: Other distributions are not part of this feature
- Changes to image-mode functionality: Existing bootc image-based deployment behavior remains unchanged
- Backward compatibility breaking changes: Existing image-mode deployments must continue to work without modification
- Automatic migration of image-mode devices to package-mode or vice versa
- Managing non-Flight Control applications or system services
- Custom Linux distributions or embedded systems beyond RHEL and Ubuntu
