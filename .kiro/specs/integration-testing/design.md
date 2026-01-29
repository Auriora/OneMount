# Design: Integration Testing

## Source
Moved from `.kiro/specs/archive/system-verification-and-fix/design.md`. Sections below are verbatim.

### Verification Framework

The verification process follows a layered approach:

```
┌─────────────────────────────────────────┐
│     End-to-End User Workflows           │
├─────────────────────────────────────────┤
│     Integration Tests                   │
├─────────────────────────────────────────┤
│     Component Verification              │
├─────────────────────────────────────────┤
│     Code Analysis & Documentation       │
└─────────────────────────────────────────┘
```

### Verification Phases

#### Phase 1: Code Analysis
- Review existing code structure
- Compare implementation against architecture docs
- Identify missing or incomplete components
- Document deviations from design

#### Phase 2: Component Verification
- Test each component in isolation
- Verify against acceptance criteria
- Document failures and root causes
- Create component-specific fix plans

#### Phase 3: Integration Verification
- Test component interactions
- Verify data flow between components
- Test error propagation and handling
- Document integration issues

#### Phase 4: End-to-End Testing
- Test complete user workflows
- Verify against user stories
- Test edge cases and error scenarios
- Document user-facing issues

## Testing Strategy

### Docker Test Environment

All tests run in isolated Docker containers to avoid affecting the host system:

**Test Runner Container** (`onemount-test-runner`):
- Based on `onemount-base` image with Go 1.23+
- Includes all dependencies: FUSE3, GTK3, Python, build tools
- Pre-built OneMount binaries for faster test execution
- Mounts workspace as volume for source code access
- Writes test artifacts to `test-artifacts/` directory
- Configured with FUSE device and SYS_ADMIN capability

**Test Types**:
1. **Unit Tests**: Lightweight, no FUSE required, run with `docker compose run unit-tests`
2. **Integration Tests**: Require FUSE, run with `docker compose run integration-tests`
3. **System Tests**: Full end-to-end, require auth tokens, run with `docker compose run system-tests`
4. **Coverage Analysis**: Generate coverage reports, run with `docker compose run coverage`

**Docker Compose Services**:
- `test-runner`: Base service with common configuration
- `unit-tests`: Extends test-runner for unit tests
- `integration-tests`: Extends test-runner with FUSE support
- `system-tests`: Extends test-runner with auth token mounting
- `coverage`: Extends test-runner for coverage analysis
- `shell`: Interactive shell for debugging

**Environment Variables**:
- `ONEMOUNT_TEST_TIMEOUT`: Test timeout duration (default: 5m)
- `ONEMOUNT_TEST_VERBOSE`: Enable verbose test output
- `GORACE`: Race detector configuration
- `DOCKER_CONTAINER`: Flag indicating tests run in Docker

### Property-Based Testing Requirements

**Library Selection**: Use Go's `testing/quick` package for property-based testing, with `github.com/leanovate/gopter` as an alternative for more complex property generation.

**Test Configuration**:
- Each property-based test MUST run a minimum of 100 iterations
- Each test MUST be tagged with a comment explicitly referencing the correctness property from the design document
- Use this exact format: `**Feature: system-verification-and-fix, Property {number}: {property_text}**`

**Property Implementation Requirements**:
- Each correctness property MUST be implemented by a SINGLE property-based test
- Property tests MUST generate random inputs within the valid domain
- Property tests MUST verify the property holds for all generated inputs
- Property tests MUST provide clear failure messages with counterexamples

**Example Property Test Structure**:
```go
// **Feature: system-verification-and-fix, Property 1: OAuth2 Token Storage Security**
func TestProperty_OAuth2TokenStorageSecurity(t *testing.T) {
    property := func(authData AuthData) bool {
        // Generate random valid OAuth2 completion
        tokens, err := completeOAuth2(authData)
        if err != nil {
            return true // Skip invalid inputs
        }
        
        // Verify tokens are stored securely
        return verifySecureTokenStorage(tokens)
    }
    
    config := &quick.Config{MaxCount: 100}
    if err := quick.Check(property, config); err != nil {
        t.Errorf("Property failed: %v", err)
    }
}
```

**Property Test Organization**:
- Group property tests by component (authentication, filesystem, etc.)
- Place property tests in `*_property_test.go` files
- Run property tests as part of integration test suite
- Include property tests in coverage analysis

### Unit Test Verification

Review existing unit tests in Docker:
- Identify gaps in coverage
- Verify tests check actual behavior, not implementation details
- Add missing tests for edge cases
- Ensure tests are deterministic
- Run with: `docker compose -f docker/compose/docker-compose.test.yml run unit-tests`

### Integration Test Creation

Create new integration tests for:
- Authentication → Mounting → File Access flow
- File Modification → Upload → Delta Sync flow
- Online → Offline → Online transition flow
- Concurrent file operations
- Error recovery scenarios
- Run with: `docker compose -f docker/compose/docker-compose.test.yml run integration-tests`

### End-to-End Test Creation

Create end-to-end tests for:
- Complete user workflow from install to file access
- Multi-file operations (copy directory, etc.)
- Long-running operations (large file upload)
- Stress testing (many concurrent operations)
- Run with: `docker compose -f docker/compose/docker-compose.test.yml run system-tests`

### Test Execution Strategy

1. **Build test images**: `docker compose -f docker/compose/docker-compose.build.yml build`
2. **Run unit tests**: Fast feedback, no external dependencies
3. **Run integration tests**: Verify component interactions
4. **Run system tests**: Full end-to-end with real OneDrive (requires auth)
5. **Generate coverage**: Analyze test coverage
6. **Review artifacts**: Check `test-artifacts/` for logs and results
7. **Debug in container**: Use `docker compose run shell` for interactive debugging
