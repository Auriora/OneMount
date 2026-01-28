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
