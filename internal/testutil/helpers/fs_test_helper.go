// Package testutil provides testing utilities for the OneMount project.
package helpers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"testing"

	"github.com/auriora/onemount/internal/graph"
	"github.com/auriora/onemount/internal/testutil"
	"github.com/auriora/onemount/internal/testutil/framework"
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

type driveChildrenPayload struct {
	Children []*graph.DriveItem `json:"value"`
	NextLink string             `json:"@odata.nextLink,omitempty"`
}

func appendChildItems(mockClient *graph.MockGraphClient, parentID string, items ...*graph.DriveItem) {
	resource := "/me/drive/items/" + parentID + "/children"
	payload := driveChildrenPayload{}
	if resp, ok := mockClient.RequestResponses[resource]; ok && len(resp.Body) > 0 {
		_ = json.Unmarshal(resp.Body, &payload)
	}
	payload.Children = append(payload.Children, items...)
	body, _ := json.Marshal(payload)
	mockClient.RequestResponses[resource] = graph.MockResponse{
		Body:       body,
		StatusCode: http.StatusOK,
	}
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

	// Stop the filesystem if it has a Stop method (to prevent goroutine leaks)
	if fixture.FS != nil {
		// Use reflection to check if the filesystem has a Stop method
		fsValue := reflect.ValueOf(fixture.FS)
		if fsValue.IsValid() && !fsValue.IsNil() {
			stopMethod := fsValue.MethodByName("Stop")
			if stopMethod.IsValid() {
				t.Logf("Calling Stop() on filesystem to clean up background goroutines")
				stopMethod.Call(nil)
			}
		}
	}

	// Clean up the mock client to prevent test interference
	if fixture.MockClient != nil {
		fixture.MockClient.Cleanup()
	}

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

	appendChildItems(mockClient, parentID, dirItem)

	return dirItem
}

// CreateMockFile creates a mock file in the filesystem.
func CreateMockFile(mockClient *graph.MockGraphClient, parentID, fileName, fileID string, content string) *graph.DriveItem {
	// Convert content to bytes and calculate hash
	contentBytes := []byte(content)
	quickXorHash := graph.QuickXORHash(&contentBytes)

	// Create a file item
	fileItem := &graph.DriveItem{
		ID:   fileID,
		Name: fileName,
		Parent: &graph.DriveItemParent{
			ID: parentID,
		},
		File: &graph.File{
			Hashes: graph.Hashes{
				QuickXorHash: quickXorHash,
			},
		},
		Size: uint64(len(content)),
	}

	// Add the file to the mock client
	mockClient.AddMockItem("/me/drive/items/"+fileID, fileItem)

	// Add content response
	contentResource := "/me/drive/items/" + fileID + "/content"
	mockClient.AddMockResponse(contentResource, []byte(content), 200, nil)

	// Get the parent's children
	appendChildItems(mockClient, parentID, fileItem)

	return fileItem
}

// NewOfflineFilesystem creates a filesystem for offline mode testing.
//
// Parameters:
//   - auth: Authentication information for the Graph API
//   - mountPoint: The directory where the filesystem will be mounted
//   - cacheTTL: Time-to-live for cached items in seconds
//
// Returns:
//   - A filesystem interface configured for offline testing
//   - An error if the filesystem could not be created
func NewOfflineFilesystem(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
	// Create a temporary cache directory for the offline filesystem
	cacheDir, err := os.MkdirTemp(mountPoint, "offline-fs-cache-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create cache directory for offline filesystem: %w", err)
	}

	// Set up a mock graph client for offline testing
	mockClient := graph.NewMockGraphClient()

	// Create a basic directory structure for offline testing
	rootID := "offline-root-id"
	rootItem := &graph.DriveItem{
		ID:   rootID,
		Name: "root",
		Folder: &graph.Folder{
			ChildCount: 2, // We'll add a test directory and file
		},
	}

	// Add the root item to the mock client
	mockClient.AddMockItem("/me/drive/root", rootItem)

	// Create a test directory for offline operations
	testDirID := "offline-test-dir-id"
	testDir := &graph.DriveItem{
		ID:   testDirID,
		Name: "test-directory",
		Parent: &graph.DriveItemParent{
			ID: rootID,
		},
		Folder: &graph.Folder{
			ChildCount: 1, // Will contain one test file
		},
	}
	mockClient.AddMockItem("/me/drive/items/"+testDirID, testDir)

	// Create a test file for offline operations
	testFileID := "offline-test-file-id"
	testFileContent := "This is test content for offline filesystem testing"
	testFileBytes := []byte(testFileContent)
	testFile := &graph.DriveItem{
		ID:   testFileID,
		Name: "test-file.txt",
		Parent: &graph.DriveItemParent{
			ID: testDirID,
		},
		File: &graph.File{
			Hashes: graph.Hashes{
				QuickXorHash: graph.QuickXORHash(&testFileBytes),
			},
		},
		Size: uint64(len(testFileContent)),
	}
	mockClient.AddMockItem("/me/drive/items/"+testFileID, testFile)
	mockClient.AddMockResponse("/me/drive/items/"+testFileID+"/content", []byte(testFileContent), 200, nil)

	// Add both items to the root's children
	mockClient.AddMockItems("/me/drive/items/"+rootID+"/children", []*graph.DriveItem{testDir, testFile})

	// Add the test file to the test directory's children
	mockClient.AddMockItems("/me/drive/items/"+testDirID+"/children", []*graph.DriveItem{testFile})

	// Create the filesystem with offline capabilities
	// Note: We need to import the fs package to create a real filesystem
	// Since this would create a circular import, we'll return a structured map
	// that contains all the necessary information for the tests to create the filesystem
	offlineFS := map[string]interface{}{
		"type":            "OfflineFilesystem",
		"auth":            auth,
		"cacheTTL":        cacheTTL,
		"mountPoint":      mountPoint,
		"cacheDir":        cacheDir,
		"mockClient":      mockClient,
		"rootID":          rootID,
		"rootItem":        rootItem,
		"testDirID":       testDirID,
		"testDir":         testDir,
		"testFileID":      testFileID,
		"testFile":        testFile,
		"testFileContent": testFileContent,
		"status":          "configured_for_offline_testing",
	}

	return offlineFS, nil
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
