# Docker-based Self-Hosted GitHub Actions Runner

This guide shows how to run the OneMount self-hosted GitHub Actions runner in a Docker container, providing a clean, isolated, and easily manageable testing environment.

## 🚀 Quick Start

### 1. Interactive Setup

```bash
# Run the interactive setup
./scripts/manage-runner.sh setup
```

This will guide you through:
- Setting up your GitHub personal access token
- Configuring the repository
- Setting up OneDrive authentication (optional)

### 2. Build and Start

```bash
# Build the runner image
./scripts/manage-runner.sh build

# Start the runner
./scripts/manage-runner.sh start

# View logs
./scripts/manage-runner.sh logs --follow
```

## 📋 Prerequisites

### Required
- Docker and Docker Compose
- GitHub personal access token with `repo` scope
- Repository admin access to configure self-hosted runners

### Optional
- OneDrive authentication tokens for system tests

## 🔧 Manual Configuration

### 1. Create GitHub Personal Access Token

1. Go to [GitHub Settings → Personal Access Tokens](https://github.com/settings/tokens)
2. Click "Generate new token (classic)"
3. Select `repo` scope
4. Copy the token

### 2. Create Environment Configuration

Create a `.env` file in the project root:

```bash
# GitHub Configuration
GITHUB_TOKEN=ghp_your_token_here
GITHUB_REPOSITORY=owner/repo
RUNNER_NAME=onemount-docker-runner
RUNNER_LABELS=self-hosted,linux,onemount-testing
RUNNER_GROUP=Default

# OneDrive Authentication (optional)
AUTH_TOKENS_B64=base64_encoded_auth_tokens
```

### 3. Generate Auth Tokens (Optional)

If you have existing OneMount authentication:

```bash
# Encode existing tokens
base64 -w 0 ~/.cache/onemount/auth_tokens.json
```

Add the output to `AUTH_TOKENS_B64` in your `.env` file.

## 🐳 Container Features

### Included Dependencies
- **Ubuntu 24.04** base image
- **Go 1.22+** for building OneMount
- **FUSE3** support for filesystem testing
- **GitHub Actions Runner** v2.311.0
- **WebKit2GTK** for GUI testing
- **IPv4-only networking** for South African networks

### Security & Isolation
- Runs as non-root `runner` user
- FUSE capabilities for filesystem testing
- Isolated network environment
- Persistent workspace volumes

### Persistent Storage
- Runner work directory: `/opt/actions-runner/_work`
- Test artifacts: `/home/runner/.onemount-tests`
- Project workspace: `/workspace`

## 📖 Usage Examples

### Basic Operations

```bash
# Build the runner image
./scripts/manage-runner.sh build

# Start the runner
./scripts/manage-runner.sh start

# Stop the runner
./scripts/manage-runner.sh stop

# Restart the runner
./scripts/manage-runner.sh restart

# View status
./scripts/manage-runner.sh status
```

### Development & Debugging

```bash
# Start development shell
./scripts/manage-runner.sh shell --dev

# Test the environment
./scripts/manage-runner.sh test

# View logs with follow
./scripts/manage-runner.sh logs --follow
```

### Direct Docker Commands

```bash
# Build image manually
docker build -f packaging/docker/Dockerfile.github-runner -t onemount-github-runner .

# Run with Docker Compose
docker-compose -f docker/compose/docker-compose.runner.yml up -d

# Interactive shell
docker-compose -f docker/compose/docker-compose.runner.yml run --rm github-runner shell
```

## 🔍 Troubleshooting

### Runner Registration Issues

1. **Invalid token error**:
   - Verify your `GITHUB_TOKEN` has `repo` scope
   - Check token hasn't expired

2. **Repository not found**:
   - Verify `GITHUB_REPOSITORY` format is `owner/repo`
   - Ensure you have admin access to the repository

3. **Runner already exists**:
   - The script uses `--replace` flag to handle existing runners
   - Check GitHub repository settings → Actions → Runners

### Authentication Issues

1. **No auth tokens**:
   ```bash
   # Check if tokens exist
   ls -la ~/.cache/onemount/auth_tokens.json
   
   # Generate tokens
   ./build/onemount --auth-only
   ```

2. **Expired tokens**:
   - Re-authenticate with OneMount
   - Update the `AUTH_TOKENS_B64` in `.env`

3. **Invalid JSON**:
   ```bash
   # Validate tokens
   jq empty ~/.cache/onemount/auth_tokens.json
   ```

### Container Issues

1. **FUSE not working**:
   - Ensure Docker has `--privileged` or proper capabilities
   - Check `/dev/fuse` is accessible

2. **Permission errors**:
   - Verify volume mounts have correct permissions
   - Check the `runner` user has access

3. **Network issues**:
   - Container uses IPv4-only configuration
   - DNS is set to 8.8.8.8 and 8.8.4.4

## 🧹 Cleanup

```bash
# Stop and remove containers
./scripts/manage-runner.sh stop

# Complete cleanup (removes volumes and images)
./scripts/manage-runner.sh clean
```

## 🔗 Integration with GitHub Actions

Once the runner is started, it will appear in your repository's Actions settings:

1. Go to **Repository Settings → Actions → Runners**
2. You should see your runner listed as "Online"
3. The runner will have labels: `self-hosted`, `linux`, `onemount-testing`

The `system-tests-self-hosted.yml` workflow will automatically use this runner when it's available.

## 📊 Advantages

1. **Easy Setup**: One-command setup and management
2. **Isolation**: Complete isolation from host system
3. **Reproducible**: Consistent environment across deployments
4. **Portable**: Can run on any Docker-capable system
5. **Persistent**: Maintains state across container restarts
6. **Secure**: Runs with minimal privileges and proper isolation

## 🔄 Workflow Integration

The runner automatically integrates with the existing `system-tests-self-hosted.yml` workflow:

- Runs all system tests with `--all` flag
- Uses pre-configured authentication
- Uploads test logs with 30-day retention
- Handles cleanup automatically

This provides a complete CI/CD solution for OneMount testing without requiring dedicated hardware or complex setup.
