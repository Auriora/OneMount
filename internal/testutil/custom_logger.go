package testutil

import (
	"github.com/auriora/onemount/pkg/logging"
)

// CustomLogger implements the Logger interface using our custom logging package.
type CustomLogger struct {
	// Optional prefix for log messages
	prefix string
}

// NewCustomLogger creates a new CustomLogger with the given prefix.
func NewCustomLogger(prefix string) *CustomLogger {
	return &CustomLogger{
		prefix: prefix,
	}
}

// Debug logs a debug message.
func (l *CustomLogger) Debug(msg string, args ...interface{}) {
	event := logging.Debug()
	if l.prefix != "" {
		event = event.Str("prefix", l.prefix)
	}

	// Add additional fields if provided
	for i := 0; i < len(args); i += 2 {
		if i+1 < len(args) {
			key, ok := args[i].(string)
			if !ok {
				continue
			}
			event = event.Interface(key, args[i+1])
		}
	}

	event.Msg(msg)
}

// Info logs an informational message.
func (l *CustomLogger) Info(msg string, args ...interface{}) {
	event := logging.Info()
	if l.prefix != "" {
		event = event.Str("prefix", l.prefix)
	}

	// Add additional fields if provided
	for i := 0; i < len(args); i += 2 {
		if i+1 < len(args) {
			key, ok := args[i].(string)
			if !ok {
				continue
			}
			event = event.Interface(key, args[i+1])
		}
	}

	event.Msg(msg)
}

// Warn logs a warning message.
func (l *CustomLogger) Warn(msg string, args ...interface{}) {
	event := logging.Warn()
	if l.prefix != "" {
		event = event.Str("prefix", l.prefix)
	}

	// Add additional fields if provided
	for i := 0; i < len(args); i += 2 {
		if i+1 < len(args) {
			key, ok := args[i].(string)
			if !ok {
				continue
			}
			event = event.Interface(key, args[i+1])
		}
	}

	event.Msg(msg)
}

// Error logs an error message.
func (l *CustomLogger) Error(msg string, args ...interface{}) {
	event := logging.Error()
	if l.prefix != "" {
		event = event.Str("prefix", l.prefix)
	}

	// Add additional fields if provided
	for i := 0; i < len(args); i += 2 {
		if i+1 < len(args) {
			key, ok := args[i].(string)
			if !ok {
				continue
			}
			event = event.Interface(key, args[i+1])
		}
	}

	event.Msg(msg)
}
