# Phase 12: Offline Mode Verification - Test Plan

**Date**: 2025-11-11  
**Status**: Test Plan Created  
**Component**: Offline Mode  
**Requirements**: 6.1, 6.2, 6.3, 6.4, 6.5

## Overview

This document provides a detailed test plan for verifying the offline mode functionality of OneMount. The tests should be executed in Docker containers to ensure isolation and reproducibility.

## Test Environment Setup

### Prerequisites
```bash
# Build test images
docker compose -f docker/compose/docker-compose.build.yml build test-runner

# Verify FUSE device is available
docker compose -f docker/compose/docker-compose.test.yml run --rm shell -c "ls -l /dev/fuse"

# Verify auth tokens are configured
ls -l test-artifacts/.auth_tokens.json
```

### Network Control in Docker

To simulate offline conditions in Docker, we can use several approaches:

1. **iptables rules** (requires NET_ADMIN capability):
```bash
# Block outgoing connections to Microsoft Graph API
iptables -A OUTPUT -d graph.microsoft.com -j DROP
iptables -A OUTPUT -d login.microsoftonline.com -j DROP

# Restore connectivity
iptables -D OUTPUT -d graph.microsoft.com -j DROP
iptables -D OUTPUT -d login.microsoftonline.com -j DROP
```

2. **Network namespace isolation**:
```bash
# Disconnect network namespace
ip netns exec isolated_ns <command>
```

3. **Mock Graph Client** (preferred for testing):
```go
// Use graph.SetOperationalOffline(true) in tests
graph.SetOperationalOffline(true)  // Simulate offline
graph.SetOperationalOffline(false) // Simulate online
```

## Test Cases

### Test 12.2: Offline Detection

**Objective**: Verify that the filesystem correctly detects when network connectivity is lost.

**Test ID**: IT-OF-05-01  
**Requirement**: 6.1

**Test Steps**:
1. Mount filesystem while online
2. Verify filesystem is in online mode (`IsOffline()` returns false)
3. Simulate network disconnection using one of the methods above
4. Trigger an operation requiring network (e.g., access uncached file)
5. Verify offline state is detected (`IsOffline()` returns true)
6. Check logs for offline detection message

**Expected Results**:
- Filesystem detects offline state within one delta sync cycle
- `IsOffline()` returns true after network failure
- Logs contain message: "Error during delta fetch, marking fs as offline"
- No crashes or hangs when network is unavailable

**Docker Command**:
```bash
docker compose -f docker/compose/docker-compose.test.yml run --rm \
  --cap-add=NET_ADMIN \
  --entrypoint /bin/bash shell -c "
    go test -v -run TestIT_OF_05_01 ./internal/fs/ -timeout 10m
  "
```

**Test Implementation** (to be added to `internal/fs/offline_integration_test.go`):
```go
func TestIT_OF_05_01_OfflineDetection_NetworkFailure_DetectedCorrectly(t *testing.T) {
    // 1. Create filesystem and mount while online
    // 2. Verify online state
    // 3. Set operational offline mode
    // 4. Trigger delta sync or file access
    // 5. Verify offline state detected
    // 6. Check logs for offline message
}
```

---

### Test 12.3: Offline Read Operations

**Objective**: Verify that cached files can be read while offline, and uncached files return appropriate errors.

**Test ID**: IT-OF-06-01  
**Requirement**: 6.2

**Test Steps**:
1. Mount filesystem while online
2. Access several files to populate cache
3. Verify files are cached
4. Simulate network disconnection
5. Read cached files
6. Verify files can be read successfully
7. Attempt to access uncached file
8. Verify appropriate error message is returned

**Expected Results**:
- Cached files remain accessible offline
- File content matches what was cached
- Uncached files return clear error (e.g., "network unavailable")
- No attempts to download uncached files while offline
- Logs indicate "Using cached content in offline mode"

**Docker Command**:
```bash
docker compose -f docker/compose/docker-compose.test.yml run --rm \
  --entrypoint /bin/bash shell -c "
    go test -v -run TestIT_OF_06_01 ./internal/fs/ -timeout 10m
  "
```

**Test Implementation**:
```go
func TestIT_OF_06_01_OfflineReadOperations_CachedFiles_AccessibleOffline(t *testing.T) {
    // 1. Create filesystem and cache files
    // 2. Set offline mode
    // 3. Read cached files - should succeed
    // 4. Attempt to read uncached file - should fail with clear error
    // 5. Verify no network requests made
}
```

---

### Test 12.4: Offline Write Restrictions

**Objective**: Verify the behavior of write operations while offline.

**Test ID**: IT-OF-07-01  
**Requirement**: 6.3

**Test Steps**:
1. Mount filesystem while online
2. Simulate network disconnection
3. Attempt to create a new file
4. Verify operation behavior (currently allows, requirements say reject)
5. Attempt to modify existing file
6. Verify operation behavior
7. Check if changes are queued for upload

**Expected Results** (per current implementation):
- File creation is allowed (logged as "cached locally")
- File modification is allowed (logged as "cached locally")
- Changes are tracked in OfflineChange database
- Files marked with `hasChanges = true`

**Expected Results** (per requirements):
- File creation should be rejected (read-only mode)
- File modification should be rejected (read-only mode)
- Clear error message: "filesystem is read-only in offline mode"

**Docker Command**:
```bash
docker compose -f docker/compose/docker-compose.test.yml run --rm \
  --entrypoint /bin/bash shell -c "
    go test -v -run TestIT_OF_07_01 ./internal/fs/ -timeout 10m
  "
```

**Test Implementation**:
```go
func TestIT_OF_07_01_OfflineWriteRestrictions_WriteOperations_BehaviorVerified(t *testing.T) {
    // 1. Create filesystem and set offline
    // 2. Attempt file creation
    // 3. Attempt file modification
    // 4. Verify behavior matches implementation (or requirements)
    // 5. Check change tracking
}
```

**Note**: This test will reveal the discrepancy between requirements (read-only) and implementation (read-write with queuing).

---

### Test 12.5: Change Queuing

**Objective**: Verify that changes made while offline are properly queued for later synchronization.

**Test ID**: IT-OF-08-01  
**Requirement**: 6.4

**Test Steps**:
1. Mount filesystem while online
2. Simulate network disconnection
3. Create new file
4. Modify existing file
5. Delete file
6. Rename file
7. Verify changes are tracked in database
8. Query offline changes from database
9. Verify all changes are recorded with correct timestamps

**Expected Results**:
- All changes stored in BBolt database (bucketOfflineChanges)
- Each change has: ID, Type, Timestamp, Path
- Changes ordered by timestamp
- Database persists across filesystem restarts
- `TrackOfflineChange()` succeeds for all operation types

**Docker Command**:
```bash
docker compose -f docker/compose/docker-compose.test.yml run --rm \
  --entrypoint /bin/bash shell -c "
    go test -v -run TestIT_OF_08_01 ./internal/fs/ -timeout 10m
  "
```

**Test Implementation**:
```go
func TestIT_OF_08_01_ChangeQueuing_OfflineChanges_ProperlyTracked(t *testing.T) {
    // 1. Create filesystem and set offline
    // 2. Perform various file operations
    // 3. Verify TrackOfflineChange() called for each
    // 4. Query database for offline changes
    // 5. Verify all changes recorded correctly
    // 6. Verify timestamp ordering
}
```

---

### Test 12.6: Online Transition

**Objective**: Verify that when network connectivity is restored, the filesystem transitions back to online mode and processes queued changes.

**Test ID**: IT-OF-09-01  
**Requirement**: 6.5

**Test Steps**:
1. Mount filesystem while online
2. Simulate network disconnection
3. Make several changes while offline
4. Verify changes are queued
5. Restore network connectivity
6. Trigger delta sync or wait for automatic sync
7. Verify online state is detected
8. Verify queued changes are processed
9. Verify delta sync resumes
10. Verify changes appear on OneDrive

**Expected Results**:
- Filesystem detects online state when delta sync succeeds
- `IsOffline()` returns false after successful sync
- Logs contain: "Delta fetch success, marking fs as online"
- `ProcessOfflineChanges()` is called
- Queued changes are uploaded to OneDrive
- Upload manager processes pending uploads
- Delta sync resumes normal operation
- Offline changes bucket is cleared after successful sync

**Docker Command**:
```bash
docker compose -f docker/compose/docker-compose.test.yml run --rm \
  --entrypoint /bin/bash shell -c "
    go test -v -run TestIT_OF_09_01 ./internal/fs/ -timeout 15m
  "
```

**Test Implementation**:
```go
func TestIT_OF_09_01_OnlineTransition_NetworkRestored_ChangesProcessed(t *testing.T) {
    // 1. Create filesystem, go offline, make changes
    // 2. Verify changes queued
    // 3. Restore network (SetOperationalOffline(false))
    // 4. Trigger delta sync
    // 5. Verify online state detected
    // 6. Verify ProcessOfflineChanges() called
    // 7. Verify changes uploaded
    // 8. Verify offline changes cleared
}
```

---

### Test 12.7: Integration Test Review

**Objective**: Review and enhance existing offline mode integration tests.

**Existing Tests** (in `internal/fs/offline_integration_test.go`):
1. ✅ `TestIT_OF_01_01`: Offline file access - basic operations
2. ✅ `TestIT_OF_02_01`: Offline filesystem operations
3. ✅ `TestIT_OF_03_01`: Offline changes cached
4. ✅ `TestIT_OF_04_01`: Offline synchronization after reconnect

**Enhancements Needed**:
- Add explicit offline detection test (12.2)
- Add cached vs uncached file access test (12.3)
- Add write restriction verification test (12.4)
- Add change queuing verification test (12.5)
- Add comprehensive online transition test (12.6)

**Docker Command to Run All Offline Tests**:
```bash
docker compose -f docker/compose/docker-compose.test.yml run --rm \
  --entrypoint /bin/bash shell -c "
    go test -v -run 'TestIT_OF' ./internal/fs/ -timeout 20m
  "
```

---

## Test Execution Summary

### Run All Offline Mode Tests

```bash
# Build images first
./docker/scripts/build-images.sh test-runner

# Run all offline integration tests
docker compose -f docker/compose/docker-compose.test.yml run --rm integration-tests \
  -run 'TestIT_OF'

# Or run specific test
docker compose -f docker/compose/docker-compose.test.yml run --rm \
  --entrypoint /bin/bash shell -c "
    go test -v -run TestIT_OF_05_01 ./internal/fs/ -timeout 10m
  "

# Check test artifacts
ls -la test-artifacts/logs/
```

### Expected Test Results

| Test ID | Test Name | Expected Result | Requirements |
|---------|-----------|-----------------|--------------|
| IT-OF-01-01 | Offline file access | ✅ Pass | 6.2 |
| IT-OF-02-01 | Offline filesystem operations | ⚠️ Pass (but violates 6.3) | 6.3, 6.4 |
| IT-OF-03-01 | Offline changes cached | ✅ Pass | 6.4 |
| IT-OF-04-01 | Offline synchronization | ✅ Pass | 6.5 |
| IT-OF-05-01 | Offline detection | 🆕 To be created | 6.1 |
| IT-OF-06-01 | Offline read operations | 🆕 To be created | 6.2 |
| IT-OF-07-01 | Offline write restrictions | 🆕 To be created | 6.3 |
| IT-OF-08-01 | Change queuing | 🆕 To be created | 6.4 |
| IT-OF-09-01 | Online transition | 🆕 To be created | 6.5 |

## Known Issues and Discrepancies

### Issue #1: Read-Write vs Read-Only Offline Mode

**Severity**: Medium  
**Requirement**: 6.3

**Description**: Requirements specify that the filesystem should be read-only while offline, but the implementation allows read-write operations with change queuing.

**Current Behavior**:
- File creation allowed offline
- File modification allowed offline
- File deletion allowed offline
- Changes queued for upload when back online

**Required Behavior**:
- Filesystem should be read-only offline
- Write operations should be rejected
- Clear error message to user

**Recommendation**: 
- **Option A**: Update requirements to match implementation (read-write with queuing)
- **Option B**: Modify implementation to enforce read-only mode
- **Option C**: Make it configurable (read-only vs read-write offline mode)

### Issue #2: Passive Offline Detection

**Severity**: Low  
**Requirement**: 6.1

**Description**: Offline state is detected passively through delta sync failures rather than actively monitoring network interfaces.

**Current Behavior**:
- Offline detected when delta sync fails
- Detection happens on next sync cycle (up to 5 minutes delay)
- No direct network interface monitoring

**Required Behavior**:
- "Network connectivity is lost" should be detected

**Recommendation**:
- Current approach is pragmatic and works well
- Active network monitoring would add complexity
- Document that detection happens via delta sync
- Consider adding explicit network connectivity check if needed

## Test Artifacts

All test artifacts will be stored in:
- `test-artifacts/logs/` - Test execution logs
- `test-artifacts/.onemount-tests/` - Test filesystem data
- `test-artifacts/.cache/` - Test cache data

## Next Steps

1. ✅ Review offline mode code (Task 12.1) - **COMPLETED**
2. 🔄 Create new integration tests (Tasks 12.2-12.6)
3. 🔄 Run all offline tests in Docker
4. 🔄 Document test results
5. 🔄 Create fix plan for identified issues (Task 12.8)
6. 🔄 Update requirements or implementation based on findings

## References

- Requirements: `docs/1-requirements/srs/requirements.md` (Section 6)
- Design: `.kiro/specs/system-verification-and-fix/design.md` (Offline Mode Component)
- Implementation: `internal/fs/offline.go`, `internal/fs/cache.go`
- Existing Tests: `internal/fs/offline_integration_test.go`
- Verification Tracking: `docs/verification-tracking.md` (Phase 11)
