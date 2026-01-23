# Manual D-Bus Fallback Testing Guide

## Overview

This guide provides step-by-step instructions for manually testing OneMount's graceful degradation when D-Bus is unavailable. These tests verify that OneMount continues to function correctly using extended attributes as a fallback mechanism when D-Bus cannot be used.

**Test Scope**: System operation without D-Bus, graceful degradation, extended attribute fallback

**Requirements Validated**: Requirement 10.4 - System continues operating when D-Bus is unavailable

---

## Prerequisites

### System Requirements

- Linux system with OneMount installed
- Valid OneDrive authentication tokens
- Ability to disable D-Bus temporarily
- Root or sudo access (for some D-Bus control operations)

### Required Tools

```bash
# Install testing utilities (Ubuntu/Debian)
sudo apt-get install attr procps

# Install testing utilities (Fedora/RHEL)
sudo dnf install attr procps-ng
```

---

## Test Environment Setup

### Understanding D-Bus Fallback Behavior

When D-Bus is unavailable, OneMount should:
- ✅ Continue mounting and operating normally
- ✅ Store file status in extended attributes (xattrs)
- ✅ Serve file status queries via xattrs
- ✅ Log D-Bus unavailability (not treat as error)
- ✅ Not crash or hang
- ✅ Maintain full functionality for file operations

### Methods to Disable D-Bus for Testing

There are several ways to test D-Bus fallback:

#### Method 1: Unset D-Bus Environment Variable (Recommended)

This is the safest method as it only affects the OneMount process:

```bash
# Unset D-Bus session bus address before launching OneMount
unset DBUS_SESSION_BUS_ADDRESS
onemount ~/test-onedrive-mount
```

#### Method 2: Stop D-Bus Session Service (More Invasive)

**Warning**: This will affect all applications in your session!

```bash
# Stop D-Bus user session
systemctl --user stop dbus

# Verify D-Bus is stopped
systemctl --user status dbus
```


#### Method 3: Use Network Namespace (Advanced)

Create an isolated environment without D-Bus:

```bash
# Create network namespace without D-Bus
sudo unshare --user --map-root-user --mount --pid --fork bash
# Inside namespace:
unset DBUS_SESSION_BUS_ADDRESS
onemount ~/test-onedrive-mount
```

### Recommended Test Setup

For this guide, we'll use **Method 1** (unset environment variable) as it's the safest and most reliable.

---

## Test Procedures

### Test 1: Mount Filesystem Without D-Bus

**Objective**: Verify OneMount can mount successfully when D-Bus is unavailable

**Steps**:

1. **Ensure no OneMount instances are running**:
   ```bash
   pkill -f onemount
   fusermount3 -uz ~/test-onedrive-mount 2>/dev/null || true
   ```

2. **Create mount point**:
   ```bash
   mkdir -p ~/test-onedrive-mount
   ```

3. **Launch OneMount without D-Bus**:
   ```bash
   unset DBUS_SESSION_BUS_ADDRESS
   onemount --log-level=debug ~/test-onedrive-mount 2>&1 | tee /tmp/onemount-no-dbus.log
   ```

4. **Verify mount succeeded**:
   ```bash
   # In another terminal
   mount | grep test-onedrive-mount
   ls ~/test-onedrive-mount/
   ```

5. **Check logs for D-Bus messages**:
   ```bash
   grep -i dbus /tmp/onemount-no-dbus.log
   ```

**Expected Results**:
- ✅ OneMount mounts successfully
- ✅ Directory listing works
- ✅ Logs indicate D-Bus is unavailable (not an error)
- ✅ No crashes or hangs
- ✅ Log message: "Failed to connect to D-Bus session bus" or similar

**Pass Criteria**:
- Mount operation completes successfully
- Filesystem is accessible and functional
- D-Bus unavailability is logged as info/warning, not error
- No process crashes or hangs

---

### Test 2: Verify Core File Operations Work

**Objective**: Confirm all core file operations function without D-Bus

**Steps**:

1. **Test directory listing**:
   ```bash
   ls -la ~/test-onedrive-mount/
   ```

2. **Test file reading**:
   ```bash
   cat ~/test-onedrive-mount/test-file.txt
   ```

3. **Test file creation**:
   ```bash
   echo "Fallback test" > ~/test-onedrive-mount/fallback-test.txt
   ```

4. **Test file modification**:
   ```bash
   echo "Modified content" >> ~/test-onedrive-mount/fallback-test.txt
   ```

5. **Test file deletion**:
   ```bash
   rm ~/test-onedrive-mount/fallback-test.txt
   ```

6. **Test directory operations**:
   ```bash
   mkdir ~/test-onedrive-mount/test-dir
   touch ~/test-onedrive-mount/test-dir/file.txt
   rmdir ~/test-onedrive-mount/test-dir
   ```

**Expected Results**:
- ✅ All operations complete successfully
- ✅ No errors related to D-Bus
- ✅ Files sync to OneDrive correctly
- ✅ Performance is comparable to D-Bus-enabled mode

**Pass Criteria**:
- All file operations succeed
- No D-Bus-related errors
- Files are correctly synced to OneDrive
- No functional degradation

---

### Test 3: Extended Attributes Still Function

**Objective**: Verify file status is available via extended attributes

**Steps**:

1. **Check extended attributes on a file**:
   ```bash
   getfattr -d ~/test-onedrive-mount/test-file.txt
   ```

2. **Look for status attribute**:
   ```bash
   getfattr -n user.onemount.status ~/test-onedrive-mount/test-file.txt
   ```

3. **Access a new file and check status**:
   ```bash
   # Clear cache to force download
   fusermount3 -uz ~/test-onedrive-mount
   rm -rf ~/.cache/onemount/
   
   # Remount without D-Bus
   unset DBUS_SESSION_BUS_ADDRESS
   onemount ~/test-onedrive-mount
   
   # Access file
   cat ~/test-onedrive-mount/large-file.pdf > /dev/null
   
   # Check status
   getfattr -n user.onemount.status ~/test-onedrive-mount/large-file.pdf
   ```

**Expected Results**:
- ✅ Extended attributes are present
- ✅ Status attribute contains valid value (Cached, Modified, etc.)
- ✅ Status updates as file state changes
- ✅ No D-Bus dependency for status queries

**Pass Criteria**:
- Extended attributes are accessible
- Status values are accurate and current
- Status updates reflect actual file state

---


### Test 4: Graceful Degradation - No Crashes or Errors

**Objective**: Confirm system handles D-Bus absence gracefully

**Steps**:

1. **Monitor system logs during operations**:
   ```bash
   # In Terminal 1: Watch logs
   tail -f /tmp/onemount-no-dbus.log | grep -i "error\|crash\|panic\|fatal"
   ```

2. **Perform stress test operations** (Terminal 2):
   ```bash
   # Create multiple files
   for i in {1..10}; do
     echo "Test $i" > ~/test-onedrive-mount/test-$i.txt
   done
   
   # Read multiple files
   for i in {1..10}; do
     cat ~/test-onedrive-mount/test-$i.txt > /dev/null
   done
   
   # Delete files
   rm ~/test-onedrive-mount/test-*.txt
   ```

3. **Check for error messages**:
   ```bash
   grep -i "error" /tmp/onemount-no-dbus.log | grep -v "D-Bus"
   ```

4. **Verify process is still running**:
   ```bash
   ps aux | grep onemount
   ```

**Expected Results**:
- ✅ No crashes or panics
- ✅ No error messages (except D-Bus unavailability notice)
- ✅ Process remains stable
- ✅ All operations complete successfully

**Pass Criteria**:
- Zero crashes during test operations
- No error-level log messages (except D-Bus connection failure)
- Process remains responsive
- All file operations succeed

---

### Test 5: Status Reporting via Alternative Methods

**Objective**: Verify status can be queried without D-Bus

**Steps**:

1. **Query status via extended attributes**:
   ```bash
   # Function to get file status
   get_status() {
     getfattr -n user.onemount.status --only-values "$1" 2>/dev/null || echo "Unknown"
   }
   
   # Test on various files
   get_status ~/test-onedrive-mount/test-file.txt
   ```

2. **Test status during file operations**:
   ```bash
   # Create file
   echo "Status test" > ~/test-onedrive-mount/status-test.txt
   get_status ~/test-onedrive-mount/status-test.txt
   # Expected: "Modified" or "Uploading"
   
   # Wait for upload
   sleep 5
   get_status ~/test-onedrive-mount/status-test.txt
   # Expected: "Cached"
   ```

3. **Compare with D-Bus-enabled mode** (if possible):
   ```bash
   # Remount with D-Bus
   fusermount3 -uz ~/test-onedrive-mount
   onemount ~/test-onedrive-mount
   
   # Check same file
   get_status ~/test-onedrive-mount/status-test.txt
   ```

**Expected Results**:
- ✅ Status is available via xattrs
- ✅ Status values match D-Bus mode
- ✅ Status updates in real-time
- ✅ No functional difference from D-Bus mode

**Pass Criteria**:
- Extended attributes provide accurate status
- Status values are identical to D-Bus mode
- No loss of functionality

---

### Test 6: Log Messages Indicate D-Bus Unavailability

**Objective**: Verify appropriate logging when D-Bus is unavailable

**Steps**:

1. **Review startup logs**:
   ```bash
   head -50 /tmp/onemount-no-dbus.log | grep -i dbus
   ```

2. **Check log level of D-Bus messages**:
   ```bash
   grep -i dbus /tmp/onemount-no-dbus.log | head -5
   ```

3. **Verify message content**:
   - Should indicate D-Bus connection failed
   - Should NOT be logged as ERROR or FATAL
   - Should be INFO or DEBUG level
   - Should mention fallback to xattrs (if applicable)

**Expected Log Messages**:
```
INFO: Failed to connect to D-Bus session bus: ...
DEBUG: D-Bus server not available, using extended attributes only
```

**NOT Expected**:
```
ERROR: D-Bus connection failed
FATAL: Cannot start without D-Bus
```

**Pass Criteria**:
- D-Bus unavailability is logged
- Log level is INFO or DEBUG (not ERROR/FATAL)
- Message indicates graceful fallback
- No misleading error messages

---

### Test 7: Comparison Testing - D-Bus Enabled vs Disabled

**Objective**: Compare functionality and performance with D-Bus enabled vs disabled

**Setup**:

Create a test script to compare both modes:

```bash
#!/bin/bash
# compare-dbus-modes.sh

MOUNT_POINT=~/test-onedrive-mount
TEST_FILE=$MOUNT_POINT/comparison-test.txt

# Function to test operations
test_operations() {
  local mode=$1
  echo "Testing in $mode mode..."
  
  # Create file
  start=$(date +%s%N)
  echo "Test content" > $TEST_FILE
  create_time=$(($(date +%s%N) - start))
  
  # Read file
  start=$(date +%s%N)
  cat $TEST_FILE > /dev/null
  read_time=$(($(date +%s%N) - start))
  
  # Modify file
  start=$(date +%s%N)
  echo "Modified" >> $TEST_FILE
  modify_time=$(($(date +%s%N) - start))
  
  # Delete file
  start=$(date +%s%N)
  rm $TEST_FILE
  delete_time=$(($(date +%s%N) - start))
  
  echo "$mode: Create=${create_time}ns Read=${read_time}ns Modify=${modify_time}ns Delete=${delete_time}ns"
}

# Test with D-Bus
fusermount3 -uz $MOUNT_POINT 2>/dev/null
onemount $MOUNT_POINT &
sleep 2
test_operations "D-Bus Enabled"
fusermount3 -uz $MOUNT_POINT

# Test without D-Bus
unset DBUS_SESSION_BUS_ADDRESS
onemount $MOUNT_POINT &
sleep 2
test_operations "D-Bus Disabled"
fusermount3 -uz $MOUNT_POINT
```

**Steps**:

1. **Run comparison script**:
   ```bash
   chmod +x compare-dbus-modes.sh
   ./compare-dbus-modes.sh
   ```

2. **Analyze results**:
   - Compare operation times
   - Check for functional differences
   - Verify both modes work correctly

**Expected Results**:
- ✅ Both modes complete all operations successfully
- ✅ Performance difference is minimal (< 10%)
- ✅ No functional differences
- ✅ Both modes produce same end result

**Pass Criteria**:
- All operations succeed in both modes
- Performance degradation < 10%
- No functional differences
- Files sync correctly in both modes

---


## Troubleshooting

### Issue: OneMount fails to start without D-Bus

**Symptoms**:
- Mount operation fails
- Error messages about D-Bus
- Process exits immediately

**Diagnosis**:
```bash
# Check if D-Bus is truly unavailable
echo $DBUS_SESSION_BUS_ADDRESS

# Check OneMount logs
grep -i dbus /tmp/onemount-no-dbus.log
```

**Solutions**:
1. Verify D-Bus environment variable is unset
2. Check if OneMount has hard dependency on D-Bus (bug)
3. Review error logs for actual failure cause
4. Try alternative D-Bus disable methods

---

### Issue: Extended attributes not working

**Symptoms**:
- `getfattr` returns no attributes
- Status queries fail
- File operations work but status unavailable

**Diagnosis**:
```bash
# Check if filesystem supports xattrs
touch /tmp/test-xattr
setfattr -n user.test -v "value" /tmp/test-xattr
getfattr -n user.test /tmp/test-xattr
rm /tmp/test-xattr
```

**Solutions**:
1. Verify filesystem supports extended attributes
2. Check mount options include `user_xattr`
3. Verify OneMount is setting xattrs (check logs)
4. Test with different filesystem if needed

---

### Issue: Performance degradation without D-Bus

**Symptoms**:
- Operations slower than with D-Bus
- Noticeable lag in file operations
- High CPU usage

**Diagnosis**:
```bash
# Monitor OneMount process
top -p $(pgrep onemount)

# Check for excessive logging
wc -l /tmp/onemount-no-dbus.log
```

**Solutions**:
1. Reduce log level if excessive logging
2. Check for polling loops or retries
3. Verify no D-Bus reconnection attempts
4. Profile code to identify bottleneck

---

### Issue: D-Bus still being used despite unsetting variable

**Symptoms**:
- D-Bus signals still emitted
- Service registered on D-Bus
- Logs show D-Bus connection succeeded

**Diagnosis**:
```bash
# Check environment
env | grep DBUS

# Check if D-Bus is available via other means
dbus-send --session --print-reply --dest=org.freedesktop.DBus /org/freedesktop/DBus org.freedesktop.DBus.ListNames
```

**Solutions**:
1. Verify environment variable is unset in correct shell
2. Use more aggressive D-Bus disable method
3. Check if OneMount has fallback D-Bus discovery
4. Stop D-Bus service entirely for testing

---

## How to Verify D-Bus is Actually Disabled

### Method 1: Check Environment

```bash
# Should be empty
echo $DBUS_SESSION_BUS_ADDRESS
```

### Method 2: Try D-Bus Command

```bash
# Should fail
dbus-send --session --print-reply --dest=org.freedesktop.DBus /org/freedesktop/DBus org.freedesktop.DBus.ListNames
# Expected: "Failed to connect to socket"
```

### Method 3: Check OneMount Logs

```bash
# Should see connection failure
grep "Failed to connect to D-Bus" /tmp/onemount-no-dbus.log
```

### Method 4: Monitor D-Bus Traffic

```bash
# Should show no traffic from OneMount
dbus-monitor --session &
# Launch OneMount
# Should see no messages from OneMount
```

---

## Results Documentation Template

### Test Execution Record

**Date**: _______________  
**Tester**: _______________  
**OneMount Version**: _______________  
**System**: _______________  
**D-Bus Disable Method**: _______________  

### Test Results

| Test # | Test Name | Status | Notes |
|--------|-----------|--------|-------|
| 1 | Mount Without D-Bus | ☐ Pass ☐ Fail | |
| 2 | Core File Operations | ☐ Pass ☐ Fail | |
| 3 | Extended Attributes | ☐ Pass ☐ Fail | |
| 4 | Graceful Degradation | ☐ Pass ☐ Fail | |
| 5 | Status Reporting | ☐ Pass ☐ Fail | |
| 6 | Log Messages | ☐ Pass ☐ Fail | |
| 7 | Comparison Testing | ☐ Pass ☐ Fail | |

### Performance Comparison

| Operation | D-Bus Enabled | D-Bus Disabled | Difference |
|-----------|---------------|----------------|------------|
| File Create | _____ ms | _____ ms | _____ % |
| File Read | _____ ms | _____ ms | _____ % |
| File Modify | _____ ms | _____ ms | _____ % |
| File Delete | _____ ms | _____ ms | _____ % |

### Functional Differences Observed

| Feature | D-Bus Enabled | D-Bus Disabled | Impact |
|---------|---------------|----------------|--------|
| File Status | | | |
| Sync Operations | | | |
| Error Handling | | | |
| Logging | | | |

### Issues Found

| Issue # | Description | Severity | Reproducible |
|---------|-------------|----------|--------------|
| | | | |

### Overall Assessment

☐ **PASS** - System functions fully without D-Bus  
☐ **FAIL** - Critical functionality lost without D-Bus  
☐ **PARTIAL** - Minor issues but core functionality works  

**Comments**: _______________________________________________

---

## Pass/Fail Criteria

### Overall Pass Criteria

The D-Bus fallback test suite **PASSES** if:

1. ✅ OneMount mounts successfully without D-Bus
2. ✅ All core file operations work (read, write, list, delete)
3. ✅ Extended attributes provide file status information
4. ✅ No crashes, hangs, or critical errors
5. ✅ Performance degradation is minimal (< 10%)
6. ✅ Appropriate log messages indicate D-Bus unavailability
7. ✅ System continues operating normally

### Critical Failures

The test suite **FAILS** if:

1. ❌ OneMount fails to mount without D-Bus
2. ❌ Core file operations fail or error
3. ❌ System crashes or hangs
4. ❌ File status is completely unavailable
5. ❌ Significant performance degradation (> 25%)
6. ❌ Error-level logs for D-Bus unavailability
7. ❌ Data loss or corruption occurs

### Acceptable Limitations

The following are acceptable when D-Bus is unavailable:

- ✅ No real-time status signals to external applications
- ✅ Nemo/Nautilus extensions don't receive updates
- ✅ Status must be queried via xattrs instead of D-Bus
- ✅ Slightly increased latency for status queries (< 10%)

### Unacceptable Limitations

The following are NOT acceptable:

- ❌ Complete loss of file status information
- ❌ File operations fail or error
- ❌ System becomes unstable
- ❌ Significant performance impact
- ❌ Data synchronization issues

---

## Known Limitations and Edge Cases

### Limitation 1: No Real-Time Notifications

**Description**: Without D-Bus, external applications cannot receive real-time file status updates.

**Impact**: File manager extensions won't update icons automatically.

**Workaround**: Applications must poll extended attributes for status.

**Acceptable**: Yes - this is expected behavior.

---

### Limitation 2: Service Discovery

**Description**: Without D-Bus, clients cannot discover OneMount service automatically.

**Impact**: Extensions must use alternative discovery methods (e.g., mount point detection).

**Workaround**: Check for OneMount mount points directly.

**Acceptable**: Yes - D-Bus is optional for discovery.

---

### Limitation 3: Method Calls

**Description**: D-Bus method calls (like `GetFileStatus`) are unavailable.

**Impact**: Clients must use extended attributes instead.

**Workaround**: Use `getfattr` to query status.

**Acceptable**: Yes - xattrs provide equivalent functionality.

---

## Additional Resources

### OneMount D-Bus Implementation

- Source: `internal/fs/dbus.go`
- Fallback logic: `internal/fs/file_status.go`
- Extended attributes: `internal/fs/xattr.go`

### Related Documentation

- [D-Bus Integration Guide](./manual-dbus-integration-guide.md)
- [Extended Attributes Documentation](../2-architecture/extended-attributes.md)
- [File Status System](../2-architecture/file-status.md)

### Related Requirements

- Requirement 10.1: Extended attribute updates
- Requirement 10.2: D-Bus signal emission
- Requirement 10.4: D-Bus fallback behavior
- Requirement 11.1: Error handling and logging

---

## Conclusion

This guide provides comprehensive testing procedures for verifying OneMount's D-Bus fallback behavior. The system must continue operating normally when D-Bus is unavailable, using extended attributes as the fallback mechanism for file status information.

**Key Takeaway**: D-Bus is an optional enhancement, not a requirement. OneMount must function fully without it.
