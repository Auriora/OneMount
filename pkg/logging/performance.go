// Package logging provides standardized logging utilities for the OneMount project.
// This file defines performance optimization utilities for logging.
package logging

import (
	"reflect"
	"sync"
)

// typeCache is used to cache type information for reflection-based logging
var typeCache = make(map[reflect.Type]string)
var typeCacheMutex sync.RWMutex

// getTypeName returns the name of a type, using a cache for performance
// This function is used to optimize reflection-based logging
func getTypeName(t reflect.Type) string {
	typeCacheMutex.RLock()
	name, ok := typeCache[t]
	typeCacheMutex.RUnlock()

	if ok {
		return name
	}

	// Compute type name
	name = t.String()

	typeCacheMutex.Lock()
	typeCache[t] = name
	typeCacheMutex.Unlock()

	return name
}

// isDebugEnabled returns true if debug logging is enabled
// This function is used to avoid expensive logging operations when debug is disabled
func isDebugEnabled() bool {
	return Debug().Enabled()
}

// isTraceEnabled returns true if trace logging is enabled
// This function is used to avoid expensive logging operations when trace is disabled
func isTraceEnabled() bool {
	return Trace().Enabled()
}

// LogComplexObjectIfDebug logs a complex object only if debug logging is enabled,
// This function is used to avoid expensive serialization when debug is disabled
func LogComplexObjectIfDebug(fieldName string, obj interface{}, msg string) {
	if isDebugEnabled() {
		Debug().
			Interface(fieldName, obj).
			Msg(msg)
	}
}

// LogComplexObjectIfTrace logs a complex object only if trace logging is enabled,
// This function is used to avoid expensive serialization when trace is disabled
func LogComplexObjectIfTrace(fieldName string, obj interface{}, msg string) {
	if isTraceEnabled() {
		Trace().
			Interface(fieldName, obj).
			Msg(msg)
	}
}
