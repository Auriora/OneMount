# Running Tests in OneMount

**Last Updated**: January 23, 2026  
**Status**: Complete

This guide provides comprehensive instructions for running tests in the OneMount project. All tests must be run in Docker containers to ensure proper isolation, reproducibility, and access to required dependencies like FUSE.

## Table of Contents

1. [Quick Start](#quick-start)
2. [Test Types and Naming Conventions](#test-types-and-naming-conventions)
3. [Running Tests](#running-tests)
4. [Test Requirements](#test-requirements)
5. [Troubleshooting](#troubleshooting)
6. [Adding New Tests](#adding-new-tests)
7. [CI/CD Integration](#cicd-integration)

---

## Quick Start

### Prerequisites

1. **Docker and Docker Compose** installed
2. **Test images built**:
   ```bash
   ./docker/scripts/build-images.sh test-runner
   ```
3. **Authentication configured** (for integration/system tests):
   ```bash
   ./scripts/setup-auth-reference.sh
   ```

### Run All Tests

```bash
# Unit tests only (fast, no auth required)
docker compose -f docker/compose/docker-compose.test.yml run --rm unit-tests

# Integration tests (requires auth)
docker compose -f docker/compose/docker-compose.test.yml \
  -f docker/compose/docker-compose.auth.yml run --rm integration-tests

# All tests with coverage
docker compose -f docker/compose/docker-compose.test.yml \
  -f docker/compose/docker-compose.auth.yml run --rm test-runner all
```

---

## Test Types and Naming Conventions

OneMount uses a strict naming convention to distinguish between different test types:

### 1. Unit Tests (`TestUT_*`)

**Purpose**: Test individual functions/methods in isolation with mocks

**Characteristics**:
- No external dependencies
- Uses mock Graph API backend
- No authentication required
- Fast execution (< 1 second per test)
- No network calls

**Naming Pattern**: `TestUT_Component_Feature_Scenario`

**Example**:
```go
func TestUT_FS_CacheHit_ReturnsLocalContent(t *testing.T) {
    fixture := helpers.SetupMockFSTestFixture(t, "CacheHitTest", newFilesystem)
    defer fixture.Teardown(t)
    
    // Test logic with mocks...
}
```

**When to use**:
- Testing internal logic and algorithms
- Testing error handling paths
- Testing edge cases
- When you don't need real OneDrive data
- During rapid development cycles

### 2. Integration Tests (`TestIT_*`)

**Purpose**: Test real API interactions and component integration

**Characteristics**:
- Requires real OneDrive authentication
- Makes actual API calls to Microsoft Graph
- Network connectivity required
- Moderate execution time (1-10 seconds per test)
- Tests against real OneDrive data

**Naming Pattern**: `TestIT_Component_Feature_Scenario`

**Example**:
```go
func TestIT_FS_FileUpload_UploadsToOneDrive(t *testing.T) {
    fixture := helpers.SetupIntegrationFSTestFixture(t, "FileUploadTest", newFilesystem)
    defer fixture.Teardown(t)
    
    // Test logic with real OneDrive...
}
```

**When to use**:
- Testing Graph API integration
- Verifying real OneDrive behavior
- Testing upload/download workflows
- Testing delta sync
- Testing conflict resolution

### 3. Property-Based Tests (`TestProperty*`)

**Purpose**: Generative testing with random inputs to verify properties

**Characteristics**:
- Uses `pgregory.net/rapid` for property-based testing
- Generates random test data
- Runs 100+ iterations per property
- Can use mocks or real API depending on property
- Verifies universal properties hold across all inputs

**Naming Pattern**: `TestProperty<Number>_<PropertyName>`

**Example**:
```go
func TestProperty24_OfflineDetection(t *testing.T) {
    rapid.Check(t, func(t *rapid.T) {
        // Generate random network failure scenarios
        // Verify offline detection works correctly
    })
}
```

**When to use**:
- Testing correctness properties
- Finding edge cases automatically
- Verifying invariants
- Testing with wide range of inputs

### 4. System Tests (`TestSystemST_*`)

**Purpose**: End-to-end testing with full mounting and user workflows

**Characteristics**:
- Full FUSE mounting
- Real OneDrive authentication
- Complete user workflows
- Slowest execution (10+ seconds per test)
- Requires FUSE device access
- Must run in Docker for isolation

**Naming Pattern**: `TestSystemST_Feature_Scenario`

**Example**:
```go
func TestSystemST_CompleteWorkflow_MountReadWrite(t *testing.T) {
    fixture := helpers.SetupSystemTestFixture(t, "CompleteWorkflowTest", newFilesystem)
    defer fixture.Teardown(t)
    
    // Test complete user workflow...
}
```

**When to use**:
- Testing complete user workflows
- Testing mounting and unmounting
- Testing file operations through FUSE
- End-to-end validation
- Performance testing

---

## Running Tests

### Unit Tests

Unit tests are fast and don't require authentication. They use mock fixtures.

**Run all unit tests**:
```bash
docker compose -f docker/compose/docker-compose.test.yml run --rm unit-tests
```

**Run specific unit test**:
```bash
docker compose -f docker/compose/docker-compose.test.yml run --rm \
  test-runner go test -v -run "^TestUT_FS_CacheHit" ./internal/fs
```

**Run unit tests in a specific package**:
```bash
docker compose -f docker/compose/docker-compose.test.yml run --rm \
  test-runner go test -v -run "^TestUT_" ./internal/fs
```

**Run unit tests with race detector**:
```bash
docker compose -f docker/compose/docker-compose.test.yml run --rm \
  test-runner go test -v -race -run "^TestUT_" ./internal/...
```

### Integration Tests

Integration tests require authentication and make real API calls.

**Setup authentication first**:
```bash
./scripts/setup-auth-reference.sh
```

**Run all integration tests**:
```bash
docker compose -f docker/compose/docker-compose.test.yml \
  -f docker/compose/docker-compose.auth.yml run --rm integration-tests
```

**Run specific integration test**:
```bash
docker compose -f docker/compose/docker-compose.test.yml \
  -f docker/compose/docker-compose.auth.yml run --rm \
  test-runner go test -v -run "^TestIT_FS_FileUpload" ./internal/fs
```

**Run integration tests with timeout protection** (recommended for potentially hanging tests):
```bash
./scripts/timeout-test-wrapper.sh "TestIT_FS_ETag" 60
```

### Property-Based Tests

Property-based tests verify correctness properties with random inputs.

**Run all property tests**:
```bash
docker compose -f docker/compose/docker-compose.test.yml run --rm \
  test-runner go test -v -run "^TestProperty" ./internal/fs
```

**Run specific property test**:
```bash
docker compose -f docker/compose/docker-compose.test.yml run --rm \
  test-runner go test -v -run "^TestProperty24_OfflineDetection" ./internal/fs
```

**Run property tests with more iterations**:
```bash
docker compose -f docker/compose/docker-compose.test.yml run --rm \
  test-runner go test -v -run "^TestProperty" -rapid.checks=1000 ./internal/fs
```

### System Tests

System tests require full FUSE mounting and authentication.

**Run all system tests**:
```bash
docker compose -f docker/compose/docker-compose.test.yml \
  -f docker/compose/docker-compose.auth.yml run --rm system-tests
```

**Run specific system test**:
```bash
docker compose -f docker/compose/docker-compose.test.yml \
  -f docker/compose/docker-compose.auth.yml run --rm \
  test-runner go test -v -run "^TestSystemST_" ./tests/system
```

### All Tests with Coverage

**Run all tests and generate coverage report**:
```bash
docker compose -f docker/compose/docker-compose.test.yml \
  -f docker/compose/docker-compose.auth.yml run --rm test-runner coverage
```

**View coverage in browser**:
```bash
go tool cover -html=coverage/coverage.out
```

### Interactive Debugging

**Open interactive shell in test container**:
```bash
docker compose -f docker/compose/docker-compose.test.yml \
  -f docker/compose/docker-compose.auth.yml run --rm shell
```

**Run tests interactively**:
```bash
# Inside the shell
go test -v -run TestUT_FS_CacheHit ./internal/fs
```

---

## Test Requirements

### Unit Tests

**Requirements**:
- ✅ Docker container
- ❌ No authentication required
- ❌ No FUSE device required
- ❌ No network connectivity required

**Fixtures**: Use `SetupMockFSTestFixture`

**Example**:
```go
func TestUT_MyFeature(t *testing.T) {
    fixture := helpers.SetupMockFSTestFixture(t, "MyTest", newFilesystem)
    defer fixture.Teardown(t)
    // ...
}
```

### Integration Tests

**Requirements**:
- ✅ Docker container
- ✅ Authentication tokens required
- ✅ FUSE device required
- ✅ Network connectivity required

**Setup**:
```bash
./scripts/setup-auth-reference.sh
```

**Fixtures**: Use `SetupIntegrationFSTestFixture`

**Example**:
```go
func TestIT_MyFeature(t *testing.T) {
    fixture := helpers.SetupIntegrationFSTestFixture(t, "MyTest", newFilesystem)
    defer fixture.Teardown(t)
    // ...
}
```

### Property-Based Tests

**Requirements**:
- ✅ Docker container
- ⚠️ Authentication depends on property being tested
- ⚠️ FUSE depends on property being tested
- ⚠️ Network depends on property being tested

**Fixtures**: Use appropriate fixture based on property

**Example**:
```go
func TestProperty24_OfflineDetection(t *testing.T) {
    rapid.Check(t, func(t *rapid.T) {
        // Generate random test data
        // Verify property holds
    })
}
```

### System Tests

**Requirements**:
- ✅ Docker container (mandatory)
- ✅ Authentication tokens required
- ✅ FUSE device required
- ✅ Network connectivity required
- ✅ Root/appropriate permissions for mounting

**Setup**:
```bash
./scripts/setup-auth-reference.sh
```

**Fixtures**: Use `SetupSystemTestFixture`

**Example**:
```go
func TestSystemST_MyFeature(t *testing.T) {
    fixture := helpers.SetupSystemTestFixture(t, "MyTest", newFilesystem)
    defer fixture.Teardown(t)
    // ...
}
```

---

## Troubleshooting

### Common Issues

#### Issue: "real auth tokens not available"

**Cause**: Authentication tokens not configured

**Solution**:
```bash
./scripts/setup-auth-reference.sh
```

Verify tokens exist:
```bash
ls -la test-artifacts/.auth_tokens.json
```

#### Issue: "FUSE device not available"

**Cause**: Not running in Docker container

**Solution**: Always use Docker:
```bash
docker compose -f docker/compose/docker-compose.test.yml run --rm unit-tests
```

#### Issue: Tests hang indefinitely

**Cause**: Some FUSE tests may hang due to kernel-level interactions

**Solution**: Use timeout wrapper:
```bash
./scripts/timeout-test-wrapper.sh "TestIT_FS_ETag" 60
```

#### Issue: "permission denied" when accessing /dev/fuse

**Cause**: Container doesn't have FUSE capabilities

**Solution**: Ensure using correct Docker compose file:
```bash
docker compose -f docker/compose/docker-compose.test.yml run --rm integration-tests
```

#### Issue: Tests are very slow

**Cause**: Running integration/system tests when unit tests would suffice

**Solution**: 
- Use unit tests for fast feedback: `docker compose -f docker/compose/docker-compose.test.yml run --rm unit-tests`
- Reserve integration tests for API verification
- Reserve system tests for end-to-end validation

#### Issue: Mock client not working

**Cause**: Mock responses not set up before testing

**Solution**: Set up mock responses:
```go
mockClient := fsFixture.MockClient
mockClient.AddMockItem("/me/drive/items/id", item)
mockClient.AddMockResponse("/me/drive/items/id/content", content, 200, nil)
```

### Debug Logging

**Enable verbose test output**:
```bash
docker compose -f docker/compose/docker-compose.test.yml run --rm \
  -e ONEMOUNT_LOG_TO_FILE=false \
  test-runner go test -v -run TestPattern ./internal/fs
```

**Check test logs**:
```bash
ls -la test-artifacts/logs/
cat test-artifacts/logs/latest-test-run.log
```

**Check debug logs**:
```bash
ls -la test-artifacts/debug/
cat test-artifacts/debug/timeout-test-*.log
```

### Verifying Test Environment

**Check Docker images**:
```bash
docker images | grep onemount
```

**Check FUSE device in container**:
```bash
docker compose -f docker/compose/docker-compose.test.yml run --rm \
  shell ls -l /dev/fuse
```

**Check Go version**:
```bash
docker compose -f docker/compose/docker-compose.test.yml run --rm \
  shell go version
```

**Check auth configuration**:
```bash
cat docker/compose/docker-compose.auth.yml
cat .env.auth
```

---

## Adding New Tests

### Step 1: Choose Test Type

Determine which type of test you need:

- **Unit test**: Testing isolated logic with mocks → `TestUT_*`
- **Integration test**: Testing real API interactions → `TestIT_*`
- **Property test**: Testing universal properties → `TestProperty*`
- **System test**: Testing end-to-end workflows → `TestSystemST_*`

### Step 2: Create Test File

Place test file next to the code being tested:

```
internal/fs/
├── cache.go
└── cache_test.go          # Tests for cache.go
```

### Step 3: Write Test

**Unit Test Example**:
```go
package fs

import (
    "testing"
    "github.com/auriora/OneMount/internal/testutil/helpers"
)

func TestUT_Cache_Hit_ReturnsLocalContent(t *testing.T) {
    // Setup mock fixture
    fixture := helpers.SetupMockFSTestFixture(t, "CacheHitTest", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
        return NewFilesystem(auth, mountPoint, cacheTTL)
    })
    defer fixture.Teardown(t)

    // Get fixture data
    fsFixture := fixture.GetFixture(t).(*helpers.FSTestFixture)
    mockClient := fsFixture.MockClient
    
    // Set up mock data
    helpers.CreateMockFile(mockClient, fsFixture.RootID, "test.txt", "file-1", "content")
    
    // Test logic
    content, err := fsFixture.FS.(*Filesystem).ReadFile("test.txt")
    if err != nil {
        t.Fatalf("ReadFile failed: %v", err)
    }
    
    if string(content) != "content" {
        t.Errorf("Expected 'content', got '%s'", string(content))
    }
}
```

**Integration Test Example**:
```go
func TestIT_Cache_RealOneDrive_DownloadsFile(t *testing.T) {
    // Setup integration fixture
    fixture := helpers.SetupIntegrationFSTestFixture(t, "RealOneDriveTest", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
        return NewFilesystem(auth, mountPoint, cacheTTL)
    })
    defer fixture.Teardown(t)

    // Get fixture data
    fsFixture := fixture.GetFixture(t).(*helpers.FSTestFixture)
    filesystem := fsFixture.FS.(*Filesystem)
    
    // Test with real OneDrive
    content, err := filesystem.ReadFile("test.txt")
    if err != nil {
        t.Fatalf("ReadFile failed: %v", err)
    }
    
    // Verify content
    if len(content) == 0 {
        t.Error("Expected non-empty content")
    }
}
```

**Property Test Example**:
```go
import "pgregory.net/rapid"

func TestProperty_Cache_AlwaysReturnsConsistentData(t *testing.T) {
    rapid.Check(t, func(t *rapid.T) {
        // Generate random file content
        content := rapid.String().Draw(t, "content")
        
        // Setup and test
        fixture := helpers.SetupMockFSTestFixture(t, "PropertyTest", newFilesystem)
        defer fixture.Teardown(t)
        
        // Verify property: reading same file twice returns same content
        content1, _ := fixture.ReadFile("test.txt")
        content2, _ := fixture.ReadFile("test.txt")
        
        if string(content1) != string(content2) {
            t.Fatalf("Inconsistent reads: %s != %s", content1, content2)
        }
    })
}
```

### Step 4: Run Test

**Run your new test**:
```bash
# Unit test
docker compose -f docker/compose/docker-compose.test.yml run --rm \
  test-runner go test -v -run "^TestUT_Cache_Hit" ./internal/fs

# Integration test
docker compose -f docker/compose/docker-compose.test.yml \
  -f docker/compose/docker-compose.auth.yml run --rm \
  test-runner go test -v -run "^TestIT_Cache_RealOneDrive" ./internal/fs
```

### Step 5: Verify Test Passes

Ensure your test:
- ✅ Compiles without errors
- ✅ Passes consistently
- ✅ Cleans up resources properly
- ✅ Uses appropriate fixtures
- ✅ Follows naming conventions
- ✅ Has clear assertions and error messages

---

## CI/CD Integration

### GitHub Actions

The project uses separate workflows for different test types:

**CI Workflow** (`.github/workflows/ci.yml`):
- Runs on every push
- Executes unit tests only (fast feedback)
- No authentication required

**Integration Test Workflow** (`.github/workflows/integration-tests.yml`):
- Runs on pull requests
- Executes integration tests
- Requires authentication secrets

**System Test Workflow** (`.github/workflows/system-tests.yml`):
- Runs on release tags
- Executes full system tests
- Requires authentication and FUSE

### Running Tests Locally Like CI

**Simulate CI unit test run**:
```bash
docker compose -f docker/compose/docker-compose.test.yml run --rm unit-tests
```

**Simulate CI integration test run**:
```bash
docker compose -f docker/compose/docker-compose.test.yml \
  -f docker/compose/docker-compose.auth.yml run --rm integration-tests
```

**Simulate CI coverage run**:
```bash
docker compose -f docker/compose/docker-compose.test.yml \
  -f docker/compose/docker-compose.auth.yml run --rm test-runner coverage
```

---

## Best Practices

### 1. Choose the Right Test Type

- **Unit tests**: Fast feedback during development
- **Integration tests**: Verify API behavior
- **Property tests**: Find edge cases automatically
- **System tests**: Validate complete workflows

### 2. Follow Naming Conventions

Always use the correct prefix:
- `TestUT_Component_Feature_Scenario` - Unit tests
- `TestIT_Component_Feature_Scenario` - Integration tests
- `TestProperty<N>_PropertyName` - Property tests
- `TestSystemST_Feature_Scenario` - System tests

### 3. Always Clean Up

Always defer fixture teardown:
```go
fixture := helpers.SetupMockFSTestFixture(t, "MyTest", newFilesystem)
defer fixture.Teardown(t)
```

### 4. Use Appropriate Fixtures

- Unit tests → `SetupMockFSTestFixture`
- Integration tests → `SetupIntegrationFSTestFixture`
- System tests → `SetupSystemTestFixture`

### 5. Add Skip Logic for Integration/System Tests

```go
func TestIT_MyFeature(t *testing.T) {
    if !authAvailable() {
        t.Skip("Auth tokens required for integration test")
    }
    // Test logic...
}
```

### 6. Use Timeout Protection

For potentially hanging tests:
```bash
./scripts/timeout-test-wrapper.sh "TestIT_FS_ETag" 60
```

### 7. Write Clear Assertions

```go
// Good: Clear error message
if got != want {
    t.Errorf("Expected %v, got %v", want, got)
}

// Bad: Unclear error message
if got != want {
    t.Error("Test failed")
}
```

---

## Related Documentation

- [Test Fixtures Guide](test-fixtures.md) - Detailed fixture usage
- [Test Audit Report](test-audit-report.md) - Test suite analysis
- [Docker Test Environment](docker-test-environment.md) - Docker setup details
- [Testing Conventions](.kiro/steering/testing-conventions.md) - Testing standards
- [Development Guidelines](../guides/developer/DEVELOPMENT.md) - Development workflow

---

## Summary

| Test Type | Prefix | Auth Required | FUSE Required | Speed | Use Case |
|-----------|--------|---------------|---------------|-------|----------|
| Unit | `TestUT_` | No | No | Fast | Isolated logic with mocks |
| Integration | `TestIT_` | Yes | Yes | Medium | Real API interactions |
| Property | `TestProperty` | Varies | Varies | Medium | Universal properties |
| System | `TestSystemST_` | Yes | Yes | Slow | End-to-end workflows |

**Key Commands**:
```bash
# Unit tests (fast, no auth)
docker compose -f docker/compose/docker-compose.test.yml run --rm unit-tests

# Integration tests (requires auth)
docker compose -f docker/compose/docker-compose.test.yml \
  -f docker/compose/docker-compose.auth.yml run --rm integration-tests

# All tests with coverage
docker compose -f docker/compose/docker-compose.test.yml \
  -f docker/compose/docker-compose.auth.yml run --rm test-runner all
```

**Remember**:
- ✅ Always use Docker containers
- ✅ Follow naming conventions
- ✅ Use appropriate fixtures
- ✅ Clean up resources
- ✅ Add skip logic for auth-required tests
- ✅ Use timeout protection for potentially hanging tests
