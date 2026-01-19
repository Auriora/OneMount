package config

// This package provides a thin wrapper around cmd/common configuration
// to enable property-based testing of XDG compliance without circular dependencies.
//
// The actual configuration implementation lives in cmd/common/config.go.
// This package re-exports the necessary functions for testing purposes.

import (
	"os"
	"path/filepath"

	"github.com/auriora/onemount/internal/logging"
)

// DefaultConfigPath returns the default config location for onemount
// This is a copy of the function from cmd/common/config.go to avoid circular dependencies
func DefaultConfigPath() string {
	confDir, err := os.UserConfigDir()
	if err != nil {
		logging.Error().Err(err).Msg("Could not determine configuration directory.")
	}
	return filepath.Join(confDir, "onemount/config.yml")
}

// createDefaultConfig returns a minimal Config struct with default cache directory
// This is a simplified version for testing XDG compliance
func createDefaultConfig() struct {
	CacheDir string
} {
	xdgCacheDir, _ := os.UserCacheDir()
	return struct {
		CacheDir string
	}{
		CacheDir: filepath.Join(xdgCacheDir, "onemount"),
	}
}
