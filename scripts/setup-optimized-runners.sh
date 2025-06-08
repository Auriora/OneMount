#!/bin/bash

# OneMount Optimized Self-Hosted Runners Setup
# This script sets up multiple optimized self-hosted runners for different workflows

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

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

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
RUNNERS_CONFIG_DIR="$PROJECT_ROOT/.runners"

# Default configuration
DEFAULT_GITHUB_REPOSITORY="Auriora/OneMount"
DEFAULT_RUNNER_LABELS="self-hosted,linux,onemount-testing"

# Runner configurations
declare -A RUNNER_CONFIGS=(
    ["ci"]="onemount-ci-runner,self-hosted,linux,onemount-testing,ci"
    ["coverage"]="onemount-coverage-runner,self-hosted,linux,onemount-testing,coverage"
    ["build"]="onemount-build-runner,self-hosted,linux,onemount-testing,build"
    ["system"]="onemount-system-runner,self-hosted,linux,onemount-testing,system"
)

show_usage() {
    cat << EOF
OneMount Optimized Self-Hosted Runners Setup

Usage: $0 [COMMAND] [OPTIONS]

Commands:
  setup-all         Set up all optimized runners
  setup [TYPE]      Set up specific runner type (ci, coverage, build, system)
  start-all         Start all configured runners
  start [TYPE]      Start specific runner type
  stop-all          Stop all runners
  stop [TYPE]       Stop specific runner type
  status            Show status of all runners
  clean             Clean up all runners and configurations

Options:
  --github-token TOKEN    GitHub personal access token
  --repository REPO       Repository in format 'owner/repo' (default: $DEFAULT_GITHUB_REPOSITORY)
  --auth-tokens FILE      Path to OneDrive auth tokens file
  --dry-run              Show what would be done without executing

Environment Variables:
  GITHUB_TOKEN           GitHub personal access token with repo scope
  GITHUB_REPOSITORY      Repository in format 'owner/repo'
  AUTH_TOKENS_FILE       Path to OneDrive auth tokens file

Examples:
  # Interactive setup for all runners
  $0 setup-all

  # Set up only CI runner
  $0 setup ci --github-token ghp_xxx

  # Start all runners
  $0 start-all

  # Check status
  $0 status

EOF
}

# Function to check prerequisites
check_prerequisites() {
    print_info "Checking prerequisites..."
    
    # Check Docker
    if ! command -v docker &> /dev/null; then
        print_error "Docker is not installed or not in PATH"
        exit 1
    fi
    
    # Check Docker Compose (v2 or v1)
    if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
        print_error "Docker Compose is not installed or not in PATH"
        exit 1
    fi
    
    # Check if Docker daemon is running
    if ! docker info &> /dev/null; then
        print_error "Docker daemon is not running"
        exit 1
    fi
    
    print_success "Prerequisites check passed"
}

# Function to create runner configuration
create_runner_config() {
    local runner_type="$1"
    local github_token="$2"
    local github_repository="$3"
    local auth_tokens_file="$4"
    
    local config_info="${RUNNER_CONFIGS[$runner_type]}"
    local runner_name="${config_info%%,*}"
    local runner_labels="${config_info#*,}"
    
    print_info "Creating configuration for $runner_type runner..."
    
    # Create runner-specific directory
    local runner_dir="$RUNNERS_CONFIG_DIR/$runner_type"
    mkdir -p "$runner_dir"
    
    # Create .env file for this runner
    cat > "$runner_dir/.env" << EOF
# OneMount $runner_type Runner Configuration
GITHUB_TOKEN=$github_token
GITHUB_REPOSITORY=$github_repository
RUNNER_NAME=$runner_name
RUNNER_LABELS=$runner_labels
RUNNER_GROUP=Default

# Performance optimizations
ONEMOUNT_TEST_TIMEOUT=30m
ONEMOUNT_TEST_VERBOSE=true
DOCKER_BUILDKIT=1
BUILDKIT_PROGRESS=plain

EOF
    
    # Add auth tokens if provided
    if [[ -n "$auth_tokens_file" && -f "$auth_tokens_file" ]]; then
        local auth_tokens_b64
        auth_tokens_b64=$(base64 -w 0 "$auth_tokens_file")
        echo "AUTH_TOKENS_B64=$auth_tokens_b64" >> "$runner_dir/.env"
        print_success "OneDrive auth tokens configured for $runner_type runner"
    fi
    
    # Create docker-compose override for this runner
    cat > "$runner_dir/docker-compose.override.yml" << EOF
version: '3.8'

services:
  github-runner:
    container_name: $runner_name
    environment:
      - RUNNER_NAME=$runner_name
      - RUNNER_LABELS=$runner_labels
    
    # Performance optimizations
    deploy:
      resources:
        limits:
          memory: 4G
        reservations:
          memory: 2G
    
    # Additional volumes for caching
    volumes:
      - $runner_name-go-cache:/home/runner/go
      - $runner_name-docker-cache:/var/lib/docker
      - $runner_name-buildkit-cache:/tmp/buildkit-cache

volumes:
  $runner_name-go-cache:
    driver: local
  $runner_name-docker-cache:
    driver: local
  $runner_name-buildkit-cache:
    driver: local

EOF
    
    chmod 600 "$runner_dir/.env"
    print_success "Configuration created for $runner_type runner: $runner_dir"
}

# Function to build optimized runner image
build_runner_image() {
    print_info "Building optimized GitHub Actions runner image..."
    
    cd "$PROJECT_ROOT"
    
    # Build with BuildKit optimizations
    DOCKER_BUILDKIT=1 docker build \
        -f packaging/docker/Dockerfile.github-runner \
        -t onemount-github-runner:optimized \
        --build-arg BUILDKIT_INLINE_CACHE=1 \
        --cache-from onemount-github-runner:latest \
        --cache-from onemount-github-runner:optimized \
        .
    
    # Tag as latest for compatibility
    docker tag onemount-github-runner:optimized onemount-github-runner:latest
    
    print_success "Optimized runner image built successfully"
}

# Function to setup specific runner
setup_runner() {
    local runner_type="$1"
    local github_token="$2"
    local github_repository="$3"
    local auth_tokens_file="$4"
    local dry_run="$5"
    
    if [[ ! -v "RUNNER_CONFIGS[$runner_type]" ]]; then
        print_error "Unknown runner type: $runner_type"
        print_info "Available types: ${!RUNNER_CONFIGS[*]}"
        exit 1
    fi
    
    if [[ "$dry_run" == "true" ]]; then
        print_info "[DRY RUN] Would set up $runner_type runner"
        return 0
    fi
    
    print_info "Setting up $runner_type runner..."
    
    # Create configuration
    create_runner_config "$runner_type" "$github_token" "$github_repository" "$auth_tokens_file"
    
    print_success "$runner_type runner setup completed"
}

# Function to start runner
start_runner() {
    local runner_type="$1"
    local dry_run="$2"
    
    local runner_dir="$RUNNERS_CONFIG_DIR/$runner_type"
    
    if [[ ! -d "$runner_dir" ]]; then
        print_error "Runner $runner_type is not configured. Run setup first."
        exit 1
    fi
    
    if [[ "$dry_run" == "true" ]]; then
        print_info "[DRY RUN] Would start $runner_type runner"
        return 0
    fi
    
    print_info "Starting $runner_type runner..."
    
    cd "$runner_dir"
    
    # Use the main docker-compose file with override
    if command -v docker-compose &> /dev/null; then
        docker-compose \
            -f "$PROJECT_ROOT/docker/compose/docker-compose.runner.yml" \
            -f docker-compose.override.yml \
            --env-file .env \
            up -d github-runner
    else
        docker compose \
            -f "$PROJECT_ROOT/docker/compose/docker-compose.runner.yml" \
            -f docker-compose.override.yml \
            --env-file .env \
            up -d github-runner
    fi
    
    print_success "$runner_type runner started"
}

# Parse command line arguments
COMMAND=""
RUNNER_TYPE=""
GITHUB_TOKEN="${GITHUB_TOKEN:-}"
GITHUB_REPOSITORY="${GITHUB_REPOSITORY:-$DEFAULT_GITHUB_REPOSITORY}"
AUTH_TOKENS_FILE="${AUTH_TOKENS_FILE:-}"
DRY_RUN="false"

while [[ $# -gt 0 ]]; do
    case $1 in
        setup-all|start-all|stop-all|status|clean)
            COMMAND="$1"
            shift
            ;;
        setup|start|stop)
            COMMAND="$1"
            if [[ $# -gt 1 && ! "$2" =~ ^-- ]]; then
                RUNNER_TYPE="$2"
                shift
            fi
            shift
            ;;
        --github-token)
            GITHUB_TOKEN="$2"
            shift 2
            ;;
        --repository)
            GITHUB_REPOSITORY="$2"
            shift 2
            ;;
        --auth-tokens)
            AUTH_TOKENS_FILE="$2"
            shift 2
            ;;
        --dry-run)
            DRY_RUN="true"
            shift
            ;;
        --help|-h|help)
            show_usage
            exit 0
            ;;
        *)
            print_error "Unknown option: $1"
            show_usage
            exit 1
            ;;
    esac
done

# Main execution
main() {
    check_prerequisites
    
    # Create runners config directory
    mkdir -p "$RUNNERS_CONFIG_DIR"
    
    case "$COMMAND" in
        setup-all)
            if [[ -z "$GITHUB_TOKEN" ]]; then
                print_error "GitHub token is required. Use --github-token or set GITHUB_TOKEN environment variable."
                exit 1
            fi
            
            print_info "Setting up all optimized runners..."
            build_runner_image
            
            for runner_type in "${!RUNNER_CONFIGS[@]}"; do
                setup_runner "$runner_type" "$GITHUB_TOKEN" "$GITHUB_REPOSITORY" "$AUTH_TOKENS_FILE" "$DRY_RUN"
            done
            
            print_success "All runners configured successfully!"
            print_info "Next steps:"
            echo "1. Run: $0 start-all"
            echo "2. Check status: $0 status"
            ;;
        setup)
            if [[ -z "$RUNNER_TYPE" ]]; then
                print_error "Runner type is required for setup command"
                print_info "Available types: ${!RUNNER_CONFIGS[*]}"
                exit 1
            fi
            
            if [[ -z "$GITHUB_TOKEN" ]]; then
                print_error "GitHub token is required. Use --github-token or set GITHUB_TOKEN environment variable."
                exit 1
            fi
            
            build_runner_image
            setup_runner "$RUNNER_TYPE" "$GITHUB_TOKEN" "$GITHUB_REPOSITORY" "$AUTH_TOKENS_FILE" "$DRY_RUN"
            ;;
        start-all)
            print_info "Starting all configured runners..."
            for runner_type in "${!RUNNER_CONFIGS[@]}"; do
                if [[ -d "$RUNNERS_CONFIG_DIR/$runner_type" ]]; then
                    start_runner "$runner_type" "$DRY_RUN"
                else
                    print_warning "Runner $runner_type is not configured, skipping..."
                fi
            done
            print_success "All configured runners started!"
            ;;
        start)
            if [[ -z "$RUNNER_TYPE" ]]; then
                print_error "Runner type is required for start command"
                exit 1
            fi
            start_runner "$RUNNER_TYPE" "$DRY_RUN"
            ;;
        stop-all)
            print_info "Stopping all runners..."
            for runner_type in "${!RUNNER_CONFIGS[@]}"; do
                if [[ -d "$RUNNERS_CONFIG_DIR/$runner_type" ]]; then
                    cd "$RUNNERS_CONFIG_DIR/$runner_type"
                    if command -v docker-compose &> /dev/null; then
                        docker-compose \
                            -f "$PROJECT_ROOT/docker/compose/docker-compose.runner.yml" \
                            -f docker-compose.override.yml \
                            --env-file .env \
                            down || true
                    else
                        docker compose \
                            -f "$PROJECT_ROOT/docker/compose/docker-compose.runner.yml" \
                            -f docker-compose.override.yml \
                            --env-file .env \
                            down || true
                    fi
                    print_success "$runner_type runner stopped"
                fi
            done
            ;;
        stop)
            if [[ -z "$RUNNER_TYPE" ]]; then
                print_error "Runner type is required for stop command"
                exit 1
            fi
            cd "$RUNNERS_CONFIG_DIR/$RUNNER_TYPE"
            docker-compose \
                -f "$PROJECT_ROOT/docker/compose/docker-compose.runner.yml" \
                -f docker-compose.override.yml \
                --env-file .env \
                down
            print_success "$RUNNER_TYPE runner stopped"
            ;;
        status)
            print_info "Runner Status:"
            for runner_type in "${!RUNNER_CONFIGS[@]}"; do
                local config_info="${RUNNER_CONFIGS[$runner_type]}"
                local runner_name="${config_info%%,*}"
                if docker ps --format "table {{.Names}}\t{{.Status}}" | grep -q "$runner_name"; then
                    local status=$(docker ps --format "{{.Status}}" --filter "name=$runner_name")
                    print_success "$runner_type ($runner_name): $status"
                else
                    print_warning "$runner_type ($runner_name): Not running"
                fi
            done
            ;;
        clean)
            print_warning "This will remove all runner configurations and containers"
            read -p "Are you sure? (y/N): " -n 1 -r
            echo
            if [[ $REPLY =~ ^[Yy]$ ]]; then
                # Stop all runners
                for runner_type in "${!RUNNER_CONFIGS[@]}"; do
                    if [[ -d "$RUNNERS_CONFIG_DIR/$runner_type" ]]; then
                        cd "$RUNNERS_CONFIG_DIR/$runner_type"
                        docker-compose \
                            -f "$PROJECT_ROOT/docker/compose/docker-compose.runner.yml" \
                            -f docker-compose.override.yml \
                            --env-file .env \
                            down -v || true
                    fi
                done
                # Remove configurations
                rm -rf "$RUNNERS_CONFIG_DIR"
                print_success "All runners cleaned up"
            fi
            ;;
        *)
            print_error "Command is required"
            show_usage
            exit 1
            ;;
    esac
}

# Run main function
main "$@"
