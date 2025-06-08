# 🏃 GitHub Actions Runners

This guide shows how to set up and manage GitHub Actions runners for OneMount using a simple 2-runner configuration.

## 🎯 Overview

The runner system provides:
- **2 Static Runners**: Manual start/stop control based on resource needs
- **Persistent Storage**: Runner data persists across restarts
- **Remote Docker**: Runs on dedicated Docker host (172.16.1.104:2376)
- **Simple Management**: Easy-to-use scripts for manual control

## ⚡ Quick Start

### 1. Prerequisites

Ensure you have:
- ✅ Docker access to remote host (172.16.1.104:2376)
- ✅ GitHub token with appropriate permissions
- ✅ Environment file (`.env`) configured

### 2. Environment Setup

Create or update your `.env` file:

```bash
# GitHub Configuration
GITHUB_TOKEN=your_github_token_here
GITHUB_REPOSITORY=Auriora/OneMount

# Optional: Runner configuration
RUNNER_GROUP=Default
AUTH_TOKENS_B64=

# Docker Configuration
DOCKER_HOST=tcp://172.16.1.104:2376
```

### 3. Build and Start

```bash
# Build the runner image
./scripts/manage-runners.sh build

# Start both runners
./scripts/manage-runners.sh start

# Check status
./scripts/manage-runners.sh status
```

## 🛠️ Management Commands

### Basic Operations

```bash
# Start runners
./scripts/manage-runners.sh start          # Start both runners
./scripts/manage-runners.sh start 1        # Start only runner-1 (primary)
./scripts/manage-runners.sh start 2        # Start only runner-2 (secondary)

# Stop runners
./scripts/manage-runners.sh stop           # Stop both runners
./scripts/manage-runners.sh stop 1         # Stop only runner-1
./scripts/manage-runners.sh stop 2         # Stop only runner-2

# Restart runners
./scripts/manage-runners.sh restart        # Restart both runners
./scripts/manage-runners.sh restart 1      # Restart only runner-1
./scripts/manage-runners.sh restart 2      # Restart only runner-2

# Check status
./scripts/manage-runners.sh status         # Show runner status
```

### Maintenance

```bash
# Rebuild runner image
./scripts/manage-runners.sh build

# Clean up stopped containers and volumes
./scripts/manage-runners.sh cleanup
```

## 🏗️ Runner Configuration

### Runner Details

- **Runner 1 (Primary)**: `onemount-runner-1`
  - Always keep running for immediate availability
  - Labels: `self-hosted,Linux,onemount-testing,optimized`
  - Persistent storage for workspace and credentials

- **Runner 2 (Secondary)**: `onemount-runner-2`
  - Start/stop based on workload and system resources
  - Same configuration as runner-1
  - Independent persistent storage

### Resource Management

**Recommended Usage Pattern:**
1. **Always keep runner-1 running** for immediate job pickup
2. **Start runner-2 when needed** for parallel jobs or heavy workloads
3. **Stop runner-2** during low activity or resource constraints
4. **Monitor system resources** and adjust accordingly

## 🔧 Advanced Configuration

### Custom Docker Host

```bash
# Use different Docker host
DOCKER_HOST=tcp://192.168.1.100:2376 ./scripts/manage-runners.sh start
```

### Environment Variables

The system supports these environment variables:

```bash
# Required
GITHUB_TOKEN              # GitHub API token
GITHUB_REPOSITORY          # Repository in format owner/repo

# Optional
RUNNER_GROUP               # Runner group (default: Default)
AUTH_TOKENS_B64           # Base64 encoded auth tokens
DOCKER_HOST               # Docker host (default: tcp://172.16.1.104:2376)
```

### Docker Compose Direct Usage

You can also use Docker Compose directly:

```bash
cd docker/compose

# Start specific runner
DOCKER_HOST=tcp://172.16.1.104:2376 docker compose -f docker-compose.runners.yml --env-file ../../.env up -d runner-1

# Stop specific runner
DOCKER_HOST=tcp://172.16.1.104:2376 docker compose -f docker-compose.runners.yml stop runner-2

# View logs
DOCKER_HOST=tcp://172.16.1.104:2376 docker compose -f docker-compose.runners.yml logs -f runner-1
```

## 📊 Monitoring

### Status Checking

The `status` command shows:
- Container status for both runners
- GitHub registration status (if `jq` and `GITHUB_TOKEN` available)
- Resource usage information

```bash
./scripts/manage-runners.sh status
```

### Log Monitoring

```bash
# View runner logs
DOCKER_HOST=tcp://172.16.1.104:2376 docker logs -f onemount-runner-1
DOCKER_HOST=tcp://172.16.1.104:2376 docker logs -f onemount-runner-2

# View all runner logs
cd docker/compose
DOCKER_HOST=tcp://172.16.1.104:2376 docker compose -f docker-compose.runners.yml logs -f
```

## 🚨 Troubleshooting

### Common Issues

**Runner not appearing in GitHub:**
1. Check GitHub token permissions
2. Verify repository name format (`owner/repo`)
3. Check runner logs for registration errors

**Docker connection issues:**
1. Verify Docker host accessibility: `DOCKER_HOST=tcp://172.16.1.104:2376 docker version`
2. Check firewall settings on Docker host
3. Ensure Docker daemon is running on remote host

**Permission issues:**
1. Runners run as root to avoid permission problems
2. Check volume mount permissions
3. Verify FUSE device access for filesystem tests

### Recovery Procedures

**Reset runner registration:**
```bash
# Stop runners
./scripts/manage-runners.sh stop

# Clean up containers and volumes
./scripts/manage-runners.sh cleanup

# Rebuild and restart
./scripts/manage-runners.sh build
./scripts/manage-runners.sh start
```

**Manual container management:**
```bash
# Remove stuck containers
DOCKER_HOST=tcp://172.16.1.104:2376 docker rm -f onemount-runner-1 onemount-runner-2

# Remove volumes (WARNING: loses runner data)
DOCKER_HOST=tcp://172.16.1.104:2376 docker volume rm onemount-runners_runner-1-data onemount-runners_runner-2-data
```

## 💡 Best Practices

1. **Keep runner-1 always running** for immediate job availability
2. **Monitor system resources** before starting runner-2
3. **Regular maintenance**: Rebuild images monthly for security updates
4. **Backup strategy**: Runner credentials persist in Docker volumes
5. **Resource planning**: Each runner can consume significant CPU/memory during builds
6. **Log rotation**: Monitor Docker logs to prevent disk space issues

## 🔗 Related Documentation

- [Docker Development Workflow](docker-development-workflow.md)
- [Remote Docker Setup](remote-docker-setup.md)
- [GitHub Actions Configuration](.github/workflows/)

For more information about GitHub Actions runners, see the [official documentation](https://docs.github.com/en/actions/hosting-your-own-runners).
