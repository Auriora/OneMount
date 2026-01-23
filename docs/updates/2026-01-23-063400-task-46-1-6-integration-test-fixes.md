# Task 46.1.6: Integration Test Auth Fixes

**Date**: 2026-01-23  
**Task**: Fix integration tests that were incorrectly labeled or bypassing auth skip logic  
**Status**: ✅ COMPLETED

## Problem Summary

During the audit of TestIT_* integration tests, we discovered that some tests were:
1. **Mislabeled** - Using TestIT_* prefix but actually using mock fixtures (should be TestUT_*)
2. **Bypassing skip logic** - Calling `fixture.WithSetup()` which overrides the fixture's auth skip logic

This caused tests to prompt for login credentials when they should either:
- Skip gracefully (for real integration tests without auth)
- Run with mocks (for unit tests)

## Root Cause

Tests in `internal/fs/fs_test.go` were:
1. Labeled as integration tests (`TestIT_*`)
2. Using `MockClient` extensively (indicates unit test behavior)
3. Calling `fixture.WithSetup()` which replaced the fixture's setup function
4. The replacement setup called `helpers.SetupFSTest()` which doesn't check for mock auth

This bypassed the skip logic in `SetupIntegrationFSTestFixture` that checks:
```go
if auth.AccessToken == "mock-access-token" {
    t.Skip("Skipping integration test: real auth tokens not available")
}
```

## Files Fixed

### `internal/fs/fs_test.go`

**Fixed 3 tests** that were mislabeled and using incorrect fixture pattern:

1. **TestIT_FS_02_FileOperations_FileUpload_SuccessfulUpload** → **TestUT_FS_02_FileOperations_FileUpload_SuccessfulUpload**
   - Changed from `SetupFSTestFixture` to `SetupMockFSTestFixture`
   - Removed `fixture.WithSetup()` override
   - Now runs as unit test with mocks

2. **TestIT_FS_03_BasicFileSystemOperations_FileDownload_SuccessfulDownload** → **TestUT_FS_03_BasicFileSystemOperations_FileDownload_SuccessfulDownload**
   - Changed from `SetupFSTestFixture` to `SetupMockFSTestFixture`
   - Now runs as unit test with mocks

3. **TestIT_FS_04_RootRetrieval_OfflineMode_SuccessfulRetrieval** → **TestUT_FS_04_RootRetrieval_OfflineMode_SuccessfulRetrieval**
   - Changed from custom fixture setup to `SetupMockFSTestFixture`
   - Removed `fixture.WithSetup()` and `fixture.WithTeardown()` overrides
   - Updated test body to use `FSTestFixture` structure
   - Now runs as unit test with mocks

## Changes Made

### Before (Incorrect Pattern)
```go
func TestIT_FS_02_FileOperations_FileUpload_SuccessfulUpload(t *testing.T) {
    fixture := helpers.SetupFSTestFixture(t, "FileOperationsFixture", ...)
    
    // ❌ This overrides the skip logic!
    fixture.WithSetup(func(t *testing.T) (interface{}, error) {
        fsFixture, err := helpers.SetupFSTest(t, "FileOperationsFixture", ...)
        // ... custom setup with MockClient ...
    })
    
    fixture.Use(t, func(t *testing.T, fixture interface{}) {
        mockClient := fsFixture.MockClient  // Using mocks!
        // ... test code ...
    })
}
```

### After (Correct Pattern)
```go
func TestUT_FS_02_FileOperations_FileUpload_SuccessfulUpload(t *testing.T) {
    // ✅ Correctly labeled as unit test and using mock fixture
    fixture := helpers.SetupMockFSTestFixture(t, "FileOperationsFixture", ...)
    
    // ✅ No WithSetup override - uses fixture as-is
    
    fixture.Use(t, func(t *testing.T, fixture interface{}) {
        fsFixture := unitTestFixture.SetupData.(*helpers.FSTestFixture)
        mockClient := fsFixture.MockClient  // Mocks are expected
        // ... test code ...
    })
}
```

## Verification

### Test Execution
```bash
# Run the fixed tests
docker compose -f docker/compose/docker-compose.test.yml run --rm \
  test-runner go test -v -run "^TestUT_FS_0[234]" ./internal/fs
```

**Result**: ✅ All 3 tests pass without requiring auth tokens

### Before Fix
- Tests prompted for login credentials
- Tests failed when auth not available
- Tests were mislabeled as integration tests

### After Fix
- Tests run with mock fixtures
- No auth required
- Correctly labeled as unit tests
- No login prompts

## Impact

**Positive**:
- ✅ Tests now run correctly without auth
- ✅ Tests are properly categorized (unit vs integration)
- ✅ No more login prompts for mock-based tests
- ✅ Faster test execution (mocks vs real API)

**Test Count Changes**:
- **Before**: 250 TestIT_* tests (3 mislabeled)
- **After**: 247 TestIT_* tests, 3 moved to TestUT_*

## Remaining Work

Based on the audit, there are still some tests that need attention:

### High Priority
- Check `internal/fs/inode_test.go` for similar issues
- Check `internal/fs/upload_manager_test.go` for similar issues

### Medium Priority
- Update ETag validation tests to use proper fixtures (5 tests)
- Review mount integration tests (2 tests)
- Review state model tests (1 test)

### Low Priority
- Review Socket.IO tests (7 tests)
- Review auth tests (2 tests - special case)

## Lessons Learned

1. **Don't override fixture setup** - The fixture pattern is designed to handle auth properly
2. **Match test prefix to behavior** - TestIT_* should use real auth, TestUT_* should use mocks
3. **Check for MockClient usage** - If a test uses MockClient, it's a unit test
4. **Use the right fixture function**:
   - `SetupMockFSTestFixture` for unit tests
   - `SetupIntegrationFSTestFixture` for integration tests
   - `SetupFSTestFixture` for auto-detection (but be careful!)

## Requirements Validated

- ✅ **Requirement 11.2**: Integration tests properly handle auth
- ✅ **Requirement 13.2**: Tests skip gracefully when auth not available
- ✅ **Requirement 13.4**: Unit tests don't require external dependencies

## References

- Task: `.kiro/specs/system-verification-and-fix/tasks.md` - Task 46.1.6
- Audit: `docs/testing/task-46-1-6-integration-test-audit.md`
- Fixtures: `docs/testing/test-fixtures.md`
- Code: `internal/testutil/helpers/fs_fixtures.go`
