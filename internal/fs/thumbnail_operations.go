package fs

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/auriora/onemount/internal/common/errors"
	"github.com/auriora/onemount/internal/fs/graph"
)

// GetThumbnail retrieves a thumbnail for a file.
// size can be "small", "medium", or "large"
func (f *Filesystem) GetThumbnail(path string, size string) ([]byte, error) {
	// Validate size
	if size != "small" && size != "medium" && size != "large" {
		return nil, errors.NewValidationError(fmt.Sprintf("invalid thumbnail size: %s", size), nil)
	}

	// Get the inode for the path
	inode, err := f.getInodeFromPath(path)
	if err != nil {
		return nil, err
	}

	// Only files can have thumbnails
	if inode.IsDir() {
		return nil, errors.NewValidationError("directories do not have thumbnails", nil)
	}

	// Check if we have a cached thumbnail
	if f.thumbnails.HasThumbnail(inode.ID(), size) {
		return f.thumbnails.Get(inode.ID(), size), nil
	}

	// If we're offline, we can't fetch thumbnails
	if f.IsOffline() {
		return nil, errors.NewNetworkError("cannot fetch thumbnails in offline mode", nil)
	}

	// Get the thumbnail from the Graph API
	thumbnailData, err := graph.GetThumbnailContent(inode.ID(), size, f.auth)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get thumbnail")
	}

	// Cache the thumbnail
	if err := f.thumbnails.Insert(inode.ID(), size, thumbnailData); err != nil {
		errors.LogError(err, "Failed to cache thumbnail", 
			errors.FieldID, inode.ID(),
			"size", size,
			errors.FieldOperation, "GetThumbnail")
	}

	return thumbnailData, nil
}

// GetThumbnailStream retrieves a thumbnail for a file and writes it to the provided writer.
// size can be "small", "medium", or "large"
func (f *Filesystem) GetThumbnailStream(path string, size string, output io.Writer) error {
	// Validate size
	if size != "small" && size != "medium" && size != "large" {
		return errors.NewValidationError(fmt.Sprintf("invalid thumbnail size: %s", size), nil)
	}

	// Get the inode for the path
	inode, err := f.getInodeFromPath(path)
	if err != nil {
		return err
	}

	// Only files can have thumbnails
	if inode.IsDir() {
		return errors.NewValidationError("directories do not have thumbnails", nil)
	}

	// Check if we have a cached thumbnail
	if f.thumbnails.HasThumbnail(inode.ID(), size) {
		// Open the cached thumbnail
		file, err := f.thumbnails.Open(inode.ID(), size)
		if err != nil {
			return err
		}
		defer func() {
			if closeErr := f.thumbnails.Close(inode.ID(), size); closeErr != nil {
				errors.LogError(closeErr, "Failed to close cached thumbnail file", 
					errors.FieldID, inode.ID(),
					"size", size,
					errors.FieldOperation, "GetThumbnailStream")
			}
		}()

		// Copy the thumbnail to the output
		_, err = io.Copy(output, file)
		return err
	}

	// If we're offline, we can't fetch thumbnails
	if f.IsOffline() {
		return errors.NewNetworkError("cannot fetch thumbnails in offline mode", nil)
	}

	// Get the thumbnail from the Graph API and stream it directly to the output
	if err := graph.GetThumbnailContentStream(inode.ID(), size, f.auth, output); err != nil {
		return errors.Wrap(err, "failed to get thumbnail stream")
	}

	// Cache the thumbnail in the background
	go func() {
		// Create a temporary file to store the thumbnail
		tempFile, err := os.CreateTemp("", "onemount-thumbnail-*")
		if err != nil {
			errors.LogError(err, "Failed to create temporary file for thumbnail caching", 
				errors.FieldID, inode.ID(),
				"size", size,
				errors.FieldOperation, "GetThumbnailStream.cacheInBackground")
			return
		}
		defer func() {
			if removeErr := os.Remove(tempFile.Name()); removeErr != nil {
				errors.LogError(removeErr, "Failed to remove temporary thumbnail file", 
					errors.FieldPath, tempFile.Name(),
					errors.FieldID, inode.ID(),
					"size", size,
					errors.FieldOperation, "GetThumbnailStream.cacheInBackground")
			}
		}()
		defer func() {
			if closeErr := tempFile.Close(); closeErr != nil {
				errors.LogError(closeErr, "Failed to close temporary thumbnail file", 
					errors.FieldPath, tempFile.Name(),
					errors.FieldID, inode.ID(),
					"size", size,
					errors.FieldOperation, "GetThumbnailStream.cacheInBackground")
			}
		}()

		// Get the thumbnail again and write it to the temporary file
		if err := graph.GetThumbnailContentStream(inode.ID(), size, f.auth, tempFile); err != nil {
			errors.LogError(err, "Failed to download thumbnail for caching", 
				errors.FieldID, inode.ID(),
				"size", size,
				errors.FieldOperation, "GetThumbnailStream.cacheInBackground")
			return
		}

		// Reset the file position to the beginning
		if _, err := tempFile.Seek(0, 0); err != nil {
			errors.LogError(err, "Failed to reset file position for thumbnail caching", 
				errors.FieldID, inode.ID(),
				"size", size,
				errors.FieldOperation, "GetThumbnailStream.cacheInBackground")
			return
		}

		// Read the thumbnail data
		thumbnailData, err := io.ReadAll(tempFile)
		if err != nil {
			errors.LogError(err, "Failed to read thumbnail data for caching", 
				errors.FieldID, inode.ID(),
				"size", size,
				errors.FieldOperation, "GetThumbnailStream.cacheInBackground")
			return
		}

		// Cache the thumbnail
		if err := f.thumbnails.Insert(inode.ID(), size, thumbnailData); err != nil {
			errors.LogError(err, "Failed to cache thumbnail", 
				errors.FieldID, inode.ID(),
				"size", size,
				errors.FieldOperation, "GetThumbnailStream.cacheInBackground")
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
		return errors.NewValidationError(fmt.Sprintf("invalid thumbnail size: %s", size), nil)
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
		return nil, errors.NewNotFoundError("root inode not found", nil)
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
			return nil, errors.Wrap(err, fmt.Sprintf("failed to get child %s", component))
		}
		if child == nil {
			return nil, errors.NewNotFoundError(fmt.Sprintf("path component not found: %s", component), nil)
		}

		current = child
	}

	return current, nil
}
