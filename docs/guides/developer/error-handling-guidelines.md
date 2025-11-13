# Error Handling Guidelines for OneMount

This guide provides comprehensive recommendations for error handling in Go applications, with a focus on the OneMount project. It covers error creation, error wrapping, error checking, and structured error logging.

## General Principles

1. **Use structured errors**: Create errors with context that helps identify where and why they occurred.
2. **Wrap errors**: Preserve the error chain by wrapping errors with additional context.
3. **Check errors appropriately**: Use the right error checking functions for different error types.
4. **Log errors with context**: Include relevant context when logging errors.
5. **Use specialized error types**: Use specialized error types for common error scenarios.

## Error Creation and Wrapping

The OneMount project uses the `internal/common/errors` package for standardized error handling.

### Creating Errors

Use the following functions to create errors:

```go
// Create a simple error
err := errors.New("something went wrong")

// Create a specialized error
err := errors.NewNotFoundError("resource not found", nil)
err := errors.NewNetworkError("network unavailable", underlyingErr)
```

### Wrapping Errors

Always wrap errors with context when returning them up the call stack:

```go
// Wrap an error with context
if err != nil {
    return errors.Wrap(err, "failed to open file")
}

// Wrap an error with formatted context
if err != nil {
    return errors.Wrapf(err, "failed to open file %s", filename)
}
```

### Error Types

The following specialized error types are available:

| Error Type | Constructor | Description |
|------------|-------------|-------------|
| NetworkError | `NewNetworkError` | Network-related errors (connection issues, etc.) |
| NotFoundError | `NewNotFoundError` | Resource not found errors |
| AuthError | `NewAuthError` | Authentication or authorization errors |
| ValidationError | `NewValidationError` | Input validation errors |
| OperationError | `NewOperationError` | Operation failures |
| TimeoutError | `NewTimeoutError` | Timeout errors |
| ResourceBusyError | `NewResourceBusyError` | Resource busy or locked errors |

### Checking Error Types

Use the following functions to check error types:

```go
// Check if an error is a specific type
if errors.IsNotFoundError(err) {
    // Handle not found error
}

// Check if an error matches a specific error
if errors.Is(err, io.EOF) {
    // Handle EOF error
}

// Extract a specific error type
var typedErr *MyErrorType
if errors.As(err, &typedErr) {
    // Use typedErr
}
```

## Structured Error Logging

The OneMount project uses structured logging with context for errors.

### Log Context

Use `LogContext` to provide context for error logs:

```go
// Create a log context
ctx := errors.NewLogContext("operation_name").
    WithMethod("method_name").
    WithPath("/path/to/resource").
    With("custom_field", "custom_value")

// Log an error with context
errors.LogErrorWithContext(err, ctx, "error message")
```

### Logging Functions

The following logging functions are available:

```go
// Log an error with context
errors.LogErrorWithContext(err, ctx, "error message")

// Log a warning with context
errors.LogWarnWithContext(err, ctx, "warning message")

// Log an info message with context
errors.LogInfoWithContext(ctx, "info message")

// Log a debug message with context
errors.LogDebugWithContext(ctx, "debug message")

// Log a trace message with context
errors.LogTraceWithContext(ctx, "trace message")
```

### Helper Functions

The following helper functions combine common error handling patterns:

```go
// Wrap an error, log it with context, and return it
err = errors.WrapAndLogErrorWithContext(err, ctx, "wrapped error message")

// Log an error with context and return it
err = errors.LogAndReturnWithContext(err, ctx, "error message")

// Add context to an error without logging it
err = errors.EnrichErrorWithContext(err, ctx, "enriched error message")
```

## Best Practices

### Error Creation

1. **Be specific**: Create errors with specific, actionable messages.
2. **Include context**: Include relevant context in error messages (file names, IDs, etc.).
3. **Use specialized types**: Use specialized error types for common error scenarios.

### Error Wrapping

1. **Always wrap errors**: Wrap errors with context when returning them up the call stack.
2. **Preserve the error chain**: Use `errors.Wrap` to preserve the error chain.
3. **Add meaningful context**: Add context that helps identify where and why the error occurred.

### Error Checking

1. **Use the right function**: Use `errors.Is` for sentinel errors, `errors.As` for error types, and specialized functions for specific error types.
2. **Check for specific errors**: Check for specific errors when you can handle them differently.
3. **Default to general handling**: Default to general error handling when specific handling isn't needed.

### Error Logging

1. **Log with context**: Always include relevant context when logging errors.
2. **Log at the appropriate level**: Use the appropriate log level for different error scenarios.
3. **Log once**: Log an error only once, typically at the point where it's handled.

## Examples

### Basic Error Handling

```go
func ReadFile(filename string) ([]byte, error) {
    file, err := os.Open(filename)
    if err != nil {
        return nil, errors.Wrap(err, "failed to open file")
    }
    defer file.Close()

    data, err := io.ReadAll(file)
    if err != nil {
        return nil, errors.Wrap(err, "failed to read file")
    }

    return data, nil
}
```

### Specialized Error Types

```go
func GetResource(id string) (*Resource, error) {
    resource, exists := resources[id]
    if !exists {
        return nil, errors.NewNotFoundError("resource not found", nil)
    }
    return resource, nil
}

func ProcessResource(id string) error {
    resource, err := GetResource(id)
    if errors.IsNotFoundError(err) {
        // Handle not found error specifically
        return errors.Wrap(err, "resource does not exist")
    }
    if err != nil {
        // Handle other errors
        return errors.Wrap(err, "failed to get resource")
    }

    // Process the resource
    return nil
}
```

### Structured Logging

```go
func ProcessRequest(req *Request) error {
    // Create a log context
    ctx := errors.NewLogContext("process_request").
        WithMethod("ProcessRequest").
        WithPath(req.Path).
        With("request_id", req.ID)

    // Log the start of the operation
    errors.LogInfoWithContext(ctx, "Processing request")

    // Process the request
    result, err := processRequestInternal(req)
    if err != nil {
        // Log and return the error with context
        return errors.WrapAndLogErrorWithContext(err, ctx, "failed to process request")
    }

    // Log the successful completion
    ctx = ctx.With("result", result)
    errors.LogInfoWithContext(ctx, "Request processed successfully")

    return nil
}
```

## Conclusion

Following these guidelines will improve the quality and consistency of error handling in the OneMount project. By using structured errors, proper error wrapping, appropriate error checking, and context-rich logging, we can make our code more maintainable and our errors more actionable.
