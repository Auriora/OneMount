#!/bin/bash
# Common functions for OneMount Docker entrypoint scripts

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Print functions
print_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
print_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
print_warning() { echo -e "${YELLOW}[WARNING]${NC} $1"; }
print_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# Environment validation
validate_workspace() {
    if [[ ! -f "go.mod" ]] || ! grep -q "github.com/auriora/onemount" go.mod; then
        print_error "Not in OneMount project directory"
        return 1
    fi
    return 0
}

# Standard auth token locations
AUTH_TOKEN_LOCATIONS=(
    "/workspace/test-artifacts/.auth_tokens.json"
    "/workspace/test-artifacts/auth_tokens.json"
    "/workspace/auth_tokens.json"
    "$HOME/.onemount-tests/.auth_tokens.json"
    "/opt/onemount-ci/auth_tokens.json"
)

# Find auth tokens in standard locations
find_auth_tokens() {
    for location in "${AUTH_TOKEN_LOCATIONS[@]}"; do
        if [[ -f "$location" ]]; then
            echo "$location"
            return 0
        fi
    done
    return 1
}

# Setup auth tokens from standard locations
setup_auth_tokens() {
    local target_dir="${1:-$HOME/.onemount-tests}"
    local target_file="$target_dir/.auth_tokens.json"
    
    # Create target directory
    mkdir -p "$target_dir"
    
    # Check if already in place
    if [[ -f "$target_file" ]]; then
        print_info "Auth tokens already configured at $target_file"
        return 0
    fi
    
    # Find tokens
    local source_file
    if source_file=$(find_auth_tokens); then
        print_info "Found auth tokens at $source_file"
        cp "$source_file" "$target_file"
        chmod 600 "$target_file"
        print_success "Auth tokens configured"
        return 0
    fi
    
    # Check environment variable
    if [[ -n "$ONEMOUNT_AUTH_TOKENS" ]]; then
        print_info "Setting up auth tokens from environment variable"
        echo "$ONEMOUNT_AUTH_TOKENS" > "$target_file"
        chmod 600 "$target_file"
        print_success "Auth tokens configured from environment"
        return 0
    fi
    
    print_warning "No auth tokens found - system tests will be skipped"
    return 1
}
