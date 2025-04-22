package fs

import (
	"github.com/rs/zerolog/log"
)

// OfflineMode represents the state of the filesystem's offline mode
type OfflineMode int

const (
	// OfflineModeDisabled means the filesystem is operating normally with network connectivity
	OfflineModeDisabled OfflineMode = iota

	// OfflineModeReadOnly means the filesystem is in offline mode but only allows read operations
	OfflineModeReadOnly

	// OfflineModeReadWrite means the filesystem is in offline mode but allows both read and write operations
	OfflineModeReadWrite
)

// SetOfflineMode sets the offline mode
func (f *Filesystem) SetOfflineMode(mode OfflineMode) {
	f.Lock()
	defer f.Unlock()

	switch mode {
	case OfflineModeDisabled:
		f.offline = false
		log.Info().Msg("Offline mode disabled")
	case OfflineModeReadOnly, OfflineModeReadWrite:
		f.offline = true
		var modeStr string
		if mode == OfflineModeReadOnly {
			modeStr = "read-only"
		} else {
			modeStr = "read-write"
		}
		log.Info().Str("mode", modeStr).Msg("Offline mode enabled")
	}
}

// GetOfflineMode returns the current offline mode
func (f *Filesystem) GetOfflineMode() OfflineMode {
	f.RLock()
	defer f.RUnlock()

	if !f.offline {
		return OfflineModeDisabled
	}

	// In the current implementation, we don't have a way to distinguish between
	// read-only and read-write modes, so we'll default to read-write
	return OfflineModeReadWrite
}

// IsReadOnly returns true if the filesystem is in read-only offline mode
func (f *Filesystem) IsReadOnly() bool {
	return f.GetOfflineMode() == OfflineModeReadOnly
}
