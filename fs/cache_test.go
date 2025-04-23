// these tests are independent of the mounted fs
package fs

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRootGet(t *testing.T) {
	cache, err := NewFilesystem(auth, filepath.Join(testDBLoc, "test_root_get"), 30)
	require.NoError(t, err)
	root, err := cache.GetPath("/", auth)
	require.NoError(t, err)
	assert.Equal(t, "/", root.Path(), "Root path did not resolve correctly.")
}

func TestRootChildrenUpdate(t *testing.T) {
	cache, err := NewFilesystem(auth, filepath.Join(testDBLoc, "test_root_children_update"), 30)
	require.NoError(t, err)
	children, err := cache.GetChildrenPath("/", auth)
	require.NoError(t, err)

	require.Contains(t, children, "documents", "Could not find documents folder.")
}

func TestSubdirGet(t *testing.T) {
	cache, err := NewFilesystem(auth, filepath.Join(testDBLoc, "test_subdir_get"), 30)
	require.NoError(t, err)
	documents, err := cache.GetPath("/Documents", auth)
	require.NoError(t, err)
	assert.Equal(t, "Documents", documents.Name(), "Failed to fetch \"/Documents\".")
}

func TestSubdirChildrenUpdate(t *testing.T) {
	cache, err := NewFilesystem(auth, filepath.Join(testDBLoc, "test_subdir_children_update"), 30)
	require.NoError(t, err)
	children, err := cache.GetChildrenPath("/Documents", auth)
	require.NoError(t, err)

	require.NotContains(t, children, "documents",
		"Documents directory found inside itself. Likely the cache did not traverse correctly.\nChildren: %v",
		children)
}

func TestSamePointer(t *testing.T) {
	cache, err := NewFilesystem(auth, filepath.Join(testDBLoc, "test_same_pointer"), 30)
	require.NoError(t, err)
	item, _ := cache.GetPath("/Documents", auth)
	item2, _ := cache.GetPath("/Documents", auth)
	require.Same(t, item, item2, "Pointers to cached items do not match: %p != %p", item, item2)
	assert.NotNil(t, item)
}
