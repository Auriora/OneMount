#!/bin/bash
# D-Bus Integration Test for Docker Environment
# Tests D-Bus signal emission and monitoring in Docker
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
DBUS_MONITOR_LOG="/tmp/dbus-monitor.log"

echo -e "${BLUE}=== D-Bus Integration Test in Docker ===${NC}"
echo "Mount point: $MOUNT_POINT"
echo ""

# Check if we're in Docker
if [ ! -f /.dockerenv ]; then
    echo -e "${RED}Error: This script must be run inside Docker${NC}"
    echo "Use: docker compose -f docker/compose/docker-compose.test.yml -f docker/compose/docker-compose.auth.yml run --rm shell"
    exit 1
fi

# Check if auth tokens are mounted
if [ ! -f "/tmp/auth-tokens/auth_tokens.json" ]; then
    echo -e "${RED}Error: Auth tokens not found at /tmp/auth-tokens/auth_tokens.json${NC}"
    echo "Make sure you're using the auth override: -f docker/compose/docker-compose.auth.yml"
    exit 1
fi

echo -e "${GREEN}✓ Auth tokens found${NC}"
ACCOUNT=$(jq -r '.account // "unknown"' /tmp/auth-tokens/auth_tokens.json 2>/dev/null || echo "unknown")
echo "Account: $ACCOUNT"
echo ""

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
    
    # Kill mount process if still running
    if [ -n "$MOUNT_PID" ] && kill -0 "$MOUNT_PID" 2>/dev/null; then
        kill "$MOUNT_PID" 2>/dev/null || true
        sleep 1
    fi
    
    # Stop D-Bus session bus if we started it
    if [ -n "$DBUS_SESSION_BUS_PID" ] && kill -0 "$DBUS_SESSION_BUS_PID" 2>/dev/null; then
        kill "$DBUS_SESSION_BUS_PID" 2>/dev/null || true
    fi
    
    rm -rf "$MOUNT_POINT"
}

trap cleanup EXIT

# Setup
mkdir -p "$MOUNT_POINT"

# Start D-Bus session bus if not already running
echo -e "${BLUE}=== Test 0: Setup D-Bus Session Bus ===${NC}"
if [ -z "$DBUS_SESSION_BUS_ADDRESS" ] || ! dbus-send --session --print-reply --dest=org.freedesktop.DBus /org/freedesktop/DBus org.freedesktop.DBus.GetId &>/dev/null; then
    echo "Starting D-Bus session bus..."
    mkdir -p /tmp/runtime-tester
    eval $(dbus-launch --sh-syntax)
    export DBUS_SESSION_BUS_ADDRESS
    export DBUS_SESSION_BUS_PID
    echo "D-Bus session bus started"
    echo "  Address: $DBUS_SESSION_BUS_ADDRESS"
    echo "  PID: $DBUS_SESSION_BUS_PID"
else
    echo "D-Bus session bus already running"
    echo "  Address: $DBUS_SESSION_BUS_ADDRESS"
fi
echo ""

echo -e "${BLUE}=== Test 1: Start D-Bus Monitor ===${NC}"
echo "Starting dbus-monitor to capture signals..."

# Start dbus-monitor in background
dbus-monitor "type='signal',interface='org.onemount.FileStatus'" > "$DBUS_MONITOR_LOG" 2>&1 &
DBUS_MONITOR_PID=$!

echo "D-Bus monitor started (PID: $DBUS_MONITOR_PID)"
echo "Monitoring interface: org.onemount.FileStatus"
echo "Log file: $DBUS_MONITOR_LOG"
sleep 2

echo ""
echo -e "${BLUE}=== Test 2: Prepare Authentication Tokens ===${NC}"
# Calculate the expected token path using systemd-escape
CACHE_BASE="$HOME/.cache/onemount"
ESCAPED_PATH=$(systemd-escape --path "$MOUNT_POINT")
EXPECTED_TOKEN_PATH="$CACHE_BASE/$ESCAPED_PATH/auth_tokens.json"

echo "Mount point: $MOUNT_POINT"
echo "Escaped path: $ESCAPED_PATH"
echo "Expected token path: $EXPECTED_TOKEN_PATH"

# Create the cache directory and copy tokens
mkdir -p "$(dirname "$EXPECTED_TOKEN_PATH")"
cp /tmp/auth-tokens/auth_tokens.json "$EXPECTED_TOKEN_PATH"
chmod 600 "$EXPECTED_TOKEN_PATH"

echo -e "${GREEN}✓ Tokens copied to expected location${NC}"
ls -la "$EXPECTED_TOKEN_PATH"
echo ""

echo -e "${BLUE}=== Test 3: Mount Filesystem ===${NC}"
echo "Mounting OneMount filesystem..."

# Mount using the pre-built binary
# It will find the tokens at the expected location
./build/onemount-nogui \
    --no-sync-tree \
    "$MOUNT_POINT" > /tmp/onemount-mount.log 2>&1 &
MOUNT_PID=$!

echo "Mount PID: $MOUNT_PID"

# Wait for mount to complete
echo "Waiting for mount to complete..."
for i in {1..30}; do
    if mountpoint -q "$MOUNT_POINT" 2>/dev/null; then
        echo -e "${GREEN}✓ Mount successful!${NC}"
        break
    fi
    sleep 1
    if [ $i -eq 30 ]; then
        echo -e "${RED}✗ Mount timeout after 30 seconds${NC}"
        echo ""
        echo -e "${YELLOW}=== Mount log ===${NC}"
        cat /tmp/onemount-mount.log
        exit 1
    fi
done

sleep 3

echo ""
echo -e "${BLUE}=== Test 4: Check D-Bus Service Registration ===${NC}"
echo "Checking if OneMount registered a D-Bus service..."
dbus-send --session --print-reply --dest=org.freedesktop.DBus \
    /org/freedesktop/DBus org.freedesktop.DBus.ListNames 2>/dev/null | \
    grep -i onemount || echo -e "${YELLOW}⚠ No OneMount service found${NC}"

echo ""
echo -e "${BLUE}=== Test 5: Trigger File Operations ===${NC}"
echo "Listing directory to trigger operations..."
ls -la "$MOUNT_POINT" 2>&1 | head -10

file_count=$(ls -1 "$MOUNT_POINT" 2>/dev/null | wc -l)
echo "Files found: $file_count"

if [ "$file_count" -gt 0 ]; then
    first_file=$(ls -1 "$MOUNT_POINT" 2>/dev/null | head -1)
    echo ""
    echo "Accessing file: $first_file"
    
    # Stat the file
    if stat "$MOUNT_POINT/$first_file" > /dev/null 2>&1; then
        echo -e "${GREEN}✓ stat operation successful${NC}"
    else
        echo -e "${YELLOW}⚠ stat operation failed${NC}"
    fi
    
    sleep 2
    
    # Read the file (if it's not too large)
    file_size=$(stat -c%s "$MOUNT_POINT/$first_file" 2>/dev/null || echo "0")
    if [ "$file_size" -lt 1048576 ]; then  # Less than 1MB
        if cat "$MOUNT_POINT/$first_file" > /dev/null 2>&1; then
            echo -e "${GREEN}✓ read operation successful${NC}"
        else
            echo -e "${YELLOW}⚠ read operation failed${NC}"
        fi
    else
        echo "File too large ($file_size bytes), skipping read"
    fi
    
    sleep 2
    
    # Check extended attributes
    echo ""
    echo "Checking extended attributes..."
    if command -v getfattr &> /dev/null; then
        getfattr -d "$MOUNT_POINT/$first_file" 2>/dev/null || echo "No extended attributes found"
    else
        echo -e "${YELLOW}getfattr command not available${NC}"
    fi
else
    echo -e "${YELLOW}⚠ No files found in mount point${NC}"
fi

sleep 3

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
    filestatus_count=$(grep -c "FileStatus" "$DBUS_MONITOR_LOG" 2>/dev/null || echo "0")
    changed_count=$(grep -c "FileStatusChanged" "$DBUS_MONITOR_LOG" 2>/dev/null || echo "0")
    
    echo ""
    echo "Analysis:"
    echo "  Total D-Bus signals: $signal_count"
    echo "  FileStatus signals: $filestatus_count"
    echo "  FileStatusChanged signals: $changed_count"
    
    if [ "$changed_count" -gt 0 ]; then
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
echo -e "${BLUE}=== Test Summary ===${NC}"
echo "Test completed. Review the D-Bus monitor output above."
echo ""
echo "Expected behavior:"
echo "  - D-Bus service should be registered on mount"
echo "  - FileStatusChanged signals should be emitted on file operations"
echo "  - Signal format: FileStatusChanged(path, status)"
echo "  - Extended attributes should be set as fallback"
echo ""

if [ "$changed_count" -gt 0 ]; then
    echo -e "${GREEN}✓ D-Bus integration appears to be working${NC}"
    exit 0
else
    echo -e "${YELLOW}⚠ D-Bus signals not detected - check service name and registration${NC}"
    echo ""
    echo "Troubleshooting tips:"
    echo "  1. Check mount log: cat /tmp/onemount-mount.log"
    echo "  2. List all D-Bus services: dbus-send --session --print-reply --dest=org.freedesktop.DBus /org/freedesktop/DBus org.freedesktop.DBus.ListNames"
    echo "  3. Check OneMount logs for D-Bus errors"
    echo "  4. Verify D-Bus service name matches between server and monitor"
    exit 1
fi
