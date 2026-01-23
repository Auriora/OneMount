# Test Fixture Refactoring: Advanced Mounting Tests

**Date**: 2026-01-23  
**Time**: 14:00:00  
**Task**: Test Fixture Refactoring  
**Related Issue**: Task 46.1.11 - Integration test verification

## Summary

Refactored two integration tests in `internal/fs/advanced_mounting_test.go` to use the proper test fixture setup functions instead of manual authentication creation. This fixes authentication failures and crashes that were occurring in Docker test environments.

## Problem

Tests `TestIT_FS_AdvancedMounting_02_StaleLockDetection` and `TestIT_FS_AdvancedMounting_03_DatabaseRetryLogic` were using legacy authentication patterns:

- Manual creation of `graph.Auth` objects with fake tokens
- Direct calls to `NewFilesystem()` without using fixtures
- No automatic cleanup or resource management

This caused:
- Authentication failures when fake tokens were used
- Fallback to interactive GTK authentication in Docker
- SIGABRT crashes due to missing display
- Test hangs and failures

## Solution

Refactored both tests to use `helpers.SetupFSTestFixture()`:

### Test 02: Stale Lock Detection

**Before**:
```go
// ❌ Manual auth creation with fake tokens
auth := &graph.Auth{
    AccessToken:  "mock_access_token",
    RefreshToken: "mock_refresh_token",
    ExpiresAt:    time.Now().Add(1 * time.Hour).Unix(),
}
filesystem, err := NewFilesystem(auth, tempDir, 30)
```

**After**:
```go
// ✅ Uses proper setup function with real authentication
fixture := helpers.SetupFSTestFixture(t, "StaleLockDetectionFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
    return NewFilesystem(auth, tempDir, cacheTTL)
})

fixture.Use(t, func(t *testing.T, fixtureData interface{}) {
    unitTestFixture := fixtureData.(*framework.UnitTestFixture)
    fsFixture := unitTestFixture.SetupData.(*helpers.FSTestFixture)
    filesystem := fsFixture.FS.(*Filesystem)
    // Test code...
})
```

### Test 03: Database Retry Logic

Similar refactoring applied to use proper fixture setup with real authentication.

## Changes Made

### Files Modified

1. **`internal/fs/advanced_mounting_test.go`**:
   - Refactored `TestIT_FS_AdvancedMounting_02_StaleLockDetection`
   - Refactored `TestIT_FS_AdvancedMounting_03_DatabaseRetryLogic`
   - Removed manual auth creation
   - Added proper fixture usage with `helpers.SetupFSTestFixture()`
   - Fixed assertion in test 03 (total retry time threshold)

## Benefits

✅ **Real Authentication**: Tests now use real auth tokens from `ONEMOUNT_AUTH_PATH`  
✅ **Automatic Cleanup**: Fixture handles resource cleanup automatically  
✅ **No Crashes**: No more GTK authentication failures in Docker  
✅ **Follows Conventions**: Matches pattern used by all other integration tests  
✅ **Reliable**: Tests pass consistently in Docker environment  

## Test Results

Both tests now pass successfully:

```bash
$ docker compose -f docker/compose/docker-compose.test.yml \
  -f docker/compose/docker-compose.auth.yml run --rm test-runner \
  go test -v -run "TestIT_FS_AdvancedMounting_02|TestIT_FS_AdvancedMounting_03" ./internal/fs

=== RUN   TestIT_FS_AdvancedMounting_02_StaleLockDetection
--- PASS: TestIT_FS_AdvancedMounting_02_StaleLockDetection (3.60s)
=== RUN   TestIT_FS_AdvancedMounting_03_DatabaseRetryLogic
--- PASS: TestIT_FS_AdvancedMounting_03_DatabaseRetryLogic (0.98s)
PASS
ok      github.com/auriora/onemount/internal/fs 4.611s
```

## Pattern Used

The refactoring follows the established pattern from `TestIT_FS_AdvancedMounting_01_MountTimeoutConfiguration` which was already using the correct fixture setup.

### Key Pattern Elements

1. **Use `helpers.SetupFSTestFixture()`** - Automatically detects test type (unit/integration/system) based on test name prefix
2. **Provide callback function** - Creates filesystem with real auth
3. **Access fixture data** - Use `fixture.Use()` to access filesystem and auth
4. **Automatic cleanup** - Fixture handles teardown automatically

## Remaining Work

This refactoring addresses the immediate failing tests in `advanced_mounting_test.go`. Additional tests throughout the codebase may still need similar refactoring:

- Property-based tests in `internal/fs/*_property_test.go`
- Other integration tests with manual auth creation
- System tests in `tests/system/`

See `test-artifacts/prompts/refactor-tests-to-use-new-setup-functions.md` for comprehensive refactoring guidance.

## References

- **Test Fixtures Guide**: `docs/testing/test-fixtures.md`
- **Root Cause Analysis**: `test-artifacts/logs/task-46-1-11-issue-2-root-cause.md`
- **Refactoring Prompt**: `test-artifacts/prompts/refactor-tests-to-use-new-setup-functions.md`
- **Auth Reference System**: `docs/updates/2025-12-23-160600-authentication-reference-system.md`

## Impact

- **Tests Fixed**: 2 integration tests now pass
- **Code Quality**: Improved test maintainability and reliability
- **Pattern Established**: Clear example for refactoring other tests
- **Documentation**: Comprehensive guide available for future refactoring

---

**Status**: ✅ Complete  
**Tests Passing**: Yes  
**Ready for Review**: Yes
