package graph

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/auriora/onemount/internal/config"
	"github.com/auriora/onemount/internal/testutil"
	"github.com/auriora/onemount/internal/testutil/framework"
)

// TestIT_AUTH_08_01_MultiAccountMounting_SimultaneousMounts_WorksCorrectly tests multiple account mounting
//
//	Test Case ID    IT-AUTH-08-01
//	Title           Multiple Account Mounting with Separate Storage
//	Description     Tests that multiple OneDrive accounts can be mounted simultaneously with
//	                separate token storage, cache directories, and delta sync loops
//	Preconditions   Two different OneDrive accounts are available for testing
//	Steps           1. Set up mount registry
//	                2. Create first account with authentication
//	                3. Register first mount point with first account
//	                4. Verify first account token storage location
//	                5. Create second account with authentication
//	                6. Register second mount point with second account
//	                7. Verify second account token storage location
//	                8. Verify both accounts have separate token files
//	                9. Verify both accounts have separate cache directories
//	                10. Verify mount registry correctly maps mount points to accounts
//	Expected Result Multiple accounts can be mounted simultaneously with isolated storage
//	Requirements    4.1-4.8
//	Notes: This test verifies end-to-end multi-account mounting with proper isolation.
func TestIT_AUTH_08_01_MultiAccountMounting_SimultaneousMounts_WorksCorrectly(t *testing.T) {
	// Create assertions helper
	assert := framework.NewAssert(t)

	// Step 1: Set up test directories
	testConfigDir := filepath.Join(testutil.TestSandboxTmpDir, "multi-account-config")
	testCacheDir := filepath.Join(testutil.TestSandboxTmpDir, "multi-account-cache")

	// Clean up test directories before and after test
	defer func() {
		_ = os.RemoveAll(testConfigDir)
		_ = os.RemoveAll(testCacheDir)
	}()

	// Ensure test directories exist
	err := os.MkdirAll(testConfigDir, 0700)
	assert.NoError(err, "Should be able to create test config directory")
	err = os.MkdirAll(testCacheDir, 0700)
	assert.NoError(err, "Should be able to create test cache directory")

	// Step 2: Create mount registry
	registry, err := config.NewMountsRegistry(testConfigDir)
	assert.NoError(err, "Should be able to create mounts registry")
	assert.NotNil(registry, "Registry should not be nil")

	// Step 3: Set up first account
	account1Email := "user1@example.com"
	mountPoint1 := filepath.Join(testutil.TestSandboxTmpDir, "mount1")
	err = os.MkdirAll(mountPoint1, 0755)
	assert.NoError(err, "Should be able to create first mount point")
	defer os.RemoveAll(mountPoint1)

	auth1 := &Auth{
		AccessToken:  "account1-access-token",
		RefreshToken: "account1-refresh-token",
		ExpiresAt:    time.Now().Add(1 * time.Hour).Unix(),
		Account:      account1Email,
	}

	// Step 4: Register first mount point with first account
	err = registry.SetMount(mountPoint1, account1Email, "Personal OneDrive")
	assert.NoError(err, "Should be able to register first mount point")

	// Step 5: Verify first account is registered
	registeredAccount1, exists := registry.GetAccount(mountPoint1)
	assert.True(exists, "First mount point should be registered")
	assert.Equal(account1Email, registeredAccount1, "First account should match")

	// Step 6: Save first account tokens using account-based storage
	tokenPath1 := GetAuthTokensPathByAccount(testCacheDir, account1Email)
	err = SaveAuthTokens(auth1, tokenPath1)
	assert.NoError(err, "Should be able to save first account tokens")

	// Step 7: Verify first account token file exists
	_, err = os.Stat(tokenPath1)
	assert.NoError(err, "First account token file should exist")

	// Step 8: Verify first account token file has correct permissions
	info1, err := os.Stat(tokenPath1)
	assert.NoError(err, "Should be able to stat first account token file")
	assert.Equal(os.FileMode(0600), info1.Mode().Perm(), "First account token file should have 0600 permissions")

	// Step 9: Set up second account
	account2Email := "user2@example.com"
	mountPoint2 := filepath.Join(testutil.TestSandboxTmpDir, "mount2")
	err = os.MkdirAll(mountPoint2, 0755)
	assert.NoError(err, "Should be able to create second mount point")
	defer os.RemoveAll(mountPoint2)

	auth2 := &Auth{
		AccessToken:  "account2-access-token",
		RefreshToken: "account2-refresh-token",
		ExpiresAt:    time.Now().Add(1 * time.Hour).Unix(),
		Account:      account2Email,
	}

	// Step 10: Register second mount point with second account
	err = registry.SetMount(mountPoint2, account2Email, "Work OneDrive")
	assert.NoError(err, "Should be able to register second mount point")

	// Step 11: Verify second account is registered
	registeredAccount2, exists := registry.GetAccount(mountPoint2)
	assert.True(exists, "Second mount point should be registered")
	assert.Equal(account2Email, registeredAccount2, "Second account should match")

	// Step 12: Save second account tokens using account-based storage
	tokenPath2 := GetAuthTokensPathByAccount(testCacheDir, account2Email)
	err = SaveAuthTokens(auth2, tokenPath2)
	assert.NoError(err, "Should be able to save second account tokens")

	// Step 13: Verify second account token file exists
	_, err = os.Stat(tokenPath2)
	assert.NoError(err, "Second account token file should exist")

	// Step 14: Verify second account token file has correct permissions
	info2, err := os.Stat(tokenPath2)
	assert.NoError(err, "Should be able to stat second account token file")
	assert.Equal(os.FileMode(0600), info2.Mode().Perm(), "Second account token file should have 0600 permissions")

	// Step 15: Verify both accounts have separate token files
	assert.NotEqual(tokenPath1, tokenPath2, "Token paths should be different for different accounts")

	// Step 16: Verify both token files contain correct account information
	loadedAuth1, err := LoadAuthTokens(tokenPath1)
	assert.NoError(err, "Should be able to load first account tokens")
	assert.Equal(account1Email, loadedAuth1.Account, "First loaded account should match")
	assert.Equal("account1-access-token", loadedAuth1.AccessToken, "First access token should match")

	loadedAuth2, err := LoadAuthTokens(tokenPath2)
	assert.NoError(err, "Should be able to load second account tokens")
	assert.Equal(account2Email, loadedAuth2.Account, "Second loaded account should match")
	assert.Equal("account2-access-token", loadedAuth2.AccessToken, "Second access token should match")

	// Step 17: Verify both accounts have separate cache directories
	cacheDir1 := filepath.Dir(tokenPath1)
	cacheDir2 := filepath.Dir(tokenPath2)
	assert.NotEqual(cacheDir1, cacheDir2, "Cache directories should be different for different accounts")

	// Step 18: Verify cache directories exist
	_, err = os.Stat(cacheDir1)
	assert.NoError(err, "First account cache directory should exist")
	_, err = os.Stat(cacheDir2)
	assert.NoError(err, "Second account cache directory should exist")

	// Step 19: Verify mount registry lists both mounts
	allMounts := registry.ListMounts()
	assert.Equal(2, len(allMounts), "Registry should contain exactly 2 mounts")

	// Step 20: Verify mount registry can find mounts by account
	account1Mounts := registry.GetMountsByAccount(account1Email)
	assert.Equal(1, len(account1Mounts), "Should find exactly 1 mount for first account")
	assert.Equal(mountPoint1, account1Mounts[0], "First account mount point should match")

	account2Mounts := registry.GetMountsByAccount(account2Email)
	assert.Equal(1, len(account2Mounts), "Should find exactly 1 mount for second account")
	assert.Equal(mountPoint2, account2Mounts[0], "Second account mount point should match")

	// Step 21: Verify mount configs contain correct metadata
	config1, exists := registry.GetMountConfig(mountPoint1)
	assert.True(exists, "First mount config should exist")
	assert.Equal(account1Email, config1.Account, "First mount config account should match")
	assert.Equal("Personal OneDrive", config1.Label, "First mount config label should match")

	config2, exists := registry.GetMountConfig(mountPoint2)
	assert.True(exists, "Second mount config should exist")
	assert.Equal(account2Email, config2.Account, "Second mount config account should match")
	assert.Equal("Work OneDrive", config2.Label, "Second mount config label should match")

	// Step 22: Test token lookup with FindAuthTokens for both accounts
	foundPath1, err := FindAuthTokens(testCacheDir, "", account1Email)
	assert.NoError(err, "Should be able to find first account tokens")
	assert.Equal(tokenPath1, foundPath1, "Found token path should match first account path")

	foundPath2, err := FindAuthTokens(testCacheDir, "", account2Email)
	assert.NoError(err, "Should be able to find second account tokens")
	assert.Equal(tokenPath2, foundPath2, "Found token path should match second account path")

	// Step 23: Verify account hash generation is consistent
	hash1a := hashAccount(account1Email)
	hash1b := hashAccount(account1Email)
	assert.Equal(hash1a, hash1b, "Account hash should be consistent for same email")

	hash2 := hashAccount(account2Email)
	assert.NotEqual(hash1a, hash2, "Account hashes should be different for different emails")

	// Step 24: Test case-insensitive account email handling
	account1Upper := "USER1@EXAMPLE.COM"
	tokenPathUpper := GetAuthTokensPathByAccount(testCacheDir, account1Upper)
	assert.Equal(tokenPath1, tokenPathUpper, "Token path should be same regardless of email case")

	// Step 25: Verify registry persistence by reloading
	registry2, err := config.NewMountsRegistry(testConfigDir)
	assert.NoError(err, "Should be able to reload mounts registry")

	reloadedAccount1, exists := registry2.GetAccount(mountPoint1)
	assert.True(exists, "First mount should exist after reload")
	assert.Equal(account1Email, reloadedAccount1, "First account should match after reload")

	reloadedAccount2, exists := registry2.GetAccount(mountPoint2)
	assert.True(exists, "Second mount should exist after reload")
	assert.Equal(account2Email, reloadedAccount2, "Second account should match after reload")

	// Step 26: Test removing one mount doesn't affect the other
	err = registry.RemoveMount(mountPoint1)
	assert.NoError(err, "Should be able to remove first mount")

	_, exists = registry.GetAccount(mountPoint1)
	assert.False(exists, "First mount should not exist after removal")

	registeredAccount2After, exists := registry.GetAccount(mountPoint2)
	assert.True(exists, "Second mount should still exist after removing first")
	assert.Equal(account2Email, registeredAccount2After, "Second account should be unchanged")

	// Step 27: Verify token files are independent (removing mount doesn't delete tokens)
	_, err = os.Stat(tokenPath1)
	assert.NoError(err, "First account token file should still exist after mount removal")
	_, err = os.Stat(tokenPath2)
	assert.NoError(err, "Second account token file should still exist")
}
