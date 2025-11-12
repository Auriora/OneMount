#!/bin/bash
# Authentication helper functions for test scripts
# Source this file in test scripts: source scripts/lib/auth-helper.sh

# Function to find and setup authentication tokens
# Returns 0 if tokens found, 1 if not found
setup_auth_tokens() {
    local AUTH_TOKEN_FOUND=false
    
    # Standard auth token locations (in priority order)
    local AUTH_LOCATIONS=(
        "$HOME/.onemount-tests/.auth_tokens.json"                    # Standard location
        "/tmp/home-tester/.onemount-tests-auth/.auth_tokens.json"   # Docker mounted location
        "./test-artifacts/.auth_tokens.json"                         # Project location
        "./auth_tokens.json"                                         # Workspace root
    )
    
    # Search for auth tokens
    for location in "${AUTH_LOCATIONS[@]}"; do
        if [ -f "$location" ]; then
            if [ -n "${VERBOSE:-}" ]; then
                echo "[INFO] Auth tokens found at: $location"
            fi
            AUTH_TOKEN_FOUND=true
            
            # Copy to expected location if needed
            local target_dir="$HOME/.onemount-tests"
            local target_file="$target_dir/.auth_tokens.json"
            
            if [ "$location" != "$target_file" ]; then
                mkdir -p "$target_dir"
                cp "$location" "$target_file"
                chmod 600 "$target_file"
                if [ -n "${VERBOSE:-}" ]; then
                    echo "[INFO] Copied auth tokens to: $target_file"
                fi
            fi
            
            # Verify tokens are valid JSON (if jq is available)
            if command -v jq > /dev/null 2>&1; then
                if ! jq empty "$target_file" 2>/dev/null; then
                    echo "[ERROR] Auth tokens are not valid JSON"
                    return 1
                fi
                
                # Check expiration
                local EXPIRES_AT=$(jq -r '.expires_at' "$target_file" 2>/dev/null || echo "0")
                local CURRENT_TIME=$(date +%s)
                
                if [ "$EXPIRES_AT" != "0" ] && [ "$EXPIRES_AT" -le "$CURRENT_TIME" ]; then
                    echo "[WARNING] Auth tokens are EXPIRED"
                    echo "[INFO] Run: ./scripts/setup-test-auth.sh --refresh"
                    return 1
                fi
            else
                # jq not available, skip validation (tokens will be validated by OneMount)
                if [ -n "${VERBOSE:-}" ]; then
                    echo "[INFO] jq not available, skipping token validation"
                fi
            fi
            
            return 0
        fi
    done
    
    if [ "$AUTH_TOKEN_FOUND" = false ]; then
        echo "[ERROR] No auth tokens found in any of the following locations:"
        for location in "${AUTH_LOCATIONS[@]}"; do
            echo "  - $location"
        done
        echo ""
        echo "[INFO] To setup authentication, run:"
        echo "  ./scripts/setup-test-auth.sh"
        echo ""
        echo "[INFO] For Docker, ensure tokens are mounted:"
        echo "  -v \$HOME/.onemount-tests:/tmp/home-tester/.onemount-tests-auth:ro"
        return 1
    fi
    
    return 0
}

# Function to check if running in Docker
is_docker() {
    if [ -f /.dockerenv ]; then
        return 0
    fi
    
    if grep -q docker /proc/1/cgroup 2>/dev/null; then
        return 0
    fi
    
    return 1
}

# Function to get auth token info
get_auth_info() {
    local token_file="$HOME/.onemount-tests/.auth_tokens.json"
    
    if [ ! -f "$token_file" ]; then
        echo "No auth tokens found"
        return 1
    fi
    
    if ! command -v jq > /dev/null 2>&1; then
        echo "jq not installed, cannot read token info"
        return 1
    fi
    
    local account=$(jq -r '.account' "$token_file" 2>/dev/null || echo "unknown")
    local expires_at=$(jq -r '.expires_at' "$token_file" 2>/dev/null || echo "0")
    
    if [ "$expires_at" != "0" ]; then
        local expires_date=$(date -d "@$expires_at" 2>/dev/null || date -r "$expires_at" 2>/dev/null || echo "unknown")
        echo "Account: $account"
        echo "Expires: $expires_date"
    else
        echo "Account: $account"
        echo "Expires: unknown"
    fi
    
    return 0
}
