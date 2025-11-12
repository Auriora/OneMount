#!/bin/bash
# Debug script for investigating mount timeout issues in Docker
# This script helps diagnose network connectivity and mount issues

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

# 11. Test a simple mount with timeout
print_info "11. Testing mount with extended logging..."
if [ -f "./build/onemount" ]; then
    print_info "OneMount binary found, attempting test mount..."
    print_info "This will timeout after 60 seconds if it hangs..."
    
    # Create mount point and cache if they don't exist
    mkdir -p /tmp/mount /tmp/cache
    
    # Run with timeout and capture output
    timeout 60s ./build/onemount \
        --cache-dir=/tmp/cache \
        --log=debug \
        --no-sync-tree \
        /tmp/mount 2>&1 | tee /tmp/mount-debug.log &
    
    MOUNT_PID=$!
    
    # Wait a bit and check if mount succeeded
    sleep 10
    
    if mountpoint -q /tmp/mount; then
        print_success "Mount succeeded!"
        # Unmount
        fusermount3 -uz /tmp/mount || true
    else
        print_warning "Mount did not complete within 10 seconds"
        print_info "Waiting for timeout or completion..."
        wait $MOUNT_PID || true
        
        print_info "Mount debug log:"
        cat /tmp/mount-debug.log
    fi
else
    print_warning "OneMount binary not found at ./build/onemount"
fi

echo ""
print_info "Diagnostic complete. Check output above for issues."
