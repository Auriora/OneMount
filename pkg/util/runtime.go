package util

import (
	"bytes"
	"runtime"
)

// GetCurrentGoroutineID returns the ID of the current goroutine
func GetCurrentGoroutineID() string {
	buf := make([]byte, 64)
	n := runtime.Stack(buf, false)
	// The format of the first line is "goroutine N [state]:"
	// where N is the goroutine ID
	buf = buf[:n]

	// Find the first space (after "goroutine")
	idStart := bytes.IndexByte(buf, ' ') + 1
	if idStart <= 0 {
		return "unknown"
	}

	// Find the next space or '[' (before "[state]:")
	idEnd := bytes.IndexAny(buf[idStart:], " [")
	if idEnd <= 0 {
		return "unknown"
	}

	return string(buf[idStart : idStart+idEnd])
}
