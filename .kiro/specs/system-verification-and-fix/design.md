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

## Components and Interfaces

### 1. Authentication Component

**Location**: `pkg/graph/oauth2*.go`, `pkg/graph/authenticator.go`

**Verification Steps**:
1. Review OAuth2 implementation (both GTK and headless flows)
2. Test token storage and retrieval
3. Test token refresh mechanism
4. Test authentication failure scenarios
5. Verify secure token storage

**Expected Interfaces**:
- `Auth` struct with `AccessToken`, `RefreshToken`, `ExpiresAt`
- `Authenticate()` method for initial auth
- `Refresh()` method for token refresh
- Token storage in `~/.config/onemount/auth_tokens.json`

**Verification Criteria**:
- Authentication succeeds with valid credentials
- Tokens are stored securely with appropriate permissions
- Token refresh works automatically before expiration
- Failed authentication provides clear error messages
- Headless mode uses device code flow correctly

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

**Location**: `pkg/errors/`, logging throughout

**Verification Steps**:
1. Review error types and handling
2. Test network error scenarios
3. Test API rate limiting
4. Test crash recovery
5. Test error logging

**Expected Interfaces**:
- Custom error types in `pkg/errors`
- Structured logging with zerolog
- Error context propagation

**Verification Criteria**:
- Errors are logged with context
- Network errors trigger retries
- Rate limits trigger backoff
- Crashes don't corrupt state
- Error messages are user-friendly

## Data Models

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

## Testing Strategy

### Unit Test Verification

Review existing unit tests:
- Identify gaps in coverage
- Verify tests check actual behavior, not implementation details
- Add missing tests for edge cases
- Ensure tests are deterministic

### Integration Test Creation

Create new integration tests for:
- Authentication → Mounting → File Access flow
- File Modification → Upload → Delta Sync flow
- Online → Offline → Online transition flow
- Concurrent file operations
- Error recovery scenarios

### End-to-End Test Creation

Create end-to-end tests for:
- Complete user workflow from install to file access
- Multi-file operations (copy directory, etc.)
- Long-running operations (large file upload)
- Stress testing (many concurrent operations)

### Test Execution Strategy

1. **Run existing tests**: Identify which pass/fail
2. **Run manual verification**: Test actual application behavior
3. **Create reproduction cases**: For each failure, create minimal test
4. **Fix and verify**: Fix issue, verify with test
5. **Regression check**: Ensure fix doesn't break other components

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
