# Personal OneDrive CI Setup Guide

This guide shows how to use your personal OneDrive account for CI system tests using GitHub's secure secret storage.

## Quick Setup (Using Your Existing Authentication)

### Step 1: Extract Your Current Auth Tokens

```bash
# Check if you have existing OneMount authentication
ls -la ~/.cache/onemount/auth_tokens.json

# If the file exists, encode it for GitHub Secrets
base64 -w 0 ~/.cache/onemount/auth_tokens.json > /tmp/encoded_tokens.txt

# Display the encoded tokens (copy this for GitHub)
cat /tmp/encoded_tokens.txt

# Clean up the temporary file
rm /tmp/encoded_tokens.txt
```

### Step 2: Add GitHub Secret

1. Go to your GitHub repository
2. Click **Settings** → **Secrets and variables** → **Actions**
3. Click **New repository secret**
4. Name: `ONEDRIVE_PERSONAL_TOKENS`
5. Value: Paste the base64-encoded string from Step 1
6. Click **Add secret**

### Step 3: Create CI Workflow

Create `.github/workflows/system-tests-personal.yml`:

```yaml
name: System Tests (Personal OneDrive)

on:
  # Run on pull requests
  pull_request:
    branches: [ main ]
    paths:
      - 'internal/**'
      - 'pkg/**'
      - 'tests/system/**'
      - 'go.mod'
      - 'go.sum'
  
  # Run on pushes to main
  push:
    branches: [ main ]
    paths:
      - 'internal/**'
      - 'pkg/**'
      - 'tests/system/**'
      - 'go.mod'
      - 'go.sum'
  
  # Allow manual triggering
  workflow_dispatch:
    inputs:
      test_category:
        description: 'Test category to run'
        required: false
        default: 'comprehensive'
        type: choice
        options:
          - comprehensive
          - performance
          - reliability
          - integration
          - stress

jobs:
  system-tests:
    runs-on: ubuntu-latest
    
    # Only run if we have the personal OneDrive tokens
    if: ${{ secrets.ONEDRIVE_PERSONAL_TOKENS != '' }}
    
    steps:
    - name: Checkout code
      uses: actions/checkout@v4
    
    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: '1.23'
        cache: true
    
    - name: Install FUSE
      run: |
        sudo apt-get update
        sudo apt-get install -y fuse3 libfuse3-dev
    
    - name: Set up personal OneDrive credentials
      env:
        ONEDRIVE_PERSONAL_TOKENS: ${{ secrets.ONEDRIVE_PERSONAL_TOKENS }}
      run: |
        # Create test directories
        mkdir -p ~/.onemount-tests/logs
        chmod 700 ~/.onemount-tests
        
        # Decode and save the auth tokens
        echo "$ONEDRIVE_PERSONAL_TOKENS" | base64 -d > ~/.onemount-tests/.auth_tokens.json
        chmod 600 ~/.onemount-tests/.auth_tokens.json
        
        # Verify the tokens file is valid JSON
        if ! jq empty ~/.onemount-tests/.auth_tokens.json; then
          echo "❌ Invalid auth tokens format"
          exit 1
        fi
        
        # Check token expiration
        EXPIRES_AT=$(jq -r '.expires_at // 0' ~/.onemount-tests/.auth_tokens.json)
        CURRENT_TIME=$(date +%s)
        
        if [ "$EXPIRES_AT" -le "$CURRENT_TIME" ]; then
          echo "⚠️  Auth tokens appear to be expired"
          echo "You may need to refresh your local authentication and update the GitHub secret"
        else
          echo "✅ Auth tokens are valid (expires in $((EXPIRES_AT - CURRENT_TIME)) seconds)"
        fi
        
        echo "✅ Personal OneDrive credentials configured"
    
    - name: Verify OneDrive access
      run: |
        # Test that we can access your OneDrive
        ACCESS_TOKEN=$(jq -r '.access_token' ~/.onemount-tests/.auth_tokens.json)
        
        RESPONSE=$(curl -s -H "Authorization: Bearer $ACCESS_TOKEN" \
          "https://graph.microsoft.com/v1.0/me/drive/root")
        
        if echo "$RESPONSE" | jq -e '.id' > /dev/null; then
          echo "✅ OneDrive access verified"
          DRIVE_NAME=$(echo "$RESPONSE" | jq -r '.name // "Unknown"')
          echo "Drive Name: $DRIVE_NAME"
        else
          echo "❌ Failed to access OneDrive"
          echo "Response: $RESPONSE"
          echo "You may need to refresh your authentication tokens"
          exit 1
        fi
    
    - name: Build OneMount
      run: make build
    
    - name: Run system tests
      env:
        TEST_CATEGORY: ${{ github.event.inputs.test_category || 'comprehensive' }}
      run: |
        # Make script executable
        chmod +x scripts/run-system-tests.sh
        
        # Run tests with appropriate timeout
        case "$TEST_CATEGORY" in
          "comprehensive")
            timeout 15m ./scripts/run-system-tests.sh --comprehensive --verbose
            ;;
          "performance")
            timeout 20m ./scripts/run-system-tests.sh --performance --verbose
            ;;
          "reliability")
            timeout 10m ./scripts/run-system-tests.sh --reliability --verbose
            ;;
          "integration")
            timeout 10m ./scripts/run-system-tests.sh --integration --verbose
            ;;
          "stress")
            timeout 25m ./scripts/run-system-tests.sh --stress --verbose
            ;;
          *)
            echo "Unknown test category: $TEST_CATEGORY"
            exit 1
            ;;
        esac
    
    - name: Upload test logs
      if: always()
      uses: actions/upload-artifact@v3
      with:
        name: system-test-logs-personal
        path: |
          ~/.onemount-tests/logs/
        retention-days: 7
    
    - name: Cleanup test data
      if: always()
      run: |
        # Clean up test files from your OneDrive
        ACCESS_TOKEN=$(jq -r '.access_token' ~/.onemount-tests/.auth_tokens.json 2>/dev/null || echo "")
        
        if [ -n "$ACCESS_TOKEN" ] && [ "$ACCESS_TOKEN" != "null" ]; then
          echo "🧹 Cleaning up test data from OneDrive..."
          
          # Delete the test directory
          curl -s -X DELETE \
            -H "Authorization: Bearer $ACCESS_TOKEN" \
            "https://graph.microsoft.com/v1.0/me/drive/root:/onemount_system_tests" || true
          
          echo "✅ Test data cleanup completed"
        fi
        
        # Clean up local test files
        rm -rf ~/.onemount-tests/ || true
        
        # Unmount any remaining FUSE mounts
        fusermount3 -uz ~/.onemount-tests/tmp/system-test-mount 2>/dev/null || true

  # Job to provide helpful information if secrets are missing
  check-setup:
    runs-on: ubuntu-latest
    if: ${{ secrets.ONEDRIVE_PERSONAL_TOKENS == '' }}
    steps:
    - name: Setup instructions
      run: |
        echo "🔧 Personal OneDrive system tests are not configured"
        echo ""
        echo "To enable system tests with your personal OneDrive:"
        echo "1. Run locally: base64 -w 0 ~/.cache/onemount/auth_tokens.json"
        echo "2. Go to Settings → Secrets and variables → Actions"
        echo "3. Add secret: ONEDRIVE_PERSONAL_TOKENS with the base64 output"
        echo ""
        echo "See docs/testing/personal-onedrive-ci-setup.md for detailed instructions"
```

## Step 4: Test the Setup

1. **Commit the workflow file**:
```bash
git add .github/workflows/system-tests-personal.yml
git commit -m "Add personal OneDrive CI system tests"
git push
```

2. **Manually trigger a test**:
   - Go to **Actions** tab in GitHub
   - Select **System Tests (Personal OneDrive)**
   - Click **Run workflow**
   - Choose **comprehensive** test category
   - Click **Run workflow**

## Security Considerations

### ✅ **Safe Practices**
- Your tokens are encrypted in GitHub Secrets
- Tests run in isolated CI environment
- Test data is automatically cleaned up
- Only repository collaborators can access secrets

### ⚠️ **Important Notes**
- Tests create files in `/onemount_system_tests/` folder in your OneDrive
- All test data is automatically deleted after tests complete
- Your personal files are never accessed or modified
- Tests only use ~100MB of storage temporarily

### 🔄 **Token Refresh**
Your OneDrive tokens will eventually expire. When they do:

1. **Refresh locally**:
```bash
./build/onemount --auth-only
```

2. **Update GitHub secret**:
```bash
base64 -w 0 ~/.cache/onemount/auth_tokens.json
# Copy output and update ONEDRIVE_PERSONAL_TOKENS secret
```

## Alternative: Create a Dedicated Test Folder

If you prefer to isolate test data, you can modify the test path:

1. **Update test constants**:
```go
// In pkg/testutil/test_constants.go
OneDriveTestPath = "/OneMount_CI_Tests"  // Instead of "/onemount_system_tests"
```

2. **Create the folder manually** in your OneDrive to ensure it exists

## Monitoring and Troubleshooting

### View Test Results
- Go to **Actions** tab to see test runs
- Download log artifacts for detailed debugging
- Check the **Cleanup test data** step to confirm cleanup

### Common Issues

**Token Expiration**:
```
❌ Failed to access OneDrive
```
**Solution**: Refresh tokens locally and update GitHub secret

**Permission Denied**:
```
❌ Invalid auth tokens format
```
**Solution**: Ensure base64 encoding was done correctly

**FUSE Mount Issues**:
```
❌ Failed to create FUSE server
```
**Solution**: This is handled automatically in the workflow

## Benefits of This Approach

### ✅ **Advantages**
- Uses your existing OneDrive account
- No additional Azure setup required
- Secure token storage in GitHub
- Automatic test data cleanup
- Easy to set up and maintain

### ✅ **Perfect For**
- Personal projects
- Small teams
- Quick validation
- Development testing

This setup gives you comprehensive CI testing with your personal OneDrive account while maintaining security and automatically cleaning up test data!
