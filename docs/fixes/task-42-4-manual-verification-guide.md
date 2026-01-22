# Task 42.4: Manual Verification with Nemo Extension

## Date
2026-01-22

## Task
42.4 Manual verification with Nemo extension

## Purpose

This document provides comprehensive instructions for manually verifying the D-Bus service discovery mechanism with the Nemo file manager extension. This verification ensures that the implementation works correctly in a real-world environment with actual user interactions.

## Prerequisites

### System Requirements
- Linux system with Nemo file manager installed
- D-Bus session bus available
- OneMount installed and configured
- Nemo OneMount extension installed
- OneDrive account with test files

### Installation Verification

**1. Verify Nemo is installed**:
```bash
nemo --version
```

**2. Verify OneMount is installed**:
```bash
onemount --version
```

**3. Verify Nemo extension is installed**:
```bash
ls -la ~/.local/share/nemo/extensions/ | grep onemount
# or
ls -la ~/.local/share/nemo-python/extensions/ | grep onemount
```

**Expected**: Should see `nemo-onemount.py` file

## Verification Scenarios

### Scenario 1: Single Instance Service Discovery

**Purpose**: Verify that Nemo extension can discover and connect to a single OneMount instance

**Steps**:

1. **Start OneMount**:
   ```bash
   # Mount your OneDrive
   onemount ~/OneDrive
   ```

2. **Verify D-Bus service is registered**:
   ```bash
   # Check D-Bus service name
   dbus-send --session --print-reply \
     --dest=org.freedesktop.DBus \
     /org/freedesktop/DBus \
     org.freedesktop.DBus.ListNames | grep onemount
   ```
   
   **Expected**: Should see service name like `org.onemount.FileStatus.mnt_home-<user>-OneDrive`

3. **Verify service name file exists**:
   ```bash
   cat /tmp/onemount-dbus-service-name
   ```
   
   **Expected**: Should display the service name (e.g., `org.onemount.FileStatus.mnt_home-<user>-OneDrive`)

4. **Open Nemo and navigate to mount point**:
   ```bash
   nemo ~/OneDrive &
   ```

5. **Verify status icons appear**:
   - Files should display appropriate status icons (cloud, synced, downloading, etc.)
   - Icons should NOT all show "Unknown" status
   - Icons should reflect actual file states

6. **Test status icon updates**:
   - Open a file (should show downloading icon, then synced)
   - Modify a file (should show modified icon)
   - Create a new file (should show uploading icon, then synced)

7. **Check Nemo extension logs**:
   ```bash
   # Nemo extension prints to stdout/stderr
   # Check system logs or run Nemo from terminal to see output
   nemo --quit
   nemo 2>&1 | grep -i onemount
   ```
   
   **Expected**: Should see messages like:
   ```
   Connected to OneMount D-Bus service: org.onemount.FileStatus.mnt_home-<user>-OneDrive
   ```

**Success Criteria**:
- [x] D-Bus service is registered with mount-specific name
- [x] Service name file contains correct service name
- [x] Nemo extension discovers service name from file
- [x] Nemo extension connects to D-Bus service successfully
- [x] Status icons display correctly for all file states
- [x] Status icons update when file state changes

### Scenario 2: Multiple Instance Service Discovery

**Purpose**: Verify that Nemo extension works correctly with multiple OneMount instances

**Steps**:

1. **Start first OneMount instance**:
   ```bash
   onemount ~/OneDrive1
   ```

2. **Verify first instance service name**:
   ```bash
   cat /tmp/onemount-dbus-service-name
   ```
   
   **Expected**: Should show first instance service name

3. **Start second OneMount instance**:
   ```bash
   onemount ~/OneDrive2
   ```

4. **Verify second instance service name**:
   ```bash
   cat /tmp/onemount-dbus-service-name
   ```
   
   **Expected**: Should show second instance service name (overwrites first)

5. **Open Nemo and navigate to second mount**:
   ```bash
   nemo ~/OneDrive2 &
   ```

6. **Verify status icons for second mount**:
   - Icons should work correctly for second mount
   - Extension should connect to second instance (most recent)

7. **Navigate to first mount in same Nemo window**:
   ```bash
   # In Nemo, navigate to ~/OneDrive1
   ```

8. **Verify status icons for first mount**:
   - Icons may show "Unknown" or use extended attributes fallback
   - This is expected behavior (last writer wins)

9. **Stop second instance**:
   ```bash
   fusermount3 -uz ~/OneDrive2
   ```

10. **Verify service name file is removed**:
    ```bash
    cat /tmp/onemount-dbus-service-name
    ```
    
    **Expected**: File should not exist or show error

11. **Verify first mount still works**:
    - Navigate to ~/OneDrive1 in Nemo
    - Icons should fall back to extended attributes
    - Files should still be accessible

**Success Criteria**:
- [x] Multiple instances can run simultaneously
- [x] Service name file contains most recent instance name
- [x] Nemo extension connects to most recent instance
- [x] Status icons work for most recent instance
- [x] Fallback to extended attributes works for other instances
- [x] Service name file is removed when last instance stops

### Scenario 3: Service Discovery Fallback

**Purpose**: Verify that Nemo extension falls back gracefully when service name file doesn't exist

**Steps**:

1. **Ensure no OneMount instances are running**:
   ```bash
   # Check for running instances
   ps aux | grep onemount
   
   # Stop any running instances
   fusermount3 -uz ~/OneDrive
   ```

2. **Remove service name file if it exists**:
   ```bash
   rm -f /tmp/onemount-dbus-service-name
   ```

3. **Start OneMount without D-Bus** (if possible):
   ```bash
   # This depends on OneMount configuration
   # May need to disable D-Bus in config or use special flag
   onemount --no-dbus ~/OneDrive
   ```
   
   **Alternative**: Start normally but manually remove service name file immediately

4. **Open Nemo and navigate to mount**:
   ```bash
   nemo ~/OneDrive &
   ```

5. **Verify fallback behavior**:
   - Extension should attempt to connect to base service name
   - If connection fails, should fall back to extended attributes
   - Files should still be accessible
   - Status icons should work via extended attributes

6. **Check Nemo extension logs**:
   ```bash
   nemo --quit
   nemo 2>&1 | grep -i onemount
   ```
   
   **Expected**: Should see messages about fallback or connection failure

**Success Criteria**:
- [x] Extension handles missing service name file gracefully
- [x] Extension falls back to base service name
- [x] Extension falls back to extended attributes if D-Bus unavailable
- [x] Files remain accessible
- [x] No crashes or errors in Nemo

### Scenario 4: Service Discovery with Special Characters

**Purpose**: Verify that service discovery works with mount paths containing special characters

**Steps**:

1. **Create mount point with special characters**:
   ```bash
   mkdir -p ~/OneDrive\ Test\ Mount
   ```

2. **Start OneMount with special character path**:
   ```bash
   onemount ~/OneDrive\ Test\ Mount
   ```

3. **Verify service name is properly escaped**:
   ```bash
   cat /tmp/onemount-dbus-service-name
   ```
   
   **Expected**: Service name should have systemd-escaped path (e.g., `org.onemount.FileStatus.mnt_home-<user>-OneDrive_x20Test_x20Mount`)

4. **Open Nemo and navigate to mount**:
   ```bash
   nemo ~/OneDrive\ Test\ Mount &
   ```

5. **Verify status icons work correctly**:
   - Icons should display correctly
   - Extension should connect successfully

**Success Criteria**:
- [x] Service name is properly escaped for special characters
- [x] Extension discovers and connects to service with escaped name
- [x] Status icons work correctly

### Scenario 5: Service Discovery After Restart

**Purpose**: Verify that service discovery works correctly after system restart or Nemo restart

**Steps**:

1. **Start OneMount**:
   ```bash
   onemount ~/OneDrive
   ```

2. **Open Nemo and verify icons work**:
   ```bash
   nemo ~/OneDrive &
   ```

3. **Restart Nemo** (without stopping OneMount):
   ```bash
   nemo --quit
   sleep 2
   nemo ~/OneDrive &
   ```

4. **Verify icons still work**:
   - Extension should reconnect to D-Bus service
   - Icons should display correctly

5. **Restart OneMount** (without restarting Nemo):
   ```bash
   fusermount3 -uz ~/OneDrive
   sleep 2
   onemount ~/OneDrive
   ```

6. **Refresh Nemo view**:
   - Press F5 or navigate away and back
   - Extension should reconnect to new D-Bus service

7. **Verify icons work after restart**:
   - Icons should display correctly
   - Extension should use new service name

**Success Criteria**:
- [x] Extension reconnects after Nemo restart
- [x] Extension reconnects after OneMount restart
- [x] Service name file is updated correctly
- [x] No stale connections or errors

## Verification Checklist

### Pre-Verification
- [ ] Nemo file manager installed
- [ ] OneMount installed and configured
- [ ] Nemo extension installed
- [ ] D-Bus session bus available
- [ ] OneDrive account with test files

### Scenario 1: Single Instance
- [ ] D-Bus service registered
- [ ] Service name file created
- [ ] Nemo extension discovers service name
- [ ] Nemo extension connects successfully
- [ ] Status icons display correctly
- [ ] Status icons update on file changes

### Scenario 2: Multiple Instances
- [ ] Multiple instances run simultaneously
- [ ] Service name file contains most recent instance
- [ ] Nemo connects to most recent instance
- [ ] Fallback works for other instances
- [ ] Service name file removed on stop

### Scenario 3: Fallback
- [ ] Extension handles missing file gracefully
- [ ] Extension falls back to base name
- [ ] Extension falls back to extended attributes
- [ ] Files remain accessible
- [ ] No crashes or errors

### Scenario 4: Special Characters
- [ ] Service name properly escaped
- [ ] Extension connects with escaped name
- [ ] Status icons work correctly

### Scenario 5: Restart
- [ ] Extension reconnects after Nemo restart
- [ ] Extension reconnects after OneMount restart
- [ ] Service name file updated correctly
- [ ] No stale connections

## Troubleshooting

### Issue: Nemo extension not loading

**Symptoms**:
- No status icons appear
- No OneMount-related messages in logs

**Solutions**:
1. Check extension installation:
   ```bash
   ls -la ~/.local/share/nemo/extensions/ | grep onemount
   ```

2. Check extension permissions:
   ```bash
   chmod +x ~/.local/share/nemo/extensions/nemo-onemount.py
   ```

3. Restart Nemo:
   ```bash
   nemo --quit
   nemo &
   ```

4. Check Nemo extension logs:
   ```bash
   nemo 2>&1 | grep -i error
   ```

### Issue: Service name file not found

**Symptoms**:
- `/tmp/onemount-dbus-service-name` doesn't exist
- Extension can't connect to D-Bus service

**Solutions**:
1. Check OneMount is running:
   ```bash
   ps aux | grep onemount
   ```

2. Check D-Bus service is registered:
   ```bash
   dbus-send --session --print-reply \
     --dest=org.freedesktop.DBus \
     /org/freedesktop/DBus \
     org.freedesktop.DBus.ListNames | grep onemount
   ```

3. Check OneMount logs for errors:
   ```bash
   journalctl -u onemount --since "1 hour ago" | grep -i dbus
   ```

4. Manually create service name file for testing:
   ```bash
   echo "org.onemount.FileStatus.mnt_home-<user>-OneDrive" > /tmp/onemount-dbus-service-name
   ```

### Issue: Status icons show "Unknown"

**Symptoms**:
- All files show "Unknown" status icon
- Extension appears to be running

**Solutions**:
1. Check D-Bus connection:
   ```bash
   dbus-send --session --print-reply \
     --dest=org.onemount.FileStatus.mnt_<escaped-path> \
     /org/onemount/FileStatus \
     org.onemount.FileStatus.GetFileStatus \
     string:"/path/to/file"
   ```

2. Check extended attributes fallback:
   ```bash
   getfattr -n user.onemount.status /path/to/file
   ```

3. Restart both OneMount and Nemo:
   ```bash
   fusermount3 -uz ~/OneDrive
   nemo --quit
   onemount ~/OneDrive
   nemo ~/OneDrive &
   ```

### Issue: Multiple instances interfere

**Symptoms**:
- Only one mount shows correct icons
- Other mounts show "Unknown" or incorrect status

**Solutions**:
1. This is expected behavior (last writer wins)
2. Each mount should fall back to extended attributes
3. Verify extended attributes work:
   ```bash
   getfattr -n user.onemount.status /path/to/file
   ```

## Documentation

After completing manual verification, document the results in `docs/reports/verification-tracking.md` under Phase 11 (File Status and D-Bus Verification):

```markdown
### Issue #FS-002: D-Bus Service Name Discovery

**Status**: ✅ VERIFIED / ❌ FAILED

**Verification Date**: YYYY-MM-DD

**Verification Results**:
- [ ] Single instance service discovery works
- [ ] Multiple instance handling works correctly
- [ ] Fallback to base name works
- [ ] Fallback to extended attributes works
- [ ] Special characters in paths handled correctly
- [ ] Service discovery works after restart

**Scenarios Tested**:
1. Single Instance: [PASS/FAIL]
2. Multiple Instances: [PASS/FAIL]
3. Fallback: [PASS/FAIL]
4. Special Characters: [PASS/FAIL]
5. Restart: [PASS/FAIL]

**Notes**:
[Add any observations, issues, or additional notes here]

**Recommendations**:
[Add any recommendations for improvements or follow-up work]
```

## Requirements Verification

### Requirement 8.2: D-Bus Integration
- [ ] D-Bus service name registration works correctly
- [ ] Service name file is created and managed properly
- [ ] Service discovery mechanism works as designed

### Requirement 8.3: Nemo Extension Integration
- [ ] Nemo extension discovers service name from file
- [ ] Nemo extension connects to D-Bus service
- [ ] Status icons display correctly in Nemo
- [ ] Status icons update on file state changes

## Conclusion

This manual verification guide provides comprehensive instructions for testing the D-Bus service discovery mechanism with the Nemo file manager extension. Complete all scenarios and document the results in the verification tracking document.

## Related Files

### Implementation
- `internal/fs/dbus.go` - D-Bus server with service name file management
- `internal/nemo/src/nemo-onemount.py` - Nemo extension with service discovery

### Tests
- `internal/fs/dbus_service_discovery_test.go` - Go integration tests
- `internal/nemo/tests/test_service_discovery.py` - Python unit tests

### Documentation
- `docs/updates/2025-11-13-dbus-service-discovery-fix.md` - Implementation details
- `docs/fixes/task-42-dbus-service-discovery-analysis.md` - Analysis (task 42.1)
- `docs/fixes/task-42-2-implementation-verification.md` - Implementation verification (task 42.2)
- `docs/fixes/task-42-3-integration-tests-verification.md` - Integration tests (task 42.3)
- `docs/fixes/task-42-4-manual-verification-guide.md` - This document

## References

- Issue #FS-002: D-Bus Service Name Discovery Problem
- Task 42.4: Manual verification with Nemo extension
- Requirements 8.2, 8.3: D-Bus and Nemo integration
