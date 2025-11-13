# Threading Implementation in OneMount

This document describes the threading implementation in onemount, including the concurrency patterns, synchronization mechanisms, and thread management strategies used throughout the codebase.

## Overview

onemount uses Go's concurrency primitives to implement a multi-threaded architecture that enables efficient file operations while maintaining responsiveness. The system runs as a single process with multiple goroutines (lightweight threads) for concurrent operations.

## Key Threading Components

### 1. Delta Synchronization

The Delta Synchronization system periodically polls for changes from OneDrive and applies them to the local filesystem.

**Implementation Details:**
- Located in `fs/delta.go`
- Runs as a background goroutine started during filesystem initialization
- Uses a WaitGroup (`deltaLoopWg`) to track the goroutine for proper shutdown
- Uses channels (`deltaLoopStop`) for signaling the goroutine to stop
- Uses context with timeouts for network operations to prevent hanging
- Spawns additional goroutines for specific tasks (subscription handling, processing offline changes)
- Uses mutexes to protect shared state (offline flag)

**Synchronization Pattern:**

Start the delta loop:
```
go f.DeltaLoop(interval)
```

Stop the delta loop:
```
close(f.deltaLoopStop)
f.deltaLoopWg.Wait()
```

### 2. Download Manager

The Download Manager handles file downloads from OneDrive in the background, using a worker pool pattern.

**Implementation Details:**
- Located in `fs/download_manager.go`
- Uses a configurable number of worker goroutines
- Uses a channel (`queue`) as a work queue to distribute download tasks to workers
- Uses a WaitGroup (`workerWg`) to track worker goroutines for proper shutdown
- Uses a channel (`stopChan`) for signaling workers to stop
- Uses mutexes to protect shared state (sessions map, session state)
- Implements a graceful shutdown mechanism with a timeout

**Worker Pool Pattern:**

Start workers:
```
for i := 0; i < dm.numWorkers; i++ {
    dm.workerWg.Add(1)
    go dm.worker()
}
```

Worker function:
```
func (dm *DownloadManager) worker() {
    defer dm.workerWg.Done()

    for {
        select {
        case id := <-dm.queue:
            dm.processDownload(id)
        case <-dm.stopChan:
            return
        }
    }
}
```

### 3. Upload Manager

The Upload Manager handles file uploads to OneDrive in the background, with support for prioritization.

**Implementation Details:**
- Located in `fs/upload_manager.go`
- Uses a single worker goroutine (`uploadLoop`) to manage uploads
- Uses multiple channels for different priority levels (`highPriorityQueue`, `lowPriorityQueue`)
- Uses a WaitGroup (`workerWg`) to track the worker goroutine for proper shutdown
- Uses a channel (`stopChan`) for signaling the worker to stop
- Uses a mutex to protect shared state (sessions map)
- Spawns additional goroutines for each upload
- Limits the number of concurrent uploads (`maxUploadsInFlight`)
- Implements a priority system for uploads

**Priority Queue Pattern:**

Queue an upload with priority:
```
// Select the appropriate queue based on priority
var targetQueue chan *UploadSession
if priority == PriorityHigh {
    targetQueue = u.highPriorityQueue
} else {
    targetQueue = u.lowPriorityQueue
}

// Try to send the session to the queue
select {
case targetQueue <- session:
    // Upload queued successfully
    return session, nil
default:
    // Queue is full
    return nil, errors.New("upload queue is full")
}
```

### 4. Cache Management

The Cache Management system handles local caching of file content and metadata, with background cleanup.

**Implementation Details:**
- Uses a background goroutine for cache cleanup
- Uses a WaitGroup (`cacheCleanupWg`) to track the goroutine for proper shutdown
- Uses mutexes to protect shared state (filesystem state, open directories map, file statuses map)

### 5. Subscription Handling

The Subscription system handles real-time notifications from OneDrive using WebSockets.

**Implementation Details:**
- Located in `fs/subscription.go`
- Uses a background goroutine to maintain the WebSocket connection
- Uses mutexes to protect shared state (connection state)
- Implements reconnection logic with exponential backoff

## Synchronization Mechanisms

onemount uses several synchronization mechanisms to coordinate goroutines and protect shared state:

### 1. Mutexes

Mutexes are used to protect shared resources from concurrent access:

- **RWMutex**: Used when there are multiple readers and fewer writers to a shared resource
  - Example: `fs.RWMutex` protects the filesystem state
  - Example: `dm.mutex` protects the download sessions map

- **Mutex**: Used for simple mutual exclusion
  - Example: `session.mutex` protects the session state
  - Example: `handleIDLock` protects the handle ID counter

### 2. WaitGroups

WaitGroups are used to wait for multiple goroutines to complete:

- Example: `deltaLoopWg` tracks the delta loop goroutine
- Example: `workerWg` tracks worker goroutines in the download and upload managers
- Example: `cacheCleanupWg` tracks the cache cleanup goroutine

### 3. Channels

Channels are used for communication and synchronization between goroutines:

- **Work Queues**: Distribute work to worker goroutines
  - Example: `dm.queue` in the download manager
  - Example: `u.highPriorityQueue` and `u.lowPriorityQueue` in the upload manager

- **Signal Channels**: Signal goroutines to stop
  - Example: `deltaLoopStop` signals the delta loop to stop
  - Example: `stopChan` signals workers to stop in the download and upload managers

- **Notification Channels**: Notify about events
  - Example: `subsc.C` notifies about subscription events

### 4. Context

Context is used for cancellation and timeout handling:

- Example: `fetchCtx` in the delta loop provides a timeout for delta fetching
- Example: `ctx` in `pollDeltas` provides a timeout for network requests

## Thread Lifecycle Management

onemount carefully manages the lifecycle of its goroutines to ensure proper startup and shutdown:

### 1. Startup

- Goroutines are started during filesystem initialization
- WaitGroups are incremented before starting goroutines
- Channels are created for communication

### 2. Operation

- Goroutines communicate through channels
- Shared state is protected by mutexes
- Errors are handled and logged

### 3. Shutdown

- Stop signals are sent through channels
- WaitGroups are waited on to ensure all goroutines complete
- Timeouts are used to prevent hanging during shutdown
- Resources are cleaned up

## Concurrency Patterns

onemount uses several concurrency patterns:

### 1. Worker Pool

Used in the Download Manager to process multiple downloads concurrently.

### 2. Priority Queue

Used in the Upload Manager to prioritize uploads based on importance.

### 3. Background Processing

Used for delta synchronization, cache cleanup, and subscription handling.

### 4. Fan-out

Used when spawning multiple worker goroutines.

### 5. Cancellation

Used to stop goroutines gracefully.

## Error Handling in Concurrent Code

onemount implements robust error handling in its concurrent code:

- Errors are logged with context
- Retries are implemented for transient failures
- Panics are recovered to prevent goroutine crashes
- Errors are propagated to the appropriate handlers

## Conclusion

The threading implementation in onemount demonstrates a well-designed concurrent system that leverages Go's goroutines, channels, and synchronization primitives to achieve efficient and reliable file operations. The system carefully manages thread lifecycle, protects shared state, and handles errors gracefully, resulting in a responsive and robust filesystem implementation.
