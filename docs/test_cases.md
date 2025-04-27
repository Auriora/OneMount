# OneDriver Test Cases

This document contains standardized test cases for the OneDriver project, covering successful sync, network failure, and invalid credentials scenarios.

## Successful Sync Scenarios

| Field           | Description                                                                                |
|-----------------|--------------------------------------------------------------------------------------------|
| Test Case ID    | TC-01                                                                                      |
| Title           | Directory Tree Synchronization                                                             |
| Description     | Verify that the filesystem can successfully synchronize the directory tree from the root   |
| Preconditions   | 1. User is authenticated with valid credentials<br>2. Network connection is available      |
| Steps           | 1. Initialize the filesystem<br>2. Call SyncDirectoryTree method<br>3. Verify all directories are cached |
| Expected Result | All directory metadata is successfully cached without errors                               |
| Actual Result   | [To be filled during test execution]                                                       |
| Status          | [Pass/Fail]                                                                                |

| Field           | Description                                                                                |
|-----------------|--------------------------------------------------------------------------------------------|
| Test Case ID    | TC-02                                                                                      |
| Title           | File Upload Synchronization                                                                |
| Description     | Verify that a file can be successfully uploaded to OneDrive                                |
| Preconditions   | 1. User is authenticated with valid credentials<br>2. Network connection is available      |
| Steps           | 1. Create a new file in the local filesystem<br>2. Write content to the file<br>3. Wait for the upload to complete<br>4. Verify the file exists on OneDrive with correct content |
| Expected Result | File is successfully uploaded to OneDrive with the correct content                         |
| Actual Result   | [To be filled during test execution]                                                       |
| Status          | [Pass/Fail]                                                                                |

| Field           | Description                                                                                |
|-----------------|--------------------------------------------------------------------------------------------|
| Test Case ID    | TC-03                                                                                      |
| Title           | Repeated File Upload                                                                       |
| Description     | Verify that the same file can be uploaded multiple times with different content            |
| Preconditions   | 1. User is authenticated with valid credentials<br>2. Network connection is available      |
| Steps           | 1. Create a file with initial content<br>2. Wait for upload to complete<br>3. Modify the file content<br>4. Wait for upload to complete<br>5. Repeat steps 3-4 multiple times |
| Expected Result | Each version of the file is successfully uploaded with the correct content                 |
| Actual Result   | [To be filled during test execution]                                                       |
| Status          | [Pass/Fail]                                                                                |

| Field           | Description                                                                                |
|-----------------|--------------------------------------------------------------------------------------------|
| Test Case ID    | TC-04                                                                                      |
| Title           | Large File Upload                                                                          |
| Description     | Verify that large files can be uploaded correctly using upload sessions                    |
| Preconditions   | 1. User is authenticated with valid credentials<br>2. Network connection is available      |
| Steps           | 1. Create a large file (>4MB)<br>2. Write content to the file<br>3. Wait for the upload to complete<br>4. Verify the file exists on OneDrive with correct content |
| Expected Result | Large file is successfully uploaded to OneDrive with the correct content                   |
| Actual Result   | [To be filled during test execution]                                                       |
| Status          | [Pass/Fail]                                                                                |

| Field           | Description                                                                                |
|-----------------|--------------------------------------------------------------------------------------------|
| Test Case ID    | TC-05                                                                                      |
| Title           | File Download Synchronization                                                              |
| Description     | Verify that a file can be successfully downloaded from OneDrive                            |
| Preconditions   | 1. User is authenticated with valid credentials<br>2. Network connection is available<br>3. File exists on OneDrive |
| Steps           | 1. Access a file that exists on OneDrive but not in local cache<br>2. Read the file content<br>3. Verify the content matches what's on OneDrive |
| Expected Result | File is successfully downloaded from OneDrive with the correct content                     |
| Actual Result   | [To be filled during test execution]                                                       |
| Status          | [Pass/Fail]                                                                                |

## Network Failure Scenarios

| Field           | Description                                                                                |
|-----------------|--------------------------------------------------------------------------------------------|
| Test Case ID    | TC-06                                                                                      |
| Title           | Offline File Access                                                                        |
| Description     | Verify that files can be accessed when offline                                             |
| Preconditions   | 1. User is authenticated<br>2. Files have been previously synchronized<br>3. Network connection is unavailable |
| Steps           | 1. Disconnect from the network<br>2. Access a previously synchronized file<br>3. Read the file content |
| Expected Result | File content is successfully read from the local cache                                     |
| Actual Result   | [To be filled during test execution]                                                       |
| Status          | [Pass/Fail]                                                                                |

| Field           | Description                                                                                |
|-----------------|--------------------------------------------------------------------------------------------|
| Test Case ID    | TC-07                                                                                      |
| Title           | Offline File Creation                                                                      |
| Description     | Verify that files can be created when offline                                              |
| Preconditions   | 1. User is authenticated<br>2. Network connection is unavailable                           |
| Steps           | 1. Disconnect from the network<br>2. Create a new file<br>3. Write content to the file<br>4. Verify the file exists locally |
| Expected Result | File is successfully created locally and marked for upload when online                     |
| Actual Result   | [To be filled during test execution]                                                       |
| Status          | [Pass/Fail]                                                                                |

| Field           | Description                                                                                |
|-----------------|--------------------------------------------------------------------------------------------|
| Test Case ID    | TC-08                                                                                      |
| Title           | Offline File Modification                                                                  |
| Description     | Verify that files can be modified when offline                                             |
| Preconditions   | 1. User is authenticated<br>2. File has been previously synchronized<br>3. Network connection is unavailable |
| Steps           | 1. Disconnect from the network<br>2. Modify a previously synchronized file<br>3. Verify the file content is updated locally |
| Expected Result | File is successfully modified locally and marked for upload when online                    |
| Actual Result   | [To be filled during test execution]                                                       |
| Status          | [Pass/Fail]                                                                                |

| Field           | Description                                                                                |
|-----------------|--------------------------------------------------------------------------------------------|
| Test Case ID    | TC-09                                                                                      |
| Title           | Offline File Deletion                                                                      |
| Description     | Verify that files can be deleted when offline                                              |
| Preconditions   | 1. User is authenticated<br>2. File has been previously synchronized<br>3. Network connection is unavailable |
| Steps           | 1. Disconnect from the network<br>2. Delete a previously synchronized file<br>3. Verify the file is no longer accessible locally |
| Expected Result | File is successfully deleted locally and marked for deletion on OneDrive when online       |
| Actual Result   | [To be filled during test execution]                                                       |
| Status          | [Pass/Fail]                                                                                |

| Field           | Description                                                                                |
|-----------------|--------------------------------------------------------------------------------------------|
| Test Case ID    | TC-10                                                                                      |
| Title           | Reconnection Synchronization                                                               |
| Description     | Verify that changes made offline are synchronized when reconnecting                        |
| Preconditions   | 1. User is authenticated<br>2. Changes have been made while offline<br>3. Network connection becomes available |
| Steps           | 1. Make changes to files while offline<br>2. Reconnect to the network<br>3. Wait for synchronization to complete<br>4. Verify changes are reflected on OneDrive |
| Expected Result | All offline changes are successfully synchronized to OneDrive                              |
| Actual Result   | [To be filled during test execution]                                                       |
| Status          | [Pass/Fail]                                                                                |

## Invalid Credentials Scenarios

| Field           | Description                                                                                |
|-----------------|--------------------------------------------------------------------------------------------|
| Test Case ID    | TC-11                                                                                      |
| Title           | Unauthenticated Request Handling                                                           |
| Description     | Verify that requests with invalid authentication are properly handled                      |
| Preconditions   | 1. User has invalid or expired credentials                                                 |
| Steps           | 1. Attempt to access OneDrive resources with invalid credentials<br>2. Verify error handling |
| Expected Result | System properly handles the authentication error and returns appropriate error message     |
| Actual Result   | [To be filled during test execution]                                                       |
| Status          | [Pass/Fail]                                                                                |

| Field           | Description                                                                                |
|-----------------|--------------------------------------------------------------------------------------------|
| Test Case ID    | TC-12                                                                                      |
| Title           | Authentication Token Refresh                                                               |
| Description     | Verify that expired authentication tokens are automatically refreshed                      |
| Preconditions   | 1. User has valid but expired access token<br>2. User has valid refresh token              |
| Steps           | 1. Set the access token to be expired<br>2. Attempt to access OneDrive resources<br>3. Verify token is refreshed automatically |
| Expected Result | Access token is refreshed and the request succeeds                                         |
| Actual Result   | [To be filled during test execution]                                                       |
| Status          | [Pass/Fail]                                                                                |

| Field           | Description                                                                                |
|-----------------|--------------------------------------------------------------------------------------------|
| Test Case ID    | TC-13                                                                                      |
| Title           | Invalid Authentication Code Format                                                         |
| Description     | Verify that invalid authentication code formats are properly handled                       |
| Preconditions   | 1. System is attempting to parse an authentication code                                    |
| Steps           | 1. Provide an invalid authentication code format<br>2. Attempt to parse the code<br>3. Verify error handling |
| Expected Result | System properly handles the invalid format and returns appropriate error                   |
| Actual Result   | [To be filled during test execution]                                                       |
| Status          | [Pass/Fail]                                                                                |

| Field           | Description                                                                                |
|-----------------|--------------------------------------------------------------------------------------------|
| Test Case ID    | TC-14                                                                                      |
| Title           | Authentication Persistence                                                                 |
| Description     | Verify that authentication tokens are properly persisted and loaded                        |
| Preconditions   | 1. User has previously authenticated<br>2. Authentication tokens are saved to file         |
| Steps           | 1. Restart the application<br>2. Verify authentication tokens are loaded from file<br>3. Verify access to OneDrive resources |
| Expected Result | Authentication tokens are successfully loaded and used for accessing resources             |
| Actual Result   | [To be filled during test execution]                                                       |
| Status          | [Pass/Fail]                                                                                |

| Field           | Description                                                                                |
|-----------------|--------------------------------------------------------------------------------------------|
| Test Case ID    | TC-15                                                                                      |
| Title           | Authentication Failure with Network Available                                              |
| Description     | Verify behavior when authentication fails but network is available                         |
| Preconditions   | 1. User has invalid credentials<br>2. Network connection is available                      |
| Steps           | 1. Attempt to authenticate with invalid credentials<br>2. Verify error handling<br>3. Verify system state after failure |
| Expected Result | System properly handles the authentication failure and provides appropriate user feedback  |
| Actual Result   | [To be filled during test execution]                                                       |
| Status          | [Pass/Fail]                                                                                |