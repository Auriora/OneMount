#!/bin/bash
# Manual test script for interactive authentication flow
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
TEST_DIR="/tmp/onemount-auth-test"
AUTH_TOKENS_FILE="$TEST_DIR/auth_tokens.json"
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
    
    print_success "Test environment ready"
}

# Test 1: Check if GTK is available
test_gtk_availability() {
    print_test_header "Test 1: Check GTK Availability"
    
    if pkg-config --exists webkit2gtk-4.1; then
        print_success "WebKit2GTK 4.1 is available"
        
        # Check if DISPLAY is set
        if [[ -z "$DISPLAY" ]]; then
            print_warning "DISPLAY environment variable not set"
            print_warning "GTK authentication will not work without X11 display"
            print_info "For headless testing, use headless mode instead"
            return 1
        else
            print_success "DISPLAY is set to: $DISPLAY"
            return 0
        fi
    else
        print_warning "WebKit2GTK 4.1 not available"
        print_info "Will use headless authentication mode"
        return 1
    fi
}

# Test 2: Interactive authentication (GTK)
test_interactive_auth_gtk() {
    print_test_header "Test 2: Interactive Authentication (GTK)"
    
    # Check if GTK is available
    if ! test_gtk_availability; then
        print_warning "Skipping GTK authentication test - GTK not available"
        return 0
    fi
    
    print_info "Starting OneMount with GTK authentication..."
    print_info "A browser window should open for Microsoft login"
    print_info ""
    print_warning "IMPORTANT: Use a TEST OneDrive account, NOT production!"
    print_info ""
    
    # Remove existing auth tokens
    rm -f "$AUTH_TOKENS_FILE"
    
    # Try to authenticate (this will open GTK window)
    if timeout 300 ./build/onemount \
        --auth-only \
        --config-file "$AUTH_TOKENS_FILE" \
        "$MOUNT_POINT"; then
        
        print_success "Authentication completed"
        
        # Verify tokens were created
        if [[ -f "$AUTH_TOKENS_FILE" ]]; then
            print_success "Auth tokens file created: $AUTH_TOKENS_FILE"
            
            # Check file permissions
            PERMS=$(stat -c "%a" "$AUTH_TOKENS_FILE")
            if [[ "$PERMS" == "600" ]]; then
                print_success "File permissions are correct: $PERMS"
            else
                print_error "File permissions are incorrect: $PERMS (expected 600)"
                return 1
            fi
            
            # Verify token structure
            if command -v jq >/dev/null 2>&1; then
                print_info "Verifying token structure..."
                
                ACCESS_TOKEN=$(jq -r '.access_token' "$AUTH_TOKENS_FILE")
                REFRESH_TOKEN=$(jq -r '.refresh_token' "$AUTH_TOKENS_FILE")
                EXPIRES_AT=$(jq -r '.expires_at' "$AUTH_TOKENS_FILE")
                
                if [[ -n "$ACCESS_TOKEN" ]] && [[ "$ACCESS_TOKEN" != "null" ]]; then
                    print_success "AccessToken present"
                else
                    print_error "AccessToken missing or null"
                    return 1
                fi
                
                if [[ -n "$REFRESH_TOKEN" ]] && [[ "$REFRESH_TOKEN" != "null" ]]; then
                    print_success "RefreshToken present"
                else
                    print_error "RefreshToken missing or null"
                    return 1
                fi
                
                if [[ -n "$EXPIRES_AT" ]] && [[ "$EXPIRES_AT" != "null" ]]; then
                    print_success "ExpiresAt present: $EXPIRES_AT"
                    
                    # Check if token is not expired
                    CURRENT_TIME=$(date +%s)
                    if [[ "$EXPIRES_AT" -gt "$CURRENT_TIME" ]]; then
                        print_success "Token is not expired"
                    else
                        print_warning "Token appears to be expired"
                    fi
                else
                    print_error "ExpiresAt missing or null"
                    return 1
                fi
            else
                print_warning "jq not available, skipping token structure verification"
            fi
            
            return 0
        else
            print_error "Auth tokens file not created"
            return 1
        fi
    else
        print_error "Authentication failed or timed out"
        return 1
    fi
}

# Test 3: Headless authentication
test_headless_auth() {
    print_test_header "Test 3: Headless Authentication"
    
    print_info "Starting OneMount with headless authentication..."
    print_info "You will be prompted to visit a URL and enter the redirect URL"
    print_info ""
    print_warning "IMPORTANT: Use a TEST OneDrive account, NOT production!"
    print_info ""
    
    # Remove existing auth tokens
    rm -f "$AUTH_TOKENS_FILE"
    
    # Try to authenticate in headless mode
    print_info "Running: ./build/onemount --auth-only --headless --config-file $AUTH_TOKENS_FILE $MOUNT_POINT"
    print_info ""
    
    if ./build/onemount \
        --auth-only \
        --headless \
        --config-file "$AUTH_TOKENS_FILE" \
        "$MOUNT_POINT"; then
        
        print_success "Headless authentication completed"
        
        # Verify tokens were created
        if [[ -f "$AUTH_TOKENS_FILE" ]]; then
            print_success "Auth tokens file created"
            
            # Verify token structure (same as GTK test)
            if command -v jq >/dev/null 2>&1; then
                ACCESS_TOKEN=$(jq -r '.access_token' "$AUTH_TOKENS_FILE")
                REFRESH_TOKEN=$(jq -r '.refresh_token' "$AUTH_TOKENS_FILE")
                EXPIRES_AT=$(jq -r '.expires_at' "$AUTH_TOKENS_FILE")
                
                if [[ -n "$ACCESS_TOKEN" ]] && [[ "$ACCESS_TOKEN" != "null" ]] && \
                   [[ -n "$REFRESH_TOKEN" ]] && [[ "$REFRESH_TOKEN" != "null" ]] && \
                   [[ -n "$EXPIRES_AT" ]] && [[ "$EXPIRES_AT" != "null" ]]; then
                    print_success "All required token fields present"
                    return 0
                else
                    print_error "Token structure incomplete"
                    return 1
                fi
            fi
            
            return 0
        else
            print_error "Auth tokens file not created"
            return 1
        fi
    else
        print_error "Headless authentication failed"
        return 1
    fi
}

# Test 4: Verify tokens can be loaded
test_load_tokens() {
    print_test_header "Test 4: Load Existing Tokens"
    
    if [[ ! -f "$AUTH_TOKENS_FILE" ]]; then
        print_warning "No auth tokens file found, skipping load test"
        print_info "Run authentication tests first"
        return 0
    fi
    
    print_info "Attempting to mount with existing tokens..."
    
    # Try to mount (should use existing tokens)
    if timeout 30 ./build/onemount \
        --config-file "$AUTH_TOKENS_FILE" \
        "$MOUNT_POINT" &
    then
        MOUNT_PID=$!
        
        # Wait for mount to complete
        sleep 5
        
        # Check if mounted
        if mountpoint -q "$MOUNT_POINT"; then
            print_success "Filesystem mounted successfully with existing tokens"
            
            # Try to list files
            if ls "$MOUNT_POINT" >/dev/null 2>&1; then
                print_success "Can list files in mount point"
            else
                print_warning "Cannot list files in mount point"
            fi
            
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
    print_info "OneMount Interactive Authentication Test Suite"
    print_info "=============================================="
    print_info ""
    print_warning "This test requires manual interaction"
    print_warning "Use a TEST OneDrive account, NOT production!"
    print_info ""
    
    # Setup
    setup
    
    # Track test results
    TESTS_RUN=0
    TESTS_PASSED=0
    TESTS_FAILED=0
    TESTS_SKIPPED=0
    
    # Run tests
    if test_gtk_availability; then
        # GTK is available, run GTK test
        TESTS_RUN=$((TESTS_RUN + 1))
        if test_interactive_auth_gtk; then
            TESTS_PASSED=$((TESTS_PASSED + 1))
        else
            TESTS_FAILED=$((TESTS_FAILED + 1))
        fi
    else
        # GTK not available, run headless test
        TESTS_RUN=$((TESTS_RUN + 1))
        if test_headless_auth; then
            TESTS_PASSED=$((TESTS_PASSED + 1))
        else
            TESTS_FAILED=$((TESTS_FAILED + 1))
        fi
    fi
    
    # Test loading tokens
    TESTS_RUN=$((TESTS_RUN + 1))
    if test_load_tokens; then
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
    echo "Tests Run:     $TESTS_RUN"
    echo "Tests Passed:  $TESTS_PASSED"
    echo "Tests Failed:  $TESTS_FAILED"
    echo "Tests Skipped: $TESTS_SKIPPED"
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
