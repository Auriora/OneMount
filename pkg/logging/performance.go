// Package logging provides standardized logging utilities for the OneMount project.
// This file defines performance optimization utilities for logging.
//
// Performance is a critical consideration in logging, especially for high-throughput
// applications. This file provides utilities to optimize logging performance, including:
//
//   - Type name caching to reduce reflection overhead
//   - Level checking functions to avoid expensive logging operations when not needed
//   - Conditional logging functions for complex objects
//
// These utilities help maintain good application performance while still providing
// comprehensive logging capabilities.
//
// This file is part of the consolidated logging package structure, which includes:
//   - logger.go: Core logger implementation and level management
//   - context.go: Context-aware logging functionality
//   - method.go: Method entry/exit logging (both with and without context)
//   - error.go: Error logging functionality
//   - performance.go (this file): Performance optimization utilities
package logging

import (
	"github.com/rs/zerolog"
	"reflect"
)

// Note: Type caching has been moved to type_cache.go
// The functions below are deprecated and will be removed in a future release.
// Use the functions in type_cache.go instead.

// getTypeName returns the name of a type, using a cache for performance
// Deprecated: Use GetTypeName from type_cache.go instead
func getTypeName(t reflect.Type) string {
	return GetTypeName(t)
}

// getTypeKind returns the kind of a type, using a cache for performance
// Deprecated: Use GetTypeKind from type_cache.go instead
func getTypeKind(t reflect.Type) reflect.Kind {
	return GetTypeKind(t)
}

// getTypeElem returns the element type of a pointer or slice, using a cache for performance
// Deprecated: Use GetTypeElem from type_cache.go instead
func getTypeElem(t reflect.Type) reflect.Type {
	return GetTypeElem(t)
}

// isPointerToByteSlice checks if a type is a pointer to a byte slice
// Deprecated: Use IsPointerToByteSlice from type_cache.go instead
func isPointerToByteSlice(t reflect.Type) bool {
	return IsPointerToByteSlice(t)
}

// Note: The functions isDebugEnabled and isTraceEnabled have been removed.
// Use IsDebugEnabled and IsTraceEnabled from logger.go instead, or
// use the more general IsLevelEnabled(DebugLevel) and IsLevelEnabled(TraceLevel).

// LogComplexObjectIfDebug logs a complex object only if debug logging is enabled,
// This function is used to avoid expensive serialization when debug is disabled
func LogComplexObjectIfDebug(fieldName string, obj interface{}, msg string) {
	if IsDebugEnabled() {
		Debug().
			Interface(fieldName, obj).
			Msg(msg)
	}
}

// LogComplexObjectIfTrace logs a complex object only if trace logging is enabled,
// This function is used to avoid expensive serialization when trace is disabled
func LogComplexObjectIfTrace(fieldName string, obj interface{}, msg string) {
	if IsTraceEnabled() {
		Trace().
			Interface(fieldName, obj).
			Msg(msg)
	}
}

// IsLevelEnabled returns true if the specified log level is enabled
// This function is used to check if a specific log level is enabled before performing expensive operations
func IsLevelEnabled(level Level) bool {
	return zerolog.GlobalLevel() <= zerolog.Level(level)
}

// LogIfEnabled executes the provided function only if the specified log level is enabled
// This is useful for expensive logging operations that should only be performed if the level is enabled
func LogIfEnabled(level Level, logFn func()) {
	if IsLevelEnabled(level) {
		logFn()
	}
}

// LogComplexObjectIfEnabled logs a complex object only if the specified level is enabled
// This is a generalized version of LogComplexObjectIfDebug and LogComplexObjectIfTrace
func LogComplexObjectIfEnabled(level Level, fieldName string, obj interface{}, msg string) {
	if IsLevelEnabled(level) {
		switch level {
		case DebugLevel:
			Debug().Interface(fieldName, obj).Msg(msg)
		case TraceLevel:
			Trace().Interface(fieldName, obj).Msg(msg)
		case InfoLevel:
			Info().Interface(fieldName, obj).Msg(msg)
		case WarnLevel:
			Warn().Interface(fieldName, obj).Msg(msg)
		case ErrorLevel:
			Error().Interface(fieldName, obj).Msg(msg)
		default:
			// For other levels, use the default logger
			Log().Interface(fieldName, obj).Msg(msg)
		}
	}
}

// Performance Optimization Notes:
//
// The logging package has been optimized for performance in several ways:
//
// 1. Level Checks:
//    - Added IsLevelEnabled function to check if a specific log level is enabled
//    - Added level checks to all method logging functions to avoid expensive operations
//      when the corresponding log level is disabled
//    - Added helper functions (LogIfEnabled, LogComplexObjectIfEnabled) for common patterns
//
// 2. Type Caching:
//    - Enhanced the type caching mechanism to cache more type information (name, kind, element type)
//    - Added helper functions (getTypeKind, getTypeElem, isPointerToByteSlice) to efficiently
//      retrieve and check type information
//
// 3. Reflection Optimization:
//    - Reduced reflection operations by caching type information
//    - Added specialized handling for common types (byte slices, structs)
//    - Used more efficient type checks with cached information
//
// These optimizations should significantly reduce the overhead of logging, especially
// when debug or trace logging is disabled in production environments.
