# Docker Configuration for OneMount

This directory contains Docker-related configuration files for OneMount development and deployment.

## Directory Structure

```
docker/
├── compose/                    # Docker Compose configurations
│   ├── docker-compose.runner.yml    # Local GitHub Actions runner
│   └── docker-compose.remote.yml    # Remote Docker host runner
└── README.md                   # This file
```

## Docker Compose Files

### `compose/docker-compose.runner.yml`
- **Purpose**: Local GitHub Actions self-hosted runner
- **Usage**: Used by `scripts/manage-runner.sh` for local development
- **Features**:
  - Interactive setup and management
  - Development and production modes
  - Persistent volumes for runner data
  - FUSE support for filesystem testing

### `compose/docker-compose.runners.yml`
- **Purpose**: Simple 2-runner setup for manual management
- **Usage**: Used by `scripts/manage-runners.sh` for remote deployment
- **Features**:
  - 2 static runners (primary and secondary)
  - Manual start/stop control
  - Persistent storage for runner data
  - FUSE support for filesystem testing

### `compose/docker-compose.remote.yml`
- **Purpose**: Remote Docker host deployment
- **Usage**: Used by `scripts/deploy-docker-remote.sh` for remote deployment
- **Features**:
  - Simplified configuration for remote deployment
  - IPv4-only networking for South African networks
  - Persistent storage for runner workspace and logs

## Related Files

### Dockerfiles
- `packaging/docker/Dockerfile.github-runner` - Main runner image
- `packaging/docker/Dockerfile.deb-builder` - Package building image
- `packaging/docker/Dockerfile.test-runner` - Test environment image

### Scripts
- `scripts/manage-runner.sh` - Local runner management
- `scripts/manage-runners.sh` - Simple 2-runner management
- `scripts/deploy-docker-remote.sh` - Remote deployment management
- `packaging/docker/runner-entrypoint.sh` - Runner container entrypoint

### Documentation
- `docs/docker-self-hosted-runner.md` - Local runner setup guide
- `docs/github-runners.md` - Simple 2-runner setup guide
- `docs/docker-remote-api-setup.md` - Remote deployment guide

## Quick Start

### Local Runner
```bash
# Setup and start local runner
./scripts/manage-runner.sh setup
./scripts/manage-runner.sh build
./scripts/manage-runner.sh start
```

### Remote Deployment
```bash
# Setup and deploy to remote Docker host
./scripts/deploy-docker-remote.sh setup
./scripts/deploy-docker-remote.sh build
./scripts/deploy-docker-remote.sh deploy
```

## Environment Variables

Both compose files support the following environment variables:

### Required
- `GITHUB_TOKEN` - GitHub personal access token with repo scope
- `GITHUB_REPOSITORY` - Repository in format 'owner/repo'

### Optional
- `RUNNER_NAME` - Custom runner name (default: auto-generated)
- `RUNNER_LABELS` - Comma-separated runner labels
- `AUTH_TOKENS_B64` - Base64-encoded OneDrive authentication tokens

## Volumes

### Persistent Volumes
- `runner-1-workspace` / `runner-2-workspace` - Project workspace (Docker volumes)
- `runner-1-work` / `runner-2-work` - GitHub Actions runner work directory
- `onemount-runner-workspace` - Remote runner workspace (for remote deployments)
- `onemount-runner-work` - Remote runner work directory
- `onemount-runner-logs` - Test logs and artifacts

### Host Mounts
- `.env` file for environment configuration
- Optional: Host auth tokens for OneDrive authentication

### Workspace Management
The runners now use Docker volumes instead of bind mounts for better performance and consistency:
- Source code is copied into the volume during container startup
- Use `ONEMOUNT_SYNC_WORKSPACE=true` to sync workspace on startup
- Manual sync available via `docker exec <container> runner-entrypoint.sh sync-workspace`

### Token Management
Authentication tokens are now managed automatically with refresh capabilities:
- Tokens are stored in persistent Docker volumes (`runner-X-tokens`)
- Automatic token refresh using OneMount's built-in capabilities
- Fallback to environment variables if refresh fails
- Periodic refresh daemon (configurable interval)
- Manual token management via `docker exec <container> runner-entrypoint.sh refresh-tokens`

#### Token Environment Variables
- `ONEMOUNT_AUTO_REFRESH_TOKENS=false` - Disable automatic token refresh (default: true)
- `ONEMOUNT_TOKEN_REFRESH_INTERVAL=3600` - Refresh interval in seconds (default: 1 hour)

## Networking

Both configurations use IPv4-only networking suitable for South African network environments:
- DNS: 8.8.8.8, 8.8.4.4
- No IPv6 dependencies
- Direct TCP connections for remote Docker API

## Security

### Container Security
- Non-root execution with `runner` user
- FUSE capabilities for filesystem testing
- AppArmor unconfined for FUSE operations
- Isolated container networking

### Credential Management
- Environment-based configuration
- Base64-encoded sensitive data
- No hardcoded credentials in compose files

## Troubleshooting

### Common Issues
1. **Permission errors**: Ensure Docker user has proper permissions
2. **Network connectivity**: Verify IPv4-only DNS configuration
3. **FUSE errors**: Check device mounting and capabilities
4. **Authentication failures**: Verify GitHub token and OneDrive credentials

### Debugging
```bash
# Check container status
docker-compose -f docker/compose/docker-compose.runner.yml ps

# View logs
docker-compose -f docker/compose/docker-compose.runner.yml logs -f

# Interactive shell
docker-compose -f docker/compose/docker-compose.runner.yml exec github-runner /bin/bash
```

## Integration

These Docker configurations integrate with:
- GitHub Actions workflows (`system-tests-self-hosted.yml`)
- OneMount build and test systems
- OneDrive authentication and testing
- Package building and distribution

For detailed setup instructions, see the documentation files in the `docs/` directory.
