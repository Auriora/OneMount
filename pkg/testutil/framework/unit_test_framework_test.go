package framework

import (
	"errors"
	"testing"
)

// TestUnitTestFixture tests the UnitTestFixture functionality
func TestUnitTestFixture(t *testing.T) {
	// Create a fixture
	fixture := NewUnitTestFixture("test-fixture")

	// Add setup and teardown functions
	setupCalled := false
	teardownCalled := false

	fixture.WithSetup(func(t *testing.T) (interface{}, error) {
		setupCalled = true
		return "fixture-data", nil
	}).WithTeardown(func(t *testing.T, fixture interface{}) error {
		teardownCalled = true
		if fixture != "fixture-data" {
			return errors.New("unexpected fixture data")
		}
		return nil
	}).WithData("key", "value")

	// Use the fixture
	fixture.Use(t, func(t *testing.T, fixture interface{}) {
		if !setupCalled {
			t.Error("Setup function was not called")
		}

		fixtureObj := fixture.(*UnitTestFixture)
		if fixtureObj.SetupData != "fixture-data" {
			t.Errorf("Expected fixture.SetupData to be 'fixture-data', got %v", fixtureObj.SetupData)
		}

		if val, ok := fixtureObj.Data["key"]; !ok || val != "value" {
			t.Errorf("Expected fixture.Data['key'] to be 'value', got %v", val)
		}
	})

	if !teardownCalled {
		t.Error("Teardown function was not called")
	}
}

// TestMock tests the Mock functionality
func TestMock(t *testing.T) {
	// Create a mock
	mock := NewMock(t, "test-mock")

	// Set up expectations
	mock.On("method1").WithArgs("arg1", "arg2").Return("result1")
	mock.On("method2").WithArgs(1, 2).Return(3).ReturnError(nil)
	mock.On("method3").WithArgs().ReturnError(errors.New("test error"))
	mock.On("method4").WithArgs("arg").SetTimes(2)

	// Call methods
	result1 := mock.Call("method1", "arg1", "arg2")
	if len(result1) != 1 || result1[0] != "result1" {
		t.Errorf("Expected result1 to be ['result1'], got %v", result1)
	}

	result2 := mock.Call("method2", 1, 2)
	if len(result2) != 1 || result2[0] != 3 {
		t.Errorf("Expected result2 to be [3], got %v", result2)
	}

	result3 := mock.Call("method3")
	if len(result3) != 1 || result3[0] == nil || result3[0].(error).Error() != "test error" {
		t.Errorf("Expected result3 to be [error('test error')], got %v", result3)
	}

	mock.Call("method4", "arg")
	mock.Call("method4", "arg")

	// Verify expectations
	mock.Verify()
}

// TestTableTests tests the table-driven test functionality
func TestTableTests(t *testing.T) {
	// Define test cases
	tests := []TableTest{
		{
			Name:     "Addition",
			Input:    []int{1, 2},
			Expected: 3,
		},
		{
			Name:     "Subtraction",
			Input:    []int{5, 3},
			Expected: 2,
		},
		{
			Name:          "Division by zero",
			Input:         []int{5, 0},
			ExpectedError: errors.New("division by zero"),
		},
	}

	// Run the tests
	RunTableTests(t, tests, func(t *testing.T, test TableTest) (interface{}, error) {
		input := test.Input.([]int)
		if len(input) != 2 {
			return nil, errors.New("input must be a slice of 2 integers")
		}

		a, b := input[0], input[1]

		switch test.Name {
		case "Addition":
			return a + b, nil
		case "Subtraction":
			return a - b, nil
		case "Division by zero":
			if b == 0 {
				return nil, errors.New("division by zero")
			}
			return a / b, nil
		default:
			return nil, errors.New("unknown operation")
		}
	})
}

// TestAssert tests the assertion utilities
func TestAssert(t *testing.T) {
	// Create a mock testing.T to capture errors
	mockT := &mockTestingT{}
	assert := NewAssert(mockT)

	// Test Equal
	assert.Equal(1, 1)
	if mockT.errorCalled {
		t.Error("Equal(1, 1) should not have failed")
	}

	mockT.reset()
	assert.Equal(1, 2)
	if !mockT.errorCalled {
		t.Error("Equal(1, 2) should have failed")
	}

	// Test NotEqual
	mockT.reset()
	assert.NotEqual(1, 2)
	if mockT.errorCalled {
		t.Error("NotEqual(1, 2) should not have failed")
	}

	mockT.reset()
	assert.NotEqual(1, 1)
	if !mockT.errorCalled {
		t.Error("NotEqual(1, 1) should have failed")
	}

	// Test Nil
	mockT.reset()
	assert.Nil(nil)
	if mockT.errorCalled {
		t.Error("Nil(nil) should not have failed")
	}

	mockT.reset()
	assert.Nil(1)
	if !mockT.errorCalled {
		t.Error("Nil(1) should have failed")
	}

	// Test NotNil
	mockT.reset()
	assert.NotNil(1)
	if mockT.errorCalled {
		t.Error("NotNil(1) should not have failed")
	}

	mockT.reset()
	assert.NotNil(nil)
	if !mockT.errorCalled {
		t.Error("NotNil(nil) should have failed")
	}

	// Test True
	mockT.reset()
	assert.True(true)
	if mockT.errorCalled {
		t.Error("True(true) should not have failed")
	}

	mockT.reset()
	assert.True(false)
	if !mockT.errorCalled {
		t.Error("True(false) should have failed")
	}

	// Test False
	mockT.reset()
	assert.False(false)
	if mockT.errorCalled {
		t.Error("False(false) should not have failed")
	}

	mockT.reset()
	assert.False(true)
	if !mockT.errorCalled {
		t.Error("False(true) should have failed")
	}

	// Test NoError
	mockT.reset()
	assert.NoError(nil)
	if mockT.errorCalled {
		t.Error("NoError(nil) should not have failed")
	}

	mockT.reset()
	assert.NoError(errors.New("test error"))
	if !mockT.errorCalled {
		t.Error("NoError(error) should have failed")
	}

	// Test Error
	mockT.reset()
	assert.Error(errors.New("test error"))
	if mockT.errorCalled {
		t.Error("Error(error) should not have failed")
	}

	mockT.reset()
	assert.Error(nil)
	if !mockT.errorCalled {
		t.Error("Error(nil) should have failed")
	}

	// Test ErrorContains
	mockT.reset()
	assert.ErrorContains(errors.New("test error message"), "error")
	if mockT.errorCalled {
		t.Error("ErrorContains(error, 'error') should not have failed")
	}

	mockT.reset()
	assert.ErrorContains(errors.New("test error message"), "not found")
	if !mockT.errorCalled {
		t.Error("ErrorContains(error, 'not found') should have failed")
	}

	// Test Len
	mockT.reset()
	assert.Len([]int{1, 2, 3}, 3)
	if mockT.errorCalled {
		t.Error("Len([]int{1, 2, 3}, 3) should not have failed")
	}

	mockT.reset()
	assert.Len([]int{1, 2, 3}, 2)
	if !mockT.errorCalled {
		t.Error("Len([]int{1, 2, 3}, 2) should have failed")
	}

	// Test Contains
	mockT.reset()
	assert.Contains([]int{1, 2, 3}, 2)
	if mockT.errorCalled {
		t.Error("Contains([]int{1, 2, 3}, 2) should not have failed")
	}

	mockT.reset()
	assert.Contains([]int{1, 2, 3}, 4)
	if !mockT.errorCalled {
		t.Error("Contains([]int{1, 2, 3}, 4) should have failed")
	}
}

// TestEdgeCaseGenerator tests the edge case generation utilities
func TestEdgeCaseGenerator(t *testing.T) {
	generator := NewEdgeCaseGenerator()

	// Test StringEdgeCases
	stringCases := generator.StringEdgeCases()
	if len(stringCases) == 0 {
		t.Error("StringEdgeCases should return a non-empty slice")
	}

	// Test IntEdgeCases
	intCases := generator.IntEdgeCases()
	if len(intCases) == 0 {
		t.Error("IntEdgeCases should return a non-empty slice")
	}

	// Test FloatEdgeCases
	floatCases := generator.FloatEdgeCases()
	if len(floatCases) == 0 {
		t.Error("FloatEdgeCases should return a non-empty slice")
	}

	// Test BoolEdgeCases
	boolCases := generator.BoolEdgeCases()
	if len(boolCases) != 2 {
		t.Errorf("BoolEdgeCases should return a slice of length 2, got %d", len(boolCases))
	}

	// Test TimeEdgeCases
	timeCases := generator.TimeEdgeCases()
	if len(timeCases) == 0 {
		t.Error("TimeEdgeCases should return a non-empty slice")
	}

	// Test SliceEdgeCases
	sliceCases := generator.SliceEdgeCases()
	if len(sliceCases) == 0 {
		t.Error("SliceEdgeCases should return a non-empty slice")
	}

	// Test MapEdgeCases
	mapCases := generator.MapEdgeCases()
	if len(mapCases) == 0 {
		t.Error("MapEdgeCases should return a non-empty slice")
	}
}

// TestErrorConditions tests the error condition utilities
func TestErrorConditions(t *testing.T) {
	// Create error conditions
	conditions := []*ErrorCondition{
		NewErrorCondition("No error").
			WithFunc(func() error {
				return nil
			}),
		NewErrorCondition("With error").
			WithFunc(func() error {
				return errors.New("test error")
			}).
			WithExpectedError(errors.New("test error")),
		NewErrorCondition("With recovery").
			WithFunc(func() error {
				return errors.New("test error")
			}).
			WithExpectedError(errors.New("test error")).
			WithRecovery(func() error {
				return nil
			}),
	}

	// Run the error conditions
	RunErrorConditions(t, conditions)
}

// mockTestingT is a mock implementation of testing.T for testing assertions
type mockTestingT struct {
	errorCalled bool
	errorMsg    string
}

func (m *mockTestingT) Errorf(format string, args ...interface{}) {
	m.errorCalled = true
	m.errorMsg = format
}

func (m *mockTestingT) reset() {
	m.errorCalled = false
	m.errorMsg = ""
}
