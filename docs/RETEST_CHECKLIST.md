# Re-Verification Checklist with Real OneDrive

**Status**: Ready to Execute  
**Authentication**: ✅ Working with saved credentials  
**Last Updated**: 2025-11-12

## Quick Reference Checklist

### High Priority (Must Complete)

- [ ] **Phase 4, Task 5.7**: Mounting Integration Tests (2-3h)
  ```bash
  docker compose -f docker/compose/docker-compose.test.yml run --rm \
    integration-tests go test -v -run TestIT_FS_Mount ./internal/fs
  ```

- [ ] **Phase 5**: ETag Validation Tests (3-4h)
  ```bash
  docker compose -f docker/compose/docker-compose.test.yml run --rm \
    integration-tests go test -v -run TestIT_FS_ETag ./internal/fs
  ```

- [ ] **Phase 5**: Conflict Detection Verification (2-3h)
  ```bash
  docker compose -f docker/compose/docker-compose.test.yml run --rm \
    integration-tests go test -v -run TestIT_FS.*Conflict ./internal/fs
  ```

- [ ] **Phase 14**: E2E Test - Complete User Workflow (1-2h)
  ```bash
  docker compose -f docker/compose/docker-compose.test.yml run --rm \
    -e RUN_E2E_TESTS=1 \
    system-tests go test -v -run TestE2E_17_01 ./internal/fs
  ```

- [ ] **Phase 14**: E2E Test - Multi-File Operations (1-2h)
  ```bash
  docker compose -f docker/compose/docker-compose.test.yml run --rm \
    -e RUN_E2E_TESTS=1 \
    system-tests go test -v -run TestE2E_17_02 ./internal/fs
  ```

### Medium Priority (Should Complete)

- [ ] **Phase 4**: Directory Deletion with Real Server (1-2h)
  ```bash
  docker compose -f docker/compose/docker-compose.test.yml run --rm \
    integration-tests go test -v -run TestIT_FS_FileWrite ./internal/fs
  ```

- [ ] **Phase 5**: Large File Operations (4-6h)
  ```bash
  docker compose -f docker/compose/docker-compose.test.yml run --rm \
    -e RUN_LONG_TESTS=1 \
    system-tests go test -v -timeout 60m -run TestSYS_LargeFile ./internal/fs
  ```

- [ ] **Phase 8**: Cache Management Manual Verification (2-3h)
  - Set short expiration time
  - Monitor cache cleanup
  - Verify statistics with large datasets

- [ ] **Phase 10**: Manual Test Scripts (2-3h)
  ```bash
  docker compose -f docker/compose/docker-compose.test.yml run --rm shell
  # Then run:
  ./tests/manual/test_file_status_updates.sh
  ./tests/manual/test_dbus_integration.sh
  ./tests/manual/test_dbus_fallback.sh
  ```

- [ ] **Phase 14**: E2E Test - Long-Running Operations (2-3h)
  ```bash
  docker compose -f docker/compose/docker-compose.test.yml run --rm \
    -e RUN_E2E_TESTS=1 \
    -e RUN_LONG_TESTS=1 \
    system-tests go test -v -timeout 60m -run TestE2E_17_03 ./internal/fs
  ```

- [ ] **Phase 14**: E2E Test - Stress Scenarios (1-2h)
  ```bash
  docker compose -f docker/compose/docker-compose.test.yml run --rm \
    -e RUN_E2E_TESTS=1 \
    -e RUN_STRESS_TESTS=1 \
    system-tests go test -v -timeout 30m -run TestE2E_17_04 ./internal/fs
  ```

### Low Priority (Nice to Have)

- [ ] **Phase 13**: Integration Tests with Real OneDrive (2-3h)
  ```bash
  docker compose -f docker/compose/docker-compose.test.yml run --rm \
    integration-tests go test -v -run TestIT_COMPREHENSIVE ./internal/fs
  ```

## Prerequisites Checklist

- [ ] Auth tokens file exists: `test-artifacts/.auth_tokens.json`
- [ ] Tokens are valid (not expired)
- [ ] Environment variable set: `ONEMOUNT_AUTH_PATH=test-artifacts/.auth_tokens.json`
- [ ] Docker images built: `./docker/scripts/build-images.sh test-runner`
- [ ] FUSE device available in containers
- [ ] Test OneDrive account prepared with test directory
- [ ] At least 5GB free space in OneDrive

## Progress Tracking

**High Priority**: 0/5 completed  
**Medium Priority**: 0/6 completed  
**Low Priority**: 0/1 completed  
**Total**: 0/12 completed (0%)

## Time Estimates

- **High Priority**: 13-18 hours
- **Medium Priority**: 9-13 hours  
- **Low Priority**: 2-3 hours
- **Total**: 24-34 hours (3-4 days)

## Quick Start

1. Verify authentication is working:
   ```bash
   ls -la test-artifacts/.auth_tokens.json
   ```

2. Build Docker images if needed:
   ```bash
   ./docker/scripts/build-images.sh test-runner
   ```

3. Start with first high-priority test:
   ```bash
   docker compose -f docker/compose/docker-compose.test.yml run --rm \
     integration-tests go test -v -run TestIT_FS_Mount ./internal/fs
   ```

4. Document results after each test in verification tracking

## Notes

- Tests can be run in any order within priority groups
- Some tests can run in parallel to save time
- Document any failures immediately
- Update `docs/verification-tracking.md` after each phase

---

**See Also**: `docs/verification-tasks-requiring-real-onedrive.md` for detailed information
