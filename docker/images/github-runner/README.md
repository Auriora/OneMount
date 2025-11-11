# GitHub Runner Image

Docker image for GitHub Actions self-hosted runners.

**Base Image**: `onemount-builder` (from `packaging/docker/Dockerfile.builder`)

## Building

```bash
# Build GitHub runner image
./docker/scripts/build-images.sh github-runner

# Build with no cache
./docker/scripts/build-images.sh github-runner --no-cache
```

## Usage

### Development Runner (Single, Interactive)

```bash
docker compose -f docker/compose/docker-compose.runners.yml --profile dev up -d
```

### Production Runners (Two, Auto-restart)

```bash
docker compose -f docker/compose/docker-compose.runners.yml --profile prod up -d
```

### Required Environment Variables

```bash
export GITHUB_TOKEN=ghp_xxxxxxxxxxxxx
export GITHUB_REPOSITORY=owner/repo
```

### Optional Environment Variables

```bash
export RUNNER_NAME=custom-runner-name
export RUNNER_LABELS=self-hosted,linux,custom
export AUTH_TOKENS_B64=$(base64 -w 0 test-artifacts/.auth_tokens.json)
```

## Features

- GitHub Actions runner v2.311.0
- Go 1.24.2
- Docker CLI for elastic manager
- FUSE3 support
- Automatic token refresh
- Workspace synchronization

## Quick Reference

```bash
# Build image
./docker/scripts/build-images.sh github-runner

# Start development runner
docker compose -f docker/compose/docker-compose.runners.yml --profile dev up -d

# Start production runners
docker compose -f docker/compose/docker-compose.runners.yml --profile prod up -d

# View logs
docker logs onemount-runner-1

# Stop runners
docker compose -f docker/compose/docker-compose.runners.yml down
```

## See Also

- Runner compose file: `docker/compose/docker-compose.runners.yml`
- Runner entrypoint: `docker/scripts/runner-entrypoint.sh`
- Runner documentation: `docs/github-runners.md`
- Build script: `docker/scripts/build-images.sh`
