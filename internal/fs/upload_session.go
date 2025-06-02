package fs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/auriora/onemount/pkg/errors"
	"github.com/auriora/onemount/pkg/graph"
	"github.com/auriora/onemount/pkg/logging"
	bolt "go.etcd.io/bbolt"
)

const (
	// 10MB is the recommended upload size according to the graph API docs
	uploadChunkSize uint64 = 10 * 1024 * 1024

	// uploads larget than 4MB must use a formal upload session
	uploadLargeSize uint64 = 4 * 1024 * 1024
)

// upload states
const (
	uploadNotStarted = iota
	uploadStarted
	uploadComplete
	uploadErrored
)

// UploadSession contains a snapshot of the file we're uploading. We have to
// take the snapshot or the file may have changed on disk during upload (which
// would break the upload). It is not recommended to directly deserialize into
// this structure from API responses in case Microsoft ever adds a size, data,
// or modTime field to the response.
type UploadSession struct {
	ID                 string    `json:"id"`
	OldID              string    `json:"oldID"`
	ParentID           string    `json:"parentID"`
	NodeID             uint64    `json:"nodeID"`
	Name               string    `json:"name"`
	ExpirationDateTime time.Time `json:"expirationDateTime"`
	Size               uint64    `json:"size,omitempty"`
	Data               []byte    `json:"data,omitempty"`
	QuickXORHash       string    `json:"quickxorhash,omitempty"`
	ModTime            time.Time `json:"modTime,omitempty"`
	retries            int

	// Recovery and progress tracking fields
	LastSuccessfulChunk int       `json:"lastSuccessfulChunk"`
	TotalChunks         int       `json:"totalChunks"`
	BytesUploaded       uint64    `json:"bytesUploaded"`
	LastProgressTime    time.Time `json:"lastProgressTime"`
	RecoveryAttempts    int       `json:"recoveryAttempts"`
	CanResume           bool      `json:"canResume"`

	sync.Mutex
	UploadURL string `json:"uploadUrl"`
	ETag      string `json:"eTag,omitempty"`
	state     int
	error     // embedded error tracks errors that killed an upload
}

// MarshalJSON implements a custom JSON marshaler to avoid race conditions
func (u *UploadSession) MarshalJSON() ([]byte, error) {
	u.Lock()
	defer u.Unlock()

	// Create a struct with the same fields but without the embedded methods
	// to avoid infinite recursion
	type SerializableUploadSession struct {
		ID                 string    `json:"id"`
		OldID              string    `json:"oldID"`
		ParentID           string    `json:"parentID"`
		NodeID             uint64    `json:"nodeID"`
		Name               string    `json:"name"`
		ExpirationDateTime time.Time `json:"expirationDateTime"`
		Size               uint64    `json:"size,omitempty"`
		Data               []byte    `json:"data,omitempty"`
		QuickXORHash       string    `json:"quickxorhash,omitempty"`
		ModTime            time.Time `json:"modTime,omitempty"`

		// Recovery and progress tracking fields
		LastSuccessfulChunk int       `json:"lastSuccessfulChunk"`
		TotalChunks         int       `json:"totalChunks"`
		BytesUploaded       uint64    `json:"bytesUploaded"`
		LastProgressTime    time.Time `json:"lastProgressTime"`
		RecoveryAttempts    int       `json:"recoveryAttempts"`
		CanResume           bool      `json:"canResume"`

		UploadURL string `json:"uploadUrl"`
		ETag      string `json:"eTag,omitempty"`
	}

	return json.Marshal(SerializableUploadSession{
		ID:                  u.ID,
		OldID:               u.OldID,
		ParentID:            u.ParentID,
		NodeID:              u.NodeID,
		Name:                u.Name,
		ExpirationDateTime:  u.ExpirationDateTime,
		Size:                u.Size,
		Data:                u.Data,
		QuickXORHash:        u.QuickXORHash,
		ModTime:             u.ModTime,
		LastSuccessfulChunk: u.LastSuccessfulChunk,
		TotalChunks:         u.TotalChunks,
		BytesUploaded:       u.BytesUploaded,
		LastProgressTime:    u.LastProgressTime,
		RecoveryAttempts:    u.RecoveryAttempts,
		CanResume:           u.CanResume,
		UploadURL:           u.UploadURL,
		ETag:                u.ETag,
	})
}

// UploadSessionPost is the initial post used to create an upload session
type UploadSessionPost struct {
	Name             string `json:"name,omitempty"`
	ConflictBehavior string `json:"@microsoft.graph.conflictBehavior,omitempty"`
	FileSystemInfo   `json:"fileSystemInfo,omitempty"`
}

// FileSystemInfo carries the filesystem metadata like Mtime/Atime
type FileSystemInfo struct {
	LastModifiedDateTime time.Time `json:"lastModifiedDateTime,omitempty"`
}

func (u *UploadSession) getState() int {
	u.Lock()
	defer u.Unlock()
	return u.state
}

// updateProgress updates the upload progress and persists it
func (u *UploadSession) updateProgress(chunkIndex int, bytesUploaded uint64) {
	u.Lock()
	defer u.Unlock()
	u.LastSuccessfulChunk = chunkIndex
	u.BytesUploaded = bytesUploaded
	u.LastProgressTime = time.Now()
}

// persistProgress saves the current upload progress to disk for recovery
func (u *UploadSession) persistProgress(db *bolt.DB) error {
	// Update recovery fields first
	u.Lock()
	u.CanResume = true
	u.LastProgressTime = time.Now()
	sessionID := u.ID // Copy ID while we have the lock
	u.Unlock()

	// Serialize without holding the lock to avoid deadlock with MarshalJSON
	contents, err := json.Marshal(u)
	if err != nil {
		return err
	}

	return db.Batch(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("uploads"))
		if err != nil {
			return err
		}
		return b.Put([]byte(sessionID), contents)
	})
}

// canResumeUpload checks if the upload can be resumed from the last checkpoint
func (u *UploadSession) canResumeUpload() bool {
	u.Lock()
	defer u.Unlock()
	return u.CanResume && u.LastSuccessfulChunk >= 0 && u.UploadURL != ""
}

// getResumeOffset returns the byte offset from which to resume the upload
func (u *UploadSession) getResumeOffset() uint64 {
	u.Lock()
	defer u.Unlock()
	if u.LastSuccessfulChunk < 0 {
		return 0
	}
	return uint64(u.LastSuccessfulChunk+1) * uploadChunkSize
}

// markAsResumable marks the session as resumable and calculates total chunks
func (u *UploadSession) markAsResumable() {
	u.Lock()
	defer u.Unlock()
	u.CanResume = true
	u.TotalChunks = int(math.Ceil(float64(u.Size) / float64(uploadChunkSize)))
	u.LastSuccessfulChunk = -1 // No chunks uploaded yet
}

// setState is just a helper method to set the UploadSession state and make error checking
// a little more straightforwards.
func (u *UploadSession) setState(state int, err error) error {
	u.Lock()
	u.state = state
	u.error = err
	u.Unlock()
	return err
}

// NewUploadSession wraps an upload of a file into an UploadSession struct
// responsible for performing uploads for a file.
func NewUploadSession(inode *Inode, data *[]byte) (*UploadSession, error) {
	if data == nil {
		return nil, errors.NewValidationError("data to upload cannot be nil", nil)
	}

	// create a generic session for all files
	inode.RLock()

	// Initialize ModTime with current time if it's nil
	var modTime time.Time
	if inode.DriveItem.ModTime != nil {
		modTime = *inode.DriveItem.ModTime
	} else {
		modTime = time.Now()
	}

	session := UploadSession{
		ID:       inode.DriveItem.ID,
		OldID:    inode.DriveItem.ID,
		ParentID: inode.DriveItem.Parent.ID,
		NodeID:   inode.nodeID,
		Name:     inode.DriveItem.Name,
		Data:     *data,
		ModTime:  modTime,

		// Initialize recovery fields
		LastSuccessfulChunk: -1,
		TotalChunks:         0,
		BytesUploaded:       0,
		LastProgressTime:    time.Now(),
		RecoveryAttempts:    0,
		CanResume:           false,
	}
	inode.RUnlock()

	// Use the size from the inode if available, otherwise use the data length
	if inode.DriveItem.Size > 0 {
		session.Size = inode.DriveItem.Size
	} else {
		session.Size = uint64(len(*data))
	}
	session.QuickXORHash = graph.QuickXORHash(data)
	return &session, nil
}

// cancel the upload session by deleting the temp file at the endpoint.
func (u *UploadSession) cancel(auth *graph.Auth) {
	u.Lock()
	// small upload sessions will also have an empty UploadURL in addition to
	// uninitialized large file uploads.
	nonemptyURL := u.UploadURL != ""
	u.Unlock()
	if nonemptyURL {
		state := u.getState()
		if state == uploadStarted || state == uploadErrored {
			// dont care about result, this is purely us being polite to the server
			go graph.Delete(u.UploadURL, auth)
		}
	}
}

// Internal method used for uploading individual chunks of a DriveItem. We have
// to make things this way because the internal Put func doesn't work all that
// well when we need to add custom headers. Will return without an error if
// irrespective of HTTP status (errors are reserved for stuff that prevented
// the HTTP request at all).
func (u *UploadSession) uploadChunk(auth *graph.Auth, offset uint64) ([]byte, int, error) {
	u.Lock()
	uploadURL := u.UploadURL
	if uploadURL == "" {
		u.Unlock()
		return nil, -1, errors.NewValidationError("UploadSession UploadURL cannot be empty", nil)
	}
	u.Unlock()

	// how much of the file are we going to upload?
	end := offset + uploadChunkSize
	var reqChunkSize uint64
	if end > u.Size {
		end = u.Size
		reqChunkSize = end - offset + 1
	}
	if offset > u.Size {
		return nil, -1, errors.NewValidationError("offset cannot be larger than DriveItem size", nil)
	}

	auth.Refresh(nil) // nil context will use context.Background() internally

	// Use the configured HTTP client (which may be a mock client for testing)
	client := graph.GetHTTPClient()
	request, _ := http.NewRequest(
		"PUT",
		uploadURL,
		bytes.NewReader((u.Data)[offset:end]),
	)
	// no Authorization header - it will throw a 401 if present
	request.Header.Add("Content-Length", strconv.Itoa(int(reqChunkSize)))
	frags := fmt.Sprintf("bytes %d-%d/%d", offset, end-1, u.Size)
	logging.Info().Str("id", u.ID).Msg("Uploading " + frags)
	request.Header.Add("Content-Range", frags)

	resp, err := client.Do(request)
	if err != nil {
		// this is a serious error, not simply one with a non-200 return code
		return nil, -1, err
	}
	defer resp.Body.Close()
	response, _ := io.ReadAll(resp.Body)
	return response, resp.StatusCode, nil
}

// Upload copies the file's contents to the server. Should only be called as a
// goroutine, or it can potentially block for a very long time. The uploadSession.error
// field contains errors to be handled if called as a goroutine.
func (u *UploadSession) Upload(auth *graph.Auth) error {
	return u.UploadWithContext(context.Background(), auth, nil)
}

// UploadWithContext uploads with context support and optional database for persistence
func (u *UploadSession) UploadWithContext(ctx context.Context, auth *graph.Auth, db *bolt.DB) error {
	logging.Info().Str("id", u.ID).Str("name", u.Name).Msg("Uploading file.")
	u.setState(uploadStarted, nil)

	// Check for context cancellation before starting upload
	select {
	case <-ctx.Done():
		logging.Info().
			Str("id", u.ID).
			Str("name", u.Name).
			Msg("Upload cancelled by context before starting")
		return u.setState(uploadErrored, errors.New("upload cancelled by context"))
	default:
		// Continue with upload
	}

	var uploadPath string
	var resp []byte
	if u.Size < uploadLargeSize {
		// Small upload sessions use a simple PUT request, but this does not support
		// adding file modification times. We don't really care though, because
		// after some experimentation, the Microsoft API doesn't seem to properly
		// support these either (this is why we have to use etags).
		if isLocalID(u.ID) {
			uploadPath = fmt.Sprintf(
				"/me/drive/items/%s:/%s:/content",
				url.PathEscape(u.ParentID),
				url.PathEscape(u.Name),
			)
		} else {
			uploadPath = fmt.Sprintf(
				"/me/drive/items/%s/content",
				url.PathEscape(u.ID),
			)
		}
		// small files handled in this block - use context-aware version
		var err error
		resp, err = graph.PutWithContext(ctx, uploadPath, auth, bytes.NewReader(u.Data))
		if err != nil {
			// Check if the error was due to context cancellation
			if ctx.Err() != nil {
				logging.Info().
					Str("id", u.ID).
					Str("name", u.Name).
					Msg("Small file upload cancelled by context")
				return u.setState(uploadErrored, errors.New("upload cancelled by context"))
			}
			if strings.Contains(err.Error(), "resourceModified") {
				// retry the request after a second, likely the server is having issues
				time.Sleep(time.Second)
				resp, err = graph.PutWithContext(ctx, uploadPath, auth, bytes.NewReader(u.Data))
				// Check for context cancellation after retry
				if err != nil && ctx.Err() != nil {
					logging.Info().
						Str("id", u.ID).
						Str("name", u.Name).
						Msg("Small file upload cancelled by context during retry")
					return u.setState(uploadErrored, errors.New("upload cancelled by context"))
				}
			}
			if err != nil {
				return u.setState(uploadErrored, errors.Wrap(err, "small upload failed"))
			}
		}

		// Update progress for small file uploads
		u.updateProgress(0, u.Size)
	} else {
		// Check if we can resume an existing upload session
		if u.canResumeUpload() {
			logging.Info().
				Str("id", u.ID).
				Str("name", u.Name).
				Int("lastChunk", u.LastSuccessfulChunk).
				Uint64("bytesUploaded", u.BytesUploaded).
				Msg("Resuming upload from last checkpoint")
		} else {
			// Create new upload session
			if isLocalID(u.ID) {
				uploadPath = fmt.Sprintf(
					"/me/drive/items/%s:/%s:/createUploadSession",
					url.PathEscape(u.ParentID),
					url.PathEscape(u.Name),
				)
			} else {
				uploadPath = fmt.Sprintf(
					"/me/drive/items/%s/createUploadSession",
					url.PathEscape(u.ID),
				)
			}
			sessionPostData, _ := json.Marshal(UploadSessionPost{
				ConflictBehavior: "replace",
				FileSystemInfo: FileSystemInfo{
					LastModifiedDateTime: u.ModTime,
				},
			})
			resp, err := graph.Post(uploadPath, auth, bytes.NewReader(sessionPostData))
			if err != nil {
				return u.setState(uploadErrored, errors.Wrap(err, "failed to create upload session"))
			}

			// populate UploadURL/expiration - we unmarshal into a fresh session here
			// just in case the API does something silly at a later date and overwrites
			// a field it shouldn't.
			tmp := UploadSession{}
			if err = json.Unmarshal(resp, &tmp); err != nil {
				return u.setState(uploadErrored,
					errors.Wrap(err, "could not unmarshal upload session post response"))
			}
			u.Lock()
			u.UploadURL = tmp.UploadURL
			u.ExpirationDateTime = tmp.ExpirationDateTime
			u.Unlock()

			// Mark session as resumable for large files
			u.markAsResumable()
		}

		// api upload session created successfully, now do actual content upload
		var status int
		var err error
		nchunks := int(math.Ceil(float64(u.Size) / float64(uploadChunkSize)))

		// Start from the next chunk after the last successful one
		startChunk := 0
		if u.canResumeUpload() {
			startChunk = u.LastSuccessfulChunk + 1
		}

		for i := startChunk; i < nchunks; i++ {
			// Check for context cancellation before each chunk
			select {
			case <-ctx.Done():
				logging.Info().
					Str("id", u.ID).
					Str("name", u.Name).
					Int("chunk", i).
					Int("totalChunks", nchunks).
					Msg("Upload cancelled by context, persisting progress for recovery")

				// Persist current progress before cancelling
				if db != nil {
					u.persistProgress(db)
				}
				return u.setState(uploadErrored, errors.New("upload cancelled by context"))
			default:
				// Continue with upload
			}

			resp, status, err = u.uploadChunk(auth, uint64(i)*uploadChunkSize)
			if err != nil {
				return u.setState(uploadErrored, errors.Wrap(err, "failed to perform chunk upload"))
			}

			// retry server-side failures with an exponential back-off strategy. Will not
			// exit this loop unless it receives a non 5xx error or serious failure
			for backoff := 1; status >= 500; backoff *= 2 {
				// Check for context cancellation during retries
				select {
				case <-ctx.Done():
					if db != nil {
						u.persistProgress(db)
					}
					return u.setState(uploadErrored, errors.New("upload cancelled during retry"))
				default:
				}

				logging.Error().
					Str("id", u.ID).
					Str("name", u.Name).
					Int("chunk", i).
					Int("nchunks", nchunks).
					Int("status", status).
					Msgf("The OneDrive server is having issues, retrying chunk upload in %ds.", backoff)
				time.Sleep(time.Duration(backoff) * time.Second)
				resp, status, err = u.uploadChunk(auth, uint64(i)*uploadChunkSize)
				if err != nil { // a serious, non 4xx/5xx error
					return u.setState(uploadErrored, errors.Wrap(err, "failed to perform chunk upload"))
				}
			}

			// Update progress after successful chunk upload
			if status < 400 {
				bytesUploaded := uint64(i+1) * uploadChunkSize
				if bytesUploaded > u.Size {
					bytesUploaded = u.Size
				}
				u.updateProgress(i, bytesUploaded)

				// Persist progress every 10 chunks or for large files
				if db != nil && (i%10 == 0 || u.Size > 100*1024*1024) {
					if err := u.persistProgress(db); err != nil {
						logging.Warn().
							Str("id", u.ID).
							Err(err).
							Msg("Failed to persist upload progress")
					}
				}

				logging.Debug().
					Str("id", u.ID).
					Int("chunk", i).
					Int("totalChunks", nchunks).
					Uint64("bytesUploaded", bytesUploaded).
					Msg("Chunk uploaded successfully")
			}

			// handle client-side errors
			if status >= 400 {
				return u.setState(uploadErrored, errors.NewOperationError(fmt.Sprintf("error uploading chunk - HTTP %d: %s", status, string(resp)), nil))
			}
		}
	}

	// server has indicated that the upload was successful - now we check to verify the
	// checksum is what it's supposed to be.
	remote := graph.DriveItem{}
	if err := json.Unmarshal(resp, &remote); err != nil {
		if len(resp) == 0 {
			// the API frequently just returns a 0-byte response for completed
			// multipart uploads, so we manually fetch the newly updated item
			var remotePtr *graph.DriveItem
			if isLocalID(u.ID) {
				remotePtr, err = graph.GetItemChild(u.ParentID, u.Name, auth)
			} else {
				remotePtr, err = graph.GetItem(u.ID, auth)
			}
			if err == nil {
				remote = *remotePtr
			} else {
				return u.setState(uploadErrored,
					errors.Wrap(err, "failed to get item post-upload"))
			}
		} else {
			return u.setState(uploadErrored,
				errors.Wrap(err, fmt.Sprintf("could not unmarshal response: %s", string(resp))),
			)
		}
	}
	if remote.File == nil && remote.Size != u.Size {
		// if we are absolutely pounding the microsoft API, a remote item may sometimes
		// come back without checksums, so we check the size of the uploaded item instead.
		return u.setState(uploadErrored, errors.NewValidationError("size mismatch when remote checksums did not exist", nil))
	} else if !remote.VerifyChecksum(u.QuickXORHash) {
		return u.setState(uploadErrored, errors.NewValidationError("remote checksum did not match", nil))
	}
	// update the UploadSession's ID, ETag, and Size in the event that we exchange a local for a remote ID
	u.Lock()
	u.ID = remote.ID
	u.ETag = remote.ETag
	u.Size = remote.Size
	u.Unlock()
	return u.setState(uploadComplete, nil)
}

// GetID returns the ID of the item being uploaded
func (u *UploadSession) GetID() string {
	u.Lock()
	defer u.Unlock()
	return u.ID
}

// GetName returns the name of the item being uploaded
func (u *UploadSession) GetName() string {
	u.Lock()
	defer u.Unlock()
	return u.Name
}

// GetSize returns the size of the item being uploaded
func (u *UploadSession) GetSize() uint64 {
	u.Lock()
	defer u.Unlock()
	return u.Size
}

// GetState returns the current state of the upload
func (u *UploadSession) GetState() int {
	return u.getState()
}
