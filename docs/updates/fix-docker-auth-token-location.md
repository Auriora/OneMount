# Fix Docker Auth Token Location Priority

**Date**: 2025-11-12  
**Type**: Bug Fix  
**Component**: Docker Test Infrastructure, Auth Setup Script

## Problem

The Docker test entrypoint was showing misleading auth token messages:
```
[SUCCESS] Auth tokens are valid JSON
[WARNING] Auth tokens appear to be expired
[INFO] Attempting to refresh tokens...
[WARNING] Failed to refresh tokens - tests may fail
```

Even after refreshing tokens with `scripts/setup-test-auth.sh`, the warnings persisted.

### Root Cause

1. **Multiple token file locations** existed with different timestamps:
   - `./auth_tokens.json` (June 6 - stale, in workspace root)
   - `./test-artifacts/.auth_tokens.json` (current location)
   - `~/.onemount-tests/.auth_tokens.json` (user home)

2. **Wrong priority order**: The entrypoint checked workspace root FIRST, finding the stale June tokens

3. **Setup script didn't update workspace**: `setup-test-auth.sh` only saved to `~/.onemount-tests/`, not to the workspace location Docker uses

## Solution

### 1. Fixed Token Location Priority in `docker/scripts/test-entrypoint.sh`

Changed the search order to check most reliable locations first:
1. ✅ `./test-artifacts/.auth_tokens.json` (preferred for Docker tests)
2. ✅ `./test-artifacts/auth_tokens.json` (alternate)
3. ✅ `~/.onemount-tests/.auth_tokens.json` (mounted from host)
4. ⚠️ `./auth_tokens.json` (legacy, with warning about staleness)

### 2. Updated `scripts/setup-test-auth.sh`

Now copies tokens to BOTH locations:
- `~/.onemount-tests/.auth_tokens.json` (user home)
- `./test-artifacts/.auth_tokens.json` (workspace - for Docker)

## Testing

Before fix:
```bash
docker compose ... run --rm test-runner go version
# [WARNING] Auth tokens appear to be expired
# [WARNING] Failed to refresh tokens - tests may fail
```

After fix:
```bash
docker compose ... run --rm test-runner go version
# [INFO] Auth tokens found in test-artifacts directory (hidden file)
# [SUCCESS] Auth tokens are valid JSON
# [SUCCESS] Auth tokens are valid
# go version go1.24.2 linux/amd64
```

## Files Modified

- `docker/scripts/test-entrypoint.sh` - Fixed token location priority
- `scripts/setup-test-auth.sh` - Copy tokens to workspace location

## Impact

- No more misleading "expired token" warnings when tokens are fresh
- Docker tests now use the correct, up-to-date token file
- Setup script ensures tokens are in the right place for Docker
- Legacy workspace root location still supported with warning

## Recommendation

Users should:
1. Run `scripts/setup-test-auth.sh` to refresh tokens
2. Remove any stale `./auth_tokens.json` in workspace root
3. Use `./test-artifacts/.auth_tokens.json` as the canonical location
