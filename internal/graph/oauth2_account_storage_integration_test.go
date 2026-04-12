package graph

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/auriora/onemount/internal/testutil/framework"
)

// TestIT_AUTH_09_01_AuthTokenMigration_InstanceToAccount_MigratesCorrectly tests that tokens
// stored in the old instance-based location are automatically migrated to account-based storage
// when FindAuthTokens is called with a known account email.
//
//	Test Case ID    IT-AUTH-09-01
//	Title           Auth Token Migration from Instance-Based to Account-Based Storage
//	Description     Verifies automatic migration of tokens from old instance-based paths
//	                to the new account-based path, preserving old files and creating correct
//	                directory structure with secure permissions.
//	Preconditions   Tokens exist in instance-based location
//	Steps           1. Create tokens in instance-based location
//	                2. Call FindAuthTokens with account email
//	                3. Verify tokens migrated to account-based location
//	                4. Verify old tokens still exist (preserved)
//	                5. Verify new file permissions are 0600
//	                6. Verify new directory permissions are 0700
//	                7. Verify token content is identical
//	Expected Result Tokens are migrated correctly with proper permissions
//	Requirements    1.2, 1.6, 13.2
func TestIT_AUTH_09_01_AuthTokenMigration_InstanceToAccount_MigratesCorrectly(t *testing.T) {
	assert := framework.NewAssert(t)
	tmpDir := t.TempDir()
	cacheDir := filepath.Join(tmpDir, "cache")

	accountEmail := "migrate-test@example.com"
	instance := "home-user-OneDrive"

	// Step 1: Create tokens in instance-based location
	instancePath := GetAuthTokensPath(cacheDir, instance)
	err := os.MkdirAll(filepath.Dir(instancePath), 0700)
	assert.NoError(err, "Should create instance directory")

	originalAuth := &Auth{
		Account:      accountEmail,
		AccessToken:  "instance-access-token",
		RefreshToken: "instance-refresh-token",
		ExpiresAt:    time.Now().Add(1 * time.Hour).Unix(),
	}
	err = originalAuth.ToFile(instancePath)
	assert.NoError(err, "Should save tokens to instance-based location")

	// Step 2: Call FindAuthTokens — should trigger migration
	resultPath, err := FindAuthTokens(cacheDir, instance, accountEmail)
	assert.NoError(err, "FindAuthTokens should succeed")

	// Step 3: Verify result points to account-based location
	expectedAccountPath := GetAuthTokensPathByAccount(cacheDir, accountEmail)
	assert.Equal(expectedAccountPath, resultPath, "Should return account-based path after migration")

	// Step 4: Verify old tokens still exist (preserved for safety)
	_, err = os.Stat(instancePath)
	assert.NoError(err, "Old instance-based token file should still exist")

	// Step 5: Verify new file exists with correct permissions
	newInfo, err := os.Stat(resultPath)
	assert.NoError(err, "Account-based token file should exist")
	assert.Equal(os.FileMode(0600), newInfo.Mode().Perm(), "Token file should have 0600 permissions")

	// Step 6: Verify new directory permissions
	dirInfo, err := os.Stat(filepath.Dir(resultPath))
	assert.NoError(err, "Account-based directory should exist")
	assert.Equal(os.FileMode(0700), dirInfo.Mode().Perm(), "Token directory should have 0700 permissions")

	// Step 7: Verify token content is identical
	migratedAuth := &Auth{}
	err = migratedAuth.FromFile(resultPath)
	assert.NoError(err, "Should load migrated tokens")
	assert.Equal(originalAuth.Account, migratedAuth.Account, "Account should match")
	assert.Equal(originalAuth.AccessToken, migratedAuth.AccessToken, "Access token should match")
	assert.Equal(originalAuth.RefreshToken, migratedAuth.RefreshToken, "Refresh token should match")
}

// TestIT_AUTH_09_02_AuthTokenMigration_LegacyToAccount_MigratesCorrectly tests migration
// from the oldest legacy location to account-based storage.
//
//	Test Case ID    IT-AUTH-09-02
//	Title           Auth Token Migration from Legacy to Account-Based Storage
//	Description     Verifies automatic migration from the legacy (cacheDir-root) token path
//	Preconditions   Tokens exist only in legacy location
//	Steps           1. Create tokens in legacy location
//	                2. Call FindAuthTokens with account email
//	                3. Verify tokens migrated to account-based location
//	                4. Verify old tokens preserved
//	Expected Result Legacy tokens are migrated correctly
//	Requirements    1.2, 1.6
func TestIT_AUTH_09_02_AuthTokenMigration_LegacyToAccount_MigratesCorrectly(t *testing.T) {
	assert := framework.NewAssert(t)
	tmpDir := t.TempDir()
	cacheDir := filepath.Join(tmpDir, "cache")

	accountEmail := "legacy-test@example.com"
	instance := "some-instance"

	// Step 1: Create tokens in legacy location only
	legacyPath := GetAuthTokensPathFromCacheDir(cacheDir)
	err := os.MkdirAll(filepath.Dir(legacyPath), 0700)
	assert.NoError(err, "Should create legacy directory")

	originalAuth := &Auth{
		Account:      accountEmail,
		AccessToken:  "legacy-access-token",
		RefreshToken: "legacy-refresh-token",
		ExpiresAt:    time.Now().Add(1 * time.Hour).Unix(),
	}
	err = originalAuth.ToFile(legacyPath)
	assert.NoError(err, "Should save tokens to legacy location")

	// Step 2: Call FindAuthTokens
	resultPath, err := FindAuthTokens(cacheDir, instance, accountEmail)
	assert.NoError(err, "FindAuthTokens should succeed")

	// Step 3: Verify result points to account-based location
	expectedAccountPath := GetAuthTokensPathByAccount(cacheDir, accountEmail)
	assert.Equal(expectedAccountPath, resultPath, "Should return account-based path after migration")

	// Step 4: Verify old tokens preserved
	_, err = os.Stat(legacyPath)
	assert.NoError(err, "Legacy token file should still exist")

	// Step 5: Verify migrated content
	migratedAuth := &Auth{}
	err = migratedAuth.FromFile(resultPath)
	assert.NoError(err, "Should load migrated tokens")
	assert.Equal(accountEmail, migratedAuth.Account, "Account should match")
	assert.Equal("legacy-access-token", migratedAuth.AccessToken, "Access token should match")
}

// TestIT_AUTH_09_03_MultipleAccountTokens_SeparateStorage_NoConflicts tests that multiple
// accounts maintain completely separate token storage with no cross-contamination.
//
//	Test Case ID    IT-AUTH-09-03
//	Title           Multiple Account Token Isolation
//	Description     Verifies that tokens for different accounts are stored in separate
//	                directories and that operations on one account don't affect another.
//	Preconditions   None
//	Steps           1. Create tokens for account A
//	                2. Create tokens for account B
//	                3. Verify separate storage paths
//	                4. Verify FindAuthTokens returns correct path per account
//	                5. Modify account A tokens
//	                6. Verify account B tokens unchanged
//	                7. Verify case-insensitive email produces same path
//	Expected Result Multiple accounts are fully isolated
//	Requirements    1.6, 4.1-4.8
func TestIT_AUTH_09_03_MultipleAccountTokens_SeparateStorage_NoConflicts(t *testing.T) {
	assert := framework.NewAssert(t)
	tmpDir := t.TempDir()
	cacheDir := filepath.Join(tmpDir, "cache")

	accountA := "personal@example.com"
	accountB := "work@company.com"

	// Step 1 & 2: Create tokens for both accounts
	pathA := GetAuthTokensPathByAccount(cacheDir, accountA)
	pathB := GetAuthTokensPathByAccount(cacheDir, accountB)

	authA := &Auth{Account: accountA, AccessToken: "token-A", RefreshToken: "refresh-A", ExpiresAt: time.Now().Add(time.Hour).Unix()}
	authB := &Auth{Account: accountB, AccessToken: "token-B", RefreshToken: "refresh-B", ExpiresAt: time.Now().Add(time.Hour).Unix()}

	err := SaveAuthTokens(authA, pathA)
	assert.NoError(err, "Should save account A tokens")
	err = SaveAuthTokens(authB, pathB)
	assert.NoError(err, "Should save account B tokens")

	// Step 3: Verify separate storage paths
	assert.NotEqual(pathA, pathB, "Token paths must be different for different accounts")

	// Step 4: Verify FindAuthTokens returns correct path per account
	foundA, err := FindAuthTokens(cacheDir, "", accountA)
	assert.NoError(err, "Should find account A tokens")
	assert.Equal(pathA, foundA, "Should return correct path for account A")

	foundB, err := FindAuthTokens(cacheDir, "", accountB)
	assert.NoError(err, "Should find account B tokens")
	assert.Equal(pathB, foundB, "Should return correct path for account B")

	// Step 5: Modify account A tokens
	authA.AccessToken = "token-A-modified"
	err = SaveAuthTokens(authA, pathA)
	assert.NoError(err, "Should update account A tokens")

	// Step 6: Verify account B tokens unchanged
	loadedB, err := LoadAuthTokens(pathB)
	assert.NoError(err, "Should load account B tokens")
	assert.Equal("token-B", loadedB.AccessToken, "Account B access token should be unchanged")
	assert.Equal(accountB, loadedB.Account, "Account B email should be unchanged")

	// Step 7: Verify case-insensitive email produces same path
	pathAUpper := GetAuthTokensPathByAccount(cacheDir, "PERSONAL@EXAMPLE.COM")
	assert.Equal(pathA, pathAUpper, "Case-insensitive email should produce same path")
}

// TestIT_AUTH_09_04_DockerTokenAccess_DifferentMountPoints_FindsTokens tests that tokens
// stored via account-based storage are accessible regardless of mount point, simulating
// the Docker test environment scenario where mount points differ from production.
//
//	Test Case ID    IT-AUTH-09-04
//	Title           Docker Token Access with Different Mount Points
//	Description     Simulates the Docker test scenario where tokens are stored from one
//	                mount point but need to be accessed from a different mount point.
//	Preconditions   Tokens stored via one mount point
//	Steps           1. Store tokens using mount point A (simulating production)
//	                2. Look up tokens using mount point B (simulating Docker)
//	                3. Verify tokens found via account email regardless of mount point
//	                4. Verify tokens found via FindAuthTokens with different instance
//	Expected Result Tokens accessible regardless of mount point
//	Requirements    1.6, 13.2, 13.4
func TestIT_AUTH_09_04_DockerTokenAccess_DifferentMountPoints_FindsTokens(t *testing.T) {
	assert := framework.NewAssert(t)
	tmpDir := t.TempDir()
	cacheDir := filepath.Join(tmpDir, "cache")

	accountEmail := "docker-test@example.com"
	productionInstance := "home-user-OneDrive"
	dockerInstance := "workspace-mount"

	// Step 1: Store tokens using production mount point (account-based)
	accountPath := GetAuthTokensPathByAccount(cacheDir, accountEmail)
	auth := &Auth{
		Account:      accountEmail,
		AccessToken:  "production-token",
		RefreshToken: "production-refresh",
		ExpiresAt:    time.Now().Add(time.Hour).Unix(),
	}
	err := SaveAuthTokens(auth, accountPath)
	assert.NoError(err, "Should save tokens via production mount point")

	// Step 2: Look up tokens using Docker mount point (different instance)
	// The key insight: account-based storage means the instance doesn't matter
	// as long as we know the account email
	foundPath, err := FindAuthTokens(cacheDir, dockerInstance, accountEmail)
	assert.NoError(err, "Should find tokens from Docker mount point")
	assert.Equal(accountPath, foundPath, "Should find same tokens regardless of mount point")

	// Step 3: Verify token content is accessible
	loadedAuth, err := LoadAuthTokens(foundPath)
	assert.NoError(err, "Should load tokens found via Docker mount point")
	assert.Equal(accountEmail, loadedAuth.Account, "Account should match")
	assert.Equal("production-token", loadedAuth.AccessToken, "Access token should match")

	// Step 4: Verify with yet another instance name (simulating CI)
	ciInstance := "ci-test-mount"
	foundPathCI, err := FindAuthTokens(cacheDir, ciInstance, accountEmail)
	assert.NoError(err, "Should find tokens from CI mount point")
	assert.Equal(accountPath, foundPathCI, "Should find same tokens from CI environment")

	// Step 5: Verify that without account email, different instances don't find each other's tokens
	// (instance-based isolation still works for backward compat)
	instancePathProd := GetAuthTokensPath(cacheDir, productionInstance)
	err = os.MkdirAll(filepath.Dir(instancePathProd), 0700)
	assert.NoError(err, "Should create production instance directory")
	err = os.WriteFile(instancePathProd, []byte(`{"access_token":"instance-only"}`), 0600)
	assert.NoError(err, "Should write instance-based token")

	// Without account email, Docker instance should NOT find production instance tokens
	foundNoEmail, err := FindAuthTokens(cacheDir, dockerInstance, "")
	assert.NoError(err, "Should not error without account email")
	// Should fall back to legacy path (not find production instance tokens)
	assert.NotEqual(instancePathProd, foundNoEmail,
		"Without account email, different instances should not find each other's tokens")
}

// TestIT_AUTH_09_05_CleanupOldTokens_AfterGracePeriod_RemovesDeprecated tests that
// the cleanup mechanism correctly removes old token files after the grace period.
//
//	Test Case ID    IT-AUTH-09-05
//	Title           Old Token Cleanup After Grace Period
//	Description     Verifies that CleanupOldTokens removes deprecated token files
//	                only after the grace period has elapsed, and only when the
//	                account-based file exists.
//	Preconditions   Tokens exist in both old and new locations
//	Steps           1. Create tokens in account-based location
//	                2. Create old tokens with recent mtime (within grace period)
//	                3. Call CleanupOldTokens — should NOT remove
//	                4. Backdate old token mtime past grace period
//	                5. Call CleanupOldTokens — should remove
//	                6. Verify account-based tokens untouched
//	Expected Result Old tokens removed only after grace period
//	Requirements    1.6
func TestIT_AUTH_09_05_CleanupOldTokens_AfterGracePeriod_RemovesDeprecated(t *testing.T) {
	assert := framework.NewAssert(t)
	tmpDir := t.TempDir()
	cacheDir := filepath.Join(tmpDir, "cache")

	accountEmail := "cleanup-test@example.com"
	instance := "test-instance"

	// Step 1: Create tokens in account-based location
	accountPath := GetAuthTokensPathByAccount(cacheDir, accountEmail)
	err := os.MkdirAll(filepath.Dir(accountPath), 0700)
	assert.NoError(err, "Should create account directory")
	err = os.WriteFile(accountPath, []byte(`{"account":"cleanup-test@example.com","access_token":"new"}`), 0600)
	assert.NoError(err, "Should write account-based tokens")

	// Step 2: Create old tokens with recent mtime
	instancePath := GetAuthTokensPath(cacheDir, instance)
	err = os.MkdirAll(filepath.Dir(instancePath), 0700)
	assert.NoError(err, "Should create instance directory")
	err = os.WriteFile(instancePath, []byte(`{"access_token":"old-instance"}`), 0600)
	assert.NoError(err, "Should write instance-based tokens")

	legacyPath := GetAuthTokensPathFromCacheDir(cacheDir)
	// legacyPath dir should already exist from account path creation, but ensure it
	err = os.MkdirAll(filepath.Dir(legacyPath), 0700)
	assert.NoError(err, "Should create legacy directory")
	err = os.WriteFile(legacyPath, []byte(`{"access_token":"old-legacy"}`), 0600)
	assert.NoError(err, "Should write legacy tokens")

	// Step 3: Call CleanupOldTokens — files are recent, should NOT remove
	removed, err := CleanupOldTokens(cacheDir, instance, accountEmail)
	assert.NoError(err, "CleanupOldTokens should not error")
	assert.Equal(0, len(removed), "Should not remove recent files")

	// Verify old files still exist
	_, err = os.Stat(instancePath)
	assert.NoError(err, "Instance-based file should still exist")
	_, err = os.Stat(legacyPath)
	assert.NoError(err, "Legacy file should still exist")

	// Step 4: Backdate old token files past grace period
	oldTime := time.Now().Add(-(DeprecationGracePeriod + 24*time.Hour))
	err = os.Chtimes(instancePath, oldTime, oldTime)
	assert.NoError(err, "Should backdate instance-based file")
	err = os.Chtimes(legacyPath, oldTime, oldTime)
	assert.NoError(err, "Should backdate legacy file")

	// Step 5: Call CleanupOldTokens — should remove both
	removed, err = CleanupOldTokens(cacheDir, instance, accountEmail)
	assert.NoError(err, "CleanupOldTokens should not error")
	assert.Equal(2, len(removed), "Should remove both deprecated files")

	// Verify old files are gone
	_, err = os.Stat(instancePath)
	assert.True(os.IsNotExist(err), "Instance-based file should be removed")
	_, err = os.Stat(legacyPath)
	assert.True(os.IsNotExist(err), "Legacy file should be removed")

	// Step 6: Verify account-based tokens untouched
	_, err = os.Stat(accountPath)
	assert.NoError(err, "Account-based token file should still exist")
	data, err := os.ReadFile(accountPath)
	assert.NoError(err, "Should read account-based tokens")
	assert.True(len(data) > 0, "Account-based token file should not be empty")
}

// TestIT_AUTH_09_06_CleanupOldTokens_NoAccountFile_SkipsCleanup tests that cleanup
// does not remove old files when the account-based file doesn't exist yet.
//
//	Test Case ID    IT-AUTH-09-06
//	Title           Old Token Cleanup Safety — No Account File
//	Description     Verifies that CleanupOldTokens does NOT remove old files when
//	                the account-based token file doesn't exist (safety guard).
//	Preconditions   Old tokens exist, account-based tokens do NOT exist
//	Steps           1. Create old tokens only (no account-based file)
//	                2. Backdate old tokens past grace period
//	                3. Call CleanupOldTokens
//	                4. Verify old tokens NOT removed
//	Expected Result Old tokens preserved when account-based file missing
//	Requirements    1.6
func TestIT_AUTH_09_06_CleanupOldTokens_NoAccountFile_SkipsCleanup(t *testing.T) {
	assert := framework.NewAssert(t)
	tmpDir := t.TempDir()
	cacheDir := filepath.Join(tmpDir, "cache")

	accountEmail := "safety-test@example.com"
	instance := "test-instance"

	// Step 1: Create old tokens only
	instancePath := GetAuthTokensPath(cacheDir, instance)
	err := os.MkdirAll(filepath.Dir(instancePath), 0700)
	assert.NoError(err, "Should create instance directory")
	err = os.WriteFile(instancePath, []byte(`{"access_token":"old"}`), 0600)
	assert.NoError(err, "Should write instance-based tokens")

	// Step 2: Backdate past grace period
	oldTime := time.Now().Add(-(DeprecationGracePeriod + 24*time.Hour))
	err = os.Chtimes(instancePath, oldTime, oldTime)
	assert.NoError(err, "Should backdate file")

	// Step 3: Call CleanupOldTokens — account-based file doesn't exist
	removed, err := CleanupOldTokens(cacheDir, instance, accountEmail)
	assert.NoError(err, "CleanupOldTokens should not error")
	assert.Equal(0, len(removed), "Should not remove anything when account-based file missing")

	// Step 4: Verify old tokens preserved
	_, err = os.Stat(instancePath)
	assert.NoError(err, "Instance-based file should still exist (safety)")
}
