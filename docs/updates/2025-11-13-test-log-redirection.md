# Test Log Redirection Feature

**Date**: 2025-11-13  
**Type**: Enhancement  
**Component**: Docker Test Environment / Logging  
**Status**: Implemented

## Summary

Added log file redirection capability to the test suite by configuring the application logger to write to files instead of stdout, making test output cleaner and easier to read.

## Problem

Test output included extensive debug logging from the application mixed with test results, making it difficult to quickly identify test pass/fail status and specific failures.

## Solution

Implemented a proper logger-based redirection approach:

1. **Logger Configuration**: Created `ConfigureTestLogging()` function that configures the application logger based on environment variables
2. **TestMain Integration**: Added `TestMain` to the fs package to initialize logging before tests run
3. **Separates Output**: Application debug logs go to timestamped files in `test-artifacts/logs/`, test framework output stays on console
4. **Configurable**: Can be enabled/disabled via `ONEMOUNT_LOG_TO_FILE` environment variable
5. **Default Enabled**: Enabled by default for all test services in docker-compose

## Changes Made

### New Files

1. **internal/fs/testing_helpers.go**
   - Created `ConfigureTestLogging()` function
   - Reads `ONEMOUNT_LOG_TO_FILE` and `ONEMOUNT_LOG_DIR` environment variables
   - Configures `logging.DefaultLogger` to write to timestamped log files
   - Falls back to stdout if log directory creation fails

2. **internal/fs/testing_main_test.go**
   - Added `TestMain` function to initialize logging before tests run
   - Calls `ConfigureTestLogging()` at test suite startup

### Modified Files

1. **docker/scripts/test-entrypoint.sh**
   - Added `LOG_TO_FILE` and `LOG_DIR` configuration variables
   - Added `--log-to-file` command-line option
   - No changes to test execution (logger handles redirection)

2. **docker/compose/docker-compose.test.yml**
   - Added `ONEMOUNT_LOG_TO_FILE=true` to all test services
   - Added `ONEMOUNT_LOG_DIR` environment variable
   - Log files are saved to mounted `test-artifacts/logs/` directory

3. **.kiro/steering/testing-conventions.md**
   - Updated documentation with log redirection examples
   - Added instructions for disabling log redirection when needed

4. **docs/testing/test-log-redirection.md**
   - Created comprehensive documentation for the feature

## Usage

### Default Behavior (Log to File)

```bash
# Verbose output goes to test-artifacts/logs/, console shows test results
docker compose -f docker/compose/docker-compose.test.yml run --rm integration-tests
```

### Disable Log Redirection

```bash
# Show all output on console
docker compose -f docker/compose/docker-compose.test.yml run --rm \
  -e ONEMOUNT_LOG_TO_FILE=false integration-tests
```

### Manual Control

```bash
# Use --log-to-file flag with test-runner
docker compose -f docker/compose/docker-compose.test.yml run --rm \
  test-runner integration --log-to-file
```

## Log File Locations

- **Container Path**: `/tmp/home-tester/.onemount-tests/logs/`
- **Host Path**: `test-artifacts/logs/` (mounted volume)
- **Naming**: `test-YYYYMMDD-HHMMSS.log`

Examples:
- `test-20251113-143022.log`
- `test-20251113-143145.log`
- `test-20251113-150230.log`

## Benefits

1. **Cleaner Console Output**: Test results are immediately visible without scrolling through debug logs
2. **Full Debug Access**: Complete application logs preserved in files for troubleshooting
3. **Proper Separation**: Application logs (from logger) vs test framework output (PASS/FAIL)
4. **Timestamped Logs**: Each test run creates a new log file with timestamp for historical tracking
5. **Flexible**: Can be disabled when full console output is needed for debugging
6. **Correct Approach**: Configures the logger itself rather than filtering shell output

## Testing

Tested with:
- Unit tests
- Integration tests
- System tests
- Coverage analysis

All test types successfully redirect verbose output while maintaining test result visibility.

## Related Files

- `internal/fs/testing_helpers.go` - Logger configuration implementation
- `internal/fs/testing_main_test.go` - TestMain initialization
- `docker/scripts/test-entrypoint.sh` - Environment variable support
- `docker/compose/docker-compose.test.yml` - Docker configuration
- `.kiro/steering/testing-conventions.md` - Testing conventions
- `docs/testing/test-log-redirection.md` - Feature documentation
- `test-artifacts/logs/` - Log output directory

## Rules Consulted

- `testing-conventions.md` (Priority 25) - Docker test environment guidelines
- `operational-best-practices.md` (Priority 40) - Tool usage and transparency
- `general-preferences.md` (Priority 50) - Code quality principles

## Rules Applied

- Followed Docker test environment conventions
- Maintained backward compatibility (can be disabled)
- Added comprehensive documentation
- Used existing log directory structure
