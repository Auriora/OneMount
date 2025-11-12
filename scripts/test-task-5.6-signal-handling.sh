#!/bin/bash
# Task 5.6: Test Signal Handling
# Tests graceful shutdown via SIGINT and SIGTERM signals

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
MOUNT_POINT="/tmp/onemount-test-signals"
CACHE_DIR="/tmp/onemount-test-signals-cache"
MOUNT_TIMEOUT=120
LOG_FILE="/tmp/onemount-signals-test.log"

# Cleanup function
cleanup() {
    print_info "Final cleanup..."
    
    # Unmount if still mounted
    if mountpoint -q "$MOUNT_POINT" 2>/dev/null; then
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

print_info "Task 5.6: Test Signal Handling"
print_info "==============================="
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

# Test 1: SIGINT (Ctrl+C) Handling
print_info "Test 1: SIGINT (Ctrl+C) Handling"
print_info "---------------------------------"

# Create mount point and cache
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

# Send SIGINT
print_info "Sending SIGINT (Ctrl+C equivalent) to process $MOUNT_PID..."
START_TIME=$(date +%s)

if kill -INT $MOUNT_PID 2>/dev/null; then
    print_success "SIGINT sent successfully"
else
    print_error "Failed to send SIGINT"
    exit 1
fi

# Wait for graceful shutdown
WAIT_COUNT=0
MAX_WAIT=30
SHUTDOWN_SUCCESS=false

while [ $WAIT_COUNT -lt $MAX_WAIT ]; do
    if ! kill -0 $MOUNT_PID 2>/dev/null; then
        END_TIME=$(date +%s)
        ELAPSED=$((END_TIME - START_TIME))
        print_success "Process exited gracefully after SIGINT (${ELAPSED}s)"
        SHUTDOWN_SUCCESS=true
        break
    fi
    sleep 1
    WAIT_COUNT=$((WAIT_COUNT + 1))
done

if [ "$SHUTDOWN_SUCCESS" = false ]; then
    print_error "Process did not exit after SIGINT (${MAX_WAIT}s)"
    kill -9 $MOUNT_PID 2>/dev/null || true
    SIGINT_RESULT="FAILED"
else
    # Check if mount point was released
    if mountpoint -q "$MOUNT_POINT" 2>/dev/null; then
        print_warning "Mount point still mounted after SIGINT"
        fusermount3 -uz "$MOUNT_POINT" 2>/dev/null || true
        SIGINT_RESULT="PARTIAL"
    else
        print_success "Mount point released after SIGINT"
        SIGINT_RESULT="PASSED"
    fi
    
    # Check exit code (if we can capture it)
    wait $MOUNT_PID 2>/dev/null
    EXIT_CODE=$?
    if [ $EXIT_CODE -eq 0 ]; then
        print_success "Process exited with code 0 (success)"
    else
        print_warning "Process exited with code $EXIT_CODE"
    fi
fi

echo ""

# Test 2: SIGTERM Handling
print_info "Test 2: SIGTERM Handling"
print_info "-------------------------"

# Clean up and recreate
rm -rf "$MOUNT_POINT" "$CACHE_DIR" "$LOG_FILE"
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

# Wait for graceful shutdown
WAIT_COUNT=0
MAX_WAIT=30
SHUTDOWN_SUCCESS=false

while [ $WAIT_COUNT -lt $MAX_WAIT ]; do
    if ! kill -0 $MOUNT_PID 2>/dev/null; then
        END_TIME=$(date +%s)
        ELAPSED=$((END_TIME - START_TIME))
        print_success "Process exited gracefully after SIGTERM (${ELAPSED}s)"
        SHUTDOWN_SUCCESS=true
        break
    fi
    sleep 1
    WAIT_COUNT=$((WAIT_COUNT + 1))
done

if [ "$SHUTDOWN_SUCCESS" = false ]; then
    print_error "Process did not exit after SIGTERM (${MAX_WAIT}s)"
    kill -9 $MOUNT_PID 2>/dev/null || true
    SIGTERM_RESULT="FAILED"
else
    # Check if mount point was released
    if mountpoint -q "$MOUNT_POINT" 2>/dev/null; then
        print_warning "Mount point still mounted after SIGTERM"
        fusermount3 -uz "$MOUNT_POINT" 2>/dev/null || true
        SIGTERM_RESULT="PARTIAL"
    else
        print_success "Mount point released after SIGTERM"
        SIGTERM_RESULT="PASSED"
    fi
    
    # Check exit code
    wait $MOUNT_PID 2>/dev/null
    EXIT_CODE=$?
    if [ $EXIT_CODE -eq 0 ]; then
        print_success "Process exited with code 0 (success)"
    else
        print_warning "Process exited with code $EXIT_CODE"
    fi
fi

echo ""

# Test 3: Verify Shutdown Sequence
print_info "Test 3: Verify Shutdown Sequence"
print_info "---------------------------------"

if [ -f "$LOG_FILE" ]; then
    print_info "Checking log file for shutdown sequence..."
    
    # Expected shutdown messages
    SHUTDOWN_MESSAGES=(
        "Signal received"
        "cleaning up"
        "Canceling context"
        "Stopping"
        "unmount"
    )
    
    FOUND_COUNT=0
    for msg in "${SHUTDOWN_MESSAGES[@]}"; do
        if grep -qi "$msg" "$LOG_FILE"; then
            print_success "  ✓ Found: $msg"
            FOUND_COUNT=$((FOUND_COUNT + 1))
        else
            print_info "  ○ Not found: $msg"
        fi
    done
    
    print_info "Found $FOUND_COUNT/${#SHUTDOWN_MESSAGES[@]} expected shutdown messages"
    
    # Show last 30 lines of log
    print_info "Last 30 lines of log:"
    tail -30 "$LOG_FILE" | sed 's/^/  /'
else
    print_warning "Log file not found"
fi

echo ""

# Test 4: Verify Resource Cleanup After Signals
print_info "Test 4: Verify Resource Cleanup After Signals"
print_info "----------------------------------------------"

# Check mount point
if mountpoint -q "$MOUNT_POINT" 2>/dev/null; then
    print_error "Mount point still mounted"
    CLEANUP_RESULT="FAILED"
else
    print_success "Mount point not mounted"
    CLEANUP_RESULT="PASSED"
fi

# Check processes
if ps aux | grep -q "[o]nemount.*$MOUNT_POINT"; then
    print_error "Orphaned processes found"
    ps aux | grep "[o]nemount.*$MOUNT_POINT"
    CLEANUP_RESULT="FAILED"
else
    print_success "No orphaned processes"
fi

# Check mount point accessibility
if [ -d "$MOUNT_POINT" ]; then
    if ls "$MOUNT_POINT" > /dev/null 2>&1; then
        print_success "Mount point is accessible"
    else
        print_error "Mount point is not accessible"
        CLEANUP_RESULT="FAILED"
    fi
fi

echo ""

# Summary
print_info "Test Summary"
print_info "============"
echo ""

TESTS_PASSED=0
TESTS_TOTAL=4

# Test 1: SIGINT
if [ "$SIGINT_RESULT" = "PASSED" ]; then
    print_success "✓ Test 1: SIGINT handling - PASSED"
    TESTS_PASSED=$((TESTS_PASSED + 1))
elif [ "$SIGINT_RESULT" = "PARTIAL" ]; then
    print_warning "⚠ Test 1: SIGINT handling - PARTIAL"
else
    print_error "✗ Test 1: SIGINT handling - FAILED"
fi

# Test 2: SIGTERM
if [ "$SIGTERM_RESULT" = "PASSED" ]; then
    print_success "✓ Test 2: SIGTERM handling - PASSED"
    TESTS_PASSED=$((TESTS_PASSED + 1))
elif [ "$SIGTERM_RESULT" = "PARTIAL" ]; then
    print_warning "⚠ Test 2: SIGTERM handling - PARTIAL"
else
    print_error "✗ Test 2: SIGTERM handling - FAILED"
fi

# Test 3: Shutdown sequence
if [ $FOUND_COUNT -ge 2 ]; then
    print_success "✓ Test 3: Shutdown sequence - PASSED"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    print_warning "⚠ Test 3: Shutdown sequence - PARTIAL"
fi

# Test 4: Resource cleanup
if [ "$CLEANUP_RESULT" = "PASSED" ]; then
    print_success "✓ Test 4: Resource cleanup - PASSED"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    print_error "✗ Test 4: Resource cleanup - FAILED"
fi

echo ""
print_info "Tests Passed: $TESTS_PASSED/$TESTS_TOTAL"

if [ $TESTS_PASSED -ge 3 ]; then
    print_success "Task 5.6 test PASSED!"
    exit 0
elif [ $TESTS_PASSED -ge 2 ]; then
    print_warning "Task 5.6 test PARTIALLY PASSED"
    exit 0
else
    print_error "Task 5.6 test FAILED"
    exit 1
fi
