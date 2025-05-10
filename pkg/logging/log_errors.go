package logging

import (
	"fmt"
	"github.com/auriora/onemount/pkg/errors"
	"github.com/rs/zerolog/log"
)

// LogErrorWithFields logs an error with additional fields
// This is a convenience function for logging errors with additional context
func LogErrorWithFields(err error, msg string, fields map[string]interface{}) {
	if err == nil {
		return
	}

	event := log.Error().Err(err)

	for key, value := range fields {
		event = event.Interface(key, value)
	}

	event.Msg(msg)
}

// LogWarnWithFields logs a warning with additional fields
// This is useful for logging potential issues that don't prevent the application from working
func LogWarnWithFields(msg string, fields map[string]interface{}) {
	event := log.Warn()

	for key, value := range fields {
		event = event.Interface(key, value)
	}

	event.Msg(msg)
}

// LogWarnWithError logs a warning with an error
// This is useful for logging non-critical errors that don't prevent the application from working
func LogWarnWithError(err error, msg string, fields ...interface{}) {
	if err == nil {
		return
	}

	event := log.Warn().Err(err)

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

// LogErrorAndReturn logs an error and returns it
// This is a convenience function for the common pattern of logging an error and then returning it
func LogErrorAndReturn(err error, msg string, fields ...interface{}) error {
	if err == nil {
		return nil
	}

	errors.LogError(err, msg, fields...)
	return err
}

// LogErrorWithContextAndReturn logs an error with context and returns it
// This is a convenience function for the common pattern of logging an error with context and then returning it
func LogErrorWithContextAndReturn(err error, ctx LogContext, msg string, fields ...interface{}) error {
	if err == nil {
		return nil
	}

	LogErrorWithContext(err, ctx, msg, fields...)
	return err
}

// FormatErrorWithContext formats an error message with additional context
// This is useful for creating descriptive error messages
func FormatErrorWithContext(err error, msg string, fields ...interface{}) error {
	if err == nil {
		return nil
	}

	// Format the message with the fields
	formattedMsg := msg
	if len(fields) > 0 {
		formattedMsg += " ("
		for i := 0; i < len(fields); i += 2 {
			if i+1 < len(fields) {
				key, ok := fields[i].(string)
				if !ok {
					continue
				}
				formattedMsg += fmt.Sprintf("%s=%v", key, fields[i+1])
				if i+2 < len(fields) {
					formattedMsg += ", "
				}
			}
		}
		formattedMsg += ")"
	}

	// Return a new error with the formatted message and the original error
	return errors.Wrap(err, formattedMsg)
}
