# Enhanced Docker Development Workflow

This document describes the improved Docker development workflow for OneMount, which addresses the issues of slow builds, dependency reinstallation, and lack of container reuse.

## Overview of Improvements

### 1. Image Tagging and Versioning
- **Semantic versioning**: Images are tagged with git commit hash and dockerfile hash
- **Development vs Production**: Separate tags for development (`dev`) and production builds
- **Multiple tags**: Images receive multiple tags for better management

### 2. Container Reuse Strategy
- **Named containers**: Containers have predictable names for easy management
- **Reuse by default**: Existing containers are reused to avoid rebuild overhead
- **Selective recreation**: Options to force container recreation when needed

### 3. Build Optimization
- **Multi-stage Dockerfile**: Separates dependency installation from source code
- **Go module caching**: Dependencies are cached between builds
- **BuildKit optimization**: Uses Docker BuildKit for improved build performance
- **Pre-built binaries**: OneMount binaries are built during image creation

### 4. Enhanced CLI Options
- `--rebuild-image`: Force rebuild of Docker image
- `--recreate-container`: Force recreation of container
- `--no-reuse`: Disable container reuse (always create new containers)
- `--dev`: Use development mode with persistent containers

## Quick Start

### Initial Setup
```bash
# Set up development environment (one-time setup)
make docker-dev-setup
```

### Daily Development Workflow
```bash
# Run tests (fast - reuses existing container)
make docker-test-unit

# Run tests with fresh container (if needed)
make docker-test-unit-fresh

# Run all tests
make docker-test-all

# Clean up when done
make docker-test-clean
```

## Detailed Usage

### Building Images

#### Standard Build (Recommended)
```bash
# Build with automatic tagging (uses Docker Compose)
./scripts/dev.py test docker build

# Build development image
./scripts/dev.py test docker build --dev

# Force rebuild without cache
./scripts/dev.py test docker build --no-cache

# Use direct Docker build (bypass Compose)
./scripts/dev.py test docker build --no-compose
```

**Key Improvement**: Images are now built once and tagged for reuse. Docker Compose files reference these pre-built images by tag, eliminating unnecessary rebuilds during test execution.

#### Image Tags
Images are automatically tagged with:
- `onemount-test-runner:latest` - Latest production build
- `onemount-test-runner:dev` - Development build
- `onemount-test-runner:git-<hash>` - Git commit specific
- `onemount-test-runner:<commit>-<dockerfile-hash>` - Unique build identifier

### Running Tests

#### Container Reuse (Default)
```bash
# First run: Creates new container
./scripts/dev.py test docker unit

# Subsequent runs: Reuses existing container (fast!)
./scripts/dev.py test docker unit
```

#### Container Management
```bash
# Force container recreation
./scripts/dev.py test docker unit --recreate-container

# Disable container reuse
./scripts/dev.py test docker unit --no-reuse

# Use development mode
./scripts/dev.py test docker unit --dev
```

### Container Lifecycle

#### Container Names
Containers follow a predictable naming pattern:
- `onemount-unit-test` - Unit test container
- `onemount-unit-dev` - Development unit test container
- `onemount-integration-test` - Integration test container
- `onemount-system-test` - System test container

#### Container States
1. **New**: Container doesn't exist, will be created
2. **Stopped**: Container exists but not running, will be started
3. **Running**: Container is running, tests will be executed directly

## Performance Improvements

### Before (Issues)
- Go dependencies reinstalled every test run (~2-3 minutes)
- No container reuse, always creating new containers
- No image caching strategy
- Slow iterative development

### After (Improvements)
- Go dependencies cached in image layers (~30 seconds first build)
- Container reuse for subsequent runs (~5-10 seconds)
- Smart image tagging and caching
- Fast iterative development workflow

## Advanced Features

### Development Mode
```bash
# Enable development mode for persistent containers
./scripts/dev.py test docker unit --dev
```

Development mode provides:
- Persistent containers that survive between test runs
- Interactive debugging capabilities
- Faster test iteration

### Custom Image Tags
```bash
# Build with custom tag
./scripts/dev.py test docker build --tag my-custom-tag

# Use custom image for tests
ONEMOUNT_TEST_IMAGE=my-custom-tag ./scripts/dev.py test docker unit
```

### Docker Compose Strategy

#### Separate Build and Run Configurations
- **`docker-compose.test.yml`**: References pre-built images by tag, no build section
- **`docker-compose.build.yml`**: Contains build configurations for creating images
- **Benefits**: Tests run immediately using tagged images, builds only when explicitly requested

#### Environment Variables
- `ONEMOUNT_TEST_IMAGE`: Override default test image (e.g., `onemount-test-runner:my-tag`)
- `ONEMOUNT_CONTAINER_NAME`: Override default container name
- `ONEMOUNT_TEST_TIMEOUT`: Set test timeout
- `ONEMOUNT_TEST_VERBOSE`: Enable verbose output

## Troubleshooting

### Container Issues
```bash
# List all OneMount containers
docker ps -a --filter name=onemount-

# Remove stuck containers
docker rm -f $(docker ps -a --filter name=onemount- -q)

# Clean up everything
./scripts/dev.py test docker clean
```

### Image Issues
```bash
# List OneMount images
docker images onemount-*

# Remove old images
docker rmi $(docker images onemount-test-runner -q)

# Force rebuild
./scripts/dev.py test docker build --no-cache
```

### Performance Issues
```bash
# Check Docker system usage
docker system df

# Clean up Docker system
docker system prune -f

# Reset development environment
make docker-dev-reset
```

## Best Practices

### For Daily Development
1. Use `make docker-dev-setup` once to set up your environment
2. Use `make docker-test-unit` for fast iterative testing
3. Use `make docker-test-unit-fresh` when you need a clean environment
4. Use `make docker-test-clean` to clean up at the end of the day

### For CI/CD
1. Use `--no-reuse` flag to ensure clean environments
2. Use `--rebuild-image` for production builds
3. Tag images with build numbers or commit hashes
4. Clean up resources after each build

### For Debugging
1. Use `--dev` flag for development mode
2. Use `./scripts/dev.py test docker shell` for interactive debugging
3. Mount additional volumes for debugging tools
4. Use `--verbose` for detailed output

## Migration Guide

### From Old Workflow
If you were using the old Docker testing approach:

1. **Replace old commands**:
   ```bash
   # Old
   make docker-test-build
   make docker-test-unit
   
   # New (equivalent)
   make docker-test-build
   make docker-test-unit
   ```

2. **Take advantage of new features**:
   ```bash
   # Fast development workflow
   make docker-dev-setup
   make docker-test-unit  # Reuses container
   make docker-test-unit  # Even faster!
   ```

3. **Update scripts**: Replace direct Docker commands with the new CLI interface

### Configuration Changes
- Update CI/CD scripts to use new flags
- Set environment variables for custom configurations
- Update documentation to reflect new workflow
