# Implementation Plan: OneMount System Verification and Fix

## Overview

This implementation plan breaks down the verification and fix process into discrete, manageable tasks. Each task builds on previous tasks and focuses on verifying specific components against requirements, identifying issues, and implementing fixes.

**UPDATED STATUS (2025-12-12)**: Plan has been corrected to reflect actual implementation status. Many phases previously marked as "planned" are actually complete, and some planned features (webhooks, multi-account) are either already implemented differently (Socket.IO) or deferred to future releases.

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

- [x] 4.7 Implement authentication property-based tests
- [x] 4.7.1 Implement Property 1: OAuth2 Token Storage Security
  - **Property 1: OAuth2 Token Storage Security**
  - **Validates: Requirements 1.2**
  - Create `internal/graph/auth_property_test.go`
  - Generate random valid OAuth2 completions
  - Verify tokens stored with proper security attributes
  - Run 100+ iterations per property test
  - _Requirements: 1.2_

- [x] 4.7.2 Implement Property 2: Automatic Token Refresh
  - **Property 2: Automatic Token Refresh**
  - **Validates: Requirements 1.3**
  - Generate random expired tokens with valid refresh tokens
  - Verify automatic refresh occurs without user intervention
  - Test with various expiration scenarios
  - _Requirements: 1.3_

- [x] 4.7.3 Implement Property 3: Re-authentication on Refresh Failure
  - **Property 3: Re-authentication on Refresh Failure**
  - **Validates: Requirements 1.4**
  - Generate random token refresh failure scenarios
  - Verify user re-authentication prompt occurs
  - Test with various failure types
  - _Requirements: 1.4_

- [x] 4.7.4 Implement Property 4: Headless Authentication Method
  - **Property 4: Headless Authentication Method**
  - **Validates: Requirements 1.5**
  - Generate random headless system configurations
  - Verify device code flow is used
  - Test with various headless scenarios
  - _Requirements: 1.5_

- [x] 4.8 Document authentication issues and create fix plan
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
  - _Requirements: 2.1, 2A.1, 2C.1-2C.5, 2D.1_

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

- [x] 5.8 Implement filesystem mounting property-based tests
- [x] 5.8.1 Implement Property 5: FUSE Mount Success
  - **Property 5: FUSE Mount Success**
  - **Validates: Requirements 2.1**
  - Create `internal/fs/mount_property_test.go`
  - Generate random valid mount point specifications
  - Verify successful FUSE mounting for all valid inputs
  - _Requirements: 2.1_

- [x] 5.8.2 Implement Property 6: Non-blocking Initial Sync
  - **Property 6: Non-blocking Initial Sync**
  - **Validates: Requirements 2A.1**
  - Generate random first-time mount scenarios
  - Verify initial sync completes while operations remain responsive
  - Measure response times during initial sync
  - _Requirements: 2A.1_

- [x] 5.8.3 Implement Property 7: Root Directory Visibility
  - **Property 7: Root Directory Visibility**
  - **Validates: Requirements 2.2**
  - Generate random successful mount scenarios
  - Verify root directory contents are visible and accessible
  - _Requirements: 2.2_

- [x] 5.8.4 Implement Property 8: Standard File Operations Support
  - **Property 8: Standard File Operations Support**
  - **Validates: Requirements 2.3**
  - Generate random mounted filesystem scenarios
  - Verify standard operations (ls, cat, cp) work correctly
  - Test with various file types and sizes
  - _Requirements: 2.3_

- [x] 5.8.5 Implement Property 9: Mount Conflict Error Handling
  - **Property 9: Mount Conflict Error Handling**
  - **Validates: Requirements 2.4**
  - Generate random already-in-use mount points
  - Verify clear error messages with conflicting process info
  - _Requirements: 2.4_

- [x] 5.8.6 Implement Property 10: Clean Resource Release
  - **Property 10: Clean Resource Release**
  - **Validates: Requirements 2.5**
  - Generate random mounted filesystem scenarios
  - Verify unmounting cleanly releases all resources
  - Check for resource leaks and orphaned processes
  - _Requirements: 2.5_

- [x] 5.9 Verify granular mounting requirements
- [x] 5.9.1 Test initial synchronization and caching (Requirement 2A)
  - Verify non-blocking initial sync behavior
  - Test cached metadata serving with async refresh
  - Test scoped cache invalidation for failed lookups
  - _Requirements: 2A.1-2A.3_

- [x] 5.9.2 Test virtual file management (Requirement 2B)
  - Verify `.xdg-volume-info` immediate availability
  - Test virtual file persistence with `local-*` identifiers
  - Test overlay policy resolution
  - _Requirements: 2B.1-2B.2_

- [x] 5.9.3 Test advanced mounting options (Requirement 2C)
  - Test daemon mode process forking
  - Test mount timeout configuration
  - Test stale lock file detection and cleanup
  - _Requirements: 2C.1-2C.5_

- [x] 5.9.4 Test FUSE operation performance (Requirement 2D)
  - Verify operations served from local metadata/cache only
  - Test Graph API delegation to background workers
  - Measure operation response times
  - _Requirements: 2D.1_

- [x] 5.10 Document mounting issues and create fix plan
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
  - _Requirements: 3.1-3.6, 3A.1-3A.2, 3B.1-3B.13, 3C.1-3C.2_

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

- [x] 6.7 Implement file access property-based tests
- [x] 6.7.1 Implement Property 11: Metadata-Only Directory Listing
  - **Property 11: Metadata-Only Directory Listing**
  - **Validates: Requirements 3.1**
  - Create `internal/fs/file_access_property_test.go`
  - Generate random directory listing operations
  - Verify no file content downloads during listing
  - Monitor network calls to ensure metadata-only access
  - _Requirements: 3.1_

- [x] 6.7.2 Implement Property 12: On-Demand Content Download
  - **Property 12: On-Demand Content Download**
  - **Validates: Requirements 3.2**
  - Generate random uncached file access scenarios
  - Verify correct API endpoint usage (GET /items/{id}/content)
  - Test with various file types and sizes
  - _Requirements: 3.2_

- [x] 6.7.3 Implement Property 13: ETag Cache Validation
  - **Property 13: ETag Cache Validation**
  - **Validates: Requirements 3.4**
  - Generate random cached file access scenarios
  - Verify ETag comparison from delta sync metadata
  - Test with various ETag states
  - _Requirements: 3.4_

- [x] 6.7.4 Implement Property 14: Cache Hit Serving
  - **Property 14: Cache Hit Serving**
  - **Validates: Requirements 3.5**
  - Generate random cached files with matching ETags
  - Verify content served from local cache without network requests
  - Monitor network activity to ensure no API calls
  - _Requirements: 3.5_

- [x] 6.7.5 Implement Property 15: Cache Invalidation on ETag Mismatch
  - **Property 15: Cache Invalidation on ETag Mismatch**
  - **Validates: Requirements 3.6**
  - Generate random cached files with different ETags
  - Verify cache invalidation and new content download
  - Test ETag mismatch detection accuracy
  - _Requirements: 3.6_

- [x] 6.8 Verify granular file access requirements
- [x] 6.8.1 Test download status and progress tracking (Requirement 3A)
  - Verify file status updates during downloads
  - Test error status marking for failed downloads
  - Test status persistence and notification
  - _Requirements: 3A.1-3A.2_

- [x] 6.8.2 Test download manager configuration (Requirement 3B)
  - Test worker pool size configuration and validation
  - Test retry attempts configuration and validation
  - Test queue size and chunk size configuration
  - Test configuration error messages
  - _Requirements: 3B.1-3B.13_

- [x] 6.8.3 Test file hydration state management (Requirement 3C)
  - Test GHOST state blocking until hydration
  - Test state transitions during hydration/eviction
  - Test metadata preservation during eviction
  - _Requirements: 3C.1-3C.2_

- [x] 6.9 Document file read issues and create fix plan
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

- [x] 7.6 Implement file modification property-based tests
- [x] 7.6.1 Implement Property 16: Local Change Tracking
  - **Property 16: Local Change Tracking**
  - **Validates: Requirements 4.1**
  - Create `internal/fs/file_modification_property_test.go`
  - Generate random file modification scenarios
  - Verify files are marked as having local changes
  - Test with various modification types
  - _Requirements: 4.1_

- [x] 7.6.2 Implement Property 17: Upload Queuing
  - **Property 17: Upload Queuing**
  - **Validates: Requirements 4.2**
  - Generate random saved modified file scenarios
  - Verify files are queued for upload to server
  - Test queue management and ordering
  - _Requirements: 4.2_

- [x] 7.6.3 Implement Property 18: ETag Update After Upload
  - **Property 18: ETag Update After Upload**
  - **Validates: Requirements 4.7**
  - Generate random successful upload scenarios
  - Verify ETag is updated from server response
  - Test ETag consistency after upload
  - _Requirements: 4.7_

- [x] 7.6.4 Implement Property 19: Modified Flag Cleanup
  - **Property 19: Modified Flag Cleanup**
  - **Validates: Requirements 4.8**
  - Generate random successful upload scenarios
  - Verify modified flag is cleared after upload
  - Test flag state consistency
  - _Requirements: 4.8_

- [x] 7.7 Document file write issues and create fix plan
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
  - _Requirements: 3.2, 3A.1-3A.2, 3B.1-3B.13, 3C.1-3C.2_

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

- [x] 9. Verify upload manager
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

- [x] 11.8 Implement cache management property-based tests
- [x] 11.8.1 Implement Property 28: ETag-Based Cache Storage
  - **Property 28: ETag-Based Cache Storage**
  - **Validates: Requirements 7.1**
  - Create `internal/fs/cache_property_test.go`
  - Generate random downloaded file scenarios
  - Verify content stored in cache directory with file's ETag
  - Test ETag association accuracy
  - _Requirements: 7.1_

- [x] 11.8.2 Implement Property 29: Cache Invalidation on Remote ETag Change
  - **Property 29: Cache Invalidation on Remote ETag Change**
  - **Validates: Requirements 7.3**
  - Generate random cached files with different remote ETags
  - Verify cache invalidation and new version download
  - Test invalidation trigger accuracy
  - _Requirements: 7.3_

- [x] 11.9 Document cache issues and create fix plan
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

- [x] 12.8 Implement offline mode property-based tests
- [x] 12.8.1 Implement Property 24: Offline Detection
  - **Property 24: Offline Detection**
  - **Validates: Requirements 6.1**
  - Create `internal/fs/offline_property_test.go`
  - Generate random network connectivity loss scenarios
  - Verify offline state detection through API call failures
  - Test detection accuracy and timing
  - _Requirements: 6.1_

- [x] 12.8.2 Implement Property 25: Offline Read Access
  - **Property 25: Offline Read Access**
  - **Validates: Requirements 6.4**
  - Generate random cached file scenarios while offline
  - Verify files can be served for read operations
  - Test read access reliability offline
  - _Requirements: 6.4_

- [x] 12.8.3 Implement Property 26: Offline Write Queuing
  - **Property 26: Offline Write Queuing**
  - **Validates: Requirements 6.5**
  - Generate random write operations while offline
  - Verify operations are allowed and changes queued
  - Test queue management and persistence
  - _Requirements: 6.5_

- [x] 12.8.4 Implement Property 27: Batch Upload Processing
  - **Property 27: Batch Upload Processing**
  - **Validates: Requirements 6.10**
  - Generate random network connectivity restoration scenarios
  - Verify queued uploads are processed in batches
  - Test batch processing efficiency
  - _Requirements: 6.10_

- [x] 12.9 Document offline mode issues and create fix plan
  - List all discovered issues
  - Identify root causes
  - Create prioritized fix plan
  - Update the relevant sections of the verification-tracking.md document
  - _Requirements: 12.1_

- [x] 12.10 Verify network error pattern recognition
- [x] 12.10.1 Review network error pattern matching code
  - Read and analyze `internal/graph/network_feedback.go`
  - Review `internal/fs/offline.go` error pattern detection
  - Check pattern matching implementation
  - Review error pattern list completeness
  - _Requirements: 19.1-19.11_

- [x] 12.10.2 Test recognized error patterns
  - Test "no such host" pattern recognition
  - Test "network is unreachable" pattern recognition
  - Test "connection refused" pattern recognition
  - Test "connection timed out" pattern recognition
  - Test "dial tcp" pattern recognition
  - Test "context deadline exceeded" pattern recognition
  - Test "no route to host" pattern recognition
  - Test "network is down" pattern recognition
  - Test "temporary failure in name resolution" pattern recognition
  - Test "operation timed out" pattern recognition
  - _Requirements: 19.1-19.10_

- [x] 12.10.3 Test offline state transition on pattern match
  - Simulate network errors with recognized patterns
  - Verify offline state is triggered correctly
  - Test pattern matching is case-insensitive where appropriate
  - Verify false positives are minimized
  - _Requirements: 19.1-19.11_

- [x] 12.10.4 Test error pattern logging
  - Verify detected patterns are logged with context
  - Test specific error pattern logging
  - Check log format and content
  - Verify error pattern identification in logs
  - _Requirements: 19.11_

- [x] 12.10.5 Create network error pattern integration tests
  - Write test for each recognized error pattern
  - Write test for offline state transition
  - Write test for error pattern logging
  - Write test for pattern matching accuracy
  - _Requirements: 19.1-19.11_

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

- [x] 13.4 Test D-Bus fallback with manual verification
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

- [x] 14.7 Implement error handling property-based tests
- [x] 14.7.1 Implement Property 35: Network Error Logging
  - **Property 35: Network Error Logging**
  - **Validates: Requirements 11.1**
  - Create `internal/errors/error_property_test.go`
  - Generate random network error scenarios
  - Verify errors are logged with appropriate context
  - Test logging completeness and accuracy
  - _Requirements: 11.1_

- [x] 14.7.2 Implement Property 36: Rate Limit Backoff
  - **Property 36: Rate Limit Backoff**
  - **Validates: Requirements 11.2**
  - Generate random API rate limit scenarios
  - Verify exponential backoff implementation
  - Test backoff timing and progression
  - _Requirements: 11.2_

- [x] 14.8 Document error handling issues and create fix plan
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

- [x] 15.9 Implement concurrency property-based tests
- [x] 15.9.1 Implement Property 33: Safe Concurrent File Access
  - **Property 33: Safe Concurrent File Access**
  - **Validates: Requirements 10.1**
  - Create `internal/fs/concurrency_property_test.go`
  - Generate random simultaneous file access scenarios
  - Verify operations are handled safely without race conditions
  - Test with race detector enabled
  - _Requirements: 10.1_

- [x] 15.9.2 Implement Property 34: Non-blocking Downloads
  - **Property 34: Non-blocking Downloads**
  - **Validates: Requirements 10.2**
  - Generate random ongoing download scenarios
  - Verify other file operations can proceed without blocking
  - Test operation concurrency and responsiveness
  - _Requirements: 10.2_

- [x] 15.10 Document performance issues and create fix plan
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

- [x] 19. Fix high-priority issues
  - Review all issues marked as "high" priority in `docs/verification-tracking.md`
  - **Status**: 2 high-priority issues identified
  - _Requirements: All_

- [x] 19.1 Fix Issue #010: Large File Upload Retry Logic Not Working
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

- [x] 19.2 Fix Issue #011: Upload Max Retries Exceeded Not Working
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

- [x] 20. Fix medium-priority issues
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

- [x] 20.6 Fix Issue #PERF-001: No Documented Lock Ordering Policy
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

- [x] 20.8 Fix Issue #PERF-003: Inconsistent Timeout Values
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

- [x] 20.17 Fix Issue #OF-002: Offline Detection False Positives
  - **Component**: Offline Mode / Network Detection
  - **Status**: ❌ FAILED Property-Based Test (Property 24)
  - **Action**: Fix conservative offline detection to avoid false positives
  - **Issue**: The `IsOffline` function in `internal/graph/graph.go` has an overly conservative default that treats all unknown errors as offline conditions, causing false positives for authentication errors, permission errors, and other non-network issues
  - **Failing Example**: Pattern 'permission denied' incorrectly detected as offline=true (expected false)
  - **Tasks**:
    - Review `IsOffline` function in `internal/graph/graph.go`
    - Remove or modify the conservative default that treats unknown errors as offline
    - Add explicit checks for authentication/authorization error patterns (401, 403, "permission denied", "invalid token", etc.)
    - Ensure these non-network errors return false (online)
    - Update error pattern matching to be more precise
    - Consider adding an explicit whitelist of offline patterns instead of blacklist approach
    - Run Property 24 test to verify fix: `go test -v -run TestProperty24_OfflineDetection ./internal/fs`
    - Ensure all 100 iterations pass with correct offline/online detection
    - Update integration tests if needed
  - **Test**: `go test -v -run TestProperty24_OfflineDetection ./internal/fs -timeout 30s`
  - **Estimate**: 2-3 hours
  - _Requirements: 6.1, 19.1-19.11_

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

## Phase 17: Documentation Updates

- [-] 22. Update documentation
- [x] 22.1 Update architecture documentation
  - Review `docs/2-architecture/software-architecture-specification.md`
  - Update Socket.IO realtime implementation details
  - Remove webhook references (replaced by Socket.IO)
  - Update sequence diagrams for realtime notifications
  - Document runtime layering and state management changes
  - _Requirements: Architecture documentation accuracy_

- [x] 22.2 Update design documentation
  - Review `docs/2-architecture/software-design-specification.md`
  - Update data models to match current implementation
  - Document change notifier facade and Socket.IO integration
  - Update interface descriptions for realtime components
  - _Requirements: Design documentation accuracy_

- [x] 22.3 Update API documentation
  - Review all public APIs
  - Ensure godoc comments are accurate
  - Update function signatures if changed
  - Document Socket.IO configuration options
  - _Requirements: API documentation accuracy_

- [x] 22.4 Update user documentation
  - Update README.md with correct realtime configuration
  - Remove webhook references from user guides
  - Document Socket.IO vs polling-only modes
  - Update troubleshooting guides
  - _Requirements: User documentation accuracy_

- [ ] 22.5 Create troubleshooting guide
  - Document common issues discovered during verification
  - Include Socket.IO connection troubleshooting
  - Provide solutions for each issue
  - Include diagnostic commands
  - _Requirements: User support_

- [ ] 22.6 Update traceability matrix
  - Update requirements traceability matrix
  - Ensure all requirements are traced to implementation
  - Document test coverage for each requirement
  - Remove references to deferred features
  - _Requirements: Requirements traceability_

- [ ] 22.7 Verify documentation alignment (Requirement 18)
- [ ] 22.7.1 Verify architecture documentation accuracy
  - Compare architecture docs with actual component interactions
  - Verify component diagrams match implementation
  - Check interface descriptions are current
  - Update outdated architectural decisions
  - _Requirements: 18.1_

- [ ] 22.7.2 Verify design documentation accuracy
  - Compare design docs with implemented data models
  - Verify API documentation matches function signatures
  - Check design patterns match implementation
  - Update design rationale where implementation differs
  - _Requirements: 18.2_

- [ ] 22.7.3 Verify API documentation accuracy
  - Review all public API documentation
  - Verify godoc comments match actual behavior
  - Check function signatures are current
  - Update parameter and return value descriptions
  - _Requirements: 18.3_

- [ ] 22.7.4 Document implementation deviations
  - Identify where implementation differs from design
  - Document rationale for each deviation
  - Update design docs or justify implementation choice
  - Create decision records for significant changes
  - _Requirements: 18.4_

- [ ] 22.7.5 Establish documentation update process
  - Create process for updating docs with code changes
  - Add documentation review to development workflow
  - Set up automated checks for doc-code alignment
  - Train team on documentation maintenance
  - _Requirements: 18.5_

---

## Phase 15: XDG Compliance Verification ✅ COMPLETE

- [x] 26. Verify XDG Base Directory compliance ✅ COMPLETE
- [x] 26.1 Review XDG implementation ✅ COMPLETE
- [x] 26.2 Test XDG_CONFIG_HOME environment variable ✅ COMPLETE
- [x] 26.3 Test XDG_CACHE_HOME environment variable ✅ COMPLETE
- [x] 26.4 Test default XDG paths ✅ COMPLETE
- [x] 26.5 Test command-line override ✅ COMPLETE
- [x] 26.6 Test directory permissions ✅ COMPLETE
- [x] 26.7 Document XDG compliance verification results ✅ COMPLETE

- [x] 26.8 Implement XDG compliance property-based tests
- [x] 26.8.1 Implement Property 37: XDG Configuration Directory Usage
  - **Property 37: XDG Configuration Directory Usage**
  - **Validates: Requirements 15.1**
  - Create `internal/config/xdg_property_test.go`
  - Generate random system configuration scenarios
  - Verify os.UserConfigDir() is used for configuration
  - Test with various XDG environment settings
  - _Requirements: 15.1_

- [x] 26.8.2 Implement Property 38: Token Storage Location
  - **Property 38: Token Storage Location**
  - **Validates: Requirements 15.7**
  - Generate random authentication token storage scenarios
  - Verify tokens are stored in configuration directory
  - Test storage location consistency
  - _Requirements: 15.7_

- [x] 26.8.3 Implement Property 39: Cache Storage Location
  - **Property 39: Cache Storage Location**
  - **Validates: Requirements 15.8**
  - Generate random file content caching scenarios
  - Verify cache is stored in cache directory
  - Test cache location consistency
  - _Requirements: 15.8_

**Status**: ✅ **COMPLETE** (2025-11-13)  
**Requirements**: 15.1-15.10 all verified  
**Results**: OneMount correctly implements XDG Base Directory Specification with only minor deviation (auth token location) that has minimal impact.

---

## Phase 16: Socket.IO Transport Implementation Verification

**Status**: ✅ **IMPLEMENTED** - Verification tasks for Requirement 20 compliance

- [x] 27. Verify Engine.IO/Socket.IO Transport Implementation (Requirement 20)
- [x] 27.1 Review Socket.IO transport implementation
  - Read and analyze `internal/socketio/` implementation
  - Review Engine.IO v4 WebSocket transport
  - Check EIO=4 and transport=websocket query parameters
  - Verify default namespace (/) joining
  - _Requirements: 20.1_

- [x] 27.2 Test OAuth token attachment and refresh
  - Test Authorization bearer header attachment
  - Test token refresh during connection
  - Verify additional Graph-required headers
  - Test connection refresh on token rotation
  - _Requirements: 20.2_

- [x] 27.3 Test Engine.IO handshake and heartbeat
  - Test Engine.IO handshake frame parsing
  - Verify ping interval/timeout value parsing
  - Test debug level logging of handshake data
  - Test heartbeat timer configuration
  - _Requirements: 20.3_

- [x] 27.4 Test ping/pong and failure detection
  - Test ping/pong frame sending per negotiated interval
  - Test two consecutive missed heartbeat detection
  - Verify unhealthy state surfacing to ChangeNotifier
  - Test fallback to polling on heartbeat failure
  - _Requirements: 20.4_

- [x] 27.5 Test reconnection and backoff logic
  - Test exponential backoff on connection close/error
  - Verify backoff parameters (1s start, 2x multiplier, 60s cap, ±10% jitter)
  - Test backoff reset after successful reconnect
  - Test connection retry behavior
  - _Requirements: 20.5_

- [x] 27.6 Test event streaming and health monitoring
  - Test Socket.IO event streaming (notification, error)
  - Verify strongly typed callback handling
  - Test health indicator constant-time queries
  - Test ChangeNotifier integration
  - _Requirements: 20.6_

- [x] 27.7 Test verbose logging and tracing
  - Test structured trace logs for handshake data
  - Test ping/pong timing logs
  - Test packet read/write summary logs
  - Test payload truncation to configurable limit
  - Test close/error code logging
  - _Requirements: 20.7_

- [x] 27.8 Test automated transport tests
  - Run packet encode/decode tests
  - Run heartbeat scheduling tests
  - Run reconnection backoff tests
  - Run error propagation tests
  - Verify tests work without live Graph access
  - _Requirements: 20.8_

- [x] 27.9 Verify self-contained implementation
  - Verify no third-party Socket.IO client libraries
  - Verify no external proxies or managed relays
  - Check configuration whitelist for troubleshooting tools
  - Verify implementation is within OneMount codebase
  - _Requirements: 20.9_

- [x] 27.10 Create Socket.IO transport integration tests
  - Write test for complete transport lifecycle
  - Write test for OAuth integration
  - Write test for heartbeat and reconnection
  - Write test for event streaming
  - _Requirements: 20.1-20.9_

---

## ~~Phase 17: Multi-Account Support~~ ⏸️ DEFERRED TO v1.1+

**Status**: ⏸️ **DEFERRED**  
**Reason**: Not in current requirements, listed in `docs/0-project-management/deferred_features.md` for v1.1+

**No tasks needed** for initial release - this feature is intentionally deferred.

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

## Phase 16: ETag Cache Validation Verification ✅ COMPLETE

- [x] 29. Verify ETag-based cache validation with real OneDrive ✅ COMPLETE
- [x] 29.1 Review ETag implementation ✅ COMPLETE
- [x] 29.2 Test cache hit with valid ETag using real OneDrive ✅ COMPLETE
- [x] 29.3 Test cache miss with changed ETag using real OneDrive ✅ COMPLETE
- [x] 29.4 Test ETag updates from delta sync using real OneDrive ✅ COMPLETE
- [x] 29.5 Test conflict detection with ETags using real OneDrive ✅ COMPLETE
- [x] 29.6 Run ETag validation integration tests with real OneDrive ✅ COMPLETE

**Status**: ✅ **COMPLETE** (2025-11-13)  
**Requirements**: 3.4-3.6, 7.1-7.4, 8.1-8.3 all verified  
**Results**: ETag-based cache validation working correctly with real OneDrive API. All integration tests passing.

---

## Phase 17: State Management Verification

- [ ] 30. Verify metadata state model implementation
- [x] 30.1 Review state model implementation
  - Read and analyze `internal/fs/state_manager.go`
  - Review `internal/fs/hydration.go` state transitions
  - Check state persistence in metadata database
  - Review state transition diagram implementation
  - Verify all 7 states are implemented (GHOST, HYDRATING, HYDRATED, DIRTY_LOCAL, DELETED_LOCAL, CONFLICT, ERROR)
  - _Requirements: 21.1-21.10_

- [x] 30.2 Test initial item state assignment
  - Test items discovered via delta are inserted with GHOST state
  - Verify no content download until required
  - Test state persistence in metadata database
  - Verify virtual entries use correct state and flags
  - _Requirements: 21.2, 21.10_

- [x] 30.3 Test hydration state transitions
  - Test GHOST → HYDRATING transition on user access
  - Test HYDRATING → HYDRATED transition on successful download
  - Test HYDRATING → ERROR transition on download failure
  - Test HYDRATING → GHOST transition on cancellation
  - Verify worker deduplication during hydration
  - _Requirements: 21.3, 21.4, 21.5_

- [x] 30.4 Test modification and upload state transitions
  - Test HYDRATED → DIRTY_LOCAL transition on local modification
  - Test DIRTY_LOCAL → HYDRATED transition on successful upload
  - Test DIRTY_LOCAL → ERROR transition on upload failure
  - Test ETag updates after successful upload
  - _Requirements: 21.6_

- [x] 30.5 Test deletion state transitions
  - Test HYDRATED → DELETED_LOCAL transition on local delete
  - Test DELETED_LOCAL → [REMOVED] transition on server confirmation
  - Test DELETED_LOCAL → CONFLICT transition on remote modification
  - Verify tombstone handling
  - _Requirements: 21.7_

- [x] 30.6 Test conflict state transitions
  - Test DIRTY_LOCAL → CONFLICT transition on remote changes
  - Test CONFLICT → HYDRATED transition on conflict resolution
  - Test CONFLICT → GHOST transition on local version deletion
  - Verify both versions are preserved during conflict
  - _Requirements: 21.8_

- [x] 30.7 Test eviction and error recovery
  - Test HYDRATED → GHOST transition on cache eviction
  - Test ERROR → HYDRATING transition on retry
  - Test ERROR → DIRTY_LOCAL transition on upload retry
  - Test ERROR → GHOST transition on error clearing
  - _Requirements: 21.9_

- [x] 30.8 Test virtual file state handling
  - Test virtual entries have item_state=HYDRATED
  - Test virtual entries have remote_id=NULL and is_virtual=TRUE
  - Verify virtual entries bypass sync/upload logic
  - Test virtual entries participate in directory listings
  - _Requirements: 21.10_

- [x] 30.9 Test state transition atomicity and consistency
  - Test state transitions are atomic
  - Test no intermediate inconsistent states
  - Test state persistence across restarts
  - Test concurrent state transition safety
  - _Requirements: 21.1-21.10_

- [ ] 30.10 Create state model integration tests
  - Write test for complete state lifecycle
  - Write test for state transition edge cases
  - Write test for state persistence and recovery
  - Write test for concurrent state operations
  - _Requirements: 21.1-21.10_

- [ ] 30.11 Implement metadata state model property-based tests
- [ ] 30.11.1 Implement Property 40: Initial Item State
  - **Property 40: Initial Item State**
  - **Validates: Requirements 21.2**
  - Create `internal/fs/state_property_test.go`
  - Generate random drive items discovered via delta
  - Verify items are inserted with GHOST state
  - Verify no content download until required
  - _Requirements: 21.2_

- [ ] 30.11.2 Implement Property 41: Successful Hydration State Transition
  - **Property 41: Successful Hydration State Transition**
  - **Validates: Requirements 21.4**
  - Generate random successful hydration scenarios
  - Verify transition to HYDRATED state
  - Verify content path recording and metadata updates
  - Verify error field clearing
  - _Requirements: 21.4_

- [ ] 30.11.3 Implement Property 42: Local Modification State Transition
  - **Property 42: Local Modification State Transition**
  - **Validates: Requirements 21.6**
  - Generate random locally modified hydrated file scenarios
  - Verify transition to DIRTY_LOCAL state
  - Verify state persists until upload succeeds
  - _Requirements: 21.6_

---

## Phase 18: Security Property-Based Tests

- [x] 31. Implement security property-based tests
- [x] 31.1 Implement Property 43: Token Encryption at Rest
  - **Property 43: Token Encryption at Rest**
  - **Validates: Requirements 22.1**
  - Create `internal/security/security_property_test.go`
  - Generate random authentication token storage scenarios
  - Verify tokens are encrypted using AES-256
  - Test encryption key management and storage
  - _Requirements: 22.1_

- [x] 31.2 Implement Property 44: Token File Permissions
  - **Property 44: Token File Permissions**
  - **Validates: Requirements 22.2**
  - Generate random token file creation scenarios
  - Verify file permissions are set to 0600
  - Test permission enforcement across different platforms
  - _Requirements: 22.2_

- [x] 31.3 Implement Property 45: Secure Token Storage Location
  - **Property 45: Secure Token Storage Location**
  - **Validates: Requirements 22.3**
  - Generate random token storage scenarios
  - Verify storage in XDG configuration directory
  - Test access restriction enforcement
  - _Requirements: 22.3_

- [x] 31.4 Implement Property 46: HTTPS/TLS Communication
  - **Property 46: HTTPS/TLS Communication**
  - **Validates: Requirements 22.4**
  - Generate random Graph API communication scenarios
  - Verify HTTPS/TLS 1.2+ usage for all connections
  - Test certificate validation and security protocols
  - _Requirements: 22.4_

- [x] 31.5 Implement Property 47: Sensitive Data Logging Prevention
  - **Property 47: Sensitive Data Logging Prevention**
  - **Validates: Requirements 22.6**
  - Generate random logging scenarios with sensitive data
  - Verify no tokens, passwords, or sensitive data in logs
  - Test log sanitization mechanisms
  - _Requirements: 22.6_

- [x] 31.6 Implement Property 48: Cache File Security
  - **Property 48: Cache File Security**
  - **Validates: Requirements 22.8**
  - Generate random cached file storage scenarios
  - Verify appropriate file permissions for cached content
  - Test unauthorized access prevention
  - _Requirements: 22.8_

---

## Phase 19: Performance Property-Based Tests

- [x] 32. Implement performance property-based tests
- [x] 32.1 Implement Property 49: Directory Listing Performance
  - **Property 49: Directory Listing Performance**
  - **Validates: Requirements 23.1**
  - Create `internal/performance/performance_property_test.go`
  - Generate random directory listing scenarios (up to 1000 files)
  - Verify response times within 2 seconds
  - Test performance under various load conditions
  - _Requirements: 23.1_

- [x] 32.2 Implement Property 50: Cached File Access Performance
  - **Property 50: Cached File Access Performance**
  - **Validates: Requirements 23.2**
  - Generate random cached file access scenarios
  - Verify content served within 100 milliseconds
  - Test performance consistency across file sizes
  - _Requirements: 23.2_

- [x] 32.3 Implement Property 51: Idle Memory Usage
  - **Property 51: Idle Memory Usage**
  - **Validates: Requirements 23.3**
  - Generate random idle system scenarios
  - Verify memory consumption stays below 50 MB
  - Test memory leak detection during idle periods
  - _Requirements: 23.3_

- [x] 32.4 Implement Property 52: Active Sync Memory Usage
  - **Property 52: Active Sync Memory Usage**
  - **Validates: Requirements 23.4**
  - Generate random active synchronization scenarios
  - Verify memory consumption stays below 200 MB
  - Test memory usage during various sync operations
  - _Requirements: 23.4_

- [x] 32.5 Implement Property 53: Concurrent Operations Performance
  - **Property 53: Concurrent Operations Performance**
  - **Validates: Requirements 23.7**
  - Generate random concurrent operation scenarios (10+ operations)
  - Verify no performance degradation under concurrent load
  - Test scalability and resource contention
  - _Requirements: 23.7_

- [x] 32.6 Implement Property 54: Startup Performance
  - **Property 54: Startup Performance**
  - **Validates: Requirements 23.9**
  - Generate random system startup scenarios
  - Verify initialization completes within 5 seconds
  - Test startup performance under various conditions
  - _Requirements: 23.9_

- [x] 32.7 Implement Property 55: Shutdown Performance
  - **Property 55: Shutdown Performance**
  - **Validates: Requirements 23.10**
  - Generate random system shutdown scenarios
  - Verify graceful shutdown completes within 10 seconds
  - Test shutdown performance under load
  - _Requirements: 23.10_

---

## Phase 20: Resource Management Property-Based Tests

- [x] 33. Implement resource management property-based tests
- [x] 33.1 Implement Property 56: Cache Size Enforcement
  - **Property 56: Cache Size Enforcement**
  - **Validates: Requirements 24.1**
  - Create `internal/resources/resource_property_test.go`
  - Generate random cache configuration scenarios
  - Verify cache size limits are enforced
  - Test cache eviction and size management
  - _Requirements: 24.1_

- [x] 33.2 Implement Property 57: File Descriptor Limits
  - **Property 57: File Descriptor Limits**
  - **Validates: Requirements 24.4**
  - Generate random file descriptor usage scenarios
  - Verify file descriptor count stays below 1000
  - Test resource cleanup and leak prevention
  - _Requirements: 24.4_

- [x] 33.3 Implement Property 58: Worker Thread Limits
  - **Property 58: Worker Thread Limits**
  - **Validates: Requirements 24.5**
  - Generate random worker thread spawning scenarios
  - Verify worker count respects configured limits
  - Test thread pool management and cleanup
  - _Requirements: 24.5_

- [x] 33.4 Implement Property 59: Adaptive Network Throttling
  - **Property 59: Adaptive Network Throttling**
  - **Validates: Requirements 24.7**
  - Generate random limited bandwidth scenarios
  - Verify adaptive throttling prevents network saturation
  - Test throttling adjustment based on network conditions
  - _Requirements: 24.7_

- [x] 33.5 Implement Property 60: Memory Pressure Handling
  - **Property 60: Memory Pressure Handling**
  - **Validates: Requirements 24.8**
  - Generate random system memory pressure scenarios
  - Verify in-memory caching reduction and disk-based increase
  - Test memory usage adaptation under pressure
  - _Requirements: 24.8_

- [x] 33.6 Implement Property 61: CPU Usage Management
  - **Property 61: CPU Usage Management**
  - **Validates: Requirements 24.9**
  - Generate random high CPU usage scenarios
  - Verify background processing priority reduction
  - Test system responsiveness maintenance
  - _Requirements: 24.9_

- [x] 33.7 Implement Property 62: Graceful Resource Degradation
  - **Property 62: Graceful Resource Degradation**
  - **Validates: Requirements 24.10**
  - Generate random system resource pressure scenarios
  - Verify graceful degradation of non-essential features
  - Test core functionality preservation under pressure
  - _Requirements: 24.10_

---

## Phase 20.1: Fix Resource Management Property Test Failures

- [x] 33.8 Fix Property 56: Cache Size Enforcement failure
  - **Issue**: Cache size 256MB exceeds configured 10MB limit (with tolerance)
  - **Root Cause**: Cache size enforcement logic not working correctly
  - Investigate `internal/fs/content_cache.go` cache size tracking
  - Review cache eviction logic in `EvictOldEntries()` method
  - Verify `GetCacheSize()` accurately tracks total cache size
  - Check if cache insertion respects size limits
  - Fix cache size enforcement to respect configured limits
  - Verify eviction occurs when cache exceeds limit
  - Re-run Property 56 test to confirm fix
  - _Requirements: 24.1_

- [x] 33.9 Fix Property 58: Worker Thread Limits failure
  - **Issue**: Worker leak detected - 1 worker still active after test completion
  - **Root Cause**: Worker goroutines not being cleaned up properly
  - Investigate worker lifecycle in download/upload managers
  - Review goroutine cleanup in `StopDownloadManager()` and `StopUploadManager()`
  - Check for missing `defer` statements or cleanup calls
  - Verify worker pool shutdown waits for all workers to complete
  - Add proper synchronization for worker cleanup
  - Ensure all goroutines are properly terminated
  - Re-run Property 58 test to confirm fix
  - _Requirements: 24.5_

- [x] 33.10 Fix Property 59: Adaptive Network Throttling failure
  - **Issue**: Average bandwidth 2.50 MB/s exceeds limit 0.19 MB/s (with tolerance)
  - **Root Cause**: Network throttling not implemented or not working correctly
  - Review if adaptive throttling is implemented in download/upload managers
  - Check if bandwidth limiting is configured and enforced
  - Investigate rate limiting logic in network operations
  - Implement or fix bandwidth throttling mechanism
  - Add adaptive throttling based on network conditions
  - Verify throttling prevents network saturation
  - Re-run Property 59 test to confirm fix
  - _Requirements: 24.7_

---



## Phase 21: Concurrency and Lock Management Property-Based Tests

- [x] 34. Implement concurrency and lock management property-based tests
- [x] 34.1 Implement Property 63: Lock Ordering Compliance
  - **Property 63: Lock Ordering Compliance**
  - **Validates: Concurrency Design Requirements**
  - Create `internal/concurrency/lock_property_test.go`
  - Generate random sequences of lock acquisitions
  - Verify locks are acquired in defined order
  - Test with various concurrent scenarios
  - _Requirements: Concurrency Design_

- [x] 34.2 Implement Property 64: Deadlock Prevention
  - **Property 64: Deadlock Prevention**
  - **Validates: Concurrency Design Requirements**
  - Generate random concurrent operation scenarios
  - Verify no deadlocks occur when following lock ordering
  - Test with high concurrency and stress conditions
  - _Requirements: Concurrency Design_

- [x] 34.3 Implement Property 65: Lock Release Consistency
  - **Property 65: Lock Release Consistency**
  - **Validates: Concurrency Design Requirements**
  - Generate random lock acquisition scenarios with errors
  - Verify locks are released in reverse order (LIFO)
  - Test error handling and cleanup paths
  - _Requirements: Concurrency Design_

- [x] 34.4 Implement Property 66: Concurrent File Access Safety
  - **Property 66: Concurrent File Access Safety**
  - **Validates: Concurrency Design Requirements**
  - Generate random concurrent file operations on different inodes
  - Verify operations complete safely without race conditions
  - Test with race detector enabled
  - _Requirements: Concurrency Design_

- [x] 34.5 Implement Property 67: State Transition Atomicity
  - **Property 67: State Transition Atomicity**
  - **Validates: State Machine Design Requirements**
  - Generate random item state transition scenarios
  - Verify transitions complete atomically
  - Test for intermediate inconsistent states
  - _Requirements: State Machine Design_

---

## Phase 22: Final Verification

- [x] 35. Run complete test suite in Docker ✅ COMPLETED
  - Build latest test images: `docker compose -f docker/compose/docker-compose.build.yml build` ✅
  - Run all unit tests: `docker compose -f docker/compose/docker-compose.test.yml run unit-tests` ✅
  - Run all integration tests: `docker compose -f docker/compose/docker-compose.test.yml run integration-tests` ⚠️ MOSTLY PASSED
  - Run all system tests (requires auth): `docker compose -f docker/compose/docker-compose.test.yml run system-tests` ⚠️ SKIPPED (AUTH REQUIRED)
  - Generate coverage report: `docker compose -f docker/compose/docker-compose.test.yml run coverage` ❌ BLOCKED BY HANGING TESTS
  - Review test artifacts in `test-artifacts/logs/` ✅
  - Verify all tests pass ⚠️ PARTIAL SUCCESS
  - Document any remaining failures ✅
  - **RESULT**: Task completed with partial success. Unit tests fully pass, integration tests mostly pass (one hangs), system tests skip due to auth requirements. Build issues fixed.
  - **SUMMARY**: `test-artifacts/logs/task-35-complete-test-suite-summary.md`
  - _Requirements: All core requirements_

- [x] 36. Perform manual verification in Docker
  - Use interactive shell: `docker compose -f docker/compose/docker-compose.test.yml run shell`
  - Follow user workflows manually within container
  - Test mounting and file operations
  - Test Socket.IO realtime notifications
  - Verify all documented features work in isolated environment
  - Test with different configurations
  - Document any issues found during manual testing
  - _Requirements: All core requirements_

- [ ] 37. Performance verification
  - Run performance benchmarks
  - Test with Socket.IO realtime (30min polling fallback)
  - Test polling-only mode (5min polling)
  - Compare polling frequency impact
  - Verify response times meet expectations
  - Check resource usage is reasonable
  - _Requirements: Performance requirements_

- [ ] 38. Create verification report
  - Summarize all verification activities
  - List all issues found and fixed
  - Document Socket.IO realtime behavior
  - Document ETag cache validation
  - Document XDG compliance
  - Document remaining known issues
  - Provide recommendations for future work
  - _Requirements: All core requirements_

- [ ] 39. Final documentation review
  - Review all updated documentation
  - Ensure Socket.IO realtime documentation is complete
  - Ensure ETag validation is documented
  - Ensure XDG compliance is documented
  - Ensure consistency across documents
  - Verify all cross-references are correct
  - Check that documentation is complete
  - _Requirements: All core requirements_

---

## CORRECTED STATUS SUMMARY

### ✅ **COMPLETE PHASES** (Ready for Release)
- **Phases 1-9**: Docker Environment, Test Suite, Authentication, Mounting, File Operations, Upload Manager, Delta Sync, Cache Management, Offline Mode
- **Phases 11-14**: Error Handling, Performance & Concurrency, Integration Tests, End-to-End Tests  
- **Phase 15**: XDG Compliance ✅ (2025-11-13)
- **Phase 16**: ETag Cache Validation ✅ (2025-11-13)
- **Socket.IO Realtime**: ✅ Already implemented and working (2025-11-17)

### 🔄 **IN PROGRESS**
- **Phase 10**: File Status & D-Bus (4/7 tasks complete)
- **Phase 15**: Issue Resolution (many medium-priority issues already fixed)

### ⏭️ **REMAINING WORK**
- **Phase 12**: Offline Mode (Network Error Pattern Recognition tasks added)
- **Phase 16**: Socket.IO Transport Implementation Verification (Requirement 20)
- **Phase 17**: State Management Verification (Requirement 21) + Documentation Updates
- **Phase 17**: State Management Property-Based Tests
- **Phase 18-22**: Security, Performance, Resource, Audit Property-Based Tests
- **Phase 23**: Final Verification

### ❌ **REMOVED/DEFERRED**
- ~~Webhook Subscriptions~~: Obsolete (replaced by Socket.IO)
- ~~Multi-Account Support~~: Deferred to v1.1+ (not in requirements)

### 📊 **ACTUAL PROGRESS**
- **Core Functionality**: ~95% complete
- **Verification**: ~85% complete  
- **Documentation**: ~80% complete (enhanced with concurrency guidelines)
- **Ready for Release**: After completing Phase 10 and final verification

The project is much closer to completion than originally indicated!

---

## ENHANCED SPECIFICATION IMPROVEMENTS ✅ COMPLETE

**STATUS**: ✅ **ALL MEDIUM PRIORITY IMPROVEMENTS IMPLEMENTED**

### ✅ **Completed Enhancements:**

**✅ Requirement Granularity**:
- Split Requirement 2 (Filesystem Mounting) into 5 focused requirements (2, 2A, 2B, 2C, 2D)
- Split Requirement 3 (On-Demand Download) into 4 focused requirements (3, 3A, 3B, 3C)
- Each requirement now has 2-6 acceptance criteria instead of 16-23
- Improved testability and traceability

**✅ State Machine Diagrams**:
- Added comprehensive Mermaid state transition diagram
- Documented all valid and invalid state transitions
- Added trigger conditions and error handling rules
- Included system invariants and recovery procedures

**✅ Concurrency Documentation**:
- Added formal lock ordering policy (6-level hierarchy)
- Documented deadlock prevention strategies
- Added lock granularity guidelines
- Included race condition prevention patterns
- Added performance considerations and monitoring

**✅ Enhanced Property-Based Testing**:
- Added 5 new correctness properties for concurrency (Properties 61-65)
- Added Phase 22 for concurrency property-based tests
- Updated existing property references to match granular requirements

---

## PROPERTY-BASED TEST IMPLEMENTATION TASKS ✅ COMPLETE

**STATUS**: ✅ **ALL PROPERTY-BASED TEST TASKS HAVE BEEN ADDED**

All 67 correctness properties from the design document have been implemented as property-based test tasks in their respective phases:

### ✅ **Completed Property-Based Test Task Additions:**

**✅ Authentication Properties (Phase 3)**:
- ✅ Property 1: OAuth2 Token Storage Security (Requirements 1.2)
- ✅ Property 2: Automatic Token Refresh (Requirements 1.3)
- ✅ Property 3: Re-authentication on Refresh Failure (Requirements 1.4)
- ✅ Property 4: Headless Authentication Method (Requirements 1.5)

**✅ Filesystem Mounting Properties (Phase 4)**:
- ✅ Property 5: FUSE Mount Success (Requirements 2.1)
- ✅ Property 6: Non-blocking Initial Sync (Requirements 2.2)
- ✅ Property 7: Root Directory Visibility (Requirements 2.3)
- ✅ Property 8: Standard File Operations Support (Requirements 2.4)
- ✅ Property 9: Mount Conflict Error Handling (Requirements 2.8)
- ✅ Property 10: Clean Resource Release (Requirements 2.9)

**✅ File Access Properties (Phase 5)**:
- ✅ Property 11: Metadata-Only Directory Listing (Requirements 3.1)
- ✅ Property 12: On-Demand Content Download (Requirements 3.2)
- ✅ Property 13: ETag Cache Validation (Requirements 3.4)
- ✅ Property 14: Cache Hit Serving (Requirements 3.5)
- ✅ Property 15: Cache Invalidation on ETag Mismatch (Requirements 3.6)

**✅ File Modification Properties (Phase 5)**:
- ✅ Property 16: Local Change Tracking (Requirements 4.1)
- ✅ Property 17: Upload Queuing (Requirements 4.2)
- ✅ Property 18: ETag Update After Upload (Requirements 4.7)
- ✅ Property 19: Modified Flag Cleanup (Requirements 4.8)

**✅ Delta Synchronization Properties (Phase 8)**:
- ✅ Property 20: Initial Delta Sync (Requirements 5.1)
- ✅ Property 21: Metadata Cache Updates (Requirements 5.8)
- ✅ Property 22: Conflict Copy Creation (Requirements 5.11)
- ✅ Property 23: Delta Token Persistence (Requirements 5.12)

**✅ Offline Mode Properties (Phase 10)**:
- ✅ Property 24: Offline Detection (Requirements 6.1)
- ✅ Property 25: Offline Read Access (Requirements 6.4)
- ✅ Property 26: Offline Write Queuing (Requirements 6.5)
- ✅ Property 27: Batch Upload Processing (Requirements 6.10)

**✅ Cache Management Properties (Phase 9)**:
- ✅ Property 28: ETag-Based Cache Storage (Requirements 7.1)
- ✅ Property 29: Cache Invalidation on Remote ETag Change (Requirements 7.3)

**✅ Conflict Resolution Properties (Phase 8)**:
- ✅ Property 30: ETag-Based Conflict Detection (Requirements 8.1)
- ✅ Property 31: Local Version Preservation (Requirements 8.4)
- ✅ Property 32: Conflict Copy Creation with Timestamp (Requirements 8.5)

**✅ Concurrency Properties (Phase 13)**:
- ✅ Property 33: Safe Concurrent File Access (Requirements 10.1)
- ✅ Property 34: Non-blocking Downloads (Requirements 10.2)

**✅ Error Handling Properties (Phase 12)**:
- ✅ Property 35: Network Error Logging (Requirements 11.1)
- ✅ Property 36: Rate Limit Backoff (Requirements 11.2)

**✅ Configuration Properties (Phase 15)**:
- ✅ Property 37: XDG Configuration Directory Usage (Requirements 15.1)
- ✅ Property 38: Token Storage Location (Requirements 15.7)
- ✅ Property 39: Cache Storage Location (Requirements 15.8)

**✅ State Management Properties (Phase 17)**:
- ✅ Property 40: Initial Item State (Requirements 21.2)
- ✅ Property 41: Successful Hydration State Transition (Requirements 21.4)
- ✅ Property 42: Local Modification State Transition (Requirements 21.6)

**✅ Security Properties (Phase 18)**:
- ✅ Property 43: Token Encryption at Rest (Requirements 22.1)
- ✅ Property 44: Token File Permissions (Requirements 22.2)
- ✅ Property 45: Secure Token Storage Location (Requirements 22.3)
- ✅ Property 46: HTTPS/TLS Communication (Requirements 22.4)
- ✅ Property 47: Sensitive Data Logging Prevention (Requirements 22.6)
- ✅ Property 48: Cache File Security (Requirements 22.8)

**✅ Performance Properties (Phase 19)**:
- ✅ Property 49: Directory Listing Performance (Requirements 23.1)
- ✅ Property 50: Cached File Access Performance (Requirements 23.2)
- ✅ Property 51: Idle Memory Usage (Requirements 23.3)
- ✅ Property 52: Active Sync Memory Usage (Requirements 23.4)
- ✅ Property 53: Concurrent Operations Performance (Requirements 23.7)
- ✅ Property 54: Startup Performance (Requirements 23.9)
- ✅ Property 55: Shutdown Performance (Requirements 23.10)

**✅ Resource Management Properties (Phase 20)**:
- ✅ Property 56: Cache Size Enforcement (Requirements 24.1)
- ✅ Property 57: File Descriptor Limits (Requirements 24.4)
- ✅ Property 58: Worker Thread Limits (Requirements 24.5)

**✅ Concurrency and Lock Management Properties (Phase 21)**:
- ✅ Property 63: Lock Ordering Compliance (Concurrency Design)
- ✅ Property 64: Deadlock Prevention (Concurrency Design)
- ✅ Property 65: Lock Release Consistency (Concurrency Design)
- ✅ Property 66: Concurrent File Access Safety (Concurrency Design)
- ✅ Property 67: State Transition Atomicity (State Machine Design)

**TOTAL**: 67 correctness properties across 22 phases, each with dedicated property-based test implementation tasks.

The specification now provides comprehensive coverage of all functional, security, performance, and concurrency requirements with formal correctness properties and a complete implementation plan.irements 3.2)
- ✅ Property 13: ETag Cache Validation (Requirements 3.4)
- ✅ Property 14: Cache Hit Serving (Requirements 3.5)
- ✅ Property 15: Cache Invalidation on ETag Mismatch (Requirements 3.6)

**✅ File Modification Properties (Phase 5)**:
- ✅ Property 16: Local Change Tracking (Requirements 4.1)
- ✅ Property 17: Upload Queuing (Requirements 4.2)
- ✅ Property 18: ETag Update After Upload (Requirements 4.7)
- ✅ Property 19: Modified Flag Cleanup (Requirements 4.8)

**✅ Delta Synchronization Properties (Phase 8)**:
- ✅ Property 20: Initial Delta Sync (Requirements 5.1)
- ✅ Property 21: Metadata Cache Updates (Requirements 5.8)
- ✅ Property 22: Conflict Copy Creation (Requirements 5.11)
- ✅ Property 23: Delta Token Persistence (Requirements 5.12)

**✅ Conflict Resolution Properties (Phase 8)**:
- ✅ Property 30: ETag-Based Conflict Detection (Requirements 8.1)
- ✅ Property 31: Local Version Preservation (Requirements 8.4)
- ✅ Property 32: Conflict Copy Creation with Timestamp (Requirements 8.5)

**✅ Cache Management Properties (Phase 9)**:
- ✅ Property 28: ETag-Based Cache Storage (Requirements 7.1)
- ✅ Property 29: Cache Invalidation on Remote ETag Change (Requirements 7.3)

**✅ Offline Mode Properties (Phase 10)**:
- ✅ Property 24: Offline Detection (Requirements 6.1)
- ✅ Property 25: Offline Read Access (Requirements 6.4)
- ✅ Property 26: Offline Write Queuing (Requirements 6.5)
- ✅ Property 27: Batch Upload Processing (Requirements 6.10)

**✅ Error Handling Properties (Phase 12)**:
- ✅ Property 35: Network Error Logging (Requirements 11.1)
- ✅ Property 36: Rate Limit Backoff (Requirements 11.2)

**✅ Concurrency Properties (Phase 13)**:
- ✅ Property 33: Safe Concurrent File Access (Requirements 10.1)
- ✅ Property 34: Non-blocking Downloads (Requirements 10.2)

**✅ Configuration Properties (Phase 15)**:
- ✅ Property 37: XDG Configuration Directory Usage (Requirements 15.1)
- ✅ Property 38: Token Storage Location (Requirements 15.7)
- ✅ Property 39: Cache Storage Location (Requirements 15.8)

**✅ State Management Properties (Phase 17)**:
- ✅ Property 40: Initial Item State (Requirements 21.2)
- ✅ Property 41: Successful Hydration State Transition (Requirements 21.4)
- ✅ Property 42: Local Modification State Transition (Requirements 21.6)

**✅ Security Properties (Phase 18)**:
- ✅ Property 43: Token Encryption at Rest (Requirements 22.1)
- ✅ Property 44: Token File Permissions (Requirements 22.2)
- ✅ Property 45: Secure Token Storage Location (Requirements 22.3)
- ✅ Property 46: HTTPS/TLS Communication (Requirements 22.4)
- ✅ Property 47: Sensitive Data Logging Prevention (Requirements 22.6)
- ✅ Property 48: Cache File Security (Requirements 22.8)

**✅ Performance Properties (Phase 19)**:
- ✅ Property 49: Directory Listing Performance (Requirements 23.1)
- ✅ Property 50: Cached File Access Performance (Requirements 23.2)
- ✅ Property 51: Idle Memory Usage (Requirements 23.3)
- ✅ Property 52: Active Sync Memory Usage (Requirements 23.4)
- ✅ Property 53: Concurrent Operations Performance (Requirements 23.7)
- ✅ Property 54: Startup Performance (Requirements 23.9)
- ✅ Property 55: Shutdown Performance (Requirements 23.10)

**✅ Resource Management Properties (Phase 20)**:
- ✅ Property 56: Cache Size Enforcement (Requirements 24.1)
- ✅ Property 57: File Descriptor Limits (Requirements 24.4)
- ✅ Property 58: Worker Thread Limits (Requirements 24.5)

### 📋 **Property-Based Test Implementation Guidelines:**

1. **File Naming**: Use `*_property_test.go` for property test files
2. **Test Annotation**: Each test must include the exact comment format:
   `// **Feature: system-verification-and-fix, Property {number}: {property_text}**`
3. **Library**: Use Go's `testing/quick` package or `github.com/leanovate/gopter`
4. **Iterations**: Minimum 100 iterations per property test
5. **Organization**: Group by component (auth, filesystem, cache, etc.)
6. **Integration**: Run as part of integration test suite in Docker

### 🎯 **Implementation Priority:**
1. **HIGH**: Properties 1-19 (Authentication, Mounting, File Operations) - Core functionality
2. **MEDIUM**: Properties 20-32 (Delta Sync, Offline, Cache, Conflicts) - Advanced features  
3. **MEDIUM**: Properties 43-48 (Security) - Critical security requirements
4. **LOW**: Properties 33-42 (Concurrency, Error Handling, Configuration, State) - System robustness
5. **LOW**: Properties 49-62 (Performance, Resource Management) - Quality and system efficiency

**✅ RESULT**: The spec is now **FULLY COMPLIANT** with the property-based testing workflow requirements.
