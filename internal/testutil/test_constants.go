// Package testutil provides utility functions and constants for testing.
package testutil

// TestSandboxDir is the directory used for test files.
const TestSandboxDir = "tmp"

// TestMountPoint is the location where the filesystem is mounted during tests.
const TestMountPoint = TestSandboxDir + "/mount"

// TestDir is the directory within the mount point used for tests.
const TestDir = TestMountPoint + "/onedriver_tests"

// TestDBLoc is the location of the test database.
const TestDBLoc = TestSandboxDir

// DeltaDir is the directory used for delta tests.
const DeltaDir = TestDir + "/delta"

// AuthTokensPath is the path to the authentication tokens file.
const AuthTokensPath = TestSandboxDir + "/.auth_tokens.json"

// TestLogPath is the path to the test log file.
const TestLogPath = TestSandboxDir + "/fusefs_tests.log"
