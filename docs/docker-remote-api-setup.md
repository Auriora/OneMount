# Remote Docker API Setup for OneMount GitHub Actions Runner

This guide shows how to deploy the OneMount GitHub Actions runner to a remote Docker host using Docker's TCP API on port 2375.

## 🚀 Quick Setup

### Prerequisites

1. **Remote Docker Host**: Docker daemon running on 172.16.1.104 with TCP API enabled on port 2375
2. **Network Access**: Port 2375 accessible from your local machine
3. **GitHub Token**: Personal access token with `repo` scope
4. **Local Docker**: Docker client installed locally

### 1. Test Connection

```bash
# Test connection to remote Docker
DOCKER_HOST=tcp://172.16.1.104:2375 docker version
```

### 2. Interactive Setup

```bash
# Run the interactive setup
./scripts/deploy-docker-remote.sh setup
```

This will:
- Test connection to the remote Docker host
- Configure GitHub authentication
- Set up OneDrive authentication (optional)
- Save configuration for future use

### 3. Build and Deploy

```bash
# Build the runner image on remote Docker
./scripts/deploy-docker-remote.sh build

# Deploy and start the runner
./scripts/deploy-docker-remote.sh deploy

# Check status
./scripts/deploy-docker-remote.sh status
```

## 🔧 Remote Docker Host Configuration

### Enable Docker TCP API

On the remote host (172.16.1.104), configure Docker to accept TCP connections:

#### Option 1: Systemd Override (Recommended)

```bash
# On the remote host
sudo mkdir -p /etc/systemd/system/docker.service.d

# Create override file
sudo tee /etc/systemd/system/docker.service.d/override.conf << EOF
[Service]
ExecStart=
ExecStart=/usr/bin/dockerd -H fd:// -H tcp://0.0.0.0:2375
EOF

# Reload and restart Docker
sudo systemctl daemon-reload
sudo systemctl restart docker

# Verify Docker is listening on port 2375
sudo netstat -tlnp | grep :2375
```

#### Option 2: Docker Daemon Configuration

```bash
# On the remote host
sudo tee /etc/docker/daemon.json << EOF
{
  "hosts": ["fd://", "tcp://0.0.0.0:2375"],
  "ipv6": false,
  "fixed-cidr": "172.17.0.0/16"
}
EOF

# Restart Docker
sudo systemctl restart docker
```

### Security Considerations

⚠️ **Warning**: Port 2375 provides unencrypted, unauthenticated access to Docker. Only use this in trusted networks.

For production environments, consider:
- Using Docker TLS (port 2376) with certificates
- Restricting access with firewall rules
- Using SSH tunneling instead

## 📋 Available Commands

### Setup and Deployment

```bash
# Interactive setup
./scripts/deploy-docker-remote.sh setup

# Build image on remote Docker
./scripts/deploy-docker-remote.sh build

# Deploy runner container
./scripts/deploy-docker-remote.sh deploy

# Build and deploy in one step
./scripts/deploy-docker-remote.sh build && ./scripts/deploy-docker-remote.sh deploy
```

### Container Management

```bash
# Start the runner
./scripts/deploy-docker-remote.sh start

# Stop the runner
./scripts/deploy-docker-remote.sh stop

# Restart the runner
./scripts/deploy-docker-remote.sh restart

# Check status
./scripts/deploy-docker-remote.sh status
```

### Monitoring and Debugging

```bash
# View logs (follows)
./scripts/deploy-docker-remote.sh logs

# Connect to runner shell
./scripts/deploy-docker-remote.sh shell

# Check container and host status
./scripts/deploy-docker-remote.sh status
```

### Maintenance

```bash
# Update runner (rebuild and redeploy)
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

## 🏗️ Container Architecture

### Image Features
- **Ubuntu 24.04** base with Go 1.22+
- **GitHub Actions Runner** v2.311.0
- **FUSE3 support** for filesystem testing
- **WebKit2GTK** for GUI testing
- **IPv4-only networking** configuration

### Persistent Volumes
- `onemount-runner-workspace`: Project workspace
- `onemount-runner-work`: Runner work directory
- `onemount-runner-logs`: Test logs and artifacts

### Container Configuration
- **Name**: `onemount-github-runner`
- **Restart Policy**: `unless-stopped`
- **Labels**: `self-hosted,linux,onemount-testing,docker-remote`
- **DNS**: 8.8.8.8, 8.8.4.4 (IPv4-only)
- **Compose File**: `docker/compose/docker-compose.remote.yml`

## 🔍 Troubleshooting

### Connection Issues

```bash
# Test Docker connection
DOCKER_HOST=tcp://172.16.1.104:2375 docker version

# Check if port is open
telnet 172.16.1.104 2375

# Test from remote host
curl http://172.16.1.104:2375/version
```

### Docker Daemon Issues

```bash
# Check Docker status on remote host
ssh user@172.16.1.104 "sudo systemctl status docker"

# Check Docker logs on remote host
ssh user@172.16.1.104 "sudo journalctl -u docker -f"

# Verify TCP binding
ssh user@172.16.1.104 "sudo netstat -tlnp | grep :2375"
```

### Runner Registration Issues

```bash
# Check runner logs
./scripts/deploy-docker-remote.sh logs

# Connect to container for debugging
./scripts/deploy-docker-remote.sh shell

# Check GitHub repository runners
# Go to Repository → Settings → Actions → Runners
```

### Container Issues

```bash
# Check container status
DOCKER_HOST=tcp://172.16.1.104:2375 docker ps -a

# Inspect container
DOCKER_HOST=tcp://172.16.1.104:2375 docker inspect onemount-github-runner

# Check container logs
DOCKER_HOST=tcp://172.16.1.104:2375 docker logs onemount-github-runner
```

## 📊 Monitoring

### Container Status

```bash
# Quick status check
./scripts/deploy-docker-remote.sh status

# Detailed container info
DOCKER_HOST=tcp://172.16.1.104:2375 docker stats onemount-github-runner

# Check volumes
DOCKER_HOST=tcp://172.16.1.104:2375 docker volume ls
```

### GitHub Integration

Once deployed, the runner will:

1. **Register** with your GitHub repository
2. **Appear** in Repository → Settings → Actions → Runners
3. **Accept jobs** from `system-tests-self-hosted.yml`
4. **Run tests** with pre-configured authentication
5. **Upload artifacts** automatically

## 🔄 Direct Docker Commands

If you prefer using Docker commands directly:

```bash
# Set Docker host
export DOCKER_HOST=tcp://172.16.1.104:2375

# Build image
docker build -f packaging/docker/Dockerfile.github-runner -t onemount-github-runner .

# Run container
docker run -d \
  --name onemount-github-runner \
  --restart unless-stopped \
  --device /dev/fuse \
  --cap-add SYS_ADMIN \
  --security-opt apparmor:unconfined \
  -e GITHUB_TOKEN=your_token \
  -e GITHUB_REPOSITORY=owner/repo \
  -e RUNNER_NAME=onemount-runner-remote \
  -v onemount-runner-workspace:/workspace \
  onemount-github-runner run

# Check status
docker ps
docker logs -f onemount-github-runner
```

## 🧹 Cleanup

To completely remove the runner:

```bash
# Clean up everything
./scripts/deploy-docker-remote.sh clean

# Or manually
DOCKER_HOST=tcp://172.16.1.104:2375 docker stop onemount-github-runner
DOCKER_HOST=tcp://172.16.1.104:2375 docker rm onemount-github-runner
DOCKER_HOST=tcp://172.16.1.104:2375 docker rmi onemount-github-runner
```

## 🎯 Advantages

1. **Simple Deployment**: No SSH setup required
2. **Direct Control**: Full Docker API access
3. **Fast Operations**: Direct TCP connection
4. **Easy Debugging**: Direct container access
5. **Persistent Storage**: Survives container restarts
6. **Automatic Restart**: Container restarts on failure

This setup provides a robust, easily manageable GitHub Actions runner that integrates seamlessly with your existing OneMount testing workflows.
