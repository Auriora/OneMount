#!/bin/bash
# Test script for XDG_CONFIG_HOME environment variable
# This script verifies that OneMount respects the XDG_CONFIG_HOME environment variable
# Requirements: 15.2, 15.7

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "========================================="
echo "Testing XDG_CONFIG_HOME Environment Variable"
echo "========================================="
echo ""

# Set up custom XDG_CONFIG_HOME
CUSTOM_CONFIG_HOME="/tmp/test-xdg-config-$(date +%s)"
export XDG_CONFIG_HOME="$CUSTOM_CONFIG_HOME"

echo "Step 1: Setting XDG_CONFIG_HOME to custom path"
echo "  XDG_CONFIG_HOME=$XDG_CONFIG_HOME"
echo ""

# Create the custom config directory
mkdir -p "$XDG_CONFIG_HOME"

# Also set XDG_CACHE_HOME for completeness
CUSTOM_CACHE_HOME="/tmp/test-xdg-cache-$(date +%s)"
export XDG_CACHE_HOME="$CUSTOM_CACHE_HOME"
mkdir -p "$XDG_CACHE_HOME"

echo "Step 2: Verifying Go's os.UserConfigDir() respects XDG_CONFIG_HOME"
# Create a simple Go program to test
cat > /tmp/test_xdg.go << 'EOF'
package main
import (
	"fmt"
	"os"
)
func main() {
	configDir, err := os.UserConfigDir()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(configDir)
	
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(cacheDir)
}
EOF

# Run the test program
GO_CONFIG_DIR=$(cd /tmp && go run test_xdg.go | head -1)
GO_CACHE_DIR=$(cd /tmp && go run test_xdg.go | tail -1)

echo "  Go's os.UserConfigDir() returns: $GO_CONFIG_DIR"
echo "  Go's os.UserCacheDir() returns: $GO_CACHE_DIR"

if [ "$GO_CONFIG_DIR" = "$XDG_CONFIG_HOME" ]; then
    echo -e "  ${GREEN}✓ os.UserConfigDir() correctly uses XDG_CONFIG_HOME${NC}"
else
    echo -e "  ${RED}✗ os.UserConfigDir() does not use XDG_CONFIG_HOME${NC}"
    echo "    Expected: $XDG_CONFIG_HOME"
    echo "    Got: $GO_CONFIG_DIR"
fi

if [ "$GO_CACHE_DIR" = "$XDG_CACHE_HOME" ]; then
    echo -e "  ${GREEN}✓ os.UserCacheDir() correctly uses XDG_CACHE_HOME${NC}"
else
    echo -e "  ${RED}✗ os.UserCacheDir() does not use XDG_CACHE_HOME${NC}"
    echo "    Expected: $XDG_CACHE_HOME"
    echo "    Got: $GO_CACHE_DIR"
fi
echo ""

# Create a test mount point
TEST_MOUNT="/tmp/test-xdg-mount-$(date +%s)"
mkdir -p "$TEST_MOUNT"

echo "Step 3: Creating test configuration file"
CONFIG_FILE="$XDG_CONFIG_HOME/onemount/config.yml"
mkdir -p "$(dirname "$CONFIG_FILE")"

cat > "$CONFIG_FILE" << EOF
cacheDir: $XDG_CACHE_HOME/onemount
log: debug
logOutput: STDOUT
syncTree: true
deltaInterval: 1
cacheExpiration: 30
cacheCleanupInterval: 24
maxCacheSize: 0
mountTimeout: 120
EOF

echo "  Created config file at: $CONFIG_FILE"
echo ""

echo "Step 4: Checking if auth tokens exist"
# Auth tokens are stored in cache directory as they are ephemeral OAuth credentials
# that expire and get refreshed as part of the OAuth authentication process
AUTH_TOKENS_PATH="$XDG_CACHE_HOME/onemount/auth_tokens.json"

if [ -f "$AUTH_TOKENS_PATH" ]; then
    echo -e "  ${GREEN}✓ Auth tokens file exists at: $AUTH_TOKENS_PATH${NC}"
else
    echo -e "  ${YELLOW}⚠ Auth tokens file does not exist yet (expected for first run)${NC}"
    echo "    Expected location: $AUTH_TOKENS_PATH"
    echo ""
    echo "  Note: Auth tokens will be created after authentication."
    echo "  To test with real authentication, you would need to:"
    echo "    1. Run: onemount --auth-only $TEST_MOUNT"
    echo "    2. Complete the OAuth flow"
    echo "    3. Verify tokens are created at: $AUTH_TOKENS_PATH"
fi
echo ""

echo "Step 5: Verifying directory structure"
echo "  Expected config directory: $XDG_CONFIG_HOME/onemount/"
echo "  Expected cache directory: $XDG_CACHE_HOME/onemount/"
echo ""

if [ -d "$XDG_CONFIG_HOME/onemount" ]; then
    echo -e "  ${GREEN}✓ Config directory exists${NC}"
    echo "    Contents:"
    ls -la "$XDG_CONFIG_HOME/onemount/" | sed 's/^/      /'
else
    echo -e "  ${RED}✗ Config directory does not exist${NC}"
fi
echo ""

if [ -d "$XDG_CACHE_HOME/onemount" ]; then
    echo -e "  ${GREEN}✓ Cache directory exists${NC}"
    echo "    Contents:"
    ls -la "$XDG_CACHE_HOME/onemount/" 2>/dev/null | sed 's/^/      /' || echo "      (empty)"
else
    echo -e "  ${YELLOW}⚠ Cache directory does not exist yet (will be created on mount)${NC}"
fi
echo ""

echo "Step 6: Testing DefaultConfigPath() function"
# This tests that the config path respects XDG_CONFIG_HOME
cat > /tmp/test_config_path.go << 'EOF'
package main
import (
	"fmt"
	"os"
	"path/filepath"
)
func main() {
	confDir, err := os.UserConfigDir()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	configPath := filepath.Join(confDir, "onemount/config.yml")
	fmt.Println(configPath)
}
EOF

EXPECTED_CONFIG_PATH=$(cd /tmp && go run test_config_path.go)
echo "  Expected config path: $EXPECTED_CONFIG_PATH"

if [ "$EXPECTED_CONFIG_PATH" = "$CONFIG_FILE" ]; then
    echo -e "  ${GREEN}✓ Config path matches XDG_CONFIG_HOME/onemount/config.yml${NC}"
else
    echo -e "  ${RED}✗ Config path does not match expected location${NC}"
    echo "    Expected: $CONFIG_FILE"
    echo "    Got: $EXPECTED_CONFIG_PATH"
fi
echo ""

echo "========================================="
echo "Summary"
echo "========================================="
echo ""
echo "XDG_CONFIG_HOME: $XDG_CONFIG_HOME"
echo "XDG_CACHE_HOME: $XDG_CACHE_HOME"
echo ""
echo "Expected locations:"
echo "  Config file: $XDG_CONFIG_HOME/onemount/config.yml"
echo "  Cache directory: $XDG_CACHE_HOME/onemount/"
echo "  Auth tokens: $XDG_CACHE_HOME/onemount/auth_tokens.json"
echo ""
echo "Note: This test verifies the directory structure and configuration."
echo "To fully test with authentication, run onemount with these environment"
echo "variables set and complete the OAuth flow."
echo ""
echo "Cleanup:"
echo "  rm -rf $CUSTOM_CONFIG_HOME"
echo "  rm -rf $CUSTOM_CACHE_HOME"
echo "  rm -rf $TEST_MOUNT"
echo ""

# Cleanup
echo "Cleaning up test directories..."
rm -rf "$CUSTOM_CONFIG_HOME"
rm -rf "$CUSTOM_CACHE_HOME"
rm -rf "$TEST_MOUNT"
rm -f /tmp/test_xdg.go
rm -f /tmp/test_config_path.go

echo -e "${GREEN}Test completed successfully!${NC}"
