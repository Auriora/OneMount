// Package logging provides standardized logging utilities for the OneMount project.
// This file defines context-aware method logging functions.
package logging

import (
	"fmt"
	"github.com/auriora/onemount/pkg/util"
	"time"
)

// WithLogContext creates a new Logger with the given context
func WithLogContext(ctx LogContext) Logger {
	logger := DefaultLogger.With()

	if ctx.RequestID != "" {
		logger = logger.Str("request_id", ctx.RequestID)
	}

	if ctx.UserID != "" {
		logger = logger.Str(FieldUser, ctx.UserID)
	}

	if ctx.Operation != "" {
		logger = logger.Str(FieldOperation, ctx.Operation)
	}

	return logger.Logger()
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

// LogErrorWithContext logs an error with the given context
func LogErrorWithContext(err error, ctx LogContext, msg string, fields ...interface{}) {
	if err == nil {
		return
	}

	logger := WithLogContext(ctx)
	event := logger.Error().Err(err)

	// Add additional fields in pairs (key, value)
	for i := 0; i < len(fields); i += 2 {
		if i+1 < len(fields) {
			key, ok := fields[i].(string)
			if !ok {
				continue
			}
			event = event.Interface(key, fields[i+1])
		}
	}

	event.Msg(msg)
}
