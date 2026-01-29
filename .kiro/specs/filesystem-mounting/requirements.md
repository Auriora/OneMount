# Requirements Document: Filesystem Mounting

## Source
Moved from `.kiro/specs/archive/system-verification-and-fix/requirements.md`. Sections below are verbatim.

## Requirements

### Requirement 2: Basic Filesystem Mounting

**User Story:** As a Linux user, I want to mount my OneDrive as a local directory so that I can access files using standard file operations.

#### Acceptance Criteria

1. WHEN the user specifies a mount point, THE OneMount System SHALL mount OneDrive at that location using FUSE
2. WHEN the filesystem is mounted, THE OneMount System SHALL display the root directory contents
3. WHILE the filesystem is mounted, THE OneMount System SHALL respond to standard file operations (ls, cat, cp, etc.)
4. IF the mount point is already in use, THEN THE OneMount System SHALL display an error message with the conflicting process
5. WHEN the user unmounts the filesystem, THE OneMount System SHALL cleanly release all resources

### Requirement 2C: Advanced Mounting Options

**User Story:** As a user, I want advanced mounting options so that I can use OneMount in different scenarios and configurations.

#### Acceptance Criteria

1. WHERE the user specifies daemon mode, THE OneMount System SHALL fork the process and detach from the terminal for background operation
2. WHEN the user specifies a mount timeout, THE OneMount System SHALL wait up to the specified duration for the mount operation to complete
3. IF the mount timeout is not specified, THEN THE OneMount System SHALL use a default timeout of 60 seconds
4. WHEN opening the metadata database, THE OneMount System SHALL detect stale lock files older than 5 minutes and attempt to remove them
5. IF a database lock file is detected and is not stale, THEN THE OneMount System SHALL retry with exponential backoff up to 10 attempts

### Requirement 15: XDG Base Directory Compliance

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
11. THE OneMount System SHALL create `.xdg-volume-info` files as local-only virtual files that are NOT synced to OneDrive
12. WHEN creating `.xdg-volume-info` files, THE OneMount System SHALL assign them a local-only ID (prefixed with "local-")
13. WHEN accessing `.xdg-volume-info` files, THE OneMount System SHALL serve content from the local cache without attempting to sync to OneDrive
