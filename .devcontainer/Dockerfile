# syntax=docker/dockerfile:1.7

ARG BASE_IMAGE=mcr.microsoft.com/devcontainers/base:ubuntu-24.04
FROM ${BASE_IMAGE}

ARG DEBIAN_FRONTEND=noninteractive
ARG GO_VERSION=1.24.2
ARG USERNAME=vscode

# Install system packages required for building and running OneMount
RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        ca-certificates \
        curl \
        fuse3 \
        git \
        libfuse3-dev \
        libjson-glib-dev \
        libwebkit2gtk-4.1-dev \
        pkg-config \
        build-essential \
        software-properties-common \
    && rm -rf /var/lib/apt/lists/*

# Install Go toolchain
RUN curl -fsSL "https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz" \
    | tar -xz -C /usr/local

ENV PATH="/usr/local/go/bin:${PATH}" \
    GO111MODULE=on \
    GOPATH=/home/${USERNAME}/go

# Ensure the non-root user can access FUSE inside the container
RUN if getent group fuse >/dev/null; then \
        usermod -a -G fuse "${USERNAME}"; \
    fi \
    && ln -sf /usr/bin/fusermount3 /usr/bin/fusermount

# Pre-create common development directories with correct ownership
RUN mkdir -p /workspaces \
    && mkdir -p /home/${USERNAME}/go/bin \
    && chown -R "${USERNAME}:${USERNAME}" /workspaces /home/${USERNAME}/go

USER ${USERNAME}
WORKDIR /workspaces

# Download Go modules to warm the module cache (speed up first build)
RUN mkdir -p /workspaces/tmp-init \
    && cd /workspaces/tmp-init \
    && go env -w GOPROXY=https://proxy.golang.org,direct \
    && rm -rf /workspaces/tmp-init
