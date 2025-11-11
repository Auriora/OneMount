# Docker Duplication Analysis and Recommendations

**Date**: 2025-11-11  
**Author**: AI Agent  
**Status**: Analysis Complete  

## Implementation Status

### ✅ Completed from Recommendations

1. **Directory Restructure** - ✅ Complete
2. **Unified Build Script** - ✅ Complete (`build-images.sh`)
3. **Common Bash Functions** - ✅ Complete (`common.sh`)
4. **Consistent Base Image References** - ✅ Complete
5. **Updated Compose Files** - ✅ Complete
6. **Documentation Updates** - ✅ Complete
7. **Production Multi-Stage Build** - ✅ Complete

### ❌ Not Yet Implemented

1. **Entrypoint Scripts Using common.sh** - Not implemented
2. **Standardized Volume Mounts** - Partially implemented
3. **Multi-stage builds for dev images** - Not implemented
4. **.dockerignore optimization** - Not reviewed
5. **BuildKit secrets** - Not implemented
6. **Image size validation** - Not implemented
7. **Dockerfile linting** - Not implemented

## Duplication Analysis

### Critical Duplication Found

#### 1. Builder Stage Duplicated (MAJOR)

**Problem**: The builder stage in `packaging/docker/Dockerfile` is nearly identical to `packaging/docker/Dockerfile.builder`

**Duplication**:
```dockerfile
# packaging/docker/Dockerfile (lines 8-32)
FROM ubuntu:24.04 AS builder
# ... 25 lines of setup ...

# packaging/docker/Dockerfile.builder (lines 1-50)
FROM ubuntu:24.04
# ... 50 lines of nearly identical setup ...
```

**Overlap**: ~80% identical
- Same Ubuntu base
- Same IPv4 config
- Same apt packages (build-essential, pkg-config, git, wget, etc.)
- Same Go installation
- Same FUSE configuration

**Impact**: 
- Maintenance burden (update in two places)
- Inconsistency risk
- Larger total image size

#### 2. Go Environment Setup Duplicated (MODERATE)

**Found in ALL derived images**:
- `docker/images/test-runner/Dockerfile` (lines 28-30)
- `docker/images/github-runner/Dockerfile` (lines 36-38)
- `docker/images/deb-builder/Dockerfile` (lines 24-26)

```dockerfile
# Repeated in 3 files
ENV GOPATH="/home/USER/go"
ENV GOCACHE="/tmp/go-cache"  # or /home/USER/.cache/go-build
ENV GOMODCACHE="/tmp/go-mod-cache"  # or /home/USER/go/pkg/mod
```

**Impact**: Minor but unnecessary

#### 3. User Creation Pattern Duplicated (MODERATE)

**Found in 3 derived images**:
```dockerfile
# test-runner
RUN useradd -m -s /bin/bash -G fuse tester

# github-runner  
RUN useradd -m -s /bin/bash -G fuse,sudo runner

# deb-builder
RUN useradd -m -s /bin/bash builder
```

**Pattern**: Similar but with variations

#### 4. Go Module Download Duplicated (MINOR)

**Found in 2 images**:
- `docker/images/test-runner/Dockerfile` (lines 54-57)
- `docker/images/deb-builder/Dockerfile` (lines 47-50)

```dockerfile
COPY --chown=USER:USER go.mod go.sum ./
RUN --mount=type=cache,target=/PATH,uid=1000 \
    go mod download
```

#### 5. Labels Duplicated (MINOR)

All 5 Dockerfiles have nearly identical labels with only title/description differences.

## Recommended Approach to Fix Duplication

### Strategy: Use Multi-Stage Builds with Shared Base

Instead of having `Dockerfile.builder` as a separate image, use it as a shared base stage that all images can reference.

### Proposed Structure

```
packaging/docker/
├── Dockerfile           # Production (uses builder stage internally)
├── Dockerfile.base      # Shared base stage (NEW - replaces Dockerfile.builder)
└── docker-compose.yml

docker/images/
├── test-runner/
│   └── Dockerfile       # FROM packaging/docker/Dockerfile.base AS builder
├── github-runner/
│   └── Dockerfile       # FROM packaging/docker/Dockerfile.base AS builder
└── deb-builder/
    └── Dockerfile       # FROM packaging/docker/Dockerfile.base AS builder
```

### Implementation Plan

#### Phase 1: Create Shared Base Stage

**Create `packaging/docker/Dockerfile.base`** (rename current Dockerfile.builder):

```dockerfile
# syntax=docker/dockerfile:1
# Shared base stage for all OneMount Docker images
# This is not meant to be built directly, but used as a base stage

FROM ubuntu:24.04 AS base

ARG ONEMOUNT_VERSION=0.1.0rc1
ARG GO_VERSION=1.24.2

ENV DEBIAN_FRONTEND=noninteractive
ENV UBUNTU_VERSION=24.04
ENV UBUNTU_CODENAME=noble

# Configure IPv4-only networking
RUN echo 'Acquire::ForceIPv4 "true";' > /etc/apt/apt.conf.d/99force-ipv4

# Configure apt for better reliability and caching
RUN echo 'Acquire::Retries "3";' > /etc/apt/apt.conf.d/80retries && \
    echo 'Acquire::http::Timeout "30";' >> /etc/apt/apt.conf.d/80retries && \
    echo 'Acquire::ftp::Timeout "30";' >> /etc/apt/apt.conf.d/80retries && \
    echo 'APT::Keep-Downloaded-Packages "true";' >> /etc/apt/apt.conf.d/80retries

# Install essential system dependencies
RUN apt-get update && apt-get install -y \
    build-essential \
    pkg-config \
    git \
    wget \
    curl \
    ca-certificates \
    fuse3 \
    libfuse3-dev \
    libgtk-3-dev \
    libwebkit2gtk-4.1-dev \
    rsync \
    lsb-release \
    gnupg \
    && rm -rf /var/lib/apt/lists/*

# Install Go
RUN wget -O go.tar.gz "https://golang.org/dl/go${GO_VERSION}.linux-amd64.tar.gz" && \
    tar -C /usr/local -xzf go.tar.gz && \
    rm go.tar.gz && \
    /usr/local/go/bin/go version

# Set Go environment
ENV PATH="/usr/local/go/bin:${PATH}"
ENV GOCACHE="/tmp/go-cache"
ENV GOMODCACHE="/tmp/go-mod-cache"
ENV GOSUMDB="sum.golang.org"

# Configure FUSE
RUN echo 'user_allow_other' >> /etc/fuse.conf && \
    groupadd -f fuse

# Create cache directories
RUN mkdir -p /tmp/go-cache /tmp/go-mod-cache /workspace && \
    chmod 777 /tmp/go-cache /tmp/go-mod-cache

WORKDIR /workspace
```

#### Phase 2: Update Production Dockerfile

**Update `packaging/docker/Dockerfile`**:

```dockerfile
# syntax=docker/dockerfile:1
# Production Dockerfile for OneMount
# Multi-stage build using shared base

# Import shared base stage
FROM packaging/docker/Dockerfile.base:latest AS builder

ARG ONEMOUNT_VERSION=0.1.0rc1

# Copy source and build (no need to repeat base setup)
COPY go.mod go.sum ./
RUN go mod download

COPY cmd/ ./cmd/
COPY internal/ ./internal/
COPY scripts/cgo-helper.sh ./scripts/cgo-helper.sh

RUN bash scripts/cgo-helper.sh && \
    mkdir -p /app && \
    CGO_CFLAGS=-Wno-deprecated-declarations go build -v \
        -o /app/onemount \
        -ldflags="-X github.com/auriora/onemount/cmd/common.commit=$(git rev-parse HEAD 2>/dev/null || echo 'unknown')" \
        ./cmd/onemount && \
    CGO_CFLAGS=-Wno-deprecated-declarations go build -v \
        -o /app/onemount-launcher \
        -ldflags="-X github.com/auriora/onemount/cmd/common.commit=$(git rev-parse HEAD 2>/dev/null || echo 'unknown')" \
        ./cmd/onemount-launcher

# Runtime stage (unchanged)
FROM ubuntu:24.04 AS runtime
# ... rest of runtime stage ...
```

**Problem with this approach**: Docker doesn't support `FROM path/to/Dockerfile:stage` syntax.

### Better Approach: Use COPY --from with Context

Actually, Docker **doesn't support** referencing stages from other Dockerfiles directly. We need a different strategy.

### RECOMMENDED SOLUTION: Consolidate with Build Args

Keep the current structure but eliminate duplication through:

1. **Use the builder image as base** (current approach is correct)
2. **Extract common patterns to scripts**
3. **Use build args for variations**

#### Specific Recommendations:

### 1. Keep Current Structure (It's Actually Good)

The current approach of having:
- `Dockerfile.builder` as a standalone base image
- Derived images using `FROM onemount-builder`
- Production using multi-stage build

**This is the RIGHT approach** because:
- ✅ Builder image is built once, reused by all dev images
- ✅ Production is self-contained (doesn't depend on builder image existing)
- ✅ Clear separation of concerns

### 2. Eliminate Minor Duplications

#### A. Create User Creation Script

**Create `docker/scripts/create-user.sh`**:
```bash
#!/bin/bash
# Create user with optional groups
USERNAME=$1
GROUPS=${2:-}

useradd -m -s /bin/bash ${GROUPS:+-G $GROUPS} $USERNAME
mkdir -p /home/$USERNAME/go
chown -R $USERNAME:$USERNAME /home/$USERNAME
```

**Use in Dockerfiles**:
```dockerfile
COPY docker/scripts/create-user.sh /tmp/create-user.sh
RUN /tmp/create-user.sh tester fuse && rm /tmp/create-user.sh
```

#### B. Consolidate Go Environment Setup

**In builder image, set defaults**:
```dockerfile
# In Dockerfile.builder
ENV GO_USER_TEMPLATE="/home/USER/go"
ENV GO_CACHE_TEMPLATE="/tmp/go-cache"
```

**In derived images, just override USER**:
```dockerfile
ENV GOPATH="/home/tester/go"
# Inherit GOCACHE and GOMODCACHE from base
```

#### C. Extract Common Labels to Script

**Create `docker/scripts/add-labels.sh`**:
```bash
#!/bin/bash
# Generates Dockerfile labels
cat << EOF
LABEL org.opencontainers.image.vendor="Auriora"
LABEL org.opencontainers.image.source="https://github.com/Auriora/OneMount"
LABEL org.opencontainers.image.version="\${ONEMOUNT_VERSION:-0.1.0rc1}"
EOF
```

### 3. The REAL Issue: Production Dockerfile Duplication

**Current Problem**: 
- `packaging/docker/Dockerfile` builder stage (lines 8-55)
- `packaging/docker/Dockerfile.builder` (lines 1-85)
- ~70% overlap

**SOLUTION**: Make production Dockerfile use builder image

**Update `packaging/docker/Dockerfile`**:

```dockerfile
# syntax=docker/dockerfile:1
# Production Dockerfile for OneMount
# Multi-stage build: Uses builder image, creates minimal runtime

ARG ONEMOUNT_VERSION=0.1.0rc1
ARG BUILDER_IMAGE=onemount-builder:${ONEMOUNT_VERSION}

# ============================================
# Builder stage - use existing builder image
# ============================================
FROM ${BUILDER_IMAGE} AS builder

# Copy source and build
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download

COPY cmd/ ./cmd/
COPY internal/ ./internal/
COPY scripts/cgo-helper.sh ./scripts/cgo-helper.sh

RUN bash scripts/cgo-helper.sh && \
    mkdir -p /app && \
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

ARG ONEMOUNT_VERSION=0.1.0rc1

ENV DEBIAN_FRONTEND=noninteractive

# Configure IPv4-only networking
RUN echo 'Acquire::ForceIPv4 "true";' > /etc/apt/apt.conf.d/99force-ipv4

# Install ONLY runtime dependencies
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

USER onemount
WORKDIR /home/onemount

ENTRYPOINT ["/usr/local/bin/onemount"]
CMD ["--help"]

LABEL org.opencontainers.image.title="OneMount"
LABEL org.opencontainers.image.description="Production OneMount image (runtime only)"
LABEL org.opencontainers.image.vendor="Auriora"
LABEL org.opencontainers.image.source="https://github.com/Auriora/OneMount"
LABEL org.opencontainers.image.version="${ONEMOUNT_VERSION:-0.1.0rc1}"
```

**Benefits**:
- ✅ Eliminates 55 lines of duplication
- ✅ Production build uses same base as development
- ✅ Consistency guaranteed
- ✅ Single source of truth for builder setup

**Trade-off**:
- ⚠️ Production build requires builder image to exist first
- ⚠️ Can't build production standalone

**Mitigation**:
- Update build script to build builder first
- Document the dependency

## Summary of Recommendations

### Priority 1: Eliminate Major Duplication (HIGH IMPACT)

**Action**: Make production Dockerfile use builder image as base

**Files to change**:
1. `packaging/docker/Dockerfile` - Use `FROM onemount-builder` in builder stage
2. `docker/scripts/build-images.sh` - Ensure builder is built before production

**Impact**: Eliminates 55 lines of duplication, ensures consistency

### Priority 2: Extract Common Patterns (MEDIUM IMPACT)

**Actions**:
1. Create `docker/scripts/create-user.sh` for user creation
2. Simplify Go environment variables in derived images
3. Consider label generation script

**Impact**: Reduces 10-15 lines of duplication per image

### Priority 3: Update Entrypoint Scripts (LOW IMPACT)

**Action**: Make entrypoint scripts source `common.sh`

**Files to change**:
- `docker/scripts/test-entrypoint.sh`
- `docker/scripts/runner-entrypoint.sh`
- `docker/scripts/build-entrypoint.sh`

**Impact**: Eliminates ~100 lines of bash duplication

## Conclusion

The current structure is actually quite good. The main issue is the duplication between:
- Production Dockerfile builder stage
- Dockerfile.builder

**Recommended fix**: Make production Dockerfile use the builder image, eliminating the duplicated builder stage.

This is a simple, high-impact change that maintains the current architecture while eliminating the largest source of duplication.
