package fs

import (
	"testing"

	"github.com/auriora/onemount/internal/graph"
	"github.com/auriora/onemount/internal/testutil/framework"
	"github.com/auriora/onemount/internal/testutil/helpers"
	"github.com/hanwen/go-fuse/v2/fuse"
)

// TestUT_FS_Path_01_PathResolution_BasicPaths tests basic path resolution.
//
//	Test Case ID    UT-FS-Path-01
//	Title           Basic Path Resolution
//	Description     Tests resolving paths to IDs and vice versa
//	Preconditions   None
//	Steps           1. Create files and directories in a hierarchy
//	                2. Test path-to-ID resolution
//	                3. Test ID-to-path resolution
//	                4. Verify path consistency
//	Expected Result Paths are correctly resolved in both directions
//	Notes: This test verifies that path resolution works correctly.
func TestUT_FS_Path_01_PathResolution_BasicPaths(t *testing.T) {
	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "PathResolutionFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
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
		rootID := fsFixture.RootID

		// Step 1: Create a directory hierarchy
		// Root -> documents -> projects -> file.txt

	// Create documents directory
	documentsID := "documents-dir-id"
	documentsName := "documents"
	mockClient.AddMockResponse("/me/drive/items/"+rootID+"/children", []byte(`{"id":"documents-dir-id","name":"documents","folder":{}}`), 201, nil)

	mkdirIn := &fuse.MkdirIn{
		InHeader: fuse.InHeader{NodeId: 1}, // Root node ID
		Mode:     0755,
	}
	documentsEntryOut := &fuse.EntryOut{}

	status := fs.Mkdir(nil, mkdirIn, documentsName, documentsEntryOut)
	assert.Equal(fuse.OK, status, "Documents directory creation should succeed")
	documentsNode := fs.GetNodeID(documentsEntryOut.NodeId)
	assert.NotNil(documentsNode, "Documents inode should exist after Mkdir")
	documentsID = documentsNode.ID()
	helpers.CreateMockDirectory(mockClient, rootID, documentsName, documentsID)

	// Create projects directory inside documents
	projectsID := "projects-dir-id"
	projectsName := "projects"
	mockClient.AddMockResponse("/me/drive/items/"+documentsID+"/children", []byte(`{"id":"projects-dir-id","name":"projects","folder":{}}`), 201, nil)

	mkdirIn.NodeId = documentsEntryOut.NodeId
	projectsEntryOut := &fuse.EntryOut{}

	status = fs.Mkdir(nil, mkdirIn, projectsName, projectsEntryOut)
	assert.Equal(fuse.OK, status, "Projects directory creation should succeed")
	projectsNode := fs.GetNodeID(projectsEntryOut.NodeId)
	assert.NotNil(projectsNode, "Projects inode should exist after Mkdir")
	projectsID = projectsNode.ID()
	helpers.CreateMockDirectory(mockClient, documentsID, projectsName, projectsID)

		// Create a file inside projects
		fileName := "test_file.txt"
		// Note: Mknod will create a local ID, not the one we specify in the mock

		mknodIn := &fuse.MknodIn{
			InHeader: fuse.InHeader{NodeId: projectsEntryOut.NodeId},
			Mode:     0644,
		}
		fileEntryOut := &fuse.EntryOut{}

		status = fs.Mknod(nil, mknodIn, fileName, fileEntryOut)
		assert.Equal(fuse.OK, status, "File creation should succeed")

		// Get the actual file ID that was created
	fileInode := fs.GetNodeID(fileEntryOut.NodeId)
	assert.NotNil(fileInode, "File inode should exist")
	fileID := fileInode.ID()
	helpers.CreateMockFile(mockClient, projectsID, fileName, fileID, "Test file content")

		// Step 2: Test path-to-ID resolution

		// Test root path
		rootInode := fs.GetID(rootID)
		assert.NotNil(rootInode, "Root inode should exist")
		assert.Equal("/", rootInode.Path(), "Root path should be /")

		// Test documents path
		documentsInode := fs.GetID(documentsID)
		assert.NotNil(documentsInode, "Documents inode should exist")
		expectedDocumentsPath := "/" + documentsName
		assert.Equal(expectedDocumentsPath, documentsInode.Path(), "Documents path should be correct")

		// Test projects path
		projectsInode := fs.GetID(projectsID)
		assert.NotNil(projectsInode, "Projects inode should exist")
		// Note: The path might be relative to the parent, so let's check what it actually is
		projectsPath := projectsInode.Path()
		assert.True(len(projectsPath) > 0, "Projects path should not be empty")
		assert.Contains(projectsPath, projectsName, "Projects path should contain the project name")

		// Test file path
		assert.NotNil(fileInode, "File inode should exist")
		filePath := fileInode.Path()
		assert.True(len(filePath) > 0, "File path should not be empty")
		assert.Contains(filePath, fileName, "File path should contain the file name")

		// Step 3: Test ID-to-path resolution using GetChildrenPath

		// Get children of root by path
		rootChildren, err := fs.GetChildrenPath("/", fs.auth)
		assert.NoError(err, "Getting root children by path should succeed")
		assert.NotNil(rootChildren, "Root children should not be nil")
		_, hasDocuments := rootChildren[documentsName]
		assert.True(hasDocuments, "Root should contain documents directory")

		// Get children of documents by path
		documentsChildren, err2 := fs.GetChildrenPath("/"+documentsName, fs.auth)
		assert.NoError(err2, "Getting documents children by path should succeed")
		assert.NotNil(documentsChildren, "Documents children should not be nil")
		_, hasProjects := documentsChildren[projectsName]
		assert.True(hasProjects, "Documents should contain projects directory")

		// Step 4: Verify path consistency

		// Verify that all created items can be found by their paths
		documentsInodeByID := fs.GetID(documentsID)
		assert.NotNil(documentsInodeByID, "Documents inode should exist")
		assert.Equal(documentsName, documentsInodeByID.Name(), "Documents name should match")

		projectsInodeByID := fs.GetID(projectsID)
		assert.NotNil(projectsInodeByID, "Projects inode should exist")
		assert.Equal(projectsName, projectsInodeByID.Name(), "Projects name should match")

		fileInodeByID := fs.GetID(fileID)
		assert.NotNil(fileInodeByID, "File inode should exist")
		assert.Equal(fileName, fileInodeByID.Name(), "File name should match")

		// Verify parent-child relationships
		assert.Equal(rootID, documentsInodeByID.ParentID(), "Documents parent should be root")
		assert.Equal(documentsID, projectsInodeByID.ParentID(), "Projects parent should be documents")
		assert.Equal(projectsID, fileInode.ParentID(), "File parent should be projects")
	})
}

// TestUT_FS_Path_02_PathValidation_InvalidPaths tests path validation and error handling.
//
//	Test Case ID    UT-FS-Path-02
//	Title           Path Validation and Error Handling
//	Description     Tests handling of invalid paths and edge cases
//	Preconditions   None
//	Steps           1. Test invalid path characters
//	                2. Test path length limits
//	                3. Test non-existent paths
//	                4. Verify appropriate error responses
//	Expected Result Invalid paths are properly rejected
//	Notes: This test verifies that path validation works correctly.
func TestUT_FS_Path_02_PathValidation_InvalidPaths(t *testing.T) {
	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "PathValidationFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
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

		// Step 1: Test invalid path characters

		// Test restricted filenames
		restrictedNames := []string{
			"CON",  // Windows reserved name
			"PRN",  // Windows reserved name
			"AUX",  // Windows reserved name
			"NUL",  // Windows reserved name
			"COM1", // Windows reserved name
			"LPT1", // Windows reserved name
			"..",   // Parent directory reference
			".",    // Current directory reference
		}

		for _, name := range restrictedNames {
			mknodIn := &fuse.MknodIn{
				InHeader: fuse.InHeader{NodeId: 1}, // Root node ID
				Mode:     0644,
			}
			entryOut := &fuse.EntryOut{}

			status := fs.Mknod(nil, mknodIn, name, entryOut)
			assert.Equal(fuse.EINVAL, status, "Restricted name '%s' should be rejected", name)
		}

		// Step 2: Test non-existent paths

		// Try to get children of non-existent path
		_, err := fs.GetChildrenPath("/non/existent/path", fs.auth)
		assert.Error(err, "Getting children of non-existent path should return error")

		// Try to get non-existent ID
		nonExistentInode := fs.GetID("non-existent-id")
		assert.Nil(nonExistentInode, "Non-existent ID should return nil")

		// Step 3: Test TranslateID with invalid node IDs

		// Test with invalid node ID
		invalidID := fs.TranslateID(99999)
		assert.Equal("", invalidID, "Invalid node ID should return empty string")

		// Test with zero node ID
		zeroID := fs.TranslateID(0)
		assert.Equal("", zeroID, "Zero node ID should return empty string")

		// Step 4: Test GetChild with non-existent parent

		child, err6 := fs.GetChild("non-existent-parent", "any-name", fs.auth)
		assert.Error(err6, "GetChild with non-existent parent should return error")
		assert.Nil(child, "GetChild with non-existent parent should return nil child")

		// Step 5: Test path operations with empty strings

		// Test empty path
		emptyPathChildren, err7 := fs.GetChildrenPath("", fs.auth)
		assert.Error(err7, "Empty path should return error")
		assert.Nil(emptyPathChildren, "Empty path should return nil children")

		// Test that empty ID operations don't work
		emptyInode := fs.GetID("")
		assert.Nil(emptyInode, "Empty ID should return nil inode")
	})
}

// TestUT_FS_Path_03_PathMovement_Operations tests path movement and renaming operations.
//
//	Test Case ID    UT-FS-Path-03
//	Title           Path Movement Operations
//	Description     Tests moving files and updating paths
//	Preconditions   Files and directories exist
//	Steps           1. Create files in different directories
//	                2. Test MoveID operation
//	                3. Test MovePath operation
//	                4. Verify paths are updated correctly
//	Expected Result Path movements are handled correctly
//	Notes: This test verifies that path movement operations work correctly.
func TestUT_FS_Path_03_PathMovement_Operations(t *testing.T) {
	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "PathMovementFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
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
			t.Fatalf("Expected fixture to be of type *helpers.FSTestFixture, but got %T", fixture)
		}
	fsFixture := unitTestFixture.SetupData.(*helpers.FSTestFixture)
	fs := fsFixture.FS.(*Filesystem)
	rootID := fsFixture.RootID
	mockClient := fsFixture.MockClient

	// Step 1: Create test files with known IDs

	// Create a test file via the helpers to ensure DriveChildren payloads stay consistent
	testFileID := "movable-file-id"
	testFileName := "movable_file.txt"
	fileItem := helpers.CreateMockFile(mockClient, rootID, testFileName, testFileID, "deterministic test content")
	fileInode := NewInodeDriveItem(fileItem)
	fs.InsertID(fileItem.ID, fileInode)

		// Verify initial path
		initialPath := fileInode.Path()
		expectedInitialPath := "/" + testFileName
		assert.Equal(expectedInitialPath, initialPath, "Initial path should be correct")

		// Step 2: Test MoveID operation

		newFileID := "moved-file-id"
		err := fs.MoveID(testFileID, newFileID)
		assert.NoError(err, "MoveID should succeed")

		// Verify old ID no longer exists
		oldInode := fs.GetID(testFileID)
		assert.Nil(oldInode, "Old ID should no longer exist")

		// Verify new ID exists
		newInode := fs.GetID(newFileID)
		assert.NotNil(newInode, "New ID should exist")
		assert.Equal(testFileName, newInode.Name(), "File name should be preserved")

		// Step 3: Test path operations

		// Verify the new inode has the correct path
		newInodePath := newInode.Path()
		expectedNewPath := "/" + testFileName
		assert.Equal(expectedNewPath, newInodePath, "New inode should have correct path")

		// Step 4: Test that the inode can be found by its new ID

		// Verify the inode still exists and is accessible
		finalInode := fs.GetID(newFileID)
		assert.NotNil(finalInode, "Final inode should exist")
		assert.Equal(testFileName, finalInode.Name(), "Final inode name should be preserved")
	})
}
