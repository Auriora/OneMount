# Development Dockerfile for OneMount GitHub Actions Self-Hosted Runner
# Extends the production base image with development-specific configurations
# For production runner, see: packaging/docker/Dockerfile.github-runner

ARG ONEMOUNT_VERSION=0.1.0rc1
ARG BASE_IMAGE=onemount-builder:${ONEMOUNT_VERSION}
FROM ${BASE_IMAGE}

# Set GitHub Actions specific environment variables
ENV RUNNER_ALLOW_RUNASROOT=1

# Install GitHub Actions runner specific dependencies
RUN apt-get update && apt-get install -y \
    # GitHub Actions runner dependencies
    curl \
    wget \
    unzip \
    tar \
    sudo \
    jq \
    # Development tools
    vim \
    less \
    && rm -rf /var/lib/apt/lists/*

# Install Docker CLI for elastic manager
RUN curl -fsSL https://download.docker.com/linux/ubuntu/gpg | gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg && \
    echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" | tee /etc/apt/sources.list.d/docker.list > /dev/null && \
    apt-get update && \
    apt-get install -y docker-ce-cli && \
    rm -rf /var/lib/apt/lists/*

# Set Go environment for runner user
ENV GOPATH="/home/runner/go"
ENV GOCACHE="/tmp/go-cache"
ENV GOMODCACHE="/home/runner/go/pkg/mod"
ENV GOSUMDB="sum.golang.org"

# Create runner user with sudo access
RUN useradd -m -s /bin/bash -G fuse,sudo runner && \
    echo "runner ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers && \
    mkdir -p /workspace && \
    chown runner:runner /workspace

# Create GitHub Actions runner directory
RUN mkdir -p /opt/actions-runner && \
    chown runner:runner /opt/actions-runner

# Switch to runner user
USER runner
WORKDIR /opt/actions-runner

# Download and install GitHub Actions runner
ARG RUNNER_VERSION=2.311.0
RUN curl -o actions-runner-linux-x64-${RUNNER_VERSION}.tar.gz \
    -L https://github.com/actions/runner/releases/download/v${RUNNER_VERSION}/actions-runner-linux-x64-${RUNNER_VERSION}.tar.gz && \
    tar xzf actions-runner-linux-x64-${RUNNER_VERSION}.tar.gz && \
    rm actions-runner-linux-x64-${RUNNER_VERSION}.tar.gz

# Set up Go environment for runner user
RUN mkdir -p /home/runner/go && \
    echo 'export GOPATH=/home/runner/go' >> /home/runner/.bashrc && \
    echo 'export PATH=/usr/local/go/bin:$GOPATH/bin:$PATH' >> /home/runner/.bashrc

# Create test directories and credentials location
RUN mkdir -p /home/runner/.onemount-tests/tmp && \
    mkdir -p /home/runner/.onemount-tests/logs && \
    mkdir -p /home/runner/.cache/onemount && \
    sudo mkdir -p /opt/onemount-ci && \
    sudo chown runner:runner /opt/onemount-ci

# Copy common functions and entrypoint scripts
COPY docker/scripts/common.sh /usr/local/bin/common.sh
COPY docker/scripts/runner-entrypoint.sh /usr/local/bin/runner-entrypoint.sh
COPY docker/scripts/init-workspace.sh /usr/local/bin/init-workspace.sh
COPY docker/scripts/token-manager.sh /usr/local/bin/token-manager.sh
COPY .env /workspace/.env
USER root
RUN chmod +x /usr/local/bin/common.sh && \
    chmod +x /usr/local/bin/runner-entrypoint.sh && \
    chmod +x /usr/local/bin/init-workspace.sh && \
    chmod +x /usr/local/bin/token-manager.sh && \
    chown runner:runner /workspace/.env && \
    chmod 644 /workspace/.env
USER runner

# Copy source code to temporary location for workspace initialization
COPY --chown=runner:runner . /tmp/onemount-source/

# Set working directory for actions
WORKDIR /workspace

# Default entrypoint
ENTRYPOINT ["/usr/local/bin/runner-entrypoint.sh"]
CMD ["--help"]

# Add labels for better image management
LABEL org.opencontainers.image.title="OneMount GitHub Runner (Development)"
LABEL org.opencontainers.image.description="Development Docker image for running GitHub Actions workflows (extends onemount-base)"
LABEL org.opencontainers.image.vendor="Auriora"
LABEL org.opencontainers.image.source="https://github.com/Auriora/OneMount"
LABEL org.opencontainers.image.version="${ONEMOUNT_VERSION:-0.1.0rc1}"
