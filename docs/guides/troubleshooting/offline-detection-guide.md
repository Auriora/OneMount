# Offline Detection Troubleshooting Guide

This guide explains how OneMount detects offline conditions and how to troubleshoot offline detection issues.

## How Offline Detection Works

OneMount uses a multi-layered approach to detect offline conditions:

### 1. Operational Offline Mode

The system can be manually set to offline mode using `SetOperationalOffline(true)`. This overrides all other detection mechanisms.

**Check operational offline state**:
```go
isOffline := graph.GetOperationalOffline()
```

### 2. Error Pattern Detection

When network operations fail, OneMount analyzes the error message to determine if it indicates a network connectivity issue.

#### Network Error Patterns (Trigger Offline Detection)

The following error patterns indicate network connectivity issues and trigger offline detection:

1. **"no such host"** - DNS resolution failure
2. **"network is unreachable"** - Network interface down or routing issue
3. **"connection refused"** - Target service not accepting connections
4. **"connection timed out"** - Connection attempt timed out
5. **"dial tcp"** - TCP connection failure
6. **"context deadline exceeded"** - Operation timeout
7. **"no route to host"** - Routing failure
8. **"network is down"** - Network interface disabled
9. **"temporary failure in name resolution"** - DNS temporary failure
10. **"operation timed out"** - General operation timeout

#### Non-Network Error Patterns (Do NOT Trigger Offline Detection)

The following error patterns indicate the system is online but experiencing other issues:

1. **HTTP Response Codes** (e.g., "HTTP 404 - Not Found", "HTTP 500 - Internal Server Error")
   - Any HTTP response indicates network connectivity
   
2. **Authentication/Authorization Errors**:
   - "401" / "unauthorized"
   - "403" / "forbidden"
   - "invalid token"
   - "permission denied"
   - "access denied"
   - "authentication failed"
   - "authorization failed"

3. **Unknown Errors**:
   - Errors that don't match any known pattern default to online
   - This prevents false positives

### 3. Network Error Type Detection

Errors explicitly marked as network errors (using `errors.IsNetworkError()`) trigger offline detection.

## Troubleshooting Offline Detection Issues

### Issue: System Not Detecting Offline State

**Symptoms**:
- Network is disconnected but system still tries to make requests
- Operations hang or timeout instead of failing fast

**Possible Causes**:

1. **Error pattern not recognized**
   - Check error logs for the exact error message
   - Verify if the error message contains one of the recognized patterns
   - Error pattern matching is case-insensitive

2. **HTTP error being returned**
   - If the error contains "HTTP [code] -", it's treated as online
   - This is correct behavior - HTTP responses indicate connectivity

3. **Authentication error being returned**
   - Authentication errors indicate the system is online
   - Check if the error contains authentication-related keywords

**Solutions**:

1. **Check error logs**:
   ```bash
   grep "Offline condition detected" ~/.onemount-tests/logs/*.log
   grep "Online condition detected" ~/.onemount-tests/logs/*.log
   ```

2. **Verify error pattern**:
   - Look for debug logs showing which pattern was matched
   - If no pattern matches, the error defaults to online

3. **Add new error pattern** (if needed):
   - If you encounter a legitimate network error that's not detected
   - Add the pattern to the `offlinePatterns` list in `internal/graph/graph.go`
   - Submit a bug report with the error message

### Issue: False Positive Offline Detection

**Symptoms**:
- System detects offline when network is actually working
- Authentication errors trigger offline mode

**Possible Causes**:

1. **Operational offline mode is set**
   - Check if `GetOperationalOffline()` returns true
   - This overrides all other detection

2. **Error message contains network-like keywords**
   - Some error messages may contain words like "connection" or "network"
   - But they're not actual network errors

**Solutions**:

1. **Check operational offline state**:
   ```go
   if graph.GetOperationalOffline() {
       // System is manually set to offline
       graph.SetOperationalOffline(false)
   }
   ```

2. **Review error logs**:
   ```bash
   grep "Offline condition detected via error pattern" ~/.onemount-tests/logs/*.log
   ```
   - Check which pattern triggered the detection
   - Verify if it's a legitimate network error

3. **Check for authentication errors**:
   - Authentication errors should NOT trigger offline detection
   - If they do, this is a bug - please report it

### Issue: Delayed Offline Detection

**Symptoms**:
- Network disconnects but system takes time to detect it
- Operations continue to fail for a while before offline mode activates

**Explanation**:
- OneMount uses **passive offline detection**
- Offline state is detected when operations fail, not proactively
- This is by design for simplicity and reliability

**Expected Behavior**:
- First operation after network disconnect will fail
- Subsequent operations will use offline mode
- Detection latency is typically < 1 second

**Workaround**:
- Use manual offline mode if you need immediate detection:
  ```go
  graph.SetOperationalOffline(true)
  ```

## Testing Offline Detection

### Unit Tests

Run the network error pattern tests:
```bash
docker compose -f docker/compose/docker-compose.test.yml run --rm test-runner \
  go test -v -run TestUT_GR_26 ./internal/graph
```

### Property-Based Tests

Run the Property 24 test (100 iterations):
```bash
docker compose -f docker/compose/docker-compose.test.yml run --rm test-runner \
  go test -v -run TestProperty24_OfflineDetection ./internal/fs -timeout 30s
```

### Integration Tests

Run the offline mode integration tests:
```bash
docker compose -f docker/compose/docker-compose.test.yml run --rm test-runner \
  go test -v -run TestIT_OF ./internal/fs
```

## Debugging Offline Detection

### Enable Debug Logging

Set the log level to debug to see offline detection decisions:
```bash
export ONEMOUNT_LOG_LEVEL=debug
```

### Check Debug Logs

Look for these log messages:

1. **Offline detected via pattern**:
   ```
   {"level":"debug","pattern":"no such host","error":"...","message":"Offline condition detected via error pattern"}
   ```

2. **Online detected via auth error**:
   ```
   {"level":"debug","pattern":"permission denied","error":"...","message":"Online condition detected via authentication/authorization error pattern"}
   ```

3. **Online assumed for unknown error**:
   ```
   {"level":"debug","error":"...","message":"Online condition assumed for unknown error type (non-conservative default)"}
   ```

4. **Offline detected via network error type**:
   ```
   {"level":"debug","error":"...","message":"Offline condition detected via network error type"}
   ```

### Verify Error Classification

Use the `IsOffline()` function directly in tests:
```go
import "github.com/auriora/onemount/internal/graph"

err := errors.New("your error message here")
isOffline := graph.IsOffline(err)
fmt.Printf("Error classified as offline: %v\n", isOffline)
```

## Common Scenarios

### Scenario 1: Network Cable Unplugged

**Expected Behavior**:
- First operation fails with "network is unreachable" or "no route to host"
- `IsOffline()` returns `true`
- Subsequent operations use offline mode

### Scenario 2: WiFi Disconnected

**Expected Behavior**:
- First operation fails with "no such host" or "network is unreachable"
- `IsOffline()` returns `true`
- Subsequent operations use offline mode

### Scenario 3: VPN Disconnected

**Expected Behavior**:
- First operation fails with "connection timed out" or "no route to host"
- `IsOffline()` returns `true`
- Subsequent operations use offline mode

### Scenario 4: Authentication Token Expired

**Expected Behavior**:
- Operation fails with "HTTP 401 - Unauthorized" or "invalid token"
- `IsOffline()` returns `false` (online)
- System attempts to refresh token
- If refresh succeeds, operation retries
- If refresh fails, user is prompted to re-authenticate

### Scenario 5: Permission Denied

**Expected Behavior**:
- Operation fails with "HTTP 403 - Forbidden" or "permission denied"
- `IsOffline()` returns `false` (online)
- Error is returned to user
- System remains in online mode

## Implementation Details

### IsOffline() Function Flow

```
IsOffline(err error) bool
    ↓
1. Check operational offline state
    ↓ (if false)
2. Check if error is nil
    ↓ (if not nil)
3. Check for HTTP response pattern
    ↓ (if no match)
4. Check for auth/authz error patterns
    ↓ (if no match)
5. Check for network error patterns
    ↓ (if no match)
6. Check if error is NetworkError type
    ↓ (if no match)
7. Default to online (false)
```

### Pattern Matching Rules

1. **Case-insensitive**: All pattern matching is case-insensitive
2. **Substring match**: Patterns can appear anywhere in the error message
3. **First match wins**: First matching pattern determines the result
4. **Logging**: All detection decisions are logged at debug level

## References

- **Implementation**: `internal/graph/graph.go` (IsOffline function)
- **Property Test**: `internal/fs/offline_property_test.go`
- **Integration Tests**: `internal/graph/network_error_patterns_test.go`
- **Requirements**: Requirements 6.1, 19.1-19.11
- **Design**: `docs/2-architecture/software-architecture-specification.md`
