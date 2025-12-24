#!/bin/bash

# Cleanup script to remove old authentication scripts that use copying/symlinking
# These are replaced by the reference-based authentication system

set -e

echo "🧹 OneMount Authentication Script Cleanup"
echo "========================================="
echo ""

echo "This script removes old authentication scripts that used copying/symlinking approaches."
echo "They have been replaced by the reference-based authentication system."
echo ""

# List of scripts to remove (copying/symlinking approaches)
OLD_SCRIPTS=(
    "scripts/copy-auth-from-devcontainer.sh"
    "scripts/fix-auth-tokens.sh"
    "scripts/setup-test-auth.sh"
    "scripts/clean-expired-tokens.sh"
    "scripts/refresh-auth-tokens.sh"
    "scripts/setup-auth-environment.sh"
    "scripts/auth-workaround.sh"
    "scripts/test-auth-tokens.sh"
)

# List of scripts to keep (reference-based or still useful)
KEEP_SCRIPTS=(
    "scripts/setup-auth-reference.sh"
    "scripts/timeout-test-wrapper.sh"
    "scripts/manual-auth.sh"
    "scripts/interactive-auth.sh"
)

echo "📋 Scripts to remove (copying/symlinking approaches):"
for script in "${OLD_SCRIPTS[@]}"; do
    if [[ -f "$script" ]]; then
        echo "  ❌ $script"
    else
        echo "  ⚪ $script (already removed)"
    fi
done

echo ""
echo "📋 Scripts to keep (reference-based or still useful):"
for script in "${KEEP_SCRIPTS[@]}"; do
    if [[ -f "$script" ]]; then
        echo "  ✅ $script"
    else
        echo "  ⚠️  $script (missing!)"
    fi
done

echo ""
read -p "Continue with cleanup? [Y/n]: " -n 1 -r
echo ""
if [[ $REPLY =~ ^[Nn]$ ]]; then
    echo "Cleanup cancelled"
    exit 0
fi

echo "🗑️  Removing old authentication scripts..."

removed_count=0
for script in "${OLD_SCRIPTS[@]}"; do
    if [[ -f "$script" ]]; then
        echo "   Removing: $script"
        rm "$script"
        removed_count=$((removed_count + 1))
    fi
done

echo ""
echo "✅ Cleanup complete!"
echo "   Removed $removed_count old scripts"
echo ""

# Check for any remaining scripts that might copy tokens
echo "🔍 Checking for any remaining scripts that copy tokens..."

remaining_copy_scripts=$(grep -r "cp.*token\|copy.*token" scripts/ 2>/dev/null | grep -v "cleanup-old-auth-scripts.sh" || true)
if [[ -n "$remaining_copy_scripts" ]]; then
    echo "⚠️  Found remaining scripts that copy tokens:"
    echo "$remaining_copy_scripts"
    echo ""
    echo "These may need manual review to convert to reference-based approach."
else
    echo "✅ No remaining scripts found that copy tokens"
fi

echo ""
echo "📋 Current authentication system:"
echo "   ✅ Reference-based: scripts/setup-auth-reference.sh"
echo "   ✅ Docker override: docker/compose/docker-compose.auth.yml"
echo "   ✅ Environment config: .env.auth"
echo "   ✅ Timeout wrapper: scripts/timeout-test-wrapper.sh"
echo ""
echo "💡 Usage:"
echo "   # Setup authentication reference:"
echo "   ./scripts/setup-auth-reference.sh"
echo ""
echo "   # Run tests with timeout protection:"
echo "   ./scripts/timeout-test-wrapper.sh \"TestPattern\" 60"
echo ""
echo "   # Or use Docker Compose directly:"
echo "   docker compose -f docker/compose/docker-compose.test.yml \\"
echo "     -f docker/compose/docker-compose.auth.yml run --rm test-runner"
echo ""
echo "🎉 Authentication system cleanup complete!"