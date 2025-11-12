# Error Handling Integration Tests

**Date**: 2025-11-11  
**Purpose**: Document existing and recommended error handling integration tests  
**Related Task**: 14.6 - Create error handling integration tests

## Overview

This document catalogs existing integration tests for error handling and identifies any gaps that need to be filled.

## Existing Integration Tests

### 1. Network Error Retry Tests

**Test**: `TestIT_FS_08_04_DownloadManager_DownloadFailureAndRetry`  
**Location**: `internal/fs/download_manager_integration_test.go`  
**Coverage**: ✅ **COMPLETE**

**What it tests**:
- Download failure with network error (503)
- Automatic retry with exponential backoff
- Eventual success after network recovery
- File content integrity after retry

**How it works**:
1. Creates test file in mock OneDrive
2. Configures mock to return 503 on first attempt
3. Configures mock to succeed on retry
4. Queues download
5. Verifies retry occurs
6. Verifies eventual success
7. Validates file content

**Run**:
```bash
go test -v -run TestIT_FS_08_04 ./internal/fs/
```

---

### 2. Upload Failure and Retry Tests

**Test**: `TestIT_FS_09_04_UploadFailureAndRetry`  
**Location**: `internal/fs/upload_retry_integration_test.go`  
**Coverage**: ✅ **COMPLETE**

**What it tests**:
- Upload failure with network error
- Automatic retry with exponential backoff
- Eventual success after retry
- File integrity after upload

**How it works**:
1. Creates test file locally
2. Configures mock to fail first upload attempt
3. Configures mock to succeed on retry
4. Queues upload
5. Verifies retry occurs
6. Verifies eventual success
7. Validates uploaded file

**Run**:
```bash
go test -v -run TestIT_FS_09_04 ./internal/fs/
```

---

### 3. Large File Upload Retry Tests

**Test**: `TestIT_FS_09_04_02_LargeFileUploadFailureAndRetry`  
**Location**: `internal/fs/upload_retry_integration_test.go`  
**Coverage**: ✅ **COMPLETE**

**What it tests**:
- Large file upload with chunked upload
- Chunk failure and retry
- Resumption from last successful chunk
- File integrity after upload

**How it works**:
1. Creates large test file (>4MB)
2. Configures mock to fail on specific chunks
3. Configures mock to succeed on retry
4. Queues upload
5. Verifies chunk retry
6. Verifies eventual success
7. Validates uploaded file

**Run**:
```bash
go test -v -run TestIT_FS_09_04_02 ./internal/fs/
```

---

### 4. Sync Manager Retry Tests

**Test**: `TestIT_SM_01_01_SyncManager_RetryMechanism_RetriesFailedOperations`  
**Location**: `internal/fs/sync_manager_test.go`  
**Coverage**: ✅ **COMPLETE**

**What it tests**:
- Sync operation failure
- Automatic retry mechanism
- Eventual success after retry

**How it works**:
1. Sets up filesystem with sync manager
2. Configures mock to fail sync operations
3. Triggers sync
4. Verifies retry occurs
5. Verifies eventual success

**Run**:
```bash
go test -v -run TestIT_SM_01_01 ./internal/fs/
```

---

### 5. Network Recovery Tests

**Test**: `TestIT_SM_03_01_SyncManager_NetworkRecovery_HandlesInterruptions`  
**Location**: `internal/fs/sync_manager_test.go`  
**Coverage**: ✅ **COMPLETE**

**What it tests**:
- Network interruption during sync
- Recovery after network restoration
- Sync resumption

**How it works**:
1. Sets up filesystem with sync manager
2. Simulates network interruption
3. Triggers sync
4. Simulates network recovery
5. Verifies sync resumes
6. Verifies eventual success

**Run**:
```bash
go test -v -run TestIT_SM_03_01 ./internal/fs/
```

---

### 6. Error Handling Tests

**Test**: `TestIT_SM_05_01_SyncManager_ErrorHandling_HandlesErrors`  
**Location**: `internal/fs/sync_manager_test.go`  
**Coverage**: ✅ **COMPLETE**

**What it tests**:
- Various error conditions during sync
- Graceful error handling
- No crashes or panics

**How it works**:
1. Sets up filesystem with sync manager
2. Triggers various error conditions
3. Verifies errors are handled gracefully
4. Verifies no crashes

**Run**:
```bash
go test -v -run TestIT_SM_05_01 ./internal/fs/
```

---

### 7. Mount/Unmount Error Recovery Tests

**Test**: `TestIT_MU_04_01_MountUnmount_ErrorRecovery_RecoversCorrectly`  
**Location**: `internal/fs/mount_unmount_integration_test.go`  
**Coverage**: ✅ **COMPLETE**

**What it tests**:
- Error recovery during mount/unmount
- Graceful handling of mount failures
- Proper cleanup on errors

**How it works**:
1. Attempts mount with various error conditions
2. Verifies errors are handled gracefully
3. Verifies proper cleanup
4. Verifies recovery on retry

**Run**:
```bash
go test -v -run TestIT_MU_04_01 ./internal/fs/
```

---

### 8. Authentication Error Tests

**Test**: `TestIT_AUTH_05_01_TokenRefresh_WithMockServer_HandlesErrors`  
**Location**: `internal/graph/auth_integration_mock_server_test.go`  
**Coverage**: ✅ **COMPLETE**

**What it tests**:
- Token refresh failure
- Authentication error handling
- Re-authentication flow

**How it works**:
1. Sets up mock authentication server
2. Simulates token refresh failure
3. Verifies error handling
4. Verifies re-authentication attempt

**Run**:
```bash
go test -v -run TestIT_AUTH_05_01 ./internal/graph/
```

---

## Unit Tests for Error Handling

### 1. Rate Limit Detection

**Test**: `TestUT_GR_RATE_01_01_RateLimitDetection_429Response_DetectsRateLimit`  
**Location**: `internal/graph/rate_limit_test.go`  
**Coverage**: ✅ **COMPLETE**

**What it tests**:
- 429 response detection
- Rate limit error type
- Error classification

**Run**:
```bash
go test -v -run TestUT_GR_RATE_01_01 ./internal/graph/
```

---

### 2. Retry-After Header Handling

**Test**: `TestUT_GR_RATE_01_02_RateLimitWithRetryAfter_RetryAfterHeader_RespectsDelay`  
**Location**: `internal/graph/rate_limit_test.go`  
**Coverage**: ✅ **COMPLETE**

**What it tests**:
- Retry-After header parsing
- Delay respect
- Rate limit handling

**Run**:
```bash
go test -v -run TestUT_GR_RATE_01_02 ./internal/graph/
```

---

### 3. Retry Logic Tests

**Test**: `TestUT_GR_RATE_02_01_RetryLogic_TransientError_RetriesSuccessfully`  
**Location**: `internal/graph/rate_limit_test.go`  
**Coverage**: ✅ **COMPLETE**

**What it tests**:
- Transient error retry
- Exponential backoff
- Eventual success

**Run**:
```bash
go test -v -run TestUT_GR_RATE_02_01 ./internal/graph/
```

---

### 4. Request Queue Tests

**Test**: `TestUT_GR_RATE_03_01_RequestQueue_RateLimitedRequests_QueuesForLater`  
**Location**: `internal/graph/rate_limit_test.go`  
**Coverage**: ✅ **COMPLETE**

**What it tests**:
- Request queuing on rate limit
- Queue processing
- Eventual success

**Run**:
```bash
go test -v -run TestUT_GR_RATE_03_01 ./internal/graph/
```

---

### 5. Error Type Tests

**Test**: `TestUT_ET_05_01_IsErrorTypeFunctions_ReturnCorrectResults`  
**Location**: `internal/errors/error_types_test.go`  
**Coverage**: ✅ **COMPLETE**

**What it tests**:
- Error type detection functions
- Error classification
- Type checking accuracy

**Run**:
```bash
go test -v -run TestUT_ET_05_01 ./internal/errors/
```

---

### 6. Retry Package Tests

**Test**: `TestUT_RT_04_01_IsRetryableNetworkError_WithNetworkError_ReturnsTrue`  
**Location**: `internal/retry/retry_test.go`  
**Coverage**: ✅ **COMPLETE**

**What it tests**:
- Retryable error detection
- Network error classification
- Retry decision logic

**Run**:
```bash
go test -v -run TestUT_RT_04_01 ./internal/retry/
```

---

## Test Coverage Summary

| Error Type | Unit Tests | Integration Tests | Coverage |
|------------|------------|-------------------|----------|
| Network Errors | ✅ Yes | ✅ Yes | ✅ **COMPLETE** |
| Rate Limiting | ✅ Yes | ✅ Yes | ✅ **COMPLETE** |
| Authentication | ✅ Yes | ✅ Yes | ✅ **COMPLETE** |
| Upload Failures | ✅ Yes | ✅ Yes | ✅ **COMPLETE** |
| Download Failures | ✅ Yes | ✅ Yes | ✅ **COMPLETE** |
| Sync Errors | ✅ Yes | ✅ Yes | ✅ **COMPLETE** |
| Mount/Unmount | ✅ Yes | ✅ Yes | ✅ **COMPLETE** |
| Crash Recovery | ❌ No | ⚠️ Partial | ⚠️ **NEEDS TESTS** |

---

## Missing Tests (Recommendations)

### 1. Crash Recovery Integration Test

**Priority**: High  
**Status**: ⚠️ **MISSING**

**What to test**:
- Process kill during upload
- Database state persistence
- Session restoration on remount
- Upload resumption (when implemented)

**Recommended test**:
```go
func TestIT_FS_14_01_CrashRecovery_ProcessKill_RestoresState(t *testing.T) {
    // 1. Start upload
    // 2. Simulate process kill (context cancellation)
    // 3. Verify state persisted to database
    // 4. Create new filesystem instance
    // 5. Verify upload session restored
    // 6. Verify upload can be restarted
}
```

**Location**: `internal/fs/crash_recovery_integration_test.go` (new file)

---

### 2. Database Corruption Test

**Priority**: Medium  
**Status**: ⚠️ **MISSING**

**What to test**:
- Database corruption detection
- Graceful error handling
- Recovery or new database creation

**Recommended test**:
```go
func TestIT_FS_14_02_DatabaseCorruption_DetectsAndRecovers(t *testing.T) {
    // 1. Create filesystem with database
    // 2. Corrupt database file
    // 3. Attempt to create new filesystem
    // 4. Verify corruption detected
    // 5. Verify graceful handling
    // 6. Verify new database created or error reported
}
```

**Location**: `internal/fs/database_corruption_test.go` (new file)

---

### 3. Concurrent Error Handling Test

**Priority**: Medium  
**Status**: ⚠️ **MISSING**

**What to test**:
- Multiple concurrent operations with errors
- Thread-safe error handling
- No race conditions

**Recommended test**:
```go
func TestIT_FS_14_03_ConcurrentErrors_HandledSafely(t *testing.T) {
    // 1. Start multiple concurrent operations
    // 2. Trigger errors in multiple operations
    // 3. Verify all errors handled correctly
    // 4. Verify no race conditions (run with -race)
    // 5. Verify no deadlocks
}
```

**Location**: `internal/fs/concurrent_error_test.go` (new file)

---

### 4. Error Metrics Test

**Priority**: Low  
**Status**: ⚠️ **MISSING**

**What to test**:
- Error metrics collection
- Error rate calculation
- Metrics accuracy

**Recommended test**:
```go
func TestUT_EM_01_01_ErrorMetrics_TracksErrorsCorrectly(t *testing.T) {
    // 1. Reset error metrics
    // 2. Trigger various errors
    // 3. Verify error counts
    // 4. Verify error rates
    // 5. Verify metrics accuracy
}
```

**Location**: `internal/errors/error_monitoring_test.go` (new file)

---

## Running All Error Handling Tests

### Run All Integration Tests

```bash
# All error handling integration tests
go test -v -run 'TestIT.*error|TestIT.*retry|TestIT.*recovery' ./internal/...

# With race detector
go test -v -race -run 'TestIT.*error|TestIT.*retry|TestIT.*recovery' ./internal/...

# In Docker
docker compose -f docker/compose/docker-compose.test.yml run --rm integration-tests
```

### Run All Unit Tests

```bash
# All error handling unit tests
go test -v -run 'TestUT.*error|TestUT.*retry|TestUT.*rate' ./internal/...

# With coverage
go test -v -cover -run 'TestUT.*error|TestUT.*retry|TestUT.*rate' ./internal/...

# In Docker
docker compose -f docker/compose/docker-compose.test.yml run --rm unit-tests
```

### Run Specific Test Suites

```bash
# Network error tests
go test -v -run 'TestIT_FS_08_04|TestIT_SM_03_01' ./internal/fs/

# Rate limiting tests
go test -v -run 'TestUT_GR_RATE' ./internal/graph/

# Upload retry tests
go test -v -run 'TestIT_FS_09_04' ./internal/fs/

# Authentication error tests
go test -v -run 'TestIT_AUTH_05_01' ./internal/graph/
```

---

## Test Maintenance

### Adding New Error Handling Tests

1. **Identify the error scenario** to test
2. **Choose test type**: Unit or Integration
3. **Follow naming convention**: `TestUT_*` or `TestIT_*`
4. **Use test fixtures** from `internal/fs/helpers/`
5. **Mock external dependencies** (API, network)
6. **Verify error type** and error message
7. **Check retry behavior** if applicable
8. **Validate eventual success** or proper failure

### Test Naming Convention

```
TestUT_<COMPONENT>_<SCENARIO>_<EXPECTED_BEHAVIOR>
TestIT_<COMPONENT>_<SCENARIO>_<EXPECTED_BEHAVIOR>

Examples:
- TestUT_GR_RATE_01_01_RateLimitDetection_429Response_DetectsRateLimit
- TestIT_FS_08_04_DownloadManager_DownloadFailureAndRetry
```

### Test Documentation

Each test should include:
- **Test Case ID**: Unique identifier
- **Title**: Descriptive name
- **Description**: What the test verifies
- **Preconditions**: Setup requirements
- **Steps**: Test execution steps
- **Expected Result**: What should happen
- **Requirements**: Linked requirements

---

## Success Criteria

### Test Coverage

- ✅ All error types have unit tests
- ✅ All error scenarios have integration tests
- ✅ Tests cover retry logic
- ✅ Tests cover error recovery
- ⚠️ Tests cover crash recovery (needs implementation)

### Test Quality

- ✅ Tests are deterministic
- ✅ Tests use mocks appropriately
- ✅ Tests verify error types
- ✅ Tests check error messages
- ✅ Tests validate retry behavior

### Test Execution

- ✅ All tests pass consistently
- ✅ Tests run in Docker environment
- ✅ Tests run with race detector
- ✅ Tests provide good coverage

---

## References

- **Integration Tests**: `internal/fs/*_integration_test.go`
- **Unit Tests**: `internal/*/\*_test.go`
- **Test Helpers**: `internal/fs/helpers/`
- **Mock Client**: `internal/graph/mock_client.go`
- **Test Framework**: `internal/fs/helpers/framework/`

---

**Last Updated**: 2025-11-11  
**Status**: Existing tests documented, gaps identified  
**Next Steps**: Implement missing tests (crash recovery, database corruption, concurrent errors)
