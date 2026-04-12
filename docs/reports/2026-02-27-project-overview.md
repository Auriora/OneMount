# OneMount Project Overview

**Date**: February 27, 2026
**Repository**: github.com/auriora/onemount
**Version**: 0.1.0 (Release Candidate)
**License**: GPL v3.0

## Executive Summary

OneMount is a production-quality Linux filesystem driver that mounts Microsoft OneDrive accounts as native FUSE filesystems. Unlike traditional sync clients, OneMount provides on-demand file access, real-time synchronization via Socket.IO, and comprehensive offline functionality with intelligent conflict resolution.

The project demonstrates professional software engineering practices with ~113,000 lines of well-structured Go code, comprehensive testing, extensive documentation, and sophisticated build automation.

## Project Description

### Core Value Proposition

OneMount solves the problem of accessing OneDrive files on Linux without the overhead of traditional sync clients:

- **On-demand downloads** - Files are only downloaded when accessed, saving disk space and sync time
- **Real-time synchronization** - Socket.IO/WebSocket integration provides instant change notifications
- **Offline functionality** - Full read-write operations while offline with conflict resolution upon reconnection
- **ETag-based caching** - Efficient validation through Microsoft Graph delta sync
- **Bidirectional sync** - Local changes upload automatically; remote changes reflect instantly
- **Native integration** - Works seamlessly with Linux desktop environments

### Origin and Evolution

Originally forked from Jeff Stafford's one-driver project, OneMount has undergone extensive modifications warranting the rename. The project has evolved into a sophisticated filesystem implementation with:

- Complete Socket.IO protocol implementation for real-time sync
- Advanced offline mode with multiple conflict resolution strategies
- Professional build and release management
- Comprehensive testing framework
- Multi-account support

## Technology Stack

### Core Technologies

| Component | Technology | Purpose |
|-----------|-----------|---------|
| **Filesystem** | FUSE3 (go-fuse v2.7.2) | Userspace filesystem implementation |
| **Backend API** | Microsoft Graph API | OneDrive integration |
| **GUI** | GTK3 (gotk3 v0.6.3) | Launcher application |
| **Authentication** | webkit2gtk-4.0/4.1 | OAuth 2.0 flow |
| **Database** | BBolt v1.4.0 | Metadata persistence |
| **IPC** | D-Bus (godbus v5.1.0) | Desktop integration |
| **Real-time** | Socket.IO (custom) | Change notifications |
| **Logging** | zerolog v1.34.0 | Structured logging |
| **Systemd** | go-systemd v22.5.0 | Service management |

### Build and Development

- **Language**: Go 1.23+ (toolchain 1.24.2)
- **Build System**: Make with complex packaging logic
- **CI/CD**: GitHub Actions with self-hosted runners
- **Package Formats**: .deb (Debian/Ubuntu), .rpm (Fedora/RHEL)
- **Dev CLI**: Python-based unified development tool
- **Testing**: Go test framework + pytest for Python components

## Architecture

### Directory Structure

```
OneMount/
├── cmd/                           # Application entry points
│   ├── onemount/                 # Main FUSE filesystem daemon
│   ├── onemount-launcher/        # GTK GUI launcher
│   └── common/                   # Shared command utilities
├── internal/                      # Internal packages (not exported)
│   ├── fs/                       # Core FUSE filesystem (largest package)
│   ├── graph/                    # Microsoft Graph API client
│   │   ├── api/                  # API request implementations
│   │   ├── mock/                 # Mock implementations for testing
│   │   └── debug/                # Debug utilities
│   ├── socketio/                 # Socket.IO protocol implementation
│   │   └── protocol/             # Protocol details
│   ├── ui/                       # GTK GUI and systemd integration
│   │   └── systemd/              # Systemd service management
│   ├── config/                   # Configuration management
│   ├── metadata/                 # Metadata caching layer
│   ├── logging/                  # Structured logging wrapper
│   ├── nemo/                     # Nemo file manager extension
│   ├── performance/              # Performance monitoring
│   ├── quickxorhash/             # OneDrive hash algorithm
│   ├── retry/                    # Retry logic with backoff
│   ├── testutil/                 # Testing framework
│   │   ├── framework/            # Test framework
│   │   ├── helpers/              # Test helpers
│   │   └── mock/                 # Mock implementations
│   └── util/                     # General utilities
├── scripts/                       # Development and build scripts
│   ├── dev                       # Unified Python CLI tool
│   ├── utils/                    # Python utilities
│   └── lib/                      # Shell script libraries
├── packaging/                     # Distribution packaging files
│   ├── debian/                   # Debian package configuration
│   └── rpm/                      # RPM package configuration
├── build/                         # Build artifacts (generated)
│   ├── binaries/                 # Compiled binaries
│   ├── packages/                 # Built packages (.deb, .rpm)
│   └── docker/                   # Docker build artifacts
├── docs/                          # Comprehensive documentation
│   ├── guides/                   # User and developer guides
│   ├── testing/                  # Testing documentation
│   ├── reports/                  # Project reports
│   └── [0-4]-*/                  # Organized by project phase
├── tests/                         # Test suites
│   ├── system/                   # System integration tests
│   └── manual/                   # Manual test procedures
├── deployments/                   # Deployment configurations
│   ├── systemd/                  # Systemd service files
│   └── desktop/                  # Desktop integration files
├── assets/                        # Static assets (icons, etc.)
├── configs/                       # Example configurations
└── .github/                       # GitHub workflows and templates
    └── workflows/                # CI/CD workflow definitions
```

### Core Components

#### 1. Filesystem Layer (internal/fs/)

**Purpose**: Implements the FUSE filesystem interface and manages file operations.

**Key Files**:
- `filesystem.go` - Main filesystem implementation
- `inode.go` - File/directory inode management
- `content_cache.go` - Content caching with LRU eviction
- `upload_manager.go` - Upload queue and coordination
- `mutation_queue.go` - Tracks pending local modifications
- `offline.go` - Offline mode and conflict resolution
- `virtual_files.go` - XDG compliance virtual files
- `dbus.go` - D-Bus signal emission for file status
- `prefetch.go` - Predictive content downloading

**Responsibilities**:
- FUSE operation handling (read, write, mkdir, etc.)
- Inode tree management with proper locking
- Content and metadata caching
- Upload queue management
- Offline operation support
- Conflict detection and resolution
- D-Bus integration for desktop environments

#### 2. Microsoft Graph Integration (internal/graph/)

**Purpose**: Manages all interactions with Microsoft Graph API.

**Key Files**:
- `graph.go` - Main Graph API client
- `oauth2.go` - OAuth 2.0 authentication flow
- `oauth2_account_storage.go` - Multi-account token management
- `http_client.go` - HTTP client with retry logic
- `request_queue.go` - Request queuing and rate limiting
- `delta_sync.go` - Efficient delta synchronization
- `socket_subscription.go` - Socket.IO subscription management
- `provider.go` - Account provider interface

**Responsibilities**:
- OAuth 2.0 authentication and token refresh
- Multi-account support
- API request execution with retry/backoff
- Delta sync for efficient updates
- Pagination handling
- Socket.IO subscription lifecycle
- Error handling and recovery

#### 3. Real-time Synchronization (internal/socketio/)

**Purpose**: Custom Socket.IO protocol implementation for instant change notifications.

**Key Files**:
- `client.go` - Socket.IO client implementation
- `protocol/parser.go` - Protocol parsing
- `transport.go` - WebSocket transport with fallback

**Responsibilities**:
- WebSocket connection management
- Socket.IO protocol parsing/encoding
- Automatic reconnection with exponential backoff
- Health monitoring and heartbeat
- Fallback to polling when needed

#### 4. User Interface (internal/ui/)

**Purpose**: GTK-based GUI and system integration.

**Key Files**:
- `launcher.go` - Main launcher application
- `auth.go` - WebKit-based OAuth flow
- `systemd/service.go` - Systemd service management

**Responsibilities**:
- Account management interface
- WebKit-based OAuth authentication
- Systemd user service integration
- Desktop environment integration

#### 5. Configuration Management (internal/config/)

**Purpose**: YAML-based configuration with XDG compliance.

**Key Features**:
- XDG Base Directory compliance
- Default configuration generation
- Command-line override support
- Validation and migration

#### 6. Testing Framework (internal/testutil/)

**Purpose**: Comprehensive testing utilities and mocks.

**Components**:
- Test framework with setup/teardown
- Mock Graph API implementation
- Test helpers for common scenarios
- Integration test utilities

## Key Features

### 1. On-Demand File Access

Files are downloaded only when accessed, providing:
- Instant access to entire OneDrive
- Minimal disk space usage
- No lengthy initial sync period
- Automatic cleanup of unused cached files

### 2. Real-time Synchronization

Three synchronization modes:

**Socket.IO Mode** (recommended):
```yaml
realtime:
  enabled: true
  pollingOnly: false
  fallbackIntervalSeconds: 1800  # 30 minutes
```

**Polling-Only Mode**:
```yaml
realtime:
  enabled: true
  pollingOnly: true
  fallbackIntervalSeconds: 300   # 5 minutes
```

**Disabled Mode**:
```yaml
realtime:
  enabled: false
```

### 3. Offline Functionality

Complete offline support including:
- Full read-write operations while disconnected
- Automatic change detection upon reconnection
- Multiple conflict resolution strategies:
  - Last-writer-wins (default)
  - Keep-both (creates conflict copies)
  - User choice (interactive)
- Intelligent merge for compatible changes

### 4. Intelligent Caching

Multi-level caching strategy:
- **In-memory cache** - Hot data for fast access
- **On-disk cache** - Persistent storage in `~/.cache/onemount/`
- **ETag validation** - Efficient cache invalidation
- **Delta sync** - Only fetch changes since last sync
- **LRU eviction** - Automatic cache size management

### 5. Performance Optimization

**Configurable Worker Pools**:
```yaml
hydration:
  workers: 4           # 1-64 concurrent downloads
  queueSize: 500       # Max queued download requests

metadataQueue:
  workers: 3           # 1-64 concurrent metadata requests
  highPrioritySize: 100
  lowPrioritySize: 1000
```

**Prefetching**:
- Predictive content downloading
- Directory-aware prefetch strategies

### 6. Multi-Account Support

- Manage multiple OneDrive accounts simultaneously
- Independent caches and configurations per account
- Account switching without unmounting

### 7. XDG Base Directory Compliance

- Configuration: `~/.config/onemount/` (respects `XDG_CONFIG_HOME`)
- Cache: `~/.cache/onemount/` (respects `XDG_CACHE_HOME`)
- Virtual files (e.g., `.xdg-volume-info`) served locally without syncing

### 8. Desktop Integration

- **D-Bus interface** - File status signals for desktop environments
- **Nemo extension** - File manager overlay icons (Python)
- **Systemd integration** - User service management
- **Desktop entry** - Application launcher integration

### 9. Statistics and Monitoring

```bash
onemount --stats /mount/path
```

Provides:
- Cache utilization (metadata and content)
- Upload queue status
- File status distribution
- Hydration queue depth and active downloads
- Metadata queue statistics
- Real-time heartbeat health
- BBolt database analytics

## Build System

### Build Targets

**Binaries**:
- `onemount` - Main filesystem daemon (requires CGO for GTK)
- `onemount-launcher` - GUI launcher (requires CGO for GTK)
- `onemount-headless` - Headless build (CGO_ENABLED=0)

**Packages**:
- `.deb` packages for Ubuntu/Debian
- `.rpm` packages for Fedora/RHEL
- Docker images for reproducible builds

### Build Tags

The build system automatically detects system libraries:

```makefile
# Detected at build time
GOTK3_GLIB_TAG = glib_2_66  # or glib_2_40, etc.
WEBKIT_TAG = webkit41        # or webkit40
GO_BUILD_TAGS = "glib_2_66 webkit41"
```

### Development CLI

Unified Python CLI tool (`scripts/dev`):

```bash
# Build operations
./scripts/dev build deb --docker
./scripts/dev build rpm --native

# Testing
./scripts/dev test coverage --threshold-line 85
./scripts/dev test system --category comprehensive

# Release management
./scripts/dev release bump patch --dry-run
./scripts/dev release bump num  # Bump RC number

# Analysis
./scripts/dev analyze test-suite --mode resolve

# Cleanup
./scripts/dev clean all
```

## Testing Strategy

### Test Categories

**1. Unit Tests**
- Go tests for all packages
- ~70-80% code coverage
- Fast feedback for development

**2. Integration Tests**
- Real OneDrive API integration
- Multi-account scenarios
- Network error handling

**3. Property-Based Tests**
- Offline mode behavior
- Lock correctness
- Cache consistency

**4. System Tests**
- End-to-end scenarios with real OneDrive
- Multiple test categories:
  - Quick smoke tests
  - Comprehensive feature tests
  - Stress tests
  - Multi-account tests

**5. Docker Tests**
- Isolated environment testing
- Reproducible test runs
- CI/CD integration

**6. Python Tests (pytest)**
- Nemo file manager extension
- D-Bus interface verification

### CI/CD Workflows

**Continuous Integration** (`ci.yml`):
- Runs on every push/PR to main
- Unit and integration tests
- Build verification
- Self-hosted runners for performance

**Package Building** (`build-packages.yml`):
- Triggered by version tags (e.g., `v0.1.0rc2`)
- Builds .deb and .rpm packages
- Creates GitHub releases
- Uploads packages as release assets

**Coverage Analysis** (`coverage.yml`):
- Tracks code coverage trends
- Comments on PRs with coverage changes
- Maintains coverage history

**System Tests** (`system-tests.yml`):
- Comprehensive E2E testing
- Real OneDrive account integration
- Multiple test scenarios

## Documentation

### User Documentation

Located in `docs/guides/user/`:
- **Quickstart Guide** - Step-by-step getting started
- **Installation Guide** - Detailed installation for all distros
- **Ubuntu Installation Guide** - Ubuntu/Linux Mint specific
- **Troubleshooting Guide** - Solutions for common issues
- **Configuration Guide** - Advanced configuration options

### Developer Documentation

Located in `docs/guides/developer/`:
- **Development Guidelines** - Project structure and practices
- **Debugging Guide** - Logs, tracing, diagnostics
- **Testing Guide** - Running and writing tests
- **Release Management** - Version management and releases

### Implementation Documentation

Located in `docs/3-implementation/`:
- Offline functionality deep dive
- Socket.IO implementation details
- Caching strategies
- Error handling patterns

### Testing Documentation

Located in `docs/4-testing/`:
- Testing framework documentation
- Test writing guidelines
- System test procedures
- Coverage requirements

## Installation and Usage

### Ubuntu 24.04 / Linux Mint 22

```bash
# Download latest release
wget https://github.com/auriora/OneMount/releases/latest/download/onemount_*.deb

# Install
sudo apt update
sudo apt install ./onemount_*.deb
```

### Build from Source

```bash
# Install dependencies
sudo apt install golang-go build-essential pkg-config \
  libwebkit2gtk-4.1-dev git fuse3

# Clone and build
git clone https://github.com/auriora/OneMount.git
cd OneMount
make all
make install
```

### Basic Usage

```bash
# Launch GUI
onemount-launcher

# Command-line mount
onemount /path/to/mount/point

# View statistics
onemount --stats /path/to/mount/point

# View help
onemount --help
man onemount
```

## Project Health Indicators

### Active Development
- Recent commits within days
- Consistent development velocity
- Well-maintained issue tracker

### Code Quality
- Comprehensive test coverage
- Professional error handling
- Structured logging throughout
- Memory-safe concurrent operations

### Documentation Quality
- Extensive user guides
- Complete developer documentation
- Well-commented code
- Regular documentation updates

### Release Management
- Automated version bumping
- Release candidate workflow
- Automated package building
- GitHub release automation

### Community
- Clear contribution guidelines
- Code of conduct
- Security policy
- Support documentation

## Known Limitations

### 1. File Browser Thumbnails
Many file browsers automatically download all files in a directory to create thumbnails. This only happens once - thumbnails persist between restarts.

**Workaround**: First directory access may be slow due to thumbnail generation.

### 2. Symbolic Links
Microsoft OneDrive doesn't support symbolic links. Attempting to create symlinks returns `ENOSYS` (function not implemented).

### 3. Large Files
OneMount loads files into memory for performance. This doesn't work well with very large files (multi-gigabyte).

**Recommendation**: Use sync clients like rclone for very large files.

### 4. OneDrive Recycle Bin
Microsoft doesn't expose Recycle Bin APIs. Must use OneDrive web UI to manage deleted files.

**Note**: OneMount uses native system trash independently.

### 5. Backups
OneDrive is not recommended as a backup solution.

**Recommendation**: Use dedicated backup tools like restic or borg.

## Configuration

### Configuration File

Location: `~/.config/onemount/config.yml`

```yaml
# Real-time synchronization
realtime:
  enabled: true
  pollingOnly: false
  fallbackIntervalSeconds: 1800

# Download performance
hydration:
  workers: 4
  queueSize: 500

# Metadata requests
metadataQueue:
  workers: 3
  highPrioritySize: 100
  lowPrioritySize: 1000

# Overlay policy for virtual files
overlay:
  defaultPolicy: REMOTE_WINS  # or LOCAL_WINS, MERGED

# Cache configuration
cache:
  maxSize: 10737418240  # 10 GB
  ttl: 3600             # 1 hour
```

### Command-Line Overrides

```bash
# Force polling-only mode
onemount --polling-only /mount/path

# Set polling interval
onemount --realtime-fallback-seconds 600 /mount/path

# Override cache directory
onemount --cache-dir /custom/cache /mount/path

# Set worker counts
onemount --hydration-workers 8 --metadata-workers 4 /mount/path

# Set overlay policy
onemount --overlay-policy LOCAL_WINS /mount/path
```

## System Requirements

### Operating System
- **Linux kernel** 4.15+ with FUSE3 support
- **Distributions**: Ubuntu 24.04 LTS, Linux Mint 22, or compatible

### Hardware
- **Memory**: 100 MB minimum, 200 MB recommended
- **Disk**: Varies by cache size (default 10 GB)
- **Network**: Broadband internet recommended

### Software Dependencies
- FUSE3 (`fuse3` package)
- systemd (for service management)
- D-Bus (for desktop integration)

### Build Dependencies
- Go 1.23+
- GCC
- webkit2gtk-4.0 or webkit2gtk-4.1 development headers
- json-glib development headers
- pkg-config

## Project Statistics

- **Total Go Code**: ~113,000 lines
- **Packages**: 16 internal packages
- **Test Files**: Extensive test coverage across all packages
- **Dependencies**: 16 direct dependencies
- **Documentation**: 100+ documentation files
- **CI/CD Workflows**: 12 GitHub Actions workflows

## Conclusion

OneMount is a mature, production-quality Linux filesystem driver for Microsoft OneDrive. The project demonstrates professional software engineering practices including:

- Well-architected codebase with clear separation of concerns
- Comprehensive testing at multiple levels
- Extensive documentation for users and developers
- Sophisticated build and release automation
- Active development and maintenance
- Strong focus on performance and reliability

The project successfully fills a significant gap in the Linux ecosystem by providing native OneDrive access without the overhead of traditional sync clients.
