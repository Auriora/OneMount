# OneMount Development Guidelines

This document provides concise guidance for new developers working on the OneMount project.

## Project Overview

OneMount is a native Linux filesystem for Microsoft OneDrive that performs on-demand file downloads rather than syncing. It's written in Go and uses FUSE to implement the filesystem.

## Project Structure

- **assets/** - Project assets
    - **examples/** - Example files
    - **icons/** - Icon files
- **build/** - Build artifacts and configuration
    - **cli/** - Command-line interface build files
    - **dev/** - Development build files
    - **package/** - Packaging files
- **cmd/** - Command-line applications
    - **common/** - Shared code between applications
    - **onemount/** - Main filesystem application
    - **onemount-launcher/** - GUI launcher application
- **configs/** - Configuration files and resources
    - **resources/** - Resource files for the application
- **data/** - Data files and resources for the project
- **deployments/** - Deployment configurations
    - **desktop/** - Desktop environment integration
    - **systemd/** - Systemd service files
- **docs/** - Documentation
    - **0-project-management/** - Project management documentation
    - **1-requirements/** - Project requirements
    - **2-architecture-and-design/** - Design documentation
    - **3-implementation/** - Implementation details
    - **4-testing/** - Testing documentation
    - **A-templates/** - Documentation templates
    - **guides/** - Development guides
    - **training/** - Training materials
- **internal/** - Internal implementation code
    - **fs/** - Filesystem implementation
    - **ui/** - GUI implementation
        - **systemd/** - Systemd integration for the UI
    - **nemo/** - Nemo file manager integration
- **pkg/** - (deprecated) previously shared public packages; code migrated under `internal/`
    - **errors/** - Error handling utilities
    - **graph/** - Microsoft Graph API client
    - **logging/** - Logging utilities
    - **quickxorhash/** - QuickXorHash implementation
    - **testutil/** - Testing utilities
    - **util/** - General utilities

## Tech Stack

- **Go** - Primary programming language
- **FUSE (go-fuse/v2)** - Filesystem implementation
- **GTK3 (gotk3)** - GUI components
- **bbolt** - Embedded database for caching
- **zerolog** - Structured logging
- **testify** - Testing framework

## Development CLI Tool

OneMount includes a unified development CLI tool that consolidates all development, build, testing, and deployment operations. This is the recommended way to perform development tasks.

### Setup

```bash
# Install CLI dependencies (first time only)
pip install -r scripts/requirements-dev-cli.txt

# Make the CLI tool executable
chmod +x scripts/dev.py

# Check development environment status
./scripts/dev info
```

### Common Development Tasks

```bash
# Build packages
./scripts/dev build deb --docker          # Build Debian packages with Docker
./scripts/dev build deb --native          # Build Debian packages natively

# Run tests
./scripts/dev test coverage --threshold-line 85    # Generate coverage reports
./scripts/dev test system --category comprehensive # Run system tests
./scripts/dev test docker all --verbose           # Run all tests in Docker

# Code analysis
./scripts/dev analyze test-suite --mode resolve    # Analyze and fix test issues
./scripts/dev analyze coverage-trends             # Analyze coverage trends

# Release management
./scripts/dev release bump patch --dry-run        # Preview version bump
./scripts/dev release bump num                    # Bump release candidate

# GitHub integration
./scripts/dev github create-issues --dry-run      # Preview GitHub issue creation
./scripts/dev github implement 123               # Implement GitHub issue #123

# Cleanup operations
./scripts/dev clean list                          # List cleanable artifacts
./scripts/dev clean all                           # Clean all artifacts

# Get help
./scripts/dev --help                             # General help
./scripts/dev build --help                       # Build command help
```

The CLI tool provides:
- **Unified interface** for all development operations
- **Rich terminal output** with colors and progress indicators
- **Built-in help** with examples for every command
- **Error handling** with prerequisite checking
- **Organized commands** in logical groups

For detailed CLI documentation, see [Development CLI Guide](../scripts/README.md).

## Building the Project

### Using the CLI Tool (Recommended)

```bash
# Build packages using the CLI tool
./scripts/dev build deb --docker    # Docker-based build (recommended)
./scripts/dev build deb --native    # Native build

# Install using manifest
./scripts/dev build manifest --target makefile --type user --action install
```

### Using Make (Traditional)

```bash
# Build the main binaries
make

# Install the application system-wide
sudo make install

# Create distribution packages
make rpm    # For RPM-based distributions
make deb    # For Debian-based distributions

# Update import paths after restructuring the project
make update-imports
```

## Running Tests

### Using the CLI Tool (Recommended)

```bash
# Run tests with coverage analysis
./scripts/dev test coverage --threshold-line 80 --threshold-func 90

# Run system tests
./scripts/dev test system --category comprehensive --verbose
./scripts/dev test system --category performance --timeout 20m

# Run tests in Docker containers
./scripts/dev test docker unit --verbose
./scripts/dev test docker all --rebuild

# Analyze test suite
./scripts/dev analyze test-suite --mode analyze
./scripts/dev analyze test-suite --mode resolve
```

### Using Make (Traditional)

```bash
# Setup test environment (first time only)
make test-init

# Run all tests
make test

# Run specific tests
go test ./internal/fs/...
go test ./cmd/...
go test ./internal/ui/...

# Run system tests
make system-test-real
make system-test-all
```

### JetBrains GoLand Run Configurations

The project includes predefined run configurations for JetBrains GoLand that replicate the functionality of the `make test` command. These configurations are stored in the `.run/` directory.

Available run configurations:
- **UI Tests** - Runs tests in the internal/ui package, excluding offline tests
- **Command Tests** - Runs tests in the cmd package
- **Graph Tests with Race Detection** - Runs tests in the internal/fs/graph package with race detection
- **FS Tests with Race Detection** - Runs tests in the internal/fs package with race detection
- **Offline Tests** - Builds the offline test binary and provides instructions for running it
- **All Tests Except Offline** - Runs all the above tests except for Offline Tests

To use these configurations:
1. Open the project in GoLand
2. Go to the Run/Debug Configurations dropdown in the toolbar
3. Select the desired configuration and click the Run button

Note: Offline tests require sudo privileges to simulate network disconnection.

## Executing Scripts

- **scripts/cgo-helper.sh** - Helps with CGO compilation
- **scripts/curl-graph.sh** - Utility for interacting with Microsoft Graph API
- **scripts/update_imports.sh** - Updates import paths after restructuring the project

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

## Best Practices

The OneMount project follows a set of comprehensive coding standards and best practices. For detailed guidelines, refer to the documents in the `docs/guides` directory:

- [Coding Standards](guides/coding-standards.md) - Main entry point for all coding standards
- [Go Coding Standards](guides/coding-standards.md#go-standards) - Comprehensive guide for Go code
- [Go Logging Best Practices](guides/logging-guidelines.md) - Guidelines for structured logging
- [Test Best Practices](guides/testing/test-guidelines.md) - Best practices for writing tests

Here's a summary of key best practices:

1. **Code Organization**
   - Group related functionality into separate files
   - Use interfaces to decouple components
   - Follow Go's standard project layout

2. **Error Handling**
   - Return errors to callers instead of handling them internally
   - Use structured logging with zerolog
   - Avoid using `log.Fatal()` in library code

3. **Testing**
   - Write both unit and integration tests
   - Use testify for assertions
   - Test edge cases, especially around network connectivity

4. **Performance**
   - Cache filesystem metadata and file contents
   - Minimize network requests
   - Use concurrent operations where appropriate

5. **Documentation**
   - Document public APIs with godoc-compatible comments
   - Add comments explaining complex logic
   - Keep the README up-to-date

## Common Tasks

- **Adding a new feature**: Create tests first, implement the feature, then verify tests pass
- **Fixing a bug**: Create a test that reproduces the bug, fix the bug, verify the test passes
- **Refactoring**: Ensure tests pass before and after refactoring

## Debugging

- Use `journalctl --user -u onemount@.service --since today` to view logs
- Set the environment variable `ONEMOUNT_DEBUG=1` for verbose logging
- Use `fusermount3 -uz $MOUNTPOINT` to unmount the filesystem if it hangs

## Architecture Summary

This architecture allows OneMount to provide a seamless experience where OneDrive files appear as local files but are only downloaded when accessed, saving bandwidth and storage space while maintaining full compatibility with the OneDrive service.
