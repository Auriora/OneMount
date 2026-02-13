# Security Testing Approach

**Last Updated**: 2026-02-13  
**Status**: Active  
**Related**: [Security Testing Guide](guides/frameworks/security-testing-guide.md)

## Overview

This document describes the comprehensive security testing approach for OneMount's authentication and account management system. It covers unit tests, integration tests, property-based tests, and manual verification procedures.

## Table of Contents

1. [Testing Strategy](#testing-strategy)
2. [Unit Tests](#unit-tests)
3. [Integration Tests](#integration-tests)
4. [Property-Based Tests](#property-based-tests)
5. [Manual Verification](#manual-verification)
6. [Security Test Scenarios](#security-test-scenarios)
7. [Continuous Security Testing](#continuous-security-testing)

## Testing Strategy

### Test Pyramid

```
        /\
       /  \      Manual Security Audits
      /____\
     /      \    Property-Based Tests
    /________\
   /          \  Integration Tests
  /____________\
 /              \ Unit Tests
/________________\
```

### Security Testing Principles

1. **Defense in Depth**: Test security at multiple layers
2. **Fail Secure**: Verify system fails safely when security checks fail
3. **Least Privilege**: Verify minimal permissions are used
4. **Secure by Default**: Verify secure defaults are applied
5. **No Security Through Obscurity**: Security doesn't rely on hidden implementation

### Test Coverage Goals

- **Unit Tests**: 80%+ coverage of security-critical code
- **Integration Tests**: All authentication flows covered
- **Property-Based Tests**: All security properties validated
- **Manual Tests**: Quarterly security audits

## Unit Tests

### Token Storage Tests

**Location**: `internal/graph/oauth2_test.go`

**Coverage**:
- Token file creation with correct permissions
- Token encryption/decryption
- Token path generation
- Account hash generation
- Token serialization/deserialization

**Example**:
```go
func TestUT_GR_AUTH_01_TokenFilePermissions(t *testing.T) {
    // Create token file
    auth := &Auth{...}
    path := "/tmp/test_tokens.json"
    
    // Save tokens
    err := auth.ToFile(path)
    require.NoError(t, err)
    
    // Verify permissions are 0600
    info, err := os.Stat(path)
    require.NoError(t, err)
    assert.Equal(t, os.FileMode(0600), info.Mode().Perm())
}
```

### Token Encryption Tests

**Coverage**:
- AES-256 encryption applied
- Decryption successful
- Encryption key derivation
- Encrypted data format

**Example**:
```go
func TestUT_GR_AUTH_02_TokenEncryption(t *testing.T) {
    // Create auth tokens
    auth := &Auth{
        AccessToken:  "test-access-token",
        RefreshToken: "test-refresh-token",
    }
    
    // Encrypt tokens
    encrypted, err := encryptTokens(auth)
    require.NoError(t, err)
    
    // Verify encrypted data is different from plaintext
    assert.NotContains(t, string(encrypted), "test-access-token")
    
    // Decrypt tokens
    decrypted, err := decryptTokens(encrypted)
    require.NoError(t, err)
    
    // Verify decrypted tokens match original
    assert.Equal(t, auth.AccessToken, decrypted.AccessToken)
    assert.Equal(t, auth.RefreshToken, decrypted.RefreshToken)
}
```

### Account Hash Tests

**Coverage**:
- Hash generation is deterministic
- Hash is collision-resistant
- Email normalization (case-insensitive, trimmed)
- Hash length is correct (16 characters)

**Example**:
```go
func TestUT_GR_AUTH_03_AccountHash(t *testing.T) {
    tests := []struct {
        email    string
        expected string
    }{
        {"user@example.com", "a1b2c3d4e5f6g7h8"},
        {"User@Example.com", "a1b2c3d4e5f6g7h8"}, // Case-insensitive
        {" user@example.com ", "a1b2c3d4e5f6g7h8"}, // Trimmed
    }
    
    for _, tt := range tests {
        hash := hashAccount(tt.email)
        assert.Equal(t, tt.expected, hash)
        assert.Len(t, hash, 16)
    }
}
```

### Token Path Tests

**Coverage**:
- Account-based path generation
- Instance-based path generation (legacy)
- Path escaping
- XDG directory compliance

**Example**:
```go
func TestUT_GR_AUTH_04_TokenPaths(t *testing.T) {
    cacheDir := "/home/user/.cache/onemount"
    email := "user@example.com"
    
    // Test account-based path
    path := GetAuthTokensPathByAccount(cacheDir, email)
    expected := "/home/user/.cache/onemount/accounts/a1b2c3d4e5f6g7h8/auth_tokens.json"
    assert.Equal(t, expected, path)
    
    // Test instance-based path (legacy)
    instance := "home-user-OneDrive"
    legacyPath := GetAuthTokensPath(cacheDir, instance)
    expectedLegacy := "/home/user/.cache/onemount/home-user-OneDrive/auth_tokens.json"
    assert.Equal(t, expectedLegacy, legacyPath)
}
```

## Integration Tests

### OAuth2 Flow Tests

**Location**: `internal/graph/oauth2_integration_test.go`

**Coverage**:
- Complete OAuth2 authentication flow
- Token refresh mechanism
- Authentication failure handling
- Headless authentication
- Interactive authentication (GTK)

**Example**:
```go
func TestIT_GR_AUTH_01_OAuth2Flow(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }
    
    // Setup mock OAuth2 server
    server := setupMockOAuth2Server(t)
    defer server.Close()
    
    // Create OAuth2 config
    config := &oauth2.Config{
        ClientID:     "test-client-id",
        ClientSecret: "test-client-secret",
        Endpoint: oauth2.Endpoint{
            AuthURL:  server.URL + "/authorize",
            TokenURL: server.URL + "/token",
        },
    }
    
    // Perform authentication
    auth, err := Authenticate(config, "/tmp/test_tokens.json", false)
    require.NoError(t, err)
    require.NotNil(t, auth)
    
    // Verify tokens were saved
    assert.FileExists(t, "/tmp/test_tokens.json")
    
    // Verify file permissions
    info, err := os.Stat("/tmp/test_tokens.json")
    require.NoError(t, err)
    assert.Equal(t, os.FileMode(0600), info.Mode().Perm())
}
```

### Token Refresh Tests

**Coverage**:
- Automatic token refresh before expiration
- Refresh with valid refresh token
- Refresh failure handling
- Re-authentication on refresh failure

**Example**:
```go
func TestIT_GR_AUTH_02_TokenRefresh(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }
    
    // Create auth with expired access token
    auth := &Auth{
        AccessToken:  "expired-token",
        RefreshToken: "valid-refresh-token",
        ExpiresAt:    time.Now().Add(-1 * time.Hour).Unix(),
    }
    
    // Setup mock refresh endpoint
    server := setupMockRefreshServer(t)
    defer server.Close()
    
    // Attempt to refresh
    err := auth.Refresh()
    require.NoError(t, err)
    
    // Verify new access token
    assert.NotEqual(t, "expired-token", auth.AccessToken)
    assert.True(t, auth.ExpiresAt > time.Now().Unix())
}
```

### Multi-Account Tests

**Location**: `internal/graph/multi_account_integration_test.go`

**Coverage**:
- Multiple accounts mounted simultaneously
- Separate token storage per account
- Separate cache directories per account
- Account isolation

**Example**:
```go
func TestIT_GR_AUTH_03_MultiAccount(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }
    
    // Authenticate two different accounts
    auth1, err := Authenticate(config1, "/tmp/account1_tokens.json", false)
    require.NoError(t, err)
    
    auth2, err := Authenticate(config2, "/tmp/account2_tokens.json", false)
    require.NoError(t, err)
    
    // Verify separate token files
    assert.FileExists(t, "/tmp/account1_tokens.json")
    assert.FileExists(t, "/tmp/account2_tokens.json")
    
    // Verify tokens are different
    assert.NotEqual(t, auth1.AccessToken, auth2.AccessToken)
    assert.NotEqual(t, auth1.RefreshToken, auth2.RefreshToken)
}
```

### TLS/HTTPS Tests

**Location**: `internal/graph/oauth2_integration_test.go`

**Coverage**:
- All API calls use HTTPS
- Certificate validation
- TLS version enforcement (1.2+)
- TLS error handling

**Example**:
```go
func TestIT_GR_AUTH_04_TLSEnforcement(t *testing.T) {
    // Setup HTTP (non-TLS) server
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
    }))
    defer server.Close()
    
    // Attempt to use HTTP endpoint
    config := &oauth2.Config{
        Endpoint: oauth2.Endpoint{
            TokenURL: server.URL, // HTTP, not HTTPS
        },
    }
    
    // Should fail due to non-HTTPS
    _, err := getAuthTokens(config, "test-code")
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "https required")
}
```

## Property-Based Tests

### Overview

Property-based tests verify security properties hold across a wide range of inputs. We use the `testing/quick` package for property-based testing.

**Location**: `internal/graph/oauth2_property_test.go`

### Property 43: Token Encryption at Rest

**Property**: *For any* authentication token storage operation, the system encrypts tokens using AES-256 encryption.

**Test**:
```go
func TestProperty43_TokenEncryptionAtRest(t *testing.T) {
    property := func(accessToken, refreshToken string) bool {
        // Create auth tokens
        auth := &Auth{
            AccessToken:  accessToken,
            RefreshToken: refreshToken,
        }
        
        // Save to file
        path := "/tmp/test_tokens.json"
        err := auth.ToFile(path)
        if err != nil {
            return false
        }
        defer os.Remove(path)
        
        // Read raw file content
        content, err := os.ReadFile(path)
        if err != nil {
            return false
        }
        
        // Verify tokens are not in plaintext
        return !bytes.Contains(content, []byte(accessToken)) &&
               !bytes.Contains(content, []byte(refreshToken))
    }
    
    if err := quick.Check(property, &quick.Config{MaxCount: 100}); err != nil {
        t.Error(err)
    }
}
```

### Property 44: Token File Permissions

**Property**: *For any* token storage file creation, the system sets file permissions to 0600.

**Test**:
```go
func TestProperty44_TokenFilePermissions(t *testing.T) {
    property := func(accessToken, refreshToken string) bool {
        // Create auth tokens
        auth := &Auth{
            AccessToken:  accessToken,
            RefreshToken: refreshToken,
        }
        
        // Save to file
        path := "/tmp/test_tokens.json"
        err := auth.ToFile(path)
        if err != nil {
            return false
        }
        defer os.Remove(path)
        
        // Check file permissions
        info, err := os.Stat(path)
        if err != nil {
            return false
        }
        
        return info.Mode().Perm() == 0600
    }
    
    if err := quick.Check(property, &quick.Config{MaxCount: 100}); err != nil {
        t.Error(err)
    }
}
```

### Property 45: Secure Token Storage Location

**Property**: *For any* authentication token storage, the system stores tokens in the XDG cache directory with restricted access.

**Test**:
```go
func TestProperty45_SecureTokenStorageLocation(t *testing.T) {
    property := func(email string) bool {
        if email == "" {
            return true // Skip empty emails
        }
        
        // Get token path
        cacheDir, _ := os.UserCacheDir()
        path := GetAuthTokensPathByAccount(cacheDir, email)
        
        // Verify path is in cache directory
        return strings.HasPrefix(path, cacheDir)
    }
    
    if err := quick.Check(property, &quick.Config{MaxCount: 100}); err != nil {
        t.Error(err)
    }
}
```

### Property 46: HTTPS/TLS Communication

**Property**: *For any* Microsoft Graph API communication, the system uses HTTPS/TLS 1.2 or higher.

**Test**:
```go
func TestProperty46_HTTPSTLSCommunication(t *testing.T) {
    property := func(endpoint string) bool {
        // Create HTTP client with TLS config
        client := &http.Client{
            Transport: &http.Transport{
                TLSClientConfig: &tls.Config{
                    MinVersion: tls.VersionTLS12,
                },
            },
        }
        
        // Attempt request
        req, _ := http.NewRequest("GET", endpoint, nil)
        resp, err := client.Do(req)
        
        // If successful, verify TLS was used
        if err == nil && resp != nil {
            defer resp.Body.Close()
            return resp.TLS != nil && resp.TLS.Version >= tls.VersionTLS12
        }
        
        // If failed, verify it's not due to TLS downgrade
        return err != nil && !strings.Contains(err.Error(), "tls: protocol version not supported")
    }
    
    if err := quick.Check(property, &quick.Config{MaxCount: 100}); err != nil {
        t.Error(err)
    }
}
```

### Property 47: Sensitive Data Logging Prevention

**Property**: *For any* logging operation, the system never logs authentication tokens, passwords, or sensitive user data.

**Test**:
```go
func TestProperty47_SensitiveDataLoggingPrevention(t *testing.T) {
    property := func(accessToken, refreshToken string) bool {
        // Create logger that captures output
        var buf bytes.Buffer
        logger := log.New(&buf, "", 0)
        
        // Create auth tokens
        auth := &Auth{
            AccessToken:  accessToken,
            RefreshToken: refreshToken,
        }
        
        // Perform operations that might log
        logger.Printf("Auth: %+v", auth)
        logger.Printf("Token refresh: %v", auth.Refresh())
        
        // Verify tokens are not in logs
        logOutput := buf.String()
        return !strings.Contains(logOutput, accessToken) &&
               !strings.Contains(logOutput, refreshToken)
    }
    
    if err := quick.Check(property, &quick.Config{MaxCount: 100}); err != nil {
        t.Error(err)
    }
}
```

### Property 48: Cache File Security

**Property**: *For any* cached file content storage, the system sets appropriate file permissions to prevent unauthorized access.

**Test**:
```go
func TestProperty48_CacheFileSecurity(t *testing.T) {
    property := func(content []byte) bool {
        // Create cache file
        cacheDir, _ := os.UserCacheDir()
        path := filepath.Join(cacheDir, "onemount", "test_cache.dat")
        
        // Write content
        err := os.WriteFile(path, content, 0600)
        if err != nil {
            return false
        }
        defer os.Remove(path)
        
        // Verify permissions
        info, err := os.Stat(path)
        if err != nil {
            return false
        }
        
        return info.Mode().Perm() == 0600
    }
    
    if err := quick.Check(property, &quick.Config{MaxCount: 100}); err != nil {
        t.Error(err)
    }
}
```

## Manual Verification

### Token File Permissions Audit

**Frequency**: Quarterly

**Procedure**:
1. List all token files:
   ```bash
   find ~/.cache/onemount -name "auth_tokens.json"
   ```

2. Check permissions:
   ```bash
   ls -l ~/.cache/onemount/accounts/*/auth_tokens.json
   ```

3. Verify all files have `0600` permissions

4. Document findings in security audit report

### TLS/HTTPS Verification

**Frequency**: Quarterly

**Procedure**:
1. Capture network traffic during authentication:
   ```bash
   tcpdump -i any -w auth_traffic.pcap host graph.microsoft.com
   ```

2. Analyze with Wireshark:
   - Verify all traffic is TLS encrypted
   - Verify TLS version is 1.2 or higher
   - Verify certificate validation occurs

3. Document findings in security audit report

### Token Logging Audit

**Frequency**: Quarterly

**Procedure**:
1. Enable debug logging:
   ```bash
   onemount --debug ~/OneDrive 2>&1 | tee onemount.log
   ```

2. Perform authentication and operations

3. Search logs for tokens:
   ```bash
   grep -i "token" onemount.log
   grep -i "access" onemount.log
   grep -i "refresh" onemount.log
   ```

4. Verify no actual token values appear in logs

5. Document findings in security audit report

### Penetration Testing

**Frequency**: Annually

**Scope**:
- Token theft attempts
- Man-in-the-middle attacks
- Privilege escalation
- Authentication bypass
- Token replay attacks

**Procedure**:
1. Engage security professional or use automated tools
2. Test in isolated environment
3. Document vulnerabilities found
4. Create remediation plan
5. Verify fixes with re-test

## Security Test Scenarios

### Scenario 1: Token Theft Prevention

**Objective**: Verify tokens cannot be stolen by unauthorized users

**Steps**:
1. Create token file as user A
2. Attempt to read token file as user B
3. Verify access denied

**Expected Result**: User B cannot read token file

### Scenario 2: Network Interception Prevention

**Objective**: Verify tokens cannot be intercepted on network

**Steps**:
1. Setup network sniffer
2. Perform authentication
3. Analyze captured traffic

**Expected Result**: All traffic is TLS encrypted, tokens not visible

### Scenario 3: Token Replay Attack Prevention

**Objective**: Verify expired tokens cannot be reused

**Steps**:
1. Capture valid access token
2. Wait for token to expire
3. Attempt to use expired token

**Expected Result**: API request fails with authentication error

### Scenario 4: Brute Force Prevention

**Objective**: Verify rate limiting prevents brute force attacks

**Steps**:
1. Attempt authentication with invalid credentials
2. Repeat rapidly (100+ attempts)
3. Observe rate limiting behavior

**Expected Result**: Rate limiting triggers, delays increase exponentially

### Scenario 5: Multi-Account Isolation

**Objective**: Verify accounts don't interfere with each other

**Steps**:
1. Mount two different accounts
2. Perform operations on account A
3. Verify account B is unaffected

**Expected Result**: Accounts remain isolated, no cross-contamination

## Continuous Security Testing

### CI/CD Integration

**Pre-Commit Hooks**:
- Run unit tests including security tests
- Check for hardcoded credentials
- Verify no tokens in commits

**Pull Request Checks**:
- Run full test suite including property-based tests
- Security code review
- Dependency vulnerability scan

**Nightly Builds**:
- Run integration tests with real OneDrive
- Run property-based tests with high iteration count
- Generate security test report

**Release Process**:
- Run full security test suite
- Perform manual security audit
- Update security documentation
- Generate security release notes

### Security Monitoring

**Metrics**:
- Test coverage of security-critical code
- Number of security tests passing/failing
- Time to fix security issues
- Number of security vulnerabilities found

**Alerts**:
- Security test failures
- Dependency vulnerabilities
- Suspicious authentication patterns
- Token file permission changes

### Security Updates

**Process**:
1. Monitor security advisories for dependencies
2. Evaluate impact on OneMount
3. Update dependencies if needed
4. Run full security test suite
5. Release security update if critical

## References

- [Security Testing Guide](guides/frameworks/security-testing-guide.md)
- [Authentication Architecture](../2-architecture/authentication.md)
- [Test Plan](test-plan.md)
- [Requirements Traceability Matrix](requirements-traceability-matrix-updated.md)
