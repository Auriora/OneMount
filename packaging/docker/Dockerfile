# Production Dockerfile for OneMount
# Multi-stage build: Uses builder image, creates minimal runtime
# This eliminates duplication by reusing the builder image

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

# Install ONLY runtime dependencies (no build tools, no git, no wget)
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

# Copy ONLY the compiled binaries from builder (no source code, no Go, no build tools)
COPY --from=builder /app/onemount /usr/local/bin/onemount
COPY --from=builder /app/onemount-launcher /usr/local/bin/onemount-launcher

# Switch to non-root user
USER onemount
WORKDIR /home/onemount

ENTRYPOINT ["/usr/local/bin/onemount"]
CMD ["--help"]

# Labels
LABEL org.opencontainers.image.title="OneMount"
LABEL org.opencontainers.image.description="Production OneMount image (runtime only, no build tools)"
LABEL org.opencontainers.image.vendor="Auriora"
LABEL org.opencontainers.image.source="https://github.com/Auriora/OneMount"
LABEL org.opencontainers.image.version="${ONEMOUNT_VERSION:-0.1.0rc1}"
