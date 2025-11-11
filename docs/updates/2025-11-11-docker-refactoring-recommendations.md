# Docker Environment Refactoring Recommendations

**Date**: 2025-11-11  
**Author**: AI Agent  
**Status**: Proposed  
**Priority**: Medium

## Executive Summary

After reviewing the Docker environment setup in `docker/` and `packaging/`, I've identified several opportunities for improvement in organization, efficiency, and maintainability. The current setup is functional but has some duplication, inconsistencies, and areas where best practices could be better applied.

## Current State Analysis

### Strengths

1. **Clear separation** between development (`docker/`) and production (`packaging/docker/`)
2. **Base image pattern** reduces duplication across derived images
3. **BuildKit cache mounts** optimize build times
4. **Comprehensive documentation** in README files
5. **Profile-based compose files** enable flexible deployment scenarios
6. **Security-conscious** with non-root users and capability management

### Issues Identified

#### 1. Image Build Process Not Streamlined

**Problem**: No unified way to build all Docker images. Users must know individual Dockerfile locations and build commands. The compose files reference images but don't provide a clear build workflow.

**Impact**: Users must manually build images with long docker build commands, leading to confusion and inconsistency.

**Evidence**:
- `docker-compose.test.yml` has `image: onemount-test-runner:latest` with no build context
- `docker-compose.build.yml` is for building binaries/packages, not Docker images (correct usage)
- No single script to build all required images with proper dependency order
- Documentation shows manual `docker build` commands

#### 2. Inconsistent Base Image References

**Problem**: Different Dockerfiles reference the base image inconsistently:
- `docker/Dockerfile.test-runner`: `ARG BASE_IMAGE=onemount-base:0.1.0rc1`
- `docker/Dockerfile.github-runner`: `FROM onemount-base` (no version)
- `packaging/deb/docker/Dockerfile`: `FROM onemount-base` (no version)

**Impact**: Version mismatches, unclear which base version is being used.

#### 3. Duplicate Environment Setup Logic

**Problem**: Similar environment setup code appears in multiple Dockerfiles:
- Go environment variables repeated in base + all derived images
- User creation patterns duplicated
- Directory creation patterns duplicated

**Evidence**: Compare lines 20-30 in each Dockerfile - nearly identical Go setup.

#### 4. Entrypoint Script Duplication

**Problem**: Common patterns repeated across entrypoint scripts:
- Color output functions (identical in all 3 scripts)
- Print functions (identical in all 3 scripts)
- Help/usage patterns (similar structure)
- Environment validation logic

**Lines of duplication**: ~100 lines of identical bash functions across 3 files.

#### 5. Volume Mount Inconsistencies

**Problem**: Test artifacts mounted to different paths:
- Test runner: `/tmp/home-tester/.onemount-tests`
- GitHub runner: `/opt/onemount-ci`
- Documentation mentions: `test-artifacts/.auth_tokens.json`

**Impact**: Confusion about where to place auth tokens, inconsistent behavior.

#### 6. Missing Multi-Stage Build Optimization

**Problem**: Each Dockerfile builds from scratch rather than using multi-stage builds to share layers. Production images include development tools, source code, and build dependencies that aren't needed at runtime.

**Impact**: 
- Larger images (production image is 1.49GB when it could be <500MB)
- Slower builds and deployments
- More disk space usage
- Security risk from unnecessary packages in production
- Increased attack surface

**Evidence**:
- Base image includes `build-essential`, `git`, `wget`, `pkg-config` (build-only tools)
- Production image would include Go compiler, source code, test files
- No separation between build-time and runtime dependencies

#### 7. Missing Unified Build Script

**Problem**: 
- No single script to build all Docker images
- Users must remember multiple `docker build` commands with different paths
- No clear documentation of build order and dependencies
- Build commands scattered across documentation
- Not using modern `docker buildx build` with enhanced features

#### 8. Using Legacy Docker Build Command

**Problem**: Build scripts and documentation use legacy `docker build` instead of modern `docker buildx build`.

**Impact**: Missing out on:
- Better BuildKit integration and features
- Multi-platform builds (arm64, amd64)
- Advanced cache management
- Better build output and progress display
- Future-proof build commands

**Evidence**: All current build commands use `docker build` instead of `docker buildx build`.

#### 9. Documentation Fragmentation

**Problem**: Docker documentation spread across 4 README files with some duplication:
- `docker/README.md` (comprehensive, 400+ lines)
- `packaging/docker/README.md` (brief, 50 lines)
- `packaging/deb/docker/README.md` (brief, 40 lines)
- Plus references in `docs/TEST_SETUP.md` and `docs/testing/docker-test-environment.md`

## Recommended Refactorings

### Priority 1: Critical Fixes

#### 1.1 Standardize Base Image References

**Change**: Use consistent versioned base image references.

```dockerfile
# In all derived Dockerfiles
ARG BASE_IMAGE=onemount-base:${ONEMOUNT_VERSION:-0.1.0rc1}
FROM ${BASE_IMAGE}
```

**Benefit**: Clear version tracking, reproducible builds.

#### 1.2 Create Unified Build Script

**Change**: Create `docker/scripts/build-images.sh` for building all images:

```bash
#!/bin/bash
# Build all OneMount Docker images with proper dependency order

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
VERSION="${ONEMOUNT_VERSION:-0.1.0rc1}"

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

print_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
print_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }

# Parse arguments
BUILD_TARGET="${1:-all}"
NO_CACHE="${2:-}"

CACHE_FLAG=""
if [[ "$NO_CACHE" == "--no-cache" ]]; then
    CACHE_FLAG="--no-cache"
fi

cd "$PROJECT_ROOT"

build_image() {
    local name=$1
    local dockerfile=$2
    local tag=$3
    
    print_info "Building $name..."
    docker buildx build $CACHE_FLAG \
        -f "$dockerfile" \
        -t "$tag:$VERSION" \
        -t "$tag:latest" \
        --build-arg ONEMOUNT_VERSION="$VERSION" \
        --load \
        .
    print_success "Built $name"
}

case "$BUILD_TARGET" in
    base)
        build_image "base image" "packaging/docker/Dockerfile" "onemount-base"
        ;;
    test-runner)
        build_image "base image" "packaging/docker/Dockerfile" "onemount-base"
        build_image "test runner" "docker/Dockerfile.test-runner" "onemount-test-runner"
        ;;
    github-runner)
        build_image "base image" "packaging/docker/Dockerfile" "onemount-base"
        build_image "GitHub runner" "docker/Dockerfile.github-runner" "onemount-github-runner"
        ;;
    deb-builder)
        build_image "base image" "packaging/docker/Dockerfile" "onemount-base"
        build_image "Debian builder" "packaging/deb/docker/Dockerfile" "onemount-deb-builder"
        ;;
    all)
        build_image "base image" "packaging/docker/Dockerfile" "onemount-base"
        build_image "test runner" "docker/Dockerfile.test-runner" "onemount-test-runner"
        build_image "GitHub runner" "docker/Dockerfile.github-runner" "onemount-github-runner"
        build_image "Debian builder" "packaging/deb/docker/Dockerfile" "onemount-deb-builder"
        ;;
    *)
        echo "Usage: $0 {base|test-runner|github-runner|deb-builder|all} [--no-cache]"
        exit 1
        ;;
esac

print_success "Build complete!"
docker images | grep onemount
```

**Usage**: 
```bash
# Build all images
./docker/scripts/build-images.sh all

# Build specific image
./docker/scripts/build-images.sh test-runner

# Build without cache
./docker/scripts/build-images.sh all --no-cache
```

**Benefit**: Single command to build all images, explicit dependencies, proper build tool usage.

#### 1.3 Standardize Volume Mounts for Auth Tokens

**Change**: Use consistent path across all containers:

```yaml
# Standard mount point for all containers
volumes:
  - ../../test-artifacts:/workspace/test-artifacts:rw
  
# Standard environment variable
environment:
  - ONEMOUNT_AUTH_TOKENS=/workspace/test-artifacts/.auth_tokens.json
```

**Update**: All entrypoint scripts to check this standard location first.

**Benefit**: Predictable behavior, easier documentation.

### Priority 2: Code Quality Improvements

#### 2.1 Extract Common Bash Functions

**Change**: Create `docker/scripts/common.sh` with shared functions:

```bash
#!/bin/bash
# Common functions for OneMount Docker entrypoint scripts

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Print functions
print_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
print_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
print_warning() { echo -e "${YELLOW}[WARNING]${NC} $1"; }
print_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# Environment validation
validate_workspace() {
    if [[ ! -f "go.mod" ]] || ! grep -q "github.com/auriora/onemount" go.mod; then
        print_error "Not in OneMount project directory"
        return 1
    fi
}

# Auth token setup
setup_auth_tokens() {
    # Unified auth token setup logic
    # ... (extracted from current scripts)
}
```

**Update**: All entrypoint scripts to source this file:

```bash
#!/bin/bash
source /usr/local/bin/common.sh
# ... rest of script
```

**Benefit**: DRY principle, easier maintenance, consistent behavior.

#### 2.2 Consolidate Environment Variables in Base Image

**Change**: Move all common environment setup to base Dockerfile:

```dockerfile
# In packaging/docker/Dockerfile
ENV PATH="/usr/local/go/bin:${PATH}"
ENV GOCACHE="/tmp/go-cache"
ENV GOMODCACHE="/tmp/go-mod-cache"
ENV GOSUMDB="sum.golang.org"
ENV DEBIAN_FRONTEND=noninteractive
```

**Remove**: These declarations from derived Dockerfiles.

**Benefit**: Single source of truth, less duplication.

#### 2.3 Use Multi-Stage Builds for Production Images

**Change**: Restructure Dockerfiles to separate build-time and runtime dependencies:

```dockerfile
# ============================================
# Builder stage - includes all build tools
# ============================================
FROM ubuntu:24.04 AS builder

# Install build dependencies only
RUN apt-get update && apt-get install -y \
    build-essential \
    pkg-config \
    git \
    wget \
    ca-certificates \
    libfuse3-dev \
    libgtk-3-dev \
    libwebkit2gtk-4.1-dev \
    && rm -rf /var/lib/apt/lists/*

# Install Go
ARG GO_VERSION=1.24.2
RUN wget -O go.tar.gz "https://golang.org/dl/go${GO_VERSION}.linux-amd64.tar.gz" && \
    tar -C /usr/local -xzf go.tar.gz && \
    rm go.tar.gz

ENV PATH="/usr/local/go/bin:${PATH}"
WORKDIR /build

# Copy source and build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_CFLAGS=-Wno-deprecated-declarations go build -v \
    -o /app/onemount \
    -ldflags="-X github.com/auriora/onemount/cmd/common.commit=$(git rev-parse HEAD)" \
    ./cmd/onemount

# ============================================
# Runtime stage - minimal production image
# ============================================
FROM ubuntu:24.04 AS runtime

# Install ONLY runtime dependencies (no build tools)
RUN apt-get update && apt-get install -y \
    fuse3 \
    libfuse3-3 \
    libgtk-3-0 \
    libwebkit2gtk-4.1-0 \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Configure FUSE
RUN echo 'user_allow_other' >> /etc/fuse.conf && \
    groupadd -f fuse

# Create non-root user
RUN useradd -m -s /bin/bash -G fuse onemount

# Copy ONLY the compiled binary from builder
COPY --from=builder /app/onemount /usr/local/bin/onemount

# No source code, no Go compiler, no build tools
USER onemount
WORKDIR /home/onemount

ENTRYPOINT ["/usr/local/bin/onemount"]
CMD ["--help"]

# Labels
LABEL org.opencontainers.image.title="OneMount"
LABEL org.opencontainers.image.description="Production OneMount image (runtime only)"
```

**Key Changes**:
- **Builder stage**: Contains all build tools, Go compiler, source code
- **Runtime stage**: Contains ONLY runtime libraries and compiled binary
- **No development packages** in production: No git, wget, build-essential, pkg-config
- **No source code** in production image
- **No Go compiler** in production image
- **Smaller image**: Target <500MB (down from 1.49GB)

**Benefit**: 
- 60-70% smaller production images
- Reduced attack surface (fewer packages = fewer vulnerabilities)
- Faster deployments
- Better security posture
- Clear separation of concerns

### Priority 3: Organizational Improvements

#### 3.1 Consolidate Docker Documentation

**Change**: Create single comprehensive Docker guide:

```
docs/docker/
├── README.md                    # Overview and quick start
├── images.md                    # Image architecture and building
├── testing.md                   # Testing with Docker
├── ci-cd.md                     # GitHub runners and CI/CD
├── production.md                # Production deployment
└── troubleshooting.md           # Common issues and solutions
```

**Update**: Other READMEs to be brief with links to main docs.

**Benefit**: Single source of truth, easier to maintain, better organization.

#### 3.2 Reorganize Build Scripts and Compose Files

**Change**: Separate image building from runtime orchestration:

```
docker/
├── scripts/
│   ├── build-images.sh          # Build Docker images (NEW)
│   ├── build-entrypoint.sh      # Build binaries in containers (KEEP)
│   ├── test-entrypoint.sh       # Test entrypoint (KEEP)
│   ├── runner-entrypoint.sh     # Runner entrypoint (KEEP)
│   └── common.sh                # Shared functions (NEW)
└── compose/
    ├── docker-compose.test.yml      # Run tests (KEEP)
    ├── docker-compose.build.yml     # Build binaries/packages (KEEP)
    ├── docker-compose.runners.yml   # GitHub runners (KEEP)
    └── docker-compose.dev.yml       # Development environment (NEW)
```

**Benefit**: Clear separation between image building (scripts) and runtime orchestration (compose).

#### 3.3 Create Development Compose File

**Change**: Add `docker-compose.dev.yml` for local development:

```yaml
name: onemount-dev

services:
  dev:
    build:
      context: ../..
      dockerfile: docker/Dockerfile.test-runner
    image: onemount-dev:latest
    volumes:
      - ../..:/workspace:rw
    devices:
      - /dev/fuse:/dev/fuse
    cap_add:
      - SYS_ADMIN
    command: ["shell"]
    stdin_open: true
    tty: true
```

**Usage**: `docker compose -f docker/compose/docker-compose.dev.yml run --rm dev`

**Benefit**: Quick development environment setup.

### Priority 4: Modern Build Tooling

#### 4.0 Migrate to Docker Buildx

**Change**: Update all build commands to use `docker buildx build`:

```bash
# In build-images.sh
docker buildx build \
    -f "$dockerfile" \
    -t "$tag:$VERSION" \
    -t "$tag:latest" \
    --build-arg ONEMOUNT_VERSION="$VERSION" \
    --load \
    .
```

**For multi-platform builds**:
```bash
docker buildx build \
    --platform linux/amd64,linux/arm64 \
    -f "$dockerfile" \
    -t "$tag:$VERSION" \
    --push \
    .
```

**Benefits**:
- Modern BuildKit features enabled by default
- Multi-platform support (amd64, arm64)
- Better cache management
- Improved build output
- Future-proof commands

**Migration**: Replace all `docker build` with `docker buildx build` in:
- `docker/scripts/build-images.sh`
- Documentation examples
- CI/CD workflows
- Makefile targets

### Priority 5: Performance Optimizations

#### 5.1 Optimize Layer Caching

**Change**: Reorder Dockerfile instructions for better caching:

```dockerfile
# 1. Install system dependencies (changes rarely)
RUN apt-get update && apt-get install -y ...

# 2. Install Go (changes rarely)
RUN wget go.tar.gz && tar ...

# 3. Copy go.mod/go.sum (changes occasionally)
COPY go.mod go.sum ./

# 4. Download dependencies (changes occasionally)
RUN go mod download

# 5. Copy source code (changes frequently)
COPY . .

# 6. Build (changes frequently)
RUN go build ...
```

**Current issue**: Some Dockerfiles copy source before downloading deps.

**Benefit**: Faster rebuilds, better cache utilization.

#### 5.2 Use .dockerignore More Effectively

**Change**: Create comprehensive `.dockerignore`:

```dockerignore
# Version control
.git/
.github/

# Build artifacts
build/
dist/
*.test
*.out

# Dependencies
.venv/
venv/
node_modules/

# Test artifacts
test-artifacts/
coverage/

# IDE
.idea/
.vscode/
*.swp

# OS
.DS_Store
Thumbs.db

# Docs (not needed in images)
docs/
*.md
!README.md
```

**Benefit**: Smaller build context, faster builds.

#### 5.3 Implement BuildKit Secrets

**Change**: Use BuildKit secrets for sensitive data:

```dockerfile
# In Dockerfile
RUN --mount=type=secret,id=github_token \
    GITHUB_TOKEN=$(cat /run/secrets/github_token) && \
    # ... use token
```

```bash
# In build command
docker build --secret id=github_token,src=token.txt .
```

**Benefit**: Secrets not stored in image layers.

### Priority 6: Testing and Validation

#### 6.1 Add Image Size Validation

**Change**: Create script to validate image sizes:

```bash
#!/bin/bash
# docker/scripts/validate-images.sh

MAX_BASE_SIZE="1.5GB"      # Development base with build tools
MAX_PRODUCTION_SIZE="500MB"  # Production runtime image (NEW)
MAX_TEST_SIZE="2.5GB"
MAX_RUNNER_SIZE="2.5GB"

check_size() {
    local image=$1
    local max_size=$2
    local actual_size=$(docker images --format "{{.Size}}" "$image")
    echo "Image $image: $actual_size (max: $max_size)"
}

check_size "onemount-base:latest" "$MAX_BASE_SIZE"
check_size "onemount:latest" "$MAX_PRODUCTION_SIZE"
check_size "onemount-test-runner:latest" "$MAX_TEST_SIZE"
check_size "onemount-github-runner:latest" "$MAX_RUNNER_SIZE"
```

**Benefit**: Catch image bloat early.

#### 6.2 Add Dockerfile Linting

**Change**: Add hadolint to CI/CD:

```yaml
# In .github/workflows/docker.yml
- name: Lint Dockerfiles
  uses: hadolint/hadolint-action@v3.1.0
  with:
    dockerfile: packaging/docker/Dockerfile
    
- name: Lint test Dockerfile
  uses: hadolint/hadolint-action@v3.1.0
  with:
    dockerfile: docker/Dockerfile.test-runner
```

**Benefit**: Catch Dockerfile issues early, enforce best practices.

## Implementation Plan

### Phase 1: Critical Fixes (Week 1)
1. Standardize base image references
2. Create unified build script (`build-images.sh`)
3. Standardize volume mounts

### Phase 2: Modern Tooling (Week 2)
1. Migrate to `docker buildx build`
2. Implement multi-stage builds for production
3. Separate runtime from build dependencies

### Phase 3: Code Quality (Week 3)
1. Extract common bash functions
2. Consolidate environment variables
3. Optimize layer caching

### Phase 4: Organization (Week 4)
1. Consolidate documentation
2. Reorganize build scripts and compose files
3. Create development compose file

### Phase 5: Validation (Week 5)
1. Add image size validation
2. Add Dockerfile linting
3. Add security scanning

## Migration Guide

### For Developers

**Before**:
```bash
# Build test image manually
docker build -f docker/Dockerfile.test-runner -t onemount-test-runner:latest .

# Run tests
docker compose -f docker/compose/docker-compose.test.yml run --rm unit-tests
```

**After**:
```bash
# Build all images using dedicated build script
./docker/scripts/build-images.sh all

# Or build specific image
./docker/scripts/build-images.sh test-runner

# Run tests (same)
docker compose -f docker/compose/docker-compose.test.yml run --rm unit-tests
```

### For CI/CD

**Before**:
```yaml
- name: Build test image
  run: docker build -f docker/Dockerfile.test-runner -t onemount-test-runner:latest .
```

**After**:
```yaml
- name: Build images
  run: ./docker/scripts/build-images.sh test-runner
```

## Risks and Mitigation

### Risk 1: Breaking Changes
**Mitigation**: Maintain backward compatibility for 1 release cycle, add deprecation warnings.

### Risk 2: Documentation Drift
**Mitigation**: Update all documentation in same PR, add validation checks.

### Risk 3: CI/CD Disruption
**Mitigation**: Test changes in feature branch CI before merging.

## Success Metrics

1. **Production image size**: Reduce from 1.49GB to <500MB (66% reduction)
2. **Build time reduction**: Target 30% faster builds with better caching
3. **Security posture**: Zero development packages in production images
4. **Code duplication**: Reduce bash duplication from ~100 lines to ~0 lines
5. **Documentation clarity**: Single source of truth for Docker setup
6. **Developer experience**: Single script to build all images with clear output
7. **Modern tooling**: 100% migration to `docker buildx build`
8. **Proper tool usage**: Docker Compose used for runtime orchestration, build scripts for image building

## Related Documentation

- Current: `docker/README.md`
- Current: `packaging/docker/README.md`
- Current: `docs/TEST_SETUP.md`
- Current: `docs/testing/docker-test-environment.md`
- Steering: `.kiro/steering/testing-conventions.md`

## Rules Consulted

- `testing-conventions.md` (Priority 25) - Docker test requirements
- `operational-best-practices.md` (Priority 40) - Tool usage and documentation
- `coding-standards.md` (Priority 100) - DRY principle, best practices
- `general-preferences.md` (Priority 50) - SOLID and DRY principles

## Rules Applied

- DRY principle: Extract common bash functions
- Documentation consistency: Consolidate Docker docs
- Security best practices: BuildKit secrets, non-root users
- Single source of truth: Centralize environment variables

## Directory Structure Analysis

### Current Structure

```
docker/                          # Development Docker files
├── compose/                     # Docker Compose files for dev/test/CI
│   ├── docker-compose.test.yml
│   ├── docker-compose.build.yml
│   └── docker-compose.runners.yml
├── scripts/                     # Container entrypoint scripts
│   ├── test-entrypoint.sh
│   ├── runner-entrypoint.sh
│   └── build-entrypoint.sh
├── Dockerfile.test-runner       # Test execution image
├── Dockerfile.github-runner     # CI/CD runner image
└── README.md

packaging/                       # Distribution packaging
├── docker/                      # Production Docker files
│   ├── Dockerfile               # Base image
│   ├── docker-compose.yml       # Production deployment
│   └── README.md
├── deb/                         # Debian packages
│   ├── docker/                  # Debian builder image
│   │   └── Dockerfile
│   └── control, rules, etc.
├── rpm/                         # RPM packages
│   └── onemount.spec
├── ubuntu/                      # Ubuntu-specific packages
└── install-manifest.json        # Central installation manifest
```

### Assessment: Does This Structure Make Sense?

**Short Answer**: Mostly yes, but with some confusion points.

#### What Works Well

1. **Clear Intent Separation**:
   - `docker/` = Development, testing, CI/CD
   - `packaging/` = Distribution and production deployment
   - This aligns with common practices

2. **Packaging Consolidation**:
   - All distribution formats (deb, rpm, docker) in one place
   - Centralized `install-manifest.json` is excellent
   - Reduces duplication across package types

3. **Compose Organization**:
   - Separate compose files for different purposes (test, build, runners)
   - Profile-based deployment options

#### What's Confusing

1. **Base Image Location** ⚠️
   - **Problem**: Base image is in `packaging/docker/Dockerfile` but used by dev images in `docker/`
   - **Confusion**: Development images depend on a file in the "packaging" directory
   - **Impact**: Not immediately obvious where to find the base image

2. **Docker Builder Image Nesting** ⚠️
   - **Problem**: `packaging/deb/docker/Dockerfile` - Docker image for building Debian packages
   - **Confusion**: Docker-in-packaging-in-docker creates deep nesting
   - **Impact**: Hard to discover, unclear relationship to other Docker files

3. **Duplicate docker-compose.yml** ⚠️
   - **Problem**: 
     - `docker/compose/docker-compose.build.yml` - Build binaries in containers
     - `packaging/docker/docker-compose.yml` - Production deployment
   - **Confusion**: Two compose files with different purposes but similar names
   - **Impact**: Users might use the wrong one

4. **Ubuntu vs Deb** ⚠️
   - **Problem**: Both `packaging/deb/` and `packaging/ubuntu/` exist
   - **Question**: Why separate? Ubuntu uses .deb packages
   - **Impact**: Unclear which to use for Ubuntu

### Recommended Structure Improvements

#### Option A: Consolidate All Docker Files (Recommended)

Move all Docker-related files to a unified location:

```
docker/
├── images/                          # All Dockerfiles
│   ├── base/
│   │   └── Dockerfile               # Base image (from packaging/)
│   ├── test-runner/
│   │   └── Dockerfile               # Test runner
│   ├── github-runner/
│   │   └── Dockerfile               # CI/CD runner
│   ├── deb-builder/
│   │   └── Dockerfile               # Debian package builder
│   └── production/
│       └── Dockerfile               # Production runtime (NEW - multi-stage)
├── compose/                         # Docker Compose files
│   ├── docker-compose.dev.yml       # Development environment
│   ├── docker-compose.test.yml      # Testing
│   ├── docker-compose.build.yml     # Build binaries/packages
│   ├── docker-compose.runners.yml   # GitHub runners
│   └── docker-compose.prod.yml      # Production deployment (from packaging/)
├── scripts/                         # Container scripts
│   ├── common.sh                    # Shared functions (NEW)
│   ├── build-images.sh              # Build all images (NEW)
│   ├── test-entrypoint.sh
│   ├── runner-entrypoint.sh
│   └── build-entrypoint.sh
└── README.md                        # Comprehensive Docker guide

packaging/
├── deb/                             # Debian packaging (no docker/)
│   ├── control, rules, changelog
│   └── source/
├── rpm/                             # RPM packaging
│   └── onemount.spec
├── install-manifest.json            # Central manifest
└── README.md                        # Packaging guide
```

**Benefits**:
- All Docker files in one place
- Clear image hierarchy in `docker/images/`
- No Docker files scattered in packaging/
- Easier to discover and understand

**Migration**:
- Move `packaging/docker/Dockerfile` → `docker/images/base/Dockerfile`
- Move `packaging/deb/docker/Dockerfile` → `docker/images/deb-builder/Dockerfile`
- Move `packaging/docker/docker-compose.yml` → `docker/compose/docker-compose.prod.yml`
- Update all references

#### Option B: Keep Separation, Improve Clarity (Conservative)

Keep current structure but improve naming and documentation:

```
docker/                              # Development & CI/CD Docker
├── dev/                             # Development images (RENAME)
│   ├── Dockerfile.test-runner
│   └── Dockerfile.github-runner
├── compose/
│   ├── docker-compose.dev.yml       # Development
│   ├── docker-compose.test.yml      # Testing
│   ├── docker-compose.build.yml     # Build binaries
│   └── docker-compose.runners.yml   # CI/CD runners
├── scripts/
└── README.md

packaging/
├── docker/                          # Production Docker (CLARIFY)
│   ├── base/                        # Base image (ORGANIZE)
│   │   └── Dockerfile
│   ├── runtime/                     # Production runtime (NEW)
│   │   └── Dockerfile
│   ├── docker-compose.prod.yml      # Production deployment (RENAME)
│   └── README.md
├── deb/
│   ├── builder/                     # Debian builder (RENAME from docker/)
│   │   └── Dockerfile
│   └── control, rules, etc.
├── rpm/
└── install-manifest.json

build/                               # Build system (EXISTING)
├── docker/                          # Docker build context (EXISTING)
└── ...
```

**Benefits**:
- Maintains current separation philosophy
- Clearer naming reduces confusion
- Less disruptive migration

**Drawbacks**:
- Still have Docker files in multiple locations
- Base image still in packaging/ but used by dev/

#### Option C: Hybrid Approach (Pragmatic)

Keep base image in packaging, consolidate derived images:

```
docker/
├── images/                          # Development images only
│   ├── test-runner/
│   │   └── Dockerfile
│   ├── github-runner/
│   │   └── Dockerfile
│   └── deb-builder/
│       └── Dockerfile               # From packaging/deb/docker/
├── compose/
├── scripts/
│   └── build-images.sh              # References packaging/docker/Dockerfile
└── README.md

packaging/
├── docker/                          # Production base & runtime
│   ├── Dockerfile.base              # Base image (RENAME)
│   ├── Dockerfile.runtime           # Production runtime (NEW)
│   ├── docker-compose.yml           # Production deployment
│   └── README.md
├── deb/                             # No docker/ subdirectory
├── rpm/
└── install-manifest.json
```

**Benefits**:
- Base image stays with production packaging (logical)
- Development images consolidated
- Minimal disruption
- Clear ownership

### Specific Issues to Address

#### 1. Ubuntu vs Deb Duplication

**Current**:
```
packaging/deb/        # Debian packages
packaging/ubuntu/     # Ubuntu packages (also .deb)
```

**Recommendation**: 
- If Ubuntu needs different control files, keep separate
- If identical, remove `packaging/ubuntu/` and document that `packaging/deb/` works for Ubuntu
- Add symlink if needed: `packaging/ubuntu -> deb`

**Investigate**: Check if there are actual differences between the two

#### 2. Build Directory Confusion

**Current**:
```
build/docker/         # What is this?
docker/               # Docker files
packaging/docker/     # More Docker files
```

**Recommendation**: 
- Document purpose of `build/docker/` in README
- If it's just build artifacts, add to .gitignore
- If it's build context, rename to `build/context/`

### Recommended Action Plan

**Phase 1: Quick Wins (No Breaking Changes)**
1. Add `docker/scripts/common.sh` for shared functions
2. Add `docker/scripts/build-images.sh` for unified builds
3. Improve README files with clear purpose statements
4. Document the relationship between directories

**Phase 2: Consolidation (Breaking Changes)**
1. Choose Option A, B, or C based on team preference
2. Move files according to chosen structure
3. Update all references in:
   - Documentation
   - CI/CD workflows
   - Makefile
   - Scripts
4. Add deprecation notices for old locations

**Phase 3: Cleanup**
1. Remove duplicate Ubuntu packaging if not needed
2. Clean up `build/docker/` or document its purpose
3. Archive old documentation

### Recommendation Summary

**For Your Project**: I recommend **Option C (Hybrid Approach)** because:

1. **Respects Current Philosophy**: Base image belongs with production packaging
2. **Minimal Disruption**: Least breaking changes
3. **Improves Clarity**: Consolidates dev images, clarifies production
4. **Pragmatic**: Balances ideal structure with migration cost

**Key Changes**:
- Move `packaging/deb/docker/` → `docker/images/deb-builder/`
- Rename `packaging/docker/Dockerfile` → `packaging/docker/Dockerfile.base`
- Add `packaging/docker/Dockerfile.runtime` (multi-stage production)
- Create `docker/scripts/build-images.sh` that knows about both locations
- Improve documentation to explain the separation

## Option C Implementation Details

### File Moves and Renames

#### 1. Reorganize Docker Images

**Move development images to unified location**:
```bash
# Create new structure
mkdir -p docker/images/test-runner
mkdir -p docker/images/github-runner
mkdir -p docker/images/deb-builder

# Move files
mv docker/Dockerfile.test-runner docker/images/test-runner/Dockerfile
mv docker/Dockerfile.github-runner docker/images/github-runner/Dockerfile
mv packaging/deb/docker/Dockerfile docker/images/deb-builder/Dockerfile
mv packaging/deb/docker/README.md docker/images/deb-builder/README.md

# Remove empty directory
rmdir packaging/deb/docker
```

**Rename production base image for clarity**:
```bash
cd packaging/docker
mv Dockerfile Dockerfile.base
```

**Create new production runtime image**:
```bash
# Create packaging/docker/Dockerfile.runtime (see multi-stage build example above)
```

#### 2. Update File References

**Files that need updates**:

1. **`docker/images/test-runner/Dockerfile`**:
   ```dockerfile
   # OLD:
   ARG BASE_IMAGE=onemount-base:0.1.0rc1
   FROM ${BASE_IMAGE}
   
   # NEW:
   ARG BASE_IMAGE=onemount-base:${ONEMOUNT_VERSION:-0.1.0rc1}
   FROM ${BASE_IMAGE}
   ```

2. **`docker/images/github-runner/Dockerfile`**:
   ```dockerfile
   # OLD:
   FROM onemount-base
   
   # NEW:
   ARG BASE_IMAGE=onemount-base:${ONEMOUNT_VERSION:-0.1.0rc1}
   FROM ${BASE_IMAGE}
   ```

3. **`docker/images/deb-builder/Dockerfile`**:
   ```dockerfile
   # OLD:
   FROM onemount-base
   
   # NEW:
   ARG BASE_IMAGE=onemount-base:${ONEMOUNT_VERSION:-0.1.0rc1}
   FROM ${BASE_IMAGE}
   ```

4. **`docker/compose/docker-compose.build.yml`**:
   ```yaml
   # OLD:
   build:
     context: ../..
     dockerfile: packaging/deb/docker/Dockerfile
   
   # NEW:
   build:
     context: ../..
     dockerfile: docker/images/deb-builder/Dockerfile
   ```

5. **`docker/compose/docker-compose.test.yml`**:
   ```yaml
   # Add build context (currently missing):
   test-runner:
     build:
       context: ../..
       dockerfile: docker/images/test-runner/Dockerfile
       args:
         - ONEMOUNT_VERSION=${ONEMOUNT_VERSION:-0.1.0rc1}
     image: onemount-test-runner:${ONEMOUNT_VERSION:-latest}
   ```

6. **`docker/compose/docker-compose.runners.yml`**:
   ```yaml
   # OLD:
   build:
     context: ../..
     dockerfile: docker/Dockerfile.github-runner
   
   # NEW:
   build:
     context: ../..
     dockerfile: docker/images/github-runner/Dockerfile
     args:
       - ONEMOUNT_VERSION=${ONEMOUNT_VERSION:-0.1.0rc1}
   ```

7. **`packaging/docker/docker-compose.yml`**:
   ```yaml
   # OLD:
   image: onemount:${ONEMOUNT_VERSION:-latest}
   
   # NEW:
   build:
     context: ../..
     dockerfile: packaging/docker/Dockerfile.runtime
     args:
       - ONEMOUNT_VERSION=${ONEMOUNT_VERSION:-0.1.0rc1}
   image: onemount:${ONEMOUNT_VERSION:-latest}
   ```

#### 3. Create New Files

**`docker/scripts/build-images.sh`**:
```bash
#!/bin/bash
# Build all OneMount Docker images with proper dependency order

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
VERSION="${ONEMOUNT_VERSION:-0.1.0rc1}"

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

print_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
print_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
print_warning() { echo -e "${YELLOW}[WARNING]${NC} $1"; }

# Parse arguments
BUILD_TARGET="${1:-all}"
NO_CACHE="${2:-}"

CACHE_FLAG=""
if [[ "$NO_CACHE" == "--no-cache" ]]; then
    CACHE_FLAG="--no-cache"
fi

cd "$PROJECT_ROOT"

build_image() {
    local name=$1
    local dockerfile=$2
    local tag=$3
    
    print_info "Building $name..."
    docker buildx build $CACHE_FLAG \
        -f "$dockerfile" \
        -t "$tag:$VERSION" \
        -t "$tag:latest" \
        --build-arg ONEMOUNT_VERSION="$VERSION" \
        --load \
        .
    print_success "Built $name ($tag:$VERSION)"
}

case "$BUILD_TARGET" in
    base)
        build_image "base image" "packaging/docker/Dockerfile.base" "onemount-base"
        ;;
    runtime|production)
        build_image "base image" "packaging/docker/Dockerfile.base" "onemount-base"
        build_image "production runtime" "packaging/docker/Dockerfile.runtime" "onemount"
        ;;
    test-runner)
        build_image "base image" "packaging/docker/Dockerfile.base" "onemount-base"
        build_image "test runner" "docker/images/test-runner/Dockerfile" "onemount-test-runner"
        ;;
    github-runner)
        build_image "base image" "packaging/docker/Dockerfile.base" "onemount-base"
        build_image "GitHub runner" "docker/images/github-runner/Dockerfile" "onemount-github-runner"
        ;;
    deb-builder)
        build_image "base image" "packaging/docker/Dockerfile.base" "onemount-base"
        build_image "Debian builder" "docker/images/deb-builder/Dockerfile" "onemount-deb-builder"
        ;;
    dev)
        build_image "base image" "packaging/docker/Dockerfile.base" "onemount-base"
        build_image "test runner" "docker/images/test-runner/Dockerfile" "onemount-test-runner"
        build_image "GitHub runner" "docker/images/github-runner/Dockerfile" "onemount-github-runner"
        build_image "Debian builder" "docker/images/deb-builder/Dockerfile" "onemount-deb-builder"
        ;;
    all)
        build_image "base image" "packaging/docker/Dockerfile.base" "onemount-base"
        build_image "production runtime" "packaging/docker/Dockerfile.runtime" "onemount"
        build_image "test runner" "docker/images/test-runner/Dockerfile" "onemount-test-runner"
        build_image "GitHub runner" "docker/images/github-runner/Dockerfile" "onemount-github-runner"
        build_image "Debian builder" "docker/images/deb-builder/Dockerfile" "onemount-deb-builder"
        ;;
    *)
        echo "Usage: $0 {base|runtime|test-runner|github-runner|deb-builder|dev|all} [--no-cache]"
        echo ""
        echo "Targets:"
        echo "  base          - Build base image only"
        echo "  runtime       - Build production runtime image (includes base)"
        echo "  test-runner   - Build test runner image (includes base)"
        echo "  github-runner - Build GitHub runner image (includes base)"
        echo "  deb-builder   - Build Debian builder image (includes base)"
        echo "  dev           - Build all development images (excludes production)"
        echo "  all           - Build all images"
        exit 1
        ;;
esac

print_success "Build complete!"
echo ""
docker images | grep onemount | head -20
```

**`docker/scripts/common.sh`**:
```bash
#!/bin/bash
# Common functions for OneMount Docker entrypoint scripts

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Print functions
print_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
print_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
print_warning() { echo -e "${YELLOW}[WARNING]${NC} $1"; }
print_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# Environment validation
validate_workspace() {
    if [[ ! -f "go.mod" ]] || ! grep -q "github.com/auriora/onemount" go.mod; then
        print_error "Not in OneMount project directory"
        return 1
    fi
    return 0
}

# Standard auth token locations
AUTH_TOKEN_LOCATIONS=(
    "/workspace/test-artifacts/.auth_tokens.json"
    "/workspace/test-artifacts/auth_tokens.json"
    "/workspace/auth_tokens.json"
    "$HOME/.onemount-tests/.auth_tokens.json"
    "/opt/onemount-ci/auth_tokens.json"
)

# Find auth tokens in standard locations
find_auth_tokens() {
    for location in "${AUTH_TOKEN_LOCATIONS[@]}"; do
        if [[ -f "$location" ]]; then
            echo "$location"
            return 0
        fi
    done
    return 1
}

# Setup auth tokens from standard locations
setup_auth_tokens() {
    local target_dir="${1:-$HOME/.onemount-tests}"
    local target_file="$target_dir/.auth_tokens.json"
    
    # Create target directory
    mkdir -p "$target_dir"
    
    # Check if already in place
    if [[ -f "$target_file" ]]; then
        print_info "Auth tokens already configured at $target_file"
        return 0
    fi
    
    # Find tokens
    local source_file
    if source_file=$(find_auth_tokens); then
        print_info "Found auth tokens at $source_file"
        cp "$source_file" "$target_file"
        chmod 600 "$target_file"
        print_success "Auth tokens configured"
        return 0
    fi
    
    # Check environment variable
    if [[ -n "$ONEMOUNT_AUTH_TOKENS" ]]; then
        print_info "Setting up auth tokens from environment variable"
        echo "$ONEMOUNT_AUTH_TOKENS" > "$target_file"
        chmod 600 "$target_file"
        print_success "Auth tokens configured from environment"
        return 0
    fi
    
    print_warning "No auth tokens found - system tests will be skipped"
    return 1
}
```

**`packaging/docker/Dockerfile.runtime`** (multi-stage production image):
```dockerfile
# ============================================
# Builder stage - includes all build tools
# ============================================
FROM ubuntu:24.04 AS builder

ARG ONEMOUNT_VERSION=0.1.0rc1
ARG GO_VERSION=1.24.2

# Install build dependencies only
RUN apt-get update && apt-get install -y \
    build-essential \
    pkg-config \
    git \
    wget \
    ca-certificates \
    libfuse3-dev \
    libgtk-3-dev \
    libwebkit2gtk-4.1-dev \
    && rm -rf /var/lib/apt/lists/*

# Install Go
RUN wget -O go.tar.gz "https://golang.org/dl/go${GO_VERSION}.linux-amd64.tar.gz" && \
    tar -C /usr/local -xzf go.tar.gz && \
    rm go.tar.gz

ENV PATH="/usr/local/go/bin:${PATH}"
WORKDIR /build

# Copy source and build
COPY go.mod go.sum ./
RUN go mod download

COPY cmd/ ./cmd/
COPY internal/ ./internal/
COPY scripts/cgo-helper.sh ./scripts/cgo-helper.sh

RUN bash scripts/cgo-helper.sh && \
    CGO_CFLAGS=-Wno-deprecated-declarations go build -v \
        -o /app/onemount \
        -ldflags="-X github.com/auriora/onemount/cmd/common.commit=$(git rev-parse HEAD 2>/dev/null || echo 'unknown')" \
        ./cmd/onemount && \
    CGO_CFLAGS=-Wno-deprecated-declarations go build -v \
        -o /app/onemount-launcher \
        -ldflags="-X github.com/auriora/onemount/cmd/common.commit=$(git rev-parse HEAD 2>/dev/null || echo 'unknown')" \
        ./cmd/onemount-launcher

# ============================================
# Runtime stage - minimal production image
# ============================================
FROM ubuntu:24.04 AS runtime

# Install ONLY runtime dependencies (no build tools)
RUN apt-get update && apt-get install -y \
    fuse3 \
    libfuse3-3 \
    libgtk-3-0 \
    libwebkit2gtk-4.1-0 \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Configure FUSE
RUN echo 'user_allow_other' >> /etc/fuse.conf && \
    groupadd -f fuse

# Create non-root user
RUN useradd -m -s /bin/bash -G fuse onemount

# Copy ONLY the compiled binaries from builder
COPY --from=builder /app/onemount /usr/local/bin/onemount
COPY --from=builder /app/onemount-launcher /usr/local/bin/onemount-launcher

# No source code, no Go compiler, no build tools
USER onemount
WORKDIR /home/onemount

ENTRYPOINT ["/usr/local/bin/onemount"]
CMD ["--help"]

# Labels
LABEL org.opencontainers.image.title="OneMount"
LABEL org.opencontainers.image.description="Production OneMount image (runtime only)"
LABEL org.opencontainers.image.vendor="Auriora"
LABEL org.opencontainers.image.source="https://github.com/Auriora/OneMount"
LABEL org.opencontainers.image.version="${ONEMOUNT_VERSION:-0.1.0rc1}"
```

#### 4. Update Documentation

**Files to update**:

1. **`docker/README.md`**:
   - Update directory structure section
   - Add explanation of image organization
   - Update build commands to use `build-images.sh`
   - Add section explaining relationship with `packaging/docker/`

2. **`packaging/docker/README.md`**:
   - Clarify this is for production deployment
   - Explain base image is used by development images
   - Add reference to `docker/` for development

3. **`docker/images/test-runner/README.md`** (new):
   ```markdown
   # Test Runner Image
   
   Docker image for running OneMount tests.
   
   **Base Image**: `onemount-base` (from `packaging/docker/Dockerfile.base`)
   
   ## Building
   
   ```bash
   ./docker/scripts/build-images.sh test-runner
   ```
   
   ## Usage
   
   See `docker/compose/docker-compose.test.yml`
   ```

4. **`docker/images/github-runner/README.md`** (new):
   ```markdown
   # GitHub Runner Image
   
   Docker image for GitHub Actions self-hosted runners.
   
   **Base Image**: `onemount-base` (from `packaging/docker/Dockerfile.base`)
   
   ## Building
   
   ```bash
   ./docker/scripts/build-images.sh github-runner
   ```
   
   ## Usage
   
   See `docker/compose/docker-compose.runners.yml`
   ```

5. **`docker/images/deb-builder/README.md`** (move and update):
   - Update paths in examples
   - Reference new location

#### 5. Update CI/CD Workflows

**`.github/workflows/*.yml`** files that reference Docker:

```yaml
# OLD:
- name: Build test image
  run: docker build -f docker/Dockerfile.test-runner -t onemount-test-runner:latest .

# NEW:
- name: Build test image
  run: ./docker/scripts/build-images.sh test-runner
```

#### 6. Update Makefile

If Makefile has Docker targets:

```makefile
# OLD:
docker-build:
	docker build -f docker/Dockerfile.test-runner -t onemount-test-runner:latest .

# NEW:
docker-build:
	./docker/scripts/build-images.sh all

docker-build-dev:
	./docker/scripts/build-images.sh dev

docker-build-prod:
	./docker/scripts/build-images.sh runtime
```

### Final Directory Structure (Option C)

```
docker/
├── images/                          # All development images
│   ├── test-runner/
│   │   ├── Dockerfile               # From docker/Dockerfile.test-runner
│   │   └── README.md                # NEW
│   ├── github-runner/
│   │   ├── Dockerfile               # From docker/Dockerfile.github-runner
│   │   └── README.md                # NEW
│   └── deb-builder/
│       ├── Dockerfile               # From packaging/deb/docker/Dockerfile
│       └── README.md                # From packaging/deb/docker/README.md
├── compose/
│   ├── docker-compose.test.yml      # UPDATED: new dockerfile path
│   ├── docker-compose.build.yml     # UPDATED: new dockerfile path
│   ├── docker-compose.runners.yml   # UPDATED: new dockerfile path
│   └── docker-compose.dev.yml       # NEW (optional)
├── scripts/
│   ├── common.sh                    # NEW: shared functions
│   ├── build-images.sh              # NEW: unified build script
│   ├── test-entrypoint.sh           # UPDATED: source common.sh
│   ├── runner-entrypoint.sh         # UPDATED: source common.sh
│   ├── build-entrypoint.sh          # UPDATED: source common.sh
│   ├── init-workspace.sh
│   ├── python-helper.sh
│   └── token-manager.sh
└── README.md                        # UPDATED: new structure

packaging/
├── docker/                          # Production Docker only
│   ├── Dockerfile.base              # RENAMED from Dockerfile
│   ├── Dockerfile.runtime           # NEW: multi-stage production
│   ├── docker-compose.yml           # UPDATED: use Dockerfile.runtime
│   ├── .dockerignore
│   └── README.md                    # UPDATED: clarify purpose
├── deb/                             # No docker/ subdirectory
│   ├── control, rules, changelog
│   └── source/
├── rpm/
│   └── onemount.spec
├── ubuntu/                          # Investigate if needed
│   └── ...
├── install-manifest.json
└── README.md

build/
├── docker/                          # Document or remove
└── ...
```

### Migration Checklist

- [ ] Create new directory structure
- [ ] Move Dockerfiles to new locations
- [ ] Rename `packaging/docker/Dockerfile` to `Dockerfile.base`
- [ ] Create `packaging/docker/Dockerfile.runtime`
- [ ] Create `docker/scripts/build-images.sh`
- [ ] Create `docker/scripts/common.sh`
- [ ] Update all Dockerfiles with consistent base image references
- [ ] Update all compose files with new dockerfile paths
- [ ] Update entrypoint scripts to source `common.sh`
- [ ] Create README files for each image directory
- [ ] Update main Docker README
- [ ] Update packaging Docker README
- [ ] Update CI/CD workflows
- [ ] Update Makefile (if applicable)
- [ ] Test all build commands
- [ ] Test all compose files
- [ ] Update documentation references
- [ ] Commit changes with clear message

### Testing Plan

After migration:

```bash
# Test base image build
./docker/scripts/build-images.sh base

# Test production runtime build
./docker/scripts/build-images.sh runtime

# Test development images
./docker/scripts/build-images.sh dev

# Test all images
./docker/scripts/build-images.sh all

# Test compose files
docker compose -f docker/compose/docker-compose.test.yml run --rm unit-tests
docker compose -f docker/compose/docker-compose.build.yml --profile binaries run --rm build-binaries
docker compose -f packaging/docker/docker-compose.yml config

# Verify image sizes
docker images | grep onemount
```

## Conclusion

The current Docker setup is functional but has room for improvement. The recommended refactorings will:

1. **Dramatically reduce production image size** from 1.49GB to <500MB through multi-stage builds
2. **Improve security** by removing development tools and source code from production images
3. **Modernize build tooling** by migrating to `docker buildx build`
4. **Reduce complexity** through consolidation and standardization
5. **Improve maintainability** by eliminating duplication
6. **Enhance developer experience** with clearer commands and documentation
7. **Optimize performance** through better caching and layer management
8. **Increase reliability** through validation and linting

**Critical Priority**: Multi-stage builds for production images should be implemented first to address the security and size concerns.

Implementation should be phased to minimize disruption while delivering incremental value.
