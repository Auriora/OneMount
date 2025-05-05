// Package testutil provides testing utilities for the OneMount project.
package testutil

import (
	"os"
	"path/filepath"
)

// TestSandboxDir is the base directory for all test artifacts.
// It follows the recommended structure from the test-sandbox-guidelines.md document.
var TestSandboxDir = filepath.Join(os.Getenv("HOME"), ".onemount-tests")

// AuthTokensPath is the path to the authentication tokens file.
// Note: This uses a different file name format (with a leading dot) than the main application.
var AuthTokensPath = filepath.Join(TestSandboxDir, ".auth_tokens.json")

// DmelfaDir is the path to the dmel.fa test file.
var DmelfaDir = filepath.Join(TestSandboxDir, "dmel.fa")

// TestLogPath is the path to the test log file.
var TestLogPath = filepath.Join(TestSandboxDir, "logs", "fusefs_tests.log")

// TestSandboxTmpDir is the path to the temporary directory within the test sandbox.
var TestSandboxTmpDir = filepath.Join(TestSandboxDir, "tmp")

// TestMountPoint is the path to the mount point for tests.
var TestMountPoint = filepath.Join(TestSandboxTmpDir, "mount")

// TestDir is the path to the test directory within the mount point.
var TestDir = filepath.Join(TestMountPoint, "onemount_tests")

// DeltaDir is the path to the delta directory within the test directory.
var DeltaDir = filepath.Join(TestDir, "delta")

// GraphTestDir is the path to the directory for graph API tests.
var GraphTestDir = filepath.Join(TestSandboxDir, "graph_test_dir")

// GetDefaultArtifactsDir returns the default directory for test artifacts.
// This function should be used when initializing the TestConfig.ArtifactsDir field.
func GetDefaultArtifactsDir() string {
	return TestSandboxDir
}
