# Docker Configuration for OneMount

This directory contains Docker configurations for OneMount development, testing, and CI/CD.

For production deployment files, see `packaging/docker/`.

## Directory Structure

```
docker/
├── images/                          # Docker image definitions
│   ├── test-runner/                 # Test execution image
│   │   ├── Dockerfile
│   │   └── README.md
│   ├── github-runner/               # CI/CD runner image
│   │   ├── Dockerfile
│   │   └── README.md
│   └── deb-builder/                 # Debian package builder
│       ├── Dockerfile
│       └── README.md
├── scripts/                         # Container scripts
│   ├── common.sh                    # Shared functions (NEW)
│   ├── build-images.sh              # Build all images (NEW)
│   ├── runner-entrypoint.sh         # Runner container entrypoint
│   ├── test-entrypoint.sh           # Test container entrypoint
│   ├── build-entrypoint.sh          # Build container entrypoint
│   ├── init-workspace.sh            # Workspace initialization
│   ├── token-manager.sh             # Token management utilities
│   └── python-helper.sh             # Python helper utilities
├── compose/                         # Docker Compose configurations
│   ├── docker-compose.build.yml     # Build binaries and packages
│   ├── docker-compose.test.yml      # Run tests
│   └── docker-compose.runners.yml   # GitHub Actions runners
└── README.md                        # This file
```

**Note**: Production and builder images are in `packaging/docker/`:
- `Dockerfile` - Production runtime (minimal, no build tools)
- `Dockerfile.builder` - Builder base (with build tools, used by dev images)

## Docker Images

### Production Image
- **Location**: `packaging/docker/Dockerfile`
- **Type**: Multi-stage (builder + runtime)
- **Runtime includes**: FUSE3 runtime, GUI runtime, compiled binaries only
- **Runtime excludes**: Build tools, Go compiler, source code
- **Purpose**: Minimal production deployment
- **Size**: <500MB
- **Build**: `./docker/scripts/build-images.sh production`

### Builder Image
- **Location**: `packaging/docker/Dockerfile.builder`
- **Base**: Ubuntu 24.04
- **Includes**: Go 1.24.2, FUSE3, build tools, GUI dependencies
- **Purpose**: Foundation for development containers
- **Size**: ~1.49GB
- **Build**: `./docker/scripts/build-images.sh builder`

### Test Runner
- **Location**: `docker/images/test-runner/Dockerfile`
- **Base**: onemount-builder
- **Includes**: Python 3.12, pytest, debugging tools (vim, less)
- **Purpose**: Running unit, integration, and system tests
- **Build**: `./docker/scripts/build-images.sh test-runner`

### GitHub Runner
- **Location**: `docker/images/github-runner/Dockerfile`
- **Base**: onemount-builder
- **Includes**: GitHub Actions runner, Docker CLI, debugging tools
- **Purpose**: Self-hosted CI/CD runners
- **Build**: `./docker/scripts/build-images.sh github-runner`

### Debian Builder
- **Location**: `docker/images/deb-builder/Dockerfile`
- **Base**: onemount-builder
- **Includes**: Debian packaging tools
- **Purpose**: Building .deb packages
- **Build**: `./docker/scripts/build-images.sh deb-builder`

## Docker Compose Files

## Building Docker Images

### Build All Images

```bash
# Build all images (base + all derived images)
./docker/scripts/build-images.sh all

# Build only development images (excludes production runtime)
./docker/scripts/build-images.sh dev

# Build specific image
./docker/scripts/build-images.sh test-runner
./docker/scripts/build-images.sh github-runner
./docker/scripts/build-images.sh deb-builder

# Build without cache
./docker/scripts/build-images.sh all --no-cache
```

### Building Binaries (`docker-compose.build.yml`)

Build OneMount binaries and Debian packages inside containers.

```bash
# Build binaries
docker compose -f docker/compose/docker-compose.build.yml --profile binaries run --rm build-binaries

# Build Debian package
docker compose -f docker/compose/docker-compose.build.yml --profile deb run --rm build-deb

# Clean build artifacts
docker compose -f docker/compose/docker-compose.build.yml --profile clean run --rm clean-build
```

**Services**: build-binaries, build-deb, clean-build  
**Profiles**: binaries, deb, package, all, clean

### Testing (`docker-compose.test.yml`)

Run tests in isolated Docker environments with FUSE support.

```bash
# Unit tests (fast, no FUSE required)
docker compose -f docker/compose/docker-compose.test.yml run --rm unit-tests

# Integration tests (requires FUSE)
docker compose -f docker/compose/docker-compose.test.yml run --rm integration-tests

# System tests (requires auth tokens)
docker compose -f docker/compose/docker-compose.test.yml run --rm system-tests

# All tests with coverage
docker compose -f docker/compose/docker-compose.test.yml run --rm test-runner all

# Interactive debugging shell
docker compose -f docker/compose/docker-compose.test.yml run --rm shell
```

**Services**: test-runner, unit-tests, integration-tests, system-tests, coverage, shell  
**Requirements**: FUSE device access, test auth tokens for system tests

### Runners (`docker-compose.runners.yml`)

GitHub Actions self-hosted runners for CI/CD.

```bash
# Development (single runner, interactive)
docker compose -f docker/compose/docker-compose.runners.yml --profile dev up -d

# Production (two runners, auto-restart)
docker compose -f docker/compose/docker-compose.runners.yml --profile prod up -d

# Remote deployment
DOCKER_HOST=tcp://remote-host:2376 docker compose -f docker/compose/docker-compose.runners.yml --profile prod up -d

# Check status
docker ps --filter "name=onemount-runner"

# View logs
docker logs onemount-runner-1
docker logs onemount-runner-2
```

**Services**: runner-dev, runner-1, runner-2  
**Profiles**: dev/development (single runner), prod/production (two runners)

## Environment Variables

### Required for Runners
- `GITHUB_TOKEN` - GitHub personal access token with repo scope
- `GITHUB_REPOSITORY` - Repository in format 'owner/repo'

### Optional
- `ONEMOUNT_VERSION` - Version tag for images (default: 0.1.0rc1)
- `RUNNER_NAME` - Custom runner name (default: auto-generated)
- `RUNNER_LABELS` - Comma-separated runner labels
- `RUNNER_GROUP` - Runner group (default: Default)
- `AUTH_TOKENS_B64` - Base64-encoded OneDrive auth tokens
- `USER_ID` / `GROUP_ID` - Container user mapping (default: 1000)

### Token Management
- `ONEMOUNT_AUTO_REFRESH_TOKENS` - Enable automatic token refresh (default: true)
- `ONEMOUNT_TOKEN_REFRESH_INTERVAL` - Refresh interval in seconds (default: 3600)
- `ONEMOUNT_SYNC_WORKSPACE` - Sync workspace on startup (default: false)

## Volumes and Storage

### Test Volumes
- `test-artifacts/` - Test output, logs, and coverage reports (host mount)
- `test-artifacts/.auth_tokens.json` - Test OneDrive credentials (gitignored)

### Runner Volumes
- `runner-X-work` - GitHub Actions runner work directory (Docker volume)
- `runner-X-workspace` - Project workspace (Docker volume)
- `runner-X-tokens` - OneDrive auth tokens (Docker volume)

### Workspace Management
Runners use Docker volumes for better performance:
- Source code copied into volume during startup
- Manual sync: `docker exec <container> runner-entrypoint.sh sync-workspace`
- Auto-sync: Set `ONEMOUNT_SYNC_WORKSPACE=true`

## Security

### Container Security
- Non-root execution where possible
- FUSE capabilities for filesystem testing
- AppArmor unconfined for FUSE operations
- Isolated container networking

### Credential Management
- **Never** mount production auth tokens directly
- Use `AUTH_TOKENS_B64` environment variable for runners
- Use dedicated test OneDrive accounts for testing
- Store test tokens in `test-artifacts/.auth_tokens.json` (gitignored)

### Auth Token Setup

For testing:
```bash
# 1. Create dedicated test OneDrive account
# 2. Authenticate with test account
./build/onemount --auth-only

# 3. Copy to test location
mkdir -p test-artifacts
cp ~/.cache/onemount/auth_tokens.json test-artifacts/.auth_tokens.json
```

For runners:
```bash
# Use environment variable (recommended)
export AUTH_TOKENS_B64=$(base64 -w 0 test-artifacts/.auth_tokens.json)
```

## Networking

All containers use IPv4-only networking:
- DNS: 8.8.8.8, 8.8.4.4
- No IPv6 dependencies
- Bridge network mode

## Common Commands

### Container Management
```bash
# List OneMount containers
docker ps --filter "name=onemount"

# Stop all OneMount containers
docker stop $(docker ps -q --filter "name=onemount")

# Remove all OneMount containers
docker rm $(docker ps -aq --filter "name=onemount")

# View logs
docker logs onemount-runner-1
docker logs -f onemount-unit-tests  # Follow logs
```

### Image Management
```bash
# List OneMount images
docker images | grep onemount

# Build images (automatic when running compose)
docker compose -f docker/compose/docker-compose.test.yml run --rm unit-tests

# Remove images
docker rmi onemount-test-runner:latest

# Check image sizes
docker images --format "table {{.Repository}}\t{{.Tag}}\t{{.Size}}" | grep onemount
```

### Volume Management
```bash
# List volumes
docker volume ls | grep onemount

# Inspect volume
docker volume inspect onemount-runner-1-work

# Remove unused volumes
docker volume prune

# Backup volume
docker run --rm -v onemount-runner-1-work:/data -v $(pwd):/backup ubuntu tar czf /backup/runner-1-work.tar.gz /data
```

### Debugging
```bash
# Interactive shell in test container
docker compose -f docker/compose/docker-compose.test.yml run --rm shell

# Run specific test
docker compose -f docker/compose/docker-compose.test.yml run --rm \
  --entrypoint /bin/bash test-runner -c "go test -v -run TestSpecific ./internal/..."

# Execute command in running container
docker exec -it onemount-runner-1 bash

# Check resource usage
docker stats onemount-runner-1
```

## Troubleshooting

### Permission Errors
```bash
# Check user mapping
docker compose -f docker/compose/docker-compose.test.yml run --rm shell
id  # Should show correct user

# Fix volume permissions
sudo chown -R $(id -u):$(id -g) test-artifacts/
```

### FUSE Errors
```bash
# Check FUSE device
ls -l /dev/fuse

# Verify in container
docker compose -f docker/compose/docker-compose.test.yml run --rm shell
ls -l /dev/fuse
```

### Auth Token Errors
```bash
# Verify test tokens exist
ls -la test-artifacts/.auth_tokens.json

# Check token format
jq . test-artifacts/.auth_tokens.json

# Refresh tokens manually
docker exec onemount-runner-1 runner-entrypoint.sh refresh-tokens
```

### Container Conflicts
```bash
# Remove conflicting container
docker rm -f onemount-test-runner

# Force recreate
docker compose -f docker/compose/docker-compose.test.yml up --force-recreate
```

### Build Failures
```bash
# Enable BuildKit
export DOCKER_BUILDKIT=1

# Build with no cache
docker compose -f docker/compose/docker-compose.build.yml build --no-cache

# Check Docker daemon
docker info
```

## Build Optimization

### BuildKit Cache
All Dockerfiles use BuildKit cache mounts for faster builds:
```dockerfile
RUN --mount=type=cache,target=/tmp/go-mod-cache,uid=1000 \
    --mount=type=cache,target=/tmp/go-cache,uid=1000 \
    go mod download
```

**Benefits**: 50-70% faster rebuilds

### .dockerignore
Minimize build context with `.dockerignore`:
```dockerignore
.git/
build/
dist/
.venv/
venv/
test-artifacts/
*_test.go
```

## Integration

These Docker configurations integrate with:
- GitHub Actions workflows (`.github/workflows/system-tests-self-hosted.yml`)
- OneMount build system (`Makefile`, `scripts/`)
- OneDrive authentication and testing
- Debian package building and distribution

## Related Documentation

- **Production Deployment**: `packaging/docker/README.md`
- **Test Setup**: `docs/TEST_SETUP.md`
- **Docker Test Environment**: `docs/testing/docker-test-environment.md`
- **GitHub Runners**: `docs/github-runners.md`
- **Remote Deployment**: `docs/docker-remote-api-setup.md`

## Quick Reference

### Building Images

| Task | Command |
|------|---------|
| Build all images | `./docker/scripts/build-images.sh all` |
| Build dev images | `./docker/scripts/build-images.sh dev` |
| Build production image | `./docker/scripts/build-images.sh production` |
| Build builder image | `./docker/scripts/build-images.sh builder` |
| Build test runner | `./docker/scripts/build-images.sh test-runner` |
| Build without cache | `./docker/scripts/build-images.sh all --no-cache` |

### Running Tests

| Task | Command |
|------|---------|
| Run unit tests | `docker compose -f docker/compose/docker-compose.test.yml run --rm unit-tests` |
| Run integration tests | `docker compose -f docker/compose/docker-compose.test.yml run --rm integration-tests` |
| Run system tests | `docker compose -f docker/compose/docker-compose.test.yml run --rm system-tests` |
| Run all tests | `docker compose -f docker/compose/docker-compose.test.yml run --rm test-runner all` |
| Interactive shell | `docker compose -f docker/compose/docker-compose.test.yml run --rm shell` |

### Building Binaries

| Task | Command |
|------|---------|
| Build binaries | `docker compose -f docker/compose/docker-compose.build.yml --profile binaries run --rm build-binaries` |
| Build .deb package | `docker compose -f docker/compose/docker-compose.build.yml --profile deb run --rm build-deb` |
| Clean artifacts | `docker compose -f docker/compose/docker-compose.build.yml --profile clean run --rm clean-build` |

### GitHub Runners

| Task | Command |
|------|---------|
| Start dev runner | `docker compose -f docker/compose/docker-compose.runners.yml --profile dev up -d` |
| Start prod runners | `docker compose -f docker/compose/docker-compose.runners.yml --profile prod up -d` |
| View logs | `docker logs onemount-runner-1` |
| Stop runners | `docker compose -f docker/compose/docker-compose.runners.yml down` |

### Maintenance

| Task | Command |
|------|---------|
| List images | `docker images | grep onemount` |
| Clean up containers | `docker rm -f $(docker ps -aq --filter "name=onemount")` |
| Clean up volumes | `docker volume prune` |
