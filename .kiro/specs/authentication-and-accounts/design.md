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
- Token storage currently uses mount-point-based paths

### Token Storage
- Tokens stored in `{cacheDir}/{instance}/auth_tokens.json`
- Instance derived from mount point, causing Docker test issues
- No account-based storage mechanism

### Issues Identified
1. **Token duplication**: Same account mounted at different points creates duplicate tokens
2. **Docker reliability**: Mount-point-based paths cause test failures
3. **Migration needed**: Existing tokens need to be migrated to new structure

## Proposed Solution Design

### Account-Based Token Storage

```
{cacheDir}/accounts/{account-hash}/auth_tokens.json
```

Where `account-hash` is SHA256 hash of account email.

### Architecture Components

1. **Token Path Resolution**
   - `GetAuthTokensPathByAccount(email string) string` - Get path by account
   - `hashAccount(email string) string` - Generate account hash
   - `FindAuthTokens(email string) (string, error)` - Search multiple locations

2. **Token Migration**
   - Automatic migration from old paths to new paths
   - Fallback to old locations for backward compatibility
   - Migration logging for troubleshooting

3. **Multi-Account Support**
   - Separate token storage per account
   - Separate cache directories per account
   - Separate delta sync loops per account

### Security Design

1. **Encryption**: AES-256 for tokens at rest
2. **File Permissions**: 0600 (owner read/write only)
3. **TLS**: HTTPS/TLS 1.2+ for all API communication
4. **Rate Limiting**: Prevent brute force attacks

## Design Constraints

- Must maintain backward compatibility with existing token locations
- Must work in Docker test environment
- Must support multiple accounts without conflicts
- Must be secure and follow best practices

## Testing Strategy

1. **Unit Tests**: Token path generation, hashing, migration logic
2. **Integration Tests**: OAuth2 flow, token refresh, multi-account
3. **Property-Based Tests**: Token storage security, automatic refresh
4. **Docker Tests**: Verify reliability in containerized environment

## Implementation Phases

See [tasks.md](tasks.md) for detailed implementation plan.

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

