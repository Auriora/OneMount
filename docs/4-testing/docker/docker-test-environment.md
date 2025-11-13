# Docker Test Environment Guide

## Overview

OneMount uses Docker containers to provide isolated, reproducible test environments. This ensures tests run consistently across different machines and don't affect the host system.

## Architecture

### Image Hierarchy

```
ubuntu:24.04
    ↓
onemount-base:latest (1.49GB)
    ├─ Ubuntu 24.04
    ├─ Go 1.24.2
    ├─ FUSE3 support
    ├─ GTK3 & WebKit2GTK
    └─ Build tools
    ↓
onemount-test-runner:latest (2.21GB)
    ├─ Python 3.12
    ├─ Test utilities
    ├─ Pre-built binaries
    └─ Test entrypoint
```

### Container Services

The `docker-compose.test.yml` defines several services:

1. **test-runner**: Base service with common configuration
2. **unit-tests**: Runs unit tests (no FUSE required)
3. **integration-tests**: Runs integration tests (requires FUSE)
4. **system-tests**: Runs system tests (requires auth tokens)
5. **coverage**: Generates coverage reports
6. **shell**: Interactive shell for debugging

## Building Images

### Prerequisites

- Docker Engine 20.10+
- Docker Compose V2
- 10GB free disk space
- Internet connection for downloading dependencies

### Build Commands

```bash
# Build base image
docker compose -f docker/compose/docker-compose.build.yml build base-build

# Build test runner
docker compose -f docker/compose/docker-compose.build.yml build test-runner-build

# Build both images
docker compose -f docker/compose/docker-compose.build.yml build

# Rebuild without cache (clean build)
docker compose -f docker/compose/docker-compose.build.yml build --no-cache
```

### Build Profiles

The build configuration supports different profiles:

- **build**: Standard build with caching
- **build-dev**: Development build variant
- **build-no-cache**: Clean rebuild without cache

```bash
# Use specific profile
docker compose -f docker/compose/docker-compose.build.yml --profile build-dev build
```

### Verifying Images

After building, verify images exist:

```bash
docker images | grep onemount
```

Expected output:
```
onemount-test-runner    latest    <image-id>    <time>    2.21GB
onemount-base           latest    <image-id>    <time>    1.49GB
```

Check image layers:

```bash
docker history onemount-test-runner:latest
```

## Running Tests

### Unit Tests

Unit tests are lightweight and don't require FUSE:

```bash
docker compose -f docker/compose/docker-compose.test.yml run --rm unit-tests
```

Options:
- `--verbose`: Enable verbose output
- `--timeout DURATION`: Set custom timeout (default: 5m)
- `--sequential`: Run tests sequentially

Example with options:
```bash
docker compose -f docker/compose/docker-compose.test.yml run --rm unit-tests --verbose --timeout 10m
```

### Integration Tests

Integration tests require FUSE device access:

```bash
docker compose -f docker/compose/docker-compose.test.yml run --rm integration-tests
```

The container is automatically configured with:
- `/dev/fuse` device
- `SYS_ADMIN` capability
- `apparmor:unconfined` security option

### System Tests

System tests require OneDrive authentication:

```bash
# Ensure auth tokens exist
ls test-artifacts/.auth_tokens.json

# Run system tests
docker compose -f docker/compose/docker-compose.test.yml run --rm system-tests
```

System tests have:
- Longer timeout (30m default)
- More resources (6GB RAM, 4 CPUs)
- Auth token mounting

### All Tests

Run all test types in sequence:

```bash
docker compose -f docker/compose/docker-compose.test.yml run --rm test-runner all
```

### Coverage Analysis

Generate coverage reports:

```bash
docker compose -f docker/compose/docker-compose.test.yml run --rm coverage
```

Coverage reports are saved to:
- `coverage/coverage.out`: Raw coverage data
- `coverage/coverage.html`: HTML report

View HTML report:
```bash
open coverage/coverage.html  # macOS
xdg-open coverage/coverage.html  # Linux
```

## Interactive Debugging

### Shell Access

Start an interactive shell in the test container:

```bash
docker compose -f docker/compose/docker-compose.test.yml run --rm shell
```

Inside the shell:
```bash
# Check environment
go version
python3 --version
ls -l /dev/fuse

# Build OneMount
bash scripts/cgo-helper.sh
go build -o build/onemount ./cmd/onemount

# Run specific tests
go test -v ./internal/fs -run TestCache

# Check auth tokens
ls -la ~/.onemount-tests/.auth_tokens.json

# Exit shell
exit
```

### Running Custom Commands

Execute specific commands in the container:

```bash
# Override entrypoint to run custom commands
docker compose -f docker/compose/docker-compose.test.yml run --rm \
  --entrypoint /bin/bash shell -c "go test -v ./internal/graph"
```

### Debugging Test Failures

1. **Run tests with verbose output**:
   ```bash
   docker compose -f docker/compose/docker-compose.test.yml run --rm unit-tests --verbose
   ```

2. **Check test logs**:
   ```bash
   ls -la test-artifacts/logs/
   cat test-artifacts/logs/fusefs_tests.log
   ```

3. **Run specific test**:
   ```bash
   docker compose -f docker/compose/docker-compose.test.yml run --rm \
     --entrypoint /bin/bash shell -c "go test -v -run TestSpecificTest ./internal/fs"
   ```

4. **Enable race detector**:
   ```bash
   docker compose -f docker/compose/docker-compose.test.yml run --rm \
     --entrypoint /bin/bash shell -c "go test -race -v ./..."
   ```

## Environment Configuration

### Environment Variables

Configure test behavior with environment variables:

```bash
# Set in docker-compose.test.yml or pass via -e flag
ONEMOUNT_TEST_TIMEOUT=10m      # Test timeout
ONEMOUNT_TEST_VERBOSE=true     # Verbose output
GORACE=log_path=race.log       # Race detector config
GOGC=50                        # Go garbage collector
GOMEMLIMIT=2GiB               # Memory limit
DOCKER_CONTAINER=true          # Flag for container environment
```

Example:
```bash
docker compose -f docker/compose/docker-compose.test.yml run --rm \
  -e ONEMOUNT_TEST_VERBOSE=true \
  -e ONEMOUNT_TEST_TIMEOUT=15m \
  unit-tests
```

### Volume Mounts

The test container mounts:

1. **Workspace**: `/workspace` (read-write)
   - Contains project source code
   - Changes persist to host

2. **Test Artifacts**: `/tmp/home-tester/.onemount-tests` (read-write)
   - Test logs and output
   - Auth tokens
   - Temporary files

3. **Coverage**: `/workspace/coverage` (read-write, coverage service only)
   - Coverage reports

### Resource Limits

Default resource limits:

**Unit/Integration Tests**:
- Memory: 4GB limit, 1GB reservation
- CPU: 2 cores limit, 0.5 cores reservation

**System Tests**:
- Memory: 6GB limit, 2GB reservation
- CPU: 4 cores limit, 1 core reservation

Override in docker-compose.test.yml if needed.

### Network Configuration

The test environment uses:
- Bridge networking mode
- IPv4-only DNS (8.8.8.8, 8.8.4.4)
- Optimized for South African networks

## Test Artifacts

### Directory Structure

```
test-artifacts/
├── .auth_tokens.json          # Auth tokens (gitignored)
├── logs/                      # Test logs
│   ├── fusefs_tests.log
│   └── fusefs_tests.race
├── system-test-data/          # System test data
│   ├── cache/
│   └── mount/
└── tmp/                       # Temporary files
```

### Accessing Artifacts

From host:
```bash
ls -la test-artifacts/
cat test-artifacts/logs/fusefs_tests.log
```

From container:
```bash
ls -la /tmp/home-tester/.onemount-tests/
```

### Cleaning Artifacts

```bash
# Clean logs
rm -rf test-artifacts/logs/*

# Clean system test data
rm -rf test-artifacts/system-test-data/*

# Clean temporary files
rm -rf test-artifacts/tmp/*

# Keep auth tokens!
```

## Troubleshooting

### Common Issues

#### 1. FUSE Device Not Available

**Error**: `/dev/fuse: No such file or directory`

**Solution**:
```bash
# Check FUSE on host
ls -l /dev/fuse

# Ensure FUSE module is loaded
sudo modprobe fuse

# Verify Docker has device access
docker compose -f docker/compose/docker-compose.test.yml run --rm \
  --entrypoint /bin/bash shell -c "ls -l /dev/fuse"
```

#### 2. Permission Denied

**Error**: `Permission denied` when accessing files

**Solution**:
```bash
# Check user/group IDs
id

# Set USER_ID and GROUP_ID in .env file
echo "USER_ID=$(id -u)" >> docker/compose/.env
echo "GROUP_ID=$(id -g)" >> docker/compose/.env
```

#### 3. Auth Token Issues

**Error**: `Auth tokens not found` or `Invalid credentials`

**Solution**:
```bash
# Verify token file exists
ls -la test-artifacts/.auth_tokens.json

# Check token format
python3 -c "import json; json.load(open('test-artifacts/.auth_tokens.json'))"

# Check token expiration
python3 -c "import json, time; data = json.load(open('test-artifacts/.auth_tokens.json')); print('Expired' if data['expires_at'] < time.time() else 'Valid')"

# Re-authenticate if needed
./build/onemount --auth
```

#### 4. Build Failures

**Error**: Build fails with dependency errors

**Solution**:
```bash
# Clean build without cache
docker compose -f docker/compose/docker-compose.build.yml build --no-cache

# Check Docker disk space
docker system df

# Clean up old images
docker system prune -a
```

#### 5. Network Timeouts

**Error**: Timeout downloading dependencies

**Solution**:
```bash
# Use IPv4-only (already configured)
# Or set custom DNS
docker compose -f docker/compose/docker-compose.test.yml run --rm \
  --dns 1.1.1.1 \
  unit-tests
```

### Debug Mode

Enable debug output:

```bash
# Set verbose mode
export ONEMOUNT_TEST_VERBOSE=true

# Run with debug logging
docker compose -f docker/compose/docker-compose.test.yml run --rm \
  -e ONEMOUNT_TEST_VERBOSE=true \
  unit-tests
```

### Container Inspection

Inspect running container:

```bash
# List running containers
docker ps

# Inspect container
docker inspect <container-id>

# View container logs
docker logs <container-id>

# Execute command in running container
docker exec -it <container-id> /bin/bash
```

## Best Practices

### 1. Use Specific Test Types

Run only the tests you need:
- Unit tests for quick feedback
- Integration tests for component interactions
- System tests for end-to-end validation

### 2. Clean Artifacts Regularly

```bash
# Clean before running tests
rm -rf test-artifacts/logs/*
rm -rf test-artifacts/tmp/*
```

### 3. Monitor Resource Usage

```bash
# Check container resource usage
docker stats

# Check disk usage
docker system df
```

### 4. Keep Images Updated

```bash
# Rebuild when dependencies change
docker compose -f docker/compose/docker-compose.build.yml build

# Pull base image updates
docker pull ubuntu:24.04
```

### 5. Use .env File

Create `docker/compose/.env` for custom configuration:

```bash
USER_ID=1000
GROUP_ID=1000
ONEMOUNT_VERSION=0.1.0rc1
ONEMOUNT_TEST_TIMEOUT=10m
```

## CI/CD Integration

### GitHub Actions

Example workflow:

```yaml
name: Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Build test images
        run: |
          docker compose -f docker/compose/docker-compose.build.yml build
      
      - name: Run unit tests
        run: |
          docker compose -f docker/compose/docker-compose.test.yml run --rm unit-tests
      
      - name: Run integration tests
        run: |
          docker compose -f docker/compose/docker-compose.test.yml run --rm integration-tests
      
      - name: Upload test artifacts
        if: always()
        uses: actions/upload-artifact@v3
        with:
          name: test-artifacts
          path: test-artifacts/logs/
```

### GitLab CI

Example `.gitlab-ci.yml`:

```yaml
test:
  image: docker:latest
  services:
    - docker:dind
  script:
    - docker compose -f docker/compose/docker-compose.build.yml build
    - docker compose -f docker/compose/docker-compose.test.yml run --rm unit-tests
  artifacts:
    paths:
      - test-artifacts/logs/
    when: always
```

## References

- [Test Setup Documentation](../../test-artifacts/TEST_SETUP.md)
- [Docker Compose Test Configuration](../../docker/compose/docker-compose.test.yml)
- [Docker Compose Build Configuration](../../docker/compose/docker-compose.build.yml)
- [Base Dockerfile](../../packaging/docker/Dockerfile)
- [Test Runner Dockerfile](../../packaging/docker/Dockerfile.test-runner)
- [Test Entrypoint Script](../../packaging/docker/test-entrypoint.sh)
- [System Verification Design](../../.kiro/specs/system-verification-and-fix/design.md)
