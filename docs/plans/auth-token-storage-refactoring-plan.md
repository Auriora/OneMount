# Auth Token Storage Refactoring Plan

**Date**: 2026-01-23  
**Status**: 📋 Planned  
**Priority**: HIGH  
**Complexity**: MEDIUM

## Overview

Refactor auth token storage from mount-point-based to account-based to improve reliability, eliminate duplication, and fix Docker test environment issues.

## Problem

Current token storage uses mount point location, causing:
- Tokens not found when mount point changes
- Duplicate tokens for same account at different mount points
- Docker tests can't reliably find production tokens
- Confusion between test and production environments

## Solution

Store tokens based on account identity (email hash) instead of mount point location.

**Current**: `~/.cache/onemount/{mount-path}/auth_tokens.json`  
**Proposed**: `~/.cache/onemount/accounts/{account-hash}/auth_tokens.json`

## Implementation Phases

### Phase 1: Investigation & Prototyping (1-2 days)

**Goal**: Validate approach and identify edge cases

**Tasks**:
1. Analyze current token storage usage across codebase
2. Prototype account-based storage functions
3. Test hash generation and collision resistance
4. Verify Docker environment behavior
5. Document findings and edge cases

**Deliverables**:
- Prototype code in feature branch
- Test results document
- Updated analysis report

### Phase 2: Core Implementation (2-3 days)

**Goal**: Implement account-based storage with backward compatibility

**Tasks**:
1. Add `GetAuthTokensPathByAccount()` function
2. Add `hashAccount()` helper function
3. Implement `FindAuthTokens()` with fallback logic
4. Add automatic token migration
5. Update `SaveAuthTokens()` to use account-based path
6. Update `LoadAuthTokens()` to search multiple locations

**Deliverables**:
- Updated `internal/graph/oauth2.go`
- Unit tests for new functions
- Migration logic with tests

### Phase 3: Integration & Testing (2-3 days)

**Goal**: Update all token usage and verify behavior

**Tasks**:
1. Update `cmd/onemount/main.go` to use account-based paths
2. Update test fixtures to use account-based paths
3. Update Docker auth setup scripts
4. Add integration tests for migration
5. Test Docker environment token access
6. Test multiple account scenarios

**Deliverables**:
- Updated main application code
- Updated test fixtures
- Integration test suite
- Docker environment verification

### Phase 4: Documentation & Cleanup (1 day)

**Goal**: Document changes and prepare for release

**Tasks**:
1. Update architecture documentation
2. Add migration guide for users
3. Update test documentation
4. Add deprecation warnings for old paths
5. Update CHANGELOG

**Deliverables**:
- Updated `docs/2-architecture/authentication.md`
- Migration guide in `docs/guides/user/`
- Updated test documentation
- Release notes

## Technical Design

### New Functions

```go
// GetAuthTokensPathByAccount returns token path based on account identity
func GetAuthTokensPathByAccount(cacheDir, accountEmail string) string {
    accountHash := hashAccount(accountEmail)
    return filepath.Join(cacheDir, "accounts", accountHash, AuthTokensFileName)
}

// hashAccount creates a stable hash of account email
func hashAccount(email string) string {
    normalized := strings.ToLower(strings.TrimSpace(email))
    hash := sha256.Sum256([]byte(normalized))
    return hex.EncodeToString(hash[:])[:16]
}

// FindAuthTokens searches for tokens with fallback and migration
func FindAuthTokens(cacheDir, instance, accountEmail string) (string, error) {
    // 1. Try account-based location (new)
    if accountEmail != "" {
        accountPath := GetAuthTokensPathByAccount(cacheDir, accountEmail)
        if fileExists(accountPath) {
            return accountPath, nil
        }
    }
    
    // 2. Try instance-based location (old)
    instancePath := GetAuthTokensPath(cacheDir, instance)
    if fileExists(instancePath) {
        // Auto-migrate if we have account email
        if accountEmail != "" {
            if err := migrateTokens(instancePath, GetAuthTokensPathByAccount(cacheDir, accountEmail)); err == nil {
                return GetAuthTokensPathByAccount(cacheDir, accountEmail), nil
            }
        }
        return instancePath, nil
    }
    
    // 3. Try legacy location (oldest)
    legacyPath := GetAuthTokensPathFromCacheDir(cacheDir)
    if fileExists(legacyPath) {
        if accountEmail != "" {
            if err := migrateTokens(legacyPath, GetAuthTokensPathByAccount(cacheDir, accountEmail)); err == nil {
                return GetAuthTokensPathByAccount(cacheDir, accountEmail), nil
            }
        }
        return legacyPath, nil
    }
    
    // 4. Return new account-based path for creation
    if accountEmail != "" {
        return GetAuthTokensPathByAccount(cacheDir, accountEmail), nil
    }
    
    return legacyPath, nil
}

// migrateTokens moves tokens from old location to new location
func migrateTokens(oldPath, newPath string) error {
    // Create directory for new location
    if err := os.MkdirAll(filepath.Dir(newPath), 0700); err != nil {
        return err
    }
    
    // Copy tokens to new location
    data, err := os.ReadFile(oldPath)
    if err != nil {
        return err
    }
    
    if err := os.WriteFile(newPath, data, 0600); err != nil {
        return err
    }
    
    // Keep old file for safety (can be removed in future version)
    // os.Remove(oldPath)
    
    return nil
}
```

### Migration Strategy

**Automatic Migration**:
- On token load, check account-based location first
- If not found, check old locations
- If found in old location, copy to new location
- Keep old file for safety (remove in future version)

**User Communication**:
- Log info message when migrating tokens
- No user action required
- Transparent migration

### Backward Compatibility

**Guaranteed**:
- Old token locations still work
- Automatic migration on first use
- No breaking changes
- No re-authentication required

**Deprecation Timeline**:
- v1.0: Add account-based storage, auto-migrate
- v1.1: Log warnings for old locations
- v2.0: Remove support for old locations (after 6+ months)

## Testing Strategy

### Unit Tests
```go
func TestHashAccount(t *testing.T) {
    // Test hash generation
    // Test case insensitivity
    // Test whitespace handling
    // Test hash stability
}

func TestGetAuthTokensPathByAccount(t *testing.T) {
    // Test path generation
    // Test directory structure
    // Test different accounts
}

func TestFindAuthTokens(t *testing.T) {
    // Test account-based location
    // Test instance-based fallback
    // Test legacy fallback
    // Test migration
}

func TestMigrateTokens(t *testing.T) {
    // Test successful migration
    // Test error handling
    // Test directory creation
    // Test file permissions
}
```

### Integration Tests
```go
func TestIT_AuthTokenMigration(t *testing.T) {
    // Create tokens in old location
    // Load tokens (should trigger migration)
    // Verify tokens in new location
    // Verify old tokens still exist
}

func TestIT_MultipleAccountTokens(t *testing.T) {
    // Create tokens for multiple accounts
    // Verify separate storage
    // Verify no conflicts
}

func TestIT_DockerTokenAccess(t *testing.T) {
    // Store tokens in account-based location
    // Access from Docker with different mount point
    // Verify tokens found correctly
}
```

### System Tests
```bash
# Test token persistence across remounts
onemount /home/user/OneDrive
# Authenticate
fusermount -u /home/user/OneDrive

# Remount at different location
onemount /mnt/onedrive
# Should use existing tokens, no re-auth

# Test multiple accounts
onemount --account personal@example.com /home/user/Personal
onemount --account work@company.com /home/user/Work
# Verify separate token storage
```

## Risks & Mitigation

### Risk 1: Token Migration Failure
**Impact**: User must re-authenticate  
**Probability**: Low  
**Mitigation**:
- Keep old tokens as backup
- Extensive testing of migration logic
- Graceful fallback to re-authentication

### Risk 2: Hash Collisions
**Impact**: Two accounts share token file  
**Probability**: Extremely low (SHA256)  
**Mitigation**:
- Use cryptographically secure hash
- Use sufficient hash length (16 chars = 64 bits)
- Monitor for collisions in logs

### Risk 3: Account Email Changes
**Impact**: Can't find old tokens after email change  
**Probability**: Low (rare for Microsoft accounts)  
**Mitigation**:
- Document email change procedure
- Provide manual migration tool
- Keep old tokens as fallback

### Risk 4: Backward Compatibility Issues
**Impact**: Old code breaks with new storage  
**Probability**: Low  
**Mitigation**:
- Maintain fallback to old locations
- Extensive compatibility testing
- Gradual deprecation timeline

## Success Criteria

### Functional
- ✅ Tokens stored by account, not mount point
- ✅ Automatic migration from old locations
- ✅ Multiple accounts work correctly
- ✅ Docker tests find tokens reliably
- ✅ No re-authentication required

### Non-Functional
- ✅ No breaking changes
- ✅ Performance unchanged (< 1ms overhead)
- ✅ All existing tests pass
- ✅ New tests cover migration scenarios
- ✅ Documentation updated

### User Experience
- ✅ Transparent migration (no user action)
- ✅ Tokens persist across remounts
- ✅ Clear error messages if issues occur
- ✅ No confusion about token locations

## Timeline

**Total Estimate**: 6-9 days

- Phase 1 (Investigation): 1-2 days
- Phase 2 (Implementation): 2-3 days
- Phase 3 (Integration): 2-3 days
- Phase 4 (Documentation): 1 day

**Dependencies**:
- None (can start immediately)

**Blockers**:
- None identified

## Resources Required

**Development**:
- 1 developer (full-time)
- Access to test OneDrive accounts
- Docker environment for testing

**Testing**:
- Multiple OneDrive accounts for testing
- Various mount point scenarios
- Docker test environment

**Documentation**:
- Architecture documentation updates
- User migration guide
- Test documentation updates

## Approval & Sign-off

**Stakeholders**:
- [ ] Development Lead
- [ ] QA Lead
- [ ] Documentation Lead

**Approvals**:
- [ ] Technical Design Review
- [ ] Security Review
- [ ] User Experience Review

## References

- Analysis Report: `docs/reports/2026-01-23-063800-auth-token-storage-architecture-analysis.md`
- Current Implementation: `internal/graph/oauth2.go`
- Test Fixtures: `internal/testutil/helpers/fs_fixtures.go`
- Docker Setup: `docker/images/test-runner/entrypoint.sh`
