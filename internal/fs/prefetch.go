package fs

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/auriora/onemount/internal/graph"
	"github.com/auriora/onemount/internal/logging"
	"github.com/auriora/onemount/internal/metadata"
)

const (
	prefetchDepthLimit = 100
	prefetchWorkerID   = "prefetch"
)

type prefetchQueueItem struct {
	id    string
	depth int
}

// StartPrefetch initiates background recursive metadata prefetching.
func (f *Filesystem) StartPrefetch() {
	if f == nil {
		return
	}
	if f.IsOffline() {
		f.prefetchPending.Store(true)
		logging.Info().Msg("Prefetch skipped because filesystem is offline")
		return
	}
	if f.prefetchActive.Swap(true) {
		return
	}
	f.prefetchPending.Store(false)
	f.Wg.Add(1)
	go func() {
		defer f.Wg.Done()
		defer f.prefetchActive.Store(false)
		f.runPrefetch(context.Background())
	}()
}

func (f *Filesystem) runPrefetch(ctx context.Context) {
	if f == nil {
		return
	}
	if f.auth == nil {
		logging.Warn().Msg("Prefetch skipped because auth is not available")
		return
	}
	if f.root == "" {
		logging.Warn().Msg("Prefetch skipped because root ID is empty")
		return
	}

	visited := make(map[string]bool)
	queue := []prefetchQueueItem{{id: f.root, depth: 0}}
	batch := 0

	for len(queue) > 0 {
		nextBatch := make([]prefetchQueueItem, 0)

		for len(queue) > 0 {
			item := queue[0]
			queue = queue[1:]

			if item.id == "" {
				continue
			}
			if item.depth > prefetchDepthLimit {
				nextBatch = append(nextBatch, prefetchQueueItem{id: item.id, depth: 0})
				continue
			}
			if visited[item.id] {
				continue
			}
			visited[item.id] = true

			subdirs, err := f.prefetchDirectory(ctx, item.id, item.depth)
			if err != nil {
				if graph.IsOffline(err) || f.IsOffline() {
					f.prefetchPending.Store(true)
					logging.Info().Msg("Prefetch paused due to offline state")
					return
				}
				continue
			}

			for _, childID := range subdirs {
				if childID == "" || visited[childID] {
					continue
				}
				nextDepth := item.depth + 1
				if nextDepth > prefetchDepthLimit {
					nextBatch = append(nextBatch, prefetchQueueItem{id: childID, depth: 0})
				} else {
					queue = append(queue, prefetchQueueItem{id: childID, depth: nextDepth})
				}
			}
		}

		if len(nextBatch) == 0 {
			break
		}
		batch++
		logging.Debug().
			Int("batch", batch).
			Int("queueDepth", len(nextBatch)).
			Msg("Prefetch batch complete, continuing with deferred directories")
		queue = nextBatch
	}

	logging.Info().Msg("Prefetch completed")
}

func (f *Filesystem) prefetchDirectory(ctx context.Context, dirID string, depth int) ([]string, error) {
	if f == nil || dirID == "" {
		return nil, nil
	}
	if f.auth == nil {
		return nil, errors.New("prefetch requires auth")
	}

	f.ensureMetadataEntry(dirID)
	start := time.Now().UTC()
	f.transitionItemState(dirID, metadata.ItemStateHydrating,
		metadata.WithHydrationEvent(),
		metadata.WithWorker(prefetchWorkerID),
		metadata.WithTransitionTimestamp(start))

	// Prefetch is metadata-only and must never download file contents.
	items, err := f.fetchPrefetchChildren(ctx, dirID, f.auth)
	if err != nil {
		f.transitionItemState(dirID, metadata.ItemStateError,
			metadata.WithHydrationEvent(),
			metadata.WithWorker(prefetchWorkerID),
			metadata.WithTransitionError(err, graph.IsOffline(err)),
			metadata.WithTransitionTimestamp(time.Now().UTC()))
		logging.Debug().
			Str(logging.FieldID, dirID).
			Int("depth", depth).
			Err(err).
			Msg("Prefetch failed for directory")
		return nil, err
	}

	children := make(map[string]*Inode, len(items))
	subdirs := make([]string, 0, len(items))

	for _, item := range items {
		if item == nil {
			continue
		}
		if strings.EqualFold(item.Name, xdgVolumeInfoName) {
			continue
		}
		child := NewInodeDriveItem(item)
		f.InsertNodeID(child)
		f.metadata.Store(child.DriveItem.ID, child)
		f.persistMetadataEntry(child.DriveItem.ID, child)

		lowerName := strings.ToLower(child.Name())
		children[lowerName] = child
		if child.IsDir() {
			subdirs = append(subdirs, child.ID())
		}
	}

	if len(children) > 0 {
		f.cacheChildrenFromMap(dirID, children)
	}

	f.assertMetadataOnlyPrefetch(items)

	f.transitionToState(dirID, metadata.ItemStateHydrated,
		metadata.WithHydrationEvent(),
		metadata.WithWorker(prefetchWorkerID),
		metadata.WithTransitionTimestamp(time.Now().UTC()),
		metadata.ClearPendingRemote())

	logging.Debug().
		Str(logging.FieldID, dirID).
		Int("depth", depth).
		Int("childCount", len(children)).
		Msg("Prefetch completed for directory")

	return subdirs, nil
}

func (f *Filesystem) fetchPrefetchChildren(ctx context.Context, dirID string, auth *graph.Auth) ([]*graph.DriveItem, error) {
	if dirID == "" {
		return nil, errors.New("prefetch requires directory ID")
	}
	if auth == nil {
		return nil, errors.New("prefetch requires auth")
	}

	timeout := 30 * time.Second
	if f != nil && f.timeoutConfig != nil && f.timeoutConfig.MetadataRequestTimeout > 0 {
		timeout = f.timeoutConfig.MetadataRequestTimeout
	}

	fetchCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	resultChan := make(chan struct {
		items []*graph.DriveItem
		err   error
	}, 1)

	if f.metadataRequestManager != nil {
		reqErr := f.metadataRequestManager.QueueChildrenRequest(dirID, auth, PriorityBackground, func(items []*graph.DriveItem, reqErr error) {
			resultChan <- struct {
				items []*graph.DriveItem
				err   error
			}{items: items, err: reqErr}
		})
		if reqErr != nil {
			items, err := graph.GetItemChildren(dirID, auth)
			return items, err
		}
	} else {
		items, err := graph.GetItemChildren(dirID, auth)
		return items, err
	}

	select {
	case result := <-resultChan:
		return result.items, result.err
	case <-fetchCtx.Done():
		return nil, context.DeadlineExceeded
	}
}

func (f *Filesystem) ensureMetadataEntry(id string) {
	if f == nil || f.metadataStore == nil || id == "" {
		return
	}
	_, err := f.metadataStore.Get(context.Background(), id)
	if err == nil {
		return
	}
	inode := f.GetID(id)
	if inode == nil {
		inode = f.ensureInodeFromMetadataStore(id)
	}
	if inode == nil {
		return
	}
	f.persistMetadataEntry(id, inode)
}

// assertMetadataOnlyPrefetch logs when prefetch encounters cached content. Prefetch never
// populates content; any cached data is from prior file opens or background hydration.
func (f *Filesystem) assertMetadataOnlyPrefetch(items []*graph.DriveItem) {
	if f == nil || f.content == nil || len(items) == 0 {
		return
	}
	for _, item := range items {
		if item == nil || item.IsDir() {
			continue
		}
		if f.content.HasContent(item.ID) {
			logging.Debug().
				Str(logging.FieldID, item.ID).
				Msg("Prefetch observed existing content; metadata prefetch does not download file data")
		}
	}
}

func (f *Filesystem) hasCachedRootChildren() bool {
	if f == nil || f.metadataStore == nil || f.root == "" {
		return false
	}
	entry, err := f.metadataStore.Get(context.Background(), f.root)
	if err != nil {
		return false
	}
	return len(entry.Children) > 0
}

func (f *Filesystem) populateRootChildrenFromMetadata() bool {
	if f == nil || f.metadataStore == nil || f.root == "" {
		return false
	}
	entry, err := f.metadataStore.Get(context.Background(), f.root)
	if err != nil || len(entry.Children) == 0 {
		return false
	}
	children := make(map[string]*Inode, len(entry.Children))
	for _, childID := range entry.Children {
		if childID == "" {
			continue
		}
		child := f.ensureInodeFromMetadataStore(childID)
		if child == nil {
			continue
		}
		children[strings.ToLower(child.Name())] = child
	}
	if len(children) == 0 {
		return false
	}
	f.cacheChildrenFromMap(f.root, children)
	return true
}
