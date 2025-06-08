#!/bin/bash

# OneMount Optimized Remote Runner Deployment via Docker Socket
# No SSH required - uses Docker remote API

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
REMOTE_HOST="${1:-172.16.1.104}"
DOCKER_PORT="${2:-2375}"
CONTAINER_NAME="onemount-github-runner"
NEW_IMAGE="onemount-github-runner:optimized"

# Logging functions
print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

# Set Docker host
export DOCKER_HOST="tcp://$REMOTE_HOST:$DOCKER_PORT"

print_info "OneMount Optimized Remote Runner Deployment"
print_info "Remote Host: $REMOTE_HOST:$DOCKER_PORT"
echo "=============================================="

# Step 1: Check Docker connection
print_info "Checking Docker connection..."
if ! docker info &> /dev/null; then
    print_error "Cannot connect to Docker daemon at $DOCKER_HOST"
    print_error "Make sure Docker daemon is running and accessible"
    exit 1
fi
print_success "Docker connection established"

# Step 2: Get current runner configuration
print_info "Getting current runner configuration..."
if ! docker inspect $CONTAINER_NAME &> /dev/null; then
    print_error "Container $CONTAINER_NAME not found"
    exit 1
fi

# Extract current environment variables
GITHUB_TOKEN=$(docker inspect $CONTAINER_NAME --format '{{range .Config.Env}}{{println .}}{{end}}' | grep "GITHUB_TOKEN=" | cut -d'=' -f2-)
GITHUB_REPOSITORY=$(docker inspect $CONTAINER_NAME --format '{{range .Config.Env}}{{println .}}{{end}}' | grep "GITHUB_REPOSITORY=" | cut -d'=' -f2-)
RUNNER_NAME=$(docker inspect $CONTAINER_NAME --format '{{range .Config.Env}}{{println .}}{{end}}' | grep "RUNNER_NAME=" | cut -d'=' -f2-)
AUTH_TOKENS_B64=$(docker inspect $CONTAINER_NAME --format '{{range .Config.Env}}{{println .}}{{end}}' | grep "AUTH_TOKENS_B64=" | cut -d'=' -f2- || echo "")

print_success "Current configuration extracted"
print_info "Repository: $GITHUB_REPOSITORY"
print_info "Runner Name: $RUNNER_NAME"

# Step 3: Stop and remove current runner
print_info "Stopping current runner..."
docker stop $CONTAINER_NAME || true
docker rm $CONTAINER_NAME || true
print_success "Current runner stopped and removed"

# Step 4: Create optimized volumes
print_info "Creating optimized volumes..."
docker volume create onemount-runner-workspace || true
docker volume create onemount-runner-work || true
docker volume create onemount-go-cache || true
docker volume create onemount-docker-cache || true
docker volume create onemount-buildkit-cache || true
print_success "Optimized volumes created"

# Step 5: Start optimized runner
print_info "Starting optimized runner..."
docker run -d \
    --name $CONTAINER_NAME \
    --restart unless-stopped \
    --device /dev/fuse \
    --cap-add SYS_ADMIN \
    --security-opt apparmor:unconfined \
    --dns 8.8.8.8 \
    --dns 8.8.4.4 \
    -e "GITHUB_TOKEN=$GITHUB_TOKEN" \
    -e "GITHUB_REPOSITORY=$GITHUB_REPOSITORY" \
    -e "RUNNER_NAME=${RUNNER_NAME}-optimized" \
    -e "RUNNER_LABELS=self-hosted,linux,onemount-testing,docker-remote,optimized" \
    -e "RUNNER_GROUP=Default" \
    -e "AUTH_TOKENS_B64=$AUTH_TOKENS_B64" \
    -e "ONEMOUNT_TEST_TIMEOUT=30m" \
    -e "ONEMOUNT_TEST_VERBOSE=true" \
    -e "DOCKER_BUILDKIT=1" \
    -e "BUILDKIT_PROGRESS=plain" \
    -e "GOPATH=/home/runner/go" \
    -e "GOCACHE=/home/runner/.cache/go-build" \
    -e "GOMODCACHE=/home/runner/go/pkg/mod" \
    -v onemount-runner-workspace:/workspace \
    -v onemount-runner-work:/opt/actions-runner/_work \
    -v onemount-go-cache:/home/runner/go \
    -v onemount-docker-cache:/var/lib/docker \
    -v onemount-buildkit-cache:/tmp/buildkit-cache \
    --memory=4g \
    --memory-reservation=2g \
    --cpus=2 \
    $NEW_IMAGE run

print_success "Optimized runner started successfully!"

# Step 6: Wait for runner to register
print_info "Waiting for runner to register with GitHub..."
sleep 10

# Step 7: Check runner status
print_info "Checking runner status..."
if docker ps --filter "name=$CONTAINER_NAME" --format "table {{.Names}}\t{{.Status}}" | grep -q "Up"; then
    print_success "Runner is running successfully"
    
    # Show logs
    print_info "Recent logs:"
    docker logs --tail 20 $CONTAINER_NAME
    
    print_success "Deployment completed!"
    echo ""
    print_info "Optimizations applied:"
    echo "  ✅ Enhanced Go module caching"
    echo "  ✅ Docker BuildKit optimization"
    echo "  ✅ Resource limits (4GB RAM, 2 CPUs)"
    echo "  ✅ Persistent build caches"
    echo "  ✅ Updated runner labels"
    echo ""
    print_info "Monitor with: docker logs -f $CONTAINER_NAME"
    print_info "Check GitHub: https://github.com/$GITHUB_REPOSITORY/settings/actions/runners"
    
else
    print_error "Runner failed to start"
    print_info "Checking logs..."
    docker logs $CONTAINER_NAME
    exit 1
fi
