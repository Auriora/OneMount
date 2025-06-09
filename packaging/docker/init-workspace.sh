#!/bin/bash
# Workspace initialization script for OneMount Docker runners
# Handles syncing source code to Docker volumes when using volume mounts

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_info() {
    echo -e "${BLUE}[WORKSPACE]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[WORKSPACE]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WORKSPACE]${NC} $1"
}

print_error() {
    echo -e "${RED}[WORKSPACE]${NC} $1"
}

# Function to initialize workspace from source
init_workspace() {
    local source_dir="/tmp/onemount-source"
    local workspace_dir="/workspace"
    
    print_info "Initializing workspace..."
    
    # Check if workspace is empty or needs updating
    if [[ ! -f "$workspace_dir/go.mod" ]]; then
        print_info "Workspace appears empty, copying source code..."
        
        # Check if source directory exists (mounted during build)
        if [[ -d "$source_dir" ]]; then
            print_info "Copying from build-time source..."
            cp -r "$source_dir"/* "$workspace_dir/"
            print_success "Source code copied to workspace"
        else
            print_warning "No source directory found at $source_dir"
            print_warning "Workspace will be empty - you may need to sync code manually"
        fi
    else
        print_info "Workspace already contains source code"
        
        # Check if we should update (optional - could be controlled by env var)
        if [[ "${ONEMOUNT_SYNC_WORKSPACE:-false}" == "true" ]]; then
            print_info "ONEMOUNT_SYNC_WORKSPACE=true, updating workspace..."
            if [[ -d "$source_dir" ]]; then
                rsync -av --delete "$source_dir"/ "$workspace_dir/"
                print_success "Workspace updated from source"
            else
                print_warning "Cannot sync - no source directory found"
            fi
        fi
    fi
    
    # Ensure proper ownership
    if [[ "$(id -u)" == "0" ]]; then
        print_info "Fixing workspace ownership..."
        chown -R runner:runner "$workspace_dir"
    fi
    
    # Verify workspace structure
    if [[ -f "$workspace_dir/go.mod" ]]; then
        print_success "Workspace initialized successfully"
        print_info "Go module: $(head -1 "$workspace_dir/go.mod")"
    else
        print_error "Workspace initialization may have failed - no go.mod found"
        return 1
    fi
}

# Function to sync workspace (for development)
sync_workspace() {
    local source_dir="/tmp/onemount-source"
    local workspace_dir="/workspace"
    
    print_info "Syncing workspace..."
    
    if [[ -d "$source_dir" ]]; then
        rsync -av --delete "$source_dir"/ "$workspace_dir/"
        print_success "Workspace synced"
    else
        print_error "No source directory found for syncing"
        return 1
    fi
}

# Main execution
case "${1:-init}" in
    init)
        init_workspace
        ;;
    sync)
        sync_workspace
        ;;
    *)
        print_error "Unknown command: $1"
        echo "Usage: $0 [init|sync]"
        exit 1
        ;;
esac
