// Package testutil provides testing utilities for the OneMount project.
package helpers

import (
	"fmt"
	"github.com/auriora/onemount/internal/testutil"
	"github.com/auriora/onemount/internal/testutil/framework"
	"os"
	"testing"

	"github.com/auriora/onemount/internal/fs/graph"
)

// FSTestFixture represents a filesystem test fixture with common setup for filesystem tests.
type FSTestFixture struct {
	// TempDir is the temporary directory for the test
	TempDir string
	// MockClient is the mock graph client
	MockClient *graph.MockGraphClient
	// RootID is the ID of the root directory
	RootID string
	// Auth is the authentication object
	Auth *graph.Auth
	// FS is the filesystem object
	FS interface{}
	// Additional data for the test
	Data map[string]interface{}
}

// SetupFSTest sets up a common filesystem test environment.
// It creates a temporary directory, mock graph client, and filesystem object.
// The caller is responsible for providing the NewFilesystem function and any additional setup.
func SetupFSTest(t *testing.T, testName string, newFilesystem func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error)) (*FSTestFixture, error) {
	// Ensure we're in online mode for test setup
	graph.SetOperationalOffline(false)

	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp(testutil.TestSandboxTmpDir, "onemount-test-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create a temporary directory: %w", err)
	}

	// Create a mock graph client
	mockClient := graph.NewMockGraphClient()

	// Set up the mock directory structure with a root ID
	rootID := "root-id"
	rootItem := &graph.DriveItem{
		ID:   rootID,
		Name: "root",
		Folder: &graph.Folder{
			ChildCount: 0,
		},
	}

	// Add the root item to the mock client
	mockClient.AddMockItem("/me/drive/root", rootItem)
	mockClient.AddMockItems("/me/drive/items/"+rootID+"/children", []*graph.DriveItem{})

	// Get auth tokens, either from existing file or create mock
	auth := GetTestAuth()

	// Create the filesystem
	fs, err := newFilesystem(auth, tempDir, 30)
	if err != nil {
		// Clean up the temporary directory if filesystem creation fails
		if cleanupErr := os.RemoveAll(tempDir); cleanupErr != nil {
			t.Logf("Warning: Failed to clean up temporary directory %s: %v", tempDir, cleanupErr)
		}
		return nil, fmt.Errorf("failed to create filesystem: %w", err)
	}

	// Create the fixture
	fixture := &FSTestFixture{
		TempDir:    tempDir,
		MockClient: mockClient,
		RootID:     rootID,
		Auth:       auth,
		FS:         fs,
		Data:       make(map[string]interface{}),
	}

	return fixture, nil
}

// CleanupFSTest cleans up the filesystem test environment.
func CleanupFSTest(t *testing.T, fixture *FSTestFixture) error {
	// Ensure we reset to online mode after the test
	graph.SetOperationalOffline(false)

	// Clean up the temporary directory
	if err := os.RemoveAll(fixture.TempDir); err != nil {
		t.Logf("Warning: Failed to clean up temporary directory %s: %v", fixture.TempDir, err)
		return err
	}
	return nil
}

// CreateMockDirectory creates a mock directory in the filesystem.
func CreateMockDirectory(mockClient *graph.MockGraphClient, parentID, dirName, dirID string) *graph.DriveItem {
	// Create a directory item
	dirItem := &graph.DriveItem{
		ID:   dirID,
		Name: dirName,
		Parent: &graph.DriveItemParent{
			ID: parentID,
		},
		Folder: &graph.Folder{
			ChildCount: 0,
		},
	}

	// Add the directory to the mock client
	mockClient.AddMockItem("/me/drive/items/"+dirID, dirItem)
	mockClient.AddMockItems("/me/drive/items/"+dirID+"/children", []*graph.DriveItem{})

	// Get the parent's children
	parentResource := "/me/drive/items/" + parentID + "/children"
	// Add the directory to the parent's children
	// We need to get the existing children first, but since we can't directly access them,
	// we'll create a new list with just this item
	mockClient.AddMockItems(parentResource, []*graph.DriveItem{dirItem})

	return dirItem
}

// CreateMockFile creates a mock file in the filesystem.
func CreateMockFile(mockClient *graph.MockGraphClient, parentID, fileName, fileID string, content string) *graph.DriveItem {
	// Create a file item
	fileItem := &graph.DriveItem{
		ID:   fileID,
		Name: fileName,
		Parent: &graph.DriveItemParent{
			ID: parentID,
		},
		File: &graph.File{},
		Size: uint64(len(content)),
	}

	// Add the file to the mock client
	mockClient.AddMockItem("/me/drive/items/"+fileID, fileItem)

	// Add content response
	contentResource := "/me/drive/items/" + fileID + "/content"
	mockClient.AddMockResponse(contentResource, []byte(content), 200, nil)

	// Get the parent's children
	parentResource := "/me/drive/items/" + parentID + "/children"
	// Add the file to the parent's children
	// We need to get the existing children first, but since we can't directly access them,
	// we'll create a new list with just this item
	mockClient.AddMockItems(parentResource, []*graph.DriveItem{fileItem})

	return fileItem
}

// NewOfflineFilesystem creates a stub filesystem for offline mode testing.
//
// Parameters:
//   - auth: Authentication information for the Graph API
//   - mountPoint: The directory where the filesystem will be mounted
//   - cacheTTL: Time-to-live for cached items in seconds
//
// Returns:
//   - A stub filesystem interface for offline testing
//   - An error if the filesystem could not be created
func NewOfflineFilesystem(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
	// This is a stub implementation that will be completed later
	return nil, fmt.Errorf("NewOfflineFilesystem stub in testutil/helpers not implemented yet")
}

// SetupFSTestFixture creates a UnitTestFixture with common filesystem test setup and teardown.
func SetupFSTestFixture(t *testing.T, fixtureName string, newFilesystem func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error)) *framework.UnitTestFixture {
	fixture := framework.NewUnitTestFixture(fixtureName)

	// Set up the fixture
	fixture.WithSetup(func(t *testing.T) (interface{}, error) {
		// Ensure we're in online mode for test setup
		graph.SetOperationalOffline(false)

		fsFixture, err := SetupFSTest(t, fixtureName, newFilesystem)
		if err != nil {
			return nil, err
		}
		return fsFixture, nil
	}).WithTeardown(func(t *testing.T, fixture interface{}) error {
		fsFixture := fixture.(*FSTestFixture)

		// Ensure we reset to online mode after the test
		graph.SetOperationalOffline(false)

		return CleanupFSTest(t, fsFixture)
	})

	return fixture
}
