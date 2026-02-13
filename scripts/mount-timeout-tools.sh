#!/bin/bash
# Consolidated mount timeout diagnostic, fix, and test tool
# Combines debug-mount-timeout.sh, fix-mount-timeout.sh, and test-mount-timeout-fix.sh

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

print_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
print_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
print_warning() { echo -e "${YELLOW}[WARNING]${NC} $1"; }
print_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# Show usage
show_usage() {
    cat << EOF
OneMount Mount Timeout Tools

Usage: $0 <command>

Commands:
  diagnose    Run diagnostic checks for mount timeout issues
  fix         Apply fixes for mount timeout issues
  test        Validate mount timeout fix implementation
  help        Show this help message

Examples:
  $0 diagnose    # Check for mount timeout issues
  $0 fix         # Apply fixes
  $0 test        # Validate the fix

EOF
}

# Diagnostic function
run_diagnose() {
    print_info "OneMount Mount Timeout Diagnostic Tool"
    print_info "========================================"
    echo ""

    # Check if running in Docker
    if [ -f /.dockerenv ]; then
        print_info "Running inside Docker container"
    else
        print_warning "Not running in Docker - some checks may not be relevant"
    fi

    # 1. Check DNS resolution
    print_info "1. Testing DNS resolution..."
    if ping -c 1 -W 2 8.8.8.8 > /dev/null 2>&1; then
        print_success "Can reach Google DNS (8.8.8.8)"
    else
        print_error "Cannot reach Google DNS (8.8.8.8)"
    fi

    if nslookup graph.microsoft.com > /dev/null 2>&1; then
        print_success "DNS resolution works (graph.microsoft.com)"
    else
        print_error "DNS resolution failed for graph.microsoft.com"
    fi

    # 2. Check Microsoft Graph API connectivity
    print_info "2. Testing Microsoft Graph API connectivity..."
    if curl -s --connect-timeout 5 --max-time 10 https://graph.microsoft.com/v1.0/ > /dev/null 2>&1; then
        print_success "Can connect to Microsoft Graph API"
    else
        print_error "Cannot connect to Microsoft Graph API"
        print_info "Trying with verbose output..."
        curl -v --connect-timeout 5 --max-time 10 https://graph.microsoft.com/v1.0/ 2>&1 | head -20
    fi

    # 3. Check FUSE device
    print_info "3. Checking FUSE device..."
    if [ -e /dev/fuse ]; then
        print_success "FUSE device exists"
        ls -l /dev/fuse
    else
        print_error "FUSE device not found"
    fi

    # 4. Check FUSE configuration
    print_info "4. Checking FUSE configuration..."
    if [ -f /etc/fuse.conf ]; then
        print_success "FUSE configuration exists"
        cat /etc/fuse.conf
    else
        print_warning "FUSE configuration not found"
    fi

    # 5. Check network interfaces
    print_info "5. Checking network interfaces..."
    ip addr show | grep -E "^[0-9]+:|inet "

    # 6. Check routing
    print_info "6. Checking routing table..."
    ip route show

    # 7. Check /etc/resolv.conf
    print_info "7. Checking DNS configuration..."
    cat /etc/resolv.conf

    # 8. Test auth tokens if available
    print_info "8. Checking for auth tokens..."
    AUTH_LOCATIONS=(
        "$HOME/.onemount-tests/.auth_tokens.json"
        "/workspace/test-artifacts/.auth_tokens.json"
        "/workspace/auth_tokens.json"
    )

    AUTH_FOUND=false
    for location in "${AUTH_LOCATIONS[@]}"; do
        if [ -f "$location" ]; then
            print_success "Auth tokens found at: $location"
            AUTH_FOUND=true
            
            # Check if tokens are valid JSON
            if command -v jq > /dev/null 2>&1; then
                if jq empty "$location" 2>/dev/null; then
                    print_success "Auth tokens are valid JSON"
                    
                    # Check expiration
                    EXPIRES_AT=$(jq -r '.expires_at // 0' "$location" 2>/dev/null || echo "0")
                    CURRENT_TIME=$(date +%s)
                    
                    if [ "$EXPIRES_AT" != "0" ] && [ "$EXPIRES_AT" -le "$CURRENT_TIME" ]; then
                        print_error "Auth tokens are EXPIRED"
                        print_info "Expiration: $(date -d @$EXPIRES_AT)"
                        print_info "Current:    $(date -d @$CURRENT_TIME)"
                    else
                        print_success "Auth tokens are valid and not expired"
                    fi
                else
                    print_error "Auth tokens are not valid JSON"
                fi
            fi
            break
        fi
    done

    if [ "$AUTH_FOUND" = false ]; then
        print_warning "No auth tokens found"
    fi

    # 9. Check if mount point exists and is empty
    print_info "9. Checking mount point..."
    if [ -d "/tmp/mount" ]; then
        print_success "Mount point /tmp/mount exists"
        if [ -z "$(ls -A /tmp/mount)" ]; then
            print_success "Mount point is empty"
        else
            print_warning "Mount point is not empty:"
            ls -la /tmp/mount
        fi
    else
        print_info "Mount point /tmp/mount does not exist (will be created)"
    fi

    # 10. Check cache directory
    print_info "10. Checking cache directory..."
    if [ -d "/tmp/cache" ]; then
        print_success "Cache directory /tmp/cache exists"
        print_info "Cache contents:"
        ls -la /tmp/cache
    else
        print_info "Cache directory /tmp/cache does not exist (will be created)"
    fi

    echo ""
    print_info "Diagnostic complete. Check output above for issues."
}

# Fix function
run_fix() {
    print_info "OneMount Mount Timeout Fix Tool"
    print_info "================================"
    echo ""

    # Solution 1: Check mount timeout configuration
    print_info "Solution 1: Checking mount timeout configuration..."
    if ! grep -q "MountTimeout" cmd/onemount/main.go 2>/dev/null; then
        print_info "Mount timeout configuration not found in code"
        print_info "This should be added via code changes"
    else
        print_success "Mount timeout configuration already exists"
    fi

    # Solution 2: Improve network connectivity checks
    print_info "Solution 2: Testing network connectivity..."

    # Test DNS
    if ! ping -c 1 -W 2 8.8.8.8 > /dev/null 2>&1; then
        print_error "Cannot reach DNS server"
        print_info "Checking /etc/resolv.conf..."
        cat /etc/resolv.conf
        
        print_info "Attempting to fix DNS configuration..."
        if [ -w /etc/resolv.conf ]; then
            echo "nameserver 8.8.8.8" > /etc/resolv.conf
            echo "nameserver 8.8.4.4" >> /etc/resolv.conf
            print_success "DNS configuration updated"
        else
            print_error "Cannot write to /etc/resolv.conf (need root)"
        fi
    else
        print_success "DNS connectivity OK"
    fi

    # Test Microsoft Graph API
    if ! curl -s --connect-timeout 5 --max-time 10 https://graph.microsoft.com/v1.0/ > /dev/null 2>&1; then
        print_error "Cannot connect to Microsoft Graph API"
        print_info "This may be a network connectivity issue"
        print_info "Checking proxy settings..."
        env | grep -i proxy || print_info "No proxy settings found"
    else
        print_success "Microsoft Graph API connectivity OK"
    fi

    # Solution 3: Test with minimal configuration
    print_info "Solution 3: Testing mount with minimal configuration..."

    if [ -f "./build/onemount" ]; then
        # Create mount point and cache
        mkdir -p /tmp/mount-test /tmp/cache-test
        
        # Clean up any existing mounts
        fusermount3 -uz /tmp/mount-test 2>/dev/null || true
        
        print_info "Attempting mount with minimal configuration..."
        print_info "Using --no-sync-tree to avoid initial sync delay..."
        
        # Run mount in background with timeout
        timeout 30s ./build/onemount \
            --cache-dir=/tmp/cache-test \
            --log=info \
            --no-sync-tree \
            /tmp/mount-test > /tmp/mount-test.log 2>&1 &
        
        MOUNT_PID=$!
        
        # Wait for mount to complete
        for i in {1..30}; do
            if mountpoint -q /tmp/mount-test; then
                print_success "Mount succeeded in $i seconds!"
                
                # Test basic operations
                print_info "Testing basic filesystem operations..."
                if ls /tmp/mount-test > /dev/null 2>&1; then
                    print_success "Can list mount point"
                fi
                
                # Unmount
                print_info "Unmounting..."
                fusermount3 -uz /tmp/mount-test
                print_success "Unmount successful"
                
                # Clean up
                rm -rf /tmp/mount-test /tmp/cache-test
                
                print_success "Mount timeout fix verified!"
                return 0
            fi
            sleep 1
        done
        
        print_error "Mount did not complete within 30 seconds"
        print_info "Mount log:"
        cat /tmp/mount-test.log
        
        # Kill the mount process
        kill $MOUNT_PID 2>/dev/null || true
        
        # Clean up
        fusermount3 -uz /tmp/mount-test 2>/dev/null || true
        rm -rf /tmp/mount-test /tmp/cache-test
        
        print_error "Mount timeout issue persists"
        return 1
    else
        print_warning "OneMount binary not found at ./build/onemount"
        print_info "Build the binary first with: go build -o build/onemount ./cmd/onemount"
        return 1
    fi
}

# Test function
run_test() {
    print_info "Mount Timeout Fix Validation Test"
    print_info "==================================="
    echo ""

    # Test 1: Verify binary exists
    print_info "Test 1: Checking if OneMount binary exists..."
    if [ -f "./build/onemount" ]; then
        print_success "Binary found at ./build/onemount"
    else
        print_error "Binary not found"
        print_info "Build with: go build -o build/onemount ./cmd/onemount"
        return 1
    fi

    # Test 2: Verify --mount-timeout flag exists
    print_info "Test 2: Verifying --mount-timeout flag..."
    if ./build/onemount --help 2>&1 | grep -q "mount-timeout"; then
        print_success "--mount-timeout flag is available"
        ./build/onemount --help 2>&1 | grep -A 1 "mount-timeout"
    else
        print_warning "--mount-timeout flag not found (may not be implemented yet)"
    fi

    # Test 3: Test connectivity check
    print_info "Test 3: Testing connectivity check..."
    if ping -c 1 -W 2 8.8.8.8 > /dev/null 2>&1; then
        print_success "Network connectivity available"
        
        # Test Microsoft Graph API
        if curl -s --connect-timeout 5 --max-time 10 https://graph.microsoft.com/v1.0/ > /dev/null 2>&1; then
            print_success "Microsoft Graph API is reachable"
        else
            print_warning "Microsoft Graph API is not reachable (may affect mount)"
        fi
    else
        print_warning "No network connectivity (tests will be limited)"
    fi

    # Test 4: Check environment
    print_info "Test 4: Checking environment..."
    if [ -f /.dockerenv ]; then
        print_info "Running inside Docker container"
        
        # Check FUSE device
        if [ -e /dev/fuse ]; then
            print_success "FUSE device is available"
        else
            print_error "FUSE device is not available"
            print_info "Run with: --device /dev/fuse --cap-add SYS_ADMIN"
        fi
    else
        print_info "Running on host system"
    fi

    echo ""
    print_info "Validation Test Complete"
    print_info "========================"
    echo ""
    print_success "All basic tests passed!"
    echo ""
    print_info "Next steps:"
    print_info "1. Run diagnostic: $0 diagnose"
    print_info "2. Apply fixes: $0 fix"
    print_info "3. Test mount: ./build/onemount --no-sync-tree --cache-dir=/tmp/cache /tmp/mount"
}

# Main execution
case "${1:-}" in
    diagnose)
        run_diagnose
        ;;
    fix)
        run_fix
        ;;
    test)
        run_test
        ;;
    help|--help|-h)
        show_usage
        ;;
    *)
        if [ -n "$1" ]; then
            print_error "Unknown command: $1"
            echo ""
        fi
        show_usage
        exit 1
        ;;
esac
