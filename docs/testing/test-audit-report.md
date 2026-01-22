# Test Naming Convention Audit Report

**Date**: January 22, 2026  
**Task**: 46.1.1 - Audit all test functions for correct naming conventions  
**Status**: CRITICAL - Blocking all test execution

## Executive Summary

This audit analyzed **752 test functions** across **170 test files** in the OneMount codebase to verify compliance with the project's test naming conventions.

### Key Findings

- **Total Tests**: 752
- **Properly Labeled**: 526 (70%)
- **Unlabeled/Mislabeled**: 226 (30%)

### Test Distribution by Type

| Test Type | Count | Percentage | Naming Convention |
|-----------|-------|------------|-------------------|
| Unit Tests | 288 | 38.3% | `TestUT_*` |
| Integration Tests | 138 | 18.4% | `TestIT_*` |
| Property-Based Tests | 94 | 12.5% | `TestProperty*` |
| System Tests | 6 | 0.8% | `TestSystemST_*` |
| **Unlabeled Tests** | **226** | **30.0%** | **None** |

## Naming Conventions

The project uses the following naming conventions to distinguish test types:

### 1. Unit Tests (`TestUT_*`)
- **Purpose**: Test individual functions/methods in isolation
- **Dependencies**: No external dependencies, use mocks only
- **Authentication**: Not required
- **Example**: `TestUT_FS_CacheHit`

### 2. Integration Tests (`TestIT_*`)
- **Purpose**: Test real API interactions and component integration
- **Dependencies**: Requires authentication tokens
- **Authentication**: Required (real OneDrive API)
- **Example**: `TestIT_FS_FileUpload`

### 3. Property-Based Tests (`TestProperty*`)
- **Purpose**: Generative testing with random inputs
- **Dependencies**: Varies by test
- **Authentication**: May or may not require auth
- **Example**: `TestProperty24_OfflineDetection`

### 4. System Tests (`TestSystemST_*`)
- **Purpose**: End-to-end testing with real OneDrive
- **Dependencies**: Full system setup with mounting
- **Authentication**: Required (real OneDrive API)
- **Example**: `TestSystemST_CompleteWorkflow`

## Critical Issues Identified

### Issue 1: 226 Unlabeled Tests (30% of total)

These tests do not follow any naming convention, making it impossible to:
- Determine which tests require authentication
- Run specific test types in isolation
- Understand test dependencies without reading the code
- Properly configure CI/CD pipelines

### Issue 2: Inconsistent Authentication Requirements

Many unlabeled tests use `SetupFSTestFixture` which requires authentication, but are not labeled as integration tests. This causes:
- Unit test runs to fail when auth tokens are missing
- Confusion about which tests can run in which environments
- Wasted CI/CD resources attempting to run tests that will fail

### Issue 3: Mixed Test Types in Same Files

Some test files contain both unit tests and integration tests without clear separation, making it difficult to:
- Run only fast unit tests during development
- Skip slow integration tests when appropriate
- Understand the scope of each test file

## Detailed Analysis by File

### Files with Unlabeled Tests Requiring Authentication

The following files contain unlabeled tests that use `SetupFSTestFixture` or similar auth-requiring fixtures:

#### `internal/fs/cache_test.go` (REQUIRES_AUTH)
- `TestGetChildrenIDUsesMetadataStoreWhenOffline`
- `TestGetPathUsesMetadataStoreWhenOffline`
- `TestGetChildrenIDReturnsQuicklyWhenUncached`
- `TestGetChildrenIDDoesNotCallGraphWhenMetadataPresent`
- `TestFallbackRootFromMetadata`

**Recommendation**: These should be labeled as `TestIT_*` or refactored to use mocks and labeled as `TestUT_*`.

#### `internal/fs/concurrency_test.go` (REQUIRES_AUTH)
- `TestConcurrentFileAccess`
- `TestConcurrentCacheOperations`
- `TestDeadlockPrevention`
- `TestDirectoryEnumerationWhileRefreshing`
- `TestHighConcurrencyStress`
- `TestConcurrentDirectoryOperations`

**Recommendation**: These should be labeled as `TestIT_*` since they test real filesystem operations.

#### `internal/fs/dbus_test.go` (REQUIRES_AUTH)
- `TestDBusServer_GetFileStatus`
- `TestDBusServer_GetFileStatus_WithRealFiles`
- `TestDBusServer_SendFileStatusUpdate`
- `TestDBusServiceNameGeneration`
- `TestSetDBusServiceNameForMount`
- `TestDBusServer_MultipleInstances`
- `TestSplitPathComponents`
- `TestFindInodeByPath_PathTraversal`

**Recommendation**: Split into unit tests (with mocks) and integration tests (with real filesystem).

#### `internal/fs/delta_test.go` (REQUIRES_AUTH)
- `TestApplyDeltaPersistsMetadataOnMetadataOnlyChange`
- `TestApplyDeltaRemoteInvalidationTransitionsMetadata`
- `TestApplyDeltaMoveUpdatesMetadataEntry`
- `TestApplyDeltaPinnedFileQueuesHydration`
- `TestDesiredDeltaIntervalUsesActiveWindow`
- `TestDesiredDeltaIntervalFallsBackAfterWindow`
- `TestDesiredDeltaIntervalUsesNotifierHealthHealthy`
- `TestDesiredDeltaIntervalUsesNotifierHealthDegraded`
- `TestDesiredDeltaIntervalUsesNotifierHealthFailedRecovery`
- `TestApplyDeltaTransitionsStateOnRemoteInvalidation`

**Recommendation**: These should be labeled as `TestIT_*` since they test delta sync with real filesystem.

### Files with Unlabeled Tests Using Mocks

The following files contain tests that appear to use mocks and could be unit tests:

#### `internal/metadata/store_test.go` (USES_MOCKS)
- `TestBoltStoreSaveAndGet`
- `TestBoltStoreUpdate`

**Recommendation**: Label as `TestUT_*` since they test the metadata store in isolation.

#### `internal/metadata/entry_test.go` (UNCLEAR)
- `TestEntryValidateDefaults`
- `TestEntryValidateRejectsBadState`
- `TestEntryValidateRejectsOverlayPolicy`

**Recommendation**: Label as `TestUT_*` since they test validation logic in isolation.

### Files with Unclear Requirements

The following files contain tests where the authentication requirement is unclear:

#### `cmd/common/config_test.go` (UNCLEAR)
- `TestDefaultDeltaIntervalIsFiveMinutes`
- `TestValidateConfigOverlayPolicy`
- `TestValidateRealtimeConfigDefaults`
- `TestDefaultActiveDeltaTuning`
- `TestValidateConfigResetsInvalidActiveDeltaTuning`
- `TestHydrationConfigDefaultsAndValidation`
- `TestMetadataQueueDefaultsAndValidation`
- `TestRealtimeFallbackValidationBounds`

**Recommendation**: Review each test. Configuration validation tests should be `TestUT_*`.

#### `internal/fs/metadata_store_test.go` (UNCLEAR)
- `TestMetadataEntryFromInodeStateInference`
- `TestBootstrapMetadataStoreMigratesLegacyEntries`
- `TestInodeFromMetadataEntry`
- `TestPendingRemoteMetadataUpdates`
- `TestGetIDLoadsFromMetadataStore`

**Recommendation**: Review each test. Metadata store tests should likely be `TestUT_*` with mocks.

### Test Framework and Helper Tests

The following files contain tests for the test framework itself:

#### `internal/testutil/framework/*_test.go`
- Multiple test framework validation tests
- Security test framework tests
- Performance test framework tests
- Network simulator tests

**Recommendation**: Label as `TestUT_*` since they test the test infrastructure itself.

## Recommendations

### Immediate Actions (Priority: CRITICAL)

1. **Label all 226 unlabeled tests** according to their actual requirements:
   - Tests using `SetupFSTestFixture` → `TestIT_*`
   - Tests using mocks only → `TestUT_*`
   - End-to-end tests → `TestSystemST_*`

2. **Create separate test fixtures** (Task 46.1.2):
   - `SetupMockFSTestFixture` for unit tests
   - `SetupIntegrationFSTestFixture` for integration tests
   - `SetupSystemTestFixture` for system tests

3. **Refactor unit tests** (Task 46.1.3):
   - Convert `TestUT_*` tests to use mock fixtures
   - Ensure no unit tests require authentication

### Short-term Actions (Priority: HIGH)

4. **Add skip logic to integration tests** (Task 46.1.4):
   ```go
   if !authAvailable {
       t.Skip("Auth tokens required for integration test")
   }
   ```

5. **Verify system tests** (Task 46.1.5):
   - Ensure proper resource cleanup
   - Add skip logic for missing auth

6. **Update documentation**:
   - Document test naming conventions in `docs/testing/`
   - Create test fixture usage guide
   - Add examples for each test type

### Long-term Actions (Priority: MEDIUM)

7. **Enforce naming conventions**:
   - Add linter rules to check test naming
   - Add pre-commit hooks to validate test names
   - Update CI/CD to fail on unlabeled tests

8. **Improve test organization**:
   - Separate unit and integration tests into different files
   - Group related tests together
   - Consider separate directories for different test types

## Impact Assessment

### Current State
- **Unit test runs fail** when auth tokens are missing
- **CI/CD pipelines waste resources** running tests that will fail
- **Developers cannot easily run** only fast unit tests
- **Test execution time is unpredictable** due to mixed test types

### After Remediation
- **Unit tests run independently** without external dependencies
- **Integration tests skip gracefully** when auth is unavailable
- **CI/CD pipelines are efficient** with proper test categorization
- **Developers can run fast feedback loops** with unit tests only

## Appendix A: Complete List of Unlabeled Tests

See `/tmp/test_audit_analysis.txt` for the complete list of all 226 unlabeled tests organized by file.

## Appendix B: Test Naming Convention Examples

### Good Examples

```go
// Unit test - uses mocks, no auth required
func TestUT_FS_CacheHit(t *testing.T) {
    fixture := SetupMockFSTestFixture(t)
    defer fixture.Cleanup()
    // Test logic...
}

// Integration test - uses real API, requires auth
func TestIT_FS_FileUpload(t *testing.T) {
    if !authAvailable() {
        t.Skip("Auth tokens required")
    }
    fixture := SetupIntegrationFSTestFixture(t)
    defer fixture.Cleanup()
    // Test logic...
}

// Property-based test
func TestProperty24_OfflineDetection(t *testing.T) {
    rapid.Check(t, func(t *rapid.T) {
        // Property test logic...
    })
}

// System test - end-to-end with mounting
func TestSystemST_CompleteWorkflow(t *testing.T) {
    if !authAvailable() {
        t.Skip("Auth tokens required")
    }
    fixture := SetupSystemTestFixture(t)
    defer fixture.Cleanup()
    // Test logic...
}
```

### Bad Examples

```go
// BAD: No prefix, unclear what type of test this is
func TestCacheHit(t *testing.T) {
    // ...
}

// BAD: Unit test prefix but requires auth
func TestUT_FS_RealAPICall(t *testing.T) {
    fixture := SetupFSTestFixture(t) // Requires auth!
    // ...
}

// BAD: Integration test prefix but uses mocks
func TestIT_FS_MockedOperation(t *testing.T) {
    mock := NewMockGraphProvider()
    // ...
}
```

## Appendix C: Test Execution Commands

### Run only unit tests (no auth required)
```bash
docker compose -f docker/compose/docker-compose.test.yml run --rm \
  test-runner go test -v -run "^TestUT_" ./...
```

### Run only integration tests (requires auth)
```bash
docker compose -f docker/compose/docker-compose.test.yml \
  -f docker/compose/docker-compose.auth.yml run --rm \
  test-runner go test -v -run "^TestIT_" ./...
```

### Run only property-based tests
```bash
docker compose -f docker/compose/docker-compose.test.yml run --rm \
  test-runner go test -v -run "^TestProperty" ./...
```

### Run only system tests (requires auth)
```bash
docker compose -f docker/compose/docker-compose.test.yml \
  -f docker/compose/docker-compose.auth.yml run --rm \
  test-runner go test -v -run "^TestSystemST_" ./tests/system
```

## Conclusion

This audit has identified significant issues with test naming conventions in the OneMount codebase. **30% of tests (226 out of 752) are unlabeled**, making it impossible to run specific test types in isolation and causing unit test runs to fail when authentication is unavailable.

**Immediate action is required** to label all tests according to their actual requirements and create appropriate test fixtures for each test type. This is blocking proper test execution and CI/CD pipeline configuration.

The next steps are outlined in tasks 46.1.2 through 46.1.5, which will:
1. Create separate test fixtures for each test type
2. Refactor unit tests to use mocks
3. Add skip logic to integration and system tests
4. Verify proper resource cleanup

Once these tasks are complete, the test suite will be properly organized, and developers will be able to run fast unit tests during development while reserving slower integration and system tests for CI/CD pipelines.
