package fs

import (
	"context"
	"errors"
	"time"

	"github.com/auriora/onemount/internal/graph"
	"github.com/auriora/onemount/internal/logging"
	"github.com/auriora/onemount/internal/metadata"
)

const (
	defaultMutationWorkers  = 2
	defaultMutationQueueLen = 128
)

type mutationJob struct {
	name string
	id   string
	work func() error
}

// startMutationQueue spins up bounded workers so FUSE entry points never block on Graph mutations.
func (f *Filesystem) startMutationQueue() {
	if f.mutationQueue != nil {
		return
	}
	f.mutationQueue = make(chan mutationJob, defaultMutationQueueLen)
	f.mutationQueueStop = make(chan struct{})
	for i := 0; i < defaultMutationWorkers; i++ {
		f.Wg.Add(1)
		go func() {
			defer f.Wg.Done()
			for {
				select {
				case <-f.ctx.Done():
					return
				case <-f.mutationQueueStop:
					return
				case job, ok := <-f.mutationQueue:
					if !ok {
						return
					}
					f.runMutation(job.name, job.id, job.work)
				}
			}
		}()
	}
}

func (f *Filesystem) stopMutationQueue() {
	f.mutationStopOnce.Do(func() {
		if f.mutationQueueStop != nil {
			close(f.mutationQueueStop)
		}
	})
}

func (f *Filesystem) queueRemoteDirCreate(parentID, tempID, name string) {
	if parentID == "" || tempID == "" || name == "" {
		return
	}
	if f.auth == nil {
		logging.Warn().
			Str("parentID", parentID).
			Str("tempID", tempID).
			Msg("Skipped remote directory create because auth is nil")
		return
	}

	work := func() error {
		item, err := graph.Mkdir(name, parentID, f.auth)
		if err != nil {
			return err
		}
		if item.ModTime == nil {
			ts := time.Now()
			item.ModTime = &ts
		}
		return f.promoteTempInode(tempID, item)
	}

	f.runMutationWithRetry("mkdir", tempID, work)
}

func (f *Filesystem) queueRemoteDelete(id string) {
	if id == "" || isLocalID(id) || f.auth == nil {
		return
	}
	f.runMutationWithRetry("delete", id, func() error {
		if err := graph.Remove(id, f.auth); err != nil {
			return err
		}
		f.clearChildPendingRemote(id)
		f.transitionItemState(id, metadata.ItemStateDeleted)
		return nil
	})
}

// queueRemoteDeleteTestHook runs a remote delete synchronously for tests, ensuring state transition to DELETED.
func (f *Filesystem) queueRemoteDeleteTestHook(id string) error {
	if id == "" || isLocalID(id) {
		return nil
	}
	// Skip network call in test hook; assume success
	f.clearChildPendingRemote(id)
	// Use DELETED_LOCAL to preserve local tombstone semantics for sync.
	if f.metadataStore != nil {
		_, _ = f.metadataStore.Update(context.Background(), id, func(entry *metadata.Entry) error {
			if entry == nil {
				return metadata.ErrNotFound
			}
			entry.State = metadata.ItemStateDeleted
			entry.Children = nil
			entry.SubdirCount = 0
			return nil
		})
	}
	f.transitionToState(id, metadata.ItemStateDeleted)
	return nil
}

func (f *Filesystem) runMutationWithRetry(operation, id string, fn func() error) {
	job := mutationJob{name: operation, id: id, work: fn}
	if f.mutationQueue == nil {
		// Fallback for tests or minimal constructors.
		f.Wg.Add(1)
		go func() {
			defer f.Wg.Done()
			f.runMutation(operation, id, fn)
		}()
		return
	}

	select {
	case f.mutationQueue <- job:
		return
	default:
		logging.Warn().
			Str("mutation", operation).
			Str("id", id).
			Msg("Mutation queue full; falling back to detached retry loop")
		f.Wg.Add(1)
		go func() {
			defer f.Wg.Done()
			f.runMutation(operation, id, fn)
		}()
	}
}

func (f *Filesystem) runMutation(operation, id string, fn func() error) {
	const maxAttempts = 5
	baseDelay := 200 * time.Millisecond
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if f.ctx != nil {
			select {
			case <-f.ctx.Done():
				return
			default:
			}
		}
		if err := fn(); err != nil {
			logging.Warn().
				Str("mutation", operation).
				Str("id", id).
				Int("attempt", attempt).
				Err(err).
				Msg("Async mutation failed")
			if attempt == maxAttempts {
				return
			}
			time.Sleep(baseDelay * time.Duration(attempt))
			continue
		}
		return
	}
}

func (f *Filesystem) queueRemoteRename(remoteID, newParentID, newName string) {
	if remoteID == "" || newName == "" || f.auth == nil || isLocalID(remoteID) {
		return
	}
	f.runMutationWithRetry("rename", remoteID, func() error {
		if err := graph.Rename(remoteID, newName, newParentID, f.auth); err != nil {
			return err
		}
		f.markHydratedState(remoteID)
		return nil
	})
}

func (f *Filesystem) promoteTempInode(tempID string, remoteItem *graph.DriveItem) error {
	if tempID == "" || remoteItem == nil || remoteItem.ID == "" {
		return errors.New("invalid promotion input")
	}
	inode := f.GetID(tempID)
	if inode == nil {
		return errors.New("temporary inode missing")
	}
	if err := f.MoveID(tempID, remoteItem.ID); err != nil {
		return err
	}

	promoted := f.GetID(remoteItem.ID)
	if promoted == nil {
		return errors.New("promoted inode missing")
	}

	promoted.mu.Lock()
	promoted.DriveItem.ID = remoteItem.ID
	promoted.DriveItem.Name = remoteItem.Name
	promoted.DriveItem.Parent = remoteItem.Parent
	promoted.DriveItem.ETag = remoteItem.ETag
	promoted.DriveItem.Size = remoteItem.Size
	promoted.DriveItem.ModTime = remoteItem.ModTime
	promoted.DriveItem.Folder = remoteItem.Folder
	promoted.DriveItem.File = remoteItem.File
	promoted.mu.Unlock()

	f.clearChildPendingRemote(remoteItem.ID)
	f.markHydratedState(remoteItem.ID)
	f.persistMetadataEntry(remoteItem.ID, promoted)

	if f.metadataStore != nil && tempID != remoteItem.ID {
		_, _ = f.metadataStore.Update(context.Background(), tempID, func(entry *metadata.Entry) error {
			if entry == nil {
				return metadata.ErrNotFound
			}
			entry.State = metadata.ItemStateDeleted
			entry.Children = nil
			entry.SubdirCount = 0
			entry.PendingRemote = false
			return nil
		})
	}
	return nil
}
