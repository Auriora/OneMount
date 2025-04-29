// Package testutil provides utility functions and constants for testing.
package testutil

// TestMountPoint is the location where the filesystem is mounted during tests.
const TestMountPoint = "tmp/mount"

// TestDir is the directory within the mount point used for tests.
const TestDir = TestMountPoint + "/onedriver_tests"

// TestDBLoc is the location of the test database.
const TestDBLoc = "tmp"

// DeltaDir is the directory used for delta tests.
const DeltaDir = TestDir + "/delta"
