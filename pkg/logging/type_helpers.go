// Package logging provides standardized logging utilities for the OneMount project.
// This file defines type-specific logging helpers to reduce reflection usage.
//
// Reflection is a powerful but expensive operation, especially in high-throughput
// applications. This file provides type-specific helpers for common types to avoid
// using reflection when possible. These helpers are used in method logging to
// efficiently log parameters and return values.
//
// This file is part of the consolidated logging package structure, which includes:
//   - logger.go: Core logger implementation and level management
//   - context.go: Context-aware logging functionality
//   - method.go: Method entry/exit logging (both with and without context)
//   - error.go: Error logging functionality
//   - performance.go: Performance optimization utilities
//   - type_helpers.go (this file): Type-specific logging helpers
package logging

import (
	"fmt"
	"reflect"
	"time"
)

// TypeLogger is an interface for type-specific logging helpers
// This interface allows for type-specific logging without using reflection
// at the point of logging. Instead, reflection is used only once to determine
// the type, and then the appropriate TypeLogger implementation is used for
// all subsequent logging operations.
type TypeLogger interface {
	// LogValue logs a value to an event using type-specific methods
	// This avoids the overhead of reflection for each logging operation
	LogValue(event Event, fieldName string, value interface{}) Event
}

// Type-specific loggers
var (
	// typeLoggers maps reflect.Kind to TypeLogger
	typeLoggers = map[reflect.Kind]TypeLogger{
		reflect.Bool:    boolLogger{},
		reflect.Int:     intLogger{},
		reflect.Int8:    int8Logger{},
		reflect.Int16:   int16Logger{},
		reflect.Int32:   int32Logger{},
		reflect.Int64:   int64Logger{},
		reflect.Uint:    uintLogger{},
		reflect.Uint8:   uint8Logger{},
		reflect.Uint16:  uint16Logger{},
		reflect.Uint32:  uint32Logger{},
		reflect.Uint64:  uint64Logger{},
		reflect.Float32: float32Logger{},
		reflect.Float64: float64Logger{},
		reflect.String:  stringLogger{},
	}

	// specialTypeLoggers maps reflect.Type to TypeLogger
	specialTypeLoggers = map[reflect.Type]TypeLogger{
		reflect.TypeOf(time.Time{}): timeLogger{},
		reflect.TypeOf([]byte{}):    byteSliceLogger{},
		reflect.TypeOf([]string{}):  stringSliceLogger{},
	}
)

// getTypeLogger returns a TypeLogger for the given type
// This function uses a multi-level lookup strategy to find the most specific
// TypeLogger for a given type:
// 1. First, it checks if there's a special logger for the exact type
// 2. Then, it checks if the type is a pointer to a special type
// 3. Then, it checks if there's a logger for the type's kind
// 4. Finally, it falls back to a generic interface logger
//
// This approach ensures that we use the most efficient logging method
// for each type, while still supporting all possible types.
func getTypeLogger(t reflect.Type) TypeLogger {
	// Check for special types first (most specific)
	if logger, ok := specialTypeLoggers[t]; ok {
		return logger
	}

	// Check for pointer to special types (second most specific)
	if t.Kind() == reflect.Ptr {
		elemType := t.Elem()
		if logger, ok := specialTypeLoggers[elemType]; ok {
			return ptrLogger{elemLogger: logger}
		}
	}

	// Check for standard kinds (third most specific)
	if logger, ok := typeLoggers[t.Kind()]; ok {
		return logger
	}

	// Default to interface logger (least specific, but works for any type)
	return interfaceLogger{}
}

// logValueWithTypeLogger logs a value using the appropriate TypeLogger
// This function is the core of the type-specific logging system. It:
// 1. Handles nil values directly
// 2. Uses reflection once to determine the type of the value
// 3. Gets the appropriate TypeLogger for that type
// 4. Uses the TypeLogger to log the value without further reflection
//
// This approach significantly reduces the amount of reflection used in logging,
// especially for repeated logging of values of the same type.
func logValueWithTypeLogger(event Event, fieldName string, value interface{}) Event {
	if value == nil {
		return event.Interface(fieldName, nil)
	}

	t := reflect.TypeOf(value)
	logger := getTypeLogger(t)
	return logger.LogValue(event, fieldName, value)
}

// Type-specific logger implementations

type boolLogger struct{}

func (l boolLogger) LogValue(event Event, fieldName string, value interface{}) Event {
	return event.Bool(fieldName, value.(bool))
}

type intLogger struct{}

func (l intLogger) LogValue(event Event, fieldName string, value interface{}) Event {
	return event.Int(fieldName, value.(int))
}

type int8Logger struct{}

func (l int8Logger) LogValue(event Event, fieldName string, value interface{}) Event {
	return event.Int(fieldName, int(value.(int8)))
}

type int16Logger struct{}

func (l int16Logger) LogValue(event Event, fieldName string, value interface{}) Event {
	return event.Int(fieldName, int(value.(int16)))
}

type int32Logger struct{}

func (l int32Logger) LogValue(event Event, fieldName string, value interface{}) Event {
	return event.Int(fieldName, int(value.(int32)))
}

type int64Logger struct{}

func (l int64Logger) LogValue(event Event, fieldName string, value interface{}) Event {
	return event.Int64(fieldName, value.(int64))
}

type uintLogger struct{}

func (l uintLogger) LogValue(event Event, fieldName string, value interface{}) Event {
	return event.Uint64(fieldName, uint64(value.(uint)))
}

type uint8Logger struct{}

func (l uint8Logger) LogValue(event Event, fieldName string, value interface{}) Event {
	return event.Uint64(fieldName, uint64(value.(uint8)))
}

type uint16Logger struct{}

func (l uint16Logger) LogValue(event Event, fieldName string, value interface{}) Event {
	return event.Uint64(fieldName, uint64(value.(uint16)))
}

type uint32Logger struct{}

func (l uint32Logger) LogValue(event Event, fieldName string, value interface{}) Event {
	return event.Uint32(fieldName, value.(uint32))
}

type uint64Logger struct{}

func (l uint64Logger) LogValue(event Event, fieldName string, value interface{}) Event {
	return event.Uint64(fieldName, value.(uint64))
}

type float32Logger struct{}

func (l float32Logger) LogValue(event Event, fieldName string, value interface{}) Event {
	return event.Float64(fieldName, float64(value.(float32)))
}

type float64Logger struct{}

func (l float64Logger) LogValue(event Event, fieldName string, value interface{}) Event {
	return event.Float64(fieldName, value.(float64))
}

type stringLogger struct{}

func (l stringLogger) LogValue(event Event, fieldName string, value interface{}) Event {
	return event.Str(fieldName, value.(string))
}

type timeLogger struct{}

func (l timeLogger) LogValue(event Event, fieldName string, value interface{}) Event {
	return event.Time(fieldName, value.(time.Time))
}

type byteSliceLogger struct{}

func (l byteSliceLogger) LogValue(event Event, fieldName string, value interface{}) Event {
	byteSlice := value.([]byte)
	return event.Int(fieldName+"_size", len(byteSlice))
}

type stringSliceLogger struct{}

func (l stringSliceLogger) LogValue(event Event, fieldName string, value interface{}) Event {
	return event.Strs(fieldName, value.([]string))
}

type ptrLogger struct {
	elemLogger TypeLogger
}

func (l ptrLogger) LogValue(event Event, fieldName string, value interface{}) Event {
	v := reflect.ValueOf(value)
	if v.IsNil() {
		return event.Interface(fieldName, nil)
	}
	return l.elemLogger.LogValue(event, fieldName, v.Elem().Interface())
}

type interfaceLogger struct{}

func (l interfaceLogger) LogValue(event Event, fieldName string, value interface{}) Event {
	return event.Interface(fieldName, value)
}

// LogParam logs a method parameter with the appropriate type-specific logger
// This function is used in method logging to log parameters efficiently.
// It:
// 1. Constructs the parameter field name using the standard format
// 2. Uses logValueWithTypeLogger to log the parameter value with the appropriate TypeLogger
//
// Using this function instead of direct reflection significantly reduces the overhead
// of logging method parameters, especially for methods that are called frequently.
func LogParam(event Event, index int, value interface{}) Event {
	fieldName := FieldParam + fmt.Sprintf("%d", index+1)
	return logValueWithTypeLogger(event, fieldName, value)
}

// LogReturn logs a method return value with the appropriate type-specific logger
// This function is used in method logging to log return values efficiently.
// It:
// 1. Constructs the return value field name using the standard format
// 2. Uses logValueWithTypeLogger to log the return value with the appropriate TypeLogger
//
// Using this function instead of direct reflection significantly reduces the overhead
// of logging method return values, especially for methods that are called frequently.
func LogReturn(event Event, index int, value interface{}) Event {
	fieldName := FieldReturn + fmt.Sprintf("%d", index+1)
	return logValueWithTypeLogger(event, fieldName, value)
}
