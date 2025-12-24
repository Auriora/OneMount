# OneMount Test Environment Setup

## Overview

This document describes the test environment setup for OneMount, including Docker configuration, reference-based authentication, and test data.

## Authentication Reference System

### Reference-Based Authentication (NEW)

**CRITICAL**: OneMount now uses a reference-based authentication system that eliminates token copying and symlinking.

#### Setup Process

```bash
# 1. Authenticate with OneMount (creates tokens in canonical location)
./build/onemount --auth-only

# 2. Set up authentication reference system
./scripts/setup-auth-reference.sh

# This creates:
# - docker/compose/docker-compose.auth.yml (Docker override)
# - .env.auth (environment configuration)
# - Direct reference to canonical token location (no copying)
```

#### How It Works

1. **Canonical Location**: Tokens remain in their original location (e.g., `~/.cache/onedriver/*/auth_tokens.json`)
2. **Reference Configuration**: `setup-auth-reference.sh` finds the newest valid tokens and creates reference configuration
3. **Docker Integration**: `docker-compose.auth.yml` mounts the canonical location into containers
4. **Automatic Updates**: When tokens are refreshed, just run `setup-auth-reference.sh` again

#### Key Benefits

- ✅ **Single source of truth**: Tokens stay in canonical location
- ✅ **No duplication**: No copying or symlinking required  
- ✅ **Automatic updates**: Refreshed tokens are immediately available
- ✅ **Container isolation**: Docker mounts reference location read-only

### Token Validation

To verify the reference system is working:

```bash
# Check if reference system is configured
ls -la docker/compose/docker-compose.auth.yml
cat .env.auth

# Test authentication in Docker container
docker compose -f docker/compose/docker-compose.test.yml \
  -f docker/compose/docker-compose.auth.yml run --rm test-runner \
  bash -c "ls -la \$ONEMOUNT_AUTH_PATH && echo 'Auth reference working'"
```

### Token Refresh

When tokens expire, refresh them and update the reference:

```bash
# 1. Re-authenticate (creates new tokens in canonical location)
./build/onemount --auth-only

# 2. Update reference system to point to new tokens
./scripts/setup-auth-reference.sh

# The reference system automatically finds the newest valid tokens
```

### Legacy Authentication (DEPRECATED)

**WARNING**: The old system of copying tokens to `test-artifacts/.auth_tokens.json` is deprecated and should not be used. Use the reference-based system instead.

## Test Data

### Sample Files

The test OneDrive account should contain sample files for testing:

- Small text files (< 1MB) for basic read/write tests
- Medium files (1-10MB) for download/upload tests
- Large files (> 10MB) for chunked upload tests
- Directories with multiple files for directory operations
- Files with special characters in names

### Test Directory Structure

```
OneDrive (Test Account)
├── test-files/
│   ├── small-file.txt (< 1MB)
│   ├── medium-file.dat (1-10MB)
│   └── large-file.bin (> 10MB)
├── test-directories/
│   ├── dir1/
│   │   ├── file1.txt
│   │   └── file2.txt
│   └── dir2/
│       └── nested/
│           └── file3.txt
└── special-chars/
    ├── file with spaces.txt
    ├── file-with-dashes.txt
    └── file_with_underscores.txt
```

## Docker Test Environment

### Images

Two Docker images are used for testing:

1. **onemount-base:latest** (1.49GB)
   - Ubuntu 24.04
   - Go 1.24.2
   - FUSE3 support
   - Build dependencies

2. **onemount-test-runner:latest** (2.21GB)
   - Extends onemount-base
   - Python 3.12 for test scripts
   - Pre-built OneMount binaries
   - Test utilities

### Building Images

```bash
# Build test runner (includes latest authentication system)
./docker/scripts/build-images.sh test-runner
```

### Running Tests

**IMPORTANT**: Always include the authentication override for integration/system tests:

```bash
# Unit tests (no authentication required)
docker compose -f docker/compose/docker-compose.test.yml run --rm unit-tests

# Integration tests (requires FUSE and authentication)
docker compose -f docker/compose/docker-compose.test.yml \
  -f docker/compose/docker-compose.auth.yml run --rm integration-tests

# System tests (requires authentication)
docker compose -f docker/compose/docker-compose.test.yml \
  -f docker/compose/docker-compose.auth.yml run --rm system-tests

# All tests with authentication
docker compose -f docker/compose/docker-compose.test.yml \
  -f docker/compose/docker-compose.auth.yml run --rm test-runner all

# Interactive shell for debugging
docker compose -f docker/compose/docker-compose.test.yml \
  -f docker/compose/docker-compose.auth.yml run --rm shell
```

### Timeout Protection for Hanging Tests

**CRITICAL**: Some FUSE filesystem tests may hang indefinitely. Use timeout protection:

```bash
# Use timeout wrapper for potentially hanging tests
./scripts/timeout-test-wrapper.sh "TestIT_FS_ETag_01_CacheValidationWithTimeoutFix" 60

# The wrapper provides:
# - Hard timeout enforcement (kills hanging processes)
# - Progress monitoring with heartbeat  
# - Container cleanup
# - Detailed logging to test-artifacts/debug/
```

### Environment Validation

The Docker test environment provides:

- **FUSE device**: `/dev/fuse` is accessible with proper permissions
- **Go environment**: Go 1.24.2 installed and configured
- **Python environment**: Python 3.12.3 with required packages
- **Workspace mounting**: Project source mounted at `/workspace`
- **Authentication reference**: Canonical tokens mounted via reference system
- **Test artifacts**: Output directory at `/tmp/home-tester/.onemount-tests`

Validation commands:

```bash
# Check FUSE device
docker compose -f docker/compose/docker-compose.test.yml run --rm shell \
  bash -c "ls -l /dev/fuse"

# Check authentication reference system
docker compose -f docker/compose/docker-compose.test.yml \
  -f docker/compose/docker-compose.auth.yml run --rm shell \
  bash -c "ls -la \$ONEMOUNT_AUTH_PATH && echo 'Auth: OK'"

# Check Go version
docker compose -f docker/compose/docker-compose.test.yml run --rm shell \
  bash -c "go version"

# Check Python version  
docker compose -f docker/compose/docker-compose.test.yml run --rm shell \
  bash -c "python3 --version"

# Check workspace
docker compose -f docker/compose/docker-compose.test.yml run --rm shell \
  bash -c "ls -la /workspace"
```

## Test Artifacts

Test artifacts are stored in `test-artifacts/` directory:

- `debug/`: Timeout wrapper logs and debugging information
- `logs/`: Test execution logs
- `system-test-data/`: System test data and cache
- `tmp/`: Temporary files created during tests

**Note**: Authentication tokens are no longer stored in `test-artifacts/`. They remain in their canonical location and are referenced via the authentication reference system.

## Troubleshooting

### Authentication Reference Issues

If system tests fail with authentication errors:

1. **Check reference system setup**:
   ```bash
   # Verify reference files exist
   ls -la docker/compose/docker-compose.auth.yml .env.auth
   
   # Check environment configuration
   cat .env.auth
   ```

2. **Verify canonical tokens exist and are valid**:
   ```bash
   # Check if setup script can find tokens
   ./scripts/setup-auth-reference.sh
   ```

3. **Test authentication in container**:
   ```bash
   docker compose -f docker/compose/docker-compose.test.yml \
     -f docker/compose/docker-compose.auth.yml run --rm shell \
     bash -c "ls -la \$ONEMOUNT_AUTH_PATH"
   ```

4. **Re-authenticate if tokens are expired**:
   ```bash
   ./build/onemount --auth-only
   ./scripts/setup-auth-reference.sh
   ```

### Docker Issues

If Docker tests fail:

1. Verify FUSE device is available: `ls -l /dev/fuse`
2. Check Docker has necessary capabilities: `--device=/dev/fuse --cap-add=SYS_ADMIN`
3. Rebuild images if dependencies changed: `./docker/scripts/build-images.sh test-runner`
4. Verify authentication reference system: `./scripts/setup-auth-reference.sh`

### Hanging Tests

If tests hang indefinitely:

1. **Use timeout wrapper**: `./scripts/timeout-test-wrapper.sh "TestPattern" 60`
2. **Check for FUSE deadlocks**: Look for goroutine dumps in logs
3. **Verify authentication**: Expired tokens can cause hangs during initialization
4. **Interactive debugging**: Use shell service to investigate manually

### Network Issues

If tests fail with network errors:

1. Check internet connectivity
2. Verify OneDrive API is accessible
3. Check firewall rules
4. Try with IPv4-only networking (already configured in docker-compose.test.yml)

## Security Best Practices

1. **Never use production credentials** for testing
2. **Use a dedicated test account** with minimal data
3. **Rotate test credentials** regularly
4. **Don't commit credentials** to version control
5. **Limit test account permissions** to minimum required
6. **Monitor test account activity** for suspicious access
7. **Use environment variables** for CI/CD credentials

## CI/CD Integration

For GitHub Actions or other CI/CD systems:

1. Store auth tokens as encrypted secrets
2. Mount secrets as environment variables
3. Use the `ONEMOUNT_AUTH_TOKENS` environment variable
4. Clean up test data after test runs
5. Use separate test accounts per CI environment

Example GitHub Actions secret setup:

```yaml
- name: Run system tests
  env:
    ONEMOUNT_AUTH_TOKENS: ${{ secrets.ONEDRIVE_TEST_TOKENS }}
  run: |
    docker compose -f docker/compose/docker-compose.test.yml run --rm system-tests
```

## Maintenance

### Regular Tasks

1. **Weekly**: Verify test credentials are still valid
2. **Monthly**: Refresh test data in OneDrive
3. **Quarterly**: Rotate test account credentials
4. **As needed**: Update Docker images when dependencies change

### Updating Test Data

To update test files in OneDrive:

1. Authenticate with test account
2. Mount OneDrive: `./build/onemount /mnt/onedrive`
3. Update files in `/mnt/onedrive/test-files/`
4. Unmount: `fusermount3 -uz /mnt/onedrive`

## References

- [Docker Test Environment Design](../../.kiro/specs/system-verification-and-fix/design.md)
- [Test Requirements](../../.kiro/specs/system-verification-and-fix/requirements.md)
- [Docker Compose Configuration](../../docker/compose/docker-compose.test.yml)
- [Test Entrypoint Script](../../packaging/docker/test-entrypoint.sh)
