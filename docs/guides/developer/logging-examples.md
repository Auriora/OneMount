# Logging Examples

This document provides examples of how to use the improved logging framework in the OneMount project.

## Basic Method Logging

The simplest way to add logging to a method is to use the `LogMethodEntry` and `LogMethodExit` functions:

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

This logs the method entry and exit, including the return value and execution duration.

## Context-Aware Logging

For operations that span multiple functions or goroutines, use the context-aware logging functions:

### Basic Context-Aware Logging

```go
func (f *Filesystem) ProcessChanges(requestID string) error {
    // Create a logging context
    ctx := NewLogContext("process_changes").
        WithRequestID(requestID)

    // Log method entry with context
    methodName, startTime, logger, ctx := LogMethodEntryWithContext("ProcessChanges", ctx)

    // Use the logger for additional logs within the method
    logger.Info().Str(FieldPath, "/some/path").Msg("Processing changes for path")

    // Process changes...
    err := f.processChangesInternal(ctx)

    // Log method exit with context and return value
    LogMethodExitWithContext(methodName, startTime, logger, ctx, err)
    return err
}

func (f *Filesystem) processChangesInternal(ctx LogContext) error {
    // Log method entry with the same context
    methodName, startTime, logger, ctx := LogMethodEntryWithContext("processChangesInternal", ctx)

    // Use the logger for additional logs
    logger.Debug().Msg("Internal processing started")

    // Process changes...
    if err := someOperation(); err != nil {
        // Log errors with context
        LogErrorWithContext(err, ctx, "Failed to process changes", 
            FieldPath, "/some/path", 
            FieldID, "123")
        return err
    }

    // Log method exit with context
    LogMethodExitWithContext(methodName, startTime, logger, ctx, nil)
    return nil
}
```

### Enhanced Context-Aware Logging with Request IDs and User IDs

For operations that span multiple functions or user-initiated operations, use the enhanced context-aware logging functions:

```go
func (f *Filesystem) ProcessUserRequest(path string) error {
    // Create a logging context with a unique request ID and the current user's ID
    ctx := NewLogContextWithRequestAndUserID("process_user_request").
        WithPath(path)

    // Log method entry with context
    methodName, startTime, logger, ctx := LogMethodEntryWithContext("ProcessUserRequest", ctx)

    // Use the logger for additional logs within the method
    logger.Info().Msg("Processing user request for path")

    // Process the request...
    err := f.processRequestInternal(ctx)

    // Log method exit with context and return value
    LogMethodExitWithContext(methodName, startTime, logger, ctx, err)
    return err
}

func (f *Filesystem) processRequestInternal(ctx LogContext) error {
    // Log method entry with the same context
    // The request ID and user ID are automatically propagated
    methodName, startTime, logger, ctx := LogMethodEntryWithContext("processRequestInternal", ctx)

    // Use the logger for additional logs
    logger.Debug().Msg("Internal processing started")

    // Process the request...
    if err := someOperation(); err != nil {
        // Log errors with context
        // The request ID and user ID are included in the error logs
        LogErrorWithContext(err, ctx, "Failed to process request")
        return err
    }

    // Log method exit with context
    LogMethodExitWithContext(methodName, startTime, logger, ctx, nil)
    return nil
}
```

## Error Logging

### Basic Error Logging

Use the `LogError` function to log errors with additional context:

```go
func (f *Filesystem) ReadFile(path string) ([]byte, error) {
    // ... method implementation ...

    data, err := os.ReadFile(path)
    if err != nil {
        // Log error with additional context
        LogError(err, "Failed to read file", 
            FieldPath, path, 
            FieldSize, fileSize)
        return nil, err
    }

    return data, nil
}
```

### Context-Aware Error Logging

Use the `LogErrorWithContext` function to log errors with a logging context:

```go
func (f *Filesystem) ReadFile(path string, ctx LogContext) ([]byte, error) {
    methodName, startTime, logger, ctx := LogMethodEntryWithContext("ReadFile", ctx)

    // ... method implementation ...

    data, err := os.ReadFile(path)
    if err != nil {
        // Log error with context
        LogErrorWithContext(err, ctx, "Failed to read file", 
            FieldPath, path, 
            FieldSize, fileSize)
        LogMethodExitWithContext(methodName, startTime, logger, ctx, nil, err)
        return nil, err
    }

    LogMethodExitWithContext(methodName, startTime, logger, ctx, len(data), nil)
    return data, nil
}
```

### Advanced Error Logging

For more advanced error logging scenarios, use the additional error logging utilities:

```go
// Log an error with a map of fields
fields := map[string]interface{}{
    FieldPath: path,
    FieldSize: fileSize,
    "retry_count": retryCount,
}
LogErrorWithFields(err, "Failed to upload file", fields)

// Log a warning with fields
LogWarnWithFields("File not found in cache, downloading from server", 
    map[string]interface{}{
        FieldPath: path,
        "cache_status": "miss",
    })

// Log a warning with an error (using the recommended function)
LogErrorAsWarn(err, "Retrying operation after error", 
    FieldPath, path, 
    "retry_count", retryCount)

// Log an error and return it separately (recommended pattern)
LogError(err, "Failed to process file", 
    FieldPath, path, 
    FieldSize, fileSize)
return err

// Log an error with context and return it separately (recommended pattern)
LogErrorWithContext(err, ctx, "Failed to process file", 
    FieldPath, path, 
    FieldSize, fileSize)
return err

// Format an error with additional context
return FormatErrorWithContext(err, "Failed to process file", 
    FieldPath, path, 
    FieldSize, fileSize)
```

## Performance Optimization

### Level Checks

For expensive logging operations, check if the log level is enabled first:

```go
func (f *Filesystem) ProcessLargeData(data []byte) error {
    // ... method implementation ...

    // Only log large data if debug is enabled
    if IsDebugEnabled() {
        Debug().
            Int(FieldSize, len(data)).
            Str(FieldPath, path).
            Msg("Processing large data")
    }

    // ... continue processing ...
}
```

### Helper Functions

Use the helper functions for level checks and complex object logging:

```go
// Check if debug logging is enabled
if IsDebugEnabled() {
    // Perform expensive operation only if debug is enabled
    details := generateDetailedReport(data)
    Debug().
        Interface("report", details).
        Msg("Generated detailed report")
}

// Log complex objects only if debug is enabled
LogComplexObjectIfDebug("fileStats", stats, "File statistics")

// Log complex objects only if trace is enabled
LogComplexObjectIfTrace("requestDetails", request, "Request details")
```

### Type Caching and Type-Specific Logging

For reflection-based logging, use the enhanced type caching mechanism and type-specific logging helpers:

```go
// Get the type name of an object using the enhanced cache
typeName := GetTypeName(reflect.TypeOf(obj))
Debug().
    Str("type", typeName).
    Msg("Processing object")

// Use type-specific logging helpers for common types
event := Debug()
event = LogParam(event, 0, "string value")  // Logs a string parameter
event = LogParam(event, 1, 42)              // Logs an int parameter
event = LogParam(event, 2, true)            // Logs a bool parameter
event.Msg("Parameters logged with type-specific helpers")

// Use the enhanced type caching mechanism for type checks
t := reflect.TypeOf(obj)
if IsStruct(t) {
    Debug().Str("type", "struct").Msg("Object is a struct")
} else if IsPointer(t) {
    Debug().Str("type", "pointer").Msg("Object is a pointer")
} else if IsSlice(t) {
    Debug().Str("type", "slice").Msg("Object is a slice")
}
```

## Standardized Field Names

Use the standardized field names defined in `constants.go` for consistent logging:

```go
Info().
    Str(FieldMethod, "UploadFile").
    Str(FieldPath, path).
    Int(FieldSize, len(data)).
    Str(FieldID, fileID).
    Msg("File uploaded successfully")
```

## Grouping Related Fields

Group related fields using the `Dict` method for better readability:

```go
Info().
    Str(FieldOperation, "file_upload").
    Dict("file", zerolog.Dict().
        Str("name", filename).
        Int(FieldSize, size).
        Str("mime", mimeType)).
    Dict("user", zerolog.Dict().
        Str(FieldID, userID).
        Str("ip", ipAddress)).
    Msg("File uploaded")
```

## Log Levels

Use the appropriate log level for different types of information:

```go
// Trace - very detailed information
Trace().
    Str(FieldMethod, "calculateHash").
    Int(FieldSize, len(data)).
    Msg("Calculating hash for data")

// Debug - information useful for debugging
Debug().
    Str(FieldPath, cachePath).
    Int("expiration_days", expirationDays).
    Msg("Cache configuration")

// Info - general information about application operation
Info().
    Str(FieldPath, mountpoint).
    Msg("Filesystem mounted successfully")

// Warn - potential issues that don't prevent the application from working
Warn().
    Str(FieldPath, path).
    Msg("File not found in cache, downloading from server")

// Error - issues that prevent a specific operation from completing
Error().
    Err(err).
    Str(FieldPath, path).
    Msg("Failed to download file")

// Fatal - critical issues that prevent the application from starting or continuing
Fatal().
    Err(err).
    Msg("Failed to initialize filesystem, cannot continue")
```

## Using the LoggedMethod Helper

For simple method logging, you can use the `LoggedMethod` helper function:

```go
func (f *Filesystem) CalculateHash(data []byte) string {
    // Use LoggedMethod to wrap the function call
    results := LoggedMethod(f.calculateHashInternal, data)

    // Extract the return value from the results
    return results[0].(string)
}

func (f *Filesystem) calculateHashInternal(data []byte) string {
    // Implementation without explicit logging
    hash := sha256.Sum256(data)
    return hex.EncodeToString(hash[:])
}
```

## Conclusion

By following these examples and using the standardized logging functions and field names, you can improve the consistency and usefulness of logs in the OneMount project, making debugging and monitoring easier while maintaining good performance.
