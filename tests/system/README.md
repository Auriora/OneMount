# System Tests

This directory contains system-level tests that verify OneMount functionality with real OneDrive integration.

## Test Types

### Authentication Tests
- Real OAuth2 flow testing
- Token refresh validation
- Multi-account scenarios

### End-to-End Integration Tests
- Full mount/unmount cycles with real OneDrive
- File operations with real API calls
- Conflict resolution with actual remote changes

### Performance Tests
- Large file upload/download with real OneDrive
- High file count scenarios
- Network condition testing

## Running System Tests

System tests require:
1. Valid OneDrive authentication tokens
2. Network connectivity to Microsoft Graph API
3. GUI support for interactive authentication (when needed)

```bash
# Run all system tests
docker compose -f docker/compose/docker-compose.test.yml run --rm system-tests

# Run specific system test
docker compose -f docker/compose/docker-compose.test.yml run --rm test-runner go test -v -run TestSystemST_Auth ./tests/system

# Interactive authentication setup
docker compose -f docker/compose/docker-compose.test.yml run --rm shell
```

## Test Data

System tests may create temporary files and directories in your OneDrive. These are cleaned up automatically, but test files are prefixed with `onemount-test-` for easy identification.

## Authentication

For GUI authentication testing, ensure:
1. X11 forwarding is working (`echo $DISPLAY` should show a value)
2. You can run GUI applications (`xeyes` or `xclock` should work)
3. Your OneDrive account has appropriate permissions

## Notes

- System tests are slower than unit/integration tests due to real API calls
- Network connectivity is required
- Some tests may be skipped if authentication fails
- Test artifacts are stored in `test-artifacts/logs/`