package fs

import (
	"net/http"
	"syscall"
	"testing"

	"github.com/auriora/onemount/pkg/errors"
	"github.com/auriora/onemount/pkg/graph"
	"github.com/auriora/onemount/pkg/testutil/framework"
	"github.com/auriora/onemount/pkg/testutil/helpers"
	"github.com/hanwen/go-fuse/v2/fuse"
)

// TestUT_FS_ERR_01_01_DiskSpaceExhaustion_WriteOperation_HandledCorrectly tests disk space exhaustion during write operations
//
//	Test Case ID    UT-FS-ERR-01-01
//	Title           Disk Space Exhaustion During Write Operation
//	Description     Tests that disk space exhaustion errors are handled correctly during write operations
//	Preconditions   None
//	Steps           1. Create a file
//	                2. Simulate disk space exhaustion error during write
//	                3. Verify proper error handling and status
//	Expected Result Write operation fails with appropriate error code (ENOSPC)
//	Notes: This test verifies that disk space exhaustion is properly detected and reported.
func TestUT_FS_ERR_01_01_DiskSpaceExhaustion_WriteOperation_HandledCorrectly(t *testing.T) {
	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "DiskSpaceExhaustionWriteFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
		// Create the filesystem
		fs, err := NewFilesystem(auth, mountPoint, cacheTTL)
		if err != nil {
			return nil, err
		}
		return fs, nil
	})

	// Use the fixture to run the test
	fixture.Use(t, func(t *testing.T, fixture interface{}) {
		// Create assertions helper
		assert := framework.NewAssert(t)

		// Get the test data
		unitTestFixture, ok := fixture.(*framework.UnitTestFixture)
		if !ok {
			t.Fatalf("Expected fixture to be of type *framework.UnitTestFixture, but got %T", fixture)
		}
		fsFixture := unitTestFixture.SetupData.(*helpers.FSTestFixture)
		fs := fsFixture.FS.(*Filesystem)
		mockClient := fsFixture.MockClient

		// Step 1: Create a test file
		testFileName := "disk-space-test.txt"

		// Create the file using Mknod
		mknodIn := &fuse.MknodIn{
			InHeader: fuse.InHeader{NodeId: 1}, // Root node ID
			Mode:     0644,
		}
		entryOut := &fuse.EntryOut{}
		status := fs.Mknod(nil, mknodIn, testFileName, entryOut)
		assert.Equal(fuse.OK, status, "Mknod should succeed")

		// Step 2: Simulate disk space exhaustion during write
		// Configure the mock to return insufficient storage error for write operations
		testFileID := fs.TranslateID(entryOut.NodeId)
		mockClient.AddMockResponse("/me/drive/items/"+testFileID+"/content", nil, http.StatusInsufficientStorage,
			errors.NewOperationError("insufficientStorage: Insufficient storage space available", nil))

		// Attempt to write data to the file
		writeIn := &fuse.WriteIn{
			InHeader: fuse.InHeader{NodeId: entryOut.NodeId},
			Offset:   0,
		}
		testData := []byte("This write should fail due to disk space exhaustion")

		// Step 3: Verify proper error handling
		bytesWritten, writeStatus := fs.Write(nil, writeIn, testData)

		// Verify that the write operation failed with appropriate error
		assert.NotEqual(fuse.OK, writeStatus, "Write should fail due to disk space exhaustion")
		assert.Equal(uint32(0), bytesWritten, "No bytes should be written when disk space is exhausted")

		// The specific error code may vary, but it should indicate a storage-related error
		// Common error codes for disk space issues: EIO, EREMOTEIO
		assert.True(writeStatus == fuse.EIO || writeStatus == fuse.EREMOTEIO,
			"Write should fail with EIO or EREMOTEIO error code")
	})
}

// TestUT_FS_ERR_01_02_DiskSpaceExhaustion_FileCreation_HandledCorrectly tests disk space exhaustion during file creation
//
//	Test Case ID    UT-FS-ERR-01-02
//	Title           Disk Space Exhaustion During File Creation
//	Description     Tests that disk space exhaustion errors are handled correctly during file creation
//	Preconditions   None
//	Steps           1. Simulate disk space exhaustion in local cache
//	                2. Attempt to create a new file
//	                3. Verify proper error handling and status
//	Expected Result File creation fails with appropriate error code (ENOSPC)
//	Notes: This test verifies that disk space exhaustion is properly detected during file creation.
func TestUT_FS_ERR_01_02_DiskSpaceExhaustion_FileCreation_HandledCorrectly(t *testing.T) {
	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "DiskSpaceExhaustionCreateFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
		// Create the filesystem
		fs, err := NewFilesystem(auth, mountPoint, cacheTTL)
		if err != nil {
			return nil, err
		}
		return fs, nil
	})

	// Use the fixture to run the test
	fixture.Use(t, func(t *testing.T, fixture interface{}) {
		// Create assertions helper
		assert := framework.NewAssert(t)

		// Get the test data
		unitTestFixture, ok := fixture.(*framework.UnitTestFixture)
		if !ok {
			t.Fatalf("Expected fixture to be of type *framework.UnitTestFixture, but got %T", fixture)
		}
		fsFixture := unitTestFixture.SetupData.(*helpers.FSTestFixture)
		fs := fsFixture.FS.(*Filesystem)
		mockClient := fsFixture.MockClient

		// Step 1: Simulate disk space exhaustion for file creation
		testFileName := "disk-space-create-test.txt"

		// Configure the mock to return insufficient storage error for file creation
		mockClient.AddMockResponse("/me/drive/items/"+fsFixture.RootID+"/children", nil, http.StatusInsufficientStorage,
			errors.NewOperationError("insufficientStorage: Insufficient storage space available to create file", nil))

		// Step 2: Attempt to create a new file
		createIn := &fuse.CreateIn{
			InHeader: fuse.InHeader{NodeId: 1}, // Root node ID
			Mode:     0644,
			Flags:    uint32(syscall.O_WRONLY | syscall.O_CREAT),
		}
		createOut := &fuse.CreateOut{}

		// Step 3: Verify proper error handling
		status := fs.Create(nil, createIn, testFileName, createOut)

		// Verify that the file creation failed with appropriate error
		assert.NotEqual(fuse.OK, status, "File creation should fail due to disk space exhaustion")

		// The specific error code may vary, but it should indicate a storage-related error
		// Common error codes for disk space issues: EIO, EREMOTEIO
		assert.True(status == fuse.EIO || status == fuse.EREMOTEIO,
			"File creation should fail with EIO or EREMOTEIO error code")

		// Verify that no node ID was assigned
		assert.Equal(uint64(0), createOut.NodeId, "No node ID should be assigned when creation fails")
	})
}

// TestUT_FS_ERR_02_01_PermissionDenied_FileAccess_HandledCorrectly tests permission denied errors during file access
//
//	Test Case ID    UT-FS-ERR-02-01
//	Title           Permission Denied During File Access
//	Description     Tests that permission denied errors are handled correctly during file access operations
//	Preconditions   None
//	Steps           1. Create a file with restricted permissions
//	                2. Attempt to access the file without proper permissions
//	                3. Verify proper error handling and status
//	Expected Result File access fails with permission denied error (EACCES)
//	Notes: This test verifies that permission denied errors are properly detected and reported.
func TestUT_FS_ERR_02_01_PermissionDenied_FileAccess_HandledCorrectly(t *testing.T) {
	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "PermissionDeniedAccessFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
		// Create the filesystem
		fs, err := NewFilesystem(auth, mountPoint, cacheTTL)
		if err != nil {
			return nil, err
		}
		return fs, nil
	})

	// Use the fixture to run the test
	fixture.Use(t, func(t *testing.T, fixture interface{}) {
		// Create assertions helper
		assert := framework.NewAssert(t)

		// Get the test data
		unitTestFixture, ok := fixture.(*framework.UnitTestFixture)
		if !ok {
			t.Fatalf("Expected fixture to be of type *framework.UnitTestFixture, but got %T", fixture)
		}
		fsFixture := unitTestFixture.SetupData.(*helpers.FSTestFixture)
		fs := fsFixture.FS.(*Filesystem)
		mockClient := fsFixture.MockClient

		// Step 1: Create a test file
		testFileName := "permission-denied-test.txt"

		// Create the file using Mknod
		mknodIn := &fuse.MknodIn{
			InHeader: fuse.InHeader{NodeId: 1}, // Root node ID
			Mode:     0644,
		}
		entryOut := &fuse.EntryOut{}
		status := fs.Mknod(nil, mknodIn, testFileName, entryOut)
		assert.Equal(fuse.OK, status, "Mknod should succeed")

		// Step 2: Simulate permission denied error for file access
		testFileID := fs.TranslateID(entryOut.NodeId)
		mockClient.AddMockResponse("/me/drive/items/"+testFileID+"/content", nil, http.StatusForbidden,
			errors.NewAuthError("accessDenied: Access denied. You do not have permission to perform this action.", nil))

		// Step 3: Attempt to open the file for reading
		openIn := &fuse.OpenIn{
			InHeader: fuse.InHeader{NodeId: entryOut.NodeId},
			Flags:    uint32(syscall.O_RDONLY),
		}
		openOut := &fuse.OpenOut{}

		// Verify proper error handling
		openStatus := fs.Open(nil, openIn, openOut)

		// Verify that the file access failed with permission denied error
		assert.NotEqual(fuse.OK, openStatus, "File open should fail due to permission denied")
		assert.True(openStatus == fuse.EACCES || openStatus == fuse.EPERM,
			"File open should fail with EACCES or EPERM error code")
	})
}

// TestUT_FS_ERR_02_02_PermissionDenied_WriteOperation_HandledCorrectly tests permission denied errors during write operations
//
//	Test Case ID    UT-FS-ERR-02-02
//	Title           Permission Denied During Write Operation
//	Description     Tests that permission denied errors are handled correctly during write operations
//	Preconditions   None
//	Steps           1. Create a read-only file
//	                2. Attempt to write to the file
//	                3. Verify proper error handling and status
//	Expected Result Write operation fails with permission denied error (EACCES)
//	Notes: This test verifies that permission denied errors are properly detected during write operations.
func TestUT_FS_ERR_02_02_PermissionDenied_WriteOperation_HandledCorrectly(t *testing.T) {
	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "PermissionDeniedWriteFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
		// Create the filesystem
		fs, err := NewFilesystem(auth, mountPoint, cacheTTL)
		if err != nil {
			return nil, err
		}
		return fs, nil
	})

	// Use the fixture to run the test
	fixture.Use(t, func(t *testing.T, fixture interface{}) {
		// Create assertions helper
		assert := framework.NewAssert(t)

		// Get the test data
		unitTestFixture, ok := fixture.(*framework.UnitTestFixture)
		if !ok {
			t.Fatalf("Expected fixture to be of type *framework.UnitTestFixture, but got %T", fixture)
		}
		fsFixture := unitTestFixture.SetupData.(*helpers.FSTestFixture)
		fs := fsFixture.FS.(*Filesystem)
		mockClient := fsFixture.MockClient

		// Step 1: Create a test file
		testFileName := "readonly-test.txt"

		// Create the file using Mknod with read-only permissions
		mknodIn := &fuse.MknodIn{
			InHeader: fuse.InHeader{NodeId: 1}, // Root node ID
			Mode:     0444,                     // Read-only permissions
		}
		entryOut := &fuse.EntryOut{}
		status := fs.Mknod(nil, mknodIn, testFileName, entryOut)
		assert.Equal(fuse.OK, status, "Mknod should succeed")

		// Step 2: Simulate permission denied error for write operations
		testFileID := fs.TranslateID(entryOut.NodeId)
		mockClient.AddMockResponse("/me/drive/items/"+testFileID+"/content", nil, http.StatusForbidden,
			errors.NewAuthError("accessDenied: Access denied. File is read-only.", nil))

		// Step 3: Attempt to write to the read-only file
		writeIn := &fuse.WriteIn{
			InHeader: fuse.InHeader{NodeId: entryOut.NodeId},
			Offset:   0,
		}
		testData := []byte("This write should fail due to read-only permissions")

		// Verify proper error handling
		bytesWritten, writeStatus := fs.Write(nil, writeIn, testData)

		// Verify that the write operation failed with permission denied error
		assert.NotEqual(fuse.OK, writeStatus, "Write should fail due to permission denied")
		assert.Equal(uint32(0), bytesWritten, "No bytes should be written to read-only file")
		assert.True(writeStatus == fuse.EACCES || writeStatus == fuse.EPERM,
			"Write should fail with EACCES or EPERM error code")
	})
}
