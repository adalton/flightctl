# Research: Package Mode Support

**Date**: 2026-01-22
**Feature**: 001-package-mode-support
**Purpose**: Research bootc detection mechanisms and agent update workflow to inform package-mode implementation

## 1. Bootc Detection Strategy

### Decision

Use existing `isBinaryAvailable()` pattern from `internal/agent/device/os/client.go:27-30`.

**Implementation**:
```go
func isBinaryAvailable(binaryName string) bool {
    _, err := exec.LookPath(binaryName)
    return err == nil
}
```

Check for bootc binary at agent startup. If absent, device operates in package-mode.

### Rationale

- Already proven pattern in codebase (`/internal/agent/device/os/client.go:27`)
- No filesystem dependencies (uses PATH lookup)
- Fast (<1ms) single syscall
- Platform-agnostic (works on RHEL, Ubuntu, any Linux)
- No new dependencies required

### Alternatives Considered

1. **Check `/usr/bin/bootc` directly**: Rejected. Hardcoded path fragile (bootc could be in `/usr/local/bin`, custom locations).
2. **Parse `/etc/os-release` for bootc indicator**: Rejected. No standard bootc identifier in os-release spec.
3. **Check for bootc package via RPM/dpkg**: Rejected. Requires distribution-specific package managers, adds complexity.
4. **Inspect filesystem for bootc artifacts**: Rejected. Implementation details may change across bootc versions.

## 2. Update Path Analysis

### Current OS Update Workflow

**File**: `internal/agent/device/device.go:315-482`

**Phases**:
1. **Before Update** (line 315-386):
   - Conditional: `if a.specManager.IsOSUpdate()` (line 340)
   - Action: Register OS manager as OCI collector → prefetch OS image
2. **Sync Device** (line 388-425):
   - No OS operations (apps, config, hooks, systemd)
3. **After Update** (line 440-482):
   - Conditional: `if !isOSReconciled && a.specManager.IsOSUpdate()` (line 460)
   - Action: `osManager.AfterUpdate()` → `bootc switch`, `osManager.Reboot()` → `bootc upgrade --apply`

### Key Conditionals

**`IsOSUpdate()`** (`internal/agent/device/spec/manager.go:474`):
```go
func (s *manager) IsOSUpdate() bool {
    return s.cache.getOSVersion(Current) != s.cache.getOSVersion(Desired)
}
```

Compares spec versions. Returns true if desired OS image differs from current.

**`CheckOsReconciliation()`** (`internal/agent/device/spec/manager.go:478`):
```go
func (s *manager) CheckOsReconciliation(ctx context.Context) (string, bool, error) {
    osStatus, err := s.osClient.Status(ctx)  // bootc status --json
    bootedOSImage := osStatus.GetBootedImage()

    desired, _ := s.Read(Desired)
    if desired.Spec.Os == nil {
        return bootedOSImage, false, nil
    }

    return bootedOSImage, desired.Spec.Os.Image == osStatus.GetBootedImage(), nil
}
```

Compares booted image with desired image. Returns reconciled=true if match.

### Package-Mode Modification Strategy

**Decision**: Early exit in `beforeUpdate()` when package-mode detected and OS update present in spec.

**Rationale**:
- Single modification point (minimal code churn)
- Prevents OS image prefetch (saves bandwidth, storage)
- Agent logs informational message: "Package-mode device, ignoring OS update"
- All other workflows (config, apps, hooks) continue normally

**Alternative Considered**: Skip OS operations in dummy client.
**Rejected**: Dummy client already used for rpm-ostree fallback. Mixing concerns. Package-mode should be explicit.

## 3. Dummy OS Client Behavior

### Current Implementation

**File**: `internal/agent/device/os/client.go:89-106`

When bootc not found, OS client factory returns `newDummyClient()`:
```go
func (d *dummy) Status(ctx context.Context) (*Status, error) {
    return &Status{container.BootcHost{}}, nil  // Empty status
}

func (d *dummy) Switch(ctx context.Context, image string) error {
    d.log.Warnf("Ignoring switch to image %s from dummy client")
    return nil  // No-op
}

func (d *dummy) Apply(ctx context.Context) error {
    d.log.Warnf("Ignoring apply from dummy client")
    return nil  // No-op
}
```

### Decision

Keep dummy client as fallback for unsupported OS managers (e.g., future Alpine, Arch).

**Package-mode distinction**: Package-mode is intentional deployment model (RHEL/Ubuntu with dnf/apt). Dummy client is unintentional (OS manager not detected).

**Implementation**:
- Add explicit package-mode detection (systeminfo collector)
- Log "Package-mode device detected" (INFO level) vs "Unsupported OS manager" (WARN level)
- Status field differentiates: `PackageMode: true` vs `PackageMode: false, UnsupportedOS: true`

### Rationale

Clear signal for operations: package-mode device is expected and supported. Unsupported OS triggers investigation.

## 4. Status Schema Extension

### Decision

Add `PackageMode` boolean to `DeviceSystemInfo` via `AdditionalProperties` (backward compatible).

**Field**: `AdditionalProperties["packageMode"] = "true"` or `"false"`

**File**: `api/core/v1beta1/openapi.yaml` (device status schema)

### Rationale

- No database migration required (`AdditionalProperties` is `map[string]string`)
- Backward compatible (older agents don't report field, defaults to false)
- API contract remains stable (existing fields unchanged)
- Console UI can filter/display based on property presence

### Alternatives Considered

1. **Add `DeploymentMode` enum field**: Rejected. Requires schema change, database migration, API version bump.
2. **Store in `CustomInfo`**: Rejected. `CustomInfo` for user-defined data, not system properties.
3. **Infer from OS image presence**: Rejected. Can't distinguish package-mode from failed bootc detection.

## 5. Detection Timing

### Decision

Detect package-mode at agent initialization in `systeminfo.Manager.Initialize()`.

**File**: `internal/agent/device/systeminfo/manager.go:68` (existing initialization pattern)

**Frequency**: Once per agent process lifetime.

### Rationale

- Boot ID check already runs at initialization (line 92-106)
- Package-mode unlikely to change during agent runtime (requires bootc installation + agent restart)
- Cached result available for all status reports (no repeated detection overhead)

### Alternatives Considered

1. **Detect on every status report**: Rejected. Unnecessary overhead (detection result doesn't change).
2. **Detect on demand when OS update triggered**: Rejected. Too late to prevent prefetch bandwidth usage.

## 6. Update Logic Modification Points

### Files to Modify

1. **`internal/agent/device/systeminfo/packagemode.go`** (NEW):
   - `func DetectPackageMode() bool`
   - Uses `isBinaryAvailable("bootc")`
   - Stores result in manager cache

2. **`internal/agent/device/systeminfo/manager.go`** (MODIFY):
   - Call `DetectPackageMode()` in `Initialize()` (line ~90)
   - Store result: `m.packageMode = DetectPackageMode()`
   - Expose via `func (m *manager) IsPackageMode() bool`

3. **`internal/agent/device/device.go`** (MODIFY line 340-342):
   ```go
   if a.specManager.IsOSUpdate() {
       if a.systemInfoManager.IsPackageMode() {
           a.log.Infof("Package-mode device, skipping OS image prefetch")
       } else {
           a.prefetchManager.RegisterOCICollector(a.osManager)
       }
   }
   ```

4. **`internal/agent/device/device.go`** (MODIFY line 460-468):
   ```go
   if !isOSReconciled && a.specManager.IsOSUpdate() {
       if a.systemInfoManager.IsPackageMode() {
           a.log.Infof("Package-mode device, ignoring OS update to %s", desired.Os.Image)
           // Continue to post-update hooks
       } else {
           if err = a.afterUpdateOS(ctx, desired); err != nil {
               return err
           }
           return nil  // OS updates return early
       }
   }
   ```

5. **`internal/agent/device/status/status.go`** (MODIFY exporter):
   - `systemInfoManager.Status()` adds `PackageMode` to `AdditionalProperties`

### Decision Rationale

- **Minimal modification points** (3 files, ~10 lines changed)
- **Preserves existing logic** for image-mode devices
- **Explicit conditionals** (readable, debuggable)
- **No architectural changes** (no new interfaces, no dependency injection changes)

## 7. Testing Strategy

### Unit Tests

**File**: `internal/agent/device/systeminfo/packagemode_test.go` (NEW)

**Test Cases**:
1. `TestDetectPackageMode_BootcPresent`: Mock `exec.LookPath("bootc")` returns nil → expect `false` (image-mode)
2. `TestDetectPackageMode_BootcAbsent`: Mock `exec.LookPath("bootc")` returns error → expect `true` (package-mode)
3. `TestSystemInfoManager_PackageModeInStatus`: Verify `AdditionalProperties["packageMode"]` set correctly

### Integration Tests

**Files**:
- `test/integration/agent/packagemode_rhel_test.go` (NEW)
- `test/integration/agent/packagemode_ubuntu_test.go` (NEW)

**Infrastructure**: RHEL 9 runner + Ubuntu 22.04 runner (GitHub Actions or equivalent)

**Test Scenarios**:
1. **Agent Install**: RPM/deb package installs successfully, agent starts
2. **Mode Detection**: Agent status reports `PackageMode: true`, no bootc warnings
3. **Config Update**: Push config change, agent applies successfully
4. **OS Update Ignored**: Push spec with OS image, agent logs "ignoring OS update", does not prefetch image
5. **No Interference**: Run `dnf update` / `apt upgrade` concurrently with Flight Control operations, no conflicts

### E2E Tests

**Scope**: Mixed fleet (package-mode RHEL + image-mode bootc devices)

**Validation**:
- Console UI displays mode correctly for each device type
- Metrics distinguish package-mode vs image-mode populations
- Rollout policies work correctly (don't attempt OS rollouts to package-mode devices)

## 8. RPM/DEB Package Assumptions

### Decision

Assume RPM/deb packages for agent already exist or will be created by packaging team.

**Scope of this feature**: Agent code modifications only. Packaging is out of scope per spec.md.

**Dependencies**:
- RPM: `systemd`, `podman`, standard RHEL 9+ libraries
- DEB: `systemd`, `podman`, standard Ubuntu 22.04+ libraries
- No bootc dependency in package spec (bootc absence is expected)

### Rationale

Feature spec explicitly lists ".deb package creation" in "Out of Scope" section. Focus on runtime behavior, not packaging.

## 9. Documentation Requirements

### User Documentation

**Files** (NEW):
1. `docs/user/installing/installing-agent-rhel-package-mode.md`:
   - RPM installation procedure
   - Agent enrollment
   - Configuration updates (no OS updates mentioned)

2. `docs/user/installing/installing-agent-ubuntu-package-mode.md`:
   - .deb installation procedure
   - systemd service management
   - Update behavior explanation

**Distinction from image-mode docs**: Explicitly state "OS updates managed by dnf/apt, not Flight Control".

### Developer Documentation

**File**: Update `docs/developer/architecture/architecture.md` (MODIFY)

**Section**: Add "Package-Mode vs Image-Mode" subsection

**Content**:
- Deployment mode detection mechanism
- Update workflow differences
- Testing guidance for package-mode scenarios

## 10. Observability & Metrics

### Logging

**Log Lines** (INFO level):
- Agent startup: "Package-mode detected: bootc not found"
- Agent startup: "Image-mode detected: bootc client initialized"
- Update reconciliation: "Package-mode device, skipping OS update to <image>"

**File**: `internal/agent/device/systeminfo/packagemode.go`, `internal/agent/device/device.go`

### Metrics

**Prometheus Metrics** (NEW):
- `flightctl_agent_deployment_mode{mode="package"|"image"}` (gauge, 0 or 1)
- `flightctl_agent_os_updates_skipped_total{reason="package_mode"}` (counter)

**Collection Point**: `internal/instrumentation/metrics.go` (MODIFY)

**Dashboard**: Fleet-wide mode distribution chart (package-mode vs image-mode device counts)

### Rationale

Operations teams need visibility into fleet composition. Metrics enable alerting on unexpected mode changes (e.g., bootc uninstalled on previously image-mode device).

## Summary of Decisions

| Aspect | Decision | Rationale |
|--------|----------|-----------|
| **Detection** | `exec.LookPath("bootc")` | Proven pattern, fast, platform-agnostic |
| **Modification Strategy** | Early exit in beforeUpdate/afterUpdate | Single modification point, minimal churn |
| **Status Field** | `AdditionalProperties["packageMode"]` | Backward compatible, no migration |
| **Detection Timing** | Agent initialization | One-time overhead, cached result |
| **Dummy Client** | Keep separate from package-mode | Clear distinction: intentional vs unsupported |
| **Testing** | Unit + Integration (RHEL/Ubuntu runners) | Platform-specific validation required |
| **Documentation** | Separate package-mode install guides | Avoid confusion with image-mode workflows |
| **Metrics** | Deployment mode gauge + skip counter | Fleet visibility and operational alerting |

All decisions align with Constitution Principle I (simplicity) and Principle III (backward compatibility).
