# State Model Reference Guide

## Overview

The OneMount metadata state model tracks the lifecycle of every filesystem item through 7 distinct states. This guide provides a quick reference for developers working with the state model.

## State Definitions

| State | Constant | Description | Typical Duration |
|-------|----------|-------------|------------------|
| GHOST | `ItemStateGhost` | Cloud metadata known, no local content | Until user accesses file |
| HYDRATING | `ItemStateHydrating` | Content download in progress | Seconds to minutes |
| HYDRATED | `ItemStateHydrated` | Local content matches remote ETag | Until modification or eviction |
| DIRTY_LOCAL | `ItemStateDirtyLocal` | Local changes pending upload | Until upload completes |
| DELETED_LOCAL | `ItemStateDeleted` | Local delete queued for upload | Until server confirms |
| CONFLICT | `ItemStateConflict` | Local + remote diverged | Until user resolves |
| ERROR | `ItemStateError` | Last operation failed | Until retry or cleanup |

## State Transition Map

```
GHOST
├─→ HYDRATING (user access)
├─→ HYDRATED (direct hydration, rare)
├─→ DELETED_LOCAL (delete before hydration)
└─→ DIRTY_LOCAL (modify before hydration)

HYDRATING
├─→ HYDRATED (download success)
└─→ ERROR (download failure)

HYDRATED
├─→ HYDRATING (re-hydration)
├─→ ERROR (operation failure)
├─→ DIRTY_LOCAL (local modification)
├─→ GHOST (cache eviction)
└─→ DELETED_LOCAL (local deletion)

DIRTY_LOCAL
├─→ HYDRATED (upload success)
├─→ CONFLICT (remote changes detected)
├─→ ERROR (upload failure)
└─→ DELETED_LOCAL (delete before upload)

CONFLICT
├─→ HYDRATED (conflict resolved)
└─→ DIRTY_LOCAL (user chooses local version)

ERROR
├─→ HYDRATING (retry download)
└─→ DELETED_LOCAL (delete failed item)

DELETED_LOCAL
└─→ [REMOVED] (terminal state)
```

## Usage Examples

### Transitioning States

```go
// Basic transition
fs.transitionItemState(id, metadata.ItemStateHydrating)

// Transition with context
fs.transitionItemState(id, metadata.ItemStateHydrated,
    metadata.WithHydrationEvent(),
    metadata.WithWorker("download:"+id),
    metadata.WithContentHash(hash),
    metadata.WithETag(etag),
    metadata.ClearPendingRemote())

// Transition with error
fs.transitionItemState(id, metadata.ItemStateError,
    metadata.WithHydrationEvent(),
    metadata.WithTransitionError(err, temporary))
```

### Checking Current State

```go
entry, err := fs.metadataStore.Get(ctx, id)
if err != nil {
    return err
}

switch entry.State {
case metadata.ItemStateGhost:
    // Need to hydrate
case metadata.ItemStateHydrated:
    // Can serve from cache
case metadata.ItemStateDirtyLocal:
    // Need to upload
case metadata.ItemStateError:
    // Handle error
}
```

### Querying by State

```go
// Get all items in a specific state
entries, err := fs.metadataStore.List(ctx, func(entry *metadata.Entry) bool {
    return entry.State == metadata.ItemStateGhost
})
```

## Transition Options

| Option | Purpose | Example |
|--------|---------|---------|
| `WithWorker(id)` | Track which worker performed transition | `WithWorker("download:"+id)` |
| `WithHydrationEvent()` | Mark as hydration-related | `WithHydrationEvent()` |
| `WithUploadEvent()` | Mark as upload-related | `WithUploadEvent()` |
| `WithTransitionError(err, temp)` | Record error details | `WithTransitionError(err, false)` |
| `WithETag(etag)` | Update ETag | `WithETag(newETag)` |
| `WithContentHash(hash)` | Update content hash | `WithContentHash(hash)` |
| `WithSize(size)` | Update file size | `WithSize(uint64(len(data)))` |
| `WithPinState(pin)` | Update pin policy | `WithPinState(pinState)` |
| `ForceTransition()` | Bypass validation (use sparingly) | `ForceTransition()` |
| `ClearPendingRemote()` | Clear pending markers | `ClearPendingRemote()` |
| `WithTransitionTimestamp(ts)` | Override timestamp | `WithTransitionTimestamp(time.Now())` |

## Common Patterns

### Download Workflow

```go
// 1. Start download
fs.transitionItemState(id, metadata.ItemStateHydrating,
    metadata.WithHydrationEvent(),
    metadata.WithWorker("download:"+id))

// 2. Download content
content, err := downloadFile(id)
if err != nil {
    // 3a. Handle failure
    fs.transitionItemState(id, metadata.ItemStateError,
        metadata.WithHydrationEvent(),
        metadata.WithTransitionError(err, isTemporary(err)))
    return err
}

// 3b. Success
fs.transitionItemState(id, metadata.ItemStateHydrated,
    metadata.WithHydrationEvent(),
    metadata.WithWorker("download:"+id),
    metadata.WithContentHash(hash),
    metadata.ClearPendingRemote())
```

### Upload Workflow

```go
// 1. Mark dirty
fs.transitionItemState(id, metadata.ItemStateDirtyLocal,
    metadata.WithUploadEvent(),
    metadata.WithWorker("upload:"+id))

// 2. Upload content
response, err := uploadFile(id, content)
if err != nil {
    // 3a. Handle failure
    fs.transitionItemState(id, metadata.ItemStateError,
        metadata.WithUploadEvent(),
        metadata.WithTransitionError(err, isTemporary(err)))
    return err
}

// 3b. Success
fs.transitionItemState(id, metadata.ItemStateHydrated,
    metadata.WithUploadEvent(),
    metadata.WithETag(response.ETag),
    metadata.WithSize(response.Size))
```

### Conflict Detection

```go
// During delta sync
if entry.State == metadata.ItemStateDirtyLocal && remoteETag != entry.ETag {
    // Remote changed while we have local changes
    fs.transitionItemState(id, metadata.ItemStateConflict)
    
    // Create conflict copy
    createConflictCopy(id, entry)
}
```

### Cache Eviction

```go
// Evict to free space
if entry.State == metadata.ItemStateHydrated {
    // Remove content from cache
    fs.content.Delete(id)
    
    // Transition to GHOST
    fs.transitionItemState(id, metadata.ItemStateGhost)
}
```

## Virtual Entries

Virtual entries (e.g., `.xdg-volume-info`) have special handling:

```go
// Virtual entries always remain HYDRATED
entry := &metadata.Entry{
    ID:            "local-xdg-volume-info",
    RemoteID:      "",  // NULL for virtual
    Name:          ".xdg-volume-info",
    ItemType:      metadata.ItemKindFile,
    State:         metadata.ItemStateHydrated,  // Always HYDRATED
    Virtual:       true,  // Mark as virtual
    OverlayPolicy: metadata.OverlayPolicyLocalWins,
}

// Attempting to transition virtual entries to non-HYDRATED states will fail
err := fs.transitionItemState("local-xdg-volume-info", metadata.ItemStateGhost)
// Returns: ErrInvalidTransition: virtual entries must remain HYDRATED
```

## Error Handling

### Temporary vs Permanent Errors

```go
// Temporary error (will retry)
fs.transitionItemState(id, metadata.ItemStateError,
    metadata.WithTransitionError(err, true))  // temporary=true

// Permanent error (won't retry automatically)
fs.transitionItemState(id, metadata.ItemStateError,
    metadata.WithTransitionError(err, false))  // temporary=false
```

### Retry from Error State

```go
entry, _ := fs.metadataStore.Get(ctx, id)
if entry.State == metadata.ItemStateError {
    if entry.LastError != nil && entry.LastError.Temporary {
        // Retry download
        fs.transitionItemState(id, metadata.ItemStateHydrating,
            metadata.WithHydrationEvent())
    }
}
```

## Monitoring and Statistics

```go
// Get state distribution
stats, err := fs.GetStats()
if err != nil {
    return err
}

fmt.Printf("Ghost: %d\n", stats.HydrationGhost)
fmt.Printf("Hydrating: %d\n", stats.HydrationHydrating)
fmt.Printf("Hydrated: %d\n", stats.HydrationHydrated)
fmt.Printf("Dirty Local: %d\n", stats.HydrationDirtyLocal)
fmt.Printf("Errored: %d\n", stats.HydrationErrored)
```

## Best Practices

1. **Always use transition options**: Provide context with `WithHydrationEvent()`, `WithUploadEvent()`, etc.
2. **Track workers**: Use `WithWorker()` to identify which component performed the transition
3. **Record errors properly**: Use `WithTransitionError()` with correct temporary flag
4. **Update metadata**: Use `WithETag()`, `WithContentHash()`, `WithSize()` when available
5. **Clear pending flags**: Use `ClearPendingRemote()` after successful operations
6. **Avoid force transitions**: Only use `ForceTransition()` when absolutely necessary
7. **Handle virtual entries**: Check `entry.Virtual` before attempting state transitions
8. **Log transitions**: Add debug logging for state transitions to aid troubleshooting

## Common Pitfalls

1. **Forgetting to transition**: Always transition state when operations complete
2. **Wrong transition direction**: Ensure transition is valid (check allowed map)
3. **Missing context**: Always provide event type (hydration/upload) and worker ID
4. **Not handling errors**: Always transition to ERROR state on failure
5. **Ignoring virtual entries**: Virtual entries must remain HYDRATED
6. **Force transition abuse**: Don't use `ForceTransition()` to bypass validation
7. **Missing metadata updates**: Update ETag, hash, size when available
8. **Not clearing pending flags**: Clear `PendingRemote` after successful operations

## Testing State Transitions

```go
func TestStateTransition(t *testing.T) {
    // Setup
    store, _ := metadata.NewBoltStore(db, bucketMetadataV2)
    stateMgr, _ := metadata.NewStateManager(store)
    
    // Create entry
    entry := &metadata.Entry{
        ID:       "test-id",
        Name:     "test.txt",
        ItemType: metadata.ItemKindFile,
        State:    metadata.ItemStateGhost,
    }
    store.Insert(ctx, entry)
    
    // Test transition
    updated, err := stateMgr.Transition(ctx, "test-id", 
        metadata.ItemStateHydrating,
        metadata.WithHydrationEvent(),
        metadata.WithWorker("test"))
    
    require.NoError(t, err)
    assert.Equal(t, metadata.ItemStateHydrating, updated.State)
    assert.NotNil(t, updated.Hydration.StartedAt)
    assert.Equal(t, "test", updated.Hydration.WorkerID)
}
```

## References

- **Implementation**: `internal/metadata/manager.go`
- **State Definitions**: `internal/metadata/state.go`
- **Entry Structure**: `internal/metadata/entry.go`
- **Filesystem Integration**: `internal/fs/metadata_store.go`
- **Design Document**: `.kiro/specs/system-verification-and-fix/design.md`
- **Review Document**: `docs/verification-phase17-state-model-review.md`
