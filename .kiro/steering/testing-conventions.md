---
inclusion: always
---

# Testing Conventions

**Priority**: 25  
**Scope**: tests/**, internal/**/*_test.go  
**Description**: Standardized testing format and policy for all projects.

## Testing Guidelines

- **Test File Placement**: Always place tests next to the code they exercise when practical (e.g., `internal/module/module_test.go` for Go).
- **Consistent Test File Naming**: Follow Go conventions: `*_test.go` for unit tests, `*_integration_test.go` for integration tests.
- **Preferred Test Runner**: Use Go's built-in test runner (`go test`) for this project.

## Docker Test Environment (OneMount Project)

**CRITICAL**: All tests for the OneMount project MUST be run inside Docker containers. Never run tests directly on the host system.

### Why Docker is Required

1. **FUSE Dependencies**: Tests require FUSE3 device access with specific capabilities
2. **Isolation**: Prevents test artifacts from polluting the host system
3. **Reproducibility**: Ensures consistent environment across all developers and CI/CD
4. **Security**: Test credentials and OneDrive access are isolated in containers
5. **Dependencies**: Specific versions of Go (1.24.2), Python (3.12), and system libraries

### Docker Test Commands

**Build images first (if not already built):**

```bash
# Build all development images
./docker/scripts/build-images.sh dev

# Or build just test runner
./docker/scripts/build-images.sh test-runner
```

**The test runner supports two modes:**

1. **Helper Commands** - Predefined test suites (unit, integration, system, all, coverage, shell)
2. **Pass-Through Mode** - Any other command is executed directly after environment setup

**Run predefined test suites (helper commands):**

```bash
# Unit tests (no FUSE required, fast)
docker compose -f docker/compose/docker-compose.test.yml run --rm unit-tests

# Integration tests (requires FUSE, moderate speed)
docker compose -f docker/compose/docker-compose.test.yml run --rm integration-tests

# System tests (requires auth tokens, slow)
docker compose -f docker/compose/docker-compose.test.yml run --rm system-tests

# All tests with coverage
docker compose -f docker/compose/docker-compose.test.yml run --rm test-runner all

# Interactive debugging shell
docker compose -f docker/compose/docker-compose.test.yml run --rm shell
```

**Run specific tests (pass-through mode):**

```bash
# IMPORTANT: Use test-runner service (not integration-tests) for pass-through mode
# The specialized services (integration-tests, unit-tests) have default commands

# Run specific test pattern - pass-through mode automatically sets up environment
docker compose -f docker/compose/docker-compose.test.yml run --rm \
  test-runner go test -v -run TestIT_FS_ETag ./internal/fs

# Run tests with custom timeout
docker compose -f docker/compose/docker-compose.test.yml run --rm \
  test-runner go test -v -timeout 10m -run TestPattern ./internal/fs

# Run tests with race detector
docker compose -f docker/compose/docker-compose.test.yml run --rm \
  test-runner go test -v -race ./internal/...

# Run benchmarks
docker compose -f docker/compose/docker-compose.test.yml run --rm \
  test-runner go test -v -bench=. -run=^$ ./internal/fs
```

**Key Points for AI Agents:**
- ✅ **Use specialized services** for full test suites: `unit-tests`, `integration-tests`, `system-tests`
- ✅ **Use test-runner service** for pass-through mode: `test-runner go test -v -run TestPattern ./path`
- ❌ **Don't use specialized services for pass-through** - they have default commands that will conflict
- ❌ **Don't use** `--entrypoint /bin/bash` workarounds - pass-through mode handles this
```

### Test Environment Details

- **Images**: 
  - `onemount-base:latest` (1.49GB) - Base image with build tools
  - `onemount-test-runner:latest` (2.21GB) - Test execution environment
- **Build Command**: `./docker/scripts/build-images.sh test-runner`
- **Image Locations**:
  - Base: `packaging/docker/Dockerfile.base`
  - Test Runner: `docker/images/test-runner/Dockerfile`
- **Workspace**: Mounted at `/workspace` with read-write access
- **Test Artifacts**: Output to `test-artifacts/` (mounted from host)
- **Auth Tokens**: Located at `test-artifacts/.auth_tokens.json` (gitignored)
- **FUSE Device**: `/dev/fuse` with SYS_ADMIN capability
- **Resources**: 4GB RAM / 2 CPUs (unit), 6GB RAM / 4 CPUs (system)

### Docker Environment Modifications

When making changes to improve the Docker test environment:

1. **Update Dockerfiles**: 
   - Base image: `packaging/docker/Dockerfile.base`
   - Test runner: `docker/images/test-runner/Dockerfile`
2. **Rebuild Images**: Run `./docker/scripts/build-images.sh test-runner`
3. **Update Documentation**: Update `docs/TEST_SETUP.md` and `docs/testing/docker-test-environment.md`
4. **Test Changes**: Verify all test types still work (unit, integration, system)
5. **Document Rationale**: Explain why the change improves the test environment

### Common Docker Test Patterns

```bash
# Run tests and keep container for debugging
docker compose -f docker/compose/docker-compose.test.yml run --rm \
  --entrypoint /bin/bash shell -c "go test -v ./... || bash"

# Run tests with verbose output
docker compose -f docker/compose/docker-compose.test.yml run --rm \
  --entrypoint /bin/bash shell -c "go test -v -count=1 ./internal/..."

# Run tests with coverage report
docker compose -f docker/compose/docker-compose.test.yml run --rm test-runner coverage

# Check test environment
docker compose -f docker/compose/docker-compose.test.yml run --rm \
  --entrypoint /bin/bash shell -c "ls -l /dev/fuse && go version && python3 --version"
```

### Troubleshooting Docker Tests

If tests fail in Docker:

1. **Check FUSE device**: `docker compose -f docker/compose/docker-compose.test.yml run --rm shell` then `ls -l /dev/fuse`
2. **Verify images are up-to-date**: Rebuild with `docker compose -f docker/compose/docker-compose.build.yml build --no-cache`
3. **Check auth tokens**: Verify `test-artifacts/.auth_tokens.json` exists and is valid
4. **Review logs**: Check `test-artifacts/logs/` for detailed error messages
5. **Interactive debugging**: Use `docker compose -f docker/compose/docker-compose.test.yml run --rm shell`

### References

- **Setup Guide**: `docs/TEST_SETUP.md` - Comprehensive test environment setup
- **Docker Guide**: `docs/testing/docker-test-environment.md` - Docker-specific details
- **Task Summary**: `docs/TASK_1_SUMMARY.md` - Docker environment validation results

## Test Organization

- Unit tests for `src/TimeLocker/module.py` → `tests/TimeLocker/test_module.py` (follow repo convention)
- Integration tests that exercise multiple modules → `tests/TimeLocker/integration/test_feature-name.py`
- Place test utilities and fixtures in appropriate `conftest.py` files

## PR Checklist Additions

When making changes that affect tests, ensure the following are considered:

-   [ ] Added/updated unit tests for changed behavior
-   [ ] Added/updated minimal integration or smoke tests if public behavior changed
-   [ ] Verified all tests pass with the changes
-   [ ] Updated test documentation if test structure or approach changed

## Testing Best Practices

- Write tests that focus on behavior, not implementation details
- Use descriptive test names that explain what is being tested
- Follow the Arrange-Act-Assert pattern for test structure
- Mock external dependencies appropriately
- Ensure tests are deterministic and can run in any order