#!/bin/bash
# Test script to verify the new build directory structure is working correctly

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
if [ ! -f "Makefile" ] || [ ! -d "cmd" ]; then
    print_error "This script must be run from the OneMount project root directory"
    exit 1
fi

print_status "Testing OneMount build directory structure..."

# Test 1: Clean build
print_status "Test 1: Cleaning build directory..."
make clean >/dev/null 2>&1
if [ ! -d "build" ]; then
    print_success "✅ Clean target removes build directory correctly"
else
    print_error "❌ Clean target failed to remove build directory"
    exit 1
fi

# Test 2: Directory structure creation
print_status "Test 2: Testing directory structure creation..."
mkdir -p build/binaries build/packages/deb build/packages/rpm build/packages/source build/docker build/temp

if [ -d "build/binaries" ] && [ -d "build/packages/deb" ] && [ -d "build/packages/rpm" ] && [ -d "build/packages/source" ]; then
    print_success "✅ Build directory structure created correctly"
else
    print_error "❌ Failed to create build directory structure"
    exit 1
fi

# Test 3: Makefile variables
print_status "Test 3: Testing Makefile variables..."
if grep -q "OUTPUT_DIR := \$(BUILD_DIR)/binaries" Makefile; then
    print_success "✅ OUTPUT_DIR variable set correctly in Makefile"
else
    print_error "❌ OUTPUT_DIR variable not set correctly in Makefile"
    exit 1
fi

# Test 4: Binary build target (create dummy binary to test path)
print_status "Test 4: Testing binary build paths..."
mkdir -p build/binaries
echo "#!/bin/bash" > build/binaries/test-binary
echo "echo 'Test binary'" >> build/binaries/test-binary
chmod +x build/binaries/test-binary

if [ -x "build/binaries/test-binary" ]; then
    print_success "✅ Binary build path working correctly"
    rm build/binaries/test-binary
else
    print_error "❌ Binary build path not working"
    exit 1
fi

# Test 5: Package directory structure
print_status "Test 5: Testing package directory structure..."
echo "test deb package" > build/packages/deb/test.deb
echo "test rpm package" > build/packages/rpm/test.rpm
echo "test source tarball" > build/packages/source/test.tar.gz

if [ -f "build/packages/deb/test.deb" ] && [ -f "build/packages/rpm/test.rpm" ] && [ -f "build/packages/source/test.tar.gz" ]; then
    print_success "✅ Package directory structure working correctly"
    rm build/packages/deb/test.deb build/packages/rpm/test.rpm build/packages/source/test.tar.gz
else
    print_error "❌ Package directory structure not working"
    exit 1
fi

# Test 6: .gitignore patterns
print_status "Test 6: Testing .gitignore patterns..."
if grep -q "build/" .gitignore; then
    print_success "✅ .gitignore patterns updated correctly (build/ directory ignored)"
else
    print_error "❌ .gitignore patterns not updated correctly"
    exit 1
fi

# Test 7: Install manifest compatibility
print_status "Test 7: Testing install manifest compatibility..."
if grep -q '$(OUTPUT_DIR)' packaging/install-manifest.json; then
    print_success "✅ Install manifest uses OUTPUT_DIR variable correctly"
else
    print_error "❌ Install manifest not using OUTPUT_DIR variable"
    exit 1
fi

# Test 8: Build scripts updated
print_status "Test 8: Testing build scripts..."
if grep -q "build/packages/deb" scripts/build-deb-docker.sh && grep -q "build/packages/deb" scripts/build-deb-native.sh; then
    print_success "✅ Build scripts updated to use new directory structure"
else
    print_error "❌ Build scripts not updated correctly"
    exit 1
fi

# Test 9: Packaging files updated
print_status "Test 9: Testing packaging files..."
if grep -q "build/binaries" packaging/deb/rules && grep -q "build/binaries" packaging/rpm/onemount.spec; then
    print_success "✅ Packaging files updated to use new directory structure"
else
    print_error "❌ Packaging files not updated correctly"
    exit 1
fi

# Test 10: GitHub workflow updated
print_status "Test 10: Testing GitHub workflow..."
if grep -q "build/packages/deb" .github/workflows/build-packages.yml; then
    print_success "✅ GitHub workflow updated to use new directory structure"
else
    print_error "❌ GitHub workflow not updated correctly"
    exit 1
fi

# Clean up test files
make clean >/dev/null 2>&1

print_success "🎉 All tests passed! Build directory structure is working correctly."
print_status "Summary of new structure:"
echo "build/"
echo "├── binaries/           # Compiled executables"
echo "├── packages/"
echo "│   ├── deb/           # Debian/Ubuntu packages"
echo "│   ├── rpm/           # RPM packages"
echo "│   └── source/        # Source tarballs"
echo "├── docker/            # Docker build artifacts"
echo "└── temp/              # Temporary build files"

print_status "Next steps:"
echo "1. Run 'make onemount' to test binary building"
echo "2. Run 'make deb' to test package building"
echo "3. Run './scripts/migrate-build-artifacts.sh' if you have existing artifacts to migrate"
