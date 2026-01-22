# Fix: D-Bus GetFileStatus Returns Unknown (Task 41 / Issue #FS-001)

## Summary

Fixed the D-Bus `GetFileStatus` method that was returning "Unknown" for all file paths by adding a proper path-to-ID conversion method to the filesystem interface.

**Date**: 2026-01-22  
**Issue**: #FS-001  
**Task**: 41  
**Requirements**: 8.2

## Problem

The D-Bus `GetFileStatus` method was returning "Unknown" status for all file paths because:

1. It relied on a complex `findInodeByPath()` method that traversed the filesystem tree
2. The traversal logic required casting the `FilesystemInterface` to `*Filesystem` to access internal fields
3. The path traversal was fragile and could fail if the filesystem structure wasn't fully populated
4. There was no clean interface method to convert paths to IDs

## Solution

### 1. Added `GetIDByPath` Method to FilesystemInterface

**File**: `internal/fs/filesystem_types.go`

Added a new method to the `FilesystemInterface`:

```go
type FilesystemInterface interface {
    // ... existing methods ...
    GetIDByPath(path string) string
    // ... existing methods ...
}
```

### 2. Implemented `GetIDByPath` in Filesystem

**File**: `internal/fs/cache.go`

Implemented the method to properly traverse the filesystem tree and return the OneDrive ID for a given path:

```go
func (f *Filesystem) GetIDByPath(path string) string {
    // Handle root path
    if path == "" || path == "/" {
        return f.root
    }

    // Remove leading slash and split into components
    // Traverse from root to find the target inode
    // Return the ID or empty string if not found
}
```

Also added a helper function `splitPathComponents()` to properly split paths into components.

### 3. Updated D-Bus GetFileStatus Method

**File**: `internal/fs/dbus.go`

Simplified the `GetFileStatus` method to use the new interface method:

```go
func (s *FileStatusDBusServer) GetFileStatus(path string) (string, *dbus.Error) {
    // Use the filesystem's GetIDByPath method
    id := s.fs.GetIDByPath(path)
    if id == "" {
        return "Unknown", nil
    }

    // Get the file status for this ID
    status := s.fs.GetFileStatus(id)
    return status.Status.String(), nil
}
```

Removed the old `findInodeByPath()` and `splitPath()` methods that were no longer needed.

### 4. Created Comprehensive Integration Tests

**File**: `internal/fs/dbus_getfilestatus_test.go`

Created new integration tests to verify the fix:

- `TestDBusServer_GetFileStatus_ValidPaths` - Tests with valid file paths
- `TestDBusServer_GetFileStatus_InvalidPaths` - Tests with non-existent paths
- `TestDBusServer_GetFileStatus_StatusChanges` - Tests status changes over time
- `TestDBusServer_GetFileStatus_SpecialCharacters` - Tests paths with special characters

### 5. Updated Existing Tests

**Files**: 
- `internal/fs/dbus_test.go` - Updated to use new method names
- `internal/fs/upload_signal_simple_test.go` - Added `GetIDByPath` to mock interface

### 6. Created Manual Verification Guide

**File**: `docs/verification-task-41-dbus-getfilestatus.md`

Created comprehensive manual verification instructions for testing with Nemo file manager.

## Benefits

1. **Cleaner Architecture**: Path-to-ID conversion is now part of the filesystem interface
2. **More Reliable**: Uses the same path traversal logic as other filesystem operations
3. **Better Testability**: Can be tested independently of D-Bus
4. **Easier Maintenance**: Simpler code with fewer dependencies
5. **Consistent Behavior**: All path-based operations use the same conversion logic

## Testing

### Unit Tests

All existing unit tests pass with the new implementation:

```bash
go test -v -run TestSplitPathComponents ./internal/fs
go test -v -run TestFindInodeByPath_PathTraversal ./internal/fs
```

### Integration Tests

New integration tests verify the fix (require D-Bus session bus):

```bash
go test -v -run TestDBusServer_GetFileStatus ./internal/fs
```

Note: D-Bus tests require a D-Bus session bus and cannot run in Docker containers without proper D-Bus configuration.

### Manual Verification

Follow the instructions in `docs/verification-task-41-dbus-getfilestatus.md` to verify:

1. D-Bus service registration
2. GetFileStatus returns correct status for existing files
3. Status icons display correctly in Nemo
4. Status updates work correctly

## Files Changed

### Modified Files

1. `internal/fs/filesystem_types.go` - Added `GetIDByPath` to interface
2. `internal/fs/cache.go` - Implemented `GetIDByPath` and `splitPathComponents`
3. `internal/fs/dbus.go` - Simplified `GetFileStatus`, removed old methods
4. `internal/fs/dbus_test.go` - Updated test to use new method names
5. `internal/fs/upload_signal_simple_test.go` - Updated mock interface

### New Files

1. `internal/fs/dbus_getfilestatus_test.go` - New integration tests
2. `docs/verification-task-41-dbus-getfilestatus.md` - Manual verification guide
3. `docs/fixes/task-41-dbus-getfilestatus-fix.md` - This document

## Backward Compatibility

This change is backward compatible:

- The `FilesystemInterface` is extended with a new method
- Existing code that doesn't use D-Bus is unaffected
- The D-Bus API remains the same (GetFileStatus method signature unchanged)
- Mock implementations need to add the new method (already done)

## Future Improvements

1. Consider caching path-to-ID mappings for frequently accessed paths
2. Add metrics to track GetFileStatus performance
3. Consider adding batch path-to-ID conversion for efficiency
4. Add more comprehensive error handling for edge cases

## Related Issues

- Issue #FS-001: D-Bus GetFileStatus returns Unknown
- Issue #FS-002: D-Bus service name discovery problem (separate issue)

## Related Requirements

- Requirement 8.2: D-Bus integration for file status updates
- Requirement 8.3: Nemo extension integration

## References

- [D-Bus Specification](https://dbus.freedesktop.org/doc/dbus-specification.html)
- [Nemo Extensions](https://github.com/linuxmint/nemo-extensions)
- [OneMount Architecture Documentation](../2-architecture/)
