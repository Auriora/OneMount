#!/bin/bash

# OneMount Elastic System Deployment Script
# Complete deployment and testing of the elastic runner system

set -euo pipefail

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

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
print_step() { echo -e "\n${BLUE}==== $1 ====${NC}"; }

# Make scripts executable
setup_permissions() {
    print_info "Setting up script permissions..."
    chmod +x "$SCRIPT_DIR/fix-remote-docker.sh"
    chmod +x "$SCRIPT_DIR/setup-elastic-env.sh"
    chmod +x "$SCRIPT_DIR/elastic-runner-manager.sh"
    chmod +x "$SCRIPT_DIR/setup-elastic-runners.sh"
    print_success "Script permissions configured"
}

# Phase 1: Fix Remote Docker Access
fix_docker_access() {
    print_step "Phase 1: Checking Remote Docker Access"

    print_info "Testing Docker remote connection..."
    local docker_host="${DOCKER_HOST:-tcp://172.16.1.104:2376}"

    if DOCKER_HOST="$docker_host" docker version >/dev/null 2>&1; then
        print_success "Docker remote access is working"
        return 0
    else
        print_error "Cannot connect to Docker daemon at $docker_host"
        echo
        print_info "To fix this, run the following commands on the remote host (172.16.1.104):"
        echo
        echo "1. Create systemd override directory:"
        echo "   sudo mkdir -p /etc/systemd/system/docker.service.d"
        echo
        echo "2. Create override configuration:"
        echo "   sudo tee /etc/systemd/system/docker.service.d/override.conf << 'EOF'"
        echo "   [Service]"
        echo "   ExecStart="
        echo "   ExecStart=/usr/bin/dockerd -H fd:// -H tcp://0.0.0.0:2376 --containerd=/run/containerd/containerd.sock"
        echo "   EOF"
        echo
        echo "3. Reload and restart Docker:"
        echo "   sudo systemctl daemon-reload"
        echo "   sudo systemctl restart docker"
        echo
        read -p "Press Enter after configuring Docker on the remote host..."

        # Test again
        if DOCKER_HOST="$docker_host" docker version >/dev/null 2>&1; then
            print_success "Docker remote access is now working"
        else
            print_error "Docker remote access still not working"
            exit 1
        fi
    fi
}

# Phase 2: Environment Setup
setup_environment() {
    print_step "Phase 2: Environment Setup"
    
    # Check if .env exists
    if [[ ! -f "$PROJECT_ROOT/.env" ]]; then
        print_info "Setting up environment configuration..."
        "$SCRIPT_DIR/setup-elastic-env.sh" setup
    else
        print_info "Validating existing environment..."
        "$SCRIPT_DIR/setup-elastic-env.sh" validate
    fi
    
    # Check prerequisites
    print_info "Checking prerequisites..."
    local missing_deps=()
    
    if ! command -v jq >/dev/null 2>&1; then
        missing_deps+=("jq")
    fi
    
    if ! command -v docker-compose >/dev/null 2>&1; then
        missing_deps+=("docker-compose")
    fi
    
    if [[ ${#missing_deps[@]} -gt 0 ]]; then
        print_error "Missing dependencies: ${missing_deps[*]}"
        print_info "Install with: sudo apt-get install ${missing_deps[*]}"
        exit 1
    fi
    
    print_success "Prerequisites check passed"
}

# Phase 3: Deploy Elastic System
deploy_elastic_system() {
    print_step "Phase 3: Deploying Elastic System"
    
    print_info "Building and starting elastic runner system..."
    "$SCRIPT_DIR/setup-elastic-runners.sh" setup
    
    # Wait for system to stabilize
    print_info "Waiting for system to stabilize..."
    sleep 30
    
    # Check status
    print_info "Checking elastic system status..."
    "$SCRIPT_DIR/elastic-runner-manager.sh" status
}

# Phase 4: Testing & Monitoring
test_and_monitor() {
    print_step "Phase 4: Testing & Monitoring"
    
    print_info "Testing elastic system functionality..."
    
    # Test manual scaling
    print_info "Testing manual scale up..."
    "$SCRIPT_DIR/elastic-runner-manager.sh" scale-up 2
    sleep 10
    
    print_info "Current status after scale up:"
    "$SCRIPT_DIR/elastic-runner-manager.sh" status
    
    # Test scale down
    print_info "Testing manual scale down..."
    "$SCRIPT_DIR/elastic-runner-manager.sh" scale-down 1
    sleep 10
    
    print_info "Current status after scale down:"
    "$SCRIPT_DIR/elastic-runner-manager.sh" status
    
    print_success "Manual scaling tests completed"
}

# Start monitoring
start_monitoring() {
    print_step "Starting Monitoring"
    
    print_info "Starting elastic manager monitoring..."
    echo "The elastic manager will now monitor GitHub Actions queue and auto-scale runners."
    echo
    print_info "Monitoring commands:"
    echo "  Status:    $SCRIPT_DIR/elastic-runner-manager.sh status"
    echo "  Dashboard: $SCRIPT_DIR/setup-elastic-runners.sh start-dashboard"
    echo "  Logs:      DOCKER_HOST=tcp://172.16.1.104:2376 docker logs -f onemount-elastic-manager"
    echo
    
    read -p "Do you want to start continuous monitoring now? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        print_info "Starting continuous monitoring (Ctrl+C to stop)..."
        "$SCRIPT_DIR/elastic-runner-manager.sh" monitor
    else
        print_info "You can start monitoring later with:"
        echo "  $SCRIPT_DIR/elastic-runner-manager.sh monitor"
    fi
}

# Install systemd service
install_service() {
    print_step "Installing Systemd Service"
    
    read -p "Do you want to install the elastic runner as a systemd service? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        "$SCRIPT_DIR/setup-elastic-runners.sh" install-service
        print_success "Systemd service installed"
        print_info "Control with:"
        echo "  sudo systemctl start onemount-elastic-runner"
        echo "  sudo systemctl status onemount-elastic-runner"
        echo "  journalctl -u onemount-elastic-runner -f"
    fi
}

# Show final status
show_final_status() {
    print_step "Deployment Complete!"
    
    echo
    print_success "🎉 Elastic runner system deployed successfully!"
    echo
    print_info "System Overview:"
    "$SCRIPT_DIR/elastic-runner-manager.sh" status
    
    echo
    print_info "Next Steps:"
    echo "1. Monitor auto-scaling: $SCRIPT_DIR/elastic-runner-manager.sh monitor"
    echo "2. View dashboard: $SCRIPT_DIR/setup-elastic-runners.sh start-dashboard"
    echo "3. Check logs: DOCKER_HOST=tcp://172.16.1.104:2376 docker logs -f onemount-elastic-manager"
    echo "4. Test workflows: Push commits or create PRs to trigger scaling"
    echo
    print_info "Management Commands:"
    echo "  Status:     $SCRIPT_DIR/elastic-runner-manager.sh status"
    echo "  Scale up:   $SCRIPT_DIR/elastic-runner-manager.sh scale-up [N]"
    echo "  Scale down: $SCRIPT_DIR/elastic-runner-manager.sh scale-down [N]"
    echo "  Cleanup:    $SCRIPT_DIR/elastic-runner-manager.sh cleanup"
}

# Show usage
show_usage() {
    cat << EOF
OneMount Elastic System Deployment

Usage: $0 [COMMAND]

Commands:
  deploy            Complete deployment (all phases)
  fix-docker        Fix Docker remote access only
  setup-env         Setup environment only
  deploy-system     Deploy elastic system only
  test              Test system functionality
  monitor           Start monitoring
  install-service   Install systemd service
  status            Show current status
  help              Show this help

Examples:
  $0 deploy                         # Complete deployment
  $0 status                         # Check current status
  $0 monitor                        # Start monitoring

EOF
}

# Main execution
main() {
    local command="${1:-deploy}"
    
    case "$command" in
        deploy)
            setup_permissions
            fix_docker_access
            setup_environment
            deploy_elastic_system
            test_and_monitor
            install_service
            show_final_status
            ;;
        fix-docker)
            setup_permissions
            fix_docker_access
            ;;
        setup-env)
            setup_permissions
            setup_environment
            ;;
        deploy-system)
            setup_permissions
            deploy_elastic_system
            ;;
        test)
            test_and_monitor
            ;;
        monitor)
            start_monitoring
            ;;
        install-service)
            install_service
            ;;
        status)
            if [[ -f "$PROJECT_ROOT/.env" ]]; then
                "$SCRIPT_DIR/elastic-runner-manager.sh" status
            else
                print_error "System not deployed. Run '$0 deploy' first."
            fi
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
