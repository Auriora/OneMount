# Manual Verification: D-Bus GetFileStatus Fix (Task 41)

## Overview

This document provides instructions for manually verifying the fix for Issue #FS-001: D-Bus GetFileStatus returns Unknown.

**Issue**: The D-Bus `GetFileStatus` method was returning "Unknown" for all file paths because it relied on complex path traversal logic that was fragile and error-prone.

**Fix**: Added a `GetIDByPath(path string) string` method to the `FilesystemInterface` that properly converts filesystem paths to OneDrive IDs. The D-Bus `GetFileStatus` method now uses this new method instead of the fragile path traversal.

## Prerequisites

- OneMount installed and configured
- Nemo file manager with OneMount extension installed
- OneDrive account with test files
- D-Bus session bus available (not in Docker)

## Verification Steps

### 1. Mount OneMount Filesystem

```bash
# Mount your OneDrive
onemount /path/to/mount/point

# Verify mount is successful
mount | grep onemount
```

### 2. Verify D-Bus Service is Running

```bash
# Check if D-Bus service is registered
dbus-send --session --print-reply \
  --dest=org.freedesktop.DBus \
  /org/freedesktop/DBus \
  org.freedesktop.DBus.ListNames | grep onemount

# Expected output should include something like:
# string "org.onemount.FileStatus.mnt_<escaped_mount_path>"
```

### 3. Test GetFileStatus with Various Paths

```bash
# Test with root directory
dbus-send --session --print-reply \
  --dest=org.onemount.FileStatus.mnt_<escaped_mount_path> \
  /org/onemount/FileStatus \
  org.onemount.FileStatus.GetFileStatus \
  string:"/"

# Expected: Should return a status (e.g., "Local", "Cloud", etc.), NOT "Unknown"

# Test with a file in root
dbus-send --session --print-reply \
  --dest=org.onemount.FileStatus.mnt_<escaped_mount_path> \
  /org/onemount/FileStatus \
  org.onemount.FileStatus.GetFileStatus \
  string:"/Documents"

# Expected: Should return actual status, NOT "Unknown"

# Test with nested path
dbus-send --session --print-reply \
  --dest=org.onemount.FileStatus.mnt_<escaped_mount_path> \
  /org/onemount/FileStatus \
  org.onemount.FileStatus.GetFileStatus \
  string:"/Documents/Work/report.pdf"

# Expected: Should return actual status, NOT "Unknown"

# Test with non-existent path
dbus-send --session --print-reply \
  --dest=org.onemount.FileStatus.mnt_<escaped_mount_path> \
  /org/onemount/FileStatus \
  org.onemount.FileStatus.GetFileStatus \
  string:"/NonExistent/file.txt"

# Expected: Should return "Unknown" (this is correct for non-existent files)
```

### 4. Verify Status Icons in Nemo

1. Open Nemo file manager
2. Navigate to the OneMount mount point
3. Observe file status icons:
   - Files should show appropriate status icons (synced, downloading, error, etc.)
   - Icons should NOT all show "Unknown" status
4. Perform file operations and verify icons update:
   - Download a file (should show downloading icon, then synced)
   - Modify a file (should show modified icon)
   - Create a new file (should show uploading icon, then synced)

### 5. Test with Various File States

Create test scenarios to verify different file states:

```bash
# 1. Cloud-only file (not downloaded)
# - Navigate to a file in Nemo
# - Verify icon shows cloud status
# - Call GetFileStatus via D-Bus
# - Expected: "Cloud" status

# 2. Downloaded file
# - Open a file to trigger download
# - Wait for download to complete
# - Verify icon shows local/synced status
# - Call GetFileStatus via D-Bus
# - Expected: "Local" status

# 3. Modified file
# - Edit a file and save changes
# - Verify icon shows modified status
# - Call GetFileStatus via D-Bus
# - Expected: "LocalModified" or "Syncing" status

# 4. File with error
# - Simulate an error condition (e.g., network disconnect during upload)
# - Verify icon shows error status
# - Call GetFileStatus via D-Bus
# - Expected: "Error" status
```

### 6. Test with Special Characters in Paths

```bash
# Test with spaces
dbus-send --session --print-reply \
  --dest=org.onemount.FileStatus.mnt_<escaped_mount_path> \
  /org/onemount/FileStatus \
  org.onemount.FileStatus.GetFileStatus \
  string:"/Documents/file with spaces.txt"

# Test with special characters
dbus-send --session --print-reply \
  --dest=org.onemount.FileStatus.mnt_<escaped_mount_path> \
  /org/onemount/FileStatus \
  org.onemount.FileStatus.GetFileStatus \
  string:"/Documents/file-with-dashes.txt"
```

## Expected Results

### Success Criteria

1. ✅ D-Bus service is registered and accessible
2. ✅ GetFileStatus returns actual status for existing files (NOT "Unknown")
3. ✅ GetFileStatus returns "Unknown" only for non-existent files
4. ✅ Status icons in Nemo display correctly for all file states
5. ✅ Status icons update correctly when file state changes
6. ✅ GetFileStatus works with paths containing special characters
7. ✅ GetFileStatus works with deeply nested paths

### Failure Indicators

- ❌ GetFileStatus returns "Unknown" for existing files
- ❌ Status icons in Nemo show "Unknown" for all files
- ❌ Status icons don't update when file state changes
- ❌ GetFileStatus fails with special characters in paths
- ❌ D-Bus service is not registered

## Troubleshooting

### D-Bus Service Not Found

```bash
# Check if OneMount is running
ps aux | grep onemount

# Check D-Bus service name file
cat /tmp/onemount-dbus-service-name

# Check OneMount logs for D-Bus errors
journalctl -u onemount --since "1 hour ago" | grep -i dbus
```

### GetFileStatus Returns "Unknown" for Existing Files

```bash
# Check if filesystem is properly mounted
ls -la /path/to/mount/point

# Verify file exists in filesystem
stat /path/to/mount/point/path/to/file

# Check OneMount logs for path resolution errors
journalctl -u onemount --since "1 hour ago" | grep -i "GetIDByPath\|GetFileStatus"
```

### Status Icons Not Updating in Nemo

```bash
# Restart Nemo
nemo -q
nemo &

# Check if Nemo extension is loaded
ls ~/.local/share/nemo/extensions/

# Check Nemo logs
journalctl --user -u nemo --since "1 hour ago"
```

## Documentation

After completing verification, document the results in `docs/verification-tracking.md` under Phase 11 (File Status and D-Bus Verification):

```markdown
### Issue #FS-001: D-Bus GetFileStatus Returns Unknown

**Status**: ✅ VERIFIED / ❌ FAILED

**Verification Date**: YYYY-MM-DD

**Verification Results**:
- [ ] D-Bus service registered successfully
- [ ] GetFileStatus returns correct status for existing files
- [ ] GetFileStatus returns "Unknown" for non-existent files
- [ ] Status icons display correctly in Nemo
- [ ] Status icons update on file state changes
- [ ] Special characters in paths handled correctly

**Notes**:
[Add any observations, issues, or additional notes here]
```

## Related Requirements

- **Requirement 8.2**: D-Bus integration for file status updates
- **Requirement 8.3**: Nemo extension integration

## Related Files

- `internal/fs/dbus.go` - D-Bus server implementation
- `internal/fs/cache.go` - GetIDByPath implementation
- `internal/fs/filesystem_types.go` - FilesystemInterface definition
- `internal/fs/dbus_getfilestatus_test.go` - Integration tests
