# OneMount Integration Test Cases

This document contains standardized integration test cases for the OneMount project. Integration tests focus on verifying that components work correctly together.

## Integration Test Case Format

Each integration test case follows this format:

| Field           | Description                                                |
|-----------------|-----------------------------------------------------------|
| Test Case ID    | IT_FS_XX_YY (where XX is a sequential number and YY is a sub-test number) |
| Title           | Brief descriptive title of the test case                  |
| Description     | Detailed description of what is being tested              |
| Preconditions   | Required state before the test can be executed            |
| Steps           | Sequence of steps to execute the test                     |
| Expected Result | Expected outcome after the test steps are executed        |
| Actual Result   | Actual outcome (to be filled during test execution)       |
| Status          | Current status of the test (Pass/Fail)                    |
| Implementation  | Details about the implementation of the test              |

## Integration Test Cases

| Field           | Description                                                                                                                        |
|-----------------|------------------------------------------------------------------------------------------------------------------------------------|
| Test Case ID    | IT_FS_01_01                                                                                                                        |
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
| Test Case ID    | IT_FS_02_01                                                                                                                                            |
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
| Test Case ID    | IT_FS_03_01                                                                                                                                                |
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
| Test Case ID    | IT_FS_04_01                                                                                                                                            |
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
| Test Case ID    | IT_FS_05_01                                                                                                                                                     |
| Title           | Reconnection Synchronization                                                                                                                                    |
| Description     | Verify that changes made offline are synchronized when reconnecting                                                                                             |
| Preconditions   | 1. User is authenticated<br>2. Changes have been made while offline<br>3. Network connection becomes available                                                  |
| Steps           | 1. Make changes to files while offline<br>2. Reconnect to the network<br>3. Wait for synchronization to complete<br>4. Verify changes are reflected on OneDrive |
| Expected Result | All offline changes are successfully synchronized to OneDrive                                                                                                   |
| Actual Result   | [To be filled during test execution]                                                                                                                            |
| Status          | [Pass/Fail]                                                                                                                                                     |
| Implementation  | **Test**: TestOfflineSynchronization<br>**File**: fs/offline/offline_test.go<br>**Notes**: Tests synchronization when reconnecting after being offline.         |

| Field           | Description                                                                                                                                                                         |
|-----------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Test Case ID    | IT_FS_06_01                                                                                                                                                                         |
| Title           | Offline Changes Caching                                                                                                                                                             |
| Description     | Verify that changes made in offline mode are properly cached                                                                                                                        |
| Preconditions   | 1. User is authenticated<br>2. Network connection is unavailable                                                                                                                    |
| Steps           | 1. Create a file in offline mode<br>2. Verify the file exists locally<br>3. Verify the file has the correct content<br>4. Verify the file is marked for synchronization when online |
| Expected Result | Changes made in offline mode are properly cached and marked for synchronization                                                                                                     |
| Actual Result   | [To be filled during test execution]                                                                                                                                                |
| Status          | [Pass/Fail]                                                                                                                                                                         |
| Implementation  | **Test**: TestOfflineChangesCached<br>**File**: fs/offline/offline_test.go<br>**Notes**: Tests that changes made in offline mode are cached and marked for synchronization.         |
