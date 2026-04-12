# Design: Authentication and Account Management

## Problem Analysis

OneMount needs to:
1. Authenticate users with Microsoft accounts using OAuth2
2. Store and refresh authentication tokens securely
3. Support multiple accounts simultaneously
4. Use account-based storage paths to prevent token duplication

## Current Implementation Review

### OAuth2 Implementation
- Located in `internal/graph/oauth2.go`, `oauth2_gtk.go`, `oauth2_headless.go`
- Supports both interactive (GTK) and headless (device code) flows
- Token storage uses account-based paths (refactored from mount-point-based)

### Token Storage (Refactored — Complete)
- Tokens now stored in `{cacheDir}/accounts/{account-hash}/auth_tokens.json` (account-based)
- Fallback to legacy `{cacheDir}/{instance}/auth_tokens.json` with automatic migration
- Implementation: `internal/graph/oauth2_account_storage.go`
- Completion report: `docs/updates/2026-01-23-task-4-9-auth-token-storage-refactoring-complete.md`

### Issues Resolved
1. **Token duplication**: ✅ Fixed — one account = one token file via account-based storage
2. **Docker reliability**: ✅ Fixed — tokens found by account email, not mount point
3. **Migration**: ✅ Implemented — automatic migration from old locations with grace-period cleanup

## Implemented Solution

### Account-Based Token Storage

```
{cacheDir}/accounts/{account-hash}/auth_tokens.json
```

Where `account-hash` is the first 16 characters of the SHA256 hash of the normalized account email.

### Architecture Components

1. **Token Path Resolution** (implemented in `internal/graph/oauth2_account_storage.go`)
   - `GetAuthTokensPathByAccount(cacheDir, email string) string` - Get path by account
   - `hashAccount(email string) string` - Generate account hash
   - `FindAuthTokens(cacheDir, instance, email string) (string, error)` - Search multiple locations with auto-migration

2. **Token Migration** (implemented in `internal/graph/oauth2_account_storage.go`)
   - Automatic migration from old paths to new paths via `migrateTokens()`
   - Fallback to old locations for backward compatibility
   - Grace-period cleanup of deprecated files via `CleanupOldTokens()`
   - Migration logging for troubleshooting

3. **Authentication Entry Point** (implemented in `internal/graph/oauth2.go`)
   - `AuthenticateWithAccountStorage()` delegates to `FindAuthTokens()` for path resolution
   - `resolveAccountFromRegistry()` looks up account email from mount-account registry
   - `loadAndRefreshAuth()` loads and refreshes tokens from a resolved path

4. **Multi-Account Support** (implemented)
   - Separate token storage per account
   - Separate cache directories per account
   - Separate delta sync loops per account

### Security Design (Implemented)

1. **Encryption**: AES-256 for tokens at rest ✅
2. **File Permissions**: 0600 (owner read/write only) ✅
3. **TLS**: HTTPS/TLS 1.2+ for all API communication ✅
4. **Rate Limiting**: Prevent brute force attacks ✅

## Design Constraints (All Satisfied)

- ✅ Maintains backward compatibility with existing token locations
- ✅ Works in Docker test environment
- ✅ Supports multiple accounts without conflicts
- ✅ Secure and follows best practices

## Testing Strategy (Implemented)

1. **Unit Tests**: Token path generation, hashing, migration logic — `internal/graph/oauth2_account_storage_test.go`
2. **Integration Tests**: Migration, multi-account isolation, Docker token access — `internal/graph/oauth2_account_storage_integration_test.go`
3. **Property-Based Tests**: Token storage security, automatic refresh — `internal/graph/oauth2_property_test.go`
4. **Docker Tests**: Verified reliability in containerized environment

## Implementation Phases (All Complete)

See [tasks.md](tasks.md) for detailed task list. All phases delivered — see `docs/updates/2026-01-23-task-4-9-auth-token-storage-refactoring-complete.md`.

## References

- Analysis: `docs/reports/2026-01-23-063800-auth-token-storage-architecture-analysis.md`
- Plan: `docs/plans/auth-token-storage-refactoring-plan.md`
- Fix: `docs/fixes/mount-account-registry-authentication-fix.md`

## Monolith Design Sections (Moved Verbatim)

Moved from `.kiro/specs/archive/system-verification-and-fix/design.md`.

### 1. Authentication Component

**Location**: `internal/graph/oauth2*.go`, `internal/graph/authenticator.go`

**Verification Steps**:
1. Review OAuth2 implementation (both GTK and headless flows)
2. Test token storage and retrieval in Docker container
3. Test token refresh mechanism
4. Test authentication failure scenarios
5. Verify secure token storage

**Expected Interfaces**:
- `Auth` struct with `AccessToken`, `RefreshToken`, `ExpiresAt`
- `Authenticator` interface with `Authenticate()`, `Refresh()`, `GetAuth()` methods
- `RealAuthenticator` and `MockAuthenticator` implementations
- Token storage in `~/.config/onemount/auth_tokens.json` or test-artifacts directory

**Verification Criteria**:
- Authentication succeeds with valid credentials (Requirements 1.1, 1.2)
- Tokens are stored securely with appropriate permissions (Requirements 1.2, 22.1, 22.2)
- Token refresh works automatically before expiration (Requirements 1.3)
- Failed authentication provides clear error messages (Requirements 1.4)
- Headless mode uses device code flow correctly (Requirements 1.5)
- Tokens are stored using account-based paths for reliability and multi-account support (Requirements 1.6)
- Tests run in isolated Docker containers without affecting host system (Requirements 16.1-16.7)

#### Account-Based Token Storage Architecture

**Problem**: The original implementation stored tokens at `{cacheDir}/{instance}/auth_tokens.json` where `instance` is derived from the mount point location. This caused:
- Docker test reliability issues (different mount points = different token paths)
- Token duplication (same account at different mount points = duplicate tokens)
- Token loss on remount (changing mount point = can't find existing tokens)
- Test environment confusion (tests can't find production tokens)

**Solution**: Store tokens at `{cacheDir}/accounts/{account-hash}/auth_tokens.json` where:
- `cacheDir`: XDG cache directory (typically `~/.cache/onemount`)
- `account-hash`: SHA256 hash of account email (first 16 characters)
- Example: `user@example.com` → `~/.cache/onemount/accounts/a1b2c3d4e5f6g7h8/auth_tokens.json`

**Benefits**:
- **Mount Point Independence**: Same tokens regardless of where you mount
- **No Token Duplication**: One account = one token file
- **Reliable Docker Testing**: Tests find tokens regardless of mount point
- **Account Isolation**: Different accounts have separate token files
- **Multi-Account Support**: Multiple accounts can be mounted simultaneously

**Implementation**:

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
    
    // 2. Try instance-based location (old, for migration)
    instancePath := GetAuthTokensPath(cacheDir, instance)
    if fileExists(instancePath) {
        // Auto-migrate if we have account email
        if accountEmail != "" {
            migrateTokens(instancePath, GetAuthTokensPathByAccount(cacheDir, accountEmail))
        }
        return instancePath, nil
    }
    
    // 3. Try legacy location (oldest)
    legacyPath := GetAuthTokensPathFromCacheDir(cacheDir)
    if fileExists(legacyPath) {
        if accountEmail != "" {
            migrateTokens(legacyPath, GetAuthTokensPathByAccount(cacheDir, accountEmail))
        }
        return legacyPath, nil
    }
    
    // 4. Return new account-based path for creation
    if accountEmail != "" {
        return GetAuthTokensPathByAccount(cacheDir, accountEmail), nil
    }
    
    return legacyPath, nil
}
```

**Migration Strategy**:
- Automatic migration on first use (copy tokens from old to new location)
- Keep old tokens as backup (don't delete immediately)
- Backward compatibility maintained (old locations still work)
- Gradual deprecation timeline (log warnings → eventual removal)

**Security Considerations**:
- Token file permissions remain `0600` (owner read/write only)
- SHA256 hash is cryptographically secure and collision-resistant
- Account email not visible in filesystem (privacy-preserving)
- Hash length (16 chars = 64 bits) provides sufficient uniqueness

**Testing Requirements**:
- Unit tests for hash generation and path generation
- Integration tests for token migration from old locations
- System tests for Docker environment token access
- Multi-account tests for token isolation

### 14. Multi-Account Mount Manager Component

**Location**: `cmd/onemount/main.go`, `internal/ui/onemount.go`

**Verification Steps**:
1. Review mount management code
2. Test mounting multiple accounts simultaneously
3. Test isolation between mounts (auth, cache, sync)
4. Test personal OneDrive mount
5. Test OneDrive for Business mount
6. Test shared drive mount
7. Test "Shared with me" access

**Expected Interfaces**:
- `MountManager` struct tracking active mounts
- `Mount(accountType, mountPoint, auth)` method
- `Unmount(mountPoint)` method
- `ListMounts()` method
- Separate `Filesystem` instance per mount
- Separate cache directory per mount
- Separate delta sync loop per mount

**Verification Criteria**:
- Multiple accounts can be mounted simultaneously
- Each mount has isolated authentication
- Each mount has isolated cache
- Each mount has independent delta sync
- Personal OneDrive accessible via `/me/drive`
- Business OneDrive accessible via `/me/drive`
- Shared drives accessible via `/drives/{drive-id}`
- "Shared with me" accessible via `/me/drive/sharedWithMe`
- No cross-contamination between mounts

### Security Properties

**Property 43: Token Encryption at Rest**
*For any* authentication token storage operation, the system should encrypt tokens using AES-256 encryption
**Validates: Requirements 22.1**

**Property 44: Token File Permissions**
*For any* token storage file creation, the system should set file permissions to 0600 (owner read/write only)
**Validates: Requirements 22.2**

**Property 45: Secure Token Storage Location**
*For any* authentication token storage, the system should store tokens in the XDG configuration directory with restricted access
**Validates: Requirements 22.3**

**Property 46: HTTPS/TLS Communication**
*For any* Microsoft Graph API communication, the system should use HTTPS/TLS 1.2 or higher for all connections
**Validates: Requirements 22.4**

**Property 47: Sensitive Data Logging Prevention**
*For any* logging operation, the system should never log authentication tokens, passwords, or sensitive user data
**Validates: Requirements 22.6**

**Property 48: Cache File Security**
*For any* cached file content storage, the system should set appropriate file permissions to prevent unauthorized access
**Validates: Requirements 22.8**

