#!/bin/bash

# OneMount Remote Docker Access Fix Script
# Diagnoses and fixes Docker daemon remote access issues

set -euo pipefail

# Configuration
REMOTE_HOST="${DOCKER_HOST_IP:-172.16.1.104}"
DOCKER_PORT="${DOCKER_PORT:-2376}"
SSH_USER="${SSH_USER:-$(whoami)}"

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

# Check if host is reachable
check_host_connectivity() {
    print_info "Checking connectivity to $REMOTE_HOST..."
    
    if ping -c 3 "$REMOTE_HOST" >/dev/null 2>&1; then
        print_success "Host $REMOTE_HOST is reachable"
    else
        print_error "Host $REMOTE_HOST is not reachable"
        exit 1
    fi
}

# Check Docker daemon status via SSH
check_docker_daemon() {
    print_info "Checking Docker daemon status on $REMOTE_HOST..."
    
    if ssh "$SSH_USER@$REMOTE_HOST" "systemctl is-active docker" >/dev/null 2>&1; then
        print_success "Docker daemon is running on $REMOTE_HOST"
    else
        print_warning "Docker daemon may not be running. Attempting to start..."
        ssh "$SSH_USER@$REMOTE_HOST" "sudo systemctl start docker"
        sleep 5
        
        if ssh "$SSH_USER@$REMOTE_HOST" "systemctl is-active docker" >/dev/null 2>&1; then
            print_success "Docker daemon started successfully"
        else
            print_error "Failed to start Docker daemon"
            return 1
        fi
    fi
}

# Check Docker remote API configuration
check_docker_remote_api() {
    print_info "Checking Docker remote API configuration..."
    
    # Check if Docker is listening on the expected port
    if ssh "$SSH_USER@$REMOTE_HOST" "ss -tlnp | grep :$DOCKER_PORT" >/dev/null 2>&1; then
        print_success "Docker is listening on port $DOCKER_PORT"
        return 0
    else
        print_warning "Docker is not listening on port $DOCKER_PORT"
        return 1
    fi
}

# Configure Docker for remote access
configure_docker_remote() {
    print_info "Configuring Docker for remote access..."
    
    # Create systemd override directory
    ssh "$SSH_USER@$REMOTE_HOST" "sudo mkdir -p /etc/systemd/system/docker.service.d"
    
    # Create override configuration
    ssh "$SSH_USER@$REMOTE_HOST" "sudo tee /etc/systemd/system/docker.service.d/override.conf" << 'EOF'
[Service]
ExecStart=
ExecStart=/usr/bin/dockerd -H fd:// -H tcp://0.0.0.0:2376 --containerd=/run/containerd/containerd.sock
EOF
    
    # Reload systemd and restart Docker
    ssh "$SSH_USER@$REMOTE_HOST" "sudo systemctl daemon-reload"
    ssh "$SSH_USER@$REMOTE_HOST" "sudo systemctl restart docker"
    
    # Wait for Docker to start
    sleep 10
    
    print_success "Docker remote access configured"
}

# Test Docker remote connection
test_docker_connection() {
    print_info "Testing Docker remote connection..."
    
    if DOCKER_HOST="tcp://$REMOTE_HOST:$DOCKER_PORT" docker version >/dev/null 2>&1; then
        print_success "Docker remote connection successful!"
        
        # Show Docker info
        print_info "Remote Docker information:"
        DOCKER_HOST="tcp://$REMOTE_HOST:$DOCKER_PORT" docker version --format "{{.Server.Version}}"
        
        return 0
    else
        print_error "Docker remote connection failed"
        return 1
    fi
}

# Show usage
show_usage() {
    cat << EOF
OneMount Remote Docker Access Fix

Usage: $0 [COMMAND] [OPTIONS]

Commands:
  check             Check current Docker remote access status
  fix               Fix Docker remote access configuration
  test              Test Docker remote connection
  help              Show this help

Environment Variables:
  DOCKER_HOST_IP    Remote Docker host IP (default: 172.16.1.104)
  DOCKER_PORT       Docker remote port (default: 2376)
  SSH_USER          SSH username (default: current user)

Examples:
  $0 check                          # Check current status
  $0 fix                            # Fix configuration
  DOCKER_HOST_IP=192.168.1.100 $0 fix  # Custom host

EOF
}

# Main execution
main() {
    local command="${1:-help}"
    
    case "$command" in
        check)
            check_host_connectivity
            check_docker_daemon
            if check_docker_remote_api; then
                test_docker_connection
            else
                print_warning "Docker remote API is not configured"
                print_info "Run '$0 fix' to configure remote access"
            fi
            ;;
        fix)
            check_host_connectivity
            check_docker_daemon
            configure_docker_remote
            test_docker_connection
            ;;
        test)
            test_docker_connection
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
