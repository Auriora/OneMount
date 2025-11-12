# Manual API Rate Limiting Testing Guide

**Date**: 2025-11-11  
**Purpose**: Manual testing procedures for verifying API rate limiting handling in OneMount  
**Related Task**: 14.3 - Test API rate limiting

## Overview

This document provides manual testing procedures to verify that OneMount correctly handles Microsoft Graph API rate limiting (HTTP 429 responses), including exponential backoff, request queuing, and eventual success.

## Prerequisites

1. OneMount installed and configured
2. Valid OneDrive authentication
3. Test OneDrive account with test files
4. Ability to trigger rapid API requests
5. Optional: Network monitoring tools (Wireshark, tcpdump)

## Understanding Microsoft Graph Rate Limits

### Rate Limit Thresholds

Microsoft Graph API has the following rate limits:

- **Per-app per-user**: ~2,000 requests per 10 seconds
- **Per-app**: ~10,000 requests per 10 seconds
- **Per-tenant**: ~30,000 requests per 10 seconds

### Rate Limit Response

When rate limited, the API returns:
- **Status Code**: 429 (Too Many Requests)
- **Error Code**: "TooManyRequests" or "activityLimitReached"
- **Retry-After Header**: Suggested wait time in seconds

Example response:
```json
{
  "error": {
    "code": "TooManyRequests",
    "message": "The request has been throttled",
    "innerError": {
      "code": "activityLimitReached",
      "date": "2025-11-11T10:00:00",
      "request-id": "abc-123-def"
    }
  }
}
```

## OneMount Rate Limiting Strategy

### Multi-Layer Approach

1. **Immediate Retry with Exponential Backoff**:
   - Max retries: 5 (increased from default 3)
   - Initial delay: 1 second
   - Max delay: 60 seconds
   - Multiplier: 2.0
   - Jitter: 20%

2. **Request Queuing**:
   - After max retries, requests are queued
   - Queue processes requests with delays
   - Delay increases if rate limit persists (1.5x multiplier)
   - Max queue delay: 60 seconds

3. **Adaptive Delay**:
   - Delay resets to 1s on successful request
   - Delay increases to 60s max on continued rate limiting

## Test Scenarios

### Scenario 1: Trigger Rate Limiting with Rapid Requests

**Objective**: Verify that rapid API requests trigger rate limiting and are handled correctly

**Steps**:
1. Mount OneMount filesystem
2. Create a script to trigger many rapid requests:
   ```bash
   #!/bin/bash
   # Trigger rapid file listings
   for i in {1..100}; do
     ls -la /mnt/onedrive/ &
   done
   wait
   ```
3. Execute the script
4. Monitor logs for rate limit detection
5. Verify eventual success

**Expected Results**:
- ✅ Some requests return 429 status
- ✅ Rate limit errors detected as `ResourceBusyError`
- ✅ Retry logic activates with exponential backoff
- ✅ Delays increase: 1s → 2s → 4s → 8s → 16s → 32s (max 60s)
- ✅ After max retries, requests are queued
- ✅ Queued requests eventually succeed
- ✅ No crashes or data loss

**Verification**:
```bash
# Monitor logs for rate limiting
journalctl -u onemount -f | grep -i "rate\|429\|TooManyRequests"

# Check for retry attempts
journalctl -u onemount | grep "retrying after delay"

# Check for queued requests
journalctl -u onemount | grep "queued due to rate limiting"
```

---

### Scenario 2: Verify Retry-After Header Handling

**Objective**: Verify that OneMount respects the Retry-After header from API responses

**Steps**:
1. Mount OneMount filesystem
2. Trigger rate limiting (use rapid requests)
3. Monitor network traffic to see Retry-After headers:
   ```bash
   sudo tcpdump -i any -A 'tcp port 443' | grep -i "retry-after"
   ```
4. Check logs for Retry-After header detection
5. Verify delays match Retry-After values

**Expected Results**:
- ✅ Retry-After header is detected and logged
- ✅ Delays respect Retry-After values when present
- ✅ Exponential backoff used when Retry-After not present
- ✅ Log message: "Rate limit detected with Retry-After header: X"

**Verification**:
```bash
# Check for Retry-After header detection
journalctl -u onemount | grep "Retry-After"

# Verify delay values
journalctl -u onemount | grep "delay" | tail -20
```

---

### Scenario 3: Test Request Queue Behavior

**Objective**: Verify that requests are queued after max retries and processed correctly

**Steps**:
1. Mount OneMount filesystem
2. Trigger sustained rate limiting:
   ```bash
   #!/bin/bash
   # Sustained rapid requests
   while true; do
     for i in {1..50}; do
       ls -la /mnt/onedrive/Documents/ &
     done
     wait
     sleep 1
   done
   ```
3. Monitor queue length in logs
4. Observe queue processing
5. Stop the script after 2 minutes
6. Verify queue drains successfully

**Expected Results**:
- ✅ Requests are queued after max retries
- ✅ Queue length increases during sustained rate limiting
- ✅ Log message: "Request queued due to rate limiting"
- ✅ Queue processes requests with delays
- ✅ Queue length decreases as rate limit eases
- ✅ All queued requests eventually complete
- ✅ No request timeouts or failures

**Verification**:
```bash
# Monitor queue length
journalctl -u onemount -f | grep "queue_length"

# Check queue processing
journalctl -u onemount | grep "Executing queued request"

# Verify queue drains
journalctl -u onemount | grep "queue_length" | tail -20
```

---

### Scenario 4: Test Adaptive Delay Adjustment

**Objective**: Verify that delays adapt based on continued rate limiting

**Steps**:
1. Mount OneMount filesystem
2. Trigger rate limiting
3. Monitor delay values in logs
4. Observe delay increases when rate limit persists
5. Observe delay resets when requests succeed

**Expected Results**:
- ✅ Initial delay: 1 second
- ✅ Delay increases by 1.5x on continued rate limiting
- ✅ Max delay: 60 seconds
- ✅ Delay resets to 1s on successful request
- ✅ Log message: "Rate limit still in effect, increasing delay"

**Verification**:
```bash
# Track delay progression
journalctl -u onemount | grep "next_delay" | tail -30

# Verify delay increases
journalctl -u onemount | grep "increasing delay"

# Verify delay resets
journalctl -u onemount | grep "delay" | grep "1s"
```

---

### Scenario 5: Test Concurrent Operations During Rate Limiting

**Objective**: Verify that concurrent operations handle rate limiting correctly

**Steps**:
1. Mount OneMount filesystem
2. Start multiple concurrent operations:
   ```bash
   #!/bin/bash
   # Concurrent file operations
   ls -laR /mnt/onedrive/ &
   find /mnt/onedrive/ -type f &
   du -sh /mnt/onedrive/* &
   cat /mnt/onedrive/Documents/*.txt > /dev/null &
   wait
   ```
3. Monitor rate limiting behavior
4. Verify all operations eventually complete

**Expected Results**:
- ✅ Multiple operations trigger rate limiting
- ✅ Each operation retries independently
- ✅ Queue handles multiple concurrent requests
- ✅ No deadlocks or race conditions
- ✅ All operations eventually succeed
- ✅ No data corruption

**Verification**:
```bash
# Monitor concurrent operations
journalctl -u onemount -f | grep -E "operation|retry|queue"

# Check for errors
journalctl -u onemount | grep -i "error\|fail" | tail -50
```

---

### Scenario 6: Test Rate Limit Recovery

**Objective**: Verify system recovers gracefully after rate limiting subsides

**Steps**:
1. Mount OneMount filesystem
2. Trigger heavy rate limiting
3. Wait for rate limiting to subside (stop triggering requests)
4. Perform normal file operations
5. Verify normal operation resumes

**Expected Results**:
- ✅ Queue drains after rate limiting subsides
- ✅ Delays return to normal (1s)
- ✅ Normal operations succeed immediately
- ✅ No lingering effects from rate limiting
- ✅ Performance returns to normal

**Verification**:
```bash
# Check queue status
journalctl -u onemount | grep "queue_length" | tail -10

# Verify normal operation
time ls -la /mnt/onedrive/
# Should complete quickly (<1s)
```

---

### Scenario 7: Test Error Metrics for Rate Limiting

**Objective**: Verify that rate limit errors are tracked in error metrics

**Steps**:
1. Mount OneMount filesystem
2. Trigger rate limiting
3. Wait 5 minutes for metrics logging
4. Check error metrics summary

**Expected Results**:
- ✅ Rate limit errors counted in `resource_busy_errors`
- ✅ Rate limit errors counted in `rate_limit_errors`
- ✅ Error metrics logged every 5 minutes
- ✅ Error rates calculated correctly

**Verification**:
```bash
# Wait for metrics summary (every 5 minutes)
journalctl -u onemount | grep "Error metrics summary" | tail -5

# Check rate limit counts
journalctl -u onemount | grep "rate_limit_errors"

# Check resource busy counts
journalctl -u onemount | grep "resource_busy_errors"
```

---

## Automated Test Verification

### Existing Unit Tests

The following unit tests verify rate limiting:

1. **TestUT_GR_RATE_01_01_RateLimitDetection_429Response_DetectsRateLimit**
   - Location: `internal/graph/rate_limit_test.go`
   - Tests: 429 response detection
   - Run: `go test -v -run TestUT_GR_RATE_01_01 ./internal/graph/`

2. **TestUT_GR_RATE_01_02_RateLimitWithRetryAfter_RetryAfterHeader_RespectsDelay**
   - Location: `internal/graph/rate_limit_test.go`
   - Tests: Retry-After header handling
   - Run: `go test -v -run TestUT_GR_RATE_01_02 ./internal/graph/`

3. **TestUT_GR_RATE_02_01_RetryLogic_TransientError_RetriesSuccessfully**
   - Location: `internal/graph/rate_limit_test.go`
   - Tests: Retry logic for transient errors
   - Run: `go test -v -run TestUT_GR_RATE_02_01 ./internal/graph/`

4. **TestUT_GR_RATE_03_01_RequestQueue_RateLimitedRequests_QueuesForLater**
   - Location: `internal/graph/rate_limit_test.go`
   - Tests: Request queuing
   - Run: `go test -v -run TestUT_GR_RATE_03_01 ./internal/graph/`

### Running Tests

```bash
# Run all rate limit tests
go test -v -run 'TestUT_GR_RATE' ./internal/graph/

# Run with race detector
go test -v -race -run 'TestUT_GR_RATE' ./internal/graph/

# Run in Docker
docker compose -f docker/compose/docker-compose.test.yml run --rm unit-tests
```

---

## Expected Log Patterns

### Rate Limit Detection

```json
{
  "level": "info",
  "resource": "/me/drive/items/ABC123",
  "method": "GET",
  "msg": "Rate limit detected with Retry-After header: 5"
}
```

### Retry Attempt

```json
{
  "level": "info",
  "error": "ResourceBusyError: Too many requests",
  "attempt": 3,
  "maxRetries": 5,
  "delay": "4s",
  "msg": "Operation failed, retrying after delay"
}
```

### Request Queued

```json
{
  "level": "info",
  "resource": "/me/drive/items/ABC123",
  "method": "GET",
  "queue_length": 5,
  "msg": "Request queued due to rate limiting"
}
```

### Queue Processing

```json
{
  "level": "info",
  "resource": "/me/drive/items/ABC123",
  "method": "GET",
  "msg": "Executing queued request after rate limit delay"
}
```

### Adaptive Delay

```json
{
  "level": "warn",
  "resource": "/me/drive/items/ABC123",
  "method": "GET",
  "next_delay": "15s",
  "msg": "Rate limit still in effect, increasing delay"
}
```

### Error Metrics

```json
{
  "level": "info",
  "total_errors": 50,
  "network_errors": 5,
  "auth_errors": 0,
  "not_found_errors": 0,
  "validation_errors": 0,
  "operation_errors": 0,
  "resource_busy_errors": 45,
  "rate_limit_errors": 45,
  "msg": "Error metrics summary"
}
```

---

## Rate Limiting Configuration

### Current Configuration

```go
// In internal/graph/graph.go
retryConfig := retry.Config{
    MaxRetries:   5,                // Increased from default 3
    InitialDelay: 1 * time.Second,
    MaxDelay:     60 * time.Second, // Increased from default 30s
    Multiplier:   2.0,
    Jitter:       0.2,
    RetryableErrors: []RetryableError{
        retry.IsRetryableNetworkError,
        retry.IsRetryableServerError,
        retry.IsRetryableRateLimitError,
    },
}
```

### Queue Configuration

```go
// In internal/graph/request_queue.go
// Initial delay: 1 second
// Delay multiplier: 1.5x on continued rate limiting
// Max delay: 60 seconds
// Queue timeout: 5 minutes per request
```

---

## Success Criteria

### Rate Limit Detection

- ✅ 429 responses detected as `ResourceBusyError`
- ✅ Retry-After header parsed and logged
- ✅ Rate limit errors tracked in metrics

### Retry Behavior

- ✅ Exponential backoff: 1s → 2s → 4s → 8s → 16s → 32s
- ✅ Max 5 retry attempts before queuing
- ✅ Jitter applied to prevent thundering herd
- ✅ Context cancellation respected

### Request Queuing

- ✅ Requests queued after max retries
- ✅ Queue processes requests with delays
- ✅ Adaptive delay adjustment (1.5x multiplier)
- ✅ Queue drains successfully
- ✅ No request timeouts or losses

### System Behavior

- ✅ No crashes during rate limiting
- ✅ Concurrent operations handled correctly
- ✅ Normal operation resumes after rate limiting
- ✅ Error metrics accurate
- ✅ All operations eventually succeed

---

## Troubleshooting

### Issue: Rate limiting not triggered

**Possible Causes**:
- Not enough concurrent requests
- API rate limits not reached

**Solution**: Increase request volume:
```bash
# More aggressive test
for i in {1..500}; do
  ls -la /mnt/onedrive/ &
done
```

### Issue: Requests timing out

**Possible Causes**:
- Queue timeout (5 minutes) exceeded
- Context cancellation

**Solution**: Check logs for timeout errors:
```bash
journalctl -u onemount | grep -i "timeout\|cancelled"
```

### Issue: Queue not draining

**Possible Causes**:
- Continued rate limiting
- Queue processing stopped

**Solution**: Check queue processing:
```bash
# Verify queue is running
journalctl -u onemount | grep "Initialized global request queue"

# Check for queue processing
journalctl -u onemount | grep "Executing queued request"
```

---

## Performance Impact

### Expected Behavior

- **Normal Operation**: <100ms per request
- **During Rate Limiting**: 1-60s delays per request
- **Queue Processing**: 1-60s between queued requests
- **Recovery Time**: <1 minute after rate limiting subsides

### Monitoring

```bash
# Monitor request latency
journalctl -u onemount -f | grep -E "duration|delay"

# Check queue length over time
watch -n 5 'journalctl -u onemount | grep "queue_length" | tail -1'
```

---

## References

- **Rate Limiting Design**: `docs/2-architecture/error-handling.md`
- **Retry Logic**: `internal/retry/retry.go`
- **Request Queue**: `internal/graph/request_queue.go`
- **Error Types**: `internal/errors/error_types.go`
- **Graph API Client**: `internal/graph/graph.go`
- **Microsoft Graph Rate Limits**: https://learn.microsoft.com/en-us/graph/throttling

---

## Test Results Template

```markdown
## API Rate Limiting Testing Results

**Date**: YYYY-MM-DD  
**Tester**: [Name]  
**OneMount Version**: [Version]

### Scenario 1: Rapid Requests
- Status: ✅ PASS / ❌ FAIL
- Rate limits triggered: Yes / No
- Max retries reached: Yes / No
- Requests queued: Yes / No
- All requests succeeded: Yes / No
- Notes: [Observations]

### Scenario 2: Retry-After Header
- Status: ✅ PASS / ❌ FAIL
- Header detected: Yes / No
- Header respected: Yes / No
- Notes: [Observations]

### Scenario 3: Request Queue
- Status: ✅ PASS / ❌ FAIL
- Max queue length: [Number]
- Queue drained: Yes / No
- Notes: [Observations]

### Scenario 4: Adaptive Delay
- Status: ✅ PASS / ❌ FAIL
- Delay increased: Yes / No
- Delay reset: Yes / No
- Max delay reached: Yes / No
- Notes: [Observations]

### Scenario 5: Concurrent Operations
- Status: ✅ PASS / ❌ FAIL
- All operations completed: Yes / No
- No deadlocks: Yes / No
- Notes: [Observations]

### Scenario 6: Recovery
- Status: ✅ PASS / ❌ FAIL
- Normal operation resumed: Yes / No
- Recovery time: [Duration]
- Notes: [Observations]

### Scenario 7: Error Metrics
- Status: ✅ PASS / ❌ FAIL
- Metrics accurate: Yes / No
- Rate limit count: [Number]
- Notes: [Observations]

### Issues Found
1. [Issue description]
2. [Issue description]

### Recommendations
1. [Recommendation]
2. [Recommendation]
```

---

**Last Updated**: 2025-11-11  
**Status**: Ready for manual testing  
**Next Steps**: Execute manual tests and document results
