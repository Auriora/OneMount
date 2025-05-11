// Package logging provides standardized logging utilities for the OneMount project.
// This file defines method logging functionality, both with and without context.
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
func LogMethodEntry(methodName string, params ...interface{}) {
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
				case paramType.Kind() == reflect.Ptr && paramType.Elem().Kind() == reflect.Slice && paramType.Elem().Elem().Kind() == reflect.Uint8:
					// For []byte pointers, just log the length
					byteSlice := reflect.ValueOf(param).Elem().Interface().([]byte)
					event = event.Int(FieldParam+fmt.Sprintf("%d_size", i+1), len(byteSlice))
				case strings.Contains(paramType.String(), "Auth"):
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
}

// LogMethodExit logs the exit of a method with its return values
func LogMethodExit(methodName string, duration time.Duration, returns ...interface{}) {
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

				// Handle different types of return values
				switch {
				case retType.Kind() == reflect.Ptr && retType.Elem().Kind() == reflect.Slice && retType.Elem().Elem().Kind() == reflect.Uint8:
					// For []byte pointers, just log the length
					byteSlice := reflect.ValueOf(ret).Elem().Interface().([]byte)
					event = event.Int(FieldReturn+fmt.Sprintf("%d_size", i+1), len(byteSlice))
				case strings.Contains(retType.String(), "Auth"):
					// Don't log auth objects which might contain sensitive information
					event = event.Str(FieldReturn+fmt.Sprintf("%d", i+1), "[Auth object]")
				case retType.Kind() == reflect.Struct || (retType.Kind() == reflect.Ptr && retType.Elem().Kind() == reflect.Struct):
					// For structs, log a simplified representation
					if retType.Kind() == reflect.Ptr {
						if reflect.ValueOf(ret).IsNil() {
							event = event.Str(FieldReturn+fmt.Sprintf("%d", i+1), "nil")
						} else {
							typeName := retType.Elem().Name()
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
	LogMethodEntry(methodName, args...)

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
	startTime := time.Now()
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

// LogMethodCallWithContext logs the entry of a method with context
func LogMethodCallWithContext(methodName string, ctx LogContext) (string, time.Time, Logger, LogContext) {
	// Create a logger with the context
	logger := WithLogContext(ctx)

	// Get the current goroutine ID
	goroutineID := util.GetCurrentGoroutineID()

	// Log method entry
	logger.Debug().
		Str(FieldMethod, methodName).
		Str(FieldPhase, PhaseEntry).
		Str(FieldGoroutine, goroutineID).
		Msg(MsgMethodCalled)

	return methodName, time.Now(), logger, ctx
}

// LogMethodReturnWithContext logs the exit of a method with context
func LogMethodReturnWithContext(methodName string, startTime time.Time, logger Logger, ctx LogContext, returns ...interface{}) {
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
			// TODO move this code into 'internal/fs' - package fs
			// Special handling for Inode objects to prevent race conditions during JSON serialization
			//if inodeInfo, ok := ret.(fs.InodeInfo); ok {
			//	// Only log the ID and name instead of the entire object
			//	event = event.Str(FieldReturn+fmt.Sprintf("%d", i+1)+".id", inodeInfo.ID()).
			//		Str(FieldReturn+fmt.Sprintf("%d", i+1)+".name", inodeInfo.Name()).
			//		Bool(FieldReturn+fmt.Sprintf("%d", i+1)+".isDir", inodeInfo.IsDir())
			//} else {
			event = event.Interface(FieldReturn+fmt.Sprintf("%d", i+1), ret)
			//}
		}
	}

	event.Msg(MsgMethodCompleted)
}
