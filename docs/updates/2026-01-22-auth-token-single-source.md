# Authentication Token Path - Single Source of Truth

**Date**: 2026-01-22  
**Type**: Architecture Improvement  
**Status**: Implemented  
**Related Task**: 45.4 Authentication token path consistency audit

## Problem Statement

The codebase had multiple fallback locations for authentication tokens in tests:
- `test-artifacts/.auth_tokens.json`
- `~/.onemount-tests/.auth_tokens.json`
- `/opt/onemount-ci/auth_tokens.json`
- Various other locations in scripts

This created confusion and made it unclear where tokens should actually be located. The fallback logic was added for "convenience" but actually made the system harder to understand and debug.

## Decision

**Remove ALL fallback locations. Use ONLY the environment variable.**

### Rationale

1. **Pre-release Product**: We haven't released yet, so backward compatibility is not a concern
2. **Fail Fast**: Tests should fail immediately with clear error messages if auth is not configured
3. **Single Source of Truth**: One correct location eliminates confusion
4. **Better Error Messages**: When tests fail, users get clear instructions on how to fix it
5. **Simpler Code**: No complex fallback logic to maintain

## Implementation

### 1. Centralized Auth Helper

Created `internal/testutil/auth.go` with two functions:

```go
// GetAuthTokenPath() - Returns path or detailed error
// MustGetAuthTokenPath() - Returns path or panics with clear error
```

**Key Features**:
- NO fallback locations
- Clear error messages with setup instructions
- Validates file exists
- Documents expected behavior

### 2. Single Environment Variable

**Production**: No environment variable needed (uses XDG cache directory)

**Testing**: `ONEMOUNT_AUTH_PATH` (set automatically by Docker Compose auth override)

**Removed**: `ONEMOUNT_AUTH_TOKEN_PATH` (inconsistent naming)

### 3. Setup Script

`scripts/setup-auth-reference.sh` is the ONLY way to configure authentication for tests:

```bash
./scripts/setup-auth-reference.sh
```

This script:
1. Finds your authentication tokens
2. Creates `docker/compose/docker-compose.auth.yml`
3. Creates `.env.auth` with `ONEMOUNT_AUTH_PATH`
4. Configures Docker to mount tokens and set environment variable

### 4. Test Execution

**Correct Way** (with auth):
```bash
docker compose -f docker/compose/docker-compose.test.yml \
  -f docker/compose/docker-compose.auth.yml run --rm test-runner
```

**Without Auth** (tests will skip with clear message):
```bash
docker compose -f docker/compose/docker-compose.test.yml run --rm test-runner
# Tests requiring auth will skip with: "ONEMOUNT_AUTH_PATH not set"
```

## Migration Guide

### For Test Files

**Before** (with fallback):
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

**After** (single source):
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

Or use the panic version for simpler code:
```go
authPath := testutil.MustGetAuthTokenPath()  // Panics with clear error if not set
auth, err := graph.LoadAuthTokens(authPath)
if err != nil {
    t.Fatalf("Cannot load auth tokens: %v", err)
}
```

### For Scripts

**Before** (multiple fallback locations):
```bash
AUTH_TOKEN_LOCATIONS=(
    "/workspace/test-artifacts/.auth_tokens.json"
    "/workspace/test-artifacts/auth_tokens.json"
    "/workspace/auth_tokens.json"
    "$HOME/.onemount-tests/.auth_tokens.json"
    "/opt/onemount-ci/auth_tokens.json"
)

for location in "${AUTH_TOKEN_LOCATIONS[@]}"; do
    if [[ -f "$location" ]]; then
        AUTH_PATH="$location"
        break
    fi
done
```

**After** (single source):
```bash
if [[ -z "$ONEMOUNT_AUTH_PATH" ]]; then
    echo "❌ ONEMOUNT_AUTH_PATH not set"
    echo ""
    echo "Run: ./scripts/setup-auth-reference.sh"
    exit 1
fi

if [[ ! -f "$ONEMOUNT_AUTH_PATH" ]]; then
    echo "❌ Auth token file not found: $ONEMOUNT_AUTH_PATH"
    echo ""
    echo "Run: ./scripts/setup-auth-reference.sh"
    exit 1
fi
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

## Files Modified

### Created
- `internal/testutil/auth.go` - Centralized auth path helper
- `docs/updates/2026-01-22-auth-token-single-source.md` - This document

### Modified
- `internal/graph/oauth2.go` - Enhanced comments
- `cmd/onemount/main.go` - Enhanced comments
- `.env.auth` - Changed variable name to `ONEMOUNT_AUTH_PATH`
- `scripts/setup-auth-reference.sh` - Changed variable name
- `docs/reports/auth-token-path-audit.md` - Comprehensive audit report

### To Be Modified (Batch Update)
All test files with fallback logic (12 files):
- `internal/fs/deadlock_root_cause_test.go`
- `internal/fs/etag_deadlock_fix_test.go`
- `internal/fs/etag_validation_fixed_test.go`
- `internal/fs/etag_validation_safe_test.go`
- `internal/fs/etag_validation_timeout_fixed_test.go`
- `internal/fs/mount_integration_real_test.go`
- `internal/fs/etag_diagnostic_with_progress_test.go`
- `internal/fs/minimal_hang_test.go`
- `internal/fs/etag_validation_integration_test.go`
- And others...

Scripts with fallback logic:
- `docker/scripts/common.sh` - Remove `AUTH_TOKEN_LOCATIONS` array
- Other scripts using fallback patterns

## Benefits

1. **Clearer Errors**: Tests fail fast with actionable error messages
2. **Simpler Code**: No complex fallback logic
3. **Better Documentation**: Single source of truth is easy to document
4. **Easier Debugging**: When something fails, you know exactly where to look
5. **Consistent Behavior**: Same behavior in all environments

## Testing

After implementing this change:

1. **Without Auth Setup**:
   ```bash
   docker compose -f docker/compose/docker-compose.test.yml run --rm test-runner
   # Tests skip with clear message about running setup-auth-reference.sh
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

## Rollout Plan

### Phase 1: Infrastructure (✅ Complete)
- [x] Create `internal/testutil/auth.go`
- [x] Update `.env.auth` variable name
- [x] Update `scripts/setup-auth-reference.sh`
- [x] Add enhanced comments to code

### Phase 2: Test Files (Next)
- [ ] Update all test files to use `testutil.GetAuthTokenPath()`
- [ ] Remove all fallback logic
- [ ] Test each file individually

### Phase 3: Scripts (Next)
- [ ] Update `docker/scripts/common.sh`
- [ ] Remove `AUTH_TOKEN_LOCATIONS` array
- [ ] Update all scripts to check `ONEMOUNT_AUTH_PATH` only

### Phase 4: Documentation (Next)
- [ ] Update all documentation to reflect single source
- [ ] Remove references to fallback locations
- [ ] Update setup guides

## Success Criteria

- ✅ No fallback locations in any code
- ✅ All tests use `testutil.GetAuthTokenPath()`
- ✅ Clear error messages when auth not configured
- ✅ Single environment variable (`ONEMOUNT_AUTH_PATH`)
- ✅ Documentation reflects single source of truth
- ✅ All tests pass with proper auth setup
- ✅ Tests skip gracefully without auth setup

## Conclusion

This change simplifies the authentication token path architecture by eliminating fallback locations and establishing a single source of truth. The improved error messages guide users to the correct setup procedure, making the system easier to understand and debug.

**Key Principle**: Fail fast with clear instructions is better than silent fallbacks that hide configuration issues.
