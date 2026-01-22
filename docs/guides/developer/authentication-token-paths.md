# Authentication Token File Paths - Developer Guide

**Last Updated**: 2026-01-22  
**Status**: Critical Documentation

## Problem Statement

There has been significant confusion about where authentication tokens are stored, leading to:
- Users having to re-authenticate multiple times
- Tests failing to find auth tokens
- Inconsistent behavior between Docker and host environments
- Manual test scripts creating tokens in unexpected locations

This document clarifies the **authoritative** token storage locations and provides guidelines for all scenarios.

---

## Token Storage Locations by Context

### 1. **Production/Normal Usage** (Main Application)

**Location**: `~/.cache/onemount/<instance>/auth_tokens.json`

Where `<instance>` is derived from the mount point path (escaped).

**Example**:
```bash
# Mounting at /home/user/OneDrive
Token path: ~/.cache/onemount/home-user-OneDrive/auth_tokens.json

# Mounting at /tmp/test-mount
Token path: ~/.cache/onemount/tmp-test-mount/auth_tokens.json
```

**Code Reference**: 
- `internal/graph/oauth2.go`: `GetAuthTokensPath(cacheDir, instance)`
- `cmd/onemount/main.go`: Uses `GetAuthTokensPathFromCacheDir(cachePath)`

**Key Points**:
- The cache directory defaults to `~/.cache/onemount` (XDG_CACHE_HOME)
- Can be overridden with `--cache-dir` flag
- Each mount point gets its own subdirectory with auth tokens
- File permissions: `0600` (owner read/write only)

---

### 2. **Integration/System Tests** (Go Tests with Real OneDrive)

**Location**: `test-artifacts/.auth_tokens.json` (workspace relative)

**Alternative**: Set `ONEMOUNT_AUTH_PATH` environment variable

**Example**:
```bash
# Default location
test-artifacts/.auth_tokens.json

# Or override with environment variable
export ONEMOUNT_AUTH_PATH=/path/to/tokens.json
```

**Code Reference**:
- All `*_integration_test.go` and `*_real_test.go` files
- Pattern:
  ```go
  authPath := os.Getenv("ONEMOUNT_AUTH_PATH")
  if authPath == "" {
      authPath = "test-artifacts/.auth_tokens.json"
  }
  ```

**Key Points**:
- Tests look for `ONEMOUNT_AUTH_PATH` first
- Falls back to `test-artifacts/.auth_tokens.json`
- This is **workspace-relative**, not absolute
- Used by integration tests that need real OneDrive access

---

### 3. **Docker Test Environment**

**Location**: Mounted from host via Docker Compose override

**Setup**:
```bash
# Run setup script to configure auth reference
./scripts/setup-auth-reference.sh
```

This creates:
- `docker/compose/docker-compose.auth.yml` - Docker override with volume mount
- `.env.auth` - Environment configuration

**Canonical Token Location** (from `.env.auth`):
```bash
ONEMOUNT_AUTH_TOKEN_PATH=/home/user/.cache/onedriver/home-user-OneDrive/auth_tokens.json
```

**Inside Docker Container**:
```bash
/tmp/auth-tokens/auth_tokens.json
```

**Usage**:
```bash
# Run tests with auth
docker compose -f docker/compose/docker-compose.test.yml \
  -f docker/compose/docker-compose.auth.yml run --rm integration-tests
```

**Key Points**:
- Uses **reference-based** mounting (no copying/symlinking)
- Tokens stay in canonical location on host
- Docker mounts the parent directory read-only
- Environment variable `ONEMOUNT_AUTH_PATH_DOCKER` points to container path

---

### 4. **Unit Tests** (Mock/Sandbox)

**Location**: `~/.onemount-tests/.auth_tokens.json`

**Code Reference**:
- `internal/testutil/test_constants.go`: `AuthTokensPath`
- Temporary sandbox: `<temp>/.auth_tokens.json`

**Key Points**:
- Unit tests use a test sandbox directory
- Tokens are typically mocked or use test fixtures
- Not for real OneDrive access

---

### 5. **Manual Testing Scripts**

**Problem**: Manual test scripts (like `tests/manual/test_dbus_integration.sh`) were creating auth tokens in **mount-point-specific** cache directories, causing confusion.

**Solution**: Manual tests should use one of these approaches:

#### Option A: Use Existing Production Tokens
```bash
# Copy from production location
cp ~/.cache/onemount/home-user-OneDrive/auth_tokens.json \
   ~/.config/onemount/.auth_tokens.json
```

#### Option B: Authenticate Once, Reuse
```bash
# Authenticate once with a known mount point
./build/onemount --auth-only ~/OneDrive

# Tokens saved to: ~/.cache/onemount/home-user-OneDrive/auth_tokens.json

# Copy to standard location for reuse
mkdir -p ~/.config/onemount
cp ~/.cache/onemount/home-user-OneDrive/auth_tokens.json \
   ~/.config/onemount/.auth_tokens.json
```

#### Option C: Use Test Artifacts Location
```bash
# For consistency with integration tests
cp test-artifacts/.auth_tokens.json ~/.config/onemount/.auth_tokens.json
```

---

## Recommended Standard Locations

To reduce confusion, we recommend these **standard** locations:

| Context | Location | Purpose |
|---------|----------|---------|
| Production | `~/.cache/onemount/<instance>/auth_tokens.json` | Per-mount-point tokens |
| Integration Tests | `test-artifacts/.auth_tokens.json` | Workspace-relative test tokens |
| Docker Tests | `/tmp/auth-tokens/auth_tokens.json` | Mounted from host canonical location |
| Manual Testing | `~/.config/onemount/.auth_tokens.json` | Shared location for manual tests |

---

## Authentication Flow

### Who Creates the Token File?

**OneMount itself** creates and manages the token file - **NOT** the Microsoft SDK or any external service.

**Code Location**: `internal/graph/oauth2.go`
- `SaveAuthTokens()` - Writes tokens to disk
- `Auth.ToFile()` - Wrapper that calls SaveAuthTokens
- `newAuth()` - Performs OAuth flow and saves tokens
- `Authenticate()` - Main entry point that loads or creates tokens

### How the Token Path is Determined

The token path is **algorithmically derived** from the mount point:

```go
// From cmd/onemount/main.go
absMountPath, _ := filepath.Abs(mountpoint)
cachePath := filepath.Join(config.CacheDir, unit.UnitNamePathEscape(absMountPath))
authPath := filepath.Join(cachePath, "auth_tokens.json")
```

**Example**:
```bash
# Mount point: /home/user/OneDrive
# Cache dir: ~/.cache/onemount (default)
# Escaped path: home-user-OneDrive
# Token path: ~/.cache/onemount/home-user-OneDrive/auth_tokens.json

# Mount point: /tmp/test-mount
# Escaped path: tmp-test-mount  
# Token path: ~/.cache/onemount/tmp-test-mount/auth_tokens.json
```

**Key Points**:
- Path escaping uses `unit.UnitNamePathEscape()` (systemd-style escaping)
- Each mount point gets its own subdirectory
- Tokens are stored alongside other mount-specific data (metadata DB, content cache)
- **No user control** over token location (it's derived automatically)

### Why Tokens Cannot Be Shared

Tokens are **mount-point-specific** by design because:

1. **Cache isolation**: Each mount has its own metadata database and content cache
2. **Instance identification**: The cache directory name identifies the mount instance
3. **Concurrent mounts**: Multiple mounts of the same account need separate state
4. **Cleanup**: Unmounting can clean up mount-specific data including tokens

**This is intentional architecture**, not a limitation to be "fixed".

### How Tokens Are Created

1. **User runs onemount** with a mount point
2. **No existing tokens** found at `~/.cache/onemount/<instance>/auth_tokens.json`
3. **Authentication triggered**:
   - GTK mode: Opens browser window
   - Headless mode (`--no-browser`): Displays URL and waits for redirect
4. **Tokens saved** to `~/.cache/onemount/<instance>/auth_tokens.json`
5. **File permissions** set to `0600`

### Token File Format

```json
{
  "config": {
    "clientID": "...",
    "codeURL": "...",
    "tokenURL": "...",
    "redirectURL": "..."
  },
  "account": "user@example.com",
  "expires_in": 3599,
  "expires_at": 1769079645,
  "access_token": "...",
  "refresh_token": "..."
}
```

---

## Best Practices

### For Developers

1. **Never hardcode paths** - Use the appropriate helper functions:
   - `graph.GetAuthTokensPath(cacheDir, instance)` for production
   - `graph.GetAuthTokensPathFromCacheDir(cacheDir)` for backward compatibility
   - Check `ONEMOUNT_AUTH_PATH` environment variable in tests

2. **Document token location** in test files and scripts

3. **Use reference-based mounting** in Docker (no copying)

4. **Set proper permissions** (`0600`) when creating token files

### For Manual Testing

1. **Authenticate once** with a known mount point
2. **Copy tokens** to a standard location for reuse
3. **Document** which account was used
4. **Never commit** auth tokens to git

### For CI/CD

1. **Use environment variables** to specify token paths
2. **Mount tokens** into containers (don't copy)
3. **Rotate tokens** regularly
4. **Use separate accounts** for testing

---

## Troubleshooting

### "Authentication failed" or "No auth tokens found"

**Check**:
1. Does the token file exist at the expected location?
2. Are file permissions correct (`0600`)?
3. Is the token expired? (check `expires_at` field)
4. Is `ONEMOUNT_AUTH_PATH` set correctly (for tests)?

**Solution**:
```bash
# Find all token files
find ~ -name "auth_tokens.json" 2>/dev/null

# Check token expiration
cat ~/.cache/onemount/*/auth_tokens.json | jq '.expires_at, .account'

# Re-authenticate if needed
./build/onemount --auth-only ~/OneDrive
```

### "Had to login again when running script"

**Cause**: Script used a different mount point, creating tokens in a new location

**Solution**: Use a consistent mount point or copy tokens to standard location

### Docker tests can't find tokens

**Cause**: Auth override not included in docker-compose command

**Solution**:
```bash
# Include auth override
docker compose -f docker/compose/docker-compose.test.yml \
  -f docker/compose/docker-compose.auth.yml run --rm integration-tests
```

---

## Future Improvements

1. ~~**Centralized token storage**~~: **INVALID** - Tokens cannot be centralized because each mount point requires its own authentication context and cache directory
2. ~~**Token sharing**~~: **INVALID** - Tokens cannot be shared between mount points for the same architectural reason
3. **Better error messages**: ✅ **RECOMMENDED** - Include expected token path in error messages when authentication fails
4. **Token management CLI**: ✅ **ALREADY EXISTS**:
   - `--auth-only` - Authenticate and save tokens without mounting
   - `--stats <mountpoint>` - Display cache/token statistics for a mount point
   - `--wipe-cache` - Remove all cached data including tokens
5. ~~**Documentation in --help**~~: **NOT PRACTICAL** - Token path is contextual (depends on mount point), cannot be shown in static help text

### Existing Token Management Commands

```bash
# Authenticate without mounting (creates/refreshes tokens)
onemount --auth-only ~/OneDrive

# View statistics including token status
onemount --stats ~/OneDrive

# Remove all cached data including tokens
onemount --wipe-cache --cache-dir ~/.cache/onemount

# Headless authentication (no browser)
onemount --auth-only --no-browser ~/OneDrive
```

---

## References

- `internal/graph/oauth2.go` - Token path functions
- `cmd/onemount/main.go` - Main application auth flow
- `.env.auth` - Docker auth configuration
- `scripts/setup-auth-reference.sh` - Docker auth setup script
- `docs/TEST_SETUP.md` - Test environment setup guide
