# Authentication and Account Management Architecture

**Last Updated**: 2026-02-13  
**Status**: Active  
**Related Specs**: `.kiro/specs/authentication-and-accounts/`

## Overview

OneMount implements secure authentication with Microsoft accounts using OAuth2, supporting multiple accounts simultaneously with account-based token storage. This document describes the authentication architecture, security measures, and multi-account support.

## Table of Contents

1. [Authentication Flow](#authentication-flow)
2. [Account-Based Token Storage](#account-based-token-storage)
3. [Token Security](#token-security)
4. [Multi-Account Support](#multi-account-support)
5. [Mount-Account Registry](#mount-account-registry)
6. [Migration Strategy](#migration-strategy)
7. [Security Properties](#security-properties)

## Authentication Flow

### OAuth2 Implementation

OneMount uses OAuth2 for authentication with Microsoft Graph API. The implementation supports two modes:

1. **Interactive Mode (GTK)**: Opens a browser window for authentication
2. **Headless Mode**: Uses device code flow for systems without a GUI

### Authentication Sequence

```
User → OneMount → Microsoft OAuth2 → Access Token + Refresh Token → Secure Storage
```

For detailed sequence diagrams, see [auth-sequence-diagram.puml](resources/auth-sequence-diagram.puml).

### Code Structure

- `internal/graph/oauth2.go` - Core OAuth2 implementation
- `internal/graph/oauth2_gtk.go` - GTK-based interactive authentication
- `internal/graph/oauth2_headless.go` - Headless device code flow
- `internal/graph/authenticator.go` - Authenticator interface and implementations

### Authentication Process

1. **Check for existing tokens** at account-based storage location
2. **If tokens exist**:
   - Load tokens from file
   - Check expiration
   - Refresh if expired
   - Save refreshed tokens
3. **If tokens don't exist or refresh fails**:
   - Initiate OAuth2 flow (interactive or headless)
   - User authenticates with Microsoft
   - Receive access token and refresh token
   - Save tokens to account-based storage with 0600 permissions

### Token Refresh

- Access tokens expire after ~1 hour
- Refresh tokens are long-lived (typically 90 days)
- Automatic refresh occurs before API requests when token is expired
- If refresh fails, user must re-authenticate

## Account-Based Token Storage

### Architecture

Tokens are stored using account identity (email hash) rather than mount point location:

```
{cacheDir}/accounts/{account-hash}/auth_tokens.json
```

Where:
- `cacheDir`: XDG cache directory (typically `~/.cache/onemount`)
- `account-hash`: First 16 characters of SHA256 hash of account email
- Example: `user@example.com` → `~/.cache/onemount/accounts/a1b2c3d4e5f6g7h8/auth_tokens.json`

### Benefits

1. **Mount Point Independence**: Same tokens regardless of where you mount
2. **No Token Duplication**: One account = one token file
3. **Reliable Docker Testing**: Tests find tokens regardless of mount point
4. **Account Isolation**: Different accounts have separate token files
5. **Multi-Account Support**: Multiple accounts can be mounted simultaneously

### Implementation

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
```

### Token Lookup with Fallback

The system searches multiple locations in priority order:

1. **Account-based location** (new): `{cacheDir}/accounts/{account-hash}/auth_tokens.json`
2. **Instance-based location** (old): `{cacheDir}/{instance}/auth_tokens.json`
3. **Legacy location** (oldest): `{cacheDir}/auth_tokens.json`

If tokens are found in a fallback location and account email is known, automatic migration occurs.

## Token Security

### Encryption at Rest

- Tokens are encrypted using AES-256 encryption
- Encryption key derived from system-specific data
- Implemented in `internal/graph/oauth2.go`

### File Permissions

- Token files created with `0600` permissions (owner read/write only)
- Prevents unauthorized access by other users
- Verified by property-based test `TestProperty1_OAuth2TokenStorageSecurity`

### Storage Location

- Tokens stored in XDG cache directory with restricted access
- Default: `~/.cache/onemount/accounts/{account-hash}/`
- Can be overridden with `--cache-dir` flag

### Transport Security

- All Microsoft Graph API communications use HTTPS/TLS 1.2 or higher
- Certificate validation enforced
- TLS errors properly handled and logged
- Verified by integration tests `TestUT_GR_ERR_07_*`

### Logging Security

- Authentication tokens never logged
- Sensitive data redacted in error messages
- Verified by property-based test `TestProperty47_SensitiveDataLoggingPrevention`

### Rate Limiting

- Authentication failures trigger rate limiting
- Prevents brute force attacks
- Exponential backoff on repeated failures

## Multi-Account Support

### Architecture

OneMount supports mounting multiple OneDrive accounts simultaneously:

- Personal OneDrive accounts
- OneDrive for Business accounts
- Shared drives
- "Shared with me" items

### Account Isolation

Each mounted account has:

1. **Separate authentication tokens**: Stored in account-specific directory
2. **Separate cache directory**: Metadata and content cache isolated per account
3. **Separate delta sync loop**: Independent change tracking per account
4. **Separate FUSE filesystem**: Each mount point has its own filesystem instance

### Drive Types

```go
// Personal OneDrive
/me/drive

// OneDrive for Business
/me/drive

// Shared drives
/drives/{drive-id}

// Shared with me
/me/drive/sharedWithMe
```

### Concurrent Mounts

Multiple accounts can be mounted simultaneously at different mount points:

```bash
# Personal account
onemount ~/OneDrive-Personal

# Work account
onemount ~/OneDrive-Work

# Shared drive
onemount --drive-id abc123 ~/OneDrive-Shared
```

Each mount maintains independent state and authentication.

## Mount-Account Registry

### Purpose

The mount registry maps mount points to account identities, enabling:

- Persistent mount-account associations
- Automatic token lookup by account
- Multi-account management
- Launcher integration

### Registry Location

```
~/.config/onemount/mounts.json
```

### Registry Format

```json
{
  "mounts": {
    "/home/user/OneDrive-Personal": {
      "account": "user@example.com",
      "label": "Personal OneDrive",
      "created": "2026-02-13T10:00:00Z",
      "last_mounted": "2026-02-13T10:00:00Z"
    },
    "/home/user/OneDrive-Work": {
      "account": "user@company.com",
      "label": "Work OneDrive",
      "created": "2026-02-13T10:05:00Z",
      "last_mounted": "2026-02-13T10:05:00Z"
    }
  }
}
```

### Registry Operations

1. **Register mount**: Add mount point and account mapping
2. **Lookup account**: Find account for a mount point
3. **Update last mounted**: Track mount usage
4. **Remove mount**: Clean up unmounted entries
5. **List mounts**: Display all registered mounts

### Launcher Integration

The launcher uses the registry to:

- Display account names for each mount point
- Check authentication status before mounting
- Provide user-friendly mount management
- Show optional labels for better identification

## Migration Strategy

### Automatic Migration

When tokens are found in old locations:

1. **Detect old token file** at instance-based or legacy location
2. **Extract account identity** from token file
3. **Create account-based directory** using account hash
4. **Copy token file** to new location
5. **Update mount registry** with mount point and account mapping
6. **Log migration details** for troubleshooting
7. **Keep old tokens** as backup (not deleted immediately)

### Backward Compatibility

During migration period:

- System reads tokens from both old and new locations
- Fallback to old locations if new location not found
- Warnings logged for deprecated storage locations
- Gradual deprecation timeline (warnings → eventual removal)

### Migration Code

```go
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

## Security Properties

### Property-Based Testing

The authentication system is validated using property-based tests:

**Property 43: Token Encryption at Rest**  
*For any* authentication token storage operation, the system encrypts tokens using AES-256 encryption.  
**Validates**: Requirements 2.1

**Property 44: Token File Permissions**  
*For any* token storage file creation, the system sets file permissions to 0600 (owner read/write only).  
**Validates**: Requirements 2.2

**Property 45: Secure Token Storage Location**  
*For any* authentication token storage, the system stores tokens in the XDG configuration directory with restricted access.  
**Validates**: Requirements 2.3

**Property 46: HTTPS/TLS Communication**  
*For any* Microsoft Graph API communication, the system uses HTTPS/TLS 1.2 or higher for all connections.  
**Validates**: Requirements 2.4

**Property 47: Sensitive Data Logging Prevention**  
*For any* logging operation, the system never logs authentication tokens, passwords, or sensitive user data.  
**Validates**: Requirements 2.6

**Property 48: Cache File Security**  
*For any* cached file content storage, the system sets appropriate file permissions to prevent unauthorized access.  
**Validates**: Requirements 2.8

### Security Testing

See [Security Testing Guide](../4-testing/guides/frameworks/security-testing-guide.md) for comprehensive security testing approach.

## References

### Design Documents

- [Mount-Account Registry Design](../designs/mount-account-registry.md)
- [Account Storage Migration](../issues/account-storage-incomplete-migration.md)

### Implementation

- `internal/graph/oauth2.go` - Core OAuth2 and token storage
- `internal/graph/authenticator.go` - Authenticator interface
- `cmd/onemount/main.go` - Main application auth flow

### Testing

- `internal/graph/oauth2_test.go` - Unit tests
- `internal/graph/oauth2_integration_test.go` - Integration tests
- `internal/graph/oauth2_property_test.go` - Property-based tests

### Documentation

- [Authentication Token Paths](../guides/developer/authentication-token-paths.md)
- [Security Testing Guide](../4-testing/guides/frameworks/security-testing-guide.md)
- [Test Setup](../4-testing/TEST_SETUP.md)

### Reports

- [Auth Token Storage Architecture Analysis](../reports/2026-01-23-063800-auth-token-storage-architecture-analysis.md)
- [Auth Token Storage Refactoring Plan](../plans/auth-token-storage-refactoring-plan.md)
