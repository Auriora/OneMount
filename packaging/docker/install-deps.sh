#!/bin/bash
# Install dependencies for OneMount Debian package building

set -e

echo "Installing build dependencies..."

# Update package lists with retries
for i in {1..3}; do
    echo "Attempt $i: Updating package lists..."
    if apt-get update; then
        echo "Package lists updated successfully"
        break
    else
        echo "Failed to update package lists, retrying in 5 seconds..."
        sleep 5
    fi
    if [ $i -eq 3 ]; then
        echo "Failed to update package lists after 3 attempts"
        exit 1
    fi
done

# Install build dependencies
echo "Installing build dependencies..."
apt-get install -y \
    build-essential \
    debhelper \
    devscripts \
    dpkg-dev \
    fakeroot \
    git \
    wget \
    curl \
    pkg-config \
    libwebkit2gtk-4.0-dev \
    libwebkit2gtk-4.1-dev \
    gcc \
    ca-certificates

# Clean up
rm -rf /var/lib/apt/lists/*

# Install Go
echo "Installing Go ${GOVERSION}..."
cd /tmp
wget -O go.tar.gz "https://golang.org/dl/go${GOVERSION}.linux-amd64.tar.gz"
tar -C /usr/local -xzf go.tar.gz
rm go.tar.gz

echo "Dependencies installed successfully!"
