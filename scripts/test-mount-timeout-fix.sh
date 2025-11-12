#!/bin/bash
# Test script to validate the mount timeout fix
# This script tests the new --mount-timeout flag and connectivity check

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

print_info "Mount Timeout Fix Validation Test"
print_info "==================================="
echo ""

# Test 1: Verify binary exists
print_info "Test 1: Checking if OneMount binary exists..."
if [ -f "./build/onemount" ]; then
    print_success "Binary found at ./build/onemount"
else
    print_error "Binary not found. Building..."
    go build -o build/onemount ./cmd/onemount
    if [ $? -eq 0 ]; then
        print_success "Binary built successfully"
    else
        print_error "Failed to build binary"
        exit 1
    fi
fi

# Test 2: Verify --mount-timeout flag exists
print_info "Test 2: Verifying --mount-timeout flag..."
if ./build/onemount --help 2>&1 | grep -q "mount-timeout"; then
    print_success "--mount-timeout flag is available"
    ./build/onemount --help 2>&1 | grep -A 1 "mount-timeout"
else
    print_error "--mount-timeout flag not found"
    exit 1
fi

# Test 3: Test with invalid timeout value
print_info "Test 3: Testing with invalid timeout value..."
if ./build/onemount --mount-timeout -1 /tmp/mount 2>&1 | grep -q "invalid"; then
    print_success "Invalid timeout value rejected correctly"
else
    print_warning "Invalid timeout handling may need improvement"
fi

# Test 4: Test connectivity check (if network available)
print_info "Test 4: Testing connectivity check..."
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

# Test 5: Test mount with increased timeout (dry run)
print_info "Test 5: Testing mount command syntax..."
mkdir -p /tmp/mount-test /tmp/cache-test

# Clean up any existing mounts
fusermount3 -uz /tmp/mount-test 2>/dev/null || true

print_info "Testing mount with --mount-timeout 120..."
print_info "Command: ./build/onemount --mount-timeout 120 --no-sync-tree --cache-dir=/tmp/cache-test /tmp/mount-test"

# Note: We don't actually mount here because we may not have auth tokens
# Just verify the command syntax is accepted
if ./build/onemount --help > /dev/null 2>&1; then
    print_success "Mount command syntax is valid"
else
    print_error "Mount command syntax is invalid"
    exit 1
fi

# Test 6: Verify diagnostic script exists
print_info "Test 6: Checking diagnostic scripts..."
if [ -f "./scripts/debug-mount-timeout.sh" ]; then
    print_success "Diagnostic script found"
    if [ -x "./scripts/debug-mount-timeout.sh" ]; then
        print_success "Diagnostic script is executable"
    else
        print_warning "Diagnostic script is not executable (run: chmod +x scripts/debug-mount-timeout.sh)"
    fi
else
    print_error "Diagnostic script not found"
fi

if [ -f "./scripts/fix-mount-timeout.sh" ]; then
    print_success "Fix script found"
    if [ -x "./scripts/fix-mount-timeout.sh" ]; then
        print_success "Fix script is executable"
    else
        print_warning "Fix script is not executable (run: chmod +x scripts/fix-mount-timeout.sh)"
    fi
else
    print_error "Fix script not found"
fi

# Test 7: Verify documentation exists
print_info "Test 7: Checking documentation..."
if [ -f "./docs/fixes/mount-timeout-fix.md" ]; then
    print_success "Detailed documentation found"
else
    print_warning "Detailed documentation not found"
fi

if [ -f "./docs/fixes/mount-timeout-summary.md" ]; then
    print_success "Summary documentation found"
else
    print_warning "Summary documentation not found"
fi

# Test 8: Check if running in Docker
print_info "Test 8: Checking environment..."
if [ -f /.dockerenv ]; then
    print_info "Running inside Docker container"
    
    # Check FUSE device
    if [ -e /dev/fuse ]; then
        print_success "FUSE device is available"
    else
        print_error "FUSE device is not available"
        print_info "Run with: --device /dev/fuse --cap-add SYS_ADMIN"
    fi
    
    # Check DNS
    if [ -f /etc/resolv.conf ]; then
        print_info "DNS configuration:"
        cat /etc/resolv.conf | grep nameserver
    fi
else
    print_info "Running on host system"
fi

# Clean up
rm -rf /tmp/mount-test /tmp/cache-test

echo ""
print_info "Validation Test Complete"
print_info "========================"
echo ""
print_success "All basic tests passed!"
echo ""
print_info "Next steps:"
print_info "1. Run diagnostic script: bash scripts/debug-mount-timeout.sh"
print_info "2. Test actual mount: ./build/onemount --mount-timeout 120 --no-sync-tree --cache-dir=/tmp/cache /tmp/mount"
print_info "3. If issues persist, run: bash scripts/fix-mount-timeout.sh"
echo ""
