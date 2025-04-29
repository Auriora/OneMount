# onedriver Development Guidelines

This document provides concise guidance for new developers working on the onedriver project.

## Project Overview

onedriver is a native Linux filesystem for Microsoft OneDrive that performs on-demand file downloads rather than syncing. It's written in Go and uses FUSE to implement the filesystem.

## Project Structure

- **cmd/** - Command-line applications
  - **common/** - Shared code between applications
  - **onedriver/** - Main filesystem application
  - **onedriver-launcher/** - GUI launcher application
- **internal/** - Internal implementation code
  - **fs/** - Filesystem implementation
    - **graph/** - Microsoft Graph API integration
    - **offline/** - Offline mode functionality
  - **ui/** - GUI implementation
    - **systemd/** - Systemd integration for the UI
  - **nemo/** - Nemo file manager integration
  - **testutil/** - Testing utilities
- **pkg/** - Resources and packaging files
- **docs/** - Documentation
  - **guides/** - Development guides
  - **design/** - Design documentation
  - **implementation/** - Implementation details
  - **requirements/** - Project requirements
  - **templates/** - Documentation templates
  - **testing/** - Testing documentation
- **.run/** - GoLand run configurations

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
```

## Running Tests

```bash
# Setup test environment (first time only)
make test-init

# Run all tests
make test

# Run specific tests
go test ./fs/...
go test ./cmd/...
go test ./ui/...
```

### JetBrains GoLand Run Configurations

The project includes predefined run configurations for JetBrains GoLand that replicate the functionality of the `make test` command. These configurations are stored in the `.run/` directory.

Available run configurations:
- **UI Tests** - Runs tests in the UI package, excluding offline tests
- **Command Tests** - Runs tests in the cmd package
- **Graph Tests with Race Detection** - Runs tests in the fs/graph package with race detection
- **FS Tests with Race Detection** - Runs tests in the fs package with race detection
- **Offline Tests** - Builds the offline test binary and provides instructions for running it
- **All Tests Except Offline** - Runs all the above tests except for Offline Tests

To use these configurations:
1. Open the project in GoLand
2. Go to the Run/Debug Configurations dropdown in the toolbar
3. Select the desired configuration and click the Run button

Note: Offline tests require sudo privileges to simulate network disconnection.

## Executing Scripts

- **cgo-helper.sh** - Helps with CGO compilation
- **curl-graph.sh** - Utility for interacting with Microsoft Graph API

## Best Practices

The onedriver project follows a set of comprehensive coding standards and best practices. For detailed guidelines, refer to the documents in the `docs/guides/` directory:

- [Coding Standards](../docs/guides/coding-standards.md) - Main entry point for all coding standards
- [Design Guidelines](../docs/guides/design-guidelines.md) - Guidelines for designing components
- [Logging Guidelines](../docs/guides/logging-guidelines.md) - Guidelines for structured logging
- [Logging Examples](../docs/guides/logging-examples.md) - Examples of proper logging
- [Test Guidelines](../docs/guides/test-guidelines.md) - Best practices for writing tests
- [Threading Guidelines](../docs/guides/threading-guidelines.md) - Guidelines for concurrent programming
- [D-Bus Integration](../docs/guides/dbus-integration.md) - Guidelines for D-Bus integration

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
   - Use 'testify' for assertions
   - Test edge cases, especially around network connectivity

4. **Performance**
   - Cache filesystem metadata and file contents
   - Minimize network requests
   - Use concurrent operations where appropriate

5. **Documentation**
   - Document public APIs with godoc-compatible comments
   - Add comments explaining complex logic
   - Keep the README up to date

6. **Method Logging**
   - Use the method logging framework for all public methods
   - Follow the patterns in `docs/guides/logging-guidelines.md`
   - Log method entry and exit, including parameters and return values

7. **D-Bus Integration**
   - Use the D-Bus interface for file status updates
   - Follow the specification in `docs/guides/dbus-integration.md`
   - Ensure backward compatibility with extended attributes

8. **Microsoft Graph API Integration**
   - Use direct API endpoints when available for better performance
   - For thumbnail retrieval, use the direct content endpoint (`/thumbnails/0/{size}/content`) instead of making separate metadata and content requests
   - Cache API responses appropriately to reduce network traffic

## Common Tasks

- **Adding a new feature**: Create tests first, implement the feature, then verify tests pass
- **Fixing a bug**: Create a test that reproduces the bug, fix the bug, verify the test passes
- **Refactoring**: Ensure tests pass before and after refactoring

## Debugging

- Use `journalctl --user -u onedriver@.service --since today` to view logs
- Set the environment variable `ONEDRIVER_DEBUG=1` for verbose logging
- Use `fusermount3 -uz $MOUNTPOINT` to unmount the filesystem if it hangs
