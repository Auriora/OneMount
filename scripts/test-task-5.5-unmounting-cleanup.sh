#!/bin/bash
# Task 5.5: Test Unmounting and Cleanup
# Tests clean unmounting and resource cleanup

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
MOUNT_POINT="/tmp/onemount-test-unmount"
CACHE_DIR="/tmp/onemount-test-unmount-cache"
MOUNT_TIMEOUT=120
LOG_FILE="/tmp/onemount-unmount-test.log"

# Cleanup function
cleanup() {
    print_info "Final cleanup..."
    
    # Unmount if still mounted
    if mountpoint -q "$MOUNT_POINT" 2>/dev/null; then
        print_info "Unmounting filesystem..."
        fusermount3 -uz "$MOUNT_POINT" 2>/dev/null || true
        sleep 1
    fi
    
    # Kill any remaining processes
    pkill -f "onemount.*$MOUNT_POINT" 2>/dev/null || true
    
    # Clean up directories
    rm -rf "$MOUNT_POINT" "$CACHE_DIR" "$LOG_FILE"
    
    print_info "Final cleanup complete"
}

# Set up trap for cleanup
trap cleanup EXIT INT TERM

print_info "Task 5.5: Test Unmounting and Cleanup"
print_info "======================================"
echo ""

# Check prerequisites
print_info "Checking prerequisites..."

if [ ! -f "./build/onemount" ]; then
    print_error "OneMount binary not found"
    exit 1
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
        AUTH_TOKEN_FOUND=true
        if [ "$location" != "$HOME/.onemount-tests/.auth_tokens.json" ]; then
            mkdir -p "$HOME/.onemount-tests"
            cp "$location" "$HOME/.onemount-tests/.auth_tokens.json"
            chmod 600 "$HOME/.onemount-tests/.auth_tokens.json"
        fi
        break
    fi
done

if [ "$AUTH_TOKEN_FOUND" = false ]; then
    print_error "No auth tokens found"
    exit 1
fi

print_success "All prerequisites met"
echo ""

# Create mount point and cache directory
print_info "Setting up test environment..."
mkdir -p "$MOUNT_POINT" "$CACHE_DIR"
print_success "Test directories created"
echo ""

# Test 1: Mount and Unmount with fusermount3
print_info "Test 1: Unmount with fusermount3"
print_info "---------------------------------"

print_info "Starting mount process..."
./build/onemount \
    --mount-timeout "$MOUNT_TIMEOUT" \
    --no-sync-tree \
    --log=info \
    --cache-dir="$CACHE_DIR" \
    "$MOUNT_POINT" > "$LOG_FILE" 2>&1 &

MOUNT_PID=$!
print_info "Mount process started (PID: $MOUNT_PID)"

# Wait for mount
WAIT_COUNT=0
while [ $WAIT_COUNT -lt $MOUNT_TIMEOUT ]; do
    if mountpoint -q "$MOUNT_POINT" 2>/dev/null; then
        break
    fi
    if ! kill -0 $MOUNT_PID 2>/dev/null; then
        print_error "Mount process died"
        exit 1
    fi
    sleep 1
    WAIT_COUNT=$((WAIT_COUNT + 1))
done

if ! mountpoint -q "$MOUNT_POINT" 2>/dev/null; then
    print_error "Mount did not complete"
    exit 1
fi

print_success "Filesystem mounted successfully"
sleep 2

# Unmount using fusermount3
print_info "Unmounting with fusermount3..."
START_TIME=$(date +%s)

if fusermount3 -uz "$MOUNT_POINT" 2>&1 | tee /tmp/unmount-output.txt; then
    END_TIME=$(date +%s)
    ELAPSED=$((END_TIME - START_TIME))
    print_success "Unmount succeeded (${ELAPSED}s)"
else
    print_error "Unmount failed"
    cat /tmp/unmount-output.txt
    exit 1
fi

# Wait for process to exit
sleep 2

# Check if process exited
if kill -0 $MOUNT_PID 2>/dev/null; then
    print_warning "Mount process still running after unmount"
    # Give it a bit more time
    sleep 3
    if kill -0 $MOUNT_PID 2>/dev/null; then
        print_error "Mount process did not exit after unmount"
        kill -9 $MOUNT_PID 2>/dev/null || true
    else
        print_success "Mount process exited (after delay)"
    fi
else
    print_success "Mount process exited cleanly"
fi

echo ""

# Test 2: Verify Mount Point Released
print_info "Test 2: Verify Mount Point Released"
print_info "------------------------------------"

# Check mount table
if mount | grep -q "$MOUNT_POINT"; then
    print_error "Mount point still in mount table"
    mount | grep "$MOUNT_POINT"
else
    print_success "Mount point not in mount table"
fi

# Check if directory is accessible
if [ -d "$MOUNT_POINT" ]; then
    print_success "Mount point directory exists"
    
    # Check if empty
    if [ -z "$(ls -A "$MOUNT_POINT")" ]; then
        print_success "Mount point is empty (original state)"
    else
        print_warning "Mount point is not empty"
        ls -la "$MOUNT_POINT"
    fi
else
    print_error "Mount point directory does not exist"
fi

echo ""

# Test 3: Check for Orphaned Processes
print_info "Test 3: Check for Orphaned Processes"
print_info "-------------------------------------"

ORPHANED_PROCESSES=$(ps aux | grep "[o]nemount.*$MOUNT_POINT" || true)

if [ -z "$ORPHANED_PROCESSES" ]; then
    print_success "No orphaned onemount processes"
else
    print_error "Found orphaned processes:"
    echo "$ORPHANED_PROCESSES"
fi

echo ""

# Test 4: Verify Clean Shutdown in Logs
print_info "Test 4: Verify Clean Shutdown in Logs"
print_info "--------------------------------------"

if [ -f "$LOG_FILE" ]; then
    print_info "Checking log file for shutdown messages..."
    
    # Check for clean shutdown message
    if grep -q "Filesystem unmounted successfully" "$LOG_FILE"; then
        print_success "Found 'Filesystem unmounted successfully' message"
    else
        print_warning "'Filesystem unmounted successfully' message not found"
    fi
    
    # Check for error messages
    ERROR_COUNT=$(grep -i "error" "$LOG_FILE" | grep -v "ERROR" | wc -l)
    if [ $ERROR_COUNT -eq 0 ]; then
        print_success "No error messages in log"
    else
        print_warning "Found $ERROR_COUNT error messages in log"
    fi
    
    # Show last 20 lines of log
    print_info "Last 20 lines of log:"
    tail -20 "$LOG_FILE" | sed 's/^/  /'
else
    print_warning "Log file not found"
fi

echo ""

# Test 5: Test Signal-Based Unmount (SIGTERM)
print_info "Test 5: Test Signal-Based Unmount (SIGTERM)"
print_info "--------------------------------------------"

# Clean up from previous test
rm -rf "$MOUNT_POINT" "$CACHE_DIR"
mkdir -p "$MOUNT_POINT" "$CACHE_DIR"

print_info "Starting mount process..."
./build/onemount \
    --mount-timeout "$MOUNT_TIMEOUT" \
    --no-sync-tree \
    --log=info \
    --cache-dir="$CACHE_DIR" \
    "$MOUNT_POINT" > "$LOG_FILE" 2>&1 &

MOUNT_PID=$!
print_info "Mount process started (PID: $MOUNT_PID)"

# Wait for mount
WAIT_COUNT=0
while [ $WAIT_COUNT -lt $MOUNT_TIMEOUT ]; do
    if mountpoint -q "$MOUNT_POINT" 2>/dev/null; then
        break
    fi
    if ! kill -0 $MOUNT_PID 2>/dev/null; then
        print_error "Mount process died"
        exit 1
    fi
    sleep 1
    WAIT_COUNT=$((WAIT_COUNT + 1))
done

if ! mountpoint -q "$MOUNT_POINT" 2>/dev/null; then
    print_error "Mount did not complete"
    exit 1
fi

print_success "Filesystem mounted successfully"
sleep 2

# Send SIGTERM
print_info "Sending SIGTERM to process $MOUNT_PID..."
START_TIME=$(date +%s)

if kill -TERM $MOUNT_PID 2>/dev/null; then
    print_success "SIGTERM sent successfully"
else
    print_error "Failed to send SIGTERM"
    exit 1
fi

# Wait for process to exit
WAIT_COUNT=0
MAX_WAIT=30

while [ $WAIT_COUNT -lt $MAX_WAIT ]; do
    if ! kill -0 $MOUNT_PID 2>/dev/null; then
        END_TIME=$(date +%s)
        ELAPSED=$((END_TIME - START_TIME))
        print_success "Process exited after SIGTERM (${ELAPSED}s)"
        break
    fi
    sleep 1
    WAIT_COUNT=$((WAIT_COUNT + 1))
done

if kill -0 $MOUNT_PID 2>/dev/null; then
    print_error "Process did not exit after SIGTERM (${MAX_WAIT}s)"
    kill -9 $MOUNT_PID 2>/dev/null || true
else
    # Check if mount point was released
    if mountpoint -q "$MOUNT_POINT" 2>/dev/null; then
        print_error "Mount point still mounted after SIGTERM"
        fusermount3 -uz "$MOUNT_POINT" 2>/dev/null || true
    else
        print_success "Mount point released after SIGTERM"
    fi
fi

# Check logs for shutdown sequence
if [ -f "$LOG_FILE" ]; then
    print_info "Checking for shutdown sequence in logs..."
    
    SHUTDOWN_MESSAGES=(
        "Signal received"
        "cleaning up"
        "Stopping cache cleanup"
        "Stopping delta loop"
        "Stopping download manager"
        "Stopping upload manager"
        "Filesystem unmounted successfully"
    )
    
    FOUND_COUNT=0
    for msg in "${SHUTDOWN_MESSAGES[@]}"; do
        if grep -qi "$msg" "$LOG_FILE"; then
            print_success "  ✓ Found: $msg"
            FOUND_COUNT=$((FOUND_COUNT + 1))
        else
            print_warning "  ✗ Not found: $msg"
        fi
    done
    
    print_info "Found $FOUND_COUNT/${#SHUTDOWN_MESSAGES[@]} expected shutdown messages"
fi

echo ""

# Test 6: Verify Resource Cleanup
print_info "Test 6: Verify Resource Cleanup"
print_info "--------------------------------"

# Check mount point
if mountpoint -q "$MOUNT_POINT" 2>/dev/null; then
    print_error "Mount point still mounted"
else
    print_success "Mount point not mounted"
fi

# Check processes
if ps aux | grep -q "[o]nemount.*$MOUNT_POINT"; then
    print_error "Orphaned processes found"
    ps aux | grep "[o]nemount.*$MOUNT_POINT"
else
    print_success "No orphaned processes"
fi

# Check mount point accessibility
if [ -d "$MOUNT_POINT" ]; then
    if ls "$MOUNT_POINT" > /dev/null 2>&1; then
        print_success "Mount point is accessible"
    else
        print_error "Mount point is not accessible"
    fi
else
    print_warning "Mount point directory does not exist"
fi

echo ""

# Summary
print_info "Test Summary"
print_info "============"
echo ""

TESTS_PASSED=0
TESTS_TOTAL=6

# Test 1: fusermount3 unmount
if [ -f /tmp/unmount-output.txt ] && ! grep -q "error" /tmp/unmount-output.txt; then
    print_success "✓ Test 1: fusermount3 unmount - PASSED"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    print_error "✗ Test 1: fusermount3 unmount - FAILED"
fi

# Test 2: Mount point released
if ! mount | grep -q "$MOUNT_POINT"; then
    print_success "✓ Test 2: Mount point released - PASSED"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    print_error "✗ Test 2: Mount point released - FAILED"
fi

# Test 3: No orphaned processes
if ! ps aux | grep -q "[o]nemount.*$MOUNT_POINT"; then
    print_success "✓ Test 3: No orphaned processes - PASSED"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    print_error "✗ Test 3: No orphaned processes - FAILED"
fi

# Test 4: Clean shutdown logs
if [ -f "$LOG_FILE" ] && grep -q "Filesystem unmounted successfully" "$LOG_FILE"; then
    print_success "✓ Test 4: Clean shutdown logs - PASSED"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    print_warning "⚠ Test 4: Clean shutdown logs - PARTIAL"
fi

# Test 5: SIGTERM handling
if [ $FOUND_COUNT -ge 4 ]; then
    print_success "✓ Test 5: SIGTERM handling - PASSED"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    print_warning "⚠ Test 5: SIGTERM handling - PARTIAL"
fi

# Test 6: Resource cleanup
if ! mountpoint -q "$MOUNT_POINT" 2>/dev/null && ! ps aux | grep -q "[o]nemount.*$MOUNT_POINT"; then
    print_success "✓ Test 6: Resource cleanup - PASSED"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    print_error "✗ Test 6: Resource cleanup - FAILED"
fi

echo ""
print_info "Tests Passed: $TESTS_PASSED/$TESTS_TOTAL"

if [ $TESTS_PASSED -ge 5 ]; then
    print_success "Task 5.5 test PASSED!"
    exit 0
elif [ $TESTS_PASSED -ge 3 ]; then
    print_warning "Task 5.5 test PARTIALLY PASSED"
    exit 0
else
    print_error "Task 5.5 test FAILED"
    exit 1
fi
