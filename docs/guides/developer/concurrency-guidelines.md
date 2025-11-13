# Concurrency and Lock Ordering Guidelines

## Overview

This document provides comprehensive guidelines for concurrent programming in OneMount, with a specific focus on lock ordering policies to prevent deadlocks. All developers working on OneMount must follow these guidelines to ensure thread-safe and deadlock-free code.

## Table of Contents

1. [Lock Ordering Policy](#lock-ordering-policy)
2. [Lock Hierarchy](#lock-hierarchy)
3. [Common Lock Patterns](#common-lock-patterns)
4. [Deadlock Prevention Rules](#deadlock-prevention-rules)
5. [Code Examples](#code-examples)
6. [Testing for Concurrency Issues](#testing-for-concurrency-issues)
7. [Common Pitfalls](#common-pitfalls)

## Lock Ordering Policy

### Core Principle

**Always acquire locks in a consistent, hierarchical order to prevent deadlocks.**

When multiple locks must be acquired, they MUST be acquired in the following order:

1. **Filesystem-level locks** (`Filesystem.RWMutex`)
2. **Manager-level locks** (DownloadManager, UploadManager, etc.)
3. **Inode-level locks** (`Inode.mu`)
4. **Session-level locks** (UploadSession, DownloadSession)
5. **Cache-level locks** (LoopbackCache, statusCache)

### Rationale

This ordering ensures that:
- Higher-level operations can safely access lower-level resources
- Lower-level operations never need to acquire higher-level locks
- Deadlocks are prevented by eliminating circular wait conditions

## Lock Hierarchy

### Level 1: Filesystem Locks

**Location**: `internal/fs/filesystem_types.go`

```go
type Filesystem struct {
    sync.RWMutex          // Protects: offline, lastNodeID, inodes
    opendirsM sync.RWMutex        // Protects: opendirs map
    statusM   sync.RWMutex        // Protects: statuses map
    statfsWarningM sync.RWMutex   // Protects: statfsWarningTime
    xattrSupportedM sync.RWMutex  // Protects: xattrSupported flag
}
```

**Purpose**: Protects filesystem-wide state including offline status, node ID allocation, and global maps.

**Acquisition Rules**:
- Acquire BEFORE any inode locks
- Acquire BEFORE any manager locks
- Use RLock for read-only operations
- Keep lock duration minimal

**Example**:
```go
// CORRECT: Filesystem lock before inode lock
f.RLock()
offline := f.offline
f.RUnlock()

inode.mu.Lock()
// ... modify inode ...
inode.mu.Unlock()
```

### Level 2: Manager Locks

**Locations**:
- `internal/fs/download_manager.go` - `DownloadManager.mutex`
- `internal/fs/upload_manager.go` - `UploadManager.mutex`
- `internal/fs/dbus.go` - `FileStatusDBusServer.mutex`
- `internal/fs/sync.go` - `SyncProgress.mutex`

```go
type DownloadManager struct {
    mutex sync.RWMutex  // Protects: sessions map
}

type UploadManager struct {
    mutex sync.RWMutex  // Protects: sessions map, state
}
```

**Purpose**: Protects manager-specific state and session maps.

**Acquisition Rules**:
- Acquire AFTER filesystem locks (if needed)
- Acquire BEFORE inode locks
- Acquire BEFORE session locks
- Never hold while making network calls

**Example**:
```go
// CORRECT: Manager lock before session lock
dm.mutex.RLock()
session, exists := dm.sessions[id]
dm.mutex.RUnlock()

if exists {
    session.mutex.Lock()
    // ... modify session ...
    session.mutex.Unlock()
}
```

### Level 3: Inode Locks

**Location**: `internal/fs/inode_types.go`

```go
type Inode struct {
    mu *sync.RWMutex  // Protects: all inode fields
}
```

**Purpose**: Protects individual inode state including metadata, children, and flags.

**Acquisition Rules**:
- Acquire AFTER filesystem locks
- Acquire AFTER manager locks
- NEVER acquire multiple inode locks simultaneously (see exception below)
- Use RLock for read-only operations
- Release before making network calls

**Exception for Multiple Inodes**:
When you must lock multiple inodes (e.g., during move operations), use **ID-based ordering**:

```go
// CORRECT: Lock inodes in ID order
inode1ID := inode1.ID()
inode2ID := inode2.ID()

if inode1ID < inode2ID {
    inode1.mu.Lock()
    inode2.mu.Lock()
} else {
    inode2.mu.Lock()
    inode1.mu.Lock()
}
defer inode1.mu.Unlock()
defer inode2.mu.Unlock()
```

### Level 4: Session Locks

**Locations**:
- `internal/fs/upload_session.go` - `UploadSession.Mutex`
- `internal/fs/download_manager.go` - `DownloadSession.mutex`

```go
type UploadSession struct {
    sync.Mutex  // Protects: session state
}

type DownloadSession struct {
    mutex sync.RWMutex  // Protects: session state
}
```

**Purpose**: Protects individual upload/download session state.

**Acquisition Rules**:
- Acquire AFTER manager locks
- Acquire AFTER inode locks
- Short-lived locks only
- Release before network I/O

### Level 5: Cache Locks

**Locations**:
- `internal/fs/content_cache.go` - `LoopbackCache.entriesM`
- `internal/fs/file_status.go` - `statusCache.mutex`
- `internal/fs/stats.go` - `CachedStats.mu`

```go
type LoopbackCache struct {
    entriesM sync.RWMutex  // Protects: entries map, totalSize
}

type statusCache struct {
    mutex sync.RWMutex  // Protects: entries map
}
```

**Purpose**: Protects cache data structures.

**Acquisition Rules**:
- Acquire AFTER all other locks
- Keep lock duration minimal
- Use RLock for lookups

## Common Lock Patterns

### Pattern 1: Read Filesystem State, Modify Inode

```go
// CORRECT: Filesystem read lock, then inode write lock
func (f *Filesystem) updateInodeOffline(inode *Inode) {
    // Check filesystem state
    f.RLock()
    isOffline := f.offline
    f.RUnlock()
    
    // Modify inode based on state
    inode.mu.Lock()
    defer inode.mu.Unlock()
    
    if isOffline {
        // ... update inode for offline mode ...
    }
}
```

### Pattern 2: Manager Operation on Session

```go
// CORRECT: Manager lock, then session lock
func (dm *DownloadManager) updateSession(id string) error {
    // Find session under manager lock
    dm.mutex.RLock()
    session, exists := dm.sessions[id]
    dm.mutex.RUnlock()
    
    if !exists {
        return errors.New("session not found")
    }
    
    // Update session under its own lock
    session.mutex.Lock()
    defer session.mutex.Unlock()
    
    session.State = StateDownloading
    return nil
}
```

### Pattern 3: Inode to Cache

```go
// CORRECT: Inode lock, then cache operation
func (f *Filesystem) cacheInodeContent(inode *Inode, data []byte) error {
    // Read inode ID under lock
    inode.mu.RLock()
    id := inode.DriveItem.ID
    inode.mu.RUnlock()
    
    // Cache operation (internally locks cache)
    return f.content.Insert(id, data)
}
```

### Pattern 4: Multiple Filesystem Locks

```go
// CORRECT: Acquire specialized locks in consistent order
func (f *Filesystem) updateFileStatus(id string, status FileStatusInfo) {
    // Order: statusM before xattrSupportedM
    f.statusM.Lock()
    f.statuses[id] = status
    f.statusM.Unlock()
    
    // Check xattr support
    f.xattrSupportedM.RLock()
    supported := f.xattrSupported
    f.xattrSupportedM.RUnlock()
    
    if supported {
        // ... update xattrs ...
    }
}
```

## Deadlock Prevention Rules

### Rule 1: Never Acquire Locks in Reverse Order

```go
// WRONG: Inode lock before filesystem lock
inode.mu.Lock()
f.Lock()  // DEADLOCK RISK!
```

```go
// CORRECT: Filesystem lock before inode lock
f.Lock()
inode.mu.Lock()
```

### Rule 2: Never Hold Locks During Network I/O

```go
// WRONG: Holding lock during network call
inode.mu.Lock()
data, err := graph.GetItemContent(inode.ID(), auth)  // BLOCKS!
inode.mu.Unlock()
```

```go
// CORRECT: Release lock before network call
inode.mu.RLock()
id := inode.DriveItem.ID
inode.mu.RUnlock()

data, err := graph.GetItemContent(id, auth)

if err == nil {
    inode.mu.Lock()
    // ... update inode with data ...
    inode.mu.Unlock()
}
```

### Rule 3: Use Defer for Lock Release

```go
// CORRECT: Always use defer for lock release
func (f *Filesystem) safeOperation() error {
    f.Lock()
    defer f.Unlock()
    
    // ... operation that might return early or panic ...
    
    return nil
}
```

### Rule 4: Minimize Lock Duration

```go
// WRONG: Holding lock too long
inode.mu.Lock()
defer inode.mu.Unlock()

// ... lots of computation ...
// ... network calls ...
// ... file I/O ...
```

```go
// CORRECT: Lock only what's necessary
inode.mu.RLock()
id := inode.DriveItem.ID
name := inode.DriveItem.Name
inode.mu.RUnlock()

// ... computation using id and name ...

inode.mu.Lock()
inode.hasChanges = true
inode.mu.Unlock()
```

### Rule 5: Document Lock Requirements

```go
// CORRECT: Document lock requirements in comments
// updateInodeMetadata updates the inode's metadata.
// Lock ordering: Must NOT hold filesystem lock when calling this method.
// This method acquires inode.mu internally.
func (f *Filesystem) updateInodeMetadata(inode *Inode, item *graph.DriveItem) {
    inode.mu.Lock()
    defer inode.mu.Unlock()
    
    inode.DriveItem = *item
}
```

## Code Examples

### Example 1: File Open Operation

```go
// Open implements the FUSE Open operation
// Lock ordering: filesystem.opendirsM -> inode.mu -> content cache (internal)
func (f *Filesystem) Open(cancel <-chan struct{}, in *fuse.OpenIn, out *fuse.OpenOut) fuse.Status {
    // Get inode without locks
    inode := f.GetID(in.NodeId)
    if inode == nil {
        return fuse.ENOENT
    }
    
    // Check filesystem state
    f.RLock()
    offline := f.offline
    f.RUnlock()
    
    // Lock inode for content verification
    inode.mu.Lock()
    id := inode.DriveItem.ID
    size := inode.DriveItem.Size
    inode.mu.Unlock()
    
    // Open cache file (cache locks internally)
    fd, err := f.content.Open(id)
    if err != nil {
        return fuse.EIO
    }
    
    // Verify content if online
    if !offline {
        // ... verification logic ...
    }
    
    return fuse.OK
}
```

### Example 2: Upload Queue Operation

```go
// QueueUpload queues a file for upload
// Lock ordering: inode.mu -> uploadManager.mutex
func (f *Filesystem) QueueUpload(inode *Inode, priority Priority) error {
    // Read inode data under lock
    inode.mu.RLock()
    id := inode.DriveItem.ID
    hasChanges := inode.hasChanges
    inode.mu.RUnlock()
    
    if !hasChanges {
        return nil  // Nothing to upload
    }
    
    // Get content (cache locks internally)
    data := f.getInodeContent(inode)
    
    // Create session
    session, err := NewUploadSession(inode, data)
    if err != nil {
        return err
    }
    
    // Queue with upload manager (manager locks internally)
    return f.uploads.QueueUpload(session, priority)
}
```

### Example 3: Delta Sync Update

```go
// applyDeltaChange applies a change from delta sync
// Lock ordering: filesystem.RWMutex -> inode.mu -> cache (internal)
func (f *Filesystem) applyDeltaChange(item *graph.DriveItem) error {
    // Check if item exists
    inode := f.GetID(item.ID)
    
    if inode == nil {
        // New item - add to filesystem
        f.Lock()
        // ... create new inode ...
        f.Unlock()
        return nil
    }
    
    // Existing item - update metadata
    inode.mu.Lock()
    oldETag := inode.DriveItem.ETag
    inode.DriveItem = *item
    newETag := item.ETag
    inode.mu.Unlock()
    
    // Invalidate cache if ETag changed
    if oldETag != newETag {
        f.content.Delete(item.ID)  // Cache locks internally
        f.MarkFileOutofSync(item.ID)
    }
    
    return nil
}
```

### Example 4: Move Operation (Multiple Inodes)

```go
// moveInode moves an inode from one parent to another
// Lock ordering: filesystem.RWMutex -> parent inodes (ID order) -> child inode
func (f *Filesystem) moveInode(child *Inode, oldParent, newParent *Inode) error {
    // Lock parents in ID order to prevent deadlock
    oldParentID := oldParent.ID()
    newParentID := newParent.ID()
    
    var first, second *Inode
    if oldParentID < newParentID {
        first, second = oldParent, newParent
    } else {
        first, second = newParent, oldParent
    }
    
    first.mu.Lock()
    defer first.mu.Unlock()
    
    if first != second {
        second.mu.Lock()
        defer second.mu.Unlock()
    }
    
    // Now lock child
    child.mu.Lock()
    defer child.mu.Unlock()
    
    // Perform move operation
    // ... update parent references ...
    // ... update children lists ...
    
    return nil
}
```

## Testing for Concurrency Issues

### Using the Race Detector

Always run tests with the race detector enabled:

```bash
# In Docker container
docker compose -f docker/compose/docker-compose.test.yml run --rm \
  test-runner go test -race ./internal/fs
```

### Stress Testing

Create tests that exercise concurrent operations:

```go
func TestConcurrentFileAccess(t *testing.T) {
    // Create filesystem
    fs := setupTestFilesystem(t)
    defer fs.Cleanup()
    
    // Create test file
    inode := createTestFile(t, fs, "test.txt")
    
    // Concurrent readers
    var wg sync.WaitGroup
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            
            // Read file multiple times
            for j := 0; j < 100; j++ {
                data := fs.GetInodeContent(inode)
                if data == nil {
                    t.Error("Failed to read file")
                }
            }
        }()
    }
    
    // Concurrent writer
    wg.Add(1)
    go func() {
        defer wg.Done()
        
        for j := 0; j < 100; j++ {
            inode.mu.Lock()
            inode.hasChanges = true
            inode.mu.Unlock()
            time.Sleep(time.Millisecond)
        }
    }()
    
    wg.Wait()
}
```

### Deadlock Detection

Use timeouts in tests to detect potential deadlocks:

```go
func TestNoDeadlock(t *testing.T) {
    done := make(chan bool)
    
    go func() {
        // Operation that might deadlock
        performComplexOperation()
        done <- true
    }()
    
    select {
    case <-done:
        // Success
    case <-time.After(5 * time.Second):
        t.Fatal("Operation timed out - possible deadlock")
    }
}
```

## Common Pitfalls

### Pitfall 1: Lock Inversion

```go
// WRONG: Different functions acquire locks in different orders
func functionA() {
    f.Lock()
    inode.mu.Lock()
    // ...
    inode.mu.Unlock()
    f.Unlock()
}

func functionB() {
    inode.mu.Lock()  // DEADLOCK RISK!
    f.Lock()
    // ...
    f.Unlock()
    inode.mu.Unlock()
}
```

**Solution**: Always follow the lock hierarchy.

### Pitfall 2: Holding Locks Across Goroutines

```go
// WRONG: Lock acquired in one goroutine, released in another
inode.mu.Lock()
go func() {
    // ... work ...
    inode.mu.Unlock()  // WRONG!
}()
```

**Solution**: Acquire and release locks in the same goroutine.

### Pitfall 3: Forgetting to Release Locks

```go
// WRONG: Lock not released on error path
func riskyOperation() error {
    inode.mu.Lock()
    
    if err := someOperation(); err != nil {
        return err  // LOCK NOT RELEASED!
    }
    
    inode.mu.Unlock()
    return nil
}
```

**Solution**: Always use defer for lock release.

### Pitfall 4: Recursive Locking

```go
// WRONG: Attempting to acquire same lock twice
func recursiveFunction(depth int) {
    f.Lock()
    defer f.Unlock()
    
    if depth > 0 {
        recursiveFunction(depth - 1)  // DEADLOCK!
    }
}
```

**Solution**: Restructure code to avoid recursive locking, or use RWMutex with RLock for read-only recursive calls.

### Pitfall 5: Lock Contention

```go
// WRONG: Single lock protecting unrelated data
type Manager struct {
    mu sync.Mutex
    sessions map[string]*Session
    stats Statistics
    config Config
}
```

**Solution**: Use separate locks for independent data:

```go
// CORRECT: Separate locks for independent data
type Manager struct {
    sessionsMu sync.RWMutex
    sessions   map[string]*Session
    
    statsMu sync.RWMutex
    stats   Statistics
    
    configMu sync.RWMutex
    config   Config
}
```

## Code Review Checklist

When reviewing code that uses locks, check for:

- [ ] Locks are acquired in the correct hierarchical order
- [ ] Locks are released using defer
- [ ] Locks are not held during network I/O or blocking operations
- [ ] Lock duration is minimized
- [ ] Multiple inode locks use ID-based ordering
- [ ] Lock requirements are documented in function comments
- [ ] Race detector passes on all tests
- [ ] No locks are held across goroutine boundaries
- [ ] RLock is used for read-only operations
- [ ] Error paths properly release locks

## References

- [Threading Guidelines](threading-guidelines.md) - Overview of threading in OneMount
- [Error Handling Guidelines](error-handling-guidelines.md) - Error handling patterns
- [Go Concurrency Patterns](https://go.dev/blog/pipelines) - Official Go blog
- [Effective Go - Concurrency](https://go.dev/doc/effective_go#concurrency) - Official Go documentation

## Revision History

- 2025-11-13: Initial version documenting lock ordering policy and concurrency guidelines
