// Package logging provides standardized logging utilities for the OneMount project.
// This file defines method logging functionality, both with and without context.
//
// Method logging is a key feature of the OneMount logging system, providing automatic
// logging of method entry and exit, including parameters, return values, and execution duration.
// This helps with debugging, performance analysis, and understanding the flow of execution.
//
// The file provides two sets of functions:
//   - Standard method logging: LogMethodEntry, LogMethodExit, LoggedMethod
//   - Context-aware method logging: LogMethodEntryWithContext, LogMethodExitWithContext
//   - Helper functions: WithMethodLogging, WithMethodLoggingAndContext
//
// This file is part of the consolidated logging package structure, which includes:
//   - logger.go: Core logger implementation and level management
//   - context.go: Context-aware logging functionality
//   - method.go (this file): Method entry/exit logging (both with and without context)
//   - error.go: Error logging functionality
//   - performance.go: Performance optimization utilities
package logging

import (
	"fmt"
	"github.com/auriora/onemount/pkg/util"
	"reflect"
	"runtime"
	"strings"
	"time"
)

// LogMethodEntry logs the entry of a method with its parameters
// It returns the method name and start time for use with LogMethodExit
func LogMethodEntry(methodName string, params ...interface{}) (string, time.Time) {
	startTime := time.Now()

	// Only perform expensive operations if debug logging is enabled
	if !IsLevelEnabled(DebugLevel) {
		return methodName, startTime
	}

	event := Debug().
		Str(FieldMethod, methodName).
		Str(FieldPhase, PhaseEntry)

	// Log parameters if any
	if len(params) > 0 {
		for i, param := range params {
			// Skip logging for large data structures or sensitive information
			if param == nil {
				event = event.Interface(FieldParam+fmt.Sprintf("%d", i+1), nil)
			} else {
				// Get the type of the parameter
				paramType := reflect.TypeOf(param)

				// Handle different types of parameters
				switch {
				case isPointerToByteSlice(paramType):
					// For []byte pointers, just log the length
					byteSlice := reflect.ValueOf(param).Elem().Interface().([]byte)
					event = event.Int(FieldParam+fmt.Sprintf("%d_size", i+1), len(byteSlice))
				case strings.Contains(getTypeName(paramType), "Auth"):
					// Don't log auth objects which might contain sensitive information
					event = event.Str(FieldParam+fmt.Sprintf("%d", i+1), "[Auth object]")
				default:
					// For other types, log the value
					event = event.Interface(FieldParam+fmt.Sprintf("%d", i+1), param)
				}
			}
		}
	}

	event.Msg(MsgMethodCalled)
	return methodName, startTime
}

// LogMethodExit logs the exit of a method with its return values
func LogMethodExit(methodName string, duration time.Duration, returns ...interface{}) {
	// Only perform expensive operations if debug logging is enabled
	if !IsLevelEnabled(DebugLevel) {
		return
	}

	event := Debug().
		Str(FieldMethod, methodName).
		Str(FieldPhase, PhaseExit).
		Dur(FieldDuration, duration)

	// Log return values if any
	if len(returns) > 0 {
		for i, ret := range returns {
			// Skip logging for large data structures or sensitive information
			if ret == nil {
				event = event.Interface(FieldReturn+fmt.Sprintf("%d", i+1), nil)
			} else {
				// Get the type of the return value
				retType := reflect.TypeOf(ret)
				retKind := getTypeKind(retType)

				// Handle different types of return values
				switch {
				case isPointerToByteSlice(retType):
					// For []byte pointers, just log the length
					byteSlice := reflect.ValueOf(ret).Elem().Interface().([]byte)
					event = event.Int(FieldReturn+fmt.Sprintf("%d_size", i+1), len(byteSlice))
				case strings.Contains(getTypeName(retType), "Auth"):
					// Don't log auth objects which might contain sensitive information
					event = event.Str(FieldReturn+fmt.Sprintf("%d", i+1), "[Auth object]")
				case retKind == reflect.Struct || (retKind == reflect.Ptr && getTypeKind(getTypeElem(retType)) == reflect.Struct):
					// For structs, log a simplified representation
					if retKind == reflect.Ptr {
						if reflect.ValueOf(ret).IsNil() {
							event = event.Str(FieldReturn+fmt.Sprintf("%d", i+1), "nil")
						} else {
							typeName := getTypeElem(retType).Name()
							event = event.Str(FieldReturn+fmt.Sprintf("%d", i+1), fmt.Sprintf("[%s object]", typeName))
						}
					} else {
						typeName := retType.Name()
						event = event.Str(FieldReturn+fmt.Sprintf("%d", i+1), fmt.Sprintf("[%s object]", typeName))
					}
				default:
					// For other types, log the value
					event = event.Interface(FieldReturn+fmt.Sprintf("%d", i+1), ret)
				}
			}
		}
	}

	event.Msg(MsgMethodCompleted)
}

// LoggedMethod wraps a function call with entry and exit logging
func LoggedMethod(f interface{}, args ...interface{}) []interface{} {
	// Get the function name
	funcName := runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
	// Extract just the method name from the full function name
	parts := strings.Split(funcName, ".")
	methodName := parts[len(parts)-1]

	// Log method entry
	methodName, startTime := LogMethodEntry(methodName, args...)

	// Prepare for function call
	fValue := reflect.ValueOf(f)
	fType := fValue.Type()

	// Create input arguments
	in := make([]reflect.Value, len(args))
	for i, arg := range args {
		if arg == nil {
			in[i] = reflect.Zero(fType.In(i))
		} else {
			in[i] = reflect.ValueOf(arg)
		}
	}

	// Call the function
	out := fValue.Call(in)
	duration := time.Since(startTime)

	// Convert output to interface slice
	returns := make([]interface{}, len(out))
	for i, val := range out {
		returns[i] = val.Interface()
	}

	// Log method exit
	LogMethodExit(methodName, duration, returns...)

	return returns
}

// LogMethodEntryWithContext logs the entry of a method with context
func LogMethodEntryWithContext(methodName string, ctx LogContext) (string, time.Time, Logger, LogContext) {
	startTime := time.Now()

	// Create a logger with the context
	logger := WithLogContext(ctx)

	// Only perform expensive operations if debug logging is enabled
	if !IsLevelEnabled(DebugLevel) {
		return methodName, startTime, logger, ctx
	}

	// Get the current goroutine ID
	goroutineID := util.GetCurrentGoroutineID()

	// Log method entry
	logger.Debug().
		Str(FieldMethod, methodName).
		Str(FieldPhase, PhaseEntry).
		Str(FieldGoroutine, goroutineID).
		Msg(MsgMethodCalled)

	return methodName, startTime, logger, ctx
}

// LogMethodCallWithContext logs the entry of a method with context
// Deprecated: Use LogMethodEntryWithContext instead
func LogMethodCallWithContext(methodName string, ctx LogContext) (string, time.Time, Logger, LogContext) {
	return LogMethodEntryWithContext(methodName, ctx)
}

// LogMethodExitWithContext logs the exit of a method with context
func LogMethodExitWithContext(methodName string, startTime time.Time, logger Logger, ctx LogContext, returns ...interface{}) {
	// Only perform expensive operations if debug logging is enabled
	if !IsLevelEnabled(DebugLevel) {
		return
	}

	duration := time.Since(startTime)

	// Get the current goroutine ID
	goroutineID := util.GetCurrentGoroutineID()

	// Create log event
	event := logger.Debug().
		Str(FieldMethod, methodName).
		Str(FieldPhase, PhaseExit).
		Str(FieldGoroutine, goroutineID).
		Dur(FieldDuration, duration)

	// Log return values if any
	for i, ret := range returns {
		if ret == nil {
			event = event.Interface(FieldReturn+fmt.Sprintf("%d", i+1), nil)
		} else {
			// Get the type of the return value
			retType := reflect.TypeOf(ret)
			retKind := getTypeKind(retType)

			// Handle different types of return values
			switch {
			case isPointerToByteSlice(retType):
				// For []byte pointers, just log the length
				byteSlice := reflect.ValueOf(ret).Elem().Interface().([]byte)
				event = event.Int(FieldReturn+fmt.Sprintf("%d_size", i+1), len(byteSlice))
			case strings.Contains(getTypeName(retType), "Auth"):
				// Don't log auth objects which might contain sensitive information
				event = event.Str(FieldReturn+fmt.Sprintf("%d", i+1), "[Auth object]")
			case retKind == reflect.Struct || (retKind == reflect.Ptr && getTypeKind(getTypeElem(retType)) == reflect.Struct):
				// For structs, log a simplified representation
				if retKind == reflect.Ptr {
					if reflect.ValueOf(ret).IsNil() {
						event = event.Str(FieldReturn+fmt.Sprintf("%d", i+1), "nil")
					} else {
						typeName := getTypeElem(retType).Name()
						event = event.Str(FieldReturn+fmt.Sprintf("%d", i+1), fmt.Sprintf("[%s object]", typeName))
					}
				} else {
					typeName := retType.Name()
					event = event.Str(FieldReturn+fmt.Sprintf("%d", i+1), fmt.Sprintf("[%s object]", typeName))
				}
			default:
				// For other types, log the value
				event = event.Interface(FieldReturn+fmt.Sprintf("%d", i+1), ret)
			}
		}
	}

	event.Msg(MsgMethodCompleted)
}

// LogMethodReturnWithContext logs the exit of a method with context
// Deprecated: Use LogMethodExitWithContext instead
func LogMethodReturnWithContext(methodName string, startTime time.Time, logger Logger, ctx LogContext, returns ...interface{}) {
	LogMethodExitWithContext(methodName, startTime, logger, ctx, returns...)
}

// WithMethodLogging executes a function with method entry and exit logging
// It logs the method entry before executing the function and the method exit after execution
// The function can return multiple values, which will be logged and returned
func WithMethodLogging(methodName string, fn interface{}, params ...interface{}) []interface{} {
	// Log method entry
	methodName, startTime := LogMethodEntry(methodName, params...)

	// Prepare for function call
	fValue := reflect.ValueOf(fn)
	fType := fValue.Type()

	// Create input arguments
	in := make([]reflect.Value, len(params))
	for i, arg := range params {
		if arg == nil {
			in[i] = reflect.Zero(fType.In(i))
		} else {
			in[i] = reflect.ValueOf(arg)
		}
	}

	// Call the function
	out := fValue.Call(in)
	duration := time.Since(startTime)

	// Convert output to interface slice
	returns := make([]interface{}, len(out))
	for i, val := range out {
		returns[i] = val.Interface()
	}

	// Log method exit
	LogMethodExit(methodName, duration, returns...)

	return returns
}

// WithMethodLoggingAndContext executes a function with context-aware method entry and exit logging
// It logs the method entry with context before executing the function and the method exit with context after execution
// The function can return multiple values, which will be logged and returned
func WithMethodLoggingAndContext(methodName string, ctx LogContext, fn interface{}, params ...interface{}) []interface{} {
	// Log method entry with context
	methodName, startTime, logger, ctx := LogMethodEntryWithContext(methodName, ctx)

	// Prepare for function call
	fValue := reflect.ValueOf(fn)
	fType := fValue.Type()

	// Create input arguments
	in := make([]reflect.Value, len(params))
	for i, arg := range params {
		if arg == nil {
			in[i] = reflect.Zero(fType.In(i))
		} else {
			in[i] = reflect.ValueOf(arg)
		}
	}

	// Call the function
	out := fValue.Call(in)

	// Convert output to interface slice
	returns := make([]interface{}, len(out))
	for i, val := range out {
		returns[i] = val.Interface()
	}

	// Log method exit with context
	LogMethodExitWithContext(methodName, startTime, logger, ctx, returns...)

	return returns
}
