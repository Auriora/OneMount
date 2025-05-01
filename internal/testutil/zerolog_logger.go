package testutil

import (
	"github.com/rs/zerolog/log"
)

// ZerologLogger implements the Logger interface using zerolog.
type ZerologLogger struct {
	// Optional prefix for log messages
	prefix string
}

// NewZerologLogger creates a new ZerologLogger with the given prefix.
func NewZerologLogger(prefix string) *ZerologLogger {
	return &ZerologLogger{
		prefix: prefix,
	}
}

// Debug logs a debug message.
func (l *ZerologLogger) Debug(msg string, args ...interface{}) {
	event := log.Debug()
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
func (l *ZerologLogger) Info(msg string, args ...interface{}) {
	event := log.Info()
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
func (l *ZerologLogger) Warn(msg string, args ...interface{}) {
	event := log.Warn()
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
func (l *ZerologLogger) Error(msg string, args ...interface{}) {
	event := log.Error()
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
