# Tasks: File Upload Modification

## Source
Moved from `.kiro/specs/archive/system-verification-and-fix/tasks.md`. Sections below are verbatim.

## Phase 5: File Operations Verification (Write/Upload)

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
