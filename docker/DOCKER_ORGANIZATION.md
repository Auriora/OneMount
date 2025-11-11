# Docker Organization Guide

This document explains the organization of Docker files in the OneMount project and the separation between production and development environments.

## Directory Structure

```
OneMount/
├── docker/                              # Development Docker files
│   ├── Dockerfile.github-runner         # Development runner
│   ├── Dockerfile.test-runner           # Development test runner
│   ├── scripts/                         # Entrypoint and helper scripts
│   │   ├── runner-entrypoint.sh         # Runner entrypoint
│   │   ├── test-entrypoint.sh           # Test entrypoint
│   │   ├── init-workspace.sh            # Workspace initialization
│   │   ├── token-manager.sh             # Token management
│   │   ├── python-helper.sh             # Python helper utilities
│   │   └── build-entrypoint.sh          # Build entrypoint
│   ├── compose/                         # Docker Compose configurations
│   │   ├── docker-compose.build.yml     # Build binaries/packages
│   │   ├── docker-compose.test.yml      # Testing workflows
│   │   └── docker-compose.runners.yml   # Runners (dev/prod/remote)
│   ├── DOCKER_ORGANIZATION.md           # This file
│   ├── QUICK_REFERENCE.md               # Quick command reference
│   └── README.md                        # Docker usage documentation
│
├── packaging/docker/                    # Production Docker files
│   ├── Dockerfile.base                  # Base image (shared by all)
│   ├── docker-compose.yml               # Production deployment (no build)
│   ├── .dockerignore                    # Build context exclusions
│   └── README.md                        # Production deployment guide
│
└── packaging/deb/docker/                # Debian package builder
    ├── Dockerfile                       # Debian builder image
    ├── install-deps.sh                  # Legacy dependency script
    └── README.md                        # Builder documentation
```

## File Organization Principles

### Production Files (`packaging/docker/`)

**Purpose**: Files used for production deployment, packaging, and distribution

**Contents**:
- `Dockerfile.base` - Foundation image with Ubuntu 24.04, Go 1.24.2, FUSE3
- `Dockerfile.deb-builder` - Debian package builder
- `docker-compose.yml` - Production packaging and deployment
- `install-deps.sh` - Dependency installation for package building

**Characteristics**:
- Minimal, optimized for size and security
- No development tools
- Production-ready configurations
- Used by CI/CD pipelines
- Used for package building

### Development Files (`docker/`)

**Purpose**: Files used for local development and testing

**Contents**:
- `Dockerfile.github-runner` - Development runner with extra tools
- `Dockerfile.test-runner` - Development test runner with debugging tools
- `scripts/` - Entrypoint and helper scripts (shared by all containers)
- `compose/` - All Docker Compose configurations for development/testing

**Characteristics**:
- Extends production Dockerfiles
- Includes development tools (vim, less, etc.)
- More verbose logging
- Interactive debugging support
- Convenient for local development

## Docker Compose Files

All compose files now include:
- **Project name** (`name:` field) for better container organization
- **Container names** for all services for easy identification
- **Consistent naming** following `onemount-<purpose>-<variant>` pattern

### `docker-compose.build.yml`
- **Project**: `onemount-build`
- **Purpose**: Building Docker images
- **Services**: base-build, test-runner-build, test-runner-dev-build, test-runner-no-cache-build
- **Usage**: `docker compose -f docker/compose/docker-compose.build.yml build`

### `docker-compose.test.yml`
- **Project**: `onemount-test`
- **Purpose**: Running tests in isolated environments
- **Services**: test-runner, unit-tests, integration-tests, system-tests, coverage, shell
- **Usage**: `docker compose -f docker/compose/docker-compose.test.yml run --rm unit-tests`

### `docker-compose.runner.yml` (Singular)
- **Project**: `onemount-runner`
- **Purpose**: Single GitHub Actions runner for development/testing
- **Services**: github-runner, runner-dev, base-build
- **Usage**: Development and debugging workflows
- **Note**: For production, use `docker-compose.runners.yml` instead

### `docker-compose.runners.yml` (Plural)
- **Project**: `onemount-runners`
- **Purpose**: Production multi-runner setup (2 runners)
- **Services**: runner-1, runner-2
- **Usage**: Production CI/CD with multiple concurrent runners
- **Features**: Persistent volumes, restart policies, proper secret management

### `docker-compose.remote.yml`
- **Project**: `onemount-remote`
- **Purpose**: Remote Docker host deployment
- **Services**: github-runner
- **Usage**: Deployment to remote Docker hosts via TCP API

## Dockerfile Relationships

### Base Image (Shared)
```
packaging/docker/Dockerfile.base
├── Ubuntu 24.04
├── Go 1.24.2
├── FUSE3 support
└── Essential build tools
```

### Production Images
```
packaging/docker/Dockerfile.base
└── packaging/docker/Dockerfile.deb-builder
    └── Package building tools
```

### Development Images
```
packaging/docker/Dockerfile.base (shared)
├── docker/Dockerfile.test-runner
│   └── Test environment with debugging tools (vim, less, etc.)
└── docker/Dockerfile.github-runner
    └── GitHub Actions runner with debugging tools (vim, less, etc.)
```

**Note**: Development images extend the shared base image directly, not production images. The separation is organizational - production files are for packaging/deployment, development files are for local testing.

## Container Naming Convention

All containers follow this pattern: `onemount-<purpose>-<variant>`

**Examples**:
- `onemount-test-runner` - Main test runner
- `onemount-unit-tests` - Unit test execution
- `onemount-integration-tests` - Integration test execution
- `onemount-system-tests` - System test execution
- `onemount-runner-1` - Production runner #1
- `onemount-runner-2` - Production runner #2
- `onemount-github-runner` - Development runner
- `onemount-base-build` - Base image builder

## Usage Examples

### Building Images

```bash
# Build all images
docker compose -f docker/compose/docker-compose.build.yml build

# Build specific image
docker compose -f docker/compose/docker-compose.build.yml build test-runner-build

# Build without cache
docker compose -f docker/compose/docker-compose.build.yml build --no-cache
```

### Running Tests

```bash
# Unit tests
docker compose -f docker/compose/docker-compose.test.yml run --rm unit-tests

# Integration tests
docker compose -f docker/compose/docker-compose.test.yml run --rm integration-tests

# System tests
docker compose -f docker/compose/docker-compose.test.yml run --rm system-tests

# Interactive shell
docker compose -f docker/compose/docker-compose.test.yml run --rm shell
```

### Managing Runners

```bash
# Development runner (single)
docker compose -f docker/compose/docker-compose.runner.yml up -d

# Production runners (multiple)
docker compose -f docker/compose/docker-compose.runners.yml up -d

# Check status
docker ps --filter "name=onemount-runner"

# View logs
docker logs onemount-runner-1
```

## Migration Notes

### From Old Structure

If you have existing scripts or documentation referencing the old structure:

**Old**: `packaging/docker/Dockerfile.github-runner` (for development)
**New**: `docker/Dockerfile.github-runner` (development) or `packaging/docker/Dockerfile.github-runner` (production)

**Old**: Compose files without project names
**New**: All compose files have `name:` field

**Old**: Containers with auto-generated names
**New**: All containers have explicit `container_name:` field

### Updating Scripts

If you have scripts that reference Docker files:

```bash
# Old
docker build -f packaging/docker/Dockerfile.github-runner .

# New (development)
docker build -f docker/Dockerfile.github-runner .

# New (production)
docker build -f packaging/docker/Dockerfile.github-runner .
```

## Best Practices

### When to Use Production vs Development

**Use Production Dockerfiles when**:
- Building for CI/CD pipelines
- Creating distribution packages
- Deploying to production environments
- Optimizing for size and security

**Use Development Dockerfiles when**:
- Local development and testing
- Debugging issues
- Interactive exploration
- Learning the codebase

### Container Management

```bash
# List all OneMount containers
docker ps -a --filter "name=onemount"

# Stop all OneMount containers
docker stop $(docker ps -q --filter "name=onemount")

# Remove all OneMount containers
docker rm $(docker ps -aq --filter "name=onemount")

# Clean up volumes
docker volume prune --filter "label=com.docker.compose.project=onemount-*"
```

### Image Management

```bash
# List all OneMount images
docker images | grep onemount

# Remove unused images
docker image prune -a --filter "label=org.opencontainers.image.vendor=Auriora"

# Check image sizes
docker images --format "table {{.Repository}}\t{{.Tag}}\t{{.Size}}" | grep onemount
```

## Troubleshooting

### Container Name Conflicts

If you get "container name already in use" errors:

```bash
# Remove the conflicting container
docker rm -f onemount-test-runner

# Or use --force-recreate
docker compose -f docker/compose/docker-compose.test.yml up --force-recreate
```

### Image Not Found

If you get "image not found" errors:

```bash
# Build the required images first
docker compose -f docker/compose/docker-compose.build.yml build

# Or pull from registry (if available)
docker pull onemount-base:latest
```

### Permission Issues

If you encounter permission issues:

```bash
# Check user mapping
docker compose -f docker/compose/docker-compose.test.yml run --rm shell
id  # Should show tester user

# Fix volume permissions
sudo chown -R $(id -u):$(id -g) test-artifacts/
```

## References

- **Docker Documentation**: `docker/README.md`
- **Compose Documentation**: `docker/compose/README.md`
- **Packaging Documentation**: `packaging/docker/README.md`
- **Testing Guide**: `docs/TEST_SETUP.md`
- **Docker Test Environment**: `docs/testing/docker-test-environment.md`
