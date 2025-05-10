# OneMount Development Guidelines

This document provides concise guidance for new developers working on the OneMount project.

## Project Overview

OneMount is a native Linux filesystem for Microsoft OneDrive that performs on-demand file downloads rather than syncing. It's written in Go and uses FUSE to implement the filesystem.

## Project Structure

- **cmd/** - Command-line applications
  - **common/** - Shared code between applications
  - **onemount/** - Main filesystem application
  - **onemount-launcher/** - GUI launcher application
- **configs/** - Configuration files and resources
  - **resources/** - Resource files for the application
- **data/** - Data files and resources for the project
- **docs/** - Documentation
  - **0-project-management/** - Project management documentation
  - **1-requirements/** - Project requirements
  - **2-architecture-and-design/** - Design documentation
  - **3-implementation/** - Implementation details
  - **4-testing/** - Testing documentation
  - **A-templates/** - Documentation templates
  - **guides/** - Development guides
  - **training/** - Training materials
- **.github/** - GitHub-specific configuration
  - **workflows/** - GitHub Actions workflows
  - **scripts/** - GitHub-specific scripts
- **internal/** - Internal implementation code
  - **fs/** - Filesystem implementation
    - **cache/** - Cache implementation
    - **operations/** - File/directory operations
    - **offline/** - Offline mode functionality
    - **upload/** - Upload management
  - **ui/** - GUI implementation
    - **systemd/** - Systemd integration for the UI
  - **nemo/** - Nemo file manager integration
- **pkg/** - Reusable public packages
  - **errors/** - Error handling utilities
  - **graph/** - Microsoft Graph API integration
  - **logging/** - Logging utilities
  - **quickxorhash/** - QuickXORHash implementation
  - **testutil/** - Testing utilities
  - **util/** - General utilities
- **assets/** - Application assets
  - **icons/** - Icon files
  - **examples/** - Example configuration files
- **deployments/** - Deployment configurations
  - **systemd/** - Systemd service files
  - **desktop/** - Desktop entry files
- **build/** - Build artifacts and scripts
  - **cli/** - Command-line interface tools
  - **package/** - Packaging scripts
    - **deb/** - Debian packaging
    - **rpm/** - RPM packaging
  - **dev/** - Development tools
    - **ai-assistant/** - AI assistant scripts
    - **tools/** - Developer utility scripts
- **.junie/** - Junie AI assistant configuration
- **.run/** - GoLand run configurations
- **scripts/** - Utility scripts (legacy, being migrated to build/)
- **tmp/** - Temporary files and script output

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
go test ./internal/fs/...
go test ./cmd/...
go test ./internal/ui/...
```

### JetBrains GoLand Run Configurations

The project includes predefined run configurations for JetBrains GoLand. These configurations are stored in the `.run/` directory.

Available run configurations:
- **all** - Builds all project binaries using the Makefile's "all" target
- **onemount-launcher** - Builds the onemount-launcher application using the Makefile
- **onemount** - Builds the onemount application using the Makefile
- **Test - Test Utils** - Runs tests in the pkg/testutil package
- **Unit Test - File System** - Runs unit tests in the internal/fs package and its subpackages that match the pattern "TestUT*"

To use these configurations:
1. Open the project in GoLand
2. Go to the Run/Debug Configurations dropdown in the toolbar
3. Select the desired configuration and click the Run button

## Executing Scripts

Developer scripts should be placed in the `build/dev/tools` directory. Script output should be directed to the `tmp/` directory.

- **cgo-helper.sh** - Helps with CGO compilation
- **curl-graph.sh** - Utility for interacting with Microsoft Graph API

## Best Practices

The OneMount project follows a set of comprehensive coding standards and best practices. For detailed guidelines, refer to the documents in the `docs/guides/` directory:

- [Coding Standards](coding-standards.md) - Main entry point for all coding standards
- [Design Guidelines](design-guidelines.md) - Guidelines for designing components
- [Logging Guidelines](logging-guidelines.md) - Guidelines for structured logging
- [Logging Examples](logging-examples.md) - Examples of proper logging
- [Test Guidelines](testing/test-guidelines.md) - Best practices for writing tests
- [Threading Guidelines](threading-guidelines.md) - Guidelines for concurrent programming
- [D-Bus Integration](dbus-integration.md) - Guidelines for D-Bus integration

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
   - Update existing documentation rather than adding new documentation
   - If new documentation is needed, place it in the 'docs/' folder or relevant sub-folder
   - Add links to new documentation in relevant existing documentation
   - Always check existing 'docs/' documentation for relevant information to a task
   - Include Junie prompts for implementing next steps, recommendations, actions, etc.

6. **Method Logging**
   - Use the method logging framework for all public methods
   - Follow the patterns in `logging-guidelines.md`
   - Log method entry and exit, including parameters and return values

7. **D-Bus Integration**
   - Use the D-Bus interface for file status updates
   - Follow the specification in `dbus-integration.md`
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

- Use `journalctl --user -u onemount@.service --since today` to view logs
- Set the environment variable `ONEMOUNT_DEBUG=1` for verbose logging
- Use `fusermount3 -uz $MOUNTPOINT` to unmount the filesystem if it hangs
