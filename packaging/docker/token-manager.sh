#!/bin/bash
# OneMount Token Manager for Docker Runners
# Handles automatic token refresh, fallback to environment, and persistent storage

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_info() {
    echo -e "${BLUE}[TOKEN-MGR]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[TOKEN-MGR]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[TOKEN-MGR]${NC} $1"
}

print_error() {
    echo -e "${RED}[TOKEN-MGR]${NC} $1"
}

# Configuration
PERSISTENT_TOKEN_DIR="/opt/onemount-ci"
PERSISTENT_TOKEN_FILE="$PERSISTENT_TOKEN_DIR/auth_tokens.json"
BACKUP_TOKEN_FILE="$PERSISTENT_TOKEN_DIR/auth_tokens_backup.json"
TEMP_TOKEN_FILE="$PERSISTENT_TOKEN_DIR/auth_tokens_temp.json"
REFRESH_LOG_FILE="$PERSISTENT_TOKEN_DIR/token_refresh.log"
ONEMOUNT_BINARY="/workspace/build/binaries/onemount"

# Ensure directories exist
mkdir -p "$PERSISTENT_TOKEN_DIR"

# Function to log with timestamp
log_with_timestamp() {
    echo "$(date '+%Y-%m-%d %H:%M:%S') $1" >> "$REFRESH_LOG_FILE"
}

# Function to check if tokens are expired
check_token_expiration() {
    local token_file="$1"
    
    if [[ ! -f "$token_file" ]]; then
        return 1  # File doesn't exist
    fi
    
    if ! jq empty "$token_file" 2>/dev/null; then
        return 1  # Invalid JSON
    fi
    
    local expires_at=$(jq -r '.expires_at // 0' "$token_file")
    local current_time=$(date +%s)
    local buffer_time=300  # 5 minutes buffer
    
    if [[ "$expires_at" -le $((current_time + buffer_time)) ]]; then
        return 0  # Expired or expiring soon
    fi
    
    return 1  # Not expired
}

# Function to validate token file
validate_token_file() {
    local token_file="$1"
    
    if [[ ! -f "$token_file" ]]; then
        print_error "Token file does not exist: $token_file"
        return 1
    fi
    
    if ! jq empty "$token_file" 2>/dev/null; then
        print_error "Token file contains invalid JSON: $token_file"
        return 1
    fi
    
    # Check required fields
    local access_token=$(jq -r '.access_token // ""' "$token_file")
    local refresh_token=$(jq -r '.refresh_token // ""' "$token_file")
    
    if [[ -z "$access_token" || -z "$refresh_token" ]]; then
        print_error "Token file missing required fields (access_token or refresh_token)"
        return 1
    fi
    
    return 0
}

# Function to backup current tokens
backup_tokens() {
    if [[ -f "$PERSISTENT_TOKEN_FILE" ]]; then
        cp "$PERSISTENT_TOKEN_FILE" "$BACKUP_TOKEN_FILE"
        print_info "Tokens backed up to $BACKUP_TOKEN_FILE"
    fi
}

# Function to restore tokens from backup
restore_from_backup() {
    if [[ -f "$BACKUP_TOKEN_FILE" ]]; then
        cp "$BACKUP_TOKEN_FILE" "$PERSISTENT_TOKEN_FILE"
        print_info "Tokens restored from backup"
        return 0
    fi
    return 1
}

# Function to refresh tokens using OneMount
refresh_tokens_with_onemount() {
    local token_file="$1"
    
    print_info "Attempting token refresh using OneMount..."
    log_with_timestamp "Starting token refresh attempt"
    
    # Create temporary directory for OneMount
    local temp_cache_dir=$(mktemp -d)
    cp "$token_file" "$temp_cache_dir/auth_tokens.json"
    
    # Attempt refresh using OneMount with timeout
    if timeout 60s "$ONEMOUNT_BINARY" --auth-only --config-file /dev/null --cache-dir "$temp_cache_dir" 2>/dev/null; then
        # Check if tokens were updated
        if [[ -f "$temp_cache_dir/auth_tokens.json" ]]; then
            if validate_token_file "$temp_cache_dir/auth_tokens.json"; then
                cp "$temp_cache_dir/auth_tokens.json" "$token_file"
                rm -rf "$temp_cache_dir"
                print_success "Token refresh successful using OneMount"
                log_with_timestamp "Token refresh successful"
                return 0
            else
                print_error "OneMount produced invalid token file"
                log_with_timestamp "Token refresh failed: invalid output"
            fi
        else
            print_error "OneMount did not produce token file"
            log_with_timestamp "Token refresh failed: no output file"
        fi
    else
        print_error "OneMount token refresh failed or timed out"
        log_with_timestamp "Token refresh failed: OneMount error or timeout"
    fi
    
    rm -rf "$temp_cache_dir"
    return 1
}

# Function to load tokens from environment
load_tokens_from_environment() {
    if [[ -n "$AUTH_TOKENS_B64" ]]; then
        print_info "Loading fresh tokens from environment..."
        
        # Decode and validate
        if echo "$AUTH_TOKENS_B64" | base64 -d > "$TEMP_TOKEN_FILE" 2>/dev/null; then
            if validate_token_file "$TEMP_TOKEN_FILE"; then
                cp "$TEMP_TOKEN_FILE" "$PERSISTENT_TOKEN_FILE"
                rm -f "$TEMP_TOKEN_FILE"
                print_success "Fresh tokens loaded from environment"
                log_with_timestamp "Fresh tokens loaded from AUTH_TOKENS_B64"
                return 0
            else
                print_error "Environment tokens are invalid"
                log_with_timestamp "Environment tokens validation failed"
            fi
        else
            print_error "Failed to decode AUTH_TOKENS_B64"
            log_with_timestamp "AUTH_TOKENS_B64 decode failed"
        fi
        rm -f "$TEMP_TOKEN_FILE"
    else
        print_warning "No AUTH_TOKENS_B64 environment variable available"
        log_with_timestamp "No AUTH_TOKENS_B64 available for fallback"
    fi
    return 1
}

# Function to setup initial tokens
setup_tokens() {
    print_info "Setting up authentication tokens..."
    
    # If persistent tokens exist and are valid, use them
    if [[ -f "$PERSISTENT_TOKEN_FILE" ]]; then
        if validate_token_file "$PERSISTENT_TOKEN_FILE"; then
            print_success "Using existing persistent tokens"
            return 0
        else
            print_warning "Existing persistent tokens are invalid, replacing..."
        fi
    fi
    
    # Try to load from environment
    if load_tokens_from_environment; then
        return 0
    fi
    
    print_error "No valid tokens available"
    return 1
}

# Function to ensure tokens are fresh
ensure_fresh_tokens() {
    print_info "Ensuring tokens are fresh..."
    
    # Check if tokens exist
    if [[ ! -f "$PERSISTENT_TOKEN_FILE" ]]; then
        print_warning "No persistent tokens found, setting up..."
        return $(setup_tokens)
    fi
    
    # Validate current tokens
    if ! validate_token_file "$PERSISTENT_TOKEN_FILE"; then
        print_warning "Current tokens are invalid, setting up fresh ones..."
        return $(setup_tokens)
    fi
    
    # Check if tokens are expired or expiring soon
    if check_token_expiration "$PERSISTENT_TOKEN_FILE"; then
        print_info "Tokens are expired or expiring soon, attempting refresh..."
        
        # Backup current tokens
        backup_tokens
        
        # Try to refresh using OneMount
        if refresh_tokens_with_onemount "$PERSISTENT_TOKEN_FILE"; then
            return 0
        fi
        
        # If refresh failed, try loading fresh tokens from environment
        print_warning "Token refresh failed, trying fresh tokens from environment..."
        if load_tokens_from_environment; then
            return 0
        fi
        
        # If all else fails, restore backup and continue
        print_warning "All refresh attempts failed, restoring backup tokens..."
        if restore_from_backup; then
            print_warning "Using potentially expired backup tokens"
            return 0
        fi
        
        print_error "All token refresh attempts failed"
        return 1
    else
        print_success "Current tokens are still valid"
        return 0
    fi
}

# Function to get token status
get_token_status() {
    if [[ ! -f "$PERSISTENT_TOKEN_FILE" ]]; then
        echo "NO_TOKENS"
        return
    fi
    
    if ! validate_token_file "$PERSISTENT_TOKEN_FILE"; then
        echo "INVALID"
        return
    fi
    
    if check_token_expiration "$PERSISTENT_TOKEN_FILE"; then
        echo "EXPIRED"
        return
    fi
    
    echo "VALID"
}

# Function to show token info
show_token_info() {
    local status=$(get_token_status)
    
    print_info "Token Status: $status"
    
    if [[ -f "$PERSISTENT_TOKEN_FILE" ]] && validate_token_file "$PERSISTENT_TOKEN_FILE"; then
        local expires_at=$(jq -r '.expires_at // 0' "$PERSISTENT_TOKEN_FILE")
        local current_time=$(date +%s)
        local account=$(jq -r '.account // "unknown"' "$PERSISTENT_TOKEN_FILE")
        
        if [[ "$expires_at" -gt 0 ]]; then
            local time_left=$((expires_at - current_time))
            if [[ "$time_left" -gt 0 ]]; then
                print_info "Account: $account"
                print_info "Expires in: $((time_left / 3600)) hours, $(((time_left % 3600) / 60)) minutes"
            else
                print_warning "Account: $account"
                print_warning "Expired $(((-time_left) / 3600)) hours, $(((-time_left % 3600) / 60)) minutes ago"
            fi
        fi
    fi
    
    if [[ -f "$REFRESH_LOG_FILE" ]]; then
        print_info "Recent refresh log entries:"
        tail -5 "$REFRESH_LOG_FILE" | while read line; do
            echo "  $line"
        done
    fi
}

# Main execution
case "${1:-ensure}" in
    setup)
        setup_tokens
        ;;
    ensure)
        ensure_fresh_tokens
        ;;
    refresh)
        if [[ -f "$PERSISTENT_TOKEN_FILE" ]]; then
            refresh_tokens_with_onemount "$PERSISTENT_TOKEN_FILE"
        else
            print_error "No tokens to refresh"
            exit 1
        fi
        ;;
    status)
        show_token_info
        ;;
    validate)
        if [[ -f "$PERSISTENT_TOKEN_FILE" ]]; then
            validate_token_file "$PERSISTENT_TOKEN_FILE"
        else
            print_error "No token file to validate"
            exit 1
        fi
        ;;
    *)
        echo "Usage: $0 [setup|ensure|refresh|status|validate]"
        echo ""
        echo "Commands:"
        echo "  setup     - Initial token setup from environment"
        echo "  ensure    - Ensure tokens are fresh (default)"
        echo "  refresh   - Force token refresh using OneMount"
        echo "  status    - Show token status and info"
        echo "  validate  - Validate current token file"
        exit 1
        ;;
esac
