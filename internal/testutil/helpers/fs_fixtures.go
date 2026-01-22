// Package helpers provides testing utilities for the OneMount project.
package helpers

import (
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/auriora/onemount/internal/graph"
	"github.com/auriora/onemount/internal/testutil"
	"github.com/auriora/onemount/internal/testutil/framework"
)

// SetupMockFSTestFixture creates a test fixture with a mock Graph API backend.
// This fixture is designed for unit tests that don't require real OneDrive authentication.
//
// Features:
//   - Uses MockGraphProvider instead of real authentication
//   - Creates filesystem with mock backend
//   - No authentication required
//   - Fast and isolated testing
//
// Usage:
//
//	fixture := helpers.SetupMockFSTestFixture(t, "MyUnitTest", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
//	    return fs.NewFilesystem(auth, mountPoint, cacheTTL)
//	})
//	defer fixture.Teardown(t)
//
//	// Access the fixture data
//	fsFixture := fixture.GetFixture(t).(*FSTestFixture)
//	mockClient := fsFixture.MockClient
//	// ... test code ...
func SetupMockFSTestFixture(t *testing.T, fixtureName string, newFilesystem func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error)) *framework.UnitTestFixture {
	fixture := framework.NewUnitTestFixture(fixtureName)

	// Set up the fixture with mock authentication
	fixture.WithSetup(func(t *testing.T) (interface{}, error) {
		// Ensure we're in online mode for test setup
		graph.SetOperationalOffline(false)

		// Create a temporary directory for the test
		tempDir, err := os.MkdirTemp(testutil.TestSandboxTmpDir, "onemount-mock-test-*")
		if err != nil {
			return nil, fmt.Errorf("failed to create a temporary directory: %w", err)
		}

		// Create a mock graph client
		mockClient := graph.NewMockGraphClient()

		// Set up the mock directory structure with a root ID
		rootID := "mock-root-id"
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

		// Create mock authentication (no real tokens needed)
		auth := createMockAuth()

		// Create the filesystem with mock backend
		fs, err := newFilesystem(auth, tempDir, 30)
		if err != nil {
			// Clean up the temporary directory if filesystem creation fails
			if cleanupErr := os.RemoveAll(tempDir); cleanupErr != nil {
				t.Logf("Warning: Failed to clean up temporary directory %s: %v", tempDir, cleanupErr)
			}
			return nil, fmt.Errorf("failed to create filesystem with mock backend: %w", err)
		}

		// Create the fixture
		fsFixture := &FSTestFixture{
			TempDir:    tempDir,
			MockClient: mockClient,
			RootID:     rootID,
			Auth:       auth,
			FS:         fs,
			Data:       make(map[string]interface{}),
		}

		return fsFixture, nil
	}).WithTeardown(func(t *testing.T, fixture interface{}) error {
		fsFixture := fixture.(*FSTestFixture)

		// Ensure we reset to online mode after the test
		graph.SetOperationalOffline(false)

		// Stop the filesystem if it has a Stop method (to prevent goroutine leaks)
		if fsFixture.FS != nil {
			fsValue := reflect.ValueOf(fsFixture.FS)
			if fsValue.IsValid() && !fsValue.IsNil() {
				stopMethod := fsValue.MethodByName("Stop")
				if stopMethod.IsValid() {
					t.Logf("Calling Stop() on filesystem to clean up background goroutines")
					stopMethod.Call(nil)
				}
			}
		}

		// Clean up the mock client to prevent test interference
		if fsFixture.MockClient != nil {
			fsFixture.MockClient.Cleanup()
		}

		// Clean up the temporary directory
		if err := os.RemoveAll(fsFixture.TempDir); err != nil {
			t.Logf("Warning: Failed to clean up temporary directory %s: %v", fsFixture.TempDir, err)
			return err
		}
		return nil
	})

	return fixture
}

// SetupIntegrationFSTestFixture creates a test fixture with real OneDrive authentication.
// This fixture is designed for integration tests that require real Graph API access.
//
// Features:
//   - Uses real auth tokens from .auth_tokens.json
//   - Creates filesystem with real OneDrive connection
//   - Requires auth tokens to be present
//   - Tests against real Microsoft Graph API
//
// Requirements:
//   - Auth tokens must be available in test-artifacts/.auth_tokens.json
//   - Network connectivity required
//   - May be slower than unit tests
//
// Usage:
//
//	fixture := helpers.SetupIntegrationFSTestFixture(t, "MyIntegrationTest", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
//	    return fs.NewFilesystem(auth, mountPoint, cacheTTL)
//	})
//	defer fixture.Teardown(t)
//
//	// Access the fixture data
//	fsFixture := fixture.GetFixture(t).(*FSTestFixture)
//	// ... test code with real OneDrive ...
func SetupIntegrationFSTestFixture(t *testing.T, fixtureName string, newFilesystem func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error)) *framework.UnitTestFixture {
	fixture := framework.NewUnitTestFixture(fixtureName)

	// Set up the fixture with real authentication
	fixture.WithSetup(func(t *testing.T) (interface{}, error) {
		// Ensure we're in online mode for test setup
		graph.SetOperationalOffline(false)

		// Create a temporary directory for the test
		tempDir, err := os.MkdirTemp(testutil.TestSandboxTmpDir, "onemount-integration-test-*")
		if err != nil {
			return nil, fmt.Errorf("failed to create a temporary directory: %w", err)
		}

		// Load real authentication tokens
		auth := GetTestAuth()

		// Verify we have real tokens (not mock)
		if auth.AccessToken == "mock-access-token" {
			// Clean up temp directory
			os.RemoveAll(tempDir)
			t.Skip("Skipping integration test: real auth tokens not available")
			return nil, fmt.Errorf("real auth tokens required for integration tests")
		}

		// Create the filesystem with real OneDrive connection
		fs, err := newFilesystem(auth, tempDir, 30)
		if err != nil {
			// Clean up the temporary directory if filesystem creation fails
			if cleanupErr := os.RemoveAll(tempDir); cleanupErr != nil {
				t.Logf("Warning: Failed to clean up temporary directory %s: %v", tempDir, cleanupErr)
			}
			return nil, fmt.Errorf("failed to create filesystem with real OneDrive: %w", err)
		}

		// Create the fixture (no mock client for integration tests)
		fsFixture := &FSTestFixture{
			TempDir:    tempDir,
			MockClient: nil, // No mock client for integration tests
			RootID:     "",  // Will be determined by real OneDrive
			Auth:       auth,
			FS:         fs,
			Data:       make(map[string]interface{}),
		}

		return fsFixture, nil
	}).WithTeardown(func(t *testing.T, fixture interface{}) error {
		fsFixture := fixture.(*FSTestFixture)

		// Ensure we reset to online mode after the test
		graph.SetOperationalOffline(false)

		// Stop the filesystem if it has a Stop method (to prevent goroutine leaks)
		if fsFixture.FS != nil {
			fsValue := reflect.ValueOf(fsFixture.FS)
			if fsValue.IsValid() && !fsValue.IsNil() {
				stopMethod := fsValue.MethodByName("Stop")
				if stopMethod.IsValid() {
					t.Logf("Calling Stop() on filesystem to clean up background goroutines")
					stopMethod.Call(nil)
				}
			}
		}

		// Clean up the temporary directory
		if err := os.RemoveAll(fsFixture.TempDir); err != nil {
			t.Logf("Warning: Failed to clean up temporary directory %s: %v", fsFixture.TempDir, err)
			return err
		}
		return nil
	})

	return fixture
}
