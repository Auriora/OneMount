# Builder Image

Base Docker image with all build tools and development dependencies for OneMount.

**Base**: Ubuntu 24.04

## What It Contains

- Go 1.24.2
- Build tools (build-essential, pkg-config, git, wget, curl)
- FUSE3 (development headers)
- GUI dependencies (development headers)
- System utilities (rsync, lsb-release, gnupg)

## Used By

This image serves as the foundation for:
- **Production builds** (`packaging/docker/Dockerfile`) - Builder stage
- **Test runner** (`docker/images/test-runner/Dockerfile`)
- **GitHub runner** (`docker/images/github-runner/Dockerfile`)
- **Debian builder** (`docker/images/deb-builder/Dockerfile`)

## Building

```bash
# Build builder image
./docker/scripts/build-images.sh builder

# Builder is automatically built when building other images
./docker/scripts/build-images.sh production  # Builds builder first
./docker/scripts/build-images.sh test-runner # Builds builder first
```

## Image Details

- **Size**: ~1.49GB (includes all build tools)
- **Purpose**: Development and building, not for production runtime
- **Tag**: `onemount-builder:${VERSION}`

## See Also

- Production image: `packaging/docker/Dockerfile`
- Build script: `docker/scripts/build-images.sh`
