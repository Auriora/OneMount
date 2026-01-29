# OneMount Specifications

## Overview

This directory contains focused specifications for OneMount functionality. Each spec is self-contained with its own requirements, design, and implementation tasks.

## Active Specs

### 1. Authentication and Account Management
**Location**: [authentication-and-accounts/](authentication-and-accounts/)  
**Status**: In Progress  
**Covers**: OAuth2 authentication, token management, multiple account support

### 2. Filesystem Mounting and Initialization
**Location**: [filesystem-mounting/](filesystem-mounting/)  
**Status**: Completed  
**Covers**: FUSE mounting, mount point validation, advanced mounting options

### 3. Directory Loading and Caching
**Location**: [directory-loading-and-caching/](directory-loading-and-caching/)  
**Status**: In Progress  
**Covers**: Lazy loading, recursive prefetch, cache population, stale cache refresh, scoped invalidation  
**Note**: Merged lazy-directory-loading-fix and initial sync/cache requirements from system verification

### 4. Virtual File Management
**Location**: [virtual-file-management/](virtual-file-management/)  
**Status**: Planned  
**Covers**: .xdg-volume-info handling, local-only files, overlay policies

### 5. FUSE Performance Optimization
**Location**: [fuse-performance/](fuse-performance/)  
**Status**: Planned  
**Covers**: Non-blocking FUSE operations, metadata prioritization, lock optimization

### 6. File Download and Hydration
**Location**: [file-download-hydration/](file-download-hydration/)  
**Status**: Completed  
**Covers**: On-demand downloads, ETag validation, download manager, hydration states

### 7. File Upload and Modification
**Location**: [file-upload-modification/](file-upload-modification/)  
**Status**: Completed  
**Covers**: Upload queue, chunked uploads, directory operations, local change tracking

### 8. Delta Sync and Realtime Updates
**Location**: [delta-sync-realtime/](delta-sync-realtime/)  
**Status**: In Progress  
**Covers**: Socket.IO subscriptions, delta polling, metadata cache updates, remote changes

### 9. Offline Mode and Sync
**Location**: [offline-mode-sync/](offline-mode-sync/)  
**Status**: Completed  
**Covers**: Offline detection, change tracking, connectivity detection, queued operations

### 10. Cache Management
**Location**: [cache-management/](cache-management/)  
**Status**: Completed  
**Covers**: Cache expiration, ETag-based invalidation, statistics, size limits, eviction

### 11. Conflict Resolution
**Location**: [conflict-resolution/](conflict-resolution/)  
**Status**: Completed  
**Covers**: Conflict detection, resolution policies, conflict copy creation, user notification

### 12. User Notifications and Status
**Location**: [notifications-and-status/](notifications-and-status/)  
**Status**: Completed  
**Covers**: D-Bus signals, status icons, feedback levels, file manager integration

### 13. Error Handling and Recovery
**Location**: [error-handling-recovery/](error-handling-recovery/)  
**Status**: Completed  
**Covers**: Network error handling, rate limit backoff, crash recovery, state persistence

### 14. Integration Testing Framework
**Location**: [integration-testing/](integration-testing/)  
**Status**: Completed  
**Covers**: Test infrastructure, Docker test environment, end-to-end scenarios, test data

## Archived Specs

### System Verification and Fix (Archived)
**Location**: [archive/system-verification-and-fix/](archive/system-verification-and-fix/)  
**Status**: Archived (2026-01-29)  
**Reason**: Broken down into focused specs listed above

### Lazy Directory Loading Fix (Archived)
**Location**: [archive/lazy-directory-loading-fix/](archive/lazy-directory-loading-fix/)  
**Status**: Archived (2026-01-29)  
**Reason**: Merged into directory-loading-and-caching spec

## Dependency Graph

```
Authentication & Accounts (1)
    ↓
Filesystem Mounting (2)
    ↓
Directory Loading & Caching (3)
    ├→ Virtual File Management (4)
    └→ FUSE Performance (5)
    
File Download & Hydration (6) ←→ Cache Management (10)
    ↓
File Upload & Modification (7)
    ↓
Delta Sync & Realtime (8) ←→ Conflict Resolution (11)
    ↓
Offline Mode & Sync (9)

Error Handling & Recovery (13) → All specs
Notifications & Status (12) → All specs
Integration Testing (14) → All specs
```

## Spec Structure

Each spec follows this structure:

```
.kiro/specs/<spec-name>/
├── README.md           # Overview and status
├── requirements.md     # User stories and acceptance criteria
├── design.md          # Solution design and architecture
└── tasks.md           # Implementation tasks
```

## Working with Specs

### Creating a New Spec

1. Create directory: `mkdir -p .kiro/specs/<spec-name>/`
2. Copy template files from an existing spec
3. Update README.md with overview and status
4. Define requirements with acceptance criteria
5. Create design document with architecture
6. Break down into implementation tasks
7. Update this index with the new spec

### Updating a Spec

1. Update the relevant document (requirements, design, or tasks)
2. Update the "Last Updated" date in README.md
3. Update status if changed
4. Update cross-references in related specs if dependencies change

### Completing a Spec

1. Mark all tasks as complete
2. Update status to "Completed" in README.md
3. Document any deviations from original design
4. Update this index

## References

- Original monolithic spec: `.kiro/specs/archive/system-verification-and-fix/`
- SRS: `docs/1-requirements/software-requirements-specification.md`
- Architecture: `docs/2-architecture/`
- Testing conventions: `.kiro/steering/testing-conventions.md`
