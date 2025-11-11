# Phase 12: Offline Mode - Issues and Fix Plan

**Date**: 2025-11-11  
**Component**: Offline Mode  
**Status**: Analysis Complete  
**Requirements**: 6.1, 6.2, 6.3, 6.4, 6.5

## Executive Summary

The offline mode implementation in OneMount is **functionally complete** with comprehensive change tracking and automatic online/offline transitions. However, there is a **significant design discrepancy** between the requirements (which specify read-only offline mode) and the implementation (which allows read-write operations with change queuing).

**Overall Assessment**: ⚠️ **Functional but Non-Compliant**
- ✅ Core offline functionality works correctly
- ✅ Change tracking and queuing implemented
- ✅ Automatic offline detection and recovery
- ⚠️ Does not enforce read-only mode as specified in requirements
- ⚠️ Passive offline detection (via delta sync failures)

## Identified Issues

### Issue #OF-001: Read-Write vs Read-Only Offline Mode

**Severity**: ⚠️ **Medium** (Design Discrepancy)  
**Status**: Open  
**Discovered**: 2025-11-11  
**Component**: Offline Mode State Management  
**Requirements Affected**: 6.3

**Description**:
Requirement 6.3 states: "WHILE offline, THE OneMount System SHALL make the filesystem read-only". However, the current implementation allows full read-write operations while offline, with changes being queued for later upload.

**Current Behavior**:
```go
// In file_operations.go
if f.IsOffline() {
    ctx.Info().Msg("File creation in offline mode will be cached locally")
    // Operation proceeds normally
}
```

**Expected Behavior** (per requirements):
```go
if f.IsOffline() {
    return syscall.EROFS // Read-only filesystem error
}
```

**Impact**:
- **Positive**: Users can continue working offline without interruption
- **Positive**: Changes are preserved and uploaded when back online
- **Negative**: Does not comply with stated requirements
- **Negative**: May confuse users who expect read-only behavior
- **Negative**: Risk of conflicts if same file modified offline and remotely

**Root Cause**:
The implementation was designed to provide a better user experience by allowing offline work with change queuing, rather than strictly enforcing read-only mode. This is a deliberate design decision that differs from the requirements.

**Affected Files**:
- `internal/fs/file_operations.go` (Create, Write, Delete operations)
- `internal/fs/dir_operations.go` (Mkdir operation)
- `internal/fs/offline.go` (OfflineMode enum)
- `internal/fs/cache.go` (TrackOfflineChange, ProcessOfflineChanges)

**Evidence**:
```go
// From file_operations.go:37-40
if f.IsOffline() {
    ctx.Info().Msg("File creation in offline mode will be cached locally")
}
// Operation continues...

// From file_operations.go:493-496
if f.IsOffline() {
    logger.Info().Msg("Write operations in offline mode will be cached locally")
}
// Write proceeds...
```

**Test Coverage**:
- ✅ `TestIT_OF_02_01`: Verifies write operations work offline
- ✅ `TestIT_OF_03_01`: Verifies changes are cached offline
- ❌ No test verifies read-only enforcement (because it's not enforced)

---

### Issue #OF-002: Passive Offline Detection

**Severity**: ℹ️ **Low** (Informational)  
**Status**: Open  
**Discovered**: 2025-11-11  
**Component**: Offline Detection  
**Requirements Affected**: 6.1

**Description**:
Requirement 6.1 states: "WHEN network connectivity is lost, THE OneMount System SHALL detect the offline state". The current implementation detects offline state passively through delta sync failures rather than actively monitoring network interfaces.

**Current Behavior**:
```go
// In delta.go:118-121
if err != nil {
    logging.Error().Err(err).Msg("Error during delta fetch, marking fs as offline.")
    f.Lock()
    f.offline = true
    f.Unlock()
}
```

**Expected Behavior** (strict interpretation):
- Active monitoring of network interfaces
- Immediate detection when network is lost
- Proactive offline state transition

**Impact**:
- **Positive**: Simple, reliable implementation
- **Positive**: No additional dependencies or complexity
- **Positive**: Works correctly in practice
- **Negative**: Detection delayed until next delta sync (up to 5 minutes)
- **Negative**: No immediate notification to user

**Root Cause**:
The implementation uses a pragmatic approach where offline state is inferred from API failures rather than directly monitoring network state. This is simpler and more reliable than trying to monitor network interfaces.

**Affected Files**:
- `internal/fs/delta.go` (Delta sync loop)
- `internal/graph/graph.go` (IsOffline error detection)

**Evidence**:
```go
// From graph.go:596-647
func IsOffline(err error) bool {
    // Checks error patterns like:
    // "no such host", "network is unreachable", "connection refused"
    // "connection timed out", "dial tcp", etc.
}
```

**Test Coverage**:
- ✅ Existing tests use `graph.SetOperationalOffline(true)` to simulate offline
- ❌ No test verifies actual network disconnection detection
- ❌ No test measures detection latency

---

### Issue #OF-003: No Explicit Cache Invalidation on Offline Transition

**Severity**: ℹ️ **Low** (Enhancement)  
**Status**: Open  
**Discovered**: 2025-11-11  
**Component**: Cache Management  
**Requirements Affected**: 6.2

**Description**:
When transitioning to offline mode, there is no explicit cache validation or cleanup. The system relies on existing cached content without verifying its freshness.

**Current Behavior**:
```go
// In file_operations.go:189-191
if f.IsOffline() {
    logger.Info().Msg("Using cached content in offline mode regardless of checksum")
    // Uses cache without validation
}
```

**Potential Enhancement**:
- Log which files are available offline
- Warn about potentially stale cached content
- Provide cache status information to user

**Impact**:
- **Positive**: Simple, fast offline access
- **Negative**: No visibility into cache freshness
- **Negative**: User doesn't know which files are available offline

**Affected Files**:
- `internal/fs/file_operations.go`
- `internal/fs/cache.go`

---

### Issue #OF-004: No User Notification of Offline State

**Severity**: ℹ️ **Low** (Enhancement)  
**Status**: Open  
**Discovered**: 2025-11-11  
**Component**: User Interface  
**Requirements Affected**: 6.1

**Description**:
When the filesystem transitions to offline mode, there is no user-visible notification. Users must check logs or file status to know they're offline.

**Current Behavior**:
- Offline state logged: `logging.Info().Msg("Offline mode enabled")`
- No D-Bus notification
- No desktop notification
- No file manager indication

**Potential Enhancement**:
- Send D-Bus signal when offline state changes
- Desktop notification: "OneDrive is now offline"
- Update file status indicators
- Add offline indicator to mount point

**Impact**:
- **Negative**: Users may not realize they're offline
- **Negative**: Confusion about why files aren't syncing
- **Positive**: Silent operation doesn't interrupt workflow

**Affected Files**:
- `internal/fs/offline.go`
- `internal/fs/dbus.go`

---

## Fix Plan

### Priority 1: Resolve Read-Only vs Read-Write Discrepancy

**Issue**: #OF-001  
**Effort**: Medium (2-4 hours)  
**Risk**: Medium (behavior change)

**Options**:

#### Option A: Update Requirements (Recommended)
**Rationale**: The current implementation provides better UX and is already working correctly.

**Changes Required**:
1. Update Requirement 6.3 to specify read-write offline mode with change queuing
2. Update design documentation to reflect current behavior
3. Add explicit requirement for change queuing (currently 6.4)
4. Document conflict resolution strategy for offline changes

**Pros**:
- ✅ No code changes required
- ✅ Preserves existing functionality
- ✅ Better user experience
- ✅ Matches user expectations from other sync tools

**Cons**:
- ⚠️ Requires requirements approval
- ⚠️ May need stakeholder buy-in

**Implementation**:
```markdown
# Updated Requirement 6.3
WHILE offline, THE OneMount System SHALL allow read and write operations
with changes queued for synchronization when connectivity is restored.

# New Requirement 6.3.1
WHEN a file is modified offline, THE OneMount System SHALL track the change
in persistent storage for later upload.

# New Requirement 6.3.2
WHEN multiple changes are made to the same file offline, THE OneMount System
SHALL preserve the most recent version for upload.
```

#### Option B: Enforce Read-Only Mode
**Rationale**: Strict compliance with current requirements.

**Changes Required**:
1. Modify file operation handlers to return `syscall.EROFS` when offline
2. Remove offline change tracking (or keep for future use)
3. Update tests to expect read-only behavior
4. Add user-facing error messages

**Pros**:
- ✅ Complies with current requirements
- ✅ Simpler behavior (no conflict resolution needed)
- ✅ Matches traditional offline filesystem behavior

**Cons**:
- ❌ Worse user experience
- ❌ Users lose work if they try to edit offline
- ❌ Breaks existing tests
- ❌ Removes useful functionality

**Implementation**:
```go
// In file_operations.go
func (f *Filesystem) Create(cancel <-chan struct{}, in *fuse.CreateIn, name string, out *fuse.CreateOut) fuse.Status {
    if f.IsOffline() {
        return fuse.Status(syscall.EROFS) // Read-only filesystem
    }
    // ... rest of implementation
}
```

#### Option C: Configurable Offline Mode
**Rationale**: Provide flexibility for different use cases.

**Changes Required**:
1. Add configuration option: `offline_mode: read-only | read-write`
2. Implement both behaviors based on configuration
3. Update documentation for both modes
4. Add tests for both modes

**Pros**:
- ✅ Satisfies both requirements and UX needs
- ✅ Flexible for different user preferences
- ✅ Maintains existing functionality

**Cons**:
- ⚠️ More complex implementation
- ⚠️ More testing required
- ⚠️ Configuration complexity

**Recommendation**: **Option A** - Update requirements to match implementation. The current behavior is superior and already working correctly.

---

### Priority 2: Improve Offline Detection Visibility

**Issue**: #OF-002, #OF-004  
**Effort**: Low (1-2 hours)  
**Risk**: Low

**Changes Required**:
1. Add D-Bus signal for offline state changes
2. Update file status when offline
3. Add desktop notification (optional)
4. Log offline detection more prominently

**Implementation**:
```go
// In offline.go
func (f *Filesystem) SetOfflineMode(mode OfflineMode) {
    f.Lock()
    defer f.Unlock()
    
    oldMode := f.offline
    
    switch mode {
    case OfflineModeDisabled:
        f.offline = false
        logging.Info().Msg("Offline mode disabled")
        if oldMode != f.offline && f.dbusServer != nil {
            f.dbusServer.SendOfflineStateUpdate(false)
        }
    case OfflineModeReadWrite:
        f.offline = true
        logging.Info().Msg("Offline mode enabled")
        if oldMode != f.offline && f.dbusServer != nil {
            f.dbusServer.SendOfflineStateUpdate(true)
        }
    }
}
```

**Benefits**:
- Users know when they're offline
- File managers can show offline indicator
- Better debugging and troubleshooting

---

### Priority 3: Add Cache Status Information

**Issue**: #OF-003  
**Effort**: Low (1-2 hours)  
**Risk**: Low

**Changes Required**:
1. Add method to query available offline files
2. Include cache status in GetStats()
3. Log cache availability when going offline

**Implementation**:
```go
// In cache.go
func (f *Filesystem) GetOfflineCacheStatus() *OfflineCacheStatus {
    return &OfflineCacheStatus{
        TotalCachedFiles: f.content.GetCachedFileCount(),
        TotalCachedSize:  f.content.GetCachedSize(),
        CachedPaths:      f.content.GetCachedPaths(),
        LastCacheUpdate:  f.content.GetLastUpdateTime(),
    }
}
```

**Benefits**:
- Users know which files are available offline
- Better planning for offline work
- Improved troubleshooting

---

## Testing Requirements

### New Tests Needed

1. **Test Offline Mode Behavior** (based on chosen option):
   - If Option A: Verify write operations work offline
   - If Option B: Verify write operations are rejected offline
   - If Option C: Verify both modes work correctly

2. **Test Offline Detection Latency**:
   - Measure time from network loss to offline detection
   - Verify detection happens within one delta sync cycle

3. **Test D-Bus Notifications**:
   - Verify offline state change signals sent
   - Verify file status updates when offline

4. **Test Cache Status Queries**:
   - Verify GetOfflineCacheStatus() returns correct data
   - Verify cache statistics include offline information

### Test Execution

```bash
# Run all offline tests
docker compose -f docker/compose/docker-compose.test.yml run --rm \
  --entrypoint /bin/bash shell -c "
    go test -v -run 'TestIT_OF' ./internal/fs/ -timeout 20m
  "

# Run with race detector
docker compose -f docker/compose/docker-compose.test.yml run --rm \
  --entrypoint /bin/bash shell -c "
    go test -v -race -run 'TestIT_OF' ./internal/fs/ -timeout 20m
  "
```

---

## Documentation Updates Required

### Requirements Document
- Update Requirement 6.3 (if Option A chosen)
- Add new requirements for change queuing behavior
- Clarify offline detection mechanism

### Design Document
- Document read-write offline mode design
- Document change tracking and queuing
- Document conflict resolution strategy
- Add sequence diagrams for offline transitions

### User Documentation
- Explain offline mode behavior
- Document which operations work offline
- Explain change synchronization
- Troubleshooting offline issues

### API Documentation
- Document `SetOfflineMode()` and `GetOfflineMode()`
- Document `TrackOfflineChange()` and `ProcessOfflineChanges()`
- Document `IsOffline()` usage patterns

---

## Implementation Timeline

### Phase 1: Decision and Planning (1 day)
- [ ] Review options with stakeholders
- [ ] Choose Option A, B, or C
- [ ] Get approval for requirements changes (if Option A)
- [ ] Finalize implementation plan

### Phase 2: Core Changes (2-3 days)
- [ ] Implement chosen option
- [ ] Add D-Bus notifications
- [ ] Add cache status queries
- [ ] Update logging

### Phase 3: Testing (2-3 days)
- [ ] Create new integration tests
- [ ] Run all offline tests in Docker
- [ ] Verify behavior matches requirements
- [ ] Performance testing

### Phase 4: Documentation (1-2 days)
- [ ] Update requirements (if needed)
- [ ] Update design documentation
- [ ] Update user documentation
- [ ] Update API documentation

### Phase 5: Review and Deployment (1 day)
- [ ] Code review
- [ ] Documentation review
- [ ] Final testing
- [ ] Merge and deploy

**Total Estimated Time**: 7-10 days

---

## Risk Assessment

### Low Risk
- ✅ Core offline functionality already works
- ✅ Comprehensive test coverage exists
- ✅ Change tracking is implemented and tested

### Medium Risk
- ⚠️ Requirements change may need approval (Option A)
- ⚠️ Behavior change may affect users (Option B)
- ⚠️ Additional complexity (Option C)

### Mitigation Strategies
- Get stakeholder approval early
- Communicate changes clearly to users
- Provide migration guide if behavior changes
- Thorough testing before deployment

---

## Recommendations

### Immediate Actions
1. ✅ **Choose Option A** - Update requirements to match implementation
2. ✅ Add D-Bus notifications for offline state changes
3. ✅ Improve logging and user visibility
4. ✅ Document current behavior thoroughly

### Future Enhancements
- Add active network monitoring (optional)
- Implement configurable offline mode (Option C)
- Add offline file browser/cache manager
- Improve conflict resolution UI
- Add offline work indicators in file manager

### No Action Required
- ✅ Change tracking and queuing (already implemented)
- ✅ Automatic offline detection (works correctly)
- ✅ Online transition and sync (works correctly)
- ✅ Integration tests (comprehensive coverage)

---

## Conclusion

The offline mode implementation in OneMount is **well-designed and functional**. The primary issue is a discrepancy between requirements (read-only) and implementation (read-write with queuing). 

**Recommended Resolution**: Update requirements to match the superior implementation rather than degrading functionality to match outdated requirements.

**Overall Status**: ⚠️ **Functional but requires requirements alignment**

**Next Steps**:
1. Get approval to update Requirement 6.3
2. Add D-Bus notifications for better user visibility
3. Update documentation to reflect actual behavior
4. Create additional integration tests for edge cases
5. Consider adding configurable offline mode in future release

---

## References

- Requirements: `.kiro/specs/system-verification-and-fix/requirements.md` (Section 6)
- Design: `.kiro/specs/system-verification-and-fix/design.md` (Offline Mode Component)
- Implementation: `internal/fs/offline.go`, `internal/fs/cache.go`
- Tests: `internal/fs/offline_integration_test.go`
- Test Plan: `docs/verification-phase12-offline-mode-test-plan.md`
- Verification Tracking: `docs/verification-tracking.md` (Phase 11)
