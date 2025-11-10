#!/bin/bash
# Test script for basic filesystem mounting verification
# Task 5.2: Test basic mounting

set -e

echo "========================================="
echo "Task 5.2: Basic Mounting Test"
echo "========================================="
echo ""

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test results
TESTS_PASSED=0
TESTS_FAILED=0

# Function to print test result
print_result() {
    if [ $1 -eq 0 ]; then
        echo -e "${GREEN}✓ PASS${NC}: $2"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "${RED}✗ FAIL${NC}: $2"
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi
}

# Check if running in Docker
if [ -f /.dockerenv ] || grep -q docker /proc/1/cgroup 2>/dev/null; then
    echo -e "${YELLOW}Running inside Docker container${NC}"
    IN_DOCKER=true
else
    echo -e "${YELLOW}Running on host system${NC}"
    IN_DOCKER=false
fi

# Setup test environment
TEST_MOUNT_POINT="/tmp/onemount-test-mount"
TEST_CACHE_DIR="/tmp/onemount-test-cache"
AUTH_TOKENS_PATH="${TEST_CACHE_DIR}/auth_tokens.json"

echo "Test mount point: ${TEST_MOUNT_POINT}"
echo "Test cache dir: ${TEST_CACHE_DIR}"
echo ""

# Cleanup function
cleanup() {
    local exit_code=$?
    echo ""
    echo "Cleaning up..."
    
    # Unmount if mounted
    if mountpoint -q "${TEST_MOUNT_POINT}" 2>/dev/null; then
        echo "Unmounting ${TEST_MOUNT_POINT}..."
        fusermount3 -uz "${TEST_MOUNT_POINT}" 2>/dev/null || true
        sleep 1
    fi
    
    # Remove test directories
    rm -rf "${TEST_MOUNT_POINT}" 2>/dev/null || true
    rm -rf "${TEST_CACHE_DIR}" 2>/dev/null || true
    
    echo "Cleanup complete"
    exit $exit_code
}

# Set trap to cleanup on exit
trap cleanup EXIT INT TERM

# Create test directories
echo "Creating test directories..."
mkdir -p "${TEST_MOUNT_POINT}"
mkdir -p "${TEST_CACHE_DIR}"
print_result $? "Create test directories"

# Check if FUSE device is available
echo ""
echo "Checking FUSE device..."
if [ -c /dev/fuse ]; then
    print_result 0 "FUSE device /dev/fuse exists"
else
    print_result 1 "FUSE device /dev/fuse not found"
    echo "FUSE is required for mounting tests"
    exit 1
fi

# Check if auth tokens exist
echo ""
echo "Checking authentication..."
if [ -f "/tmp/home-tester/.onemount-tests/.auth_tokens.json" ]; then
    echo "Using auth tokens from mounted test-artifacts"
    cp "/tmp/home-tester/.onemount-tests/.auth_tokens.json" "${AUTH_TOKENS_PATH}"
    print_result 0 "Auth tokens available"
elif [ -f "test-artifacts/.auth_tokens.json" ]; then
    echo "Using auth tokens from test-artifacts/.auth_tokens.json"
    cp "test-artifacts/.auth_tokens.json" "${AUTH_TOKENS_PATH}"
    print_result 0 "Auth tokens available"
elif [ -f "auth_tokens.json" ]; then
    echo "Using auth tokens from auth_tokens.json"
    cp "auth_tokens.json" "${AUTH_TOKENS_PATH}"
    print_result 0 "Auth tokens available"
else
    echo -e "${YELLOW}WARNING: No auth tokens found${NC}"
    echo "This test requires authentication to OneDrive"
    echo "Please run: ./onemount --auth-only ${TEST_MOUNT_POINT}"
    print_result 1 "Auth tokens not found"
    exit 1
fi

# Build onemount if not already built
echo ""
echo "Building onemount..."
if [ ! -f "build/onemount" ]; then
    make build
    BUILD_RESULT=$?
else
    echo "Using existing build/onemount"
    BUILD_RESULT=0
fi
print_result ${BUILD_RESULT} "Build onemount binary"

if [ ${BUILD_RESULT} -ne 0 ]; then
    echo "Build failed, cannot proceed with mounting test"
    exit 1
fi

# Test 1: Mount filesystem
echo ""
echo "========================================="
echo "Test 1: Mount filesystem"
echo "========================================="

# Start onemount in background
echo "Starting onemount..."
./build/onemount \
    --cache-dir="${TEST_CACHE_DIR}" \
    --log-level=debug \
    --no-sync-tree \
    "${TEST_MOUNT_POINT}" > /tmp/onemount-test.log 2>&1 &

ONEMOUNT_PID=$!
echo "OneMount PID: ${ONEMOUNT_PID}"

# Wait for mount to complete (max 30 seconds)
echo "Waiting for mount to complete..."
MOUNT_TIMEOUT=30
MOUNT_ELAPSED=0
MOUNT_SUCCESS=false

while [ ${MOUNT_ELAPSED} -lt ${MOUNT_TIMEOUT} ]; do
    if mountpoint -q "${TEST_MOUNT_POINT}" 2>/dev/null; then
        MOUNT_SUCCESS=true
        break
    fi
    sleep 1
    ((MOUNT_ELAPSED++))
    echo -n "."
done
echo ""

if [ "${MOUNT_SUCCESS}" = true ]; then
    print_result 0 "Filesystem mounted successfully"
else
    print_result 1 "Filesystem failed to mount within ${MOUNT_TIMEOUT} seconds"
    echo "OneMount log:"
    cat /tmp/onemount-test.log
    kill ${ONEMOUNT_PID} 2>/dev/null || true
    exit 1
fi

# Test 2: Verify mount appears in mount command
echo ""
echo "========================================="
echo "Test 2: Verify mount in mount output"
echo "========================================="

if mount | grep -q "${TEST_MOUNT_POINT}"; then
    print_result 0 "Mount point appears in mount output"
    echo "Mount entry:"
    mount | grep "${TEST_MOUNT_POINT}"
else
    print_result 1 "Mount point not found in mount output"
fi

# Test 3: Verify mount point is accessible
echo ""
echo "========================================="
echo "Test 3: Verify mount point accessibility"
echo "========================================="

if [ -d "${TEST_MOUNT_POINT}" ]; then
    print_result 0 "Mount point is accessible as directory"
else
    print_result 1 "Mount point is not accessible"
fi

# Test 4: Check root directory is visible
echo ""
echo "========================================="
echo "Test 4: Check root directory visibility"
echo "========================================="

# Try to list the root directory
echo "Listing root directory..."
if ls -la "${TEST_MOUNT_POINT}" > /tmp/ls-output.txt 2>&1; then
    print_result 0 "Root directory is listable"
    echo "Root directory contents:"
    cat /tmp/ls-output.txt | head -10
else
    print_result 1 "Failed to list root directory"
    echo "Error output:"
    cat /tmp/ls-output.txt
fi

# Test 5: Verify filesystem responds to stat
echo ""
echo "========================================="
echo "Test 5: Verify stat on mount point"
echo "========================================="

if stat "${TEST_MOUNT_POINT}" > /tmp/stat-output.txt 2>&1; then
    print_result 0 "Stat command successful on mount point"
    echo "Stat output:"
    cat /tmp/stat-output.txt
else
    print_result 1 "Stat command failed on mount point"
    echo "Error output:"
    cat /tmp/stat-output.txt
fi

# Test 6: Graceful unmount
echo ""
echo "========================================="
echo "Test 6: Graceful unmount"
echo "========================================="

echo "Sending SIGTERM to onemount (PID: ${ONEMOUNT_PID})..."
kill -TERM ${ONEMOUNT_PID} 2>/dev/null || true

# Wait for unmount (max 10 seconds)
UNMOUNT_TIMEOUT=10
UNMOUNT_ELAPSED=0
UNMOUNT_SUCCESS=false

while [ ${UNMOUNT_ELAPSED} -lt ${UNMOUNT_TIMEOUT} ]; do
    if ! mountpoint -q "${TEST_MOUNT_POINT}" 2>/dev/null; then
        UNMOUNT_SUCCESS=true
        break
    fi
    sleep 1
    ((UNMOUNT_ELAPSED++))
    echo -n "."
done
echo ""

if [ "${UNMOUNT_SUCCESS}" = true ]; then
    print_result 0 "Filesystem unmounted successfully"
else
    print_result 1 "Filesystem failed to unmount within ${UNMOUNT_TIMEOUT} seconds"
    echo "Forcing unmount..."
    fusermount3 -uz "${TEST_MOUNT_POINT}" 2>/dev/null || true
fi

# Check if process exited
if ! kill -0 ${ONEMOUNT_PID} 2>/dev/null; then
    print_result 0 "OneMount process exited cleanly"
else
    print_result 1 "OneMount process still running"
    kill -9 ${ONEMOUNT_PID} 2>/dev/null || true
fi

# Print summary
echo ""
echo "========================================="
echo "Test Summary"
echo "========================================="
echo -e "Tests passed: ${GREEN}${TESTS_PASSED}${NC}"
echo -e "Tests failed: ${RED}${TESTS_FAILED}${NC}"
echo ""

if [ ${TESTS_FAILED} -eq 0 ]; then
    echo -e "${GREEN}All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}Some tests failed${NC}"
    echo ""
    echo "OneMount log (last 50 lines):"
    tail -50 /tmp/onemount-test.log
    exit 1
fi
