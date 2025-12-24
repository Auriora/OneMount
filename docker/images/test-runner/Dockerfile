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
    # JSON processing for auth tokens
    jq \
    # GUI and webkit dependencies for authentication
    libwebkit2gtk-4.1-dev \
    libgtk-3-dev \
    libgtk-3-0 \
    pkg-config \
    dbus \
    dbus-user-session \
    # X11 and display dependencies
    libx11-6 \
    libx11-xcb1 \
    libxcomposite1 \
    libxcursor1 \
    libxdamage1 \
    libxext6 \
    libxfixes3 \
    libxi6 \
    libxinerama1 \
    libxrandr2 \
    libxrender1 \
    libxss1 \
    libxtst6 \
    xauth \
    x11-apps \
    xdg-utils \
    # Graphics and Mesa drivers for WebKit
    mesa-utils \
    libegl1-mesa-dev \
    libgl1-mesa-dev \
    libglu1-mesa \
    libgl1-mesa-dri \
    # Additional GUI runtime dependencies (matching devcontainer)
    libasound2t64 \
    libatk-bridge2.0-0t64 \
    libatk1.0-0t64 \
    libcairo2 \
    libdbus-1-3 \
    libfontconfig1 \
    libfreetype6 \
    libgdk-pixbuf-2.0-0 \
    libglib2.0-0t64 \
    libharfbuzz0b \
    libnss3 \
    libpango-1.0-0 \
    libpangocairo-1.0-0 \
    libsecret-1-0 \
    # Additional packages from devcontainer for full GUI support
    libgl1 \
    libglu1-mesa \
    libxkbcommon0 \
    libxkbcommon-x11-0 \
    libxv1 \
    libxcb1 \
    libxcb-render0 \
    libxcb-render-util0 \
    libxcb-shape0 \
    libxcb-shm0 \
    libxcb-util1 \
    libxcb-xfixes0 \
    libxcb-icccm4 \
    libxcb-image0 \
    libxcb-keysyms1 \
    libxcb-xkb1 \
    libxslt1.1 \
    # Cleanup
    && rm -rf /var/lib/apt/lists/*

# Set Go environment for test user
# Note: GOCACHE and GOMODCACHE are set dynamically in entrypoint to avoid permission issues
ENV GOPATH="/home/tester/go" \
    GIO_USE_PORTALS=1 \
    GTK_USE_PORTAL=1

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

# Download dependencies (cache mounts already set up in builder image)
RUN go mod download

# Copy only required source to maximize cache hits
COPY --chown=tester:tester cmd/ ./cmd
COPY --chown=tester:tester internal/ ./internal
COPY --chown=tester:tester scripts/cgo-helper.sh ./scripts/cgo-helper.sh

# Create test directories
RUN mkdir -p /home/tester/.onemount-tests/tmp && \
    mkdir -p /home/tester/.onemount-tests/logs && \
    mkdir -p /home/tester/.cache/onemount

# Switch to root to copy scripts
USER root

# Copy common functions and entrypoint scripts
COPY docker/scripts/common.sh /usr/local/bin/common.sh
COPY docker/scripts/test-entrypoint.sh /usr/local/bin/test-entrypoint.sh
COPY docker/scripts/python-helper.sh /usr/local/bin/python-helper.sh
COPY docker/scripts/host-open.sh /usr/local/bin/host-open

RUN chmod +x /usr/local/bin/common.sh && \
    chmod +x /usr/local/bin/test-entrypoint.sh && \
    chmod +x /usr/local/bin/python-helper.sh && \
    chmod +x /usr/local/bin/host-open

# Switch back to tester user
USER tester

# Pre-build OneMount binaries for faster test execution
# First run CGO helper with CGO enabled to update webkit version in source
RUN CGO_ENABLED=1 bash scripts/cgo-helper.sh && \
    mkdir -p build/binaries && \
    CGO_ENABLED=0 go build -v \
        -o build/binaries/onemount \
        -ldflags="-X github.com/auriora/onemount/cmd/common.commit=$(git rev-parse HEAD 2>/dev/null || echo 'unknown')" \
        ./cmd/onemount && \
    echo "Built onemount binary without CGO/GUI support for testing"

# Default command
ENTRYPOINT ["/usr/local/bin/test-entrypoint.sh"]
CMD ["help"]

# Add labels for better image management
LABEL org.opencontainers.image.title="OneMount Test Runner (Development)"
LABEL org.opencontainers.image.description="Development Docker image for running OneMount tests (extends onemount-base)"
LABEL org.opencontainers.image.vendor="Auriora"
LABEL org.opencontainers.image.source="https://github.com/Auriora/OneMount"
LABEL org.opencontainers.image.version="${ONEMOUNT_VERSION:-0.1.0rc1}"
