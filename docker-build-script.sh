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

# Get version from spec file
VERSION=$(grep Version packaging/rpm/onemount.spec | sed 's/Version: *//g')
RELEASE=$(grep -oP "Release: *[0-9]+" packaging/rpm/onemount.spec | sed 's/Release: *//g')

print_status "Inside Docker: Building OneMount v${VERSION}-${RELEASE}..."

# Create source tarball
print_status "Creating source tarball..."
mkdir -p "onemount-${VERSION}"

# Copy source files
git ls-files > filelist.txt
git rev-parse HEAD > .commit
echo .commit >> filelist.txt
rsync -a --files-from=filelist.txt . "onemount-${VERSION}/"

# Move debian packaging
mv "onemount-${VERSION}/packaging/deb" "onemount-${VERSION}/debian"

# Create vendor directory
print_status "Creating Go vendor directory..."
go mod vendor
cp -R vendor/ "onemount-${VERSION}/"

# Create tarballs
print_status "Creating source tarballs..."
tar -czf "onemount_${VERSION}.orig.tar.gz" "onemount-${VERSION}"

print_success "Source tarball created"

# Build source package
print_status "Building source package..."
cd "onemount-${VERSION}"
dpkg-buildpackage -S -sa -d -us -uc
cd ..

print_success "Source package built"

# Build binary package
print_status "Building binary package..."
cd "onemount-${VERSION}"
dpkg-buildpackage -b -d -us -uc
cd ..

print_success "Binary package built"

# Clean up build artifacts but keep packages
print_status "Cleaning up build artifacts..."
rm -f filelist.txt .commit
rm -rf "onemount-${VERSION}" vendor/

print_success "Docker build completed!"
print_status "Built packages:"
ls -la *.deb *.dsc *.changes 2>/dev/null || echo "No package files found"
