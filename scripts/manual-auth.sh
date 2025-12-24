#!/bin/bash

# Manual authentication script for OneMount
# This script uses the --no-browser option to get a URL for manual authentication

set -e

echo "🔐 OneMount Manual Authentication"
echo "================================="
echo ""

AUTH_DIR="$HOME/.onemount-tests"
AUTH_FILE="$AUTH_DIR/.auth_tokens.json"

mkdir -p "$AUTH_DIR"

echo "This script will provide you with a URL to visit in your host browser"
echo "for Microsoft OAuth authentication."
echo ""

# Check if auth tokens already exist
if [[ -f "$AUTH_FILE" ]]; then
    echo "📄 Existing auth tokens found:"
    echo "   File: $AUTH_FILE"
    echo "   Size: $(stat -c%s "$AUTH_FILE" 2>/dev/null || echo "unknown") bytes"
    echo "   Modified: $(stat -c%y "$AUTH_FILE" 2>/dev/null || echo "unknown")"
    echo ""
    
    read -p "🤔 Do you want to re-authenticate (overwrite existing tokens)? [y/N]: " -n 1 -r
    echo ""
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "✋ Authentication cancelled"
        exit 0
    fi
    echo ""
fi

echo "🚀 Starting manual authentication..."
echo ""
echo "📋 Instructions:"
echo "   1. Copy the URL that appears below"
echo "   2. Open it in your browser (on your host machine)"
echo "   3. Complete the Microsoft OAuth flow"
echo "   4. When redirected to a blank page, copy the full URL"
echo "   5. Paste the redirect URL back into this terminal"
echo ""

# Run the authentication command in Docker
docker compose -f docker/compose/docker-compose.test.yml run --rm \
    -e ONEMOUNT_AUTH_PATH="/tmp/home-tester/.onemount-tests-auth/.auth_tokens.json" \
    test-runner \
    bash -c '
        echo "🔧 Setting up authentication environment..."
        
        # Ensure auth directory exists in container
        mkdir -p /tmp/home-tester/.onemount-tests-auth
        mkdir -p /tmp/mount
        
        echo "🌐 Starting manual OAuth authentication flow..."
        echo ""
        echo "=========================================="
        echo "COPY THE URL BELOW AND OPEN IN YOUR BROWSER:"
        echo "=========================================="
        echo ""
        
        # Run the authentication command with no-browser option
        cd /workspace
        go run cmd/onemount/main.go --auth-only --no-browser /tmp/mount
        
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
if [[ -f "$AUTH_FILE" ]]; then
    echo ""
    echo "🎉 Authentication successful!"
    echo "📄 Auth tokens saved to: $AUTH_FILE"
    echo "   Size: $(stat -c%s "$AUTH_FILE") bytes"
    echo "   Modified: $(stat -c%y "$AUTH_FILE")"
    echo ""
    
    # Test the tokens
    echo "🧪 Testing authentication..."
    ./scripts/test-with-progress-monitor.sh "TestDeadlockRootCauseAnalysis" 30
    
    echo ""
    echo "✅ You can now run tests that require authentication:"
    echo "   ./scripts/test-with-progress-monitor.sh \"TestIT_FS_ETag_01_CacheValidationSafe\" 60"
    echo "   docker compose -f docker/compose/docker-compose.test.yml run --rm system-tests"
else
    echo ""
    echo "❌ Authentication failed!"
    echo "   Auth tokens file not found at: $AUTH_FILE"
    echo ""
    echo "🔍 Troubleshooting:"
    echo "   1. Make sure you completed the OAuth flow in your browser"
    echo "   2. Ensure you copied the full redirect URL correctly"
    echo "   3. Check that the redirect URL starts with 'https://login.live.com/oauth20_desktop.srf'"
    echo "   4. Try running the script again"
    exit 1
fi