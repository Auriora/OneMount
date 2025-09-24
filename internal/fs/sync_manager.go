package fs

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/auriora/onemount/internal/graph"
	"github.com/auriora/onemount/internal/logging"
	"github.com/auriora/onemount/internal/retry"
)

// SyncManager handles synchronization with retry mechanisms and error recovery
type SyncManager struct {
	fs               *Filesystem
	conflictResolver *ConflictResolver
	retryConfig      retry.Config
}

// SyncResult represents the result of a synchronization operation
type SyncResult struct {
	ProcessedChanges  int
	ConflictsFound    int
	ConflictsResolved int
	Errors            []error
	Duration          time.Duration
}

// NewSyncManager creates a new synchronization manager
func NewSyncManager(fs *Filesystem) *SyncManager {
	// Configure retry settings for synchronization
	retryConfig := retry.Config{
		MaxRetries:   5,
		InitialDelay: 1 * time.Second,
		MaxDelay:     30 * time.Second,
		Multiplier:   2.0,
		Jitter:       0.1,
		RetryableErrors: []retry.RetryableError{
			retry.IsRetryableNetworkError,
			retry.IsRetryableServerError,
			retry.IsRetryableRateLimitError,
		},
	}

	conflictResolver := NewConflictResolver(fs, StrategyKeepBoth)

	return &SyncManager{
		fs:               fs,
		conflictResolver: conflictResolver,
		retryConfig:      retryConfig,
	}
}

// ProcessOfflineChangesWithRetry processes offline changes with comprehensive error handling and retry logic
func (sm *SyncManager) ProcessOfflineChangesWithRetry(ctx context.Context) (*SyncResult, error) {
	startTime := time.Now()
	result := &SyncResult{
		ProcessedChanges:  0,
		ConflictsFound:    0,
		ConflictsResolved: 0,
		Errors:            make([]error, 0),
	}

	logCtx := logging.NewLogContextWithRequestAndUserID("sync_manager")
	logCtx = logCtx.WithComponent("fs").WithMethod("ProcessOfflineChangesWithRetry")
	logger := logging.WithLogContext(logCtx)

	logger.Info().Msg("Starting offline changes synchronization with retry logic")

	// Get all offline changes with retry
	changes, err := sm.getOfflineChangesWithRetry(ctx)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Errorf("failed to get offline changes: %w", err))
		result.Duration = time.Since(startTime)
		return result, err
	}

	logger.Info().Int("changeCount", len(changes)).Msg("Retrieved offline changes")

	// Process each change with individual retry logic
	for _, change := range changes {
		select {
		case <-ctx.Done():
			logger.Debug().Msg("Context cancelled during change processing")
			result.Duration = time.Since(startTime)
			return result, ctx.Err()
		default:
			// Continue processing
		}

		err := sm.processChangeWithRetry(ctx, change, result)
		if err != nil {
			logger.Error().Err(err).
				Str("changeID", change.ID).
				Str("changeType", change.Type).
				Msg("Failed to process change after retries")
			result.Errors = append(result.Errors, err)
			// Continue with other changes instead of failing completely
		} else {
			result.ProcessedChanges++
		}
	}

	result.Duration = time.Since(startTime)
	logger.Info().
		Int("processed", result.ProcessedChanges).
		Int("conflicts", result.ConflictsFound).
		Int("resolved", result.ConflictsResolved).
		Int("errors", len(result.Errors)).
		Dur("duration", result.Duration).
		Msg("Completed offline changes synchronization")

	return result, nil
}

// getOfflineChangesWithRetry retrieves offline changes with retry logic
func (sm *SyncManager) getOfflineChangesWithRetry(ctx context.Context) ([]*OfflineChange, error) {
	return retry.DoWithResult(ctx, func() ([]*OfflineChange, error) {
		return sm.fs.getOfflineChanges(ctx)
	}, sm.retryConfig)
}

// processChangeWithRetry processes a single offline change with retry logic and conflict detection
func (sm *SyncManager) processChangeWithRetry(ctx context.Context, change *OfflineChange, result *SyncResult) error {
	logCtx := logging.NewLogContextWithRequestAndUserID("process_change")
	logCtx = logCtx.WithComponent("fs").WithMethod("processChangeWithRetry")
	logCtx = logCtx.With("changeID", change.ID).With("changeType", change.Type)
	logger := logging.WithLogContext(logCtx)

	return retry.Do(ctx, func() error {
		// Check for conflicts before processing
		conflict, err := sm.detectConflictForChange(ctx, change)
		if err != nil {
			return fmt.Errorf("failed to detect conflicts: %w", err)
		}

		if conflict != nil {
			result.ConflictsFound++
			logger.Info().Str("conflictType", fmt.Sprintf("%d", conflict.Type)).Msg("Conflict detected")

			// Resolve the conflict
			err = sm.conflictResolver.ResolveConflict(ctx, conflict)
			if err != nil {
				return fmt.Errorf("failed to resolve conflict: %w", err)
			}

			result.ConflictsResolved++
			logger.Info().Msg("Conflict resolved successfully")
		}

		// Process the change based on its type
		switch change.Type {
		case "create", "modify":
			return sm.processCreateOrModifyChange(ctx, change)
		case "delete":
			return sm.processDeleteChange(ctx, change)
		case "rename":
			return sm.processRenameChange(ctx, change)
		default:
			return fmt.Errorf("unknown change type: %s", change.Type)
		}
	}, sm.retryConfig)
}

// detectConflictForChange detects conflicts for a specific offline change
func (sm *SyncManager) detectConflictForChange(ctx context.Context, change *OfflineChange) (*ConflictInfo, error) {
	localItem := sm.fs.GetID(change.ID)

	// Get the current remote state
	remoteItem, err := sm.getRemoteItemWithRetry(ctx, change.ID)
	if err != nil {
		// If item doesn't exist remotely, it might be a delete conflict
		if isNotFoundError(err) {
			remoteItem = nil
		} else {
			return nil, fmt.Errorf("failed to get remote item: %w", err)
		}
	}

	return sm.conflictResolver.DetectConflict(ctx, localItem, remoteItem, change)
}

// getRemoteItemWithRetry gets remote item state with retry logic
func (sm *SyncManager) getRemoteItemWithRetry(ctx context.Context, itemID string) (*graph.DriveItem, error) {
	return retry.DoWithResult(ctx, func() (*graph.DriveItem, error) {
		return graph.GetItem(itemID, sm.fs.auth)
	}, sm.retryConfig)
}

// processCreateOrModifyChange processes create or modify changes
func (sm *SyncManager) processCreateOrModifyChange(ctx context.Context, change *OfflineChange) error {
	inode := sm.fs.GetID(change.ID)
	if inode == nil {
		return fmt.Errorf("inode not found for change ID: %s", change.ID)
	}

	// Queue upload with retry logic built into the upload manager
	if sm.fs.uploads != nil {
		_, err := sm.fs.uploads.QueueUploadWithPriority(inode, PriorityLow)
		if err != nil {
			return fmt.Errorf("failed to queue upload: %w", err)
		}
	}

	return nil
}

// processDeleteChange processes delete changes
func (sm *SyncManager) processDeleteChange(ctx context.Context, change *OfflineChange) error {
	if isLocalID(change.ID) {
		// Local-only item, nothing to delete remotely
		return nil
	}

	return retry.Do(ctx, func() error {
		return graph.Remove(change.ID, sm.fs.auth)
	}, sm.retryConfig)
}

// processRenameChange processes rename changes
func (sm *SyncManager) processRenameChange(ctx context.Context, change *OfflineChange) error {
	if change.NewPath == "" {
		return fmt.Errorf("rename change missing new path")
	}

	// For now, treat rename as a modify operation
	// More sophisticated rename handling could be implemented later
	return sm.processCreateOrModifyChange(ctx, change)
}

// RecoverFromNetworkInterruption handles recovery after network interruptions
func (sm *SyncManager) RecoverFromNetworkInterruption(ctx context.Context) error {
	logCtx := logging.NewLogContextWithRequestAndUserID("network_recovery")
	logCtx = logCtx.WithComponent("fs").WithMethod("RecoverFromNetworkInterruption")
	logger := logging.WithLogContext(logCtx)

	logger.Info().Msg("Starting network interruption recovery")

	// Wait for network to be available with exponential backoff
	err := retry.Do(ctx, func() error {
		if sm.fs.IsOffline() {
			return fmt.Errorf("network still unavailable")
		}
		return nil
	}, retry.Config{
		MaxRetries:   10,
		InitialDelay: 5 * time.Second,
		MaxDelay:     2 * time.Minute,
		Multiplier:   1.5,
		Jitter:       0.2,
		RetryableErrors: []retry.RetryableError{
			func(err error) bool { return true }, // Retry all errors for network recovery
		},
	})

	if err != nil {
		return fmt.Errorf("network recovery failed: %w", err)
	}

	logger.Info().Msg("Network recovered, processing pending changes")

	// Process any pending offline changes
	result, err := sm.ProcessOfflineChangesWithRetry(ctx)
	if err != nil {
		return fmt.Errorf("failed to process changes after network recovery: %w", err)
	}

	logger.Info().
		Int("processed", result.ProcessedChanges).
		Int("errors", len(result.Errors)).
		Msg("Network recovery completed")

	return nil
}

// GetSyncStatus returns the current synchronization status
func (sm *SyncManager) GetSyncStatus(ctx context.Context) (map[string]interface{}, error) {
	changes, err := sm.fs.getOfflineChanges(ctx)
	if err != nil {
		return nil, err
	}

	status := map[string]interface{}{
		"pending_changes": len(changes),
		"is_offline":      sm.fs.IsOffline(),
		"last_sync":       time.Now(), // This could be tracked more precisely
	}

	return status, nil
}

// isNotFoundError checks if an error indicates that a resource was not found
func isNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	errorStr := strings.ToLower(err.Error())
	return strings.Contains(errorStr, "not found") || strings.Contains(errorStr, "404")
}
