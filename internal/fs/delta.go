package fs

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/auriora/onemount/internal/graph"
	"github.com/auriora/onemount/internal/logging"
	"github.com/auriora/onemount/internal/metadata"
	"github.com/auriora/onemount/internal/socketio"
	bolt "go.etcd.io/bbolt"
)

const (
	defaultPollingInterval          = 5 * time.Minute
	defaultRealtimeFallbackInterval = 30 * time.Minute
	defaultRecoveryInterval         = 10 * time.Second

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
	if interval, ok := f.deltaIntervalFromNotifier(); ok {
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

func (f *Filesystem) deltaIntervalFromNotifier() (time.Duration, bool) {
	if f.subscriptionManager == nil || f.realtimeOptions == nil || !f.realtimeOptions.Enabled {
		return 0, false
	}

	health, hasHealth := f.notifierHealthSnapshot()
	recoverySince := f.notifierRecoverySince.Load()
	recoveryWindowOpen := recoverySince != 0

	if hasHealth && isNotifierFailed(health.Status) {
		f.logDeltaInterval(defaultRecoveryInterval, "realtime-recovery", defaultRecoveryInterval)
		if !recoveryWindowOpen {
			f.notifierRecoverySince.Store(time.Now().UnixNano())
		}
		return defaultRecoveryInterval, true
	}

	if hasHealth && isNotifierDegraded(health.Status) {
		f.logDeltaInterval(defaultPollingInterval, "realtime-degraded", defaultPollingInterval)
		return defaultPollingInterval, true
	}

	if f.subscriptionManager.IsActive() {
		expected := defaultRealtimeFallbackInterval
		interval := expected
		if f.realtimeOptions.FallbackInterval > 0 {
			interval = f.realtimeOptions.FallbackInterval
		}
		// Requirement 5.4: no more frequent than defaultRealtimeFallbackInterval
		if interval < defaultRealtimeFallbackInterval {
			interval = defaultRealtimeFallbackInterval
		}
		f.logDeltaInterval(interval, "realtime", expected)
		return interval, true
	}

	// Clear recovery window when healthy/active again
	f.notifierRecoverySince.Store(0)
	return 0, false
}

func (f *Filesystem) notifierHealthSnapshot() (socketio.HealthState, bool) {
	provider, ok := f.subscriptionManager.(interface {
		HealthSnapshot() socketio.HealthState
	})
	if !ok {
		return socketio.HealthState{}, false
	}
	state := provider.HealthSnapshot()
	f.noteNotifierHealth(state)
	return state, true
}

func isNotifierDegraded(status socketio.StatusCode) bool {
	return status == socketio.StatusDegraded
}

func isNotifierFailed(status socketio.StatusCode) bool {
	return status == socketio.StatusFailed
}

func (f *Filesystem) noteNotifierHealth(state socketio.HealthState) {
	now := time.Now()
	prev := socketio.StatusUnknown
	if v := f.notifierLastStatus.Load(); v != nil {
		if s, ok := v.(socketio.StatusCode); ok {
			prev = s
		}
	}

	enteredDegraded := !isNotifierDegraded(prev) && !isNotifierFailed(prev) && (isNotifierDegraded(state.Status) || isNotifierFailed(state.Status))
	leftDegraded := (isNotifierDegraded(prev) || isNotifierFailed(prev)) && !(isNotifierDegraded(state.Status) || isNotifierFailed(state.Status))

	if enteredDegraded {
		f.notifierDegradedSince.Store(now.UnixNano())
		logging.Warn().
			Str("status", string(state.Status)).
			Int("consecutiveFailures", state.ConsecutiveFailures).
			Int("missedHeartbeats", state.MissedHeartbeats).
			Err(state.LastError).
			Msg("Realtime notifier degraded; switching to fallback polling cadence")
	}

	if leftDegraded {
		start := f.notifierDegradedSince.Swap(0)
		if start != 0 {
			duration := time.Since(time.Unix(0, start))
			logging.Info().
				Dur("duration", duration).
				Msg("Realtime notifier recovered; restoring realtime cadence")
		}
	}

	if isNotifierFailed(state.Status) {
		if f.notifierRecoverySince.Load() == 0 {
			f.notifierRecoverySince.Store(now.UnixNano())
			logging.Warn().
				Err(state.LastError).
				Int("consecutiveFailures", state.ConsecutiveFailures).
				Msg("Realtime notifier failed; entering 10-second recovery polling window")
		}
	} else {
		f.notifierRecoverySince.Store(0)
	}

	f.notifierLastStatus.Store(state.Status)
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

	ctx := context.Background()

	// Ensure the parent exists in the structured metadata store. If not, skip quietly.
	if f.metadataStore != nil {
		if _, err := f.metadataStore.Get(ctx, parentID); err != nil {
			if !errors.Is(err, metadata.ErrNotFound) {
				logger.Debug().Err(err).Str("parentID", parentID).Msg("Failed to read parent metadata")
			} else {
				logger.Debug().Str("parentID", parentID).Msg("Skipping delta; parent metadata not present yet")
			}
			return nil
		}
	}

	// was it deleted?
	if delta.Deleted != nil {
		logger.Debug().Msg("Processing deletion delta")
		logger.Info().Str("delta", "delete").
			Msg("Applying server-side deletion of item.")
		_ = f.removeChildFromParent(ctx, parentID, id, delta.IsDir())
		f.markEntryDeleted(id)
		f.DeleteID(id)
		return nil
	}

	// Pull prior metadata if present to evaluate transitions.
	var prior *metadata.Entry
	if f.metadataStore != nil {
		if entry, err := f.metadataStore.Get(ctx, id); err == nil {
			prior = entry
		}
	}

	updated, previous, err := f.upsertDriveItemEntry(ctx, delta, time.Now().UTC())
	if err != nil {
		logger.Debug().Err(err).Str("id", id).Msg("Failed to upsert metadata entry for delta")
		return err
	}
	if previous == nil {
		previous = prior
	}

	// Honor parent/child relationships in metadata.
	if previous != nil && previous.ParentID != updated.ParentID {
		f.moveChildBetweenParents(ctx, previous.ParentID, updated.ParentID, updated)
	} else {
		_ = f.addChildToParent(ctx, updated.ParentID, updated)
	}

	// Hydrate inode cache for downstream consumers (metadata remains source of truth).
	f.ensureInodeFromMetadataStore(updated.ID)

	// was the item moved?
	if previous != nil && (previous.ParentID != parentID || previous.Name != name) {
		logger.Debug().Msg("Processing move/rename delta")
		logging.Info().
			Str("parent", previous.ParentID).
			Str("name", previous.Name).
			Str("newParent", parentID).
			Str("newName", name).
			Str("id", id).
			Str("delta", "rename").
			Msg("Applying server-side rename")
	}

	etagChanged := previous != nil && previous.ETag != "" && delta.ETag != "" && previous.ETag != delta.ETag

	switch {
	case delta.IsDir():
		f.transitionToState(id, metadata.ItemStateHydrated, metadata.ClearPendingRemote())
	default:
		if etagChanged {
			logger.Info().Str("delta", "invalidate").
				Msg("Content has changed, invalidating cache and marking file as out of sync")
			if f.content != nil {
				if err := f.content.Delete(id); err != nil {
					logger.Warn().Err(err).Msg("Failed to delete cached content during invalidation")
				}
			}
			f.handleContentEvicted(id)
			f.MarkFileOutofSync(id)

			priorMode := metadata.PinModeUnset
			if previous != nil {
				priorMode = previous.Pin.Mode
			}
			priorPinned := priorMode == metadata.PinModeAlways
			currentPin := metadata.PinModeUnset
			if entry, _ := f.GetMetadataEntry(id); entry != nil {
				currentPin = entry.Pin.Mode
			}
			currentPinned := currentPin == metadata.PinModeAlways
			logging.Debug().
				Str("id", id).
				Str("pin_prev", string(priorMode)).
				Str("pin_curr", string(currentPin)).
				Msg("Evaluating auto-hydration for invalidated item")

			// Respect pinning: if the item was pinned before the remote change, restore the
			// pin metadata (if the upsert cleared it) and queue hydration.
			if priorPinned {
				_, _ = f.UpdateMetadataEntry(id, func(entry *metadata.Entry) error {
					entry.Pin = previous.Pin
					return nil
				})
			}
			// Queue hydration when the item is (or was) pinned to ALWAYS.
			if priorPinned || currentPinned {
				f.autoHydratePinned(id)
			}

			if previous.State == metadata.ItemStateDirtyLocal {
				f.transitionToState(id, metadata.ItemStateConflict, metadata.ClearPendingRemote())
			} else {
				f.transitionToState(id, metadata.ItemStateGhost, metadata.ClearPendingRemote())
			}
		} else {
			f.transitionToState(id, metadata.ItemStateHydrated, metadata.ClearPendingRemote())
		}
	}

	return nil
}
