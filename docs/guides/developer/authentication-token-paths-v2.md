# Authentication Token File Paths - Developer Guide (v2.0)

**Last Updated**: 2026-01-23  
**Status**: Current (Account-Based Storage)  
**Version**: 2.0 - Account-Based Token Storage

## Overview

OneMount now uses **account-based token storage** instead of mount-point-based storage. This provides:
- Mount point independence - Same tokens regardless of where you mount
- No token duplication - One account = one token file
- Reliable Docker testing - Tests find tokens regardless of mount point
- Better multi-account support - Clear separation of credentials

---

## Token Storage Architecture

### New Architecture (v2.0 - Current)

**Location**: `~/.cache/onemount/accounts/{account-hash}/auth_tokens.json`

Where `{account-hash}` is the first 16 characters of the SHA256 hash of the account email (normalized to lowercase).

**Example**:
```bash
# Account: user@example.com
# Hash: b4c9a289323b21a0
# Token path: ~/.cache/onemount/accounts/b4c9a289323b21a0/auth_tokens.json

# Account: work@company.com  
# Hash: 7f3a8b2c9d1e4f5a
# Token path: ~/.cache/onemount/accounts/7f3a8b2c9d1e4f5a/auth_tokens.json
```

**Benefits**:
- **Mount Point Independent**: Tokens accessible regardless of mount location
- **Privacy**: Email not visible in filesystem (only hash)
- **Collision Resistant**: SHA256 provides 2^64 possible values
- **Deterministic**: Same email always produces same hash

### Old Architecture (v1.0 - Deprecated)

**Location**: `~/.cache/onemount/{instance}/auth_tokens.json`

Where `{instance}` was derived from the mount point path (escaped).

**Example**:
```bash
# Mount at /home/user/OneDrive
# Token path: ~/.cache/onemount/home-user-OneDrive/auth_tokens.json
```

**Status**: Still supported for backward compatibility with automatic migration

---

## Automatic Migration

OneMount automatically migrates tokens from old locations to the new account-based location:

### Migration Process

1. **First Token Access**: When OneMount needs authentication tokens
2. **Search Order**:
   - Account-based location (new)
   - Instance-based location (old)
   - Legacy location (oldest)
3. **Auto-Migration**: If tokens found in old location, automatically copy to new location
4. **Safety**: Old tokens preserved (not deleted) for safety

### Migration Example

```bash
# Before migration
~/.cache/onemount/home-user-OneDrive/auth_tokens.json  # Old location

# After first mount
~/.cache/onemount/accounts/b4c9a289323b21a0/auth_tokens.json  # New location
~/.cache/onemount/home-user-OneDrive/auth_tokens.json  # Old location (preserved)
```

### Migration Logging

```
INFO: Loaded auth tokens from instance-based location
INFO: Migrated auth tokens to account-based location
      from: ~/.cache/onemount/home-user-OneDrive/auth_tokens.json
      to:   ~/.cache/onemount/accounts/b4c9a289323b21a0/auth_tokens.json
```

---

## Token Storage Locations by Context

### 1. **Production/Normal Usage** (Main Application)

**Location**: `~/.cache/onemount/accounts/{account-hash}/auth_tokens.json`

**Code Reference**: 
- `internal/graph/oauth2_account_storage.go`: `GetAuthTokensPathByAccount(cacheDir, accountEmail)`
- `cmd/onemount/main.go`: Uses `AuthenticateWithAccountStorage()`

**Key Points**:
- Account-based storage (mount point independent)
- Automatic migration from old locations
- File permissions: `0600` (owner read/write only)
- Directory permissions: `0700` (owner read/write/execute only)

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
- Works with both old and new token storage locations

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

## Authentication Flow

### Token Path Determination (New)

The token path is now **account-based** instead of mount-point-based:

```go
// From internal/graph/oauth2_account_storage.go
accountHash := hashAccount(accountEmail)  // SHA256 hash (first 16 chars)
tokenPath := filepath.Join(cacheDir, "accounts", accountHash, "auth_tokens.json")
```

**Example**:
```bash
# Account: user@example.com
# Hash: b4c9a289323b21a0
# Token path: ~/.cache/onemount/accounts/b4c9a289323b21a0/auth_tokens.json
```

### How Tokens Are Created

1. **User runs onemount** with a mount point
2. **Check for existing tokens**:
   - Try account-based location (if account email known)
   - Try instance-based location (old)
   - Try legacy location (oldest)
3. **If no tokens found**, trigger authentication:
   - GTK mode: Opens browser window
   - Headless mode (`--no-browser`): Displays URL and waits for redirect
4. **Get account email** from Microsoft Graph API
5. **Save tokens** to account-based location
6. **Set permissions**: File `0600`, Directory `0700`

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

## Hash Algorithm Details

### SHA256 Hash Generation

```go
// Normalize email: lowercase and trim whitespace
normalized := strings.ToLower(strings.TrimSpace(email))

// Compute SHA256 hash
hash := sha256.Sum256([]byte(normalized))

// Return first 16 hex characters (64 bits)
return hex.EncodeToString(hash[:])[:16]
```

### Collision Probability

| Number of Accounts | Collision Probability |
|-------------------|----------------------|
| 1,000 | ~0.0000000027% |
| 10,000 | ~0.000027% |
| 100,000 | ~0.027% |
| 1,000,000 | ~2.7% |

For typical use cases (< 10,000 accounts per system), collision risk is negligible.

### Hash Properties

- **Deterministic**: Same email always produces same hash
- **Case-Insensitive**: "User@Example.com" and "user@example.com" produce same hash
- **Privacy-Preserving**: Email not visible in filesystem
- **Collision-Resistant**: SHA256 provides strong collision resistance

---

## Best Practices

### For Developers

1. **Use account-based storage** for new code:
   ```go
   auth, err := graph.AuthenticateWithAccountStorage(ctx, config, cacheDir, instance, headless)
   ```

2. **Never hardcode paths** - Use helper functions:
   - `graph.GetAuthTokensPathByAccount(cacheDir, accountEmail)` for account-based storage
   - `graph.FindAuthTokens(cacheDir, instance, accountEmail)` for search with fallback
   - Check `ONEMOUNT_AUTH_PATH` environment variable in tests

3. **Document token location** in test files and scripts

4. **Set proper permissions** (`0600` for files, `0700` for directories)

### For Manual Testing

1. **Authenticate once** - Tokens will be stored in account-based location
2. **Remount anywhere** - Same tokens will be used
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
# Find all token files (old and new locations)
find ~/.cache/onemount -name "auth_tokens.json" 2>/dev/null

# Check token expiration and account
cat ~/.cache/onemount/accounts/*/auth_tokens.json | jq '.expires_at, .account'

# Re-authenticate if needed
./build/onemount --auth-only ~/OneDrive
```

### Tokens not migrating automatically

**Cause**: Account email not available during token search

**Solution**: 
- Ensure old tokens contain `account` field
- If missing, re-authenticate to create new tokens with account email

### Multiple token files for same account

**Cause**: Tokens created before migration to account-based storage

**Solution**:
- Old tokens will be automatically migrated on first use
- Old files are preserved for safety
- Can be manually deleted after successful migration

---

## Migration Guide for Users

### Automatic Migration (Recommended)

1. **Just mount normally** - Migration happens automatically
2. **Check logs** for migration messages
3. **Verify new location**:
   ```bash
   ls -la ~/.cache/onemount/accounts/*/auth_tokens.json
   ```

### Manual Migration (If Needed)

```bash
# Find your account email
cat ~/.cache/onemount/home-user-OneDrive/auth_tokens.json | jq -r '.account'

# Calculate hash (requires Python)
python3 -c "import hashlib; print(hashlib.sha256(b'user@example.com').hexdigest()[:16])"

# Create new directory
mkdir -p ~/.cache/onemount/accounts/b4c9a289323b21a0

# Copy tokens
cp ~/.cache/onemount/home-user-OneDrive/auth_tokens.json \
   ~/.cache/onemount/accounts/b4c9a289323b21a0/auth_tokens.json

# Set permissions
chmod 700 ~/.cache/onemount/accounts/b4c9a289323b21a0
chmod 600 ~/.cache/onemount/accounts/b4c9a289323b21a0/auth_tokens.json
```

### Cleanup Old Tokens (Optional)

After successful migration and verification:

```bash
# Remove old instance-based tokens
rm -rf ~/.cache/onemount/home-user-OneDrive/auth_tokens.json

# Remove old legacy tokens
rm -f ~/.cache/onemount/auth_tokens.json
```

**Warning**: Only remove old tokens after verifying new tokens work correctly!

---

## API Reference

### New Functions (v2.0)

```go
// Get account-based token path
func GetAuthTokensPathByAccount(cacheDir, accountEmail string) string

// Hash account email
func hashAccount(email string) string

// Find tokens with fallback and migration
func FindAuthTokens(cacheDir, instance, accountEmail string) (string, error)

// Migrate tokens from old to new location
func migrateTokens(oldPath, newPath string) error

// Authenticate with account-based storage
func AuthenticateWithAccountStorage(ctx context.Context, config AuthConfig, cacheDir, instance string, headless bool) (*Auth, error)
```

### Legacy Functions (v1.0 - Still Supported)

```go
// Get instance-based token path (deprecated)
func GetAuthTokensPath(cacheDir, instance string) string

// Get legacy token path (deprecated)
func GetAuthTokensPathFromCacheDir(cacheDir string) string

// Authenticate with explicit path (deprecated)
func Authenticate(ctx context.Context, config AuthConfig, path string, headless bool) (*Auth, error)
```

---

## References

- `internal/graph/oauth2_account_storage.go` - Account-based storage implementation
- `internal/graph/oauth2.go` - Legacy token path functions
- `cmd/onemount/main.go` - Main application auth flow
- `docs/reports/2026-01-23-063800-auth-token-storage-architecture-analysis.md` - Architecture analysis
- `docs/plans/auth-token-storage-refactoring-plan.md` - Refactoring plan
- `docs/reports/2026-01-23-task-4-9-1-investigation-findings.md` - Investigation findings

---

## Changelog

### v2.0 (2026-01-23)
- **BREAKING**: Changed to account-based token storage
- **ADDED**: Automatic migration from old locations
- **ADDED**: `GetAuthTokensPathByAccount()` function
- **ADDED**: `AuthenticateWithAccountStorage()` function
- **ADDED**: `FindAuthTokens()` with fallback logic
- **DEPRECATED**: Instance-based token storage (still supported)
- **IMPROVED**: Mount point independence
- **IMPROVED**: Docker test reliability
- **IMPROVED**: Multi-account support

### v1.0 (2025-11-13)
- Initial documentation
- Instance-based token storage
- XDG compliance

