#!/bin/bash
# Task 5.4: Test Filesystem Operations While Mounted
# Tests basic filesystem operations after successful mount

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

# Test configuration
MOUNT_POINT="/tmp/onemount-test-mount"
CACHE_DIR="/tmp/onemount-test-cache"
MOUNT_TIMEOUT=120
TEST_TIMEOUT=30

# Cleanup function
cleanup() {
    print_info "Cleaning up..."
    
    # Unmount if mounted
    if mountpoint -q "$MOUNT_POINT" 2>/dev/null; then
        print_info "Unmounting filesystem..."
        fusermount3 -uz "$MOUNT_POINT" 2>/dev/null || true
        sleep 1
    fi
    
    # Kill any remaining onemount processes
    pkill -f "onemount.*$MOUNT_POINT" 2>/dev/null || true
    
    # Clean up directories
    rm -rf "$MOUNT_POINT" "$CACHE_DIR"
    
    print_info "Cleanup complete"
}

# Set up trap for cleanup
trap cleanup EXIT INT TERM

print_info "Task 5.4: Test Filesystem Operations While Mounted"
print_info "==================================================="
echo ""

# Check prerequisites
print_info "Checking prerequisites..."

# Check if binary exists
if [ ! -f "./build/onemount" ]; then
    print_error "OneMount binary not found. Building..."
    go build -o build/onemount ./cmd/onemount
    if [ $? -ne 0 ]; then
        print_error "Failed to build binary"
        exit 1
    fi
    print_success "Binary built successfully"
fi

# Check for auth tokens
AUTH_TOKEN_FOUND=false
AUTH_LOCATIONS=(
    "$HOME/.onemount-tests/.auth_tokens.json"
    "./test-artifacts/.auth_tokens.json"
    "./auth_tokens.json"
)

for location in "${AUTH_LOCATIONS[@]}"; do
    if [ -f "$location" ]; then
        print_success "Auth tokens found at: $location"
        AUTH_TOKEN_FOUND=true
        
        # Copy to expected location if needed
        if [ "$location" != "$HOME/.onemount-tests/.auth_tokens.json" ]; then
            mkdir -p "$HOME/.onemount-tests"
            cp "$location" "$HOME/.onemount-tests/.auth_tokens.json"
            chmod 600 "$HOME/.onemount-tests/.auth_tokens.json"
            print_info "Copied auth tokens to expected location"
        fi
        break
    fi
done

if [ "$AUTH_TOKEN_FOUND" = false ]; then
    print_error "No auth tokens found. Cannot test with real OneDrive."
    print_info "This test requires OneDrive authentication."
    print_info "Please run authentication first or provide auth tokens."
    exit 1
fi

# Check FUSE availability
if [ ! -e /dev/fuse ]; then
    print_error "FUSE device not available"
    print_info "This test requires FUSE support"
    exit 1
fi

print_success "All prerequisites met"
echo ""

# Create mount point and cache directory
print_info "Setting up test environment..."
mkdir -p "$MOUNT_POINT" "$CACHE_DIR"
print_success "Test directories created"
echo ""

# Start mount process
print_info "Starting mount process..."
print_info "Command: ./build/onemount --mount-timeout $MOUNT_TIMEOUT --no-sync-tree --log=info --cache-dir=$CACHE_DIR $MOUNT_POINT"

# Start mount in background
./build/onemount \
    --mount-timeout "$MOUNT_TIMEOUT" \
    --no-sync-tree \
    --log=info \
    --cache-dir="$CACHE_DIR" \
    "$MOUNT_POINT" > /tmp/onemount-test.log 2>&1 &

MOUNT_PID=$!
print_info "Mount process started (PID: $MOUNT_PID)"

# Wait for mount to complete
print_info "Waiting for mount to complete (timeout: ${MOUNT_TIMEOUT}s)..."
WAIT_COUNT=0
MOUNT_SUCCESS=false

while [ $WAIT_COUNT -lt $MOUNT_TIMEOUT ]; do
    if mountpoint -q "$MOUNT_POINT" 2>/dev/null; then
        MOUNT_SUCCESS=true
        break
    fi
    
    # Check if process is still running
    if ! kill -0 $MOUNT_PID 2>/dev/null; then
        print_error "Mount process died unexpectedly"
        print_info "Log output:"
        cat /tmp/onemount-test.log
        exit 1
    fi
    
    sleep 1
    WAIT_COUNT=$((WAIT_COUNT + 1))
    
    # Show progress every 10 seconds
    if [ $((WAIT_COUNT % 10)) -eq 0 ]; then
        print_info "Still waiting... (${WAIT_COUNT}s elapsed)"
    fi
done

if [ "$MOUNT_SUCCESS" = false ]; then
    print_error "Mount did not complete within ${MOUNT_TIMEOUT} seconds"
    print_info "Log output:"
    cat /tmp/onemount-test.log
    exit 1
fi

print_success "Mount completed successfully in ${WAIT_COUNT} seconds!"
echo ""

# Give filesystem a moment to stabilize
sleep 2

# Test 1: List Directory (ls)
print_info "Test 1: List Directory (ls)"
print_info "----------------------------"

START_TIME=$(date +%s)
if timeout $TEST_TIMEOUT ls -la "$MOUNT_POINT" > /tmp/test-ls.txt 2>&1; then
    END_TIME=$(date +%s)
    ELAPSED=$((END_TIME - START_TIME))
    
    print_success "Directory listing succeeded (${ELAPSED}s)"
    
    # Show first few entries
    print_info "First 10 entries:"
    head -10 /tmp/test-ls.txt | sed 's/^/  /'
    
    # Check if directory is empty
    ENTRY_COUNT=$(ls -A "$MOUNT_POINT" | wc -l)
    if [ $ENTRY_COUNT -eq 0 ]; then
        print_warning "Mount point appears empty (OneDrive may be empty)"
    else
        print_success "Found $ENTRY_COUNT entries in mount point"
    fi
else
    print_error "Directory listing failed or timed out"
    cat /tmp/test-ls.txt
fi
echo ""

# Test 2: Stat Operations
print_info "Test 2: Stat Operations"
print_info "------------------------"

# Get first file/directory
FIRST_ENTRY=$(ls -A "$MOUNT_POINT" | head -1)

if [ -n "$FIRST_ENTRY" ]; then
    print_info "Testing stat on: $FIRST_ENTRY"
    
    START_TIME=$(date +%s)
    if timeout $TEST_TIMEOUT stat "$MOUNT_POINT/$FIRST_ENTRY" > /tmp/test-stat.txt 2>&1; then
        END_TIME=$(date +%s)
        ELAPSED=$((END_TIME - START_TIME))
        
        print_success "Stat operation succeeded (${ELAPSED}s)"
        print_info "Stat output:"
        cat /tmp/test-stat.txt | sed 's/^/  /'
    else
        print_error "Stat operation failed or timed out"
        cat /tmp/test-stat.txt
    fi
else
    print_warning "No entries found to test stat operation"
fi
echo ""

# Test 3: Read File (if text file exists)
print_info "Test 3: Read File Operations"
print_info "-----------------------------"

# Look for a small text file
TEXT_FILE=$(find "$MOUNT_POINT" -maxdepth 2 -type f -name "*.txt" -o -name "*.md" 2>/dev/null | head -1)

if [ -n "$TEXT_FILE" ]; then
    FILE_SIZE=$(stat -c%s "$TEXT_FILE" 2>/dev/null || echo "0")
    print_info "Testing read on: $(basename "$TEXT_FILE") (size: $FILE_SIZE bytes)"
    
    # Only read if file is reasonably small (< 1MB)
    if [ "$FILE_SIZE" -lt 1048576 ]; then
        START_TIME=$(date +%s)
        if timeout $TEST_TIMEOUT head -20 "$TEXT_FILE" > /tmp/test-read.txt 2>&1; then
            END_TIME=$(date +%s)
            ELAPSED=$((END_TIME - START_TIME))
            
            print_success "Read operation succeeded (${ELAPSED}s)"
            print_info "First 10 lines:"
            head -10 /tmp/test-read.txt | sed 's/^/  /'
        else
            print_error "Read operation failed or timed out"
            cat /tmp/test-read.txt
        fi
    else
        print_warning "File too large for test, skipping read"
    fi
else
    print_warning "No suitable text file found for read test"
    print_info "Creating a test file for future tests..."
    
    # Try to create a test file (if write operations work)
    if timeout $TEST_TIMEOUT bash -c "echo 'Test content' > '$MOUNT_POINT/onemount-test.txt'" 2>/dev/null; then
        print_success "Test file created successfully"
        
        # Try to read it back
        if timeout $TEST_TIMEOUT cat "$MOUNT_POINT/onemount-test.txt" > /tmp/test-read.txt 2>&1; then
            print_success "Read back test file successfully"
            cat /tmp/test-read.txt | sed 's/^/  /'
        fi
    else
        print_warning "Could not create test file (may be read-only or offline)"
    fi
fi
echo ""

# Test 4: Directory Traversal
print_info "Test 4: Directory Traversal"
print_info "----------------------------"

# Find first subdirectory
FIRST_DIR=$(find "$MOUNT_POINT" -maxdepth 1 -type d ! -path "$MOUNT_POINT" | head -1)

if [ -n "$FIRST_DIR" ]; then
    print_info "Testing traversal into: $(basename "$FIRST_DIR")"
    
    START_TIME=$(date +%s)
    if timeout $TEST_TIMEOUT ls -la "$FIRST_DIR" > /tmp/test-traverse.txt 2>&1; then
        END_TIME=$(date +%s)
        ELAPSED=$((END_TIME - START_TIME))
        
        print_success "Directory traversal succeeded (${ELAPSED}s)"
        
        SUBENTRY_COUNT=$(ls -A "$FIRST_DIR" | wc -l)
        print_info "Found $SUBENTRY_COUNT entries in subdirectory"
        
        # Show first few entries
        if [ $SUBENTRY_COUNT -gt 0 ]; then
            print_info "First 5 entries:"
            head -5 /tmp/test-traverse.txt | sed 's/^/  /'
        fi
    else
        print_error "Directory traversal failed or timed out"
        cat /tmp/test-traverse.txt
    fi
else
    print_warning "No subdirectories found for traversal test"
fi
echo ""

# Test 5: Multiple Sequential Operations
print_info "Test 5: Multiple Sequential Operations"
print_info "---------------------------------------"

print_info "Running multiple operations in sequence..."
START_TIME=$(date +%s)

OPERATIONS_PASSED=0
OPERATIONS_TOTAL=5

# Operation 1: ls
if timeout $TEST_TIMEOUT ls "$MOUNT_POINT" > /dev/null 2>&1; then
    OPERATIONS_PASSED=$((OPERATIONS_PASSED + 1))
    print_success "  [1/5] ls operation passed"
else
    print_error "  [1/5] ls operation failed"
fi

# Operation 2: stat on mount point
if timeout $TEST_TIMEOUT stat "$MOUNT_POINT" > /dev/null 2>&1; then
    OPERATIONS_PASSED=$((OPERATIONS_PASSED + 1))
    print_success "  [2/5] stat operation passed"
else
    print_error "  [2/5] stat operation failed"
fi

# Operation 3: find (limited depth)
if timeout $TEST_TIMEOUT find "$MOUNT_POINT" -maxdepth 1 > /dev/null 2>&1; then
    OPERATIONS_PASSED=$((OPERATIONS_PASSED + 1))
    print_success "  [3/5] find operation passed"
else
    print_error "  [3/5] find operation failed"
fi

# Operation 4: du (disk usage)
if timeout $TEST_TIMEOUT du -sh "$MOUNT_POINT" > /dev/null 2>&1; then
    OPERATIONS_PASSED=$((OPERATIONS_PASSED + 1))
    print_success "  [4/5] du operation passed"
else
    print_error "  [4/5] du operation failed"
fi

# Operation 5: ls again (test caching)
if timeout $TEST_TIMEOUT ls "$MOUNT_POINT" > /dev/null 2>&1; then
    OPERATIONS_PASSED=$((OPERATIONS_PASSED + 1))
    print_success "  [5/5] ls (cached) operation passed"
else
    print_error "  [5/5] ls (cached) operation failed"
fi

END_TIME=$(date +%s)
ELAPSED=$((END_TIME - START_TIME))

print_info "Sequential operations completed: $OPERATIONS_PASSED/$OPERATIONS_TOTAL passed (${ELAPSED}s total)"
echo ""

# Summary
print_info "Test Summary"
print_info "============"
echo ""

if mountpoint -q "$MOUNT_POINT"; then
    print_success "✓ Filesystem is mounted"
else
    print_error "✗ Filesystem is not mounted"
fi

if [ $OPERATIONS_PASSED -ge 4 ]; then
    print_success "✓ Most operations passed ($OPERATIONS_PASSED/$OPERATIONS_TOTAL)"
else
    print_warning "⚠ Some operations failed ($OPERATIONS_PASSED/$OPERATIONS_TOTAL)"
fi

# Check mount process is still running
if kill -0 $MOUNT_PID 2>/dev/null; then
    print_success "✓ Mount process is still running"
else
    print_warning "⚠ Mount process has terminated"
fi

echo ""
print_info "Task 5.4 test complete!"
print_info "Mount will be cleaned up automatically..."

# Cleanup will happen via trap
exit 0
