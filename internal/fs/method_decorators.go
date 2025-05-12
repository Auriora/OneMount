package fs

import (
	"github.com/auriora/onemount/pkg/logging"
	"runtime"
	"time"
)

// LogMethodCall wraps a function call with entry and exit logging
// This is a helper function to be used in each public method
// Deprecated: Use logging.LogMethodEntry instead
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

	// Use the new logging.LogMethodEntry function
	return logging.LogMethodEntry(methodName)
}

// LogMethodReturn logs the exit of a method with its return values
// This should be deferred at the beginning of each public method
// Deprecated: Use logging.LogMethodExit instead
func LogMethodReturn(methodName string, startTime time.Time, returns ...interface{}) {
	// Use the new logging.LogMethodExit function
	logging.LogMethodExit(methodName, time.Since(startTime), returns...)
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
