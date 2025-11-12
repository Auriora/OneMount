#!/bin/bash
# Setup persistent authentication tokens for testing
# This script helps configure auth tokens that work across test runs and Docker

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

print_info "OneMount Test Authentication Setup"
print_info "===================================="
echo ""

# Standard test token location
TEST_TOKEN_DIR="$HOME/.onemount-tests"
TEST_TOKEN_FILE="$TEST_TOKEN_DIR/.auth_tokens.json"

# Check if tokens already exist
if [ -f "$TEST_TOKEN_FILE" ]; then
    print_info "Existing tokens found at: $TEST_TOKEN_FILE"
    
    # Check if valid JSON
    if jq empty "$TEST_TOKEN_FILE" 2>/dev/null; then
        print_success "Tokens are valid JSON"
        
        # Check expiration
        EXPIRES_AT=$(jq -r '.expires_at' "$TEST_TOKEN_FILE" 2>/dev/null || echo "0")
        CURRENT_TIME=$(date +%s)
        ACCOUNT=$(jq -r '.account' "$TEST_TOKEN_FILE" 2>/dev/null || echo "unknown")
        
        print_info "Account: $ACCOUNT"
        
        if [ "$EXPIRES_AT" != "0" ] && [ "$EXPIRES_AT" -gt "$CURRENT_TIME" ]; then
            EXPIRES_DATE=$(date -d "@$EXPIRES_AT" 2>/dev/null || date -r "$EXPIRES_AT" 2>/dev/null || echo "unknown")
            print_success "Tokens are valid until: $EXPIRES_DATE"
            echo ""
            print_info "Your authentication is already set up!"
            print_info "You can run tests without re-authenticating."
            echo ""
            print_info "To refresh tokens, run: $0 --refresh"
            exit 0
        else
            print_warning "Tokens are EXPIRED"
            print_info "Will need to re-authenticate"
        fi
    else
        print_error "Existing tokens are not valid JSON"
        print_info "Will need to re-authenticate"
    fi
    echo ""
fi

# Check for existing tokens in other locations
print_info "Searching for existing auth tokens..."
FOUND_TOKEN=""

# Search in cache directories
for cache_dir in ~/.cache/onemount/*/auth_tokens.json; do
    if [ -f "$cache_dir" ]; then
        print_info "Found tokens in cache: $cache_dir"
        
        # Check if valid and not expired
        if jq empty "$cache_dir" 2>/dev/null; then
            EXPIRES_AT=$(jq -r '.expires_at' "$cache_dir" 2>/dev/null || echo "0")
            CURRENT_TIME=$(date +%s)
            
            if [ "$EXPIRES_AT" != "0" ] && [ "$EXPIRES_AT" -gt "$CURRENT_TIME" ]; then
                FOUND_TOKEN="$cache_dir"
                print_success "Found valid tokens!"
                break
            fi
        fi
    fi
done

# Check project locations
for proj_loc in "./test-artifacts/.auth_tokens.json" "./auth_tokens.json"; do
    if [ -f "$proj_loc" ]; then
        print_info "Found tokens in project: $proj_loc"
        
        if jq empty "$proj_loc" 2>/dev/null; then
            EXPIRES_AT=$(jq -r '.expires_at' "$proj_loc" 2>/dev/null || echo "0")
            CURRENT_TIME=$(date +%s)
            
            if [ "$EXPIRES_AT" != "0" ] && [ "$EXPIRES_AT" -gt "$CURRENT_TIME" ]; then
                FOUND_TOKEN="$proj_loc"
                print_success "Found valid tokens!"
                break
            fi
        fi
    fi
done

if [ -n "$FOUND_TOKEN" ]; then
    echo ""
    print_success "Found valid authentication tokens!"
    print_info "Copying to standard test location..."
    
    # Create directory
    mkdir -p "$TEST_TOKEN_DIR"
    
    # Copy tokens
    cp "$FOUND_TOKEN" "$TEST_TOKEN_FILE"
    chmod 600 "$TEST_TOKEN_FILE"
    
    print_success "Tokens copied to: $TEST_TOKEN_FILE"
    
    # Show account info
    ACCOUNT=$(jq -r '.account' "$TEST_TOKEN_FILE")
    EXPIRES_AT=$(jq -r '.expires_at' "$TEST_TOKEN_FILE")
    EXPIRES_DATE=$(date -d "@$EXPIRES_AT" 2>/dev/null || date -r "$EXPIRES_AT" 2>/dev/null || echo "unknown")
    
    echo ""
    print_info "Authentication Details:"
    print_info "  Account: $ACCOUNT"
    print_info "  Expires: $EXPIRES_DATE"
    echo ""
    print_success "Setup complete! You can now run tests without re-authenticating."
    exit 0
fi

# No valid tokens found, need to authenticate
echo ""
print_warning "No valid authentication tokens found"
print_info "You need to authenticate with your Microsoft account"
echo ""
print_info "This will:"
print_info "  1. Build OneMount (if needed)"
print_info "  2. Open a browser for Microsoft login"
print_info "  3. Mount your OneDrive temporarily"
print_info "  4. Save tokens for future test runs"
echo ""

read -p "Continue with authentication? (y/n) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    print_info "Authentication cancelled"
    exit 0
fi

# Check if binary exists
if [ ! -f "./build/onemount" ]; then
    print_info "Building OneMount..."
    if ! make build; then
        print_error "Failed to build OneMount"
        exit 1
    fi
    print_success "Build complete"
fi

# Create temporary directories
TEMP_MOUNT=$(mktemp -d)
TEMP_CACHE=$(mktemp -d)

print_info "Temporary mount: $TEMP_MOUNT"
print_info "Temporary cache: $TEMP_CACHE"
echo ""

# Cleanup function
cleanup() {
    print_info "Cleaning up..."
    
    # Unmount if mounted
    if mountpoint -q "$TEMP_MOUNT" 2>/dev/null; then
        fusermount3 -uz "$TEMP_MOUNT" 2>/dev/null || true
    fi
    
    # Remove temp directories
    rm -rf "$TEMP_MOUNT"
    
    # Don't remove cache yet - we need to extract tokens
}

trap cleanup EXIT

# Run OneMount to authenticate
print_info "Starting OneMount for authentication..."
print_info "A browser window will open for Microsoft login"
echo ""

# Start OneMount in background
./build/onemount --cache-dir="$TEMP_CACHE" "$TEMP_MOUNT" &
MOUNT_PID=$!

# Wait for mount or timeout
WAIT_COUNT=0
MAX_WAIT=120

while [ $WAIT_COUNT -lt $MAX_WAIT ]; do
    if mountpoint -q "$TEMP_MOUNT" 2>/dev/null; then
        print_success "Mount successful!"
        break
    fi
    
    if ! kill -0 $MOUNT_PID 2>/dev/null; then
        print_error "OneMount process died"
        exit 1
    fi
    
    sleep 1
    WAIT_COUNT=$((WAIT_COUNT + 1))
done

if [ $WAIT_COUNT -ge $MAX_WAIT ]; then
    print_error "Mount timeout after ${MAX_WAIT}s"
    kill $MOUNT_PID 2>/dev/null || true
    exit 1
fi

# Give it a moment to stabilize
sleep 2

# Kill the mount process
print_info "Stopping OneMount..."
kill $MOUNT_PID 2>/dev/null || true
sleep 2

# Unmount
if mountpoint -q "$TEMP_MOUNT" 2>/dev/null; then
    fusermount3 -uz "$TEMP_MOUNT" 2>/dev/null || true
fi

# Find and copy auth tokens
print_info "Extracting authentication tokens..."

TOKEN_FOUND=false
for token_file in "$TEMP_CACHE"/*/auth_tokens.json; do
    if [ -f "$token_file" ]; then
        print_success "Found tokens: $token_file"
        
        # Verify JSON
        if jq empty "$token_file" 2>/dev/null; then
            # Create test token directory
            mkdir -p "$TEST_TOKEN_DIR"
            
            # Copy tokens
            cp "$token_file" "$TEST_TOKEN_FILE"
            chmod 600 "$TEST_TOKEN_FILE"
            
            print_success "Tokens saved to: $TEST_TOKEN_FILE"
            TOKEN_FOUND=true
            
            # Show account info
            ACCOUNT=$(jq -r '.account' "$TEST_TOKEN_FILE")
            EXPIRES_AT=$(jq -r '.expires_at' "$TEST_TOKEN_FILE")
            EXPIRES_DATE=$(date -d "@$EXPIRES_AT" 2>/dev/null || date -r "$EXPIRES_AT" 2>/dev/null || echo "unknown")
            
            echo ""
            print_info "Authentication Details:"
            print_info "  Account: $ACCOUNT"
            print_info "  Expires: $EXPIRES_DATE"
            
            break
        fi
    fi
done

# Clean up temp cache
rm -rf "$TEMP_CACHE"

if [ "$TOKEN_FOUND" = false ]; then
    print_error "Failed to extract authentication tokens"
    print_info "Please try running OneMount manually:"
    print_info "  ./build/onemount --cache-dir=/tmp/cache /tmp/mount"
    exit 1
fi

echo ""
print_success "Authentication setup complete!"
echo ""
print_info "Your tokens are saved at: $TEST_TOKEN_FILE"
print_info "They will be used automatically by all test scripts"
echo ""
print_info "Next steps:"
print_info "  1. Run tests: ./scripts/test-task-5.4-filesystem-operations.sh"
print_info "  2. For Docker: See docs/testing/persistent-authentication-setup.md"
print_info "  3. To refresh: $0 --refresh"
echo ""
