# Remote Docker Host Setup for OneMount GitHub Actions Runner

This guide shows how to deploy the OneMount GitHub Actions runner to a remote Docker host at 172.16.1.104.

## 🚀 Quick Setup

### Prerequisites

1. **SSH Access**: Ensure you have SSH key-based access to the remote host
2. **Docker**: Docker and Docker Compose must be installed on the remote host
3. **GitHub Token**: Personal access token with `repo` scope
4. **Permissions**: User should have Docker permissions (in `docker` group)

### 1. Interactive Setup

```bash
# Run the interactive setup
./scripts/deploy-remote-runner.sh setup
```

This will:
- Configure remote host connection details
- Set up GitHub authentication
- Configure OneDrive authentication (optional)
- Test connectivity and Docker availability

### 2. Deploy and Start

```bash
# Deploy the runner to remote host
./scripts/deploy-remote-runner.sh deploy

# Start the runner
./scripts/deploy-remote-runner.sh start

# Check status
./scripts/deploy-remote-runner.sh status

# View logs
./scripts/deploy-remote-runner.sh logs
```

## 🔧 Manual Configuration

### Remote Host Requirements

The remote host (172.16.1.104) should have:

```bash
# Install Docker (if not already installed)
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# Install Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# Add user to docker group
sudo usermod -aG docker $USER
# Log out and back in for group changes to take effect
```

### SSH Key Setup

Ensure SSH key authentication is configured:

```bash
# Copy your SSH key to the remote host
ssh-copy-id username@172.16.1.104

# Test SSH connection
ssh username@172.16.1.104 "echo 'SSH working'"
```

## 📋 Available Commands

### Deployment Commands

```bash
# Interactive setup
./scripts/deploy-remote-runner.sh setup

# Deploy runner to remote host
./scripts/deploy-remote-runner.sh deploy

# Deploy with custom settings
./scripts/deploy-remote-runner.sh deploy --host 172.16.1.104 --user myuser
```

### Management Commands

```bash
# Start the runner
./scripts/deploy-remote-runner.sh start

# Stop the runner
./scripts/deploy-remote-runner.sh stop

# Check status
./scripts/deploy-remote-runner.sh status

# View logs (follows)
./scripts/deploy-remote-runner.sh logs

# Connect to runner shell
./scripts/deploy-remote-runner.sh shell
```

### Maintenance Commands

```bash
# Update runner code and restart
./scripts/deploy-remote-runner.sh update

# Clean up everything
./scripts/deploy-remote-runner.sh clean
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

## 🏗️ Remote Directory Structure

The runner will be deployed to `/opt/onemount-runner/` on the remote host:

```
/opt/onemount-runner/
├── .env                           # Environment configuration
├── docker-compose.runner.yml     # Docker Compose config
├── packaging/docker/              # Docker files
├── scripts/                       # Management scripts
└── ... (OneMount source code)
```

## 🔍 Troubleshooting

### SSH Connection Issues

```bash
# Test SSH connectivity
ssh -v username@172.16.1.104

# Check SSH key
ssh-add -l

# Copy SSH key if needed
ssh-copy-id username@172.16.1.104
```

### Docker Issues on Remote Host

```bash
# Check Docker status
ssh username@172.16.1.104 "docker --version && docker ps"

# Check Docker Compose
ssh username@172.16.1.104 "docker-compose --version"

# Check user permissions
ssh username@172.16.1.104 "groups"
```

### Runner Registration Issues

```bash
# Check runner logs
./scripts/deploy-remote-runner.sh logs

# Connect to runner shell for debugging
./scripts/deploy-remote-runner.sh shell

# Check GitHub repository settings
# Go to Repository → Settings → Actions → Runners
```

### Network Issues

The runner is configured for IPv4-only networking (suitable for South African networks):

```bash
# Test network connectivity from container
./scripts/deploy-remote-runner.sh shell
# Inside container: ping 8.8.8.8
```

## 📊 Monitoring

### Check Runner Status

```bash
# Quick status check
./scripts/deploy-remote-runner.sh status

# Detailed logs
./scripts/deploy-remote-runner.sh logs

# GitHub repository runners page
# Repository → Settings → Actions → Runners
```

### Container Resources

```bash
# Check container resource usage
ssh username@172.16.1.104 "docker stats"

# Check disk usage
ssh username@172.16.1.104 "df -h"
```

## 🔄 Updates and Maintenance

### Updating the Runner

```bash
# Update runner code and restart
./scripts/deploy-remote-runner.sh update
```

This will:
1. Stop the current runner
2. Deploy updated code
3. Rebuild the Docker image
4. Start the updated runner

### Backup and Restore

```bash
# Backup runner configuration
scp username@172.16.1.104:/opt/onemount-runner/.env ./backup-runner.env

# Restore configuration
scp ./backup-runner.env username@172.16.1.104:/opt/onemount-runner/.env
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
./scripts/deploy-remote-runner.sh clean

# This will:
# - Stop and remove containers
# - Remove Docker images
# - Optionally remove the project directory
```

## 📞 Support

If you encounter issues:

1. Check the troubleshooting section above
2. View runner logs: `./scripts/deploy-remote-runner.sh logs`
3. Connect to runner shell: `./scripts/deploy-remote-runner.sh shell`
4. Check GitHub repository runner status
