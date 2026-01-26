# Changelog

All notable changes to the OneMount project will be documented in this file.

## [0.1.0] - 2026-01-26

**First Public Release**

### Major Features

#### Realtime Synchronization with Socket.IO
- **Socket.IO-based realtime notifications** - Instant change detection using Microsoft Graph Socket.IO subscriptions
- **Automatic fallback to polling** - Seamless degradation when Socket.IO unavailable
- **Configurable polling intervals** - Adaptive polling based on connection health (5-30 minutes)
- **Health monitoring** - Real-time subscription status tracking via `onemount --stats`
- **Three notification modes** - Socket.IO (recommended), polling-only, and disabled modes

#### ETag-Based Cache Validation
- **Proactive cache invalidation** - Delta sync detects remote changes and invalidates stale cache entries
- **Efficient bandwidth usage** - Only downloads files when ETags indicate changes
- **QuickXORHash verification** - Content integrity checking for downloaded files
- **Batch metadata updates** - Efficient processing of multiple file changes
- **No conditional GET required** - Works with Microsoft Graph pre-authenticated download URLs

#### XDG Base Directory Compliance
- **Standard configuration paths** - Follows XDG Base Directory specification
- **Automatic directory creation** - Creates `~/.config/onemount/` and `~/.cache/onemount/` as needed
- **Environment variable support** - Respects `XDG_CONFIG_HOME` and `XDG_CACHE_HOME`
- **Virtual file management** - `.xdg-volume-info` files served locally without OneDrive sync
- **Custom path override** - Command-line flags for non-standard locations

#### Account-Based Token Storage
- **Mount point independence** - Tokens stored by account identity (email hash) instead of mount point
- **Automatic migration** - Seamless migration from old token locations
- **Multi-account support** - Separate token storage for each OneDrive account
- **Docker test reliability** - Eliminates token path issues in containerized environments

#### Metadata State Machine
- **Explicit state tracking** - Clear state transitions (GHOST → HYDRATING → HYDRATED → DIRTY_LOCAL)
- **Conflict detection** - Automatic detection of local and remote changes
- **Error recovery** - Retry mechanisms for failed hydration and upload operations
- **Virtual file support** - Local-only files with `local-*` identifiers

#### Offline Mode with Conflict Resolution
- **Full read-write offline** - Complete file operations without network connectivity
- **Intelligent conflict detection** - ETag-based comparison when reconnecting
- **Multiple resolution strategies** - Last-writer-wins, keep-both, user-choice, merge, rename
- **Queued synchronization** - Automatic upload of offline changes when online
- **Persistent change tracking** - Offline modifications stored in metadata database

### Added

#### Core Features
- Socket.IO subscription manager for realtime notifications
- Engine.IO v4 WebSocket transport implementation
- ETag-based cache validation via delta sync
- XDG Base Directory compliance for config and cache
- Account-based authentication token storage
- Metadata state machine with explicit state transitions
- Offline mode with comprehensive conflict resolution
- Virtual file management for `.xdg-volume-info`
- Overlay policy system (LOCAL_WINS, REMOTE_WINS, MERGED)

#### Configuration Options
- `realtime.enabled` - Enable/disable realtime notifications
- `realtime.pollingOnly` - Force polling-only mode
- `realtime.fallbackIntervalSeconds` - Polling interval (30-7200 seconds)
- `hydration.workers` - Concurrent download workers (1-64, default 4)
- `hydration.queueSize` - Download queue size (1-100000, default 500)
- `metadataQueue.workers` - Metadata request workers (1-64, default 3)
- `metadataQueue.highPrioritySize` - High-priority queue size (default 100)
- `metadataQueue.lowPrioritySize` - Low-priority queue size (default 1000)
- `overlay.defaultPolicy` - Virtual file overlay policy

#### Command-Line Options
- `--polling-only` - Force polling-only mode
- `--realtime-fallback-seconds N` - Set polling interval
- `--hydration-workers N` - Set download worker count
- `--hydration-queue-size N` - Set download queue size
- `--metadata-workers N` - Set metadata worker count
- `--metadata-high-queue-size N` - Set high-priority queue size
- `--metadata-low-queue-size N` - Set low-priority queue size
- `--overlay-policy POLICY` - Set overlay policy
- `--stats` - Display comprehensive statistics

#### API Functions
- `GetAuthTokensPathByAccount()` - Generate account-based token path
- `hashAccount()` - Create stable SHA256 hash of account email
- `FindAuthTokens()` - Search for tokens with fallback to legacy locations
- `AuthenticateWithAccountStorage()` - Authenticate using account-based storage

#### Testing Infrastructure
- Docker-based test environment with FUSE support
- Comprehensive unit tests (100% pass rate)
- Integration tests with real OneDrive accounts
- Property-based tests for correctness verification
- System tests for end-to-end workflows
- Timeout protection for hanging tests



### Fixed

#### Authentication
- Account-based token storage using email hash for stable identification
- Token storage independent of mount point location
- Support for multiple OneDrive accounts with separate token storage

#### Cache Management
- ETag-based cache validation through delta sync
- Automatic cache invalidation when remote files change
- Efficient bandwidth usage with selective downloads
- XDG-compliant cache directory structure

#### Synchronization
- Socket.IO-based realtime change notifications
- Automatic fallback to polling when Socket.IO unavailable
- Adaptive polling intervals based on connection health
- Comprehensive conflict detection for simultaneous changes

#### File Operations
- Virtual files (`.xdg-volume-info`) served locally without OneDrive sync
- Scoped cache invalidation for efficient metadata management
- FUSE operations served from local metadata/cache only
- Prioritized metadata requests for user-facing operations

### Deprecated

None - this is the first release.

### Known Issues

#### Resource Management (Deferred to v1.0.1)
- **Property 56: Cache Size Enforcement** - Cache may exceed configured size limits
  - Workaround: Manually run `onemount --cleanup` to enforce limits
  - Fix planned: Implement LRU eviction with size tracking
- **Property 58: Worker Thread Limits** - Worker goroutines may not fully cleanup
  - Workaround: Restart OneMount if memory usage grows
  - Fix planned: Improve worker pool synchronization
- **Property 59: Adaptive Network Throttling** - Bandwidth throttling not fully adaptive
  - Workaround: Configure fixed bandwidth limits in config
  - Fix planned: Implement adaptive throttling based on network conditions

#### Limitations
- **Large files** - Files loaded into memory; not suitable for multi-gigabyte files
- **Symbolic links** - Not supported by Microsoft OneDrive API
- **OneDrive Recycle Bin** - Not exposed by Microsoft Graph API; use web UI
- **File browser thumbnails** - May trigger automatic downloads for thumbnail generation

### Security

- Authentication tokens encrypted at rest using AES-256
- Token files created with 0600 permissions (owner read/write only)
- All Microsoft Graph API communication uses HTTPS/TLS 1.2+
- Certificate chain validation enforced
- No sensitive data logged (tokens, passwords, user data)
- Rate limiting on authentication failures
- Secure deletion of temporary authentication data

### Performance

- Directory listing: < 2 seconds for 1000 files
- Cached file access: < 100 milliseconds
- Idle memory usage: < 50 MB
- Active sync memory usage: < 200 MB
- Concurrent operations: 10+ simultaneous without degradation
- Startup time: < 5 seconds
- Shutdown time: < 10 seconds
- Delta sync: 1000 changed files in < 30 seconds

### Documentation

- [v0.1.0 Feature Guide](docs/guides/user/v0.1.0-features.md)
- [Socket.IO Configuration Guide](docs/guides/socketio-configuration.md)
- [Offline Functionality Guide](docs/3-implementation/offline-functionality.md)
- [Authentication Architecture](docs/guides/developer/authentication-token-paths-v2.md)
- [Running Tests Guide](docs/testing/running-tests.md)
- [Docker Test Environment](docs/testing/docker-test-environment.md)
- [Troubleshooting Guide](docs/guides/user/troubleshooting-guide.md)

