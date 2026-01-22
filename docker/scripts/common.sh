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

# Get authentication token path from environment variable
# This is the SINGLE source of truth - no fallback locations
get_auth_token_path() {
    if [[ -z "$ONEMOUNT_AUTH_PATH" ]]; then
        print_error "ONEMOUNT_AUTH_PATH environment variable is not set"
        echo ""
        print_info "To fix this:"
        print_info "1. Run: ./scripts/setup-auth-reference.sh"
        print_info "2. This will configure Docker Compose to set ONEMOUNT_AUTH_PATH"
        print_info "3. Run tests with: docker compose -f docker/compose/docker-compose.test.yml -f docker/compose/docker-compose.auth.yml run --rm test-runner"
        return 1
    fi
    
    if [[ ! -f "$ONEMOUNT_AUTH_PATH" ]]; then
        print_error "Auth token file not found: $ONEMOUNT_AUTH_PATH"
        echo ""
        print_info "The ONEMOUNT_AUTH_PATH is set but the file doesn't exist."
        print_info "Run: ./scripts/setup-auth-reference.sh"
        return 1
    fi
    
    echo "$ONEMOUNT_AUTH_PATH"
    return 0
}

# Validate authentication is configured
validate_auth_configured() {
    local auth_path
    if ! auth_path=$(get_auth_token_path); then
        return 1
    fi
    
    print_success "Authentication configured: $auth_path"
    return 0
}

