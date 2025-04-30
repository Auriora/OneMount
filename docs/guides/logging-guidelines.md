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
    methodName, startTime := LogMethodCall()
    defer LogMethodReturn(methodName, startTime)
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

1. `LogMethodCall()` - A function that logs method entry and returns the method name and start time.
2. `LogMethodReturn()` - A function that logs method exit, including return values and execution duration.

These functions use the zerolog library to produce structured logs that can be easily parsed and analyzed.

## How to Use

To add logging to a method, follow these patterns:

### For methods with simple return values

```go
func (f *Filesystem) IsOffline() bool {
    methodName, startTime := LogMethodCall()
    f.RLock()
    defer f.RUnlock()

    result := f.offline
    defer LogMethodReturn(methodName, startTime, result)
    return result
}
```

### For methods with error returns

```go
func (f *Filesystem) TrackOfflineChange(change *OfflineChange) error {
    methodName, startTime := LogMethodCall()
    defer func() {
        // We can't capture the return value directly in a defer, so we'll just log completion
        LogMethodReturn(methodName, startTime)
    }()

    // Method implementation...
    return someError
}
```

### For methods with pointer returns

```go
func (f *Filesystem) GetNodeID(nodeID uint64) *Inode {
    methodName, startTime := LogMethodCall()

    // Early return case
    if someCondition {
        defer LogMethodReturn(methodName, startTime, nil)
        return nil
    }

    result := someOperation()
    defer LogMethodReturn(methodName, startTime, result)
    return result
}
```

### For methods with multiple return values

For methods with multiple return values, you'll need to use named return values and a defer function:

```go
func (f *Filesystem) SomeMethod() (result1 Type1, result2 Type2, err error) {
    methodName, startTime := LogMethodCall()
    defer func() {
        LogMethodReturn(methodName, startTime, result1, result2, err)
    }()

    // Method implementation...
    result1 = ...
    result2 = ...
    err = ...
    return
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
{"level":"debug","method":"IsOffline","phase":"exit","goroutine":"1","duration_ms":0.123,"return1":false,"time":"2023-04-27T21:00:00Z","message":"Method completed"}
```

The `goroutine` field contains the ID of the goroutine (Go's lightweight thread) that executed the method. This is useful for tracking method calls across different threads, especially in concurrent operations.

## Testing

The logging framework includes tests in `logging_test.go` that verify:

1. Basic functionality of the logging functions
2. Integration with instrumented methods

Run the tests with:
```bash
go test -v ./fs/...
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

### Using Context

Go's `context.Context` can be used to propagate request-scoped values, including logging context:

```go
// Create a context with a logger
ctx := log.With().
    Str("request_id", requestID).
    Str("user", userID).
    Logger().WithContext(context.Background())

// Pass the context to other functions
processRequest(ctx)

// In the called function
func processRequest(ctx context.Context) {
    // Get the logger from context
    logger := log.Ctx(ctx)

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
   if log.Debug().Enabled() {
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

   The current method logging uses reflection to log parameters and return values, which can be expensive. Consider caching type information or providing type-specific logging helpers.

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

The onemount project uses zerolog for structured logging and has a method logging framework that logs method entry and exit with parameters and return values.

### Recommendations for Improvement

1. **Standardize field names**:
   - Define constants for common field names
   - Use consistent field names across the codebase

2. **Enhance context propagation**:
   - Use request IDs for operations that span multiple functions
   - Consider using context.Context for propagating logging context

3. **Optimize performance**:
   - Add level checks before expensive logging operations
   - Consider caching type information for reflection-based logging

4. **Improve error logging**:
   - Always include the error using Err()
   - Add relevant context to error logs
   - Use error wrapping for better error context

5. **Document logging standards**:
   - Update the logging documentation with these best practices
   - Provide examples for common logging scenarios

## Conclusion

Following these best practices will improve the quality and usefulness of logs in the onemount project, making debugging and monitoring easier while maintaining good performance. The method logging framework provides a consistent way to instrument methods with entry and exit logging, which is particularly valuable for tracing execution flow and diagnosing issues in a complex filesystem implementation.
