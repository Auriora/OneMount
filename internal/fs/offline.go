package fs

import (
	"github.com/auriora/onemount/internal/logging"
)

// OfflineMode represents the state of the filesystem's offline mode
type OfflineMode int

const (
	// OfflineModeDisabled means the filesystem is operating normally with network connectivity
	OfflineModeDisabled OfflineMode = iota

	// OfflineModeReadWrite means the filesystem is in offline mode
	// The offline state only affects the process to download and upload updates to OneDrive cloud
	OfflineModeReadWrite
)

// SetOfflineMode sets the offline mode
func (f *Filesystem) SetOfflineMode(mode OfflineMode) {
	f.Lock()
	defer f.Unlock()

	switch mode {
	case OfflineModeDisabled:
		f.offline = false
		logging.Info().Msg("Offline mode disabled")
	case OfflineModeReadWrite:
		f.offline = true
		logging.Info().Msg("Offline mode enabled")
	}
}

// GetOfflineMode returns the current offline mode
func (f *Filesystem) GetOfflineMode() OfflineMode {
	f.RLock()
	defer f.RUnlock()

	if !f.offline {
		return OfflineModeDisabled
	}

	return OfflineModeReadWrite
}
