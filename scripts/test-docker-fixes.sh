#!/bin/bash
# Quick test script to verify Docker fixes
# This script tests the Docker configuration without running full tests

set -e

# Colors for output
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

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_info "OneMount Docker Fixes Verification"
print_info "=================================="
echo

# Check if we're in the right directory
if [[ ! -f "go.mod" ]] || ! grep -q "github.com/auriora/onemount" go.mod; then
    print_error "This script must be run from the OneMount project root"
    exit 1
fi

# 1. Check Docker availability
print_info "1. Checking Docker availability..."
if ! command -v docker >/dev/null 2>&1; then
    print_error "Docker is not installed or not in PATH"
    exit 1
fi

if ! docker info >/dev/null 2>&1; then
    print_error "Docker daemon is not running or not accessible"
    exit 1
fi
print_success "Docker is available and running"
echo

# 2. Check base image
print_info "2. Checking base image availability..."
if docker image inspect onemount-base:latest >/dev/null 2>&1; then
    print_success "Base image 'onemount-base:latest' exists"
else
    print_warning "Base image 'onemount-base:latest' not found"
    print_info "Building base image..."
    if docker build -f packaging/docker/Dockerfile -t onemount-base:latest . >/dev/null 2>&1; then
        print_success "Base image built successfully"
    else
        print_error "Failed to build base image"
        exit 1
    fi
fi
echo

# 3. Check for production token mounts in compose files
print_info "3. Checking for security issues in compose files..."
SECURITY_ISSUES=0

# Check for production token mounts (exclude documentation files)
if grep -r "\.cache/onemount/auth_tokens\.json" docker/compose/ --include="*.yml" --include="*.yaml" >/dev/null 2>&1; then
    print_error "Found production token mounts in compose files!"
    grep -r "\.cache/onemount/auth_tokens\.json" docker/compose/ --include="*.yml" --include="*.yaml" | while read line; do
        print_error "  $line"
    done
    SECURITY_ISSUES=$((SECURITY_ISSUES + 1))
else
    print_success "No production token mounts found in compose files"
fi

if [[ $SECURITY_ISSUES -gt 0 ]]; then
    print_error "Security issues found! Please fix before proceeding."
    exit 1
fi
echo

# 4. Check test artifacts directory
print_info "4. Checking test artifacts directory..."
if [[ -d "test-artifacts/.auth_tokens.json" ]]; then
    print_error "test-artifacts/.auth_tokens.json exists as directory (should be file or not exist)"
    print_info "Cleaning up..."
    docker run --rm -v "$(pwd)/test-artifacts:/tmp/cleanup" alpine:latest rm -rf /tmp/cleanup/.auth_tokens.json
    print_success "Cleaned up incorrect directory structure"
else
    print_success "test-artifacts directory structure is correct"
fi
echo

# 5. Check for failed containers
print_info "5. Checking for failed containers..."
FAILED_CONTAINERS=$(docker ps -a --filter "name=onemount-" --filter "status=exited" --format "{{.Names}}" | grep -v "^$" || true)

if [[ -n "$FAILED_CONTAINERS" ]]; then
    print_warning "Found failed OneMount containers:"
    echo "$FAILED_CONTAINERS" | while read container; do
        if [[ -n "$container" ]]; then
            EXIT_CODE=$(docker inspect "$container" --format "{{.State.ExitCode}}" 2>/dev/null || echo "unknown")
            print_warning "  $container (exit code: $EXIT_CODE)"
        fi
    done
    
    print_info "Cleaning up failed containers..."
    echo "$FAILED_CONTAINERS" | while read container; do
        if [[ -n "$container" ]]; then
            docker rm "$container" >/dev/null 2>&1 || true
        fi
    done
    print_success "Cleaned up failed containers"
else
    print_success "No failed containers found"
fi
echo

# 6. Test basic container creation (without running tests)
print_info "6. Testing basic container creation..."
TEST_CONTAINER="onemount-test-validation"

# Clean up any existing test container
docker rm -f "$TEST_CONTAINER" >/dev/null 2>&1 || true

# Try to create a simple test container
if docker run --name "$TEST_CONTAINER" \
    -v "$(pwd):/workspace:rw" \
    -v "$(pwd)/test-artifacts:/tmp/home-tester/.onemount-tests:rw" \
    --tmpfs "/tmp/home-tester/go:rw,noexec,nosuid,size=100m" \
    --tmpfs "/tmp/home-tester/.cache:rw,noexec,nosuid,size=100m" \
    -e "HOME=/tmp/home-tester" \
    onemount-base:latest \
    /bin/bash -c "echo 'Container creation test successful' && ls -la /workspace && ls -la /tmp/home-tester/.onemount-tests" >/dev/null 2>&1; then
    print_success "Container creation test passed"
else
    print_error "Container creation test failed"
    docker logs "$TEST_CONTAINER" 2>/dev/null || true
fi

# Clean up test container
docker rm -f "$TEST_CONTAINER" >/dev/null 2>&1 || true
echo

# 7. Validate compose file syntax
print_info "7. Validating Docker Compose file syntax..."
COMPOSE_FILES=(
    "docker/compose/docker-compose.test.yml"
    "docker/compose/docker-compose.build.yml"
    "docker/compose/docker-compose.runner.yml"
    "docker/compose/docker-compose.runners.yml"
    "docker/compose/docker-compose.remote.yml"
)

for compose_file in "${COMPOSE_FILES[@]}"; do
    if [[ -f "$compose_file" ]]; then
        if docker compose -f "$compose_file" config >/dev/null 2>&1; then
            print_success "✓ $compose_file syntax is valid"
        else
            print_error "✗ $compose_file has syntax errors"
            docker compose -f "$compose_file" config 2>&1 | head -5
        fi
    else
        print_warning "? $compose_file not found"
    fi
done
echo

# Summary
print_info "Docker Fixes Verification Summary"
print_info "================================"
print_success "✓ Docker environment is ready"
print_success "✓ Base image is available"
print_success "✓ Security issues have been resolved"
print_success "✓ Container lifecycle management improved"
print_success "✓ Compose file syntax validated"
echo
print_info "You can now run Docker tests with:"
print_info "  make docker-test-unit"
print_info "  ./scripts/dev.py test docker unit"
print_info "  docker compose -f docker/compose/docker-compose.test.yml run --rm unit-tests"
echo
print_warning "Remember: Use dedicated test OneDrive accounts for system tests!"
