// Package logging provides standardized logging utilities for the OneMount project.
//
// The logging package is organized into several files, each with a specific purpose:
//   - logger.go: Core logger implementation and level management
//   - context.go: Context-aware logging functionality
//   - method.go: Method entry/exit logging (both with and without context)
//   - error.go: Error logging functionality
//   - performance.go: Performance optimization utilities
//   - constants.go: Constants used throughout the logging package
//   - console_writer.go: Console writer functionality
//   - structured_logging.go: Structured logging functions
//
// This file (logger.go) defines the core Logger and Event types that encapsulate zerolog functionality,
// as well as level-related functionality. It provides the foundation for all logging operations
// in the OneMount project.
package logging

import (
	"fmt"
	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
	"io"
	"os"
	"time"
)

// Logger is a wrapper around zerolog.Logger that provides the same functionality
// without exposing zerolog directly.
type Logger struct {
	zl zerolog.Logger
}

// Event is a wrapper around zerolog.Event that provides the same functionality
// without exposing zerolog directly.
type Event struct {
	ze *zerolog.Event
}

var (
	// DefaultLogger is the default logger used by the package-level functions.
	DefaultLogger = Logger{zl: zlog.Logger}
)

// SetGlobalLevel sets the global log level.
func SetGlobalLevel(level Level) {
	zerolog.SetGlobalLevel(zerolog.Level(level))
}

// Level represents a log level.
type Level int8

// Log levels.
const (
	DebugLevel Level = Level(zerolog.DebugLevel)
	InfoLevel  Level = Level(zerolog.InfoLevel)
	WarnLevel  Level = Level(zerolog.WarnLevel)
	ErrorLevel Level = Level(zerolog.ErrorLevel)
	FatalLevel Level = Level(zerolog.FatalLevel)
	PanicLevel Level = Level(zerolog.PanicLevel)
	NoLevel    Level = Level(zerolog.NoLevel)
	Disabled   Level = Level(zerolog.Disabled)
	TraceLevel Level = Level(zerolog.TraceLevel)
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

// New creates a new Logger with the given writer.
func New(w io.Writer) Logger {
	return Logger{zl: zerolog.New(w)}
}

// NewConsoleWriter creates a new console writer.
func NewConsoleWriter() io.Writer {
	return zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
}

// Output duplicates the current logger and sets w as its output.
func (l Logger) Output(w io.Writer) Logger {
	return Logger{zl: l.zl.Output(w)}
}

// Context is a wrapper around zerolog.Context that provides the same functionality
// without exposing zerolog directly.
type Context struct {
	zc zerolog.Context
}

// With creates a child logger with the field added to its context.
func (l Logger) With() Context {
	return Context{zc: l.zl.With()}
}

// Logger returns a Logger from the Context.
func (c Context) Logger() Logger {
	return Logger{zl: c.zc.Logger()}
}

// Str adds a string field to the context.
func (c Context) Str(key, val string) Context {
	return Context{zc: c.zc.Str(key, val)}
}

// Int adds an int field to the context.
func (c Context) Int(key string, val int) Context {
	return Context{zc: c.zc.Int(key, val)}
}

// Int64 adds an int64 field to the context.
func (c Context) Int64(key string, val int64) Context {
	return Context{zc: c.zc.Int64(key, val)}
}

// Float64 adds a float64 field to the context.
func (c Context) Float64(key string, val float64) Context {
	return Context{zc: c.zc.Float64(key, val)}
}

// Bool adds a bool field to the context.
func (c Context) Bool(key string, val bool) Context {
	return Context{zc: c.zc.Bool(key, val)}
}

// Err adds an error field to the context.
func (c Context) Err(err error) Context {
	return Context{zc: c.zc.Err(err)}
}

// Dur adds a duration field to the context.
func (c Context) Dur(key string, val time.Duration) Context {
	return Context{zc: c.zc.Dur(key, val)}
}

// Time adds a time field to the context.
func (c Context) Time(key string, val time.Time) Context {
	return Context{zc: c.zc.Time(key, val)}
}

// Interface adds an interface field to the context.
func (c Context) Interface(key string, val interface{}) Context {
	return Context{zc: c.zc.Interface(key, val)}
}

// Uint64 adds a uint64 field to the context.
func (c Context) Uint64(key string, val uint64) Context {
	return Context{zc: c.zc.Uint64(key, val)}
}

// Level creates a child logger with the minimum accepted level set to level.
func (l Logger) Level(level Level) Logger {
	return Logger{zl: l.zl.Level(zerolog.Level(level))}
}

// Debug starts a new message with debug level.
func (l Logger) Debug() Event {
	return Event{ze: l.zl.Debug()}
}

// Info starts a new message with info level.
func (l Logger) Info() Event {
	return Event{ze: l.zl.Info()}
}

// Warn starts a new message with warn level.
func (l Logger) Warn() Event {
	return Event{ze: l.zl.Warn()}
}

// Error starts a new message with error level.
func (l Logger) Error() Event {
	return Event{ze: l.zl.Error()}
}

// Fatal starts a new message with fatal level.
func (l Logger) Fatal() Event {
	return Event{ze: l.zl.Fatal()}
}

// Panic starts a new message with panic level.
func (l Logger) Panic() Event {
	return Event{ze: l.zl.Panic()}
}

// Trace starts a new message with trace level.
func (l Logger) Trace() Event {
	return Event{ze: l.zl.Trace()}
}

// Log starts a new message with no level.
func (l Logger) Log() Event {
	return Event{ze: l.zl.Log()}
}

// Str adds a string field to the event.
func (e Event) Str(key, val string) Event {
	return Event{ze: e.ze.Str(key, val)}
}

// Int adds an int field to the event.
func (e Event) Int(key string, val int) Event {
	return Event{ze: e.ze.Int(key, val)}
}

// Int64 adds an int64 field to the event.
func (e Event) Int64(key string, val int64) Event {
	return Event{ze: e.ze.Int64(key, val)}
}

// Float64 adds a float64 field to the event.
func (e Event) Float64(key string, val float64) Event {
	return Event{ze: e.ze.Float64(key, val)}
}

// Bool adds a bool field to the event.
func (e Event) Bool(key string, val bool) Event {
	return Event{ze: e.ze.Bool(key, val)}
}

// Err adds an error field to the event.
func (e Event) Err(err error) Event {
	return Event{ze: e.ze.Err(err)}
}

// Dur adds a duration field to the event.
func (e Event) Dur(key string, val time.Duration) Event {
	return Event{ze: e.ze.Dur(key, val)}
}

// Time adds a time field to the event.
func (e Event) Time(key string, val time.Time) Event {
	return Event{ze: e.ze.Time(key, val)}
}

// Interface adds an interface field to the event.
func (e Event) Interface(key string, val interface{}) Event {
	return Event{ze: e.ze.Interface(key, val)}
}

// Uint64 adds a uint64 field to the event.
func (e Event) Uint64(key string, val uint64) Event {
	return Event{ze: e.ze.Uint64(key, val)}
}

// Uint32 adds a uint32 field to the event.
func (e Event) Uint32(key string, val uint32) Event {
	return Event{ze: e.ze.Uint32(key, val)}
}

// Strs adds a string slice field to the event.
func (e Event) Strs(key string, vals []string) Event {
	return Event{ze: e.ze.Strs(key, vals)}
}

// Msg sends the event with the given message.
func (e Event) Msg(msg string) {
	e.ze.Msg(msg)
}

// Msgf sends the event with the given formatted message.
func (e Event) Msgf(format string, v ...interface{}) {
	e.ze.Msgf(format, v...)
}

// Send sends the event.
func (e Event) Send() {
	e.ze.Send()
}

// Enabled returns true if the event is enabled.
func (e Event) Enabled() bool {
	return e.ze.Enabled()
}

// Debug returns a debug logger.
func Debug() Event {
	return DefaultLogger.Debug()
}

// Info returns an info logger.
func Info() Event {
	return DefaultLogger.Info()
}

// Warn returns a warn logger.
func Warn() Event {
	return DefaultLogger.Warn()
}

// Error returns an error logger.
func Error() Event {
	return DefaultLogger.Error()
}

// Fatal returns a fatal logger.
func Fatal() Event {
	return DefaultLogger.Fatal()
}

// Panic returns a panic logger.
func Panic() Event {
	return DefaultLogger.Panic()
}

// Trace returns a trace logger.
func Trace() Event {
	return DefaultLogger.Trace()
}

// Log returns a logger with no level.
func Log() Event {
	return DefaultLogger.Log()
}
