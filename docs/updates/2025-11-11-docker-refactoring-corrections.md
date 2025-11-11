# Docker Refactoring Corrections

**Date**: 2025-11-11  
**Author**: AI Agent  
**Status**: Completed  
**Related**: `docs/updates/2025-11-11-docker-refactoring-implementation.md`

## Issue

Initial implementation incorrectly:
1. Named production base image as `Dockerfile.base` instead of `Dockerfile`
2. Included build tools in what should be the production image
3. Did not create a proper minimal runtime-only production image

## Root Cause

Misunderstood the requirement. The production image should be:
- Named `Dockerfile` (standard convention)
- Multi-stage build with minimal runtime
- NO build tools in final image (no git, wget, build-essential, rsync, pkg-config, Go compiler)
- Only runtime libraries and compiled binaries

## Corrections Made

### 1. File Renames

- `packaging/docker/Dockerfile.base` → `packaging/docker/Dockerfile.builder`
- Created new `packaging/docker/Dockerfile` (production, multi-stage, minimal)

### 2. Production Dockerfile (packaging/docker/Dockerfile)

**Builder stage** (discarded after build):
- Ubuntu 24.04
- Go 1.24.2
- Build tools: build-essential, pkg-config, git, wget
- Compiles OneMount binaries

**Runtime stage** (final image):
```dockerfile
# Install ONLY runtime dependencies (no build tools, no git, no wget)
RUN apt-get update && apt-get install -y \
    fuse3 \
    libfuse3-3 \
    libgtk-3-0 \
    libwebkit2gtk-4.1-0 \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*
```

**What's excluded from production**:
- ❌ build-essential
- ❌ pkg-config
- ❌ git
- ❌ wget
- ❌ rsync
- ❌ Go compiler
- ❌ Source code
- ❌ Development headers (-dev packages)

**What's included in production**:
- ✅ FUSE3 runtime (fuse3, libfuse3-3)
- ✅ GUI runtime (libgtk-3-0, libwebkit2gtk-4.1-0)
- ✅ CA certificates
- ✅ Compiled binaries only

**Expected size**: <500MB (vs 1.49GB for builder image)

### 3. Builder Image (packaging/docker/Dockerfile.builder)

This is the former "base" image, now correctly named:
- Contains all build tools
- Used by development images (test-runner, github-runner, deb-builder)
- Size: ~1.49GB

### 4. Updated References

All Dockerfiles now reference `onemount-builder` instead of `onemount-base`:
- `docker/images/test-runner/Dockerfile`
- `docker/images/github-runner/Dockerfile`
- `docker/images/deb-builder/Dockerfile`

### 5. Updated Build Script

`docker/scripts/build-images.sh` now has correct targets:
- `builder` - Build builder image (with build tools)
- `production` - Build production image (runtime only)
- `dev` - Build all development images
- `all` - Build everything

### 6. Updated Documentation

All READMEs updated to reflect:
- Production image is `Dockerfile` (not `Dockerfile.base`)
- Builder image is `Dockerfile.builder` (not `Dockerfile`)
- Production image has NO build tools
- Clear distinction between builder (dev) and production (runtime)

## Verification

### Production Image Has NO Build Tools

```bash
$ sed -n '/^# Runtime stage/,/^LABEL/p' packaging/docker/Dockerfile | grep -A 20 "apt-get install"
RUN apt-get update && apt-get install -y \
    fuse3 \
    libfuse3-3 \
    libgtk-3-0 \
    libwebkit2gtk-4.1-0 \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*
```

✅ No build-essential  
✅ No git  
✅ No wget  
✅ No rsync  
✅ No pkg-config  
✅ No development headers  

### Builder Image Has Build Tools

```bash
$ grep "build-essential" packaging/docker/Dockerfile.builder
    build-essential \
```

✅ Contains all necessary build tools for development

## File Structure

```
packaging/docker/
├── Dockerfile           # Production (multi-stage, runtime only, NO build tools)
├── Dockerfile.builder   # Builder (with build tools, for development)
├── docker-compose.yml   # Production deployment
└── README.md            # Updated documentation
```

## Build Commands

```bash
# Build production image (minimal, no build tools)
./docker/scripts/build-images.sh production

# Build builder image (with build tools)
./docker/scripts/build-images.sh builder

# Build all development images
./docker/scripts/build-images.sh dev

# Build everything
./docker/scripts/build-images.sh all
```

## Summary

Corrected the Docker setup to properly separate:
1. **Production image** (`Dockerfile`) - Multi-stage, minimal runtime, NO build tools
2. **Builder image** (`Dockerfile.builder`) - Full build environment for development

This follows Docker best practices:
- Production images are minimal and secure
- Build tools are not shipped to production
- Multi-stage builds reduce final image size
- Standard naming convention (Dockerfile for production)
