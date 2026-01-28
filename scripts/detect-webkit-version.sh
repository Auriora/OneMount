#!/usr/bin/env bash
# Detect which version of webkit2gtk is available and output the appropriate version
# Supports both webkit2gtk-4.0 and webkit2gtk-4.1

set -euo pipefail

if ! command -v pkg-config >/dev/null 2>&1; then
    echo "webkit2gtk-4.0"  # Default fallback
    exit 0
fi

# Check for webkit2gtk-4.1 first (newer version)
if pkg-config --exists webkit2gtk-4.1 2>/dev/null; then
    echo "webkit2gtk-4.1"
    exit 0
fi

# Fall back to webkit2gtk-4.0
if pkg-config --exists webkit2gtk-4.0 2>/dev/null; then
    echo "webkit2gtk-4.0"
    exit 0
fi

# If neither is found, default to 4.0 (will fail at compile time if not available)
echo "webkit2gtk-4.0"
exit 0
