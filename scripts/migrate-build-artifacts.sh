#!/bin/bash
# Script to migrate existing build artifacts to the new organized build/ directory structure

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

print_status "Migrating build artifacts to new organized structure..."

# Create the new directory structure
print_status "Creating build directory structure..."
mkdir -p build/binaries
mkdir -p build/packages/deb
mkdir -p build/packages/rpm
mkdir -p build/packages/source
mkdir -p build/docker
mkdir -p build/temp

# Move existing binaries from build/ to build/binaries/
if [ -d "build" ] && [ "$(ls -A build/ 2>/dev/null | grep -v '^packages$\|^binaries$\|^docker$\|^temp$' || true)" ]; then
    print_status "Moving existing binaries to build/binaries/..."
    for file in build/*; do
        if [ -f "$file" ] && [[ "$(basename "$file")" =~ ^(onemount|onemount-launcher|onemount-headless)$ ]]; then
            mv "$file" build/binaries/
            print_success "Moved $(basename "$file") to build/binaries/"
        fi
    done
fi

# Move existing package artifacts from root to appropriate directories
print_status "Moving package artifacts to organized directories..."

# Move Debian/Ubuntu packages
deb_files_moved=0
shopt -s nullglob  # Enable nullglob to handle empty matches
for file in *.deb *.dsc *.changes *.ddeb *.buildinfo; do
    if [ -f "$file" ]; then
        mv "$file" build/packages/deb/
        print_success "Moved $(basename "$file") to build/packages/deb/"
        deb_files_moved=$((deb_files_moved + 1))
    fi
done

# Move RPM packages
rpm_files_moved=0
for file in *.rpm; do
    if [ -f "$file" ]; then
        mv "$file" build/packages/rpm/
        print_success "Moved $(basename "$file") to build/packages/rpm/"
        rpm_files_moved=$((rpm_files_moved + 1))
    fi
done

# Move source tarballs
source_files_moved=0
for file in *.tar.gz *.tar.xz; do
    if [ -f "$file" ] && [[ "$(basename "$file")" =~ ^(onemount|v[0-9]) ]]; then
        mv "$file" build/packages/source/
        print_success "Moved $(basename "$file") to build/packages/source/"
        source_files_moved=$((source_files_moved + 1))
    fi
done
shopt -u nullglob  # Disable nullglob

# Clean up temporary build files
print_status "Cleaning up temporary build files..."
temp_files_cleaned=0
for file in filelist.txt .commit onemount-*/ *.build* *.upload; do
    if [ -e "$file" ]; then
        rm -rf "$file"
        print_success "Removed temporary file/directory: $(basename "$file")"
        temp_files_cleaned=$((temp_files_cleaned + 1))
    fi
done

# Summary
print_success "Migration completed!"
print_status "Summary:"
echo "  - Debian/Ubuntu packages moved: $deb_files_moved"
echo "  - RPM packages moved: $rpm_files_moved"
echo "  - Source tarballs moved: $source_files_moved"
echo "  - Temporary files cleaned: $temp_files_cleaned"

print_status "New build directory structure:"
echo "build/"
echo "├── binaries/           # Compiled executables"
echo "├── packages/"
echo "│   ├── deb/           # Debian/Ubuntu packages"
echo "│   ├── rpm/           # RPM packages"
echo "│   └── source/        # Source tarballs"
echo "├── docker/            # Docker build artifacts"
echo "└── temp/              # Temporary build files"

if [ $deb_files_moved -gt 0 ] || [ $rpm_files_moved -gt 0 ] || [ $source_files_moved -gt 0 ]; then
    print_warning "Build artifacts have been moved. You may want to update any scripts or documentation that reference the old locations."
fi

print_success "Build artifact migration completed successfully!"
