#!/bin/bash
# Docker-based Ubuntu package builder for OneMount
# Uses the onemount-ubuntu-builder Docker image for clean, reproducible builds
# Optimized for Ubuntu 24.04 LTS and Linux Mint 22

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
if [ ! -f "Makefile" ] || [ ! -d "packaging/ubuntu" ]; then
    print_error "This script must be run from the OneMount project root directory"
    exit 1
fi

# Check if Docker is available
if ! command -v docker >/dev/null 2>&1; then
    print_error "Docker is not installed or not in PATH"
    exit 1
fi

# Check if Docker daemon is running
if ! docker info >/dev/null 2>&1; then
    print_error "Docker daemon is not running"
    exit 1
fi

# Check if the Docker image exists
if ! docker image inspect onemount-ubuntu-builder >/dev/null 2>&1; then
    print_error "Docker image 'onemount-ubuntu-builder' not found"
    print_status "Please build it first with:"
    echo "docker build -t onemount-ubuntu-builder -f packaging/docker/Dockerfile.deb-builder ."
    exit 1
fi

# Get version from spec file
VERSION=$(grep Version packaging/rpm/onemount.spec | sed 's/Version: *//g')
RELEASE=$(grep -oP "Release: *[0-9]+" packaging/rpm/onemount.spec | sed 's/Release: *//g')

print_status "Building OneMount v${VERSION}-${RELEASE} Ubuntu package using Docker..."

# Clean up any previous builds
print_status "Cleaning up previous builds..."
rm -rf "onemount-${VERSION}" *.deb *.dsc *.changes *.tar.* *.build* *.upload filelist.txt .commit

# Create build script for inside Docker
print_status "Creating Docker build script..."
cat > docker-build-script.sh << 'EOF'
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

# Move Ubuntu packaging (compatible with Debian)
mv "onemount-${VERSION}/packaging/ubuntu" "onemount-${VERSION}/debian"

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
EOF

chmod +x docker-build-script.sh

# Run the build in Docker
print_status "Starting Docker build container..."
docker run --rm \
    --user builder \
    -v "$(pwd):/build" \
    -w /build \
    onemount-ubuntu-builder \
    ./docker-build-script.sh

# Clean up the build script
rm -f docker-build-script.sh

print_success "Docker-based Ubuntu package build completed!"
print_status "Built packages:"
ls -la *.deb *.dsc *.changes 2>/dev/null || print_warning "No package files found"
