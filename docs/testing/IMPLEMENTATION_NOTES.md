# Test Log Redirection - Implementation Notes

## Approach

The log redirection feature uses a **logger-based approach** rather than shell output filtering. This is the correct way to handle application logging during tests.

## Why Logger-Based?

1. **Separation of Concerns**: Application logs (from `logging.Debug()`, etc.) are separate from test framework output (PASS/FAIL)
2. **Clean Implementation**: No need for grep filtering or complex shell piping
3. **Proper Control**: The application logger is configured once at test startup
4. **Maintainable**: Changes to log format don't break the redirection

## How It Works

```
Test Execution Flow:
1. go test starts
2. TestMain() is called (internal/fs/testing_main_test.go)
3. ConfigureTestLogging() checks ONEMOUNT_LOG_TO_FILE env var
4. If true: logging.DefaultLogger is reconfigured to write to file
5. Tests run with logs going to file
6. Test framework output (PASS/FAIL) still goes to stdout
```

## Key Files

- `internal/fs/testing_helpers.go` - Logger configuration logic
- `internal/fs/testing_main_test.go` - TestMain hook
- `docker/compose/docker-compose.test.yml` - Environment variables

## Environment Variables

- `ONEMOUNT_LOG_TO_FILE=true` - Enable log file redirection
- `ONEMOUNT_LOG_DIR=/path/to/logs` - Where to write log files (default: ~/.onemount-tests/logs)

## Testing

To test locally:

```bash
# With log redirection (default in Docker)
docker compose -f docker/compose/docker-compose.test.yml run --rm integration-tests

# Without log redirection
docker compose -f docker/compose/docker-compose.test.yml run --rm \
  -e ONEMOUNT_LOG_TO_FILE=false integration-tests
```


## Before vs After

### Before (Shell Filtering Approach - Incorrect)
```bash
# Attempted to filter output with grep
go test -v ./... 2>&1 | tee log.txt | grep -E '(PASS|FAIL|RUN)'
```

**Problems:**
- Fragile: breaks if test output format changes
- Incomplete: might miss important test information
- Wrong layer: filtering at shell level instead of logger level

### After (Logger Configuration - Correct)
```go
// Configure logger in TestMain
func TestMain(m *testing.M) {
    ConfigureTestLogging()  // Redirects logging.Debug(), etc. to file
    code := m.Run()
    os.Exit(code)
}
```

**Benefits:**
- Robust: works regardless of test output format
- Complete: all application logs captured
- Correct layer: configures the logger itself
- Clean separation: app logs vs test framework output

## Example Output

### Console (with log redirection enabled)
```
=== RUN   TestIT_FS_ETag
--- PASS: TestIT_FS_ETag (0.05s)
=== RUN   TestIT_FS_Metadata
--- PASS: TestIT_FS_Metadata (0.12s)
PASS
ok      github.com/auriora/onemount/internal/fs 0.234s
```

### Log File (test-artifacts/logs/test-20251113-071700.log)
```
{"level":"debug","op":"OpenDir","nodeID":1,"id":"root","time":"2025-11-13T07:17:00Z","message":"Starting OpenDir operation"}
{"level":"debug","op":"OpenDir","time":"2025-11-13T07:17:00Z","message":"About to call GetChildrenID"}
{"level":"debug","op":"OpenDir","childrenCount":5,"time":"2025-11-13T07:17:00Z","message":"Returned from GetChildrenID"}
...
```
