# Manual Crash Recovery Testing Guide

**Date**: 2025-11-11  
**Purpose**: Manual testing procedures for verifying crash recovery in OneMount  
**Related Task**: 14.4 - Test crash recovery

## Overview

This document provides manual testing procedures to verify that OneMount correctly recovers from crashes, process kills, and unexpected shutdowns, including state recovery from the bbolt database.

## Prerequisites

1. OneMount installed and configured
2. Valid OneDrive authentication
3. Test OneDrive account with test files
4. Root/sudo access for process management
5. Ability to monitor system logs

## Understanding OneMount State Persistence

### State Storage

OneMount uses bbolt (embedded key-value database) to persist state:

**Location**: `~/.local/share/onemount/onemount.db` (or `$XDG_DATA_HOME/onemount/onemount.db`)

**Persisted Data**:
1. **Upload Sessions** (`bucketUploads`):
   - Session ID, file ID, file name
   - Upload state (not started, in progress, completed, errored)
   - Progress tracking (bytes uploaded, last successful chunk)
   - Recovery metadata (can resume, last progress time, recovery attempts)
   - Upload URL and expiration

2. **Delta Sync State** (`bucketDelta`):
   - Delta link (for incremental sync)
   - Last sync timestamp
   - Sync state

3. **File Metadata** (in-memory, reconstructed from API):
   - File/folder hierarchy
   - Inode mappings
   - File attributes

### Recovery Strategy

**On Startup**:
1. Load persisted upload sessions from database
2. Cancel incomplete uploads (currently non-resumable per code comment)
3. Reconstruct file system state from OneDrive API
4. Resume delta sync from last delta link

**During Operation**:
1. Persist upload progress after each chunk
2. Persist delta link after each sync
3. Handle graceful shutdown (SIGTERM, SIGINT, SIGHUP)
4. Force shutdown after 30-second timeout

## Test Scenarios

### Scenario 1: Graceful Shutdown During Upload

**Objective**: Verify that graceful shutdown persists upload state correctly

**Steps**:
1. Mount OneMount filesystem
2. Start uploading a large file (>100MB):
   ```bash
   dd if=/dev/urandom of=/tmp/testfile.bin bs=1M count=100
   cp /tmp/testfile.bin /mnt/onedrive/ &
   ```
3. Wait for upload to start (check logs)
4. Send SIGTERM signal:
   ```bash
   sudo killall -TERM onemount
   ```
5. Wait for graceful shutdown (max 30 seconds)
6. Check database for persisted state
7. Remount filesystem
8. Verify upload session state

**Expected Results**:
- ✅ SIGTERM signal received
- ✅ Upload progress persisted to database
- ✅ Graceful shutdown completes within 30 seconds
- ✅ Database contains upload session
- ✅ On remount, upload session is loaded
- ✅ Upload session is cancelled (non-resumable)
- ✅ File can be re-uploaded successfully

**Verification**:
```bash
# Check for graceful shutdown
journalctl -u onemount | grep -i "shutdown\|SIGTERM"

# Check database contents
sqlite3 ~/.local/share/onemount/onemount.db ".dump" 2>/dev/null || \
  echo "Database is bbolt format, use Go tool to inspect"

# Check upload persistence
journalctl -u onemount | grep "Persisting upload progress"

# Verify remount
journalctl -u onemount | grep "restoring upload sessions"
```

---

### Scenario 2: Forced Kill During Upload (SIGKILL)

**Objective**: Verify recovery from forced process termination

**Steps**:
1. Mount OneMount filesystem
2. Start uploading a large file:
   ```bash
   dd if=/dev/urandom of=/tmp/testfile.bin bs=1M count=100
   cp /tmp/testfile.bin /mnt/onedrive/ &
   ```
3. Wait for upload to be in progress
4. Force kill the process:
   ```bash
   sudo killall -9 onemount
   ```
5. Verify process is killed immediately
6. Remount filesystem
7. Check upload session recovery

**Expected Results**:
- ✅ Process killed immediately (no graceful shutdown)
- ✅ Last persisted state available in database
- ✅ On remount, upload session loaded from database
- ✅ Upload session cancelled (non-resumable)
- ✅ No database corruption
- ✅ File system state reconstructed from API
- ✅ File can be re-uploaded successfully

**Verification**:
```bash
# Verify process killed
ps aux | grep onemount

# Check for database corruption
# Remount should succeed without errors
onemount mount /mnt/onedrive

# Check logs for session restoration
journalctl -u onemount | grep "restoring upload sessions"

# Verify file system accessible
ls -la /mnt/onedrive/
```

---

### Scenario 3: System Crash During Upload

**Objective**: Verify recovery from system crash or power loss

**Steps**:
1. Mount OneMount filesystem
2. Start uploading multiple files:
   ```bash
   for i in {1..5}; do
     dd if=/dev/urandom of=/tmp/testfile$i.bin bs=1M count=50
     cp /tmp/testfile$i.bin /mnt/onedrive/ &
   done
   ```
3. Simulate system crash:
   ```bash
   # WARNING: This will reboot the system!
   # Only do this on a test system
   sudo sync
   sudo reboot -f
   ```
4. After reboot, remount filesystem
5. Check upload session recovery

**Expected Results**:
- ✅ Database survives system crash
- ✅ All upload sessions loaded on remount
- ✅ Upload sessions cancelled (non-resumable)
- ✅ No data corruption
- ✅ Files can be re-uploaded successfully

**Verification**:
```bash
# After reboot, check database integrity
onemount mount /mnt/onedrive

# Check logs
journalctl -u onemount | grep "restoring upload sessions"

# Verify file system state
ls -la /mnt/onedrive/
```

---

### Scenario 4: Crash During Delta Sync

**Objective**: Verify delta sync recovery after crash

**Steps**:
1. Mount OneMount filesystem
2. Trigger delta sync by modifying files on OneDrive
3. During sync, kill the process:
   ```bash
   sudo killall -9 onemount
   ```
4. Remount filesystem
5. Verify delta sync resumes from last delta link

**Expected Results**:
- ✅ Delta link persisted before crash
- ✅ On remount, delta sync resumes from last link
- ✅ No duplicate sync operations
- ✅ File system state consistent with OneDrive

**Verification**:
```bash
# Check delta link persistence
journalctl -u onemount | grep "delta"

# Verify sync resumes
journalctl -u onemount | grep "delta sync"

# Check file system consistency
diff -r /mnt/onedrive/ /path/to/onedrive/web/
```

---

### Scenario 5: Database Corruption Recovery

**Objective**: Verify handling of database corruption

**Steps**:
1. Unmount OneMount filesystem
2. Corrupt the database:
   ```bash
   # Backup first!
   cp ~/.local/share/onemount/onemount.db ~/.local/share/onemount/onemount.db.backup
   
   # Corrupt the database
   dd if=/dev/urandom of=~/.local/share/onemount/onemount.db bs=1K count=1 seek=10 conv=notrunc
   ```
3. Attempt to mount filesystem
4. Observe error handling

**Expected Results**:
- ✅ Database corruption detected
- ✅ Clear error message logged
- ✅ Mount fails gracefully (or creates new database)
- ✅ User notified of corruption
- ✅ Backup/recovery instructions provided

**Verification**:
```bash
# Attempt mount
onemount mount /mnt/onedrive 2>&1 | tee /tmp/mount-error.log

# Check error logs
journalctl -u onemount | grep -i "error\|corrupt"

# Restore backup if needed
mv ~/.local/share/onemount/onemount.db.backup ~/.local/share/onemount/onemount.db
```

---

### Scenario 6: Multiple Crashes During Upload

**Objective**: Verify recovery from repeated crashes

**Steps**:
1. Mount OneMount filesystem
2. Start uploading a large file
3. Kill process after 10 seconds:
   ```bash
   sleep 10 && sudo killall -9 onemount &
   ```
4. Remount and restart upload
5. Kill process again after 10 seconds
6. Repeat 3 times
7. Finally allow upload to complete

**Expected Results**:
- ✅ Each crash persists current state
- ✅ Each remount loads persisted state
- ✅ Upload eventually completes
- ✅ No database corruption from repeated crashes
- ✅ Recovery attempts tracked correctly

**Verification**:
```bash
# Check recovery attempts
journalctl -u onemount | grep "recoveryAttempts"

# Verify upload completion
ls -la /mnt/onedrive/ | grep testfile
```

---

### Scenario 7: Graceful Shutdown Timeout

**Objective**: Verify forced shutdown after graceful timeout

**Steps**:
1. Mount OneMount filesystem
2. Start uploading a very large file (>1GB)
3. Send SIGTERM signal
4. Observe graceful shutdown attempt
5. Wait for 30-second timeout
6. Verify forced shutdown

**Expected Results**:
- ✅ Graceful shutdown initiated
- ✅ Upload progress persisted
- ✅ After 30 seconds, forced shutdown occurs
- ✅ Warning logged: "Timeout reached, forcing shutdown"
- ✅ Final persistence before forced shutdown
- ✅ Database not corrupted

**Verification**:
```bash
# Monitor shutdown process
journalctl -u onemount -f | grep -i "shutdown\|timeout"

# Check for forced shutdown
journalctl -u onemount | grep "forcing shutdown"

# Verify database integrity
onemount mount /mnt/onedrive
```

---

## Current Limitations

### Upload Resumption

**Current Behavior**: Upload sessions are loaded from database but then **cancelled** (non-resumable).

**Code Reference**:
```go
// internal/fs/upload_manager.go:179
session.cancel(auth) // uploads are currently non-resumable
```

**Impact**:
- ✅ Upload state is persisted
- ✅ Upload sessions are loaded on restart
- ❌ Uploads are NOT resumed from last checkpoint
- ❌ Uploads must be restarted from beginning

**Future Enhancement**: Implement upload resumption using Microsoft Graph upload sessions.

### Delta Sync Recovery

**Current Behavior**: Delta link is persisted and sync resumes from last link.

**Status**: ✅ **IMPLEMENTED**

---

## Expected Log Patterns

### Graceful Shutdown

```json
{
  "level": "info",
  "signal": "terminated",
  "msg": "Received shutdown signal, initiating graceful shutdown"
}
```

### Upload Persistence

```json
{
  "level": "info",
  "id": "ABC123",
  "name": "testfile.bin",
  "lastChunk": 5,
  "bytesUploaded": 5242880,
  "msg": "Persisting upload progress for recovery"
}
```

### Session Restoration

```json
{
  "level": "info",
  "msg": "Restoring upload sessions from disk"
}
```

### Upload Cancellation

```json
{
  "level": "warn",
  "id": "ABC123",
  "msg": "Upload session cancelled (non-resumable)"
}
```

### Forced Shutdown

```json
{
  "level": "warn",
  "msg": "Timeout reached, forcing shutdown with active uploads"
}
```

### Database Error

```json
{
  "level": "error",
  "error": "database corruption detected",
  "msg": "Failure restoring upload sessions from disk"
}
```

---

## Success Criteria

### State Persistence

- ✅ Upload progress persisted after each chunk
- ✅ Delta link persisted after each sync
- ✅ Database survives crashes and power loss
- ✅ No database corruption from crashes

### Graceful Shutdown

- ✅ SIGTERM/SIGINT/SIGHUP handled correctly
- ✅ Upload progress persisted before shutdown
- ✅ Graceful timeout (30 seconds) enforced
- ✅ Forced shutdown after timeout

### Recovery

- ✅ Upload sessions loaded from database
- ✅ Delta sync resumes from last link
- ✅ File system state reconstructed from API
- ✅ No data loss or corruption

### Error Handling

- ✅ Database corruption detected and reported
- ✅ Clear error messages for users
- ✅ Graceful degradation on errors

---

## Known Issues and Limitations

### Issue 1: Upload Resumption Not Implemented

**Description**: Upload sessions are persisted but not resumed after crash.

**Impact**: Large file uploads must restart from beginning after crash.

**Workaround**: None currently.

**Future Fix**: Implement upload resumption using Microsoft Graph upload sessions (requires storing upload URL and chunk offsets).

**Priority**: Medium (affects user experience for large files)

### Issue 2: In-Memory State Lost on Crash

**Description**: File system state (inodes, hierarchy) is in-memory only.

**Impact**: File system must be reconstructed from API on remount (slow for large directories).

**Workaround**: None currently.

**Future Fix**: Consider persisting file system state to database.

**Priority**: Low (reconstruction is fast enough for most use cases)

---

## Troubleshooting

### Issue: Database locked error

**Possible Causes**:
- Another OneMount instance running
- Stale lock file

**Solution**:
```bash
# Check for running instances
ps aux | grep onemount

# Kill stale instances
sudo killall onemount

# Remove lock file if needed
rm ~/.local/share/onemount/onemount.db.lock
```

### Issue: Upload sessions not restored

**Possible Causes**:
- Database corruption
- Bucket not created

**Solution**:
```bash
# Check logs for errors
journalctl -u onemount | grep -i "error\|fail"

# Verify database exists
ls -la ~/.local/share/onemount/onemount.db

# Try remounting
onemount unmount /mnt/onedrive
onemount mount /mnt/onedrive
```

### Issue: Graceful shutdown hangs

**Possible Causes**:
- Large upload in progress
- Network timeout

**Solution**:
```bash
# Wait for 30-second timeout
# Or force kill if needed
sudo killall -9 onemount
```

---

## Performance Impact

### Expected Behavior

- **Graceful Shutdown**: <30 seconds
- **Database Persistence**: <100ms per operation
- **Session Restoration**: <1 second for 100 sessions
- **File System Reconstruction**: Depends on directory size

### Monitoring

```bash
# Monitor shutdown time
time sudo killall -TERM onemount

# Check database size
du -h ~/.local/share/onemount/onemount.db

# Monitor persistence operations
journalctl -u onemount -f | grep "Persisting"
```

---

## References

- **Upload Manager**: `internal/fs/upload_manager.go`
- **Database Schema**: bbolt buckets (`bucketUploads`, `bucketDelta`)
- **Signal Handling**: `signalHandler()` in upload_manager.go
- **State Persistence**: `persistActiveUploads()` in upload_manager.go
- **Session Restoration**: `NewUploadManager()` in upload_manager.go

---

## Test Results Template

```markdown
## Crash Recovery Testing Results

**Date**: YYYY-MM-DD  
**Tester**: [Name]  
**OneMount Version**: [Version]

### Scenario 1: Graceful Shutdown
- Status: ✅ PASS / ❌ FAIL
- Shutdown time: [Duration]
- State persisted: Yes / No
- Notes: [Observations]

### Scenario 2: Forced Kill (SIGKILL)
- Status: ✅ PASS / ❌ FAIL
- State recovered: Yes / No
- Database intact: Yes / No
- Notes: [Observations]

### Scenario 3: System Crash
- Status: ✅ PASS / ❌ FAIL
- Database survived: Yes / No
- State recovered: Yes / No
- Notes: [Observations]

### Scenario 4: Delta Sync Crash
- Status: ✅ PASS / ❌ FAIL
- Delta link persisted: Yes / No
- Sync resumed: Yes / No
- Notes: [Observations]

### Scenario 5: Database Corruption
- Status: ✅ PASS / ❌ FAIL
- Error detected: Yes / No
- Graceful handling: Yes / No
- Notes: [Observations]

### Scenario 6: Multiple Crashes
- Status: ✅ PASS / ❌ FAIL
- Recovery attempts: [Number]
- Final success: Yes / No
- Notes: [Observations]

### Scenario 7: Shutdown Timeout
- Status: ✅ PASS / ❌ FAIL
- Timeout enforced: Yes / No
- Forced shutdown: Yes / No
- Notes: [Observations]

### Issues Found
1. [Issue description]
2. [Issue description]

### Recommendations
1. [Recommendation]
2. [Recommendation]
```

---

**Last Updated**: 2025-11-11  
**Status**: Ready for manual testing  
**Next Steps**: Execute manual tests and document results

**Note**: Upload resumption is currently not implemented. Upload sessions are persisted but cancelled on restart. This is a known limitation documented in the code.
