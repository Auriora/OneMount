# OneMount Test Environment Setup

## Overview

This document describes the test environment setup for OneMount, including Docker configuration, authentication, and test data.

## Test Credentials

### Authentication Tokens

Authentication tokens for system tests are stored in `test-artifacts/.auth_tokens.json`. This file contains:

- `access_token`: OAuth2 access token for OneDrive API
- `refresh_token`: OAuth2 refresh token for token renewal
- `expires_at`: Token expiration timestamp
- `config`: Configuration details
- `account`: Account information

**SECURITY WARNING**: 
- Use a dedicated test OneDrive account, NOT your production account
- Never commit production credentials to version control
- The `.auth_tokens.json` file is in `.gitignore` to prevent accidental commits
- Tokens should be refreshed regularly to ensure tests can run

### Token Validation

To verify tokens are valid:

```bash
python3 -c "import json; data = json.load(open('test-artifacts/.auth_tokens.json')); print('Valid JSON'); print('Keys:', list(data.keys()))"
```

### Token Refresh

If tokens are expired, you'll need to re-authenticate:

1. Run OneMount with authentication: `./build/onemount --auth`
2. Complete the OAuth2 flow
3. Copy the new tokens to `test-artifacts/.auth_tokens.json`

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
# Build base image
docker compose -f docker/compose/docker-compose.build.yml build base-build

# Build test runner
docker compose -f docker/compose/docker-compose.build.yml build test-runner-build
```

### Running Tests

```bash
# Unit tests (no FUSE required)
docker compose -f docker/compose/docker-compose.test.yml run --rm unit-tests

# Integration tests (requires FUSE)
docker compose -f docker/compose/docker-compose.test.yml run --rm integration-tests

# System tests (requires auth tokens)
docker compose -f docker/compose/docker-compose.test.yml run --rm system-tests

# All tests
docker compose -f docker/compose/docker-compose.test.yml run --rm test-runner all

# Interactive shell for debugging
docker compose -f docker/compose/docker-compose.test.yml run --rm shell
```

### Environment Validation

The Docker test environment provides:

- **FUSE device**: `/dev/fuse` is accessible with proper permissions
- **Go environment**: Go 1.24.2 installed and configured
- **Python environment**: Python 3.12.3 with required packages
- **Workspace mounting**: Project source mounted at `/workspace`
- **Test artifacts**: Output directory at `/tmp/home-tester/.onemount-tests`

Validation commands:

```bash
# Check FUSE device
docker compose -f docker/compose/docker-compose.test.yml run --rm --entrypoint /bin/bash shell -c "ls -l /dev/fuse"

# Check Go version
docker compose -f docker/compose/docker-compose.test.yml run --rm --entrypoint /bin/bash shell -c "go version"

# Check Python version
docker compose -f docker/compose/docker-compose.test.yml run --rm --entrypoint /bin/bash shell -c "python3 --version"

# Check workspace
docker compose -f docker/compose/docker-compose.test.yml run --rm --entrypoint /bin/bash shell -c "ls -la /workspace"

# Check test artifacts directory
docker compose -f docker/compose/docker-compose.test.yml run --rm --entrypoint /bin/bash shell -c "ls -la /tmp/home-tester/.onemount-tests"
```

## Test Artifacts

Test artifacts are stored in `test-artifacts/` directory:

- `.auth_tokens.json`: Authentication tokens (gitignored)
- `logs/`: Test execution logs
- `system-test-data/`: System test data and cache
- `tmp/`: Temporary files created during tests

## Troubleshooting

### Auth Token Issues

If system tests fail with authentication errors:

1. Check token expiration: `python3 -c "import json, time; data = json.load(open('test-artifacts/.auth_tokens.json')); print('Expired' if data['expires_at'] < time.time() else 'Valid')"`
2. Refresh tokens by re-authenticating
3. Verify token file permissions: `ls -l test-artifacts/.auth_tokens.json` (should be 600)

### Docker Issues

If Docker tests fail:

1. Verify FUSE device is available: `ls -l /dev/fuse`
2. Check Docker has necessary capabilities: `--device=/dev/fuse --cap-add=SYS_ADMIN`
3. Rebuild images if dependencies changed: `docker compose -f docker/compose/docker-compose.build.yml build --no-cache`

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
