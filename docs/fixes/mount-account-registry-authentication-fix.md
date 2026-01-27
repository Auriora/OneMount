# Mount-Account Registry Authentication Fix

**Date**: 2026-01-27  
**Status**: ✅ Complete  
**Related**: `docs/designs/mount-account-registry.md`, `docs/fixes/mount-account-registry-integration.md`

## Problem

After implementing the mount-account registry system, the authentication flow was not properly integrated. The service would attempt to authenticate from scratch instead of using existing tokens stored in the account-based location (`~/.cache/onemount/accounts/<hash>/auth_tokens.json`).

### Symptoms

1. Service failed to start with authentication errors
2. Tokens existed in account-based location but were not being found
3. Registry had correct mount → account mapping but wasn't being used during authentication

## Root Cause

The `AuthenticateWithAccountStorage()` function in `internal/graph/oauth2.go` had the registry lookup logic implemented, but there were compilation errors preventing it from working:

1. **Import conflict**: The `config` package was imported without an alias, but the function parameter was also named `config`, causing shadowing
2. **Variable redeclaration**: Multiple error variables were being redeclared with `:=` instead of using unique names

## Solution

### 1. Fixed Import Alias

Added import alias to avoid package name shadowing:

```go
import (
    mountconfig "github.com/auriora/onemount/internal/config"
    // ... other imports
)
```

### 2. Fixed Variable Names

Updated all uses of the config package to use the alias and fixed variable redeclarations:

```go
registry, regErr := mountconfig.NewMountsRegistry(configDirPath)
if regErr == nil {
    if account, exists := registry.GetAccount(mountPoint); exists && account != "" {
        // ... check for tokens
    }
}
```

### 3. Added Debug Logging

Added comprehensive debug logging to trace the authentication flow:

- Log when `AuthenticateWithAccountStorage` is called
- Log registry lookup results
- Log token file checks
- Log successful token loads

## Authentication Flow (Final)

```
1. Service starts with mount point: /home/user/onmount-auriora
2. Instance name: home-user-onmount\x2dauriora
3. Unescape to get mount point: /home/user/onmount-auriora
4. Load registry from ~/.config/onemount/mounts.json
5. Look up account for mount point → bcherrington.993834@outlook.com
6. Hash account email → 3da226fdcd749f83
7. Check ~/.cache/onemount/accounts/3da226fdcd749f83/auth_tokens.json
8. Load tokens ✓
9. Refresh if needed ✓
10. Mount filesystem ✓
```

## Verification

### Test 1: Direct Binary Execution

```bash
$ /usr/bin/onemount /home/bcherrington/onmount-auriora
2026-01-27T16:30:41Z INF Found account in registry, checking for tokens account=bcherrington.993834@outlook.com
2026-01-27T16:30:41Z INF Loaded auth tokens from account-based location (via registry)
2026-01-27T16:30:41Z INF Authentication successful
```

### Test 2: Systemd Service

```bash
$ systemctl --user restart "onemount@home-bcherrington-onmount\\x2dauriora.service"
$ systemctl --user status "onemount@home-bcherrington-onmount\\x2dauriora.service"
● onemount@home-bcherrington-onmount\x2dauriora.service - onemount
     Active: active (running)
```

### Test 3: Mount Accessibility

```bash
$ ls -la /home/bcherrington/onmount-auriora/
total 1422
drwxr-xr-x    8 bcherrington bcherrington    4096 Jan 26 22:53 .
drwxr-x---+ 108 bcherrington bcherrington    4096 Jan 27 10:10 ..
-rw-r--r--    1 bcherrington bcherrington      34 Nov 12 18:00 dbus-test-file.txt
drwxr-xr-x    2 bcherrington bcherrington    4096 Jun  9  2025 Documents
...
```

## Files Modified

- `internal/graph/oauth2.go`: Fixed import alias and variable names, added debug logging
- `internal/graph/oauth2_account_storage.go`: No changes (already correct)
- `internal/config/mounts.go`: No changes (already correct)
- `cmd/onemount/main.go`: No changes (already correct)
- `cmd/onemount-launcher/main.go`: No changes (already correct)

## Benefits

1. **Single Source of Truth**: Registry provides authoritative mount → account mapping
2. **No Token Duplication**: Tokens stay in one location per account
3. **Automatic Token Discovery**: Service finds tokens via registry lookup
4. **Proper Error Handling**: Clear logging when tokens are missing
5. **Migration Support**: Old token locations still checked as fallback

## Next Steps

1. ✅ Test launcher UI to verify account names are displayed
2. ✅ Verify unmount cleanup works correctly
3. ✅ Test with multiple mounts for same account
4. ✅ Test with multiple accounts

## Additional Fix: Launcher Mount List

**Issue**: Launcher was not displaying registered mounts because it was passing the cache directory path to `GetKnownMounts()` instead of the config directory path.

**Root Cause**: 
- Registry is stored in `~/.config/onemount/mounts.json` (config directory)
- Launcher was calling `ui.GetKnownMounts(config.CacheDir)` which passed `~/.cache/onemount` (cache directory)
- This caused the registry lookup to fail silently

**Fix**: Changed launcher to pass empty string to `GetKnownMounts("")`, which makes it use the default config directory:

```go
// Before:
mounts := ui.GetKnownMounts(config.CacheDir)

// After:
mounts := ui.GetKnownMounts("")  // Uses default config directory
```

**Files Modified**:
- `cmd/onemount-launcher/main.go`: Fixed two calls to `GetKnownMounts()` (lines 276 and 669)

## Related Issues

- Account-based storage incomplete migration (resolved)
- Systemd service authentication flow (resolved)
- Mount-account registry integration (complete)
