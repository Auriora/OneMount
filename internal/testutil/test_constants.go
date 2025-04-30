// Package testutil provides utility functions and constants for testing.
package testutil

// All paths in this file are relative to the project root directory.
// Tests should ensure they are running from the project root before using these paths.

// TestSandboxDir is the directory used for test files.
const TestSandboxDir = "test-sandbox"

// TestSandboxTmpDir is the directory used for temporary files. These files should be deleted after testing is complete.
const TestSandboxTmpDir = TestSandboxDir + "/tmp"

// TestMountPoint is the location where the filesystem is mounted during tests.
const TestMountPoint = TestSandboxTmpDir + "/mount"

// TestDir is the directory within the mount point used for tests.
const TestDir = TestMountPoint + "/onemount_tests"

// TestDBLoc is the location of the test database.
const TestDBLoc = TestSandboxTmpDir

// DeltaDir is the directory used for delta tests.
const DeltaDir = TestDir + "/delta"

// DmelfaDir is the path to the dmel.fa file used for tests.
const DmelfaDir = TestSandboxDir + "/dmel.fa"

// AuthTokensPath is the path to the authentication tokens file.
const AuthTokensPath = TestSandboxDir + "/.auth_tokens.json"

// TestLogPath is the path to the test log file.
const TestLogPath = TestSandboxDir + "/fusefs_tests.log"
