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
    - **graph/** - Microsoft Graph API integration
    - **offline/** - Offline mode functionality
  - **ui/** - GUI implementation
    - **systemd/** - Systemd integration for the UI
  - **nemo/** - Nemo file manager integration
  - **testutil/** - Testing utilities
- **.junie/** - Junie AI assistant configuration
- **.run/** - GoLand run configurations
- **scripts/** - Utility scripts
  - **ai-assistant/** - AI assistant scripts
  - **debian/** - Debian packaging scripts
  - **developer/** - Developer utility scripts
- **docker/** - Docker configurations
  - **compose/** - Docker Compose files
- **build/** - Build artifacts
  - **binaries/** - Compiled executables
  - **packages/** - Package files (deb, rpm)
  - **temp/** - Temporary build files
- **deployments/** - Deployment configurations
  - **desktop/** - Desktop entry files
  - **systemd/** - Systemd service files
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
# Build the main binaries (outputs to build/binaries/)
make build  # Alias for 'make all'

# Install the application system-wide
sudo make install-system

# Create distribution packages
make rpm    # For RPM-based distributions
make deb    # For Debian-based distributions (uses Docker)
```

### Build Directory Structure

The project now uses a structured build directory:

```
build/
├── binaries/           # Compiled executables
├── packages/           # All package formats
│   ├── deb/           # Debian/Ubuntu packages
│   ├── rpm/           # RPM packages
│   └── source/        # Source tarballs
└── temp/              # Temporary build files
```

## Running Tests

```bash
# Run all tests
make test

# Run specific test categories
make unit-test         # Unit tests only
make integration-test  # Integration tests only
make system-test       # Basic system tests

# Run system tests with real OneDrive account
make system-test-real  # Comprehensive tests
make system-test-all   # All test categories

# Run Docker-based tests
make docker-test-unit        # Unit tests in Docker
make docker-test-integration # Integration tests in Docker
make docker-test-system      # System tests in Docker
make docker-test-all         # All tests in Docker
```

### System Tests

The project includes a comprehensive system test suite that uses a real OneDrive account for end-to-end testing. To set up system tests:

```bash
# Set up authentication (one-time)
make onemount
./build/binaries/onemount --auth-only
mkdir -p ~/.onemount-tests
cp ~/.cache/onemount/auth_tokens.json ~/.onemount-tests/.auth_tokens.json
```

### JetBrains GoLand Run Configurations

The project includes predefined run configurations for JetBrains GoLand. These configurations are stored in the `.run/` directory.

Available run configurations:
- **all** - Builds all project binaries using the Makefile's "all" target
- **onemount-launcher** - Builds the onemount-launcher application using the Makefile
- **onemount** - Builds the onemount application using the Makefile
- **Test - Test Utils** - Runs tests in the internal/testutil package
- **Unit Test - File System** - Runs unit tests in the internal/fs package that match the pattern "TestUT*"

To use these configurations:
1. Open the project in GoLand
2. Go to the Run/Debug Configurations dropdown in the toolbar
3. Select the desired configuration and click the Run button

## Docker Support

The project includes Docker configurations for development, testing, and CI/CD:

```bash
# Build and run tests in Docker
make docker-test-build  # Build Docker test image
make docker-test-all    # Run all tests in Docker

# Run GitHub Actions self-hosted runner
./scripts/manage-runner.sh setup
./scripts/manage-runner.sh start
```

## Version Management

The project uses bump2version for version management:

```bash
# Install bump2version
python3 -m venv .venv
.venv/bin/pip install bump2version

# Bump version
.venv/bin/bumpversion patch  # 0.1.0 -> 0.1.1
.venv/bin/bumpversion minor  # 0.1.0 -> 0.2.0
.venv/bin/bumpversion major  # 0.1.0 -> 1.0.0

# Create release candidates
.venv/bin/bumpversion --new-version 0.2.0rc1 minor  # Start new RC
.venv/bin/bumpversion num                           # 0.2.0rc1 -> 0.2.0rc2
.venv/bin/bumpversion release                       # 0.2.0rc1 -> 0.2.0
```

## CI/CD Workflows

The project has several GitHub Actions workflows:

1. **Continuous Integration** (ci.yml)
   - Runs on every push to main and pull requests
   - Runs tests, linting, and basic builds

2. **Package Building** (build-packages.yml)
   - Triggered by version tags (e.g., v0.1.0)
   - Builds packages and creates GitHub releases

3. **System Tests** (system-tests.yml)
   - Manual or scheduled execution
   - End-to-end testing with real OneDrive accounts

## Best Practices

The OneMount project follows a set of comprehensive coding standards and best practices. For detailed guidelines, refer to the documents in the `docs/guides/` directory:

- [Coding Standards](../docs/guides/coding-standards.md) - Main entry point for all coding standards
- [Design Guidelines](../docs/guides/design-guidelines.md) - Guidelines for designing components
- [Logging Guidelines](../docs/guides/logging-guidelines.md) - Guidelines for structured logging
- [Logging Examples](../docs/guides/logging-examples.md) - Examples of proper logging
- [Test Guidelines](../docs/guides/testing/test-guidelines.md) - Best practices for writing tests
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
   - Use system tests for end-to-end validation

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
- **Creating a release**: Bump version, push tags, let CI/CD create the release

## Debugging

- Use `journalctl --user -u onemount@.service --since today` to view logs
- Set the environment variable `ONEMOUNT_DEBUG=1` for verbose logging
- Use `fusermount3 -uz $MOUNTPOINT` to unmount the filesystem if it hangs
