# Auth Token Storage Architecture Analysis

**Date**: 2026-01-23  
**Type**: Architecture Analysis  
**Status**: 🔍 Investigation Required  
**Priority**: HIGH

## Problem Statement

Auth tokens are currently stored based on **mount point location** rather than **account identity**, causing several issues:

1. **Unreliable token availability in Docker** - Different mount points in tests vs production cause different token paths
2. **Token duplication** - Same account mounted at different locations creates duplicate token files
3. **Token loss on remount** - Changing mount point location loses access to existing tokens
4. **Test environment confusion** - Tests can't reliably find tokens because mount points vary

## Current Architecture

### Token Path Formula
```
{cacheDir}/{instance}/auth_tokens.json
```

Where:
- `cacheDir`: XDG cache directory (typically `~/.cache/onemount`)
- `instance`: **Escaped mount path** using systemd unit name escaping
- Example: Mount at `/home/user/OneDrive` → `~/.cache/onemount/home-user-OneDrive/auth_tokens.json`

### Code Location
- **Implementation**: `internal/graph/oauth2.go:47-50`
- **Function**: `GetAuthTokensPath(cacheDir, instance string)`

### Current Behavior
```go
func GetAuthTokensPath(cacheDir, instance string) string {
    return filepath.Join(cacheDir, instance, AuthTokensFileName)
}
```

**Problem**: `instance` is derived from mount point, not account!

## Issues Identified

### 1. Mount Point Dependency
**Current**: Token path depends on where you mount  
**Problem**: Same account, different mount points = different token files

**Example**:
```bash
# Production mount
onemount /home/user/OneDrive
# Tokens: ~/.cache/onemount/home-user-OneDrive/auth_tokens.json

# Test mount (different location)
onemount /tmp/test-mount
# Tokens: ~/.cache/onemount/tmp-test-mount/auth_tokens.json

# Result: Can't find production tokens in test environment!
```

### 2. Token Duplication
**Scenario**: User mounts same account at multiple locations

```bash
# Mount 1
onemount /home/user/OneDrive
# Tokens: ~/.cache/onemount/home-user-OneDrive/auth_tokens.json

# Mount 2 (same account, different location)
onemount /mnt/onedrive
# Tokens: ~/.cache/onemount/mnt-onedrive/auth_tokens.json

# Result: Two copies of same tokens, can get out of sync!
```

### 3. Docker Environment Issues
**Test Environment**:
```bash
# Docker test mount point
/tmp/onemount-integration-test-xyz123/

# Token path in Docker
~/.cache/onemount/tmp-onemount-integration-test-xyz123/auth_tokens.json
```

**Production Environment**:
```bash
# Production mount point
/home/user/OneDrive/

# Token path in production
~/.cache/onemount/home-user-OneDrive/auth_tokens.json
```

**Result**: Tests can't find production tokens because paths don't match!

### 4. Token Loss on Remount
**Scenario**: User changes mount point location

```bash
# Original mount
onemount /home/user/OneDrive
# Tokens stored: ~/.cache/onemount/home-user-OneDrive/auth_tokens.json

# User decides to change mount location
onemount /mnt/onedrive
# Looks for tokens: ~/.cache/onemount/mnt-onedrive/auth_tokens.json
# Result: Tokens not found, must re-authenticate!
```

## Proposed Solution

### Account-Based Token Storage

**New Formula**:
```
{cacheDir}/accounts/{account-hash}/auth_tokens.json
```

Where:
- `cacheDir`: XDG cache directory (unchanged)
- `account-hash`: SHA256 hash of account email (first 16 chars)
- Example: `user@example.com` → `~/.cache/onemount/accounts/a1b2c3d4e5f6g7h8/auth_tokens.json`

### Benefits

1. **Mount Point Independence** ✅
   - Same tokens regardless of where you mount
   - Tests and production use same token location
   - Remounting doesn't lose tokens

2. **No Token Duplication** ✅
   - One account = one token file
   - Token refresh updates single location
   - No sync issues

3. **Reliable Docker Testing** ✅
   - Tests find tokens regardless of mount point
   - Consistent token location across environments
   - No path confusion

4. **Account Isolation** ✅
   - Different accounts have separate token files
   - Multiple accounts can be mounted simultaneously
   - Clear separation of credentials

### Implementation Strategy

#### Phase 1: Add Account-Based Storage (New Code)
```go
// GetAuthTokensPathByAccount returns token path based on account identity
func GetAuthTokensPathByAccount(cacheDir, accountEmail string) string {
    accountHash := hashAccount(accountEmail)
    return filepath.Join(cacheDir, "accounts", accountHash, AuthTokensFileName)
}

// hashAccount creates a stable hash of account email
func hashAccount(email string) string {
    hash := sha256.Sum256([]byte(strings.ToLower(email)))
    return hex.EncodeToString(hash[:])[:16]
}
```

#### Phase 2: Migration Path (Backward Compatibility)
```go
// FindAuthTokens searches for tokens in order:
// 1. Account-based location (new)
// 2. Instance-based location (old, for migration)
// 3. Legacy location (oldest)
func FindAuthTokens(cacheDir, instance, accountEmail string) (string, error) {
    // Try new account-based location first
    if accountEmail != "" {
        accountPath := GetAuthTokensPathByAccount(cacheDir, accountEmail)
        if _, err := os.Stat(accountPath); err == nil {
            return accountPath, nil
        }
    }
    
    // Fall back to instance-based location
    instancePath := GetAuthTokensPath(cacheDir, instance)
    if _, err := os.Stat(instancePath); err == nil {
        // Migrate to account-based location
        if accountEmail != "" {
            migrateTokens(instancePath, GetAuthTokensPathByAccount(cacheDir, accountEmail))
        }
        return instancePath, nil
    }
    
    // Fall back to legacy location
    legacyPath := GetAuthTokensPathFromCacheDir(cacheDir)
    return legacyPath, nil
}
```

#### Phase 3: Deprecation (Future)
- Log warnings when using instance-based paths
- Auto-migrate tokens to account-based location
- Eventually remove instance-based support

## Alternative Approaches Considered

### Option 1: Account Email as Directory Name
**Path**: `~/.cache/onemount/accounts/user@example.com/auth_tokens.json`

**Pros**:
- Human-readable
- Easy to debug

**Cons**:
- Special characters in email (e.g., `+`, `.`) cause filesystem issues
- Privacy concern (email visible in filesystem)
- Case sensitivity issues (User@Example.com vs user@example.com)

**Decision**: ❌ Rejected - Too many edge cases

### Option 2: Account ID from Microsoft Graph
**Path**: `~/.cache/onemount/accounts/{graph-user-id}/auth_tokens.json`

**Pros**:
- Guaranteed unique
- No special characters
- Official Microsoft identifier

**Cons**:
- Requires API call to get user ID
- Can't determine path before authentication
- Chicken-and-egg problem for token storage

**Decision**: ❌ Rejected - Can't use before auth

### Option 3: Hash of Account Email (RECOMMENDED)
**Path**: `~/.cache/onemount/accounts/{hash}/auth_tokens.json`

**Pros**:
- Stable and deterministic
- No special characters
- Privacy-preserving
- Works before full authentication
- Case-insensitive (normalize before hashing)

**Cons**:
- Not human-readable
- Requires account email to find tokens

**Decision**: ✅ **RECOMMENDED** - Best balance of benefits

## Migration Strategy

### Step 1: Implement Account-Based Storage
- Add new functions for account-based paths
- Keep existing instance-based functions
- No breaking changes

### Step 2: Add Migration Logic
- Check account-based location first
- Fall back to instance-based location
- Auto-migrate when found in old location

### Step 3: Update Authentication Flow
- Store account email in token file
- Use account email to determine storage path
- Update all token save/load operations

### Step 4: Update Tests
- Update test fixtures to use account-based paths
- Add migration tests
- Verify backward compatibility

### Step 5: Documentation
- Update architecture docs
- Add migration guide for users
- Document new token storage location

## Impact Analysis

### Breaking Changes
**None** - Migration path maintains backward compatibility

### User Impact
**Positive**:
- Tokens persist across mount point changes
- No re-authentication needed when remounting
- Clearer token organization

**Neutral**:
- Tokens automatically migrated on first use
- No user action required

### Test Impact
**Positive**:
- Tests reliably find tokens
- No mount point confusion
- Consistent test environment

### Performance Impact
**Negligible**:
- Hash calculation is fast (< 1ms)
- Same number of file operations
- No additional API calls

## Security Considerations

### Token File Permissions
**Current**: `0600` (owner read/write only)  
**Proposed**: `0600` (unchanged)

### Hash Algorithm
**Choice**: SHA256  
**Rationale**:
- Cryptographically secure
- Collision-resistant
- Standard library support
- Fast computation

### Privacy
**Current**: Mount point visible in path (e.g., `/home/user/OneDrive`)  
**Proposed**: Account hash (e.g., `a1b2c3d4e5f6g7h8`)  
**Improvement**: Email not visible in filesystem

## Testing Requirements

### Unit Tests
- [ ] Test account hash generation
- [ ] Test account-based path generation
- [ ] Test token migration logic
- [ ] Test backward compatibility

### Integration Tests
- [ ] Test token storage with real account
- [ ] Test token retrieval after remount
- [ ] Test migration from old to new location
- [ ] Test multiple accounts

### System Tests
- [ ] Test Docker environment token access
- [ ] Test production token reuse in tests
- [ ] Test mount point changes
- [ ] Test multiple simultaneous mounts

## Recommended Next Steps

### Immediate (This Sprint)
1. **Create Investigation Task** - Add to task list for detailed analysis
2. **Prototype Implementation** - Test account-based storage approach
3. **Validate Docker Behavior** - Verify tokens accessible in Docker

### Short Term (Next Sprint)
4. **Implement Account-Based Storage** - Add new functions
5. **Add Migration Logic** - Backward compatibility
6. **Update Tests** - Use new token paths
7. **Documentation** - Update architecture docs

### Long Term (Future Release)
8. **Deprecate Instance-Based Storage** - Log warnings
9. **Remove Old Code** - Clean up after migration period
10. **Monitor Adoption** - Track migration success

## References

### Code Locations
- Token path logic: `internal/graph/oauth2.go:47-58`
- Token save/load: `internal/graph/oauth2.go:60-95`
- Test token path: `internal/testutil/test_constants.go:15`
- Docker auth setup: `docker/images/test-runner/entrypoint.sh`

### Related Issues
- Task 46.1.6: Integration test auth handling
- Docker auth reference system: `scripts/setup-auth-reference.sh`
- Test fixtures: `internal/testutil/helpers/fs_fixtures.go`

### Documentation
- Architecture: `docs/2-architecture/authentication.md` (needs update)
- Testing: `docs/testing/test-fixtures.md`
- Docker: `docs/testing/docker-test-environment.md`

## Conclusion

The current mount-point-based token storage causes reliability issues in Docker environments and creates unnecessary token duplication. **Account-based storage using email hash** is the recommended solution, providing:

- ✅ Mount point independence
- ✅ No token duplication
- ✅ Reliable Docker testing
- ✅ Backward compatibility via migration
- ✅ Better privacy (hashed email)

**Priority**: HIGH - This affects test reliability and user experience  
**Complexity**: MEDIUM - Requires careful migration strategy  
**Risk**: LOW - Backward compatibility maintained
