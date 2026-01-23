---
inclusion: always
---

# Testing Conventions

**Priority**: 25  
**Scope**: tests/**, internal/**/*_test.go  
**Description**: Standardized testing format and policy for all projects.

---

## 🚨 CRITICAL - READ THIS FIRST 🚨

**ALL TESTS MUST BE RUN IN DOCKER CONTAINERS - NO EXCEPTIONS**

**NEVER run tests directly on the host system with `go test`**

**ALWAYS use the timeout wrapper script for integration tests**

If you are an AI agent and you see yourself about to run `go test` directly, STOP and use Docker instead.

---

## AI Agent Test Execution Checklist

Before running ANY test, verify ALL of these:

- [ ] Am I using Docker? (`docker compose -f ...`)
- [ ] Am I using the timeout wrapper for integration tests? (`./scripts/timeout-test-wrapper.sh`)
- [ ] Am I in the correct working directory? (workspace root, not a subdirectory)
- [ ] Have I included the auth override for integration/system tests? (`-f docker/compose/docker-compose.auth.yml`)
- [ ] Am I NOT using the `cd` command? (it's forbidden - use `cwd` parameter instead)

**If ANY checkbox is unchecked, you are doing it wrong!**

---

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

**Set up authentication (required for integration/system tests):**

```bash
# Set up reference-based authentication (REQUIRED for tests with auth)
./scripts/setup-auth-reference.sh

# This creates:
# - docker/compose/docker-compose.auth.yml (Docker override)
# - .env.auth (environment configuration)
# - References canonical token location (no copying/symlinking)
```

**The test runner supports two modes:**

1. **Helper Commands** - Predefined test suites (unit, integration, system, all, coverage, shell)
2. **Pass-Through Mode** - Any other command is executed directly after environment setup

**Run predefined test suites (helper commands):**

```bash
# Unit tests (no FUSE required, fast)
# Verbose output is automatically redirected to test-artifacts/logs/
docker compose -f docker/compose/docker-compose.test.yml run --rm unit-tests

# Integration tests (requires FUSE and auth, moderate speed)
docker compose -f docker/compose/docker-compose.test.yml \
  -f docker/compose/docker-compose.auth.yml run --rm integration-tests

# System tests (requires auth tokens, slow)
docker compose -f docker/compose/docker-compose.test.yml \
  -f docker/compose/docker-compose.auth.yml run --rm system-tests

# All tests with coverage
docker compose -f docker/compose/docker-compose.test.yml \
  -f docker/compose/docker-compose.auth.yml run --rm test-runner all

# Interactive debugging shell
docker compose -f docker/compose/docker-compose.test.yml \
  -f docker/compose/docker-compose.auth.yml run --rm shell

# Disable log file redirection (show all output on console)
docker compose -f docker/compose/docker-compose.test.yml \
  -f docker/compose/docker-compose.auth.yml run --rm \
  -e ONEMOUNT_LOG_TO_FILE=false integration-tests
```

**Run specific tests (pass-through mode):**

```bash
# IMPORTANT: Use test-runner service (not integration-tests) for pass-through mode
# The specialized services (integration-tests, unit-tests) have default commands

# Run specific test pattern with timeout protection
./scripts/timeout-test-wrapper.sh "TestIT_FS_ETag" 60

# Run specific test pattern - pass-through mode automatically sets up environment
docker compose -f docker/compose/docker-compose.test.yml \
  -f docker/compose/docker-compose.auth.yml run --rm \
  test-runner go test -v -run TestIT_FS_ETag ./internal/fs

# Run tests with custom timeout
docker compose -f docker/compose/docker-compose.test.yml \
  -f docker/compose/docker-compose.auth.yml run --rm \
  test-runner go test -v -timeout 10m -run TestPattern ./internal/fs

# Run tests with race detector
docker compose -f docker/compose/docker-compose.test.yml \
  -f docker/compose/docker-compose.auth.yml run --rm \
  test-runner go test -v -race ./internal/...

# Run benchmarks
docker compose -f docker/compose/docker-compose.test.yml \
  -f docker/compose/docker-compose.auth.yml run --rm \
  test-runner go test -v -bench=. -run=^$ ./internal/fs
```

**Key Points for AI Agents:**
- ✅ **Use specialized services** for full test suites: `unit-tests`, `integration-tests`, `system-tests`
- ✅ **Use test-runner service** for pass-through mode: `test-runner go test -v -run TestPattern ./path`
- ✅ **Always include auth override** for integration/system tests: `-f docker/compose/docker-compose.auth.yml`
- ✅ **Use timeout wrapper** for potentially hanging tests: `./scripts/timeout-test-wrapper.sh "TestPattern" 60`
- ❌ **Don't use specialized services for pass-through** - they have default commands that will conflict
- ❌ **Don't use** `--entrypoint /bin/bash` workarounds - pass-through mode handles this

### ❌ WRONG - DO NOT DO THIS

```bash
# WRONG: Running tests directly on host
go test -v -run TestIT_FS ./internal/fs

# WRONG: Using cd command
cd OneMount && go test ...

# WRONG: Not using timeout wrapper for integration tests
docker compose -f docker/compose/docker-compose.test.yml run --rm \
  test-runner go test -run TestIT_FS_30_08 ./internal/fs

# WRONG: Missing auth override for integration tests
docker compose -f docker/compose/docker-compose.test.yml run --rm integration-tests
```

### ✅ CORRECT - DO THIS

```bash
# CORRECT: Using timeout wrapper for integration tests
./scripts/timeout-test-wrapper.sh "TestIT_FS_30_08" 60

# CORRECT: Using Docker directly with auth override
docker compose -f docker/compose/docker-compose.test.yml \
  -f docker/compose/docker-compose.auth.yml run --rm \
  test-runner go test -v -run TestIT_FS_30_08 ./internal/fs

# CORRECT: Unit tests in Docker
docker compose -f docker/compose/docker-compose.test.yml run --rm unit-tests

# CORRECT: Integration tests with auth
docker compose -f docker/compose/docker-compose.test.yml \
  -f docker/compose/docker-compose.auth.yml run --rm integration-tests
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
- **Auth Reference System**: 
  - Setup: `./scripts/setup-auth-reference.sh`
  - Override: `docker/compose/docker-compose.auth.yml`
  - Environment: `.env.auth`
  - Canonical location referenced (no copying/symlinking)
- **FUSE Device**: `/dev/fuse` with SYS_ADMIN capability
- **Resources**: 4GB RAM / 2 CPUs (unit), 6GB RAM / 4 CPUs (system)

### Authentication Reference System

**CRITICAL**: Use reference-based authentication - never copy or symlink tokens.

```bash
# Set up authentication reference (run once after token refresh)
./scripts/setup-auth-reference.sh

# This creates:
# - docker/compose/docker-compose.auth.yml (mounts canonical token location)
# - .env.auth (environment configuration with paths)
# - Direct reference to canonical token location

# When tokens are refreshed, just run setup again:
./scripts/setup-auth-reference.sh
```

**Key Benefits:**
- ✅ **Single source of truth**: Tokens stay in canonical location
- ✅ **Automatic updates**: Refreshed tokens are immediately available
- ✅ **No duplication**: No copying or symlinking required
- ✅ **Container isolation**: Docker mounts reference location read-only

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
2. **Verify images are up-to-date**: Rebuild with `./docker/scripts/build-images.sh test-runner`
3. **Check auth reference system**: 
   - Run `./scripts/setup-auth-reference.sh` to configure authentication
   - Verify `docker/compose/docker-compose.auth.yml` exists
   - Check `.env.auth` for correct paths
4. **Review logs**: Check `test-artifacts/logs/` for detailed error messages
5. **Interactive debugging**: Use `docker compose -f docker/compose/docker-compose.test.yml -f docker/compose/docker-compose.auth.yml run --rm shell`
6. **Test hanging prevention**: Use `./scripts/timeout-test-wrapper.sh "TestPattern" 60` for potentially hanging tests

### Hanging Test Prevention

**CRITICAL**: Some FUSE filesystem tests may hang indefinitely. Always use timeout protection:

```bash
# Use timeout wrapper for potentially hanging tests
./scripts/timeout-test-wrapper.sh "TestIT_FS_ETag_01_CacheValidationWithTimeoutFix" 60

# The wrapper provides:
# - Hard timeout enforcement (kills hanging processes)
# - Progress monitoring with heartbeat
# - Container cleanup
# - Detailed logging to test-artifacts/debug/
```

## Why These Rules Exist

Understanding the rationale helps ensure compliance:

### Docker Requirement

- **FUSE Dependencies**: Tests require FUSE3 device access with specific capabilities that are only available in the Docker environment
- **Isolation**: Prevents test artifacts from polluting the host system and interfering with other processes
- **Reproducibility**: Ensures consistent environment across all developers and CI/CD pipelines
- **Security**: Test credentials and OneDrive access are isolated in containers, preventing accidental exposure
- **Dependencies**: Specific versions of Go (1.24.2), Python (3.12), and system libraries are pre-configured

### Timeout Wrapper Requirement

- **Prevents Hanging**: Some FUSE filesystem tests may hang indefinitely due to kernel-level interactions
- **Resource Cleanup**: Ensures containers are properly cleaned up even on timeout or failure
- **Debugging**: Provides detailed logs in `test-artifacts/debug/` for post-mortem analysis
- **CI/CD Protection**: Prevents CI/CD pipelines from hanging indefinitely

### No `cd` Command

- **Shell Context**: The `cd` command doesn't work as expected in tool execution contexts
- **Working Directory**: Use the `cwd` parameter for bash commands instead
- **Consistency**: Ensures all operations start from the workspace root

### References

- **Running Tests Guide**: `docs/testing/running-tests.md` - Complete guide to running and writing tests
- **Test Fixtures Guide**: `docs/testing/test-fixtures.md` - Detailed fixture usage and examples
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