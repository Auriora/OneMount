
# onedriver Project Structure

onedriver is a native Linux filesystem for Microsoft OneDrive that provides on-demand access to files rather than syncing them. Here's an overview of the project structure:

## Core Components

### 1. Command-Line Applications (`cmd/`)
- **onedriver** - The main filesystem mount command that implements the FUSE interface
- **onedriver-launcher** - A GUI launcher for the filesystem
- **common** - Shared code between commands (config handling, etc.)

### 2. Filesystem Implementation (`fs/`)
- Core filesystem implementation that handles FUSE operations
- **graph/** - Microsoft Graph API integration for OneDrive communication
  - Handles authentication (OAuth2)
  - Implements DriveItem representation
  - Manages file hashing (QuickXORHash)
- **offline/** - Handles offline functionality

### 3. User Interface (`ui/`)
- GTK-based user interface for managing OneDrive accounts
- **systemd/** - Integration with systemd for service management
- Widgets and UI components

### 4. Packaging and Resources (`pkg/`)
- **debian/** - Debian packaging files
- **resources/** - Application resources (icons, man pages, etc.)

## Key Technical Features

1. **FUSE Filesystem Implementation**
   - Implements the low-level FUSE API
   - Handles file operations (read, write, create, delete, etc.)

2. **Caching System**
   - Local content cache for files
   - Metadata caching using BoltDB
   - Delta synchronization to efficiently track changes

3. **Authentication**
   - OAuth2 authentication with Microsoft
   - Support for both GUI and headless authentication

4. **Upload Management**
   - Handles file uploads to OneDrive
   - Supports large file uploads via upload sessions

## Project Dependencies

The project uses several key libraries:
- `github.com/hanwen/go-fuse/v2` - FUSE bindings for Go
- `github.com/gotk3/gotk3` - GTK3 bindings for Go
- `go.etcd.io/bbolt` - Key/value store for caching
- `github.com/coreos/go-systemd` - systemd integration

## Build System

The project uses a Makefile for building and packaging, with support for:
- RPM packages (via COPR)
- Debian packages
- Direct installation

This architecture allows onedriver to provide a seamless experience where OneDrive files appear as local files but are only downloaded when accessed, saving bandwidth and storage space while maintaining full compatibility with the OneDrive service.