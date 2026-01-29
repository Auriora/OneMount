# Design: Fuse Performance

## Source
Moved from `.kiro/specs/archive/system-verification-and-fix/design.md`. Sections below are verbatim.

### 2D. FUSE Operation Performance Component

**Location**: `internal/fs/fuse_operations.go`, `internal/fs/metadata_cache.go`

**Verification Steps**:
1. Review FUSE operation handlers
2. Test metadata-only operation serving
3. Test background worker delegation
4. Test operation response times
5. Test concurrent operation handling

**Expected Interfaces**:
- FUSE operation handlers (`readdir`, `getattr`, etc.)
- `ServeFromMetadata()` for local-only operations
- `DelegateToWorker()` for Graph API interactions
- Performance monitoring and metrics

**Verification Criteria**:
- All FUSE operations served from local metadata/cache
- No Graph API calls block FUSE threads
- Graph interactions delegated to background workers
- Operation response times meet performance requirements
- Concurrent operations handled safely

#### Daemon Mode

**Purpose**: Allow OneMount to run as a background service without blocking the terminal.

**Implementation** (`cmd/onemount/main.go`):
- Accepts `--daemon` flag to enable daemon mode
- Forks the process using `syscall.ForkExec`
- Creates new process group and session
- Redirects logs to file in cache directory
- Removes `--daemon` flag from child process arguments to prevent infinite forking
- Parent process exits after successful fork

**Behavior**:
- When `--daemon` is specified, the process forks and the parent exits immediately
- The child process continues running in the background
- All output is redirected to log files
- The daemon process can be stopped using standard signals (SIGTERM, SIGINT)

#### Mount Timeout Configuration

**Purpose**: Prevent indefinite hanging during mount operations, especially in containerized environments.

**Implementation** (`cmd/onemount/main.go`):
- Accepts `--mount-timeout` flag to specify timeout duration (e.g., "120s", "2m")
- Default timeout: 60 seconds
- Recommended for Docker: 120 seconds (due to network initialization delays)
- Uses context with timeout to enforce the limit
- Provides clear error message if mount times out

**Behavior**:
- Mount operation is wrapped in a context with the specified timeout
- If mount doesn't complete within the timeout, the operation is cancelled
- Error message indicates timeout occurred and suggests increasing the value
- Pre-mount connectivity check helps identify network issues early

#### Stale Lock File Detection and Cleanup

**Purpose**: Recover from crashes or improper shutdowns that leave database lock files behind.

**Implementation** (`internal/fs/cache.go`):
- Database initialization includes retry logic with exponential backoff
- Max retries: 10 attempts
- Initial backoff: 200ms, max backoff: 5 seconds
- Database timeout: 10 seconds per attempt
- **Stale lock detection**: Checks if lock file is older than 5 minutes
- If stale, attempts to remove the lock file and retry
- Provides clear error message if database remains locked after all retries

**Behavior**:
- When opening the BBolt database, if a lock file exists:
  1. Check the modification time of the lock file
  2. If older than 5 minutes, consider it stale
  3. Attempt to remove the stale lock file
  4. Retry database open operation
  5. If lock is not stale or removal fails, retry with exponential backoff
- Logs each retry attempt with backoff duration
- After 10 failed attempts, returns error with diagnostic information

## Concurrency and Lock Management

### Lock Ordering Policy

To prevent deadlocks, all components MUST acquire locks in the following order:

1. **Global Filesystem Lock** (`filesystem.mutex`)
2. **Mount Manager Lock** (`mountManager.mutex`) 
3. **Cache Manager Lock** (`cacheManager.mutex`)
4. **Individual Inode Locks** (`inode.mutex`)
5. **Worker Pool Locks** (`downloadManager.mutex`, `uploadManager.mutex`)
6. **Network State Lock** (`networkState.mutex`)

**Lock Ordering Rules**:
- Always acquire locks in the order listed above
- Never acquire a higher-numbered lock while holding a lower-numbered lock
- Release locks in reverse order (LIFO)
- Use `defer` statements to ensure proper lock release
- Minimize lock hold time by preparing data before acquiring locks

### Concurrent Operation Guidelines

**Safe Concurrent Operations**:
- Multiple file reads from different inodes (each inode has its own lock)
- Directory listings while files are being downloaded (read-only metadata access)
- Delta sync while serving file operations (background sync uses separate locks)
- Upload and download operations on different files (separate worker pools)

**Operations Requiring Coordination**:
- File modification + delta sync (coordinate via inode state)
- Cache eviction + file access (coordinate via cache manager)
- Filesystem shutdown + active operations (coordinate via context cancellation)

### Deadlock Prevention Strategies

**Lock Timeout Policy**:
```go
// Example lock acquisition with timeout
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

if !filesystem.mutex.TryLockContext(ctx) {
    return fmt.Errorf("failed to acquire filesystem lock within timeout")
}
defer filesystem.mutex.Unlock()
```

**Lock-Free Operations Where Possible**:
- File status queries use atomic operations
- Statistics collection uses read-only snapshots
- Network state checks use atomic boolean flags

**Goroutine Management**:
- All long-running goroutines MUST be tracked with `sync.WaitGroup`
- Use context cancellation for graceful shutdown
- Set reasonable timeouts for all blocking operations
- Implement circuit breakers for external API calls

### Lock Granularity Guidelines

**Filesystem Level** (Coarse-grained):
- Mount/unmount operations
- Global configuration changes
- Shutdown coordination

**Cache Level** (Medium-grained):
- Cache cleanup operations
- Cache size enforcement
- Statistics collection

**Inode Level** (Fine-grained):
- Individual file operations
- Metadata updates
- State transitions

**Operation Level** (Finest-grained):
- Network request queuing
- Worker thread coordination
- Status updates

### Race Condition Prevention

**Common Race Conditions and Solutions**:

1. **File State Changes During Access**:
   - Solution: Use atomic state transitions with compare-and-swap
   - Lock inode before checking and updating state

2. **Cache Invalidation During Read**:
   - Solution: Use reference counting for cache entries
   - Delay invalidation until all readers complete

3. **Concurrent Uploads of Same File**:
   - Solution: Use file-level upload locks
   - Queue subsequent uploads until first completes

4. **Delta Sync vs Local Modifications**:
   - Solution: Use ETag comparison with atomic updates
   - Detect conflicts before applying changes

### Performance Considerations

**Lock Contention Reduction**:
- Use read-write locks where appropriate (`sync.RWMutex`)
- Implement lock-free fast paths for common operations
- Batch operations to reduce lock acquisition frequency
- Use channels for producer-consumer patterns instead of shared state

**Monitoring and Debugging**:
- Log lock acquisition times in debug mode
- Implement lock contention metrics
- Use Go's race detector during testing
- Profile lock usage under load

### Lock Ordering Policy

To prevent deadlocks, all components MUST acquire locks in the following order:

1. **Global Filesystem Lock** (`filesystem.mutex`)
2. **Mount Manager Lock** (`mountManager.mutex`) 
3. **Cache Manager Lock** (`cacheManager.mutex`)
4. **Individual Inode Locks** (`inode.mutex`)
5. **Worker Pool Locks** (`downloadManager.mutex`, `uploadManager.mutex`)
6. **Network State Lock** (`networkState.mutex`)

**Lock Ordering Rules**:
- Always acquire locks in the order listed above
- Never acquire a higher-numbered lock while holding a lower-numbered lock
- Release locks in reverse order (LIFO)
- Use `defer` statements to ensure proper lock release
- Minimize lock hold time by preparing data before acquiring locks

### Concurrent Operation Guidelines

**Safe Concurrent Operations**:
- Multiple file reads from different inodes (each inode has its own lock)
- Directory listings while files are being downloaded (read-only metadata access)
- Delta sync while serving file operations (background sync uses separate locks)
- Upload and download operations on different files (separate worker pools)

**Operations Requiring Coordination**:
- File modification + delta sync (coordinate via inode state)
- Cache eviction + file access (coordinate via cache manager)
- Filesystem shutdown + active operations (coordinate via context cancellation)

### Deadlock Prevention Strategies

**Lock Timeout Policy**:
```go
// Example lock acquisition with timeout
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

if !filesystem.mutex.TryLockContext(ctx) {
    return fmt.Errorf("failed to acquire filesystem lock within timeout")
}
defer filesystem.mutex.Unlock()
```

**Lock-Free Operations Where Possible**:
- File status queries use atomic operations
- Statistics collection uses read-only snapshots
- Network state checks use atomic boolean flags

**Goroutine Management**:
- All long-running goroutines MUST be tracked with `sync.WaitGroup`
- Use context cancellation for graceful shutdown
- Set reasonable timeouts for all blocking operations
- Implement circuit breakers for external API calls

### Lock Granularity Guidelines

**Filesystem Level** (Coarse-grained):
- Mount/unmount operations
- Global configuration changes
- Shutdown coordination

**Cache Level** (Medium-grained):
- Cache cleanup operations
- Cache size enforcement
- Statistics collection

**Inode Level** (Fine-grained):
- Individual file operations
- Metadata updates
- State transitions

**Operation Level** (Finest-grained):
- Network request queuing
- Worker thread coordination
- Status updates

### Race Condition Prevention

**Common Race Conditions and Solutions**:

1. **File State Changes During Access**:
   - Solution: Use atomic state transitions with compare-and-swap
   - Lock inode before checking and updating state

2. **Cache Invalidation During Read**:
   - Solution: Use reference counting for cache entries
   - Delay invalidation until all readers complete

3. **Concurrent Uploads of Same File**:
   - Solution: Use file-level upload locks
   - Queue subsequent uploads until first completes

4. **Delta Sync vs Local Modifications**:
   - Solution: Use ETag comparison with atomic updates
   - Detect conflicts before applying changes

### Performance Considerations

**Lock Contention Reduction**:
- Use read-write locks where appropriate (`sync.RWMutex`)
- Implement lock-free fast paths for common operations
- Batch operations to reduce lock acquisition frequency
- Use channels for producer-consumer patterns instead of shared state

**Monitoring and Debugging**:
- Log lock acquisition times in debug mode
- Implement lock contention metrics
- Use Go's race detector during testing
- Profile lock usage under load

### Concurrency Properties

**Property 33: Safe Concurrent File Access**
*For any* simultaneous file access operations, the system should handle concurrent operations safely without race conditions
**Validates: Requirements 10.1**

**Property 34: Non-blocking Downloads**
*For any* ongoing download operations, other file operations should be allowed to proceed without blocking
**Validates: Requirements 10.2**

### Performance Properties

**Property 49: Directory Listing Performance**
*For any* directory listing operation with up to 1000 files, the system should respond within 2 seconds
**Validates: Requirements 23.1**

**Property 50: Cached File Access Performance**
*For any* cached file opening operation, the system should serve content within 100 milliseconds
**Validates: Requirements 23.2**

**Property 51: Idle Memory Usage**
*For any* idle system state, the system should consume no more than 50 MB of RAM
**Validates: Requirements 23.3**

**Property 52: Active Sync Memory Usage**
*For any* active file synchronization operation, the system should consume no more than 200 MB of RAM
**Validates: Requirements 23.4**

**Property 53: Concurrent Operations Performance**
*For any* concurrent file operations scenario with at least 10 simultaneous operations, the system should handle them without performance degradation
**Validates: Requirements 23.7**

**Property 54: Startup Performance**
*For any* system startup scenario, the system should complete initialization and be ready for file operations within 5 seconds
**Validates: Requirements 23.9**

**Property 55: Shutdown Performance**
*For any* system shutdown scenario, the system should complete graceful shutdown within 10 seconds
**Validates: Requirements 23.10**

### Resource Management Properties

**Property 56: Cache Size Enforcement**
*For any* cache size configuration, the system should enforce the specified maximum cache size limit
**Validates: Requirements 24.1**

**Property 57: File Descriptor Limits**
*For any* file descriptor management scenario, the system should not exceed 1000 open file descriptors simultaneously
**Validates: Requirements 24.4**

**Property 58: Worker Thread Limits**
*For any* worker thread spawning scenario, the system should limit concurrent workers to the configured maximum (default: 10)
**Validates: Requirements 24.5**

**Property 59: Adaptive Network Throttling**
*For any* limited network bandwidth scenario, the system should implement adaptive throttling to prevent network saturation
**Validates: Requirements 24.7**

**Property 60: Memory Pressure Handling**
*For any* system memory pressure scenario, the system should reduce in-memory caching and increase disk-based caching
**Validates: Requirements 24.8**

**Property 61: CPU Usage Management**
*For any* high CPU usage scenario, the system should reduce background processing priority to maintain system responsiveness
**Validates: Requirements 24.9**

**Property 62: Graceful Resource Degradation**
*For any* system resource pressure scenario, the system should gracefully degrade non-essential features while maintaining core functionality
**Validates: Requirements 24.10**

### Concurrency and Lock Management Properties

**Property 63: Lock Ordering Compliance**
*For any* sequence of lock acquisitions, the system should acquire locks in the defined order (filesystem → mount manager → cache manager → inode → worker pool → network state)
**Validates: Concurrency Design Requirements**

**Property 64: Deadlock Prevention**
*For any* concurrent operation scenario, the system should complete all operations without deadlocks when following the lock ordering policy
**Validates: Concurrency Design Requirements**

**Property 65: Lock Release Consistency**
*For any* lock acquisition, the system should release locks in reverse order (LIFO) and handle all error conditions properly
**Validates: Concurrency Design Requirements**

**Property 66: Concurrent File Access Safety**
*For any* set of concurrent file operations on different inodes, the system should handle all operations safely without race conditions
**Validates: Concurrency Design Requirements**

**Property 67: State Transition Atomicity**
*For any* item state transition, the system should complete the transition atomically without intermediate inconsistent states
**Validates: State Machine Design Requirements**

## Timeout Configuration

### Overview

All timeout values across OneMount components are centralized in the `TimeoutConfig` struct to ensure consistency and ease of configuration. This addresses Issue #PERF-003 (Inconsistent Timeout Values) by providing a single source of truth for all timeout-related settings.

### Timeout Categories

**Short Operations (< 5 seconds)**:
- Download Worker Shutdown: 5 seconds
- Network Callback Shutdown: 5 seconds
- Content Stats Timeout: 5 seconds

**Medium Operations (5-30 seconds)**:
- Metadata Request Timeout: 30 seconds

**Long Operations (30 seconds - 2 minutes)**:
- Upload Graceful Shutdown: 30 seconds

**Graceful Shutdown (10-60 seconds)**:
- Filesystem Shutdown: 10 seconds

### Configuration Structure

```go
type TimeoutConfig struct {
    DownloadWorkerShutdown  time.Duration // Time to wait for download workers to finish
    UploadGracefulShutdown  time.Duration // Time to wait for active uploads to complete
    FilesystemShutdown      time.Duration // Time to wait for all filesystem goroutines to stop
    NetworkCallbackShutdown time.Duration // Time to wait for network feedback callbacks
    MetadataRequestTimeout  time.Duration // Time to wait for metadata requests
    ContentStatsTimeout     time.Duration // Time to wait for content cache statistics
}
```

### Default Values

```go
DefaultTimeoutConfig() returns:
- DownloadWorkerShutdown:  5 * time.Second
- UploadGracefulShutdown:  30 * time.Second
- FilesystemShutdown:      10 * time.Second
- NetworkCallbackShutdown: 5 * time.Second
- MetadataRequestTimeout:  30 * time.Second
- ContentStatsTimeout:     5 * time.Second
```

### Validation

All timeout values are validated on initialization:
- Must be positive (> 0)
- Should be at least 1 second (for most timeouts)
- Should not exceed 5 minutes (to prevent indefinite hangs)

Invalid configurations result in clear error messages indicating the problem field and valid ranges.

### Usage in Components

**Download Manager**:
- Uses `DownloadWorkerShutdown` when stopping workers
- Logs timeout duration if workers don't finish in time

**Upload Manager**:
- Uses `UploadGracefulShutdown` for active upload completion
- Type asserts to `*Filesystem` to access configuration

**Filesystem**:
- Uses `FilesystemShutdown` when stopping all goroutines
- Initializes with `DefaultTimeoutConfig()` on creation

**Metadata Requests**:
- Uses `MetadataRequestTimeout` for fetch operations
- Logs timeout duration if requests don't complete

**Content Statistics**:
- Uses `ContentStatsTimeout` for statistics collection
- Falls back to partial results on timeout

### Future Enhancements

1. **Command-Line Configuration**: Add flags for timeout customization
2. **Configuration File**: Support timeout settings in config file
3. **Dynamic Adjustment**: Automatically adjust based on network conditions
4. **Timeout Metrics**: Collect metrics on timeout occurrences
5. **Timeout Profiles**: Predefined profiles for different environments

### Documentation

Detailed timeout policy documentation is available in:
- `docs/guides/developer/timeout-policy.md`

This document includes:
- Rationale for each timeout value
- Guidelines for when to adjust timeouts
- Troubleshooting timeout-related issues
- Best practices for timeout configuration
