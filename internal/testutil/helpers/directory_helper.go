// Package helpers provides testing utilities for the OneMount project.
package helpers

import (
	"github.com/auriora/onemount/pkg/logging"
	"github.com/auriora/onemount/internal/testutil"
	"os"
	"path/filepath"
)

// EnsureTestDirectories ensures that all required test directories exist
func EnsureTestDirectories() error {
	// Create the test sandbox directory if it doesn't exist
	if err := os.MkdirAll(testutil.TestSandboxDir, 0755); err != nil {
		logging.Error().Err(err).Str("path", testutil.TestSandboxDir).Msg("Failed to create test sandbox directory")
		return err
	}

	// Create the temporary directory if it doesn't exist
	if err := os.MkdirAll(testutil.TestSandboxTmpDir, 0755); err != nil {
		logging.Error().Err(err).Str("path", testutil.TestSandboxTmpDir).Msg("Failed to create test sandbox tmp directory")
		return err
	}

	// Create the logs directory if it doesn't exist
	logsDir := filepath.Dir(testutil.TestLogPath)
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		logging.Error().Err(err).Str("path", logsDir).Msg("Failed to create logs directory")
		return err
	}

	// Create the graph test directory if it doesn't exist
	if err := os.MkdirAll(testutil.GraphTestDir, 0755); err != nil {
		logging.Error().Err(err).Str("path", testutil.GraphTestDir).Msg("Failed to create graph test directory")
		return err
	}

	return nil
}
