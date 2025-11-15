package fs

import (
	"regexp"
	"strings"
	"time"

	"github.com/auriora/onemount/internal/graph"
	"github.com/auriora/onemount/internal/logging"
)

// Package fs implements the core FUSE-based filesystem logic, including caching, file operations, and OneDrive integration for OneMount.

const timeout = time.Second

func (f *Filesystem) getInodeContent(i *Inode) *[]byte {
	i.mu.RLock()
	defer i.mu.RUnlock()
	data := f.content.Get(i.DriveItem.ID)
	return &data
}

// GetInodeContent returns the content of an inode.
// This method is part of the FilesystemInterface.
func (f *Filesystem) GetInodeContent(i *Inode) *[]byte {
	return f.getInodeContent(i)
}

// GetInodeContentPath returns the file path to the inode's content in the cache.
// This is more memory-efficient than GetInodeContent for large files.
// This method is part of the FilesystemInterface.
func (f *Filesystem) GetInodeContentPath(i *Inode) string {
	i.mu.RLock()
	id := i.DriveItem.ID
	i.mu.RUnlock()
	return f.content.contentPath(id)
}

// StoreContent stores content in the cache for the given inode ID.
// This is useful for creating local-only virtual files that don't sync to OneDrive.
func (f *Filesystem) StoreContent(id string, content []byte) error {
	return f.content.Insert(id, content)
}

// remoteID uploads a file to obtain a Onedrive ID if it doesn't already
// have one. This is necessary to avoid race conditions against uploads if the
// file has not already been uploaded.
func (f *Filesystem) remoteID(i *Inode) (string, error) {
	if i.IsDir() {
		// Directories are always created with an ID. (And this method is only
		// really used for files anyways...)
		return i.ID(), nil
	}

	originalID := i.ID()
	if isLocalID(originalID) && f.auth.AccessToken != "" {
		// perform a blocking upload of the item
		data := f.getInodeContent(i)
		session, err := NewUploadSession(i, data)
		if err != nil {
			return originalID, err
		}

		i.mu.Lock()
		name := i.DriveItem.Name
		err = session.Upload(f.auth)
		if err != nil {
			i.mu.Unlock()

			if strings.Contains(err.Error(), "nameAlreadyExists") {
				// A file with this name already exists on the server, get its ID and
				// use that. This is probably the same file, but just got uploaded
				// earlier.
				children, err := graph.GetItemChildren(i.ParentID(), f.auth)
				if err != nil {
					return originalID, err
				}
				for _, child := range children {
					if child.Name == name {
						logging.Info().
							Str("name", name).
							Str("originalID", originalID).
							Str("newID", child.ID).
							Msg("Exchanged ID.")
						return child.ID, f.MoveID(originalID, child.ID)
					}
				}
			}
			// failed to obtain an ID, return whatever it was beforehand
			return originalID, err
		}

		// we just successfully uploaded a copy, no need to do it again
		i.hasChanges = false
		i.DriveItem.ETag = session.ETag
		i.mu.Unlock()

		// this is all we really wanted from this transaction
		err = f.MoveID(originalID, session.ID)
		logging.Info().
			Str("name", name).
			Str("originalID", originalID).
			Str("newID", session.ID).
			Msg("Exchanged ID.")
		return session.ID, err
	}
	return originalID, nil
}

var disallowedRexp = regexp.MustCompile(`(?i)LPT[0-9]|COM[0-9]|_vti_|["*:<>?/\\|]`)

// isNameRestricted returns true if the name is disallowed according to the doc here:
// https://support.microsoft.com/en-us/office/restrictions-and-limitations-in-onedrive-and-sharepoint-64883a5d-228e-48f5-b3d2-eb39e07630fa
func isNameRestricted(name string) bool {
	if name == "." || name == ".." {
		return true
	}
	if strings.EqualFold(name, "CON") {
		return true
	}
	if strings.EqualFold(name, "AUX") {
		return true
	}
	if strings.EqualFold(name, "PRN") {
		return true
	}
	if strings.EqualFold(name, "NUL") {
		return true
	}
	if strings.EqualFold(name, ".lock") {
		return true
	}
	if strings.EqualFold(name, "desktop.ini") {
		return true
	}
	return disallowedRexp.FindStringIndex(name) != nil
}
