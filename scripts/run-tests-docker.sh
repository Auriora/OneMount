#!/bin/bash
# Wrapper script for running OneMount tests in Docker containers
# Provides a convenient interface for Docker-based testing

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Print functions
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

# Help function
show_help() {
    cat << EOF
OneMount Docker Test Runner Script

Usage: $0 [COMMAND] [OPTIONS]

Commands:
  build                  Build the test Docker image
  unit                   Run unit tests in Docker
  integration            Run integration tests in Docker
  system                 Run system tests in Docker (requires auth setup)
  all                    Run all tests in Docker
  coverage               Run tests with coverage analysis
  shell                  Start interactive shell in test container
  clean                  Clean up Docker containers and images
  setup-auth             Help setup OneDrive authentication for system tests

Options:
  --verbose              Enable verbose output
  --timeout DURATION     Set test timeout (default: 5m)
  --sequential           Run tests sequentially
  --rebuild              Force rebuild of Docker image
  --no-cache             Build Docker image without cache

Examples:
  # Build test image
  $0 build

  # Run unit tests
  $0 unit

  # Run all tests with verbose output
  $0 all --verbose

  # Run system tests with custom timeout
  $0 system --timeout 30m

  # Start interactive shell for debugging
  $0 shell

  # Clean up Docker resources
  $0 clean

Environment Variables:
  DOCKER_BUILDKIT        Enable Docker BuildKit (default: 1)
  ONEMOUNT_TEST_TIMEOUT  Default test timeout
  ONEMOUNT_TEST_VERBOSE  Default verbose setting

Notes:
  - System tests require OneDrive authentication tokens
  - Use 'setup-auth' command for authentication setup instructions
  - Docker must be installed and running
  - User must have permission to run Docker commands
EOF
}

# Check Docker availability
check_docker() {
    if ! command -v docker &> /dev/null; then
        print_error "Docker is not installed or not in PATH"
        exit 1
    fi
    
    if ! docker info &> /dev/null; then
        print_error "Docker daemon is not running or user lacks permissions"
        print_info "Try: sudo systemctl start docker"
        print_info "Or add user to docker group: sudo usermod -aG docker \$USER"
        exit 1
    fi
}

# Check Docker Compose availability
check_docker_compose() {
    if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
        print_error "Docker Compose is not available"
        print_info "Install docker-compose or use Docker with compose plugin"
        exit 1
    fi
}

# Get Docker Compose command
get_compose_cmd() {
    if command -v docker-compose &> /dev/null; then
        echo "docker-compose"
    else
        echo "docker compose"
    fi
}

# Build Docker image
build_image() {
    print_info "Building OneMount test Docker image..."
    
    local build_args=""
    if [[ "$NO_CACHE" == "true" ]]; then
        build_args="--no-cache"
    fi
    
    # Enable BuildKit for better build performance
    export DOCKER_BUILDKIT=1
    
    if docker build $build_args -f packaging/docker/Dockerfile.test-runner -t onemount-test-runner .; then
        print_success "Docker image built successfully"
    else
        print_error "Failed to build Docker image"
        exit 1
    fi
}

# Setup authentication help
setup_auth_help() {
    cat << EOF

${BLUE}Setting up OneDrive Authentication for System Tests${NC}

System tests require valid OneDrive authentication tokens. Follow these steps:

1. Build OneMount:
   ${YELLOW}make onemount${NC}

2. Authenticate with your test OneDrive account:
   ${YELLOW}./build/onemount --auth-only${NC}

3. Create test directory and copy tokens:
   ${YELLOW}mkdir -p ~/.onemount-tests
   cp ~/.cache/onemount/auth_tokens.json ~/.onemount-tests/.auth_tokens.json${NC}

4. Verify the tokens file exists:
   ${YELLOW}ls -la ~/.onemount-tests/.auth_tokens.json${NC}

5. Now you can run system tests:
   ${YELLOW}$0 system${NC}

${YELLOW}Important Notes:${NC}
- Use a dedicated test OneDrive account, not your production account
- The auth tokens file will be mounted into the Docker container
- System tests create and delete files in /onemount_system_tests/ on OneDrive

EOF
}

# Run tests with Docker Compose
run_tests() {
    local service="$1"
    shift
    local extra_args=("$@")
    
    check_docker_compose
    
    local compose_cmd=$(get_compose_cmd)
    local compose_file="docker-compose.test.yml"
    
    if [[ ! -f "$compose_file" ]]; then
        print_error "Docker Compose file not found: $compose_file"
        exit 1
    fi
    
    # Check if auth tokens exist for system tests
    if [[ "$service" == "system-tests" ]] && [[ ! -f "$HOME/.onemount-tests/.auth_tokens.json" ]]; then
        print_error "OneDrive auth tokens not found for system tests"
        print_info "Run '$0 setup-auth' for setup instructions"
        exit 1
    fi
    
    print_info "Running $service with Docker Compose..."
    
    # Build image if it doesn't exist or if rebuild requested
    if [[ "$REBUILD" == "true" ]] || ! docker image inspect onemount-test-runner:latest &> /dev/null; then
        build_image
    fi
    
    # Run the service
    if $compose_cmd -f "$compose_file" run --rm "$service" "${extra_args[@]}"; then
        print_success "$service completed successfully"
    else
        print_error "$service failed"
        exit 1
    fi
}

# Clean up Docker resources
clean_docker() {
    print_info "Cleaning up OneMount Docker test resources..."
    
    local compose_cmd=$(get_compose_cmd)
    
    # Stop and remove containers
    if [[ -f "docker-compose.test.yml" ]]; then
        $compose_cmd -f docker-compose.test.yml down --remove-orphans 2>/dev/null || true
    fi
    
    # Remove test containers
    docker ps -a --filter "name=onemount-" --format "{{.Names}}" | xargs -r docker rm -f 2>/dev/null || true
    
    # Remove test image
    docker rmi onemount-test-runner:latest 2>/dev/null || true
    
    # Clean up test artifacts
    if [[ -d "test-artifacts" ]]; then
        print_info "Cleaning up test artifacts..."
        rm -rf test-artifacts
    fi
    
    print_success "Docker cleanup complete"
}

# Parse command line arguments
COMMAND="${1:-help}"
shift || true

# Default values
REBUILD="false"
NO_CACHE="false"
VERBOSE="false"
TIMEOUT=""
SEQUENTIAL="false"

# Parse options
while [[ $# -gt 0 ]]; do
    case $1 in
        --verbose)
            VERBOSE="true"
            shift
            ;;
        --timeout)
            TIMEOUT="$2"
            shift 2
            ;;
        --sequential)
            SEQUENTIAL="true"
            shift
            ;;
        --rebuild)
            REBUILD="true"
            shift
            ;;
        --no-cache)
            NO_CACHE="true"
            shift
            ;;
        *)
            print_error "Unknown option: $1"
            show_help
            exit 1
            ;;
    esac
done

# Build extra arguments for test commands
extra_args=()
if [[ "$VERBOSE" == "true" ]]; then
    extra_args+=("--verbose")
fi
if [[ -n "$TIMEOUT" ]]; then
    extra_args+=("--timeout" "$TIMEOUT")
fi
if [[ "$SEQUENTIAL" == "true" ]]; then
    extra_args+=("--sequential")
fi

# Main execution
case "$COMMAND" in
    help|--help|-h)
        show_help
        ;;
    build)
        check_docker
        build_image
        ;;
    unit)
        run_tests "unit-tests" "${extra_args[@]}"
        ;;
    integration)
        run_tests "integration-tests" "${extra_args[@]}"
        ;;
    system)
        run_tests "system-tests" "${extra_args[@]}"
        ;;
    all)
        run_tests "test-runner" "all" "${extra_args[@]}"
        ;;
    coverage)
        run_tests "coverage" "${extra_args[@]}"
        ;;
    shell)
        run_tests "shell"
        ;;
    setup-auth)
        setup_auth_help
        ;;
    clean)
        check_docker
        clean_docker
        ;;
    *)
        print_error "Unknown command: $COMMAND"
        show_help
        exit 1
        ;;
esac
