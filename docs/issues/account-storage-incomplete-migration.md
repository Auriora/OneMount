# Account-Based Storage Incomplete Migration

**Date**: 2026-01-27  
**Severity**: CRITICAL  
**Status**: Identified - Requires Fix

## Problem Statement

The account-based token storage system was implemented in `oauth2_account_storage.go`, but the migration is incomplete. Several functions still use the old path-based system (`~/.cache/onemount/{mount-path}/`), causing the launcher to fail when trying to display account information.

## Evidence

User reported error when running launcher:
```
Could not determine user principal name. error="open /home/bcherrington/.cache/onemount/home-bcherrington-onmount\\x2dauriora/auth_tokens.json: no such file or directory"
```

This shows the code is looking in the OLD location based on mount path, not the NEW account-based location.

## Root Cause Analysis

### What Was Implemented Correctly

1. **`oauth2_account_storage.go`**: New functions for account-based storage
   - `GetAuthTokensPathByAccount()` - Returns path based on account email hash
   - `FindAuthTokens()` - Searches all locations (account, instance, legacy)
   - `migrateTokens()` - Migrates from old to new location

2. **`cmd/onemount/main.go`**: Updated to use `AuthenticateWithAccountStorage()`
   - Properly calls the new authentication function
   - Should save tokens to account-based location

### What Was NOT Updated

1. **`graph.GetAccountName()`** (internal/graph/oauth2.go:95)
   - Still uses `GetAuthTokensPath(cacheDir, instance)` - OLD path-based system
   - Should use `FindAuthTokens(cacheDir, instance, "")` to search all locations

2. **`ui.GetKnownMounts()`** (internal/ui/onemount.go:56)
   - Scans cache directory for subdirectories with auth tokens
   - Only finds OLD instance-based tokens
   - Doesn't scan `accounts/` subdirectory for NEW account-based tokens

3. **Launcher** (cmd/onemount-launcher/main.go:356)
   - Calls `graph.GetAccountName()` which uses old path
   - Fails to find tokens in new account-based location

## Why This Happened

Looking at the git history and test files, it appears:

1. Account-based storage was implemented as a new feature
2. Tests were written that work with BOTH old and new systems
3. The migration path was implemented (old → new)
4. BUT the lookup functions were not updated to search new locations
5. Tests pass because they use the old locations OR explicitly set up new locations
6. Real-world usage fails because tokens are in new location but lookups use old paths

**This is a classic case of "tests passing but production broken"** - the tests were adapted to the implementation rather than the implementation being fully completed.

## Impact

- **Launcher cannot display account names** - Shows mount path instead
- **User experience degraded** - No indication of which OneDrive account is mounted
- **Confusion for users** - Multiple mounts look identical
- **Breaks the purpose of account-based storage** - Tokens are stored by account but never found

## Required Fixes

### 1. Fix `GetAccountName()` (HIGH PRIORITY)

```go
// BEFORE (WRONG):
func GetAccountName(cacheDir, instance string) (string, error) {
    tokenFile := GetAuthTokensPath(cacheDir, instance)  // OLD PATH
    auth, err := LoadAuthTokens(tokenFile)
    if err != nil {
        return "", err
    }
    return auth.Account, nil
}

// AFTER (CORRECT):
func GetAccountName(cacheDir, instance string) (string, error) {
    // Search all locations (account-based, instance-based, legacy)
    tokenFile, err := FindAuthTokens(cacheDir, instance, "")
    if err != nil {
        return "", err
    }
    
    auth, err := LoadAuthTokens(tokenFile)
    if err != nil {
        return "", err
    }
    return auth.Account, nil
}
```

### 2. Fix `GetKnownMounts()` (MEDIUM PRIORITY)

This is more complex because account-based tokens don't map directly to mount points. Options:

**Option A**: Query systemd for active `onemount@*` units (RECOMMENDED)
- More reliable - shows what's actually mounted
- Works with both old and new token storage
- Requires systemd integration

**Option B**: Maintain a mount registry file
- Store mapping of account → mount path
- Update on mount/unmount
- Additional complexity

**Option C**: Scan both old and new locations
- Check instance-based directories (old)
- Check accounts/ directory (new) but can't map back to mount path
- Incomplete solution

### 3. Update Tests (HIGH PRIORITY)

Tests need to be reviewed to ensure they:
- Test the ACTUAL production code paths
- Don't work around incomplete implementations
- Verify account-based storage is used
- Verify old tokens are migrated

## Recommended Approach

1. **Immediate Fix**: Update `GetAccountName()` to use `FindAuthTokens()`
2. **Short Term**: Implement systemd-based mount discovery for launcher
3. **Medium Term**: Review all functions that access auth tokens
4. **Long Term**: Remove old path-based functions after migration period

## Testing Strategy

After fixes:
1. Remove all old token files
2. Authenticate fresh (should create account-based tokens)
3. Verify launcher shows account name
4. Verify mount/unmount works
5. Test with multiple accounts

## Lessons Learned

1. **Complete migrations before merging** - Don't leave half-migrated code
2. **Test real-world scenarios** - Not just unit tests
3. **Review all callers** - When changing storage, update all access points
4. **Integration tests** - Test the full flow, not just individual functions
5. **Code review focus** - Look for incomplete migrations

## Related Files

- `internal/graph/oauth2.go` - Contains `GetAccountName()` that needs fixing
- `internal/graph/oauth2_account_storage.go` - New account-based storage (correct)
- `internal/ui/onemount.go` - Contains `GetKnownMounts()` that needs updating
- `cmd/onemount-launcher/main.go` - Launcher that calls broken functions
- `cmd/onemount/main.go` - Main binary (correctly uses new system)

## Priority

**CRITICAL** - This breaks core functionality (launcher) and defeats the purpose of the account-based storage refactoring.
