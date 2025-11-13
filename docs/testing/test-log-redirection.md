# Test Log Redirection

## Overview

The test suite now supports redirecting application debug logs to files while keeping test framework output (PASS/FAIL) visible on the console. This is achieved by configuring the application logger itself, not by filtering shell output.

## Quick Start

### Default Behavior (Recommended)

```bash
# Verbose output automatically goes to test-artifacts/logs/
docker compose -f docker/compose/docker-compose.test.yml run --rm integration-tests
```

Console output shows:
- Test names (RUN)
- Test results (PASS/FAIL)
- Coverage summaries
- Test framework output

Application debug logs are saved to: `test-artifacts/logs/test-TIMESTAMP.log`

### Show All Output on Console

```bash
# Disable log redirection to see everything
docker compose -f docker/compose/docker-compose.test.yml run --rm \
  -e ONEMOUNT_LOG_TO_FILE=false integration-tests
```

## Configuration

### Environment Variables

- `ONEMOUNT_LOG_TO_FILE`: Enable/disable log redirection (default: `true`)
- `ONEMOUNT_LOG_DIR`: Directory for log files (default: `~/.onemount-tests/logs`)

### Command-Line Flag

```bash
docker compose -f docker/compose/docker-compose.test.yml run --rm \
  test-runner integration --log-to-file
```

## Log File Locations


**Container**: `/tmp/home-tester/.onemount-tests/logs/`  
**Host**: `test-artifacts/logs/` (mounted volume)

Log files are named: `test-YYYYMMDD-HHMMSS.log`

## Examples

### View Recent Test Logs

```bash
# List recent log files
ls -lht test-artifacts/logs/ | head -10

# View latest test log
tail -f test-artifacts/logs/test-*.log
```

### Troubleshooting Failed Tests

When a test fails, the console shows which test failed. Check the application logs for debug details:

```bash
# Find the most recent log file
LOG_FILE=$(ls -t test-artifacts/logs/test-*.log | head -1)

# Search for debug logs related to the failed test
grep -A 20 "TestIT_FS_ETag" "$LOG_FILE"
```

## Benefits

- **Cleaner Console**: Test framework output is not mixed with application debug logs
- **Full Debug Access**: Complete application logs preserved for troubleshooting
- **Proper Separation**: Logger-based approach, not shell filtering
- **Historical Tracking**: Timestamped logs for each test run
- **Flexible**: Can be disabled when needed

## Implementation Details

The feature works by:

1. **TestMain Hook**: `internal/fs/testing_main_test.go` calls `ConfigureTestLogging()` before tests run
2. **Logger Configuration**: `internal/fs/testing_helpers.go` reconfigures `logging.DefaultLogger` to write to a file
3. **Environment Variables**: Docker compose sets `ONEMOUNT_LOG_TO_FILE=true` and `ONEMOUNT_LOG_DIR`
4. **Result**: Application logs (Debug, Info, etc.) go to file, test output (PASS/FAIL) stays on console
