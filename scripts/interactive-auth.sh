#!/bin/bash

# Interactive authentication script for Docker test environment
# This script runs the authentication process with X11 forwarding enabled

set -e

echo "🔐 OneMount Interactive Authentication"
echo "======================================"
echo ""

# Check if DISPLAY is set
if [[ -z "$DISPLAY" ]]; then
    echo "❌ DISPLAY environment variable is not set"
    echo "Please ensure X11 forwarding is enabled:"
    echo "  export DISPLAY=:0"
    echo "  xhost +local:docker  # Allow Docker containers to access X11"
    exit 1
fi

echo "🖥️  Display: $DISPLAY"
echo "🏠 Home: $HOME"
echo ""

# Check if X11 is accessible
if ! xset q >/dev/null 2>&1; then
    echo "⚠️  Warning: Cannot connect to X11 display"
    echo "You may need to run: xhost +local:docker"
    echo ""
fi

# Ensure auth directory exists
AUTH_DIR="$HOME/.onemount-tests"
mkdir -p "$AUTH_DIR"

echo "📁 Auth directory: $AUTH_DIR"
echo "🎯 Target auth file: $AUTH_DIR/.auth_tokens.json"
echo ""

# Check if auth tokens already exist
if [[ -f "$AUTH_DIR/.auth_tokens.json" ]]; then
    echo "📄 Existing auth tokens found:"
    echo "   File: $AUTH_DIR/.auth_tokens.json"
    echo "   Size: $(stat -c%s "$AUTH_DIR/.auth_tokens.json" 2>/dev/null || echo "unknown") bytes"
    echo "   Modified: $(stat -c%y "$AUTH_DIR/.auth_tokens.json" 2>/dev/null || echo "unknown")"
    echo ""
    
    read -p "🤔 Do you want to re-authenticate (overwrite existing tokens)? [y/N]: " -n 1 -r
    echo ""
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "✋ Authentication cancelled"
        exit 0
    fi
    echo ""
fi

echo "🚀 Starting interactive authentication in Docker..."
echo "   This will open a browser window for OAuth authentication"
echo ""

# Run the authentication command in Docker with X11 forwarding
docker compose -f docker/compose/docker-compose.test.yml run --rm \
    -e DISPLAY="$DISPLAY" \
    -e ONEMOUNT_AUTH_PATH="/tmp/home-tester/.onemount-tests-auth/.auth_tokens.json" \
    test-runner \
    bash -c '
        echo "🔧 Setting up authentication environment..."
        
        # Ensure auth directory exists in container
        mkdir -p /tmp/home-tester/.onemount-tests-auth
        
        # Check X11 connectivity
        if command -v xset >/dev/null 2>&1; then
            if xset q >/dev/null 2>&1; then
                echo "✅ X11 display accessible"
            else
                echo "⚠️  X11 display not accessible, authentication may fail"
            fi
        else
            echo "⚠️  xset not available, cannot test X11 connectivity"
        fi
        
        echo "🌐 Starting OAuth authentication flow..."
        echo "   A browser window should open shortly"
        echo "   Please complete the authentication in the browser"
        echo "   You have 5 minutes to complete the login process"
        echo ""
        
        # Run the authentication command with extended timeout
        cd /workspace
        mkdir -p /tmp/test-mount
        timeout 300 go run cmd/onemount/main.go -a /tmp/test-mount
        
        echo ""
        echo "✅ Authentication completed!"
        
        # Show the created auth file
        if [[ -f "/tmp/home-tester/.onemount-tests-auth/.auth_tokens.json" ]]; then
            echo "📄 Auth tokens saved to: /tmp/home-tester/.onemount-tests-auth/.auth_tokens.json"
            echo "   Size: $(stat -c%s "/tmp/home-tester/.onemount-tests-auth/.auth_tokens.json") bytes"
        else
            echo "❌ Auth tokens file not found - authentication may have failed"
            exit 1
        fi
    '

# Check if authentication was successful
if [[ -f "$AUTH_DIR/.auth_tokens.json" ]]; then
    echo ""
    echo "🎉 Authentication successful!"
    echo "📄 Auth tokens saved to: $AUTH_DIR/.auth_tokens.json"
    echo "   Size: $(stat -c%s "$AUTH_DIR/.auth_tokens.json") bytes"
    echo "   Modified: $(stat -c%y "$AUTH_DIR/.auth_tokens.json")"
    echo ""
    echo "✅ You can now run tests that require authentication:"
    echo "   ./scripts/test-with-progress-monitor.sh \"TestIT_FS_ETag_01_CacheValidationSafe\" 60"
    echo "   docker compose -f docker/compose/docker-compose.test.yml run --rm system-tests"
else
    echo ""
    echo "❌ Authentication failed!"
    echo "   Auth tokens file not found at: $AUTH_DIR/.auth_tokens.json"
    echo ""
    echo "🔍 Troubleshooting:"
    echo "   1. Ensure X11 forwarding is working: xhost +local:docker"
    echo "   2. Check DISPLAY variable: echo \$DISPLAY"
    echo "   3. Try running: xeyes (should open a window)"
    echo "   4. Check Docker logs for error messages"
    exit 1
fi