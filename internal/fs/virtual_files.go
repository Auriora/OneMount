package fs

// registerVirtualFileInternal stores a virtual inode and makes it visible to the filesystem.
func (f *Filesystem) registerVirtualFileInternal(inode *Inode) {
	if inode == nil {
		return
	}
	f.virtualMu.Lock()
	if f.virtualFiles == nil {
		f.virtualFiles = make(map[string]*Inode)
	}
	f.virtualFiles[inode.ID()] = inode
	f.virtualMu.Unlock()
	f.metadata.Store(inode.ID(), inode)

	// Ensure the inode has a node ID so FUSE can reference it
	f.InsertNodeID(inode)
	f.persistMetadataEntry(inode.ID(), inode)

	if parentID := inode.ParentID(); parentID != "" {
		parent := f.GetID(parentID)
		if parent == nil {
			parent = f.ensureInodeFromMetadataStore(parentID)
		}
		if parent != nil {
			parent.mu.Lock()
			alreadyPresent := false
			for _, childID := range parent.children {
				if childID == inode.ID() {
					alreadyPresent = true
					break
				}
			}
			if !alreadyPresent {
				parent.children = append(parent.children, inode.ID())
			}
			parent.mu.Unlock()
			f.persistMetadataEntry(parentID, parent)
		}
	}
}

// RegisterVirtualFile exposes virtual file registration to other packages (e.g., cmd/common).
func (f *Filesystem) RegisterVirtualFile(inode *Inode) {
	f.registerVirtualFileInternal(inode)
}

// getVirtualFile returns the virtual inode for the given ID, if any.
func (f *Filesystem) getVirtualFile(id string) (*Inode, bool) {
	f.virtualMu.RLock()
	defer f.virtualMu.RUnlock()
	inode, ok := f.virtualFiles[id]
	return inode, ok
}

// collectVirtualChildSnapshots gathers metadata for all virtual children of the
// specified parent. The caller must NOT hold the parent lock while calling this
// helper because it takes per-child locks to snapshot their state.
func (f *Filesystem) collectVirtualChildSnapshots(parentID string) []childSnapshot {
	if parentID == "" {
		return nil
	}
	f.virtualMu.RLock()
	defer f.virtualMu.RUnlock()
	var snapshots []childSnapshot
	for _, inode := range f.virtualFiles {
		if inode.ParentID() != parentID {
			continue
		}
		if snapshots == nil {
			snapshots = make([]childSnapshot, 0, 4)
		}
		snapshots = append(snapshots, newChildSnapshot(inode))
	}
	return snapshots
}

// appendVirtualChildrenLocked appends precomputed virtual children snapshots to
// the parent. The caller must hold the parent's lock.
func (f *Filesystem) appendVirtualChildrenLocked(parent *Inode, virtualChildren []childSnapshot) {
	if parent == nil || len(virtualChildren) == 0 {
		return
	}
	existing := make(map[string]struct{}, len(parent.children))
	for _, childID := range parent.children {
		existing[childID] = struct{}{}
	}
	for _, snapshot := range virtualChildren {
		if snapshot.inode == nil {
			continue
		}
		if _, exists := existing[snapshot.id]; exists {
			continue
		}
		parent.children = append(parent.children, snapshot.id)
		if snapshot.isDir {
			parent.subdir++
		}
	}
}
