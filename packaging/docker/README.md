# OneMount Production Docker Packaging

This directory contains Docker configurations for **production packaging and deployment** of OneMount.

For **development and testing** Docker files, see `docker/` directory.

## Overview

The production Docker setup includes:

- **Dockerfile.base**: Foundation image with Ubuntu 24.04, Go 1.24.2, FUSE3
- **Dockerfile.deb-builder**: Debian package builder
- **docker-compose.yml**: Production packaging and deployment
- **install-deps.sh**: Dependency installation script

## Quick Start

### 1. Build Production Images

```bash
# Build base image
docker compose -f packaging/docker/docker-compose.yml build base

# Build Debian package builder
docker compose -f packaging/docker/docker-compose.yml build deb-builder

# Build production deployment image
docker compose -f packaging/docker/docker-compose.yml build production
```

### 2. Build Packages

```bash
# Build Debian package
docker compose -f packaging/docker/docker-compose.yml run --rm deb-builder

# Output will be in dist/ directory
ls -la dist/*.deb
```

### 3. Run Production Container

```bash
# Start production container
docker compose -f packaging/docker/docker-compose.yml --profile production up -d

# Check status
docker ps --filter "name=onemount-production"

# View logs
docker logs onemount-production
```

## Available Commands

### Building Images

```bash
# Build base image
docker compose -f packaging/docker/docker-compose.yml build base

# Build Debian package builder
docker compose -f packaging/docker/docker-compose.yml build deb-builder

# Build production deployment image
docker compose -f packaging/docker/docker-compose.yml build production

# Build all images
docker compose -f packaging/docker/docker-compose.yml build
```

### Building Packages

```bash
# Build Debian package
docker compose -f packaging/docker/docker-compose.yml run --rm deb-builder

# Build with specific version
ONEMOUNT_VERSION=1.0.0 docker compose -f packaging/docker/docker-compose.yml run --rm deb-builder

# Output location
ls -la dist/*.deb
```

### Running Production Container

```bash
# Start production container
docker compose -f packaging/docker/docker-compose.yml --profile production up -d

# Stop production container
docker compose -f packaging/docker/docker-compose.yml --profile production down

# View logs
docker logs onemount-production

# Check health
docker inspect onemount-production | jq '.[0].State.Health'
```

## Production Deployment

### Prerequisites

- Docker Engine 20.10+
- Docker Compose V2
- Valid OneDrive authentication tokens (for production use)

### Configuration

1. **Set environment variables**:
   ```bash
   export ONEMOUNT_VERSION=1.0.0
   ```

2. **Configure volumes**:
   - `onemount-data`: Mount point for OneDrive files
   - `onemount-config`: Configuration files
   - `onemount-cache`: Cache directory

3. **Deploy**:
   ```bash
   docker compose -f packaging/docker/docker-compose.yml --profile production up -d
   ```

## Container Features

### Base Image

- **Ubuntu 24.04 LTS**: Stable foundation
- **Go 1.24.2**: Specific version for reproducibility
- **FUSE3 Support**: Full FUSE support for filesystem operations
- **GUI Dependencies**: WebKit2GTK for launcher
- **Build Tools**: Complete build environment with CGO support
- **IPv4-only networking**: Optimized for South African networks

### Debian Package Builder

- **Debian packaging tools**: debhelper, devscripts, dpkg-dev
- **Automated builds**: Single command package creation
- **Output to dist/**: Built packages in project directory
- **Non-root builder**: Security-focused build process

### Production Image

- **Minimal runtime**: Only essential dependencies
- **FUSE support**: Full filesystem capabilities
- **Resource limits**: 2GB RAM, 2 CPUs
- **Health checks**: Automatic mount point monitoring
- **Restart policy**: Automatic recovery from failures

## Troubleshooting

### Common Issues

#### Docker Permission Errors
```bash
# Add user to docker group
sudo usermod -aG docker $USER
newgrp docker

# Or run with sudo
sudo python scripts/dev.py test docker unit
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

# Using Docker Compose directly (Recommended)
docker compose -f docker/compose/docker-compose.test.yml run --rm shell

# Or legacy script (deprecated)
# ./scripts/run-tests-docker.sh shell

# Inside the container
cd /workspace
go test -v ./internal/... -run TestSpecificFunction
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
  go test -v ./internal/... -run TestPattern

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
