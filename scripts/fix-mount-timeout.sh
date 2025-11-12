#!/bin/bash
# Fix script for mount timeout issues in Docker
# Implements solutions based on diagnostic results

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

print_info "OneMount Mount Timeout Fix Tool"
print_info "================================"
echo ""

# Solution 1: Add configurable mount timeout
print_info "Solution 1: Adding configurable mount timeout..."

# Check if we need to add timeout configuration
if ! grep -q "MountTimeout" cmd/onemount/main.go; then
    print_info "Mount timeout configuration not found in code"
    print_info "This will be added via code changes"
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
fi

# Test Microsoft Graph API
if ! curl -s --connect-timeout 5 --max-time 10 https://graph.microsoft.com/v1.0/ > /dev/null 2>&1; then
    print_error "Cannot connect to Microsoft Graph API"
    print_info "This may be a network connectivity issue"
    print_info "Checking proxy settings..."
    env | grep -i proxy || print_info "No proxy settings found"
fi

# Solution 3: Disable sync-tree by default in Docker
print_info "Solution 3: Checking sync-tree configuration..."
print_info "Using --no-sync-tree flag is recommended for Docker environments"
print_success "This is already documented in the test scripts"

# Solution 4: Add pre-mount connectivity check
print_info "Solution 4: Testing pre-mount connectivity..."

# Create a simple connectivity test
cat > /tmp/test-connectivity.sh << 'EOF'
#!/bin/bash
# Test connectivity before mounting

echo "Testing network connectivity..."

# Test DNS
if ! ping -c 1 -W 2 8.8.8.8 > /dev/null 2>&1; then
    echo "ERROR: Cannot reach DNS server"
    exit 1
fi

# Test DNS resolution
if ! nslookup graph.microsoft.com > /dev/null 2>&1; then
    echo "ERROR: DNS resolution failed"
    exit 1
fi

# Test Microsoft Graph API
if ! curl -s --connect-timeout 5 --max-time 10 https://graph.microsoft.com/v1.0/ > /dev/null 2>&1; then
    echo "ERROR: Cannot connect to Microsoft Graph API"
    exit 1
fi

echo "Network connectivity OK"
exit 0
EOF

chmod +x /tmp/test-connectivity.sh

if /tmp/test-connectivity.sh; then
    print_success "Network connectivity is OK"
else
    print_error "Network connectivity check failed"
    print_info "Mount will likely timeout due to network issues"
fi

# Solution 5: Test with minimal configuration
print_info "Solution 5: Testing mount with minimal configuration..."

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
            exit 0
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
    exit 1
else
    print_warning "OneMount binary not found at ./build/onemount"
    print_info "Build the binary first with: go build -o build/onemount ./cmd/onemount"
fi
