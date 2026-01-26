# OneMount v0.1.0 Release Notes

**Release Date**: January 26, 2026

We're excited to announce the first public release of OneMount v0.1.0, a native Linux filesystem client for Microsoft OneDrive with realtime synchronization, intelligent cache management, and comprehensive offline functionality.

---

## 🎉 Highlights

### Realtime Synchronization with Socket.IO

OneMount now uses Microsoft Graph Socket.IO subscriptions for instant change notifications. When files change in OneDrive, you'll see updates within seconds instead of waiting for polling intervals.

**Key Benefits:**
- **Instant Updates**: Changes appear immediately without manual refresh
- **Reduced Network Usage**: Less frequent polling when Socket.IO is healthy
- **Better Battery Life**: Fewer network requests mean longer battery life
- **No Webhooks Required**: No need for publicly accessible endpoints

**How to Use:**
```yaml
# ~/.config/onemount/config.yml
realtime:
  enabled: true
  pollingOnly: false
  fallbackIntervalSeconds: 1800  # 30 minutes
```

For corporate networks where WebSocket connections may be blocked, use polling-only mode:
```bash
onemount --polling-only /mnt/onedrive
```

See the [Socket.IO Configuration Guide](docs/guides/socketio-configuration.md) for details.

### ETag-Based Cache Validation

OneMount now uses ETags from delta sync to efficiently validate cached files, ensuring you always have the latest versions without unnecessary downloads.

**Key Benefits:**
- **Always Current**: Automatic cache invalidation when remote files change
- **Bandwidth Efficient**: Only downloads files that have actually changed
- **Proactive Detection**: Changes detected before you access the file
- **Works with Pre-authenticated URLs**: Compatible with Microsoft Graph's download URLs

**How It Works:**
1. Background delta sync fetches file metadata including ETags
2. When you access a file, OneMount compares cached ETag with current ETag
3. If ETags differ, cache is invalidated and file is re-downloaded
4. QuickXORHash verifies content integrity

### XDG Base Directory Compliance

OneMount now follows the XDG Base Directory Specification, storing configuration and cache files in standard Linux locations.

**Standard Locations:**
- Configuration: `~/.config/onemount/` (respects `$XDG_CONFIG_HOME`)
- Cache: `~/.cache/onemount/` (respects `$XDG_CACHE_HOME`)
- Virtual files like `.xdg-volume-info` served locally without OneDrive sync

**Benefits:**
- Standards compliance with Linux filesystem hierarchy
- Easier backups with clear separation of config and cache
- Better integration with backup tools
- Cleaner home directory

---

## 🚀 Major Features

### Account-Based Token Storage

Authentication tokens are stored based on account identity (email hash) for stable, mount-point-independent storage.

**Benefits:**
- Mount same account at different locations without issues
- Clean organization with one token location per account
- Support for multiple OneDrive accounts

**Token Location:**
- Storage: `~/.config/onemount/accounts/{account-hash}/auth_tokens.json`
- Hash: SHA256 of account email provides stable identifier

### Metadata State Machine

Explicit state tracking for file lifecycle with clear state transitions:

**States:**
- **GHOST**: Cloud metadata known, no local content
- **HYDRATING**: Content download in progress
- **HYDRATED**: Local content matches remote ETag
- **DIRTY_LOCAL**: Local changes pending upload
- **DELETED_LOCAL**: Local delete queued for upload
- **CONFLICT**: Local and remote versions diverged
- **ERROR**: Last operation failed

**Benefits:**
- Clear file status at all times
- Better error recovery with retry mechanisms
- Automatic conflict detection
- Progress tracking for operations

### Offline Mode with Conflict Resolution

Comprehensive offline functionality with intelligent conflict resolution:

**Offline Capabilities:**
- Full read-write access to cached files
- Create, modify, and delete files while offline
- Automatic change queuing for synchronization
- Passive network monitoring and active connectivity checks

**Conflict Resolution Strategies:**
- **Last-Writer-Wins**: Most recent modification wins (default)
- **Keep-Both**: Preserve both local and remote versions
- **User-Choice**: Manual selection of resolution
- **Merge**: Automatic merging for compatible changes
- **Rename**: Create separate versions with conflict indicators

**Configuration:**
```yaml
offline:
  connectivityCheckInterval: 15
  connectivityTimeout: 10
  maxPendingChanges: 1000
  conflictResolution: "keep-both"
```

---

## ✨ New Features

### Configuration Options

Extensive configuration options for customizing behavior:

**Realtime Notifications:**
- `realtime.enabled` - Enable/disable realtime notifications
- `realtime.pollingOnly` - Force polling-only mode
- `realtime.fallbackIntervalSeconds` - Polling interval (30-7200 seconds)

**Download Management:**
- `hydration.workers` - Concurrent download workers (1-64, default 4)
- `hydration.queueSize` - Download queue size (1-100000, default 500)
- `hydration.retryAttempts` - Retry attempts for failed downloads (default 3)

**Metadata Requests:**
- `metadataQueue.workers` - Metadata request workers (1-64, default 3)
- `metadataQueue.highPrioritySize` - High-priority queue size (default 100)
- `metadataQueue.lowPrioritySize` - Low-priority queue size (default 1000)

**Cache Management:**
- `cache.maxSizeMB` - Maximum cache size in megabytes
- `cache.expirationDays` - Cache expiration in days
- `cache.cleanupIntervalHours` - Cleanup interval in hours

**Virtual File Overlay:**
- `overlay.defaultPolicy` - Overlay policy (LOCAL_WINS, REMOTE_WINS, MERGED)

### Command-Line Options

New command-line flags for runtime configuration:

```bash
# Realtime
onemount --polling-only /mnt/onedrive
onemount --realtime-fallback-seconds 900 /mnt/onedrive

# Hydration
onemount --hydration-workers 8 /mnt/onedrive
onemount --hydration-queue-size 1000 /mnt/onedrive

# Metadata
onemount --metadata-workers 5 /mnt/onedrive
onemount --metadata-high-queue-size 200 /mnt/onedrive

# Overlay
onemount --overlay-policy LOCAL_WINS /mnt/onedrive

# Statistics
onemount --stats /mnt/onedrive
```

### Enhanced Statistics

Comprehensive statistics via `onemount --stats`:

- **Realtime Notifications**: Mode, status, heartbeat health, reconnect count
- **Hydration Queue**: Queue depth, active downloads, worker utilization
- **Metadata Queue**: Queue depth, average wait time, worker status
- **Cache Statistics**: Hit rate, size, invalidations, file count
- **File States**: Count of files in each state (GHOST, HYDRATED, etc.)
- **Upload Queue**: Pending uploads, active uploads, retry count
- **Offline Status**: Network state, pending changes, last connectivity check

---

## 🔧 Improvements

### Performance

- **Directory Listing**: < 2 seconds for 1000 files
- **Cached File Access**: < 100 milliseconds
- **Idle Memory Usage**: < 50 MB
- **Active Sync Memory Usage**: < 200 MB
- **Concurrent Operations**: 10+ simultaneous without degradation
- **Startup Time**: < 5 seconds
- **Shutdown Time**: < 10 seconds
- **Delta Sync**: 1000 changed files in < 30 seconds

### Reliability

- **Comprehensive Error Handling**: Retry mechanisms with exponential backoff
- **Automatic Network Recovery**: Seamless transition between online and offline
- **State Persistence**: Operations resume after restart
- **Graceful Degradation**: Falls back to polling when Socket.IO unavailable

### Security

- **Token Encryption**: AES-256 encryption for authentication tokens at rest
- **File Permissions**: Token files created with 0600 permissions (owner only)
- **TLS 1.2+**: All Microsoft Graph API communication uses HTTPS/TLS 1.2 or higher
- **Certificate Validation**: Certificate chain validation enforced
- **No Sensitive Logging**: Tokens, passwords, and sensitive data never logged
- **Rate Limiting**: Authentication failure rate limiting to prevent brute force

---

## 🐛 Bug Fixes

This is the first release, so there are no bug fixes to report. All features are new implementations.

---

## ⚠️ Known Issues

The following issues are known and will be addressed in v0.2.0:

### Resource Management

**Property 56: Cache Size Enforcement**
- **Issue**: Cache may exceed configured size limits
- **Impact**: Disk space usage may be higher than expected
- **Workaround**: Manually run `onemount --cleanup` to enforce limits
- **Fix Planned**: Implement LRU eviction with size tracking in v0.2.0

**Property 58: Worker Thread Limits**
- **Issue**: Worker goroutines may not fully cleanup
- **Impact**: Memory usage may grow over time
- **Workaround**: Restart OneMount if memory usage grows excessively
- **Fix Planned**: Improve worker pool synchronization in v0.2.0

**Property 59: Adaptive Network Throttling**
- **Issue**: Bandwidth throttling not fully adaptive
- **Impact**: May not optimally utilize available bandwidth
- **Workaround**: Configure fixed bandwidth limits in config
- **Fix Planned**: Implement adaptive throttling based on network conditions in v0.2.0

### Limitations

**Large Files**
- Files are loaded into memory during access
- Not suitable for multi-gigabyte files
- Use sync clients like [rclone](https://rclone.org/) for very large files

**Symbolic Links**
- Not supported by Microsoft OneDrive API
- Attempting to create symlinks returns ENOSYS (function not implemented)

**OneDrive Recycle Bin**
- Not exposed by Microsoft Graph API
- Use OneDrive web UI to empty or restore Recycle Bin

**File Browser Thumbnails**
- Many file browsers automatically download files for thumbnail generation
- Only needs to happen once; thumbnails persist between restarts

---

## 📚 Documentation

### New Documentation

- [v0.1.0 Feature Guide](docs/guides/user/v0.1.0-features.md) - Comprehensive overview of all features
- [Socket.IO Configuration Guide](docs/guides/socketio-configuration.md) - Detailed Socket.IO setup and troubleshooting
- [Authentication Architecture](docs/guides/developer/authentication-token-paths-v2.md) - Account-based token storage design

### Updated Documentation

- [README.md](README.md) - Complete project overview and feature list
- [CHANGELOG.md](CHANGELOG.md) - Complete v0.1.0 changelog
- [Troubleshooting Guide](docs/guides/user/troubleshooting-guide.md) - Comprehensive troubleshooting for all features
- [Offline Functionality Guide](docs/3-implementation/offline-functionality.md) - Detailed offline mode documentation
- [Running Tests Guide](docs/testing/running-tests.md) - Docker test environment documentation

---

## 🔄 Getting Started

This is the first public release of OneMount. There is no migration required.

### Installation Steps

1. **Install OneMount**: Follow the [Installation Guide](docs/guides/user/installation-guide.md)
2. **Authenticate**: Run `onemount /mnt/onedrive` and complete Microsoft OAuth2 flow
3. **Configure** (optional): Edit `~/.config/onemount/config.yml` to customize settings
4. **Monitor Status**: Use `onemount --stats /mnt/onedrive` to check system health

### Recommended Configuration

For optimal performance, consider these settings:

```yaml
# ~/.config/onemount/config.yml
realtime:
  enabled: true
  pollingOnly: false  # Set to true if WebSocket blocked
  fallbackIntervalSeconds: 1800

hydration:
  workers: 4
  queueSize: 500

cache:
  maxSizeMB: 10000
  expirationDays: 30
```

---

## 🙏 Acknowledgments

This first release represents months of development and testing. Special thanks to:

- The original one-driver project by Jeff Stafford, which provided the foundation
- Early testers who provided valuable feedback
- The Microsoft Graph API team for excellent documentation
- The Linux FUSE community for their support

---

## 📦 Installation

### Ubuntu 24.04 LTS / Linux Mint 22

```bash
# Download the latest release
wget https://github.com/auriora/OneMount/releases/latest/download/onemount_*.deb

# Install the package
sudo apt update
sudo apt install ./onemount_*.deb
```

### Build from Source

```bash
# Install dependencies
sudo apt update
sudo apt install golang-go build-essential pkg-config libwebkit2gtk-4.1-dev git fuse3

# Clone and build
git clone https://github.com/auriora/OneMount.git
cd OneMount
make all
sudo make install
```

See the [Installation Guide](docs/guides/user/installation-guide.md) for detailed instructions.

---

## 🔗 Resources

- **GitHub Repository**: [https://github.com/auriora/OneMount](https://github.com/auriora/OneMount)
- **Issue Tracker**: [https://github.com/auriora/OneMount/issues](https://github.com/auriora/OneMount/issues)
- **Documentation**: [https://github.com/auriora/OneMount/tree/main/docs](https://github.com/auriora/OneMount/tree/main/docs)
- **Releases**: [https://github.com/auriora/OneMount/releases](https://github.com/auriora/OneMount/releases)

---

## 📄 License

OneMount is licensed under the [GNU General Public License v3.0 (GPLv3)](LICENSE).

---

**Enjoy OneMount v0.1.0!** 🎉

This is the first public release. We welcome your feedback and bug reports on [GitHub Issues](https://github.com/auriora/OneMount/issues).

For help getting started, see the [Troubleshooting Guide](docs/guides/user/troubleshooting-guide.md).
