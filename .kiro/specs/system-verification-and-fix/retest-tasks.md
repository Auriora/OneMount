# Re-Verification Task List with Real OneDrive

## Overview

This task list focuses on running integration and end-to-end tests that require real OneDrive authentication. These tests verify that the OneMount system works correctly with actual Microsoft OneDrive services, not just mocks.

**Prerequisites**:
- Auth tokens file exists: `test-artifacts/.auth_tokens.json`
- Tokens are valid (not expired)
- Environment variable set: `ONEMOUNT_AUTH_PATH=test-artifacts/.auth_tokens.json`
- Docker images built: `./docker/scripts/build-images.sh test-runner`
- FUSE device available in containers
- Test OneDrive account prepared with test directory
- At least 5GB free space in OneDrive

---

## High Priority Tests (Must Complete)

- [x] 1. Phase 4, Task 5.7: Mounting Integration Tests
  - Run mounting integration tests with real OneDrive
  - Execute: `docker compose -f docker/compose/docker-compose.test.yml run --rm integration-tests go test -v -run TestIT_FS_Mount ./internal/fs`
  - Verify filesystem mounts successfully
  - Verify root directory is accessible
  - Verify mount point validation works
  - Document results in verification tracking
  - _Requirements: 2.1, 2.2, 2.4, 2.5_
  - _Estimated Time: 2-3 hours_


- [x] 2. Phase 5: ETag Validation Tests
  - Run ETag validation integration tests
  - Execute: `docker compose -f docker/compose/docker-compose.test.yml run --rm integration-tests go test -v -run TestIT_FS_ETag ./internal/fs`
  - Verify cache validation with if-none-match header
  - Verify 304 Not Modified responses are handled
  - Verify cache updates on ETag changes
  - Document results in existing verification tracking documents
  - _Requirements: 3.4, 3.5, 3.6, 7.3_
  - _Estimated Time: 3-4 hours_

- [x] 3. Phase 5: Conflict Detection Verification
  - Run conflict detection integration tests
  - Execute: `docker compose -f docker/compose/docker-compose.test.yml run --rm integration-tests go test -v -run TestIT_FS.*Conflict ./internal/fs`
  - Verify conflicts are detected when file modified locally and remotely
  - Verify conflict copies are created correctly
  - Verify both versions are preserved
  - Document results in existing verification tracking documents
  - _Requirements: 5.4, 8.1, 8.2, 8.3_
  - _Estimated Time: 2-3 hours_

- [ ] 4. Phase 14: E2E Test - Complete User Workflow
  - Run end-to-end test for complete user workflow
  - Execute: `docker compose -f docker/compose/docker-compose.test.yml run --rm -e RUN_E2E_TESTS=1 system-tests go test -v -run TestE2E_17_01 ./internal/fs`
  - Verify authentication → mount → list → read → write → sync workflow
  - Verify all steps complete successfully
  - Verify state is preserved across mount/unmount
  - Document results in existing verification tracking documents
  - _Requirements: All major requirements_
  - _Estimated Time: 1-2 hours_

- [ ] 5. Phase 14: E2E Test - Multi-File Operations
  - Run end-to-end test for multi-file operations
  - Execute: `docker compose -f docker/compose/docker-compose.test.yml run --rm -e RUN_E2E_TESTS=1 system-tests go test -v -run TestE2E_17_02 ./internal/fs`
  - Verify copying entire directories to OneDrive
  - Verify all files upload correctly
  - Verify copying directories from OneDrive
  - Verify all files download correctly
  - Document results in existing verification tracking documents
  - _Requirements: 3.2, 4.3, 10.1, 10.2_
  - _Estimated Time: 1-2 hours_

---

## Medium Priority Tests (Should Complete)

- [ ] 6. Phase 4: Directory Deletion with Real Server
  - Run file write integration tests including directory operations
  - Execute: `docker compose -f docker/compose/docker-compose.test.yml run --rm integration-tests go test -v -run TestIT_FS_FileWrite ./internal/fs`
  - Verify directory creation works
  - Verify directory deletion syncs to OneDrive
  - Verify nested directory operations
  - Document results in existing verification tracking documents
  - _Requirements: 4.1_
  - _Estimated Time: 1-2 hours_


- [ ] 7. Phase 5: Large File Operations
  - Run large file system tests
  - Execute: `docker compose -f docker/compose/docker-compose.test.yml run --rm -e RUN_LONG_TESTS=1 system-tests go test -v -timeout 60m -run TestSYS_LargeFile ./internal/fs`
  - Verify large file uploads (>250MB) use chunked upload
  - Verify large file downloads work correctly
  - Verify upload sessions are managed properly
  - Monitor progress and verify completion
  - Document results in existing verification tracking documents
  - _Requirements: 4.3, 4.4, 4.5_
  - _Estimated Time: 4-6 hours_

- [ ] 8. Phase 8: Cache Management Manual Verification
  - Perform manual cache management verification
  - Set short cache expiration time in configuration
  - Access multiple files to populate cache
  - Monitor cache cleanup process
  - Verify old files are removed based on expiration
  - Verify cache statistics with large datasets
  - Test with different cache size limits
  - Document results in existing verification tracking documents
  - _Requirements: 7.1, 7.2, 7.3, 7.4, 7.5_
  - _Estimated Time: 2-3 hours_

- [ ] 9. Phase 10: Manual Test Scripts
  - Run manual test scripts for file status and D-Bus
  - Execute: `docker compose -f docker/compose/docker-compose.test.yml run --rm shell`
  - Run: `./tests/manual/test_file_status_updates.sh`
  - Run: `./tests/manual/test_dbus_integration.sh`
  - Run: `./tests/manual/test_dbus_fallback.sh`
  - Verify file status updates work correctly
  - Verify D-Bus signals are emitted
  - Verify fallback to extended attributes works
  - Document results in existing verification tracking documents
  - _Requirements: 8.1, 8.2, 8.3, 8.4_
  - _Estimated Time: 2-3 hours_

- [ ] 10. Phase 14: E2E Test - Long-Running Operations
  - Run end-to-end test for long-running operations
  - Execute: `docker compose -f docker/compose/docker-compose.test.yml run --rm -e RUN_E2E_TESTS=1 -e RUN_LONG_TESTS=1 system-tests go test -v -timeout 60m -run TestE2E_17_03 ./internal/fs`
  - Verify very large file uploads (1GB+)
  - Monitor progress throughout operation
  - Test interruption and resume functionality
  - Verify upload completes successfully
  - Document results in existing verification tracking documents
  - _Requirements: 4.3, 4.4_
  - _Estimated Time: 2-3 hours_

- [ ] 11. Phase 14: E2E Test - Stress Scenarios
  - Run end-to-end stress test
  - Execute: `docker compose -f docker/compose/docker-compose.test.yml run --rm -e RUN_E2E_TESTS=1 -e RUN_STRESS_TESTS=1 system-tests go test -v -timeout 30m -run TestE2E_17_04 ./internal/fs`
  - Verify many concurrent operations
  - Monitor resource usage (CPU, memory, network)
  - Verify system remains stable under load
  - Check for memory leaks
  - Document results in existing verification tracking documents
  - _Requirements: 10.1, 10.2_
  - _Estimated Time: 1-2 hours_

---

## Low Priority Tests (Nice to Have)

- [ ] 12. Phase 13: Comprehensive Integration Tests
  - Run comprehensive integration tests with real OneDrive
  - Execute: `docker compose -f docker/compose/docker-compose.test.yml run --rm integration-tests go test -v -run TestIT_COMPREHENSIVE ./internal/fs`
  - Verify all components work together end-to-end
  - Test complete workflows with real API
  - Verify error handling with real network conditions
  - Document results in existing verification tracking documents
  - _Requirements: All integration requirements_
  - _Estimated Time: 2-3 hours_

---

## Progress Tracking

**High Priority**: 0/5 completed (0%)  
**Medium Priority**: 0/6 completed (0%)  
**Low Priority**: 0/1 completed (0%)  
**Total**: 0/12 completed (0%)

---

## Time Estimates

- **High Priority**: 13-18 hours
- **Medium Priority**: 9-13 hours  
- **Low Priority**: 2-3 hours
- **Total**: 24-34 hours (3-4 days)

---

## Quick Start Guide

### 1. Verify Prerequisites

```bash
# Check auth tokens exist
ls -la test-artifacts/.auth_tokens.json

# Verify tokens are valid (check expiration)
cat test-artifacts/.auth_tokens.json | grep -i expires
```

### 2. Build Docker Images

```bash
# Build test runner image
./docker/scripts/build-images.sh test-runner

# Verify images are built
docker images | grep onemount
```

### 3. Run First Test

```bash
# Start with mounting integration tests
docker compose -f docker/compose/docker-compose.test.yml run --rm \
  integration-tests go test -v -run TestIT_FS_Mount ./internal/fs
```

### 4. Document Results

After each test:
- Update progress tracking in this file
- Document results in `docs/verification-tracking.md`
- Note any failures or issues discovered
- Update issue tracker if problems found

---

## Test Execution Notes

### Running Tests in Order

Tests can be run in any order within priority groups, but consider:
- Start with high-priority tests first
- Some tests can run in parallel to save time
- Long-running tests (large files, stress) should be scheduled appropriately

### Handling Test Failures

If a test fails:
1. Review test output and logs in `test-artifacts/logs/`
2. Check OneDrive account state (web interface)
3. Verify auth tokens are still valid
4. Document the failure with full context
5. Create issue in verification tracking
6. Determine if failure blocks other tests

### Environment Variables

Key environment variables for tests:
- `ONEMOUNT_AUTH_PATH`: Path to auth tokens file
- `RUN_E2E_TESTS=1`: Enable end-to-end tests
- `RUN_LONG_TESTS=1`: Enable long-running tests
- `RUN_STRESS_TESTS=1`: Enable stress tests

### Docker Container Access

For debugging or manual testing:
```bash
# Interactive shell in test container
docker compose -f docker/compose/docker-compose.test.yml run --rm shell

# Run specific test with debugging
docker compose -f docker/compose/docker-compose.test.yml run --rm \
  --entrypoint /bin/bash shell -c "go test -v -run TestName ./internal/package"
```

---

## References

- **Main Spec**: `.kiro/specs/system-verification-and-fix/`
- **Requirements**: `.kiro/specs/system-verification-and-fix/requirements.md`
- **Full Task List**: `.kiro/specs/system-verification-and-fix/tasks.md`
- **Verification Tracking**: `docs/verification-tracking.md`
- **Test Setup Guide**: `docs/TEST_SETUP.md`
- **Docker Test Environment**: `docs/testing/docker-test-environment.md`
- **Original Checklist**: `docs/RETEST_CHECKLIST.md`
