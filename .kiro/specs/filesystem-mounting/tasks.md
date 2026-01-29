# Tasks: Filesystem Mounting

## Source
Moved from `.kiro/specs/archive/system-verification-and-fix/tasks.md`. Sections below are verbatim.

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
