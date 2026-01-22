# Test Fixtures Guide

This document describes the test fixture helpers available in the OneMount project and how to use them effectively.

## Overview

The OneMount project provides three types of test fixtures to support different testing scenarios:

1. **Mock Fixtures** - For unit tests with mock Graph API backend
2. **Integration Fixtures** - For integration tests with real OneDrive
3. **System Fixtures** - For end-to-end tests with full mounting

## Test Fixture Types

### 1. Mock Fixtures (Unit Tests)

**Purpose**: Fast, isolated unit tests that don't require real OneDrive authentication.

**Function**: `SetupMockFSTestFixture(t, fixtureName, newFilesystem)`

**Features**:
- Uses `MockGraphProvider` instead of real authentication
- Creates filesystem with mock backend
- No authentication required
- Fast execution
- No network calls
- Fully isolated

**When to use**:
- Testing internal logic and algorithms
- Testing error handling paths
- Testing edge cases
- When you don't need real OneDrive data
- When speed is important

**Example**:

```go
func TestUT_MyFeature_BasicLogic(t *testing.T) {
    fixture := helpers.SetupMockFSTestFixture(t, "BasicLogicTest", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
        return fs.NewFilesystem(auth, mountPoint, cacheTTL)
    })
    defer fixture.Teardown(t)

    // Access the fixture data
    fsFixture := fixture.GetFixture(t).(*helpers.FSTestFixture)
    mockClient := fsFixture.MockClient
    
    // Set up mock responses
    mockClient.AddMockItem("/me/drive/items/test-id", &graph.DriveItem{
        ID:   "test-id",
        Name: "test-file.txt",
    })
    
    // Test your code
    // ...
}
```

### 2. Integration Fixtures (Integration Tests)

**Purpose**: Integration tests that verify behavior against real Microsoft Graph API.

**Function**: `SetupIntegrationFSTestFixture(t, fixtureName, newFilesystem)`

**Features**:
- Uses real auth tokens from `.auth_tokens.json`
- Creates filesystem with real OneDrive connection
- Tests against real Microsoft Graph API
- Requires network connectivity
- Slower than unit tests

**Requirements**:
- Auth tokens must be available in `test-artifacts/.auth_tokens.json`
- Network connectivity required
- Should be run in Docker container

**When to use**:
- Testing Graph API integration
- Verifying real OneDrive behavior
- Testing upload/download workflows
- Testing delta sync
- Testing conflict resolution

**Example**:

```go
func TestIT_MyFeature_RealOneDrive(t *testing.T) {
    fixture := helpers.SetupIntegrationFSTestFixture(t, "RealOneDriveTest", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
        return fs.NewFilesystem(auth, mountPoint, cacheTTL)
    })
    defer fixture.Teardown(t)

    // Access the fixture data
    fsFixture := fixture.GetFixture(t).(*helpers.FSTestFixture)
    filesystem := fsFixture.FS.(*fs.Filesystem)
    
    // Test with real OneDrive
    // ...
}
```

### 3. System Fixtures (System Tests)

**Purpose**: Full end-to-end system tests with complete mounting and user workflows.

**Function**: `SetupSystemTestFixture(t, fixtureName, newFilesystem)`

**Features**:
- Full end-to-end setup with FUSE mounting
- Uses real auth and real OneDrive
- Tests complete user workflows
- Requires FUSE device access
- Most comprehensive but slowest

**Requirements**:
- Auth tokens must be available in `test-artifacts/.auth_tokens.json`
- FUSE device must be accessible (`/dev/fuse`)
- Network connectivity required
- Root or appropriate permissions for mounting
- **Must be run in Docker container for isolation**

**When to use**:
- Testing complete user workflows
- Testing mounting and unmounting
- Testing file operations through FUSE
- End-to-end validation
- Performance testing

**Example**:

```go
func TestST_MyFeature_EndToEnd(t *testing.T) {
    fixture := helpers.SetupSystemTestFixture(t, "EndToEndTest", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
        return fs.NewFilesystem(auth, mountPoint, cacheTTL)
    })
    defer fixture.Teardown(t)

    // Access the fixture data
    sysFixture := fixture.GetFixture(t).(*helpers.SystemTestFixture)
    mountPoint := sysFixture.MountPoint
    
    // Perform mounting if needed
    // Test file operations through mount point
    // ...
}
```

## Automatic Test Type Detection

The `SetupFSTestFixture` function automatically detects the test type based on the test name prefix:

- `TestUT_*` → Uses `SetupMockFSTestFixture` (unit tests)
- `TestIT_*` → Uses `SetupIntegrationFSTestFixture` (integration tests)
- `TestST_*` or `TestE2E_*` → Uses `SetupSystemTestFixture` (system tests)

**Example**:

```go
func TestUT_MyFeature(t *testing.T) {
    // Automatically uses mock fixture
    fixture := helpers.SetupFSTestFixture(t, "MyTest", newFilesystem)
    defer fixture.Teardown(t)
    // ...
}

func TestIT_MyFeature(t *testing.T) {
    // Automatically uses integration fixture
    fixture := helpers.SetupFSTestFixture(t, "MyTest", newFilesystem)
    defer fixture.Teardown(t)
    // ...
}
```

## Best Practices

### 1. Choose the Right Fixture Type

- **Unit tests**: Use mock fixtures for fast, isolated tests
- **Integration tests**: Use integration fixtures when you need real API behavior
- **System tests**: Use system fixtures for complete end-to-end validation

### 2. Test Naming Convention

Follow the standard test naming convention:

- `TestUT_Component_Feature_Scenario` - Unit tests
- `TestIT_Component_Feature_Scenario` - Integration tests
- `TestST_Component_Feature_Scenario` - System tests
- `TestE2E_Feature_Scenario` - End-to-end tests

### 3. Always Clean Up

Always defer the fixture teardown to ensure proper cleanup:

```go
fixture := helpers.SetupMockFSTestFixture(t, "MyTest", newFilesystem)
defer fixture.Teardown(t)
```

### 4. Use Explicit Fixture Functions

For clarity and explicit control, prefer using the specific fixture functions:

```go
// Good - explicit and clear
fixture := helpers.SetupMockFSTestFixture(t, "MyTest", newFilesystem)

// Also good - automatic detection
fixture := helpers.SetupFSTestFixture(t, "MyTest", newFilesystem)
```

### 5. Mock Setup for Unit Tests

When using mock fixtures, set up your mock responses before testing:

```go
fsFixture := fixture.GetFixture(t).(*helpers.FSTestFixture)
mockClient := fsFixture.MockClient

// Set up mock directory
helpers.CreateMockDirectory(mockClient, "root-id", "test-dir", "dir-id")

// Set up mock file
helpers.CreateMockFile(mockClient, "dir-id", "test.txt", "file-id", "content")
```

### 6. Authentication for Integration/System Tests

Integration and system tests require real authentication tokens:

1. Set up authentication: `./scripts/setup-auth-reference.sh`
2. Ensure tokens are in `test-artifacts/.auth_tokens.json`
3. Run tests in Docker with auth override:
   ```bash
   docker compose -f docker/compose/docker-compose.test.yml \
     -f docker/compose/docker-compose.auth.yml run --rm integration-tests
   ```

### 7. Running Tests

**Unit Tests** (fast, no auth required):
```bash
docker compose -f docker/compose/docker-compose.test.yml run --rm unit-tests
```

**Integration Tests** (requires auth):
```bash
docker compose -f docker/compose/docker-compose.test.yml \
  -f docker/compose/docker-compose.auth.yml run --rm integration-tests
```

**System Tests** (requires auth and FUSE):
```bash
docker compose -f docker/compose/docker-compose.test.yml \
  -f docker/compose/docker-compose.auth.yml run --rm system-tests
```

## Common Patterns

### Pattern 1: Unit Test with Mock Data

```go
func TestUT_Feature_Scenario(t *testing.T) {
    fixture := helpers.SetupMockFSTestFixture(t, "ScenarioTest", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
        return fs.NewFilesystem(auth, mountPoint, cacheTTL)
    })
    defer fixture.Teardown(t)

    fsFixture := fixture.GetFixture(t).(*helpers.FSTestFixture)
    mockClient := fsFixture.MockClient
    
    // Set up test data
    helpers.CreateMockFile(mockClient, fsFixture.RootID, "test.txt", "file-1", "content")
    
    // Run test
    // ...
}
```

### Pattern 2: Integration Test with Real OneDrive

```go
func TestIT_Feature_Scenario(t *testing.T) {
    fixture := helpers.SetupIntegrationFSTestFixture(t, "ScenarioTest", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
        return fs.NewFilesystem(auth, mountPoint, cacheTTL)
    })
    defer fixture.Teardown(t)

    fsFixture := fixture.GetFixture(t).(*helpers.FSTestFixture)
    filesystem := fsFixture.FS.(*fs.Filesystem)
    
    // Test with real OneDrive
    // ...
}
```

### Pattern 3: System Test with Mounting

```go
func TestST_Feature_Scenario(t *testing.T) {
    fixture := helpers.SetupSystemTestFixture(t, "ScenarioTest", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
        return fs.NewFilesystem(auth, mountPoint, cacheTTL)
    })
    defer fixture.Teardown(t)

    sysFixture := fixture.GetFixture(t).(*helpers.SystemTestFixture)
    
    // Mount and test
    // ...
}
```

## Troubleshooting

### Issue: "real auth tokens not available"

**Solution**: Set up authentication tokens:
```bash
./scripts/setup-auth-reference.sh
```

### Issue: "FUSE device not available"

**Solution**: Run tests in Docker container:
```bash
docker compose -f docker/compose/docker-compose.test.yml run --rm system-tests
```

### Issue: Tests are slow

**Solution**: 
- Use mock fixtures for unit tests (fastest)
- Use integration fixtures only when needed
- Use system fixtures only for end-to-end validation

### Issue: Mock client not working

**Solution**: Ensure you're setting up mock responses before testing:
```go
mockClient.AddMockItem("/me/drive/items/id", item)
mockClient.AddMockResponse("/me/drive/items/id/content", content, 200, nil)
```

## Related Documentation

- [Test Setup Guide](../TEST_SETUP.md) - Complete test environment setup
- [Docker Test Environment](docker-test-environment.md) - Docker-specific details
- [Testing Conventions](.kiro/steering/testing-conventions.md) - Testing standards

## Summary

| Fixture Type | Function | Use Case | Auth Required | Speed |
|-------------|----------|----------|---------------|-------|
| Mock | `SetupMockFSTestFixture` | Unit tests | No | Fast |
| Integration | `SetupIntegrationFSTestFixture` | API integration | Yes | Medium |
| System | `SetupSystemTestFixture` | End-to-end | Yes | Slow |

Choose the appropriate fixture type based on your testing needs, and always follow the naming conventions for automatic test type detection.
