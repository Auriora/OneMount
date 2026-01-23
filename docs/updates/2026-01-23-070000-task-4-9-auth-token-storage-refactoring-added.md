# Task 4.9: Auth Token Storage Refactoring Added to Spec

**Date**: 2026-01-23  
**Time**: 07:00:00  
**Type**: Spec Update  
**Status**: ✅ Complete  
**Related Tasks**: Task 46.1.6, Task 4.9

## Summary

Added new task 4.9 to Phase 3 (Authentication Component Verification) of the system-verification-and-fix spec to address the auth token storage architecture issue discovered during integration test verification.

## Problem Identified

During task 46.1.6 (integration test auth verification), we discovered that auth tokens are stored based on mount point location rather than account identity:

**Current**: `~/.cache/onemount/{mount-path}/auth_tokens.json`  
**Problem**: Different mount points = different token paths

This causes:
- Docker test reliability issues (tests can't find tokens)
- Token duplication (same account at different locations)
- Token loss on remount (changing mount point loses tokens)
- Test environment confusion

## Solution Designed

Store tokens based on account identity (email hash):

**Proposed**: `~/.cache/onemount/accounts/{account-hash}/auth_tokens.json`

Where `account-hash` is the first 16 characters of SHA256 hash of the account email.

## Changes Made

### 1. Tasks Document (`.kiro/specs/system-verification-and-fix/tasks.md`)

Added new task 4.9 with 4 sub-tasks:
- **Task 4.9**: Refactor auth token storage to use account-based paths
  - **Task 4.9.1**: Phase 1: Investigation & Prototyping (1-2 days)
  - **Task 4.9.2**: Phase 2: Core Implementation (2-3 days)
  - **Task 4.9.3**: Phase 3: Integration & Testing (2-3 days)
  - **Task 4.9.4**: Phase 4: Documentation & Cleanup (1 day)

**Rationale for placement**: Added as sub-task of Phase 3 (Authentication Component Verification) to avoid renumbering subsequent phases. This makes sense because:
- It's an authentication-related issue
- It was discovered during authentication verification
- It logically belongs with authentication work

### 2. Requirements Document (`.kiro/specs/system-verification-and-fix/requirements.md`)

Added new acceptance criterion to Requirement 1 (Authentication Verification):

**Requirement 1.6**: WHEN storing authentication tokens, THE OneMount System SHALL use account-based storage paths derived from account identity (email hash) rather than mount point location, ensuring tokens are accessible regardless of mount point changes and preventing token duplication across multiple mounts of the same account

### 3. Design Document (`.kiro/specs/system-verification-and-fix/design.md`)

Added comprehensive section on "Account-Based Token Storage Architecture" including:
- Problem statement with specific examples
- Solution design with code examples
- Benefits (mount point independence, no duplication, reliable testing, account isolation)
- Implementation details (hash function, migration strategy)
- Security considerations (file permissions, hash algorithm, privacy)
- Testing requirements

Added new correctness property:

**Property 4.1: Account-Based Token Storage**: *For any* authentication token storage operation, the system should store tokens using account-based paths (account email hash) rather than mount-point-based paths, ensuring tokens are accessible regardless of mount point location and preventing duplication  
**Validates: Requirements 1.6**

## Implementation Plan

The refactoring follows a 4-phase approach (6-9 days total):

### Phase 1: Investigation & Prototyping (1-2 days)
- Analyze current token storage usage
- Prototype account-based storage functions
- Test hash generation and collision resistance
- Verify Docker environment behavior

### Phase 2: Core Implementation (2-3 days)
- Add `GetAuthTokensPathByAccount()` function
- Add `hashAccount()` helper function
- Implement `FindAuthTokens()` with fallback logic
- Add automatic token migration
- Update `SaveAuthTokens()` and `LoadAuthTokens()`
- Add unit tests

### Phase 3: Integration & Testing (2-3 days)
- Update `cmd/onemount/main.go`
- Update test fixtures
- Update Docker auth setup scripts
- Add integration tests for migration
- Test Docker environment token access
- Test multiple account scenarios

### Phase 4: Documentation & Cleanup (1 day)
- Update architecture documentation
- Add migration guide for users
- Update test documentation
- Add deprecation warnings
- Update CHANGELOG

## Migration Strategy

**Backward Compatibility**:
- Old token locations still work (fallback mechanism)
- Automatic migration on first use
- No breaking changes
- No re-authentication required

**Deprecation Timeline**:
- v1.0: Add account-based storage, auto-migrate
- v1.1: Log warnings for old locations
- v2.0: Remove support for old locations (after 6+ months)

## Benefits

1. **Mount Point Independence**: Tokens accessible regardless of mount location
2. **No Token Duplication**: One account = one token file
3. **Reliable Docker Testing**: Tests find tokens consistently
4. **Account Isolation**: Different accounts have separate token files
5. **Multi-Account Support**: Multiple accounts can be mounted simultaneously
6. **Better Privacy**: Email not visible in filesystem (hashed)

## References

### Analysis and Planning Documents
- **Analysis Report**: `docs/reports/2026-01-23-063800-auth-token-storage-architecture-analysis.md`
- **Implementation Plan**: `docs/plans/auth-token-storage-refactoring-plan.md`

### Related Tasks
- **Task 46.1.6**: Integration test auth verification (completed)
- **Task 4.9**: Auth token storage refactoring (added, not started)

### Code Locations
- Token path logic: `internal/graph/oauth2.go:47-58`
- Token save/load: `internal/graph/oauth2.go:60-95`
- Test fixtures: `internal/testutil/helpers/fs_fixtures.go`
- Docker auth setup: `docker/images/test-runner/entrypoint.sh`

### Documentation
- Spec requirements: `.kiro/specs/system-verification-and-fix/requirements.md`
- Spec design: `.kiro/specs/system-verification-and-fix/design.md`
- Spec tasks: `.kiro/specs/system-verification-and-fix/tasks.md`

## Next Steps

1. Begin Phase 1 (Investigation & Prototyping) when ready to start implementation
2. Follow the 4-phase plan outlined in the tasks
3. Ensure all tests pass with backward compatibility
4. Document migration process for users

## Rules Applied

**Rules consulted**: 
- `documentation-conventions.md` (Priority 20) - Documentation structure and placement
- `general-preferences.md` (Priority 50) - Rule application and documentation
- `coding-standards.md` (Priority 100) - Code quality and design principles

**Rules applied**:
- Placed update in `docs/updates/` with timestamp (documentation-conventions.md)
- Updated spec requirements, design, and tasks (documentation-conventions.md)
- Followed SOLID principles in design (coding-standards.md, general-preferences.md)
- Documented rationale and traceability (general-preferences.md)

**Overrides**: None

## Conclusion

Successfully added auth token storage refactoring as task 4.9 to the system-verification-and-fix spec. The task is well-documented with clear requirements, design, implementation plan, and testing strategy. The solution maintains backward compatibility while fixing the underlying architectural issue that was causing Docker test reliability problems.

The work is ready to begin when the team is ready to implement it, with comprehensive documentation to guide the implementation.
