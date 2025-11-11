# Production Docker Configuration

Base Docker image and production deployment configuration for OneMount.

For development and testing, see `docker/` directory.

## Files

- **Dockerfile** - Production image (multi-stage: builder + minimal runtime, NO build tools)
- **Dockerfile.builder** - Builder image (Ubuntu 24.04, Go 1.24.2, FUSE3, build tools)
- **docker-compose.yml** - Production deployment configuration
- **.dockerignore** - Build context exclusions

## Production Image (Dockerfile)

The production image is a **multi-stage build** that creates a minimal runtime-only image:

**Builder stage**:
- Ubuntu 24.04 LTS
- Go 1.24.2
- Build tools (build-essential, pkg-config, git, wget)
- Compiles OneMount binaries

**Runtime stage** (final image):
- Ubuntu 24.04 LTS (minimal)
- FUSE3 runtime libraries only
- GUI runtime libraries only
- Compiled binaries from builder
- **NO build tools** (no git, wget, build-essential, Go compiler)
- **NO source code**
- Target size: <500MB (vs 1.49GB for builder image)

**Building**:
```bash
./docker/scripts/build-images.sh production
```

## Builder Image (Dockerfile.builder)

The builder image provides the foundation for development containers:

- Ubuntu 24.04 LTS
- Go 1.24.2
- FUSE3 support
- GUI dependencies (WebKit2GTK)
- Build tools with CGO support
- IPv4-only networking

**Used by**:
- `docker/images/test-runner/` - Test execution
- `docker/images/github-runner/` - CI/CD runners
- `docker/images/deb-builder/` - Debian package builder

**Building**:
```bash
./docker/scripts/build-images.sh builder
```

## Production Deployment

### Start Production Container

```bash
docker compose -f packaging/docker/docker-compose.yml up -d
```

### Configuration

The production container includes:
- FUSE support for filesystem operations
- Resource limits (2GB RAM, 2 CPUs)
- Health checks (mount point monitoring)
- Automatic restart policy
- Persistent volumes for data, config, and cache

### Volumes

- `onemount-data` - OneDrive mount point
- `onemount-config` - Configuration files
- `onemount-cache` - Cache directory

### Environment Variables

- `ONEMOUNT_VERSION` - Version tag (default: latest)
- `ONEMOUNT_LOG_LEVEL` - Log level (default: info)
- `ONEMOUNT_MOUNT_POINT` - Mount point (default: /mnt/onedrive)

## Quick Reference

### Building

```bash
# Build production image (runtime only, no build tools)
./docker/scripts/build-images.sh production

# Build builder image (with build tools, for development)
./docker/scripts/build-images.sh builder
```

### Deployment

```bash
# Start production container
docker compose -f packaging/docker/docker-compose.yml up -d

# View logs
docker logs onemount

# Stop container
docker compose -f packaging/docker/docker-compose.yml down
```

## See Also

- Development Docker: `docker/README.md`
- Build script: `docker/scripts/build-images.sh`
- Test configuration: `docker/compose/docker-compose.test.yml`
- Build configuration: `docker/compose/docker-compose.build.yml`
