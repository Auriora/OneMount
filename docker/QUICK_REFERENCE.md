# Docker Quick Reference

Quick reference for common Docker operations in the OneMount project.

## File Locations

| Purpose | Location |
|---------|----------|
| Development Dockerfiles | `docker/Dockerfile.*` |
| Base Dockerfile | `packaging/docker/Dockerfile.base` |
| Debian Builder Dockerfile | `packaging/deb/docker/Dockerfile` |
| Entrypoint Scripts | `docker/scripts/` |
| Development Compose | `docker/compose/` |
| Production Compose | `packaging/docker/docker-compose.yml` |

## Common Commands

### Building Docker Images

Images are built automatically when running compose files if not present.

```bash
# Enable BuildKit for faster builds
export DOCKER_BUILDKIT=1

# Images build automatically when running:
docker compose -f docker/compose/docker-compose.test.yml run --rm unit-tests
docker compose -f docker/compose/docker-compose.build.yml --profile deb run --rm build-deb

# Or build manually if needed:
docker build -f packaging/docker/Dockerfile.base -t onemount-base:latest .
docker build -f docker/Dockerfile.test-runner -t onemount-test-runner:latest .
docker build -f docker/Dockerfile.github-runner -t onemount-github-runner:latest .
docker build -f packaging/deb/docker/Dockerfile -t onemount-deb-builder:latest .
```

### Building OneMount Binaries/Packages

```bash
# Build binaries
docker compose -f docker/compose/docker-compose.build.yml --profile binaries run --rm build-binaries

# Build Debian package
docker compose -f docker/compose/docker-compose.build.yml --profile deb run --rm build-deb

# Clean build artifacts
docker compose -f docker/compose/docker-compose.build.yml --profile clean run --rm clean-build

# Using build entrypoint directly
docker run --rm -v $(pwd):/build -w /build onemount-deb-builder binaries
docker run --rm -v $(pwd):/build -w /build onemount-deb-builder deb --output /dist
```

### Running Tests

```bash
# Unit tests
docker compose -f docker/compose/docker-compose.test.yml run --rm unit-tests

# Integration tests
docker compose -f docker/compose/docker-compose.test.yml run --rm integration-tests

# System tests
docker compose -f docker/compose/docker-compose.test.yml run --rm system-tests

# All tests
docker compose -f docker/compose/docker-compose.test.yml run --rm test-runner all

# Coverage
docker compose -f docker/compose/docker-compose.test.yml run --rm coverage

# Interactive shell
docker compose -f docker/compose/docker-compose.test.yml run --rm shell
```

### Building Packages

```bash
# Build Debian package
docker compose -f packaging/docker/docker-compose.yml run --rm deb-builder

# With specific version
ONEMOUNT_VERSION=1.0.0 docker compose -f packaging/docker/docker-compose.yml run --rm deb-builder

# Check output
ls -la dist/*.deb
```

### Managing Runners

```bash
# Development runner (single)
docker compose -f docker/compose/docker-compose.runners.yml --profile dev up -d
docker compose -f docker/compose/docker-compose.runners.yml --profile dev down

# Production runners (multiple)
docker compose -f docker/compose/docker-compose.runners.yml --profile prod up -d
docker compose -f docker/compose/docker-compose.runners.yml --profile prod down

# Remote deployment (use DOCKER_HOST)
DOCKER_HOST=tcp://remote-host:2376 docker compose -f docker/compose/docker-compose.runners.yml --profile prod up -d

# Check status
docker ps --filter "name=onemount-runner"

# View logs
docker logs onemount-runner-1
docker logs onemount-runner-2
docker logs onemount-runner-dev
```

### Production Deployment

```bash
# Start production container
docker compose -f packaging/docker/docker-compose.yml up -d

# Stop production container
docker compose -f packaging/docker/docker-compose.yml down

# View logs
docker logs onemount

# Check health
docker inspect onemount | jq '.[0].State.Health'
```

### Container Management

```bash
# List all OneMount containers
docker ps --filter "name=onemount"

# List all (including stopped)
docker ps -a --filter "name=onemount"

# Stop all OneMount containers
docker stop $(docker ps -q --filter "name=onemount")

# Remove all OneMount containers
docker rm $(docker ps -aq --filter "name=onemount")

# View logs
docker logs onemount-unit-tests
docker logs -f onemount-runner-1  # Follow logs

# Execute command in running container
docker exec -it onemount-runner-1 bash

# Inspect container
docker inspect onemount-production
```

### Image Management

```bash
# List OneMount images
docker images | grep onemount

# Remove specific image
docker rmi onemount-test-runner:latest

# Remove all OneMount images
docker rmi $(docker images -q "onemount-*")

# Prune unused images
docker image prune -a

# Check image size
docker images --format "table {{.Repository}}\t{{.Tag}}\t{{.Size}}" | grep onemount
```

### Volume Management

```bash
# List OneMount volumes
docker volume ls | grep onemount

# Inspect volume
docker volume inspect onemount-runner-1-work

# Remove specific volume
docker volume rm onemount-runner-1-work

# Remove all unused volumes
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

# Check container logs
docker logs onemount-unit-tests

# Follow logs
docker logs -f onemount-runner-1

# Inspect container
docker inspect onemount-production | jq '.[0].State'

# Check resource usage
docker stats onemount-production

# Execute command in running container
docker exec -it onemount-runner-1 bash
```

### Cleanup

```bash
# Stop and remove all OneMount containers
docker compose -f docker/compose/docker-compose.test.yml down
docker compose -f docker/compose/docker-compose.runners.yml down
docker compose -f packaging/docker/docker-compose.yml --profile production down

# Remove all OneMount containers
docker rm -f $(docker ps -aq --filter "name=onemount")

# Remove all OneMount images
docker rmi $(docker images -q "onemount-*")

# Remove all OneMount volumes
docker volume rm $(docker volume ls -q | grep onemount)

# Complete cleanup (use with caution!)
docker system prune -a --volumes
```

## Container Names Reference

| Container Name | Purpose | Compose File |
|----------------|---------|--------------|
| `onemount-test-runner` | Main test runner | docker-compose.test.yml |
| `onemount-unit-tests` | Unit test execution | docker-compose.test.yml |
| `onemount-integration-tests` | Integration tests | docker-compose.test.yml |
| `onemount-system-tests` | System tests | docker-compose.test.yml |
| `onemount-coverage` | Coverage analysis | docker-compose.test.yml |
| `onemount-shell` | Interactive debugging | docker-compose.test.yml |
| `onemount-runner-1` | Production runner #1 | docker-compose.runners.yml |
| `onemount-runner-2` | Production runner #2 | docker-compose.runners.yml |
| `onemount-github-runner` | Development runner | docker-compose.runner.yml |
| `onemount-base-build` | Base image builder | docker-compose.build.yml |
| `onemount-deb-builder` | Debian package builder | packaging/docker-compose.yml |
| `onemount-production` | Production deployment | packaging/docker-compose.yml |

## Project Names Reference

| Project Name | Compose File | Purpose |
|--------------|--------------|---------|
| `onemount-build` | docker-compose.build.yml | Image building |
| `onemount-test` | docker-compose.test.yml | Testing workflows |
| `onemount-runner` | docker-compose.runner.yml | Single runner (dev) |
| `onemount-runners` | docker-compose.runners.yml | Multi-runner (prod) |
| `onemount-remote` | docker-compose.remote.yml | Remote deployment |
| `onemount-packaging` | packaging/docker-compose.yml | Production packaging |

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `ONEMOUNT_VERSION` | `0.1.0rc1` | Version tag for images |
| `USER_ID` | `1000` | User ID for container |
| `GROUP_ID` | `1000` | Group ID for container |
| `GITHUB_TOKEN` | - | GitHub PAT for runners |
| `GITHUB_REPOSITORY` | - | Repository for runners |
| `RUNNER_NAME` | auto | Custom runner name |
| `RUNNER_LABELS` | auto | Runner labels |
| `AUTH_TOKENS_B64` | - | Base64 OneDrive tokens |

## Useful Aliases

Add these to your `.bashrc` or `.zshrc`:

```bash
# Docker Compose shortcuts
alias dc='docker compose'
alias dcb='docker compose -f docker/compose/docker-compose.build.yml'
alias dct='docker compose -f docker/compose/docker-compose.test.yml'
alias dcp='docker compose -f packaging/docker/docker-compose.yml'

# OneMount specific
alias om-test='docker compose -f docker/compose/docker-compose.test.yml run --rm'
alias om-build='docker compose -f docker/compose/docker-compose.build.yml build'
alias om-shell='docker compose -f docker/compose/docker-compose.test.yml run --rm shell'
alias om-clean='docker rm -f $(docker ps -aq --filter "name=onemount")'
alias om-ps='docker ps --filter "name=onemount"'
alias om-logs='docker logs'

# Usage examples:
# om-test unit-tests
# om-build test-runner-build
# om-shell
# om-ps
```

## Documentation

- **Organization**: `docker/DOCKER_ORGANIZATION.md`
- **Migration**: `docker/MIGRATION_GUIDE.md`
- **Summary**: `docker/REFACTORING_SUMMARY.md`
- **Development**: `docker/README.md`
- **Production**: `packaging/docker/README.md`
- **Compose**: `docker/compose/README.md`
- **Change Log**: `docs/updates/2025-11-11-docker-refactoring.md`

## Getting Help

```bash
# View compose file services
docker compose -f docker/compose/docker-compose.test.yml config --services

# View compose file
docker compose -f docker/compose/docker-compose.test.yml config

# Check Docker version
docker --version
docker compose version

# Check system info
docker info
```
