package fs

import "github.com/auriora/onemount/internal/metadata"

// markDirtyLocalState records a local modification: sets the runtime flag and transitions metadata to DIRTY_LOCAL.
func (f *Filesystem) markDirtyLocalState(id string) {
	if id == "" {
		return
	}
	f.transitionItemState(id, metadata.ItemStateDirtyLocal)
}

// markHydratedState marks an entry as hydrated (e.g., post-hydration) and clears pending-remote markers.
func (f *Filesystem) markHydratedState(id string) {
	if id == "" {
		return
	}
	f.transitionItemState(id, metadata.ItemStateHydrated, metadata.ClearPendingRemote())
	if inode := f.GetID(id); inode != nil {
		inode.mu.Lock()
		inode.hasChanges = false
		inode.mu.Unlock()
	}
}

// markCleanLocalState clears local-dirty flags after upload/reconcile and transitions to HYDRATED.
func (f *Filesystem) markCleanLocalState(id string) {
	if id == "" {
		return
	}
	if inode := f.GetID(id); inode != nil {
		inode.mu.Lock()
		inode.hasChanges = false
		inode.mu.Unlock()
	}
	f.transitionToState(id, metadata.ItemStateHydrated, metadata.ClearPendingRemote())
}

// markPendingUpload records that local content exists and needs upload.
func (f *Filesystem) markPendingUpload(id string) {
	if id == "" {
		return
	}
	if inode := f.GetID(id); inode != nil {
		inode.mu.Lock()
		inode.hasChanges = true
		inode.mu.Unlock()
	}
	f.transitionItemState(id, metadata.ItemStateDirtyLocal)
}
