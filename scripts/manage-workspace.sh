#!/bin/bash
# OneMount Workspace Management Script
# Helps manage Docker volumes for runner workspaces

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

show_usage() {
    cat << EOF
OneMount Workspace Management

Usage: $0 [COMMAND] [OPTIONS]

Commands:
  list              List all workspace and token volumes
  sync RUNNER       Sync source code to runner workspace
  clean RUNNER      Clean runner workspace volume
  clean-all         Clean all workspace volumes
  inspect RUNNER    Inspect runner workspace volume
  tokens RUNNER     Manage authentication tokens for runner
  backup RUNNER     Backup runner workspace to tar file
  restore RUNNER    Restore runner workspace from tar file

Runner names:
  runner-1          Primary production runner
  runner-2          Secondary production runner
  remote            Remote runner (docker-compose.remote.yml)

Examples:
  $0 list                    # List all workspace and token volumes
  $0 sync runner-1           # Sync code to runner-1 workspace
  $0 clean runner-2          # Clean runner-2 workspace
  $0 tokens runner-1         # Manage runner-1 authentication tokens
  $0 backup runner-1         # Backup runner-1 workspace

EOF
}

# Function to list workspace and token volumes
list_volumes() {
    print_info "OneMount workspace volumes:"
    echo

    # Use DOCKER_HOST from environment if set
    local docker_cmd="docker"
    if [[ -n "${DOCKER_HOST:-}" ]]; then
        docker_cmd="docker"
    fi

    for volume in onemount-runners_runner-1-workspace onemount-runners_runner-2-workspace onemount-runner-workspace; do
        if $docker_cmd volume inspect "$volume" >/dev/null 2>&1; then
            size=$($docker_cmd run --rm -v "$volume":/workspace alpine du -sh /workspace 2>/dev/null | cut -f1 || echo "unknown")
            print_success "$volume (size: $size)"
        else
            print_warning "$volume (not found)"
        fi
    done

    echo
    print_info "OneMount token volumes:"
    echo

    for volume in onemount-runners_runner-1-tokens onemount-runners_runner-2-tokens onemount-runner-tokens; do
        if $docker_cmd volume inspect "$volume" >/dev/null 2>&1; then
            size=$($docker_cmd run --rm -v "$volume":/tokens alpine du -sh /tokens 2>/dev/null | cut -f1 || echo "unknown")
            print_success "$volume (size: $size)"
        else
            print_warning "$volume (not found)"
        fi
    done
}

# Function to sync source code to workspace
sync_workspace() {
    local runner="$1"
    local volume_name
    
    case "$runner" in
        runner-1)
            volume_name="onemount-runners_runner-1-workspace"
            ;;
        runner-2)
            volume_name="onemount-runners_runner-2-workspace"
            ;;
        remote)
            volume_name="onemount-runner-workspace"
            ;;
        *)
            print_error "Unknown runner: $runner"
            print_info "Valid runners: runner-1, runner-2, remote"
            return 1
            ;;
    esac
    
    print_info "Syncing source code to $volume_name..."
    
    # Use a temporary container to sync the code
    docker run --rm \
        -v "$(pwd)":/source:ro \
        -v "$volume_name":/workspace \
        alpine sh -c "
            rm -rf /workspace/* /workspace/.[^.]* 2>/dev/null || true
            cp -r /source/* /workspace/
            cp -r /source/.* /workspace/ 2>/dev/null || true
            echo 'Source code synced successfully'
        "
    
    print_success "Workspace $volume_name synced"
}

# Function to clean workspace volume
clean_workspace() {
    local runner="$1"
    local volume_name
    
    case "$runner" in
        runner-1)
            volume_name="runner-1-workspace"
            ;;
        runner-2)
            volume_name="runner-2-workspace"
            ;;
        remote)
            volume_name="onemount-runner-workspace"
            ;;
        *)
            print_error "Unknown runner: $runner"
            return 1
            ;;
    esac
    
    print_warning "This will delete all data in $volume_name"
    read -p "Are you sure? (y/N): " -n 1 -r
    echo
    
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        print_info "Cleaning workspace $volume_name..."
        docker run --rm -v "$volume_name":/workspace alpine rm -rf /workspace/*
        print_success "Workspace $volume_name cleaned"
    else
        print_info "Operation cancelled"
    fi
}

# Function to clean all workspace volumes
clean_all_workspaces() {
    print_warning "This will delete all data in ALL workspace volumes"
    read -p "Are you sure? (y/N): " -n 1 -r
    echo
    
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        for runner in runner-1 runner-2 remote; do
            print_info "Cleaning $runner workspace..."
            clean_workspace "$runner" >/dev/null 2>&1 || true
        done
        print_success "All workspaces cleaned"
    else
        print_info "Operation cancelled"
    fi
}

# Function to inspect workspace volume
inspect_workspace() {
    local runner="$1"
    local volume_name
    
    case "$runner" in
        runner-1)
            volume_name="runner-1-workspace"
            ;;
        runner-2)
            volume_name="runner-2-workspace"
            ;;
        remote)
            volume_name="onemount-runner-workspace"
            ;;
        *)
            print_error "Unknown runner: $runner"
            return 1
            ;;
    esac
    
    print_info "Inspecting workspace $volume_name..."
    
    docker run --rm -v "$volume_name":/workspace alpine sh -c "
        echo 'Volume: $volume_name'
        echo 'Size:' \$(du -sh /workspace | cut -f1)
        echo 'Files:'
        ls -la /workspace/ | head -20
        if [ \$(ls -la /workspace/ | wc -l) -gt 21 ]; then
            echo '... (truncated)'
        fi
    "
}

# Function to manage authentication tokens
manage_tokens() {
    local runner="$1"
    local container_name

    case "$runner" in
        runner-1)
            container_name="onemount-runner-1"
            ;;
        runner-2)
            container_name="onemount-runner-2"
            ;;
        remote)
            container_name="onemount-remote-runner"
            ;;
        *)
            print_error "Unknown runner: $runner"
            return 1
            ;;
    esac

    print_info "Managing tokens for $runner..."

    # Check if container is running
    if ! docker ps --format "table {{.Names}}" | grep -q "^$container_name$"; then
        print_warning "Container $container_name is not running"
        print_info "Available commands for stopped containers:"
        echo "  - Start container first, then use token commands"
        return 1
    fi

    print_info "Available token management commands:"
    echo "  1. Show token status"
    echo "  2. Refresh tokens"
    echo "  3. Setup new tokens"
    echo "  4. Validate tokens"
    echo ""

    read -p "Select option (1-4): " -n 1 -r
    echo

    case $REPLY in
        1)
            print_info "Showing token status..."
            docker exec "$container_name" /usr/local/bin/token-manager.sh status
            ;;
        2)
            print_info "Refreshing tokens..."
            docker exec "$container_name" /usr/local/bin/token-manager.sh refresh
            ;;
        3)
            print_info "Setting up new tokens..."
            docker exec "$container_name" /usr/local/bin/token-manager.sh setup
            ;;
        4)
            print_info "Validating tokens..."
            docker exec "$container_name" /usr/local/bin/token-manager.sh validate
            ;;
        *)
            print_error "Invalid option"
            return 1
            ;;
    esac
}

# Main execution
case "${1:-}" in
    list)
        list_volumes
        ;;
    sync)
        if [[ -z "$2" ]]; then
            print_error "Runner name required for sync command"
            show_usage
            exit 1
        fi
        sync_workspace "$2"
        ;;
    clean)
        if [[ -z "$2" ]]; then
            print_error "Runner name required for clean command"
            show_usage
            exit 1
        fi
        clean_workspace "$2"
        ;;
    clean-all)
        clean_all_workspaces
        ;;
    inspect)
        if [[ -z "$2" ]]; then
            print_error "Runner name required for inspect command"
            show_usage
            exit 1
        fi
        inspect_workspace "$2"
        ;;
    tokens)
        if [[ -z "$2" ]]; then
            print_error "Runner name required for tokens command"
            show_usage
            exit 1
        fi
        manage_tokens "$2"
        ;;
    backup|restore)
        print_error "Backup/restore functionality not yet implemented"
        exit 1
        ;;
    --help|-h|help)
        show_usage
        ;;
    *)
        if [[ -n "$1" ]]; then
            print_error "Unknown command: $1"
        fi
        show_usage
        exit 1
        ;;
esac
