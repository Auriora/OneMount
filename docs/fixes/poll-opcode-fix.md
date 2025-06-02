# POLL Opcode Fix - Resolving OneMount Crashes

## Problem Description

OneMount was experiencing frequent crashes, particularly when copying large files (e.g., 2.5GB files that would pause at 2.4GB and then crash). The analysis of the logs revealed several critical issues:

### Primary Issues Identified

1. **POLL Opcode Implementation Conflict**
   - OneMount had implemented custom POLL opcode support
   - However, go-fuse v2.7.2 intentionally disables the POLL opcode to prevent deadlocks
   - This caused "Unimplemented opcode POLL" errors leading to crashes

2. **Service Timeout Issues**
   - Systemd service was timing out during startup (90-second timeout)
   - Service had restarted 72 times, indicating persistent startup issues
   - Large file operations were causing the service to hang

3. **Mount Point Issues**
   - fusermount3 reported "entry not found in /etc/mtab"
   - Filesystem wasn't properly mounting or was crashing during mount

## Root Cause Analysis

The main issue was that OneMount was trying to implement POLL support, but the go-fuse library intentionally disables this opcode. According to the [go-fuse documentation](https://pkg.go.dev/github.com/hanwen/go-fuse/v2/fs):

> "To prevent this from happening, Go-FUSE disables the POLL opcode on mount."

This is done to prevent deadlocks that can occur when the Go runtime multiplexes Goroutines onto operating system threads and makes assumptions about system calls not blocking.

## Solution Implemented

### 1. Removed POLL Opcode Implementation

**Files Modified:**
- `internal/fs/raw_filesystem.go` - Removed Poll method and updated comments
- `internal/fs/file_operations.go` - Removed Poll and PollOperationHandler methods
- `internal/fs/filesystem_types.go` - Removed Poll method from FilesystemInterface
- `internal/fs/cache.go` - Updated comment about CustomRawFileSystem purpose

**Changes Made:**
```go
// Before: CustomRawFileSystem with POLL support
func (c *CustomRawFileSystem) Poll(cancel <-chan struct{}, in *fuse.InHeader, out *fuse.OutHeader) fuse.Status {
    // ... POLL implementation
}

// After: CustomRawFileSystem without POLL support
// POLL opcode support has been removed as go-fuse intentionally disables it
// to prevent deadlocks. See: https://pkg.go.dev/github.com/hanwen/go-fuse/v2/fs
```

### 2. Enhanced Cancellation Handling

**Files Modified:**
- `internal/fs/file_operations.go` - Added cancellation checks to Read and Write operations

**Changes Made:**
```go
func (f *Filesystem) Write(cancel <-chan struct{}, in *fuse.WriteIn, data []byte) (uint32, fuse.Status) {
    // Check for cancellation before starting the write operation
    select {
    case <-cancel:
        logger.Debug().Msg("Write operation cancelled")
        return 0, fuse.EINTR
    default:
    }
    // ... rest of implementation
}
```

### 3. Added Large File Operation Warnings

**Files Modified:**
- `internal/fs/file_operations.go` - Added warnings for large file operations

**Changes Made:**
```go
// Check for large file operations and log warnings
const largeFileThreshold = 1024 * 1024 * 1024 // 1GB
if nWrite > largeFileThreshold {
    logger.Warn().
        Int("writeSize", nWrite).
        Msg("Large write operation detected - this may take some time")
}
```

### 4. Created Tests

**Files Created:**
- `tests/unit/poll_fix_test.go` - Tests to verify POLL removal and functionality

## Benefits of the Fix

1. **Eliminates POLL Opcode Errors**: No more "Unimplemented opcode POLL" errors
2. **Prevents Deadlocks**: Aligns with go-fuse's design to prevent runtime deadlocks
3. **Better Cancellation**: Operations can be properly cancelled to prevent hangs
4. **Improved Monitoring**: Large file operations are logged with warnings
5. **Stability**: Reduces service restarts and crashes

## Testing

The fix has been tested with:
- Unit tests verifying POLL method removal
- Compilation tests ensuring code builds correctly
- Verification that CustomRawFileSystem creates successfully

## Deployment Notes

1. **No Configuration Changes Required**: The fix is purely code-based
2. **Backward Compatible**: No breaking changes to the API
3. **Immediate Effect**: The fix takes effect as soon as the service is restarted

## Monitoring

After deployment, monitor for:
- Reduction in service restarts
- Absence of "Unimplemented opcode POLL" errors in logs
- Successful completion of large file transfers
- Proper mount/unmount operations

## References

- [go-fuse v2.7.2 Documentation](https://pkg.go.dev/github.com/hanwen/go-fuse/v2/fs)
- [FUSE POLL Opcode Documentation](https://www.kernel.org/doc/html/latest/filesystems/fuse.html)
- [Go Runtime and System Calls](https://golang.org/doc/faq#goroutines)
