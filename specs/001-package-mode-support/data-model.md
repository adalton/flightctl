# Data Model: Package Mode Support

**Date**: 2026-01-22
**Feature**: 001-package-mode-support
**Purpose**: Define data structures for package-mode detection and status reporting

## Entity: DeviceSystemInfo Extension

### Schema Change

**File**: `api/core/v1beta1/openapi.yaml`

**Change Type**: Extension (backward compatible)

**Existing Schema**:
```yaml
DeviceSystemInfo:
  type: object
  properties:
    agentVersion:
      type: string
    architecture:
      type: string
    bootID:
      type: string
    operatingSystem:
      type: string
    customInfo:
      $ref: '#/components/schemas/CustomDeviceInfo'
  additionalProperties:  # <-- USED FOR PACKAGE-MODE
    type: string
```

### New Property

**Key**: `packageMode`
**Type**: `string` (values: `"true"` | `"false"`)
**Location**: `DeviceSystemInfo.additionalProperties["packageMode"]`

**Rationale**: `additionalProperties` is a map[string]string that allows arbitrary key-value pairs without schema changes. No database migration required.

### Example Status Response

**Before** (image-mode device):
```json
{
  "status": {
    "systemInfo": {
      "agentVersion": "v0.9.0",
      "architecture": "amd64",
      "bootID": "a3b2c1d4-...",
      "operatingSystem": "linux",
      "additionalProperties": {}
    },
    "os": {
      "image": "quay.io/flightctl/flightctl-agent-fedora",
      "imageDigest": "sha256:6adcbcf..."
    }
  }
}
```

**After** (package-mode device):
```json
{
  "status": {
    "systemInfo": {
      "agentVersion": "v0.9.0",
      "architecture": "amd64",
      "bootID": "e5f6g7h8-...",
      "operatingSystem": "linux",
      "additionalProperties": {
        "packageMode": "true"
      }
    },
    "os": {
      "image": "",
      "imageDigest": ""
    }
  }
}
```

### Field Semantics

| Value | Meaning | Bootc Presence | Update Behavior |
|-------|---------|----------------|-----------------|
| `"true"` | Package-mode deployment | Absent | Config/app updates only, OS updates skipped |
| `"false"` | Image-mode deployment | Present | Full management including OS images |
| (absent) | Unknown/legacy agent | N/A | Assume image-mode for compatibility |

**Backward Compatibility**: Older agents don't report `packageMode` field. API defaults to image-mode behavior (safe default).

## Entity: Device Model (Internal)

### Internal Representation

**File**: `internal/store/model/device.go`

**No changes required**. Device model stores raw JSON in `Status` field (PostgreSQL JSONB type). `additionalProperties` automatically persists.

**Existing Schema**:
```go
type Device struct {
    Resource  // ID, Name, Labels, etc.

    Spec   *v1beta1.DeviceSpec   `json:"spec,omitempty"`
    Status *v1beta1.DeviceStatus `json:"status,omitempty"`  // <-- Contains systemInfo
}
```

**Query Support**: PostgreSQL JSONB operators enable filtering:
```sql
SELECT * FROM devices
WHERE status->'systemInfo'->'additionalProperties'->>'packageMode' = 'true';
```

No index required initially (fleet sizes <100k devices). Add if query performance degrades.

## Entity: SystemInfo Manager (Runtime)

### In-Memory Cache

**File**: `internal/agent/device/systeminfo/manager.go`

**New Field**:
```go
type manager struct {
    // ... existing fields ...
    bootTime    time.Time
    bootID      string
    isRebooted  bool
    packageMode bool  // NEW: cached package-mode detection result
    // ... other fields ...
}
```

**Initialization** (line ~90):
```go
func (m *manager) Initialize(ctx context.Context) (err error) {
    m.bootTime, err = getBootTime(ctx, m.exec)
    // ... existing boot ID logic ...

    m.packageMode = DetectPackageMode()  // NEW
    if m.packageMode {
        m.log.Info("Package-mode detected: bootc not found")
    } else {
        m.log.Info("Image-mode detected: bootc client available")
    }

    // ... existing reboot detection ...
}
```

**Accessor**:
```go
func (m *manager) IsPackageMode() bool {
    return m.packageMode
}
```

**Status Export** (existing `Status()` method):
```go
func (m *manager) Status(
    ctx context.Context,
    status *v1beta1.DeviceStatus,
    _ ...CollectorOpt,
) error {
    systemInfo, err := collectDeviceSystemInfo(/* ... */)

    // NEW: Add package-mode to additional properties
    if m.packageMode {
        systemInfo.AdditionalProperties["packageMode"] = "true"
    } else {
        systemInfo.AdditionalProperties["packageMode"] = "false"
    }

    status.SystemInfo = systemInfo
    return nil
}
```

## Validation Rules

### Agent-Side Validation

1. **Detection Consistency**: Package-mode flag must not change during agent process lifetime
   - **Check**: Compare cached value with runtime detection on every status collection
   - **Action**: Log ERROR if mismatch detected (indicates bootc install/uninstall without agent restart)

2. **Status Completeness**: If `packageMode="true"`, `os.image` must be empty
   - **Check**: In status collector
   - **Action**: Log WARN if package-mode device has populated OS image field

### API-Side Validation

**None required**. `additionalProperties` accepts arbitrary string key-value pairs per OpenAPI schema.

### Database Validation

**None required**. JSONB field stores arbitrary JSON. No constraints needed.

## State Transitions

### Package-Mode Detection State Machine

```
[Agent Start]
     |
     v
Detect bootc binary
     |
     +-- Found --> packageMode = false (image-mode)
     |
     +-- Not Found --> packageMode = true (package-mode)
     |
     v
[Cache Result]
     |
     v
[Report in Status]
     |
     v
[Never Changes Until Agent Restart]
```

**No runtime transitions**. Deployment mode is process-lifetime constant.

### Mode Change Scenario

**Scenario**: Administrator installs bootc on package-mode device

**Steps**:
1. Device currently running in package-mode (`packageMode="true"`)
2. Admin runs: `dnf install bootc` (RHEL) or `apt install bootc` (Ubuntu)
3. Agent continues running (no restart)
4. **Result**: Agent still reports `packageMode="true"` (cached value)
5. **Resolution**: Restart agent → re-detection → `packageMode="false"`

**Rationale**: Avoids mid-process mode confusion. Clear boundary: restart required for mode change.

## Relationships

### Package-Mode → OS Update Skipping

**Relationship**: Package-mode flag controls OS update workflow

**Constraint**: If `packageMode="true"`, agent MUST NOT:
- Prefetch OS images
- Execute `bootc switch`
- Execute `bootc upgrade --apply`
- Reboot for OS updates

**Implementation**: Conditional in `internal/agent/device/device.go:340,460`

### Package-Mode → Console UI Display

**Relationship**: Console UI renders mode indicator from `additionalProperties`

**Query**: `GET /api/v1/devices/{deviceId}`

**Response Field**: `status.systemInfo.additionalProperties.packageMode`

**UI Rendering**:
- `"true"` → Display badge: "Package Mode" (blue)
- `"false"` → Display badge: "Image Mode" (green)
- (absent) → Display badge: "Image Mode*" (gray, tooltip: "Legacy agent, mode unknown")

## Database Schema

### No Migration Required

**Existing Column**: `devices.status` (JSONB type)

**New Data**: `status->systemInfo->additionalProperties->packageMode`

**Index Consideration**:
```sql
-- Optional: Create GIN index for package-mode queries (if fleet size >100k)
CREATE INDEX CONCURRENTLY idx_devices_package_mode
ON devices USING gin ((status->'systemInfo'->'additionalProperties'));
```

**Defer until performance testing**. GIN indexes add write overhead.

## API Contract Changes

### Device Status Endpoint

**Endpoint**: `GET /api/v1/devices/{deviceId}`

**Schema Change**: None (uses `additionalProperties`)

**Example Response Diff**:

```diff
 {
   "status": {
     "systemInfo": {
       "agentVersion": "v0.9.0",
       "architecture": "amd64",
       "bootID": "...",
       "operatingSystem": "linux",
       "additionalProperties": {
+        "packageMode": "true"
       }
     }
   }
 }
```

### Device List Endpoint

**Endpoint**: `GET /api/v1/devices`

**Filtering Support** (optional enhancement):
```
GET /api/v1/devices?packageMode=true
```

**Implementation**: Parse query param, add PostgreSQL JSONB filter

**Priority**: P2 (nice-to-have, not required for MVP)

## Summary Table

| Entity | Change Type | File | Impact |
|--------|-------------|------|--------|
| DeviceSystemInfo | Extension | `api/core/v1beta1/openapi.yaml` | Add `packageMode` to `additionalProperties` |
| SystemInfo Manager | Runtime field | `internal/agent/device/systeminfo/manager.go` | Cache `packageMode` boolean |
| Device Model | No change | `internal/store/model/device.go` | JSONB auto-persists new field |
| Status API | No change | `api/core/v1beta1/openapi.yaml` | Backward compatible extension |

**Database Migration**: Not required
**API Version Bump**: Not required
**Breaking Changes**: None
