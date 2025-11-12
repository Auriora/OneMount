# Implementation Plan: OneMount System Verification and Fix

## Overview

This implementation plan breaks down the verification and fix process into discrete, manageable tasks. Each task builds on previous tasks and focuses on verifying specific components against requirements, identifying issues, and implementing fixes.

---

## Phase 1: Docker Environment Setup and Validation

- [x] 1. Review and validate Docker test environment
- [x] 1.1 Review Docker configuration files
  - Review `.devcontainer/Dockerfile` and `.devcontainer/devcontainer.json`
  - Review `docker/compose/docker-compose.test.yml`
  - Review `packaging/docker/Dockerfile.test-runner`
  - Review `packaging/docker/test-entrypoint.sh`
  - Verify all required dependencies are included
  - _Requirements: 13.1, 13.2, 13.3, 13.6, 13.7_

- [x] 1.2 Build Docker test images
  - Build base image: `docker compose -f docker/compose/docker-compose.build.yml build base-image`
  - Build test runner: `docker compose -f docker/compose/docker-compose.build.yml build test-runner`
  - Verify images are created successfully
  - Check image sizes and layers
  - _Requirements: 13.7_

- [x] 1.3 Validate Docker test environment
  - Test shell access: `docker compose -f docker/compose/docker-compose.test.yml run shell`
  - Verify FUSE device is accessible: `ls -l /dev/fuse`
  - Verify Go environment: `go version`
  - Verify Python environment: `python3 --version`
  - Test workspace mounting: `ls -la /workspace`
  - Test artifact directory: `ls -la /tmp/home-tester/.onemount-tests`
  - _Requirements: 13.4, 13.5, 13.6_

- [x] 1.4 Setup test credentials and data
  - Create test OneDrive account with sample files (if not already available)
  - Configure auth tokens in `test-artifacts/.auth_tokens.json` for system tests
  - Create sample test files in OneDrive for verification
  - Document test account setup and credentials storage
  - _Requirements: 13.5_

- [x] 1.5 Document Docker test environment
  - Document how to build images
  - Document how to run different test types
  - Document how to access test artifacts
  - Document how to debug in containers
  - Document environment variables and configuration options
  - _Requirements: 13.1, 13.2, 13.3, 13.4, 13.5_

## Phase 2: Initial Test Suite Analysis

- [x] 2. Analyze existing test suite
  - Run all existing unit tests in Docker: `docker compose -f docker/compose/docker-compose.test.yml run unit-tests`
  - Run all existing integration tests in Docker: `docker compose -f docker/compose/docker-compose.test.yml run integration-tests`
  - Document test results from `test-artifacts/logs/`
  - Identify which tests pass vs fail
  - Analyze test coverage gaps
  - Create test results summary document
  - _Requirements: 11.1, 11.2, 11.3, 11.4, 11.5, 13.1, 13.2, 13.4, 13.5_

- [x] 3. Create verification tracking document
  - Create spreadsheet or markdown table for tracking component verification status
  - Set up issue tracking for discovered problems
  - Create template for test result documentation
  - Set up traceability matrix linking requirements to tests
  - **Document created**: `docs/verification-tracking.md`
  - _Requirements: 12.1, 12.2, 12.3_

---

## Phase 3: Authentication Component Verification

- [x] 4. Verify authentication implementation
- [x] 4.1 Review OAuth2 code structure
  - Read and analyze `internal/graph/oauth2.go`, `oauth2_gtk.go`, `oauth2_headless.go`
  - Review `internal/graph/authenticator.go` interface and implementations
  - Compare implementation against design document
  - Document any deviations from architecture
  - _Requirements: 1.1, 1.5_

- [x] 4.2 Test interactive authentication flow
  - Use Docker shell for interactive testing: `docker compose -f docker/compose/docker-compose.test.yml run shell`
  - Launch OneMount with GUI authentication (if GTK available in container)
  - Complete Microsoft OAuth2 flow
  - Verify tokens are stored in `test-artifacts/.auth_tokens.json`
  - Check file permissions on token storage
  - Verify tokens contain AccessToken, RefreshToken, and ExpiresAt
  - _Requirements: 1.1, 1.2, 13.4, 13.5_

- [x] 4.3 Test token refresh mechanism
  - Manually expire access token (modify ExpiresAt)
  - Trigger operation requiring authentication
  - Verify automatic token refresh occurs
  - Check that new tokens are persisted
  - _Requirements: 1.3_

- [x] 4.4 Test authentication failure scenarios
  - Test with invalid credentials
  - Test with network disconnection during auth
  - Test with expired refresh token
  - Verify error messages are clear and actionable
  - _Requirements: 1.4_

- [x] 4.5 Test headless authentication
  - Run OneMount in headless mode (no GUI)
  - Verify device code flow is used
  - Complete authentication via browser
  - Verify tokens are stored correctly
  - _Requirements: 1.5_

- [x] 4.6 Create authentication integration tests
  - Write test for complete OAuth2 flow with mock server
  - Write test for token refresh with mock responses
  - Write test for authentication failure scenarios
  - Run tests in Docker: `docker compose -f docker/compose/docker-compose.test.yml run integration-tests`
  - _Requirements: 1.1, 1.2, 1.3, 1.4, 13.2, 13.4_

- [x] 4.7 Document authentication issues and create fix plan
  - List all discovered issues with severity
  - Identify root causes
  - Create prioritized fix plan
  - Update architecture docs if implementation differs
  - _Requirements: 12.1, 12.4_

---

## Phase 4: Filesystem Mounting Verification

- [x] 5. Verify filesystem mounting
- [x] 5.1 Review FUSE initialization code
  - Read and analyze `internal/fs/raw_filesystem.go`
  - Review `cmd/onemount/main.go` mount logic
  - Compare against design document
  - _Requirements: 2.1, 2.2_

- [x] 5.2 Test basic mounting
  - Mount filesystem at test mount point inside a docker container
  - Verify mount appears in `mount` command output
  - Verify mount point is accessible
  - Check that root directory is visible
  - _Requirements: 2.1, 2.2_

- [x] 5.3 Test mount point validation (in a docker container)
  - Attempt to mount at non-existent directory
  - Attempt to mount at already-mounted location
  - Attempt to mount at file (not directory)
  - Verify appropriate error messages
  - _Requirements: 2.4_

- [x] 5.4 Test filesystem operations while mounted
  - Run `ls` on mount point
  - Run `cat` on a file
  - Run `cp` to copy a file
  - Verify operations complete without hanging
  - _Requirements: 2.3_

- [x] 5.5 Test unmounting and cleanup
  - Unmount filesystem using `fusermount3 -uz`
  - Verify mount point is released
  - Check for orphaned processes
  - Verify clean shutdown in logs
  - _Requirements: 2.5_

- [x] 5.6 Test signal handling
  - Mount filesystem
  - Send SIGINT (Ctrl+C)
  - Verify graceful shutdown
  - Repeat with SIGTERM
  - _Requirements: 2.5_

- [x] 5.7 Create mounting integration tests
  - Write test for successful mount
  - Write test for mount failure scenarios
  - Write test for graceful unmount
  - _Requirements: 2.1, 2.2, 2.4, 2.5_

- [x] 5.8 Document mounting issues and create fix plan
  - List all discovered issues
  - Identify root causes
  - Create prioritized fix plan
  - _Requirements: 12.1_

---

## Phase 5: File Operations Verification

- [x] 6. Verify file read operations
- [x] 6.1 Review file operation code
  - Read and analyze `internal/fs/file_operations.go`
  - Review FUSE operation handlers (Open, Read, Release)
  - Compare against design document
  - _Requirements: 3.1, 3.2, 3.3_

- [x] 6.2 Test reading uncached files
  - Clear cache
  - Read a file that hasn't been accessed
  - Verify file downloads from OneDrive
  - Check file content is correct
  - Verify file is cached after read
  - _Requirements: 3.2_

- [x] 6.3 Test reading cached files
  - Read a previously accessed file
  - Verify no network request is made (check logs)
  - Verify content is served from cache
  - Check read performance is fast
  - _Requirements: 3.3_

- [x] 6.4 Test directory listing
  - List a directory with many files
  - Verify all files appear
  - Verify no file content is downloaded
  - Check that metadata is displayed correctly
  - _Requirements: 3.1_

- [x] 6.5 Test file metadata operations
  - Run `stat` on files
  - Check file size, timestamps, permissions
  - Verify metadata matches OneDrive
  - _Requirements: 3.1_

- [x] 6.6 Create file read integration tests
  - Write test for uncached file read
  - Write test for cached file read
  - Write test for directory listing
  - Write test for metadata operations
  - _Requirements: 3.1, 3.2, 3.3_

- [x] 6.7 Document file read issues and create fix plan
  - List all discovered issues
  - Identify root causes
  - Create prioritized fix plan
  - _Requirements: 12.1_

- [x] 7. Verify file write operations
- [x] 7.1 Test file creation
  - Create a new file in mounted directory
  - Write content to the file
  - Verify file appears in directory listing
  - Check that file is marked for upload
  - _Requirements: 4.1, 4.2_

- [x] 7.2 Test file modification
  - Modify an existing file
  - Save changes
  - Verify file is marked as modified
  - Check that upload is queued
  - _Requirements: 4.1, 4.2_

- [x] 7.3 Test file deletion
  - Delete a file
  - Verify file is removed from directory listing
  - Check that deletion is synced to OneDrive
  - _Requirements: 4.1_

- [x] 7.4 Test directory operations
  - Create a new directory
  - Create files within the directory
  - Delete the directory
  - Verify operations sync correctly
  - _Requirements: 4.1_

- [x] 7.5 Create file write integration tests
  - Write test for file creation and upload
  - Write test for file modification and upload
  - Write test for file deletion
  - Write test for directory operations
  - _Requirements: 4.1, 4.2_

- [x] 7.6 Document file write issues and create fix plan
  - List all discovered issues
  - Identify root causes
  - Create prioritized fix plan
  - _Requirements: 12.1_

---

## Phase 6: Download Manager Verification

- [x] 8. Verify download manager
- [x] 8.1 Review download manager code
  - Read and analyze `internal/fs/download_manager.go`
  - Review worker pool implementation
  - Check queue management
  - _Requirements: 3.2, 3.4, 3.5_

- [x] 8.2 Test single file download
  - Trigger download of one file
  - Monitor download progress in logs
  - Verify file content is correct
  - Check that file is cached
  - _Requirements: 3.2_

- [x] 8.3 Test concurrent downloads
  - Trigger downloads of multiple files simultaneously
  - Verify downloads proceed concurrently
  - Check that all downloads complete
  - Verify no race conditions or deadlocks
  - _Requirements: 3.4, 10.1_

- [x] 8.4 Test download failure and retry
  - Simulate network failure during download
  - Verify download is retried
  - Check exponential backoff is used
  - Verify eventual success or clear error
  - _Requirements: 3.5, 9.1_

- [x] 8.5 Test download status tracking
  - Monitor file status during download
  - Verify status changes from "not cached" to "downloading" to "cached"
  - Check that status is visible via extended attributes
  - _Requirements: 3.4, 8.1_

- [x] 8.6 Create download manager integration tests
  - Write test for single download
  - Write test for concurrent downloads
  - Write test for download retry logic
  - Write test for download cancellation
  - _Requirements: 3.2, 3.4, 3.5_

- [x] 8.7 Document download manager issues and create fix plan
  - List all discovered issues
  - Identify root causes
  - Create prioritized fix plan
  - _Requirements: 12.1_

---

## Phase 7: Upload Manager Verification

- [-] 9. Verify upload manager
- [x] 9.1 Review upload manager code
  - Read and analyze `internal/fs/upload_manager.go`
  - Review `internal/fs/upload_session.go`
  - Check queue and retry logic
  - _Requirements: 4.2, 4.3, 4.4, 4.5_

- [x] 9.2 Test small file upload
  - Create and modify a small file (< 4MB)
  - Verify upload is queued
  - Monitor upload progress
  - Verify file appears on OneDrive
  - Check ETag is updated
  - _Requirements: 4.2, 4.3, 4.5_

- [x] 9.3 Test large file upload
  - Create a large file (> 10MB)
  - Verify chunked upload is used
  - Monitor upload progress
  - Verify complete file on OneDrive
  - _Requirements: 4.3_

- [x] 9.4 Test upload failure and retry
  - Simulate network failure during upload
  - Verify upload is retried
  - Check exponential backoff
  - Verify eventual success
  - _Requirements: 4.4_

- [x] 9.5 Test upload conflict detection with real OneDrive
  - Modify file locally
  - Modify same file on OneDrive web interface
  - Trigger upload
  - Verify conflict is detected
  - Check conflict resolution (should be tested in delta sync)
  - **Retest with real OneDrive**: `docker compose -f docker/compose/docker-compose.test.yml run --rm integration-tests go test -v -run TestIT_FS.*Conflict ./internal/fs`
  - Verify conflicts are detected when file modified locally and remotely
  - Verify conflict copies are created correctly
  - Verify both versions are preserved
  - Document results in `docs/verification-tracking.md` Phase 6 section
  - _Requirements: 4.4, 5.4, 5.4, 8.1, 8.2, 8.3_

- [x] 9.6 Create upload manager integration tests
  - Write test for small file upload
  - Write test for large file chunked upload
  - Write test for upload retry logic
  - Write test for upload queue management
  - _Requirements: 4.2, 4.3, 4.4_

- [x] 9.7 Document upload manager issues and create fix plan
  - List all discovered issues
  - Identify root causes
  - Create prioritized fix plan
  - Update the relevant sections of the verification-tracking.md document
  - _Requirements: 12.1_

---

## Phase 8: Delta Synchronization Verification

- [ ] 10. Verify delta synchronization
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

- [x] 10.8 Document delta sync issues and create fix plan
  - List all discovered issues
  - Identify root causes
  - Create prioritized fix plan
  - Update the relevant sections of the verification-tracking.md document
  - _Requirements: 12.1_

---

## Phase 9: Cache Management Verification

- [x] 11. Verify cache management
- [x] 11.1 Review cache code
  - Read and analyze `internal/fs/cache.go`
  - Review `internal/fs/content_cache.go`
  - Check cache cleanup implementation
  - Review bbolt database usage
  - _Requirements: 7.1, 7.2, 7.3, 7.4, 7.5_

- [x] 11.2 Test content caching
  - Access several files
  - Verify content is stored in cache directory
  - Check cache directory structure
  - Verify cached content is correct
  - _Requirements: 7.1_

- [x] 11.3 Test cache hit/miss
  - Access a cached file (should be cache hit)
  - Access an uncached file (should be cache miss)
  - Verify cache statistics reflect hits and misses
  - _Requirements: 7.5_

- [x] 11.4 Test cache expiration with manual verification
  - Configure short cache expiration (e.g., 1 day)
  - Create files with old access times
  - Trigger cache cleanup
  - Verify old files are removed
  - Verify recent files are retained
  - **Retest**: Perform manual cache management verification in Docker
  - Set short cache expiration time in configuration
  - Access multiple files to populate cache
  - Monitor cache cleanup process
  - Verify cache statistics with large datasets
  - Test with different cache size limits
  - Run in Docker: `docker compose -f docker/compose/docker-compose.test.yml run --rm shell`
  - Document results in `docs/verification-tracking.md` Phase 9 section
  - _Requirements: 7.1, 7.2, 7.3, 7.4, 7.5_

- [x] 11.5 Test cache statistics
  - Run `onemount --stats /mount/path`
  - Verify statistics show cache size
  - Check file count
  - Verify hit rate calculation
  - _Requirements: 7.5_

- [x] 11.6 Test metadata cache persistence
  - Access files to populate metadata cache
  - Unmount filesystem
  - Remount filesystem
  - Verify metadata is still cached (no refetch)
  - _Requirements: 7.1_

- [x] 11.7 Create cache management integration tests
  - Write test for cache storage and retrieval
  - Write test for cache expiration
  - Write test for cache cleanup
  - Write test for cache statistics
  - _Requirements: 7.1, 7.2, 7.3, 7.5_

- [x] 11.8 Document cache issues and create fix plan
  - List all discovered issues
  - Identify root causes
  - Create prioritized fix plan
  - Update the relevant sections of the verification-tracking.md document
  - _Requirements: 12.1_

---

## Phase 10: Offline Mode Verification

- [x] 12. Verify offline mode
- [x] 12.1 Review offline mode code
  - Read and analyze `internal/fs/offline.go`
  - Check offline detection logic
  - Review change queuing implementation
  - _Requirements: 6.1, 6.2, 6.3, 6.4, 6.5_

- [x] 12.2 Test offline detection
  - Mount filesystem while online
  - Disconnect network (disable network interface)
  - Trigger operation requiring network
  - Verify offline state is detected
  - Check logs for offline detection message
  - _Requirements: 6.1_

- [x] 12.3 Test offline read operations
  - While offline, access cached files
  - Verify files can be read
  - Attempt to access uncached file
  - Verify appropriate error message
  - _Requirements: 6.2_

- [x] 12.4 Test offline write restrictions
  - While offline, attempt to create file
  - Verify operation is rejected (read-only)
  - Attempt to modify file
  - Verify operation is rejected
  - _Requirements: 6.3_

- [x] 12.5 Test change queuing (if implemented)
  - If system allows queuing changes while offline
  - Make changes while offline
  - Verify changes are queued
  - _Requirements: 6.4_

- [x] 12.6 Test online transition
  - While offline, reconnect network
  - Trigger operation requiring network
  - Verify online state is detected
  - Check that queued changes are processed
  - Verify delta sync resumes
  - _Requirements: 6.5_

- [x] 12.7 Create offline mode integration tests
  - Write test for offline detection
  - Write test for offline read operations
  - Write test for offline write restrictions
  - Write test for online transition
  - _Requirements: 6.1, 6.2, 6.3, 6.5_

- [x] 12.8 Document offline mode issues and create fix plan
  - List all discovered issues
  - Identify root causes
  - Create prioritized fix plan
  - Update the relevant sections of the verification-tracking.md document
  - _Requirements: 12.1_

---

## Phase 11: File Status and D-Bus Verification

- [x] 13. Verify file status tracking
- [x] 13.1 Review file status code
  - Read and analyze `internal/fs/file_status.go`
  - Review `internal/fs/dbus.go`
  - Check extended attribute implementation
  - Review Nemo extension code
  - _Requirements: 8.1, 8.2, 8.3, 8.4, 8.5_

- [ ] 13.2 Test file status updates with manual verification
  - Monitor file status during various operations
  - Verify status changes appropriately (synced, downloading, error, etc.)
  - Check extended attributes are set correctly
  - Run: `./tests/manual/test_file_status_updates.sh`
  - Verify file status updates work correctly with real OneDrive
  - Document results in `docs/verification-tracking.md` Phase 11 section
  - _Requirements: 8.1_

- [ ] 13.3 Test D-Bus integration with manual verification
  - Verify D-Bus server starts successfully
  - Monitor D-Bus signals during file operations
  - Use `dbus-monitor` to observe signals
  - Verify signal format and content
  - Run: `./tests/manual/test_dbus_integration.sh`
  - Verify D-Bus signals are emitted correctly with real OneDrive
  - Document results in `docs/verification-tracking.md` Phase 11 section
  - _Requirements: 8.2_

- [ ] 13.4 Test D-Bus fallback with manual verification
  - Disable D-Bus (or run in environment without D-Bus)
  - Verify system continues operating
  - Check that extended attributes still work
  - Run: `./tests/manual/test_dbus_fallback.sh`
  - Verify fallback to extended attributes works with real OneDrive
  - Document results in `docs/verification-tracking.md` Phase 11 section
  - _Requirements: 8.4_

- [ ] 13.5 Test Nemo extension with manual verification
  - Open Nemo file manager
  - Navigate to mounted OneDrive
  - Verify status icons appear on files
  - Trigger file operations and watch icons update
  - Test with real OneDrive mount in Docker
  - Document results in `docs/verification-tracking.md` Phase 11 section
  - _Requirements: 8.3_

- [x] 13.6 Create file status integration tests
  - Write test for status tracking
  - Write test for D-Bus signal emission
  - Write test for extended attribute fallback
  - _Requirements: 8.1, 8.2, 8.4_

- [x] 13.7 Document file status issues and create fix plan
  - List all discovered issues
  - Identify root causes
  - Create prioritized fix plan
  - Update the relevant sections of the verification-tracking.md document
  - _Requirements: 12.1_

---

## Phase 12: Error Handling and Recovery Verification

- [x] 14. Verify error handling
- [x] 14.1 Review error handling code
  - Read and analyze `internal/errors/`
  - Review `internal/logging/` implementation
  - Review error handling throughout codebase
  - Check structured logging with zerolog
  - _Requirements: 9.1, 9.2, 9.3, 9.4, 9.5_

- [x] 14.2 Test network error handling
  - Simulate various network errors (timeout, connection refused, etc.)
  - Verify errors are logged with context
  - Check that retries occur
  - Verify eventual success or clear failure
  - _Requirements: 9.1_

- [x] 14.3 Test API rate limiting
  - Trigger many API requests rapidly
  - Verify rate limit detection
  - Check that exponential backoff is used
  - Verify operations eventually succeed
  - _Requirements: 9.2_

- [x] 14.4 Test crash recovery
  - Mount filesystem
  - Forcefully kill process (kill -9)
  - Remount filesystem
  - Verify state is recovered from database
  - Check that incomplete uploads resume
  - _Requirements: 9.3, 9.4_

- [x] 14.5 Test error messages
  - Trigger various error conditions
  - Review error messages shown to user
  - Verify messages are clear and actionable
  - Check that technical details are logged but not shown to user
  - _Requirements: 9.5_

- [x] 14.6 Create error handling integration tests
  - Write test for network error retry
  - Write test for rate limit handling
  - Write test for crash recovery
  - _Requirements: 9.1, 9.2, 9.3, 9.4_

- [x] 14.7 Document error handling issues and create fix plan
  - List all discovered issues
  - Identify root causes
  - Create prioritized fix plan
  - Update the relevant sections of the verification-tracking.md document
  - _Requirements: 12.1_

---

## Phase 13: Performance and Concurrency Verification

- [x] 15. Verify performance and concurrency
- [x] 15.1 Review concurrency implementation
  - Review goroutine usage throughout codebase
  - Check locking mechanisms (mutexes, RWMutexes)
  - Verify wait groups for cleanup
  - Look for potential race conditions or deadlocks
  - _Requirements: 10.1, 10.2, 10.3, 10.4, 10.5_

- [x] 15.2 Test concurrent file access
  - Access multiple files simultaneously from different processes
  - Verify no race conditions occur
  - Check that all operations complete successfully
  - _Requirements: 10.1_

- [x] 15.3 Test concurrent downloads
  - Trigger many downloads simultaneously
  - Verify downloads proceed concurrently
  - Check that worker pool limits are respected
  - Verify no deadlocks occur
  - _Requirements: 10.2_

- [x] 15.4 Test directory listing performance
  - List a directory with many files (100+)
  - Measure response time
  - Verify response is under 2 seconds
  - _Requirements: 10.3_

- [x] 15.5 Test locking granularity
  - Review lock usage in hot paths
  - Verify locks are held for minimal time
  - Check for unnecessary global locks
  - _Requirements: 10.4_

- [x] 15.6 Test graceful shutdown
  - Mount filesystem
  - Start several long-running operations
  - Trigger shutdown (SIGTERM)
  - Verify all goroutines complete
  - Check that wait groups are used correctly
  - _Requirements: 10.5_

- [x] 15.7 Run race detector
  - Run tests with `-race` flag
  - Run application with race detector enabled
  - Fix any detected race conditions
  - _Requirements: 10.1_

- [x] 15.8 Create performance benchmarks
  - Write benchmark for file read operations
  - Write benchmark for directory listing
  - Write benchmark for concurrent operations
  - _Requirements: 10.2, 10.3_

- [x] 15.9 Document performance issues and create fix plan
  - List all discovered issues
  - Identify bottlenecks
  - Create prioritized fix plan
  - Update the relevant sections of the verification-tracking.md document
  - _Requirements: 12.1_

---

## Phase 14: Integration and End-to-End Testing

- [ ] 16. Run comprehensive integration tests with real OneDrive
- [ ] 16.1 Test authentication to file access with real OneDrive
  - Test complete flow: authenticate → mount → list files → read file
  - Verify each step works correctly
  - Check error handling at each step
  - `docker compose -f docker/compose/docker-compose.test.yml run --rm integration-tests go test -v -run TestIT_COMPREHENSIVE ./internal/fs`
  - Verify all components work together end-to-end
  - Test complete workflows with real API
  - Verify error handling with real network conditions
  - Document results in `docs/verification-tracking.md` Phase 13 section
  - _Requirements: 11.1_

- [ ] 16.2 Test file modification to sync with real OneDrive
  - Test flow: create file → modify → upload → verify on OneDrive
  - Check that all steps complete
  - Verify file appears correctly on OneDrive
  - **Covered by TestIT_COMPREHENSIVE integration test above**
  - _Requirements: 11.2_

- [ ] 16.3 Test offline mode with real OneDrive
  - Test flow: online → access files → go offline → access cached files → go online
  - Verify offline detection works
  - Check that cached files remain accessible
  - Verify online transition works
  - **Covered by TestIT_COMPREHENSIVE integration test above**
  - _Requirements: 11.3_

- [ ] 16.4 Test conflict resolution with real OneDrive
  - Test flow: modify file locally → modify remotely → sync → verify conflict copy
  - Check that both versions are preserved
  - Verify conflict is detected correctly
  - **Covered by TestIT_COMPREHENSIVE integration test above**
  - _Requirements: 11.4_

- [ ] 16.5 Test cache cleanup with real OneDrive
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

- [ ] 17.3 Test long-running operations with real OneDrive
  - Upload a very large file (1GB+)
  - Monitor progress
  - Verify upload completes successfully
  - Test interruption and resume
  - **Retest with real OneDrive**: `docker compose -f docker/compose/docker-compose.test.yml run --rm -e RUN_E2E_TESTS=1 -e RUN_LONG_TESTS=1 system-tests go test -v -timeout 60m -run TestE2E_17_03 ./internal/fs`
  - Verify very large file uploads (1GB+)
  - Monitor progress throughout operation
  - Test interruption and resume functionality
  - Document results in `docs/verification-tracking.md` Phase 14 section
  - _Requirements: 4.3, 4.4_

- [ ] 17.4 Test stress scenarios with real OneDrive
  - Perform many concurrent operations
  - Monitor resource usage (CPU, memory, network)
  - Verify system remains stable
  - Check for memory leaks
  - **Retest with real OneDrive**: `docker compose -f docker/compose/docker-compose.test.yml run --rm -e RUN_E2E_TESTS=1 -e RUN_STRESS_TESTS=1 system-tests go test -v -timeout 30m -run TestE2E_17_04 ./internal/fs`
  - Verify many concurrent operations work correctly
  - Monitor resource usage (CPU, memory, network)
  - Verify system remains stable under load
  - Check for memory leaks
  - Document results in `docs/verification-tracking.md` Phase 14 section
  - _Requirements: 10.1, 10.2_

---

## Phase 15: Issue Resolution

- [ ] 18. Fix critical issues
  - Review all issues marked as "critical" priority
  - For each critical issue:
    - Analyze root cause
    - Design fix
    - Implement fix
    - Write test to verify fix
    - Run regression tests
    - Update documentation
  - _Requirements: All_

- [ ] 19. Fix high-priority issues
  - Review all issues marked as "high" priority
  - For each high-priority issue:
    - Analyze root cause
    - Design fix
    - Implement fix
    - Write test to verify fix
    - Run regression tests
    - Update documentation
  - _Requirements: All_

- [ ] 20. Fix medium-priority issues
  - Review all issues marked as "medium" priority
  - Prioritize based on impact and effort
  - For each selected issue:
    - Analyze root cause
    - Design fix
    - Implement fix
    - Write test to verify fix
    - Run regression tests
    - Update documentation
  - _Requirements: All_

---

## Phase 16: Documentation Updates

- [ ] 21. Update architecture documentation
  - Review `docs/2-architecture-and-design/software-architecture-specification.md`
  - Update component descriptions to match implementation
  - Update sequence diagrams if flows have changed
  - Document any architectural decisions made during fixes
  - _Requirements: 12.1_

- [ ] 22. Update design documentation
  - Review `docs/2-architecture-and-design/software-design-specification.md`
  - Update data models to match implementation
  - Update interface descriptions
  - Document design patterns used
  - _Requirements: 12.2_

- [ ] 23. Update API documentation
  - Review all public APIs
  - Ensure godoc comments are accurate
  - Update function signatures if changed
  - Document any breaking changes
  - _Requirements: 12.3_

- [ ] 24. Create troubleshooting guide
  - Document common issues discovered during verification
  - Provide solutions for each issue
  - Include diagnostic commands
  - Add to user documentation
  - _Requirements: 9.5, 12.5_

- [ ] 25. Update traceability matrix
  - Update `docs/2-architecture-and-design/sas-requirements-traceability-matrix.md`
  - Ensure all requirements are traced to implementation
  - Document test coverage for each requirement
  - _Requirements: 12.1, 12.2, 12.3_

---

## Phase 17: XDG Compliance Verification

- [ ] 26. Verify XDG Base Directory compliance
- [ ] 26.1 Review XDG implementation
  - Read and analyze `cmd/common/config.go`
  - Verify use of `os.UserConfigDir()` and `os.UserCacheDir()`
  - Check directory creation and permissions
  - Run in Docker: `docker compose -f docker/compose/docker-compose.test.yml run --rm shell`
  - _Requirements: 15.1, 15.4_

- [ ] 26.2 Test XDG_CONFIG_HOME environment variable
  - Set `XDG_CONFIG_HOME` to custom path
  - Mount filesystem in Docker container
  - Verify config stored in `$XDG_CONFIG_HOME/onemount/`
  - Verify auth tokens in config directory
  - Run in Docker: `docker compose -f docker/compose/docker-compose.test.yml run --rm shell`
  - _Requirements: 15.2, 15.7_

- [ ] 26.3 Test XDG_CACHE_HOME environment variable
  - Set `XDG_CACHE_HOME` to custom path
  - Mount filesystem in Docker container
  - Verify cache stored in `$XDG_CACHE_HOME/onemount/`
  - Verify metadata database in cache directory
  - Run in Docker: `docker compose -f docker/compose/docker-compose.test.yml run --rm shell`
  - _Requirements: 15.5, 15.9_

- [ ] 26.4 Test default XDG paths
  - Unset XDG environment variables
  - Mount filesystem in Docker container
  - Verify config in `~/.config/onemount/`
  - Verify cache in `~/.cache/onemount/`
  - Run in Docker: `docker compose -f docker/compose/docker-compose.test.yml run --rm shell`
  - _Requirements: 15.3, 15.6_

- [ ] 26.5 Test command-line override
  - Use `--config-file` and `--cache-dir` flags
  - Verify custom paths are used in Docker container
  - Verify XDG paths are not used
  - _Requirements: 15.10_

- [ ] 26.6 Test directory permissions
  - Check config directory permissions (should be 0700)
  - Check cache directory permissions (should be 0755)
  - Verify auth tokens are not world-readable
  - Run in Docker: `docker compose -f docker/compose/docker-compose.test.yml run --rm shell`
  - _Requirements: 15.7_

- [ ] 26.7 Document XDG compliance verification results
  - Update `docs/verification-tracking.md` with Phase 17 results
  - Document any issues found
  - Create fix plan if needed
  - _Requirements: 15.1-15.10_

---

## Phase 18: Webhook Subscription Verification

- [ ] 27. Verify webhook subscription implementation
- [ ] 27.1 Review subscription code
  - Read and analyze `internal/fs/subscription.go`
  - Review subscription API calls in `internal/graph/`
  - Check subscription manager implementation
  - Run in Docker: `docker compose -f docker/compose/docker-compose.test.yml run --rm shell`
  - _Requirements: 14.1, 14.2, 14.3, 14.4, 14.5_

- [ ] 27.2 Test subscription creation on mount
  - Mount filesystem in Docker container
  - Verify POST `/subscriptions` API call
  - Check subscription ID is stored
  - Verify expiration time is tracked
  - Run in Docker: `docker compose -f docker/compose/docker-compose.test.yml run --rm shell`
  - _Requirements: 14.1, 14.5, 5.2_

- [ ] 27.3 Test webhook notification reception
  - Set up webhook listener in Docker container
  - Trigger change on OneDrive
  - Verify notification is received
  - Check notification validation
  - Verify delta query is triggered
  - Run in Docker: `docker compose -f docker/compose/docker-compose.test.yml run --rm shell`
  - _Requirements: 14.6, 14.7, 5.6_

- [ ] 27.4 Test subscription renewal
  - Create subscription with short expiration
  - Wait until within 24h of expiration
  - Verify PATCH `/subscriptions/{id}` is called
  - Check new expiration time is stored
  - Run in Docker: `docker compose -f docker/compose/docker-compose.test.yml run --rm shell`
  - _Requirements: 14.9, 5.13_

- [ ] 27.5 Test subscription failure fallback
  - Simulate subscription creation failure
  - Verify system continues with polling
  - Check polling interval is shorter (5 min)
  - Run in Docker: `docker compose -f docker/compose/docker-compose.test.yml run --rm shell`
  - _Requirements: 14.10, 5.7, 5.14_

- [ ] 27.6 Test subscription deletion on unmount
  - Mount filesystem with subscription in Docker
  - Unmount filesystem
  - Verify DELETE `/subscriptions/{id}` is called
  - Run in Docker: `docker compose -f docker/compose/docker-compose.test.yml run --rm shell`
  - _Requirements: 14.12_

- [ ] 27.7 Test personal vs business subscription limits
  - Test subscription to subfolder on personal OneDrive
  - Test subscription to root only on business OneDrive
  - Verify appropriate restrictions
  - Run in Docker: `docker compose -f docker/compose/docker-compose.test.yml run --rm shell`
  - _Requirements: 5.3, 5.4_

- [ ] 27.8 Create webhook subscription integration tests
  - Write test for subscription lifecycle
  - Write test for notification handling
  - Write test for renewal logic
  - Write test for fallback to polling
  - Run in Docker: `docker compose -f docker/compose/docker-compose.test.yml run --rm integration-tests`
  - _Requirements: 14.1-14.12, 5.2-5.14_

- [ ] 27.9 Document webhook subscription verification results
  - Update `docs/verification-tracking.md` with Phase 18 results
  - Document any issues found
  - Create fix plan if needed
  - _Requirements: 14.1-14.12, 5.2-5.14_

---

## Phase 19: Multiple Account Support Verification

- [ ] 28. Verify multiple account support
- [ ] 28.1 Review multi-account code
  - Read and analyze mount manager implementation
  - Check account isolation (auth, cache, sync)
  - Review drive type handling (personal, business, shared)
  - Run in Docker: `docker compose -f docker/compose/docker-compose.test.yml run --rm shell`
  - _Requirements: 13.1, 13.6, 13.7, 13.8_

- [ ] 28.2 Test mounting personal OneDrive
  - Authenticate with personal account in Docker
  - Mount at `/mnt/onedrive-personal`
  - Verify access to `/me/drive`
  - Check files are accessible
  - Run in Docker: `docker compose -f docker/compose/docker-compose.test.yml run --rm shell`
  - _Requirements: 13.2_

- [ ] 28.3 Test mounting business OneDrive
  - Authenticate with work account in Docker
  - Mount at `/mnt/onedrive-work`
  - Verify access to `/me/drive`
  - Check files are accessible
  - Run in Docker: `docker compose -f docker/compose/docker-compose.test.yml run --rm shell`
  - _Requirements: 13.3_

- [ ] 28.4 Test simultaneous mounts
  - Mount personal OneDrive in Docker
  - Mount business OneDrive in Docker
  - Verify both are accessible
  - Check no cross-contamination
  - Run in Docker: `docker compose -f docker/compose/docker-compose.test.yml run --rm shell`
  - _Requirements: 13.1_

- [ ] 28.5 Test shared drive mount
  - Get shared drive ID
  - Mount using `/drives/{drive-id}` in Docker
  - Verify access to shared files
  - Run in Docker: `docker compose -f docker/compose/docker-compose.test.yml run --rm shell`
  - _Requirements: 13.4_

- [ ] 28.6 Test "Shared with me" access
  - Access `/me/drive/sharedWithMe` in Docker
  - Verify shared items are visible
  - Test accessing shared files
  - Run in Docker: `docker compose -f docker/compose/docker-compose.test.yml run --rm shell`
  - _Requirements: 13.5_

- [ ] 28.7 Test cache isolation
  - Mount multiple accounts in Docker
  - Access files in each mount
  - Verify separate cache directories
  - Check no cache conflicts
  - Run in Docker: `docker compose -f docker/compose/docker-compose.test.yml run --rm shell`
  - _Requirements: 13.7_

- [ ] 28.8 Test delta sync isolation
  - Mount multiple accounts in Docker
  - Trigger changes in each OneDrive
  - Verify independent delta sync loops
  - Check correct updates per mount
  - Run in Docker: `docker compose -f docker/compose/docker-compose.test.yml run --rm shell`
  - _Requirements: 13.8_

- [ ] 28.9 Create multi-account integration tests
  - Write test for simultaneous mounts
  - Write test for cache isolation
  - Write test for sync isolation
  - Write test for different drive types
  - Run in Docker: `docker compose -f docker/compose/docker-compose.test.yml run --rm integration-tests`
  - _Requirements: 13.1-13.8_

- [ ] 28.10 Document multi-account verification results
  - Update `docs/verification-tracking.md` with Phase 19 results
  - Document any issues found
  - Create fix plan if needed
  - _Requirements: 13.1-13.8_

---

## Phase 20: ETag Cache Validation Verification

- [ ] 29. Verify ETag-based cache validation with real OneDrive
- [ ] 29.1 Review ETag implementation
  - Read and analyze `internal/fs/cache.go`
  - Review `internal/fs/content_cache.go`
  - Check ETag storage in cache entries
  - Review `if-none-match` header usage
  - Run in Docker: `docker compose -f docker/compose/docker-compose.test.yml run --rm shell`
  - _Requirements: 7.1, 7.3_

- [ ] 29.2 Test cache hit with valid ETag using real OneDrive
  - Download a file (cache it)
  - Access the same file again
  - Verify `if-none-match` header is sent
  - Check 304 Not Modified response
  - Verify content served from cache
  - **Retest with real OneDrive**: `docker compose -f docker/compose/docker-compose.test.yml run --rm integration-tests go test -v -run TestIT_FS_ETag ./internal/fs`
  - Document results in `docs/verification-tracking.md` Phase 20 section
  - _Requirements: 3.4, 3.5, 7.3_

- [ ] 29.3 Test cache miss with changed ETag using real OneDrive
  - Download a file (cache it)
  - Modify file on OneDrive web interface
  - Access the file again
  - Verify `if-none-match` header is sent
  - Check 200 OK response with new content
  - Verify cache is updated with new ETag
  - **Covered by TestIT_FS_ETag integration test above**
  - _Requirements: 3.6, 7.3_

- [ ] 29.4 Test ETag updates from delta sync with real OneDrive
  - Cache several files
  - Modify files on OneDrive web interface
  - Run delta sync
  - Verify ETags are updated in metadata
  - Check cache entries are invalidated
  - **Covered by delta sync integration tests**
  - _Requirements: 5.10, 7.4_

- [ ] 29.5 Test conflict detection with ETags using real OneDrive
  - Download a file (cache it with ETag)
  - Modify file locally
  - Modify same file on OneDrive web interface (changes ETag)
  - Attempt to upload
  - Verify conflict is detected via ETag mismatch
  - Check conflict copy is created
  - **Covered by upload and delta sync tests**
  - _Requirements: 8.1, 8.2, 8.3_

- [ ] 29.6 Run ETag validation integration tests with real OneDrive
  - Run: `docker compose -f docker/compose/docker-compose.test.yml run --rm integration-tests go test -v -run TestIT_FS_ETag ./internal/fs`
  - Verify cache validation flow works with real API
  - Verify ETag-based conflict detection works with real API
  - Verify delta sync ETag updates work with real API
  - Document results in `docs/verification-tracking.md` Phase 20 section
  - _Requirements: 3.4-3.6, 7.1-7.4, 8.1-8.3_

---

## Phase 21: Final Verification

- [ ] 30. Run complete test suite in Docker
  - Build latest test images: `docker compose -f docker/compose/docker-compose.build.yml build`
  - Run all unit tests: `docker compose -f docker/compose/docker-compose.test.yml run unit-tests`
  - Run all integration tests: `docker compose -f docker/compose/docker-compose.test.yml run integration-tests`
  - Run all system tests (requires auth): `docker compose -f docker/compose/docker-compose.test.yml run system-tests`
  - Generate coverage report: `docker compose -f docker/compose/docker-compose.test.yml run coverage`
  - Review test artifacts in `test-artifacts/logs/`
  - Verify all tests pass
  - Document any remaining failures
  - _Requirements: All, 17.1, 17.2, 17.3, 17.4, 17.5_

- [ ] 31. Perform manual verification in Docker
  - Use interactive shell: `docker compose -f docker/compose/docker-compose.test.yml run shell`
  - Follow user workflows manually within container
  - Test mounting and file operations
  - Test multiple account mounts
  - Test webhook subscriptions
  - Verify all documented features work in isolated environment
  - Test with different configurations
  - Document any issues found during manual testing
  - _Requirements: All, 17.4, 17.5_

- [ ] 32. Performance verification
  - Run performance benchmarks
  - Test with webhook subscriptions (30min polling)
  - Test without subscriptions (5min polling)
  - Compare polling frequency impact
  - Verify response times meet expectations
  - Check resource usage is reasonable
  - Test multiple simultaneous mounts
  - _Requirements: 11.3, 5.5, 5.7_

- [ ] 33. Create verification report
  - Summarize all verification activities
  - List all issues found and fixed
  - Document webhook subscription behavior
  - Document multi-account support
  - Document ETag cache validation
  - Document XDG compliance
  - Document remaining known issues
  - Provide recommendations for future work
  - _Requirements: 16.1, 16.2, 16.3, 16.4, 16.5_

- [ ] 34. Final documentation review
  - Review all updated documentation
  - Ensure webhook subscription documentation is complete
  - Ensure multi-account support is documented
  - Ensure ETag validation is documented
  - Ensure XDG compliance is documented
  - Ensure consistency across documents
  - Verify all cross-references are correct
  - Check that documentation is complete
  - _Requirements: 16.1, 16.2, 16.3, 16.4, 16.5_
