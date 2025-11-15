package fs

import "strings"

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

	// Ensure the inode has a node ID so FUSE can reference it
	f.InsertNodeID(inode)

	if parentID := inode.ParentID(); parentID != "" {
		if parent := f.GetID(parentID); parent != nil {
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

// appendVirtualChildrenLocked appends any virtual children for the given parent into the children map.
// The caller must hold the parent's lock while calling this function.
func (f *Filesystem) appendVirtualChildrenLocked(parent *Inode, children map[string]*Inode) {
	if parent == nil {
		return
	}
	// The caller must hold parent.mu before invoking this helper. Access the
	// parent's ID directly instead of via parent.ID() to avoid deadlocking on
	// the same mutex (parent.ID() takes an RLock).
	parentID := parent.DriveItem.ID
	f.virtualMu.RLock()
	defer f.virtualMu.RUnlock()
	for _, inode := range f.virtualFiles {
		if inode.ParentID() != parentID {
			continue
		}
		key := strings.ToLower(inode.Name())
		children[key] = inode
		// Ensure the parent child list contains this inode ID
		alreadyPresent := false
		inodeID := inode.ID()
		for _, childID := range parent.children {
			if childID == inodeID {
				alreadyPresent = true
				break
			}
		}
		if !alreadyPresent {
			parent.children = append(parent.children, inodeID)
		}
	}
}
