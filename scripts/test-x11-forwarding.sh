#!/bin/bash

# Test X11 forwarding in Docker environment
# This script verifies that X11 forwarding is properly configured

echo "🖥️  Testing X11 Forwarding for Docker"
echo "====================================="
echo ""

# Check if DISPLAY is set
if [[ -z "$DISPLAY" ]]; then
    echo "❌ DISPLAY environment variable is not set"
    echo ""
    echo "To fix this, run:"
    echo "  export DISPLAY=:0"
    exit 1
fi

echo "✅ DISPLAY is set to: $DISPLAY"
echo ""

# Check if X11 is accessible from host
echo "🔍 Testing X11 on host..."
if command -v xset >/dev/null 2>&1; then
    if xset q >/dev/null 2>&1; then
        echo "✅ X11 is accessible on host"
    else
        echo "❌ X11 is not accessible on host"
        echo "   Make sure X server is running"
        exit 1
    fi
else
    echo "⚠️  xset command not found on host"
    echo "   Installing: sudo apt-get install x11-xserver-utils"
fi

echo ""

# Check xhost permissions
echo "🔍 Checking xhost permissions..."
if command -v xhost >/dev/null 2>&1; then
    echo "Current xhost access control:"
    xhost | head -5
    echo ""
    echo "To allow Docker containers to access X11, run:"
    echo "  xhost +local:docker"
else
    echo "⚠️  xhost command not found"
    echo "   Installing: sudo apt-get install x11-xserver-utils"
fi

echo ""

# Test X11 in Docker
echo "🐳 Testing X11 in Docker container..."
docker compose -f docker/compose/docker-compose.test.yml run --rm test-runner \
    bash -c '
        echo "Container environment:"
        echo "  DISPLAY=$DISPLAY"
        echo "  XAUTHORITY=$XAUTHORITY"
        echo ""
        
        # Check if X11 utilities are available
        if command -v xset >/dev/null 2>&1; then
            echo "Testing X11 connectivity with xset..."
            if xset q >/dev/null 2>&1; then
                echo "✅ X11 is accessible from Docker container!"
                echo ""
                echo "X11 server info:"
                xset q | head -10
            else
                echo "❌ X11 is not accessible from Docker container"
                echo "   Error: $(xset q 2>&1)"
            fi
        else
            echo "⚠️  xset not available in container"
            echo "   This is expected - X11 forwarding may still work"
        fi
        
        echo ""
        echo "Checking X11 socket..."
        if [[ -S /tmp/.X11-unix/X0 ]]; then
            echo "✅ X11 socket found: /tmp/.X11-unix/X0"
            ls -la /tmp/.X11-unix/
        else
            echo "❌ X11 socket not found"
            echo "   Expected: /tmp/.X11-unix/X0"
            ls -la /tmp/.X11-unix/ 2>/dev/null || echo "   Directory does not exist"
        fi
    '

echo ""
echo "📋 Summary"
echo "=========="
echo ""
echo "If X11 is working, you should be able to:"
echo "  1. Run interactive authentication: ./scripts/interactive-auth.sh"
echo "  2. Open GUI applications from Docker containers"
echo ""
echo "If X11 is not working:"
echo "  1. Ensure X server is running on your host"
echo "  2. Run: xhost +local:docker"
echo "  3. Check DISPLAY variable: echo \$DISPLAY"
echo "  4. Try: docker run --rm -e DISPLAY=\$DISPLAY -v /tmp/.X11-unix:/tmp/.X11-unix:rw alpine sh -c 'apk add xeyes && xeyes'"