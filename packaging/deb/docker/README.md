# Debian Package Builder Docker Image

This directory contains the Docker configuration for building OneMount Debian packages.

## Overview

The Debian builder image extends the base OneMount image with Debian packaging tools and provides a standardized build environment.

## Building the Image

The image will be built automatically when using the build compose file:

```bash
# Build Debian package (image builds automatically if needed)
docker compose -f docker/compose/docker-compose.build.yml --profile deb run --rm build-deb
```

Or build manually:

```bash
docker build -f packaging/deb/docker/Dockerfile -t onemount-deb-builder:latest .
```

## Using the Builder

### Via Docker Compose (Recommended)

```bash
# Build Debian package
docker compose -f docker/compose/docker-compose.build.yml --profile deb run --rm build-deb

# Build binaries
docker compose -f docker/compose/docker-compose.build.yml --profile binaries run --rm build-binaries

# Clean build artifacts
docker compose -f docker/compose/docker-compose.build.yml --profile clean run --rm clean-build
```

### Via Docker Run

```bash
# Build Debian package
docker run --rm -v $(pwd):/build -w /build onemount-deb-builder deb --output /dist

# Build binaries
docker run --rm -v $(pwd):/build -w /build onemount-deb-builder binaries

# Show help
docker run --rm onemount-deb-builder help
```

## Build Entrypoint

The image includes a build entrypoint script that provides:

- **binaries** - Build OneMount binaries (onemount, onemount-launcher)
- **deb** - Build Debian package
- **test** - Run tests
- **clean** - Clean build artifacts
- **help** - Show help message

### Options

- `--verbose` - Enable verbose output
- `--no-cache` - Disable build cache
- `--output DIR` - Output directory (default: /workspace/build)

### Examples

```bash
# Build with verbose output
docker run --rm -v $(pwd):/build -w /build onemount-deb-builder binaries --verbose

# Build to custom output directory
docker run --rm -v $(pwd):/build -w /build onemount-deb-builder deb --output /custom/path

# Build without cache
docker run --rm -v $(pwd):/build -w /build onemount-deb-builder binaries --no-cache
```

## BuildKit Cache

The Dockerfile uses BuildKit cache mounts for faster builds:

```bash
# Enable BuildKit
export DOCKER_BUILDKIT=1

# Build with cache
docker build -f packaging/deb/docker/Dockerfile -t onemount-deb-builder:latest .
```

## Files

- `Dockerfile` - Debian builder image definition
- `README.md` - This file

## Output

### Binaries

Built binaries are output to `build/binaries/`:
- `onemount` - Main OneMount binary
- `onemount-launcher` - GUI launcher

### Debian Package

Built packages are output to `dist/`:
- `onemount_*.deb` - Debian package
- `onemount_*.buildinfo` - Build information
- `onemount_*.changes` - Changes file

## Requirements

- Docker with BuildKit support
- Base OneMount image (`onemount-base`)
- Source code mounted at `/build`

## Troubleshooting

### Build Fails

Check that the base image is available:
```bash
docker images | grep onemount-base
```

If not, build it first:
```bash
docker build -f packaging/docker/Dockerfile.base -t onemount-base:latest .
```

### Permission Issues

Ensure the builder user has write access to output directories:
```bash
chmod -R 777 build/ dist/
```

### Cache Issues

Clear BuildKit cache:
```bash
docker builder prune
```

## See Also

- Base image: `packaging/docker/Dockerfile.base`
- Build compose: `docker/compose/docker-compose.build.yml`
- Build entrypoint: `docker/scripts/build-entrypoint.sh`
