# Contributing to onedriver

Thank you for your interest in contributing to onedriver! This document provides guidelines and instructions for contributing to the project.

## Table of Contents

1. [Code of Conduct](#code-of-conduct)
2. [Getting Started](#getting-started)
3. [Development Environment](#development-environment)
4. [Building the Project](#building-the-project)
5. [Running Tests](#running-tests)
6. [Coding Standards](#coding-standards)
7. [Pull Request Process](#pull-request-process)
8. [Reporting Bugs](#reporting-bugs)
9. [Feature Requests](#feature-requests)

## Code of Conduct

This project and everyone participating in it is governed by the [onedriver Code of Conduct](CODE_OF_CONDUCT.md). By participating, you are expected to uphold this code. Please report unacceptable behavior to the project maintainers.

## Getting Started

1. Fork the repository on GitHub
2. Clone your fork locally
3. Add the original repository as a remote named "upstream"
4. Create a new branch for your changes

```bash
git clone https://github.com/yourusername/onedriver.git
cd onedriver
git remote add upstream https://github.com/bcherrington/onedriver.git
git checkout -b feature/your-feature-name
```

## Development Environment

Before you begin development, ensure you have the necessary dependencies installed:

### Required Dependencies

- Go programming language
- GCC compiler
- webkit2gtk-4.0 and json-glib development headers

#### On Fedora:
```bash
dnf install golang gcc pkg-config webkit2gtk3-devel json-glib-devel
```

#### On Ubuntu:
```bash
apt install golang gcc pkg-config libwebkit2gtk-4.0-dev libjson-glib-dev
```

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

The tests will write and delete files/folders on your OneDrive account at the path `/onedriver_tests`. Note that the offline test suite requires `sudo` to remove network access to simulate being offline.

```bash
# Setup test environment for first time run
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

## Coding Standards

The onedriver project follows a set of comprehensive coding standards and best practices. For detailed guidelines, refer to the documents in the `docs/guides/` directory:

- [Coding Standards](docs/guides/coding-standards.md) - Main entry point for all coding standards
- [Go Coding Standards](docs/guides/go-coding-standards.md) - Comprehensive guide for Go code
- [Go Logging Best Practices](docs/guides/go-logging-best-practices.md) - Guidelines for structured logging
- [Test Best Practices](docs/guides/test-guidelines.md) - Best practices for writing tests

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

## Pull Request Process

1. Ensure your code follows the project's coding standards
2. Update documentation as necessary
3. Add or update tests as appropriate
4. Ensure all tests pass
5. Submit a pull request with a clear description of the changes

## Reporting Bugs

When reporting bugs, please include:

1. A clear and descriptive title
2. Steps to reproduce the issue
3. Expected behavior
4. Actual behavior
5. Log output (`journalctl --user -u onedriver@.service --since today`)
6. Your Linux distribution and version

## Feature Requests

Feature requests are welcome. Please provide:

1. A clear and descriptive title
2. A detailed description of the proposed feature
3. Any relevant use cases
4. If possible, a suggestion for how to implement the feature
