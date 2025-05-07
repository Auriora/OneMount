# OneMount System Test Cases

This document contains standardized system test cases for the OneMount project. System tests verify the complete integrated system and end-to-end functionality.

## System Test Case Format

Each system test case follows this format:

| Field           | Description                                                |
|-----------------|-----------------------------------------------------------|
| Test Case ID    | ST-XX (where XX is a sequential number)                   |
| Title           | Brief descriptive title of the test case                  |
| Description     | Detailed description of what is being tested              |
| Preconditions   | Required state before the test can be executed            |
| Steps           | Sequence of steps to execute the test                     |
| Expected Result | Expected outcome after the test steps are executed        |
| Actual Result   | Actual outcome (to be filled during test execution)       |
| Status          | Current status of the test (Pass/Fail)                    |
| Implementation  | Details about the implementation of the test              |

## System Test Cases

| Field           | Description                                                                                                                                                                      |
|-----------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Test Case ID    | ST_FS_01_01                                                                                                                                                                      |
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
| Test Case ID    | ST_FS_02_01                                                                                                                                                                                          |
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
| Test Case ID    | ST_FS_03_01                                                                                                                                                                              |
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
| Test Case ID    | ST_FS_04_01                                                                                                                                                               |
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
| Test Case ID    | ST_FS_05_01                                                                                                                                                |
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
| Test Case ID    | ST_FS_06_01                                                                                                                            |
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
| Test Case ID    | ST_FS_07_01                                                                                                                                                                                                    |
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
| Test Case ID    | ST_FS_08_01                                                                                                                                                                          |
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
| Test Case ID    | ST_FS_09_01                                                                                                                                                                                     |
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
| Test Case ID    | ST_FS_10_01                                                                                                                                                                                   |
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
| Test Case ID    | ST_FS_11_01                                                                                                                                                                   |
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
| Test Case ID    | ST_FS_12_01                                                                                                                                                                              |
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
| Test Case ID    | ST_FS_13_01                                                                                                                                                                   |
| Title           | Application-Specific Save Patterns                                                                                                                                            |
| Description     | Verify that the filesystem correctly handles application-specific save patterns                                                                                               |
| Preconditions   | 1. User is authenticated with valid credentials<br>2. Network connection is available                                                                                         |
| Steps           | 1. Simulate LibreOffice save pattern (create temporary file, rename to target)<br>2. Verify the file is saved correctly<br>3. Verify the temporary files are handled properly |
| Expected Result | Application-specific save patterns are handled correctly                                                                                                                      |
| Actual Result   | [To be filled during test execution]                                                                                                                                          |
| Status          | [Pass/Fail]                                                                                                                                                                   |
| Implementation  | **Test**: TestLibreOfficeSavePattern<br>**File**: fs/fs_test.go<br>**Notes**: Tests handling of application-specific save patterns, specifically LibreOffice's pattern.       |
