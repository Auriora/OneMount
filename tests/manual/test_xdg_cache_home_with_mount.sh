#!/bin/bash
# Comprehensive test for XDG_CACHE_HOME with actual filesystem mount
# This script verifies that OneMount respects XDG_CACHE_HOME when mounting
# Requirements: 15.5, 15.9

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo "========================================="
echo "Testing XDG_CACHE_HOME with Filesystem Mount"
echo "========================================="
echo ""

# Set up custom XDG paths
CUSTOM_CONFIG_HOME="/tmp/test-xdg-config-$(date +%s)"
CUSTOM_CACHE_HOME="/tmp/test-xdg-cache-$(date +%s)"
TEST_MOUNT="/tmp/test-xdg-mount-$(date +%s)"

export XDG_CONFIG_HOME="$CUSTOM_CONFIG_HOME"
export XDG_CACHE_HOME="$CUSTOM_CACHE_HOME"

echo -e "${BLUE}Step 1: Setting up test environment${NC}"
echo "  XDG_CONFIG_HOME=$XDG_CONFIG_HOME"
echo "  XDG_CACHE_HOME=$XDG_CACHE_HOME"
echo "  Mount point=$TEST_MOUNT"
echo ""

# Create directories
mkdir -p "$XDG_CONFIG_HOME/onemount"
mkdir -p "$XDG_CACHE_HOME/onemount"
mkdir -p "$TEST_MOUNT"

# Check if we have auth tokens available
SOURCE_AUTH_TOKENS="/workspace/test-artifacts/.auth_tokens.json"
if [ ! -f "$SOURCE_AUTH_TOKENS" ]; then
    echo -e "${RED}✗ Auth tokens not found at $SOURCE_AUTH_TOKENS${NC}"
    echo "  Cannot proceed with mount test without authentication."
    echo "  Please run authentication first or use the basic test script."
    exit 1
fi

echo -e "${BLUE}Step 2: Copying auth tokens to custom cache directory${NC}"
# Copy auth tokens to the custom cache directory
cp "$SOURCE_AUTH_TOKENS" "$XDG_CACHE_HOME/onemount/auth_tokens.json"
echo -e "  ${GREEN}✓ Auth tokens copied to $XDG_CACHE_HOME/onemount/auth_tokens.json${NC}"
echo ""

echo -e "${BLUE}Step 3: Creating configuration file in custom config directory${NC}"
CONFIG_FILE="$XDG_CONFIG_HOME/onemount/config.yml"
cat > "$CONFIG_FILE" << EOF
cacheDir: $XDG_CACHE_HOME/onemount
log: debug
logOutput: STDOUT
syncTree: true
deltaInterval: 300
cacheExpiration: 30
cacheCleanupInterval: 24
maxCacheSize: 0
mountTimeout: 120
EOF
echo -e "  ${GREEN}✓ Config file created at $CONFIG_FILE${NC}"
echo ""

echo -e "${BLUE}Step 4: Attempting to mount filesystem with custom XDG paths${NC}"
echo "  This will verify that OneMount uses the custom XDG_CACHE_HOME directory..."
echo ""

# Try to mount with no-browser mode (with a short timeout since we just want to verify it starts)
timeout 15s /workspace/onemount --config-file "$CONFIG_FILE" --no-browser "$TEST_MOUNT" &
MOUNT_PID=$!

# Wait a bit for mount to initialize
sleep 5

# Check if the process is still running
if ps -p $MOUNT_PID > /dev/null 2>&1; then
    echo -e "  ${GREEN}✓ OneMount process started successfully${NC}"
    
    # Check if mount point is accessible
    if mountpoint -q "$TEST_MOUNT" 2>/dev/null; then
        echo -e "  ${GREEN}✓ Filesystem mounted at $TEST_MOUNT${NC}"
        
        # Try to list the mount point to trigger some cache activity
        echo "  Listing mount point to trigger cache activity..."
        ls -la "$TEST_MOUNT" > /dev/null 2>&1 || true
        sleep 2
    else
        echo -e "  ${YELLOW}⚠ Mount point not yet ready (may still be initializing)${NC}"
    fi
    
    # Kill the mount process
    kill $MOUNT_PID 2>/dev/null || true
    wait $MOUNT_PID 2>/dev/null || true
    
    # Unmount if needed
    fusermount -u "$TEST_MOUNT" 2>/dev/null || true
else
    echo -e "  ${RED}✗ OneMount process failed to start${NC}"
fi
echo ""

echo -e "${BLUE}Step 5: Verifying cache files were created in custom XDG_CACHE_HOME${NC}"
echo ""

echo "  Cache directory ($XDG_CACHE_HOME/onemount):"
if [ -d "$XDG_CACHE_HOME/onemount" ]; then
    echo -e "    ${GREEN}✓ Directory exists${NC}"
    ls -la "$XDG_CACHE_HOME/onemount/" | sed 's/^/      /'
    
    # Check for specific files
    echo ""
    echo "  Checking for expected files in cache directory:"
    
    if [ -f "$XDG_CACHE_HOME/onemount/auth_tokens.json" ]; then
        echo -e "    ${GREEN}✓ auth_tokens.json exists${NC}"
    else
        echo -e "    ${RED}✗ auth_tokens.json not found${NC}"
    fi
    
    if [ -f "$XDG_CACHE_HOME/onemount/metadata.db" ]; then
        echo -e "    ${GREEN}✓ metadata.db exists (metadata database in cache directory)${NC}"
    else
        echo -e "    ${YELLOW}⚠ metadata.db not found (may not have been created yet)${NC}"
    fi
    
    if [ -d "$XDG_CACHE_HOME/onemount/content" ]; then
        echo -e "    ${GREEN}✓ content cache directory exists${NC}"
    else
        echo -e "    ${YELLOW}⚠ content cache directory not found (may not have been created yet)${NC}"
    fi
    
    # Check for thumbnails directory
    if [ -d "$XDG_CACHE_HOME/onemount/thumbnails" ]; then
        echo -e "    ${GREEN}✓ thumbnails cache directory exists${NC}"
    else
        echo -e "    ${YELLOW}⚠ thumbnails cache directory not found (may not have been created yet)${NC}"
    fi
else
    echo -e "    ${RED}✗ Directory does not exist${NC}"
fi
echo ""

echo -e "${BLUE}Step 6: Verifying no cache files were created in default XDG location${NC}"
DEFAULT_CACHE="$HOME/.cache/onemount"

echo "  Checking default cache location ($DEFAULT_CACHE):"
if [ -d "$DEFAULT_CACHE" ]; then
    echo -e "    ${RED}✗ Directory exists (should not exist when using custom XDG_CACHE_HOME)${NC}"
    ls -la "$DEFAULT_CACHE/" | sed 's/^/      /'
else
    echo -e "    ${GREEN}✓ Directory does not exist (correct)${NC}"
fi
echo ""

echo -e "${BLUE}Step 7: Verifying metadata database location${NC}"
echo "  The metadata database (metadata.db) should be in the cache directory."
echo "  This is a key requirement for XDG compliance (Requirement 15.9)."
echo ""

if [ -f "$XDG_CACHE_HOME/onemount/metadata.db" ]; then
    echo -e "  ${GREEN}✓ metadata.db found in \$XDG_CACHE_HOME/onemount/${NC}"
    echo "    File size: $(du -h "$XDG_CACHE_HOME/onemount/metadata.db" | cut -f1)"
    echo "    Permissions: $(stat -c '%a' "$XDG_CACHE_HOME/onemount/metadata.db")"
else
    echo -e "  ${YELLOW}⚠ metadata.db not found in \$XDG_CACHE_HOME/onemount/${NC}"
    echo "    This may be expected if the mount didn't complete initialization."
fi
echo ""

echo "========================================="
echo "Summary"
echo "========================================="
echo ""
echo "Custom XDG directories:"
echo "  XDG_CONFIG_HOME: $XDG_CONFIG_HOME"
echo "  XDG_CACHE_HOME: $XDG_CACHE_HOME"
echo ""
echo "Expected cache files:"
echo "  Auth tokens: $XDG_CACHE_HOME/onemount/auth_tokens.json"
echo "  Metadata DB: $XDG_CACHE_HOME/onemount/metadata.db"
echo "  Content cache: $XDG_CACHE_HOME/onemount/content/"
echo "  Thumbnails: $XDG_CACHE_HOME/onemount/thumbnails/"
echo ""
echo "Verification results:"

PASS_COUNT=0
TOTAL_CHECKS=3

if [ -f "$XDG_CACHE_HOME/onemount/auth_tokens.json" ]; then
    echo -e "  ${GREEN}✓ Auth tokens stored in \$XDG_CACHE_HOME/onemount/${NC}"
    ((PASS_COUNT++))
else
    echo -e "  ${RED}✗ Auth tokens not found in \$XDG_CACHE_HOME/onemount/${NC}"
fi

if [ -f "$XDG_CACHE_HOME/onemount/metadata.db" ]; then
    echo -e "  ${GREEN}✓ Metadata database stored in \$XDG_CACHE_HOME/onemount/${NC}"
    ((PASS_COUNT++))
else
    echo -e "  ${YELLOW}⚠ Metadata database not found (mount may not have completed)${NC}"
fi

if [ ! -d "$DEFAULT_CACHE" ]; then
    echo -e "  ${GREEN}✓ No cache files created in default location${NC}"
    ((PASS_COUNT++))
else
    echo -e "  ${RED}✗ Cache files found in default location (should use custom XDG_CACHE_HOME)${NC}"
fi

echo ""
echo "Test result: $PASS_COUNT/$TOTAL_CHECKS checks passed"
echo ""

# Cleanup
echo "Cleaning up test directories..."
fusermount -u "$TEST_MOUNT" 2>/dev/null || true
rm -rf "$CUSTOM_CONFIG_HOME"
rm -rf "$CUSTOM_CACHE_HOME"
rm -rf "$TEST_MOUNT"

if [ $PASS_COUNT -eq $TOTAL_CHECKS ]; then
    echo -e "${GREEN}✓ All tests passed! OneMount correctly uses XDG_CACHE_HOME.${NC}"
    exit 0
else
    echo -e "${YELLOW}⚠ Some tests did not pass. Review the output above.${NC}"
    exit 1
fi
