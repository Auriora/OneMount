package fs

// The upload_manager.go file implements a background upload manager for OneDrive files.
// This helps decouple the local file system logic from the OneDrive sync logic by running
// file uploads in separate worker threads. This improves performance by handling the
// OneDrive cloud file sync in the background.

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/auriora/onemount/internal/fs/graph"
	"github.com/rs/zerolog/log"
	bolt "go.etcd.io/bbolt"
)

const maxUploadsInFlight = 5

var bucketUploads = []byte("uploads")

// UploadState represents the state of an upload
type UploadState int

const (
	// uploadNotStarted indicates the upload is queued but not started
	uploadNotStartedState UploadState = iota
	// uploadStarted indicates the upload is in progress
	uploadStartedState
	// uploadCompleted indicates the upload completed successfully
	uploadCompletedState
	// uploadErrored indicates the upload failed
	uploadErroredState
)

// UploadManager is used to manage and retry uploads.
type UploadManager struct {
	highPriorityQueue chan *UploadSession
	lowPriorityQueue  chan *UploadSession
	queue             chan *UploadSession // Legacy queue for backward compatibility
	deletionQueue     chan string
	sessions          map[string]*UploadSession
	sessionPriorities map[string]UploadPriority // Track priority of each session
	inFlight          uint8                     // number of sessions in flight
	auth              *graph.Auth
	fs                *Filesystem
	db                *bolt.DB
	mutex             sync.RWMutex
	stopChan          chan struct{}
	workerWg          sync.WaitGroup
}

// UploadPriority defines the priority level for uploads
type UploadPriority int

const (
	// PriorityLow is for background tasks
	PriorityLow UploadPriority = iota
	// PriorityHigh is for mount point requests
	PriorityHigh
)

// NewUploadManager creates a new queue/thread for uploads
func NewUploadManager(duration time.Duration, db *bolt.DB, fs *Filesystem, auth *graph.Auth) *UploadManager {
	manager := UploadManager{
		highPriorityQueue: make(chan *UploadSession),
		lowPriorityQueue:  make(chan *UploadSession),
		queue:             make(chan *UploadSession), // Legacy queue for backward compatibility
		deletionQueue:     make(chan string, 1000),   // FIXME - why does this chan need to be buffered now???
		sessions:          make(map[string]*UploadSession),
		sessionPriorities: make(map[string]UploadPriority),
		auth:              auth,
		db:                db,
		fs:                fs,
		stopChan:          make(chan struct{}),
	}
	db.View(func(tx *bolt.Tx) error {
		// Add any incomplete sessions from disk - any sessions here were never
		// finished. The most likely cause of this is that the user shut off
		// their computer or closed the program after starting the upload.
		b := tx.Bucket(bucketUploads)
		if b == nil {
			// bucket does not exist yet, bail out early
			return nil
		}
		return b.ForEach(func(key []byte, val []byte) error {
			session := &UploadSession{}
			err := json.Unmarshal(val, session)
			if err != nil {
				log.Error().Err(err).Msg("Failure restoring upload sessions from disk.")
				return err
			}
			if session.getState() != uploadNotStarted {
				manager.inFlight++
			}
			session.cancel(auth) // uploads are currently non-resumable
			manager.sessions[session.ID] = session
			return nil
		})
	})

	// Add the uploadLoop goroutine to the wait group
	manager.workerWg.Add(1)
	go manager.uploadLoop(duration)

	return &manager
}

// uploadLoop manages the deduplication and tracking of uploads
func (u *UploadManager) uploadLoop(duration time.Duration) {
	defer u.workerWg.Done()

	ticker := time.NewTicker(duration)
	defer ticker.Stop()

	for {
		select {
		case session := <-u.highPriorityQueue: // high priority sessions
			// deduplicate sessions for the same item
			u.mutex.Lock()
			if old, exists := u.sessions[session.ID]; exists {
				old.cancel(u.auth)
			}
			contents, _ := json.Marshal(session)
			u.db.Batch(func(tx *bolt.Tx) error {
				// persist to disk in case the user shuts off their computer or
				// kills onemount prematurely
				b, _ := tx.CreateBucketIfNotExists(bucketUploads)
				return b.Put([]byte(session.ID), contents)
			})
			u.sessions[session.ID] = session
			u.sessionPriorities[session.ID] = PriorityHigh
			u.mutex.Unlock()

		case session := <-u.lowPriorityQueue: // low priority sessions
			// deduplicate sessions for the same item
			u.mutex.Lock()
			if old, exists := u.sessions[session.ID]; exists {
				old.cancel(u.auth)
			}
			contents, _ := json.Marshal(session)
			u.db.Batch(func(tx *bolt.Tx) error {
				// persist to disk in case the user shuts off their computer or
				// kills onemount prematurely
				b, _ := tx.CreateBucketIfNotExists(bucketUploads)
				return b.Put([]byte(session.ID), contents)
			})
			u.sessions[session.ID] = session
			u.sessionPriorities[session.ID] = PriorityLow
			u.mutex.Unlock()

		case session := <-u.queue: // legacy queue for backward compatibility
			// deduplicate sessions for the same item
			u.mutex.Lock()
			if old, exists := u.sessions[session.ID]; exists {
				old.cancel(u.auth)
			}
			contents, _ := json.Marshal(session)
			u.db.Batch(func(tx *bolt.Tx) error {
				// persist to disk in case the user shuts off their computer or
				// kills onemount prematurely
				b, _ := tx.CreateBucketIfNotExists(bucketUploads)
				return b.Put([]byte(session.ID), contents)
			})
			u.sessions[session.ID] = session
			u.sessionPriorities[session.ID] = PriorityLow // Default to low priority for legacy queue
			u.mutex.Unlock()

		case cancelID := <-u.deletionQueue: // remove uploads for deleted items
			u.finishUpload(cancelID)

		case <-ticker.C: // periodically start uploads, or remove them if done/failed
			u.mutex.RLock()
			sessionsCopy := make(map[string]*UploadSession)
			prioritiesCopy := make(map[string]UploadPriority)
			for id, session := range u.sessions {
				sessionsCopy[id] = session
				priority, exists := u.sessionPriorities[id]
				if !exists {
					priority = PriorityLow // Default to low priority if not specified
				}
				prioritiesCopy[id] = priority
			}
			u.mutex.RUnlock()

			// Sort sessions by priority (high priority first)
			type sessionWithPriority struct {
				id       string
				session  *UploadSession
				priority UploadPriority
			}
			prioritizedSessions := make([]sessionWithPriority, 0, len(sessionsCopy))
			for id, session := range sessionsCopy {
				priority := prioritiesCopy[id]
				prioritizedSessions = append(prioritizedSessions, sessionWithPriority{
					id:       id,
					session:  session,
					priority: priority,
				})
			}
			// Sort by priority (high priority first)
			sort.Slice(prioritizedSessions, func(i, j int) bool {
				return prioritizedSessions[i].priority > prioritizedSessions[j].priority
			})

			for _, s := range prioritizedSessions {
				id := s.id
				session := s.session
				switch session.getState() {
				case uploadNotStarted:
					// max active upload sessions are capped at this limit for faster
					// uploads of individual files and also to prevent possible server-
					// side throttling that can cause errors.
					if u.inFlight < maxUploadsInFlight {
						u.inFlight++
						// Update status to syncing
						u.fs.SetFileStatus(id, FileStatusInfo{
							Status:    StatusSyncing,
							Timestamp: time.Now(),
						})
						go session.Upload(u.auth)
					}

				case uploadErrored:
					session.retries++
					if session.retries > 5 {
						log.Error().
							Str("id", session.ID).
							Str("name", session.Name).
							Err(session).
							Int("retries", session.retries).
							Msg("Upload session failed too many times, cancelling session.")
						// Update status to error
						u.fs.MarkFileError(session.ID, session.error)
						u.finishUpload(session.ID)
					}

					log.Warn().
						Str("id", session.ID).
						Str("name", session.Name).
						Err(session).
						Msg("Upload session failed, will retry from beginning.")
					session.cancel(u.auth) // cancel large sessions
					session.setState(uploadNotStarted, nil)

				case uploadComplete:
					log.Info().
						Str("id", session.ID).
						Str("oldID", session.OldID).
						Str("name", session.Name).
						Msg("Upload completed!")

					// ID changed during upload, move to new ID
					if session.OldID != session.ID {
						err := u.fs.MoveID(session.OldID, session.ID)
						if err != nil {
							log.Error().
								Str("id", session.ID).
								Str("oldID", session.OldID).
								Str("name", session.Name).
								Err(err).
								Msg("Could not move inode to new ID!")
						}
					}

					// inode will exist at the new ID now, but we check if inode
					// is nil to see if the item has been deleted since upload start
					if inode := u.fs.GetID(session.ID); inode != nil {
						inode.Lock()
						inode.DriveItem.ETag = session.ETag
						inode.Unlock()

						// Update status to local (synced)
						u.fs.SetFileStatus(session.ID, FileStatusInfo{
							Status:    StatusLocal,
							Timestamp: time.Now(),
						})

						// Update file status attributes
						u.fs.updateFileStatus(inode)
					}

					// the old ID is the one that was used to add it to the queue.
					// cleanup the session.
					u.finishUpload(session.OldID)
				}
			}

		case <-u.stopChan:
			// Stop the upload loop
			return
		}
	}
}

// QueueUpload queues an item for upload with the specified priority.
// If no priority is specified (using QueueUploadWithPriority), it defaults to low priority.
func (u *UploadManager) QueueUpload(inode *Inode) (*UploadSession, error) {
	return u.QueueUploadWithPriority(inode, PriorityLow)
}

// QueueUploadWithPriority queues an item for upload with the specified priority.
func (u *UploadManager) QueueUploadWithPriority(inode *Inode, priority UploadPriority) (*UploadSession, error) {
	data := u.fs.getInodeContent(inode)
	session, err := NewUploadSession(inode, data)
	if err != nil {
		return nil, err
	}

	// Check if there's already an upload session for this ID
	u.mutex.RLock()
	existingSession, exists := u.sessions[session.ID]
	existingPriority := PriorityLow
	if exists {
		existingPriority, _ = u.sessionPriorities[session.ID]
	}
	u.mutex.RUnlock()

	if exists {
		// If the existing session has lower priority than the requested priority,
		// update its priority
		if existingPriority < priority {
			u.mutex.Lock()
			u.sessionPriorities[session.ID] = priority
			u.mutex.Unlock()
		}
		// Return the existing session
		return existingSession, nil
	}

	if u.fs.IsOffline() {
		// If offline, store the session for later but don't start upload
		contents, _ := json.Marshal(session)
		u.db.Batch(func(tx *bolt.Tx) error {
			b, _ := tx.CreateBucketIfNotExists(bucketUploads)
			return b.Put([]byte(session.ID), contents)
		})

		log.Info().
			Str("id", session.ID).
			Str("name", session.Name).
			Str("priority", priorityToString(priority)).
			Msg("Queued upload for when connectivity is restored.")

		// Store the session in memory too
		u.mutex.Lock()
		u.sessions[session.ID] = session
		u.sessionPriorities[session.ID] = priority
		u.mutex.Unlock()

		return session, nil
	}

	// Normal online behavior
	var targetQueue chan *UploadSession
	if priority == PriorityHigh {
		targetQueue = u.highPriorityQueue
	} else {
		targetQueue = u.lowPriorityQueue
	}

	select {
	case targetQueue <- session:
		log.Info().
			Str("id", session.ID).
			Str("name", session.Name).
			Str("priority", priorityToString(priority)).
			Msg("File queued for upload")
		return session, nil
	default:
		// Queue is full, return error
		return nil, errors.New("upload queue is full")
	}
}

// priorityToString converts an UploadPriority to a string for logging
func priorityToString(priority UploadPriority) string {
	if priority == PriorityHigh {
		return "high"
	}
	return "low"
}

// CancelUpload is used to kill any pending uploads for a session
func (u *UploadManager) CancelUpload(id string) {
	u.deletionQueue <- id
}

// finishUpload is an internal method that gets called when a session is
// completed. It cancels the session if one was in progress, and then deletes
// it from both memory and disk.
func (u *UploadManager) finishUpload(id string) {
	u.mutex.Lock()
	defer u.mutex.Unlock()

	if session, exists := u.sessions[id]; exists {
		session.cancel(u.auth)
	}
	u.db.Batch(func(tx *bolt.Tx) error {
		if b := tx.Bucket(bucketUploads); b != nil {
			b.Delete([]byte(id))
		}
		return nil
	})
	if u.inFlight > 0 {
		u.inFlight--
	}
	delete(u.sessions, id)
	delete(u.sessionPriorities, id) // Also remove from sessionPriorities map
}

// GetSession returns the upload session with the given ID
func (u *UploadManager) GetSession(id string) (*UploadSession, bool) {
	u.mutex.RLock()
	defer u.mutex.RUnlock()
	session, exists := u.sessions[id]
	return session, exists
}

// GetUploadStatus returns the status of an upload
func (u *UploadManager) GetUploadStatus(id string) (UploadState, error) {
	session, exists := u.GetSession(id)
	if !exists {
		return 0, errors.New("upload session not found")
	}

	state := session.getState()
	switch state {
	case uploadNotStarted:
		return uploadNotStartedState, nil
	case uploadStarted:
		return uploadStartedState, nil
	case uploadComplete:
		return uploadCompletedState, nil
	case uploadErrored:
		return uploadErroredState, nil
	default:
		return 0, errors.New("unknown upload state")
	}
}

// WaitForUpload waits for an upload to complete
func (u *UploadManager) WaitForUpload(id string) error {
	// Maximum time to wait for a session to be created
	const sessionCreationTimeout = 5 * time.Second
	sessionCreationDeadline := time.Now().Add(sessionCreationTimeout)

	for {
		session, exists := u.GetSession(id)

		if !exists {
			// If the session doesn't exist yet, wait for it to be created
			if time.Now().After(sessionCreationDeadline) {
				return fmt.Errorf("upload session not found after waiting %v: id=%s", sessionCreationTimeout, id)
			}
			// Wait a bit and check again
			time.Sleep(50 * time.Millisecond)
			continue
		}

		state := session.getState()
		switch state {
		case uploadComplete:
			return nil
		case uploadErrored:
			return session.error
		default:
			// Still in progress, wait a bit
			time.Sleep(100 * time.Millisecond)
		}
	}
}

// Stop stops the upload manager and waits for all uploads to finish
func (u *UploadManager) Stop() {
	log.Info().Msg("Stopping upload manager...")
	close(u.stopChan)

	// Wait for all workers to finish with a timeout
	done := make(chan struct{})
	go func() {
		u.workerWg.Wait()
		close(done)
	}()

	// Wait for workers to finish or timeout after 5 seconds
	select {
	case <-done:
		log.Info().Msg("Upload manager stopped successfully")
	case <-time.After(5 * time.Second):
		log.Warn().Msg("Timed out waiting for upload manager to stop")
	}
}
