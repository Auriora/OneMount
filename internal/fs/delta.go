package fs

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/bcherrington/onemount/internal/fs/graph"
	"github.com/rs/zerolog/log"
	bolt "go.etcd.io/bbolt"
)

// DeltaLoop creates a new thread to poll the server for changes and should be
// called as a goroutine
func (f *Filesystem) DeltaLoop(interval time.Duration) {
	log.Info().Msg("Starting delta goroutine.")

	subsc := newSubscription(f.subscribeChanges)
	go subsc.Start()
	defer subsc.Stop()

	// Add to wait group to track this goroutine
	f.deltaLoopWg.Add(1)
	defer func() {
		log.Debug().Msg("Delta goroutine exiting, calling Done() on wait group")
		f.deltaLoopWg.Done()
		log.Debug().Msg("Delta goroutine completed")
	}()

	// Create a ticker for the interval
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Create a shorter ticker for offline mode
	offlineTicker := time.NewTicker(2 * time.Second)
	defer offlineTicker.Stop()

	// Use the normal ticker by default
	currentTicker := ticker

	for { // eva
		// Check if we should stop before starting a new cycle
		select {
		case <-f.deltaLoopStop:
			log.Info().Msg("Stopping delta goroutine.")
			log.Debug().Msg("Delta goroutine received stop signal, exiting loop")
			return
		default:
			// Continue with normal operation
		}

		// get deltas
		log.Debug().Msg("Starting delta fetch cycle")
		log.Trace().Msg("Fetching deltas from server.")
		pollSuccess := false
		deltas := make(map[string]*graph.DriveItem)

		// Use a timeout for the entire delta fetch cycle
		fetchCtx, fetchCancel := context.WithTimeout(f.deltaLoopCtx, 2*time.Minute)

	fetchLoop:
		for {
			// Check if we should stop before making network call
			select {
			case <-f.deltaLoopStop:
				log.Debug().Msg("Delta loop stop signal received during polling loop")
				fetchCancel()
				return
			case <-fetchCtx.Done():
				log.Debug().Msg("Delta fetch cycle timed out")
				break fetchLoop
			default:
				// Continue with normal operation
			}

			log.Debug().Msg("Calling pollDeltas to fetch changes from server")
			incoming, cont, err := f.pollDeltas(f.auth)
			log.Debug().Bool("continue", cont).Int("incomingCount", len(incoming)).Msg("pollDeltas returned")

			// Check again if we should stop after network call
			select {
			case <-f.deltaLoopStop:
				log.Debug().Msg("Delta loop stop signal received after pollDeltas")
				fetchCancel()
				return
			default:
				// Continue processing
			}

			if err != nil {
				// Check if it's a context cancellation error
				if fetchCtx.Err() != nil || f.deltaLoopCtx.Err() != nil {
					log.Debug().Msg("Delta fetch was cancelled by context")
					fetchCancel()
					return
				}

				// the only thing that should be able to bring the FS out
				// of a read-only state is a successful delta call
				log.Error().Err(err).
					Msg("Error during delta fetch, marking fs as offline.")
				f.Lock()
				f.offline = true
				f.Unlock()
				break
			}

			for _, delta := range incoming {
				// As per the API docs, the last delta received from the server
				// for an item is the one we should use.
				deltas[delta.ID] = delta
			}
			if !cont {
				log.Info().Msgf("Fetched %d deltas.", len(deltas))
				pollSuccess = true
				break
			}
			log.Debug().Msg("Need to continue polling for more deltas")
		}

		// Clean up the fetch context
		fetchCancel()

		// Check if we should stop before applying deltas
		select {
		case <-f.deltaLoopStop:
			log.Debug().Msg("Delta loop stop signal received before applying deltas")
			return
		default:
			// Continue with normal operation
		}

		// now apply deltas
		log.Debug().Int("deltaCount", len(deltas)).Msg("Starting to apply deltas")
		secondPass := make([]string, 0)

		// Use a timeout for delta application
		applyCtx, applyCancel := context.WithTimeout(f.deltaLoopCtx, 1*time.Minute)

	applyLoop:
		for _, delta := range deltas {
			// Check if we should stop before applying delta
			select {
			case <-f.deltaLoopStop:
				log.Debug().Msg("Delta loop stop signal received during delta application")
				applyCancel()
				return
			case <-applyCtx.Done():
				log.Debug().Msg("Delta application timed out")
				break applyLoop
			default:
				// Continue with normal operation
			}

			err := f.applyDelta(delta)
			// retry deletion of non-empty directories after all other deltas applied
			if err != nil && err.Error() == "directory is non-empty" {
				secondPass = append(secondPass, delta.ID)
			}
		}

		waitDur := interval

		// Check if we should stop before second pass
		select {
		case <-f.deltaLoopStop:
			log.Debug().Msg("Delta loop stop signal received before second pass")
			applyCancel()
			return
		case <-applyCtx.Done():
			log.Debug().Msg("Delta application timed out before second pass")
			applyCancel()
			// Continue to next cycle
			goto nextCycle
		default:
			// Continue with normal operation
		}

		log.Debug().Int("secondPassCount", len(secondPass)).Msg("Starting second pass for non-empty directories")
		for _, id := range secondPass {
			// Check if we should stop before processing each item in second pass
			select {
			case <-f.deltaLoopStop:
				log.Debug().Msg("Delta loop stop signal received during second pass")
				applyCancel()
				return
			case <-applyCtx.Done():
				log.Debug().Msg("Second pass timed out")
				break
			default:
				// Continue with normal operation
			}

			// failures should explicitly be ignored the second time around as per docs
			if err := f.applyDelta(deltas[id]); err != nil {
				log.Debug().Err(err).Str("id", id).Msg("Ignoring error in second pass delta application")
			}
		}

		// Clean up the apply context
		applyCancel()

		log.Debug().Msg("Finished applying deltas")

		// Check if we should stop before serialization
		select {
		case <-f.deltaLoopStop:
			log.Debug().Msg("Delta loop stop signal received before serialization")
			return
		default:
			// Continue with normal operation
		}

		if !f.IsOffline() {
			log.Debug().Msg("Serializing filesystem state")
			f.SerializeAll()
		}

		if pollSuccess {
			f.Lock()
			wasOffline := f.offline
			if f.offline {
				log.Info().Msg("Delta fetch success, marking fs as online.")
			}
			f.offline = false
			f.Unlock()

			// Switch to normal ticker if we were using offline ticker
			if currentTicker == offlineTicker {
				currentTicker = ticker
			}

			log.Debug().Msg("Saving delta link to database")
			if err := f.db.Batch(func(tx *bolt.Tx) error {
				return tx.Bucket(bucketDelta).Put([]byte("deltaLink"), []byte(f.deltaLink))
			}); err != nil {
				log.Error().Err(err).Msg("Failed to save delta link to database")
			}

			// If we were offline and now we're online, process offline changes
			if wasOffline {
				log.Info().Msg("Transitioning from offline to online, processing offline changes")
				// Use a goroutine with proper error handling
				go func() {
					defer func() {
						if r := recover(); r != nil {
							log.Error().Interface("recover", r).Msg("Panic in ProcessOfflineChanges")
						}
					}()
					f.ProcessOfflineChanges()
				}()
			}
		} else {
			// Switch to offline ticker for shorter retry intervals
			if currentTicker == ticker {
				currentTicker = offlineTicker
			}
			waitDur = 2 * time.Second
		}

	nextCycle:
		// Wait for next interval or stop signal
		select {
		case <-time.After(waitDur):
		case <-subsc.C:
		case <-currentTicker.C:
			// Time to run the next cycle
			log.Debug().Msg("Ticker triggered, starting next delta cycle")
		case <-f.deltaLoopStop:
			log.Info().Msg("Stopping delta goroutine during wait interval.")
			return
		}
	}
}

type deltaResponse struct {
	NextLink  string             `json:"@odata.nextLink,omitempty"`
	DeltaLink string             `json:"@odata.deltaLink,omitempty"`
	Values    []*graph.DriveItem `json:"value,omitempty"`
}

// Polls the delta endpoint and return deltas + whether or not to continue
// polling. Does not perform deduplication. Note that changes from the local
// client will actually appear as deltas from the server (there is no
// distinction between local and remote changes from the server's perspective,
// everything is a delta, regardless of where it came from).
func (f *Filesystem) pollDeltas(auth *graph.Auth) ([]*graph.DriveItem, bool, error) {
	log.Debug().Str("deltaLink", f.deltaLink).Msg("Making network request to delta endpoint")

	// Check if context is already cancelled before making request
	if f.deltaLoopCtx.Err() != nil {
		log.Debug().Err(f.deltaLoopCtx.Err()).Msg("Context already cancelled before making delta request")
		return make([]*graph.DriveItem, 0), false, f.deltaLoopCtx.Err()
	}

	// Create a timeout context that's a child of the main context
	// This ensures we don't get stuck in network calls indefinitely
	ctx, cancel := context.WithTimeout(f.deltaLoopCtx, 30*time.Second)
	defer cancel()

	// Make the network request with context that can be cancelled during shutdown
	log.Debug().Msg("Starting delta request with cancellable context")
	resp, err := graph.GetWithContext(ctx, f.deltaLink, auth)
	log.Debug().Msg("Delta request completed or cancelled")

	// Check for context cancellation first
	if ctx.Err() != nil {
		log.Debug().Err(ctx.Err()).Msg("Delta request context was cancelled")
		// Check if it was our parent context or just the timeout
		if f.deltaLoopCtx.Err() != nil {
			log.Debug().Msg("Parent context was cancelled, stopping delta loop")
			return make([]*graph.DriveItem, 0), false, f.deltaLoopCtx.Err()
		}
		log.Debug().Msg("Request timed out, will retry on next cycle")
		return make([]*graph.DriveItem, 0), false, ctx.Err()
	}

	if err != nil {
		log.Error().Err(err).Msg("Error fetching deltas from server")
		return make([]*graph.DriveItem, 0), false, err
	}

	log.Debug().Int("responseSize", len(resp)).Msg("Received response from delta endpoint")

	page := deltaResponse{}
	err = json.Unmarshal(resp, &page)
	if err != nil {
		log.Error().Err(err).Msg("Error unmarshaling delta response")
		return make([]*graph.DriveItem, 0), false, err
	}

	// If the server does not provide a `@odata.nextLink` item, it means we've
	// reached the end of this polling cycle and should not continue until the
	// next poll interval.
	if page.NextLink != "" {
		newLink := strings.TrimPrefix(page.NextLink, graph.GraphURL)
		log.Debug().Str("oldLink", f.deltaLink).Str("newLink", newLink).Bool("continue", true).Int("itemCount", len(page.Values)).Msg("Delta page has nextLink, continuing")
		f.deltaLink = newLink
		return page.Values, true, nil
	}

	newLink := strings.TrimPrefix(page.DeltaLink, graph.GraphURL)
	log.Debug().Str("oldLink", f.deltaLink).Str("newLink", newLink).Bool("continue", false).Int("itemCount", len(page.Values)).Msg("Delta page has deltaLink, finished polling")
	f.deltaLink = newLink
	return page.Values, false, nil
}

// applyDelta diagnoses and applies a server-side change to our local state.
// Things we care about (present in the local cache):
// * Deleted items
// * Changed content remotely, but not locally
// * New items in a folder we have locally
func (f *Filesystem) applyDelta(delta *graph.DriveItem) error {
	id := delta.ID
	name := delta.Name
	parentID := delta.Parent.ID
	ctx := log.With().
		Str("id", id).
		Str("parentID", parentID).
		Str("name", name).
		Logger()
	ctx.Debug().Msg("Applying delta")

	// diagnose and act on what type of delta we're dealing with

	// do we have it at all?
	parent := f.GetID(parentID)
	if parent == nil {
		// Nothing needs to be applied, item not in cache, so latest copy will
		// be pulled down next time it's accessed.
		ctx.Debug().
			Str("delta", "skip").
			Msg("Skipping delta, item's parent not in cache.")
		return nil
	}

	ctx.Debug().
		Str("parentPath", parent.Path()).
		Msg("Found parent in cache")

	local := f.GetID(id)
	if local != nil {
		ctx.Debug().
			Str("localPath", local.Path()).
			Msg("Found item in cache")
	} else {
		ctx.Debug().Msg("Item not found in cache")
	}

	// was it deleted?
	if delta.Deleted != nil {
		ctx.Debug().Msg("Processing deletion delta")
		if delta.IsDir() && local != nil && local.HasChildren() {
			// from docs: you should only delete a folder locally if it is empty
			// after syncing all the changes.
			ctx.Warn().Str("delta", "delete").
				Msg("Refusing delta deletion of non-empty folder as per API docs.")
			return errors.New("directory is non-empty")
		}
		ctx.Info().Str("delta", "delete").
			Msg("Applying server-side deletion of item.")
		f.DeleteID(id)
		return nil
	}

	// does the item exist locally? if not, add the delta to the cache under the
	// appropriate parent
	if local == nil {
		ctx.Debug().Msg("Item not in cache by ID, checking by name")
		// check if we don't have it here first
		local, _ = f.GetChild(parentID, name, nil)
		if local != nil {
			localID := local.ID()
			ctx.Info().
				Str("localID", localID).
				Msg("Local item already exists under different ID.")
			if isLocalID(localID) {
				ctx.Debug().Msg("Local ID is a temporary ID, moving to permanent ID")
				if err := f.MoveID(localID, id); err != nil {
					ctx.Error().
						Str("localID", localID).
						Err(err).
						Msg("Could not move item to new, nonlocal ID!")
				} else {
					ctx.Debug().Msg("Successfully moved item to new ID")
				}
			}
		} else {
			ctx.Info().Str("delta", "create").
				Msg("Creating inode from delta.")
			f.InsertChild(parentID, NewInodeDriveItem(delta))
			ctx.Debug().Msg("Successfully created inode from delta")
			return nil
		}
	}

	// was the item moved?
	localName := local.Name()
	if local.ParentID() != parentID || local.Name() != name {
		ctx.Debug().Msg("Processing move/rename delta")
		log.Info().
			Str("parent", local.ParentID()).
			Str("name", localName).
			Str("newParent", parentID).
			Str("newName", name).
			Str("id", id).
			Str("delta", "rename").
			Msg("Applying server-side rename")
		oldParentID := local.ParentID()
		// local rename only
		ctx.Debug().Msg("Calling MovePath to rename/move item")
		if err := f.MovePath(oldParentID, parentID, localName, name, f.auth); err != nil {
			ctx.Error().Err(err).Msg("Failed to rename/move item")
			// Continue processing as there may be additional changes
		} else {
			ctx.Debug().Msg("Successfully renamed/moved item")
		}
		// do not return, there may be additional changes
	}

	// Finally, check if the content/metadata of the remote has changed.
	// "Interesting" changes must be synced back to our local state without
	// data loss or corruption. Currently the only thing the local filesystem
	// actually modifies remotely is the actual file data, so we simply accept
	// the remote metadata changes that do not deal with the file's content
	// changing.
	ctx.Debug().
		Time("localModTime", *local.DriveItem.ModTime).
		Time("remoteModTime", *delta.ModTime).
		Str("localETag", local.ETag).
		Str("remoteETag", delta.ETag).
		Msg("Checking for content changes")

	if delta.ModTimeUnix() > local.ModTime() && !delta.ETagIsMatch(local.ETag) {
		ctx.Debug().Msg("Remote item is newer than local item")
		sameContent := false
		if !delta.IsDir() && delta.File != nil {
			ctx.Debug().Msg("Verifying file content checksum")
			local.RLock()
			sameContent = local.VerifyChecksum(delta.File.Hashes.QuickXorHash)
			local.RUnlock()
			ctx.Debug().Bool("sameContent", sameContent).Msg("Checksum verification result")
		}

		if !sameContent {
			ctx.Debug().Msg("Content has changed, checking for local modifications")
			// Check if we have local changes
			hasLocalChanges := false

			// Check if the item has been modified offline
			ctx.Debug().Msg("Checking for offline changes")
			if err := f.db.View(func(tx *bolt.Tx) error {
				b := tx.Bucket(bucketOfflineChanges)
				if b == nil {
					return nil
				}

				c := b.Cursor()
				prefix := []byte(delta.ID + "-")
				k, _ := c.Seek(prefix)
				if k != nil && bytes.HasPrefix(k, prefix) {
					hasLocalChanges = true
				}
				return nil
			}); err != nil {
				ctx.Error().Err(err).Msg("Failed to check for offline changes")
			}

			// Also check if the item has pending uploads
			if !hasLocalChanges {
				ctx.Debug().Msg("Checking for pending uploads")
				if err := f.db.View(func(tx *bolt.Tx) error {
					b := tx.Bucket(bucketUploads)
					if b == nil {
						return nil
					}

					if b.Get([]byte(delta.ID)) != nil {
						hasLocalChanges = true
					}
					return nil
				}); err != nil {
					ctx.Error().Err(err).Msg("Failed to check for pending uploads")
				}
			}

			ctx.Debug().Bool("hasLocalChanges", hasLocalChanges).Msg("Local modification check result")

			if hasLocalChanges {
				// Conflict detected - create a conflict copy
				ctx.Info().Str("delta", "conflict").
					Msg("Conflict detected, creating conflict copy.")

				// Mark the file as having a conflict
				ctx.Debug().Msg("Marking file as having a conflict")
				f.MarkFileConflict(delta.ID, "Conflict detected between local and remote changes")

				// Create a conflict copy of the remote file
				conflictName := delta.Name + " (Conflict Copy " + time.Now().Format("2006-01-02 15:04:05") + ")"
				conflictItem := *delta
				conflictItem.Name = conflictName

				// Add the conflict copy to the filesystem
				ctx.Debug().Str("conflictName", conflictName).Msg("Creating conflict copy")
				conflictInode := NewInodeDriveItem(&conflictItem)
				f.InsertChild(parentID, conflictInode)
				ctx.Debug().Msg("Successfully created conflict copy")

				// Keep the local version as is
				return nil
			}

			// No local changes, accept remote changes
			ctx.Info().Str("delta", "overwrite").
				Msg("Overwriting local item, no local changes to preserve.")

			// Mark the file as out of sync
			ctx.Debug().Msg("Marking file as out of sync")
			f.MarkFileOutofSync(delta.ID)

			// update modtime, hashes, purge any local content in memory
			ctx.Debug().Msg("Updating local item with remote metadata")
			local.Lock()
			local.DriveItem.ModTime = delta.ModTime
			local.DriveItem.Size = delta.Size
			local.DriveItem.ETag = delta.ETag
			// the rest of these are harmless when this is a directory
			// as they will be null anyways
			local.DriveItem.File = delta.File
			local.hasChanges = false
			local.Unlock()

			// Update file status attributes after releasing the lock
			ctx.Debug().Msg("Updating file status attributes")
			f.updateFileStatus(local)

			// Queue a download for the file to ensure content is synced
			if !delta.IsDir() {
				ctx.Debug().Msg("Queueing download for changed file")
				if _, err := f.downloads.QueueDownload(delta.ID); err != nil {
					ctx.Error().Err(err).Msg("Failed to queue download for changed file")
				} else {
					ctx.Debug().Msg("Successfully queued download for changed file")
				}
			}

			ctx.Debug().Msg("Successfully updated local item with remote metadata")

			return nil
		}
	}

	ctx.Debug().Str("delta", "skip").Msg("Skipping, no changes relative to local state.")
	return nil
}
