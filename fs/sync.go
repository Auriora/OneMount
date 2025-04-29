package fs

import (
	"github.com/bcherrington/onedriver/fs/graph"
	"github.com/rs/zerolog/log"
)

// SyncDirectoryTree recursively traverses the filesystem from the root
// and caches all directory metadata (not file contents)
func (f *Filesystem) SyncDirectoryTree(auth *graph.Auth) error {
	log.Info().Msg("Starting full directory tree synchronization...")
	return f.syncDirectoryTreeRecursive(f.root, auth)
}

// syncDirectoryTreeRecursive recursively traverses the filesystem
// starting from the given directory ID
func (f *Filesystem) syncDirectoryTreeRecursive(dirID string, auth *graph.Auth) error {
	// Get all children of the current directory
	children, err := f.GetChildrenID(dirID, auth)
	if err != nil {
		log.Error().Err(err).Str("dirID", dirID).Msg("Failed to get children")
		return err
	}

	// Recursively process all subdirectories
	for _, child := range children {
		if child.IsDir() {
			if err := f.syncDirectoryTreeRecursive(child.ID(), auth); err != nil {
				// Log the error but continue with other directories
				log.Warn().Err(err).Str("dirID", child.ID()).Msg("Error syncing directory")
			}
		}
	}

	return nil
}
