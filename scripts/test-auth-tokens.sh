#!/bin/bash

# OneMount Authentication Token Test and Refresh Script
# This script validates authentication tokens and attempts to refresh them if expired

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

log_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

log_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

log_error() {
    echo -e "${RED}❌ $1${NC}"
}

# Function to check if OneMount binary exists
check_onemount_binary() {
    if [ ! -f "$PROJECT_ROOT/build/onemount" ]; then
        log_error "OneMount binary not found at $PROJECT_ROOT/build/onemount"
        log_info "Please build OneMount first: make onemount"
        return 1
    fi
    log_success "OneMount binary found"
    return 0
}

# Function to validate JSON file
validate_json() {
    local file="$1"
    if [ ! -f "$file" ]; then
        log_error "File not found: $file"
        return 1
    fi
    
    if ! jq empty "$file" 2>/dev/null; then
        log_error "Invalid JSON in file: $file"
        return 1
    fi
    
    log_success "Valid JSON file: $file"
    return 0
}

# Function to check token expiration
check_token_expiration() {
    local auth_file="$1"
    local expires_at current_time
    
    expires_at=$(jq -r '.expires_at // 0' "$auth_file")
    current_time=$(date +%s)
    
    log_info "Current time: $current_time"
    log_info "Token expires at: $expires_at"
    
    if [ "$expires_at" -le "$current_time" ]; then
        log_warning "Token has expired"
        return 1
    else
        local time_left=$((expires_at - current_time))
        log_success "Token is valid (expires in $time_left seconds)"
        return 0
    fi
}

# Function to test OneDrive API access
test_onedrive_access() {
    local auth_file="$1"
    local access_token response http_code response_body
    
    log_info "Testing OneDrive API access..."
    
    access_token=$(jq -r '.access_token' "$auth_file")
    
    if [ "$access_token" = "null" ] || [ -z "$access_token" ]; then
        log_error "Access token is null or empty"
        return 1
    fi
    
    response=$(curl -s -w "HTTP_CODE:%{http_code}" -H "Authorization: Bearer $access_token" \
        "https://graph.microsoft.com/v1.0/me/drive/root")
    
    http_code=$(echo "$response" | grep -o "HTTP_CODE:[0-9]*" | cut -d: -f2)
    response_body=$(echo "$response" | sed 's/HTTP_CODE:[0-9]*$//')
    
    log_info "HTTP Status Code: $http_code"
    
    if [ "$http_code" = "200" ] && echo "$response_body" | jq -e '.id' > /dev/null 2>&1; then
        local drive_name
        drive_name=$(echo "$response_body" | jq -r '.name // "Unknown"')
        log_success "OneDrive access verified - Drive: $drive_name"
        return 0
    else
        log_error "Failed to access OneDrive"
        log_info "HTTP Status: $http_code"
        log_info "Response: $response_body"
        return 1
    fi
}

# Function to attempt token refresh
refresh_tokens() {
    local auth_file="$1"
    local refresh_token temp_dir
    
    log_info "Attempting to refresh authentication tokens..."
    
    refresh_token=$(jq -r '.refresh_token // ""' "$auth_file")
    
    if [ -z "$refresh_token" ] || [ "$refresh_token" = "null" ]; then
        log_error "No refresh token available - full re-authentication required"
        return 1
    fi
    
    # Create temporary directory for refresh
    temp_dir=$(mktemp -d)
    cp "$auth_file" "$temp_dir/auth_tokens.json"
    
    log_info "Using temporary directory: $temp_dir"
    
    # Attempt refresh using OneMount
    if timeout 30s "$PROJECT_ROOT/build/onemount" --auth-only --config-file /dev/null --cache-dir "$temp_dir" 2>/dev/null; then
        log_success "Token refresh successful"
        
        # Copy refreshed tokens back
        if [ -f "$temp_dir/auth_tokens.json" ]; then
            cp "$temp_dir/auth_tokens.json" "$auth_file"
            log_success "Updated auth tokens file: $auth_file"
        fi
        
        # Clean up
        rm -rf "$temp_dir"
        return 0
    else
        log_error "Token refresh failed"
        rm -rf "$temp_dir"
        return 1
    fi
}

# Main function
main() {
    local auth_file="${1:-$HOME/.onemount-tests/.auth_tokens.json}"
    local refresh_flag="${2:-}"
    
    echo "🔧 OneMount Authentication Token Test"
    echo "====================================="
    echo ""
    
    log_info "Testing auth file: $auth_file"
    echo ""
    
    # Check if OneMount binary exists
    if ! check_onemount_binary; then
        exit 1
    fi
    echo ""
    
    # Validate auth file
    if ! validate_json "$auth_file"; then
        exit 1
    fi
    echo ""
    
    # Check token expiration
    local token_expired=false
    if ! check_token_expiration "$auth_file"; then
        token_expired=true
    fi
    echo ""
    
    # Test OneDrive access
    local api_access_failed=false
    if ! test_onedrive_access "$auth_file"; then
        api_access_failed=true
    fi
    echo ""
    
    # If tokens are expired or API access failed, attempt refresh
    if [ "$token_expired" = true ] || [ "$api_access_failed" = true ] || [ "$refresh_flag" = "--refresh" ]; then
        if refresh_tokens "$auth_file"; then
            echo ""
            log_info "Re-testing after refresh..."
            echo ""
            
            # Re-test expiration
            check_token_expiration "$auth_file"
            echo ""
            
            # Re-test API access
            test_onedrive_access "$auth_file"
            echo ""
        else
            log_error "Token refresh failed - manual re-authentication required"
            echo ""
            log_info "To manually re-authenticate:"
            log_info "1. Run: $PROJECT_ROOT/build/onemount --auth-only"
            log_info "2. Copy tokens: cp ~/.cache/onemount/auth_tokens.json $auth_file"
            exit 1
        fi
    fi
    
    log_success "Authentication token test completed successfully!"
}

# Show usage if requested
if [ "$1" = "--help" ] || [ "$1" = "-h" ]; then
    echo "Usage: $0 [AUTH_FILE] [--refresh]"
    echo ""
    echo "Arguments:"
    echo "  AUTH_FILE    Path to auth tokens file (default: ~/.onemount-tests/.auth_tokens.json)"
    echo "  --refresh    Force token refresh even if tokens appear valid"
    echo ""
    echo "Examples:"
    echo "  $0                                    # Test default auth file"
    echo "  $0 /path/to/auth_tokens.json         # Test specific auth file"
    echo "  $0 ~/.onemount-tests/.auth_tokens.json --refresh  # Force refresh"
    exit 0
fi

# Run main function
main "$@"
