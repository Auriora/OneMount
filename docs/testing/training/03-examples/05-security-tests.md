# Security Test Examples

This file contains examples of security tests using the OneMount test framework. Security tests help you identify and address potential security vulnerabilities in your code.

> **Note**: All code examples in this document are for illustration purposes only and may need to be adapted to your specific project structure and imports. The examples are not meant to be compiled directly but rather to demonstrate concepts and patterns.

## Table of Contents

1. [Introduction to Security Testing](#introduction-to-security-testing)
2. [Authentication Tests](#authentication-tests)
3. [Authorization Tests](#authorization-tests)
4. [Input Validation Tests](#input-validation-tests)
5. [Secure Storage Tests](#secure-storage-tests)
6. [Network Security Tests](#network-security-tests)

## Introduction to Security Testing

Security testing is a critical part of ensuring that your application is protected against potential threats. It helps you:

- Identify and address security vulnerabilities
- Ensure that sensitive data is properly protected
- Verify that authentication and authorization mechanisms work correctly
- Validate input to prevent injection attacks
- Ensure secure communication over networks

The OneMount test framework provides tools for writing and running security tests, including the SecurityTestEnvironment component.

## Authentication Tests

Here's an example of a security test that verifies the authentication mechanism:

```go
package security_test

import (
    "context"
    "testing"
    
    "github.com/stretchr/testify/require"
    "github.com/yourusername/onemount/internal/auth"
    "github.com/yourusername/onemount/internal/testutil"
)

// TestAuthenticationSecurity tests the security of the authentication mechanism
func TestAuthenticationSecurity(t *testing.T) {
    // Create a logger
    logger := testutil.NewTestLogger()
    
    // Create a security test environment
    env := testutil.NewSecurityTestEnvironment(context.Background(), logger)
    require.NotNil(t, env)
    
    // Set up the environment
    err := env.SetupEnvironment()
    require.NoError(t, err)
    
    // Add cleanup using t.Cleanup to ensure resources are cleaned up
    t.Cleanup(func() {
        env.TeardownEnvironment()
    })
    
    // Get the auth component
    authComponent, err := env.GetComponent("auth")
    require.NoError(t, err)
    authService := authComponent.(*auth.AuthService)
    
    // Test cases for authentication security
    testCases := []struct {
        name        string
        username    string
        password    string
        expectError bool
        errorType   error
    }{
        {
            name:        "ValidCredentials",
            username:    "validuser",
            password:    "ValidP@ssw0rd",
            expectError: false,
        },
        {
            name:        "EmptyUsername",
            username:    "",
            password:    "ValidP@ssw0rd",
            expectError: true,
            errorType:   auth.ErrInvalidCredentials,
        },
        {
            name:        "EmptyPassword",
            username:    "validuser",
            password:    "",
            expectError: true,
            errorType:   auth.ErrInvalidCredentials,
        },
        {
            name:        "InvalidUsername",
            username:    "invaliduser",
            password:    "ValidP@ssw0rd",
            expectError: true,
            errorType:   auth.ErrInvalidCredentials,
        },
        {
            name:        "InvalidPassword",
            username:    "validuser",
            password:    "invalidpassword",
            expectError: true,
            errorType:   auth.ErrInvalidCredentials,
        },
        {
            name:        "SQLInjectionAttempt",
            username:    "validuser' OR '1'='1",
            password:    "ValidP@ssw0rd",
            expectError: true,
            errorType:   auth.ErrInvalidCredentials,
        },
        {
            name:        "BruteForceProtection",
            username:    "validuser",
            password:    "attempt1",
            expectError: true,
            errorType:   auth.ErrInvalidCredentials,
        },
    }
    
    // Run test cases
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            // If testing brute force protection, make multiple failed attempts first
            if tc.name == "BruteForceProtection" {
                for i := 0; i < 5; i++ {
                    _, err := authService.Authenticate(tc.username, "wrong"+string(i))
                    require.Error(t, err, "Authentication should fail with wrong password")
                }
                
                // Now try with the correct password, but it should be blocked due to too many failed attempts
                tc.password = "ValidP@ssw0rd" // Use the correct password
                tc.errorType = auth.ErrTooManyFailedAttempts
            }
            
            // Attempt authentication
            token, err := authService.Authenticate(tc.username, tc.password)
            
            // Verify expectations
            if tc.expectError {
                require.Error(t, err, "Authentication should fail")
                if tc.errorType != nil {
                    require.ErrorIs(t, err, tc.errorType, "Error type should match")
                }
                require.Empty(t, token, "Token should be empty on error")
            } else {
                require.NoError(t, err, "Authentication should succeed")
                require.NotEmpty(t, token, "Token should not be empty on success")
                
                // Verify token is valid
                valid, err := authService.ValidateToken(token)
                require.NoError(t, err, "Token validation should succeed")
                require.True(t, valid, "Token should be valid")
            }
        })
    }
    
    // Test token expiration
    t.Run("TokenExpiration", func(t *testing.T) {
        // Authenticate to get a token
        token, err := authService.Authenticate("validuser", "ValidP@ssw0rd")
        require.NoError(t, err, "Authentication should succeed")
        
        // Verify token is valid
        valid, err := authService.ValidateToken(token)
        require.NoError(t, err, "Token validation should succeed")
        require.True(t, valid, "Token should be valid")
        
        // Simulate token expiration
        env.SimulateTimePassage(24 * time.Hour) // Advance time by 24 hours
        
        // Verify token is now expired
        valid, err = authService.ValidateToken(token)
        require.NoError(t, err, "Token validation should not error for expired token")
        require.False(t, valid, "Token should be expired")
    })
}
```

This test:
1. Sets up a security test environment
2. Tests various authentication scenarios, including valid credentials, invalid credentials, and potential attack vectors
3. Verifies that the authentication mechanism correctly handles each scenario
4. Tests token expiration to ensure that tokens are not valid indefinitely

## Authorization Tests

Here's an example of a security test that verifies the authorization mechanism:

```go
package security_test

import (
    "context"
    "testing"
    
    "github.com/stretchr/testify/require"
    "github.com/yourusername/onemount/internal/auth"
    "github.com/yourusername/onemount/internal/fs"
    "github.com/yourusername/onemount/internal/testutil"
)

// TestAuthorizationSecurity tests the security of the authorization mechanism
func TestAuthorizationSecurity(t *testing.T) {
    // Create a logger
    logger := testutil.NewTestLogger()
    
    // Create a security test environment
    env := testutil.NewSecurityTestEnvironment(context.Background(), logger)
    require.NotNil(t, env)
    
    // Set up the environment
    err := env.SetupEnvironment()
    require.NoError(t, err)
    
    // Add cleanup using t.Cleanup to ensure resources are cleaned up
    t.Cleanup(func() {
        env.TeardownEnvironment()
    })
    
    // Get the auth component
    authComponent, err := env.GetComponent("auth")
    require.NoError(t, err)
    authService := authComponent.(*auth.AuthService)
    
    // Get the file manager component
    fsComponent, err := env.GetComponent("fs")
    require.NoError(t, err)
    fileManager := fsComponent.(*fs.FileManager)
    
    // Create test users with different roles
    adminUser := &auth.User{
        Username: "admin",
        Password: "AdminP@ss123",
        Roles:    []string{"admin", "user"},
    }
    
    regularUser := &auth.User{
        Username: "user",
        Password: "UserP@ss123",
        Roles:    []string{"user"},
    }
    
    guestUser := &auth.User{
        Username: "guest",
        Password: "GuestP@ss123",
        Roles:    []string{"guest"},
    }
    
    // Add users to the auth service
    err = authService.AddUser(adminUser)
    require.NoError(t, err, "Failed to add admin user")
    
    err = authService.AddUser(regularUser)
    require.NoError(t, err, "Failed to add regular user")
    
    err = authService.AddUser(guestUser)
    require.NoError(t, err, "Failed to add guest user")
    
    // Create test files with different permissions
    adminFile := "/admin-only.txt"
    userFile := "/user-readable.txt"
    publicFile := "/public.txt"
    
    // Create the files
    err = fileManager.CreateFile(adminFile, []byte("Admin only content"), fs.FilePermissions{
        ReadRoles:  []string{"admin"},
        WriteRoles: []string{"admin"},
    })
    require.NoError(t, err, "Failed to create admin file")
    
    err = fileManager.CreateFile(userFile, []byte("User readable content"), fs.FilePermissions{
        ReadRoles:  []string{"admin", "user"},
        WriteRoles: []string{"admin"},
    })
    require.NoError(t, err, "Failed to create user file")
    
    err = fileManager.CreateFile(publicFile, []byte("Public content"), fs.FilePermissions{
        ReadRoles:  []string{"admin", "user", "guest"},
        WriteRoles: []string{"admin", "user"},
    })
    require.NoError(t, err, "Failed to create public file")
    
    // Test cases for authorization
    testCases := []struct {
        name        string
        user        *auth.User
        operation   string
        file        string
        expectError bool
    }{
        // Admin user tests
        {
            name:        "AdminReadAdminFile",
            user:        adminUser,
            operation:   "read",
            file:        adminFile,
            expectError: false,
        },
        {
            name:        "AdminWriteAdminFile",
            user:        adminUser,
            operation:   "write",
            file:        adminFile,
            expectError: false,
        },
        {
            name:        "AdminReadUserFile",
            user:        adminUser,
            operation:   "read",
            file:        userFile,
            expectError: false,
        },
        {
            name:        "AdminWriteUserFile",
            user:        adminUser,
            operation:   "write",
            file:        userFile,
            expectError: false,
        },
        {
            name:        "AdminReadPublicFile",
            user:        adminUser,
            operation:   "read",
            file:        publicFile,
            expectError: false,
        },
        {
            name:        "AdminWritePublicFile",
            user:        adminUser,
            operation:   "write",
            file:        publicFile,
            expectError: false,
        },
        
        // Regular user tests
        {
            name:        "UserReadAdminFile",
            user:        regularUser,
            operation:   "read",
            file:        adminFile,
            expectError: true,
        },
        {
            name:        "UserWriteAdminFile",
            user:        regularUser,
            operation:   "write",
            file:        adminFile,
            expectError: true,
        },
        {
            name:        "UserReadUserFile",
            user:        regularUser,
            operation:   "read",
            file:        userFile,
            expectError: false,
        },
        {
            name:        "UserWriteUserFile",
            user:        regularUser,
            operation:   "write",
            file:        userFile,
            expectError: true,
        },
        {
            name:        "UserReadPublicFile",
            user:        regularUser,
            operation:   "read",
            file:        publicFile,
            expectError: false,
        },
        {
            name:        "UserWritePublicFile",
            user:        regularUser,
            operation:   "write",
            file:        publicFile,
            expectError: false,
        },
        
        // Guest user tests
        {
            name:        "GuestReadAdminFile",
            user:        guestUser,
            operation:   "read",
            file:        adminFile,
            expectError: true,
        },
        {
            name:        "GuestWriteAdminFile",
            user:        guestUser,
            operation:   "write",
            file:        adminFile,
            expectError: true,
        },
        {
            name:        "GuestReadUserFile",
            user:        guestUser,
            operation:   "read",
            file:        userFile,
            expectError: true,
        },
        {
            name:        "GuestWriteUserFile",
            user:        guestUser,
            operation:   "write",
            file:        userFile,
            expectError: true,
        },
        {
            name:        "GuestReadPublicFile",
            user:        guestUser,
            operation:   "read",
            file:        publicFile,
            expectError: false,
        },
        {
            name:        "GuestWritePublicFile",
            user:        guestUser,
            operation:   "write",
            file:        publicFile,
            expectError: true,
        },
    }
    
    // Run test cases
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            // Authenticate as the user
            token, err := authService.Authenticate(tc.user.Username, tc.user.Password)
            require.NoError(t, err, "Authentication should succeed")
            
            // Create a context with the authentication token
            ctx := auth.ContextWithToken(context.Background(), token)
            
            // Perform the operation
            var opErr error
            if tc.operation == "read" {
                _, opErr = fileManager.GetFile(ctx, tc.file)
            } else if tc.operation == "write" {
                opErr = fileManager.UpdateFile(ctx, tc.file, []byte("Updated content"))
            }
            
            // Verify expectations
            if tc.expectError {
                require.Error(t, opErr, "Operation should fail due to insufficient permissions")
                require.ErrorIs(t, opErr, fs.ErrPermissionDenied, "Error should be permission denied")
            } else {
                require.NoError(t, opErr, "Operation should succeed with sufficient permissions")
            }
        })
    }
    
    // Test privilege escalation attempt
    t.Run("PrivilegeEscalationAttempt", func(t *testing.T) {
        // Authenticate as a regular user
        token, err := authService.Authenticate(regularUser.Username, regularUser.Password)
        require.NoError(t, err, "Authentication should succeed")
        
        // Create a context with the authentication token
        ctx := auth.ContextWithToken(context.Background(), token)
        
        // Attempt to modify the token to add admin role
        modifiedCtx := auth.ContextWithRoles(ctx, []string{"admin", "user"})
        
        // Attempt to access admin file with modified context
        _, err = fileManager.GetFile(modifiedCtx, adminFile)
        
        // Verify that the attempt fails
        require.Error(t, err, "Access should be denied despite role modification attempt")
        require.ErrorIs(t, err, fs.ErrPermissionDenied, "Error should be permission denied")
    })
}
```

This test:
1. Sets up a security test environment
2. Creates users with different roles (admin, user, guest)
3. Creates files with different permission levels
4. Tests various access scenarios to verify that the authorization mechanism correctly enforces permissions
5. Tests a privilege escalation attempt to ensure that users cannot gain unauthorized access

## Input Validation Tests

Here's an example of a security test that verifies input validation:

```go
package security_test

import (
    "context"
    "testing"
    
    "github.com/stretchr/testify/require"
    "github.com/yourusername/onemount/internal/fs"
    "github.com/yourusername/onemount/internal/testutil"
)

// TestInputValidationSecurity tests the security of input validation
func TestInputValidationSecurity(t *testing.T) {
    // Create a logger
    logger := testutil.NewTestLogger()
    
    // Create a security test environment
    env := testutil.NewSecurityTestEnvironment(context.Background(), logger)
    require.NotNil(t, env)
    
    // Set up the environment
    err := env.SetupEnvironment()
    require.NoError(t, err)
    
    // Add cleanup using t.Cleanup to ensure resources are cleaned up
    t.Cleanup(func() {
        env.TeardownEnvironment()
    })
    
    // Get the file manager component
    fsComponent, err := env.GetComponent("fs")
    require.NoError(t, err)
    fileManager := fsComponent.(*fs.FileManager)
    
    // Test cases for path validation
    pathTestCases := []struct {
        name        string
        path        string
        expectError bool
        errorType   error
    }{
        {
            name:        "ValidPath",
            path:        "/valid/path/file.txt",
            expectError: false,
        },
        {
            name:        "PathTraversal",
            path:        "/../../../etc/passwd",
            expectError: true,
            errorType:   fs.ErrInvalidPath,
        },
        {
            name:        "NullByteInjection",
            path:        "/valid/path\x00/etc/passwd",
            expectError: true,
            errorType:   fs.ErrInvalidPath,
        },
        {
            name:        "CommandInjection",
            path:        "/valid/path; rm -rf /",
            expectError: true,
            errorType:   fs.ErrInvalidPath,
        },
        {
            name:        "SpecialCharacters",
            path:        "/valid/path/<script>alert('XSS')</script>",
            expectError: true,
            errorType:   fs.ErrInvalidPath,
        },
        {
            name:        "ExcessiveLength",
            path:        "/" + string(make([]byte, 4096)), // Very long path
            expectError: true,
            errorType:   fs.ErrInvalidPath,
        },
    }
    
    // Run path validation test cases
    for _, tc := range pathTestCases {
        t.Run(tc.name, func(t *testing.T) {
            // Attempt to create a file with the test path
            err := fileManager.CreateFile(tc.path, []byte("Test content"))
            
            // Verify expectations
            if tc.expectError {
                require.Error(t, err, "Path validation should fail")
                if tc.errorType != nil {
                    require.ErrorIs(t, err, tc.errorType, "Error type should match")
                }
            } else {
                require.NoError(t, err, "Path validation should succeed")
                
                // Clean up the created file
                err = fileManager.DeleteFile(tc.path)
                require.NoError(t, err, "Failed to clean up test file")
            }
        })
    }
    
    // Test cases for content validation
    contentTestCases := []struct {
        name        string
        content     []byte
        expectError bool
        errorType   error
    }{
        {
            name:        "ValidContent",
            content:     []byte("Valid content"),
            expectError: false,
        },
        {
            name:        "NullByteInContent",
            content:     []byte("Content with null byte\x00"),
            expectError: true,
            errorType:   fs.ErrInvalidContent,
        },
        {
            name:        "ExcessiveSize",
            content:     make([]byte, 100*1024*1024), // 100 MB (assuming there's a size limit)
            expectError: true,
            errorType:   fs.ErrContentTooLarge,
        },
        {
            name:        "MaliciousContent",
            content:     []byte("<script>alert('XSS')</script>"),
            expectError: false, // Content validation might not block this, but it should be sanitized when displayed
        },
    }
    
    // Run content validation test cases
    for _, tc := range contentTestCases {
        t.Run(tc.name, func(t *testing.T) {
            // Create a valid path for testing content
            path := "/test-content-" + tc.name + ".txt"
            
            // Attempt to create a file with the test content
            err := fileManager.CreateFile(path, tc.content)
            
            // Verify expectations
            if tc.expectError {
                require.Error(t, err, "Content validation should fail")
                if tc.errorType != nil {
                    require.ErrorIs(t, err, tc.errorType, "Error type should match")
                }
            } else {
                require.NoError(t, err, "Content validation should succeed")
                
                // Verify that the content was stored correctly
                file, err := fileManager.GetFile(path)
                require.NoError(t, err, "Failed to get file")
                
                content, err := fileManager.GetFileContent(file)
                require.NoError(t, err, "Failed to get file content")
                require.Equal(t, tc.content, content, "File content should match")
                
                // Clean up the created file
                err = fileManager.DeleteFile(path)
                require.NoError(t, err, "Failed to clean up test file")
            }
        })
    }
}
```

This test:
1. Sets up a security test environment
2. Tests path validation to prevent path traversal and other path-based attacks
3. Tests content validation to ensure that file content is properly validated
4. Verifies that the system correctly handles both valid and invalid inputs

## Secure Storage Tests

Here's an example of a security test that verifies secure storage:

```go
package security_test

import (
    "context"
    "testing"
    
    "github.com/stretchr/testify/require"
    "github.com/yourusername/onemount/internal/auth"
    "github.com/yourusername/onemount/internal/storage"
    "github.com/yourusername/onemount/internal/testutil"
)

// TestSecureStorageSecurity tests the security of secure storage
func TestSecureStorageSecurity(t *testing.T) {
    // Create a logger
    logger := testutil.NewTestLogger()
    
    // Create a security test environment
    env := testutil.NewSecurityTestEnvironment(context.Background(), logger)
    require.NotNil(t, env)
    
    // Set up the environment
    err := env.SetupEnvironment()
    require.NoError(t, err)
    
    // Add cleanup using t.Cleanup to ensure resources are cleaned up
    t.Cleanup(func() {
        env.TeardownEnvironment()
    })
    
    // Get the secure storage component
    storageComponent, err := env.GetComponent("storage")
    require.NoError(t, err)
    secureStorage := storageComponent.(*storage.SecureStorage)
    
    // Test storing and retrieving sensitive data
    t.Run("StoreAndRetrieveSensitiveData", func(t *testing.T) {
        // Define sensitive data
        key := "api_key"
        value := "secret_api_key_12345"
        
        // Store the sensitive data
        err := secureStorage.Store(key, []byte(value))
        require.NoError(t, err, "Failed to store sensitive data")
        
        // Retrieve the sensitive data
        retrievedValue, err := secureStorage.Retrieve(key)
        require.NoError(t, err, "Failed to retrieve sensitive data")
        require.Equal(t, value, string(retrievedValue), "Retrieved value should match stored value")
    })
    
    // Test data encryption
    t.Run("DataEncryption", func(t *testing.T) {
        // Define sensitive data
        key := "password"
        value := "secure_password_123"
        
        // Store the sensitive data
        err := secureStorage.Store(key, []byte(value))
        require.NoError(t, err, "Failed to store sensitive data")
        
        // Get the raw storage to verify encryption
        rawStorage := env.GetRawStorage()
        
        // Verify that the data is encrypted in the raw storage
        rawValue, exists := rawStorage[key]
        require.True(t, exists, "Key should exist in raw storage")
        require.NotEqual(t, value, string(rawValue), "Value should be encrypted in raw storage")
        
        // Retrieve the sensitive data through the secure storage
        retrievedValue, err := secureStorage.Retrieve(key)
        require.NoError(t, err, "Failed to retrieve sensitive data")
        require.Equal(t, value, string(retrievedValue), "Retrieved value should match original value")
    })
    
    // Test key rotation
    t.Run("KeyRotation", func(t *testing.T) {
        // Define sensitive data
        key := "secret_note"
        value := "This is a secret note"
        
        // Store the sensitive data
        err := secureStorage.Store(key, []byte(value))
        require.NoError(t, err, "Failed to store sensitive data")
        
        // Rotate the encryption key
        err = secureStorage.RotateKey()
        require.NoError(t, err, "Failed to rotate encryption key")
        
        // Verify that the data can still be retrieved
        retrievedValue, err := secureStorage.Retrieve(key)
        require.NoError(t, err, "Failed to retrieve sensitive data after key rotation")
        require.Equal(t, value, string(retrievedValue), "Retrieved value should match original value after key rotation")
    })
    
    // Test access control
    t.Run("AccessControl", func(t *testing.T) {
        // Create users with different roles
        adminUser := &auth.User{
            Username: "admin",
            Password: "AdminP@ss123",
            Roles:    []string{"admin"},
        }
        
        regularUser := &auth.User{
            Username: "user",
            Password: "UserP@ss123",
            Roles:    []string{"user"},
        }
        
        // Get the auth component
        authComponent, err := env.GetComponent("auth")
        require.NoError(t, err)
        authService := authComponent.(*auth.AuthService)
        
        // Add users to the auth service
        err = authService.AddUser(adminUser)
        require.NoError(t, err, "Failed to add admin user")
        
        err = authService.AddUser(regularUser)
        require.NoError(t, err, "Failed to add regular user")
        
        // Store sensitive data with admin access only
        adminKey := "admin_only_key"
        adminValue := "admin_only_value"
        
        // Authenticate as admin
        adminToken, err := authService.Authenticate(adminUser.Username, adminUser.Password)
        require.NoError(t, err, "Admin authentication should succeed")
        
        // Create admin context
        adminCtx := auth.ContextWithToken(context.Background(), adminToken)
        
        // Store data with admin context
        err = secureStorage.StoreWithContext(adminCtx, adminKey, []byte(adminValue))
        require.NoError(t, err, "Admin should be able to store data")
        
        // Authenticate as regular user
        userToken, err := authService.Authenticate(regularUser.Username, regularUser.Password)
        require.NoError(t, err, "User authentication should succeed")
        
        // Create user context
        userCtx := auth.ContextWithToken(context.Background(), userToken)
        
        // Attempt to retrieve admin data with user context
        _, err = secureStorage.RetrieveWithContext(userCtx, adminKey)
        require.Error(t, err, "Regular user should not be able to retrieve admin data")
        require.ErrorIs(t, err, storage.ErrAccessDenied, "Error should be access denied")
        
        // Store user data
        userKey := "user_key"
        userValue := "user_value"
        
        // Store data with user context
        err = secureStorage.StoreWithContext(userCtx, userKey, []byte(userValue))
        require.NoError(t, err, "User should be able to store data")
        
        // Retrieve user data with user context
        retrievedUserValue, err := secureStorage.RetrieveWithContext(userCtx, userKey)
        require.NoError(t, err, "User should be able to retrieve their own data")
        require.Equal(t, userValue, string(retrievedUserValue), "Retrieved value should match stored value")
        
        // Retrieve user data with admin context
        retrievedUserValueByAdmin, err := secureStorage.RetrieveWithContext(adminCtx, userKey)
        require.NoError(t, err, "Admin should be able to retrieve user data")
        require.Equal(t, userValue, string(retrievedUserValueByAdmin), "Retrieved value should match stored value")
    })
}
```

This test:
1. Sets up a security test environment
2. Tests storing and retrieving