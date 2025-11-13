# Error Recovery Guidelines for OneMount

This guide provides recommendations for implementing error recovery mechanisms in the OneMount project. It covers retry strategies, error classification, and best practices for handling transient failures.

## General Principles

1. **Classify errors**: Determine which errors are transient and can be retried.
2. **Use exponential backoff**: Increase the delay between retries to avoid overwhelming the system.
3. **Add jitter**: Add randomness to retry delays to prevent thundering herd problems.
4. **Set maximum retries**: Limit the number of retry attempts to avoid infinite loops.
5. **Log retry attempts**: Log each retry attempt with relevant context for debugging.

## Retry Utility Package

The OneMount project includes a centralized retry utility package (`internal/retry`) that provides a consistent approach to error recovery. This package should be used for all operations that may encounter transient failures.

### Retry Configuration

The retry package provides a `Config` struct for configuring retry parameters:

```
type Config struct {
    // MaxRetries is the maximum number of retry attempts
    MaxRetries int

    // InitialDelay is the initial delay between retries
    InitialDelay time.Duration

    // MaxDelay is the maximum delay between retries
    MaxDelay time.Duration

    // Multiplier is the factor by which the delay increases after each retry
    Multiplier float64

    // Jitter is the maximum random jitter added to the delay
    Jitter float64

    // RetryableErrors is a list of error types that should be retried
    RetryableErrors []RetryableError
}
```

The default configuration can be obtained using `DefaultConfig()`:

```
config := retry.DefaultConfig()
```

This provides a reasonable starting point with:
- 3 maximum retries
- 1 second initial delay
- 30 second maximum delay
- 2.0 multiplier (exponential backoff)
- 0.2 jitter (20% randomness)
- Retryable errors include network errors, server errors, and rate limit errors

### Retrying Operations

The retry package provides two main functions for retrying operations:

#### 1. `Do` - For operations that don't return a result

```
err := retry.Do(ctx, func() error {
    // Operation that may fail
    return someOperation()
}, config)
```

#### 2. `DoWithResult` - For operations that return a result

```
result, err := retry.DoWithResult(ctx, func() (ResultType, error) {
    // Operation that may fail
    return someOperation()
}, config)
```

### Error Classification

The retry package includes functions for classifying errors as retryable:

- `IsRetryableNetworkError`: Network-related errors (connection issues, etc.)
- `IsRetryableServerError`: Server errors (typically 5xx errors)
- `IsRetryableRateLimitError`: Rate limit errors (typically 429 errors)

You can add custom error classification functions to the retry configuration:

```
config := retry.DefaultConfig()
config.RetryableErrors = append(config.RetryableErrors, func(err error) bool {
    // Custom logic to determine if an error is retryable
    return strings.Contains(err.Error(), "specific error message")
})
```

## Examples

### Basic Retry

```
import (
    "context"
    "github.com/auriora/onemount/internal/retry"
)

func performOperation() error {
    ctx := context.Background()
    config := retry.DefaultConfig()
    
    return retry.Do(ctx, func() error {
        // Operation that may fail
        return someOperation()
    }, config)
}
```

### Retry with Custom Configuration

```
import (
    "context"
    "time"
    "github.com/auriora/onemount/internal/retry"
)

func performOperation() error {
    ctx := context.Background()
    config := retry.Config{
        MaxRetries:   5,
        InitialDelay: 500 * time.Millisecond,
        MaxDelay:     10 * time.Second,
        Multiplier:   1.5,
        Jitter:       0.1,
        RetryableErrors: []retry.RetryableError{
            retry.IsRetryableNetworkError,
            retry.IsRetryableServerError,
        },
    }
    
    return retry.Do(ctx, func() error {
        // Operation that may fail
        return someOperation()
    }, config)
}
```

### Retry with Result

```
import (
    "context"
    "github.com/auriora/onemount/internal/retry"
)

func fetchData() (Data, error) {
    ctx := context.Background()
    config := retry.DefaultConfig()
    
    return retry.DoWithResult(ctx, func() (Data, error) {
        // Operation that may fail
        return fetchDataFromAPI()
    }, config)
}
```

### Retry with Cleanup

```
import (
    "context"
    "github.com/auriora/onemount/internal/retry"
)

func processFile() error {
    ctx := context.Background()
    config := retry.DefaultConfig()
    
    return retry.Do(ctx, func() error {
        // Reset state before each attempt
        if err := resetState(); err != nil {
            return err
        }
        
        // Operation that may fail
        return processFileOperation()
    }, config)
}
```

## Best Practices

1. **Use the retry package for all critical operations**: File uploads/downloads, API calls, and other network operations.
2. **Clean up before each retry attempt**: Reset file positions, truncate files, etc.
3. **Use appropriate retry parameters**: Adjust the retry configuration based on the operation's importance and expected failure rate.
4. **Add context to error messages**: Include relevant context in error messages to aid debugging.
5. **Log retry attempts**: Log each retry attempt with relevant context for debugging.
6. **Consider the impact of retries**: Be mindful of the impact of retries on system resources and external services.
7. **Use timeouts**: Set appropriate timeouts for operations to avoid blocking indefinitely.
8. **Handle non-retryable errors appropriately**: Some errors should not be retried (e.g., validation errors, permission errors).

## Conclusion

By following these guidelines and using the retry utility package, you can implement robust error recovery mechanisms that handle transient failures gracefully. This will improve the reliability and user experience of the OneMount project.