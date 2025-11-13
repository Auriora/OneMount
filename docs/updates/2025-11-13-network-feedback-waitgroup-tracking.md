# Network Feedback Wait Group Tracking Fix

**Date**: 2025-11-13  
**Issue**: #PERF-002  
**Component**: Network Feedback / Goroutine Management  
**Priority**: Medium  
**Status**: ✅ RESOLVED

## Problem

The `NetworkFeedbackManager` spawned callback goroutines without tracking them with a wait group. This meant that during shutdown, there was no way to ensure all callbacks completed gracefully before the application terminated. This could lead to:

1. Incomplete callback execution during shutdown
2. Potential resource leaks
3. Lost notifications
4. Difficulty in testing graceful shutdown behavior

## Root Cause

The `NetworkFeedbackManager` methods (`NotifyConnected`, `NotifyDisconnected`, `NotifyStatusUpdate`) spawned goroutines for each handler callback but did not track them:

```go
for _, handler := range handlers {
    go func(h NetworkFeedbackHandler) {
        defer func() {
            if r := recover(); r != nil {
                logging.Error().Interface("panic", r).Msg("Network feedback handler panicked")
            }
        }()
        h.OnNetworkConnected()
    }(handler)
}
```

There was no mechanism to:
- Track how many callback goroutines were active
- Wait for callbacks to complete during shutdown
- Enforce a timeout for callback completion

## Solution

Added wait group tracking and graceful shutdown support to `NetworkFeedbackManager`:

### 1. Added WaitGroup Field

```go
type NetworkFeedbackManager struct {
    handlers []NetworkFeedbackHandler
    mutex    sync.RWMutex
    wg       sync.WaitGroup // Track callback goroutines for graceful shutdown
}
```

### 2. Track Callback Goroutines

Updated all notification methods to track goroutines:

```go
for _, handler := range handlers {
    m.wg.Add(1)
    go func(h NetworkFeedbackHandler) {
        defer m.wg.Done()
        defer func() {
            if r := recover(); r != nil {
                logging.Error().Interface("panic", r).Msg("Network feedback handler panicked")
            }
        }()
        h.OnNetworkConnected()
    }(handler)
}
```

Key changes:
- `m.wg.Add(1)` before spawning each goroutine
- `defer m.wg.Done()` as the first defer in the goroutine
- Ensures `Done()` is called even if the handler panics

### 3. Added Shutdown Method

```go
func (m *NetworkFeedbackManager) Shutdown(timeout time.Duration) bool {
    done := make(chan struct{})
    go func() {
        m.wg.Wait()
        close(done)
    }()

    select {
    case <-done:
        logging.Debug().Msg("All network feedback callbacks completed")
        return true
    case <-time.After(timeout):
        logging.Warn().
            Dur("timeout", timeout).
            Msg("Network feedback callbacks did not complete within timeout")
        return false
    }
}
```

Features:
- Waits for all callback goroutines to complete
- Enforces a configurable timeout
- Returns `true` if all callbacks completed, `false` if timeout occurred
- Logs appropriate messages for debugging

## Testing

Created comprehensive tests in `internal/graph/network_feedback_test.go`:

### Test Coverage

1. **TestNetworkFeedbackManager_WaitGroupTracking**
   - Verifies callbacks are tracked correctly
   - Ensures shutdown completes when all callbacks are done

2. **TestNetworkFeedbackManager_ShutdownTimeout**
   - Tests timeout behavior with blocking callbacks
   - Verifies timeout is respected

3. **TestNetworkFeedbackManager_ShutdownWithMultipleCallbacks**
   - Tests shutdown with multiple handlers
   - Verifies all callbacks complete before shutdown returns

4. **TestNetworkFeedbackManager_PanicRecovery**
   - Ensures panicking handlers don't prevent `Done()` call
   - Verifies other handlers still execute

5. **TestNetworkFeedbackManager_ConcurrentNotifications**
   - Tests concurrent notification calls
   - Verifies wait group handles concurrent operations correctly

### Test Results

```
=== RUN   TestNetworkFeedbackManager_WaitGroupTracking
--- PASS: TestNetworkFeedbackManager_WaitGroupTracking (0.00s)
=== RUN   TestNetworkFeedbackManager_ShutdownTimeout
--- PASS: TestNetworkFeedbackManager_ShutdownTimeout (0.60s)
=== RUN   TestNetworkFeedbackManager_ShutdownWithMultipleCallbacks
--- PASS: TestNetworkFeedbackManager_ShutdownWithMultipleCallbacks (0.20s)
=== RUN   TestNetworkFeedbackManager_PanicRecovery
--- PASS: TestNetworkFeedbackManager_PanicRecovery (0.00s)
=== RUN   TestNetworkFeedbackManager_ConcurrentNotifications
--- PASS: TestNetworkFeedbackManager_ConcurrentNotifications (0.05s)
PASS
ok      github.com/auriora/onemount/internal/graph      0.907s
```

All tests pass successfully.

## Usage

To use the shutdown functionality:

```go
// During application shutdown
feedbackManager := GetGlobalFeedbackManager()
timeout := 5 * time.Second

if !feedbackManager.Shutdown(timeout) {
    logging.Warn().Msg("Some network feedback callbacks did not complete in time")
}
```

## Benefits

1. **Graceful Shutdown**: All callback goroutines complete before shutdown
2. **Timeout Protection**: Prevents indefinite waiting for misbehaving callbacks
3. **Resource Cleanup**: Ensures proper cleanup of goroutines
4. **Testability**: Can verify shutdown behavior in tests
5. **Observability**: Logs when callbacks don't complete in time

## Requirements Satisfied

- **Requirement 10.5**: "WHEN goroutines are spawned, THE OneMount System SHALL track them with wait groups for clean shutdown"

## Related Issues

- Issue #PERF-002: Network Callbacks Lack Wait Group Tracking

## Files Modified

- `internal/graph/network_feedback.go`: Added wait group tracking and Shutdown method
- `internal/graph/network_feedback_test.go`: Added comprehensive tests

## Recommendations

1. **Integration with Filesystem Shutdown**: The filesystem shutdown process should call `feedbackManager.Shutdown()` with an appropriate timeout (e.g., 5 seconds)

2. **Timeout Configuration**: Consider making the shutdown timeout configurable via command-line flag or configuration file

3. **Monitoring**: Consider adding metrics for:
   - Number of active callbacks
   - Shutdown timeout occurrences
   - Average callback completion time

4. **Handler Guidelines**: Document best practices for handler implementations:
   - Keep callbacks short and non-blocking
   - Use timeouts for any network operations
   - Handle context cancellation appropriately

## Verification

- ✅ Wait group tracks all callback goroutines
- ✅ Shutdown waits for callbacks with timeout
- ✅ Panic recovery doesn't prevent Done() call
- ✅ Concurrent notifications handled correctly
- ✅ All tests pass
- ✅ No diagnostics or linting issues
