// Package logging provides standardized logging utilities for the OneMount project.
// This file defines level-related functionality.
package logging

import (
	"fmt"

	"github.com/rs/zerolog"
)

// ParseLevel parses a level string into a Level.
// It returns an error if the level string is invalid.
func ParseLevel(levelStr string) (Level, error) {
	level, err := zerolog.ParseLevel(levelStr)
	if err != nil {
		return Level(0), fmt.Errorf("invalid log level %q: %w", levelStr, err)
	}
	return Level(level), nil
}

// String returns the string representation of the log level.
func (l Level) String() string {
	return zerolog.Level(l).String()
}

// MarshalText implements encoding.TextMarshaler interface.
func (l Level) MarshalText() ([]byte, error) {
	return []byte(l.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler interface.
func (l *Level) UnmarshalText(text []byte) error {
	level, err := ParseLevel(string(text))
	if err != nil {
		return err
	}
	*l = level
	return nil
}

// IsDebugEnabled returns true if debug logging is enabled.
func IsDebugEnabled() bool {
	return Debug().Enabled()
}

// IsTraceEnabled returns true if trace logging is enabled.
func IsTraceEnabled() bool {
	return Trace().Enabled()
}
