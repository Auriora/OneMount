// Package testutil provides testing utilities for the OneMount project.
package testutil

import (
	"fmt"
	"os"
	"path/filepath"
)

// EnsureDirectoriesExist ensures that all required test directories exist
// This function is used by tests to set up the necessary directory structure
// without creating import cycles.
func EnsureDirectoriesExist() error {
	// Create the test sandbox directory if it doesn't exist
	if err := os.MkdirAll(TestSandboxDir, 0755); err != nil {
		fmt.Printf("ERROR: Failed to create test sandbox directory: %v\n", err)
		return err
	}

	// Create the temporary directory if it doesn't exist
	if err := os.MkdirAll(TestSandboxTmpDir, 0755); err != nil {
		fmt.Printf("ERROR: Failed to create test sandbox tmp directory: %v\n", err)
		return err
	}

	// Create the logs directory if it doesn't exist
	logsDir := filepath.Dir(TestLogPath)
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		fmt.Printf("ERROR: Failed to create logs directory: %v\n", err)
		return err
	}

	// Create the graph test directory if it doesn't exist
	if err := os.MkdirAll(GraphTestDir, 0755); err != nil {
		fmt.Printf("ERROR: Failed to create graph test directory: %v\n", err)
		return err
	}

	return nil
}
