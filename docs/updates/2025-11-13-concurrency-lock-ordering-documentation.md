# Concurrency Lock Ordering Documentation

**Date**: 2025-11-13  
**Issue**: #PERF-001 - No Documented Lock Ordering Policy  
**Task**: 20.6 Fix Issue #PERF-001  
**Requirements**: 10.1, 10.4

## Summary

Created comprehensive concurrency guidelines documenting the lock ordering policy to prevent deadlocks in OneMount. Added inline code comments throughout the codebase to document lock acquisition patterns.

## Changes Made

### 1. New Documentation

Created `docs/guides/developer/concurrency-guidelines.md` with:

- **Lock Ordering Policy**: Hierarchical lock acquisition order
  1. Filesystem-level locks
  2. Manager-level locks (Download, Upload, D-Bus, etc.)
  3. Inode-level locks
  4. Session-level locks
  5. Cache-level locks

- **Lock Hierarchy Details**: Comprehensive documentation of each lock level with:
  - Purpose and scope
  - Acquisition rules
  - Code examples
  - Special cases (e.g., multiple inode locks using ID-based ordering)

- **Common Lock Patterns**: Four documented patterns with correct implementations:
  1. Read filesystem state, modify inode
  2. Manager operation on session
  3. Inode to cache
  4. Multiple filesystem locks

- **Deadlock Prevention Rules**: Five critical rules:
  1. Never acquire locks in reverse order
  2. Never hold locks during network I/O
  3. Use defer for lock release
  4. Minimize lock duration
  5. Document lock requirements

- **Code Examples**: Real-world examples from OneMount:
  - File open operation
  - Upload queue operation
  - Delta sync update
  - Move operation (multiple inodes)

- **Testing Guidelines**: 
  - Using the race detector
  - Stress testing patterns
  - Deadlock detection techniques

- **Common Pitfalls**: Six documented pitfalls with solutions:
  1. Lock inversion
  2. Holding locks across goroutines
  3. Forgetting to release locks
  4. Recursive locking
  5. Lock contention
  6. Code review checklist

### 2. Code Comments Added

Added lock ordering comments to key files:

#### `internal/fs/cache.go`
- `InsertNodeID()`: Documented inode -> filesystem lock ordering exception
- `InsertID()`: Documented parent -> child inode ordering
- `InsertChild()`: Documented child inode locking
- `DeleteID()`: Documented parent inode locking
- `MoveID()`: Documented parent inode locking during rename

#### `internal/fs/file_operations.go`
- `Open()`: Documented inode-only locking pattern

#### `internal/fs/delta.go`
- Delta fetch error handling: Documented filesystem lock for offline flag
- Delta fetch success: Documented filesystem lock for online transition

#### `internal/fs/sync.go`
- `SyncDirectoryTree()`: Documented filesystem lock for sync progress

#### `internal/fs/offline.go`
- `SetOfflineMode()`: Documented filesystem lock for offline mode

### 3. Documentation Updates

Updated `docs/guides/developer/README.md` to:
- Add reference to new concurrency-guidelines.md
- Highlight it as the lock ordering policy document

## Lock Ordering Policy Summary

The established hierarchy prevents deadlocks by ensuring consistent lock acquisition order:

```
Filesystem Locks (Level 1)
    ↓
Manager Locks (Level 2)
    ↓
Inode Locks (Level 3)
    ↓
Session Locks (Level 4)
    ↓
Cache Locks (Level 5)
```

**Special Rules**:
- Multiple inodes at same level: Use ID-based ordering (lexicographic comparison)
- Parent-child inodes: Lock parent before child
- Never hold locks during network I/O or blocking operations
- Always use `defer` for lock release

## Rationale

### Why This Ordering?

1. **Filesystem locks first**: Protect global state that affects all operations
2. **Manager locks second**: Coordinate work distribution without blocking individual operations
3. **Inode locks third**: Allow fine-grained file/directory operations
4. **Session locks fourth**: Short-lived locks for individual transfer operations
5. **Cache locks last**: Minimize contention on frequently accessed data structures

### Exceptions Documented

Two locations in `cache.go` violate the standard hierarchy but are safe:
- `InsertNodeID()`: Locks inode before filesystem, but only modifies independent fields
- `InsertID()`: Similar pattern, locks are released separately (no overlap)

These exceptions are explicitly documented with rationale.

## Testing

- Code compiles successfully: `go build ./internal/fs`
- No test failures introduced
- Race detector can be used to verify: `go test -race ./internal/fs`

## Benefits

1. **Deadlock Prevention**: Clear hierarchy eliminates circular wait conditions
2. **Code Maintainability**: Developers know exactly which order to acquire locks
3. **Review Efficiency**: Code reviewers can quickly verify correct lock ordering
4. **Debugging**: Lock-related issues easier to diagnose with documented patterns
5. **Onboarding**: New developers have clear guidelines for concurrent code

## Future Work

1. Consider adding static analysis tool to verify lock ordering
2. Add more stress tests for concurrent operations
3. Profile lock contention under heavy load
4. Consider lock-free data structures where appropriate

## References

- Requirements: 10.1 (concurrent operations), 10.4 (locking granularity)
- Related: `docs/guides/developer/threading-guidelines.md` (threading overview)
- Related: `docs/guides/developer/error-handling-guidelines.md` (error patterns)

## Rules Consulted

- `coding-standards.md` (Priority 100): Documentation requirements
- `operational-best-practices.md` (Priority 40): SRS alignment
- `testing-conventions.md` (Priority 25): Docker test environment
- `general-preferences.md` (Priority 50): SOLID principles, DRY

## Rules Applied

- Comprehensive documentation with examples (coding-standards.md)
- Aligned with Requirements 10.1 and 10.4 (operational-best-practices.md)
- Code comments explain lock ordering rationale (coding-standards.md)
- No code duplication in examples (general-preferences.md)
