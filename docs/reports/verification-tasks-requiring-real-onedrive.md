# Tasks Requiring Re-Verification with Real OneDrive Authentication

**Date**: 2025-11-12  
**Status**: Ready for Re-Verification  
**Context**: Authentication has been fixed to allow tests to connect to OneDrive with saved credentials

## Overview

Several verification tasks were either skipped, marked as optional, or only tested with mocks due to authentication challenges or requiring real OneDrive connectivity. Now that authentication is working properly with saved credentials, these tasks should be re-verified with actual OneDrive integration.

---

## High Priority Tasks for Re-Verification

### Phase 4: Filesystem Mounting - Task 5.7 (Integration Tests)

**Status**: ⏭️ Optional - Not Completed  
**Priority**: HIGH  
**Reason Skipped**: Marked as optional during initial verification

**What Needs Testing**:
- Integration tests for mounting with real OneDrive
- Full mount/unmount cycle with actual server
- Directory structure synchronization
- Mount validation with real data

**Test Location**: Should be created in `internal/fs/mount_integration_test.go`

**Estimated Time**: 2-3 hours

**Commands to Run**:
```bash
# Run mounting integration tests
docker compose -f docker/compose/docker-compose.test.yml run --rm \
  integration-tests go test -v -run TestIT_FS_Mount ./internal/fs
```

---

### Phase 5: File Operations - ETag Validation (Issue #002)

**Status**: ⚠️ Needs Verification  
**Priority**: HIGH  
**Reason Skipped**: ETag validation location unclear, needs real server testing

**What Needs Testing**:
1. **HTTP `if-none-match` header usage**
   - Verify ETag is sent in download requests
   - Confirm 304 Not Modified responses are handled
   - Test cache is served on 304 responses

2. **Cache invalidation on ETag mismatch**
   - Modify file remotely
   - Verify local cache detects ETag change
   - Confirm new content is downloaded

3. **200 OK with new content**
   - Verify cache is updated on 200 OK
   - Confirm new ETag is stored

**Test Location**: Create `internal/fs/etag_validation_integration_test.go`

**Estimated Time**: 3-4 hours

**Commands to Run**:
```bash
# Run ETag validation tests
docker compose -f docker/compose/docker-compose.test.yml run --rm \
  integration-tests go test -v -run TestIT_FS_ETag ./internal/fs
```

---

### Phase 5: File Operations - Conflict Detection (Issue #005)

**Status**: ⚠️ Needs Verification  
**Priority**: HIGH  
**Reason Skipped**: Conflict scenarios require real server interaction

**What Needs Testing**:
1. **Local and remote modifications**
   - Modify file locally
   - Modify same file remotely (different content)
   - Trigger upload
   - Verify conflict is detected (412 Precondition Failed)
   - Confirm conflict copy is created

2. **Conflict resolution strategies**
   - Test KeepBoth strategy (creates conflict copy)
   - Verify both versions are preserved
   - Check file naming for conflict copies

**Test Location**: Already exists in `internal/fs/upload_conflict_integration_test.go` but needs real server verification

**Estimated Time**: 2-3 hours

**Commands to Run**:
```bash
# Run conflict detection tests with real OneDrive
docker compose -f docker/compose/docker-compose.test.yml run --rm \
  integration-tests go test -v -run TestIT_FS.*Conflict ./internal/fs
```

---

### Phase 5: File Operations - Large File Operations

**Status**: ⚠️ Needs System Test  
**Priority**: MEDIUM  
**Reason Skipped**: Large file testing requires real server and time

**What Needs Testing**:
1. **Large file downloads** (> 100MB)
   - Chunk-based download
   - Progress tracking
   - Resume capability

2. **Large file uploads** (> 100MB)
   - Chunked upload sessions
   - Progress tracking
   - Checkpoint recovery

**Test Location**: Create `internal/fs/large_file_system_test.go`

**Estimated Time**: 4-6 hours (includes upload/download time)

**Commands to Run**:
```bash
# Run large file tests (requires RUN_LONG_TESTS=1)
docker compose -f docker/compose/docker-compose.test.yml run --rm \
  -e RUN_LONG_TESTS=1 \
  system-tests go test -v -timeout 60m -run TestSYS_LargeFile ./internal/fs
```

---

### Phase 14: End-to-End Tests - All E2E Tests

**Status**: ✅ Created but Not Executed  
**Priority**: HIGH  
**Reason Skipped**: Require real OneDrive account and RUN_E2E_TESTS flag

**What Needs Testing**:

#### E2E-17-01: Complete User Workflow
- Authenticate with Microsoft account
- Mount OneDrive filesystem
- Create, modify, delete files
- Verify changes sync to OneDrive
- Unmount and remount
- Verify state persistence

#### E2E-17-02: Multi-File Operations
- Create directory with multiple files
- Copy to OneDrive mount point
- Verify all files upload
- Copy from OneDrive to local
- Verify all files download

#### E2E-17-03: Long-Running Operations
- Create 1GB file
- Upload to OneDrive
- Monitor progress
- Verify completion

#### E2E-17-04: Stress Scenarios
- 20 workers × 50 operations
- Monitor resource usage
- Verify system stability
- Check for memory leaks

**Test Location**: `internal/fs/end_to_end_workflow_test.go`

**Estimated Time**: 6-8 hours (includes long-running tests)

**Commands to Run**:
```bash
# Run all E2E tests (except long-running and stress)
docker compose -f docker/compose/docker-compose.test.yml run --rm \
  -e RUN_E2E_TESTS=1 \
  system-tests go test -v -run TestE2E ./internal/fs

# Run specific E2E test
docker compose -f docker/compose/docker-compose.test.yml run --rm \
  -e RUN_E2E_TESTS=1 \
  system-tests go test -v -run TestE2E_17_01 ./internal/fs

# Run all tests including long-running and stress
docker compose -f docker/compose/docker-compose.test.yml run --rm \
  -e RUN_E2E_TESTS=1 \
  -e RUN_LONG_TESTS=1 \
  -e RUN_STRESS_TESTS=1 \
  system-tests go test -v -timeout 60m -run TestE2E ./internal/fs
```

---

## Medium Priority Tasks for Re-Verification

### Phase 4: File Write Operations - Directory Deletion

**Status**: ⚠️ Mock Limitation  
**Priority**: MEDIUM  
**Reason Skipped**: Directory deletion requires server synchronization not supported in mocks

**What Needs Testing**:
- Delete empty directories
- Delete directories with files
- Verify server synchronization
- Check cleanup of local cache

**Test Location**: Add to `internal/fs/file_write_verification_test.go` or create new integration test

**Estimated Time**: 1-2 hours

**Commands to Run**:
```bash
# Run file write integration tests
docker compose -f docker/compose/docker-compose.test.yml run --rm \
  integration-tests go test -v -run TestIT_FS_FileWrite ./internal/fs
```

---

### Phase 8: Cache Management - Manual Verification

**Status**: ⚠️ Recommended  
**Priority**: MEDIUM  
**Reason Skipped**: Some cache behaviors best verified manually with real data

**What Needs Testing**:
1. **Cache expiration**
   - Set short expiration time (e.g., 1 hour)
   - Access files
   - Wait for expiration
   - Verify cleanup removes old files

2. **Cache statistics with large datasets**
   - Mount OneDrive with many files (1000+)
   - Check GetStats() performance
   - Verify statistics accuracy

3. **Cache size limits** (if implemented)
   - Fill cache to limit
   - Verify LRU eviction
   - Check disk space management

**Test Location**: Manual testing or create `internal/fs/cache_manual_verification_test.go`

**Estimated Time**: 2-3 hours

---

### Phase 10: File Status & D-Bus - Manual Test Scripts

**Status**: ✅ Scripts Created, Not Executed  
**Priority**: MEDIUM  
**Reason Skipped**: Manual scripts require interactive verification

**What Needs Testing**:
1. **File status updates** (`tests/manual/test_file_status_updates.sh`)
   - Create, modify, delete files
   - Monitor status changes via xattrs
   - Verify status accuracy

2. **D-Bus integration** (`tests/manual/test_dbus_integration.sh`)
   - Start D-Bus server
   - Monitor FileStatusChanged signals
   - Verify signal content

3. **D-Bus fallback** (`tests/manual/test_dbus_fallback.sh`)
   - Run without D-Bus
   - Verify graceful degradation
   - Check no panics or errors

**Test Location**: `tests/manual/test_*.sh`

**Estimated Time**: 2-3 hours

**Commands to Run**:
```bash
# Run manual test scripts in Docker
docker compose -f docker/compose/docker-compose.test.yml run --rm shell

# Inside container:
./tests/manual/test_file_status_updates.sh
./tests/manual/test_dbus_integration.sh
./tests/manual/test_dbus_fallback.sh
```

---

## Low Priority Tasks for Re-Verification

### Phase 13: Integration Tests - Real OneDrive Verification

**Status**: ✅ Created with Mocks  
**Priority**: LOW  
**Reason**: Tests work with mocks but should be verified with real server

**What Needs Testing**:
- Run all 5 comprehensive integration tests with real OneDrive
- Verify behavior matches mock expectations
- Check for any real-world edge cases

**Test Location**: `internal/fs/comprehensive_integration_test.go`

**Estimated Time**: 2-3 hours

**Commands to Run**:
```bash
# Run comprehensive integration tests
docker compose -f docker/compose/docker-compose.test.yml run --rm \
  integration-tests go test -v -run TestIT_COMPREHENSIVE ./internal/fs
```

---

## Summary of Re-Verification Needs

### By Priority

**HIGH PRIORITY** (Must Do):
1. ✅ Phase 4, Task 5.7: Mounting integration tests (2-3 hours)
2. ✅ Phase 5: ETag validation with real server (3-4 hours)
3. ✅ Phase 5: Conflict detection verification (2-3 hours)
4. ✅ Phase 14: All E2E tests (6-8 hours)

**MEDIUM PRIORITY** (Should Do):
5. ✅ Phase 4: Directory deletion with real server (1-2 hours)
6. ✅ Phase 5: Large file operations (4-6 hours)
7. ✅ Phase 8: Cache management manual verification (2-3 hours)
8. ✅ Phase 10: Manual test scripts execution (2-3 hours)

**LOW PRIORITY** (Nice to Have):
9. ✅ Phase 13: Integration tests with real OneDrive (2-3 hours)

### Total Estimated Time

- **High Priority**: 13-18 hours
- **Medium Priority**: 9-13 hours
- **Low Priority**: 2-3 hours
- **Total**: 24-34 hours (3-4 days of focused testing)

---

## Prerequisites

### Authentication Setup

Ensure authentication is properly configured:

1. **Auth tokens file exists**: `test-artifacts/.auth_tokens.json`
2. **Tokens are valid**: Not expired
3. **Environment variable set**: `ONEMOUNT_AUTH_PATH=test-artifacts/.auth_tokens.json`

### Docker Environment

Ensure Docker test environment is ready:

1. **Images built**: Run `./docker/scripts/build-images.sh test-runner`
2. **FUSE device available**: `/dev/fuse` accessible in containers
3. **Sufficient resources**: 6GB RAM, 4 CPUs for system tests

### Test Data

Prepare test OneDrive account:

1. **Test directory**: Create `/OneMount-Tests/` in OneDrive
2. **Test files**: Some existing files for download tests
3. **Sufficient space**: At least 5GB free for large file tests

---

## Execution Plan

### Phase 1: High Priority (Week 1)

**Day 1-2**: ETag Validation and Conflict Detection
- Create ETag validation integration tests
- Run conflict detection tests with real server
- Document findings

**Day 3-4**: E2E Tests
- Run E2E-17-01 (Complete User Workflow)
- Run E2E-17-02 (Multi-File Operations)
- Document results

**Day 5**: Mounting Integration Tests
- Create mounting integration tests
- Run with real OneDrive
- Verify all mount scenarios

### Phase 2: Medium Priority (Week 2)

**Day 1**: Large File Operations
- Run large file upload/download tests
- Monitor progress and performance
- Document timing and resource usage

**Day 2**: Directory Deletion and Cache
- Test directory deletion with real server
- Run cache management manual verification
- Document cache behavior

**Day 3**: Manual Test Scripts
- Execute all manual test scripts
- Verify D-Bus integration
- Document interactive testing results

### Phase 3: Low Priority (Week 2)

**Day 4**: Integration Tests Verification
- Run comprehensive integration tests with real OneDrive
- Compare results with mock-based tests
- Document any discrepancies

**Day 5**: Final Review and Documentation
- Review all test results
- Update verification tracking document
- Create summary report

---

## Success Criteria

### For Each Test

- ✅ Test executes without errors
- ✅ All assertions pass
- ✅ Behavior matches requirements
- ✅ Performance is acceptable
- ✅ No resource leaks detected
- ✅ Results documented

### Overall

- ✅ All high-priority tests pass
- ✅ At least 80% of medium-priority tests pass
- ✅ Any failures are documented with root cause
- ✅ Verification tracking document updated
- ✅ Issues logged for any problems found

---

## Documentation Updates Required

After re-verification, update:

1. **`docs/verification-tracking.md`**
   - Update test counts
   - Mark tasks as completed
   - Add any new issues found

2. **Phase documentation files**
   - Add real OneDrive test results
   - Document any differences from mock tests
   - Update status and findings

3. **Test result reports**
   - Create new reports for each phase tested
   - Include performance metrics
   - Document any edge cases discovered

4. **Issue tracking**
   - Log any new issues found
   - Update existing issues if resolved
   - Prioritize issues for fixing

---

## Notes

- **Authentication is now working**: Tests can connect to real OneDrive with saved credentials
- **Docker environment is ready**: All infrastructure is in place
- **Tests are already written**: Most tests exist, just need to be run with real server
- **Time estimates are conservative**: Actual time may be less if tests pass smoothly
- **Parallel execution possible**: Some tests can run in parallel to save time

---

**Created By**: Kiro AI  
**Date**: 2025-11-12  
**Status**: Ready for Execution  
**Next Step**: Begin with High Priority Phase 1 tasks
