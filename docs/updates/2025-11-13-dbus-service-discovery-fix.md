# D-Bus Service Name Discovery Fix

**Date**: 2025-11-13  
**Issue**: #FS-002  
**Component**: D-Bus Server / Nemo Extension  
**Status**: ✅ Resolved

## Problem

The D-Bus service name includes a unique suffix (PID + timestamp) to avoid conflicts between multiple OneMount instances, but the Nemo extension used a hardcoded base name `org.onemount.FileStatus`. This mismatch prevented the Nemo extension from connecting to the D-Bus service via method calls.

### Symptoms

- Nemo extension could not connect to D-Bus service
- `GetFileStatus` method calls failed
- Extension fell back to extended attributes only
- D-Bus signals worked (if client subscribed correctly) but method calls did not

### Root Cause

Mismatch between:
- **Server**: Dynamic service name generation (e.g., `org.onemount.FileStatus.instance_12345_67890`)
- **Client**: Static service name (e.g., `org.onemount.FileStatus`)

## Solution

Implemented **Option 3**: Write service name to a known location for discovery.

### Implementation Details

#### 1. D-Bus Server Changes (`internal/fs/dbus.go`)

Added functionality to write the service name to a file when the server starts:

- **File Location**: `/tmp/onemount-dbus-service-name`
- **File Permissions**: 0600 (owner read/write only)
- **Atomic Write**: Uses temporary file + rename for atomicity
- **Cleanup**: Removes file on server stop (only if it contains our service name)

**New Functions**:
- `writeServiceNameFile()`: Writes service name to file atomically
- `removeServiceNameFile()`: Removes file on cleanup (with safety check)

**Modified Functions**:
- `Start()`: Calls `writeServiceNameFile()` after successful D-Bus registration
- `Stop()`: Calls `removeServiceNameFile()` during cleanup

#### 2. Nemo Extension Changes (`internal/nemo/src/nemo-onemount.py`)

Added service discovery functionality:

**New Method**:
- `_discover_dbus_service_name()`: Reads service name from file, falls back to base name

**Modified Method**:
- `connect_to_dbus()`: Uses discovered service name instead of hardcoded name

### Discovery Flow

```
1. OneMount starts D-Bus server
   └─> Generates unique service name: org.onemount.FileStatus.instance_12345_67890
   └─> Registers with D-Bus
   └─> Writes service name to /tmp/onemount-dbus-service-name

2. Nemo extension initializes
   └─> Calls _discover_dbus_service_name()
   └─> Reads /tmp/onemount-dbus-service-name
   └─> Gets actual service name: org.onemount.FileStatus.instance_12345_67890
   └─> Connects to D-Bus using discovered name

3. OneMount stops
   └─> Checks if file contains our service name
   └─> Removes /tmp/onemount-dbus-service-name
```

### Multiple Instance Support

The solution handles multiple OneMount instances gracefully:

1. **Last Writer Wins**: When multiple instances run, the most recent one's service name is in the file
2. **Safe Cleanup**: Each instance only removes the file if it contains its own service name
3. **Fallback**: If file doesn't exist or is unreadable, extension falls back to base name

## Testing

### Go Tests (`internal/fs/dbus_service_discovery_test.go`)

Created comprehensive tests:

1. **TestDBusServiceNameFileCreation**: Verifies file is created with correct content
2. **TestDBusServiceNameFileCleanup**: Verifies file is removed on server stop
3. **TestDBusServiceNameFileMultipleInstances**: Verifies multiple instances don't interfere

All tests pass:
```
PASS: TestDBusServiceNameFileCreation (0.11s)
PASS: TestDBusServiceNameFileCleanup (0.10s)
PASS: TestDBusServiceNameFileMultipleInstances (0.11s)
```

### Python Tests (`internal/nemo/tests/test_service_discovery.py`)

Created unit tests for discovery logic:

1. **test_discover_service_name_from_file**: Verifies reading from file
2. **test_discover_service_name_with_whitespace**: Verifies whitespace handling
3. **test_discover_service_name_fallback_nonexistent**: Verifies fallback when file missing
4. **test_discover_service_name_fallback_empty**: Verifies fallback when file empty
5. **test_discover_service_name_fallback_whitespace_only**: Verifies fallback when file has only whitespace

All tests pass:
```
Ran 5 tests in 0.002s
OK
```

## Benefits

1. **Compatibility**: Works with multiple OneMount instances
2. **Simplicity**: Simple file-based discovery mechanism
3. **Robustness**: Graceful fallback to base name if file unavailable
4. **Security**: File has restricted permissions (0600)
5. **Atomicity**: Uses atomic rename to prevent race conditions
6. **Safety**: Only removes file if it contains our service name

## Affected Files

### Modified
- `internal/fs/dbus.go`: Added service name file management
- `internal/nemo/src/nemo-onemount.py`: Added service discovery

### Created
- `internal/fs/dbus_service_discovery_test.go`: Go tests
- `internal/nemo/tests/test_service_discovery.py`: Python tests
- `docs/updates/2025-11-13-dbus-service-discovery-fix.md`: This document

## Requirements Addressed

- **Requirement 8.2**: D-Bus integration for file status updates
- **Requirement 8.3**: Nemo extension integration with status icons

## Future Considerations

### Alternative Approaches (Not Implemented)

1. **Option 1**: Use well-known service name without unique suffix
   - **Pros**: Simpler, no discovery needed
   - **Cons**: Conflicts with multiple instances

2. **Option 2**: D-Bus introspection for service discovery
   - **Pros**: More "D-Bus native"
   - **Cons**: More complex, requires listing all services

### Potential Improvements

1. **XDG Base Directory**: Consider using `$XDG_RUNTIME_DIR` instead of `/tmp`
2. **File Locking**: Add file locking to prevent race conditions
3. **Service Registry**: Maintain a registry of all active OneMount instances
4. **Automatic Cleanup**: Add systemd-tmpfiles or similar for stale file cleanup

## Verification

To verify the fix works:

1. Start OneMount with D-Bus enabled
2. Check that `/tmp/onemount-dbus-service-name` exists and contains the service name
3. Open Nemo file manager and navigate to a OneMount mount
4. Verify that file status icons appear correctly
5. Stop OneMount and verify the file is removed

## Related Issues

- Issue #FS-001: D-Bus GetFileStatus returns Unknown (separate issue, not addressed here)
- Issue #FS-003: No error handling for extended attributes (separate issue)

## References

- Task: 20.16 Fix Issue #FS-002: D-Bus Service Name Discovery Problem
- Verification Document: `docs/reports/verification-tracking.md` (Phase 11)
