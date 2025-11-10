# Design Document: OneMount System Verification and Fix

## Overview

This design document outlines the systematic approach to verifying and fixing the OneMount application. The strategy involves:

1. **Component-by-component verification** - Test each major component against requirements
2. **Integration testing** - Verify components work together correctly
3. **Gap analysis** - Document discrepancies between docs and implementation
4. **Prioritized fixes** - Address critical issues first
5. **Documentation updates** - Ensure docs reflect actual behavior

## Architecture

### Verification Framework

The verification process follows a layered approach:

```
┌─────────────────────────────────────────┐
│     End-to-End User Workflows           │
├─────────────────────────────────────────┤
│     Integration Tests                   │
├─────────────────────────────────────────┤
│     Component Verification              │
├─────────────────────────────────────────┤
│     Code Analysis & Documentation       │
└─────────────────────────────────────────┘
```

### Verification Phases

#### Phase 1: Code Analysis
- Review existing code structure
- Compare implementation against architecture docs
- Identify missing or incomplete components
- Document deviations from design

#### Phase 2: Component Verification
- Test each component in isolation
- Verify against acceptance criteria
- Document failures and root causes
- Create component-specific fix plans

#### Phase 3: Integration Verification
- Test component interactions
- Verify data flow between components
- Test error propagation and handling
- Document integration issues

#### Phase 4: End-to-End Testing
- Test complete user workflows
- Verify against user stories
- Test edge cases and error scenarios
- Document user-facing issues

## Architecture Enhancements

### Multi-Account Architecture

The system supports multiple simultaneous OneDrive mounts through isolated filesystem instances:

```
┌─────────────────────────────────────────────────────────────┐
│                    OneMount Application                      │
├─────────────────────────────────────────────────────────────┤
│  Mount Manager                                               │
│  ├─ Personal OneDrive Mount (/mnt/onedrive-personal)       │
│  │  ├─ Filesystem Instance #1                              │
│  │  ├─ Auth (personal account)                             │
│  │  ├─ Cache (personal)                                    │
│  │  ├─ Delta Sync Loop #1                                  │
│  │  └─ Webhook Subscription #1                             │
│  │                                                          │
│  ├─ Work OneDrive Mount (/mnt/onedrive-work)              │
│  │  ├─ Filesystem Instance #2                              │
│  │  ├─ Auth (work account)                                 │
│  │  ├─ Cache (work)                                        │
│  │  ├─ Delta Sync Loop #2                                  │
│  │  └─ Webhook Subscription #2                             │
│  │                                                          │
│  └─ Shared Drive Mount (/mnt/onedrive-shared)             │
│     ├─ Filesystem Instance #3                              │
│     ├─ Auth (work account, shared drive access)            │
│     ├─ Cache (shared)                                       │
│     ├─ Delta Sync Loop #3                                  │
│     └─ Webhook Subscription #3                             │
└─────────────────────────────────────────────────────────────┘
```

Each mount is completely isolated with:
- Separate FUSE mount point
- Independent authentication tokens
- Isolated metadata and content caches
- Dedicated delta sync goroutine
- Individual webhook subscription

### Webhook Subscription Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                  Webhook Notification Flow                   │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  OneDrive API                                               │
│       │                                                      │
│       │ 1. POST /subscriptions                             │
│       │    (create subscription)                            │
│       ▼                                                      │
│  ┌──────────────┐                                          │
│  │ Subscription │                                          │
│  │   Manager    │                                          │
│  └──────┬───────┘                                          │
│         │                                                    │
│         │ 2. Store subscription ID & expiration            │
│         │                                                    │
│         ▼                                                    │
│  ┌──────────────┐         ┌──────────────┐               │
│  │  Webhook     │◄────────│   OneDrive   │               │
│  │  Listener    │ 3. POST │   Service    │               │
│  │  (HTTP)      │ notification            │               │
│  └──────┬───────┘         └──────────────┘               │
│         │                                                    │
│         │ 4. Validate notification                          │
│         │                                                    │
│         ▼                                                    │
│  ┌──────────────┐                                          │
│  │ Delta Sync   │                                          │
│  │   Trigger    │                                          │
│  └──────┬───────┘                                          │
│         │                                                    │
│         │ 5. Immediate delta query                          │
│         │                                                    │
│         ▼                                                    │
│  ┌──────────────┐                                          │
│  │  Metadata    │                                          │
│  │   Cache      │                                          │
│  └──────────────┘                                          │
│                                                              │
│  Background: Subscription Renewal Loop                      │
│  ┌──────────────┐                                          │
│  │   Monitor    │──► Check expiration every hour           │
│  │  Expiration  │──► Renew if < 24h remaining             │
│  └──────────────┘                                          │
└─────────────────────────────────────────────────────────────┘
```

### ETag-Based Cache Validation

```
┌─────────────────────────────────────────────────────────────┐
│              ETag Cache Validation Flow                      │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  File Access Request                                         │
│       │                                                      │
│       ▼                                                      │
│  ┌──────────────┐                                          │
│  │ Check Cache  │                                          │
│  │   Exists?    │                                          │
│  └──────┬───────┘                                          │
│         │                                                    │
│    Yes  │  No                                               │
│    ┌────┴────┐                                             │
│    │         │                                              │
│    ▼         ▼                                              │
│  ┌────┐  ┌────────┐                                        │
│  │Get │  │Download│                                        │
│  │ETag│  │  Full  │                                        │
│  └─┬──┘  │  File  │                                        │
│    │     └────────┘                                        │
│    │                                                        │
│    ▼                                                        │
│  GET /items/{id}/content                                   │
│  Header: if-none-match: "{etag}"                           │
│    │                                                        │
│    ├──► 304 Not Modified                                   │
│    │     └─► Serve from cache                              │
│    │                                                        │
│    └──► 200 OK (new content)                               │
│          ├─► Update cache                                   │
│          ├─► Store new ETag                                 │
│          └─► Serve new content                              │
│                                                              │
│  Delta Sync Updates ETags:                                  │
│  ┌──────────────┐                                          │
│  │ Delta Query  │──► Detects remote changes                │
│  └──────┬───────┘                                          │
│         │                                                    │
│         ▼                                                    │
│  ┌──────────────┐                                          │
│  │ Update ETag  │──► Invalidates cache entry               │
│  │  in Metadata │                                          │
│  └──────────────┘                                          │
└─────────────────────────────────────────────────────────────┘
```

## Components and Interfaces

### 1. Authentication Component

**Location**: `internal/graph/oauth2*.go`, `internal/graph/authenticator.go`

**Verification Steps**:
1. Review OAuth2 implementation (both GTK and headless flows)
2. Test token storage and retrieval in Docker container
3. Test token refresh mechanism
4. Test authentication failure scenarios
5. Verify secure token storage

**Expected Interfaces**:
- `Auth` struct with `AccessToken`, `RefreshToken`, `ExpiresAt`
- `Authenticator` interface with `Authenticate()`, `Refresh()`, `GetAuth()` methods
- `RealAuthenticator` and `MockAuthenticator` implementations
- Token storage in `~/.config/onemount/auth_tokens.json` or test-artifacts directory

**Verification Criteria**:
- Authentication succeeds with valid credentials
- Tokens are stored securely with appropriate permissions
- Token refresh works automatically before expiration
- Failed authentication provides clear error messages
- Headless mode uses device code flow correctly
- Tests run in isolated Docker containers without affecting host system

### 2. Filesystem Mounting Component

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

### 3. File Operations Component

**Location**: `internal/fs/file_operations.go`, `internal/fs/dir_operations.go`

**Verification Steps**:
1. Review FUSE operation implementations
2. Test read operations (Open, Read, Release)
3. Test write operations (Create, Write, Flush, Fsync)
4. Test directory operations (OpenDir, ReadDir, ReleaseDir)
5. Test metadata operations (GetAttr, SetAttr)
6. Test file/directory creation and deletion

**Expected Interfaces**:
- FUSE operation handlers (Open, Read, Write, etc.)
- Inode management (GetID, InsertNodeID, DeleteNodeID)
- Content caching (Get, Set from LoopbackCache)

**Verification Criteria**:
- Files can be read without errors
- File writes are queued for upload
- Directory listings show all files
- File metadata is accurate
- Operations return appropriate FUSE status codes

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

**Verification Criteria**:
- Files download on first access
- Multiple files download concurrently
- Failed downloads retry with backoff
- Downloaded content is cached correctly
- Download status is tracked and reported

### 5. Upload Manager Component

**Location**: `internal/fs/upload_manager.go`, `internal/fs/upload_session.go`

**Verification Steps**:
1. Review upload queue implementation
2. Test upload session creation
3. Test chunked uploads for large files
4. Test upload retry logic
5. Test conflict detection

**Expected Interfaces**:
- `UploadManager` struct with queue
- `UploadSession` for managing uploads
- `QueueUpload()` method
- `Upload()` method with retry logic

**Verification Criteria**:
- Modified files are queued for upload
- Uploads complete successfully
- Large files use chunked upload
- Failed uploads retry appropriately
- Upload conflicts are detected and handled

### 6. Delta Synchronization Component

**Location**: `internal/fs/delta.go`, `internal/fs/sync.go`

**Verification Steps**:
1. Review delta sync loop implementation
2. Test initial sync
3. Test incremental sync
4. Test conflict detection
5. Test delta link persistence

**Expected Interfaces**:
- `DeltaLoop()` goroutine
- `FetchDeltas()` method
- `ApplyChanges()` method
- Delta link storage in bbolt database

**Verification Criteria**:
- Initial sync fetches all metadata
- Incremental syncs fetch only changes
- Remote changes update local cache
- Conflicts create conflict copies
- Delta link persists across restarts

### 7. Cache Management Component

**Location**: `internal/fs/cache.go`, `internal/fs/content_cache.go`

**Verification Steps**:
1. Review cache implementation (metadata and content)
2. Test cache hit/miss scenarios
3. Test cache expiration
4. Test cache cleanup
5. Test cache statistics

**Expected Interfaces**:
- `LoopbackCache` for content
- `ThumbnailCache` for thumbnails
- bbolt database for metadata
- Cache cleanup goroutine

**Verification Criteria**:
- Cached files are served without network access
- Cache respects expiration settings
- Cleanup removes old files
- Statistics accurately reflect cache state
- Cache survives filesystem restarts

### 8. Offline Mode Component

**Location**: `internal/fs/offline.go`

**Verification Steps**:
1. Review offline detection logic
2. Test transition to offline mode
3. Test read-only enforcement
4. Test change queuing
5. Test transition back to online mode

**Expected Interfaces**:
- `IsOffline()` method
- `SetOffline()` method
- Offline change tracking in database

**Verification Criteria**:
- Network loss is detected automatically
- Filesystem becomes read-only when offline
- Cached files remain accessible
- Changes are queued for later upload
- Online transition processes queued changes

### 9. File Status and D-Bus Component

**Location**: `internal/fs/file_status.go`, `internal/fs/dbus.go`

**Verification Steps**:
1. Review file status tracking
2. Test D-Bus server initialization
3. Test status signal emission
4. Test extended attribute fallback
5. Test Nemo extension integration

**Expected Interfaces**:
- `FileStatusInfo` struct
- `FileStatusDBusServer` for D-Bus communication
- Extended attributes (user.onemount.status)
- Nemo extension Python script

**Verification Criteria**:
- File status updates correctly
- D-Bus signals are sent when available
- Extended attributes work as fallback
- Nemo extension displays status icons
- Status persists across filesystem restarts

### 10. Error Handling Component

**Location**: `internal/errors/`, `internal/logging/` throughout codebase

**Verification Steps**:
1. Review error types and handling
2. Test network error scenarios in Docker
3. Test API rate limiting
4. Test crash recovery
5. Test error logging

**Expected Interfaces**:
- Custom error types in `internal/errors`
- Structured logging with zerolog in `internal/logging`
- Error context propagation and monitoring
- Error wrapping with context

**Verification Criteria**:
- Errors are logged with context
- Network errors trigger retries
- Rate limits trigger backoff
- Crashes don't corrupt state
- Error messages are user-friendly
- All tests run in Docker containers

### 11. Webhook Subscription Component

**Location**: `internal/fs/subscription.go`, `internal/graph/` (subscription API calls)

**Verification Steps**:
1. Review subscription creation and management code
2. Test subscription creation on mount
3. Test webhook notification reception
4. Test subscription renewal before expiration
5. Test fallback to polling when subscriptions fail
6. Test subscription deletion on unmount

**Expected Interfaces**:
- `SubscriptionManager` struct with subscription state
- `CreateSubscription(resource, notificationUrl)` method
- `RenewSubscription(subscriptionId)` method
- `DeleteSubscription(subscriptionId)` method
- `HandleWebhookNotification(notification)` method
- HTTP server for receiving webhook notifications
- Subscription expiration monitoring goroutine

**Verification Criteria**:
- Subscriptions are created successfully on mount
- Webhook notifications trigger immediate delta queries
- Subscriptions are renewed before expiration (within 24h)
- System falls back to polling if subscription fails
- Subscriptions are deleted cleanly on unmount
- Personal OneDrive: can subscribe to any folder
- Business OneDrive: can only subscribe to root
- Polling interval is longer (30min) when subscription active
- Polling interval is shorter (5min) when no subscription

### 12. Multi-Account Mount Manager Component

**Location**: `cmd/onemount/main.go`, `internal/ui/onemount.go`

**Verification Steps**:
1. Review mount management code
2. Test mounting multiple accounts simultaneously
3. Test isolation between mounts (auth, cache, sync)
4. Test personal OneDrive mount
5. Test OneDrive for Business mount
6. Test shared drive mount
7. Test "Shared with me" access

**Expected Interfaces**:
- `MountManager` struct tracking active mounts
- `Mount(accountType, mountPoint, auth)` method
- `Unmount(mountPoint)` method
- `ListMounts()` method
- Separate `Filesystem` instance per mount
- Separate cache directory per mount
- Separate delta sync loop per mount

**Verification Criteria**:
- Multiple accounts can be mounted simultaneously
- Each mount has isolated authentication
- Each mount has isolated cache
- Each mount has independent delta sync
- Personal OneDrive accessible via `/me/drive`
- Business OneDrive accessible via `/me/drive`
- Shared drives accessible via `/drives/{drive-id}`
- "Shared with me" accessible via `/me/drive/sharedWithMe`
- No cross-contamination between mounts

### 13. ETag Cache Validation Component

**Location**: `internal/fs/cache.go`, `internal/fs/content_cache.go`

**Verification Steps**:
1. Review ETag storage and validation code
2. Test cache hit with valid ETag (304 Not Modified)
3. Test cache miss with changed ETag (200 OK)
4. Test ETag updates from delta sync
5. Test conflict detection using ETag comparison

**Expected Interfaces**:
- Cache entries store ETag alongside content
- `ValidateCache(itemId, etag)` method
- `UpdateETag(itemId, newEtag)` method
- HTTP requests include `if-none-match` header
- Delta sync updates ETags in metadata cache

**Verification Criteria**:
- ETags are stored with cached files
- Cache validation uses `if-none-match` header
- 304 response serves from cache
- 200 response updates cache and ETag
- Delta sync invalidates cache when ETag changes
- Conflict detection compares local and remote ETags
- Upload checks remote ETag before overwriting

## Data Models

### Webhook Subscription Data Model

```go
type Subscription struct {
    ID                 string    // Subscription ID from OneDrive API
    Resource           string    // Resource path (e.g., "/me/drive/root")
    ChangeType         string    // "updated"
    NotificationURL    string    // Public URL for webhook notifications
    ExpirationDateTime time.Time // When subscription expires (max 3 days)
    ClientState        string    // Optional validation token
    CreatedAt          time.Time // When subscription was created
    LastRenewed        time.Time // Last renewal timestamp
}

type WebhookNotification struct {
    SubscriptionID          string    // ID of the subscription
    SubscriptionExpiration  time.Time // Expiration time
    ChangeType              string    // Type of change
    Resource                string    // Resource that changed
    ClientState             string    // Validation token
    ResourceData            map[string]interface{} // Additional data
}

type SubscriptionManager struct {
    subscriptions      map[string]*Subscription // subscriptionID -> Subscription
    filesystem         *Filesystem              // Associated filesystem
    auth               *graph.Auth              // Authentication
    notificationServer *http.Server             // HTTP server for webhooks
    renewalTicker      *time.Ticker             // Periodic renewal check
    mutex              sync.RWMutex             // Thread safety
}
```

### Multi-Account Mount Data Model

```go
type MountConfig struct {
    AccountType    AccountType // Personal, Business, Shared
    MountPoint     string      // Filesystem mount path
    DriveID        string      // For shared drives
    AuthTokenPath  string      // Path to auth tokens
    CacheDir       string      // Cache directory for this mount
}

type AccountType int
const (
    PersonalOneDrive AccountType = iota
    BusinessOneDrive
    SharedDrive
    SharedWithMe
)

type MountManager struct {
    mounts map[string]*MountInstance // mountPoint -> MountInstance
    mutex  sync.RWMutex
}

type MountInstance struct {
    Config       MountConfig
    Filesystem   *Filesystem
    Auth         *graph.Auth
    DeltaLoop    *DeltaLoop
    Subscription *SubscriptionManager
    MountedAt    time.Time
}
```

### XDG Base Directory Structure

OneMount follows XDG Base Directory Specification:

```
Configuration Directory (XDG_CONFIG_HOME or ~/.config):
└── onemount/
    ├── config.yml                    # Main configuration file
    ├── auth_tokens.json              # Authentication tokens (personal)
    ├── auth_tokens_work.json         # Authentication tokens (work)
    └── auth_tokens_shared.json       # Authentication tokens (shared drives)

Cache Directory (XDG_CACHE_HOME or ~/.cache):
└── onemount/
    ├── personal/                     # Personal OneDrive cache
    │   ├── metadata.db               # BBolt metadata database
    │   ├── content/                  # File content cache
    │   │   ├── {item-id-1}
    │   │   └── {item-id-2}
    │   └── thumbnails/               # Thumbnail cache
    │       ├── {item-id-1}.jpg
    │       └── {item-id-2}.jpg
    │
    ├── work/                         # Work OneDrive cache
    │   ├── metadata.db
    │   ├── content/
    │   └── thumbnails/
    │
    └── shared-{drive-id}/            # Shared drive cache
        ├── metadata.db
        ├── content/
        └── thumbnails/

Runtime Data (XDG_RUNTIME_DIR or /tmp):
└── onemount/
    ├── mount-{pid}.sock              # Unix socket for IPC
    └── dbus-{pid}.sock               # D-Bus socket
```

**Implementation**:
- Use `os.UserConfigDir()` which respects `XDG_CONFIG_HOME`
- Use `os.UserCacheDir()` which respects `XDG_CACHE_HOME`
- Allow override via command-line flags (`--config-file`, `--cache-dir`)
- Create directories with appropriate permissions (0700 for config, 0755 for cache)
- Separate cache directories per mount to avoid conflicts

### ETag Cache Entry Data Model

```go
type CacheEntry struct {
    ItemID           string    // OneDrive item ID
    Content          []byte    // File content
    ETag             string    // ETag from OneDrive
    CTag             string    // CTag from OneDrive (optional)
    Size             int64     // File size
    LastModified     time.Time // Last modified time
    CachedAt         time.Time // When cached
    LastAccessed     time.Time // Last access time
    DownloadURL      string    // Pre-authenticated download URL (temporary)
    DownloadURLExpiry time.Time // When download URL expires
}

type MetadataCache struct {
    items map[string]*CachedMetadata // itemID -> metadata
    mutex sync.RWMutex
}

type CachedMetadata struct {
    DriveItem        *graph.DriveItem // Full metadata
    ETag             string           // Current ETag
    CachedAt         time.Time        // When metadata was cached
    LastValidated    time.Time        // Last ETag validation
    HasLocalChanges  bool             // Pending upload
    LocalETag        string           // ETag when file was downloaded
}
```

### Verification Data Model

For each component, we track:

```go
type ComponentVerification struct {
    ComponentName    string
    RequirementIDs   []string
    Status           VerificationStatus
    TestResults      []TestResult
    Issues           []Issue
    FixPriority      Priority
}

type VerificationStatus string
const (
    NotStarted VerificationStatus = "not_started"
    InProgress VerificationStatus = "in_progress"
    Passed     VerificationStatus = "passed"
    Failed     VerificationStatus = "failed"
)

type TestResult struct {
    TestName        string
    Passed          bool
    ErrorMessage    string
    ExpectedBehavior string
    ActualBehavior   string
}

type Issue struct {
    IssueID         string
    Description     string
    RootCause       string
    AffectedFiles   []string
    Severity        Severity
    FixEstimate     string
}

type Priority string
const (
    Critical Priority = "critical"  // Blocks core functionality
    High     Priority = "high"      // Major feature broken
    Medium   Priority = "medium"    // Minor feature broken
    Low      Priority = "low"       // Enhancement or optimization
)
```

### Gap Analysis Data Model

```go
type DocumentationGap struct {
    DocumentPath     string
    Section          string
    DocumentedBehavior string
    ActualBehavior     string
    GapType            GapType
    UpdateRequired     bool
}

type GapType string
const (
    Missing      GapType = "missing"       // Feature documented but not implemented
    Extra        GapType = "extra"         // Feature implemented but not documented
    Inconsistent GapType = "inconsistent"  // Implementation differs from docs
    Outdated     GapType = "outdated"      // Docs describe old implementation
)
```

## Error Handling

### Verification Error Handling

During verification, we handle errors as follows:

1. **Test Failures**: Log failure details, continue with other tests
2. **Component Crashes**: Capture stack trace, mark component as failed
3. **Integration Failures**: Identify which components are involved
4. **Timeout Errors**: Mark as inconclusive, retry with longer timeout

### Fix Error Handling

During fixes, we handle errors as follows:

1. **Compilation Errors**: Fix immediately before proceeding
2. **Test Failures**: Analyze root cause, update fix approach
3. **Regression**: Revert change, analyze why it broke other components
4. **Documentation Errors**: Update docs to match working implementation

## OneDrive API Endpoint Mapping

This section maps design components to specific OneDrive API endpoints based on Microsoft Graph API documentation.

### Authentication Endpoints
- **OAuth2 Authorization**: `https://login.microsoftonline.com/{tenant}/oauth2/v2.0/authorize`
- **Token Exchange**: `https://login.microsoftonline.com/{tenant}/oauth2/v2.0/token`
- **Token Refresh**: `https://login.microsoftonline.com/{tenant}/oauth2/v2.0/token` (with refresh_token)

### Drive Access Endpoints
- **Personal OneDrive**: `GET /me/drive`
- **Business OneDrive**: `GET /me/drive` (same endpoint, different tenant)
- **Specific Drive**: `GET /drives/{drive-id}`
- **Shared With Me**: `GET /me/drive/sharedWithMe`
- **Drive Root**: `GET /me/drive/root`

### File Operations Endpoints
- **Get Metadata**: `GET /drives/{drive-id}/items/{item-id}`
- **Get Content**: `GET /drives/{drive-id}/items/{item-id}/content`
  - Returns: `302 Found` with `Location` header to download URL
  - Supports: `if-none-match` header for ETag validation
  - Response: `304 Not Modified` if ETag matches, `200 OK` with content if changed
- **Upload Small File (<250MB)**: `PUT /drives/{drive-id}/items/{item-id}/content`
  - Request Body: Binary file content
  - Response: DriveItem with new ETag
- **Create Upload Session (≥250MB)**: `POST /drives/{drive-id}/items/{parent-id}:/{filename}:/createUploadSession`
  - Response: `uploadUrl` and `expirationDateTime`
- **Upload Chunk**: `PUT {uploadUrl}` (from upload session)
  - Headers: `Content-Range: bytes {start}-{end}/{total}`

### Delta Sync Endpoints
- **Initial Delta**: `GET /drives/{drive-id}/root/delta`
  - Returns: Collection of DriveItems + `@odata.nextLink` or `@odata.deltaLink`
- **Subsequent Delta**: `GET /drives/{drive-id}/root/delta?token={deltaToken}`
  - Uses token from previous `@odata.deltaLink`
  - Returns: Only changed items since last query
- **Delta with Sharing Changes**: `GET /drives/{drive-id}/root/delta`
  - Header: `Prefer: deltashowsharingchanges`
  - Response includes: `@microsoft.graph.sharedChanged: "True"` annotation

### Webhook Subscription Endpoints
- **Create Subscription**: `POST /subscriptions`
  ```json
  {
    "changeType": "updated",
    "notificationUrl": "https://your-server.com/webhook",
    "resource": "/me/drive/root",
    "expirationDateTime": "2024-12-31T10:00:00Z"
  }
  ```
  - Personal OneDrive: Can subscribe to any folder
  - Business OneDrive: Can only subscribe to root folder
  - Max expiration: 3 days (4230 minutes)
  
- **Renew Subscription**: `PATCH /subscriptions/{subscription-id}`
  ```json
  {
    "expirationDateTime": "2024-12-31T10:00:00Z"
  }
  ```
  
- **Delete Subscription**: `DELETE /subscriptions/{subscription-id}`

- **Webhook Notification Format**:
  ```json
  {
    "value": [{
      "subscriptionId": "subscription-guid",
      "subscriptionExpirationDateTime": "2024-12-31T10:00:00Z",
      "changeType": "updated",
      "resource": "/me/drive/root",
      "clientState": "validation-token"
    }]
  }
  ```

### Directory Listing Endpoints
- **List Children**: `GET /drives/{drive-id}/items/{item-id}/children`
  - Returns: Collection of DriveItems (files and folders)
- **Get Item by Path**: `GET /drives/{drive-id}/root:/{path}`

### Metadata Properties
All DriveItem responses include:
- `id`: Unique identifier
- `name`: File/folder name
- `size`: Size in bytes
- `eTag`: Entity tag for cache validation
- `cTag`: Change tag (optional)
- `lastModifiedDateTime`: Last modification timestamp
- `createdDateTime`: Creation timestamp
- `file`: Present if item is a file (with `mimeType`)
- `folder`: Present if item is a folder (with `childCount`)
- `deleted`: Present if item is deleted
- `parentReference`: Parent folder information
- `@microsoft.graph.downloadUrl`: Pre-authenticated download URL (temporary, ~1 hour)

### API Limitations and Constraints
- **File Size Limits**:
  - Simple upload (PUT): Max 250 MB
  - Upload session: Max 250 GB (personal), 15 GB (business)
  
- **Subscription Limits**:
  - Max expiration: 4230 minutes (3 days)
  - Personal OneDrive: Subscribe to any folder
  - Business OneDrive: Subscribe to root only
  - Requires publicly accessible notification URL
  
- **Rate Limiting**:
  - Implement exponential backoff on 429 responses
  - Respect `Retry-After` header
  
- **Delta Query**:
  - Delta tokens expire after 30 days of inactivity
  - Use `token=latest` to get current state without enumeration

## Testing Strategy

### Docker Test Environment

All tests run in isolated Docker containers to avoid affecting the host system:

**Test Runner Container** (`onemount-test-runner`):
- Based on `onemount-base` image with Go 1.23+
- Includes all dependencies: FUSE3, GTK3, Python, build tools
- Pre-built OneMount binaries for faster test execution
- Mounts workspace as volume for source code access
- Writes test artifacts to `test-artifacts/` directory
- Configured with FUSE device and SYS_ADMIN capability

**Test Types**:
1. **Unit Tests**: Lightweight, no FUSE required, run with `docker compose run unit-tests`
2. **Integration Tests**: Require FUSE, run with `docker compose run integration-tests`
3. **System Tests**: Full end-to-end, require auth tokens, run with `docker compose run system-tests`
4. **Coverage Analysis**: Generate coverage reports, run with `docker compose run coverage`

**Docker Compose Services**:
- `test-runner`: Base service with common configuration
- `unit-tests`: Extends test-runner for unit tests
- `integration-tests`: Extends test-runner with FUSE support
- `system-tests`: Extends test-runner with auth token mounting
- `coverage`: Extends test-runner for coverage analysis
- `shell`: Interactive shell for debugging

**Environment Variables**:
- `ONEMOUNT_TEST_TIMEOUT`: Test timeout duration (default: 5m)
- `ONEMOUNT_TEST_VERBOSE`: Enable verbose test output
- `GORACE`: Race detector configuration
- `DOCKER_CONTAINER`: Flag indicating tests run in Docker

### Unit Test Verification

Review existing unit tests in Docker:
- Identify gaps in coverage
- Verify tests check actual behavior, not implementation details
- Add missing tests for edge cases
- Ensure tests are deterministic
- Run with: `docker compose -f docker/compose/docker-compose.test.yml run unit-tests`

### Integration Test Creation

Create new integration tests for:
- Authentication → Mounting → File Access flow
- File Modification → Upload → Delta Sync flow
- Online → Offline → Online transition flow
- Concurrent file operations
- Error recovery scenarios
- Run with: `docker compose -f docker/compose/docker-compose.test.yml run integration-tests`

### End-to-End Test Creation

Create end-to-end tests for:
- Complete user workflow from install to file access
- Multi-file operations (copy directory, etc.)
- Long-running operations (large file upload)
- Stress testing (many concurrent operations)
- Run with: `docker compose -f docker/compose/docker-compose.test.yml run system-tests`

### Test Execution Strategy

1. **Build test images**: `docker compose -f docker/compose/docker-compose.build.yml build`
2. **Run unit tests**: Fast feedback, no external dependencies
3. **Run integration tests**: Verify component interactions
4. **Run system tests**: Full end-to-end with real OneDrive (requires auth)
5. **Generate coverage**: Analyze test coverage
6. **Review artifacts**: Check `test-artifacts/` for logs and results
7. **Debug in container**: Use `docker compose run shell` for interactive debugging

## Implementation Plan Overview

The implementation will follow this sequence:

1. **Setup verification environment**
   - Create test OneDrive account
   - Set up test mount point
   - Configure logging for detailed output

2. **Component verification (in dependency order)**
   - Authentication (foundational)
   - Filesystem mounting (depends on auth)
   - File operations (depends on mounting)
   - Download/Upload managers (depends on file ops)
   - Delta sync (depends on download/upload)
   - Cache management (depends on all above)
   - Offline mode (depends on cache)
   - File status/D-Bus (depends on file ops)
   - Error handling (cross-cutting)

3. **Integration verification**
   - Test component interactions
   - Verify data flow
   - Test error propagation

4. **End-to-end verification**
   - Test complete workflows
   - Test edge cases
   - Performance testing

5. **Documentation updates**
   - Update architecture docs
   - Update design docs
   - Update API docs
   - Create troubleshooting guide

## Success Criteria

The verification and fix process is complete when:

1. All acceptance criteria from requirements are met
2. All integration tests pass
3. End-to-end workflows work correctly
4. Documentation accurately reflects implementation
5. No critical or high-priority issues remain
6. Performance meets documented requirements
7. Error handling is robust and user-friendly

## Risks and Mitigation

### Risk: Breaking working components while fixing others
**Mitigation**: 
- Create comprehensive test suite before making changes
- Use feature branches for each fix
- Run full test suite after each change

### Risk: Documentation updates lag behind code changes
**Mitigation**:
- Update docs as part of each fix task
- Review docs before marking task complete
- Use traceability matrix to track doc updates

### Risk: Fixes introduce new bugs
**Mitigation**:
- Thorough code review for each change
- Integration testing after each fix
- Regression testing before moving to next component

### Risk: Time estimates are inaccurate
**Mitigation**:
- Start with smallest, most isolated components
- Adjust estimates based on actual progress
- Prioritize critical issues if time is limited
