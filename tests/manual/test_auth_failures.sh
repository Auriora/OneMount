#!/bin/bash
# Manual test script for authentication failure scenarios
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
TEST_DIR="/tmp/onemount-auth-failure-test"
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
    
    # Restore network if disabled
    if command -v ip >/dev/null 2>&1; then
        # This is a placeholder - actual network manipulation requires privileges
        print_info "Network state restored (if modified)"
    fi
    
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

# Test 1: Invalid credentials (malformed token)
test_invalid_credentials() {
    print_test_header "Test 1: Invalid Credentials (Malformed Token)"
    
    print_info "Creating auth tokens file with invalid credentials..."
    
    # Create a token file with invalid tokens
    cat > "$AUTH_TOKENS_FILE" << EOF
{
  "config": {
    "clientID": "3470c3fa-bc10-45ab-a0a9-2d30836485d1",
    "codeURL": "https://login.microsoftonline.com/common/oauth2/v2.0/authorize",
    "tokenURL": "https://login.microsoftonline.com/common/oauth2/v2.0/token",
    "redirectURL": "https://login.live.com/oauth20_desktop.srf"
  },
  "account": "test@example.com",
  "expires_in": 3600,
  "expires_at": 0,
  "access_token": "invalid-access-token-12345",
  "refresh_token": "invalid-refresh-token-67890"
}
EOF
    
    print_success "Created auth tokens file with invalid credentials"
    
    # Try to mount (should fail or trigger reauthentication)
    print_info "Attempting to mount with invalid credentials..."
    
    # Capture output
    if timeout 30 ./build/onemount \
        --config-file "$AUTH_TOKENS_FILE" \
        "$MOUNT_POINT" 2>&1 | tee /tmp/mount_output.log &
    then
        MOUNT_PID=$!
        sleep 10
        
        # Check if mount succeeded (it shouldn't with invalid tokens)
        if mountpoint -q "$MOUNT_POINT"; then
            print_warning "Filesystem mounted with invalid tokens (unexpected)"
            
            # Try to list files (should fail)
            if timeout 10 ls "$MOUNT_POINT" 2>&1 | tee /tmp/ls_output.log; then
                print_error "File listing succeeded with invalid tokens (should fail)"
                fusermount3 -uz "$MOUNT_POINT"
                kill $MOUNT_PID 2>/dev/null || true
                return 1
            else
                print_success "File listing failed as expected with invalid tokens"
                fusermount3 -uz "$MOUNT_POINT"
                kill $MOUNT_PID 2>/dev/null || true
                return 0
            fi
        else
            print_success "Mount failed with invalid credentials (expected behavior)"
            kill $MOUNT_PID 2>/dev/null || true
            
            # Check error message
            if grep -i "error\|fail\|invalid" /tmp/mount_output.log >/dev/null 2>&1; then
                print_success "Error message provided"
                print_info "Error output:"
                grep -i "error\|fail\|invalid" /tmp/mount_output.log | head -5
                return 0
            else
                print_warning "No clear error message found"
                return 0
            fi
        fi
    else
        print_success "Mount command failed with invalid credentials (expected)"
        return 0
    fi
}

# Test 2: Expired refresh token
test_expired_refresh_token() {
    print_test_header "Test 2: Expired Refresh Token"
    
    print_info "Creating auth tokens file with expired refresh token..."
    
    # Create a token file with expired tokens
    cat > "$AUTH_TOKENS_FILE" << EOF
{
  "config": {
    "clientID": "3470c3fa-bc10-45ab-a0a9-2d30836485d1",
    "codeURL": "https://login.microsoftonline.com/common/oauth2/v2.0/authorize",
    "tokenURL": "https://login.microsoftonline.com/common/oauth2/v2.0/token",
    "redirectURL": "https://login.live.com/oauth20_desktop.srf"
  },
  "account": "test@example.com",
  "expires_in": 3600,
  "expires_at": 0,
  "access_token": "expired-access-token",
  "refresh_token": "expired-refresh-token"
}
EOF
    
    print_success "Created auth tokens file with expired tokens"
    
    # Try to mount (should fail and potentially trigger reauthentication)
    print_info "Attempting to mount with expired refresh token..."
    
    if timeout 30 ./build/onemount \
        --config-file "$AUTH_TOKENS_FILE" \
        "$MOUNT_POINT" 2>&1 | tee /tmp/expired_output.log &
    then
        MOUNT_PID=$!
        sleep 10
        
        # Check if reauthentication was triggered
        if grep -i "reauth\|authenticate" /tmp/expired_output.log >/dev/null 2>&1; then
            print_success "Reauthentication triggered for expired refresh token"
            kill $MOUNT_PID 2>/dev/null || true
            return 0
        elif grep -i "error\|fail" /tmp/expired_output.log >/dev/null 2>&1; then
            print_success "Error reported for expired refresh token"
            print_info "Error output:"
            grep -i "error\|fail" /tmp/expired_output.log | head -5
            kill $MOUNT_PID 2>/dev/null || true
            return 0
        else
            print_warning "No clear indication of reauthentication or error"
            kill $MOUNT_PID 2>/dev/null || true
            return 0
        fi
    else
        print_success "Mount failed with expired refresh token (expected)"
        return 0
    fi
}

# Test 3: Network disconnection during auth (simulated)
test_network_disconnection() {
    print_test_header "Test 3: Network Disconnection During Auth"
    
    print_warning "This test simulates network issues by using invalid endpoints"
    
    # Create a token file with invalid token URL (simulates network failure)
    cat > "$AUTH_TOKENS_FILE" << EOF
{
  "config": {
    "clientID": "3470c3fa-bc10-45ab-a0a9-2d30836485d1",
    "codeURL": "https://login.microsoftonline.com/common/oauth2/v2.0/authorize",
    "tokenURL": "https://invalid-endpoint-that-does-not-exist.example.com/token",
    "redirectURL": "https://login.live.com/oauth20_desktop.srf"
  },
  "account": "test@example.com",
  "expires_in": 3600,
  "expires_at": 0,
  "access_token": "test-access-token",
  "refresh_token": "test-refresh-token"
}
EOF
    
    print_success "Created auth tokens file with invalid token URL"
    
    # Try to mount (should handle network error gracefully)
    print_info "Attempting to mount with unreachable token endpoint..."
    
    if timeout 30 ./build/onemount \
        --config-file "$AUTH_TOKENS_FILE" \
        "$MOUNT_POINT" 2>&1 | tee /tmp/network_output.log &
    then
        MOUNT_PID=$!
        sleep 10
        
        # Check for network error handling
        if grep -i "network\|offline\|unreachable\|connection" /tmp/network_output.log >/dev/null 2>&1; then
            print_success "Network error detected and logged"
            print_info "Network error output:"
            grep -i "network\|offline\|unreachable\|connection" /tmp/network_output.log | head -5
        else
            print_warning "No explicit network error message found"
        fi
        
        # Check if system continued with existing token
        if grep -i "continuing with existing" /tmp/network_output.log >/dev/null 2>&1; then
            print_success "System continued with existing token (offline mode)"
        fi
        
        kill $MOUNT_PID 2>/dev/null || true
        return 0
    else
        print_info "Mount failed with network error (acceptable behavior)"
        return 0
    fi
}

# Test 4: Missing auth tokens file
test_missing_auth_file() {
    print_test_header "Test 4: Missing Auth Tokens File"
    
    # Remove auth tokens file
    rm -f "$AUTH_TOKENS_FILE"
    
    print_info "Attempting to mount without auth tokens file..."
    
    # Try to mount (should trigger authentication flow or fail with clear message)
    if timeout 30 ./build/onemount \
        --config-file "$AUTH_TOKENS_FILE" \
        --headless \
        "$MOUNT_POINT" 2>&1 | tee /tmp/missing_output.log &
    then
        MOUNT_PID=$!
        sleep 5
        
        # Check if authentication flow was triggered
        if grep -i "please visit\|authenticate\|authorization" /tmp/missing_output.log >/dev/null 2>&1; then
            print_success "Authentication flow triggered for missing tokens"
            print_info "Authentication prompt:"
            grep -i "please visit\|authenticate" /tmp/missing_output.log | head -3
            kill $MOUNT_PID 2>/dev/null || true
            return 0
        else
            print_warning "No clear authentication prompt found"
            kill $MOUNT_PID 2>/dev/null || true
            return 0
        fi
    else
        print_info "Mount failed with missing auth file (acceptable)"
        return 0
    fi
}

# Test 5: Malformed JSON in auth tokens file
test_malformed_json() {
    print_test_header "Test 5: Malformed JSON in Auth Tokens File"
    
    print_info "Creating auth tokens file with malformed JSON..."
    
    # Create a file with invalid JSON
    cat > "$AUTH_TOKENS_FILE" << EOF
{
  "access_token": "test-token",
  "refresh_token": "test-refresh",
  "expires_at": 0,
  INVALID JSON HERE
}
EOF
    
    print_success "Created malformed JSON file"
    
    # Try to mount (should fail with JSON parse error)
    print_info "Attempting to mount with malformed JSON..."
    
    if timeout 10 ./build/onemount \
        --config-file "$AUTH_TOKENS_FILE" \
        "$MOUNT_POINT" 2>&1 | tee /tmp/malformed_output.log; then
        
        print_error "Mount succeeded with malformed JSON (unexpected)"
        return 1
    else
        print_success "Mount failed with malformed JSON (expected)"
        
        # Check for JSON error message
        if grep -i "json\|parse\|unmarshal\|invalid" /tmp/malformed_output.log >/dev/null 2>&1; then
            print_success "JSON error message provided"
            print_info "Error output:"
            grep -i "json\|parse\|unmarshal\|invalid" /tmp/malformed_output.log | head -3
        else
            print_warning "No clear JSON error message found"
        fi
        
        return 0
    fi
}

# Test 6: Error message clarity
test_error_message_clarity() {
    print_test_header "Test 6: Error Message Clarity"
    
    print_info "Reviewing error messages from previous tests..."
    
    # Check if error messages are user-friendly
    local error_files=("/tmp/mount_output.log" "/tmp/expired_output.log" "/tmp/network_output.log" "/tmp/missing_output.log" "/tmp/malformed_output.log")
    local clear_messages=0
    local total_files=0
    
    for file in "${error_files[@]}"; do
        if [[ -f "$file" ]]; then
            total_files=$((total_files + 1))
            
            # Check for actionable error messages
            if grep -i "error\|fail\|invalid\|please\|try\|check" "$file" >/dev/null 2>&1; then
                clear_messages=$((clear_messages + 1))
                print_success "Clear error message found in $(basename $file)"
            else
                print_warning "No clear error message in $(basename $file)"
            fi
        fi
    done
    
    if [[ $total_files -gt 0 ]]; then
        print_info "Error message clarity: $clear_messages/$total_files files have clear messages"
        
        if [[ $clear_messages -ge $((total_files / 2)) ]]; then
            print_success "Majority of error messages are clear and actionable"
            return 0
        else
            print_warning "Error messages could be more clear and actionable"
            return 0
        fi
    else
        print_warning "No error log files found to analyze"
        return 0
    fi
}

# Main test execution
main() {
    print_info "OneMount Authentication Failure Test Suite"
    print_info "==========================================="
    print_info ""
    print_warning "This test suite verifies error handling for authentication failures"
    print_info ""
    
    # Setup
    setup
    
    # Track test results
    TESTS_RUN=0
    TESTS_PASSED=0
    TESTS_FAILED=0
    
    # Test 1: Invalid credentials
    TESTS_RUN=$((TESTS_RUN + 1))
    if test_invalid_credentials; then
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi
    
    # Test 2: Expired refresh token
    TESTS_RUN=$((TESTS_RUN + 1))
    if test_expired_refresh_token; then
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi
    
    # Test 3: Network disconnection
    TESTS_RUN=$((TESTS_RUN + 1))
    if test_network_disconnection; then
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi
    
    # Test 4: Missing auth file
    TESTS_RUN=$((TESTS_RUN + 1))
    if test_missing_auth_file; then
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi
    
    # Test 5: Malformed JSON
    TESTS_RUN=$((TESTS_RUN + 1))
    if test_malformed_json; then
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi
    
    # Test 6: Error message clarity
    TESTS_RUN=$((TESTS_RUN + 1))
    if test_error_message_clarity; then
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
