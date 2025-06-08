#!/bin/bash

# OneMount Elastic Runners Setup Script
# Sets up auto-scaling GitHub Actions runners on remote Docker host

set -euo pipefail

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
DOCKER_HOST="${DOCKER_HOST:-172.16.1.104:2376}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
print_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
print_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
print_warning() { echo -e "${YELLOW}[WARNING]${NC} $1"; }
print_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# Check prerequisites
check_prerequisites() {
    print_info "Checking prerequisites..."
    
    # Check if .env file exists
    if [[ ! -f "$PROJECT_ROOT/.env" ]]; then
        print_error "Environment file not found: $PROJECT_ROOT/.env"
        print_info "Run './scripts/manage-runner.sh setup' first"
        exit 1
    fi
    
    # Check Docker connectivity
    if ! DOCKER_HOST="tcp://$DOCKER_HOST" docker info &>/dev/null; then
        print_error "Cannot connect to Docker host: $DOCKER_HOST"
        print_info "Check your Docker host configuration"
        exit 1
    fi
    
    # Check if jq is installed
    if ! command -v jq &>/dev/null; then
        print_error "jq is required but not installed"
        print_info "Install with: sudo apt-get install jq"
        exit 1
    fi
    
    print_success "Prerequisites check passed"
}

# Build runner image
build_image() {
    print_info "Building elastic runner image..."
    
    cd "$PROJECT_ROOT"
    
    # Build with BuildKit optimizations
    DOCKER_BUILDKIT=1 DOCKER_HOST="tcp://$DOCKER_HOST" docker build \
        -f packaging/docker/Dockerfile.github-runner \
        -t onemount-github-runner:elastic \
        --build-arg BUILDKIT_INLINE_CACHE=1 \
        --cache-from onemount-github-runner:latest \
        .
    
    # Tag as latest for compatibility
    DOCKER_HOST="tcp://$DOCKER_HOST" docker tag onemount-github-runner:elastic onemount-github-runner:latest
    
    print_success "Elastic runner image built successfully"
}

# Setup elastic manager
setup_manager() {
    print_info "Setting up elastic manager..."
    
    # Make scripts executable
    chmod +x "$PROJECT_ROOT/scripts/elastic-runner-manager.sh"
    
    # Copy environment file for Docker Compose
    cp "$PROJECT_ROOT/.env" "$PROJECT_ROOT/docker/compose/.env"
    
    print_success "Elastic manager setup completed"
}

# Start elastic system
start_elastic() {
    print_info "Starting elastic runner system..."
    
    cd "$PROJECT_ROOT/docker/compose"
    
    # Start the elastic manager
    DOCKER_HOST="tcp://$DOCKER_HOST" docker compose \
        -f docker-compose.elastic.yml \
        --env-file .env \
        up -d elastic-manager
    
    print_success "Elastic runner system started"
    print_info "Monitor with: ./scripts/elastic-runner-manager.sh status"
}

# Stop elastic system
stop_elastic() {
    print_info "Stopping elastic runner system..."

    cd "$PROJECT_ROOT/docker/compose"

    # Stop the elastic manager
    DOCKER_HOST="tcp://$DOCKER_HOST" docker compose \
        -f docker-compose.elastic.yml \
        down

    # Clean up any running elastic runners
    "$PROJECT_ROOT/scripts/elastic-runner-manager.sh" cleanup

    print_success "Elastic runner system stopped"
}

# Restart elastic system
restart_elastic() {
    print_info "Restarting elastic runner system..."

    # Stop the system first
    stop_elastic

    # Wait a moment for cleanup to complete
    sleep 2

    # Start the system
    check_prerequisites
    setup_manager
    start_elastic

    print_success "Elastic runner system restarted"
}

# Show status
show_status() {
    print_info "Checking elastic runner status..."
    
    # Check if manager is running
    if DOCKER_HOST="tcp://$DOCKER_HOST" docker ps --filter "name=onemount-elastic-manager" --format "table {{.Names}}\t{{.Status}}" | grep -q "onemount-elastic-manager"; then
        print_success "Elastic manager is running"
    else
        print_warning "Elastic manager is not running"
    fi
    
    # Show runner status
    if [[ -f "$PROJECT_ROOT/.env" ]]; then
        "$PROJECT_ROOT/scripts/elastic-runner-manager.sh" status
    else
        print_warning "Cannot show runner status - .env file not found"
    fi
}

# Install systemd service
install_service() {
    print_info "Installing systemd service..."
    
    # Copy service file
    sudo cp "$PROJECT_ROOT/scripts/onemount-elastic-runner.service" /etc/systemd/system/
    
    # Update service file with correct paths
    sudo sed -i "s|/opt/onemount|$PROJECT_ROOT|g" /etc/systemd/system/onemount-elastic-runner.service
    sudo sed -i "s|User=runner|User=$USER|g" /etc/systemd/system/onemount-elastic-runner.service
    
    # Reload systemd and enable service
    sudo systemctl daemon-reload
    sudo systemctl enable onemount-elastic-runner.service
    
    print_success "Systemd service installed"
    print_info "Start with: sudo systemctl start onemount-elastic-runner"
    print_info "Check status: sudo systemctl status onemount-elastic-runner"
}

# Start dashboard
start_dashboard() {
    print_info "Starting monitoring dashboard..."
    
    cd "$PROJECT_ROOT/docker/compose"
    
    # Start the dashboard
    DOCKER_HOST="tcp://$DOCKER_HOST" docker-compose \
        -f docker-compose.elastic.yml \
        --profile dashboard \
        --env-file .env \
        up -d dashboard
    
    print_success "Dashboard started at http://$DOCKER_HOST:8080"
}

# Show usage
show_usage() {
    cat << EOF
OneMount Elastic Runners Setup

Usage: $0 [COMMAND] [OPTIONS]

Commands:
  setup             Complete setup (build + configure + start)
  build             Build elastic runner image
  start             Start elastic runner system
  stop              Stop elastic runner system
  restart           Restart elastic runner system
  status            Show current status
  install-service   Install systemd service
  start-dashboard   Start monitoring dashboard
  help              Show this help

Environment Variables:
  DOCKER_HOST       Remote Docker host (default: 172.16.1.104:2376)

Examples:
  $0 setup                      # Complete setup
  $0 restart                    # Restart the system
  $0 status                     # Check status
  DOCKER_HOST=192.168.1.100:2376 $0 setup  # Custom Docker host

EOF
}

# Main execution
main() {
    local command="${1:-help}"
    
    case "$command" in
        setup)
            check_prerequisites
            build_image
            setup_manager
            start_elastic
            print_success "🎉 Elastic runners setup completed!"
            print_info "Next steps:"
            echo "1. Monitor: ./scripts/elastic-runner-manager.sh status"
            echo "2. Dashboard: $0 start-dashboard"
            echo "3. Service: $0 install-service"
            ;;
        build)
            check_prerequisites
            build_image
            ;;
        start)
            check_prerequisites
            setup_manager
            start_elastic
            ;;
        stop)
            stop_elastic
            ;;
        restart)
            restart_elastic
            ;;
        status)
            show_status
            ;;
        install-service)
            install_service
            ;;
        start-dashboard)
            start_dashboard
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
