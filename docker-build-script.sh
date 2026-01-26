#!/bin/bash
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

cd /build

# Set up environment for build user
export HOME=/tmp
export GOPATH=/tmp/go
export GOCACHE=/tmp/go-cache
export GOMODCACHE=/tmp/go/pkg/mod
mkdir -p "$GOPATH" "$GOCACHE" "$GOMODCACHE"

VERSION="0.1.0rc1"
RELEASE="1%{?dist}"

print_status "Inside Docker: Building OneMount v${VERSION}-${RELEASE}..."

# Create source tarball
print_status "Creating source tarball..."
mkdir -p "build/temp/onemount-${VERSION}"

# Copy source files
git ls-files > build/temp/filelist.txt
git rev-parse HEAD > build/temp/.commit
rsync -a --files-from=build/temp/filelist.txt . "build/temp/onemount-${VERSION}/"
# Copy the commit file separately since it's generated in build/temp
cp build/temp/.commit "build/temp/onemount-${VERSION}/"

# Move Ubuntu packaging (compatible with Debian)
mv "build/temp/onemount-${VERSION}/packaging/ubuntu" "build/temp/onemount-${VERSION}/debian"

# Create vendor directory
print_status "Creating Go vendor directory..."
go mod vendor
cp -R vendor/ "build/temp/onemount-${VERSION}/"

# Create tarballs
print_status "Creating source tarballs..."
cd build/temp && tar -czf "onemount_${VERSION}.orig.tar.gz" "onemount-${VERSION}"
cd /build

print_success "Source tarball created"

# Build source package
print_status "Building source package..."
cd "build/temp/onemount-${VERSION}"
dpkg-buildpackage -S -sa -d -us -uc
cd /build

# Move source package files to deb directory
mv build/temp/onemount_${VERSION}*.dsc build/packages/deb/
mv build/temp/onemount_${VERSION}*_source.* build/packages/deb/

print_success "Source package built"

# Build binary package
print_status "Building binary package..."
cd "build/temp/onemount-${VERSION}"
dpkg-buildpackage -b -d -us -uc
cd /build

# Move binary package files to deb directory
mv build/temp/onemount_${VERSION}*.deb build/packages/deb/
mv build/temp/onemount*_${VERSION}*.deb build/packages/deb/ 2>/dev/null || true
mv build/temp/onemount_${VERSION}*_amd64.* build/packages/deb/ 2>/dev/null || true
# Move source tarball to packages directory
mv build/temp/onemount_${VERSION}.orig.tar.gz build/packages/deb/ 2>/dev/null || true

print_success "Binary package built"

# Clean up build artifacts but keep packages
print_status "Cleaning up build artifacts..."
rm -rf build/temp/* vendor/

print_success "Docker build completed!"
print_status "Built packages:"
ls -la build/packages/deb/ 2>/dev/null || echo "No package files found"
