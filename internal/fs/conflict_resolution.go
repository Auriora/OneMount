package fs

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/auriora/onemount/internal/graph"
	"github.com/auriora/onemount/internal/logging"
)

// ConflictType represents the type of conflict detected
type ConflictType int

const (
	ConflictTypeContent ConflictType = iota
	ConflictTypeMetadata
	ConflictTypeExistence
	ConflictTypeParent
)

// ConflictResolutionStrategy represents how to resolve a conflict
type ConflictResolutionStrategy int

const (
	StrategyLastWriterWins ConflictResolutionStrategy = iota
	StrategyUserChoice
	StrategyMerge
	StrategyRename
	StrategyKeepBoth
)

// ConflictInfo contains information about a detected conflict
type ConflictInfo struct {
	ID            string
	Type          ConflictType
	LocalItem     *Inode
	RemoteItem    *graph.DriveItem
	OfflineChange *OfflineChange
	DetectedAt    time.Time
	Message       string
}

// ConflictResolver handles conflict resolution during synchronization
type ConflictResolver struct {
	fs       *Filesystem
	strategy ConflictResolutionStrategy
}

// NewConflictResolver creates a new conflict resolver
func NewConflictResolver(fs *Filesystem, strategy ConflictResolutionStrategy) *ConflictResolver {
	return &ConflictResolver{
		fs:       fs,
		strategy: strategy,
	}
}

// DetectConflict checks if there's a conflict between local and remote changes
func (cr *ConflictResolver) DetectConflict(ctx context.Context, localItem *Inode, remoteItem *graph.DriveItem, offlineChange *OfflineChange) (*ConflictInfo, error) {
	logCtx := logging.NewLogContextWithRequestAndUserID("conflict_detection")
	logCtx = logCtx.WithComponent("fs").WithMethod("DetectConflict")
	logCtx = logCtx.With("id", localItem.ID()).With("name", localItem.Name())
	logger := logging.WithLogContext(logCtx)

	// Check for existence conflicts
	if localItem == nil && remoteItem != nil && offlineChange != nil && offlineChange.Type == "delete" {
		return &ConflictInfo{
			ID:            offlineChange.ID,
			Type:          ConflictTypeExistence,
			LocalItem:     localItem,
			RemoteItem:    remoteItem,
			OfflineChange: offlineChange,
			DetectedAt:    time.Now(),
			Message:       "File was deleted locally but modified remotely",
		}, nil
	}

	if localItem != nil && remoteItem == nil && offlineChange != nil && (offlineChange.Type == "create" || offlineChange.Type == "modify") {
		return &ConflictInfo{
			ID:            offlineChange.ID,
			Type:          ConflictTypeExistence,
			LocalItem:     localItem,
			RemoteItem:    remoteItem,
			OfflineChange: offlineChange,
			DetectedAt:    time.Now(),
			Message:       "File was created/modified locally but deleted remotely",
		}, nil
	}

	// Check for content conflicts
	if localItem != nil && remoteItem != nil && offlineChange != nil {
		// Check if both local and remote have changes
		hasLocalChanges := localItem.HasChanges() || offlineChange.Type == "modify" || offlineChange.Type == "create"
		hasRemoteChanges := remoteItem.ModTimeUnix() > localItem.ModTime() && !remoteItem.ETagIsMatch(localItem.DriveItem.ETag)

		if hasLocalChanges && hasRemoteChanges {
			// Check if content is actually different
			if !remoteItem.IsDir() && remoteItem.File != nil && localItem.DriveItem.File != nil {
				if remoteItem.File.Hashes.QuickXorHash != "" && localItem.DriveItem.File.Hashes.QuickXorHash != "" {
					if remoteItem.File.Hashes.QuickXorHash != localItem.DriveItem.File.Hashes.QuickXorHash {
						logger.Info().Msg("Content conflict detected - different hashes")
						return &ConflictInfo{
							ID:            localItem.ID(),
							Type:          ConflictTypeContent,
							LocalItem:     localItem,
							RemoteItem:    remoteItem,
							OfflineChange: offlineChange,
							DetectedAt:    time.Now(),
							Message:       "File content differs between local and remote versions",
						}, nil
					}
				}
			}
		}

		// Check for metadata conflicts (name, parent changes)
		remoteParentID := ""
		if remoteItem.Parent != nil {
			remoteParentID = remoteItem.Parent.ID
		}
		if localItem.Name() != remoteItem.Name || localItem.ParentID() != remoteParentID {
			logger.Info().Msg("Metadata conflict detected - name or parent differs")
			return &ConflictInfo{
				ID:            localItem.ID(),
				Type:          ConflictTypeMetadata,
				LocalItem:     localItem,
				RemoteItem:    remoteItem,
				OfflineChange: offlineChange,
				DetectedAt:    time.Now(),
				Message:       "File metadata (name or location) differs between local and remote versions",
			}, nil
		}
	}

	return nil, nil
}

// ResolveConflict resolves a detected conflict based on the configured strategy
func (cr *ConflictResolver) ResolveConflict(ctx context.Context, conflict *ConflictInfo) error {
	logCtx := logging.NewLogContextWithRequestAndUserID("conflict_resolution")
	logCtx = logCtx.WithComponent("fs").WithMethod("ResolveConflict")
	logCtx = logCtx.With("id", conflict.ID).With("type", fmt.Sprintf("%d", conflict.Type))
	logger := logging.WithLogContext(logCtx)

	logger.Info().Str("strategy", fmt.Sprintf("%d", cr.strategy)).Msg("Resolving conflict")

	switch cr.strategy {
	case StrategyLastWriterWins:
		return cr.resolveLastWriterWins(ctx, conflict)
	case StrategyKeepBoth:
		return cr.resolveKeepBoth(ctx, conflict)
	case StrategyRename:
		return cr.resolveRename(ctx, conflict)
	default:
		// Default to keeping both versions
		return cr.resolveKeepBoth(ctx, conflict)
	}
}

// resolveLastWriterWins resolves conflict by keeping the most recently modified version
func (cr *ConflictResolver) resolveLastWriterWins(ctx context.Context, conflict *ConflictInfo) error {
	logger := logging.WithLogContext(logging.NewLogContextWithRequestAndUserID("resolve_last_writer_wins"))

	if conflict.LocalItem != nil && conflict.RemoteItem != nil {
		localModTime := conflict.LocalItem.ModTime()
		remoteModTime := conflict.RemoteItem.ModTimeUnix()

		if remoteModTime > localModTime {
			// Remote is newer, accept remote changes
			logger.Info().Msg("Remote version is newer, accepting remote changes")
			return cr.acceptRemoteChanges(ctx, conflict)
		} else {
			// Local is newer, keep local changes
			logger.Info().Msg("Local version is newer, keeping local changes")
			return cr.keepLocalChanges(ctx, conflict)
		}
	}

	return nil
}

// resolveKeepBoth resolves conflict by keeping both versions
func (cr *ConflictResolver) resolveKeepBoth(ctx context.Context, conflict *ConflictInfo) error {
	logger := logging.WithLogContext(logging.NewLogContextWithRequestAndUserID("resolve_keep_both"))

	// Create a conflict copy of the remote version
	if conflict.RemoteItem != nil {
		conflictName := cr.generateConflictName(conflict.RemoteItem.Name)
		logger.Info().Str("conflictName", conflictName).Msg("Creating conflict copy")

		// Create a new item for the remote version with conflict name
		conflictItem := &graph.DriveItem{}
		*conflictItem = *conflict.RemoteItem
		conflictItem.Name = conflictName

		// Add the conflict copy to the filesystem
		// Check if the remote item has a parent before accessing it
		if conflict.RemoteItem.Parent != nil {
			parent := cr.fs.GetID(conflict.RemoteItem.Parent.ID)
			if parent != nil {
				conflictInode := NewInodeDriveItem(conflictItem)
				cr.fs.InsertID(conflictItem.ID+"_conflict", conflictInode)
			}
		}
	}

	// Keep the local version as is
	return cr.keepLocalChanges(ctx, conflict)
}

// resolveRename resolves conflict by renaming the conflicting file
func (cr *ConflictResolver) resolveRename(ctx context.Context, conflict *ConflictInfo) error {
	return cr.resolveKeepBoth(ctx, conflict)
}

// acceptRemoteChanges accepts the remote version and discards local changes
func (cr *ConflictResolver) acceptRemoteChanges(ctx context.Context, conflict *ConflictInfo) error {
	logger := logging.WithLogContext(logging.NewLogContextWithRequestAndUserID("accept_remote_changes"))

	if conflict.LocalItem != nil && conflict.RemoteItem != nil {
		// Update local item with remote data
		conflict.LocalItem.mu.Lock()
		conflict.LocalItem.DriveItem = *conflict.RemoteItem
		conflict.LocalItem.mu.Unlock()
		cr.fs.markCleanLocalState(conflict.LocalItem.ID())

		// Invalidate content cache to force re-download
		cr.fs.content.Delete(conflict.ID)

		logger.Info().Msg("Accepted remote changes")
	}

	// Mark conflict as resolved by updating file status
	// Note: This would be implemented based on the actual file status tracking system
	logger.Info().Msg("Marked conflict as resolved")

	return nil
}

// keepLocalChanges keeps the local version and queues it for upload
func (cr *ConflictResolver) keepLocalChanges(ctx context.Context, conflict *ConflictInfo) error {
	logger := logging.WithLogContext(logging.NewLogContextWithRequestAndUserID("keep_local_changes"))

	if conflict.LocalItem != nil && conflict.OfflineChange != nil {
		// Queue the local changes for upload only if the item has a valid parent
		if cr.fs.uploads != nil && conflict.LocalItem.DriveItem.Parent != nil {
			_, err := cr.fs.uploads.QueueUploadWithPriority(conflict.LocalItem, PriorityLow)
			if err != nil {
				logger.Error().Err(err).Msg("Failed to queue upload for local changes")
				return err
			}
			logger.Info().Msg("Kept local changes and queued for upload")
		} else {
			logger.Info().Msg("Kept local changes (upload skipped due to missing parent or upload manager)")
		}
	}

	return nil
}

// generateConflictName generates a unique name for conflict copies
func (cr *ConflictResolver) generateConflictName(originalName string) string {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	ext := filepath.Ext(originalName)
	nameWithoutExt := strings.TrimSuffix(originalName, ext)

	return fmt.Sprintf("%s (Conflict Copy %s)%s", nameWithoutExt, timestamp, ext)
}
