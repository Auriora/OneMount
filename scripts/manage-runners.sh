#!/bin/bash

# OneMount GitHub Actions Runners Management
# Simple 2-runner setup for manual management

set -euo pipefail

# Configuration
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DOCKER_HOST="${DOCKER_HOST:-tcp://172.16.1.104:2376}"
COMPOSE_FILE="$PROJECT_ROOT/docker/compose/docker-compose.runners.yml"
STACK_NAME="onemount-runners"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
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

# Check prerequisites
check_prerequisites() {
    if ! command -v docker >/dev/null 2>&1; then
        print_error "Docker is not installed or not in PATH"
        exit 1
    fi

    if ! command -v docker-compose >/dev/null 2>&1 && ! docker compose version >/dev/null 2>&1; then
        print_error "Docker Compose is not installed or not in PATH"
        exit 1
    fi

    if [[ ! -f "$PROJECT_ROOT/.env" ]]; then
        print_error "Environment file not found: $PROJECT_ROOT/.env"
        print_info "Please create .env file with required GitHub configuration"
        exit 1
    fi

    if [[ ! -f "$COMPOSE_FILE" ]]; then
        print_error "Docker Compose file not found: $COMPOSE_FILE"
        exit 1
    fi
}

# Get Docker Compose command
get_compose_cmd() {
    if docker compose version >/dev/null 2>&1; then
        echo "docker compose"
    else
        echo "docker-compose"
    fi
}

# Start runners
start_runners() {
    local runner="${1:-all}"
    
    print_info "Starting GitHub Actions runners..."
    
    cd "$PROJECT_ROOT/docker/compose"
    
    local compose_cmd
    compose_cmd=$(get_compose_cmd)
    
    case "$runner" in
        "1"|"runner-1")
            print_info "Starting runner-1 (primary)..."
            DOCKER_HOST="$DOCKER_HOST" $compose_cmd -f docker-compose.runners.yml --env-file ../../.env up -d runner-1
            ;;
        "2"|"runner-2")
            print_info "Starting runner-2 (secondary)..."
            DOCKER_HOST="$DOCKER_HOST" $compose_cmd -f docker-compose.runners.yml --env-file ../../.env up -d runner-2
            ;;
        "all"|"")
            print_info "Starting both runners..."
            DOCKER_HOST="$DOCKER_HOST" $compose_cmd -f docker-compose.runners.yml --env-file ../../.env up -d
            ;;
        *)
            print_error "Invalid runner: $runner. Use '1', '2', or 'all'"
            exit 1
            ;;
    esac
    
    print_success "Runners started successfully"
}

# Stop runners
stop_runners() {
    local runner="${1:-all}"
    
    print_info "Stopping GitHub Actions runners..."
    
    cd "$PROJECT_ROOT/docker/compose"
    
    local compose_cmd
    compose_cmd=$(get_compose_cmd)
    
    case "$runner" in
        "1"|"runner-1")
            print_info "Stopping runner-1..."
            DOCKER_HOST="$DOCKER_HOST" $compose_cmd -f docker-compose.runners.yml stop runner-1
            ;;
        "2"|"runner-2")
            print_info "Stopping runner-2..."
            DOCKER_HOST="$DOCKER_HOST" $compose_cmd -f docker-compose.runners.yml stop runner-2
            ;;
        "all"|"")
            print_info "Stopping both runners..."
            DOCKER_HOST="$DOCKER_HOST" $compose_cmd -f docker-compose.runners.yml stop
            ;;
        *)
            print_error "Invalid runner: $runner. Use '1', '2', or 'all'"
            exit 1
            ;;
    esac
    
    print_success "Runners stopped successfully"
}

# Restart runners
restart_runners() {
    local runner="${1:-all}"
    
    print_info "Restarting GitHub Actions runners..."
    stop_runners "$runner"
    sleep 2
    start_runners "$runner"
}

# Show status
show_status() {
    print_info "GitHub Actions Runners Status"
    echo
    
    cd "$PROJECT_ROOT/docker/compose"
    
    local compose_cmd
    compose_cmd=$(get_compose_cmd)
    
    # Show container status
    print_info "Container Status:"
    DOCKER_HOST="$DOCKER_HOST" $compose_cmd -f docker-compose.runners.yml ps
    
    echo
    
    # Show GitHub runner status if possible
    if command -v jq >/dev/null 2>&1 && [[ -n "${GITHUB_TOKEN:-}" ]]; then
        print_info "GitHub Runner Status:"
        
        local repo="${GITHUB_REPOSITORY:-}"
        if [[ -n "$repo" ]]; then
            curl -s -H "Authorization: token $GITHUB_TOKEN" \
                 -H "Accept: application/vnd.github.v3+json" \
                 "https://api.github.com/repos/$repo/actions/runners" | \
            jq -r '.runners[] | select(.name | startswith("onemount-runner")) | "\(.name): \(.status) (\(.busy))"' 2>/dev/null || \
            print_warning "Could not fetch GitHub runner status"
        else
            print_warning "GITHUB_REPOSITORY not set, cannot check GitHub status"
        fi
    else
        print_warning "jq or GITHUB_TOKEN not available, cannot check GitHub status"
    fi
}

# Build runner image
build_image() {
    print_info "Building GitHub Actions runner image..."
    
    cd "$PROJECT_ROOT"
    
    DOCKER_HOST="$DOCKER_HOST" docker build \
        -f packaging/docker/Dockerfile.github-runner \
        -t onemount-github-runner:latest \
        .
    
    print_success "Runner image built successfully"
}

# Clean up stopped containers and unused volumes
cleanup() {
    print_info "Cleaning up stopped containers and unused volumes..."
    
    cd "$PROJECT_ROOT/docker/compose"
    
    local compose_cmd
    compose_cmd=$(get_compose_cmd)
    
    # Remove stopped containers
    DOCKER_HOST="$DOCKER_HOST" $compose_cmd -f docker-compose.runners.yml rm -f
    
    # Clean up unused volumes (be careful with this)
    read -p "Do you want to remove unused Docker volumes? This will delete runner data! (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        DOCKER_HOST="$DOCKER_HOST" docker volume prune -f
        print_success "Cleanup completed"
    else
        print_info "Skipped volume cleanup"
    fi
}

# Show usage
show_usage() {
    cat << EOF
OneMount GitHub Actions Runners Management

Usage: $0 [COMMAND] [RUNNER]

Commands:
  start [RUNNER]    Start runners (1, 2, or all)
  stop [RUNNER]     Stop runners (1, 2, or all)
  restart [RUNNER]  Restart runners (1, 2, or all)
  status            Show runner status
  build             Build runner Docker image
  cleanup           Clean up stopped containers and volumes
  help              Show this help

Runners:
  1, runner-1       Primary runner (keep running)
  2, runner-2       Secondary runner (start/stop as needed)
  all               Both runners (default)

Environment Variables:
  DOCKER_HOST       Remote Docker host (default: tcp://172.16.1.104:2376)

Examples:
  $0 start          # Start both runners
  $0 start 1        # Start only runner-1
  $0 stop 2         # Stop only runner-2
  $0 status         # Show current status
  $0 restart all    # Restart both runners

EOF
}

# Main execution
main() {
    local command="${1:-help}"
    local runner="${2:-}"
    
    case "$command" in
        start)
            check_prerequisites
            start_runners "$runner"
            ;;
        stop)
            check_prerequisites
            stop_runners "$runner"
            ;;
        restart)
            check_prerequisites
            restart_runners "$runner"
            ;;
        status)
            check_prerequisites
            show_status
            ;;
        build)
            check_prerequisites
            build_image
            ;;
        cleanup)
            check_prerequisites
            cleanup
            ;;
        help|--help|-h)
            show_usage
            ;;
        *)
            print_error "Unknown command: $command"
            show_usage
            exit 1
            ;;
    esac
}

main "$@"
