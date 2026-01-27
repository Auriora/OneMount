# Prompt: Break Down System Verification Spec into Focused Specs

## Context

The `.kiro/specs/system-verification-and-fix/` spec is too large and covers multiple distinct functional areas. It needs to be broken down into separate, focused specs that can be worked on independently.

## Objective

Break down the monolithic system verification spec into multiple focused specs, each covering a specific functional area. Each spec should be self-contained with its own requirements, design, and tasks.

## Current Structure

The system verification spec currently contains these requirements:

1. **Requirement 1**: Authentication Verification
2. **Requirement 2**: Basic Filesystem Mounting
3. **Requirement 2A**: Initial Synchronization and Caching (lazy directory loading)
4. **Requirement 2B**: Virtual File Management
5. **Requirement 2C**: Advanced Mounting Options
6. **Requirement 2D**: FUSE Operation Performance
7. **Requirement 3**: Basic On-Demand File Access
8. **Requirement 3A**: Download Status and Progress Tracking
9. **Requirement 3B**: Download Manager Configuration
10. **Requirement 3C**: File Hydration State Management
11. **Requirement 4**: File Modification and Upload Verification
12. **Requirement 5**: Delta Synchronization Verification
13. **Requirement 6**: Offline Mode Verification
14. **Requirement 7**: Cache Management Verification
15. **Requirement 8**: Conflict Resolution Verification
16. **Requirement 9**: User Notifications and Feedback
17. **Requirement 10**: File Status and D-Bus Integration Verification
18. **Requirement 11**: Error Handling and Recovery Verification
19. **Requirement 12**: Performance and Concurrency Verification
20. **Requirement 13**: Integration Test Coverage
21. **Requirement 14**: Multiple Account and Drive Support

## Proposed Breakdown

Create separate specs for these functional areas:

### 1. Authentication and Account Management
**Location**: `.kiro/specs/authentication-and-accounts/`
**Covers**:
- Requirement 1: Authentication Verification
- Requirement 14: Multiple Account and Drive Support
- Account-based storage paths
- Token management and refresh
- Multi-account mounting

### 2. Filesystem Mounting and Initialization
**Location**: `.kiro/specs/filesystem-mounting/`
**Covers**:
- Requirement 2: Basic Filesystem Mounting
- Requirement 2C: Advanced Mounting Options
- Mount point management
- Database initialization
- Resource allocation

### 3. Directory Loading and Caching (Lazy Loading Fix)
**Location**: `.kiro/specs/directory-loading-and-caching/`
**Covers**:
- Requirement 2A: Initial Synchronization and Caching
- Recursive prefetch implementation
- Cache population strategy
- NEVER return empty directories
- Stale cache refresh with timeout
**Note**: This should merge/replace the existing `.kiro/specs/lazy-directory-loading-fix/`

### 4. Virtual File Management
**Location**: `.kiro/specs/virtual-file-management/`
**Covers**:
- Requirement 2B: Virtual File Management
- `.xdg-volume-info` handling
- Local-only files with overlay policies
- Virtual file persistence

### 5. FUSE Performance Optimization
**Location**: `.kiro/specs/fuse-performance/`
**Covers**:
- Requirement 2D: FUSE Operation Performance
- Requirement 12: Performance and Concurrency Verification
- Non-blocking FUSE operations
- Metadata request prioritization
- Lock granularity optimization

### 6. File Download and Hydration
**Location**: `.kiro/specs/file-download-hydration/`
**Covers**:
- Requirement 3: Basic On-Demand File Access
- Requirement 3A: Download Status and Progress Tracking
- Requirement 3B: Download Manager Configuration
- Requirement 3C: File Hydration State Management
- ETag validation
- Download manager workers
- State transitions (GHOST → HYDRATING → HYDRATED)

### 7. File Upload and Modification
**Location**: `.kiro/specs/file-upload-modification/`
**Covers**:
- Requirement 4: File Modification and Upload Verification
- Upload queue management
- Chunked uploads for large files
- Directory creation and deletion
- Local change tracking

### 8. Delta Sync and Realtime Updates
**Location**: `.kiro/specs/delta-sync-realtime/`
**Covers**:
- Requirement 5: Delta Synchronization Verification
- Socket.IO subscriptions
- Delta polling fallback
- Metadata cache updates
- Remote change detection

### 9. Offline Mode and Sync
**Location**: `.kiro/specs/offline-mode-sync/`
**Covers**:
- Requirement 6: Offline Mode Verification
- Offline change tracking
- Connectivity detection
- Offline-to-online synchronization
- Queued operations

### 10. Cache Management
**Location**: `.kiro/specs/cache-management/`
**Covers**:
- Requirement 7: Cache Management Verification
- Cache expiration and cleanup
- ETag-based invalidation
- Cache statistics
- Size limits and eviction

### 11. Conflict Resolution
**Location**: `.kiro/specs/conflict-resolution/`
**Covers**:
- Requirement 8: Conflict Resolution Verification
- Conflict detection strategies
- Resolution policies (last-writer-wins, keep-both, etc.)
- Conflict copy creation
- User notification

### 12. User Notifications and Status
**Location**: `.kiro/specs/notifications-and-status/`
**Covers**:
- Requirement 9: User Notifications and Feedback
- Requirement 10: File Status and D-Bus Integration Verification
- D-Bus signals
- Status icons for file managers
- Feedback levels (none, basic, detailed)

### 13. Error Handling and Recovery
**Location**: `.kiro/specs/error-handling-recovery/`
**Covers**:
- Requirement 11: Error Handling and Recovery Verification
- Network error handling
- Rate limit backoff
- Crash recovery
- State persistence

### 14. Integration Testing Framework
**Location**: `.kiro/specs/integration-testing/`
**Covers**:
- Requirement 13: Integration Test Coverage
- Test infrastructure
- Docker test environment
- End-to-end test scenarios
- Test data management

## Task Instructions

For each spec area listed above:

### Step 1: Create Spec Directory Structure
```bash
mkdir -p .kiro/specs/<spec-name>/
```

### Step 2: Extract Requirements
Create `requirements.md` with:
- Overview section explaining the functional area
- User stories from the original spec
- Acceptance criteria (copy relevant requirements)
- Dependencies on other specs
- References to SRS and ADRs

### Step 3: Create Design Document
Create `design.md` with:
- Problem analysis
- Current implementation review
- Proposed solution design
- Architecture diagrams (if needed)
- Design constraints
- Testing strategy

### Step 4: Create Tasks Document
Create `tasks.md` with:
- Phased implementation plan
- Specific, actionable tasks
- Task dependencies
- Estimated effort
- Testing requirements

### Step 5: Create README
Create `README.md` with:
- Brief description of the spec
- Current status
- Links to requirements, design, and tasks
- Related specs and dependencies

### Step 6: Handle Special Cases

**For Directory Loading and Caching**:
- Merge with existing `.kiro/specs/lazy-directory-loading-fix/`
- Consolidate requirements from both specs
- Keep the more detailed design and tasks
- Archive the old spec if needed

**For Integration Testing**:
- Reference existing test infrastructure
- Link to `docs/testing/` documentation
- Include Docker test environment setup

### Step 7: Update Cross-References
After creating all specs:
- Update each spec's dependencies section
- Add cross-references between related specs
- Update `.kiro/specs/README.md` with the new structure
- Create a dependency graph showing relationships

### Step 8: Archive Original Spec
- Move `.kiro/specs/system-verification-and-fix/` to `.kiro/specs/archive/system-verification-and-fix/`
- Add a README explaining it was broken down
- Include links to the new focused specs

## Spec Template Structure

Each spec should follow this structure:

```
.kiro/specs/<spec-name>/
├── README.md           # Overview and status
├── requirements.md     # User stories and acceptance criteria
├── design.md          # Solution design and architecture
└── tasks.md           # Implementation tasks
```

## Quality Checklist

For each spec, ensure:
- [ ] Requirements are clear and testable
- [ ] Design addresses all requirements
- [ ] Tasks are specific and actionable
- [ ] Dependencies on other specs are documented
- [ ] Cross-references are accurate
- [ ] No duplication between specs
- [ ] Each spec is independently understandable
- [ ] Testing approach is defined

## Success Criteria

The breakdown is successful when:
1. All 14 functional areas have their own focused specs
2. Each spec is self-contained and independently workable
3. Dependencies between specs are clearly documented
4. No requirements are lost in the breakdown
5. The original monolithic spec is archived with proper references
6. A dependency graph shows relationships between specs

## Notes

- **Prioritize clarity over completeness** - it's better to have clear, focused specs than comprehensive but confusing ones
- **Avoid duplication** - if requirements overlap, choose the most appropriate spec and reference it from others
- **Keep specs independent** - minimize dependencies to allow parallel work
- **Document assumptions** - if a spec assumes another is complete, document it clearly
- **Use consistent terminology** - maintain glossary terms across all specs

## Example: Directory Loading and Caching Spec

Here's what the directory loading spec should look like after consolidation:

**Location**: `.kiro/specs/directory-loading-and-caching/`

**Requirements** (from both system verification and lazy loading specs):
- NEVER return empty directories
- Recursive prefetch on mount (metadata only, not content)
- Stale cache refresh with 2 second timeout
- Serve stale data on timeout, continue refresh in background
- Block on first access until data available (10 second timeout)

**Design**:
- Implement `StartPrefetch()` and `prefetchRecursive()`
- Update `GetChildrenID()` to be prefetch-aware
- Add helper methods for prefetch state tracking
- Use metadata states (GHOST → HYDRATING → HYDRATED)

**Tasks**:
- Phase 1: Implement recursive prefetch infrastructure
- Phase 2: Update GetChildrenID for prefetch awareness
- Phase 3: Testing (unit, integration, system)
- Phase 4: Performance validation

## References

- Original spec: `.kiro/specs/system-verification-and-fix/`
- Existing lazy loading spec: `.kiro/specs/lazy-directory-loading-fix/`
- SRS: `docs/1-requirements/software-requirements-specification.md`
- Architecture docs: `docs/2-architecture/`
- Testing conventions: `.kiro/steering/testing-conventions.md`
