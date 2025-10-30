# syntax=docker/dockerfile:1.7

ARG BASE_IMAGE=mcr.microsoft.com/devcontainers/base:ubuntu-24.04
FROM ${BASE_IMAGE}

ARG DEBIAN_FRONTEND=noninteractive
ARG GO_VERSION=1.24.2
ARG USERNAME=vscode
ARG USER_UID=1000
ARG USER_GID=1000

# Install system packages required for building and running OneMount and JetBrains Remote Development
RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        build-essential \
        ca-certificates \
        curl \
        fuse3 \
        git \
        libasound2t64 \
        libatk-bridge2.0-0t64 \
        libatk1.0-0t64 \
        libcairo2 \
        libdbus-1-3 \
        libfontconfig1 \
        libfreetype6 \
        libfuse3-dev \
        libglib2.0-0t64 \
        libgtk-3-0 \
        libjson-glib-dev \
        libnss3 \
        libpango-1.0-0 \
        libsecret-1-0 \
        libwebkit2gtk-4.1-dev \
        libx11-6 \
        libxcomposite1 \
        libxdamage1 \
        libxext6 \
        libxkbcommon0 \
        libxrandr2 \
        libxrender1 \
        libxss1 \
        libxtst6 \
        locales \
        openssh-server \
        pkg-config \
        procps \
        software-properties-common \
        sudo \
        unzip \
        xz-utils \
    && rm -rf /var/lib/apt/lists/*

# Configure locale so JetBrains backend and Go tooling agree on UTF-8 defaults
RUN locale-gen en_US.UTF-8

# Ensure the requested devcontainer user exists (supports overrides)
RUN if ! id -u "${USERNAME}" >/dev/null 2>&1; then \
        groupadd --gid "${USER_GID}" "${USERNAME}"; \
        useradd --uid "${USER_UID}" --gid "${USER_GID}" -m -s /bin/bash "${USERNAME}"; \
    fi

# Allow the dev container user to sudo without prompting for a password
RUN usermod -aG sudo "${USERNAME}" \
    && echo "${USERNAME} ALL=(ALL) NOPASSWD:ALL" > /etc/sudoers.d/99-${USERNAME}

# Install Go toolchain
RUN curl -fsSL "https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz" \
    | tar -xz -C /usr/local

ENV PATH="/usr/local/go/bin:${PATH}" \
    GO111MODULE=on \
    GOPATH=/home/${USERNAME}/go \
    HOME=/home/${USERNAME} \
    LANG=en_US.UTF-8 \
    LC_ALL=en_US.UTF-8

# Ensure the non-root user can access FUSE inside the container
RUN if getent group fuse >/dev/null; then \
        usermod -a -G fuse "${USERNAME}"; \
    fi \
    && ln -sf /usr/bin/fusermount3 /usr/bin/fusermount

# Pre-create common development directories with correct ownership
RUN mkdir -p /workspaces \
    && mkdir -p /home/${USERNAME}/go/bin \
    && mkdir -p /.jbdevcontainer \
    && chown -R "${USERNAME}:${USERNAME}" /workspaces /home/${USERNAME}/go /.jbdevcontainer

# Prepare OpenSSH so JetBrains Gateway can attach over SSH inside the container
RUN mkdir -p /var/run/sshd

USER ${USERNAME}
WORKDIR /workspaces

# Download Go modules to warm the module cache (speed up first build)
RUN mkdir -p /workspaces/tmp-init \
    && cd /workspaces/tmp-init \
    && go env -w GOPROXY=https://proxy.golang.org,direct \
    && rm -rf /workspaces/tmp-init
