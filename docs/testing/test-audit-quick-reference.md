# Test Audit Quick Reference

**Task**: 46.1.1 - Audit all test functions for correct naming conventions  
**Status**: ✅ COMPLETED  
**Date**: January 22, 2026

## Quick Stats

| Metric | Count | Percentage |
|--------|-------|------------|
| **Total Tests** | **752** | **100%** |
| Correctly Labeled | 460 | 61.2% |
| **Issues Found** | **292** | **38.8%** |

## Issues Breakdown

| Issue Type | Count | Severity |
|------------|-------|----------|
| **Mislabeled Tests** | **66** | 🔴 CRITICAL |
| **Unlabeled Tests** | **226** | 🔴 CRITICAL |
| Unclear Requirements | 167 | 🟡 HIGH |

## Critical Findings

### 1. Mislabeled Tests (66 tests) 🔴

Tests labeled as `TestUT_` (unit tests) but requiring authentication:

```
internal/fs/advanced_mounting_test.go:
  - TestUT_FS_AdvancedMounting_01_MountTimeoutConfiguration
  - TestUT_FS_AdvancedMounting_02_StaleLockDetection
  - TestUT_FS_AdvancedMounting_03_DatabaseRetryLogic
  - TestUT_FS_AdvancedMounting_04_ConfigurationValidation

internal/fs/cache_management_test.go:
  - TestUT_FS_Cache_01_CacheInvalidation_WorksCorrectly
  - TestUT_FS_Cache_02_ContentCache_Operations
  - TestUT_FS_Cache_03_CacheConsistency_MultipleOperations
  - TestUT_FS_Cache_04_CacheInvalidation_Comprehensive
  - TestUT_FS_Cache_05_CachePerformance_Operations

... (61 more)
```

**Impact**: These tests will fail when run as unit tests without auth tokens.

**Action Required**: Either:
- Relabel as `TestIT_*` (integration tests)
- Refactor to use mocks and keep as `TestUT_*`

### 2. Unlabeled Tests (226 tests) 🔴

Tests with no type prefix, making it impossible to determine requirements:

```
internal/fs/cache_test.go:
  - TestGetChildrenIDUsesMetadataStoreWhenOffline
  - TestGetPathUsesMetadataStoreWhenOffline
  - TestGetChildrenIDReturnsQuicklyWhenUncached
  - TestGetChildrenIDDoesNotCallGraphWhenMetadataPresent
  - TestFallbackRootFromMetadata

internal/fs/concurrency_test.go:
  - TestConcurrentFileAccess
  - TestConcurrentCacheOperations
  - TestDeadlockPrevention
  - TestDirectoryEnumerationWhileRefreshing
  - TestHighConcurrencyStress
  - TestConcurrentDirectoryOperations

... (214 more)
```

**Impact**: Cannot run specific test types in isolation, blocks CI/CD configuration.

**Action Required**: Add appropriate prefix based on actual requirements.

## Test Naming Conventions

| Prefix | Purpose | Auth Required | Example |
|--------|---------|---------------|---------|
| `TestUT_*` | Unit tests | ❌ No | `TestUT_FS_CacheHit` |
| `TestIT_*` | Integration tests | ✅ Yes | `TestIT_FS_FileUpload` |
| `TestProperty*` | Property-based tests | Varies | `TestProperty24_OfflineDetection` |
| `TestSystemST_*` | System tests | ✅ Yes | `TestSystemST_CompleteWorkflow` |

## Files Generated

1. **Detailed Report**: `docs/testing/test-audit-report.md`
   - Complete analysis with recommendations
   - Examples and best practices
   - Appendices with detailed lists

2. **Detailed CSV**: `docs/testing/test-audit-detailed.csv`
   - All 752 tests with analysis
   - Columns: File, Test Name, Current Label, Auth Required, Recommendation
   - Sortable and filterable for analysis

3. **Summary Statistics**: `docs/testing/test-audit-summary.txt`
   - Quick statistics and counts
   - Top issues by file
   - Action items prioritized

4. **Quick Reference**: `docs/testing/test-audit-quick-reference.md` (this file)
   - At-a-glance summary
   - Critical findings highlighted
   - Quick action items

## Immediate Actions Required

### Priority 1: Fix Mislabeled Tests (66 tests)

```bash
# Review mislabeled tests
grep "MISLABELED" docs/testing/test-audit-detailed.csv

# For each mislabeled test, either:
# Option A: Relabel as integration test
# Option B: Refactor to use mocks
```

### Priority 2: Label Unlabeled Tests (226 tests)

```bash
# Review unlabeled tests by file
grep ",NONE," docs/testing/test-audit-detailed.csv | cut -d',' -f1 | sort | uniq -c

# Add appropriate prefix based on auth requirements
```

### Priority 3: Create Test Fixtures (Task 46.1.2)

- `SetupMockFSTestFixture` for unit tests (no auth)
- `SetupIntegrationFSTestFixture` for integration tests (with auth)
- `SetupSystemTestFixture` for system tests (full setup)

## How to Use This Audit

### For Developers

1. **Check your tests**: Search for your files in the CSV
2. **Fix mislabeled tests**: Refactor or relabel as needed
3. **Label unlabeled tests**: Add appropriate prefix
4. **Use correct fixtures**: Switch to mock fixtures for unit tests

### For CI/CD

1. **Run unit tests only**: `go test -v -run "^TestUT_" ./...`
2. **Run integration tests**: `go test -v -run "^TestIT_" ./...` (with auth)
3. **Skip tests without auth**: Add skip logic to integration tests

### For Code Review

1. **Check test naming**: Ensure new tests follow conventions
2. **Verify fixtures**: Unit tests should use mocks
3. **Add skip logic**: Integration tests should skip without auth

## Next Steps

1. ✅ **Task 46.1.1**: Audit complete (this task)
2. ⏭️ **Task 46.1.2**: Create separate test fixtures
3. ⏭️ **Task 46.1.3**: Refactor unit tests to use mocks
4. ⏭️ **Task 46.1.4**: Add skip logic to integration tests
5. ⏭️ **Task 46.1.5**: Verify system tests

## Commands

### View mislabeled tests
```bash
grep "MISLABELED" docs/testing/test-audit-detailed.csv | less
```

### View unlabeled tests by file
```bash
grep ",NONE," docs/testing/test-audit-detailed.csv | cut -d',' -f1,2 | sort
```

### Count issues by file
```bash
grep -E "MISLABELED|ADD:" docs/testing/test-audit-detailed.csv | \
  cut -d',' -f1 | sort | uniq -c | sort -rn
```

### Run only unit tests (should work without auth)
```bash
docker compose -f docker/compose/docker-compose.test.yml run --rm \
  test-runner go test -v -run "^TestUT_" ./...
```

### Run only integration tests (requires auth)
```bash
docker compose -f docker/compose/docker-compose.test.yml \
  -f docker/compose/docker-compose.auth.yml run --rm \
  test-runner go test -v -run "^TestIT_" ./...
```

## References

- **Full Report**: `docs/testing/test-audit-report.md`
- **Detailed CSV**: `docs/testing/test-audit-detailed.csv`
- **Summary Stats**: `docs/testing/test-audit-summary.txt`
- **Task List**: `.kiro/specs/system-verification-and-fix/tasks.md`

---

**Status**: ✅ Audit Complete - Ready for remediation (Tasks 46.1.2-46.1.5)
