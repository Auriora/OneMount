// Package testutil provides testing utilities for the OneMount project.
package testutil

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"
)

// UnitTestFixture represents a test fixture for unit tests.
type UnitTestFixture struct {
	// Name of the fixture
	Name string
	// Setup function to prepare the fixture
	Setup func(t *testing.T) (interface{}, error)
	// Teardown function to clean up the fixture
	Teardown func(t *testing.T, fixture interface{}) error
	// Data associated with the fixture
	Data map[string]interface{}
	// SetupData holds the data returned by the Setup function
	SetupData interface{}
}

// NewUnitTestFixture creates a new UnitTestFixture with the given name.
func NewUnitTestFixture(name string) *UnitTestFixture {
	return &UnitTestFixture{
		Name: name,
		Data: make(map[string]interface{}),
	}
}

// WithSetup adds a setup function to the fixture.
func (f *UnitTestFixture) WithSetup(setup func(t *testing.T) (interface{}, error)) *UnitTestFixture {
	f.Setup = setup
	return f
}

// WithTeardown adds a teardown function to the fixture.
func (f *UnitTestFixture) WithTeardown(teardown func(t *testing.T, fixture interface{}) error) *UnitTestFixture {
	f.Teardown = teardown
	return f
}

// WithData adds data to the fixture.
func (f *UnitTestFixture) WithData(key string, value interface{}) *UnitTestFixture {
	f.Data[key] = value
	return f
}

// Use sets up the fixture, runs the test function, and tears down the fixture.
func (f *UnitTestFixture) Use(t *testing.T, testFunc func(t *testing.T, fixture interface{})) {
	var err error

	// Setup the fixture
	if f.Setup != nil {
		f.SetupData, err = f.Setup(t)
		if err != nil {
			t.Fatalf("Failed to set up fixture %s: %v", f.Name, err)
		}
	}

	// Register cleanup to ensure teardown is called even if the test fails
	t.Cleanup(func() {
		if f.Teardown != nil {
			if err := f.Teardown(t, f.SetupData); err != nil {
				t.Logf("Warning: Failed to tear down fixture %s: %v", f.Name, err)
			}
		}
	})

	// Run the test function
	testFunc(t, f)

	// Call teardown directly to ensure it's called before the test function returns
	if f.Teardown != nil {
		if err := f.Teardown(t, f.SetupData); err != nil {
			t.Errorf("Failed to tear down fixture %s: %v", f.Name, err)
		}
	}
}

// MockExpectation represents an expectation for a mock function call.
type MockExpectation struct {
	// Method name
	Method string
	// Arguments to expect
	Args []interface{}
	// Return values
	Returns []interface{}
	// Error to return
	Error error
	// Times the method should be called
	Times int
	// Actual call count
	CallCount int
}

// MockFunction represents a mock function.
type MockFunction struct {
	// Name of the function
	Name string
	// Function implementation
	Func interface{}
	// Expectations for the function
	Expectations []*MockExpectation
}

// Mock represents a mock object.
type Mock struct {
	// Name of the mock
	Name string
	// Functions in the mock
	Functions map[string]*MockFunction
	// t is the testing.T instance
	t *testing.T
}

// NewMock creates a new Mock with the given name.
func NewMock(t *testing.T, name string) *Mock {
	return &Mock{
		Name:      name,
		Functions: make(map[string]*MockFunction),
		t:         t,
	}
}

// On sets up an expectation for a method call.
func (m *Mock) On(method string) *MockExpectation {
	expectation := &MockExpectation{
		Method:    method,
		Args:      make([]interface{}, 0),
		Returns:   make([]interface{}, 0),
		Times:     1,
		CallCount: 0,
	}

	if _, exists := m.Functions[method]; !exists {
		m.Functions[method] = &MockFunction{
			Name:         method,
			Expectations: make([]*MockExpectation, 0),
		}
	}

	m.Functions[method].Expectations = append(m.Functions[method].Expectations, expectation)
	return expectation
}

// WithArgs sets the expected arguments for the expectation.
func (e *MockExpectation) WithArgs(args ...interface{}) *MockExpectation {
	e.Args = args
	return e
}

// Return sets the return values for the expectation.
func (e *MockExpectation) Return(values ...interface{}) *MockExpectation {
	e.Returns = values
	return e
}

// ReturnError sets the error to return for the expectation.
func (e *MockExpectation) ReturnError(err error) *MockExpectation {
	e.Error = err
	return e
}

// SetTimes sets the number of times the method should be called.
func (e *MockExpectation) SetTimes(times int) *MockExpectation {
	e.Times = times
	return e
}

// Call calls a mock function with the given arguments.
func (m *Mock) Call(method string, args ...interface{}) []interface{} {
	function, exists := m.Functions[method]
	if !exists {
		m.t.Fatalf("Mock %s: Method %s not found", m.Name, method)
	}

	// Find a matching expectation
	for _, expectation := range function.Expectations {
		if matchArgs(expectation.Args, args) {
			expectation.CallCount++
			if expectation.Error != nil {
				return append(expectation.Returns, expectation.Error)
			}
			return expectation.Returns
		}
	}

	m.t.Fatalf("Mock %s: No expectation found for method %s with args %v", m.Name, method, args)
	return nil
}

// Verify verifies that all expectations were met.
func (m *Mock) Verify() {
	for methodName, function := range m.Functions {
		for _, expectation := range function.Expectations {
			if expectation.CallCount != expectation.Times {
				m.t.Errorf("Mock %s: Method %s expected to be called %d times, but was called %d times",
					m.Name, methodName, expectation.Times, expectation.CallCount)
			}
		}
	}
}

// matchArgs checks if the actual arguments match the expected arguments.
func matchArgs(expected, actual []interface{}) bool {
	if len(expected) != len(actual) {
		return false
	}
	for i, exp := range expected {
		if exp != nil && !reflect.DeepEqual(exp, actual[i]) {
			return false
		}
	}
	return true
}

// TableTest represents a table-driven test.
type TableTest struct {
	// Name of the test
	Name string
	// Input data for the test
	Input interface{}
	// Expected output for the test
	Expected interface{}
	// Error expected from the test
	ExpectedError error
	// Setup function to run before the test
	Setup func(t *testing.T) error
	// Teardown function to run after the test
	Teardown func(t *testing.T) error
}

// RunTableTests runs a set of table-driven tests.
func RunTableTests(t *testing.T, tests []TableTest, testFunc func(t *testing.T, test TableTest) (interface{}, error)) {
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			// Run setup if provided
			if test.Setup != nil {
				if err := test.Setup(t); err != nil {
					t.Fatalf("Setup failed: %v", err)
				}
			}

			// Register cleanup to ensure teardown is called even if the test fails
			if test.Teardown != nil {
				t.Cleanup(func() {
					if err := test.Teardown(t); err != nil {
						t.Logf("Warning: Teardown failed: %v", err)
					}
				})
			}

			// Run the test function
			result, err := testFunc(t, test)

			// Check for expected error
			if test.ExpectedError != nil {
				if err == nil {
					t.Errorf("Expected error %v, but got nil", test.ExpectedError)
				} else if err.Error() != test.ExpectedError.Error() {
					t.Errorf("Expected error %v, but got %v", test.ExpectedError, err)
				}
			} else if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// Check for expected result
			if test.Expected != nil && !reflect.DeepEqual(result, test.Expected) {
				t.Errorf("Expected %v, but got %v", test.Expected, result)
			}
		})
	}
}

// TestingT is an interface for testing.T functionality used by Assert.
type TestingT interface {
	Errorf(format string, args ...interface{})
}

// Assert provides assertion utilities for unit tests.
type Assert struct {
	t TestingT
}

// NewAssert creates a new Assert with the given testing.T.
func NewAssert(t TestingT) *Assert {
	return &Assert{t: t}
}

// Equal asserts that two values are equal.
func (a *Assert) Equal(expected, actual interface{}, msgAndArgs ...interface{}) bool {
	if !reflect.DeepEqual(expected, actual) {
		a.fail("Equal", fmt.Sprintf("Expected %v, but got %v", expected, actual), msgAndArgs...)
		return false
	}
	return true
}

// NotEqual asserts that two values are not equal.
func (a *Assert) NotEqual(expected, actual interface{}, msgAndArgs ...interface{}) bool {
	if reflect.DeepEqual(expected, actual) {
		a.fail("NotEqual", fmt.Sprintf("Expected %v to not equal %v", expected, actual), msgAndArgs...)
		return false
	}
	return true
}

// Nil asserts that a value is nil.
func (a *Assert) Nil(value interface{}, msgAndArgs ...interface{}) bool {
	if value == nil {
		return true
	}

	v := reflect.ValueOf(value)
	kind := v.Kind()
	// IsNil() can only be called on chan, func, interface, map, pointer, or slice types
	if kind == reflect.Chan || kind == reflect.Func || kind == reflect.Interface ||
		kind == reflect.Map || kind == reflect.Ptr || kind == reflect.Slice {
		if !v.IsNil() {
			a.fail("Nil", fmt.Sprintf("Expected nil, but got %v", value), msgAndArgs...)
			return false
		}
		return true
	}

	// For non-nilable types, we can't call IsNil() so we just fail the assertion
	a.fail("Nil", fmt.Sprintf("Expected nil, but got non-nilable type %v with value %v", kind, value), msgAndArgs...)
	return false
}

// NotNil asserts that a value is not nil.
func (a *Assert) NotNil(value interface{}, msgAndArgs ...interface{}) bool {
	if value == nil {
		a.fail("NotNil", "Expected value to not be nil", msgAndArgs...)
		return false
	}

	v := reflect.ValueOf(value)
	kind := v.Kind()
	// IsNil() can only be called on chan, func, interface, map, pointer, or slice types
	if kind == reflect.Chan || kind == reflect.Func || kind == reflect.Interface ||
		kind == reflect.Map || kind == reflect.Ptr || kind == reflect.Slice {
		if v.IsNil() {
			a.fail("NotNil", "Expected value to not be nil", msgAndArgs...)
			return false
		}
	}

	return true
}

// True asserts that a value is true.
func (a *Assert) True(value bool, msgAndArgs ...interface{}) bool {
	if !value {
		a.fail("True", "Expected true, but got false", msgAndArgs...)
		return false
	}
	return true
}

// False asserts that a value is false.
func (a *Assert) False(value bool, msgAndArgs ...interface{}) bool {
	if value {
		a.fail("False", "Expected false, but got true", msgAndArgs...)
		return false
	}
	return true
}

// NoError asserts that an error is nil.
func (a *Assert) NoError(err error, msgAndArgs ...interface{}) bool {
	if err != nil {
		a.fail("NoError", fmt.Sprintf("Expected no error, but got %v", err), msgAndArgs...)
		return false
	}
	return true
}

// Error asserts that an error is not nil.
func (a *Assert) Error(err error, msgAndArgs ...interface{}) bool {
	if err == nil {
		a.fail("Error", "Expected error, but got nil", msgAndArgs...)
		return false
	}
	return true
}

// ErrorContains asserts that an error is not nil and contains the expected text.
func (a *Assert) ErrorContains(err error, contains string, msgAndArgs ...interface{}) bool {
	if err == nil {
		a.fail("ErrorContains", "Expected error, but got nil", msgAndArgs...)
		return false
	}
	if !strings.Contains(err.Error(), contains) {
		a.fail("ErrorContains", fmt.Sprintf("Expected error to contain '%s', but got '%s'", contains, err.Error()), msgAndArgs...)
		return false
	}
	return true
}

// Len asserts that a collection has the expected length.
func (a *Assert) Len(collection interface{}, length int, msgAndArgs ...interface{}) bool {
	val := reflect.ValueOf(collection)
	if val.Kind() != reflect.Array && val.Kind() != reflect.Slice && val.Kind() != reflect.Map && val.Kind() != reflect.String {
		a.fail("Len", fmt.Sprintf("Expected a collection, but got %v", collection), msgAndArgs...)
		return false
	}
	if val.Len() != length {
		a.fail("Len", fmt.Sprintf("Expected length %d, but got %d", length, val.Len()), msgAndArgs...)
		return false
	}
	return true
}

// Contains asserts that a collection contains an element.
func (a *Assert) Contains(collection, element interface{}, msgAndArgs ...interface{}) bool {
	val := reflect.ValueOf(collection)
	if val.Kind() != reflect.Array && val.Kind() != reflect.Slice && val.Kind() != reflect.Map && val.Kind() != reflect.String {
		a.fail("Contains", fmt.Sprintf("Expected a collection, but got %v", collection), msgAndArgs...)
		return false
	}

	switch val.Kind() {
	case reflect.String:
		str := val.String()
		elementStr, ok := element.(string)
		if !ok {
			a.fail("Contains", fmt.Sprintf("Expected element to be string, but got %v", element), msgAndArgs...)
			return false
		}
		if !strings.Contains(str, elementStr) {
			a.fail("Contains", fmt.Sprintf("Expected '%s' to contain '%s'", str, elementStr), msgAndArgs...)
			return false
		}
	case reflect.Array, reflect.Slice:
		found := false
		for i := 0; i < val.Len(); i++ {
			if reflect.DeepEqual(val.Index(i).Interface(), element) {
				found = true
				break
			}
		}
		if !found {
			a.fail("Contains", fmt.Sprintf("Expected %v to contain %v", collection, element), msgAndArgs...)
			return false
		}
	case reflect.Map:
		found := false
		for _, key := range val.MapKeys() {
			if reflect.DeepEqual(val.MapIndex(key).Interface(), element) {
				found = true
				break
			}
		}
		if !found {
			a.fail("Contains", fmt.Sprintf("Expected %v to contain %v", collection, element), msgAndArgs...)
			return false
		}
	}
	return true
}

// fail reports a test failure.
func (a *Assert) fail(assertion, message string, msgAndArgs ...interface{}) {
	_, file, line, _ := runtime.Caller(2)
	if len(msgAndArgs) > 0 {
		message = fmt.Sprintf("%s: %s", message, fmt.Sprint(msgAndArgs...))
	}
	a.t.Errorf("\nAssertion: %s\nError: %s\nLocation: %s:%d", assertion, message, file, line)
}

// EdgeCaseGenerator provides utilities for generating edge cases for testing.
type EdgeCaseGenerator struct{}

// NewEdgeCaseGenerator creates a new EdgeCaseGenerator.
func NewEdgeCaseGenerator() *EdgeCaseGenerator {
	return &EdgeCaseGenerator{}
}

// StringEdgeCases returns a set of edge cases for string testing.
func (g *EdgeCaseGenerator) StringEdgeCases() []string {
	return []string{
		"",                              // Empty string
		" ",                             // Space
		"  ",                            // Multiple spaces
		"\t",                            // Tab
		"\n",                            // Newline
		"\r\n",                          // Windows newline
		"a",                             // Single character
		"abcdefghijklmnopqrstuvwxyz",    // Long string
		"1234567890",                    // Numeric string
		"!@#$%^&*()",                    // Special characters
		"こんにちは",                         // Unicode characters
		"😀😁😂",                           // Emoji
		"<script>alert('XSS')</script>", // Potential XSS
		"'; DROP TABLE users; --",       // SQL injection
	}
}

// IntEdgeCases returns a set of edge cases for integer testing.
func (g *EdgeCaseGenerator) IntEdgeCases() []int {
	return []int{
		0,           // Zero
		1,           // One
		-1,          // Negative one
		2,           // Two
		-2,          // Negative two
		10,          // Ten
		-10,         // Negative ten
		100,         // Hundred
		-100,        // Negative hundred
		1000000,     // Million
		-1000000,    // Negative million
		2147483647,  // Max int32
		-2147483648, // Min int32
	}
}

// FloatEdgeCases returns a set of edge cases for float testing.
func (g *EdgeCaseGenerator) FloatEdgeCases() []float64 {
	return []float64{
		0.0,                     // Zero
		1.0,                     // One
		-1.0,                    // Negative one
		0.1,                     // Small positive
		-0.1,                    // Small negative
		0.00001,                 // Very small positive
		-0.00001,                // Very small negative
		1000000.0,               // Large positive
		-1000000.0,              // Large negative
		1.7976931348623157e+308, // Max float64
		4.9406564584124654e-324, // Min positive float64
	}
}

// BoolEdgeCases returns a set of edge cases for boolean testing.
func (g *EdgeCaseGenerator) BoolEdgeCases() []bool {
	return []bool{
		true,
		false,
	}
}

// TimeEdgeCases returns a set of edge cases for time testing.
func (g *EdgeCaseGenerator) TimeEdgeCases() []time.Time {
	return []time.Time{
		time.Time{},                                              // Zero time
		time.Now(),                                               // Current time
		time.Now().Add(24 * time.Hour),                           // Future time
		time.Now().Add(-24 * time.Hour),                          // Past time
		time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC),              // Unix epoch
		time.Date(2038, 1, 19, 3, 14, 7, 0, time.UTC),            // Year 2038 problem
		time.Date(9999, 12, 31, 23, 59, 59, 999999999, time.UTC), // Far future
	}
}

// SliceEdgeCases returns a set of edge cases for slice testing.
func (g *EdgeCaseGenerator) SliceEdgeCases() [][]interface{} {
	return [][]interface{}{
		{},                  // Empty slice
		{nil},               // Slice with nil
		{1},                 // Slice with one element
		{1, 2, 3},           // Slice with multiple elements
		{1, "string", true}, // Slice with mixed types
	}
}

// MapEdgeCases returns a set of edge cases for map testing.
func (g *EdgeCaseGenerator) MapEdgeCases() []map[string]interface{} {
	return []map[string]interface{}{
		{},                     // Empty map
		{"key": nil},           // Map with nil value
		{"key": "value"},       // Map with string value
		{"key1": 1, "key2": 2}, // Map with multiple entries
		{"key1": 1, "key2": "string", "key3": true}, // Map with mixed types
	}
}

// ErrorCondition represents an error condition for testing.
type ErrorCondition struct {
	// Name of the error condition
	Name string
	// Function to simulate the error condition
	Func func() error
	// Expected error
	ExpectedError error
	// Recovery function to clean up after the error condition
	Recovery func() error
}

// NewErrorCondition creates a new ErrorCondition with the given name.
func NewErrorCondition(name string) *ErrorCondition {
	return &ErrorCondition{
		Name: name,
	}
}

// WithFunc sets the function to simulate the error condition.
func (e *ErrorCondition) WithFunc(f func() error) *ErrorCondition {
	e.Func = f
	return e
}

// WithExpectedError sets the expected error for the error condition.
func (e *ErrorCondition) WithExpectedError(err error) *ErrorCondition {
	e.ExpectedError = err
	return e
}

// WithRecovery sets the recovery function for the error condition.
func (e *ErrorCondition) WithRecovery(f func() error) *ErrorCondition {
	e.Recovery = f
	return e
}

// Test tests the error condition.
func (e *ErrorCondition) Test(t *testing.T) {
	if e.Func == nil {
		t.Fatalf("Error condition %s has no function", e.Name)
	}

	// Run the error condition function
	err := e.Func()

	// Check for expected error
	if e.ExpectedError != nil {
		if err == nil {
			t.Errorf("Error condition %s: Expected error %v, but got nil", e.Name, e.ExpectedError)
		} else if err.Error() != e.ExpectedError.Error() {
			t.Errorf("Error condition %s: Expected error %v, but got %v", e.Name, e.ExpectedError, err)
		}
	} else if err != nil {
		t.Errorf("Error condition %s: Unexpected error: %v", e.Name, err)
	}

	// Run recovery function if provided
	if e.Recovery != nil {
		if err := e.Recovery(); err != nil {
			t.Logf("Warning: Error condition %s: Recovery failed: %v", e.Name, err)
		}
	}
}

// RunErrorConditions runs a set of error conditions.
func RunErrorConditions(t *testing.T, conditions []*ErrorCondition) {
	for _, condition := range conditions {
		t.Run(condition.Name, func(t *testing.T) {
			condition.Test(t)
		})
	}
}
