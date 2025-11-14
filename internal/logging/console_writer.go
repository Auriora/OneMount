// Package logging provides standardized logging utilities for the OneMount project.
// This file defines console writer functionality.
//
// While the default logging output is JSON for machine parsing, human-readable
// console output is often more useful during development and debugging. This file
// provides a console writer that formats log entries in a human-friendly way.
//
// The console writer can be customized with different output destinations and
// time formats to suit different development environments.
//
// This file is part of the consolidated logging package structure, which includes:
//   - logger.go: Core logger implementation and level management
//   - context.go: Context-aware logging functionality
//   - method.go: Method entry/exit logging (both with and without context)
//   - error.go: Error logging functionality
//   - performance.go: Performance optimization utilities
//   - constants.go: Constants used throughout the logging package
//   - console_writer.go (this file): Console writer functionality
package logging

import (
	"fmt"
	"io"
	"time"

	"github.com/rs/zerolog"
)

// NewConsoleWriterWithOptions creates a new console writer with custom settings.
func NewConsoleWriterWithOptions(output io.Writer, timeFormat string) io.Writer {
	writer := zerolog.ConsoleWriter{Out: output, TimeFormat: timeFormat}
	writer.FormatTimestamp = func(input interface{}) string {
		switch v := input.(type) {
		case time.Time:
			return v.Format(timeFormat)
		case string:
			return v
		default:
			return fmt.Sprint(v)
		}
	}
	return writer
}
