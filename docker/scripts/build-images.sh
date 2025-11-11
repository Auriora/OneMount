#!/bin/bash
# Build all OneMount Docker images with proper dependency order

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
VERSION="${ONEMOUNT_VERSION:-0.1.0rc1}"

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

print_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
print_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
print_warning() { echo -e "${YELLOW}[WARNING]${NC} $1"; }

# Parse arguments
BUILD_TARGET="${1:-all}"
NO_CACHE="${2:-}"

CACHE_FLAG=""
if [[ "$NO_CACHE" == "--no-cache" ]]; then
    CACHE_FLAG="--no-cache"
fi

cd "$PROJECT_ROOT"

build_image() {
    local name=$1
    local dockerfile=$2
    local tag=$3
    
    print_info "Building $name..."
    docker buildx build $CACHE_FLAG \
        -f "$dockerfile" \
        -t "$tag:$VERSION" \
        -t "$tag:latest" \
        --build-arg ONEMOUNT_VERSION="$VERSION" \
        --load \
        .
    print_success "Built $name ($tag:$VERSION)"
}

case "$BUILD_TARGET" in
    builder)
        build_image "builder image (with build tools)" "docker/images/builder/Dockerfile" "onemount-builder"
        ;;
    production|runtime)
        # Production requires builder image first
        build_image "builder image" "docker/images/builder/Dockerfile" "onemount-builder"
        build_image "production image (runtime only)" "packaging/docker/Dockerfile" "onemount"
        ;;
    test-runner)
        build_image "builder image" "docker/images/builder/Dockerfile" "onemount-builder"
        build_image "test runner" "docker/images/test-runner/Dockerfile" "onemount-test-runner"
        ;;
    github-runner)
        build_image "builder image" "docker/images/builder/Dockerfile" "onemount-builder"
        build_image "GitHub runner" "docker/images/github-runner/Dockerfile" "onemount-github-runner"
        ;;
    deb-builder)
        build_image "builder image" "docker/images/builder/Dockerfile" "onemount-builder"
        build_image "Debian builder" "docker/images/deb-builder/Dockerfile" "onemount-deb-builder"
        ;;
    dev)
        build_image "builder image" "docker/images/builder/Dockerfile" "onemount-builder"
        build_image "test runner" "docker/images/test-runner/Dockerfile" "onemount-test-runner"
        build_image "GitHub runner" "docker/images/github-runner/Dockerfile" "onemount-github-runner"
        build_image "Debian builder" "docker/images/deb-builder/Dockerfile" "onemount-deb-builder"
        ;;
    all)
        build_image "builder image" "docker/images/builder/Dockerfile" "onemount-builder"
        build_image "production image" "packaging/docker/Dockerfile" "onemount"
        build_image "test runner" "docker/images/test-runner/Dockerfile" "onemount-test-runner"
        build_image "GitHub runner" "docker/images/github-runner/Dockerfile" "onemount-github-runner"
        build_image "Debian builder" "docker/images/deb-builder/Dockerfile" "onemount-deb-builder"
        ;;
    *)
        echo "Usage: $0 {builder|production|test-runner|github-runner|deb-builder|dev|all} [--no-cache]"
        echo ""
        echo "Targets:"
        echo "  builder       - Build builder image (with build tools, for development)"
        echo "  production    - Build production image (runtime only, requires builder)"
        echo "  test-runner   - Build test runner image (includes builder)"
        echo "  github-runner - Build GitHub runner image (includes builder)"
        echo "  deb-builder   - Build Debian builder image (includes builder)"
        echo "  dev           - Build all development images (excludes production)"
        echo "  all           - Build all images"
        exit 1
        ;;
esac

print_success "Build complete!"
echo ""
docker images | grep onemount | head -20
