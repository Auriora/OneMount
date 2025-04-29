# Go Coding Standards for onedriver

This document outlines the coding standards and best practices for the onedriver project. It serves as a comprehensive guide for developers to ensure consistency, maintainability, and quality across the codebase.

## Table of Contents

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

## Code Organization

### Package Structure

- Follow the standard Go project layout:
  - `/cmd` - Main applications
  - `/fs` - Filesystem implementation
  - `/ui` - User interface code
  - `/pkg` - Code that can be used by external applications
  - `/testutil` - Testing utilities

- Keep packages focused on a single responsibility
- Avoid circular dependencies between packages
- Use internal packages for code that shouldn't be imported by other projects

## Project Structure

In this guideline, we synthesize the community-driven **Standard Go Project Layout** patterns, the **official Go module** recommendations, and expert advice from Smart Byte Labs and Alex Edwards to propose a flexible yet consistent structure for your OneDriver repository.

### 1. Core Principles

#### 1.1 Embrace Go Modules  
All projects should use Go Modules. Place a `go.mod` file in the repository root to declare your module path and track dependencies, eliminating `$GOPATH` constraints.

#### 1.2 Start Small and Evolve  
Begin with only `main.go` and `go.mod` in the root for prototypes or small tools. As functionality grows, let your code organically drive the creation of packages and directories—avoid over-structuring upfront.

#### 1.3 Prioritize Effectiveness Over Perfection  
Aim for a structure that is easy to navigate, supports testing, and scales with team size. Resist the urge to chase an elusive "perfect" layout—iteratively refine as real-world needs emerge.

### 2. Standard Skeletons

#### 2.1 Basic Layout  
For simple libraries or command-line tools with minimal assets:
```
├── go.mod
├── main.go
├── foo.go
├── foo_test.go
└── README.md
```  
Use this when only one package exists; it keeps discovery trivial.

#### 2.2 Supporting-Packages Layout  
When you need internal helpers but still have a single `main`:
```
├── go.mod
├── main.go
├── internal/
│   └── helpers/
│       └── helpers.go
└── README.md
```  
Leverage the `internal/` directory to restrict package use to your module, enabling safe refactoring.

#### 2.3 Server (Cmd-Internal-Assets) Layout  
For larger applications with multiple executables or non-Go assets:
```
├── go.mod
├── cmd/
│   ├── onedriver/
│   │   └── main.go
│   └── cli/
│       └── main.go
├── internal/
│   └── storage/
│       └── storage.go
├── pkg/             # optional, for public libraries
├── configs/         # YAML/JSON configurations
├── migrations/      # DB migrations
├── scripts/         # build/test/deploy scripts
└── README.md
```  
This clear separation of **cmd/**, **internal/**, and other assets aligns with the official Go module guidelines for medium-to-large codebases.

### 3. Directory Roles

#### 3.1 `cmd/`  
Houses entry points—each subdirectory under `cmd/` produces a distinct executable. Names should match the binary name, e.g., `onedriver`.

#### 3.2 `internal/`  
Contains code not intended for external consumption. Go's compiler enforces import restrictions on `internal/`, safeguarding your module's private APIs.

#### 3.3 `pkg/` (Optional)  
Place libraries you intend to share externally. If OneDriver exports reusable functionality, reside it here; otherwise, you can omit `pkg/` to avoid confusion.

#### 3.4 Ancillary Directories  
- **`configs/`**: Configuration templates and defaults  
- **`migrations/`**: Database schema evolution  
- **`scripts/`**: CI/CD, linting, or deployment scripts  
- **`docs/`**: Design documents, API specs, and usage guides

### 4. Naming Conventions

#### 4.1 Packages  
Choose **short, clear, lowercase** names without underscores or mixedCaps. Each package name should reflect its purpose concisely, e.g., `storage`, `auth`, `cmd`.

#### 4.2 Files  
Group related types and functions in the same file. Avoid creating directories solely to tidy files—only create a new package when there's a clear need.

#### 4.3 Modules and Paths  
Set the module path in `go.mod` to your repository root, e.g., `module github.com/bcherrington/onedriver`, so imports resolve correctly.

### 5. Testing and Documentation

#### 5.1 Tests  
Place tests adjacent to implementation in `<file>_test.go`. Keep test packages consistent with code packages to access unexported identifiers when needed.

#### 5.2 Documentation  
Write `godoc`-style comments immediately before declarations. Maintain a high-level README.md describing setup, usage, and contribution guidelines.

### 6. Evolving the Layout

- **Monitor warning signs**: frequent import cycles, difficulty locating code, or monolithic packages may indicate the need for refactoring.
- **Iterate**: As OneDriver grows—adding features like condition monitoring or IoT integration—refine the project skeleton, splitting or merging packages as warranted.
- **Balance**: Apply the principle "effective, not perfect" to adapt structure pragmatically without over-engineering.

### File Organization

- Group related functionality in the same file
- Keep files to a reasonable size (generally under 500 lines)
- Organize file content in the following order:
  1. Package declaration
  2. Imports
  3. Constants
  4. Variables
  5. Types/Structs
  6. Methods
  7. Functions

### Interface Design

- Define interfaces at the point of use, not at the point of implementation
- Keep interfaces small and focused on a single responsibility
- Use interfaces to decouple components and facilitate testing

## Naming Conventions

### General Guidelines

- Use meaningful, descriptive names
- Prefer clarity over brevity
- Be consistent with existing code

### Specific Naming Rules

- **Packages**: Use short, lowercase names without underscores (e.g., `fs`, `graph`)
- **Files**: Use lowercase with underscores for multi-word names (e.g., `drive_item.go`)
- **Variables**:
  - Use camelCase for local variables
  - Use descriptive names that indicate purpose
  - Avoid single-letter variables except for short-lived loop counters
- **Constants**: Use MixedCaps or ALL_CAPS for constants
- **Functions and Methods**:
  - Use MixedCaps (e.g., `GetItem`, `IsDir`)
  - Begin with a verb for functions that perform actions
  - Use Get/Set prefix for accessor/mutator methods
  - Use Is/Has/Can prefix for boolean methods
- **Types/Structs**: Use MixedCaps (e.g., `DriveItem`, `Filesystem`)
- **Interfaces**: Use MixedCaps, often ending with 'er' (e.g., `Reader`, `Writer`)
- **Exported vs. Unexported**:
  - Capitalize exported names (visible outside the package)
  - Use lowercase for unexported names (package-private)

## Constants and Literals

### Minimizing Literals

- Avoid using literals (string, numeric, boolean) directly in code
- Define constants for values that are used multiple times or have special meaning
- Place constants at the package level near the top of the file, after imports
- Group related constants together
- Consider creating a dedicated constants file for values used across multiple files

### When to Use Constants

- Magic numbers (e.g., `const MaxRetries = 3` instead of using `3` directly)
- Status codes and error messages
- Configuration values
- File paths and URLs
- Regular expression patterns
- Any value that might change in the future

### Constants Organization

- **File-level constants**: Place at the top of the file after imports
  ```go
  package example

  import "time"

  const (
      defaultTimeout = 30 * time.Second
      maxRetries     = 3
      apiVersion     = "v1.0"
  )
  ```

- **Package-level constants**: Create a dedicated file (e.g., `constants.go`) for constants used across multiple files in the same package
  ```go
  // constants.go
  package mypackage

  const (
      // StatusCodes represents HTTP status codes used in the application
      StatusOK           = 200
      StatusBadRequest   = 400
      StatusUnauthorized = 401
      StatusNotFound     = 404
  )
  ```

- **Project-wide constants**: Create a dedicated package for constants used across multiple packages
  ```go
  // pkg/constants/api.go
  package constants

  const (
      // API endpoints
      BaseURL     = "https://api.example.com"
      AuthEndpoint = "/auth"
      UserEndpoint = "/users"
  )
  ```

### Examples

#### Bad Practice (Using Literals)

```go
package example

import (
    "errors"
    "strings"
)

// Bad practice: Using string literals directly in code
func isNameRestricted(name string) bool {
    if strings.EqualFold(name, "CON") {
        return true
    }
    if strings.EqualFold(name, "AUX") {
        return true
    }
    if strings.EqualFold(name, "PRN") {
        return true
    }
    // More literals...
    return false
}

// Bad practice: Using string literals in error handling
func checkFileName(name string) error {
    // Some logic...
    if isRestricted(name) {
        return errors.New("nameAlreadyExists")
    }
    return nil
}

func handleFile(name string) {
    err := checkFileName(name)
    if err != nil && strings.Contains(err.Error(), "nameAlreadyExists") {
        // Handle error...
    }
}

func isRestricted(name string) bool {
    return name == "restricted"
}
```

#### Good Practice (Using Constants)

```go
package example

import (
    "errors"
    "strings"
)

const (
    // RestrictedNames are file names that are not allowed in Windows
    RestrictedNameCON = "CON"
    RestrictedNameAUX = "AUX"
    RestrictedNamePRN = "PRN"
    // More constants...

    // ErrorMessages
    ErrNameAlreadyExists = "nameAlreadyExists"
)

// Good practice: Using constants instead of literals
func isNameRestricted(name string) bool {
    if strings.EqualFold(name, RestrictedNameCON) {
        return true
    }
    if strings.EqualFold(name, RestrictedNameAUX) {
        return true
    }
    if strings.EqualFold(name, RestrictedNamePRN) {
        return true
    }
    // More checks using constants...
    return false
}

// Good practice: Using constants in error handling
func checkFileName(name string) error {
    // Some logic...
    if isRestricted(name) {
        return errors.New(ErrNameAlreadyExists)
    }
    return nil
}

func handleFile(name string) {
    err := checkFileName(name)
    if err != nil && strings.Contains(err.Error(), ErrNameAlreadyExists) {
        // Handle error...
    }
}

func isRestricted(name string) bool {
    return name == "restricted"
}
```

### Benefits of Using Constants

- Improves code readability and maintainability
- Reduces the risk of typos and errors
- Makes it easier to change values in the future
- Provides a central place to document the meaning of values
- Helps with code completion in IDEs

## Formatting and Style

- Use `gofmt` or `goimports` to format code
- Follow the [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Limit line length to 100 characters when possible
- Use blank lines to separate logical sections of code
- Group related imports and separate standard library, third-party, and local imports

## Error Handling

- Always check errors returned from functions
- Return errors rather than using panic (except in truly unrecoverable situations)
- Use error wrapping to add context: `fmt.Errorf("failed to process file %s: %w", filename, err)`
- Be specific about error messages to aid debugging
- For detailed error handling guidelines, see the [Error Handling section in the Logging Best Practices](go-logging-best-practices.md#error-handling-and-logging)

## Comments and Documentation

### Code Comments

- Write comments for non-obvious code
- Focus on explaining "why" rather than "what" (the code should be self-explanatory)
- Keep comments up-to-date with code changes
- Use complete sentences with proper punctuation

### Documentation Comments

- Document all exported functions, types, and constants
- Follow the godoc format for documentation comments
- Begin function documentation with the function name
- Include examples where appropriate

Example:

```
// GetItem retrieves a DriveItem by its ID.
// It returns the DriveItem and any error encountered during the request.
// If the item doesn't exist, it returns nil and an appropriate error.
func GetItem(id string, auth *Auth) (*DriveItem, error) {
    // Implementation...
}
```

## Testing

For detailed testing guidelines, refer to the [Test Best Practices](test-best-practices.md) document. Key points include:

- Write both unit and integration tests
- Use table-driven tests for multiple test cases
- Use testify for assertions and test setup
- Test edge cases and error conditions
- Keep tests independent and idempotent
- Aim for high test coverage, especially for critical paths

## Logging

For detailed logging guidelines, refer to the [Go Logging Best Practices](go-logging-best-practices.md) document. Key points include:

- Use structured logging with zerolog
- Log at appropriate levels (trace, debug, info, warn, error, fatal)
- Include relevant context in log entries
- Be consistent with field names
- Consider performance implications of logging

## Performance Considerations

- Profile before optimizing
- Avoid premature optimization
- Consider memory usage and allocations
- Use buffered I/O for file operations
- Cache expensive computations
- Be mindful of API rate limits and network latency
- Use appropriate data structures for the task

### Specific to onedriver

- Minimize network requests to Microsoft Graph API
- Cache filesystem metadata and file contents
- Use concurrent operations where appropriate
- Implement efficient delta synchronization

## Concurrency

- Use goroutines judiciously
- Protect shared resources with appropriate synchronization (mutex, channels)
- Be aware of race conditions
- Consider using sync.WaitGroup for managing groups of goroutines
- Use context for cancellation and timeouts
- Follow the [threading guidelines](threading-guidelines.md)

## Security Best Practices

- Never store sensitive information (tokens, passwords) in plaintext
- Validate all user input
- Use secure defaults
- Keep dependencies up to date
- Follow the principle of least privilege
- Be cautious with file permissions
- Sanitize file paths to prevent path traversal attacks

## Dependencies

- Minimize external dependencies
- Vendor dependencies or use Go modules
- Regularly update dependencies for security fixes
- Evaluate the quality and maintenance status of dependencies before adding them
- Consider the license of dependencies

## Version Control

- Write clear, descriptive commit messages
- Keep commits focused on a single change
- Use feature branches for new development
- Run tests before committing
- Review code before merging

## Conclusion

Following these coding standards will help maintain a consistent, high-quality codebase for the onedriver project. These standards should be applied to all new code and, when feasible, to existing code during refactoring.

For specific aspects of development, refer to the more detailed best practice documents:
- [Test Best Practices](test-best-practices.md)
- [Go Logging Best Practices](go-logging-best-practices.md)
