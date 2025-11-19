package fs

import "time"

// RealtimeOptions controls Socket.IO subscription behaviour.
type RealtimeOptions struct {
	Enabled          bool
	PollingOnly      bool
	ClientState      string
	Resource         string
	FallbackInterval time.Duration
}

// ConfigureRealtime stores the realtime options for the filesystem. Must be invoked before DeltaLoop starts.
func (f *Filesystem) ConfigureRealtime(opts RealtimeOptions) {
	f.realtimeOptions = &opts
}
