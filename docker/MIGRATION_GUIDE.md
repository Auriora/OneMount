# Docker Refactoring Migration Guide

This guide helps you migrate from the old Docker structure to the new refactored structure.

## What Changed?

### Directory Structure

**Before**:
```
packaging/docker/
├── Dockerfile.base
├── Dockerfile.github-runner (production + development mixed)
├── Dockerfile.test-runner (production + development mixed)
├── Dockerfile.deb-builder
├── runner-entrypoint.sh
├── test-entrypoint.sh
├── init-workspace.sh
├── token-manager.sh
└── python-helper.sh

docker/compose/
├── docker-compose.build.yml
├── docker-compose.test.yml
├── docker-compose.runner.yml
├── docker-compose.runners.yml
└── docker-compose.remote.yml
```

**After**:
```
packaging/docker/              # Production only
├── Dockerfile.base            # Shared base image
├── Dockerfile.deb-builder     # Package builder
├── docker-compose.yml         # Production packaging/deployment
├── install-deps.sh
└── README.md

docker/                        # Development only
├── Dockerfile.github-runner   # Development runner
├── Dockerfile.test-runner     # Development test runner
├── scripts/                   # Entrypoint scripts
│   ├── runner-entrypoint.sh
│   ├── test-entrypoint.sh
│   ├── init-workspace.sh
│   ├── token-manager.sh
│   └── python-helper.sh
├── compose/                   # Development compose files
│   ├── docker-compose.build.yml
│   ├── docker-compose.test.yml
│   ├── docker-compose.runner.yml
│   ├── docker-compose.runners.yml
│   └── docker-compose.remote.yml
└── README.md
```

## Migration Steps

### For Developers (Local Testing)

**No action required!** All existing commands continue to work:

```bash
# These still work exactly the same
docker compose -f docker/compose/docker-compose.test.yml run --rm unit-tests
docker compose -f docker/compose/docker-compose.build.yml build
python scripts/dev.py test docker unit
```

**What you'll notice**:
- Container names are now readable (e.g., `onemount-unit-tests`)
- Easier to find containers: `docker ps --filter "name=onemount"`

### For CI/CD Pipelines

**No changes required.** All compose file paths remain the same.

### For Scripts That Build Images

If you have custom scripts that build Docker images, update the Dockerfile paths:

**Before**:
```bash
docker build -f packaging/docker/Dockerfile.github-runner .
docker build -f packaging/docker/Dockerfile.test-runner .
```

**After**:
```bash
# For development/testing
docker build -f docker/Dockerfile.github-runner .
docker build -f docker/Dockerfile.test-runner .

# For production packaging
docker compose -f packaging/docker/docker-compose.yml build
```

### For Production Deployment

**New capability!** You can now use the production compose file:

```bash
# Build production images
docker compose -f packaging/docker/docker-compose.yml build

# Build Debian packages
docker compose -f packaging/docker/docker-compose.yml run --rm deb-builder

# Deploy production container
docker compose -f packaging/docker/docker-compose.yml --profile production up -d
```

## Key Improvements

### 1. Clear Separation
- **Production** (`packaging/docker/`): For packaging and deployment
- **Development** (`docker/`): For local testing and development

### 2. Better Container Names
All containers now have descriptive names:
- `onemount-unit-tests` instead of `docker-compose-test-1-abc123`
- `onemount-runner-1` instead of `docker-compose-runner-1-xyz789`

### 3. Project Names
All compose files have project names for better organization:
- `onemount-build`
- `onemount-test`
- `onemount-runner`
- `onemount-runners`
- `onemount-packaging`

### 4. Centralized Scripts
All entrypoint and helper scripts are now in `docker/scripts/`:
- Easier to find and maintain
- Single source of truth
- Used by all Docker images

### 5. Production Compose File
New `packaging/docker/docker-compose.yml` for:
- Building production images
- Creating Debian packages
- Deploying production containers

## Common Tasks

### Building Images

**Development images**:
```bash
docker compose -f docker/compose/docker-compose.build.yml build
```

**Production images**:
```bash
docker compose -f packaging/docker/docker-compose.yml build
```

### Running Tests

```bash
# Unit tests
docker compose -f docker/compose/docker-compose.test.yml run --rm unit-tests

# Integration tests
docker compose -f docker/compose/docker-compose.test.yml run --rm integration-tests

# System tests
docker compose -f docker/compose/docker-compose.test.yml run --rm system-tests
```

### Building Packages

```bash
# Build Debian package
docker compose -f packaging/docker/docker-compose.yml run --rm deb-builder

# Output in dist/
ls -la dist/*.deb
```

### Managing Runners

**Development (single runner)**:
```bash
docker compose -f docker/compose/docker-compose.runner.yml up -d
```

**Production (multiple runners)**:
```bash
docker compose -f docker/compose/docker-compose.runners.yml up -d
```

### Production Deployment

```bash
# Start production container
docker compose -f packaging/docker/docker-compose.yml --profile production up -d

# Check status
docker ps --filter "name=onemount-production"

# View logs
docker logs onemount-production

# Stop
docker compose -f packaging/docker/docker-compose.yml --profile production down
```

## Troubleshooting

### "File not found" errors during build

If you see errors like `COPY failed: file not found: packaging/docker/runner-entrypoint.sh`:

**Cause**: Dockerfile references old script location

**Fix**: Scripts moved to `docker/scripts/`. Rebuild images:
```bash
docker compose -f docker/compose/docker-compose.build.yml build --no-cache
```

### Container name conflicts

If you see "container name already in use":

**Cause**: Old containers with same names

**Fix**: Remove old containers:
```bash
docker rm -f onemount-test-runner
# Or remove all OneMount containers
docker rm -f $(docker ps -aq --filter "name=onemount")
```

### Image not found

If you see "image not found" errors:

**Cause**: Images need to be rebuilt

**Fix**: Build the required images:
```bash
# Development images
docker compose -f docker/compose/docker-compose.build.yml build

# Production images
docker compose -f packaging/docker/docker-compose.yml build
```

## FAQ

### Q: Do I need to rebuild all images?

**A**: Not immediately. Existing images will continue to work. Rebuild when:
- You need to update to the latest code
- You encounter "file not found" errors
- You want to use the new production compose file

### Q: Will this break my CI/CD pipeline?

**A**: No. All compose file paths remain the same. Container behavior is identical.

### Q: Can I still use the old Dockerfiles?

**A**: No. The old Dockerfiles in `packaging/docker/` have been removed. Use:
- `docker/Dockerfile.*` for development
- `packaging/docker/docker-compose.yml` for production

### Q: Where did the entrypoint scripts go?

**A**: Moved to `docker/scripts/`. The Dockerfiles have been updated to reference the new location.

### Q: What's the difference between runner.yml and runners.yml?

**A**: 
- `docker-compose.runner.yml` (singular): Single runner for development/testing
- `docker-compose.runners.yml` (plural): Two runners for production CI/CD

### Q: How do I build production packages now?

**A**: Use the new production compose file:
```bash
docker compose -f packaging/docker/docker-compose.yml run --rm deb-builder
```

## Getting Help

- **Organization Guide**: `docker/DOCKER_ORGANIZATION.md`
- **Development Guide**: `docker/README.md`
- **Production Guide**: `packaging/docker/README.md`
- **Compose Guide**: `docker/compose/README.md`
- **Change Log**: `docs/updates/2025-11-11-docker-refactoring.md`

## Rollback (If Needed)

If you need to rollback to the old structure:

```bash
# Checkout the previous commit
git log --oneline | grep -B1 "docker-refactoring"
git checkout <previous-commit-hash>

# Or revert the refactoring commit
git revert <refactoring-commit-hash>
```

**Note**: Rollback should not be necessary as the refactoring maintains backward compatibility.
