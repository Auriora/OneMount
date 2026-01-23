# Changelog

All notable changes to the OneMount project will be documented in this file.

## [Unreleased]

### Changed
- **BREAKING**: Authentication tokens now use account-based storage instead of mount-point-based storage
  - Tokens stored at `~/.cache/onemount/accounts/{account-hash}/auth_tokens.json`
  - Provides mount point independence and eliminates token duplication
  - Automatic migration from old locations (`~/.cache/onemount/{instance}/auth_tokens.json`)
  - Old tokens preserved for safety during migration
  - See `docs/guides/developer/authentication-token-paths-v2.md` for details

### Added
- `GetAuthTokensPathByAccount()` - Generate account-based token path
- `hashAccount()` - Create stable SHA256 hash of account email
- `FindAuthTokens()` - Search for tokens with automatic migration
- `migrateTokens()` - Copy tokens from old to new location
- `AuthenticateWithAccountStorage()` - Authenticate using account-based storage
- Comprehensive unit tests for account-based storage (100% pass rate)
- Documentation: `docs/guides/developer/authentication-token-paths-v2.md`
- Investigation report: `docs/reports/2026-01-23-task-4-9-1-investigation-findings.md`

### Deprecated
- `GetAuthTokensPath()` - Use `GetAuthTokensPathByAccount()` instead (still supported for backward compatibility)
- `GetAuthTokensPathFromCacheDir()` - Use `FindAuthTokens()` instead (still supported for backward compatibility)
- Instance-based token storage - Automatically migrated to account-based storage

### Fixed
- Docker test reliability issues caused by mount-point-dependent token paths
- Token duplication when same account mounted at different locations
- Token loss when mount point location changes
- Confusion between test and production token locations

