// Package api defines interfaces and types for the graph package.
package api

import (
	"time"
)

// Header represents an HTTP header.
type Header struct {
	Key   string
	Value string
}

// DriveItemParent represents the parent of a drive item.
type DriveItemParent struct {
	DriveID   string `json:"driveId,omitempty"`
	DriveType string `json:"driveType,omitempty"`
	ID        string `json:"id,omitempty"`
	Path      string `json:"path,omitempty"`
}

// Folder represents a folder in a drive.
type Folder struct {
	ChildCount uint32 `json:"childCount,omitempty"`
}

// Hashes represents integrity hashes for a file.
type Hashes struct {
	SHA1Hash     string `json:"sha1Hash,omitempty"`
	QuickXorHash string `json:"quickXorHash,omitempty"`
}

// File represents a file in a drive.
type File struct {
	Hashes Hashes `json:"hashes,omitempty"`
}

// Deleted represents a deleted item.
type Deleted struct {
	State string `json:"state,omitempty"`
}

// DriveItem represents an item in a drive.
type DriveItem struct {
	ID               string           `json:"id,omitempty"`
	Name             string           `json:"name,omitempty"`
	Size             uint64           `json:"size,omitempty"`
	ModTime          *time.Time       `json:"lastModifiedDatetime,omitempty"`
	Parent           *DriveItemParent `json:"parentReference,omitempty"`
	Folder           *Folder          `json:"folder,omitempty"`
	File             *File            `json:"file,omitempty"`
	Deleted          *Deleted         `json:"deleted,omitempty"`
	ConflictBehavior string           `json:"@microsoft.graph.conflictBehavior,omitempty"`
	ETag             string           `json:"eTag,omitempty"`
}

// IsDir returns if the DriveItem represents a directory or not.
func (d *DriveItem) IsDir() bool {
	return d.Folder != nil
}

// ModTimeUnix returns the modification time as a unix uint64 time.
func (d *DriveItem) ModTimeUnix() uint64 {
	if d.ModTime == nil {
		return 0
	}
	return uint64(d.ModTime.Unix())
}

// DriveChildren represents a collection of DriveItems with pagination support.
type DriveChildren struct {
	Children []*DriveItem `json:"value"`
	NextLink string       `json:"@odata.nextLink"`
}

// DriveTypePersonal is the constant for a personal drive type.
const DriveTypePersonal = "personal"

// DriveQuota is used to parse the User's current storage quotas from the API.
type DriveQuota struct {
	Deleted   uint64 `json:"deleted"`
	FileCount uint64 `json:"fileCount"`
	Remaining uint64 `json:"remaining"`
	State     string `json:"state"`
	Total     uint64 `json:"total"`
	Used      uint64 `json:"used"`
}

// Drive represents a drive in the Microsoft Graph API.
type Drive struct {
	ID        string     `json:"id,omitempty"`
	DriveType string     `json:"driveType,omitempty"`
	Name      string     `json:"name,omitempty"`
	Quota     DriveQuota `json:"quota,omitempty"`
}

// User represents a user in the Microsoft Graph API.
type User struct {
	ID                string `json:"id,omitempty"`
	DisplayName       string `json:"displayName,omitempty"`
	Email             string `json:"mail,omitempty"`
	UserPrincipalName string `json:"userPrincipalName,omitempty"`
}
