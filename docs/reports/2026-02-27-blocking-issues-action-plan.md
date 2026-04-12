# OneMount v1.0 Blocking Issues - Detailed Action Plan
**Date**: February 27, 2026  
**Status**: READY TO EXECUTE  
**Estimated Time**: 3-5 days  
**Priority**: CRITICAL - Blocks v1.0 Release

---

## Executive Summary

This document provides a detailed, step-by-step action plan to resolve the two critical issues blocking the OneMount v1.0 release:

1. **Test Naming Convention Violations** - 19 remaining unlabeled tests
2. **QuickXORHash Comprehensive Testing** - Missing critical file integrity tests

**Good News**: The test labeling work is 92% complete (733/752 tests labeled). Only 19 tests remain, primarily `TestMain` functions and utility tests that don't need labeling.

---

## Issue #1: Test Naming Convention Violations

### Current Status: 92% COMPLETE ✅

**Progress**:
- ✅ 733 out of 752 tests properly labeled (97.5%)
- ⚠️ 19 tests remaining (2.5%)
- ✅ All critical filesystem and integration tests labeled
- ✅ Documentation updated with labeling summary

**Remaining Unlabeled Tests** (19 total):

#### Category A: TestMain Functions (8 tests) - NO ACTION NEEDED
These are test setup functions, not actual tests:
```
./internal/ui/systemd/setup_test.go:func TestMain(m *testing.M)
./internal/ui/setup_test.go:func TestMain(m *testing.M)
./internal/fs/testing_main_test.go:func TestMain(m *testing.M)
./internal/testutil/framework/setup_test.go:func TestMain(m *testing.M)
./internal/graph/setup_test.go:func TestMain(m *testing.M)
./cmd/common/setup_test.go:func TestMain(m *testing.M)
```
**Action**: None required - `TestMain` is a special Go testing function

#### Category B: Utility Tests (11 tests) - NEEDS LABELING
These are actual tests that need proper prefixes:

**File**: `internal/config/mounts_test.go` (4 tests)
```go
func TestMountsRegistry(t *testing.T)
func TestMountsRegistry_MultipleAccounts(t *testing.T)
func TestMountsRegistry_EmptyConfigDir(t *testing.T)
func TestMountsRegistry_UpdateMount(t *testing.T)
```
**Recommendation**: Label as `TestUT_` (unit tests, no auth required)

**File**: `internal/graph/oauth2_account_storage_test.go` (7 tests)
```go
func TestHashAccount(t *testing.T)
func TestHashAccountStability(t *testing.T)
func TestHashAccountCollisionResistance(t *testing.T)
func TestGetAuthTokensPathByAccount(t *testing.T)
func TestFindAuthTokens(t *testing.T)
func TestMigrateTokens(t *testing.T)
func TestMigrateTokensPermissions(t *testing.T)
func TestFileExists(t *testing.T)
func TestAuthenticateWithAccountStorage_Migration(t *testing.T)
```
**Recommendation**: Label as `TestUT_` (unit tests, testing utility functions)

### Action Plan for Issue #1

#### Step 1: Label Remaining Tests (1-2 hours)

**Task 1.1**: Label `internal/config/mounts_test.go` tests
```bash
# Rename tests to include TestUT_ prefix
sed -i 's/func TestMountsRegistry(/func TestUT_Config_MountsRegistry(/g' internal/config/mounts_test.go
sed -i 's/func TestMountsRegistry_MultipleAccounts(/func TestUT_Config_MountsRegistry_MultipleAccounts(/g' internal/config/mounts_test.go
sed -i 's/func TestMountsRegistry_EmptyConfigDir(/func TestUT_Config_MountsRegistry_EmptyConfigDir(/g' internal/config/mounts_test.go
sed -i 's/func TestMountsRegistry_UpdateMount(/func TestUT_Config_MountsRegistry_UpdateMount(/g' internal/config/mounts_test.go
```

**Task 1.2**: Label `internal/graph/oauth2_account_storage_test.go` tests
```bash
# Rename tests to include TestUT_ prefix
sed -i 's/func TestHashAccount(/func TestUT_Graph_HashAccount(/g' internal/graph/oauth2_account_storage_test.go
sed -i 's/func TestHashAccountStability(/func TestUT_Graph_HashAccountStability(/g' internal/graph/oauth2_account_storage_test.go
sed -i 's/func TestHashAccountCollisionResistance(/func TestUT_Graph_HashAccountCollisionResistance(/g' internal/graph/oauth2_account_storage_test.go
sed -i 's/func TestGetAuthTokensPathByAccount(/func TestUT_Graph_GetAuthTokensPathByAccount(/g' internal/graph/oauth2_account_storage_test.go
sed -i 's/func TestFindAuthTokens(/func TestUT_Graph_FindAuthTokens(/g' internal/graph/oauth2_account_storage_test.go
sed -i 's/func TestMigrateTokens(/func TestUT_Graph_MigrateTokens(/g' internal/graph/oauth2_account_storage_test.go
sed -i 's/func TestMigrateTokensPermissions(/func TestUT_Graph_MigrateTokensPermissions(/g' internal/graph/oauth2_account_storage_test.go
sed -i 's/func TestFileExists(/func TestUT_Graph_FileExists(/g' internal/graph/oauth2_account_storage_test.go
sed -i 's/func TestAuthenticateWithAccountStorage_Migration(/func TestUT_Graph_AuthenticateWithAccountStorage_Migration(/g' internal/graph/oauth2_account_storage_test.go
```

#### Step 2: Verify Labeling (15 minutes)

```bash
# Check for remaining unlabeled tests (should be 0 or only TestMain functions)
grep -r "^func Test[A-Z]" --include="*_test.go" . | \
  grep -v "TestUT_" | grep -v "TestIT_" | \
  grep -v "TestProperty" | grep -v "TestSystemST_" | \
  grep -v "TestMain"

# Expected output: 0 lines (all tests labeled)
```

#### Step 3: Compile and Test (30 minutes)

```bash
# Verify all tests compile
go build ./...

# Run unit tests only (should work without auth)
go test -v -run "^TestUT_" ./internal/config/... ./internal/graph/...

# Run full test suite in Docker
docker compose -f docker/compose/docker-compose.test.yml run --rm unit
```

#### Step 4: Update Documentation (15 minutes)

Update `docs/testing/test-labeling-summary.md`:
- Change status to "100% COMPLETE"
- Update statistics: 752/752 tests labeled
- Add completion date

**Total Time for Issue #1**: 2-3 hours

---

## Issue #2: QuickXORHash Comprehensive Testing

### Current Status: PARTIALLY COMPLETE ⚠️

**Existing Tests** (Good coverage):
- ✅ `TestUT_GR_29_01_QuickXORHash_Sum_CalculatesCorrectHashes` - Tests with Microsoft test vectors
- ✅ `TestUT_GR_30_01_QuickXORHash_WriteByBlocks_CalculatesCorrectHashes` - Block writing tests
- ✅ `TestUT_GR_32_01_QuickXORHash_BlockSize_Returns64Bytes` - BlockSize method test
- ✅ `TestUT_GR_33_01_QuickXORHash_Reset_RestoresInitialState` - Reset method test
- ✅ `TestUT_GR_09_01_QuickXORHash_ReaderInput_MatchesDirectCalculation` - Stream vs direct comparison

**Missing Tests** (Critical gaps):
- ❌ Comprehensive test with ALL Microsoft test vectors (0-256 bytes)
- ❌ Large file testing (>1MB, >100MB)
- ❌ Edge cases (empty files, single byte, boundary conditions)
- ❌ Performance benchmarks
- ❌ Integration test with actual OneDrive file verification

**Location of Test Vectors**:
- File: `internal/quickxorhash/quickxorhash_test.go`
- Contains 18 test vectors from Microsoft (sizes 0-256 bytes)
- Already being used in existing tests

**Critical Finding**: The existing tests in `internal/quickxorhash/quickxorhash_test.go` are **already comprehensive** and use Microsoft's official test vectors. However, there are TODO comments in `internal/graph/hash_functions_test.go` indicating incomplete tests.

### Action Plan for Issue #2

#### Step 1: Review Existing Test Coverage (30 minutes)

**Task 2.1**: Audit existing QuickXORHash tests
```bash
# List all QuickXORHash-related tests
grep -r "QuickXOR" --include="*_test.go" . | grep "^func Test"

# Run existing QuickXORHash tests
go test -v -run "QuickXOR" ./internal/quickxorhash/...
go test -v -run "QuickXOR" ./internal/graph/...
```

**Expected findings**:
- `internal/quickxorhash/quickxorhash_test.go` has comprehensive tests
- `internal/graph/hash_functions_test.go` has TODO placeholders

#### Step 2: Implement Missing Tests (4-6 hours)

**Task 2.2**: Complete `internal/graph/hash_functions_test.go` tests

**File**: `internal/graph/hash_functions_test.go`
**Line**: 185-196 (TODO comment)

**Test to implement**: `TestUT_GR_15_01_QuickXORHash_VariousInputs_ReturnsCorrectHash`

```go
func TestUT_GR_15_01_QuickXORHash_VariousInputs_ReturnsCorrectHash(t *testing.T) {
	fixture := framework.NewUnitTestFixture("QuickXORHashFixture")

	fixture.Use(t, func(t *testing.T, fixture interface{}) {
		assert := framework.NewAssert(t)

		// Test 1: Empty byte array
		emptyData := []byte("")
		emptyHash := QuickXORHash(&emptyData)
		assert.Equal("AAAAAAAAAAAAAAAAAAAAAAAAAAA=", emptyHash,
			"Empty data should produce known hash")

		// Test 2: Small content (Microsoft test vector)
		smallData := []byte{0x4a} // Single byte
		smallHash := QuickXORHash(&smallData)
		assert.Equal("SgAAAAAAAAAAAAAAAQAAAAAAAAA=", smallHash,
			"Single byte should match Microsoft test vector")

		// Test 3: Known content from Microsoft documentation
		// Using test vector for size 128
		testData := base64.StdEncoding.DecodeString(
			"ikwCorI7PKWz17EI50jZCGbV9JU2E8bXVfxNMg5zdmqSZ2NlsQPp0kqYIPjzwTg1MBtfWPg53k0h" +
			"0P2naJNEVgrqpoHTfV2b3pJ4m0zYPTJmUX4Bg/lOxcnCxAYKU29Y5F0U8Quz7ZXFBEweftXxJ7RS" +
			"4r6N7BzJrPsLhY7hgck=")
		expectedHash := "imAoFvCWlDn4yVw3/oq1PDbbm6U="
		actualHash := QuickXORHash(&testData)
		assert.Equal(expectedHash, actualHash,
			"128-byte content should match Microsoft test vector")

		// Test 4: Large file (1MB)
		largeData := make([]byte, 1024*1024)
		for i := range largeData {
			largeData[i] = byte(i % 256)
		}
		largeHash := QuickXORHash(&largeData)
		assert.NotEqual("", largeHash, "Large file should produce valid hash")
		assert.Equal(28, len(largeHash), "Hash should be base64 encoded (28 chars)")

		// Test 5: Verify hash consistency
		hash1 := QuickXORHash(&largeData)
		hash2 := QuickXORHash(&largeData)
		assert.Equal(hash1, hash2, "Same data should produce same hash")

		// Test 6: Verify hash uniqueness
		modifiedData := make([]byte, len(largeData))
		copy(modifiedData, largeData)
		modifiedData[0] ^= 0x01 // Flip one bit
		modifiedHash := QuickXORHash(&modifiedData)
		assert.NotEqual(largeHash, modifiedHash,
			"Different data should produce different hash")
	})
}
```

**Task 2.3**: Implement `TestUT_GR_16_01_QuickXORHashStream_VariousInputs_ReturnsCorrectHash`

**File**: `internal/graph/hash_functions_test.go`
**Line**: 211-222 (TODO comment)

```go
func TestUT_GR_16_01_QuickXORHashStream_VariousInputs_ReturnsCorrectHash(t *testing.T) {
	fixture := framework.NewUnitTestFixture("QuickXORHashStreamFixture")

	fixture.Use(t, func(t *testing.T, fixture interface{}) {
		assert := framework.NewAssert(t)

		// Test 1: Empty reader
		emptyReader := strings.NewReader("")
		emptyHash := QuickXORHashStream(emptyReader)
		assert.Equal("AAAAAAAAAAAAAAAAAAAAAAAAAAA=", emptyHash,
			"Empty stream should produce known hash")

		// Test 2: Small content
		smallReader := strings.NewReader("Hello, World!")
		smallHash := QuickXORHashStream(smallReader)

		// Verify it matches direct calculation
		smallData := []byte("Hello, World!")
		directHash := QuickXORHash(&smallData)
		assert.Equal(directHash, smallHash,
			"Stream hash should match direct hash")

		// Test 3: Large content (1MB)
		largeData := make([]byte, 1024*1024)
		for i := range largeData {
			largeData[i] = byte(i % 256)
		}
		largeReader := bytes.NewReader(largeData)
		streamHash := QuickXORHashStream(largeReader)
		directHash = QuickXORHash(&largeData)
		assert.Equal(directHash, streamHash,
			"Large stream hash should match direct hash")

		// Test 4: Verify seek position is reset
		testReader := strings.NewReader("Test content")
		testReader.Seek(5, 0) // Move to middle
		QuickXORHashStream(testReader)
		pos, _ := testReader.Seek(0, 1) // Get current position
		assert.Equal(int64(0), pos,
			"Seek position should be reset after hashing")

		// Test 5: Multiple reads produce same hash
		reader := strings.NewReader("Consistent content")
		hash1 := QuickXORHashStream(reader)
		hash2 := QuickXORHashStream(reader)
		assert.Equal(hash1, hash2,
			"Multiple reads should produce same hash")
	})
}
```

#### Step 3: Add Integration Test with Real Files (2 hours)

**Task 2.4**: Create integration test for file integrity verification

**File**: Create `internal/fs/quickxorhash_integration_test.go`

```go
package fs

import (
	"testing"
	"github.com/auriora/onemount/internal/graph"
	"github.com/auriora/onemount/internal/testutil/helpers"
)

// TestIT_FS_QuickXORHash_FileIntegrity tests QuickXORHash with real file operations
func TestIT_FS_QuickXORHash_FileIntegrity(t *testing.T) {
	fixture := helpers.SetupFSTestFixture(t, "QuickXORHashIntegrity",
		func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
			return NewFilesystem(auth, mountPoint, cacheTTL)
		})
	defer fixture.Teardown(t)

	fs := fixture.GetFixture(t).(*helpers.FSTestFixture).Filesystem

	// Test 1: Upload file and verify hash
	testContent := []byte("Test file content for QuickXORHash verification")
	expectedHash := graph.QuickXORHash(&testContent)

	// Create and upload file
	// ... (implementation details)

	// Download file and verify hash matches
	// ... (implementation details)

	// Test 2: Large file (>100MB) streaming
	// ... (implementation details)
}
```

#### Step 4: Add Performance Benchmarks (1 hour)

**Task 2.5**: Add benchmarks to `internal/quickxorhash/quickxorhash_test.go`

```go
func BenchmarkQuickXORHash_Small(b *testing.B) {
	data := make([]byte, 1024) // 1KB
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Sum(data)
	}
}

func BenchmarkQuickXORHash_Medium(b *testing.B) {
	data := make([]byte, 1024*1024) // 1MB
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Sum(data)
	}
}

func BenchmarkQuickXORHash_Large(b *testing.B) {
	data := make([]byte, 100*1024*1024) // 100MB
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Sum(data)
	}
}
```

#### Step 5: Run and Verify Tests (1 hour)

```bash
# Run all QuickXORHash tests
go test -v -run "QuickXOR" ./internal/quickxorhash/...
go test -v -run "QuickXOR" ./internal/graph/...
go test -v -run "QuickXOR" ./internal/fs/...

# Run benchmarks
go test -bench="QuickXOR" ./internal/quickxorhash/...

# Verify test coverage
go test -cover -run "QuickXOR" ./internal/quickxorhash/...
go test -cover -run "QuickXOR" ./internal/graph/...
```

#### Step 6: Update Documentation (30 minutes)

**Task 2.6**: Update test documentation

Files to update:
- `docs/testing/test-audit-report.md` - Mark QuickXORHash tests as complete
- `docs/0-project-management/todo_comments_summary.md` - Remove TODO entries
- `docs/testing/running-tests.md` - Add QuickXORHash test examples

**Total Time for Issue #2**: 8-10 hours (1-1.5 days)

---

## Combined Timeline and Milestones

### Day 1: Test Labeling + QuickXORHash Planning
**Duration**: 6-8 hours

**Morning** (3-4 hours):
- ✅ Label remaining 11 tests in config and graph packages
- ✅ Verify all tests compile
- ✅ Run unit tests to confirm labeling works
- ✅ Update documentation

**Afternoon** (3-4 hours):
- ✅ Audit existing QuickXORHash test coverage
- ✅ Design comprehensive test cases
- ✅ Set up test fixtures and utilities
- ✅ Begin implementing missing tests

**Deliverables**:
- All 752 tests properly labeled
- QuickXORHash test plan documented
- 50% of QuickXORHash tests implemented

### Day 2: QuickXORHash Implementation
**Duration**: 8 hours

**Morning** (4 hours):
- ✅ Complete `TestUT_GR_15_01` implementation
- ✅ Complete `TestUT_GR_16_01` implementation
- ✅ Add edge case tests
- ✅ Run and verify all unit tests pass

**Afternoon** (4 hours):
- ✅ Create integration test file
- ✅ Implement file integrity verification test
- ✅ Add performance benchmarks
- ✅ Run full test suite

**Deliverables**:
- All QuickXORHash tests implemented
- Integration tests passing
- Benchmarks showing acceptable performance

### Day 3: Verification and Documentation
**Duration**: 4-6 hours

**Morning** (2-3 hours):
- ✅ Run full test suite in Docker
- ✅ Verify test coverage metrics
- ✅ Fix any failing tests
- ✅ Run benchmarks and document results

**Afternoon** (2-3 hours):
- ✅ Update all documentation
- ✅ Remove TODO comments
- ✅ Create release notes
- ✅ Final verification

**Deliverables**:
- All tests passing
- Documentation updated
- Ready for v1.0 release

---

## Success Criteria

### Issue #1: Test Labeling
- [ ] All 752 tests have proper naming conventions
- [ ] `grep` command returns 0 unlabeled tests (excluding TestMain)
- [ ] Unit tests run without auth tokens
- [ ] Integration tests skip gracefully when auth unavailable
- [ ] Documentation updated with 100% completion status

### Issue #2: QuickXORHash Testing
- [ ] All TODO comments removed from hash test files
- [ ] Comprehensive tests with Microsoft test vectors
- [ ] Edge case tests (empty, single byte, large files)
- [ ] Integration test with real file operations
- [ ] Performance benchmarks documented
- [ ] Test coverage >90% for QuickXORHash code
- [ ] All tests passing in CI/CD

### Overall v1.0 Readiness
- [ ] Both blocking issues resolved
- [ ] Full test suite passing
- [ ] Documentation updated
- [ ] Release notes prepared
- [ ] Package validation complete
- [ ] Security review complete

---

## Risk Mitigation

### Risk 1: Test Labeling Breaks Existing Tests
**Likelihood**: Low
**Impact**: Medium
**Mitigation**:
- Compile after each batch of changes
- Run tests incrementally
- Keep git commits small for easy rollback
- Test in Docker environment first

### Risk 2: QuickXORHash Tests Reveal Implementation Bugs
**Likelihood**: Medium
**Impact**: High
**Mitigation**:
- Existing implementation already uses Microsoft's code
- Test vectors from Microsoft documentation
- If bugs found, fix before v1.0 release
- Document any known limitations

### Risk 3: Integration Tests Fail in CI/CD
**Likelihood**: Low
**Impact**: Medium
**Mitigation**:
- Test locally with Docker first
- Verify auth token handling
- Add proper skip logic for missing auth
- Document CI/CD requirements

### Risk 4: Timeline Overruns
**Likelihood**: Medium
**Impact**: Low
**Mitigation**:
- Conservative time estimates (3-5 days)
- Can defer non-critical tests to v1.1
- Focus on blocking issues only
- Parallel work where possible

---

## Execution Checklist

### Pre-Execution
- [ ] Create feature branch: `fix/v1.0-blocking-issues`
- [ ] Backup current state
- [ ] Verify Docker environment working
- [ ] Ensure auth tokens available for testing
- [ ] Review this action plan

### During Execution
- [ ] Commit after each major step
- [ ] Run tests after each change
- [ ] Document any issues encountered
- [ ] Update progress in this document
- [ ] Communicate blockers immediately

### Post-Execution
- [ ] Run full test suite in Docker
- [ ] Verify CI/CD pipeline passes
- [ ] Update all documentation
- [ ] Create pull request
- [ ] Request code review
- [ ] Merge to main after approval

---

## Commands Reference

### Test Labeling Commands

```bash
# Check for unlabeled tests
grep -r "^func Test[A-Z]" --include="*_test.go" . | \
  grep -v "TestUT_" | grep -v "TestIT_" | \
  grep -v "TestProperty" | grep -v "TestSystemST_" | \
  grep -v "TestMain"

# Run unit tests only
go test -v -run "^TestUT_" ./...

# Run integration tests only
go test -v -run "^TestIT_" ./...

# Compile all tests
go build ./...
```

### QuickXORHash Testing Commands

```bash
# Run QuickXORHash tests
go test -v -run "QuickXOR" ./internal/quickxorhash/...
go test -v -run "QuickXOR" ./internal/graph/...

# Run with coverage
go test -cover -run "QuickXOR" ./internal/quickxorhash/...

# Run benchmarks
go test -bench="QuickXOR" ./internal/quickxorhash/...

# Run in Docker
docker compose -f docker/compose/docker-compose.test.yml run --rm \
  test-runner go test -v -run "QuickXOR" ./...
```

### Full Test Suite Commands

```bash
# Run all tests in Docker
docker compose -f docker/compose/docker-compose.test.yml run --rm all

# Run with auth
docker compose -f docker/compose/docker-compose.test.yml \
  -f docker/compose/docker-compose.auth.yml run --rm all

# Generate coverage report
./scripts/dev test coverage --threshold-line 85
```

---

## Contact and Escalation

### Questions or Issues
- Review this document first
- Check existing documentation in `docs/testing/`
- Consult `docs/testing/running-tests.md` for test execution
- Consult `docs/testing/test-fixtures.md` for test setup

### Escalation Path
1. Document the issue in this file
2. Check if it's a known risk (see Risk Mitigation section)
3. Determine if it blocks v1.0 release
4. If blocking: escalate immediately
5. If non-blocking: defer to v1.1

---

## Appendix A: File Locations

### Files to Modify

**Test Labeling**:
- `internal/config/mounts_test.go` - 4 tests to label
- `internal/graph/oauth2_account_storage_test.go` - 7 tests to label

**QuickXORHash Testing**:
- `internal/graph/hash_functions_test.go` - Implement 2 TODO tests
- `internal/fs/quickxorhash_integration_test.go` - Create new file
- `internal/quickxorhash/quickxorhash_test.go` - Add benchmarks

**Documentation**:
- `docs/testing/test-labeling-summary.md` - Update completion status
- `docs/testing/test-audit-report.md` - Mark QuickXORHash complete
- `docs/0-project-management/todo_comments_summary.md` - Remove TODOs
- `docs/testing/running-tests.md` - Add examples

### Reference Files

**Test Vectors**:
- `internal/quickxorhash/quickxorhash_test.go` - Microsoft test vectors

**Existing Tests**:
- `internal/quickxorhash/quickxorhash_test.go` - Comprehensive unit tests
- `internal/graph/hashes_test.go` - Hash function tests
- `internal/graph/hashes.go` - Hash implementations

**Documentation**:
- `docs/testing/running-tests.md` - Test execution guide
- `docs/testing/test-fixtures.md` - Test fixture guide
- `docs/testing/test-audit-report.md` - Test audit results

---

## Appendix B: Expected Test Output

### After Test Labeling

```bash
$ grep -r "^func Test[A-Z]" --include="*_test.go" . | \
  grep -v "TestUT_" | grep -v "TestIT_" | \
  grep -v "TestProperty" | grep -v "TestSystemST_" | \
  grep -v "TestMain" | wc -l
0
```

### After QuickXORHash Testing

```bash
$ go test -v -run "QuickXOR" ./internal/quickxorhash/...
=== RUN   TestUT_GR_29_01_QuickXORHash_Sum_CalculatesCorrectHashes
--- PASS: TestUT_GR_29_01_QuickXORHash_Sum_CalculatesCorrectHashes (0.01s)
=== RUN   TestUT_GR_30_01_QuickXORHash_WriteByBlocks_CalculatesCorrectHashes
--- PASS: TestUT_GR_30_01_QuickXORHash_WriteByBlocks_CalculatesCorrectHashes (0.05s)
=== RUN   TestUT_GR_32_01_QuickXORHash_BlockSize_Returns64Bytes
--- PASS: TestUT_GR_32_01_QuickXORHash_BlockSize_Returns64Bytes (0.00s)
=== RUN   TestUT_GR_33_01_QuickXORHash_Reset_RestoresInitialState
--- PASS: TestUT_GR_33_01_QuickXORHash_Reset_RestoresInitialState (0.00s)
PASS
ok      github.com/auriora/onemount/internal/quickxorhash    0.123s
```

---

**End of Action Plan**

**Next Steps**: Begin execution with Day 1 tasks. Create feature branch and start with test labeling.


