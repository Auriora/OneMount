# onedriver Coding Standards

This document serves as the comprehensive guide for all coding standards and best practices in the onedriver project. It provides detailed guidelines to ensure consistency, maintainability, and quality across the codebase.

## Table of Contents

1. [Overview](#overview)
2. [Go Standards](#go-standards)
   1. [Code Organization](#code-organization)
   2. [Project Structure](#project-structure)
   3. [Naming Conventions](#naming-conventions)
   4. [Constants and Literals](#constants-and-literals)
   5. [Formatting and Style](#formatting-and-style)
   6. [Error Handling](#error-handling)
   7. [Comments and Documentation](#comments-and-documentation)
   8. [Testing](#testing)
   9. [Logging](#logging)
   10. [Performance Considerations](#performance-considerations)
   11. [Concurrency](#concurrency)
   12. [Security Best Practices](#security-best-practices)
   13. [Dependencies](#dependencies)
   14. [Version Control](#version-control)
3. [Shell Script Standards](#shell-script-standards)
4. [Python Standards](#python-standards)
5. [Documentation Standards](#documentation-standards)
6. [Version Control Standards](#version-control-standards)
7. [Code Review Standards](#code-review-standards)
8. [Security Standards](#security-standards)
9. [Performance Standards](#performance-standards)
10. [Accessibility Standards](#accessibility-standards)
11. [Conclusion](#conclusion)

## Overview

The onedriver project follows industry best practices for software development, with a focus on code quality, maintainability, and performance. These standards ensure consistency across the codebase and make it easier for new developers to contribute to the project.

For specific aspects of development, we also provide additional detailed guides:
- [Go Logging Best Practices](go_logging_best_practices.md) - Guidelines for structured logging with zerolog
- [Test Best Practices](test_best_practices.md) - Best practices for writing effective tests

## Go Standards

The majority of the onedriver codebase is written in Go. This section provides comprehensive guidelines for Go development.

### Code Organization

#### Package Structure

- Follow the standard Go project layout:
  - `/cmd` - Main applications
  - `/fs` - Filesystem implementation
  - `/ui` - User interface code
  - `/pkg` - Code that can be used by external applications
  - `/testutil` - Testing utilities

- Keep packages focused on a single responsibility
- Avoid circular dependencies between packages
- Use internal packages for code that shouldn't be imported by other projects

### Project Structure

In this guideline, we synthesize the community-driven **Standard Go Project Layout** patterns, the **official Go module** recommendations, and expert advice from Smart Byte Labs and Alex Edwards to propose a flexible yet consistent structure for your OneDriver repository.

#### 1. Core Principles

##### 1.1 Embrace Go Modules  
All projects should use Go Modules. Place a `go.mod` file in the repository root to declare your module path and track dependencies, eliminating `$GOPATH` constraints.

##### 1.2 Start Small and Evolve  
Begin with only `main.go` and `go.mod` in the root for prototypes or small tools. As functionality grows, let your code organically drive the creation of packages and directories—avoid over-structuring upfront.

##### 1.3 Prioritize Effectiveness Over Perfection  
Aim for a structure that is easy to navigate, supports testing, and scales with team size. Resist the urge to chase an elusive "perfect" layout—iteratively refine as real-world needs emerge.

#### 2. Standard Skeletons

##### 2.1 Basic Layout  
For simple libraries or command-line tools with minimal assets:
```
├── go.mod
├── main.go
├── foo.go
├── foo_test.go
└── README.md
```

##### 2.2 Medium-Sized Project  
For projects with multiple packages and commands:
```
├── cmd/
│   ├── myapp/
│   │   └── main.go
│   └── myutil/
│       └── main.go
├── internal/
│   ├── config/
│   │   └── config.go
│   └── handler/
│       └── handler.go
├── pkg/
│   └── mylib/
│       └── mylib.go
├── go.mod
└── README.md
```

##### 2.3 Complex Project  
For large applications with multiple components:
```
├── api/
│   └── openapi.yaml
├── cmd/
│   ├── server/
│   │   └── main.go
│   └── client/
│       └── main.go
├── configs/
│   └── app.yaml
├── docs/
│   └── README.md
├── internal/
│   ├── auth/
│   │   └── auth.go
│   ├── server/
│   │   └── server.go
│   └── storage/
│       └── storage.go
├── pkg/
│   └── api/
│       └── client.go
├── scripts/
│   └── setup.sh
├── web/
│   ├── static/
│   │   └── style.css
│   └── template/
│       └── index.html
├── go.mod
└── README.md
```

### Naming Conventions

- Use descriptive, meaningful names
- Follow Go's standard naming conventions:
  - `camelCase` for unexported identifiers
  - `PascalCase` for exported identifiers
  - `snake_case` for file names
  - `ALL_CAPS` for constants that are truly unchanging

- Package names should be:
  - Short, concise, and descriptive
  - Lower case, single word (no underscores)
  - Not plural (e.g., `item` not `items`)

- Interface names:
  - Should describe behavior, not implementation
  - Often end with `-er` (e.g., `Reader`, `Writer`)
  - Single-method interfaces named after the method with `-er` suffix

- Avoid stutter in names:
  - Don't repeat package name in identifiers (e.g., `fs.FSFile` → `fs.File`)

- Use consistent abbreviations:
  - Common abbreviations like `HTTP`, `URL`, `ID` should be used consistently
  - Treat abbreviations as single words for capitalization (e.g., `HttpClient` → `HTTPClient`)

### Constants and Literals

- Group related constants in a const block
- Use iota for related constants when appropriate
- Prefer named constants over magic numbers
- Use raw string literals (``) for regular expressions and multi-line strings
- Define string constants for error messages to ensure consistency

### Formatting and Style

- Use `gofmt` or `goimports` to format code
- Follow the style recommendations in [Effective Go](https://golang.org/doc/effective_go.html)
- Limit line length to 100-120 characters for better readability
- Group imports into standard library, external packages, and internal packages
- Use blank lines to separate logical sections of code
- Align struct field tags for better readability

### Error Handling

- Always check errors
- Return errors rather than using panic
- Use error wrapping to add context: `fmt.Errorf("failed to open file: %w", err)`
- Create custom error types for specific error conditions
- Use sentinel errors (`var ErrNotFound = errors.New("not found")`) for expected error conditions
- Avoid using `_` to ignore errors unless you have a good reason
- Log errors at the appropriate level (debug, info, warn, error)
- Consider using error handling packages like `pkg/errors` for more complex applications

### Comments and Documentation

- Write godoc-compatible comments for all exported functions, types, and constants
- Begin comments with the name of the thing being described
- Use complete sentences with proper punctuation
- Add examples where appropriate
- Include package-level documentation in a file named `doc.go`
- Comment complex or non-obvious code
- Keep comments up-to-date with code changes
- Use `// TODO: description` or `// FIXME: description` for temporary comments

### Testing

- Write tests for all exported functions and methods
- Use table-driven tests for testing multiple cases
- Use subtests for organizing test cases: `t.Run("case name", func(t *testing.T) { ... })`
- Use testify/assert for cleaner assertions
- Mock external dependencies for unit tests
- Aim for high test coverage, but focus on critical paths
- Write benchmarks for performance-critical code
- Use examples in documentation that also serve as tests
- For more detailed guidelines, see [Test Best Practices](test_best_practices.md)

### Logging

- Use structured logging with zerolog
- Log at the appropriate level (debug, info, warn, error)
- Include relevant context in log messages
- Avoid logging sensitive information
- Use log levels consistently across the codebase
- Consider performance implications of logging in hot paths
- For more detailed guidelines, see [Go Logging Best Practices](go_logging_best_practices.md)

### Performance Considerations

- Profile before optimizing
- Avoid premature optimization
- Minimize allocations in hot paths
- Use sync.Pool for frequently allocated objects
- Consider using buffered I/O for better performance
- Be mindful of API rate limits and network latency
- Use appropriate data structures for the task
- Consider memory usage and allocations

### Concurrency

- Use goroutines judiciously
- Always use synchronization primitives (mutex, channels) to protect shared data
- Prefer channels for communication, mutexes for state
- Consider using sync.WaitGroup for managing groups of goroutines
- Be aware of goroutine leaks and ensure proper cleanup
- Use context for cancellation and timeouts
- Consider using worker pools for limiting concurrency
- Be careful with closures in goroutines (capture variables explicitly)

### Security Best Practices

- Never store sensitive information (tokens, passwords) in plaintext
- Validate all user input
- Use secure defaults
- Keep dependencies up to date
- Follow the principle of least privilege
- Be cautious with file permissions
- Sanitize file paths to prevent path traversal attacks
- Use crypto/rand for generating random values, not math/rand

### Dependencies

- Minimize external dependencies
- Pin dependency versions in go.mod
- Regularly update dependencies for security fixes
- Consider vendoring dependencies for reproducible builds
- Evaluate the quality, maintenance status, and license of dependencies before adding them
- Prefer standard library solutions when available

### Version Control

- Write clear, descriptive commit messages
- Keep commits focused on a single change
- Use feature branches for new development
- Run tests before committing
- Review code before merging

## Shell Script Standards

For shell scripts (bash), follow these guidelines:

- Use shellcheck to validate scripts
- Include a shebang line (`#!/bin/bash`)
- Use double quotes around variables to prevent word splitting
- Use meaningful variable names
- Add comments for complex operations
- Check for error conditions and handle them appropriately

## Python Standards

For Python code (used in testing and utilities), follow these guidelines:

- Follow PEP 8 style guide
- Use Python 3 features
- Document functions and classes with docstrings
- Use type hints where appropriate
- Handle exceptions properly

## Documentation Standards

- Use Markdown for documentation
- Keep documentation up-to-date with code changes
- Include examples where appropriate
- Use proper headings and formatting for readability
- Link related documents together

## Version Control Standards

- Write clear, descriptive commit messages
- Keep commits focused on a single change
- Use feature branches for new development
- Run tests before committing
- Review code before merging

## Code Review Standards

- Review all code changes before merging
- Check for adherence to coding standards
- Verify that tests are included and pass
- Look for potential security issues
- Ensure documentation is updated

## Security Standards

- Never store sensitive information (tokens, passwords) in plaintext
- Validate all user input
- Use secure defaults
- Keep dependencies up to date
- Follow the principle of least privilege
- Be cautious with file permissions
- Sanitize file paths to prevent path traversal attacks

## Performance Standards

- Profile before optimizing
- Avoid premature optimization
- Consider memory usage and allocations
- Use appropriate data structures for the task
- Be mindful of API rate limits and network latency

## Accessibility Standards

- Ensure UI elements are accessible
- Provide keyboard shortcuts for common actions
- Use high-contrast colors
- Support screen readers
- Test with accessibility tools

## Conclusion

Following these coding standards will help maintain a consistent, high-quality codebase for the onedriver project. These standards should be applied to all new code and, when feasible, to existing code during refactoring.

For questions or suggestions regarding these standards, please open an issue or pull request in the project repository.