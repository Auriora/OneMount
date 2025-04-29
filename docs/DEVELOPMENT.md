# Onedriver Development Guidelines

This document provides concise guidance for new developers working on the onedriver project.

## Project Overview

onedriver is a native Linux filesystem for Microsoft OneDrive that performs on-demand file downloads rather than syncing. It's written in Go and uses FUSE to implement the filesystem.

## Project Structure

- **cmd/** - Command-line applications
  - **common/** - Shared code between applications
  - **onedriver/** - Main filesystem application
  - **onedriver-launcher/** - GUI launcher application
- **internal/** - Internal packages not meant for external use
  - **fs/** - Filesystem implementation
    - **graph/** - Microsoft Graph API integration
    - **offline/** - Offline mode functionality
  - **ui/** - GUI implementation
    - **systemd/** - Systemd integration for the UI
  - **testutil/** - Testing utilities
  - **nemo/** - Nemo file manager integration
- **configs/** - Configuration and resource files
  - **resources/** - Application resources (icons, man pages, etc.)
- **scripts/** - Utility scripts
  - **debian/** - Debian packaging files

## Tech Stack

- **Go** - Primary programming language
- **FUSE (go-fuse/v2)** - Filesystem implementation
- **GTK3 (gotk3)** - GUI components
- **bbolt** - Embedded database for caching
- **zerolog** - Structured logging
- **testify** - Testing framework

## Building the Project

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

```bash
# Setup test environment (first time only)
make test-init

# Run all tests
make test

# Run specific tests
go test ./internal/fs/...
go test ./cmd/...
go test ./internal/ui/...
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

The onedriver project follows a set of comprehensive coding standards and best practices. For detailed guidelines, refer to the documents in the `docs/guides` directory:

- [Coding Standards](guides/coding-standards.md) - Main entry point for all coding standards
- [Go Coding Standards](guides/coding-standards.md#go-standards) - Comprehensive guide for Go code
- [Go Logging Best Practices](guides/logging-guidelines.md) - Guidelines for structured logging
- [Test Best Practices](guides/test-guidelines.md) - Best practices for writing tests

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

- Use `journalctl --user -u onedriver@.service --since today` to view logs
- Set the environment variable `ONEDRIVER_DEBUG=1` for verbose logging
- Use `fusermount3 -uz $MOUNTPOINT` to unmount the filesystem if it hangs

## Architecture Summary

This architecture allows onedriver to provide a seamless experience where OneDrive files appear as local files but are only downloaded when accessed, saving bandwidth and storage space while maintaining full compatibility with the OneDrive service.
