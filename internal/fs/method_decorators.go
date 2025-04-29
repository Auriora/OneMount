package fs

import (
	"bytes"
	"fmt"
	"runtime"
	"time"

	"github.com/rs/zerolog/log"
)

// LogMethodCall wraps a function call with entry and exit logging
// This is a helper function to be used in each public method
func LogMethodCall() (string, time.Time) {
	// Get the caller function name
	pc, _, _, _ := runtime.Caller(1)
	funcObj := runtime.FuncForPC(pc)
	if funcObj == nil {
		return "unknown", time.Now()
	}

	fullName := funcObj.Name()
	// Extract just the method name from the full function name
	// Format is typically: github.com/bcherrington/onedriver/internal/fs.(*Filesystem).MethodName
	methodName := fullName
	if lastDot := lastIndexDot(fullName); lastDot >= 0 {
		methodName = fullName[lastDot+1:]
	}

	// Get the current goroutine ID
	goroutineID := getCurrentGoroutineID()

	// Log method entry
	log.Debug().
		Str(FieldMethod, methodName).
		Str(FieldPhase, PhaseEntry).
		Str(FieldGoroutine, goroutineID).
		Msg(MsgMethodCalled)

	return methodName, time.Now()
}

// LogMethodReturn logs the exit of a method with its return values
// This should be deferred at the beginning of each public method
func LogMethodReturn(methodName string, startTime time.Time, returns ...interface{}) {
	duration := time.Since(startTime)

	// Get the current goroutine ID
	goroutineID := getCurrentGoroutineID()

	// Create log event
	event := log.Debug().
		Str(FieldMethod, methodName).
		Str(FieldPhase, PhaseExit).
		Str(FieldGoroutine, goroutineID).
		Dur(FieldDuration, duration)

	// Log return values if any
	for i, ret := range returns {
		if ret == nil {
			event = event.Interface(FieldReturn+fmt.Sprintf("%d", i+1), nil)
		} else {
			// Special handling for Inode objects to prevent race conditions during JSON serialization
			if inode, ok := ret.(*Inode); ok {
				// Only log the ID and name instead of the entire object
				if inode != nil {
					event = event.Str(FieldReturn+fmt.Sprintf("%d", i+1)+".id", inode.ID()).
						Str(FieldReturn+fmt.Sprintf("%d", i+1)+".name", inode.Name()).
						Bool(FieldReturn+fmt.Sprintf("%d", i+1)+".isDir", inode.IsDir())
				} else {
					event = event.Interface(FieldReturn+fmt.Sprintf("%d", i+1), nil)
				}
			} else {
				event = event.Interface(FieldReturn+fmt.Sprintf("%d", i+1), ret)
			}
		}
	}

	event.Msg(MsgMethodCompleted)
}

// lastIndexDot returns the last index of '.' in the string
func lastIndexDot(s string) int {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == '.' {
			return i
		}
	}
	return -1
}

// getCurrentGoroutineID returns the ID of the current goroutine
func getCurrentGoroutineID() string {
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

// FilesystemMethodsToInstrument returns a list of public methods in the Filesystem struct
// This is used for documentation purposes
func FilesystemMethodsToInstrument() []string {
	return []string{
		"IsOffline",
		"TrackOfflineChange",
		"ProcessOfflineChanges",
		"TranslateID",
		"GetNodeID",
		"InsertNodeID",
		"GetID",
		"InsertID",
		"InsertChild",
		"DeleteID",
		"GetChild",
		"GetChildrenID",
		"GetChildrenPath",
		"GetPath",
		"DeletePath",
		"InsertPath",
		"MoveID",
		"MovePath",
		"StartCacheCleanup",
		"StopCacheCleanup",
		"StopDeltaLoop",
		"StopDownloadManager",
		"StopUploadManager",
		"SerializeAll",
	}
}

// InodeMethodsToInstrument returns a list of public methods in the Inode struct
// This is used for documentation purposes
func InodeMethodsToInstrument() []string {
	return []string{
		"AsJSON",
		"String",
		"Name",
		"SetName",
		"NodeID",
		"SetNodeID",
		"ID",
		"ParentID",
		"Path",
		"HasChanges",
		"HasChildren",
		"IsDir",
		"Mode",
		"ModTime",
		"NLink",
		"Size",
	}
}
