# Remote Docker Host Setup for OneMount GitHub Actions Runner

This guide shows how to deploy the OneMount GitHub Actions runner to a remote Docker host at 172.16.1.104 using direct Docker API access.

## 🚀 Quick Setup

### Prerequisites

1. **Docker API Access**: Remote Docker daemon must be accessible on port 2376
2. **Docker**: Docker must be installed and configured on the remote host
3. **GitHub Token**: Personal access token with `repo` scope
4. **Network Access**: Port 2376 must be accessible from your local machine

### 1. Interactive Setup

```bash
# Run the interactive setup
./scripts/deploy-docker-remote.sh setup
```

This will:
- Test Docker API connectivity
- Set up GitHub authentication
- Configure OneDrive authentication (optional)
- Save configuration for future use

### 2. Deploy and Start

```bash
# Build the runner image on remote host
./scripts/deploy-docker-remote.sh build

# Deploy and start the runner
./scripts/deploy-docker-remote.sh deploy

# Check status
./scripts/deploy-docker-remote.sh status

# View logs
./scripts/deploy-docker-remote.sh logs
```

## 🔧 Manual Configuration

### Remote Host Requirements

The remote host (172.16.1.104) should have:

```bash
# Install Docker (if not already installed)
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# Configure Docker daemon to accept API connections
# Edit /etc/docker/daemon.json
sudo tee /etc/docker/daemon.json << EOF
{
  "hosts": ["unix:///var/run/docker.sock", "tcp://0.0.0.0:2376"]
}
EOF

# Restart Docker daemon
sudo systemctl restart docker
```

### Docker API Access Test

Test Docker API connectivity from your local machine:

```bash
# Test connection to remote Docker API
DOCKER_HOST="tcp://172.16.1.104:2376" docker version

# Should show both client and server information
```

## 📋 Available Commands

### Deployment Commands

```bash
# Interactive setup
./scripts/deploy-docker-remote.sh setup

# Build runner image on remote host
./scripts/deploy-docker-remote.sh build

# Deploy and start runner
./scripts/deploy-docker-remote.sh deploy

# Deploy with custom Docker host
./scripts/deploy-docker-remote.sh deploy --host 172.16.1.104:2376
```

### Management Commands

```bash
# Start the runner
./scripts/deploy-docker-remote.sh start

# Stop the runner
./scripts/deploy-docker-remote.sh stop

# Restart the runner
./scripts/deploy-docker-remote.sh restart

# Check status
./scripts/deploy-docker-remote.sh status

# View logs (follows)
./scripts/deploy-docker-remote.sh logs

# Connect to runner shell
./scripts/deploy-docker-remote.sh shell
```

### Maintenance Commands

```bash
# Update runner image and restart
./scripts/deploy-docker-remote.sh update

# Clean up everything
./scripts/deploy-docker-remote.sh clean
```

## 🔐 Authentication Setup

### GitHub Personal Access Token

1. Go to [GitHub Settings → Personal Access Tokens](https://github.com/settings/tokens)
2. Click "Generate new token (classic)"
3. Select `repo` scope
4. Copy the token for use in setup

### OneDrive Authentication (Optional)

If you have existing OneMount authentication:

```bash
# Check for existing tokens
ls -la ~/.cache/onemount/auth_tokens.json

# The setup script will automatically detect and use these
```

Or provide a custom path during setup.

## 🏗️ Remote Container Structure

The runner operates as a Docker container on the remote host with:

```
Container: onemount-github-runner
├── Persistent Volumes:
│   ├── onemount-runner-workspace:/workspace     # Runner workspace
│   └── onemount-runner-work:/opt/actions-runner/_work  # GitHub Actions work directory
├── Environment Variables:
│   ├── GITHUB_TOKEN                             # GitHub authentication
│   ├── GITHUB_REPOSITORY                       # Target repository
│   ├── RUNNER_NAME                             # Runner identifier
│   └── AUTH_TOKENS_B64                         # OneDrive authentication (optional)
└── Labels: self-hosted,linux,onemount-testing,docker-remote
```

## 🔍 Troubleshooting

### Docker API Connection Issues

```bash
# Test Docker API connectivity
DOCKER_HOST="tcp://172.16.1.104:2376" docker version

# Test port connectivity
telnet 172.16.1.104 2376

# Check if Docker daemon is configured for API access
# On remote host: sudo systemctl status docker
```

### Docker Issues on Remote Host

```bash
# Test Docker API directly
DOCKER_HOST="tcp://172.16.1.104:2376" docker ps

# Check Docker daemon configuration
DOCKER_HOST="tcp://172.16.1.104:2376" docker info

# Check available resources
DOCKER_HOST="tcp://172.16.1.104:2376" docker system df
```

### Runner Registration Issues

```bash
# Check runner logs
./scripts/deploy-docker-remote.sh logs

# Connect to runner shell for debugging
./scripts/deploy-docker-remote.sh shell

# Check GitHub repository settings
# Go to Repository → Settings → Actions → Runners
```

### Network Issues

The runner is configured for IPv4-only networking:

```bash
# Test network connectivity from container
./scripts/deploy-docker-remote.sh shell
# Inside container: ping 8.8.8.8
```

## 📊 Monitoring

### Check Runner Status

```bash
# Quick status check
./scripts/deploy-docker-remote.sh status

# Detailed logs
./scripts/deploy-docker-remote.sh logs

# GitHub repository runners page
# Repository → Settings → Actions → Runners
```

### Container Resources

```bash
# Check container resource usage
DOCKER_HOST="tcp://172.16.1.104:2376" docker stats

# Check disk usage on remote host
DOCKER_HOST="tcp://172.16.1.104:2376" docker system df
```

## 🔄 Updates and Maintenance

### Updating the Runner

```bash
# Update runner image and restart
./scripts/deploy-docker-remote.sh update
```

This will:
1. Build a new runner image on the remote host
2. Stop the current runner container
3. Deploy the updated container
4. Start the updated runner

### Configuration Management

```bash
# Configuration is stored in .docker-remote-config
# View current configuration
cat .docker-remote-config

# Reconfigure if needed
./scripts/deploy-docker-remote.sh setup
```

## 🎯 Integration with GitHub Actions

Once deployed and started, the runner will:

1. **Register** with your GitHub repository
2. **Appear** in Repository → Settings → Actions → Runners
3. **Accept jobs** from the `system-tests-self-hosted.yml` workflow
4. **Run tests** with pre-configured OneDrive authentication
5. **Upload artifacts** and logs automatically

The runner will have labels: `self-hosted`, `linux`, `onemount-testing`, `docker-remote`

## 🧹 Cleanup

To completely remove the runner:

```bash
# Clean up everything
./scripts/deploy-docker-remote.sh clean

# This will:
# - Stop and remove containers
# - Remove Docker images
# - Optionally remove persistent volumes
```

## 📞 Support

If you encounter issues:

1. Check the troubleshooting section above
2. View runner logs: `./scripts/deploy-docker-remote.sh logs`
3. Connect to runner shell: `./scripts/deploy-docker-remote.sh shell`
4. Check GitHub repository runner status
