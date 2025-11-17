package fs

import "strings"

// childSnapshot captures metadata about a child inode so callers can safely
// update parent structures without re-locking each child while holding the
// parent lock.
type childSnapshot struct {
	inode     *Inode
	lowerName string
	id        string
	isDir     bool
}

func newChildSnapshot(inode *Inode) childSnapshot {
	if inode == nil {
		return childSnapshot{}
	}
	return childSnapshot{
		inode:     inode,
		lowerName: strings.ToLower(inode.Name()),
		id:        inode.ID(),
		isDir:     inode.IsDir(),
	}
}

func snapshotChildrenFromMap(children map[string]*Inode) []childSnapshot {
	snapshots := make([]childSnapshot, 0, len(children))
	for _, child := range children {
		snapshots = append(snapshots, newChildSnapshot(child))
	}
	return snapshots
}
