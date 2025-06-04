#!/bin/bash
# Native Debian package builder for OneMount (without Docker)

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if we're in the right directory
if [ ! -f "Makefile" ] || [ ! -d "packaging/deb" ]; then
    print_error "This script must be run from the OneMount project root directory"
    exit 1
fi

# Check for required tools
print_status "Checking for required build tools..."
missing_tools=()

for tool in dpkg-buildpackage debuild go git rsync; do
    if ! command -v "$tool" >/dev/null 2>&1; then
        missing_tools+=("$tool")
    fi
done

if [ ${#missing_tools[@]} -ne 0 ]; then
    print_error "Missing required tools: ${missing_tools[*]}"
    print_status "Please install them with:"
    echo "sudo apt install build-essential debhelper devscripts dpkg-dev golang git rsync"
    exit 1
fi

# Get version from spec file
VERSION=$(grep Version packaging/rpm/onemount.spec | sed 's/Version: *//g')
RELEASE=$(grep -oP "Release: *[0-9]+" packaging/rpm/onemount.spec | sed 's/Release: *//g')

print_status "Building OneMount v${VERSION}-${RELEASE} Debian package natively..."

# Create build directory structure
print_status "Creating build directory structure..."
mkdir -p build/packages/deb build/temp

# Clean up any previous builds
print_status "Cleaning up previous builds..."
rm -rf build/temp/* build/packages/deb/* "onemount-${VERSION}" *.deb *.dsc *.changes *.tar.* *.build* *.upload filelist.txt .commit

# Create source tarball
print_status "Creating source tarball..."
mkdir -p "build/temp/onemount-${VERSION}"

# Copy source files
git ls-files > build/temp/filelist.txt
git rev-parse HEAD > build/temp/.commit
echo .commit >> build/temp/filelist.txt
rsync -a --files-from=build/temp/filelist.txt . "build/temp/onemount-${VERSION}/"

# Move debian packaging
mv "build/temp/onemount-${VERSION}/packaging/deb" "build/temp/onemount-${VERSION}/debian"

# Create vendor directory
print_status "Creating Go vendor directory..."
go mod vendor
cp -R vendor/ "build/temp/onemount-${VERSION}/"

# Create tarballs
print_status "Creating source tarballs..."
cd build/temp && tar -czf "../packages/deb/onemount_${VERSION}.orig.tar.gz" "onemount-${VERSION}"
cd ../..

print_success "Source tarball created"

# Build source package
print_status "Building source package..."
cd "build/temp/onemount-${VERSION}"
dpkg-buildpackage -S -sa -d -us -uc
cd ../../..

# Move source package files to deb directory
mv build/temp/onemount_${VERSION}*.dsc build/packages/deb/
mv build/temp/onemount_${VERSION}*_source.* build/packages/deb/

print_success "Source package built"

# Build binary package
print_status "Building binary package..."
cd "build/temp/onemount-${VERSION}"
dpkg-buildpackage -b -us -uc
cd ../../..

# Move binary package files to deb directory
mv build/temp/onemount_${VERSION}*.deb build/packages/deb/
mv build/temp/onemount*_${VERSION}*.deb build/packages/deb/ 2>/dev/null || true
mv build/temp/onemount_${VERSION}*_amd64.* build/packages/deb/ 2>/dev/null || true

print_success "Binary package built"

# Clean up
print_status "Cleaning up build artifacts..."
rm -rf build/temp/* vendor/

print_success "Native Debian package build completed!"
print_status "Built packages:"
ls -la build/packages/deb/ 2>/dev/null || print_warning "No package files found in build/packages/deb/"
