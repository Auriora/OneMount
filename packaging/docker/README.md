# OneMount Docker Testing

This directory contains Docker configurations for running OneMount tests in isolated environments. The Docker-based testing provides a clean, reproducible environment that eliminates host system dependencies and ensures consistent test results.

## Overview

The Docker testing setup includes:

- **Dockerfile.test-runner**: Main test container with all dependencies
- **test-entrypoint.sh**: Test execution script with unified interface
- **docker-compose.test.yml**: Docker Compose configuration for easy test execution
- **run-tests-docker.sh**: Wrapper script for convenient Docker test execution

## Quick Start

### 1. Build the Test Image

```bash
# Using Make
make docker-test-build

# Or directly
./scripts/run-tests-docker.sh build
```

### 2. Run Tests

```bash
# Run unit tests
make docker-test-unit

# Run integration tests
make docker-test-integration

# Run all tests
make docker-test-all

# Run with coverage analysis
make docker-test-coverage
```

## Available Test Commands

### Make Targets

| Target | Description |
|--------|-------------|
| `make docker-test-build` | Build the Docker test image |
| `make docker-test-unit` | Run unit tests in Docker |
| `make docker-test-integration` | Run integration tests in Docker |
| `make docker-test-system` | Run system tests in Docker |
| `make docker-test-all` | Run all tests in Docker |
| `make docker-test-coverage` | Run coverage analysis in Docker |
| `make docker-test-shell` | Start interactive shell in test container |
| `make docker-test-clean` | Clean up Docker test resources |

### Script Interface

```bash
# Build test image
./scripts/run-tests-docker.sh build

# Run different test types
./scripts/run-tests-docker.sh unit
./scripts/run-tests-docker.sh integration
./scripts/run-tests-docker.sh system
./scripts/run-tests-docker.sh all

# Run with options
./scripts/run-tests-docker.sh unit --verbose --sequential
./scripts/run-tests-docker.sh system --timeout 30m

# Interactive debugging
./scripts/run-tests-docker.sh shell

# Cleanup
./scripts/run-tests-docker.sh clean
```

### Docker Compose

```bash
# Run specific test services
docker-compose -f docker-compose.test.yml run --rm unit-tests
docker-compose -f docker-compose.test.yml run --rm integration-tests
docker-compose -f docker-compose.test.yml run --rm system-tests

# Interactive shell
docker-compose -f docker-compose.test.yml run --rm shell
```

## System Tests with OneDrive Authentication

System tests require valid OneDrive authentication tokens. Follow these steps:

### 1. Setup Authentication

```bash
# Build OneMount
make onemount

# Authenticate with your test OneDrive account
./build/onemount --auth-only

# Create test directory and copy tokens
mkdir -p ~/.onemount-tests
cp ~/.cache/onemount/auth_tokens.json ~/.onemount-tests/.auth_tokens.json
```

### 2. Run System Tests

```bash
# Using Make
make docker-test-system

# Or directly
./scripts/run-tests-docker.sh system
```

### Authentication Notes

- Use a dedicated test OneDrive account, not your production account
- System tests create and delete files in `/onemount_system_tests/` on OneDrive
- The auth tokens file is automatically mounted into the Docker container
- Run `./scripts/run-tests-docker.sh setup-auth` for detailed setup instructions

## Container Features

### Included Dependencies

- **Go 1.22+**: Latest Go version from Ubuntu 24.04
- **FUSE Support**: Full FUSE3 support for filesystem testing
- **GUI Dependencies**: WebKit2GTK for launcher testing
- **Build Tools**: Complete build environment with CGO support
- **Network Tools**: IPv4-only configuration for South African networks

### Security and Isolation

- **Non-root User**: Tests run as `tester` user for security
- **FUSE Access**: Proper FUSE device access for filesystem tests
- **Volume Mounts**: Source code mounted read-only for safety
- **Network Isolation**: Isolated network environment

### Performance Optimizations

- **BuildKit**: Enabled for faster image builds
- **Layer Caching**: Optimized Dockerfile for better caching
- **Parallel Tests**: Support for parallel test execution
- **Resource Limits**: Configurable timeouts and resource limits

## Troubleshooting

### Common Issues

#### Docker Permission Errors
```bash
# Add user to docker group
sudo usermod -aG docker $USER
newgrp docker

# Or run with sudo
sudo ./scripts/run-tests-docker.sh unit
```

#### FUSE Not Available
```bash
# Run with FUSE support
docker run --device /dev/fuse --cap-add SYS_ADMIN --security-opt apparmor:unconfined onemount-test-runner
```

#### System Test Authentication Failures
```bash
# Check auth tokens exist
ls -la ~/.onemount-tests/.auth_tokens.json

# Re-authenticate if needed
./build/onemount --auth-only
cp ~/.cache/onemount/auth_tokens.json ~/.onemount-tests/.auth_tokens.json
```

#### Network Issues
```bash
# Check IPv4 connectivity
docker run --rm onemount-test-runner ping -c 3 8.8.8.8

# Force IPv4-only DNS
docker run --rm --dns 8.8.8.8 onemount-test-runner
```

### Debug Mode

Start an interactive shell in the test container for debugging:

```bash
# Using Make
make docker-test-shell

# Or directly
./scripts/run-tests-docker.sh shell

# Inside the container
cd /workspace
go test -v ./pkg/... -run TestSpecificFunction
```

### Logs and Artifacts

Test artifacts are stored in:
- **Host**: `./test-artifacts/` directory
- **Container**: `/home/tester/.onemount-tests/` directory

Coverage reports are generated in:
- **Host**: `./coverage/` directory
- **Container**: `/workspace/coverage/` directory

## Advanced Usage

### Custom Test Execution

```bash
# Run specific test patterns
docker run --rm -v $(pwd):/workspace:ro onemount-test-runner \
  go test -v ./pkg/... -run TestPattern

# Run with custom environment
docker run --rm -v $(pwd):/workspace:ro \
  -e ONEMOUNT_TEST_TIMEOUT=60m \
  -e ONEMOUNT_TEST_VERBOSE=true \
  onemount-test-runner all
```

### CI/CD Integration

The Docker test setup is designed for CI/CD pipelines:

```yaml
# Example GitHub Actions step
- name: Run Docker Tests
  run: |
    make docker-test-build
    make docker-test-unit
    make docker-test-integration
```

### Performance Testing

```bash
# Run with performance monitoring
docker run --rm -v $(pwd):/workspace:ro \
  --cpus="2" --memory="4g" \
  onemount-test-runner coverage --verbose
```

## Container Architecture

The test container is based on Ubuntu 24.04 and includes:

1. **Base System**: Ubuntu 24.04 LTS with IPv4-only networking
2. **Go Environment**: Go 1.22+ with proper module support
3. **FUSE Support**: FUSE3 with user permissions configured
4. **GUI Dependencies**: WebKit2GTK and GTK3 for launcher testing
5. **Test User**: Non-root `tester` user with appropriate permissions
6. **Test Directories**: Pre-configured test artifact directories

The container provides complete isolation from the host system while maintaining access to necessary devices and capabilities for comprehensive testing.
