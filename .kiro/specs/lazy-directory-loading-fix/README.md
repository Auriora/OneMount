# Lazy Directory Loading Performance Fix - Spec Summary

## Overview

This spec addresses the issue where directories appear empty on first access in OneMount, requiring users to access them twice to see contents.

## Problem

**Current behavior (WRONG)**:
- First `ls` of a directory returns empty
- Background refresh fetches data (5-10 seconds)
- Second `ls` shows contents

**This is terrible UX!**

## Solution

**New behavior (CORRECT)**:
1. **First access**: Block and fetch data synchronously (up to 10 seconds)
   - **NEVER return empty**
   - Return complete data or error
2. **Subsequent access with fresh cache**: Return immediately (< 50ms)
3. **Subsequent access with stale cache**:
   - Try to refresh with 2 second timeout
   - If refresh succeeds: Return fresh data
   - If refresh times out: Serve stale data and continue refresh in background

## Key Principle

**NEVER RETURN EMPTY** - Always wait for data or serve stale cache.

## Implementation Tasks

See `tasks.md` for detailed implementation tasks:
1. Update `GetChildrenID()` to block on first access
2. Implement synchronous fetch with timeout
3. Implement stale cache refresh with timeout
4. Update metadata request manager
5. Update all tests to reflect new behavior
6. Add new tests for **NEVER empty** requirement

## Files

- `requirements.md` - Detailed requirements and acceptance criteria
- `design.md` - Technical design and implementation approach
- `tasks.md` - Step-by-step implementation tasks

## Success Criteria

- ✅ Directories **NEVER** appear empty
- ✅ First access shows complete contents (blocks up to 10 seconds)
- ✅ Subsequent accesses are fast (< 50ms from cache)
- ✅ Stale cache tries to refresh (2 second timeout) before serving stale data
- ✅ All tests updated and passing
- ✅ File managers work correctly

## Next Steps

1. Review and approve this spec
2. Begin implementation following `tasks.md`
3. Update tests as you implement
4. Validate with manual testing
5. Document the fix

## References

- Issue: `docs/issues/lazy-directory-loading-performance.md`
- ADR-003: `docs/2-architecture/decisions/ADR-003-metadata-request-prioritization.md`
- Current implementation: `internal/fs/cache.go`
- Tests: `internal/fs/cache_test.go`
