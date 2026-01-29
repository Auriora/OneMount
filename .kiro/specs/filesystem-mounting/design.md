# Design: Filesystem Mounting

## Source
Moved from `.kiro/specs/archive/system-verification-and-fix/design.md`. Sections below are verbatim.

### 2. Basic Filesystem Mounting Component

**Location**: `internal/fs/raw_filesystem.go`, `cmd/onemount/main.go`

**Verification Steps**:
1. Review FUSE initialization code
2. Test mounting at various mount points
3. Test mount point validation
4. Test unmounting and cleanup
5. Verify signal handling for graceful shutdown

**Expected Interfaces**:
- `NewFilesystem()` constructor
- `Mount()` method to mount filesystem
- `Unmount()` method for cleanup
- Signal handlers for SIGINT, SIGTERM

**Verification Criteria**:
- Filesystem mounts successfully at specified path
- Mount fails gracefully if path is in use
- Unmount releases all resources
- Signal handlers trigger clean shutdown
- No orphaned processes or mount points after exit
- Root directory contents are visible after mount
- Standard file operations work correctly

### 2C. Advanced Mounting Options Component

**Location**: `cmd/onemount/main.go`, `internal/fs/daemon.go`, `internal/fs/database.go`

**Verification Steps**:
1. Review daemon mode implementation
2. Test mount timeout configuration
3. Test stale lock file detection and cleanup
4. Test database retry logic with exponential backoff
5. Test command-line option parsing

**Expected Interfaces**:
- `ForkDaemon()` method for background operation
- `SetMountTimeout()` for timeout configuration
- `DetectStaleLocks()` for lock file cleanup
- `RetryWithBackoff()` for database operations

**Verification Criteria**:
- Daemon mode forks process and detaches from terminal
- Mount timeout is configurable and enforced
- Stale lock files (>5 minutes) are detected and removed
- Database retries use exponential backoff (max 10 attempts)
- Configuration validation provides clear error messages

### Filesystem Mounting Properties

**Property 5: FUSE Mount Success**
*For any* valid mount point specification, the system should successfully mount OneDrive using FUSE
**Validates: Requirements 2.1**

**Property 6: Non-blocking Initial Sync**
*For any* first-time filesystem mount, the initial directory structure fetch should complete while keeping interactive operations responsive
**Validates: Requirements 2A.1**

**Property 7: Root Directory Visibility**
*For any* successfully mounted filesystem, the root directory contents should be visible and accessible
**Validates: Requirements 2.2**

**Property 8: Standard File Operations Support**
*For any* mounted filesystem, standard file operations (ls, cat, cp, etc.) should work correctly
**Validates: Requirements 2.3**

**Property 9: Mount Conflict Error Handling**
*For any* mount point that is already in use, the system should display a clear error message identifying the conflicting process
**Validates: Requirements 2.4**

**Property 10: Clean Resource Release**
*For any* mounted filesystem, unmounting should cleanly release all associated resources
**Validates: Requirements 2.5**
