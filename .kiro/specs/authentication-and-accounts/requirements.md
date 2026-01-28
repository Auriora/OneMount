# Requirements Document: Authentication and Account Management

## Introduction

This specification defines the requirements for authentication and multi-account support in OneMount. The system must securely authenticate users with their Microsoft accounts, manage authentication tokens, and support mounting multiple OneDrive accounts simultaneously. This spec focuses on the authentication flow, token management, account-based storage, and multi-account mounting capabilities.

## Glossary

- **OneMount System**: The complete OneDrive filesystem client for Linux including FUSE filesystem, Graph API integration, caching, and UI components
- **OAuth2**: Open Authorization 2.0 protocol used for secure authentication with Microsoft accounts
- **Authentication Token**: Access token and refresh token pair used to authenticate API requests to Microsoft Graph
- **Refresh Token**: Long-lived token used to obtain new access tokens when they expire
- **Access Token**: Short-lived token used to authenticate individual API requests
- **Account Hash**: SHA-256 hash of the account email address used as a directory name for account-based storage
- **Account-Based Storage**: Storage system that organizes tokens and metadata by account identity rather than mount point location
- **Mount Point**: Directory path where a OneDrive account is mounted in the Linux filesystem
- **Mount Registry**: Configuration file mapping mount points to account identities
- **Device Code Flow**: OAuth2 authentication flow for headless systems without a web browser
- **Microsoft Graph API**: Microsoft's REST API for accessing OneDrive and other Microsoft 365 resources
- **Personal OneDrive**: OneDrive account associated with a personal Microsoft account
- **OneDrive for Business**: OneDrive account associated with a work or school Microsoft 365 account
- **Shared Drive**: OneDrive drive shared with the user by another account
- **XDG Base Directory**: Linux standard for organizing user configuration and cache files
- **Token Encryption**: AES-256 encryption applied to authentication tokens at rest
- **Token Refresh**: Process of obtaining a new access token using a refresh token

## Requirements

### Requirement 1: Authentication Verification

**User Story:** As a Linux user, I want to authenticate with my Microsoft account so that I can access my OneDrive files.

#### Acceptance Criteria

1. WHEN the user launches OneMount for the first time, THE OneMount System SHALL display an authentication dialog
2. WHEN the user completes Microsoft OAuth2 authentication, THE OneMount System SHALL store authentication tokens securely
3. WHEN authentication tokens expire, THE OneMount System SHALL automatically refresh them using the refresh token
4. IF token refresh fails, THEN THE OneMount System SHALL prompt the user to re-authenticate
5. WHERE the system is running in headless mode, THE OneMount System SHALL use device code flow for authentication
6. WHEN storing authentication tokens, THE OneMount System SHALL use account-based storage paths derived from account identity (email hash) rather than mount point location, ensuring tokens are accessible regardless of mount point changes and preventing token duplication across multiple mounts of the same account

### Requirement 2: Token Security

**User Story:** As a security-conscious user, I want my authentication tokens to be protected from unauthorized access so that my OneDrive account remains secure.

#### Acceptance Criteria

1. WHEN storing authentication tokens, THE OneMount System SHALL encrypt tokens at rest using AES-256 encryption
2. WHEN creating token storage files, THE OneMount System SHALL set file permissions to 0600 (owner read/write only)
3. WHEN storing authentication tokens, THE OneMount System SHALL store them in the XDG configuration directory with restricted access
4. WHEN communicating with Microsoft Graph API, THE OneMount System SHALL use HTTPS/TLS 1.2 or higher for all connections
5. WHEN validating TLS certificates, THE OneMount System SHALL verify certificate chains and reject invalid certificates
6. WHEN logging operations, THE OneMount System SHALL never log authentication tokens, passwords, or sensitive user data
7. WHEN handling authentication failures, THE OneMount System SHALL implement rate limiting to prevent brute force attacks

### Requirement 3: Mount-Account Registry

**User Story:** As a user with multiple mount points, I want the system to remember which account is associated with each mount point so that I don't need to re-authenticate every time.

#### Acceptance Criteria

1. THE OneMount System SHALL maintain a mount registry file mapping mount points to account identities
2. WHEN a new mount point is created, THE OneMount System SHALL register the mount point and associated account in the registry
3. WHEN the user mounts an existing mount point, THE OneMount System SHALL look up the associated account from the registry
4. WHEN the registry contains a mount point entry, THE OneMount System SHALL use the account identity to locate authentication tokens
5. IF the registry contains a mount point but tokens are missing, THEN THE OneMount System SHALL prompt for re-authentication
6. WHEN a mount point is removed, THE OneMount System SHALL remove the mount point entry from the registry
7. THE OneMount System SHALL store the mount registry in the XDG configuration directory at `~/.config/onemount/mounts.json`
8. WHEN the mount registry file is corrupt or missing, THE OneMount System SHALL start with an empty registry and log a warning

### Requirement 4: Multiple Account and Drive Support

**User Story:** As a user with multiple OneDrive accounts, I want to mount my personal OneDrive, work OneDrive, and shared drives simultaneously so that I can access all my files.

#### Acceptance Criteria

1. THE OneMount System SHALL support mounting multiple OneDrive accounts simultaneously at different mount points
2. WHEN mounting a personal OneDrive account, THE OneMount System SHALL access the user's personal drive using `/me/drive`
3. WHEN mounting a OneDrive for Business account, THE OneMount System SHALL access the user's work drive using `/me/drive`
4. THE OneMount System SHALL support mounting shared drives using `/drives/{drive-id}`
5. THE OneMount System SHALL support accessing "Shared with me" items using `/me/drive/sharedWithMe`
6. WHEN multiple accounts are mounted, THE OneMount System SHALL maintain separate authentication tokens for each account
7. WHEN multiple accounts are mounted, THE OneMount System SHALL maintain separate caches for each account
8. WHEN multiple accounts are mounted, THE OneMount System SHALL maintain separate delta sync loops for each account

### Requirement 5: XDG Base Directory Compliance

**User Story:** As a Linux user, I want OneMount to follow XDG Base Directory standards so that my configuration and cache files are stored in standard locations.

#### Acceptance Criteria

1. THE OneMount System SHALL use `os.UserConfigDir()` to determine the configuration directory
2. WHEN `XDG_CONFIG_HOME` is set, THE OneMount System SHALL store configuration in `$XDG_CONFIG_HOME/onemount/`
3. WHEN `XDG_CONFIG_HOME` is not set, THE OneMount System SHALL store configuration in `$HOME/.config/onemount/`
4. THE OneMount System SHALL use `os.UserCacheDir()` to determine the cache directory
5. WHEN `XDG_CACHE_HOME` is set, THE OneMount System SHALL store cache in `$XDG_CACHE_HOME/onemount/`
6. WHEN `XDG_CACHE_HOME` is not set, THE OneMount System SHALL store cache in `$HOME/.cache/onemount/`
7. THE OneMount System SHALL store authentication tokens in the configuration directory
8. THE OneMount System SHALL store file content cache in the cache directory
9. THE OneMount System SHALL store metadata database (bbolt) in the cache directory
10. WHERE the user specifies custom paths via command-line flags, THE OneMount System SHALL use the specified paths instead of XDG defaults

### Requirement 6: Account-Based Storage Migration

**User Story:** As a user upgrading from an older version, I want my existing authentication tokens to be migrated to the new account-based storage system so that I don't need to re-authenticate.

#### Acceptance Criteria

1. WHEN the system detects old instance-based token files, THE OneMount System SHALL migrate them to account-based storage
2. WHEN migrating tokens, THE OneMount System SHALL extract the account identity from the token file
3. WHEN migrating tokens, THE OneMount System SHALL create the account-based storage directory using the account hash
4. WHEN migrating tokens, THE OneMount System SHALL copy the token file to the new location
5. WHEN migrating tokens, THE OneMount System SHALL update the mount registry with the mount point and account mapping
6. WHEN migration completes successfully, THE OneMount System SHALL log the migration details
7. IF migration fails, THEN THE OneMount System SHALL log the error and continue with the old token location
8. THE OneMount System SHALL support reading tokens from both old and new locations during the migration period

### Requirement 7: Token Lookup and Discovery

**User Story:** As a developer, I want a reliable way to find authentication tokens regardless of storage location so that the system works during migration periods.

#### Acceptance Criteria

1. WHEN looking up tokens for a mount point, THE OneMount System SHALL first check the mount registry for the associated account
2. IF the mount registry contains an account, THEN THE OneMount System SHALL look for tokens in the account-based storage location
3. IF tokens are not found in account-based storage, THEN THE OneMount System SHALL fall back to instance-based storage
4. IF tokens are not found in instance-based storage, THEN THE OneMount System SHALL fall back to legacy storage locations
5. WHEN tokens are found in a fallback location, THE OneMount System SHALL log a warning about the deprecated storage location
6. THE OneMount System SHALL provide a function to search all token storage locations in priority order
7. WHEN multiple token files exist for the same account, THE OneMount System SHALL use the most recently modified token file

### Requirement 8: Launcher Integration

**User Story:** As a user, I want the launcher to display my account names for each mount point so that I can easily identify which OneDrive account is mounted where.

#### Acceptance Criteria

1. WHEN the launcher displays mount points, THE OneMount System SHALL look up the associated account from the mount registry
2. IF a mount point has an associated account, THEN THE OneMount System SHALL display the account email address
3. IF a mount point does not have an associated account, THEN THE OneMount System SHALL display "Not configured"
4. WHEN the user clicks on a mount point, THE OneMount System SHALL check if authentication tokens exist for the associated account
5. IF tokens exist and are valid, THEN THE OneMount System SHALL mount the drive
6. IF tokens are missing or invalid, THEN THE OneMount System SHALL prompt for authentication before mounting
7. THE OneMount System SHALL allow the user to add optional labels to mount points in the registry for better identification

## Dependencies

### Internal Dependencies

- **Cache Management Spec**: Account-based storage requires separate cache directories per account
- **Lazy Directory Loading Spec**: Authentication must complete before directory structure can be loaded
- **System Verification Spec**: Authentication is a prerequisite for all filesystem operations

### External Dependencies

- **Microsoft Graph API**: OAuth2 authentication endpoints
- **XDG Base Directory Specification**: Standard for configuration and cache file locations
- **systemd**: Optional integration for managing mount units

## References

### SRS Requirements

- **FR-AUTH-001**: OAuth 2.0 authentication (docs/1-requirements/srs/3-specific-requirements.md)
- **FR-AUTH-002**: Secure token storage (docs/1-requirements/srs/3-specific-requirements.md)
- **FR-AUTH-003**: Automatic token refresh (docs/1-requirements/srs/3-specific-requirements.md)
- **FR-AUTH-004**: Re-authentication support (docs/1-requirements/srs/3-specific-requirements.md)
- **NFR-SEC-001**: Token file permissions (docs/1-requirements/srs/3-specific-requirements.md)
- **NFR-SEC-002**: HTTPS for API communications (docs/1-requirements/srs/3-specific-requirements.md)
- **NFR-SEC-003**: Token exposure prevention (docs/1-requirements/srs/3-specific-requirements.md)

### Design Documents

- **Mount-Account Registry Design**: docs/designs/mount-account-registry.md
- **Account Storage Migration**: docs/issues/account-storage-incomplete-migration.md

### ADRs

- None currently - this spec may result in ADRs for authentication architecture decisions
