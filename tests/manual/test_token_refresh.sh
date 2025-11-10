#!/bin/bash
# Manual test script for token refresh mechanism
# This script should be run inside the Docker test container

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

print_test_header() {
    echo ""
    echo "=========================================="
    echo "$1"
    echo "=========================================="
    echo ""
}

# Test configuration
TEST_DIR="/tmp/onemount-refresh-test"
AUTH_TOKENS_FILE="$TEST_DIR/auth_tokens.json"
AUTH_TOKENS_BACKUP="$TEST_DIR/auth_tokens_backup.json"
MOUNT_POINT="$TEST_DIR/mount"

# Cleanup function
cleanup() {
    print_info "Cleaning up test environment..."
    
    # Unmount if mounted
    if mountpoint -q "$MOUNT_POINT" 2>/dev/null; then
        fusermount3 -uz "$MOUNT_POINT" 2>/dev/null || true
    fi
    
    # Remove test directory
    rm -rf "$TEST_DIR"
    
    print_success "Cleanup complete"
}

# Setup test environment
setup() {
    print_info "Setting up test environment..."
    
    # Create test directories
    mkdir -p "$TEST_DIR"
    mkdir -p "$MOUNT_POINT"
    
    # Check if FUSE is available
    if [[ ! -e /dev/fuse ]]; then
        print_error "FUSE device not available"
        print_error "Run container with: --device /dev/fuse --cap-add SYS_ADMIN"
        exit 1
    fi
    
    # Check for existing auth tokens
    if [[ -f "/workspace/test-artifacts/.auth_tokens.json" ]]; then
        print_info "Using existing auth tokens from test-artifacts"
        cp "/workspace/test-artifacts/.auth_tokens.json" "$AUTH_TOKENS_FILE"
    elif [[ -f "/workspace/auth_tokens.json" ]]; then
        print_info "Using existing auth tokens from workspace"
        cp "/workspace/auth_tokens.json" "$AUTH_TOKENS_FILE"
    else
        print_error "No auth tokens found"
        print_error "Please run authentication test first or provide auth tokens"
        exit 1
    fi
    
    # Backup original tokens
    cp "$AUTH_TOKENS_FILE" "$AUTH_TOKENS_BACKUP"
    
    print_success "Test environment ready"
}

# Test 1: Verify token structure
test_token_structure() {
    print_test_header "Test 1: Verify Token Structure"
    
    if ! command -v jq >/dev/null 2>&1; then
        print_error "jq not available, cannot verify token structure"
        return 1
    fi
    
    print_info "Checking token file structure..."
    
    # Verify JSON is valid
    if ! jq empty "$AUTH_TOKENS_FILE" 2>/dev/null; then
        print_error "Invalid JSON in auth tokens file"
        return 1
    fi
    
    # Check required fields
    ACCESS_TOKEN=$(jq -r '.access_token' "$AUTH_TOKENS_FILE")
    REFRESH_TOKEN=$(jq -r '.refresh_token' "$AUTH_TOKENS_FILE")
    EXPIRES_AT=$(jq -r '.expires_at' "$AUTH_TOKENS_FILE")
    EXPIRES_IN=$(jq -r '.expires_in' "$AUTH_TOKENS_FILE")
    
    if [[ -z "$ACCESS_TOKEN" ]] || [[ "$ACCESS_TOKEN" == "null" ]]; then
        print_error "AccessToken missing"
        return 1
    fi
    print_success "AccessToken present"
    
    if [[ -z "$REFRESH_TOKEN" ]] || [[ "$REFRESH_TOKEN" == "null" ]]; then
        print_error "RefreshToken missing"
        return 1
    fi
    print_success "RefreshToken present"
    
    if [[ -z "$EXPIRES_AT" ]] || [[ "$EXPIRES_AT" == "null" ]]; then
        print_error "ExpiresAt missing"
        return 1
    fi
    print_success "ExpiresAt present: $EXPIRES_AT"
    
    # Check if token is expired
    CURRENT_TIME=$(date +%s)
    if [[ "$EXPIRES_AT" -gt "$CURRENT_TIME" ]]; then
        SECONDS_UNTIL_EXPIRY=$((EXPIRES_AT - CURRENT_TIME))
        print_success "Token is valid (expires in $SECONDS_UNTIL_EXPIRY seconds)"
    else
        print_warning "Token is already expired"
    fi
    
    return 0
}

# Test 2: Force token expiration and test refresh
test_force_expiration() {
    print_test_header "Test 2: Force Token Expiration and Refresh"
    
    if ! command -v jq >/dev/null 2>&1; then
        print_error "jq not available, cannot modify tokens"
        return 1
    fi
    
    print_info "Forcing token expiration by setting ExpiresAt to 0..."
    
    # Get original values
    ORIGINAL_ACCESS_TOKEN=$(jq -r '.access_token' "$AUTH_TOKENS_FILE")
    ORIGINAL_EXPIRES_AT=$(jq -r '.expires_at' "$AUTH_TOKENS_FILE")
    
    print_info "Original ExpiresAt: $ORIGINAL_EXPIRES_AT"
    print_info "Original AccessToken (first 20 chars): ${ORIGINAL_ACCESS_TOKEN:0:20}..."
    
    # Set ExpiresAt to 0 to force refresh
    jq '.expires_at = 0' "$AUTH_TOKENS_FILE" > "$AUTH_TOKENS_FILE.tmp"
    mv "$AUTH_TOKENS_FILE.tmp" "$AUTH_TOKENS_FILE"
    
    print_success "Token expiration forced (ExpiresAt = 0)"
    
    # Try to mount (this should trigger token refresh)
    print_info "Attempting to mount filesystem (should trigger token refresh)..."
    
    if timeout 30 ./build/onemount \
        --config-file "$AUTH_TOKENS_FILE" \
        "$MOUNT_POINT" &
    then
        MOUNT_PID=$!
        
        # Wait for mount and potential refresh
        sleep 10
        
        # Check if tokens were refreshed
        NEW_ACCESS_TOKEN=$(jq -r '.access_token' "$AUTH_TOKENS_FILE")
        NEW_EXPIRES_AT=$(jq -r '.expires_at' "$AUTH_TOKENS_FILE")
        
        print_info "New ExpiresAt: $NEW_EXPIRES_AT"
        print_info "New AccessToken (first 20 chars): ${NEW_ACCESS_TOKEN:0:20}..."
        
        # Verify refresh occurred
        if [[ "$NEW_EXPIRES_AT" != "0" ]] && [[ "$NEW_EXPIRES_AT" != "$ORIGINAL_EXPIRES_AT" ]]; then
            print_success "ExpiresAt was updated (refresh occurred)"
            
            # Check if ExpiresAt is in the future
            CURRENT_TIME=$(date +%s)
            if [[ "$NEW_EXPIRES_AT" -gt "$CURRENT_TIME" ]]; then
                SECONDS_UNTIL_EXPIRY=$((NEW_EXPIRES_AT - CURRENT_TIME))
                print_success "New token is valid (expires in $SECONDS_UNTIL_EXPIRY seconds)"
            else
                print_error "New token is still expired"
                kill $MOUNT_PID 2>/dev/null || true
                return 1
            fi
        else
            print_error "Token refresh did not occur (ExpiresAt unchanged)"
            kill $MOUNT_PID 2>/dev/null || true
            return 1
        fi
        
        # Check if access token changed (it should)
        if [[ "$NEW_ACCESS_TOKEN" != "$ORIGINAL_ACCESS_TOKEN" ]]; then
            print_success "AccessToken was updated"
        else
            print_warning "AccessToken unchanged (may indicate refresh didn't occur)"
        fi
        
        # Check if filesystem is mounted
        if mountpoint -q "$MOUNT_POINT"; then
            print_success "Filesystem mounted successfully after refresh"
            
            # Try to list files
            if timeout 10 ls "$MOUNT_POINT" >/dev/null 2>&1; then
                print_success "Can list files after token refresh"
            else
                print_warning "Cannot list files after token refresh"
            fi
            
            # Unmount
            fusermount3 -uz "$MOUNT_POINT"
            wait $MOUNT_PID 2>/dev/null || true
        else
            print_error "Filesystem not mounted after refresh attempt"
            kill $MOUNT_PID 2>/dev/null || true
            return 1
        fi
        
        return 0
    else
        print_error "Failed to start OneMount"
        return 1
    fi
}

# Test 3: Verify token persistence after refresh
test_token_persistence() {
    print_test_header "Test 3: Verify Token Persistence After Refresh"
    
    if ! command -v jq >/dev/null 2>&1; then
        print_error "jq not available"
        return 1
    fi
    
    print_info "Checking if refreshed tokens are persisted to disk..."
    
    # Verify file exists
    if [[ ! -f "$AUTH_TOKENS_FILE" ]]; then
        print_error "Auth tokens file not found"
        return 1
    fi
    
    # Verify file is valid JSON
    if ! jq empty "$AUTH_TOKENS_FILE" 2>/dev/null; then
        print_error "Auth tokens file is not valid JSON"
        return 1
    fi
    
    print_success "Auth tokens file is valid JSON"
    
    # Check file permissions
    PERMS=$(stat -c "%a" "$AUTH_TOKENS_FILE")
    if [[ "$PERMS" == "600" ]]; then
        print_success "File permissions are correct: $PERMS"
    else
        print_warning "File permissions are: $PERMS (expected 600)"
    fi
    
    # Verify token structure
    ACCESS_TOKEN=$(jq -r '.access_token' "$AUTH_TOKENS_FILE")
    REFRESH_TOKEN=$(jq -r '.refresh_token' "$AUTH_TOKENS_FILE")
    EXPIRES_AT=$(jq -r '.expires_at' "$AUTH_TOKENS_FILE")
    
    if [[ -n "$ACCESS_TOKEN" ]] && [[ "$ACCESS_TOKEN" != "null" ]] && \
       [[ -n "$REFRESH_TOKEN" ]] && [[ "$REFRESH_TOKEN" != "null" ]] && \
       [[ -n "$EXPIRES_AT" ]] && [[ "$EXPIRES_AT" != "null" ]]; then
        print_success "All required token fields are present and persisted"
        
        # Verify token is not expired
        CURRENT_TIME=$(date +%s)
        if [[ "$EXPIRES_AT" -gt "$CURRENT_TIME" ]]; then
            print_success "Persisted token is valid"
            return 0
        else
            print_error "Persisted token is expired"
            return 1
        fi
    else
        print_error "Token structure incomplete after persistence"
        return 1
    fi
}

# Test 4: Test automatic refresh on subsequent operations
test_automatic_refresh() {
    print_test_header "Test 4: Test Automatic Refresh on Subsequent Operations"
    
    print_info "Restoring original tokens and forcing expiration..."
    
    # Restore backup
    cp "$AUTH_TOKENS_BACKUP" "$AUTH_TOKENS_FILE"
    
    if ! command -v jq >/dev/null 2>&1; then
        print_error "jq not available"
        return 1
    fi
    
    # Force expiration
    jq '.expires_at = 0' "$AUTH_TOKENS_FILE" > "$AUTH_TOKENS_FILE.tmp"
    mv "$AUTH_TOKENS_FILE.tmp" "$AUTH_TOKENS_FILE"
    
    print_info "Mounting filesystem with expired token..."
    
    if timeout 30 ./build/onemount \
        --config-file "$AUTH_TOKENS_FILE" \
        "$MOUNT_POINT" &
    then
        MOUNT_PID=$!
        sleep 10
        
        if mountpoint -q "$MOUNT_POINT"; then
            print_success "Filesystem mounted (token should have been refreshed)"
            
            # Perform multiple operations to verify token remains valid
            print_info "Performing multiple file operations..."
            
            for i in {1..3}; do
                print_info "Operation $i: Listing files..."
                if timeout 10 ls "$MOUNT_POINT" >/dev/null 2>&1; then
                    print_success "Operation $i succeeded"
                else
                    print_error "Operation $i failed"
                    fusermount3 -uz "$MOUNT_POINT"
                    kill $MOUNT_PID 2>/dev/null || true
                    return 1
                fi
                sleep 2
            done
            
            print_success "All operations succeeded with refreshed token"
            
            # Unmount
            fusermount3 -uz "$MOUNT_POINT"
            wait $MOUNT_PID 2>/dev/null || true
            
            return 0
        else
            print_error "Filesystem not mounted"
            kill $MOUNT_PID 2>/dev/null || true
            return 1
        fi
    else
        print_error "Failed to start OneMount"
        return 1
    fi
}

# Main test execution
main() {
    print_info "OneMount Token Refresh Test Suite"
    print_info "=================================="
    print_info ""
    
    # Setup
    setup
    
    # Track test results
    TESTS_RUN=0
    TESTS_PASSED=0
    TESTS_FAILED=0
    
    # Test 1: Verify token structure
    TESTS_RUN=$((TESTS_RUN + 1))
    if test_token_structure; then
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        TESTS_FAILED=$((TESTS_FAILED + 1))
        print_error "Token structure test failed, cannot continue"
        cleanup
        exit 1
    fi
    
    # Test 2: Force expiration and refresh
    TESTS_RUN=$((TESTS_RUN + 1))
    if test_force_expiration; then
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi
    
    # Test 3: Verify persistence
    TESTS_RUN=$((TESTS_RUN + 1))
    if test_token_persistence; then
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi
    
    # Test 4: Automatic refresh
    TESTS_RUN=$((TESTS_RUN + 1))
    if test_automatic_refresh; then
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi
    
    # Cleanup
    cleanup
    
    # Print summary
    echo ""
    echo "=========================================="
    echo "Test Summary"
    echo "=========================================="
    echo "Tests Run:    $TESTS_RUN"
    echo "Tests Passed: $TESTS_PASSED"
    echo "Tests Failed: $TESTS_FAILED"
    echo "=========================================="
    
    if [[ $TESTS_FAILED -eq 0 ]]; then
        print_success "All tests passed!"
        return 0
    else
        print_error "Some tests failed"
        return 1
    fi
}

# Run main function
main "$@"
