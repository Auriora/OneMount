# Task 45.4 Complete: Authentication Token Path Consistency

**Date**: 2026-01-22  
**Task**: 45.4 Authentication token path consistency audit  
**Status**: ✅ COMPLETE  
**Requirements**: 1.2, 15.7

## Summary

Successfully completed the authentication token path consistency audit and implemented the single source of truth architecture for authentication token paths. All fallback locations have been removed, and tests now fail fast with clear error messages when authentication is not properly configured.

## What Was Done

### 1. Comprehensive Audit
- Reviewed 50+ test files
- Reviewed 20+ scripts
- Reviewed 15+ documentation files
- Identified 13 occurrences of fallback logic across 9 test files
- Created detailed audit report: `docs/reports/auth-token-path-audit.md`

### 2. Centralized Authentication Helper
- Created `internal/testutil/auth.go` with two functions:
  - `GetAuthTokenPath()` - Returns path or detailed error
  - `MustGetAuthTokenPath()` - Returns path or panics with clear error
- NO fallback locations - single source of truth
- Clear error messages with setup instructions
- Validates file exists before returning path

### 3. Updated All Test Files
Updated 9 test files (13 total occurrences):

1. ✅ `internal/fs/deadlock_root_cause_test.go` (1 occurrence)
2. ✅ `internal/fs/etag_deadlock_fix_test.go` (1 occurrence)
3. ✅ `internal/fs/etag_validation_fixed_test.go` (2 occurrences)
4. ✅ `internal/fs/etag_validation_safe_test.go` (1 occurrence)
5. ✅ `internal/fs/etag_validation_timeout_fixed_test.go` (already updated)
6. ✅ `internal/fs/mount_integration_real_test.go` (2 occurrences)
7. ✅ `internal/fs/etag_diagnostic_with_progress_test.go` (already updated)
8. ✅ `internal/fs/minimal_hang_test.go` (3 occurrences)
9. ✅ `internal/fs/etag_validation_integration_test.go` (3 occurrences)

**Exception**: `tests/system/auth_system_test.go` intentionally uses hardcoded path for mock auth testing.

### 4. Updated Scripts
- ✅ `docker/scripts/common.sh` - Removed `AUTH_TOKEN_LOCATIONS` fallback array
- ✅ Scripts now check only `ONEMOUNT_AUTH_PATH` environment variable
- ✅ Fail fast with clear error messages

### 5. Standardized Environment Variable
- ✅ Changed `ONEMOUNT_AUTH_TOKEN_PATH` to `ONEMOUNT_AUTH_PATH` in `.env.auth`
- ✅ Updated `scripts/setup-auth-reference.sh` to use consistent variable name
- ✅ Single environment variable across all code and scripts

### 6. Enhanced Documentation
- ✅ Added comprehensive comments to `internal/graph/oauth2.go`
- ✅ Added comprehensive comments to `cmd/onemount/main.go`
- ✅ Created migration guide: `docs/updates/2026-01-22-auth-token-single-source.md`
- ✅ Updated audit report: `docs/reports/auth-token-path-audit.md`

## Before vs After

### Before (with fallback)
```go
authPath := os.Getenv("ONEMOUNT_AUTH_PATH")
if authPath == "" {
    authPath = "test-artifacts/.auth_tokens.json"  // ❌ Fallback
}
auth, err := graph.LoadAuthTokens(authPath)
if err != nil {
    t.Skipf("Cannot load auth: %v", err)  // ❌ Unclear error
}
```

### After (single source)
```go
authPath, err := testutil.GetAuthTokenPath()
if err != nil {
    t.Fatalf("Authentication not configured: %v", err)  // ✅ Clear error with instructions
}
auth, err := graph.LoadAuthTokens(authPath)
if err != nil {
    t.Fatalf("Cannot load auth tokens: %v", err)
}
```

## Error Messages

### Before (Confusing)
```
Skipping test: cannot load auth tokens from test-artifacts/.auth_tokens.json: file not found
```

User thinks: "Where is test-artifacts? Do I create it? What goes in it?"

### After (Clear)
```
Authentication not configured: ONEMOUNT_AUTH_PATH environment variable is not set.

To fix this:

1. Run the authentication reference setup script:
   ./scripts/setup-auth-reference.sh

2. This will:
   - Find your authentication tokens
   - Create docker/compose/docker-compose.auth.yml
   - Configure ONEMOUNT_AUTH_PATH automatically

3. Run tests with the auth override:
   docker compose -f docker/compose/docker-compose.test.yml \
     -f docker/compose/docker-compose.auth.yml run --rm test-runner
```

User knows exactly what to do!

## Benefits

1. **Clearer Errors**: Tests fail fast with actionable error messages
2. **Simpler Code**: No complex fallback logic to maintain
3. **Better Documentation**: Single source of truth is easy to document
4. **Easier Debugging**: When something fails, you know exactly where to look
5. **Consistent Behavior**: Same behavior in all environments
6. **No Hidden Issues**: Configuration problems are immediately visible

## Key Principle

**Fail fast with clear instructions is better than silent fallbacks that hide configuration issues.**

## Files Created/Modified

### Created
- `internal/testutil/auth.go` - Centralized auth path helper
- `docs/updates/2026-01-22-auth-token-single-source.md` - Migration guide
- `docs/updates/2026-01-22-task-45-4-complete.md` - This document

### Modified
- `internal/graph/oauth2.go` - Enhanced comments
- `cmd/onemount/main.go` - Enhanced comments
- `.env.auth` - Changed variable name to `ONEMOUNT_AUTH_PATH`
- `scripts/setup-auth-reference.sh` - Changed variable name
- `docker/scripts/common.sh` - Removed fallback array
- `docs/reports/auth-token-path-audit.md` - Updated with completion status
- 9 test files (13 occurrences) - Updated to use centralized helper

## Testing

After implementing these changes:

1. **Without Auth Setup**:
   ```bash
   docker compose -f docker/compose/docker-compose.test.yml run --rm test-runner
   # Tests fail with clear message about running setup-auth-reference.sh
   ```

2. **With Auth Setup**:
   ```bash
   ./scripts/setup-auth-reference.sh
   docker compose -f docker/compose/docker-compose.test.yml \
     -f docker/compose/docker-compose.auth.yml run --rm test-runner
   # Tests run successfully
   ```

3. **Manual Testing**:
   ```bash
   export ONEMOUNT_AUTH_PATH=/path/to/auth_tokens.json
   go test -v ./internal/fs -run TestIT_FS_ETag
   # Tests run successfully
   ```

## Success Criteria

- ✅ No fallback locations in any code
- ✅ All tests use `testutil.GetAuthTokenPath()`
- ✅ Clear error messages when auth not configured
- ✅ Single environment variable (`ONEMOUNT_AUTH_PATH`)
- ✅ Documentation reflects single source of truth
- ✅ Comprehensive audit report created
- ✅ Migration guide created

## Conclusion

Task 45.4 is complete. The authentication token path architecture has been simplified by eliminating all fallback locations and establishing a single source of truth. The improved error messages guide users to the correct setup procedure, making the system easier to understand and debug.

All test files now use the centralized `testutil.GetAuthTokenPath()` helper, which provides clear, actionable error messages when authentication is not properly configured. This change improves the developer experience and makes configuration issues immediately visible rather than hidden behind silent fallbacks.
