package fs

import "time"

// DeltaTuning configures adaptive delta loop behavior.
type DeltaTuning struct {
	ActiveInterval time.Duration
	ActiveWindow   time.Duration
}

// ConfigureDeltaTuning stores the delta tuning options for the filesystem.
func (f *Filesystem) ConfigureDeltaTuning(opts DeltaTuning) {
	if opts.ActiveInterval > 0 {
		f.activeDeltaInterval = opts.ActiveInterval
	}
	if opts.ActiveWindow > 0 {
		f.activeDeltaWindow = opts.ActiveWindow
	}
}

// RecordForegroundActivity marks the time a foreground metadata request was
// scheduled so the delta loop can temporarily switch to the faster interval.
func (f *Filesystem) RecordForegroundActivity() {
	if f == nil {
		return
	}
	if f.activeDeltaInterval <= 0 || f.activeDeltaWindow <= 0 {
		return
	}
	f.lastForegroundActivity.Store(time.Now().UnixNano())
}
