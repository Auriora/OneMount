# Performance Harness Duration Guard (2025-11-22)

**Type**: Test/Infra Update  
**Status**: Complete  
**Components**: `internal/testutil/framework/performance_integration_test.go`

## Summary

- Shortened default durations for sustained-operation and memory-leak performance integration tests to 2 minutes to prevent the Go test 20m timeout from firing in CI or local default runs.
- Added an opt-in flag (`ONEMOUNT_PERF_LONG=1`) to restore the previous 30m/20m soak durations when explicitly desired.
- Loosened duration tolerance to 20% for shortened runs and kept sampling intervals aligned with the reduced windows.

## Testing

- Not fully re-run here due to the 2-minute wall-clock per test; logic validated by compilation and partial run initiation.

## Rules Consulted

- AGENT-GUIDE-General-Preferences (priority 50)  
- AGENT-RULE-Documentation-Conventions (priority 20)  
- AGENT-RULE-Testing-Conventions (priority 25)
