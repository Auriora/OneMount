# CI System Tests Setup Guide

This guide explains how to set up system tests in GitHub Actions CI/CD pipelines using real OneDrive accounts.

## Overview

There are three approaches for running system tests in CI:

1. **Service Principal (Recommended)** - Most secure, uses Azure App Registration
2. **Encrypted Test Account** - Simple, uses dedicated test account credentials
3. **Self-Hosted Runner** - For organizations with persistent infrastructure

## Option 1: Service Principal Setup (Recommended)

### Prerequisites
- Azure subscription with admin access
- Azure CLI installed locally

### Step 1: Create Azure App Registration

```bash
# Login to Azure
az login

# Create app registration
az ad app create \
  --display-name "OneMount-CI-Tests" \
  --required-resource-accesses '[{
    "resourceAppId": "00000003-0000-0000-c000-000000000000",
    "resourceAccess": [{
      "id": "75359482-378d-4052-8f01-80520e7db3cd",
      "type": "Scope"
    }]
  }]'

# Get the application ID
APP_ID=$(az ad app list --display-name "OneMount-CI-Tests" --query "[0].appId" -o tsv)
echo "Application ID: $APP_ID"

# Create client secret
SECRET_RESULT=$(az ad app credential reset --id $APP_ID --append)
CLIENT_SECRET=$(echo $SECRET_RESULT | jq -r '.password')
echo "Client Secret: $CLIENT_SECRET"

# Get tenant ID
TENANT_ID=$(az account show --query tenantId -o tsv)
echo "Tenant ID: $TENANT_ID"
```

### Step 2: Grant Permissions

1. Go to [Azure Portal](https://portal.azure.com)
2. Navigate to **App Registrations** → **OneMount-CI-Tests**
3. Go to **API Permissions**
4. Click **Grant admin consent for [Your Organization]**
5. Verify the status shows "Granted"

### Step 3: Test OneDrive Access

```bash
# Test the service principal can access OneDrive
curl -X POST "https://login.microsoftonline.com/$TENANT_ID/oauth2/v2.0/token" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=client_credentials" \
  -d "client_id=$APP_ID" \
  -d "client_secret=$CLIENT_SECRET" \
  -d "scope=https://graph.microsoft.com/.default"
```

### Step 4: Configure GitHub Secrets

In your GitHub repository:

1. Go to **Settings** → **Secrets and variables** → **Actions**
2. Add these repository secrets:
   - `AZURE_CLIENT_ID`: The Application ID from Step 1
   - `AZURE_CLIENT_SECRET`: The Client Secret from Step 1  
   - `AZURE_TENANT_ID`: The Tenant ID from Step 1

### Step 5: Enable Workflow

The service principal workflow (`.github/workflows/system-tests.yml`) will automatically run when:
- Pull requests are created/updated
- Code is pushed to main branch
- Manually triggered via GitHub Actions UI

## Option 2: Encrypted Test Account Setup

### Prerequisites
- Dedicated OneDrive test account (not production)
- OneMount installed locally

### Step 1: Generate Test Account Tokens

```bash
# Authenticate with test account
./build/onemount --auth-only

# Copy tokens to temporary location
cp ~/.cache/onemount/auth_tokens.json /tmp/test_tokens.json

# Encode tokens for GitHub Secrets
base64 -w 0 /tmp/test_tokens.json > /tmp/encoded_tokens.txt
cat /tmp/encoded_tokens.txt
```

### Step 2: Configure GitHub Secret

1. Go to **Settings** → **Secrets and variables** → **Actions**
2. Add repository secret:
   - `ONEDRIVE_TEST_TOKENS`: The base64-encoded content from Step 1

### Step 3: Enable Simple Workflow

Rename `.github/workflows/system-tests-simple.yml` to `.github/workflows/system-tests.yml` to enable this approach.

## Option 3: Self-Hosted Runner Setup

### Prerequisites
- Self-hosted GitHub Actions runner
- Persistent storage for credentials

### Step 1: Set Up Runner Credentials

```bash
# On the self-hosted runner machine
sudo mkdir -p /opt/onemount-ci
sudo chmod 755 /opt/onemount-ci

# Authenticate with test account (run as the runner user)
./build/onemount --auth-only

# Copy tokens to persistent location
sudo cp ~/.cache/onemount/auth_tokens.json /opt/onemount-ci/.auth_tokens.json
sudo chmod 600 /opt/onemount-ci/.auth_tokens.json
sudo chown runner:runner /opt/onemount-ci/.auth_tokens.json
```

### Step 2: Configure Runner Labels

Add the label `onemount-testing` to your self-hosted runner:

1. Go to **Settings** → **Actions** → **Runners**
2. Click on your runner
3. Add label: `onemount-testing`

### Step 3: Enable Self-Hosted Workflow

Rename `.github/workflows/system-tests-self-hosted.yml` to `.github/workflows/system-tests.yml`.

## Workflow Features

### Test Categories

All workflows support different test categories:

- **Comprehensive** (default): Basic functionality tests
- **Performance**: Upload/download speed tests  
- **Reliability**: Error handling tests
- **Integration**: Mount/unmount tests
- **Stress**: High load tests
- **All**: Run all categories

### Manual Triggering

You can manually trigger tests:

1. Go to **Actions** tab in GitHub
2. Select **System Tests** workflow
3. Click **Run workflow**
4. Choose test category
5. Click **Run workflow**

### Artifacts

Test logs are automatically uploaded as artifacts:

- **Retention**: 7-30 days depending on workflow
- **Contents**: Test logs, debug information, performance metrics
- **Access**: Download from the workflow run page

## Security Considerations

### Service Principal (Option 1)
✅ **Most Secure**
- No long-lived credentials in GitHub
- Automatic token rotation
- Minimal required permissions
- Audit trail in Azure AD

### Encrypted Test Account (Option 2)  
⚠️ **Moderate Security**
- Refresh tokens stored in GitHub Secrets
- Tokens can be rotated manually
- Requires dedicated test account
- Limited audit capabilities

### Self-Hosted Runner (Option 3)
⚠️ **Requires Careful Management**
- Credentials stored on runner machine
- Need secure runner environment
- Manual credential rotation
- Full control over environment

## Troubleshooting

### Authentication Failures

```bash
# Test service principal authentication
curl -X POST "https://login.microsoftonline.com/$TENANT_ID/oauth2/v2.0/token" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=client_credentials" \
  -d "client_id=$CLIENT_ID" \
  -d "client_secret=$CLIENT_SECRET" \
  -d "scope=https://graph.microsoft.com/.default"
```

### Permission Issues

Common permission errors:
- **Insufficient privileges**: Grant admin consent in Azure Portal
- **Wrong scope**: Ensure using `https://graph.microsoft.com/.default`
- **Expired secret**: Rotate client secret in Azure Portal

### FUSE Mount Issues

```bash
# Check FUSE availability in CI
ls -la /dev/fuse
lsmod | grep fuse

# Install FUSE if missing
sudo apt-get update
sudo apt-get install -y fuse3 libfuse3-dev
```

### Test Timeouts

Adjust timeouts in workflow files:
```yaml
# Increase timeout for slow networks
timeout 30m ./scripts/run-system-tests.sh --comprehensive
```

## Best Practices

### Security
1. **Use Service Principal** for production CI/CD
2. **Rotate credentials** regularly
3. **Monitor access logs** in Azure AD
4. **Use dedicated test accounts** only
5. **Restrict repository access** to authorized users

### Performance
1. **Cache Go modules** in workflows
2. **Run tests in parallel** when possible
3. **Use appropriate timeouts** for different test categories
4. **Clean up test data** after each run

### Reliability
1. **Handle network failures** gracefully
2. **Retry failed operations** with backoff
3. **Upload logs** for debugging
4. **Clean up resources** even on failure

## Monitoring

### Metrics to Track
- Test execution time
- Success/failure rates
- OneDrive API response times
- Resource usage (CPU, memory, disk)

### Alerts
Set up alerts for:
- Consecutive test failures
- Authentication failures
- Timeout issues
- Resource exhaustion

## Cost Considerations

### OneDrive Usage
- System tests use minimal storage (< 100MB)
- API calls are within free tier limits
- Consider using OneDrive for Business for higher limits

### GitHub Actions Minutes
- Service Principal: ~5-15 minutes per run
- Simple Account: ~5-15 minutes per run  
- Self-Hosted: No GitHub minutes consumed

Choose the approach that best fits your organization's security requirements, infrastructure, and maintenance capabilities.
