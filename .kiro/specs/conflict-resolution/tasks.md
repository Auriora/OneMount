# Tasks: Conflict Resolution

## Source
Moved from `.kiro/specs/archive/system-verification-and-fix/tasks.md`. Sections below are verbatim.

## Phase 8: Delta Synchronization Verification

- [x] 10. Verify delta synchronization
- [x] 10.1 Review delta sync code
  - Read and analyze `internal/fs/delta.go`
  - Review `internal/fs/sync.go`
  - Check delta loop implementation
  - Review delta link persistence
  - _Requirements: 5.1, 5.2, 5.3, 5.4, 5.5_

- [x] 10.2 Test initial delta sync
  - Start with empty cache
  - Mount filesystem
  - Verify initial sync fetches all metadata
  - Check that delta link is stored
  - _Requirements: 5.1, 5.5_

- [x] 10.3 Test incremental delta sync
  - Create a file on OneDrive web interface
  - Wait for delta sync to run
  - Verify new file appears in mounted filesystem
  - Check that only changes were fetched
  - _Requirements: 5.1, 5.2_

- [x] 10.4 Test remote file modification
  - Modify a file on OneDrive web interface
  - Wait for delta sync
  - Access the file locally
  - Verify new version is downloaded
  - _Requirements: 5.3_

- [x] 10.5 Test conflict detection and resolution with real OneDrive
  - Modify a file locally (don't let it upload yet)
  - Modify same file on OneDrive web interface
  - Trigger delta sync
  - Verify conflict is detected
  - Check that conflict copy is created
  - Verify local version is preserved
  - **Retest with real OneDrive**: `docker compose -f docker/compose/docker-compose.test.yml run --rm integration-tests go test -v -run TestIT_FS.*Conflict ./internal/fs`
  - Verify delta sync conflict detection works with real API
  - Verify local changes preserved when remote changes detected
  - Verify ETag comparison mechanism for conflict detection
  - Document results in `docs/verification-tracking.md` Phase 7 section
  - _Requirements: 5.4, 8.1, 8.2, 8.3_

- [x] 10.6 Test delta sync persistence
  - Run delta sync
  - Unmount filesystem
  - Remount filesystem
  - Verify delta sync resumes from last position
  - _Requirements: 5.5_

- [x] 10.7 Create delta sync integration tests
  - Write test for initial sync
  - Write test for incremental sync
  - Write test for conflict detection
  - Write test for delta link persistence
  - _Requirements: 5.1, 5.2, 5.3, 5.4, 5.5_

- [x] 10.8 Implement delta synchronization property-based tests
- [x] 10.8.1 Implement Property 20: Initial Delta Sync
  - **Property 20: Initial Delta Sync**
  - **Validates: Requirements 5.1**
  - Create `internal/fs/delta_property_test.go`
  - Generate random first filesystem mount scenarios
  - Verify complete directory structure fetch using delta API
  - Test with various OneDrive structures
  - _Requirements: 5.1_

- [x] 10.8.2 Implement Property 21: Metadata Cache Updates
  - **Property 21: Metadata Cache Updates**
  - **Validates: Requirements 5.8**
  - Generate random remote change scenarios via delta query
  - Verify local metadata cache is updated correctly
  - Test with various change types (create, modify, delete)
  - _Requirements: 5.8_

- [x] 10.8.3 Implement Property 22: Conflict Copy Creation
  - **Property 22: Conflict Copy Creation**
  - **Validates: Requirements 5.11**
  - Generate random files with both local and remote changes
  - Verify conflict copy is created correctly
  - Test conflict detection accuracy
  - _Requirements: 5.11_

- [x] 10.8.4 Implement Property 23: Delta Token Persistence
  - **Property 23: Delta Token Persistence**
  - **Validates: Requirements 5.12**
  - Generate random delta sync completion scenarios
  - Verify @odata.deltaLink token is stored for next cycle
  - Test token persistence across restarts
  - _Requirements: 5.12_

- [x] 10.8.5 Implement Property 30: ETag-Based Conflict Detection
  - **Property 30: ETag-Based Conflict Detection**
  - **Validates: Requirements 8.1**
  - Generate random files modified both locally and remotely
  - Verify conflict detection using ETag comparison
  - Test ETag comparison accuracy
  - _Requirements: 8.1_

- [x] 10.8.6 Implement Property 31: Local Version Preservation
  - **Property 31: Local Version Preservation**
  - **Validates: Requirements 8.4**
  - Generate random conflict scenarios
  - Verify local version is preserved with original name
  - Test version preservation integrity
  - _Requirements: 8.4_

- [x] 10.8.7 Implement Property 32: Conflict Copy Creation with Timestamp
  - **Property 32: Conflict Copy Creation with Timestamp**
  - **Validates: Requirements 8.5**
  - Generate random conflict scenarios
  - Verify conflict copy creation with timestamp suffix
  - Test timestamp format and uniqueness
  - _Requirements: 8.5_

- [x] 10.9 Document delta sync issues and create fix plan
  - List all discovered issues
  - Identify root causes
  - Create prioritized fix plan
  - Update the relevant sections of the verification-tracking.md document
  - _Requirements: 12.1_

---

## Phase 14: Integration and End-to-End Testing

- [x] 16. Run comprehensive integration tests with real OneDrive
- [x] 16.1 Test authentication to file access with real OneDrive
  - Test complete flow: authenticate → mount → list files → read file
  - Verify each step works correctly
  - Check error handling at each step
  - `docker compose -f docker/compose/docker-compose.test.yml run --rm integration-tests go test -v -run TestIT_COMPREHENSIVE ./internal/fs`
  - Verify all components work together end-to-end
  - Test complete workflows with real API
  - Verify error handling with real network conditions
  - Document results in `docs/verification-tracking.md` Phase 13 section
  - _Requirements: 11.1_

- [x] 16.2 Test file modification to sync with real OneDrive
  - Test flow: create file → modify → upload → verify on OneDrive
  - Check that all steps complete
  - Verify file appears correctly on OneDrive
  - **Covered by TestIT_COMPREHENSIVE integration test above**
  - _Requirements: 11.2_

- [x] 16.3 Test offline mode with real OneDrive
  - Test flow: online → access files → go offline → access cached files → go online
  - Verify offline detection works
  - Check that cached files remain accessible
  - Verify online transition works
  - **Covered by TestIT_COMPREHENSIVE integration test above**
  - _Requirements: 11.3_

- [x] 16.4 Test conflict resolution with real OneDrive
  - Test flow: modify file locally → modify remotely → sync → verify conflict copy
  - Check that both versions are preserved
  - Verify conflict is detected correctly
  - **Covered by TestIT_COMPREHENSIVE integration test above**
  - _Requirements: 11.4_

- [x] 16.5 Test cache cleanup with real OneDrive
  - Test flow: access files → wait for expiration → trigger cleanup → verify old files removed
  - Check that cleanup respects expiration settings
  - Verify recent files are retained
  - **Covered by TestIT_COMPREHENSIVE integration test above**
  - _Requirements: 11.5_

- [x] 17. Create end-to-end workflow tests
- [x] 17.1 Test complete user workflow
  - Install OneMount
  - Authenticate with Microsoft account
  - Mount OneDrive
  - Create, modify, and delete files
  - Verify changes sync to OneDrive
  - Unmount and remount
  - Verify state is preserved
  - _Requirements: All_

- [x] 17.2 Test multi-file operations
  - Copy entire directory to OneDrive
  - Verify all files upload correctly
  - Copy directory from OneDrive to local
  - Verify all files download correctly
  - _Requirements: 3.2, 4.3, 10.1, 10.2_

- [x] 17.3 Test long-running operations with real OneDrive
  - Upload a very large file (1GB+)
  - Monitor progress
  - Verify upload completes successfully
  - Test interruption and resume
  - `docker compose -f docker/compose/docker-compose.test.yml run --rm -e RUN_E2E_TESTS=1 -e RUN_LONG_TESTS=1 system-tests go test -v -timeout 60m -run TestE2E_17_03 ./internal/fs`
  - Verify very large file uploads (1GB+)
  - Monitor progress throughout operation
  - Test interruption and resume functionality
  - Document results in `docs/verification-tracking.md` Phase 14 section
  - _Requirements: 4.3, 4.4_

- [x] 17.4 Test stress scenarios with real OneDrive
  - Perform many concurrent operations
  - Monitor resource usage (CPU, memory, network)
  - Verify system remains stable
  - Check for memory leaks
  - `docker compose -f docker/compose/docker-compose.test.yml run --rm -e RUN_E2E_TESTS=1 -e RUN_STRESS_TESTS=1 system-tests go test -v -timeout 30m -run TestE2E_17_04 ./internal/fs`
  - Verify many concurrent operations work correctly
  - Monitor resource usage (CPU, memory, network)
  - Verify system remains stable under load
  - Check for memory leaks
  - Document results in `docs/verification-tracking.md` Phase 14 section
  - _Requirements: 10.1, 10.2_

---
