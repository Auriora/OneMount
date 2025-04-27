# OneDriver Test Case Mapping Matrix

This document maps the standardized test cases from `test_cases.md` to the actual Go test implementations in the codebase.

## Mapping Matrix

| Test Case ID | Test Case Title | Go Test Implementation | File | Notes |
|--------------|----------------|------------------------|------|-------|
| TC-01 | Directory Tree Synchronization | TestSyncDirectoryTree | fs/sync_test.go | Directly tests the `SyncDirectoryTree` function to verify directory tree synchronization. |
| TC-02 | File Upload Synchronization | TestFileOperations | fs/fs_test.go | Tests basic file creation and writing, which triggers uploads. |
| TC-03 | Repeated File Upload | TestRepeatedUploads | fs/upload_manager_test.go | Directly tests uploading the same file multiple times with different content. |
| TC-04 | Large File Upload | TestUploadDiskSerialization | fs/upload_manager_test.go | Tests uploading large files using upload sessions. |
| TC-05 | File Download Synchronization | TestBasicFileSystemOperations | fs/fs_test.go | Tests reading files, which triggers downloads if not in cache. |
| TC-06 | Offline File Access | TestOfflineFileAccess | fs/offline/offline_test.go | Directly tests accessing files when offline. |
| TC-07 | Offline File Creation | TestOfflineFileSystemOperations | fs/offline/offline_test.go | Contains test cases for file creation in offline mode. |
| TC-08 | Offline File Modification | TestOfflineFileSystemOperations | fs/offline/offline_test.go | Contains test cases for file modification in offline mode. |
| TC-09 | Offline File Deletion | TestOfflineFileSystemOperations | fs/offline/offline_test.go | Contains test cases for file deletion in offline mode. |
| TC-10 | Reconnection Synchronization | TestOfflineSynchronization | fs/offline/offline_test.go | Tests synchronization when reconnecting after being offline. |
| TC-11 | Unauthenticated Request Handling | TestRequestUnauthenticated | fs/graph/graph_test.go | Tests handling of requests with invalid authentication. |
| TC-12 | Authentication Token Refresh | TestAuthRefresh | fs/graph/oauth2_test.go | Tests refreshing expired authentication tokens. |
| TC-13 | Invalid Authentication Code Format | TestAuthCodeFormat | fs/graph/oauth2_test.go | Tests parsing of various authentication code formats, including invalid ones. |
| TC-14 | Authentication Persistence | TestAuthFromfile | fs/graph/oauth2_test.go | Tests loading authentication tokens from file. |
| TC-15 | Authentication Failure with Network Available | TestAuthFailureWithNetworkAvailable | fs/graph/oauth2_test.go | Tests the behavior when authentication fails but network is available. |

## Additional Tests Mapped to New Test Cases

The following tests in the codebase have been mapped to new standardized test cases:

1. **File System Operations**
   - TestReaddir (fs/fs_test.go) → TC-16 (Directory Reading)
   - TestLs (fs/fs_test.go) → TC-17 (Directory Listing with Shell Commands)
   - TestTouchOperations (fs/fs_test.go) → TC-18 (File Creation and Modification Time)
   - TestFilePermissions (fs/fs_test.go) → TC-19 (File Permissions)
   - TestDirectoryOperations (fs/fs_test.go) → TC-20 (Directory Creation and Removal)
   - TestDirectoryRemoval (fs/fs_test.go) → TC-20 (Directory Creation and Removal)
   - TestWriteOffset (fs/fs_test.go) → TC-21 (File Writing with Offset)
   - TestFileMovementOperations (fs/fs_test.go) → TC-22 (File Movement Operations)
   - TestPositionalFileOperations (fs/fs_test.go) → TC-23 (Positional File Operations)
   - TestCaseSensitivityHandling (fs/fs_test.go) → TC-24 (Case Sensitivity Handling)
   - TestFilenameCase (fs/fs_test.go) → TC-24 (Case Sensitivity Handling)
   - TestShellFileOperations (fs/fs_test.go) → TC-17 (Directory Listing with Shell Commands)
   - TestFileInfo (fs/fs_test.go) → TC-19 (File Permissions)
   - TestNoQuestionMarks (fs/fs_test.go) → TC-25 (Special Characters in Filenames)
   - TestGIOTrash (fs/fs_test.go) → TC-26 (Trash Functionality)
   - TestListChildrenPaging (fs/fs_test.go) → TC-27 (Directory Paging)
   - TestLibreOfficeSavePattern (fs/fs_test.go) → TC-28 (Application-Specific Save Patterns)
   - TestDisallowedFilenames (fs/fs_test.go) → TC-25 (Special Characters in Filenames)

2. **Authentication and Configuration**
   - TestAuthConfigMerge (fs/graph/oauth2_test.go) → TC-29 (Authentication Configuration Merging)
   - TestResourcePath (fs/graph/graph_test.go) → TC-30 (Resource Path Handling)

3. **Offline Functionality**
   - TestOfflineChangesCached (fs/offline/offline_test.go) → TC-31 (Offline Changes Caching)

## Coverage Analysis

The mapping shows that all standardized test cases now have corresponding implementations in the codebase. The previously identified gaps have been addressed:

1. **TC-01 (Directory Tree Synchronization)**: Now tested by `TestSyncDirectoryTree` in `fs/sync_test.go`.
2. **TC-15 (Authentication Failure with Network Available)**: Now tested by `TestAuthFailureWithNetworkAvailable` in `fs/graph/oauth2_test.go`.

All test cases are now covered by specific tests in the codebase.
