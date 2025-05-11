// Package logging provides standardized logging utilities for the OneMount project.
// This file defines console writer functionality.
package logging

import (
	"io"

	"github.com/rs/zerolog"
)

// NewConsoleWriterWithOptions creates a new console writer with custom settings.
func NewConsoleWriterWithOptions(output io.Writer, timeFormat string) io.Writer {
	return zerolog.ConsoleWriter{Out: output, TimeFormat: timeFormat}
}
