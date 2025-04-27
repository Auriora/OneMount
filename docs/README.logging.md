# Method Logging Framework

This document describes the method logging framework implemented in the onedriver filesystem module. The framework provides a way to log method entry and exit, including parameters and return values, for all public methods in the core module.

## Overview

The logging framework consists of two main components:

1. `LogMethodCall()` - A function that logs method entry and returns the method name and start time.
2. `LogMethodReturn()` - A function that logs method exit, including return values and execution duration.

These functions use the zerolog library to produce structured logs that can be easily parsed and analyzed.

## How to Use

To add logging to a method, follow these patterns:

### For methods with simple return values

```go
func (f *Filesystem) IsOffline() bool {
    methodName, startTime := LogMethodCall()
    f.RLock()
    defer f.RUnlock()

    result := f.offline
    defer LogMethodReturn(methodName, startTime, result)
    return result
}
```

### For methods with error returns

```go
func (f *Filesystem) TrackOfflineChange(change *OfflineChange) error {
    methodName, startTime := LogMethodCall()
    defer func() {
        // We can't capture the return value directly in a defer, so we'll just log completion
        LogMethodReturn(methodName, startTime)
    }()

    // Method implementation...
    return someError
}
```

### For methods with pointer returns

```go
func (f *Filesystem) GetNodeID(nodeID uint64) *Inode {
    methodName, startTime := LogMethodCall()

    // Early return case
    if someCondition {
        defer LogMethodReturn(methodName, startTime, nil)
        return nil
    }

    result := someOperation()
    defer LogMethodReturn(methodName, startTime, result)
    return result
}
```

### For methods with multiple return values

For methods with multiple return values, you'll need to use named return values and a defer function:

```go
func (f *Filesystem) SomeMethod() (result1 Type1, result2 Type2, err error) {
    methodName, startTime := LogMethodCall()
    defer func() {
        LogMethodReturn(methodName, startTime, result1, result2, err)
    }()

    // Method implementation...
    result1 = ...
    result2 = ...
    err = ...
    return
}
```

## Methods to Instrument

The following methods should be instrumented with logging:

### Filesystem Methods

- IsOffline
- TrackOfflineChange
- ProcessOfflineChanges
- TranslateID
- GetNodeID
- InsertNodeID
- GetID
- InsertID
- InsertChild
- DeleteID
- GetChild
- GetChildrenID
- GetChildrenPath
- GetPath
- DeletePath
- InsertPath
- MoveID
- MovePath
- StartCacheCleanup
- StopCacheCleanup
- StopDeltaLoop
- StopDownloadManager
- StopUploadManager
- SerializeAll

### Inode Methods

- AsJSON
- String
- Name
- SetName
- NodeID
- SetNodeID
- ID
- ParentID
- Path
- HasChanges
- HasChildren
- IsDir
- Mode
- ModTime
- NLink
- Size

## Log Output

The logs produced by this framework include:

- Method name
- Entry/exit phase
- Goroutine ID (thread identifier)
- Parameters (for entry)
- Return values (for exit)
- Execution duration (for exit)

Example log entry:
```json
{"level":"debug","method":"IsOffline","phase":"entry","goroutine":"1","time":"2023-04-27T21:00:00Z","message":"Method called"}
{"level":"debug","method":"IsOffline","phase":"exit","goroutine":"1","duration_ms":0.123,"return1":false,"time":"2023-04-27T21:00:00Z","message":"Method completed"}
```

The `goroutine` field contains the ID of the goroutine (Go's lightweight thread) that executed the method. This is useful for tracking method calls across different threads, especially in concurrent operations.

## Testing

The logging framework includes tests in `logging_test.go` that verify:

1. Basic functionality of the logging functions
2. Integration with instrumented methods

Run the tests with:
```bash
go test -v ./fs/...
```
