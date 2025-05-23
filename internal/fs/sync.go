package fs

import (
	"github.com/auriora/onemount/pkg/graph"
	"github.com/auriora/onemount/pkg/logging"
)

// SyncDirectoryTree recursively traverses the filesystem from the root
// and caches all directory metadata (not file contents)
func (f *Filesystem) SyncDirectoryTree(auth *graph.Auth) error {
	logging.Info().Msg("Starting full directory tree synchronization...")
	// Create a map to track visited directories to prevent cycles
	visited := make(map[string]bool)
	// Set a reasonable maximum depth to prevent excessive recursion
	const maxDepth = 20
	return f.syncDirectoryTreeRecursive(f.root, auth, visited, 0, maxDepth)
}

// syncDirectoryTreeRecursive recursively traverses the filesystem
// starting from the given directory ID
func (f *Filesystem) syncDirectoryTreeRecursive(dirID string, auth *graph.Auth, visited map[string]bool, depth int, maxDepth int) error {
	// Check if we've already visited this directory to prevent cycles
	if visited[dirID] {
		logging.Debug().Str("dirID", dirID).Msg("Skipping already visited directory to prevent cycle")
		return nil
	}

	// Check if we've reached the maximum depth
	if depth >= maxDepth {
		logging.Warn().Str("dirID", dirID).Int("depth", depth).Int("maxDepth", maxDepth).Msg("Reached maximum recursion depth, stopping")
		return nil
	}

	// Mark this directory as visited
	visited[dirID] = true

	// Get all children of the current directory
	children, err := f.GetChildrenID(dirID, auth)
	if err != nil {
		logging.Error().Err(err).Str("dirID", dirID).Msg("Failed to get children")
		return err
	}

	// Recursively process all subdirectories
	for _, child := range children {
		if child.IsDir() {
			if err := f.syncDirectoryTreeRecursive(child.ID(), auth, visited, depth+1, maxDepth); err != nil {
				// Log the error but continue with other directories
				logging.Warn().Err(err).Str("dirID", child.ID()).Msg("Error syncing directory")
			}
		}
	}

	return nil
}
