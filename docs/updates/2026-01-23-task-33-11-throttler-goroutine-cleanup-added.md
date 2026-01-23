# Task 33.11 Added: Fix Goroutine Cleanup in Throttler

**Date**: 2026-01-23  
**Type**: Task Addition  
**Priority**: HIGH  
**Phase**: 20.1 - Fix Resource Management Property Test Failures

## Summary

Added task 33.11 to fix goroutine leak in `internal/util/throttler.go` that causes TestProperty59_AdaptiveNetworkThrottling to timeout after 600 seconds.

## Issue Description

**Test**: TestProperty59_AdaptiveNetworkThrottling  
**Symptom**: Test times out after 600 seconds with goroutine leak  
**Root Cause**: Goroutines in `BandwidthThrottler.Wait()` not properly cleaned up when test completes

## Technical Analysis

The `BandwidthThrottler.Wait()` method uses `time.After()` in a select statement:

```go
select {
case <-time.After(sleepDuration):
    // Sleep completed normally
case <-ctx.Done():
    bt.mutex.Lock()
    return ctx.Err()
}
```

**Problem**: `time.After()` creates a timer that cannot be stopped. If the context is cancelled before the timer expires, the timer goroutine continues running until it fires, causing a goroutine leak.

## Solution Approach

Replace `time.After()` with `time.NewTimer()` to allow proper cleanup:

```go
timer := time.NewTimer(sleepDuration)
defer timer.Stop()

select {
case <-timer.C:
    // Sleep completed normally
case <-ctx.Done():
    bt.mutex.Lock()
    return ctx.Err()
}
```

This ensures the timer is stopped and resources are released when the function returns, regardless of which case is selected.

## Task Details

**Task ID**: 33.11  
**Title**: Fix Property 59: Goroutine Cleanup in Throttler  
**Location**: `.kiro/specs/system-verification-and-fix/tasks.md` (Phase 20.1)

### Action Items

1. Review `internal/util/throttler.go` for goroutine cleanup issues
2. Replace `time.After()` with `time.NewTimer()` to allow proper cleanup
3. Add `defer timer.Stop()` to ensure timer resources are released
4. Add context cancellation checks before sleeping
5. Ensure all goroutines are tracked with wait groups if needed
6. Add timeout protection to throttler operations
7. Test with race detector: `go test -race -run TestProperty59`
8. Re-run Property 59 test to confirm fix

### Test Command

```bash
docker compose -f docker/compose/docker-compose.test.yml run --rm test-runner \
  go test -v -run "^TestProperty59" ./internal/fs
```

## Requirements Validated

- **Requirement 24.7**: Adaptive Network Throttling
- **Requirement 10.5**: Goroutine Management and Cleanup

## Estimated Effort

2-3 hours

## Priority Justification

**HIGH** - This issue is blocking the completion of the property-based test suite. While 66/67 tests pass, having a single test timeout prevents full validation of the system's correctness properties.

## Related Documents

- **Audit Report**: `test-artifacts/logs/task-46-1-8-property-test-isolation-audit.md`
- **Test Results**: Property 59 timeout documented in audit report
- **Code Location**: `internal/util/throttler.go` lines 60-75

## Next Steps

1. Implement the fix in `internal/util/throttler.go`
2. Run tests with race detector to verify no race conditions
3. Run Property 59 test to confirm timeout is resolved
4. Update audit report with fix results
5. Mark task 33.11 as complete

---

**Added By**: Kiro AI Agent  
**Date**: 2026-01-23  
**Related Task**: 46.1.8 (Property-Based Test Isolation Verification)
