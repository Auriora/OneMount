// these tests are independent of the mounted fs
package fs

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCacheOperations tests various cache operations using a table-driven approach
func TestCacheOperations(t *testing.T) {
	// Define test cases
	testCases := []struct {
		name        string
		dbName      string
		operation   string // "get_path", "get_children", or "check_pointers"
		path        string
		verifyFunc  func(t *testing.T, cache *Filesystem, path string)
		description string
	}{
		{
			name:        "GetRootPath_ShouldReturnRootItem",
			dbName:      "test_root_get",
			operation:   "get_path",
			path:        "/",
			description: "Get the root directory from the cache",
			verifyFunc: func(t *testing.T, cache *Filesystem, path string) {
				// Get the root item
				root, err := cache.GetPath(path, auth)
				require.NoError(t, err, "Failed to get root path")
				assert.Equal(t, "/", root.Path(), "Root path did not resolve correctly")
			},
		},
		{
			name:        "GetRootChildren_ShouldContainDocumentsFolder",
			dbName:      "test_root_children_update",
			operation:   "get_children",
			path:        "/",
			description: "Get the children of the root directory",
			verifyFunc: func(t *testing.T, cache *Filesystem, path string) {
				// Get the children of the root
				children, err := cache.GetChildrenPath(path, auth)
				require.NoError(t, err, "Failed to get root children")
				require.Contains(t, children, "documents", "Could not find documents folder")

				// Log the children for debugging
				t.Logf("Root children: %v", children)
			},
		},
		{
			name:        "GetDocumentsPath_ShouldReturnDocumentsItem",
			dbName:      "test_subdir_get",
			operation:   "get_path",
			path:        "/Documents",
			description: "Get the Documents directory from the cache",
			verifyFunc: func(t *testing.T, cache *Filesystem, path string) {
				// Get the Documents item
				documents, err := cache.GetPath(path, auth)
				require.NoError(t, err, "Failed to get Documents path")
				assert.Equal(t, "Documents", documents.Name(), "Failed to fetch \"/Documents\"")
			},
		},
		{
			name:        "GetDocumentsChildren_ShouldNotContainDocumentsFolder",
			dbName:      "test_subdir_children_update",
			operation:   "get_children",
			path:        "/Documents",
			description: "Get the children of the Documents directory",
			verifyFunc: func(t *testing.T, cache *Filesystem, path string) {
				// Get the children of Documents
				children, err := cache.GetChildrenPath(path, auth)
				require.NoError(t, err, "Failed to get Documents children")
				require.NotContains(t, children, "documents",
					"Documents directory found inside itself. Likely the cache did not traverse correctly.\nChildren: %v",
					children)

				// Log the children for debugging
				t.Logf("Documents children: %v", children)
			},
		},
		{
			name:        "GetSamePathTwice_ShouldReturnSamePointer",
			dbName:      "test_same_pointer",
			operation:   "check_pointers",
			path:        "/Documents",
			description: "Check that getting the same item twice returns the same pointer",
			verifyFunc: func(t *testing.T, cache *Filesystem, path string) {
				// Get the item twice
				item1, err := cache.GetPath(path, auth)
				require.NoError(t, err, "Failed to get item first time")
				require.NotNil(t, item1, "First item should not be nil")

				item2, err := cache.GetPath(path, auth)
				require.NoError(t, err, "Failed to get item second time")
				require.NotNil(t, item2, "Second item should not be nil")

				// Check that they are the same pointer
				require.Same(t, item1, item2, "Pointers to cached items do not match: %p != %p", item1, item2)
			},
		},
	}

	// Run each test case
	for _, tc := range testCases {
		tc := tc // Capture range variable for parallel execution
		t.Run(tc.name, func(t *testing.T) {
			// Create a unique database location for this test
			dbPath := filepath.Join(testDBLoc, tc.dbName+"_"+t.Name())

			// Create a new filesystem cache
			cache, err := NewFilesystem(auth, dbPath, 30)
			require.NoError(t, err, "Failed to create filesystem cache")

			// Run the verification function
			tc.verifyFunc(t, cache, tc.path)
		})
	}
}
