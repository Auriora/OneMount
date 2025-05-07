# OneMount Security Test Cases

This document contains standardized security test cases for the OneMount project. Security tests identify vulnerabilities and security issues in the system.

## Security Test Case Format

Each security test case follows this format:

| Field           | Description                                                |
|-----------------|-----------------------------------------------------------|
| Test Case ID    | ST_SC_XX_YY (where XX is a sequential number and YY is a sub-test number) |
| Title           | Brief descriptive title of the test case                  |
| Description     | Detailed description of what is being tested              |
| Preconditions   | Required state before the test can be executed            |
| Steps           | Sequence of steps to execute the test                     |
| Expected Result | Expected outcome after the test steps are executed        |
| Actual Result   | Actual outcome (to be filled during test execution)       |
| Status          | Current status of the test (Pass/Fail)                    |
| Implementation  | Details about the implementation of the test              |

## Security Test Cases

| Field           | Description                                                                                                                                    |
|-----------------|------------------------------------------------------------------------------------------------------------------------------------------------|
| Test Case ID    | ST_SC_01_01                                                                                                                                    |
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
| Test Case ID    | ST_SC_02_01                                                                                                                    |
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
| Test Case ID    | ST_SC_03_01                                                                                                                                                   |
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
| Test Case ID    | ST_SC_04_01                                                                                                                  |
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
| Test Case ID    | ST_SC_05_01                                                                                                                                                             |
| Title           | Authentication Failure with Network Available                                                                                                                           |
| Description     | Verify behavior when authentication fails but network is available                                                                                                      |
| Preconditions   | 1. User has invalid credentials<br>2. Network connection is available                                                                                                   |
| Steps           | 1. Attempt to authenticate with invalid credentials<br>2. Verify error handling<br>3. Verify system state after failure                                                 |
| Expected Result | System properly handles the authentication failure and provides appropriate user feedback                                                                               |
| Actual Result   | [To be filled during test execution]                                                                                                                                    |
| Status          | [Pass/Fail]                                                                                                                                                             |
| Implementation  | **Test**: TestAuthFailureWithNetworkAvailable<br>**File**: fs/graph/oauth2_test.go<br>**Notes**: Tests the behavior when authentication fails but network is available. |

| Field           | Description                                                                                                                                                                        |
|-----------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Test Case ID    | ST_SC_06_01                                                                                                                                                                        |
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
| Test Case ID    | ST_SC_07_01                                                                                                                                                                                                |
| Title           | Resource Path Handling                                                                                                                                                                                     |
| Description     | Verify that local filesystem paths are correctly converted to OneDrive API resource paths                                                                                                                  |
| Preconditions   | 1. Various types of local filesystem paths are available                                                                                                                                                   |
| Steps           | 1. Convert local paths with special characters to resource paths<br>2. Convert root path to resource path<br>3. Convert simple and nested paths to resource paths<br>4. Verify all conversions are correct |
| Expected Result | Local filesystem paths are correctly converted to OneDrive API resource paths                                                                                                                              |
| Actual Result   | [To be filled during test execution]                                                                                                                                                                       |
| Status          | [Pass/Fail]                                                                                                                                                                                                |
| Implementation  | **Test**: TestResourcePath<br>**File**: fs/graph/graph_test.go<br>**Notes**: Tests conversion of local filesystem paths to OneDrive API resource paths.                                                    |
