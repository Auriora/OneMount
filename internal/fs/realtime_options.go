package fs

import "time"

// RealtimeOptions controls Socket.IO subscription behavior for the filesystem.
// These options determine how the filesystem connects to Microsoft Graph's realtime
// notification system to receive immediate change notifications.
type RealtimeOptions struct {
	// Enabled controls whether realtime notifications are active.
	// When false, the filesystem relies entirely on periodic delta polling.
	Enabled bool

	// PollingOnly forces delta polling even when realtime subscriptions are configured.
	// This disables the Socket.IO transport while keeping the realtime infrastructure active.
	// Useful for debugging or environments where WebSocket connections are problematic.
	PollingOnly bool

	// ClientState is a validation token echoed in notification events.
	// Used to verify that notifications are intended for this client instance.
	ClientState string

	// Resource specifies the Microsoft Graph resource path to monitor for changes.
	// Examples: "/me/drive/root" (personal OneDrive), "/drives/{drive-id}" (shared drive).
	Resource string

	// FallbackInterval is the polling interval when Socket.IO is unavailable or degraded.
	// When Socket.IO is healthy, polling occurs much less frequently.
	// When Socket.IO fails, polling falls back to this interval.
	FallbackInterval time.Duration
}

// ConfigureRealtime stores the realtime options for the filesystem.
// This method must be called before starting the delta loop to enable realtime notifications.
// The options control how the filesystem connects to Microsoft Graph's Socket.IO endpoint
// and how it falls back to polling when the realtime connection is unavailable.
func (f *Filesystem) ConfigureRealtime(opts RealtimeOptions) {
	f.realtimeOptions = &opts
}
