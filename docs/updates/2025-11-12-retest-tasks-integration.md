# Retest Tasks Integration into Main Task List

**Date**: 2025-11-12  
**Type**: Task List Restructuring  
**Status**: Completed  
**Related Spec**: `.kiro/specs/system-verification-and-fix/`

## Problem

The `retest-tasks.md` file was causing confusion because:
1. It had incorrect phase numbers that didn't match the actual phases in `tasks.md`
2. Tasks were duplicated between the two files
3. Kiro was getting confused about which phase a task belonged to
4. Documentation updates weren't consistent with the original task structure

## Solution

Integrated all retest tasks back into the main `tasks.md` file by:
1. Using **requirements numbers** (not phase numbers) to find the correct tasks
2. Unchecking tasks that need to be re-run with real OneDrive
3. Adding "Retest with real OneDrive" steps with Docker commands
4. Deleting the `retest-tasks.md` file

## Tasks Updated

### Tasks That Remain Checked (Already Completed with Real OneDrive)
- ✅ Phase 4, Task 5.7: Mounting Integration Tests (Req 2.1, 2.2, 2.4, 2.5)
- ✅ Phase 5, Task 7.x: Directory Deletion (Req 4.1)
- ✅ Phase 14, Task 17.1: Complete User Workflow E2E Test
- ✅ Phase 14, Task 17.2: Multi-File Operations E2E Test

### Tasks Unchecked for Retest (Need Real OneDrive Verification)

**Phase 6: Upload Manager**
- ❌ Task 9.5: Test upload conflict detection
  - Requirements: 4.4, 5.4, 8.1, 8.2, 8.3
  - Command: `docker compose -f docker/compose/docker-compose.test.yml run --rm integration-tests go test -v -run TestIT_FS.*Conflict ./internal/fs`

**Phase 7: Delta Synchronization**
- ❌ Task 10.5: Test conflict detection and resolution
  - Requirements: 5.4, 8.1, 8.2, 8.3
  - Command: `docker compose -f docker/compose/docker-compose.test.yml run --rm integration-tests go test -v -run TestIT_FS.*Conflict ./internal/fs`

**Phase 9: Cache Management**
- ❌ Task 11.4: Test cache expiration with manual verification
  - Requirements: 7.1, 7.2, 7.3, 7.4, 7.5
  - Command: `docker compose -f docker/compose/docker-compose.test.yml run --rm shell`
  - Manual testing required

**Phase 11: File Status and D-Bus**
- ❌ Task 13.2: Test file status updates with manual verification (Req 8.1)
- ❌ Task 13.3: Test D-Bus integration with manual verification (Req 8.2)
- ❌ Task 13.4: Test D-Bus fallback with manual verification (Req 8.4)
- ❌ Task 13.5: Test Nemo extension with manual verification (Req 8.3)
  - Commands: Run manual test scripts in Docker shell

**Phase 13: Comprehensive Integration Tests**
- ❌ Task 16.1-16.5: Run comprehensive integration tests with real OneDrive
  - Requirements: 11.1-11.5
  - Command: `docker compose -f docker/compose/docker-compose.test.yml run --rm integration-tests go test -v -run TestIT_COMPREHENSIVE ./internal/fs`

**Phase 14: End-to-End Tests**
- ❌ Task 17.3: Test long-running operations with real OneDrive
  - Requirements: 4.3, 4.4
  - Command: `docker compose -f docker/compose/docker-compose.test.yml run --rm -e RUN_E2E_TESTS=1 -e RUN_LONG_TESTS=1 system-tests go test -v -timeout 60m -run TestE2E_17_03 ./internal/fs`

- ❌ Task 17.4: Test stress scenarios with real OneDrive
  - Requirements: 10.1, 10.2
  - Command: `docker compose -f docker/compose/docker-compose.test.yml run --rm -e RUN_E2E_TESTS=1 -e RUN_STRESS_TESTS=1 system-tests go test -v -timeout 30m -run TestE2E_17_04 ./internal/fs`

**Phase 20: ETag Cache Validation**
- ❌ Task 29.1-29.6: Verify ETag-based cache validation with real OneDrive
  - Requirements: 3.4, 3.5, 3.6, 7.1, 7.3, 7.4, 8.1, 8.2, 8.3
  - Command: `docker compose -f docker/compose/docker-compose.test.yml run --rm integration-tests go test -v -run TestIT_FS_ETag ./internal/fs`

## Key Changes

1. **Single Source of Truth**: `tasks.md` is now the only task list
2. **Requirements-Based Mapping**: Used requirements numbers to correctly identify tasks (phase numbers in retest-tasks.md were incorrect)
3. **Retest Steps Added**: Each unchecked task now has explicit "Retest with real OneDrive" steps
4. **Docker Commands**: All retest steps include the exact Docker command to run
5. **Documentation Instructions**: Each task includes instructions to update `docs/verification-tracking.md`

## Phase Number Corrections

The retest-tasks.md file had incorrect phase numbers:
- "Phase 5: ETag Validation" → Actually Phase 20
- "Phase 5: Conflict Detection" → Actually Phase 6 & 7
- "Phase 4: Directory Deletion" → Actually Phase 5
- "Phase 8: Cache Management" → Actually Phase 9
- "Phase 10: Manual Test Scripts" → Actually Phase 11

## Files Modified

- `.kiro/specs/system-verification-and-fix/tasks.md` - Updated with retest steps, unchecked tasks
- `.kiro/specs/system-verification-and-fix/retest-tasks.md` - Deleted
- `docs/updates/2025-11-12-retest-tasks-integration.md` - Created (this file)

## Next Steps

To run the remaining retests:
1. Open `.kiro/specs/system-verification-and-fix/tasks.md`
2. Find unchecked tasks (marked with `[ ]`)
3. Run the Docker command specified in the "Retest with real OneDrive" section
4. Mark the task as complete `[x]` when done
5. Update `docs/verification-tracking.md` with results

## Rules Consulted

- `testing-conventions.md` (Priority 25) - Docker test environment requirements
- `documentation-conventions.md` (Priority 20) - Documentation structure
- `operational-best-practices.md` (Priority 40) - Single source of truth

## Rules Applied

- All tests must run in Docker containers
- Single source of truth for task definitions
- Documentation updates are explicit tasks
- Requirements-based task identification
