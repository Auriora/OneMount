#!/bin/bash
# Setup script for OneMount test authentication tokens
# This script helps create and validate test auth tokens safely

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_security_warning() {
    echo -e "${RED}🔒 SECURITY WARNING${NC}"
    echo -e "${RED}===================${NC}"
    echo -e "${RED}NEVER use production OneDrive accounts for testing!${NC}"
    echo -e "${RED}Create a dedicated test account with test data only.${NC}"
    echo -e "${RED}System tests will create/delete files in OneDrive.${NC}"
    echo
}

print_info "OneMount Test Authentication Setup"
print_info "=================================="
echo

print_security_warning

print_info "This script will help you set up authentication tokens for OneMount testing."
print_info "The process involves:"
print_info "1. Building OneMount if needed"
print_info "2. Running authentication with a temporary mountpoint"
print_info "3. Copying tokens to the correct test locations"
print_info "4. Validating the setup"
echo

# Check if we're in the right directory
if [[ ! -f "go.mod" ]] || ! grep -q "github.com/auriora/onemount" go.mod; then
    print_error "This script must be run from the OneMount project root"
    exit 1
fi

# Check if OneMount binary exists
if [[ ! -f "build/onemount" ]]; then
    print_info "OneMount binary not found. Building..."
    if ! make onemount; then
        print_error "Failed to build OneMount binary"
        exit 1
    fi
    print_success "OneMount binary built successfully"
fi

# Check for existing production tokens
PRODUCTION_TOKENS="$HOME/.cache/onemount/auth_tokens.json"
TEST_TOKENS="$HOME/.onemount-tests/.auth_tokens.json"
WORKSPACE_TOKENS="test-artifacts/.auth_tokens.json"

if [[ -f "$PRODUCTION_TOKENS" ]]; then
    print_warning "Production auth tokens found at: $PRODUCTION_TOKENS"
    echo
    echo "Please confirm this is a DEDICATED TEST ACCOUNT:"
    echo "1. Is this a separate OneDrive account used ONLY for testing? (y/N)"
    read -r response
    if [[ ! "$response" =~ ^[Yy]$ ]]; then
        print_error "Please create a dedicated test OneDrive account first!"
        echo
        print_info "Steps to create test account:"
        print_info "1. Go to https://outlook.com and create a new Microsoft account"
        print_info "2. Sign up for OneDrive with this new account"
        print_info "3. Add some test files (they may be modified/deleted during testing)"
        print_info "4. Run this script again and authenticate with the test account"
        exit 1
    fi
    
    print_info "Using existing tokens as test tokens..."
    
    # Create test directory
    mkdir -p "$HOME/.onemount-tests"
    
    # Copy to test location
    cp "$PRODUCTION_TOKENS" "$TEST_TOKENS"
    print_success "Copied tokens to test location: $TEST_TOKENS"
    
else
    print_info "No existing auth tokens found. Starting authentication..."
    echo
    print_warning "Make sure to authenticate with a DEDICATED TEST ACCOUNT!"
    echo
    read -p "Press Enter to continue with authentication..."

    # Create a temporary mountpoint for authentication
    TEMP_MOUNTPOINT="/tmp/onemount-auth-$$"
    mkdir -p "$TEMP_MOUNTPOINT"

    print_info "Authenticating with temporary mountpoint: $TEMP_MOUNTPOINT"

    # Authenticate (this will create tokens and exit)
    if ! ./build/onemount --auth-only "$TEMP_MOUNTPOINT"; then
        print_error "Authentication failed"
        rmdir "$TEMP_MOUNTPOINT" 2>/dev/null || true
        exit 1
    fi

    # Clean up temporary mountpoint
    rmdir "$TEMP_MOUNTPOINT" 2>/dev/null || true
    
    # Look for tokens in various locations
    FOUND_TOKENS=""

    # Check standard location
    if [[ -f "$PRODUCTION_TOKENS" ]]; then
        FOUND_TOKENS="$PRODUCTION_TOKENS"
        print_success "Found tokens at standard location: $PRODUCTION_TOKENS"
    else
        # Look for tokens in cache directories (OneMount creates subdirectories based on mountpoint)
        FOUND_TOKENS=$(find ~/.cache/onemount -name "auth_tokens.json" 2>/dev/null | sort -r | head -1)
        if [[ -n "$FOUND_TOKENS" ]]; then
            print_success "Found tokens at: $FOUND_TOKENS"
        else
            print_error "Authentication completed but no tokens file found"
            print_info "Checked locations:"
            print_info "  - $PRODUCTION_TOKENS"
            print_info "  - ~/.cache/onemount/*/auth_tokens.json"
            print_info ""
            print_info "Available cache directories:"
            ls -la ~/.cache/onemount/ 2>/dev/null || print_info "  (none found)"
            print_info ""
            print_info "Please check the authentication output above for any error messages."
            exit 1
        fi
    fi
    
    # Create test directory and copy tokens
    mkdir -p "$HOME/.onemount-tests"
    cp "$FOUND_TOKENS" "$TEST_TOKENS"
    print_success "Authentication completed and tokens saved to: $TEST_TOKENS"
fi

# Validate tokens
print_info "Validating auth tokens..."
if command -v jq >/dev/null 2>&1; then
    if jq empty "$TEST_TOKENS" 2>/dev/null; then
        print_success "Auth tokens are valid JSON"
        
        # Check expiration
        EXPIRES_AT=$(jq -r '.expires_at // 0' "$TEST_TOKENS" 2>/dev/null || echo "0")
        CURRENT_TIME=$(date +%s)
        
        if [[ "$EXPIRES_AT" != "0" ]] && [[ "$EXPIRES_AT" -le "$CURRENT_TIME" ]]; then
            print_warning "Auth tokens appear to be expired"
            print_info "You may need to re-authenticate before running tests"
        else
            print_success "Auth tokens are valid and not expired"
        fi
    else
        print_error "Invalid auth tokens format"
        exit 1
    fi
else
    print_warning "jq not available - skipping JSON validation"
fi

# Set up workspace tokens
print_info "Setting up workspace auth tokens..."

# Clean up any existing problematic files
if [[ -d "$WORKSPACE_TOKENS" ]]; then
    print_info "Removing incorrect directory structure..."
    rm -rf "$WORKSPACE_TOKENS"
fi

# Ensure test-artifacts directory exists with correct permissions
mkdir -p test-artifacts
chmod 755 test-artifacts

# Copy tokens to workspace
cp "$TEST_TOKENS" "$WORKSPACE_TOKENS"
chmod 600 "$WORKSPACE_TOKENS"
print_success "Workspace auth tokens created: $WORKSPACE_TOKENS"

# Final validation
print_info "Running final validation..."
ls -la "$WORKSPACE_TOKENS"

echo
print_success "✅ Test authentication setup complete!"
echo
print_info "You can now run system tests with:"
print_info "  ./scripts/dev.py test docker system"
print_info "  make docker-test-system"
echo
print_warning "Remember: These tokens are for TESTING ONLY!"
print_warning "System tests will create/delete files in your OneDrive test account."
