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

# Clean up any previous builds
print_status "Cleaning up previous builds..."
rm -rf "onemount-${VERSION}" *.deb *.dsc *.changes *.tar.* *.build* *.upload filelist.txt .commit

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
dpkg-buildpackage -b -us -uc
cd ..

print_success "Binary package built"

# Clean up
print_status "Cleaning up build artifacts..."
rm -f filelist.txt .commit
rm -rf "onemount-${VERSION}" vendor/

print_success "Native Debian package build completed!"
print_status "Built packages:"
ls -la *.deb *.dsc *.changes 2>/dev/null || print_warning "No package files found"
