package graph

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/auriora/onemount/internal/logging"
)

// DeprecationGracePeriod is the duration after migration before old token files
// are eligible for cleanup. This gives users time to verify the migration worked
// before old files are removed.
const DeprecationGracePeriod = 90 * 24 * time.Hour // 90 days

// GetAuthTokensPathByAccount returns the full path to the auth tokens file based on account identity.
//
// Account-Based Token Storage Architecture:
// OneMount uses account-based token storage to ensure tokens are accessible regardless of mount point location.
// This approach provides better reliability, eliminates token duplication, and improves Docker test environment support.
//
// Path Formula: {cacheDir}/accounts/{account-hash}/auth_tokens.json
//
// Where:
//   - cacheDir: XDG cache directory (typically ~/.cache/onemount)
//   - account-hash: First 16 characters of SHA256 hash of normalized account email
//   - AuthTokensFileName: Constant "auth_tokens.json"
//
// Example:
//
//	Account: user@example.com → Hash: a1b2c3d4e5f6g7h8
//	Path: ~/.cache/onemount/accounts/a1b2c3d4e5f6g7h8/auth_tokens.json
//
// Benefits:
//  1. Mount point independence - Same tokens regardless of where you mount
//  2. No token duplication - One account = one token file
//  3. Reliable Docker testing - Tests find tokens regardless of mount point
//  4. Account isolation - Different accounts have separate token files
//  5. Privacy - Email not visible in filesystem (only hash)
//
// Hash Algorithm:
//   - SHA256 for cryptographic security and collision resistance
//   - First 16 characters (64 bits) provides sufficient uniqueness
//   - Email normalized (lowercase, trimmed) before hashing for consistency
func GetAuthTokensPathByAccount(cacheDir, accountEmail string) string {
	if accountEmail == "" {
		logging.Warn().Msg("GetAuthTokensPathByAccount called with empty account email")
		return ""
	}
	accountHash := hashAccount(accountEmail)
	return filepath.Join(cacheDir, "accounts", accountHash, AuthTokensFileName)
}

// hashAccount creates a stable, deterministic hash of an account email address.
//
// The hash is used as a directory name for storing account-specific tokens, providing:
//   - Stability: Same email always produces same hash
//   - Privacy: Email not visible in filesystem
//   - Uniqueness: SHA256 collision resistance
//   - Consistency: Case-insensitive (normalized to lowercase)
//
// Implementation:
//  1. Normalize email: lowercase and trim whitespace
//  2. Compute SHA256 hash
//  3. Return first 16 hex characters (64 bits)
//
// Example:
//
//	"User@Example.com" → "a1b2c3d4e5f6g7h8"
//	"user@example.com" → "a1b2c3d4e5f6g7h8" (same hash)
func hashAccount(email string) string {
	// Normalize email: lowercase and trim whitespace for consistency
	normalized := strings.ToLower(strings.TrimSpace(email))

	// Compute SHA256 hash
	hash := sha256.Sum256([]byte(normalized))

	// Return first 16 hex characters (64 bits)
	// This provides sufficient uniqueness while keeping paths manageable
	return hex.EncodeToString(hash[:])[:16]
}

// FindAuthTokens searches for authentication tokens in multiple locations with automatic migration.
//
// Search Order:
//  1. Account-based location (new): {cacheDir}/accounts/{account-hash}/auth_tokens.json
//  2. Instance-based location (old): {cacheDir}/{instance}/auth_tokens.json
//  3. Legacy location (oldest): {cacheDir}/auth_tokens.json
//
// Migration Strategy:
//   - If tokens found in old location and account email is available, automatically migrate to new location
//   - Old tokens are preserved (not deleted) for safety
//   - Migration is transparent to the user
//
// Parameters:
//   - cacheDir: XDG cache directory
//   - instance: Mount point instance name (for backward compatibility)
//   - accountEmail: Account email address (for account-based storage)
//
// Returns:
//   - Token file path (either existing or where new tokens should be created)
//   - Error if search fails
//
// Example:
//
//	path, err := FindAuthTokens(cacheDir, instance, "user@example.com")
//	// Returns: ~/.cache/onemount/accounts/a1b2c3d4e5f6g7h8/auth_tokens.json
func FindAuthTokens(cacheDir, instance, accountEmail string) (string, error) {
	// 1. Try account-based location (new, preferred)
	if accountEmail != "" {
		accountPath := GetAuthTokensPathByAccount(cacheDir, accountEmail)
		if accountPath != "" {
			if _, err := os.Stat(accountPath); err == nil {
				logging.Debug().
					Str("path", accountPath).
					Str("location", "account-based").
					Msg("Found auth tokens in account-based location")
				return accountPath, nil
			}
		}
	}

	// 2. Try instance-based location (old, for migration)
	if instance != "" {
		instancePath := GetAuthTokensPath(cacheDir, instance)
		if _, err := os.Stat(instancePath); err == nil {
			logging.Info().
				Str("path", instancePath).
				Str("location", "instance-based").
				Msg("Found auth tokens in instance-based location")

			// Auto-migrate if we have account email
			if accountEmail != "" {
				accountPath := GetAuthTokensPathByAccount(cacheDir, accountEmail)
				if accountPath != "" {
					if err := migrateTokens(instancePath, accountPath); err == nil {
						logging.Info().
							Str("from", instancePath).
							Str("to", accountPath).
							Msg("Successfully migrated auth tokens to account-based location")
						return accountPath, nil
					} else {
						logging.Warn().
							Err(err).
							Str("from", instancePath).
							Str("to", accountPath).
							Msg("Failed to migrate auth tokens, using instance-based location")
					}
				}
			}
			return instancePath, nil
		}
	}

	// 3. Try legacy location (oldest, for backward compatibility)
	legacyPath := GetAuthTokensPathFromCacheDir(cacheDir)
	if _, err := os.Stat(legacyPath); err == nil {
		logging.Info().
			Str("path", legacyPath).
			Str("location", "legacy").
			Msg("Found auth tokens in legacy location")

		// Auto-migrate if we have account email
		if accountEmail != "" {
			accountPath := GetAuthTokensPathByAccount(cacheDir, accountEmail)
			if accountPath != "" {
				if err := migrateTokens(legacyPath, accountPath); err == nil {
					logging.Info().
						Str("from", legacyPath).
						Str("to", accountPath).
						Msg("Successfully migrated auth tokens from legacy location to account-based location")
					return accountPath, nil
				} else {
					logging.Warn().
						Err(err).
						Str("from", legacyPath).
						Str("to", accountPath).
						Msg("Failed to migrate auth tokens from legacy location, using legacy location")
				}
			}
		}
		return legacyPath, nil
	}

	// 4. No existing tokens found - return new account-based path for creation
	if accountEmail != "" {
		accountPath := GetAuthTokensPathByAccount(cacheDir, accountEmail)
		if accountPath != "" {
			logging.Debug().
				Str("path", accountPath).
				Msg("No existing tokens found, will create in account-based location")
			return accountPath, nil
		}
	}

	// 5. Fallback to legacy location if no account email available
	logging.Debug().
		Str("path", legacyPath).
		Msg("No account email available, falling back to legacy location")
	return legacyPath, nil
}

// migrateTokens copies authentication tokens from an old location to a new location.
//
// Migration Process:
//  1. Create directory for new location (with secure permissions 0700)
//  2. Read tokens from old location
//  3. Write tokens to new location (with secure permissions 0600)
//  4. Preserve old tokens (not deleted) for safety
//
// Security:
//   - Directory permissions: 0700 (owner read/write/execute only)
//   - File permissions: 0600 (owner read/write only)
//   - Old tokens preserved as backup
//
// Parameters:
//   - oldPath: Current token file path
//   - newPath: New token file path
//
// Returns:
//   - Error if migration fails
//
// Example:
//
//	err := migrateTokens(
//	    "~/.cache/onemount/home-user-OneDrive/auth_tokens.json",
//	    "~/.cache/onemount/accounts/a1b2c3d4e5f6g7h8/auth_tokens.json",
//	)
func migrateTokens(oldPath, newPath string) error {
	// Validate paths
	if oldPath == "" || newPath == "" {
		return fmt.Errorf("invalid migration paths: oldPath=%q, newPath=%q", oldPath, newPath)
	}

	if oldPath == newPath {
		logging.Debug().Str("path", oldPath).Msg("Migration not needed, paths are identical")
		return nil
	}

	// Check if new location already exists
	if _, err := os.Stat(newPath); err == nil {
		logging.Debug().Str("path", newPath).Msg("Migration not needed, new location already exists")
		return nil
	}

	// Create directory for new location with secure permissions
	newDir := filepath.Dir(newPath)
	if err := os.MkdirAll(newDir, 0700); err != nil {
		return fmt.Errorf("failed to create directory for new token location: %w", err)
	}

	// Read tokens from old location
	data, err := os.ReadFile(oldPath)
	if err != nil {
		return fmt.Errorf("failed to read tokens from old location: %w", err)
	}

	// Write tokens to new location with secure permissions
	if err := os.WriteFile(newPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write tokens to new location: %w", err)
	}

	// Verify the migration succeeded
	if _, err := os.Stat(newPath); err != nil {
		return fmt.Errorf("migration verification failed, new file not found: %w", err)
	}

	logging.Info().
		Str("from", oldPath).
		Str("to", newPath).
		Msg("Auth token migration completed successfully")

	// Log deprecation warning for the old file location
	logging.Warn().
		Str("path", oldPath).
		Str("migratedTo", newPath).
		Msg("Old token file is deprecated and will be eligible for cleanup after the grace period. " +
			"Please verify the new location works correctly.")

	return nil
}

// CleanupOldTokens removes old token files that have been successfully migrated
// to account-based storage and are past the deprecation grace period.
//
// A file is eligible for cleanup when:
//  1. The corresponding account-based token file exists and is valid
//  2. The old file has not been modified since migration (mtime older than grace period)
//
// Parameters:
//   - cacheDir: XDG cache directory
//   - instance: Mount point instance name
//   - accountEmail: Account email address
//
// Returns:
//   - removed: list of paths that were cleaned up
//   - error: any error encountered (non-fatal; partial cleanup may have occurred)
func CleanupOldTokens(cacheDir, instance, accountEmail string) ([]string, error) {
	if accountEmail == "" {
		return nil, nil
	}

	accountPath := GetAuthTokensPathByAccount(cacheDir, accountEmail)
	if accountPath == "" {
		return nil, nil
	}

	// Account-based file must exist and be valid before we remove anything
	if !fileExists(accountPath) {
		return nil, nil
	}

	var removed []string
	cutoff := time.Now().Add(-DeprecationGracePeriod)

	// Check instance-based location
	if instance != "" {
		instancePath := GetAuthTokensPath(cacheDir, instance)
		if r, err := cleanupIfEligible(instancePath, cutoff); err != nil {
			logging.Warn().Err(err).Str("path", instancePath).Msg("Failed to clean up old instance-based token file")
		} else if r {
			removed = append(removed, instancePath)
		}
	}

	// Check legacy location
	legacyPath := GetAuthTokensPathFromCacheDir(cacheDir)
	if r, err := cleanupIfEligible(legacyPath, cutoff); err != nil {
		logging.Warn().Err(err).Str("path", legacyPath).Msg("Failed to clean up old legacy token file")
	} else if r {
		removed = append(removed, legacyPath)
	}

	if len(removed) > 0 {
		logging.Info().
			Strs("removed", removed).
			Msg("Cleaned up deprecated token files")
	}

	return removed, nil
}

// cleanupIfEligible removes a file if it exists and its modification time is before the cutoff.
// Returns true if the file was removed, false otherwise.
func cleanupIfEligible(path string, cutoff time.Time) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		return false, nil // file doesn't exist, nothing to do
	}
	if info.IsDir() {
		return false, nil
	}
	if info.ModTime().After(cutoff) {
		logging.Debug().
			Str("path", path).
			Time("modTime", info.ModTime()).
			Time("cutoff", cutoff).
			Msg("Old token file not yet eligible for cleanup (within grace period)")
		return false, nil
	}

	if err := os.Remove(path); err != nil {
		return false, fmt.Errorf("failed to remove deprecated token file %q: %w", path, err)
	}
	return true, nil
}

// fileExists checks if a file exists and is not a directory.
func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}
