# Error Handling Examples

This document provides examples of how to use the standardized error handling approach in the OneMount project. It includes examples of creating errors, wrapping errors, checking errors, and logging errors with context.

## Creating Errors

### Basic Error Creation

```go
// Create a simple error
err := errors.New("something went wrong")
```

### Specialized Error Types

```go
// Create a validation error
err := errors.NewValidationError("invalid thumbnail size: large", nil)

// Create a network error
err := errors.NewNetworkError("cannot fetch thumbnails in offline mode", nil)

// Create a not found error
err := errors.NewNotFoundError("resource not found", nil)

// Create an authentication error
err := errors.NewAuthError("invalid credentials", nil)

// Create an operation error
err := errors.NewOperationError("failed to process request", nil)

// Create a timeout error
err := errors.NewTimeoutError("request timed out", nil)

// Create a resource busy error
err := errors.NewResourceBusyError("resource is locked", nil)
```

## Wrapping Errors

### Basic Error Wrapping

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

### Wrapping with Specialized Error Types

```go
// Wrap an error as a validation error
if err != nil {
    return errors.NewValidationError("invalid input", err)
}

// Wrap an error as a network error
if err != nil {
    return errors.NewNetworkError("failed to connect", err)
}
```

## Checking Errors

### Checking Error Types

```go
// Check if an error is a specific type
if errors.IsNotFoundError(err) {
    // Handle not found error
}

// Check if an error is a network error
if errors.IsNetworkError(err) {
    // Handle network error
}

// Check if an error matches a specific error
if errors.Is(err, io.EOF) {
    // Handle EOF error
}

// Extract a specific error type
var validationErr *errors.TypedError
if errors.As(err, &validationErr) && validationErr.Type == errors.ErrorTypeValidation {
    // Use validationErr
}
```

## Logging Errors

### Basic Error Logging

```go
// Log an error with context
errors.LogError(err, "Failed to open file", 
    errors.FieldPath, "/path/to/file",
    errors.FieldOperation, "OpenFile")

// Log a warning with context
errors.LogWarn(err, "Using fallback method", 
    errors.FieldOperation, "ProcessRequest")
```

### Structured Logging with Context

```go
// Create a log context
ctx := errors.NewLogContext("process_request").
    WithMethod("ProcessRequest").
    WithPath("/api/v1/resource").
    With("request_id", "12345")

// Log an error with context
errors.LogErrorWithContext(err, ctx, "Failed to process request")

// Log and return an error with context
return errors.LogAndReturnWithContext(err, ctx, "Failed to process request")

// Wrap, log, and return an error with context
return errors.WrapAndLogWithContext(err, ctx, "Failed to process request")
```

## Real-World Examples

### Thumbnail Operations

```go
// Validate thumbnail size
if size != "small" && size != "medium" && size != "large" {
    return nil, errors.NewValidationError(fmt.Sprintf("invalid thumbnail size: %s", size), nil)
}

// Handle offline mode
if f.IsOffline() {
    return nil, errors.NewNetworkError("cannot fetch thumbnails in offline mode", nil)
}

// Handle API errors
thumbnailData, err := graph.GetThumbnailContent(inode.ID(), size, f.auth)
if err != nil {
    return nil, errors.Wrap(err, "failed to get thumbnail")
}

// Log errors with context
if err := f.thumbnails.Insert(inode.ID(), size, thumbnailData); err != nil {
    errors.LogError(err, "Failed to cache thumbnail", 
        errors.FieldID, inode.ID(),
        "size", size,
        errors.FieldOperation, "GetThumbnail")
}
```

### File Operations

```go
// Handle file open errors
fd, err := f.content.Open(i.DriveItem.ID)
if err != nil {
    errors.LogError(err, "Failed to open file for truncation", 
        errors.FieldID, i.DriveItem.ID,
        errors.FieldOperation, "SetAttr.truncate",
        errors.FieldPath, path)
    i.Unlock()
    return fuse.EIO
}

// Handle file truncate errors
if err := fd.Truncate(int64(size)); err != nil {
    errors.LogError(err, "Failed to truncate file", 
        errors.FieldID, i.DriveItem.ID,
        errors.FieldOperation, "SetAttr.truncate",
        errors.FieldPath, path,
        "size", size)
    i.Unlock()
    return fuse.EIO
}
```

### API Operations

```go
// Handle API errors
if err = graph.Rename(id, newName, newParentID, f.auth); err != nil {
    errors.LogError(err, "Failed to rename remote item", 
        errors.FieldOperation, "Rename.remoteRename",
        errors.FieldID, id,
        errors.FieldPath, path,
        "dest", dest,
        "newName", newName,
        "newParentID", newParentID)
    return fuse.EREMOTEIO
}
```

## Best Practices

1. **Use specialized error types** for common error scenarios to make error handling more precise.
2. **Wrap errors with context** to preserve the error chain and provide more information about where and why the error occurred.
3. **Use consistent field names** for logging to make logs more consistent and easier to parse.
4. **Add relevant context** to error logs to make them more useful for debugging.
5. **Check for specific error types** when you can handle them differently.
6. **Log errors at the appropriate level** based on their severity and impact.
7. **Use structured logging with context** to propagate context between functions.

## Conclusion

By following these examples and best practices, you can ensure consistent, maintainable, and debuggable error handling throughout the OneMount project. If you have any questions or need further guidance, refer to the [Error Handling Guidelines](../docs/guides/error-handling-guidelines.md) or ask for help in the developer chat.