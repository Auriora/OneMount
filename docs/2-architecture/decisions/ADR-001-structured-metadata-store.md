# ADR-001: Structured Metadata Store

## Status

Accepted

## Context

The original design used a simple `sync.Map` for in-memory metadata storage. As the system evolved, we needed:

1. **State Management**: Track file states (GHOST, HYDRATING, HYDRATED, DIRTY_LOCAL, etc.)
2. **Persistence**: Reliably persist metadata across restarts
3. **Query Capabilities**: Efficiently query metadata by various criteria
4. **Consistency**: Ensure metadata consistency across concurrent operations
5. **Validation**: Validate state transitions to prevent invalid states

The simple `sync.Map` approach had limitations:
- No structured state management
- Limited query capabilities
- No validation of state transitions
- Difficult to maintain consistency
- No clear separation of concerns

## Decision

We implemented a structured metadata store (`internal/metadata/`) with:

1. **Metadata Store Interface**: Defines operations for metadata persistence
2. **State Controller**: Validates and manages item state transitions
3. **Overlay Policy System**: Manages virtual files and local-only entries
4. **BBolt Integration**: Provides reliable persistence with ACID properties
5. **Structured Queries**: Enables efficient querying by state, parent, etc.

**Key Components**:

```go
// Metadata Store Interface
type Store interface {
    Get(id string) (*Item, error)
    Put(item *Item) error
    Delete(id string) error
    List(parent string) ([]*Item, error)
    QueryByState(state ItemState) ([]*Item, error)
}

// State Controller Interface
type StateController interface {
    TransitionState(id string, from, to ItemState) error
    GetState(id string) (ItemState, error)
    ValidateTransition(from, to ItemState) error
}

// Item States
const (
    StateGhost        ItemState = "GHOST"
    StateHydrating    ItemState = "HYDRATING"
    StateHydrated     ItemState = "HYDRATED"
    StateDirtyLocal   ItemState = "DIRTY_LOCAL"
    StateDeletedLocal ItemState = "DELETED_LOCAL"
    StateConflict     ItemState = "CONFLICT"
    StateError        ItemState = "ERROR"
)
```

## Consequences

### Positive

1. **Better State Management**: Clear state machine with validated transitions
2. **Improved Reliability**: ACID properties from BBolt ensure consistency
3. **Enhanced Query Capabilities**: Efficient queries by state, parent, etc.
4. **Clearer Code**: Separation of concerns makes code easier to understand
5. **Better Testing**: State transitions can be tested independently
6. **Conflict Detection**: Easier to detect and handle conflicts
7. **Virtual File Support**: Clean implementation of local-only files

### Negative

1. **Increased Complexity**: More components to understand and maintain
2. **Performance Overhead**: Additional layer adds some latency
3. **Migration Required**: Existing deployments need migration
4. **Learning Curve**: Developers need to understand state machine

### Neutral

1. **Code Organization**: More files and packages to navigate
2. **Testing Requirements**: More comprehensive testing needed
3. **Documentation Needs**: Requires detailed documentation

## Alternatives Considered

### 1. Enhanced sync.Map

**Pros**:
- Simpler implementation
- Lower overhead
- Familiar to developers

**Cons**:
- No persistence
- No state validation
- Limited query capabilities
- Difficult to maintain consistency

**Rejected**: Insufficient for reliability requirements

### 2. External Database (PostgreSQL, MySQL)

**Pros**:
- Powerful query capabilities
- Well-understood technology
- Rich ecosystem

**Cons**:
- External dependency
- Deployment complexity
- Overkill for use case
- Network latency

**Rejected**: Too complex for embedded use case

### 3. SQLite

**Pros**:
- Embedded database
- SQL query capabilities
- Well-tested

**Cons**:
- CGo dependency
- Larger binary size
- More complex than needed
- Locking issues with FUSE

**Rejected**: BBolt better suited for key-value use case

## Implementation Notes

### Migration Path

1. **Detect Old Format**: Check for sync.Map-based metadata
2. **Convert to New Format**: Migrate to structured metadata store
3. **Validate Migration**: Ensure all data migrated correctly
4. **Clean Up**: Remove old metadata format

### State Transition Rules

Valid transitions documented in `docs/2-architecture/software-design-specification.md`:

- GHOST → HYDRATING (user access or pinning)
- HYDRATING → HYDRATED (download success)
- HYDRATING → ERROR (download failure)
- HYDRATED → DIRTY_LOCAL (local modification)
- HYDRATED → GHOST (cache eviction)
- DIRTY_LOCAL → HYDRATED (upload success)
- DIRTY_LOCAL → CONFLICT (remote change detected)
- etc.

### Performance Considerations

- **Caching**: In-memory cache for frequently accessed metadata
- **Batch Operations**: Support batch updates for efficiency
- **Indexing**: Appropriate indexes for common queries
- **Connection Pooling**: Reuse BBolt transactions

## References

- [Requirement 21: Metadata State Model Verification](../../1-requirements/software-requirements-specification.md#requirement-21)
- [State Machine Diagram](../resources/state-machine-diagram.puml)
- [Metadata Store Implementation](../../../internal/metadata/)
- [State Controller Implementation](../../../internal/fs/filesystem_types.go)

## Related ADRs

- ADR-002: Socket.IO Realtime Subscriptions
- ADR-003: Metadata Request Prioritization
- ADR-005: Virtual File Handling

## Revision History

| Date | Version | Author | Changes |
|------|---------|--------|---------|
| 2025-01-21 | 1.0 | AI Agent | Initial version |
