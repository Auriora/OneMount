# Manual D-Bus Integration Testing Guide

## Overview

This guide provides step-by-step instructions for manually testing OneMount's D-Bus integration. These tests verify that D-Bus signals are emitted correctly when file status changes occur, and that external clients (like Nemo file manager extensions) can receive and process these signals.

**Test Scope**: D-Bus signal emission, signal monitoring, and integration with file manager extensions

**Requirements Validated**: Requirement 10.2 - D-Bus signal emission for file status updates

---

## Prerequisites

### System Requirements

- Linux system with D-Bus installed and running
- OneMount installed and configured
- Valid OneDrive authentication tokens
- D-Bus development tools installed

### Required Tools Installation

```bash
# Install D-Bus tools (Ubuntu/Debian)
sudo apt-get install dbus dbus-x11 d-feet

# Install D-Bus tools (Fedora/RHEL)
sudo dnf install dbus dbus-x11 d-feet

# Verify D-Bus is running
systemctl --user status dbus
```

### Environment Configuration

1. **Ensure D-Bus session bus is available**:
   ```bash
   echo $DBUS_SESSION_BUS_ADDRESS
   # Should output something like: unix:path=/run/user/1000/bus
   ```

2. **Verify OneMount is built with D-Bus support**:
   ```bash
   ldd $(which onemount) | grep dbus
   # Should show libdbus-1.so or similar
   ```

3. **Set up test mount point**:
   ```bash
   mkdir -p ~/test-onedrive-mount
   ```

---

## Test Environment Setup

### 1. Prepare Host System for D-Bus Testing

**Important**: D-Bus testing must be performed on the host system, not in Docker containers, because D-Bus session bus communication requires proper user session integration.

1. **Stop any running OneMount instances**:
   ```bash
   # Find and unmount any existing OneMount mounts
   mount | grep onemount
   fusermount3 -uz ~/test-onedrive-mount 2>/dev/null || true
   
   # Kill any running OneMount processes
   pkill -f onemount
   ```

2. **Clear previous D-Bus service name files**:
   ```bash
   rm -f /tmp/onemount-dbus-service-name
   ```

3. **Start D-Bus monitor in a separate terminal**:
   ```bash
   # Terminal 1: Monitor all D-Bus signals from OneMount
   dbus-monitor --session "type='signal',interface='org.onemount.FileStatus'"
   ```

### 2. Launch OneMount with D-Bus Enabled

In a second terminal, mount OneMount:

```bash
# Terminal 2: Mount OneMount
onemount ~/test-onedrive-mount

# Verify mount succeeded
mount | grep test-onedrive-mount
```

**Expected Output**:
- OneMount should mount successfully
- D-Bus monitor (Terminal 1) should show connection activity
- Log should indicate "D-Bus server started"

---

## Automation Status

**Important**: Most D-Bus tests are now automated! See `docs/testing/dbus-test-automation-status.md` for details.

**Automated Tests** (Tests 1-6): These tests are fully automated and run in Docker:
- ✅ Test 1: D-Bus Service Registration - `TestIT_FS_DBus_Service*` tests
- ✅ Test 1a: D-Bus Service Discovery - `TestIT_FS_DBus_ServiceDiscovery`
- ✅ Test 1b: D-Bus Introspection Validation - `TestIT_FS_DBus_IntrospectionValidation`
- ✅ Test 2: File Status Signal Emission - `TestIT_FS_DBus_SendFileStatusUpdate`
- ✅ Test 3: Signal Content Validation - `TestIT_FS_DBus_SignalContent_*` tests
- ✅ Test 4: Signal Timing and Ordering - `TestIT_FS_DBus_SignalSequence_*` tests
- ✅ Test 5: Multiple Client Subscription - `TestIT_FS_DBus_MultipleInstances`
- ✅ Test 6: GetFileStatus Method Call - `TestIT_FS_DBus_GetFileStatus*` tests
- ✅ Test 7: External Client Simulation - `TestIT_FS_DBus_ExternalClientSimulation`

**Manual Tests** (Test 8 only): Only GUI-based testing requires manual verification:
- ❌ Test 8: D-Feet GUI Integration - Cannot be automated (requires GUI)

**When to Use Manual Tests**:
- Debugging D-Bus issues with visual tools
- Verifying GUI integration with D-Feet
- Exploring D-Bus interface interactively
- Learning how D-Bus works

**When to Use Automated Tests**:
- Regression testing during development
- CI/CD pipeline validation
- Quick verification of D-Bus functionality
- Testing without GUI environment

---

## Test Procedures

### Test 1: D-Bus Service Registration

**Status**: ✅ **AUTOMATED** - See `TestIT_FS_DBus_Service*` tests

**Objective**: Verify OneMount registers its D-Bus service correctly

**Steps**:

1. **Check D-Bus service name file**:
   ```bash
   cat /tmp/onemount-dbus-service-name
   ```
   
   **Expected**: Should show service name like `org.onemount.FileStatus.mnt_home_user_test_onedrive_mount`

2. **List D-Bus services**:
   ```bash
   dbus-send --session --print-reply \
     --dest=org.freedesktop.DBus \
     /org/freedesktop/DBus \
     org.freedesktop.DBus.ListNames | grep onemount
   ```
   
   **Expected**: Should show the OneMount service name

3. **Introspect D-Bus interface**:
   ```bash
   SERVICE_NAME=$(cat /tmp/onemount-dbus-service-name)
   dbus-send --session --print-reply \
     --dest=$SERVICE_NAME \
     /org/onemount/FileStatus \
     org.freedesktop.DBus.Introspectable.Introspect
   ```
   
   **Expected**: Should show XML introspection data with:
   - Interface: `org.onemount.FileStatus`
   - Method: `GetFileStatus` (input: path string, output: status string)
   - Signal: `FileStatusChanged` (args: path string, status string)

**Pass Criteria**:
- ✅ Service name file exists and contains valid service name
- ✅ Service is registered on D-Bus session bus
- ✅ Introspection shows correct interface, method, and signal definitions

---

### Test 2: File Status Signal Emission on File Access

**Status**: ✅ **AUTOMATED** - See `TestIT_FS_DBus_SendFileStatusUpdate` test

**Objective**: Verify D-Bus signals are emitted when files are accessed

**Steps**:

1. **Ensure D-Bus monitor is running** (Terminal 1)

2. **List directory contents** (Terminal 2):
   ```bash
   ls ~/test-onedrive-mount/
   ```

3. **Observe D-Bus monitor output** (Terminal 1):
   
   **Expected Signals**: Multiple `FileStatusChanged` signals for files in the directory
   
   Example signal format:
   ```
   signal time=1234567890.123456 sender=:1.234 -> destination=(null destination) serial=42 path=/org/onemount/FileStatus; interface=org.onemount.FileStatus; member=FileStatusChanged
      string "/path/to/file.txt"
      string "Cached"
   ```

4. **Access a specific file**:
   ```bash
   cat ~/test-onedrive-mount/test-file.txt
   ```

5. **Verify signal emission** (Terminal 1):
   
   **Expected**: Should see `FileStatusChanged` signals for the accessed file with status transitions:
   - Initial: `"Ghost"` or `"Cached"` (if previously accessed)
   - During download: `"Downloading"` (if not cached)
   - After download: `"Cached"`

**Pass Criteria**:
- ✅ Signals are emitted for directory listing operations
- ✅ Signals are emitted when individual files are accessed
- ✅ Signal format matches expected structure (path + status)
- ✅ Status values are valid (Ghost, Downloading, Cached, Modified, Error, etc.)

---

### Test 3: Signal Content Validation

**Status**: ✅ **AUTOMATED** - See `TestIT_FS_DBus_SignalContent_*` tests

**Objective**: Verify signal parameters contain correct data types and values

**Steps**:

1. **Create a test file**:
   ```bash
   echo "Test content" > ~/test-onedrive-mount/dbus-test-file.txt
   ```

2. **Monitor signals with detailed output**:
   ```bash
   dbus-monitor --session "type='signal',interface='org.onemount.FileStatus'" \
     | grep -A 5 FileStatusChanged
   ```

3. **Verify signal parameters**:
   - **Parameter 1 (path)**: Should be a string containing the full path
   - **Parameter 2 (status)**: Should be a string with valid status value

4. **Test with various file operations**:
   ```bash
   # Modify file
   echo "Modified content" >> ~/test-onedrive-mount/dbus-test-file.txt
   
   # Delete file
   rm ~/test-onedrive-mount/dbus-test-file.txt
   ```

5. **Verify each operation emits appropriate signals**:
   - Modify: Should emit signal with status `"Modified"` or `"Uploading"`
   - Delete: Should emit signal with status `"Deleted"` or similar

**Pass Criteria**:
- ✅ Path parameter is always a valid string
- ✅ Status parameter is always a valid string
- ✅ Status values match documented states (Ghost, Downloading, Cached, Modified, Uploading, Error, Deleted)
- ✅ Signals are emitted for all file state changes

---

### Test 4: Signal Timing and Ordering

**Status**: ✅ **AUTOMATED** - See `TestIT_FS_DBus_SignalSequence_*` tests

**Objective**: Verify signals are emitted in correct order and at appropriate times

**Steps**:

1. **Clear cache to force download**:
   ```bash
   # Unmount and clear cache
   fusermount3 -uz ~/test-onedrive-mount
   rm -rf ~/.cache/onemount/
   
   # Remount
   onemount ~/test-onedrive-mount
   ```

2. **Access a large file and monitor signal sequence**:
   ```bash
   # In Terminal 2
   cat ~/test-onedrive-mount/large-file.pdf > /dev/null
   ```

3. **Observe signal sequence in Terminal 1**:
   
   **Expected Order**:
   1. Initial signal: `"Ghost"` (file not cached)
   2. Download start: `"Downloading"`
   3. Download complete: `"Cached"`

4. **Verify timing**:
   - Signals should be emitted immediately when state changes
   - No significant delays between state change and signal emission
   - Signals should arrive in chronological order

**Pass Criteria**:
- ✅ Signals are emitted in correct state transition order
- ✅ Signal timing is immediate (< 100ms after state change)
- ✅ No duplicate signals for the same state
- ✅ No missing signals in the transition sequence

---

### Test 5: Multiple Client Subscription

**Status**: ✅ **AUTOMATED** - See `TestIT_FS_DBus_MultipleInstances` test

**Objective**: Verify multiple D-Bus clients can receive signals simultaneously

**Steps**:

1. **Start multiple D-Bus monitors** (3 separate terminals):
   
   ```bash
   # Terminal 3
   dbus-monitor --session "type='signal',interface='org.onemount.FileStatus'" \
     > /tmp/dbus-monitor-1.log 2>&1
   
   # Terminal 4
   dbus-monitor --session "type='signal',interface='org.onemount.FileStatus'" \
     > /tmp/dbus-monitor-2.log 2>&1
   
   # Terminal 5
   dbus-monitor --session "type='signal',interface='org.onemount.FileStatus'" \
     > /tmp/dbus-monitor-3.log 2>&1
   ```

2. **Perform file operations** (Terminal 2):
   ```bash
   # Create file
   echo "Multi-client test" > ~/test-onedrive-mount/multi-test.txt
   
   # Wait for upload
   sleep 5
   
   # Modify file
   echo "Modified" >> ~/test-onedrive-mount/multi-test.txt
   ```

3. **Stop monitors and compare logs**:
   ```bash
   # Stop monitors (Ctrl+C in each terminal)
   
   # Compare logs
   diff /tmp/dbus-monitor-1.log /tmp/dbus-monitor-2.log
   diff /tmp/dbus-monitor-2.log /tmp/dbus-monitor-3.log
   ```

**Pass Criteria**:
- ✅ All monitors receive the same signals
- ✅ Signal order is consistent across all monitors
- ✅ No signals are lost or duplicated
- ✅ All clients receive signals simultaneously (within timing tolerance)

---

### Test 6: GetFileStatus Method Call

**Status**: ✅ **AUTOMATED** - See `TestIT_FS_DBus_GetFileStatus*` tests

**Objective**: Verify the D-Bus method for querying file status works correctly

**Steps**:

1. **Get service name**:
   ```bash
   SERVICE_NAME=$(cat /tmp/onemount-dbus-service-name)
   ```

2. **Query status of a known file**:
   ```bash
   dbus-send --session --print-reply \
     --dest=$SERVICE_NAME \
     /org/onemount/FileStatus \
     org.onemount.FileStatus.GetFileStatus \
     string:"/test-file.txt"
   ```
   
   **Expected Output**:
   ```
   method return time=... sender=... -> destination=... serial=...
      string "Cached"
   ```

3. **Query status of non-existent file**:
   ```bash
   dbus-send --session --print-reply \
     --dest=$SERVICE_NAME \
     /org/onemount/FileStatus \
     org.onemount.FileStatus.GetFileStatus \
     string:"/non-existent-file.txt"
   ```
   
   **Expected Output**:
   ```
   method return time=... sender=... -> destination=... serial=...
      string "Unknown"
   ```

4. **Query status during download**:
   ```bash
   # Clear cache
   fusermount3 -uz ~/test-onedrive-mount
   rm -rf ~/.cache/onemount/
   onemount ~/test-onedrive-mount
   
   # Start download in background
   cat ~/test-onedrive-mount/large-file.pdf > /dev/null &
   
   # Query status immediately
   dbus-send --session --print-reply \
     --dest=$SERVICE_NAME \
     /org/onemount/FileStatus \
     org.onemount.FileStatus.GetFileStatus \
     string:"/large-file.pdf"
   ```
   
   **Expected**: Status should be `"Downloading"` or `"Cached"` depending on timing

**Pass Criteria**:
- ✅ Method returns correct status for existing files
- ✅ Method returns "Unknown" for non-existent files
- ✅ Method returns current status (not stale cached status)
- ✅ Method responds quickly (< 100ms)

---

## Optional Debugging Tools

### Test 7: D-Bus Integration with D-Feet GUI Tool

**Status**: ❌ **MANUAL ONLY** - Cannot be automated (requires GUI)

**Objective**: Verify D-Bus interface using graphical inspection tool

**Note**: This test is optional and primarily useful for debugging D-Bus issues. The automated tests provide comprehensive coverage of D-Bus functionality without requiring GUI tools.

**When to Use D-Feet**:
- Debugging D-Bus communication issues
- Exploring D-Bus interface structure interactively
- Learning how D-Bus signals work
- Verifying signal emission visually
- Testing method calls manually

**Steps**:

1. **Launch D-Feet**:
   ```bash
   d-feet &
   ```

2. **Connect to Session Bus**:
   - Click "Session Bus" tab
   - Search for "onemount" in the service list

3. **Inspect OneMount service**:
   - Click on the OneMount service (e.g., `org.onemount.FileStatus.mnt_...`)
   - Navigate to `/org/onemount/FileStatus` object path
   - Verify interface `org.onemount.FileStatus` is present

4. **Test GetFileStatus method**:
   - Select `GetFileStatus` method
   - Enter a file path (e.g., `/test-file.txt`)
   - Click "Execute"
   - Verify status is returned

5. **Monitor signals**:
   - Keep D-Feet open
   - Perform file operations in OneMount
   - Verify `FileStatusChanged` signals appear in D-Feet's signal log

**Pass Criteria**:
- ✅ OneMount service is visible in D-Feet
- ✅ Interface and methods are correctly exposed
- ✅ Method calls work from D-Feet GUI
- ✅ Signals are visible in D-Feet's signal monitor

---

## Expected Results Summary

### Signal Format

All `FileStatusChanged` signals should follow this format:

```
signal sender=:1.XXX -> destination=(null destination)
  path=/org/onemount/FileStatus
  interface=org.onemount.FileStatus
  member=FileStatusChanged
  string "<file-path>"
  string "<status>"
```

### Valid Status Values

- `"Ghost"` - File metadata known, content not cached
- `"Downloading"` - File content being downloaded
- `"Cached"` - File content cached locally
- `"Modified"` - File modified locally, pending upload
- `"Uploading"` - File being uploaded to OneDrive
- `"Error"` - Error occurred during operation
- `"Deleted"` - File deleted
- `"Unknown"` - File not found in filesystem

### Signal Emission Triggers

Signals should be emitted when:
- File is accessed for the first time (Ghost → Downloading)
- Download completes (Downloading → Cached)
- File is modified (Cached → Modified)
- Upload starts (Modified → Uploading)
- Upload completes (Uploading → Cached)
- Error occurs (any state → Error)
- File is deleted (any state → Deleted)

---

## Troubleshooting

### Issue: No D-Bus signals received

**Possible Causes**:
1. D-Bus session bus not running
2. OneMount not connected to D-Bus
3. D-Bus monitor not listening to correct interface

**Solutions**:
```bash
# Check D-Bus session bus
echo $DBUS_SESSION_BUS_ADDRESS
systemctl --user status dbus

# Check OneMount logs for D-Bus errors
journalctl --user -u onemount -f | grep -i dbus

# Verify D-Bus monitor command
dbus-monitor --session "type='signal',interface='org.onemount.FileStatus'"
```

### Issue: Service name not found

**Possible Causes**:
1. OneMount not running
2. D-Bus server failed to start
3. Service name file not created

**Solutions**:
```bash
# Check if OneMount is running
ps aux | grep onemount

# Check service name file
cat /tmp/onemount-dbus-service-name

# Check OneMount logs
journalctl --user -u onemount | grep "D-Bus"
```

### Issue: Signals emitted but with wrong data

**Possible Causes**:
1. Bug in signal emission code
2. Race condition in status updates
3. Stale status cache

**Solutions**:
```bash
# Enable debug logging
onemount --log-level=debug ~/test-onedrive-mount

# Check logs for status update sequence
journalctl --user -u onemount | grep "file status"
```

### Issue: Multiple signals for same state

**Possible Causes**:
1. Status update called multiple times
2. Race condition in file operations
3. Cache invalidation triggering duplicate updates

**Solutions**:
- Review logs for duplicate status update calls
- Check if file operations are being retried
- Verify cache invalidation logic

---

## Results Documentation Template

### Test Execution Record

**Date**: _______________  
**Tester**: _______________  
**OneMount Version**: _______________  
**System**: _______________  

### Test Results

| Test # | Test Name | Status | Notes |
|--------|-----------|--------|-------|
| 1 | D-Bus Service Registration | ☐ Pass ☐ Fail | |
| 2 | File Status Signal Emission | ☐ Pass ☐ Fail | |
| 3 | Signal Content Validation | ☐ Pass ☐ Fail | |
| 4 | Signal Timing and Ordering | ☐ Pass ☐ Fail | |
| 5 | Multiple Client Subscription | ☐ Pass ☐ Fail | |
| 6 | GetFileStatus Method Call | ☐ Pass ☐ Fail | |
| 7 | D-Feet GUI Integration | ☐ Pass ☐ Fail | |

### Issues Found

| Issue # | Description | Severity | Steps to Reproduce |
|---------|-------------|----------|-------------------|
| | | | |

### Overall Assessment

☐ **PASS** - All tests passed, D-Bus integration working correctly  
☐ **FAIL** - One or more critical tests failed  
☐ **PARTIAL** - Minor issues found but core functionality works  

**Comments**: _______________________________________________

---

## Pass/Fail Criteria

### Overall Pass Criteria

The D-Bus integration test suite **PASSES** if:

1. ✅ OneMount successfully registers its D-Bus service on mount
2. ✅ D-Bus signals are emitted for all file status changes
3. ✅ Signal format and content are correct (path + status strings)
4. ✅ Signals are received by multiple clients simultaneously
5. ✅ GetFileStatus method returns accurate current status
6. ✅ No crashes or errors in D-Bus communication
7. ✅ Signal timing is immediate (< 100ms after state change)

### Critical Failures

The test suite **FAILS** if:

1. ❌ D-Bus service fails to register
2. ❌ No signals are emitted for file operations
3. ❌ Signals contain incorrect or malformed data
4. ❌ GetFileStatus method returns wrong status
5. ❌ D-Bus errors cause OneMount to crash or hang
6. ❌ Signals are not received by external clients

### Known Limitations

- D-Bus testing requires a running session bus (cannot be fully automated in Docker)
- Signal timing may vary based on system load
- Some file operations may emit multiple signals (e.g., modify + upload)
- Status transitions may be too fast to observe for small files

---

## Additional Resources

### D-Bus Documentation

- [D-Bus Specification](https://dbus.freedesktop.org/doc/dbus-specification.html)
- [D-Bus Tutorial](https://dbus.freedesktop.org/doc/dbus-tutorial.html)
- [godbus/dbus Go Library](https://github.com/godbus/dbus)

### OneMount D-Bus Implementation

- Source: `internal/fs/dbus.go`
- Interface: `org.onemount.FileStatus`
- Object Path: `/org/onemount/FileStatus`
- Service Name: `org.onemount.FileStatus.<mount-path-escaped>`

### Related Requirements

- Requirement 10.1: Extended attribute updates
- Requirement 10.2: D-Bus signal emission
- Requirement 10.3: Nemo extension integration
- Requirement 10.4: D-Bus fallback behavior
- Requirement 10.5: Download progress updates
