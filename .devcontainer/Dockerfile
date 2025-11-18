# syntax=docker/dockerfile:1.7

ARG BASE_IMAGE=mcr.microsoft.com/devcontainers/go:1-1.23-bookworm
FROM ${BASE_IMAGE}

ARG DEBIAN_FRONTEND=noninteractive
ARG USERNAME=vscode
ARG USER_UID=1000
ARG USER_GID=1000

# Install system packages required for JetBrains Remote Development support
RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        build-essential \
        ca-certificates \
        curl \
        fuse3 \
        git \
        libasound2 \
        libatk-bridge2.0-0 \
        libatk1.0-0 \
        libcairo2 \
        libdbus-1-3 \
        libfontconfig1 \
        libfreetype6 \
        libfuse3-dev \
        libgdk-pixbuf2.0-0 \
        libglib2.0-0 \
        libglib2.0-bin \
        libgl1 \
        libglu1-mesa \
        libgtk-3-0 \
        libharfbuzz0b \
        libjson-glib-dev \
        libnss3 \
        libpango-1.0-0 \
        libpangocairo-1.0-0 \
        libsecret-1-0 \
        libwebkit2gtk-4.0-dev \
        libx11-6 \
        libx11-xcb1 \
        libxcomposite1 \
        libxcursor1 \
        libxdamage1 \
        libxext6 \
        libxfixes3 \
        libxi6 \
        libxinerama1 \
        libxkbcommon0 \
        libxkbcommon-x11-0 \
        libxrandr2 \
        libxrender1 \
        libxss1 \
        libxtst6 \
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
        locales \
        openssh-server \
        pkg-config \
        python-is-python3 \
        python3 \
        python3-dev \
        python3-pip \
        python3-venv \
        procps \
        software-properties-common \
        sudo \
        unzip \
        xz-utils \
        x11-apps \
        xdg-utils \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

COPY --chmod=0755 .devcontainer/bin/host-open /usr/local/bin/host-open

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

# Ensure Go tooling uses a writable GOPATH/GOCACHE for the dev user
ENV GOPATH=/home/${USERNAME}/go \
    GOCACHE=/home/${USERNAME}/.cache/go-build \
    PATH=/usr/local/go/bin:/home/${USERNAME}/go/bin:${PATH} \
    REMOTE_DEV_SSH_LINK_HOST=localhost \
    REMOTE_DEV_SSH_LINK_PORT=2222 \
    REMOTE_DEV_SSH_LINK_USER=${USERNAME}

RUN mkdir -p "${GOPATH}/bin" "${GOPATH}/pkg" "${GOCACHE}" \
    && chown -R "${USERNAME}:${USERNAME}" "${GOPATH}" "${GOCACHE}"

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
