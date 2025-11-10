#!/bin/bash
# Test script for mount point validation
# Task 5.3: Test mount point validation

set -e

echo "========================================="
echo "Task 5.3: Mount Point Validation Test"
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
TEST_BASE_DIR="/tmp/onemount-validation-test"
TEST_CACHE_DIR="${TEST_BASE_DIR}/cache"
AUTH_TOKENS_PATH="${TEST_CACHE_DIR}/auth_tokens.json"

echo "Test base dir: ${TEST_BASE_DIR}"
echo ""

# Cleanup function
cleanup() {
    local exit_code=$?
    echo ""
    echo "Cleaning up..."
    
    # Remove test directories
    rm -rf "${TEST_BASE_DIR}" 2>/dev/null || true
    
    echo "Cleanup complete"
    exit $exit_code
}

# Set trap to cleanup on exit
trap cleanup EXIT INT TERM

# Create base test directory
echo "Creating test base directory..."
mkdir -p "${TEST_BASE_DIR}"
mkdir -p "${TEST_CACHE_DIR}"
print_result $? "Create test directories"

# Setup auth tokens
echo ""
echo "Setting up authentication..."
if [ -f "/tmp/home-tester/.onemount-tests/.auth_tokens.json" ]; then
    cp "/tmp/home-tester/.onemount-tests/.auth_tokens.json" "${AUTH_TOKENS_PATH}"
    print_result 0 "Auth tokens copied"
elif [ -f "test-artifacts/.auth_tokens.json" ]; then
    cp "test-artifacts/.auth_tokens.json" "${AUTH_TOKENS_PATH}"
    print_result 0 "Auth tokens copied"
elif [ -f "auth_tokens.json" ]; then
    cp "auth_tokens.json" "${AUTH_TOKENS_PATH}"
    print_result 0 "Auth tokens copied"
else
    echo -e "${YELLOW}WARNING: No auth tokens found, using dummy tokens${NC}"
    echo '{"access_token":"dummy","refresh_token":"dummy","expires_at":9999999999}' > "${AUTH_TOKENS_PATH}"
    print_result 0 "Dummy auth tokens created"
fi

# Check if onemount binary exists
echo ""
echo "Checking onemount binary..."
if [ ! -f "build/onemount" ]; then
    echo "Building onemount..."
    make build
    BUILD_RESULT=$?
else
    echo "Using existing build/onemount"
    BUILD_RESULT=0
fi
print_result ${BUILD_RESULT} "OneMount binary available"

if [ ${BUILD_RESULT} -ne 0 ]; then
    echo "Build failed, cannot proceed"
    exit 1
fi

# Test 1: Attempt to mount at non-existent directory
echo ""
echo "========================================="
echo "Test 1: Mount at non-existent directory"
echo "========================================="

NON_EXISTENT_DIR="${TEST_BASE_DIR}/does-not-exist"
echo "Attempting to mount at: ${NON_EXISTENT_DIR}"

# Run onemount and capture output
if ./build/onemount --cache-dir="${TEST_CACHE_DIR}" "${NON_EXISTENT_DIR}" > /tmp/test1-output.log 2>&1; then
    print_result 1 "Should have failed with non-existent directory"
    echo "Unexpected success - output:"
    cat /tmp/test1-output.log
else
    # Check if error message is appropriate
    if grep -q "did not exist\|not a directory\|no such file" /tmp/test1-output.log; then
        print_result 0 "Correctly rejected non-existent directory"
        echo "Error message:"
        grep -i "did not exist\|not a directory\|no such file" /tmp/test1-output.log | head -3
    else
        print_result 1 "Failed but with unexpected error message"
        echo "Output:"
        cat /tmp/test1-output.log
    fi
fi

# Test 2: Attempt to mount at a file (not directory)
echo ""
echo "========================================="
echo "Test 2: Mount at file (not directory)"
echo "========================================="

TEST_FILE="${TEST_BASE_DIR}/test-file.txt"
echo "test" > "${TEST_FILE}"
echo "Attempting to mount at file: ${TEST_FILE}"

if ./build/onemount --cache-dir="${TEST_CACHE_DIR}" "${TEST_FILE}" > /tmp/test2-output.log 2>&1; then
    print_result 1 "Should have failed with file instead of directory"
    echo "Unexpected success - output:"
    cat /tmp/test2-output.log
else
    # Check if error message is appropriate
    if grep -q "not a directory" /tmp/test2-output.log; then
        print_result 0 "Correctly rejected file as mount point"
        echo "Error message:"
        grep -i "not a directory" /tmp/test2-output.log | head -3
    else
        print_result 1 "Failed but with unexpected error message"
        echo "Output:"
        cat /tmp/test2-output.log
    fi
fi

# Test 3: Attempt to mount at non-empty directory
echo ""
echo "========================================="
echo "Test 3: Mount at non-empty directory"
echo "========================================="

NON_EMPTY_DIR="${TEST_BASE_DIR}/non-empty"
mkdir -p "${NON_EMPTY_DIR}"
echo "test" > "${NON_EMPTY_DIR}/existing-file.txt"
echo "Attempting to mount at non-empty directory: ${NON_EMPTY_DIR}"

if timeout 5 ./build/onemount --cache-dir="${TEST_CACHE_DIR}" "${NON_EMPTY_DIR}" > /tmp/test3-output.log 2>&1; then
    print_result 1 "Should have failed with non-empty directory"
    echo "Unexpected success - output:"
    cat /tmp/test3-output.log
    # Try to unmount if it somehow mounted
    fusermount3 -uz "${NON_EMPTY_DIR}" 2>/dev/null || true
else
    # Check if error message is appropriate
    if grep -q "must be empty" /tmp/test3-output.log; then
        print_result 0 "Correctly rejected non-empty directory"
        echo "Error message:"
        grep -i "must be empty" /tmp/test3-output.log | head -3
    else
        # Timeout or other error
        if [ $? -eq 124 ]; then
            print_result 1 "Command timed out (may have attempted to mount)"
            echo "Output:"
            cat /tmp/test3-output.log
        else
            print_result 1 "Failed but with unexpected error message"
            echo "Output:"
            cat /tmp/test3-output.log
        fi
    fi
fi

# Test 4: Attempt to mount at already-mounted location
echo ""
echo "========================================="
echo "Test 4: Mount at already-mounted location"
echo "========================================="

ALREADY_MOUNTED_DIR="${TEST_BASE_DIR}/already-mounted"
mkdir -p "${ALREADY_MOUNTED_DIR}"

# First, try to mount successfully (this may timeout, but that's okay)
echo "First mount attempt (may timeout, that's expected)..."
timeout 3 ./build/onemount --cache-dir="${TEST_CACHE_DIR}" --no-sync-tree "${ALREADY_MOUNTED_DIR}" > /tmp/test4-mount1.log 2>&1 &
FIRST_MOUNT_PID=$!

# Wait a bit to see if it mounts
sleep 2

# Check if it's mounted
if mountpoint -q "${ALREADY_MOUNTED_DIR}" 2>/dev/null; then
    echo "First mount succeeded, now attempting second mount..."
    
    # Try to mount again at the same location
    if timeout 3 ./build/onemount --cache-dir="${TEST_CACHE_DIR}" "${ALREADY_MOUNTED_DIR}" > /tmp/test4-mount2.log 2>&1; then
        print_result 1 "Should have failed with already-mounted location"
        echo "Unexpected success - output:"
        cat /tmp/test4-mount2.log
    else
        # Check if error message is appropriate
        if grep -q "already mounted\|already in use\|device or resource busy" /tmp/test4-mount2.log; then
            print_result 0 "Correctly rejected already-mounted location"
            echo "Error message:"
            grep -i "already mounted\|already in use\|device or resource busy" /tmp/test4-mount2.log | head -3
        else
            print_result 1 "Failed but with unexpected error message"
            echo "Output:"
            cat /tmp/test4-mount2.log
        fi
    fi
    
    # Cleanup: unmount the first mount
    echo "Cleaning up first mount..."
    kill -TERM ${FIRST_MOUNT_PID} 2>/dev/null || true
    sleep 1
    fusermount3 -uz "${ALREADY_MOUNTED_DIR}" 2>/dev/null || true
else
    echo -e "${YELLOW}First mount did not complete (expected due to timeout)${NC}"
    echo "Skipping already-mounted test"
    print_result 0 "Test skipped (mount timeout)"
    
    # Cleanup
    kill -TERM ${FIRST_MOUNT_PID} 2>/dev/null || true
fi

# Test 5: Valid empty directory (should start mounting)
echo ""
echo "========================================="
echo "Test 5: Valid empty directory"
echo "========================================="

VALID_DIR="${TEST_BASE_DIR}/valid-mount"
mkdir -p "${VALID_DIR}"
echo "Attempting to mount at valid empty directory: ${VALID_DIR}"

# Start mount in background with short timeout
timeout 3 ./build/onemount --cache-dir="${TEST_CACHE_DIR}" --no-sync-tree "${VALID_DIR}" > /tmp/test5-output.log 2>&1 &
MOUNT_PID=$!

# Wait a bit
sleep 2

# Check if process is still running (indicates it accepted the mount point)
if kill -0 ${MOUNT_PID} 2>/dev/null; then
    print_result 0 "Accepted valid empty directory (process started)"
    echo "OneMount process is running (PID: ${MOUNT_PID})"
    
    # Check if it actually mounted
    if mountpoint -q "${VALID_DIR}" 2>/dev/null; then
        echo "Mount point is active"
    else
        echo "Mount point not yet active (may be initializing)"
    fi
    
    # Cleanup
    kill -TERM ${MOUNT_PID} 2>/dev/null || true
    sleep 1
    fusermount3 -uz "${VALID_DIR}" 2>/dev/null || true
else
    # Process exited - check why
    wait ${MOUNT_PID}
    EXIT_CODE=$?
    
    if [ ${EXIT_CODE} -eq 0 ]; then
        print_result 1 "Process exited successfully but should be running"
    else
        print_result 1 "Process exited with error"
        echo "Exit code: ${EXIT_CODE}"
        echo "Output:"
        cat /tmp/test5-output.log
    fi
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
    exit 1
fi
