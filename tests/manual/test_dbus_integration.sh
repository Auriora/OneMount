#!/bin/bash
# Manual test script for D-Bus integration
# Tests D-Bus signal emission and monitoring
# Requirements: 8.2

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test configuration
MOUNT_POINT="/tmp/onemount-dbus-test"
CACHE_DIR="/tmp/onemount-dbus-cache"
TEST_FILE="dbus-test-file.txt"
DBUS_MONITOR_LOG="/tmp/dbus-monitor.log"

echo -e "${BLUE}=== D-Bus Integration Test ===${NC}"
echo "This test monitors D-Bus signals during file operations"
echo ""

# Check if dbus-monitor is available
if ! command -v dbus-monitor &> /dev/null; then
    echo -e "${RED}Error: dbus-monitor command not found${NC}"
    echo "Please install dbus-monitor: sudo apt-get install dbus"
    exit 1
fi

# Check if dbus-send is available
if ! command -v dbus-send &> /dev/null; then
    echo -e "${YELLOW}Warning: dbus-send command not found${NC}"
    echo "Some tests will be skipped"
fi

# Cleanup function
cleanup() {
    echo -e "${YELLOW}Cleaning up...${NC}"
    
    # Stop dbus-monitor if running
    if [ -n "$DBUS_MONITOR_PID" ] && kill -0 "$DBUS_MONITOR_PID" 2>/dev/null; then
        kill "$DBUS_MONITOR_PID" 2>/dev/null || true
    fi
    
    # Unmount filesystem
    if mountpoint -q "$MOUNT_POINT" 2>/dev/null; then
        fusermount -u "$MOUNT_POINT" 2>/dev/null || true
        sleep 1
    fi
    
    rm -rf "$MOUNT_POINT" "$CACHE_DIR" "$DBUS_MONITOR_LOG"
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

echo ""
echo -e "${BLUE}=== Test 1: D-Bus Service Discovery ===${NC}"
echo "Checking for existing OneMount D-Bus services..."

# List all D-Bus services matching onemount
if command -v dbus-send &> /dev/null; then
    echo "Querying D-Bus for OneMount services..."
    dbus-send --session --print-reply --dest=org.freedesktop.DBus \
        /org/freedesktop/DBus org.freedesktop.DBus.ListNames 2>/dev/null | \
        grep -i onemount || echo "No OneMount services found (expected before mount)"
else
    echo -e "${YELLOW}Skipping: dbus-send not available${NC}"
fi

echo ""
echo -e "${BLUE}=== Test 2: Start D-Bus Monitor ===${NC}"
echo "Starting dbus-monitor to capture signals..."

# Start dbus-monitor in background
dbus-monitor "type='signal',interface='org.onemount.FileStatus'" > "$DBUS_MONITOR_LOG" 2>&1 &
DBUS_MONITOR_PID=$!

echo "D-Bus monitor started (PID: $DBUS_MONITOR_PID)"
echo "Monitoring interface: org.onemount.FileStatus"
echo "Log file: $DBUS_MONITOR_LOG"
sleep 1

echo ""
echo -e "${BLUE}=== Test 3: Mount Filesystem ===${NC}"
echo "Mounting OneMount filesystem..."
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
echo -e "${BLUE}=== Test 4: Check D-Bus Service Registration ===${NC}"
if command -v dbus-send &> /dev/null; then
    echo "Checking if OneMount registered a D-Bus service..."
    dbus-send --session --print-reply --dest=org.freedesktop.DBus \
        /org/freedesktop/DBus org.freedesktop.DBus.ListNames 2>/dev/null | \
        grep -i onemount || echo -e "${YELLOW}No OneMount service found${NC}"
    
    echo ""
    echo "Attempting to introspect OneMount D-Bus service..."
    # Try base name first
    dbus-send --session --print-reply --dest=org.onemount.FileStatus \
        /org/onemount/FileStatus org.freedesktop.DBus.Introspectable.Introspect 2>/dev/null || \
        echo -e "${YELLOW}Could not introspect base service name${NC}"
else
    echo -e "${YELLOW}Skipping: dbus-send not available${NC}"
fi

echo ""
echo -e "${BLUE}=== Test 5: Trigger File Operations ===${NC}"
echo "Creating a file to trigger D-Bus signals..."
TEST_FILE_PATH="$MOUNT_POINT/$TEST_FILE"
echo "Test content for D-Bus monitoring" > "$TEST_FILE_PATH"
echo "File created: $TEST_FILE_PATH"

sleep 2

echo ""
echo "Modifying the file..."
echo "Modified content" >> "$TEST_FILE_PATH"
echo "File modified"

sleep 2

echo ""
echo "Reading the file..."
cat "$TEST_FILE_PATH" > /dev/null
echo "File read"

sleep 2

echo ""
echo -e "${BLUE}=== Test 6: Analyze D-Bus Monitor Output ===${NC}"
echo "Stopping D-Bus monitor..."
if [ -n "$DBUS_MONITOR_PID" ] && kill -0 "$DBUS_MONITOR_PID" 2>/dev/null; then
    kill "$DBUS_MONITOR_PID" 2>/dev/null || true
    sleep 1
fi

echo ""
echo "D-Bus monitor log contents:"
echo "---"
if [ -f "$DBUS_MONITOR_LOG" ] && [ -s "$DBUS_MONITOR_LOG" ]; then
    cat "$DBUS_MONITOR_LOG"
    echo "---"
    
    # Analyze the log
    signal_count=$(grep -c "signal" "$DBUS_MONITOR_LOG" 2>/dev/null || echo "0")
    filestatus_count=$(grep -c "FileStatusChanged" "$DBUS_MONITOR_LOG" 2>/dev/null || echo "0")
    
    echo ""
    echo "Analysis:"
    echo "  Total D-Bus signals: $signal_count"
    echo "  FileStatusChanged signals: $filestatus_count"
    
    if [ "$filestatus_count" -gt 0 ]; then
        echo -e "  ${GREEN}✓ D-Bus signals detected!${NC}"
    else
        echo -e "  ${YELLOW}⚠ No FileStatusChanged signals detected${NC}"
        echo "  This may indicate:"
        echo "    - D-Bus server not started"
        echo "    - Service name mismatch"
        echo "    - Signals not being emitted"
    fi
else
    echo "(empty or not found)"
    echo "---"
    echo -e "${YELLOW}⚠ No D-Bus monitor output captured${NC}"
fi

echo ""
echo -e "${BLUE}=== Test 7: Check Extended Attributes ===${NC}"
echo "Verifying extended attributes are set (D-Bus fallback)..."
if command -v getfattr &> /dev/null; then
    echo "Extended attributes for $TEST_FILE_PATH:"
    getfattr -d "$TEST_FILE_PATH" 2>/dev/null || echo "No extended attributes found"
else
    echo -e "${YELLOW}getfattr command not available${NC}"
fi

echo ""
echo -e "${BLUE}=== Test Summary ===${NC}"
echo "Test completed. Review the D-Bus monitor output above."
echo ""
echo "Expected behavior:"
echo "  - D-Bus service should be registered on mount"
echo "  - FileStatusChanged signals should be emitted on file operations"
echo "  - Signal format: FileStatusChanged(path, status)"
echo "  - Extended attributes should be set as fallback"
echo ""

if [ "$filestatus_count" -gt 0 ]; then
    echo -e "${GREEN}✓ D-Bus integration appears to be working${NC}"
else
    echo -e "${YELLOW}⚠ D-Bus signals not detected - check service name and registration${NC}"
    echo ""
    echo "Troubleshooting tips:"
    echo "  1. Check if D-Bus session bus is running: echo \$DBUS_SESSION_BUS_ADDRESS"
    echo "  2. List all D-Bus services: dbus-send --session --print-reply --dest=org.freedesktop.DBus /org/freedesktop/DBus org.freedesktop.DBus.ListNames"
    echo "  3. Check OneMount logs for D-Bus errors"
    echo "  4. Verify D-Bus service name matches between server and monitor"
fi

echo ""
echo -e "${YELLOW}Mount point is still active at: $MOUNT_POINT${NC}"
echo "You can manually inspect files and D-Bus activity."
echo "Press Enter to unmount and cleanup..."
read

echo -e "${GREEN}Test completed!${NC}"
