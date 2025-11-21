package fs

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/auriora/onemount/internal/graph"
	"github.com/auriora/onemount/internal/logging"
	bolt "go.etcd.io/bbolt"
)

const (
	defaultPollingInterval          = 5 * time.Minute
	defaultRealtimeFallbackInterval = 30 * time.Minute

	deltaIntervalDeviationSuffix = "-deviation"
)

func (f *Filesystem) startRealtimeManager() (<-chan struct{}, error) {
	if f.realtimeOptions == nil {
		return nil, nil
	}
	notifier := NewChangeNotifier(*f.realtimeOptions, f.auth)
	if err := notifier.Start(f.deltaLoopCtx); err != nil {
		return nil, err
	}
	f.subscriptionManager = notifier
	return notifier.Notifications(), nil
}

func (f *Filesystem) stopRealtimeManager() {
	if f.subscriptionManager == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := f.subscriptionManager.Stop(ctx); err != nil {
		logging.Warn().Err(err).Msg("Failed to stop realtime subscription manager")
	}
	f.subscriptionManager = nil
}

func (f *Filesystem) logDeltaInterval(interval time.Duration, mode string, expected time.Duration) {
	reason := mode
	if interval != expected {
		reason += deltaIntervalDeviationSuffix
	}
	if f.lastDeltaInterval == interval && f.lastDeltaReason == reason {
		return
	}

	f.lastDeltaInterval = interval
	f.lastDeltaReason = reason

	if interval == expected {
		logging.Info().
			Str("mode", mode).
			Dur("interval", interval).
			Msg("Delta polling interval set to default")
		return
	}

	logging.Warn().
		Str("mode", mode).
		Dur("interval", interval).
		Dur("expected", expected).
		Msg("Delta polling interval deviates from requirement; continuing with configured value")
}

func (f *Filesystem) shouldUseActiveInterval(baseInterval time.Duration) bool {
	if f.activeDeltaInterval <= 0 || f.activeDeltaWindow <= 0 {
		return false
	}
	if f.activeDeltaInterval >= baseInterval {
		return false
	}
	last := f.lastForegroundActivity.Load()
	if last == 0 {
		return false
	}
	lastTime := time.Unix(0, last)
	return time.Since(lastTime) <= f.activeDeltaWindow
}

func (f *Filesystem) desiredDeltaInterval() time.Duration {
	if f.subscriptionManager != nil && f.subscriptionManager.IsActive() {
		expected := defaultRealtimeFallbackInterval
		interval := expected
		if f.realtimeOptions != nil && f.realtimeOptions.FallbackInterval > 0 {
			interval = f.realtimeOptions.FallbackInterval
		}
		f.logDeltaInterval(interval, "realtime", expected)
		return interval
	}
	baseInterval := f.deltaInterval
	if baseInterval <= 0 {
		baseInterval = defaultPollingInterval
	}
	interval := baseInterval
	if f.shouldUseActiveInterval(baseInterval) {
		interval = f.activeDeltaInterval
	}
	f.logDeltaInterval(interval, "polling", baseInterval)
	return interval
}

// DeltaLoop creates a new thread to poll the server for changes and should be
// called as a goroutine.
//
// ETag-Based Cache Invalidation:
// This delta sync process is the primary mechanism for ETag-based cache validation.
// When remote files change, the delta query returns updated metadata including new ETags.
// The sync process:
// 1. Fetches changed items from OneDrive API (via delta query)
// 2. Compares new ETags with cached metadata ETags
// 3. Invalidates content cache entries when ETags differ
// 4. Updates metadata cache with new ETags
// 5. Next file access triggers re-download of invalidated content
//
// This approach is more efficient than using HTTP if-none-match headers for conditional
// GET requests because:
// - Batch metadata updates reduce API calls
// - Changes are detected proactively before file access
// - Pre-authenticated download URLs don't support conditional GET
// - Only changed files are re-downloaded
func (f *Filesystem) DeltaLoop(interval time.Duration) {
	logging.Info().Msg("Starting delta goroutine.")

	f.deltaInterval = interval

	// Add to wait groups to track this goroutine
	f.deltaLoopWg.Add(1)
	f.Wg.Add(1)
	defer func() {
		logging.Debug().Msg("Delta goroutine exiting, calling Done() on wait groups")
		f.deltaLoopWg.Done()
		f.Wg.Done()
		logging.Debug().Msg("Delta goroutine completed")
	}()

	notificationCh, err := f.startRealtimeManager()
	if err != nil {
		logging.Error().Err(err).Msg("Failed to start realtime subscription; continuing with polling only")
	} else if notificationCh != nil {
		logging.Info().Msg("Realtime subscription started; using extended delta interval when active")
		defer f.stopRealtimeManager()
	}

	currentInterval := f.desiredDeltaInterval()
	waitDur := currentInterval

	// Create a ticker for the interval
	ticker := time.NewTicker(currentInterval)
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
			logging.Info().Msg("Stopping delta goroutine via stop channel.")
			logging.Debug().Msg("Delta goroutine received stop signal, exiting loop")
			return
		case <-f.deltaLoopCtx.Done():
			logging.Info().Msg("Stopping delta goroutine via context cancellation.")
			logging.Debug().Msg("Delta goroutine context cancelled, exiting loop")
			return
		default:
			// Continue with normal operation
		}

		// get deltas
		logging.Debug().Msg("Starting delta fetch cycle")
		logging.Trace().Msg("Fetching deltas from server.")
		pollSuccess := false
		deltas := make(map[string]*graph.DriveItem)

		// Use a timeout for the entire delta fetch cycle
		fetchCtx, fetchCancel := context.WithTimeout(f.deltaLoopCtx, 2*time.Minute)

	fetchLoop:
		for {
			// Check if we should stop before making network call
			select {
			case <-f.deltaLoopStop:
				logging.Debug().Msg("Delta loop stop signal received during polling loop")
				fetchCancel()
				return
			case <-f.deltaLoopCtx.Done():
				logging.Debug().Msg("Delta loop context cancelled during polling loop")
				fetchCancel()
				return
			case <-fetchCtx.Done():
				logging.Debug().Msg("Delta fetch cycle timed out")
				break fetchLoop
			default:
				// Continue with normal operation
			}

			logging.Debug().Msg("Calling pollDeltas to fetch changes from server")
			incoming, cont, err := f.pollDeltas(f.auth)
			logging.Debug().Bool("continue", cont).Int("incomingCount", len(incoming)).Msg("pollDeltas returned")

			// Check again if we should stop after network call
			select {
			case <-f.deltaLoopStop:
				logging.Debug().Msg("Delta loop stop signal received after pollDeltas")
				fetchCancel()
				return
			case <-f.deltaLoopCtx.Done():
				logging.Debug().Msg("Delta loop context cancelled after pollDeltas")
				fetchCancel()
				return
			default:
				// Continue processing
			}

			if err != nil {
				// Check if it's a context cancellation error
				if fetchCtx.Err() != nil || f.deltaLoopCtx.Err() != nil {
					logging.Debug().Msg("Delta fetch was cancelled by context")
					fetchCancel()
					return
				}

				// the only thing that should be able to bring the FS out
				// of a read-only state is a successful delta call
				logging.Error().Err(err).
					Msg("Error during delta fetch, marking fs as offline.")
				// Lock ordering: filesystem.RWMutex only (no other locks held)
				// See docs/guides/developer/concurrency-guidelines.md
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
				logging.Info().Msgf("Fetched %d deltas.", len(deltas))
				pollSuccess = true
				break
			}
			logging.Debug().Msg("Need to continue polling for more deltas")
		}

		// Clean up the fetch context
		fetchCancel()

		// Check if we should stop before applying deltas
		select {
		case <-f.deltaLoopStop:
			logging.Debug().Msg("Delta loop stop signal received before applying deltas")
			return
		default:
			// Continue with normal operation
		}

		// now apply deltas
		logging.Debug().Int("deltaCount", len(deltas)).Msg("Starting to apply deltas")
		secondPass := make([]string, 0)

		// Use a timeout for delta application
		applyCtx, applyCancel := context.WithTimeout(f.deltaLoopCtx, 1*time.Minute)

	applyLoop:
		for _, delta := range deltas {
			// Check if we should stop before applying delta
			select {
			case <-f.deltaLoopStop:
				logging.Debug().Msg("Delta loop stop signal received during delta application")
				applyCancel()
				return
			case <-applyCtx.Done():
				logging.Debug().Msg("Delta application timed out")
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

		waitDur = currentInterval

		// Check if we should stop before second pass
		select {
		case <-f.deltaLoopStop:
			logging.Debug().Msg("Delta loop stop signal received before second pass")
			applyCancel()
			return
		case <-applyCtx.Done():
			logging.Debug().Msg("Delta application timed out before second pass")
			applyCancel()
			// Continue to next cycle
			goto nextCycle
		default:
			// Continue with normal operation
		}

		logging.Debug().Int("secondPassCount", len(secondPass)).Msg("Starting second pass for non-empty directories")
		for _, id := range secondPass {
			// Check if we should stop before processing each item in second pass
			select {
			case <-f.deltaLoopStop:
				logging.Debug().Msg("Delta loop stop signal received during second pass")
				applyCancel()
				return
			case <-applyCtx.Done():
				logging.Debug().Msg("Second pass timed out")
				break
			default:
				// Continue with normal operation
			}

			// failures should explicitly be ignored the second time around as per docs
			if err := f.applyDelta(deltas[id]); err != nil {
				logging.Debug().Err(err).Str("id", id).Msg("Ignoring error in second pass delta application")
			}
		}

		// Clean up the apply context
		applyCancel()

		logging.Debug().Msg("Finished applying deltas")

		// Check if we should stop before serialization
		select {
		case <-f.deltaLoopStop:
			logging.Debug().Msg("Delta loop stop signal received before serialization")
			return
		default:
			// Continue with normal operation
		}

		if !f.IsOffline() {
			logging.Debug().Msg("Serializing filesystem state")
			f.SerializeAll()
		}

		if pollSuccess {
			// Lock ordering: filesystem.RWMutex only (no other locks held)
			// See docs/guides/developer/concurrency-guidelines.md
			f.Lock()
			wasOffline := f.offline
			if f.offline {
				logging.Info().Msg("Delta fetch success, marking fs as online.")
			}
			f.offline = false
			f.Unlock()

			// Switch to normal ticker if we were using offline ticker
			if currentTicker == offlineTicker {
				currentTicker = ticker
			}

			logging.Debug().Msg("Saving delta link to database")
			if err := f.db.Batch(func(tx *bolt.Tx) error {
				return tx.Bucket(bucketDelta).Put([]byte("deltaLink"), []byte(f.deltaLink))
			}); err != nil {
				logging.Error().Err(err).Msg("Failed to save delta link to database")
			}

			// If we were offline and now we're online, process offline changes
			if wasOffline {
				logging.Info().Msg("Transitioning from offline to online, processing offline changes with enhanced sync manager")
				// Use a goroutine with proper error handling
				f.Wg.Add(1)
				go func(ctx context.Context) {
					defer f.Wg.Done()
					defer func() {
						if r := recover(); r != nil {
							logging.Error().Interface("recover", r).Msg("Panic in ProcessOfflineChanges")
						}
					}()

					// Create a child context with timeout
					processCtx, cancel := context.WithTimeout(ctx, 10*time.Minute)
					defer cancel()

					// Check if context is already cancelled
					select {
					case <-processCtx.Done():
						logging.Debug().Msg("Context cancelled, skipping offline changes processing")
						return
					default:
						// Continue with processing
					}

					// Use the enhanced sync manager for better error handling and conflict resolution
					result, err := f.ProcessOfflineChangesWithSyncManager(processCtx)
					if err != nil {
						logging.Error().Err(err).Msg("Failed to process offline changes with sync manager")
						// Fall back to the original method
						f.ProcessOfflineChangesWithContext(processCtx)
					} else {
						logging.Info().
							Int("processed", result.ProcessedChanges).
							Int("conflicts", result.ConflictsFound).
							Int("resolved", result.ConflictsResolved).
							Int("errors", len(result.Errors)).
							Dur("duration", result.Duration).
							Msg("Successfully processed offline changes with sync manager")
					}
				}(f.ctx)
			}
		} else {
			// Switch to offline ticker for shorter retry intervals
			if currentTicker == ticker {
				currentTicker = offlineTicker
			}
			waitDur = 2 * time.Second
		}

		if currentTicker != offlineTicker {
			desired := f.desiredDeltaInterval()
			if desired != currentInterval {
				ticker.Stop()
				ticker = time.NewTicker(desired)
				currentInterval = desired
				currentTicker = ticker
				waitDur = desired
				logging.Info().
					Dur("interval", desired).
					Msg("Adjusted delta polling interval based on realtime state")
			}
		}

	nextCycle:
		// Wait for next interval or stop signal
		select {
		case <-time.After(waitDur):
		case <-notificationCh:
			if notificationCh != nil {
				logging.Info().Msg("Realtime notification received; triggering immediate delta sync")
			}
		case <-currentTicker.C:
			// Time to run the next cycle
			logging.Debug().Msg("Ticker triggered, starting next delta cycle")
		case <-f.deltaLoopStop:
			logging.Info().Msg("Stopping delta goroutine during wait interval.")
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
	logging.Debug().Str("deltaLink", f.deltaLink).Msg("Making network request to delta endpoint")

	// Check if context is already cancelled before making request
	if f.deltaLoopCtx.Err() != nil {
		logging.Debug().Err(f.deltaLoopCtx.Err()).Msg("Context already cancelled before making delta request")
		return make([]*graph.DriveItem, 0), false, f.deltaLoopCtx.Err()
	}

	// Create a timeout context that's a child of the main context
	// This ensures we don't get stuck in network calls indefinitely
	ctx, cancel := context.WithTimeout(f.deltaLoopCtx, 30*time.Second)
	defer cancel()

	// Make the network request with context that can be cancelled during shutdown
	logging.Debug().Msg("Starting delta request with cancellable context")
	resp, err := graph.GetWithContext(ctx, f.deltaLink, auth)
	logging.Debug().Msg("Delta request completed or cancelled")

	// Check for context cancellation first
	if ctx.Err() != nil {
		logging.Debug().Err(ctx.Err()).Msg("Delta request context was cancelled")
		// Check if it was our parent context or just the timeout
		if f.deltaLoopCtx.Err() != nil {
			logging.Debug().Msg("Parent context was cancelled, stopping delta loop")
			return make([]*graph.DriveItem, 0), false, f.deltaLoopCtx.Err()
		}
		logging.Debug().Msg("Request timed out, will retry on next cycle")
		return make([]*graph.DriveItem, 0), false, ctx.Err()
	}

	if err != nil {
		logging.Error().Err(err).Msg("Error fetching deltas from server")
		return make([]*graph.DriveItem, 0), false, err
	}

	logging.Debug().Int("responseSize", len(resp)).Msg("Received response from delta endpoint")

	page := deltaResponse{}
	err = json.Unmarshal(resp, &page)
	if err != nil {
		logging.Error().Err(err).Msg("Error unmarshaling delta response")
		return make([]*graph.DriveItem, 0), false, err
	}

	// If the server does not provide a `@odata.nextLink` item, it means we've
	// reached the end of this polling cycle and should not continue until the
	// next poll interval.
	if page.NextLink != "" {
		newLink := strings.TrimPrefix(page.NextLink, graph.GraphURL)
		logging.Debug().Str("oldLink", f.deltaLink).Str("newLink", newLink).Bool("continue", true).Int("itemCount", len(page.Values)).Msg("Delta page has nextLink, continuing")
		f.deltaLink = newLink
		return page.Values, true, nil
	}

	newLink := strings.TrimPrefix(page.DeltaLink, graph.GraphURL)
	logging.Debug().Str("oldLink", f.deltaLink).Str("newLink", newLink).Bool("continue", false).Int("itemCount", len(page.Values)).Msg("Delta page has deltaLink, finished polling")
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
	// Create a context for this operation with request ID and user ID
	logCtx := logging.NewLogContextWithRequestAndUserID("delta")
	logCtx = logCtx.WithComponent("fs").WithMethod("applyDelta")
	logCtx = logCtx.With("id", id).With("parentID", parentID).With("name", name)
	logger := logging.WithLogContext(logCtx)
	logger.Debug().Msg("Applying delta")

	// diagnose and act on what type of delta we're dealing with

	// do we have it at all?
	parent := f.GetID(parentID)
	if parent == nil {
		// Nothing needs to be applied, item not in cache, so latest copy will
		// be pulled down next time it's accessed.
		logger.Debug().
			Str("delta", "skip").
			Msg("Skipping delta, item's parent not in cache.")
		return nil
	}

	logger.Debug().
		Str("parentPath", parent.Path()).
		Msg("Found parent in cache")

	local := f.GetID(id)
	if local != nil {
		logger.Debug().
			Str("localPath", local.Path()).
			Msg("Found item in cache")
	} else {
		logger.Debug().Msg("Item not found in cache")
	}

	// was it deleted?
	if delta.Deleted != nil {
		logger.Debug().Msg("Processing deletion delta")
		if delta.IsDir() && local != nil && local.HasChildren() {
			// from docs: you should only delete a folder locally if it is empty
			// after syncing all the changes.
			logger.Warn().Str("delta", "delete").
				Msg("Refusing delta deletion of non-empty folder as per API docs.")
			return errors.New("directory is non-empty")
		}
		logger.Info().Str("delta", "delete").
			Msg("Applying server-side deletion of item.")
		f.DeleteID(id)
		return nil
	}

	// does the item exist locally? if not, add the delta to the cache under the
	// appropriate parent
	if local == nil {
		logger.Debug().Msg("Item not in cache by ID, checking by name")
		// check if we don't have it here first
		local, _ = f.GetChild(parentID, name, nil)
		if local != nil {
			localID := local.ID()
			logger.Debug().
				Str("localID", localID).
				Msg("Local item already exists under different ID.")
			if isLocalID(localID) {
				logger.Debug().Msg("Local ID is a temporary ID, moving to permanent ID")
				if err := f.MoveID(localID, id); err != nil {
					logger.Debug().
						Str("localID", localID).
						Err(err).
						Msg("Could not move item to new, nonlocal ID!")
				} else {
					logger.Debug().Msg("Successfully moved item to new ID")
				}
			}
		} else {
			logger.Info().Str("delta", "create").
				Msg("Creating inode from delta.")
			f.InsertChild(parentID, NewInodeDriveItem(delta))
			f.updateMetadataFromDelta(id, delta)
			logger.Debug().Msg("Successfully created inode from delta")
			return nil
		}
	}

	// was the item moved?
	localName := local.Name()
	if local.ParentID() != parentID || local.Name() != name {
		logger.Debug().Msg("Processing move/rename delta")
		logging.Info().
			Str("parent", local.ParentID()).
			Str("name", localName).
			Str("newParent", parentID).
			Str("newName", name).
			Str("id", id).
			Str("delta", "rename").
			Msg("Applying server-side rename")
		oldParentID := local.ParentID()
		// local rename only
		logger.Debug().Msg("Calling MovePath to rename/move item")
		if err := f.MovePath(oldParentID, parentID, localName, name, f.auth); err != nil {
			logger.Debug().Err(err).Msg("Failed to rename/move item")
			// Continue processing as there may be additional changes
		} else {
			logger.Debug().Msg("Successfully renamed/moved item")
			f.updateMetadataFromDelta(id, delta)
		}
		// do not return, there may be additional changes
	}

	// Finally, check if the content/metadata of the remote has changed.
	// "Interesting" changes must be synced back to our local state without
	// data loss or corruption. Currently the only thing the local filesystem
	// actually modifies remotely is the actual file data, so we simply accept
	// the remote metadata changes that do not deal with the file's content
	// changing.
	logger.Debug().
		Time("localModTime", *local.DriveItem.ModTime).
		Time("remoteModTime", *delta.ModTime).
		Str("localETag", local.DriveItem.ETag).
		Str("remoteETag", delta.ETag).
		Msg("Checking for content changes")

	if delta.ModTimeUnix() > local.ModTime() && !delta.ETagIsMatch(local.DriveItem.ETag) {
		logger.Debug().Msg("Remote item is newer than local item")
		sameContent := false
		if !delta.IsDir() && delta.File != nil {
			logger.Debug().Msg("Checking file hashes")
			// check if the content is the same
			if delta.File.Hashes.QuickXorHash != "" && local.DriveItem.File != nil &&
				local.DriveItem.File.Hashes.QuickXorHash != "" {
				sameContent = delta.File.Hashes.QuickXorHash == local.DriveItem.File.Hashes.QuickXorHash
				logger.Debug().Bool("sameContent", sameContent).Msg("Compared QuickXorHash values")
			}
		}

		if sameContent {
			logger.Info().Str("delta", "update").
				Msg("Updating metadata only, content is the same")
			// update the metadata only
			local.mu.Lock()
			local.DriveItem.ModTime = delta.ModTime
			local.DriveItem.Size = delta.Size
			local.DriveItem.ETag = delta.ETag
			local.DriveItem.File = delta.File
			local.hasChanges = false
			local.mu.Unlock()
			logger.Debug().Msg("Updated metadata")
			f.persistMetadataEntry(id, local)
			f.updateMetadataFromDelta(id, delta)
		} else {
			logger.Info().Str("delta", "invalidate").
				Msg("Content has changed, invalidating cache and marking file as out of sync")
			// Explicitly invalidate the cache by deleting cached content
			// This ensures stale content is not served to users
			if err := f.content.Delete(id); err != nil {
				logger.Warn().Err(err).Msg("Failed to delete cached content during invalidation")
			}
			f.handleContentEvicted(id)
			// Mark file status as OutofSync to indicate it needs to be re-downloaded
			f.MarkFileOutofSync(id)
			// update the metadata with new ETag and size
			local.mu.Lock()
			local.DriveItem.ModTime = delta.ModTime
			local.DriveItem.Size = delta.Size
			local.DriveItem.ETag = delta.ETag
			local.DriveItem.File = delta.File
			local.hasChanges = false
			local.mu.Unlock()
			// Update file status extended attributes for UI integration
			f.updateFileStatus(local)
			logger.Debug().Msg("Updated metadata, invalidated content cache, and marked file as OutofSync")
			f.persistMetadataEntry(id, local)
			f.updateMetadataFromDelta(id, delta)
		}
	} else {
		logger.Debug().Msg("Local item is up to date")
	}

	return nil
}
