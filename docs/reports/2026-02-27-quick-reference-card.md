# OneMount v1.0 Blocking Issues - Quick Reference Card
**Date**: February 27, 2026  
**Status**: READY TO EXECUTE  
**Estimated Time**: 3-5 days

---

## 🎯 Quick Summary

**Two blocking issues** prevent v1.0 release:
1. ✅ **92% DONE** - Test labeling (11 tests remaining, 2-3 hours)
2. ⚠️ **PARTIAL** - QuickXORHash testing (8-10 hours)

**Total effort**: 10-13 hours (1.5-2 days of focused work)

---

## 📋 Issue #1: Test Labeling (2-3 hours)

### What to Do
Label 11 remaining tests with `TestUT_` prefix:
- 4 tests in `internal/config/mounts_test.go`
- 7 tests in `internal/graph/oauth2_account_storage_test.go`

### Quick Commands

```bash
# 1. Label config tests (copy-paste all 4 lines)
sed -i 's/func TestMountsRegistry(/func TestUT_Config_MountsRegistry(/g' internal/config/mounts_test.go
sed -i 's/func TestMountsRegistry_MultipleAccounts(/func TestUT_Config_MountsRegistry_MultipleAccounts(/g' internal/config/mounts_test.go
sed -i 's/func TestMountsRegistry_EmptyConfigDir(/func TestUT_Config_MountsRegistry_EmptyConfigDir(/g' internal/config/mounts_test.go
sed -i 's/func TestMountsRegistry_UpdateMount(/func TestUT_Config_MountsRegistry_UpdateMount(/g' internal/config/mounts_test.go

# 2. Label graph tests (copy-paste all 9 lines)
sed -i 's/func TestHashAccount(/func TestUT_Graph_HashAccount(/g' internal/graph/oauth2_account_storage_test.go
sed -i 's/func TestHashAccountStability(/func TestUT_Graph_HashAccountStability(/g' internal/graph/oauth2_account_storage_test.go
sed -i 's/func TestHashAccountCollisionResistance(/func TestUT_Graph_HashAccountCollisionResistance(/g' internal/graph/oauth2_account_storage_test.go
sed -i 's/func TestGetAuthTokensPathByAccount(/func TestUT_Graph_GetAuthTokensPathByAccount(/g' internal/graph/oauth2_account_storage_test.go
sed -i 's/func TestFindAuthTokens(/func TestUT_Graph_FindAuthTokens(/g' internal/graph/oauth2_account_storage_test.go
sed -i 's/func TestMigrateTokens(/func TestUT_Graph_MigrateTokens(/g' internal/graph/oauth2_account_storage_test.go
sed -i 's/func TestMigrateTokensPermissions(/func TestUT_Graph_MigrateTokensPermissions(/g' internal/graph/oauth2_account_storage_test.go
sed -i 's/func TestFileExists(/func TestUT_Graph_FileExists(/g' internal/graph/oauth2_account_storage_test.go
sed -i 's/func TestAuthenticateWithAccountStorage_Migration(/func TestUT_Graph_AuthenticateWithAccountStorage_Migration(/g' internal/graph/oauth2_account_storage_test.go

# 3. Verify (should return 0)
grep -r "^func Test[A-Z]" --include="*_test.go" . | grep -v "TestUT_" | grep -v "TestIT_" | grep -v "TestProperty" | grep -v "TestSystemST_" | grep -v "TestMain" | wc -l

# 4. Test
go build ./...
go test -v -run "^TestUT_" ./internal/config/... ./internal/graph/...
```

### Success Criteria
- ✅ Grep returns 0 unlabeled tests
- ✅ All tests compile
- ✅ Unit tests run without auth

---

## 📋 Issue #2: QuickXORHash Testing (8-10 hours)

### What to Do
1. Implement 2 TODO tests in `internal/graph/hash_functions_test.go`
2. Add integration test for file integrity
3. Add performance benchmarks
4. Verify all tests pass

### Files to Edit

**File 1**: `internal/graph/hash_functions_test.go` (lines 185-196)
- Remove TODO comment
- Implement `TestUT_GR_15_01_QuickXORHash_VariousInputs_ReturnsCorrectHash`
- Test with Microsoft test vectors, empty data, large files

**File 2**: `internal/graph/hash_functions_test.go` (lines 211-222)
- Remove TODO comment
- Implement `TestUT_GR_16_01_QuickXORHashStream_VariousInputs_ReturnsCorrectHash`
- Test stream hashing, seek position reset, consistency

**File 3**: Create `internal/fs/quickxorhash_integration_test.go`
- Add integration test for file upload/download integrity
- Test with real OneDrive operations

**File 4**: `internal/quickxorhash/quickxorhash_test.go`
- Add benchmarks for small, medium, large files

### Quick Test Commands

```bash
# Run QuickXORHash tests
go test -v -run "QuickXOR" ./internal/quickxorhash/...
go test -v -run "QuickXOR" ./internal/graph/...

# Run with coverage
go test -cover -run "QuickXOR" ./internal/quickxorhash/...

# Run benchmarks
go test -bench="QuickXOR" ./internal/quickxorhash/...

# Full test in Docker
docker compose -f docker/compose/docker-compose.test.yml run --rm unit
```

### Success Criteria
- ✅ All TODO comments removed
- ✅ Tests with Microsoft test vectors
- ✅ Edge cases covered
- ✅ Integration test passing
- ✅ Benchmarks documented
- ✅ Coverage >90%

---

## 📅 3-Day Timeline

### Day 1 (6-8 hours)
**Morning**: Test labeling (3-4 hours)
- Label 11 tests
- Verify compilation
- Run unit tests
- Update docs

**Afternoon**: QuickXORHash planning (3-4 hours)
- Audit existing tests
- Design test cases
- Begin implementation

### Day 2 (8 hours)
**Morning**: QuickXORHash unit tests (4 hours)
- Implement TestUT_GR_15_01
- Implement TestUT_GR_16_01
- Add edge cases

**Afternoon**: Integration & benchmarks (4 hours)
- Create integration test
- Add benchmarks
- Run full suite

### Day 3 (4-6 hours)
**Morning**: Verification (2-3 hours)
- Run in Docker
- Fix failures
- Check coverage

**Afternoon**: Documentation (2-3 hours)
- Update docs
- Remove TODOs
- Prepare release

---

## ✅ Final Checklist

### Before Starting
- [ ] Create branch: `fix/v1.0-blocking-issues`
- [ ] Verify Docker working
- [ ] Auth tokens available

### Issue #1 Complete
- [ ] 11 tests labeled
- [ ] Grep returns 0
- [ ] Tests compile
- [ ] Unit tests pass

### Issue #2 Complete
- [ ] 2 TODO tests implemented
- [ ] Integration test created
- [ ] Benchmarks added
- [ ] All tests passing
- [ ] Coverage >90%

### Ready for v1.0
- [ ] Full test suite passing
- [ ] Documentation updated
- [ ] TODOs removed
- [ ] CI/CD passing
- [ ] Release notes ready

---

## 🚀 Quick Start

```bash
# 1. Create branch
git checkout -b fix/v1.0-blocking-issues

# 2. Run test labeling commands (see Issue #1 above)

# 3. Verify
grep -r "^func Test[A-Z]" --include="*_test.go" . | \
  grep -v "TestUT_" | grep -v "TestIT_" | \
  grep -v "TestProperty" | grep -v "TestSystemST_" | \
  grep -v "TestMain" | wc -l

# 4. Commit
git add .
git commit -m "fix(tests): label remaining 11 tests with proper prefixes"

# 5. Move to QuickXORHash testing (see detailed plan)
```

---

## 📚 Reference Documents

- **Detailed Plan**: `docs/reports/2026-02-27-blocking-issues-action-plan.md`
- **Full Review**: `docs/reports/2026-02-27-comprehensive-project-review.md`
- **Executive Summary**: `docs/reports/2026-02-27-project-review-executive-summary.md`
- **Test Guide**: `docs/testing/running-tests.md`
- **Test Fixtures**: `docs/testing/test-fixtures.md`

---

**Ready to execute!** Start with test labeling (2-3 hours), then move to QuickXORHash testing.

