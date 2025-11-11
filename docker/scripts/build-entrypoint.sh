#!/bin/bash
# Build entrypoint script for OneMount Docker images
# Handles building OneMount binaries inside containers

set -e

# Source common functions
if [[ -f /usr/local/bin/common.sh ]]; then
    source /usr/local/bin/common.sh
else
    # Fallback if common.sh not available
    print_info() { echo "[INFO] $1"; }
    print_success() { echo "[SUCCESS] $1"; }
    print_warning() { echo "[WARNING] $1"; }
    print_error() { echo "[ERROR] $1"; }
fi

# Function to show usage
show_usage() {
    cat << EOF
OneMount Build Entrypoint

Usage: build-entrypoint.sh [COMMAND] [OPTIONS]

Commands:
    binaries        Build OneMount binaries (onemount, onemount-launcher)
    deb             Build Debian package
    test            Run tests
    clean           Clean build artifacts
    help            Show this help message

Options:
    --verbose       Enable verbose output
    --no-cache      Disable build cache
    --output DIR    Output directory (default: /workspace/build)

Examples:
    build-entrypoint.sh binaries
    build-entrypoint.sh deb --output /dist
    build-entrypoint.sh test --verbose

EOF
}

# Parse arguments
COMMAND="${1:-help}"
VERBOSE=false
NO_CACHE=false
OUTPUT_DIR="/workspace/build"

shift || true
while [[ $# -gt 0 ]]; do
    case $1 in
        --verbose)
            VERBOSE=true
            shift
            ;;
        --no-cache)
            NO_CACHE=true
            shift
            ;;
        --output)
            OUTPUT_DIR="$2"
            shift 2
            ;;
        *)
            print_error "Unknown option: $1"
            show_usage
            exit 1
            ;;
    esac
done

# Set verbose mode
if [ "$VERBOSE" = true ]; then
    set -x
fi

# Function to build binaries
build_binaries() {
    print_info "Building OneMount binaries..."
    
    # Create output directory
    mkdir -p "$OUTPUT_DIR/binaries"
    
    # Get commit hash
    COMMIT=$(git rev-parse HEAD 2>/dev/null || echo 'unknown')
    
    # Build flags
    BUILD_FLAGS="-v"
    if [ "$NO_CACHE" = true ]; then
        BUILD_FLAGS="$BUILD_FLAGS -a"
    fi
    
    # Set CGO flags
    export CGO_CFLAGS="-Wno-deprecated-declarations"
    
    # Build onemount
    print_info "Building onemount..."
    go build $BUILD_FLAGS \
        -o "$OUTPUT_DIR/binaries/onemount" \
        -ldflags="-X github.com/auriora/onemount/cmd/common.commit=$COMMIT" \
        ./cmd/onemount
    
    # Build onemount-launcher
    print_info "Building onemount-launcher..."
    go build $BUILD_FLAGS \
        -o "$OUTPUT_DIR/binaries/onemount-launcher" \
        -ldflags="-X github.com/auriora/onemount/cmd/common.commit=$COMMIT" \
        ./cmd/onemount-launcher
    
    print_info "Binaries built successfully:"
    ls -lh "$OUTPUT_DIR/binaries/"
}

# Function to build Debian package
build_deb() {
    print_info "Building Debian package..."
    
    # Check if we're in the builder image
    if [ ! -f "/etc/apt/sources.list.d/docker.list" ] && [ "$(whoami)" != "builder" ]; then
        print_warning "Not in deb-builder image, attempting anyway..."
    fi
    
    # Create output directory
    mkdir -p "$OUTPUT_DIR"
    
    # Build package
    print_info "Running dpkg-buildpackage..."
    dpkg-buildpackage -us -uc -b
    
    # Move packages to output directory
    print_info "Moving packages to $OUTPUT_DIR..."
    mv ../*.deb "$OUTPUT_DIR/" 2>/dev/null || true
    mv ../*.buildinfo "$OUTPUT_DIR/" 2>/dev/null || true
    mv ../*.changes "$OUTPUT_DIR/" 2>/dev/null || true
    
    print_info "Debian package built successfully:"
    ls -lh "$OUTPUT_DIR"/*.deb
}

# Function to run tests
run_tests() {
    print_info "Running tests..."
    
    # Run Go tests
    go test -v ./...
    
    print_info "Tests completed successfully"
}

# Function to clean build artifacts
clean_build() {
    print_info "Cleaning build artifacts..."
    
    rm -rf "$OUTPUT_DIR"
    rm -rf build/
    rm -rf dist/
    
    print_info "Build artifacts cleaned"
}

# Main command handler
case $COMMAND in
    binaries)
        build_binaries
        ;;
    deb)
        build_deb
        ;;
    test)
        run_tests
        ;;
    clean)
        clean_build
        ;;
    help|--help|-h)
        show_usage
        ;;
    *)
        print_error "Unknown command: $COMMAND"
        show_usage
        exit 1
        ;;
esac

print_info "Done!"
