#!/bin/bash

# OneMount GitHub Secrets Setup Script
# This script helps you set up the required GitHub secrets for system tests

set -e

echo "🔧 OneMount GitHub Secrets Setup"
echo "================================"
echo ""

# Check if .env file exists
if [ ! -f ".env" ]; then
    echo "❌ .env file not found in the current directory"
    echo "Please ensure you're running this from the OneMount project root"
    echo "and that you have a .env file with AUTH_TOKENS_B64"
    exit 1
fi

# Extract AUTH_TOKENS_B64 from .env file
if ! grep -q "AUTH_TOKENS_B64=" .env; then
    echo "❌ AUTH_TOKENS_B64 not found in .env file"
    echo "Please ensure your .env file contains the AUTH_TOKENS_B64 variable"
    exit 1
fi

AUTH_TOKENS_B64=$(grep "AUTH_TOKENS_B64=" .env | cut -d'=' -f2)

if [ -z "$AUTH_TOKENS_B64" ]; then
    echo "❌ AUTH_TOKENS_B64 is empty in .env file"
    exit 1
fi

echo "✅ Found AUTH_TOKENS_B64 in .env file"
echo ""

# Validate the base64 content
echo "🔍 Validating auth tokens..."
if echo "$AUTH_TOKENS_B64" | base64 -d | jq empty 2>/dev/null; then
    echo "✅ Auth tokens are valid JSON"
    
    # Check expiration
    EXPIRES_AT=$(echo "$AUTH_TOKENS_B64" | base64 -d | jq -r '.expires_at // 0')
    CURRENT_TIME=$(date +%s)
    
    if [ "$EXPIRES_AT" -le "$CURRENT_TIME" ]; then
        echo "⚠️  Warning: Auth tokens appear to be expired"
        echo "You may want to refresh them before setting up the GitHub secret"
    else
        echo "✅ Auth tokens are valid (expires in $((EXPIRES_AT - CURRENT_TIME)) seconds)"
    fi
else
    echo "❌ Auth tokens are not valid JSON"
    exit 1
fi

echo ""
echo "📋 GitHub Secret Setup Instructions"
echo "==================================="
echo ""
echo "1. Go to your GitHub repository: https://github.com/Auriora/OneMount"
echo "2. Click on 'Settings' tab"
echo "3. In the left sidebar, click 'Secrets and variables' > 'Actions'"
echo "4. Click 'New repository secret'"
echo "5. Set the name to: ONEDRIVE_PERSONAL_TOKENS"
echo "6. Copy and paste the following value:"
echo ""
echo "--- START COPYING FROM HERE ---"
echo "$AUTH_TOKENS_B64"
echo "--- END COPYING HERE ---"
echo ""
echo "7. Click 'Add secret'"
echo ""
echo "✅ Once you've added the secret, the system tests should work in GitHub Actions!"
echo ""
echo "Note: For self-hosted runners, the .env file will be used automatically."
