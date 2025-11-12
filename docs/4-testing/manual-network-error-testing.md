# Manual Network Error Testing Guide

**Date**: 2025-11-11  
**Purpose**: Manual testing procedures for verifying network error handling in OneMount  
**Related Task**: 14.2 - Test network error handling

## Overview

This document provides manual testing procedures to verify that OneMount correctly handles various network error scenarios, including timeouts, connection failures, and transient errors.

## Prerequisites

1. OneMount installed and configured
2. Valid OneDrive authentication
3. Network simulation tools (optional):
   - `tc` (traffic control) for Linux
   - `iptables` for blocking connections
   - `toxiproxy` for advanced network simulation
4. Test OneDrive account with test files

## Test Scenarios

### Scenario 1: Network Timeout During Download

**Objective**: Verify that download operations handle network timeouts gracefully with retry

**Steps**:
1. Mount OneMount filesystem
2. Create a large file (>10MB) in OneDrive
3. Start downloading the file
4. During download, simulate network timeout:
   ```bash
   # Block OneDrive API temporarily
   sudo iptables -A OUTPUT -d graph.microsoft.com -j DROP
   ```
5. Wait 30 seconds
6. Restore network:
   ```bash
   sudo iptables -D OUTPUT -d graph.microsoft.com -j DROP
   ```
7. Observe download behavior

**Expected Results**:
- ✅ Download fails initially with network error
- ✅ Error is logged with context (file ID, path, operation)
- ✅ Download is retried automatically
- ✅ Retry uses exponential backoff (1s, 2s, 4s delays)
- ✅ Download eventually succeeds after network restoration
- ✅ File content is correct and complete

**Verification**:
```bash
# Check logs for retry attempts
journalctl -u onemount -f | grep -i "retry\|network"

# Verify file integrity
md5sum /path/to/downloaded/file
```

---

### Scenario 2: Connection Refused Error

**Objective**: Verify handling of connection refused errors

**Steps**:
1. Mount OneMount filesystem
2. Block all connections to Microsoft Graph API:
   ```bash
   sudo iptables -A OUTPUT -d graph.microsoft.com -j REJECT --reject-with icmp-port-unreachable
   ```
3. Attempt to read a file
4. Observe error handling
5. Restore network:
   ```bash
   sudo iptables -D OUTPUT -d graph.microsoft.com -j REJECT --reject-with icmp-port-unreachable
   ```

**Expected Results**:
- ✅ Operation fails with network error
- ✅ Error message is clear: "network request failed"
- ✅ Technical details logged but not shown to user
- ✅ Retry attempts are made (up to 3 times)
- ✅ Final error is returned after max retries

**Verification**:
```bash
# Check error logs
tail -f ~/.local/share/onemount/logs/onemount.log | grep -i "network\|connection"
```

---

### Scenario 3: DNS Resolution Failure

**Objective**: Verify handling of DNS resolution failures

**Steps**:
1. Mount OneMount filesystem
2. Temporarily break DNS resolution:
   ```bash
   # Backup DNS config
   sudo cp /etc/resolv.conf /etc/resolv.conf.backup
   
   # Point to invalid DNS
   echo "nameserver 192.0.2.1" | sudo tee /etc/resolv.conf
   ```
3. Attempt to access files
4. Observe error handling
5. Restore DNS:
   ```bash
   sudo mv /etc/resolv.conf.backup /etc/resolv.conf
   ```

**Expected Results**:
- ✅ Network error detected
- ✅ Retry logic activated
- ✅ Clear error message after retries exhausted
- ✅ System recovers when DNS restored

---

### Scenario 4: Intermittent Network Failures

**Objective**: Verify handling of intermittent network issues

**Steps**:
1. Mount OneMount filesystem
2. Use `tc` to add packet loss:
   ```bash
   # Add 50% packet loss
   sudo tc qdisc add dev eth0 root netem loss 50%
   ```
3. Perform multiple file operations (read, write, list)
4. Observe behavior
5. Remove packet loss:
   ```bash
   sudo tc qdisc del dev eth0 root
   ```

**Expected Results**:
- ✅ Some operations fail initially
- ✅ Failed operations are retried
- ✅ Most operations eventually succeed
- ✅ Error rate is logged in metrics
- ✅ No crashes or hangs

**Verification**:
```bash
# Monitor error metrics
# Check logs every 5 minutes for error rate summary
tail -f ~/.local/share/onemount/logs/onemount.log | grep "Error metrics summary"
```

---

### Scenario 5: Network Latency

**Objective**: Verify handling of high network latency

**Steps**:
1. Mount OneMount filesystem
2. Add network delay:
   ```bash
   # Add 2 second delay
   sudo tc qdisc add dev eth0 root netem delay 2000ms
   ```
3. Perform file operations
4. Observe behavior and timeouts
5. Remove delay:
   ```bash
   sudo tc qdisc del dev eth0 root
   ```

**Expected Results**:
- ✅ Operations complete successfully (slower)
- ✅ No premature timeouts
- ✅ Timeout errors only after reasonable wait (30s+)
- ✅ Retry logic handles timeout errors

---

### Scenario 6: Network Restoration After Failure

**Objective**: Verify system recovers gracefully after network restoration

**Steps**:
1. Mount OneMount filesystem
2. Perform some file operations (establish baseline)
3. Disconnect network completely:
   ```bash
   sudo ip link set eth0 down
   ```
4. Attempt file operations (should fail)
5. Wait 1 minute
6. Restore network:
   ```bash
   sudo ip link set eth0 up
   ```
7. Retry file operations

**Expected Results**:
- ✅ Operations fail during network outage
- ✅ Clear error messages indicating network issue
- ✅ Operations automatically succeed after network restoration
- ✅ No manual intervention required
- ✅ Delta sync resumes automatically

---

## Automated Test Verification

### Existing Integration Tests

The following integration tests verify network error handling:

1. **TestIT_FS_08_04_DownloadManager_DownloadFailureAndRetry**
   - Location: `internal/fs/download_manager_integration_test.go`
   - Tests: Download retry with network failures
   - Run: `go test -v -run TestIT_FS_08_04 ./internal/fs/`

2. **TestIT_FS_09_04_UploadFailureAndRetry**
   - Location: `internal/fs/upload_retry_integration_test.go`
   - Tests: Upload retry with network failures
   - Run: `go test -v -run TestIT_FS_09_04 ./internal/fs/`

3. **TestIT_SM_03_01_SyncManager_NetworkRecovery**
   - Location: `internal/fs/sync_manager_test.go`
   - Tests: Sync manager network recovery
   - Run: `go test -v -run TestIT_SM_03_01 ./internal/fs/`

### Running Tests in Docker

```bash
# Run all integration tests
docker compose -f docker/compose/docker-compose.test.yml run --rm integration-tests

# Run specific network error tests
docker compose -f docker/compose/docker-compose.test.yml run --rm shell
# Then inside container:
go test -v -run 'TestIT.*network|TestIT.*retry' ./internal/fs/
```

---

## Error Logging Verification

### Expected Log Patterns

**Network Error**:
```json
{
  "level": "error",
  "error": "NetworkError: network request failed: connection refused",
  "operation": "download",
  "file_id": "ABC123",
  "path": "/Documents/test.txt",
  "msg": "Network request failed"
}
```

**Retry Attempt**:
```json
{
  "level": "info",
  "error": "network error",
  "attempt": 2,
  "maxRetries": 3,
  "delay": "2s",
  "msg": "Operation failed, retrying after delay"
}
```

**Error Metrics** (every 5 minutes):
```json
{
  "level": "info",
  "total_errors": 15,
  "network_errors": 10,
  "auth_errors": 0,
  "not_found_errors": 2,
  "validation_errors": 0,
  "operation_errors": 3,
  "resource_busy_errors": 0,
  "rate_limit_errors": 0,
  "msg": "Error metrics summary"
}
```

### Log Verification Commands

```bash
# Monitor real-time logs
journalctl -u onemount -f

# Filter for network errors
journalctl -u onemount | grep -i "network"

# Filter for retry attempts
journalctl -u onemount | grep -i "retry"

# Check error metrics
journalctl -u onemount | grep "Error metrics summary"

# Count error types
journalctl -u onemount | grep "network_errors" | tail -1
```

---

## Success Criteria

### Network Error Handling

- ✅ All network errors are caught and typed correctly
- ✅ Errors are logged with full context (operation, file, path)
- ✅ Retry logic activates for transient errors
- ✅ Exponential backoff is used (1s, 2s, 4s, max 30s)
- ✅ Max 3 retry attempts before giving up
- ✅ Clear error messages for users
- ✅ Technical details in logs only

### Error Logging

- ✅ Structured logging with zerolog
- ✅ Context propagation through error chain
- ✅ Error metrics collected automatically
- ✅ Error rates calculated and logged every 5 minutes
- ✅ No sensitive information in logs

### System Behavior

- ✅ No crashes or hangs during network errors
- ✅ Graceful degradation during outages
- ✅ Automatic recovery after network restoration
- ✅ Delta sync resumes after recovery
- ✅ File operations eventually succeed

---

## Troubleshooting

### Issue: Tests fail with permission errors

**Solution**: Run tests in Docker with proper user mapping:
```bash
USER_ID=$(id -u) GROUP_ID=$(id -g) docker compose -f docker/compose/docker-compose.test.yml run --rm integration-tests
```

### Issue: Network simulation doesn't work

**Solution**: Ensure you have root privileges and required tools:
```bash
# Install traffic control tools
sudo apt-get install iproute2 iptables

# Verify tc is available
which tc

# Verify iptables is available
which iptables
```

### Issue: Logs not showing retry attempts

**Solution**: Enable debug logging:
```bash
# Set log level to debug
export ONEMOUNT_LOG_LEVEL=debug

# Remount filesystem
onemount unmount /mnt/onedrive
onemount mount /mnt/onedrive
```

---

## References

- **Error Handling Design**: `docs/2-architecture/error-handling.md`
- **Retry Logic**: `internal/retry/retry.go`
- **Error Types**: `internal/errors/error_types.go`
- **Logging**: `internal/logging/`
- **Integration Tests**: `internal/fs/*_integration_test.go`

---

## Test Results Template

```markdown
## Network Error Testing Results

**Date**: YYYY-MM-DD  
**Tester**: [Name]  
**OneMount Version**: [Version]

### Scenario 1: Network Timeout
- Status: ✅ PASS / ❌ FAIL
- Notes: [Observations]

### Scenario 2: Connection Refused
- Status: ✅ PASS / ❌ FAIL
- Notes: [Observations]

### Scenario 3: DNS Failure
- Status: ✅ PASS / ❌ FAIL
- Notes: [Observations]

### Scenario 4: Intermittent Failures
- Status: ✅ PASS / ❌ FAIL
- Notes: [Observations]

### Scenario 5: High Latency
- Status: ✅ PASS / ❌ FAIL
- Notes: [Observations]

### Scenario 6: Network Restoration
- Status: ✅ PASS / ❌ FAIL
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
