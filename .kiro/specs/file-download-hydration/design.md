# Design: File Download Hydration

## Source
Moved from `.kiro/specs/archive/system-verification-and-fix/design.md`. Sections below are verbatim.

### 3. Basic On-Demand File Access Component

**Location**: `internal/fs/file_operations.go`, `internal/fs/dir_operations.go`

**Verification Steps**:
1. Review FUSE operation implementations
2. Test read operations (Open, Read, Release)
3. Test directory operations (OpenDir, ReadDir, ReleaseDir)
4. Test metadata operations (GetAttr, SetAttr)
5. Test ETag-based cache validation
6. Test redirect handling for downloads

**Expected Interfaces**:
- FUSE operation handlers (Open, Read, etc.)
- `RequestFileContent()` for uncached files
- `ValidateCache()` for ETag comparison
- `FollowRedirect()` for download URLs

**Verification Criteria**:
- Directory listings use metadata without downloading content
- Uncached files trigger download via GET `/items/{id}/content`
- 302 redirects are followed to pre-authenticated URLs
- Cached files validated using ETag comparison from delta sync
- Cache hits serve from local storage without network requests
- Cache misses invalidate and re-download content

### 3A. Download Status and Progress Tracking Component

**Location**: `internal/fs/file_status.go`, `internal/fs/download_manager.go`

**Verification Steps**:
1. Review file status tracking implementation
2. Test status updates during download
3. Test error status marking
4. Test status persistence and retrieval
5. Test status notification mechanisms

**Expected Interfaces**:
- `UpdateFileStatus()` for status changes
- `SetDownloadingStatus()` for active downloads
- `SetErrorStatus()` for failed downloads
- `GetFileStatus()` for status queries

**Verification Criteria**:
- File status updates to "downloading" during active downloads
- Failed downloads marked with error status and logged
- Status changes are persistent and queryable
- Status updates trigger appropriate notifications
- Status information available via extended attributes and D-Bus

### 3B. Download Manager Configuration Component

**Location**: `internal/fs/download_manager.go`, `internal/config/download_config.go`

**Verification Steps**:
1. Review download configuration implementation
2. Test worker pool size configuration
3. Test retry attempts configuration
4. Test queue size configuration
5. Test chunk size configuration
6. Test configuration validation

**Expected Interfaces**:
- `ConfigureWorkerPool()` for worker management
- `SetRetryAttempts()` for retry configuration
- `SetQueueSize()` for queue management
- `SetChunkSize()` for large file handling
- `ValidateConfig()` for parameter validation

**Verification Criteria**:
- Worker pool size configurable (1-10, default: 3)
- Retry attempts configurable (1-10, default: 3)
- Queue size configurable (100-5000, default: 500)
- Chunk size configurable (1MB-100MB, default: 10MB)
- Invalid configurations display clear error messages with valid ranges
- Configuration parameters validated on startup

### 3C. File Hydration State Management Component

**Location**: `internal/fs/state_manager.go`, `internal/fs/hydration.go`

**Verification Steps**:
1. Review item state management implementation
2. Test GHOST state handling
3. Test hydration state transitions
4. Test eviction state transitions
5. Test state persistence and recovery

**Expected Interfaces**:
- `GetItemState()` for state queries
- `TransitionToHydrating()` for download initiation
- `TransitionToHydrated()` for successful completion
- `TransitionToGhost()` for eviction
- `BlockUntilHydrated()` for access control

**Verification Criteria**:
- GHOST state items block access until hydration
- Hydration transitions to HYDRATED on success or ERROR on failure
- Evicted files transition back to GHOST without metadata loss
- State transitions are atomic and persistent
- Future FUSE requests can immediately rehydrate on demand

### 4. Download Manager Component

**Location**: `internal/fs/download_manager.go`

**Verification Steps**:
1. Review download queue implementation
2. Test concurrent downloads
3. Test download retry logic
4. Test download cancellation
5. Test cache integration

**Expected Interfaces**:
- `DownloadManager` struct with worker pool
- `QueueDownload()` method
- `CancelDownload()` method
- Integration with `LoopbackCache`

**Configuration Parameters**:
- **Worker Pool Size**: Number of concurrent download workers
  - Default: 3 workers
  - Valid Range: 1-10 workers
  - Configurable via: Command-line flag or configuration file
  - Purpose: Controls download concurrency and resource usage
  
- **Recovery Attempts Limit**: Maximum retry attempts for failed downloads
  - Default: 3 attempts
  - Valid Range: 1-10 attempts
  - Configurable via: Command-line flag or configuration file
  - Purpose: Balances reliability with avoiding infinite retries
  
- **Queue Size**: Buffer capacity for pending download requests
  - Default: 500 requests
  - Valid Range: 100-5000 requests
  - Configurable via: Command-line flag or configuration file
  - Purpose: Prevents memory exhaustion while allowing burst traffic
  
- **Chunk Size**: Size of chunks for large file downloads
  - Default: 10 MB (10485760 bytes)
  - Valid Range: 1 MB - 100 MB
  - Configurable via: Command-line flag or configuration file
  - Purpose: Balances memory usage with download efficiency and resume granularity

**Verification Criteria**:
- Files download on first access
- Multiple files download concurrently
- Failed downloads retry with backoff
- Downloaded content is cached correctly
- Download status is tracked and reported
- Configuration parameters are validated on startup
- Invalid configuration values display clear error messages with valid ranges

### File Access Properties

**Property 11: Metadata-Only Directory Listing**
*For any* directory listing operation, the system should display all files using cached metadata without downloading file content
**Validates: Requirements 3.1**

**Property 12: On-Demand Content Download**
*For any* uncached file access, the system should request file content using the correct API endpoint (GET /items/{id}/content)
**Validates: Requirements 3.2**

**Property 13: ETag Cache Validation**
*For any* cached file access, the system should validate the cache using ETag comparison from delta sync metadata
**Validates: Requirements 3.4**

**Property 14: Cache Hit Serving**
*For any* cached file with matching ETag, the system should serve content from local cache without network requests
**Validates: Requirements 3.5**

**Property 15: Cache Invalidation on ETag Mismatch**
*For any* cached file with different ETag, the system should invalidate the cache entry and download new content
**Validates: Requirements 3.6**
