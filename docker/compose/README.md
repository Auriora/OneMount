# OneMount Docker Compose Configurations

This directory contains multiple Docker Compose configurations for different purposes. Each file serves a specific role in the OneMount development and deployment workflow.

## File Overview

### `docker-compose.test.yml`
**Purpose**: Testing and development workflows
- **Services**: test-runner, unit-tests, integration-tests, system-tests, coverage, shell
- **Usage**: Local development testing, CI/CD pipelines
- **Security**: Uses test-specific auth tokens, no production data access
- **Volume Mounts**: Project source, test artifacts
- **Key Features**:
  - Isolated test environment
  - FUSE support for filesystem testing
  - Multiple test type configurations
  - Coverage analysis support

### `docker-compose.build.yml`
**Purpose**: Building Docker images
- **Services**: Image building configurations
- **Usage**: CI/CD image building, development image creation
- **Key Features**:
  - Multi-stage builds
  - Caching optimization
  - Development vs production builds

### `docker-compose.runner.yml` (Singular)
**Purpose**: Single GitHub Actions runner for development/testing
- **Services**: github-runner, runner-dev
- **Usage**: Development testing, debugging GitHub Actions workflows
- **Security**: ⚠️ **DEPRECATED** - Previously mounted production tokens (now removed)
- **Key Features**:
  - Single runner instance
  - Development-focused
  - Interactive debugging support

### `docker-compose.runners.yml` (Plural)
**Purpose**: Production GitHub Actions runners (2-runner setup)
- **Services**: runner-1, runner-2
- **Usage**: Production CI/CD, automated testing
- **Security**: ✅ **SECURE** - Uses `AUTH_TOKENS_B64` environment variable
- **Key Features**:
  - Two persistent runners
  - Production-ready configuration
  - Proper secret management
  - Restart policies for reliability

### `docker-compose.remote.yml`
**Purpose**: Remote deployment configurations
- **Services**: Remote runner configurations
- **Usage**: Deployment to remote Docker hosts
- **Key Features**:
  - Remote deployment optimized
  - Network configuration for remote access
  - Persistent storage management

## Security Best Practices

### ✅ SECURE Configurations
- `docker-compose.runners.yml`: Uses `AUTH_TOKENS_B64` environment variable
- `docker-compose.test.yml`: Uses dedicated test tokens in test-artifacts
- `docker-compose.remote.yml`: Proper secret management

### ⚠️ SECURITY WARNINGS
- **NEVER** mount production auth tokens directly into containers
- **ALWAYS** use dedicated test OneDrive accounts for testing
- **AVOID** mounting `${HOME}/.cache/onemount/auth_tokens.json` in any container

## Usage Examples

### Running Tests
```bash
# Unit tests
docker compose -f docker-compose.test.yml run --rm unit-tests

# System tests (requires test auth tokens)
docker compose -f docker-compose.test.yml run --rm system-tests

# Interactive debugging
docker compose -f docker-compose.test.yml run --rm shell
```

### GitHub Actions Runners
```bash
# Production runners (recommended)
docker compose -f docker-compose.runners.yml up -d

# Development runner
docker compose -f docker-compose.runner.yml run --rm runner-dev shell
```

### Building Images

### Unified CLI (recommended)
```bash
# Build with deterministic, content-based tags (derived from git)
./scripts/dev.py build images --profile build

# Dev variant
./scripts/dev.py build images --profile build-dev

# Force a specific tag
./scripts/dev.py build images --profile build --version v1.2.3+local
```

Notes:
- Tags are derived as <tag>.<distance>+<sha>[-dirty] when available, else 0.0.0+<sha>[-dirty].
- No :latest is used; cache_from and image tags align with ONEMOUNT_VERSION.

```bash
# Build test runner image
docker compose -f docker-compose.build.yml build test-runner

# Build with no cache
docker compose -f docker-compose.build.yml build --no-cache test-runner
```

## Build caching and .dockerignore guidance

To keep builds fast, deterministic, and idempotent:

- Minimize build context to stabilize cache keys and reduce transfer time
- Avoid copying the entire repo in Dockerfiles (no `COPY . .`); copy only what is needed
- Use a root `.dockerignore` to exclude volatile and large directories/files
- Leverage BuildKit and inline cache via Compose `cache_from` aligned with content-based tags

### Recommended `.dockerignore` (snippet)
```dockerignore
.git/
build/
dist/
.venv/
venv/
node_modules/
docs/
test-artifacts/
*_test.go
```

### Dockerfile cache pattern (already applied)
- Stage modules separately for maximal cache reuse:
  1) `COPY go.mod go.sum ./` → `RUN go mod download`
  2) `COPY cmd/ internal/ scripts/cgo-helper.sh ...`
- Define `ARG ONEMOUNT_VERSION` with a default to avoid UndefinedVar warnings
- Keep final build steps dependent only on copied source, not on `.git` or other volatile inputs

### BuildKit and idempotence check
```bash
# Enable BuildKit and use compose build
export DOCKER_BUILDKIT=1 COMPOSE_DOCKER_CLI_BUILD=1

# Example (service name may be test-runner or test-runner-build depending on compose file)
docker compose -f docker-compose.build.yml build test-runner-build
# Immediately repeat (no changes)
docker compose -f docker-compose.build.yml build test-runner-build
# Expectation: second run should report CACHED for all steps
```


## Auth Token Management

### For Testing
1. **Create dedicated test OneDrive account**
2. **Authenticate with test account**:
   ```bash
   ./build/onemount --auth-only
   ```
3. **Copy to test location**:
   ```bash
   mkdir -p ~/.onemount-tests
   cp ~/.cache/onemount/auth_tokens.json ~/.onemount-tests/.auth_tokens.json
   ```

### For Production Runners
1. **Use environment variable**:
   ```bash
   export AUTH_TOKENS_B64=$(base64 -w 0 ~/.onemount-tests/.auth_tokens.json)
   ```
2. **Never mount production token files directly**

## Troubleshooting

### Common Issues

1. **Permission Errors**:
   - Ensure proper user mapping: `USER_ID` and `GROUP_ID`
   - Check volume mount permissions

2. **Auth Token Errors**:
   - Verify test tokens exist: `ls -la ~/.onemount-tests/.auth_tokens.json`
   - Check token format: `jq . ~/.onemount-tests/.auth_tokens.json`

3. **Container Conflicts**:
   - Remove failed containers: `docker rm -f onemount-*-test`
   - Clean up volumes: `docker volume prune`

4. **Image Build Failures**:
   - Build base image first: `docker compose -f docker-compose.build.yml build base-build`
   - Check Docker daemon status

### Debug Commands
```bash
# Check container status
docker ps -a | grep onemount

# View container logs
docker logs onemount-unit-test

# Interactive shell in test container
docker compose -f docker-compose.test.yml run --rm shell

# Clean up all OneMount containers
docker rm -f $(docker ps -aq --filter "name=onemount-")
```

## Migration Notes

### From Legacy Scripts
- Replace `./scripts/run-tests-docker.sh` with `docker compose` commands
- Use `scripts/dev.py test docker` for unified CLI experience
- Update CI/CD pipelines to use compose files

### Security Migration
- Remove any direct production token mounts
- Migrate to `AUTH_TOKENS_B64` environment variable
- Update documentation to reflect secure practices
