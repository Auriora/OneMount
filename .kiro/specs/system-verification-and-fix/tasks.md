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

- [x] 12.4 Test offline write operations with change queuing
  - While offline, attempt to create file
  - Verify operation succeeds and change is queued
  - Attempt to modify file
  - Verify modification succeeds and is tracked
  - Verify changes are stored in persistent storage
  - _Requirements: 6.3, 6.4_

- [x] 12.5 Test multiple changes to same file offline
  - Make multiple changes to the same file while offline
  - Verify most recent version is preserved
  - Verify change tracking updates correctly
  - _Requirements: 6.5_

- [x] 12.6 Test online transition
  - While offline, reconnect network
  - Trigger operation requiring network
  - Verify online state is detected
  - Check that queued changes are processed
  - Verify delta sync resumes
  - _Requirements: 6.6_

- [x] 12.7 Create offline mode integration tests
  - Write test for offline detection
  - Write test for offline read operations
  - Write test for offline write operations with change queuing
  - Write test for multiple changes to same file
  - Write test for online transition
  - _Requirements: 6.1, 6.2, 6.3, 6.4, 6.5, 6.6_

- [x] 12.8 Document offline mode issues and create fix plan
  - List all discovered issues
  - Identify root causes
  - Create prioritized fix plan
  - Update the relevant sections of the verification-tracking.md document
  - _Requirements: 12.1_

---

## Phase 11: File Status and D-Bus Verification

- [-] 13. Verify file status tracking
- [x] 13.1 Review file status code
  - Read and analyze `internal/fs/file_status.go`
  - Review `internal/fs/dbus.go`
  - Check extended attribute implementation
  - Review Nemo extension code
  - _Requirements: 8.1, 8.2, 8.3, 8.4, 8.5_

- [x] 13.2 Test file status updates with manual verification
  - Monitor file status during various operations
  - Verify status changes appropriately (synced, downloading, error, etc.)
  - Check extended attributes are set correctly
  - Run: `./tests/manual/test_file_status_updates.sh`
  - Verify file status updates work correctly with real OneDrive
  - Document results in `docs/verification-tracking.md` Phase 11 section
  - _Requirements: 8.1_

- [x] 13.3 Test D-Bus integration with manual verification
  - Verify D-Bus server starts successfully
  - Monitor D-Bus signals during file operations
  - Use `dbus-monitor` to observe signals
  - Verify signal format and content
  - Run outside of docker: `./tests/manual/test_dbus_integration.sh`
  - Verify D-Bus signals are emitted correctly with real OneDrive
  - Document results in `docs/verification-tracking.md` Phase 11 section
  - _Requirements: 8.2_

- [ ] 13.4 Test D-Bus fallback with manual verification
  - Disable D-Bus (or run in environment without D-Bus)
  - Verify system continues operating
  - Check that extended attributes still work
  - Run outside of docker: `./tests/manual/test_dbus_fallback.sh`
  - Verify fallback to extended attributes works with real OneDrive
  - Document results in `docs/verification-tracking.md` Phase 11 section
  - _Requirements: 8.4_

- [x] 13.5 Test Nemo extension with manual verification
  - Open Nemo file manager
  - Navigate to mounted OneDrive
  - Verify status icons appear on files
  - Trigger file operations and watch icons update
  - Test with real OneDrive mount outside Docker
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

## Phase 15: Issue Resolution

- [x] 18. Fix critical issues
  - Review all issues marked as "critical" priority in `docs/verification-tracking.md`
  - **Status**: No critical issues identified (0 issues)
  - _Requirements: All_

- [ ] 19. Fix high-priority issues
  - Review all issues marked as "high" priority in `docs/verification-tracking.md`
  - **Status**: 2 high-priority issues identified
  - _Requirements: All_

- [ ] 19.1 Fix Issue #010: Large File Upload Retry Logic Not Working
  - **Component**: Upload Manager / Upload Session
  - **Action**: Fix chunk upload retry mechanism
  - **Tasks**:
    - Review chunk upload retry logic in `PerformChunkedUpload()`
    - Verify retry attempt tracking is working correctly
    - Ensure exponential backoff is applied between retries
    - Fix ETag update after successful upload
    - Fix file status transition after upload completion
    - Add logging to track retry attempts
    - Update tests to verify retry behavior
  - **Test**: `docker compose -f docker/compose/docker-compose.test.yml run --rm test-runner go test -v -run TestIT_FS_09_04_02_LargeFileUploadFailureAndRetry ./internal/fs`
  - **Estimate**: 4-6 hours
  - _Requirements: 4.4, 4.5, 8.1_

- [ ] 19.2 Fix Issue #011: Upload Max Retries Exceeded Not Working
  - **Component**: Upload Manager / Upload Session
  - **Action**: Fix upload session state machine for max retries
  - **Tasks**:
    - Review upload session state machine transitions
    - Ensure Error state (3) is set when max retries exceeded
    - Update file status determination to reflect upload errors
    - Add user notification for upload failures
    - Ensure file remains accessible locally after failure
    - Add logging for max retries exceeded
    - Update tests to verify error state behavior
  - **Test**: `docker compose -f docker/compose/docker-compose.test.yml run --rm test-runner go test -v -run TestIT_FS_09_04_03_UploadMaxRetriesExceeded ./internal/fs`
  - **Estimate**: 3-4 hours
  - _Requirements: 4.4, 8.1, 9.5_

- [-] 20. Fix medium-priority issues
  - Review all issues marked as "medium" priority in `docs/verification-tracking.md`
  - **Total**: 16 medium-priority issues identified
  - Prioritize based on impact and effort
  
- [x] 20.1 Fix Issue #001: Mount Timeout in Docker Container
  - **Status**: ✅ RESOLVED (2025-11-12)
  - **Component**: Filesystem Mounting
  - **Fix**: Added `--mount-timeout` flag with configurable timeout
  - **Documentation**: `docs/fixes/mount-timeout-fix.md`
  - _Requirements: 2.1, 2.2_

- [x] 20.2 Fix Issue #002: ETag-Based Cache Validation Location Unclear
  - **Component**: File Operations / Download Manager
  - **Action**: Review download manager and Graph API layer for ETag validation
  - **Tasks**:
    - Review `internal/fs/download_manager.go` for ETag validation
    - Review `internal/graph/` HTTP request code for `if-none-match` header
    - Update design documentation to clarify where ETag validation occurs
    - Add integration tests to verify 304 Not Modified handling
    - Add code comments explaining the validation flow
  - **Estimate**: 4 hours
  - _Requirements: 3.4, 3.5, 3.6_

- [x] 20.3 Fix Issue #008: Upload Manager - Memory Usage for Large Files
  - **Component**: Upload Manager
  - **Action**: Implement streaming upload for files > 100MB
  - **Tasks**:
    - Replace `Data []byte` field with `ContentReader io.ReadSeeker`
    - Implement chunked reading from disk
    - Add memory usage metrics to upload manager
    - Test with large files (> 100MB)
    - Document memory requirements
  - **Estimate**: 8 hours
  - _Requirements: 4.3, 11.1_

- [x] 20.4 Fix Issue #OF-001: Read-Write vs Read-Only Offline Mode
  - **Component**: Offline Mode
  - **Action**: Update requirements to match implementation (RECOMMENDED)
  - **Tasks**:
    - Update Requirement 6.3 to specify read-write offline mode with change queuing
    - Update design documentation to reflect current behavior
    - Add explicit requirement for change queuing
    - Document conflict resolution strategy for offline changes
  - **Estimate**: 1 hour (requirements update)
  - _Requirements: 6.3_

- [x] 20.5 Fix Issue #FS-001: D-Bus GetFileStatus Returns Unknown
  - **Component**: File Status / D-Bus Server
  - **Action**: Add GetPath() method or implement path-to-ID mapping
  - **Tasks**:
    - Option 1: Add `GetPath(id string) string` method to FilesystemInterface
    - Option 2: Implement path-to-ID mapping in D-Bus server
    - Update GetFileStatus to return actual status
    - Add integration tests for D-Bus method calls
  - **Estimate**: 2-3 hours
  - _Requirements: 8.2_

- [ ] 20.6 Fix Issue #PERF-001: No Documented Lock Ordering Policy
  - **Component**: Concurrency / Locking
  - **Action**: Document lock ordering policy
  - **Tasks**:
    - Create `docs/guides/developer/concurrency-guidelines.md`
    - Document lock acquisition order (e.g., filesystem lock before inode lock)
    - Add examples of correct usage
    - Update code comments with lock ordering notes
    - Review all code that acquires multiple locks
  - **Estimate**: 4 hours
  - _Requirements: 10.1, 10.4_

- [x] 20.7 Fix Issue #PERF-002: Network Callbacks Lack Wait Group Tracking
  - **Component**: Network Feedback / Goroutine Management
  - **Action**: Add wait group tracking for callback goroutines
  - **Tasks**:
    - Add WaitGroup to NetworkFeedbackManager
    - Track callback goroutines with WaitGroup.Add/Done
    - Add timeout for callback completion during shutdown
    - Test graceful shutdown with active callbacks
  - **Estimate**: 2 hours
  - _Requirements: 10.5_

- [ ] 20.8 Fix Issue #PERF-003: Inconsistent Timeout Values
  - **Component**: Shutdown / Configuration
  - **Action**: Standardize timeout values across components
  - **Tasks**:
    - Create configuration for timeout values
    - Update all managers to use standard timeouts
    - Document timeout policy
    - Make timeouts configurable via command-line or config file
  - **Estimate**: 2 hours
  - _Requirements: 10.5_

- [x] 20.9 Fix Issue #PERF-004: Inode Embeds Mutex
  - **Component**: Inode / Locking
  - **Action**: Change Inode to use pointer to mutex
  - **Tasks**:
    - Modify Inode struct to use `*sync.RWMutex` instead of embedded mutex
    - Update all code that accesses inode mutex
    - Run full test suite to verify no regressions
    - Run race detector to verify thread safety
  - **Estimate**: 4 hours
  - _Requirements: 10.1_

- [x] 20.10 Fix Issue #CACHE-001: No Cache Size Limit Enforcement
  - **Component**: Cache Management
  - **Action**: Implement LRU eviction with size limit
  - **Tasks**:
    - Add cache size tracking to LoopbackCache
    - Implement LRU eviction algorithm
    - Add configuration for max cache size
    - Update CleanupCache to enforce size limits
    - Add cache size metrics to GetStats()
    - Document cache management behavior
  - **Estimate**: 6-8 hours
  - _Requirements: 7.2, 7.3_

- [x] 20.11 Fix Issue #CACHE-002: No Explicit Cache Invalidation When ETag Changes
  - **Component**: Cache Management / Delta Sync
  - **Action**: Add explicit cache invalidation in delta sync
  - **Tasks**:
    - Add explicit cache invalidation when ETag changes in delta sync
    - Call `content.Delete(id)` for modified files
    - Mark file status as OutofSync
    - Add integration test for ETag-based invalidation
    - Document cache invalidation behavior
  - **Estimate**: 3-4 hours
  - _Requirements: 7.3, 7.4, 5.3_

- [x] 20.12 Fix Issue #CACHE-003: Statistics Collection Slow for Large Filesystems
  - **Component**: Cache Management / Statistics
  - **Action**: Optimize statistics collection
  - **Tasks**:
    - Implement incremental statistics updates
    - Cache frequently accessed statistics with TTL
    - Use background goroutines for expensive calculations
    - Implement sampling for very large datasets
    - Add pagination support for statistics display
    - Optimize database queries with better indexing
  - **Estimate**: 8-12 hours
  - _Requirements: 7.5, 10.3_

- [x] 20.13 Fix Issue #CACHE-004: Fixed 24-Hour Cleanup Interval
  - **Component**: Cache Management
  - **Action**: Make cleanup interval configurable
  - **Tasks**:
    - Add `--cache-cleanup-interval` command-line flag
    - Add configuration option to config file
    - Update `StartCacheCleanup()` to use configured interval
    - Document cleanup interval configuration
    - Add validation for reasonable intervals (1 hour to 30 days)
  - **Estimate**: 2-3 hours
  - _Requirements: 7.2_

- [x] 20.14 Fix Issue #FS-003: No Error Handling for Extended Attributes
  - **Component**: File Status
  - **Action**: Add error handling for xattr operations
  - **Tasks**:
    - Add error handling for xattr operations in updateFileStatus()
    - Log warnings when xattr operations fail
    - Track xattr support status per mount point
    - Document filesystem requirements for full functionality
    - Consider adding status to GetStats() output
  - **Estimate**: 1-2 hours
  - _Requirements: 8.1, 8.4_

- [x] 20.15 Fix Issue #FS-004: Status Determination Performance
  - **Component**: File Status
  - **Action**: Optimize status determination
  - **Tasks**:
    - Profile status determination performance
    - Add caching of determination results with TTL
    - Batch database queries for multiple files
    - Optimize hash calculation (only when needed)
    - Add invalidation on relevant events
    - Consider lazy evaluation for non-visible files
  - **Estimate**: 4-6 hours
  - _Requirements: 8.1, 10.3_

- [x] 20.16 Fix Issue #FS-002: D-Bus Service Name Discovery Problem
  - **Component**: D-Bus Server / Nemo Extension
  - **Action**: Implement service discovery mechanism
  - **Tasks**:
    - Option 1: Use well-known service name without unique suffix
    - Option 2: Implement service discovery mechanism (e.g., via D-Bus introspection)
    - Option 3: Write service name to known location (e.g., /tmp/onemount-dbus-name)
    - Update Nemo extension to discover service name
    - Test with multiple OneMount instances
  - **Estimate**: 3-4 hours
  - _Requirements: 8.2, 8.3_

- [x] 21. Address ACTION REQUIRED items
  - Review all "ACTION REQUIRED" items in verification documents
  - **Total**: 8 action items identified

- [x] 21.1 Update Requirements for Offline Mode
  - **Source**: Issue #OF-001, `docs/verification-tracking.md`
  - **Action**: Update Requirement 6.3 to specify read-write offline mode
  - **Tasks**:
    - Update Requirement 6.3: "WHILE offline, THE OneMount System SHALL allow read and write operations with changes queued for synchronization when connectivity is restored"
    - Add Requirement 6.3.1: "WHEN a file is modified offline, THE OneMount System SHALL track the change in persistent storage for later upload"
    - Add Requirement 6.3.2: "WHEN multiple changes are made to the same file offline, THE OneMount System SHALL preserve the most recent version for upload"
    - Ensure requirements match implementation for change queuing
    - Ensure requirements match implementation for online transition
    - Ensure requirements match implementation for file status integration
  - **Estimate**: 2 hours
  - _Requirements: 6.3, 6.4, 6.5_

- [x] 21.2 Document Mounting Features in Requirements
  - **Source**: `docs/verification-phase4-mounting.md`
  - **Action**: Update requirements to document daemon mode and stale lock detection
  - **Tasks**:
    - Add requirement for daemon mode functionality (background operation)
    - Add requirement for stale lock file detection and cleanup mechanism (>5 minutes threshold)
    - Document mount timeout configuration
    - Update design documentation with these features
  - **Estimate**: 1 hour
  - _Requirements: 2.1, 2.2_

- [x] 21.3 Add Download Manager Configuration Requirements
  - **Source**: `docs/verification-phase5-download-manager-review.md`
  - **Action**: Add requirements for configurable download manager parameters
  - **Tasks**:
    - Add requirement for worker pool size configuration (default: 3, range: 1-10)
    - Add requirement for recovery attempts limit (default: 3, range: 1-10)
    - Add requirement for queue size configuration (default: 500, range: 100-5000)
    - Add requirement for chunk size configuration (default: 10MB, range: 1MB-100MB)
    - Document reasonable defaults and valid ranges
  - **Estimate**: 1 hour
  - _Requirements: 3.2, 3.4, 3.5_

- [x] 21.4 Document Cache Behavior for Deleted Files
  - **Source**: `docs/verification-phase4-file-write-operations.md`
  - **Action**: Document why deleted files remain in cache
  - **Tasks**:
    - Add to cache management requirements
    - deleted files should be removed from the cache
    - Update design documentation
  - **Estimate**: 1 hour
  - _Requirements: 7.1, 7.2_

- [x] 21.5 Add Directory Deletion Testing
  - **Source**: `docs/verification-phase4-file-write-operations.md`
  - **Action**: Ensure directory deletion is properly tested and documented
  - **Tasks**:
    - Add unit tests for directory deletion logic (without server sync)
    - Add integration tests with real OneDrive to verify server synchronization
    - Verify directory deletion is properly handled in the code
    - Note: Task 5.4 retest results already verified directory operations work correctly
  - **Estimate**: 3 hours
  - _Requirements: 4.1_

- [x] 21.6 Review Offline Functionality Documentation
  - **Source**: `docs/verification-tracking.md` Phase 9
  - **Action**: Review docs/offline-functionality.md for requirements/design elements
  - **Tasks**:
    - Review `docs/offline-functionality.md` (if exists)
    - Identify requirements or design elements worth incorporating
    - Update requirements specification with any missing elements
    - Ensure consistency between documentation and implementation
  - **Estimate**: 1 hour
  - _Requirements: 6.1-6.5_

- [x] 21.7 Make XDG Volume Info Files Virtual
  - **Source**: Issue #XDG-001, `docs/verification-tracking.md`
  - **Action**: Ensure .xdg-volume-info files are virtual (not synced to OneDrive)
  - **Tasks**:
    - Review `cmd/common/xdg.go` CreateXDGVolumeInfo function
    - Investigate why file causes I/O errors
    - Make file virtual (not synced to OneDrive)
    - Update the requirements with this
    - Fix file permissions and attributes
    - Test with various desktop environments
    - Add error handling to prevent I/O errors
  - **Estimate**: 1-2 hours
  - _Requirements: 15.1_

- [x] 21.8 Add Requirements for User Notifications
  - **Source**: Issues #OF-002, #OF-003, #OF-004
  - **Action**: Add requirements for offline state notifications and visibility
  - **Tasks**:
    - Add requirement for D-Bus notifications for offline state changes
    - Add requirement for user visibility of offline status
    - Add requirement for cache status information for offline planning
    - Add requirement for manual offline mode option (command-line/config)
    - Document expected user experience for offline mode
  - **Estimate**: 1 hour
  - _Requirements: 6.1, 6.2_

---

## Phase 16: Documentation Updates

- [ ] 22. Update documentation
- [ ] 22.1 Update architecture documentation
  - Review `docs/2-architecture-and-design/software-architecture-specification.md`
  - Update component descriptions to match implementation
  - Update sequence diagrams if flows have changed
  - Document any architectural decisions made during fixes
  - _Requirements: 12.1_

- [ ] 22.2 Update design documentation
  - Review `docs/2-architecture-and-design/software-design-specification.md`
  - Update data models to match implementation
  - Update interface descriptions
  - Document design patterns used
  - _Requirements: 12.2_

- [ ] 22.3 Update API documentation
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
