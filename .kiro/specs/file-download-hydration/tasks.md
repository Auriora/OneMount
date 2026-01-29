# Tasks: File Download Hydration

## Source
Moved from `.kiro/specs/archive/system-verification-and-fix/tasks.md`. Sections below are verbatim.

## Phase 5: File Operations Verification (Read/Metadata)

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

