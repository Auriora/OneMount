# Design: Offline Mode Sync

## Source
Moved from `.kiro/specs/archive/system-verification-and-fix/design.md`. Sections below are verbatim.

### 8. Offline Mode Component

**Location**: `internal/fs/offline.go`, `internal/graph/network_feedback.go`

**Verification Steps**:
1. Review offline detection logic (passive and active)
2. Test transition to offline mode
3. Test read-write operations with change queuing
4. Test change tracking in persistent storage
5. Test transition back to online mode
6. Test user notification mechanisms
7. Test configuration options

**Expected Interfaces**:
- `IsOffline()` method
- `SetOffline()` method
- `NetworkStateMonitor` for connectivity monitoring
- `ConnectivityChecker` for active checks
- `FeedbackManager` for user notifications
- Offline change tracking in database
- `OfflineChange` struct for tracking modifications
- `ProcessOfflineChanges()` method for upload queue processing

**Network Detection Mechanisms**:
1. **Passive Detection**: Monitors API call failures for network error patterns
2. **Active Detection**: Periodic connectivity checks to Microsoft Graph endpoints
3. **Error Pattern Analysis**: Recognizes specific network error strings

**User Feedback Levels**:
- **None**: No user notifications (logging only)
- **Basic**: Simple connectivity status messages
- **Detailed**: Comprehensive network and sync information

**Notification Types**:
- Network Connected/Disconnected
- Sync Started/Completed
- Conflicts Detected
- Sync Failed

**Configuration Options**:
- `checkInterval`: Network connectivity check frequency (default: 15s)
- `connectivityTimeout`: Timeout for connectivity checks (default: 10s)
- `feedbackLevel`: Level of user feedback (default: Basic)
- `cacheRetention`: Cache retention duration
- `maxPendingChanges`: Maximum pending changes to track (default: 1000)
- `conflictResolution`: Default conflict resolution strategy

**Offline Change Queuing**:

The system implements a read-write offline mode where all file operations are allowed while offline:

1. **File Modifications**: When a file is modified offline:
   - The change is written to the local cache
   - An `OfflineChange` record is created in the persistent database
   - The record includes: file path, operation type (create/modify/delete), timestamp, and local ETag
   - Multiple changes to the same file update the existing record with the most recent version

2. **File Creation**: When a file is created offline:
   - The file content is stored in the local cache
   - An `OfflineChange` record is created with operation type "create"
   - The file is assigned a temporary local ID until synchronized

3. **File Deletion**: When a file is deleted offline:
   - The file is removed from the local cache
   - An `OfflineChange` record is created with operation type "delete"
   - The deletion is queued for synchronization

4. **Change Processing**: When connectivity is restored:
   - All pending `OfflineChange` records are retrieved from the database
   - Changes are processed in batches to avoid overwhelming the server
   - For each change:
     - Check for conflicts by comparing local ETag with remote ETag
     - If no conflict, upload the change
     - If conflict detected, apply configured conflict resolution strategy
     - On successful upload, remove the `OfflineChange` record
     - On failure, retry with exponential backoff

5. **Conflict Resolution**: During offline-to-online synchronization:
   - Compare local ETag (from when file was downloaded) with current remote ETag
   - If ETags differ, a conflict exists
   - Apply configured strategy: last-writer-wins, keep-both, user-choice, merge, or rename
   - Default strategy is keep-both (creates conflict copy)

**Verification Criteria**:
- Network loss is detected automatically via passive monitoring
- Active connectivity checks run at configured intervals
- Filesystem allows read and write operations when offline
- Cached files remain accessible
- File modifications are tracked in persistent storage
- Multiple changes to the same file preserve the most recent version
- File creation and deletion operations are queued correctly
- Changes are queued for upload when online
- Online transition processes queued changes in batches
- Conflicts are detected using ETag comparison
- Configured conflict resolution strategy is applied
- User notifications are emitted according to feedback level
- Configuration options are respected

### 9. Offline-to-Online Synchronization Process

**Location**: `internal/fs/sync_manager.go`, `internal/fs/offline.go`

**Synchronization Steps**:
1. **Change Detection**: Identify all pending offline changes from database
2. **Conflict Analysis**: Check for server-side changes that conflict with local changes
3. **Upload Queue**: Queue local changes for upload to the server
4. **Batch Processing**: Process changes in batches to avoid overwhelming the server
5. **Verification**: Verify that all changes were successfully synchronized
6. **Cleanup**: Remove successfully synchronized changes from the pending queue

**Conflict Resolution Strategies**:
1. **Last Writer Wins**: Most recent modification takes precedence (compare timestamps)
2. **User Choice**: Present options to user for manual resolution
3. **Merge**: Attempt automatic merging for compatible changes
4. **Rename**: Create separate versions with conflict indicators
5. **Keep Both**: Create separate versions for both local and remote (default)

**Conflict Types**:
- **Content Conflicts**: Local and server versions have different content
- **Metadata Conflicts**: File properties differ between local and server
- **Existence Conflicts**: File exists locally but was deleted on server, or vice versa
- **Parent Conflicts**: Parent directory was moved or deleted on server

**Offline Change Data Model**:

```go
type OfflineChangeType string
const (
    OfflineChangeCreate OfflineChangeType = "create"
    OfflineChangeModify OfflineChangeType = "modify"
    OfflineChangeDelete OfflineChangeType = "delete"
)

type OfflineChange struct {
    ID            string            // Unique change ID
    ItemID        string            // OneDrive item ID (or temporary local ID)
    Path          string            // File path
    OperationType OfflineChangeType // Type of operation
    Timestamp     time.Time         // When change was made
    LocalETag     string            // ETag when file was downloaded
    ContentHash   string            // Hash of local content
    Size          int64             // File size
    RetryCount    int               // Number of upload attempts
    LastError     string            // Last error message (if any)
}

type OfflineChangeQueue struct {
    changes map[string]*OfflineChange // itemID -> change
    db      *bbolt.DB                  // Persistent storage
    mutex   sync.RWMutex               // Thread safety
}
```

**Change Queue Operations**:
- `AddChange(change *OfflineChange)`: Add or update a change in the queue
- `GetPendingChanges() []*OfflineChange`: Retrieve all pending changes
- `RemoveChange(changeID string)`: Remove a successfully synchronized change
- `UpdateRetryCount(changeID string)`: Increment retry counter after failed attempt
- `GetChangeByItemID(itemID string) *OfflineChange`: Get change for specific item

**Verification Criteria**:
- All pending changes are identified correctly
- Conflicts are detected before upload attempts using ETag comparison
- Appropriate resolution strategy is applied based on configuration
- Both versions are preserved when using keep-both strategy
- Synchronization completes successfully for all queued changes
- Failed changes are retried with exponential backoff
- Successfully synchronized changes are removed from the queue
- Change queue persists across filesystem restarts

### Offline Mode Properties

**Property 24: Offline Detection**
*For any* network connectivity loss, the system should detect the offline state through passive monitoring of API call failures
**Validates: Requirements 6.1**

**Property 25: Offline Read Access**
*For any* cached file, the system should serve the file for read operations while offline
**Validates: Requirements 6.4**

**Property 26: Offline Write Queuing**
*For any* write operation while offline, the system should allow the operation and queue changes for synchronization when connectivity is restored
**Validates: Requirements 6.5**

**Property 27: Batch Upload Processing**
*For any* network connectivity restoration, the system should process queued uploads in batches
**Validates: Requirements 6.10**
