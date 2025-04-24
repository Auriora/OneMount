package fs

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/jstaf/onedriver/fs/graph"
	"github.com/rs/zerolog/log"
)

// GetThumbnail retrieves a thumbnail for a file.
// size can be "small", "medium", or "large"
func (f *Filesystem) GetThumbnail(path string, size string) ([]byte, error) {
	// Validate size
	if size != "small" && size != "medium" && size != "large" {
		return nil, fmt.Errorf("invalid thumbnail size: %s", size)
	}

	// Get the inode for the path
	inode, err := f.getInodeFromPath(path)
	if err != nil {
		return nil, err
	}

	// Only files can have thumbnails
	if inode.IsDir() {
		return nil, fmt.Errorf("directories do not have thumbnails")
	}

	// Check if we have a cached thumbnail
	if f.thumbnails.HasThumbnail(inode.ID(), size) {
		return f.thumbnails.Get(inode.ID(), size), nil
	}

	// If we're offline, we can't fetch thumbnails
	if f.IsOffline() {
		return nil, fmt.Errorf("cannot fetch thumbnails in offline mode")
	}

	// Get the thumbnail from the Graph API
	thumbnailData, err := graph.GetThumbnailContent(inode.ID(), size, f.auth)
	if err != nil {
		return nil, fmt.Errorf("failed to get thumbnail: %w", err)
	}

	// Cache the thumbnail
	if err := f.thumbnails.Insert(inode.ID(), size, thumbnailData); err != nil {
		log.Error().Err(err).
			Str("id", inode.ID()).
			Str("size", size).
			Msg("Failed to cache thumbnail")
	}

	return thumbnailData, nil
}

// GetThumbnailStream retrieves a thumbnail for a file and writes it to the provided writer.
// size can be "small", "medium", or "large"
func (f *Filesystem) GetThumbnailStream(path string, size string, output io.Writer) error {
	// Validate size
	if size != "small" && size != "medium" && size != "large" {
		return fmt.Errorf("invalid thumbnail size: %s", size)
	}

	// Get the inode for the path
	inode, err := f.getInodeFromPath(path)
	if err != nil {
		return err
	}

	// Only files can have thumbnails
	if inode.IsDir() {
		return fmt.Errorf("directories do not have thumbnails")
	}

	// Check if we have a cached thumbnail
	if f.thumbnails.HasThumbnail(inode.ID(), size) {
		// Open the cached thumbnail
		file, err := f.thumbnails.Open(inode.ID(), size)
		if err != nil {
			return err
		}
		defer f.thumbnails.Close(inode.ID(), size)

		// Copy the thumbnail to the output
		_, err = io.Copy(output, file)
		return err
	}

	// If we're offline, we can't fetch thumbnails
	if f.IsOffline() {
		return fmt.Errorf("cannot fetch thumbnails in offline mode")
	}

	// Get the thumbnail from the Graph API and stream it directly to the output
	if err := graph.GetThumbnailContentStream(inode.ID(), size, f.auth, output); err != nil {
		return fmt.Errorf("failed to get thumbnail stream: %w", err)
	}

	// Cache the thumbnail in the background
	go func() {
		// Create a temporary file to store the thumbnail
		tempFile, err := os.CreateTemp("", "onedriver-thumbnail-*")
		if err != nil {
			log.Error().Err(err).
				Str("id", inode.ID()).
				Str("size", size).
				Msg("Failed to create temporary file for thumbnail caching")
			return
		}
		defer os.Remove(tempFile.Name())
		defer tempFile.Close()

		// Get the thumbnail again and write it to the temporary file
		if err := graph.GetThumbnailContentStream(inode.ID(), size, f.auth, tempFile); err != nil {
			log.Error().Err(err).
				Str("id", inode.ID()).
				Str("size", size).
				Msg("Failed to download thumbnail for caching")
			return
		}

		// Reset the file position to the beginning
		if _, err := tempFile.Seek(0, 0); err != nil {
			log.Error().Err(err).
				Str("id", inode.ID()).
				Str("size", size).
				Msg("Failed to reset file position for thumbnail caching")
			return
		}

		// Read the thumbnail data
		thumbnailData, err := io.ReadAll(tempFile)
		if err != nil {
			log.Error().Err(err).
				Str("id", inode.ID()).
				Str("size", size).
				Msg("Failed to read thumbnail data for caching")
			return
		}

		// Cache the thumbnail
		if err := f.thumbnails.Insert(inode.ID(), size, thumbnailData); err != nil {
			log.Error().Err(err).
				Str("id", inode.ID()).
				Str("size", size).
				Msg("Failed to cache thumbnail")
		}
	}()

	return nil
}

// DeleteThumbnail deletes a cached thumbnail for a file.
// size can be "small", "medium", or "large"
// If size is empty, all thumbnails for the file are deleted.
func (f *Filesystem) DeleteThumbnail(path string, size string) error {
	// Get the inode for the path
	inode, err := f.getInodeFromPath(path)
	if err != nil {
		return err
	}

	// Delete the thumbnail(s)
	if size == "" {
		return f.thumbnails.DeleteAll(inode.ID())
	}

	// Validate size
	if size != "small" && size != "medium" && size != "large" {
		return fmt.Errorf("invalid thumbnail size: %s", size)
	}

	return f.thumbnails.Delete(inode.ID(), size)
}

// CleanupThumbnailCache cleans up the thumbnail cache
func (f *Filesystem) CleanupThumbnailCache() (int, error) {
	return f.thumbnails.CleanupCache(f.cacheExpirationDays)
}

// getInodeFromPath gets an inode from a path
func (f *Filesystem) getInodeFromPath(path string) (*Inode, error) {
	// Clean the path
	path = filepath.Clean(path)

	// Split the path into components
	components := strings.Split(path, "/")
	if len(components) > 0 && components[0] == "" {
		components = components[1:]
	}

	// Start at the root
	current := f.GetID(f.root)
	if current == nil {
		return nil, fmt.Errorf("root inode not found")
	}

	// Empty path means root
	if len(components) == 0 || (len(components) == 1 && components[0] == "") {
		return current, nil
	}

	// Traverse the path
	for _, component := range components {
		if component == "" {
			continue
		}

		// Get the child with the given name
		child, err := f.GetChild(current.ID(), component, f.auth)
		if err != nil {
			return nil, fmt.Errorf("failed to get child %s: %w", component, err)
		}
		if child == nil {
			return nil, fmt.Errorf("path component not found: %s", component)
		}

		current = child
	}

	return current, nil
}
