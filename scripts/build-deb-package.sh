#!/bin/bash
# Simple script to build OneMount Debian package in Docker

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Get version from spec file
VERSION=$(grep "^Version:" packaging/rpm/onemount.spec | sed 's/Version: *//')
RELEASE=$(grep "^Release:" packaging/rpm/onemount.spec | sed 's/Release: *//' | awk '{print $1}')

print_info "Building OneMount v${VERSION}-${RELEASE} Debian package"

# Ensure build directories exist
mkdir -p build/packages/deb build/temp

# Clean previous builds
print_info "Cleaning previous builds..."
rm -f build/packages/deb/*.deb build/packages/deb/*.dsc build/packages/deb/*.changes build/packages/deb/*.tar.*
rm -rf build/temp/*

# Build the Docker image if needed
print_info "Ensuring Docker image is available..."
if ! docker image inspect onemount-deb-builder:latest >/dev/null 2>&1; then
    print_info "Building Docker image..."
    docker build -t onemount-deb-builder:latest -f docker/images/deb-builder/Dockerfile .
fi

# Run the build in Docker
print_info "Starting Docker build..."
docker run --rm \
    -v "$(pwd):/build:rw" \
    -w /build \
    -u "$(id -u):$(id -g)" \
    -e HOME=/tmp \
    -e GOPATH=/tmp/go \
    -e GOCACHE=/tmp/go-cache \
    -e GOMODCACHE=/tmp/go/pkg/mod \
    --entrypoint /bin/bash \
    onemount-deb-builder:latest \
    -c '
set -e

VERSION="'"${VERSION}"'"
RELEASE="'"${RELEASE}"'"

echo "[INFO] Creating source tarball..."
mkdir -p build/temp/onemount-${VERSION}

# Copy source files
git ls-files > build/temp/filelist.txt
git rev-parse HEAD > build/temp/.commit
rsync -a --files-from=build/temp/filelist.txt . build/temp/onemount-${VERSION}/
cp build/temp/.commit build/temp/onemount-${VERSION}/

# Use Ubuntu packaging (compatible with Debian)
mv build/temp/onemount-${VERSION}/packaging/ubuntu build/temp/onemount-${VERSION}/debian

# Create vendor directory
echo "[INFO] Creating Go vendor directory..."
go mod vendor
cp -R vendor/ build/temp/onemount-${VERSION}/

# Create source tarball
echo "[INFO] Creating source tarball..."
cd build/temp && tar -czf onemount_${VERSION}.orig.tar.gz onemount-${VERSION}
cd /build

# Build source package
echo "[INFO] Building source package..."
cd build/temp/onemount-${VERSION}
dpkg-buildpackage -S -sa -d -us -uc
cd /build

# Move source package files
mv build/temp/onemount_${VERSION}*.dsc build/packages/deb/ 2>/dev/null || true
mv build/temp/onemount_${VERSION}*_source.* build/packages/deb/ 2>/dev/null || true

# Build binary package
echo "[INFO] Building binary package..."
cd build/temp/onemount-${VERSION}
dpkg-buildpackage -b -d -us -uc
cd /build

# Move binary package files
mv build/temp/onemount_${VERSION}*.deb build/packages/deb/ 2>/dev/null || true
mv build/temp/onemount*_${VERSION}*.deb build/packages/deb/ 2>/dev/null || true
mv build/temp/onemount_${VERSION}*_amd64.* build/packages/deb/ 2>/dev/null || true
mv build/temp/onemount_${VERSION}.orig.tar.gz build/packages/deb/ 2>/dev/null || true

# Clean up
echo "[INFO] Cleaning up build artifacts..."
rm -rf build/temp/* vendor/

echo "[SUCCESS] Build completed!"
ls -lh build/packages/deb/
'

print_success "Debian package built successfully!"
print_info "Package location: build/packages/deb/"
ls -lh build/packages/deb/*.deb
