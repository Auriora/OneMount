# OneMount Unit Test Cases

This document contains standardized unit test cases for the OneMount project. Unit tests focus on testing individual functions or methods in isolation.

## Unit Test Case Format

Each unit test case follows this format:

| Field           | Description                                                |
|-----------------|-----------------------------------------------------------|
| Test Case ID    | UT-XX (where XX is a sequential number)                   |
| Title           | Brief descriptive title of the test case                  |
| Description     | Detailed description of what is being tested              |
| Preconditions   | Required state before the test can be executed            |
| Steps           | Sequence of steps to execute the test                     |
| Expected Result | Expected outcome after the test steps are executed        |
| Actual Result   | Actual outcome (to be filled during test execution)       |
| Status          | Current status of the test (Pass/Fail)                    |
| Implementation  | Details about the implementation of the test              |

## Unit Test Cases

| Field           | Description                                                                                                                                                          |
|-----------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Test Case ID    | UT-01                                                                                                                                                                |
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
| Test Case ID    | UT-02                                                                                                                                                                            |
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
| Test Case ID    | UT-03                                                                                                                                                                         |
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
| Test Case ID    | UT-04                                                                                                                                                             |
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
| Test Case ID    | UT-05                                                                                                                                           |
| Title           | File Download Synchronization                                                                                                                   |
| Description     | Verify that a file can be successfully downloaded from OneDrive                                                                                 |
| Preconditions   | 1. User is authenticated with valid credentials<br>2. Network connection is available<br>3. File exists on OneDrive                             |
| Steps           | 1. Access a file that exists on OneDrive but not in local cache<br>2. Read the file content<br>3. Verify the content matches what's on OneDrive |
| Expected Result | File is successfully downloaded from OneDrive with the correct content                                                                          |
| Actual Result   | [To be filled during test execution]                                                                                                            |
| Status          | [Pass/Fail]                                                                                                                                     |
| Implementation  | **Test**: TestBasicFileSystemOperations<br>**File**: fs/fs_test.go<br>**Notes**: Tests reading files, which triggers downloads if not in cache. |