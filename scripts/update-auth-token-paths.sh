#!/bin/bash

# Script to update all test files to use centralized auth token path helper
# This removes fallback logic and enforces single source of truth

set -e

echo "🔧 Updating Authentication Token Path References"
echo "================================================"
echo ""

# List of test files to update
TEST_FILES=(
    "internal/fs/deadlock_root_cause_test.go"
    "internal/fs/etag_deadlock_fix_test.go"
    "internal/fs/etag_validation_fixed_test.go"
    "internal/fs/etag_validation_safe_test.go"
    "internal/fs/etag_validation_timeout_fixed_test.go"
    "internal/fs/mount_integration_real_test.go"
    "internal/fs/etag_diagnostic_with_progress_test.go"
    "internal/fs/minimal_hang_test.go"
    "internal/fs/etag_validation_integration_test.go"
)

# Pattern to find and replace
OLD_PATTERN='authPath := os.Getenv("ONEMOUNT_AUTH_PATH")
	if authPath == "" {
		authPath = "test-artifacts/.auth_tokens.json"
	}'

NEW_PATTERN='authPath, err := testutil.GetAuthTokenPath()
	if err != nil {
		t.Fatalf("Authentication not configured: %v", err)
	}'

echo "📝 Files to update:"
for file in "${TEST_FILES[@]}"; do
    if [[ -f "$file" ]]; then
        echo "   ✓ $file"
    else
        echo "   ✗ $file (not found)"
    fi
done
echo ""

read -p "Continue with updates? (y/N) " -n 1 -r
echo ""
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "❌ Aborted"
    exit 1
fi

echo ""
echo "🔄 Updating files..."
echo ""

UPDATED=0
FAILED=0

for file in "${TEST_FILES[@]}"; do
    if [[ ! -f "$file" ]]; then
        echo "⚠️  Skipping $file (not found)"
        continue
    fi
    
    echo "Processing: $file"
    
    # Check if file contains the old pattern
    if grep -q 'authPath := os.Getenv("ONEMOUNT_AUTH_PATH")' "$file"; then
        # Create backup
        cp "$file" "$file.bak"
        
        # Use sed to replace the pattern
        # This is a multi-line replacement, so we need to be careful
        if sed -i '/authPath := os.Getenv("ONEMOUNT_AUTH_PATH")/,/authPath = "test-artifacts\/.auth_tokens.json"/c\
	authPath, err := testutil.GetAuthTokenPath()\
	if err != nil {\
		t.Fatalf("Authentication not configured: %v", err)\
	}' "$file"; then
            echo "   ✓ Updated successfully"
            UPDATED=$((UPDATED + 1))
            rm "$file.bak"
        else
            echo "   ✗ Update failed, restoring backup"
            mv "$file.bak" "$file"
            FAILED=$((FAILED + 1))
        fi
    else
        echo "   ⊘ No changes needed (pattern not found)"
    fi
    echo ""
done

echo "================================================"
echo "📊 Summary:"
echo "   Updated: $UPDATED files"
echo "   Failed: $FAILED files"
echo "   Total: ${#TEST_FILES[@]} files"
echo ""

if [[ $FAILED -gt 0 ]]; then
    echo "⚠️  Some files failed to update. Please review manually."
    exit 1
fi

echo "✅ All files updated successfully!"
echo ""
echo "📋 Next steps:"
echo "   1. Review the changes: git diff"
echo "   2. Run tests to verify: docker compose -f docker/compose/docker-compose.test.yml -f docker/compose/docker-compose.auth.yml run --rm test-runner"
echo "   3. Commit the changes: git add . && git commit -m 'Remove auth token fallback locations'"
