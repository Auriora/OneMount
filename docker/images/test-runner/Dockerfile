# Development Dockerfile for running OneMount tests
# Extends the production base image with development-specific configurations
# For production test runner, see: packaging/docker/Dockerfile.test-runner

ARG ONEMOUNT_VERSION=0.1.0rc1
ARG BASE_IMAGE=onemount-builder:${ONEMOUNT_VERSION}
FROM ${BASE_IMAGE}

# Define version arg for labels
ARG ONEMOUNT_VERSION=0.1.0rc1

# Install additional dependencies for testing
RUN apt-get update && apt-get install -y \
    # Python for test scripts
    python3 \
    python3-pip \
    python3-venv \
    python3.12-venv \
    # Network tools for testing
    iputils-ping \
    netcat-openbsd \
    # Process management
    psmisc \
    # Development tools
    vim \
    less \
    # Cleanup
    && rm -rf /var/lib/apt/lists/*

# Set Go environment for test user
ENV GOPATH="/home/tester/go"
ENV GOCACHE="/tmp/go-cache"
ENV GOMODCACHE="/tmp/go-mod-cache"

# Install Python dependencies for OneMount development CLI
COPY scripts/requirements-dev-cli.txt /tmp/requirements-dev-cli.txt
RUN pip3 install --no-cache-dir --break-system-packages -r /tmp/requirements-dev-cli.txt && \
    rm /tmp/requirements-dev-cli.txt

# Create test user and add to existing FUSE group
RUN useradd -m -s /bin/bash -G fuse tester && \
    mkdir -p /workspace && \
    chown tester:tester /workspace

# Set up Go environment for test user
USER tester
RUN mkdir -p /home/tester/go && \
    echo 'export GOPATH=/home/tester/go' >> /home/tester/.bashrc && \
    echo 'export PATH=/usr/local/go/bin:$GOPATH/bin:$PATH' >> /home/tester/.bashrc

# Set working directory
WORKDIR /workspace

# Copy go.mod and go.sum first for better caching
COPY --chown=tester:tester go.mod go.sum ./

# Download dependencies with BuildKit cache mount
RUN --mount=type=cache,target=/tmp/go-mod-cache,uid=1000 \
    --mount=type=cache,target=/tmp/go-cache,uid=1000 \
    go mod download

# Copy only required source to maximize cache hits
COPY --chown=tester:tester cmd/ ./cmd
COPY --chown=tester:tester internal/ ./internal
COPY --chown=tester:tester scripts/cgo-helper.sh ./scripts/cgo-helper.sh

# Create test directories
RUN mkdir -p /home/tester/.onemount-tests/tmp && \
    mkdir -p /home/tester/.onemount-tests/logs && \
    mkdir -p /home/tester/.cache/onemount

# Copy common functions and entrypoint scripts
COPY docker/scripts/common.sh /usr/local/bin/common.sh
COPY docker/scripts/test-entrypoint.sh /usr/local/bin/test-entrypoint.sh
COPY docker/scripts/python-helper.sh /usr/local/bin/python-helper.sh
USER root
RUN chmod +x /usr/local/bin/common.sh && \
    chmod +x /usr/local/bin/test-entrypoint.sh && \
    chmod +x /usr/local/bin/python-helper.sh
USER tester

# Pre-build OneMount binaries for faster test execution with BuildKit cache
RUN --mount=type=cache,target=/tmp/go-mod-cache,uid=1000 \
    --mount=type=cache,target=/tmp/go-cache,uid=1000 \
    bash scripts/cgo-helper.sh && \
    mkdir -p build/binaries && \
    CGO_CFLAGS=-Wno-deprecated-declarations go build -v \
        -o build/binaries/onemount \
        -ldflags="-X github.com/auriora/onemount/cmd/common.commit=$(git rev-parse HEAD 2>/dev/null || echo 'unknown')" \
        ./cmd/onemount && \
    CGO_CFLAGS=-Wno-deprecated-declarations go build -v \
        -o build/binaries/onemount-launcher \
        -ldflags="-X github.com/auriora/onemount/cmd/common.commit=$(git rev-parse HEAD 2>/dev/null || echo 'unknown')" \
        ./cmd/onemount-launcher

# Default command
ENTRYPOINT ["/usr/local/bin/test-entrypoint.sh"]
CMD ["help"]

# Add labels for better image management
LABEL org.opencontainers.image.title="OneMount Test Runner (Development)"
LABEL org.opencontainers.image.description="Development Docker image for running OneMount tests (extends onemount-base)"
LABEL org.opencontainers.image.vendor="Auriora"
LABEL org.opencontainers.image.source="https://github.com/Auriora/OneMount"
LABEL org.opencontainers.image.version="${ONEMOUNT_VERSION:-0.1.0rc1}"
