# Go Logging Guidelines for OneMount

This guide provides comprehensive recommendations for logging in Go applications, with a focus on the onemount project. It covers structured logging, log levels, the method logging framework, context propagation, performance considerations, and error handling.

## General Principles

1. **Use structured logging**: Log in a format that is machine-parseable (JSON) with consistent field names.
2. **Include context**: Add relevant context to logs to make them more useful for debugging.
3. **Be consistent**: Use consistent log levels and field names throughout the application.
4. **Consider performance**: Logging can impact performance, especially in high-throughput applications.
5. **Respect privacy**: Avoid logging sensitive information like authentication tokens or personal data.

## Structured Logging with zerolog

The onemount project uses [zerolog](https://github.com/rs/zerolog) for structured logging, which is a good choice for performance and usability.

### Best Practices

1. **Use typed fields instead of string formatting**:

   ```go
   // Good
   log.Info().
       Str("user", username).
       Int("items", count).
       Msg("User purchased items")

   // Avoid
   log.Info().Msgf("User %s purchased %d items", username, count)
   ```

2. **Use consistent field names**:

   Define a set of standard field names for common concepts and use them consistently:

   ```go
   // Standard field names
   const (
       FieldMethod    = "method"     // Method or function name
       FieldOperation = "operation"  // Higher-level operation
       FieldComponent = "component"  // Component or module
       FieldDuration  = "duration"   // Duration of operation in milliseconds
       FieldError     = "error"      // Error message
       FieldPath      = "path"       // File or resource path
       FieldID        = "id"         // Identifier
       FieldUser      = "user"       // User identifier
       FieldStatus    = "status"     // Status code or string
       FieldSize      = "size"       // Size in bytes
       FieldGoroutine = "goroutine"  // Goroutine ID
       FieldPhase     = "phase"      // Phase of operation (e.g., "entry", "exit")
   )
   ```

3. **Group related fields**:

   Use sub-objects for related fields to improve readability:

   ```go
   log.Info().
       Str(FieldOperation, "file_upload").
       Dict("file", zerolog.Dict().
           Str("name", filename).
           Int("size", size).
           Str("mime", mimeType)).
       Dict("user", zerolog.Dict().
           Str("id", userID).
           Str("ip", ipAddress)).
       Msg("File uploaded")
   ```

## Log Levels

zerolog provides several log levels. Use them consistently according to these guidelines:

1. **Trace**: Very detailed information, useful for debugging specific issues.
   - Method entry/exit with parameters and return values
   - Internal state changes
   - Detailed algorithm steps

2. **Debug**: Information useful for debugging during development.
   - Configuration settings at startup
   - Cache operations
   - API request/response details

3. **Info**: General information about application operation.
   - Application startup/shutdown
   - User actions
   - Successful operations
   - Periodic status updates

4. **Warn**: Potential issues that don't prevent the application from working.
   - Deprecated feature usage
   - Recoverable errors
   - Performance issues
   - Unexpected but handled conditions

5. **Error**: Issues that prevent a specific operation from completing.
   - API request failures
   - Database errors
   - File system errors
   - Network connectivity issues

6. **Fatal**: Critical issues that prevent the application from starting or continuing.
   - Configuration errors that prevent startup
   - Required resource unavailability
   - Unrecoverable errors

### Example

```go
// Trace - detailed method tracing
func SomeMethod() {
    methodName, startTime := LogMethodEntry("SomeMethod")
    defer LogMethodExit(methodName, time.Since(startTime))
    // Method implementation...
}

// Debug - configuration details
log.Debug().
    Str("cache_dir", config.CacheDir).
    Str("log_level", config.LogLevel).
    Bool("sync_tree", config.SyncTree).
    Msg("Configuration loaded")

// Info - normal operations
log.Info().
    Str("mountpoint", mountpoint).
    Msg("Filesystem mounted successfully")

// Warn - potential issues
log.Warn().
    Str("feature", "legacy_auth").
    Msg("Using deprecated authentication method")

// Error - operation failures
log.Error().
    Err(err).
    Str("file", filename).
    Msg("Failed to read file")

// Fatal - application cannot continue
log.Fatal().
    Err(err).
    Msg("Failed to initialize database, cannot continue")
```

## Method Logging Framework

The onemount project implements a method logging framework that provides a way to log method entry and exit, including parameters and return values, for all public methods in the core module.

### Overview

The logging framework consists of two main components:

1. `LogMethodEntry()` - A function that logs method entry with its parameters.
2. `LogMethodExit()` - A function that logs method exit, including return values and execution duration.

For context-aware method logging, the framework provides:

1. `LogMethodEntryWithContext()` - A function that logs method entry with context.
2. `LogMethodExitWithContext()` - A function that logs method exit with context.

There's also a helper function `LoggedMethod()` that wraps a function call with entry and exit logging.

These functions use the zerolog library to produce structured logs that can be easily parsed and analyzed.

## How to Use

To add logging to a method, follow these patterns:

### For methods with simple return values

```go
func (f *Filesystem) IsOffline() bool {
    methodName, startTime := LogMethodEntry("IsOffline")
    f.RLock()
    defer f.RUnlock()

    result := f.offline
    defer LogMethodExit(methodName, time.Since(startTime), result)
    return result
}
```

### For methods with error returns

```go
func (f *Filesystem) TrackOfflineChange(change *OfflineChange) error {
    methodName, startTime := LogMethodEntry("TrackOfflineChange", change)
    defer func() {
        // We can't capture the return value directly in a defer, so we'll just log completion
        LogMethodExit(methodName, time.Since(startTime))
    }()

    // Method implementation...
    return someError
}
```

### For methods with pointer returns

```go
func (f *Filesystem) GetNodeID(nodeID uint64) *Inode {
    methodName, startTime := LogMethodEntry("GetNodeID", nodeID)

    // Early return case
    if someCondition {
        defer LogMethodExit(methodName, time.Since(startTime), nil)
        return nil
    }

    result := someOperation()
    defer LogMethodExit(methodName, time.Since(startTime), result)
    return result
}
```

### For methods with multiple return values

For methods with multiple return values, you'll need to use named return values and a defer function:

```go
func (f *Filesystem) SomeMethod() (result1 Type1, result2 Type2, err error) {
    methodName, startTime := LogMethodEntry("SomeMethod")
    defer func() {
        LogMethodExit(methodName, time.Since(startTime), result1, result2, err)
    }()

    // Method implementation...
    result1 = ...
    result2 = ...
    err = ...
    return
}
```

### Using context-aware method logging

For methods that need to include context information:

```go
func (f *Filesystem) ProcessWithContext(ctx LogContext, data []byte) error {
    methodName, startTime, logger, ctx := LogMethodEntryWithContext("ProcessWithContext", ctx)

    // Method implementation...

    // Log method exit with context
    LogMethodExitWithContext(methodName, startTime, logger, ctx, err)
    return err
}
```

## Methods to Instrument

The following methods should be instrumented with logging:

### Filesystem Methods

- IsOffline
- TrackOfflineChange
- ProcessOfflineChanges
- TranslateID
- GetNodeID
- InsertNodeID
- GetID
- InsertID
- InsertChild
- DeleteID
- GetChild
- GetChildrenID
- GetChildrenPath
- GetPath
- DeletePath
- InsertPath
- MoveID
- MovePath
- StartCacheCleanup
- StopCacheCleanup
- StopDeltaLoop
- StopDownloadManager
- StopUploadManager
- SerializeAll

### Inode Methods

- AsJSON
- String
- Name
- SetName
- NodeID
- SetNodeID
- ID
- ParentID
- Path
- HasChanges
- HasChildren
- IsDir
- Mode
- ModTime
- NLink
- Size

## Log Output

The logs produced by this framework include:

- Method name
- Entry/exit phase
- Goroutine ID (thread identifier)
- Parameters (for entry)
- Return values (for exit)
- Execution duration (for exit)

Example log entry:
```json
{"level":"debug","method":"IsOffline","phase":"entry","goroutine":"1","time":"2023-04-27T21:00:00Z","message":"Method called"}
```
```json
{"level":"debug","method":"IsOffline","phase":"exit","goroutine":"1","duration_ms":0.123,"return1":false,"time":"2023-04-27T21:00:00Z","message":"Method completed"}
```

The `goroutine` field contains the ID of the goroutine (Go's lightweight thread) that executed the method. This is useful for tracking method calls across different threads, especially in concurrent operations.

## Testing

The logging framework includes tests in `method_logging_test.go` that verify:

1. Basic functionality of the logging functions
2. Integration with instrumented methods

Run the tests with:
```bash
go test -v ./internal/logging/...
```

## Context Propagation

Context propagation is important for tracking related log entries across different functions and goroutines.

### Using Request IDs

For operations that span multiple functions or goroutines, use a unique identifier:

```go
// Generate a request ID
requestID := uuid.New().String()

// Add it to all related log entries
log.Info().
    Str("request_id", requestID).
    Str("operation", "sync").
    Msg("Starting sync operation")

// Pass it to other functions
syncFiles(ctx, requestID)

// In the called function
func syncFiles(ctx context.Context, requestID string) {
    log.Info().
        Str("request_id", requestID).
        Str("operation", "sync_files").
        Msg("Syncing files")
}
```

### Using LogContext

The logging package provides a `LogContext` struct for propagating logging context:

```go
// Create a context with a logger
ctx := NewLogContext("sync_operation").
    WithRequestID(requestID).
    WithUserID(userID)

// Pass the context to other functions
processRequest(ctx)

// In the called function
func processRequest(ctx LogContext) {
    // Get a logger with the context
    logger := ctx.Logger()

    // Log with the context already included
    logger.Info().Str("operation", "process").Msg("Processing request")

    // Pass to other functions
    validateInput(ctx)
}
```

## Performance Considerations

Logging can impact performance, especially in high-throughput applications.

### Best Practices

1. **Avoid expensive operations in log statements**:

   ```go
   // Bad - JSON marshaling happens even if debug is disabled
   log.Debug().Msg("Request body: " + string(json.Marshal(body)))

   // Good - Only executes if debug is enabled
   if IsDebugEnabled() {
       bodyJSON, _ := json.Marshal(body)
       log.Debug().RawJSON("body", bodyJSON).Msg("Request body")
   }
   ```

2. **Use sampling for high-volume logs**:

   ```go
   // Only log 1% of these messages
   if rand.Float64() < 0.01 {
       log.Debug().Msg("High volume operation")
   }
   ```

3. **Optimize reflection-based logging**:

   The method logging uses reflection to log parameters and return values, which can be expensive. The following utilities are provided to optimize this:

   - Type-specific logging helpers in `type_helpers.go` for common types (string, int, bool, etc.)
   - Enhanced type caching mechanism in `type_cache.go` to reduce reflection overhead
   - Conditional logging functions in `performance.go` to avoid expensive operations when not needed

## Error Handling and Logging

Proper error handling and logging is crucial for debugging and monitoring.

### Best Practices

1. **Always include the error**:

   ```go
   // Good
   if err != nil {
       log.Error().Err(err).Msg("Failed to open file")
   }

   // Avoid
   if err != nil {
       log.Error().Msg("Failed to open file: " + err.Error())
   }
   ```

2. **Add context to errors**:

   ```go
   // Good
   if err != nil {
       log.Error().
           Err(err).
           Str("file", filename).
           Int("attempt", attempt).
           Msg("Failed to open file")
   }
   ```

3. **Use error wrapping**:

   ```go
   if err != nil {
       return fmt.Errorf("failed to process file %s: %w", filename, err)
   }
   ```

4. **Log at the appropriate level**:

   ```go
   // For expected errors that are handled
   if err == ErrNotFound {
       log.Debug().Str("id", id).Msg("Item not found, creating new one")
       // Handle the error...
   }

   // For unexpected errors that affect the operation
   if err != nil {
       log.Error().Err(err).Str("id", id).Msg("Failed to retrieve item")
       // Handle the error...
   }
   ```

## Implementation in OneMount

### Current Implementation

The onemount project uses zerolog for structured logging and has a consolidated logging package structure:

- `logger.go`: Core logger implementation and level management
- `context.go`: Context-aware logging functionality
- `method.go`: Method entry/exit logging (both with and without context)
- `error.go`: Error logging functionality
- `performance.go`: Performance optimization utilities
- `constants.go`: Constants used throughout the logging package
- `console_writer.go`: Console writer functionality
- `structured_logging.go`: Structured logging functions

### API Overview

The logging package provides a simplified and consistent API for logging in the OneMount project. The key components are:

1. **Core Logging API** (`logger.go`):
   - `Logger` struct: The main logger type
   - `Event` struct: Represents a log event
   - Log level functions: `Debug()`, `Info()`, `Warn()`, `Error()`, `Fatal()`, `Panic()`, `Trace()`
   - Level management: `SetGlobalLevel()`, `IsDebugEnabled()`, `IsTraceEnabled()`

2. **Method Logging API** (`method.go`):
   - Basic method logging: `LogMethodEntry()`, `LogMethodExit()`
   - Context-aware method logging: `LogMethodEntryWithContext()`, `LogMethodExitWithContext()`
   - Helper functions: `LoggedMethod()`, `WithMethodLogging()`, `WithMethodLoggingAndContext()`

3. **Error Logging API** (`error.go`):
   - Basic error logging: `LogError()`, `LogErrorWithFields()`
   - Warning-level error logging: `LogErrorAsWarn()`, `LogErrorAsWarnWithFields()`
   - Context-aware error logging: `LogErrorWithContext()`
   - Error wrapping: `WrapAndLogError()`, `WrapAndLogErrorWithContext()`

4. **Context-Aware Logging API** (`context.go`):
   - `LogContext` struct: For propagating logging context
   - Context building methods: `WithRequestID()`, `WithUserID()`, `WithComponent()`, etc.
   - Logger creation: `Logger()` method to create a logger with context

5. **Performance Optimization API** (`performance.go`):
   - Level checks: `IsLevelEnabled()`, `LogIfEnabled()`
   - Complex object logging: `LogComplexObjectIfDebug()`, `LogComplexObjectIfTrace()`, `LogComplexObjectIfEnabled()`
   - Type caching: `getTypeName()`, `getTypeKind()`, `getTypeElem()`

6. **Structured Logging API** (`structured_logging.go`):
   - Context-aware logging at different levels: `LogInfoWithContext()`, `LogDebugWithContext()`, `LogTraceWithContext()`
   - Error enrichment: `EnrichErrorWithContext()`

7. **Constants** (`constants.go`):
   - Standard field names: `FieldMethod`, `FieldOperation`, `FieldComponent`, etc.
   - Method logging specific fields: `FieldReturn`, `FieldParam`
   - Phase values: `PhaseEntry`, `PhaseExit`
   - Message templates: `MsgMethodCalled`, `MsgMethodCompleted`

### Recommended Patterns

1. **Use structured logging with typed fields**:
   ```go
   Info().
       Str(FieldMethod, "UploadFile").
       Str(FieldPath, path).
       Int(FieldSize, len(data)).
       Msg("File uploaded successfully")
   ```

2. **Use context-aware logging for operations that span multiple functions**:
   ```go
   ctx := NewLogContext("sync_operation").
       WithRequestID(requestID).
       WithUserID(userID)

   LogMethodEntryWithContext("ProcessChanges", ctx)
   ```

3. **Check log levels before expensive operations**:
   ```go
   if IsDebugEnabled() {
       details := generateDetailedReport(data)
       Debug().Interface("report", details).Msg("Generated detailed report")
   }
   ```

4. **Use the standardized error logging functions**:
   ```go
   // For basic error logging
   LogError(err, "Failed to read file", FieldPath, path)

   // For context-aware error logging
   LogErrorWithContext(err, ctx, "Failed to read file", FieldPath, path)
   ```

5. **Use the helper functions for method logging**:
   ```go
   // For simple methods
   results := LoggedMethod(calculateHash, data)

   // For methods with context
   results := WithMethodLoggingAndContext("CalculateHash", ctx, calculateHash, data)
   ```

## Conclusion

Following these best practices will improve the quality and usefulness of logs in the onemount project, making debugging and monitoring easier while maintaining good performance. The method logging framework provides a consistent way to instrument methods with entry and exit logging, which is particularly valuable for tracing execution flow and diagnosing issues in a complex filesystem implementation.
