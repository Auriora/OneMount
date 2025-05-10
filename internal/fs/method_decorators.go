package fs

import (
	"fmt"
	"github.com/auriora/onemount/pkg/logging"
	"github.com/auriora/onemount/pkg/util"
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
	// Format is typically: github.com/auriora/onemount/internal/fs.(*Filesystem).MethodName
	methodName := fullName
	if lastDot := lastIndexDot(fullName); lastDot >= 0 {
		methodName = fullName[lastDot+1:]
	}

	// Get the current goroutine ID
	goroutineID := util.GetCurrentGoroutineID()

	// Log method entry
	log.Debug().
		Str(logging.FieldMethod, methodName).
		Str(logging.FieldPhase, logging.PhaseEntry).
		Str(logging.FieldGoroutine, goroutineID).
		Msg(logging.MsgMethodCalled)

	return methodName, time.Now()
}

// LogMethodReturn logs the exit of a method with its return values
// This should be deferred at the beginning of each public method
func LogMethodReturn(methodName string, startTime time.Time, returns ...interface{}) {
	duration := time.Since(startTime)

	// Get the current goroutine ID
	goroutineID := util.GetCurrentGoroutineID()

	// Create log event
	event := log.Debug().
		Str(logging.FieldMethod, methodName).
		Str(logging.FieldPhase, logging.PhaseExit).
		Str(logging.FieldGoroutine, goroutineID).
		Dur(logging.FieldDuration, duration)

	// Log return values if any
	for i, ret := range returns {
		if ret == nil {
			event = event.Interface(logging.FieldReturn+fmt.Sprintf("%d", i+1), nil)
		} else {
			// Special handling for Inode objects to prevent race conditions during JSON serialization
			if inode, ok := ret.(*Inode); ok {
				// Only log the ID and name instead of the entire object
				if inode != nil {
					event = event.Str(logging.FieldReturn+fmt.Sprintf("%d", i+1)+".id", inode.ID()).
						Str(logging.FieldReturn+fmt.Sprintf("%d", i+1)+".name", inode.Name()).
						Bool(logging.FieldReturn+fmt.Sprintf("%d", i+1)+".isDir", inode.IsDir())
				} else {
					event = event.Interface(logging.FieldReturn+fmt.Sprintf("%d", i+1), nil)
				}
			} else {
				event = event.Interface(logging.FieldReturn+fmt.Sprintf("%d", i+1), ret)
			}
		}
	}

	event.Msg(logging.MsgMethodCompleted)
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
