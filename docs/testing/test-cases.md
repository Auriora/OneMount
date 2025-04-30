# OneMount Test Cases

This document contains standardized test cases for the OneMount project, covering successful sync, network failure, invalid credentials scenarios, file system operations, authentication and configuration, and offline functionality.

## Successful Sync Scenarios

| Field           | Description                                                                                                                                                          |
|-----------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Test Case ID    | TC-01                                                                                                                                                                |
| Title           | Directory Tree Synchronization                                                                                                                                       |
| Description     | Verify that the filesystem can successfully synchronize the directory tree from the root                                                                             |
| Preconditions   | 1. User is authenticated with valid credentials<br>2. Network connection is available                                                                                |
| Steps           | 1. Initialize the filesystem<br>2. Call SyncDirectoryTree method<br>3. Verify all directories are cached                                                             |
| Expected Result | All directory metadata is successfully cached without errors                                                                                                         |
| Actual Result   | [To be filled during test execution]                                                                                                                                 |
| Status          | [Pass/Fail]                                                                                                                                                          |
| Implementation  | **Test**: TestSyncDirectoryTree<br>**File**: fs/sync_test.go<br>**Notes**: Directly tests the `SyncDirectoryTree` function to verify directory tree synchronization. |

| Field           | Description                                                                                                                                                                      |
|-----------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Test Case ID    | TC-02                                                                                                                                                                            |
| Title           | File Upload Synchronization                                                                                                                                                      |
| Description     | Verify that a file can be successfully uploaded to OneDrive                                                                                                                      |
| Preconditions   | 1. User is authenticated with valid credentials<br>2. Network connection is available                                                                                            |
| Steps           | 1. Create a new file in the local filesystem<br>2. Write content to the file<br>3. Wait for the upload to complete<br>4. Verify the file exists on OneDrive with correct content |
| Expected Result | File is successfully uploaded to OneDrive with the correct content                                                                                                               |
| Actual Result   | [To be filled during test execution]                                                                                                                                             |
| Status          | [Pass/Fail]                                                                                                                                                                      |
| Implementation  | **Test**: TestFileOperations<br>**File**: fs/fs_test.go<br>**Notes**: Tests basic file creation and writing, which triggers uploads.                                             |

| Field           | Description                                                                                                                                                                   |
|-----------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Test Case ID    | TC-03                                                                                                                                                                         |
| Title           | Repeated File Upload                                                                                                                                                          |
| Description     | Verify that the same file can be uploaded multiple times with different content                                                                                               |
| Preconditions   | 1. User is authenticated with valid credentials<br>2. Network connection is available                                                                                         |
| Steps           | 1. Create a file with initial content<br>2. Wait for upload to complete<br>3. Modify the file content<br>4. Wait for upload to complete<br>5. Repeat steps 3-4 multiple times |
| Expected Result | Each version of the file is successfully uploaded with the correct content                                                                                                    |
| Actual Result   | [To be filled during test execution]                                                                                                                                          |
| Status          | [Pass/Fail]                                                                                                                                                                   |
| Implementation  | **Test**: TestRepeatedUploads<br>**File**: fs/upload_manager_test.go<br>**Notes**: Directly tests uploading the same file multiple times with different content.              |

| Field           | Description                                                                                                                                                       |
|-----------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Test Case ID    | TC-04                                                                                                                                                             |
| Title           | Large File Upload                                                                                                                                                 |
| Description     | Verify that large files can be uploaded correctly using upload sessions                                                                                           |
| Preconditions   | 1. User is authenticated with valid credentials<br>2. Network connection is available                                                                             |
| Steps           | 1. Create a large file (>4MB)<br>2. Write content to the file<br>3. Wait for the upload to complete<br>4. Verify the file exists on OneDrive with correct content |
| Expected Result | Large file is successfully uploaded to OneDrive with the correct content                                                                                          |
| Actual Result   | [To be filled during test execution]                                                                                                                              |
| Status          | [Pass/Fail]                                                                                                                                                       |
| Implementation  | **Test**: TestUploadDiskSerialization<br>**File**: fs/upload_manager_test.go<br>**Notes**: Tests uploading large files using upload sessions.                     |

| Field           | Description                                                                                                                                     |
|-----------------|-------------------------------------------------------------------------------------------------------------------------------------------------|
| Test Case ID    | TC-05                                                                                                                                           |
| Title           | File Download Synchronization                                                                                                                   |
| Description     | Verify that a file can be successfully downloaded from OneDrive                                                                                 |
| Preconditions   | 1. User is authenticated with valid credentials<br>2. Network connection is available<br>3. File exists on OneDrive                             |
| Steps           | 1. Access a file that exists on OneDrive but not in local cache<br>2. Read the file content<br>3. Verify the content matches what's on OneDrive |
| Expected Result | File is successfully downloaded from OneDrive with the correct content                                                                          |
| Actual Result   | [To be filled during test execution]                                                                                                            |
| Status          | [Pass/Fail]                                                                                                                                     |
| Implementation  | **Test**: TestBasicFileSystemOperations<br>**File**: fs/fs_test.go<br>**Notes**: Tests reading files, which triggers downloads if not in cache. |

## Network Failure Scenarios

| Field           | Description                                                                                                                        |
|-----------------|------------------------------------------------------------------------------------------------------------------------------------|
| Test Case ID    | TC-06                                                                                                                              |
| Title           | Offline File Access                                                                                                                |
| Description     | Verify that files can be accessed when offline                                                                                     |
| Preconditions   | 1. User is authenticated<br>2. Files have been previously synchronized<br>3. Network connection is unavailable                     |
| Steps           | 1. Disconnect from the network<br>2. Access a previously synchronized file<br>3. Read the file content                             |
| Expected Result | File content is successfully read from the local cache                                                                             |
| Actual Result   | [To be filled during test execution]                                                                                               |
| Status          | [Pass/Fail]                                                                                                                        |
| Implementation  | **Test**: TestOfflineFileAccess<br>**File**: fs/offline/offline_test.go<br>**Notes**: Directly tests accessing files when offline. |

| Field           | Description                                                                                                                                            |
|-----------------|--------------------------------------------------------------------------------------------------------------------------------------------------------|
| Test Case ID    | TC-07                                                                                                                                                  |
| Title           | Offline File Creation                                                                                                                                  |
| Description     | Verify that files can be created when offline                                                                                                          |
| Preconditions   | 1. User is authenticated<br>2. Network connection is unavailable                                                                                       |
| Steps           | 1. Disconnect from the network<br>2. Create a new file<br>3. Write content to the file<br>4. Verify the file exists locally                            |
| Expected Result | File is successfully created locally and marked for upload when online                                                                                 |
| Actual Result   | [To be filled during test execution]                                                                                                                   |
| Status          | [Pass/Fail]                                                                                                                                            |
| Implementation  | **Test**: TestOfflineFileSystemOperations<br>**File**: fs/offline/offline_test.go<br>**Notes**: Contains test cases for file creation in offline mode. |

| Field           | Description                                                                                                                                                |
|-----------------|------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Test Case ID    | TC-08                                                                                                                                                      |
| Title           | Offline File Modification                                                                                                                                  |
| Description     | Verify that files can be modified when offline                                                                                                             |
| Preconditions   | 1. User is authenticated<br>2. File has been previously synchronized<br>3. Network connection is unavailable                                               |
| Steps           | 1. Disconnect from the network<br>2. Modify a previously synchronized file<br>3. Verify the file content is updated locally                                |
| Expected Result | File is successfully modified locally and marked for upload when online                                                                                    |
| Actual Result   | [To be filled during test execution]                                                                                                                       |
| Status          | [Pass/Fail]                                                                                                                                                |
| Implementation  | **Test**: TestOfflineFileSystemOperations<br>**File**: fs/offline/offline_test.go<br>**Notes**: Contains test cases for file modification in offline mode. |

| Field           | Description                                                                                                                                            |
|-----------------|--------------------------------------------------------------------------------------------------------------------------------------------------------|
| Test Case ID    | TC-09                                                                                                                                                  |
| Title           | Offline File Deletion                                                                                                                                  |
| Description     | Verify that files can be deleted when offline                                                                                                          |
| Preconditions   | 1. User is authenticated<br>2. File has been previously synchronized<br>3. Network connection is unavailable                                           |
| Steps           | 1. Disconnect from the network<br>2. Delete a previously synchronized file<br>3. Verify the file is no longer accessible locally                       |
| Expected Result | File is successfully deleted locally and marked for deletion on OneDrive when online                                                                   |
| Actual Result   | [To be filled during test execution]                                                                                                                   |
| Status          | [Pass/Fail]                                                                                                                                            |
| Implementation  | **Test**: TestOfflineFileSystemOperations<br>**File**: fs/offline/offline_test.go<br>**Notes**: Contains test cases for file deletion in offline mode. |

| Field           | Description                                                                                                                                                     |
|-----------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Test Case ID    | TC-10                                                                                                                                                           |
| Title           | Reconnection Synchronization                                                                                                                                    |
| Description     | Verify that changes made offline are synchronized when reconnecting                                                                                             |
| Preconditions   | 1. User is authenticated<br>2. Changes have been made while offline<br>3. Network connection becomes available                                                  |
| Steps           | 1. Make changes to files while offline<br>2. Reconnect to the network<br>3. Wait for synchronization to complete<br>4. Verify changes are reflected on OneDrive |
| Expected Result | All offline changes are successfully synchronized to OneDrive                                                                                                   |
| Actual Result   | [To be filled during test execution]                                                                                                                            |
| Status          | [Pass/Fail]                                                                                                                                                     |
| Implementation  | **Test**: TestOfflineSynchronization<br>**File**: fs/offline/offline_test.go<br>**Notes**: Tests synchronization when reconnecting after being offline.         |

## Invalid Credentials Scenarios

| Field           | Description                                                                                                                                    |
|-----------------|------------------------------------------------------------------------------------------------------------------------------------------------|
| Test Case ID    | TC-11                                                                                                                                          |
| Title           | Unauthenticated Request Handling                                                                                                               |
| Description     | Verify that requests with invalid authentication are properly handled                                                                          |
| Preconditions   | 1. User has invalid or expired credentials                                                                                                     |
| Steps           | 1. Attempt to access OneDrive resources with invalid credentials<br>2. Verify error handling                                                   |
| Expected Result | System properly handles the authentication error and returns appropriate error message                                                         |
| Actual Result   | [To be filled during test execution]                                                                                                           |
| Status          | [Pass/Fail]                                                                                                                                    |
| Implementation  | **Test**: TestRequestUnauthenticated<br>**File**: fs/graph/graph_test.go<br>**Notes**: Tests handling of requests with invalid authentication. |

| Field           | Description                                                                                                                    |
|-----------------|--------------------------------------------------------------------------------------------------------------------------------|
| Test Case ID    | TC-12                                                                                                                          |
| Title           | Authentication Token Refresh                                                                                                   |
| Description     | Verify that expired authentication tokens are automatically refreshed                                                          |
| Preconditions   | 1. User has valid but expired access token<br>2. User has valid refresh token                                                  |
| Steps           | 1. Set the access token to be expired<br>2. Attempt to access OneDrive resources<br>3. Verify token is refreshed automatically |
| Expected Result | Access token is refreshed and the request succeeds                                                                             |
| Actual Result   | [To be filled during test execution]                                                                                           |
| Status          | [Pass/Fail]                                                                                                                    |
| Implementation  | **Test**: TestAuthRefresh<br>**File**: fs/graph/oauth2_test.go<br>**Notes**: Tests refreshing expired authentication tokens.   |

| Field           | Description                                                                                                                                                   |
|-----------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Test Case ID    | TC-13                                                                                                                                                         |
| Title           | Invalid Authentication Code Format                                                                                                                            |
| Description     | Verify that invalid authentication code formats are properly handled                                                                                          |
| Preconditions   | 1. System is attempting to parse an authentication code                                                                                                       |
| Steps           | 1. Provide an invalid authentication code format<br>2. Attempt to parse the code<br>3. Verify error handling                                                  |
| Expected Result | System properly handles the invalid format and returns appropriate error                                                                                      |
| Actual Result   | [To be filled during test execution]                                                                                                                          |
| Status          | [Pass/Fail]                                                                                                                                                   |
| Implementation  | **Test**: TestAuthCodeFormat<br>**File**: fs/graph/oauth2_test.go<br>**Notes**: Tests parsing of various authentication code formats, including invalid ones. |

| Field           | Description                                                                                                                  |
|-----------------|------------------------------------------------------------------------------------------------------------------------------|
| Test Case ID    | TC-14                                                                                                                        |
| Title           | Authentication Persistence                                                                                                   |
| Description     | Verify that authentication tokens are properly persisted and loaded                                                          |
| Preconditions   | 1. User has previously authenticated<br>2. Authentication tokens are saved to file                                           |
| Steps           | 1. Restart the application<br>2. Verify authentication tokens are loaded from file<br>3. Verify access to OneDrive resources |
| Expected Result | Authentication tokens are successfully loaded and used for accessing resources                                               |
| Actual Result   | [To be filled during test execution]                                                                                         |
| Status          | [Pass/Fail]                                                                                                                  |
| Implementation  | **Test**: TestAuthFromfile<br>**File**: fs/graph/oauth2_test.go<br>**Notes**: Tests loading authentication tokens from file. |

| Field           | Description                                                                                                                                                             |
|-----------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Test Case ID    | TC-15                                                                                                                                                                   |
| Title           | Authentication Failure with Network Available                                                                                                                           |
| Description     | Verify behavior when authentication fails but network is available                                                                                                      |
| Preconditions   | 1. User has invalid credentials<br>2. Network connection is available                                                                                                   |
| Steps           | 1. Attempt to authenticate with invalid credentials<br>2. Verify error handling<br>3. Verify system state after failure                                                 |
| Expected Result | System properly handles the authentication failure and provides appropriate user feedback                                                                               |
| Actual Result   | [To be filled during test execution]                                                                                                                                    |
| Status          | [Pass/Fail]                                                                                                                                                             |
| Implementation  | **Test**: TestAuthFailureWithNetworkAvailable<br>**File**: fs/graph/oauth2_test.go<br>**Notes**: Tests the behavior when authentication fails but network is available. |

## File System Operations

| Field           | Description                                                                                                                                                                      |
|-----------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Test Case ID    | TC-16                                                                                                                                                                            |
| Title           | Directory Reading                                                                                                                                                                |
| Description     | Verify that the filesystem can correctly list directory contents                                                                                                                 |
| Preconditions   | 1. User is authenticated with valid credentials<br>2. Network connection is available<br>3. Directories with known content exist                                                 |
| Steps           | 1. Read the contents of various directories using os.ReadDir<br>2. Verify that expected items are present<br>3. Verify that item types (file/directory) are correctly identified |
| Expected Result | Directory contents are correctly listed with proper item types                                                                                                                   |
| Actual Result   | [To be filled during test execution]                                                                                                                                             |
| Status          | [Pass/Fail]                                                                                                                                                                      |
| Implementation  | **Test**: TestReaddir<br>**File**: fs/fs_test.go<br>**Notes**: Tests reading directory contents using Go's internal ReadDir function.                                            |

| Field           | Description                                                                                                                                                                                          |
|-----------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Test Case ID    | TC-17                                                                                                                                                                                                |
| Title           | Directory Listing with Shell Commands                                                                                                                                                                |
| Description     | Verify that directory contents can be listed using shell commands                                                                                                                                    |
| Preconditions   | 1. User is authenticated with valid credentials<br>2. Network connection is available<br>3. Directories with known content exist                                                                     |
| Steps           | 1. List directory contents using the 'ls' command with various options<br>2. Verify that expected items are present<br>3. Verify that item properties are correctly displayed with different options |
| Expected Result | Directory contents are correctly listed with proper item properties                                                                                                                                  |
| Actual Result   | [To be filled during test execution]                                                                                                                                                                 |
| Status          | [Pass/Fail]                                                                                                                                                                                          |
| Implementation  | **Test**: TestLs, TestShellFileOperations<br>**File**: fs/fs_test.go<br>**Notes**: Tests directory listing using shell commands like 'ls' with various options.                                      |

| Field           | Description                                                                                                                                                                              |
|-----------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Test Case ID    | TC-18                                                                                                                                                                                    |
| Title           | File Creation and Modification Time                                                                                                                                                      |
| Description     | Verify that files can be created and their modification times can be updated                                                                                                             |
| Preconditions   | 1. User is authenticated with valid credentials<br>2. Network connection is available                                                                                                    |
| Steps           | 1. Create a new empty file using the 'touch' command<br>2. Verify file properties<br>3. Update the file's modification time using 'touch'<br>4. Verify the modification time has changed |
| Expected Result | Files are created with correct properties and modification times can be updated                                                                                                          |
| Actual Result   | [To be filled during test execution]                                                                                                                                                     |
| Status          | [Pass/Fail]                                                                                                                                                                              |
| Implementation  | **Test**: TestTouchOperations<br>**File**: fs/fs_test.go<br>**Notes**: Tests file creation and modification time updates using the 'touch' command.                                      |

| Field           | Description                                                                                                                                                               |
|-----------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Test Case ID    | TC-19                                                                                                                                                                     |
| Title           | File Permissions                                                                                                                                                          |
| Description     | Verify that file permissions are correctly handled                                                                                                                        |
| Preconditions   | 1. User is authenticated with valid credentials<br>2. Network connection is available                                                                                     |
| Steps           | 1. Create files with specific permissions<br>2. Verify the permissions are set correctly<br>3. Change file permissions<br>4. Verify the permissions are updated correctly |
| Expected Result | File permissions are correctly set and can be modified                                                                                                                    |
| Actual Result   | [To be filled during test execution]                                                                                                                                      |
| Status          | [Pass/Fail]                                                                                                                                                               |
| Implementation  | **Test**: TestFilePermissions, TestFileInfo<br>**File**: fs/fs_test.go<br>**Notes**: Tests file permission handling and verification.                                     |

| Field           | Description                                                                                                                                                |
|-----------------|------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Test Case ID    | TC-20                                                                                                                                                      |
| Title           | Directory Creation and Removal                                                                                                                             |
| Description     | Verify that directories can be created and removed                                                                                                         |
| Preconditions   | 1. User is authenticated with valid credentials<br>2. Network connection is available                                                                      |
| Steps           | 1. Create a new directory<br>2. Verify the directory exists with correct properties<br>3. Remove the directory<br>4. Verify the directory no longer exists |
| Expected Result | Directories can be created and removed successfully                                                                                                        |
| Actual Result   | [To be filled during test execution]                                                                                                                       |
| Status          | [Pass/Fail]                                                                                                                                                |
| Implementation  | **Test**: TestDirectoryOperations, TestDirectoryRemoval<br>**File**: fs/fs_test.go<br>**Notes**: Tests directory creation and removal operations.          |

| Field           | Description                                                                                                                            |
|-----------------|----------------------------------------------------------------------------------------------------------------------------------------|
| Test Case ID    | TC-21                                                                                                                                  |
| Title           | File Writing with Offset                                                                                                               |
| Description     | Verify that files can be written to at specific offsets                                                                                |
| Preconditions   | 1. User is authenticated with valid credentials<br>2. Network connection is available                                                  |
| Steps           | 1. Create a file with initial content<br>2. Write new content at a specific offset<br>3. Verify the file content is correctly modified |
| Expected Result | File content can be modified at specific offsets                                                                                       |
| Actual Result   | [To be filled during test execution]                                                                                                   |
| Status          | [Pass/Fail]                                                                                                                            |
| Implementation  | **Test**: TestWriteOffset<br>**File**: fs/fs_test.go<br>**Notes**: Tests writing to files at specific offsets.                         |

| Field           | Description                                                                                                                                                                                                    |
|-----------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Test Case ID    | TC-22                                                                                                                                                                                                          |
| Title           | File Movement Operations                                                                                                                                                                                       |
| Description     | Verify that files can be moved and renamed                                                                                                                                                                     |
| Preconditions   | 1. User is authenticated with valid credentials<br>2. Network connection is available                                                                                                                          |
| Steps           | 1. Create a file<br>2. Move the file to a different location<br>3. Verify the file exists at the new location and not at the old location<br>4. Rename the file<br>5. Verify the file exists with the new name |
| Expected Result | Files can be moved and renamed successfully                                                                                                                                                                    |
| Actual Result   | [To be filled during test execution]                                                                                                                                                                           |
| Status          | [Pass/Fail]                                                                                                                                                                                                    |
| Implementation  | **Test**: TestFileMovementOperations<br>**File**: fs/fs_test.go<br>**Notes**: Tests moving and renaming files.                                                                                                 |

| Field           | Description                                                                                                                                                                          |
|-----------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Test Case ID    | TC-23                                                                                                                                                                                |
| Title           | Positional File Operations                                                                                                                                                           |
| Description     | Verify that files can be read and written at specific positions                                                                                                                      |
| Preconditions   | 1. User is authenticated with valid credentials<br>2. Network connection is available                                                                                                |
| Steps           | 1. Create a file with known content<br>2. Read from specific positions in the file<br>3. Write to specific positions in the file<br>4. Verify the file content is correctly modified |
| Expected Result | Files can be read and written at specific positions                                                                                                                                  |
| Actual Result   | [To be filled during test execution]                                                                                                                                                 |
| Status          | [Pass/Fail]                                                                                                                                                                          |
| Implementation  | **Test**: TestPositionalFileOperations<br>**File**: fs/fs_test.go<br>**Notes**: Tests reading and writing files at specific positions.                                               |

| Field           | Description                                                                                                                                                                                     |
|-----------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Test Case ID    | TC-24                                                                                                                                                                                           |
| Title           | Case Sensitivity Handling                                                                                                                                                                       |
| Description     | Verify that the filesystem correctly handles case sensitivity in filenames                                                                                                                      |
| Preconditions   | 1. User is authenticated with valid credentials<br>2. Network connection is available                                                                                                           |
| Steps           | 1. Create files with names differing only in case<br>2. Verify both files exist and are distinct<br>3. Access files using different case variations<br>4. Verify the correct files are accessed |
| Expected Result | Case sensitivity in filenames is handled correctly                                                                                                                                              |
| Actual Result   | [To be filled during test execution]                                                                                                                                                            |
| Status          | [Pass/Fail]                                                                                                                                                                                     |
| Implementation  | **Test**: TestCaseSensitivityHandling, TestFilenameCase<br>**File**: fs/fs_test.go<br>**Notes**: Tests handling of case sensitivity in filenames.                                               |

| Field           | Description                                                                                                                                                                                   |
|-----------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Test Case ID    | TC-25                                                                                                                                                                                         |
| Title           | Special Characters in Filenames                                                                                                                                                               |
| Description     | Verify that the filesystem correctly handles special characters in filenames                                                                                                                  |
| Preconditions   | 1. User is authenticated with valid credentials<br>2. Network connection is available                                                                                                         |
| Steps           | 1. Create files with special characters in their names<br>2. Verify the files exist and can be accessed<br>3. Create files with disallowed characters<br>4. Verify appropriate error handling |
| Expected Result | Special characters in filenames are handled correctly, and disallowed characters are properly rejected                                                                                        |
| Actual Result   | [To be filled during test execution]                                                                                                                                                          |
| Status          | [Pass/Fail]                                                                                                                                                                                   |
| Implementation  | **Test**: TestNoQuestionMarks, TestDisallowedFilenames<br>**File**: fs/fs_test.go<br>**Notes**: Tests handling of special characters and disallowed characters in filenames.                  |

| Field           | Description                                                                                                                                                                   |
|-----------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Test Case ID    | TC-26                                                                                                                                                                         |
| Title           | Trash Functionality                                                                                                                                                           |
| Description     | Verify that the filesystem correctly handles trash operations                                                                                                                 |
| Preconditions   | 1. User is authenticated with valid credentials<br>2. Network connection is available                                                                                         |
| Steps           | 1. Create a file<br>2. Move the file to trash using GIO<br>3. Verify the file is no longer accessible in its original location<br>4. Verify the file is in the OneDrive trash |
| Expected Result | Files can be moved to trash successfully                                                                                                                                      |
| Actual Result   | [To be filled during test execution]                                                                                                                                          |
| Status          | [Pass/Fail]                                                                                                                                                                   |
| Implementation  | **Test**: TestGIOTrash<br>**File**: fs/fs_test.go<br>**Notes**: Tests moving files to trash using GIO.                                                                        |

| Field           | Description                                                                                                                                                                              |
|-----------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Test Case ID    | TC-27                                                                                                                                                                                    |
| Title           | Directory Paging                                                                                                                                                                         |
| Description     | Verify that the filesystem correctly handles directories with many items requiring paging                                                                                                |
| Preconditions   | 1. User is authenticated with valid credentials<br>2. Network connection is available<br>3. A directory with many items exists                                                           |
| Steps           | 1. List the contents of a directory with many items<br>2. Verify all items are correctly listed<br>3. Access items from different pages<br>4. Verify the items can be accessed correctly |
| Expected Result | Directories with many items are handled correctly with proper paging                                                                                                                     |
| Actual Result   | [To be filled during test execution]                                                                                                                                                     |
| Status          | [Pass/Fail]                                                                                                                                                                              |
| Implementation  | **Test**: TestListChildrenPaging<br>**File**: fs/fs_test.go<br>**Notes**: Tests handling of directories with many items requiring paging.                                                |

| Field           | Description                                                                                                                                                                   |
|-----------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Test Case ID    | TC-28                                                                                                                                                                         |
| Title           | Application-Specific Save Patterns                                                                                                                                            |
| Description     | Verify that the filesystem correctly handles application-specific save patterns                                                                                               |
| Preconditions   | 1. User is authenticated with valid credentials<br>2. Network connection is available                                                                                         |
| Steps           | 1. Simulate LibreOffice save pattern (create temporary file, rename to target)<br>2. Verify the file is saved correctly<br>3. Verify the temporary files are handled properly |
| Expected Result | Application-specific save patterns are handled correctly                                                                                                                      |
| Actual Result   | [To be filled during test execution]                                                                                                                                          |
| Status          | [Pass/Fail]                                                                                                                                                                   |
| Implementation  | **Test**: TestLibreOfficeSavePattern<br>**File**: fs/fs_test.go<br>**Notes**: Tests handling of application-specific save patterns, specifically LibreOffice's pattern.       |

## Authentication and Configuration

| Field           | Description                                                                                                                                                                        |
|-----------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Test Case ID    | TC-29                                                                                                                                                                              |
| Title           | Authentication Configuration Merging                                                                                                                                               |
| Description     | Verify that authentication configuration can be merged with defaults                                                                                                               |
| Preconditions   | 1. Authentication configuration with custom values exists                                                                                                                          |
| Steps           | 1. Create a custom authentication configuration<br>2. Apply default values<br>3. Verify custom values are preserved<br>4. Verify default values are applied for unspecified fields |
| Expected Result | Authentication configuration is correctly merged with defaults                                                                                                                     |
| Actual Result   | [To be filled during test execution]                                                                                                                                               |
| Status          | [Pass/Fail]                                                                                                                                                                        |
| Implementation  | **Test**: TestAuthConfigMerge<br>**File**: fs/graph/oauth2_test.go<br>**Notes**: Tests merging of authentication configuration with defaults.                                      |

| Field           | Description                                                                                                                                                                                                |
|-----------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Test Case ID    | TC-30                                                                                                                                                                                                      |
| Title           | Resource Path Handling                                                                                                                                                                                     |
| Description     | Verify that local filesystem paths are correctly converted to OneDrive API resource paths                                                                                                                  |
| Preconditions   | 1. Various types of local filesystem paths are available                                                                                                                                                   |
| Steps           | 1. Convert local paths with special characters to resource paths<br>2. Convert root path to resource path<br>3. Convert simple and nested paths to resource paths<br>4. Verify all conversions are correct |
| Expected Result | Local filesystem paths are correctly converted to OneDrive API resource paths                                                                                                                              |
| Actual Result   | [To be filled during test execution]                                                                                                                                                                       |
| Status          | [Pass/Fail]                                                                                                                                                                                                |
| Implementation  | **Test**: TestResourcePath<br>**File**: fs/graph/graph_test.go<br>**Notes**: Tests conversion of local filesystem paths to OneDrive API resource paths.                                                    |

## Offline Functionality

| Field           | Description                                                                                                                                                                         |
|-----------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Test Case ID    | TC-31                                                                                                                                                                               |
| Title           | Offline Changes Caching                                                                                                                                                             |
| Description     | Verify that changes made in offline mode are properly cached                                                                                                                        |
| Preconditions   | 1. User is authenticated<br>2. Network connection is unavailable                                                                                                                    |
| Steps           | 1. Create a file in offline mode<br>2. Verify the file exists locally<br>3. Verify the file has the correct content<br>4. Verify the file is marked for synchronization when online |
| Expected Result | Changes made in offline mode are properly cached and marked for synchronization                                                                                                     |
| Actual Result   | [To be filled during test execution]                                                                                                                                                |
| Status          | [Pass/Fail]                                                                                                                                                                         |
| Implementation  | **Test**: TestOfflineChangesCached<br>**File**: fs/offline/offline_test.go<br>**Notes**: Tests that changes made in offline mode are cached and marked for synchronization.         |
