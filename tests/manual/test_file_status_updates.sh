#!/bin/bash
# Manual test script for file status updates
# Tests file status tracking during various operations
# Requirements: 8.1

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test configuration
MOUNT_POINT="/tmp/onemount-status-test"
CACHE_DIR="/tmp/onemount-status-cache"
TEST_FILE="status-test-file.txt"
TEST_DIR="status-test-dir"

echo -e "${BLUE}=== File Status Updates Test ===${NC}"
echo "This test monitors file status during various operations"
echo ""

# Cleanup function
cleanup() {
    echo -e "${YELLOW}Cleaning up...${NC}"
    if mountpoint -q "$MOUNT_POINT" 2>/dev/null; then
        fusermount -u "$MOUNT_POINT" 2>/dev/null || true
        sleep 1
    fi
    rm -rf "$MOUNT_POINT" "$CACHE_DIR"
}

trap cleanup EXIT

# Setup
echo -e "${BLUE}Setting up test environment...${NC}"
cleanup
mkdir -p "$MOUNT_POINT" "$CACHE_DIR"

# Check if onemount binary exists
if [ ! -f "./build/onemount" ]; then
    echo -e "${RED}Error: onemount binary not found at ./build/onemount${NC}"
    echo "Please build the project first: go build -o build/onemount ./cmd/onemount"
    exit 1
fi

# Function to get file status via extended attributes
get_file_status() {
    local file_path="$1"
    if [ -f "$file_path" ] || [ -d "$file_path" ]; then
        status=$(getfattr -n user.onemount.status --only-values "$file_path" 2>/dev/null || echo "NoXattr")
        echo "$status"
    else
        echo "NotFound"
    fi
}

# Function to check if file has error attribute
get_file_error() {
    local file_path="$1"
    if [ -f "$file_path" ] || [ -d "$file_path" ]; then
        error=$(getfattr -n user.onemount.error --only-values "$file_path" 2>/dev/null || echo "")
        echo "$error"
    else
        echo ""
    fi
}

# Function to display status with color
display_status() {
    local status="$1"
    case "$status" in
        "Cloud")
            echo -e "${BLUE}Cloud${NC}"
            ;;
        "Local")
            echo -e "${GREEN}Local${NC}"
            ;;
        "LocalModified")
            echo -e "${YELLOW}LocalModified${NC}"
            ;;
        "Syncing")
            echo -e "${YELLOW}Syncing${NC}"
            ;;
        "Downloading")
            echo -e "${BLUE}Downloading${NC}"
            ;;
        "OutofSync")
            echo -e "${YELLOW}OutofSync${NC}"
            ;;
        "Error")
            echo -e "${RED}Error${NC}"
            ;;
        "Conflict")
            echo -e "${RED}Conflict${NC}"
            ;;
        "NoXattr")
            echo -e "${YELLOW}NoXattr${NC}"
            ;;
        "NotFound")
            echo -e "${RED}NotFound${NC}"
            ;;
        *)
            echo -e "${YELLOW}$status${NC}"
            ;;
    esac
}

# Mount filesystem
echo -e "${BLUE}Mounting OneMount filesystem...${NC}"
./build/onemount --cache-dir="$CACHE_DIR" --no-sync-tree "$MOUNT_POINT" &
MOUNT_PID=$!

# Wait for mount to complete
echo "Waiting for mount to complete..."
for i in {1..30}; do
    if mountpoint -q "$MOUNT_POINT" 2>/dev/null; then
        echo -e "${GREEN}Mount successful!${NC}"
        break
    fi
    sleep 1
    if [ $i -eq 30 ]; then
        echo -e "${RED}Mount timeout after 30 seconds${NC}"
        exit 1
    fi
done

sleep 2

echo ""
echo -e "${BLUE}=== Test 1: File Creation Status ===${NC}"
echo "Creating a new file..."
TEST_FILE_PATH="$MOUNT_POINT/$TEST_FILE"
echo "Test content for status tracking" > "$TEST_FILE_PATH"

echo "Checking status immediately after creation..."
status=$(get_file_status "$TEST_FILE_PATH")
echo -n "Status: "
display_status "$status"

sleep 2
echo "Checking status after 2 seconds..."
status=$(get_file_status "$TEST_FILE_PATH")
echo -n "Status: "
display_status "$status"

echo ""
echo -e "${BLUE}=== Test 2: File Modification Status ===${NC}"
echo "Modifying the file..."
echo "Modified content" >> "$TEST_FILE_PATH"

echo "Checking status immediately after modification..."
status=$(get_file_status "$TEST_FILE_PATH")
echo -n "Status: "
display_status "$status"

sleep 2
echo "Checking status after 2 seconds..."
status=$(get_file_status "$TEST_FILE_PATH")
echo -n "Status: "
display_status "$status"

echo ""
echo -e "${BLUE}=== Test 3: File Read Status ===${NC}"
echo "Reading the file..."
cat "$TEST_FILE_PATH" > /dev/null

echo "Checking status after read..."
status=$(get_file_status "$TEST_FILE_PATH")
echo -n "Status: "
display_status "$status"

echo ""
echo -e "${BLUE}=== Test 4: Directory Creation Status ===${NC}"
TEST_DIR_PATH="$MOUNT_POINT/$TEST_DIR"
echo "Creating a new directory..."
mkdir -p "$TEST_DIR_PATH"

echo "Checking directory status..."
status=$(get_file_status "$TEST_DIR_PATH")
echo -n "Status: "
display_status "$status"

echo ""
echo -e "${BLUE}=== Test 5: File in Directory Status ===${NC}"
TEST_SUBFILE_PATH="$TEST_DIR_PATH/subfile.txt"
echo "Creating file in directory..."
echo "Subfile content" > "$TEST_SUBFILE_PATH"

echo "Checking subfile status..."
status=$(get_file_status "$TEST_SUBFILE_PATH")
echo -n "Status: "
display_status "$status"

echo ""
echo -e "${BLUE}=== Test 6: Extended Attributes Verification ===${NC}"
echo "Checking all extended attributes on test file..."
if command -v getfattr &> /dev/null; then
    echo "Extended attributes for $TEST_FILE_PATH:"
    getfattr -d "$TEST_FILE_PATH" 2>/dev/null || echo "No extended attributes found"
else
    echo -e "${YELLOW}getfattr command not available${NC}"
fi

echo ""
echo -e "${BLUE}=== Test 7: Error Attribute Check ===${NC}"
echo "Checking for error attributes..."
error=$(get_file_error "$TEST_FILE_PATH")
if [ -n "$error" ]; then
    echo -e "${RED}Error found: $error${NC}"
else
    echo -e "${GREEN}No error attribute${NC}"
fi

echo ""
echo -e "${BLUE}=== Test 8: Status Consistency Check ===${NC}"
echo "Checking status multiple times for consistency..."
for i in {1..5}; do
    status=$(get_file_status "$TEST_FILE_PATH")
    echo -n "Check $i: "
    display_status "$status"
    sleep 0.5
done

echo ""
echo -e "${BLUE}=== Test Summary ===${NC}"
echo "Test completed. Review the status changes above."
echo ""
echo "Expected behavior:"
echo "  - New files should show 'LocalModified' or 'Syncing'"
echo "  - Modified files should show 'LocalModified' or 'Syncing'"
echo "  - Read operations should not change status"
echo "  - Status should be consistent across multiple checks"
echo "  - Extended attributes should be set on all files"
echo ""

# Keep mount active for manual inspection
echo -e "${YELLOW}Mount point is still active at: $MOUNT_POINT${NC}"
echo "You can manually inspect files and their status."
echo "Press Enter to unmount and cleanup..."
read

echo -e "${GREEN}Test completed!${NC}"
