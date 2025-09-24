// Package logging provides standardized logging utilities for the OneMount project.
// This file defines enhanced type caching mechanisms to reduce reflection overhead.
//
// Reflection is a powerful but expensive operation, especially in high-throughput
// applications. This file provides an enhanced type caching mechanism to reduce
// the overhead of reflection operations used in logging.
//
// This file is part of the consolidated logging package structure, which includes:
//   - logger.go: Core logger implementation and level management
//   - context.go: Context-aware logging functionality
//   - method.go: Method entry/exit logging (both with and without context)
//   - error.go: Error logging functionality
//   - performance.go: Performance optimization utilities
//   - type_helpers.go: Type-specific logging helpers
//   - type_cache.go (this file): Enhanced type caching mechanisms
package logging

import (
	"reflect"
	"sync"
	"time"
)

// TypeCache provides a centralized cache for type-related information
// to reduce the overhead of reflection operations.
//
// Reflection is a powerful but expensive operation in Go. Each reflection operation
// requires runtime type lookups and other overhead. By caching the results of common
// reflection operations, we can significantly reduce this overhead, especially for
// operations that are performed repeatedly on the same types.
//
// The TypeCache is organized into several categories of cached information:
// 1. Basic type information (name, kind, element type)
// 2. Enhanced type information (fields, methods, assignability, implementation)
// 3. Type check results (isPointer, isSlice, isMap, isStruct)
// 4. Specialized type check results (isByteSlice, isPointerToByteSlice, isStringSlice, isTime)
//
// All cache operations are thread-safe, protected by a read-write mutex that allows
// concurrent reads but exclusive writes. This ensures that the cache can be safely
// used from multiple goroutines.
type TypeCache struct {
	// Basic type information
	nameCache map[reflect.Type]string       // Cache for type names
	kindCache map[reflect.Type]reflect.Kind // Cache for type kinds
	elemCache map[reflect.Type]reflect.Type // Cache for element types (for pointers, slices, maps, channels)

	// Enhanced type information
	fieldCache      map[reflect.Type][]reflect.StructField // Cache for struct fields
	methodCache     map[reflect.Type][]reflect.Method      // Cache for type methods
	assignableCache map[typePair]bool                      // Cache for assignability checks
	implementsCache map[typePair]bool                      // Cache for implementation checks

	// Cache for type checks
	isPointerCache map[reflect.Type]bool // Cache for pointer type checks
	isSliceCache   map[reflect.Type]bool // Cache for slice type checks
	isMapCache     map[reflect.Type]bool // Cache for map type checks
	isStructCache  map[reflect.Type]bool // Cache for struct type checks

	// Cache for specialized type checks
	isByteSliceCache          map[reflect.Type]bool // Cache for byte slice type checks
	isPointerToByteSliceCache map[reflect.Type]bool // Cache for pointer to byte slice type checks
	isStringSliceCache        map[reflect.Type]bool // Cache for string slice type checks
	isTimeCache               map[reflect.Type]bool // Cache for time.Time type checks

	// Mutex to protect the cache
	mutex sync.RWMutex // Read-write mutex for thread safety
}

// typePair represents a pair of types for caching assignability and implementation checks
type typePair struct {
	t1, t2 reflect.Type
}

// Global type cache instance
var globalTypeCache = newTypeCache()

// newTypeCache creates a new TypeCache instance
func newTypeCache() *TypeCache {
	return &TypeCache{
		nameCache:                 make(map[reflect.Type]string),
		kindCache:                 make(map[reflect.Type]reflect.Kind),
		elemCache:                 make(map[reflect.Type]reflect.Type),
		fieldCache:                make(map[reflect.Type][]reflect.StructField),
		methodCache:               make(map[reflect.Type][]reflect.Method),
		assignableCache:           make(map[typePair]bool),
		implementsCache:           make(map[typePair]bool),
		isPointerCache:            make(map[reflect.Type]bool),
		isSliceCache:              make(map[reflect.Type]bool),
		isMapCache:                make(map[reflect.Type]bool),
		isStructCache:             make(map[reflect.Type]bool),
		isByteSliceCache:          make(map[reflect.Type]bool),
		isPointerToByteSliceCache: make(map[reflect.Type]bool),
		isStringSliceCache:        make(map[reflect.Type]bool),
		isTimeCache:               make(map[reflect.Type]bool),
	}
}

// GetTypeName returns the name of a type, using the cache for performance
func (tc *TypeCache) GetTypeName(t reflect.Type) string {
	tc.mutex.RLock()
	name, ok := tc.nameCache[t]
	tc.mutex.RUnlock()

	if ok {
		return name
	}

	// Compute type name
	name = t.String()

	tc.mutex.Lock()
	tc.nameCache[t] = name
	tc.mutex.Unlock()

	return name
}

// GetTypeKind returns the kind of a type, using the cache for performance
func (tc *TypeCache) GetTypeKind(t reflect.Type) reflect.Kind {
	tc.mutex.RLock()
	kind, ok := tc.kindCache[t]
	tc.mutex.RUnlock()

	if ok {
		return kind
	}

	// Compute type kind
	kind = t.Kind()

	tc.mutex.Lock()
	tc.kindCache[t] = kind
	tc.mutex.Unlock()

	return kind
}

// GetTypeElem returns the element type of a pointer, slice, map, or channel,
// using the cache for performance
func (tc *TypeCache) GetTypeElem(t reflect.Type) reflect.Type {
	tc.mutex.RLock()
	elem, ok := tc.elemCache[t]
	tc.mutex.RUnlock()

	if ok {
		return elem
	}

	// Compute element type
	elem = t.Elem()

	tc.mutex.Lock()
	tc.elemCache[t] = elem
	tc.mutex.Unlock()

	return elem
}

// GetStructFields returns the fields of a struct type, using the cache for performance
func (tc *TypeCache) GetStructFields(t reflect.Type) []reflect.StructField {
	tc.mutex.RLock()
	fields, ok := tc.fieldCache[t]
	tc.mutex.RUnlock()

	if ok {
		return fields
	}

	// Compute struct fields
	numFields := t.NumField()
	fields = make([]reflect.StructField, numFields)
	for i := 0; i < numFields; i++ {
		fields[i] = t.Field(i)
	}

	tc.mutex.Lock()
	tc.fieldCache[t] = fields
	tc.mutex.Unlock()

	return fields
}

// GetTypeMethods returns the methods of a type, using the cache for performance
func (tc *TypeCache) GetTypeMethods(t reflect.Type) []reflect.Method {
	tc.mutex.RLock()
	methods, ok := tc.methodCache[t]
	tc.mutex.RUnlock()

	if ok {
		return methods
	}

	// Compute type methods
	numMethods := t.NumMethod()
	methods = make([]reflect.Method, numMethods)
	for i := 0; i < numMethods; i++ {
		methods[i] = t.Method(i)
	}

	tc.mutex.Lock()
	tc.methodCache[t] = methods
	tc.mutex.Unlock()

	return methods
}

// IsAssignableTo checks if a type is assignable to another type,
// using the cache for performance
func (tc *TypeCache) IsAssignableTo(t1, t2 reflect.Type) bool {
	pair := typePair{t1, t2}

	tc.mutex.RLock()
	result, ok := tc.assignableCache[pair]
	tc.mutex.RUnlock()

	if ok {
		return result
	}

	// Compute assignability
	result = t1.AssignableTo(t2)

	tc.mutex.Lock()
	tc.assignableCache[pair] = result
	tc.mutex.Unlock()

	return result
}

// Implements checks if a type implements an interface,
// using the cache for performance
func (tc *TypeCache) Implements(t1, t2 reflect.Type) bool {
	pair := typePair{t1, t2}

	tc.mutex.RLock()
	result, ok := tc.implementsCache[pair]
	tc.mutex.RUnlock()

	if ok {
		return result
	}

	// Compute implementation
	result = t1.Implements(t2)

	tc.mutex.Lock()
	tc.implementsCache[pair] = result
	tc.mutex.Unlock()

	return result
}

// IsPointer checks if a type is a pointer, using the cache for performance
func (tc *TypeCache) IsPointer(t reflect.Type) bool {
	tc.mutex.RLock()
	result, ok := tc.isPointerCache[t]
	tc.mutex.RUnlock()

	if ok {
		return result
	}

	// Compute pointer check
	result = t.Kind() == reflect.Ptr

	tc.mutex.Lock()
	tc.isPointerCache[t] = result
	tc.mutex.Unlock()

	return result
}

// IsSlice checks if a type is a slice, using the cache for performance
func (tc *TypeCache) IsSlice(t reflect.Type) bool {
	tc.mutex.RLock()
	result, ok := tc.isSliceCache[t]
	tc.mutex.RUnlock()

	if ok {
		return result
	}

	// Compute slice check
	result = t.Kind() == reflect.Slice

	tc.mutex.Lock()
	tc.isSliceCache[t] = result
	tc.mutex.Unlock()

	return result
}

// IsMap checks if a type is a map, using the cache for performance
func (tc *TypeCache) IsMap(t reflect.Type) bool {
	tc.mutex.RLock()
	result, ok := tc.isMapCache[t]
	tc.mutex.RUnlock()

	if ok {
		return result
	}

	// Compute map check
	result = t.Kind() == reflect.Map

	tc.mutex.Lock()
	tc.isMapCache[t] = result
	tc.mutex.Unlock()

	return result
}

// IsStruct checks if a type is a struct, using the cache for performance
func (tc *TypeCache) IsStruct(t reflect.Type) bool {
	tc.mutex.RLock()
	result, ok := tc.isStructCache[t]
	tc.mutex.RUnlock()

	if ok {
		return result
	}

	// Compute struct check
	result = t.Kind() == reflect.Struct

	tc.mutex.Lock()
	tc.isStructCache[t] = result
	tc.mutex.Unlock()

	return result
}

// IsByteSlice checks if a type is a byte slice, using the cache for performance
func (tc *TypeCache) IsByteSlice(t reflect.Type) bool {
	tc.mutex.RLock()
	result, ok := tc.isByteSliceCache[t]
	tc.mutex.RUnlock()

	if ok {
		return result
	}

	// Compute byte slice check
	result = t.Kind() == reflect.Slice && t.Elem().Kind() == reflect.Uint8

	tc.mutex.Lock()
	tc.isByteSliceCache[t] = result
	tc.mutex.Unlock()

	return result
}

// IsPointerToByteSlice checks if a type is a pointer to a byte slice,
// using the cache for performance
func (tc *TypeCache) IsPointerToByteSlice(t reflect.Type) bool {
	tc.mutex.RLock()
	result, ok := tc.isPointerToByteSliceCache[t]
	tc.mutex.RUnlock()

	if ok {
		return result
	}

	// Compute pointer to byte slice check
	result = t.Kind() == reflect.Ptr && tc.IsByteSlice(t.Elem())

	tc.mutex.Lock()
	tc.isPointerToByteSliceCache[t] = result
	tc.mutex.Unlock()

	return result
}

// IsStringSlice checks if a type is a string slice, using the cache for performance
func (tc *TypeCache) IsStringSlice(t reflect.Type) bool {
	tc.mutex.RLock()
	result, ok := tc.isStringSliceCache[t]
	tc.mutex.RUnlock()

	if ok {
		return result
	}

	// Compute string slice check
	result = t.Kind() == reflect.Slice && t.Elem().Kind() == reflect.String

	tc.mutex.Lock()
	tc.isStringSliceCache[t] = result
	tc.mutex.Unlock()

	return result
}

// IsTime checks if a type is time.Time, using the cache for performance
func (tc *TypeCache) IsTime(t reflect.Type) bool {
	tc.mutex.RLock()
	result, ok := tc.isTimeCache[t]
	tc.mutex.RUnlock()

	if ok {
		return result
	}

	// Compute time check
	result = t.String() == "time.Time"

	tc.mutex.Lock()
	tc.isTimeCache[t] = result
	tc.mutex.Unlock()

	return result
}

// Global functions that use the global type cache

// GetTypeName returns the name of a type, using the global cache for performance
func GetTypeName(t reflect.Type) string {
	return globalTypeCache.GetTypeName(t)
}

// GetTypeKind returns the kind of a type, using the global cache for performance
func GetTypeKind(t reflect.Type) reflect.Kind {
	return globalTypeCache.GetTypeKind(t)
}

// GetTypeElem returns the element type of a pointer, slice, map, or channel,
// using the global cache for performance
func GetTypeElem(t reflect.Type) reflect.Type {
	return globalTypeCache.GetTypeElem(t)
}

// GetStructFields returns the fields of a struct type, using the global cache for performance
func GetStructFields(t reflect.Type) []reflect.StructField {
	return globalTypeCache.GetStructFields(t)
}

// GetTypeMethods returns the methods of a type, using the global cache for performance
func GetTypeMethods(t reflect.Type) []reflect.Method {
	return globalTypeCache.GetTypeMethods(t)
}

// IsAssignableTo checks if a type is assignable to another type,
// using the global cache for performance
func IsAssignableTo(t1, t2 reflect.Type) bool {
	return globalTypeCache.IsAssignableTo(t1, t2)
}

// Implements checks if a type implements an interface,
// using the global cache for performance
func Implements(t1, t2 reflect.Type) bool {
	return globalTypeCache.Implements(t1, t2)
}

// IsPointer checks if a type is a pointer, using the global cache for performance
func IsPointer(t reflect.Type) bool {
	return globalTypeCache.IsPointer(t)
}

// IsSlice checks if a type is a slice, using the global cache for performance
func IsSlice(t reflect.Type) bool {
	return globalTypeCache.IsSlice(t)
}

// IsMap checks if a type is a map, using the global cache for performance
func IsMap(t reflect.Type) bool {
	return globalTypeCache.IsMap(t)
}

// IsStruct checks if a type is a struct, using the global cache for performance
func IsStruct(t reflect.Type) bool {
	return globalTypeCache.IsStruct(t)
}

// IsByteSlice checks if a type is a byte slice, using the global cache for performance
func IsByteSlice(t reflect.Type) bool {
	return globalTypeCache.IsByteSlice(t)
}

// IsPointerToByteSlice checks if a type is a pointer to a byte slice,
// using the global cache for performance
func IsPointerToByteSlice(t reflect.Type) bool {
	return globalTypeCache.IsPointerToByteSlice(t)
}

// IsStringSlice checks if a type is a string slice, using the global cache for performance
func IsStringSlice(t reflect.Type) bool {
	return globalTypeCache.IsStringSlice(t)
}

// IsTime checks if a type is time.Time, using the global cache for performance
func IsTime(t reflect.Type) bool {
	return globalTypeCache.IsTime(t)
}

// Initialize common types in the cache
// This function is called automatically when the package is imported.
// It preloads the cache with information about common types that are
// likely to be used in logging operations. This provides two benefits:
//
//  1. It ensures that the first use of these types in logging doesn't
//     incur the full cost of reflection.
//
//  2. It populates the cache with types that are likely to be used
//     frequently, improving overall logging performance.
//
// The types preloaded include:
// - Basic types: string, int, float, bool, etc.
// - Common composite types: time.Time, []byte, []string, maps
//
// For each type, we cache:
// - The type name and kind
// - Element type for composite types
// - Struct fields for struct types
// - Results of common type checks
//
// This aggressive preloading strategy significantly reduces the
// reflection overhead for logging operations involving these types.
func init() {
	// Preload common types into the cache
	commonTypes := []interface{}{
		"", // string
		0,  // int
		int8(0),
		int16(0),
		int32(0),
		int64(0),
		uint(0),
		uint8(0),
		uint16(0),
		uint32(0),
		uint64(0),
		float32(0),
		float64(0),
		true, // bool
		time.Time{},
		[]byte{},
		[]string{},
		map[string]string{},
		map[string]interface{}{},
	}

	for _, v := range commonTypes {
		t := reflect.TypeOf(v)

		// Cache basic type information
		GetTypeName(t)
		GetTypeKind(t)

		// For composite types, cache element type and other relevant information
		switch t.Kind() {
		case reflect.Ptr, reflect.Slice, reflect.Map, reflect.Chan:
			GetTypeElem(t)
		case reflect.Struct:
			GetStructFields(t)
		}

		// Cache results of common type checks
		IsPointer(t)
		IsSlice(t)
		IsMap(t)
		IsStruct(t)
		IsByteSlice(t)
		IsPointerToByteSlice(t)
		IsStringSlice(t)
		IsTime(t)
	}
}
