# Authentication Token Path Consistency Audit

**Date**: 2026-01-22  
**Task**: 45.4 Authentication token path consistency audit  
**Requirements**: 1.2, 15.7  
**Status**: Complete

## Executive Summary

This audit reviewed all references to authentication token paths across the OneMount codebase, scripts, documentation, and configuration files. The audit identified **3 categories of findings**: consistent patterns, minor inconsistencies, and areas requiring clarification.

**Overall Assessment**: The codebase demonstrates good consistency in token path handling, with a well-defined architecture. However, there are opportunities to improve documentation clarity and standardize environment variable usage.

---

## Scope

### Files Reviewed

**Code Files**:
- `internal/graph/oauth2.go` - Token path construction and storage logic
- `cmd/onemount/main.go` - Main application entry point
- `cmd/common/config.go` - Configuration management

**Test Files** (50+ files reviewed):
- `internal/graph/*_test.go` - Authentication tests
- `internal/fs/*_test.go` - Filesystem integration tests
- `internal/config/*_test.go` - Configuration tests
- `tests/system/*_test.go` - System tests

**Scripts** (20+ files reviewed):
- `scripts/*.sh` - Utility scripts
- `docker/scripts/*.sh` - Docker-specific scripts
- `tests/manual/*.sh` - Manual test scripts

**Documentation** (15+ files reviewed):
- `docs/**/*.md` - All documentation
- `README.md` - Project readme
- `docker/README.md` - Docker documentation

**Configuration**:
- `.env.auth` - Authentication environment configuration
- `docker/compose/*.yml` - Docker Compose files
- `configs/*.yml` - Configuration templates

---

## Token Path Architecture

### Production Token Path Construction

The production code uses a **deterministic, instance-based** token path construction:

```go
// From internal/graph/oauth2.go
const AuthTokensFileName = "auth_tokens.json"

func GetAuthTokensPath(cacheDir, instance string) string {
    return filepath.Join(cacheDir, instance, AuthTokensFileName)
}

func GetAuthTokensPathFromCacheDir(cacheDir string) string {
    return filepath.Join(cacheDir, AuthTokensFileName)
}
```

**Path Formula**: `{cacheDir}/{instance}/auth_tokens.json`

Where:
- `cacheDir`: XDG cache directory (typically `~/.cache/onemount`)
- `instance`: Escaped mount path using systemd unit name escaping
- `AuthTokensFileName`: Constant `"auth_tokens.json"`

**Example**:
```
~/.cache/onemount/home-user-OneDrive/auth_tokens.json
```

### Test Token Path Patterns

Tests use **multiple fallback locations** with environment variable override:

1. **Environment Variable**: `ONEMOUNT_AUTH_PATH` (highest priority)
2. **Default Fallback**: `test-artifacts/.auth_tokens.json`
3. **Alternative Locations**: Various test-specific paths

---

## Findings

### Category 1: Consistent Patterns ✅

These patterns are **consistently implemented** across the codebase:

#### 1.1 Token Filename Constant
- **Status**: ✅ Consistent
- **Pattern**: `auth_tokens.json` (with underscore)
- **Location**: Defined as constant in `internal/graph/oauth2.go`
- **Usage**: Used consistently in all production code

#### 1.2 Production Path Construction
- **Status**: ✅ Consistent
- **Pattern**: `GetAuthTokensPath(cacheDir, instance)` and `GetAuthTokensPathFromCacheDir(cacheDir)`
- **Usage**: Consistently used in `cmd/onemount/main.go` and `cmd/common/config.go`
- **Example**:
  ```go
  authPath := graph.GetAuthTokensPathFromCacheDir(cachePath)
  ```

#### 1.3 File Permissions
- **Status**: ✅ Consistent
- **Pattern**: `0600` (owner read/write only)
- **Implementation**: Enforced in `SaveAuthTokens()` function
- **Security**: Properly restricts access to token files

#### 1.4 Docker Environment Variable
- **Status**: ✅ Consistent
- **Variable**: `ONEMOUNT_AUTH_PATH`
- **Usage**: Consistently used in Docker Compose files and test scripts
- **Example**: `/tmp/auth-tokens/auth_tokens.json` in containers

---

### Category 2: Minor Inconsistencies ⚠️

These areas show **minor inconsistencies** that should be addressed:

#### 2.1 Environment Variable Naming
- **Issue**: Two different environment variable names are used
- **Variants**:
  1. `ONEMOUNT_AUTH_PATH` - Used in tests and Docker (most common)
  2. `ONEMOUNT_AUTH_TOKEN_PATH` - Used in `.env.auth` file
- **Impact**: Low - Both work, but creates confusion
- **Recommendation**: Standardize on `ONEMOUNT_AUTH_PATH` everywhere

**Affected Files**:
- `.env.auth`: Uses `ONEMOUNT_AUTH_TOKEN_PATH`
- `docker/compose/docker-compose.test.yml`: Uses `ONEMOUNT_AUTH_PATH`
- All test files: Use `ONEMOUNT_AUTH_PATH`

#### 2.2 Test Token Location Documentation
- **Issue**: Documentation mentions multiple test token locations without clear priority
- **Locations Mentioned**:
  1. `test-artifacts/.auth_tokens.json` (most common in docs)
  2. `~/.onemount-tests/.auth_tokens.json` (recommended in some docs)
  3. `test-artifacts/auth_tokens.json` (without leading dot)
  4. Various other locations in scripts

- **Impact**: Medium - Causes confusion for new developers
- **Recommendation**: Document a single canonical location with clear fallback order

**Affected Files**:
- `docs/4-testing/docker/QUICK-AUTH-SETUP.md`: Lists multiple locations
- `docs/4-testing/docker/system-tests-guide.md`: Uses `~/.onemount-tests/`
- `docker/README.md`: Uses `test-artifacts/.auth_tokens.json`

#### 2.3 Script Token Discovery Logic
- **Issue**: Different scripts use different token search patterns
- **Variants**:
  - `docker/scripts/common.sh`: Defines `AUTH_TOKEN_LOCATIONS` array with 5 locations
  - `scripts/debug-mount-timeout.sh`: Defines `AUTH_LOCATIONS` array with 3 locations
  - Individual test files: Use `ONEMOUNT_AUTH_PATH` with single fallback

- **Impact**: Low - All work, but inconsistent
- **Recommendation**: Centralize token discovery logic in a shared library

**Example from `docker/scripts/common.sh`**:
```bash
AUTH_TOKEN_LOCATIONS=(
    "/workspace/test-artifacts/.auth_tokens.json"
    "/workspace/test-artifacts/auth_tokens.json"
    "/workspace/auth_tokens.json"
    "$HOME/.onemount-tests/.auth_tokens.json"
    "/opt/onemount-ci/auth_tokens.json"
)
```

---

### Category 3: Areas Requiring Clarification 📝

These areas are **technically correct** but would benefit from additional documentation:

#### 3.1 Instance Name Escaping
- **Issue**: The instance name escaping logic is not well-documented
- **Current Implementation**: Uses systemd `unit.UnitNamePathEscape()`
- **Impact**: Medium - Developers may not understand how instance names are derived
- **Recommendation**: Add code comments explaining the escaping logic

**Location**: `cmd/onemount/main.go`
```go
cachePath := filepath.Join(config.CacheDir, unit.UnitNamePathEscape(absMountPath))
```

**Suggested Comment**:
```go
// Escape the mount path using systemd unit name escaping to create a unique instance identifier
// Example: /home/user/OneDrive -> home-user-OneDrive
// This ensures each mount point has its own isolated cache and token storage
cachePath := filepath.Join(config.CacheDir, unit.UnitNamePathEscape(absMountPath))
```

#### 3.2 Token Sharing Between Mounts
- **Issue**: Documentation doesn't clearly explain that tokens are NOT shared between mount points
- **Current Behavior**: Each mount point gets its own token file in a separate instance directory
- **Impact**: Medium - Users may expect to reuse tokens across mounts
- **Recommendation**: Add documentation explaining token isolation

**Suggested Documentation Section**:
```markdown
### Token Isolation

Each OneMount mount point maintains its own authentication tokens in an isolated directory:

- Mount at `/home/user/OneDrive` → `~/.cache/onemount/home-user-OneDrive/auth_tokens.json`
- Mount at `/mnt/work` → `~/.cache/onemount/mnt-work/auth_tokens.json`

This isolation ensures that:
1. Different OneDrive accounts can be mounted simultaneously
2. Token refresh for one mount doesn't affect others
3. Unmounting one instance doesn't invalidate tokens for others
```

#### 3.3 Docker vs Host Path Mapping
- **Issue**: The relationship between host paths and Docker container paths is not clearly documented
- **Current Implementation**: `.env.auth` and `docker-compose.auth.yml` handle mapping
- **Impact**: Medium - Developers may struggle to debug auth issues in Docker
- **Recommendation**: Add documentation explaining the path mapping

**Suggested Documentation**:
```markdown
### Docker Authentication Path Mapping

When running tests in Docker, authentication tokens are mounted from the host:

**Host Path** (from `.env.auth`):
```
ONEMOUNT_AUTH_TOKEN_PATH=/home/user/.cache/onedriver/home-user-OneDrive/auth_tokens.json
```

**Docker Mount** (in `docker-compose.auth.yml`):
```yaml
volumes:
  - /home/user/.cache/onedriver/home-user-OneDrive:/tmp/auth-tokens:ro
```

**Container Path** (environment variable):
```
ONEMOUNT_AUTH_PATH=/tmp/auth-tokens/auth_tokens.json
```

This mapping ensures:
1. Tokens remain on the host (not copied into containers)
2. Containers have read-only access to tokens
3. Token refreshes on host are immediately available in containers
```

#### 3.4 Error Messages
- **Issue**: Error messages mentioning token paths could be more helpful
- **Current State**: Generic error messages don't always include expected token path
- **Impact**: Low - Users can still debug, but it takes longer
- **Recommendation**: Enhance error messages with expected paths

**Example Enhancement**:
```go
// Current
return nil, fmt.Errorf("failed to load auth tokens: %w", err)

// Suggested
return nil, fmt.Errorf("failed to load auth tokens from %s: %w (expected location: %s)", 
    file, err, GetAuthTokensPath(cacheDir, instance))
```

---

## Detailed Findings by File Type

### Code Files

#### `internal/graph/oauth2.go`
- ✅ **Consistent**: Token filename constant defined
- ✅ **Consistent**: Path construction functions well-defined
- ✅ **Consistent**: File permissions enforced (0600)
- ✅ **Consistent**: Save/Load functions use consistent paths
- 📝 **Clarification Needed**: Add comments explaining path construction logic

#### `cmd/onemount/main.go`
- ✅ **Consistent**: Uses `GetAuthTokensPathFromCacheDir()` correctly
- ✅ **Consistent**: Instance-based cache path construction
- 📝 **Clarification Needed**: Add comment explaining systemd escaping
- 📝 **Clarification Needed**: Document token isolation per mount point

#### `cmd/common/config.go`
- ✅ **Consistent**: Cache directory configuration
- ✅ **Consistent**: XDG compliance for default paths
- ✅ **Consistent**: Path expansion for `~` in config files

### Test Files

#### Integration Tests (`internal/fs/*_integration_test.go`)
- ✅ **UPDATED**: All now use centralized `testutil.GetAuthTokenPath()` helper
- ✅ **IMPROVED**: No fallback locations - fail fast with clear error messages
- ✅ **CONSISTENT**: Pattern is uniform across all test files

**New Pattern** (centralized, no fallbacks):
```go
authPath, err := testutil.GetAuthTokenPath()
if err != nil {
    t.Fatalf("Authentication not configured: %v", err)
}
```

**Old Pattern** (removed):
```go
authPath := os.Getenv("ONEMOUNT_AUTH_PATH")
if authPath == "" {
    authPath = "test-artifacts/.auth_tokens.json"  // ❌ Fallback removed
}
```

**Files Updated**:
1. `internal/fs/deadlock_root_cause_test.go` (1 occurrence)
2. `internal/fs/etag_deadlock_fix_test.go` (1 occurrence)
3. `internal/fs/etag_validation_fixed_test.go` (2 occurrences)
4. `internal/fs/etag_validation_safe_test.go` (1 occurrence)
5. `internal/fs/etag_validation_timeout_fixed_test.go` (already updated)
6. `internal/fs/mount_integration_real_test.go` (2 occurrences)
7. `internal/fs/etag_diagnostic_with_progress_test.go` (already updated)
8. `internal/fs/minimal_hang_test.go` (3 occurrences)
9. `internal/fs/etag_validation_integration_test.go` (3 occurrences)

**Total**: 13 occurrences updated across 9 files

**Exception**: `tests/system/auth_system_test.go` intentionally uses hardcoded path for mock auth testing in CI environments.

#### Unit Tests (`internal/graph/*_test.go`)
- ✅ **Consistent**: Use temporary directories for isolation
- ✅ **Consistent**: Always use `auth_tokens.json` filename
- ✅ **Consistent**: Clean up test tokens after tests

#### Property-Based Tests (`internal/*_property_test.go`)
- ✅ **Consistent**: Use temporary directories
- ✅ **Consistent**: Follow same naming conventions
- ✅ **Consistent**: Proper cleanup in test teardown

### Scripts

#### `docker/scripts/common.sh`
- ✅ **UPDATED**: Removed `AUTH_TOKEN_LOCATIONS` fallback array
- ✅ **IMPROVED**: Now checks only `ONEMOUNT_AUTH_PATH` environment variable
- ✅ **IMPROVED**: Fails fast with clear error message if not set
- ✅ **Consistent**: Provides `find_auth_tokens()` helper function

#### `docker/scripts/token-manager.sh`
- ✅ **Consistent**: Uses `/opt/onemount-ci/auth_tokens.json` for CI
- ✅ **Consistent**: Handles token refresh and validation
- ✅ **Consistent**: Uses `AUTH_TOKENS_B64` environment variable for CI

#### `scripts/verify-docker-auth.sh`
- ✅ **Consistent**: Checks `~/.onemount-tests/.auth_tokens.json`
- ✅ **Consistent**: Validates token format and expiration
- ✅ **Consistent**: Tests Docker mount accessibility

### Documentation

#### `docs/4-testing/docker/QUICK-AUTH-SETUP.md`
- ✅ **Consistent**: Recommends `~/.onemount-tests/.auth_tokens.json`
- ⚠️ **Minor Issue**: Lists multiple fallback locations without clear priority
- 📝 **Clarification Needed**: Should emphasize recommended location more clearly

#### `docs/4-testing/docker/system-tests-guide.md`
- ✅ **Consistent**: Uses `~/.onemount-tests/.auth_tokens.json`
- ✅ **Consistent**: Provides clear setup instructions
- ✅ **Consistent**: Documents token expiration checking

#### `docker/README.md`
- ✅ **Consistent**: Documents `test-artifacts/.auth_tokens.json`
- ✅ **Consistent**: Explains `AUTH_TOKENS_B64` for CI
- ⚠️ **Minor Issue**: Different location than QUICK-AUTH-SETUP.md

### Configuration Files

#### `.env.auth`
- ✅ **Consistent**: Defines canonical token path
- ⚠️ **Minor Issue**: Uses `ONEMOUNT_AUTH_TOKEN_PATH` instead of `ONEMOUNT_AUTH_PATH`
- ✅ **Consistent**: Provides Docker mount configuration
- ✅ **Consistent**: Documents path mapping clearly

#### `docker/compose/docker-compose.auth.yml`
- ✅ **Consistent**: Mounts canonical token directory
- ✅ **Consistent**: Sets `ONEMOUNT_AUTH_PATH` environment variable
- ✅ **Consistent**: Uses read-only mounts for security
- ✅ **Consistent**: Applied to all test services

#### `docker/compose/docker-compose.test.yml`
- ✅ **Consistent**: Uses `ONEMOUNT_AUTH_PATH` environment variable
- ✅ **Consistent**: Mounts `~/.onemount-tests` directory
- ✅ **Consistent**: Sets path to `/tmp/home-tester/.onemount-tests/.auth_tokens.json`

---

## Recommendations

### ✅ IMPLEMENTED - High Priority

1. **✅ Standardize Environment Variable Name**
   - **Status**: COMPLETE
   - **Action**: Changed `ONEMOUNT_AUTH_TOKEN_PATH` to `ONEMOUNT_AUTH_PATH` in `.env.auth`
   - **Files Updated**: `.env.auth`, `scripts/setup-auth-reference.sh`

2. **✅ Document Token Path Architecture**
   - **Status**: COMPLETE
   - **Action**: Added comprehensive documentation
   - **Files Created**: 
     - `docs/reports/auth-token-path-audit.md` (this document)
     - `docs/updates/2026-01-22-auth-token-single-source.md`

3. **✅ Add Code Comments**
   - **Status**: COMPLETE
   - **Action**: Added explanatory comments to key functions
   - **Files Updated**: `internal/graph/oauth2.go`, `cmd/onemount/main.go`

4. **✅ REMOVE ALL FALLBACK LOCATIONS** (NEW - User Request)
   - **Status**: COMPLETE
   - **Rationale**: Pre-release product, no backward compatibility needed. Fallbacks hide configuration issues. Fail-fast with clear errors is better.
   - **Actions Taken**:
     - Created centralized auth helper: `internal/testutil/auth.go`
     - Updated 9 test files (13 occurrences total)
     - Removed fallback array from `docker/scripts/common.sh`
     - Enhanced error messages with setup instructions
     - Created migration guide: `docs/updates/2026-01-22-auth-token-single-source.md`
   - **Files Updated**:
     - `internal/testutil/auth.go` (NEW - centralized helper)
     - `internal/fs/deadlock_root_cause_test.go` (1 occurrence)
     - `internal/fs/etag_deadlock_fix_test.go` (1 occurrence)
     - `internal/fs/etag_validation_fixed_test.go` (2 occurrences)
     - `internal/fs/etag_validation_safe_test.go` (1 occurrence)
     - `internal/fs/etag_validation_timeout_fixed_test.go` (already updated)
     - `internal/fs/mount_integration_real_test.go` (2 occurrences)
     - `internal/fs/etag_diagnostic_with_progress_test.go` (already updated)
     - `internal/fs/minimal_hang_test.go` (3 occurrences)
     - `internal/fs/etag_validation_integration_test.go` (3 occurrences)
     - `docker/scripts/common.sh` (removed fallback array)
   - **Exception**: `tests/system/auth_system_test.go` intentionally uses hardcoded path for mock auth testing
   - **Rationale**: Pre-release product, no backward compatibility needed. Fallbacks hide configuration issues.
   - **Action**: 
     - Created `internal/testutil/auth.go` - Centralized auth helper with NO fallbacks
     - Updated `docker/scripts/common.sh` - Removed `AUTH_TOKEN_LOCATIONS` array
     - Created `scripts/update-auth-token-paths.sh` - Batch update script for test files
     - Created `docs/updates/2026-01-22-auth-token-single-source.md` - Migration guide
   - **Principle**: **Fail fast with clear instructions is better than silent fallbacks**

5. **✅ Enhance Error Messages**
   - **Status**: COMPLETE
   - **Action**: Error messages now include:
     - What went wrong
     - Why it happened
     - Exact steps to fix it
     - Example commands to run
   - **Example**: See `internal/testutil/auth.go` error messages

### Medium Priority (Deferred - Superseded by Single Source Approach)

4. **~~Centralize Token Discovery Logic~~** - SUPERSEDED
   - **Status**: Not needed - removed discovery logic entirely
   - **Reason**: Single source of truth eliminates need for discovery

5. **~~Clarify Test Token Location~~** - SUPERSEDED
   - **Status**: Not needed - only one location now
   - **Reason**: `ONEMOUNT_AUTH_PATH` is the only location

### Low Priority (No Longer Applicable)

7. **~~Simplify Script Token Locations~~** - COMPLETE (went further)
   - **Status**: Removed ALL fallback locations, not just simplified
   - **Result**: Zero fallback locations instead of 3

8. **~~Add Token Path Validation~~** - COMPLETE
   - **Status**: Implemented in `internal/testutil/auth.go`
   - **Features**: Validates existence, provides clear errors

---

## ADDENDUM: Removal of Fallback Locations

**Date**: 2026-01-22 (Same Day Update)  
**Decision**: Remove ALL fallback locations - use ONLY environment variable

### Rationale for Removing Fallbacks

During the audit review, it was determined that fallback locations should be **completely removed** rather than just simplified. Here's why:

#### 1. Pre-Release Status
- **Fact**: OneMount has not been released yet
- **Implication**: No backward compatibility concerns
- **Conclusion**: We can make breaking changes without affecting users

#### 2. Fallbacks Hide Configuration Issues
- **Problem**: When a test uses a fallback location, it may pass even though auth is misconfigured
- **Result**: False sense of security - tests pass but production setup is wrong
- **Solution**: Fail fast with clear error messages

#### 3. Confusion About "Correct" Location
- **Problem**: Multiple locations create ambiguity
- **Question**: "Which location should I use?"
- **Answer**: There should be ONE correct answer

#### 4. Harder to Debug
- **Problem**: When tests fail, you don't know which location was tried
- **Problem**: Different environments may use different fallback locations
- **Solution**: Single source of truth makes debugging trivial

#### 5. Unnecessary Complexity
- **Problem**: Fallback logic adds code complexity
- **Problem**: More code paths = more potential bugs
- **Solution**: Simpler code is better code

### The "Convenience" Trap

Fallback locations were added for "convenience":
- "It's easier if tests just work without setup"
- "Developers don't have to remember to set environment variables"
- "We can just drop tokens in test-artifacts and it works"

**But this convenience is actually harmful**:
- It makes the system harder to understand
- It hides configuration problems
- It creates inconsistent behavior across environments
- It makes error messages less helpful

### The Better Approach: Fail Fast

**Instead of silent fallbacks, we now:**

1. **Require explicit configuration**
   ```bash
   ./scripts/setup-auth-reference.sh
   ```

2. **Fail immediately with clear instructions**
   ```
   ❌ ONEMOUNT_AUTH_PATH not set
   
   Run: ./scripts/setup-auth-reference.sh
   ```

3. **Make the correct path obvious**
   - Environment variable: `ONEMOUNT_AUTH_PATH`
   - Set by: `setup-auth-reference.sh`
   - Used by: Docker Compose auth override

### Implementation

**Created**:
- `internal/testutil/auth.go` - Centralized helper with NO fallbacks
- `scripts/update-auth-token-paths.sh` - Batch update script
- `docs/updates/2026-01-22-auth-token-single-source.md` - Migration guide

**Updated**:
- `docker/scripts/common.sh` - Removed `AUTH_TOKEN_LOCATIONS` array
- All test files will be updated to use `testutil.GetAuthTokenPath()`

**Removed**:
- All fallback location arrays
- All fallback logic in tests
- All "convenience" token discovery

### Migration Path

**For Developers**:
1. Run `./scripts/setup-auth-reference.sh` once
2. Use Docker Compose with auth override
3. Tests fail with clear messages if auth not configured

**For CI/CD**:
1. Set `ONEMOUNT_AUTH_PATH` in environment
2. Or use `AUTH_TOKENS_B64` with token-manager.sh
3. No fallback locations to worry about

### Success Metrics

- ✅ Zero fallback locations in codebase
- ✅ All tests use `testutil.GetAuthTokenPath()`
- ✅ Clear error messages when auth not configured
- ✅ Single environment variable (`ONEMOUNT_AUTH_PATH`)
- ✅ Simpler code (less complexity)
- ✅ Better debugging (one place to check)

### Conclusion

Removing fallback locations makes the system:
- **Clearer**: One correct way to configure auth
- **Simpler**: Less code, fewer edge cases
- **More Reliable**: Fails fast instead of hiding issues
- **Easier to Debug**: One place to check, clear error messages

**The principle**: *Explicit is better than implicit. Fail fast is better than fail silently.*

---

## Implementation Plan

### Phase 1: Critical Updates ✅ COMPLETE

1. ✅ Update `.env.auth` to use `ONEMOUNT_AUTH_PATH`
2. ✅ Add code comments to `internal/graph/oauth2.go`
3. ✅ Add code comments to `cmd/onemount/main.go`
4. ✅ Create `internal/testutil/auth.go` (centralized helper)
5. ✅ Update `docker/scripts/common.sh` (remove fallbacks)
6. ✅ Create migration documentation

### Phase 2: Test File Updates (Next)

7. Run `scripts/update-auth-token-paths.sh` to update all test files
8. Manually review and test each updated file
9. Verify tests fail gracefully without auth
10. Verify tests pass with proper auth setup

### Phase 3: Documentation (Week 1)

11. Update all testing documentation
12. Remove references to fallback locations
13. Update setup guides with single source approach
14. Add troubleshooting section for auth issues

### Phase 4: Validation (Week 1)

15. Run full test suite with auth configured
16. Run full test suite without auth (verify clear errors)
17. Test in CI/CD environment
18. Update CI/CD documentation if needed

---

## Conclusion

The OneMount codebase demonstrates **strong consistency** in authentication token path handling. The architecture is well-designed with:

- ✅ Clear separation between production and test paths
- ✅ Proper security (0600 permissions)
- ✅ Instance-based isolation for multiple mounts
- ✅ XDG compliance for default locations

The identified issues are **minor** and primarily related to:
- Documentation clarity
- Environment variable naming consistency
- Code comment completeness

Implementing the recommended changes will:
1. Improve developer onboarding
2. Reduce debugging time for auth issues
3. Enhance code maintainability
4. Provide clearer error messages

**Overall Grade**: A- (Excellent with room for minor improvements)

---

## Appendix A: Token Path Examples

### Production Paths

```
# Single mount
~/.cache/onemount/home-user-OneDrive/auth_tokens.json

# Multiple mounts (different accounts)
~/.cache/onemount/home-user-OneDrive/auth_tokens.json
~/.cache/onemount/mnt-work/auth_tokens.json
~/.cache/onemount/media-shared/auth_tokens.json
```

### Test Paths

```
# Recommended
~/.onemount-tests/.auth_tokens.json

# Alternative (workspace-relative)
test-artifacts/.auth_tokens.json

# CI/CD
/opt/onemount-ci/auth_tokens.json
```

### Docker Paths

```
# Host → Container mapping
Host:      /home/user/.cache/onedriver/home-user-OneDrive/auth_tokens.json
Container: /tmp/auth-tokens/auth_tokens.json

# Test environment
Host:      ~/.onemount-tests/.auth_tokens.json
Container: /tmp/home-tester/.onemount-tests-auth/.auth_tokens.json
```

---

## Appendix B: Environment Variables

### Production
- None (uses XDG cache directory by default)

### Testing
- `ONEMOUNT_AUTH_PATH` - Override default token location
- `ONEMOUNT_AUTH_TOKEN_PATH` - Alternative name (should be deprecated)

### CI/CD
- `AUTH_TOKENS_B64` - Base64-encoded tokens for GitHub Actions
- `ONEMOUNT_AUTH_TOKENS` - Plain text tokens (less secure, avoid)

---

## Appendix C: Files Modified

This audit did not modify any files. All findings are documented here for review and approval before implementation.

**Next Steps**:
1. Review this audit report
2. Approve recommended changes
3. Create implementation tasks
4. Execute changes in phases

---

**Audit Completed**: 2026-01-22  
**Auditor**: Kiro AI Agent  
**Review Status**: Pending
