package graph

import (
	"errors"
	"testing"

	"github.com/auriora/onemount/internal/testutil/framework"
)

// TestUT_GR_23_01_OfflineState_SetAndGet_StateCorrectlyManaged tests setting and getting the operational offline state.
//
//	Test Case ID    UT-GR-23-01
//	Title           Operational Offline State Management
//	Description     Tests setting and getting the operational offline state
//	Preconditions   None
//	Steps           1. Reset the operational offline state
//	                2. Check the default state
//	                3. Set the state to true and check it
//	                4. Set the state back to false and check it
//	Expected Result The operational offline state is correctly set and retrieved
//	Notes: This test verifies that the operational offline state can be set and retrieved correctly.
func TestUT_GR_23_01_OfflineState_SetAndGet_StateCorrectlyManaged(t *testing.T) {
	// Create a test fixture
	fixture := framework.NewUnitTestFixture("OperationalOfflineStateFixture")

	// Use the fixture to run the test
	fixture.Use(t, func(t *testing.T, fixture interface{}) {
		assert := framework.NewAssert(t)

		// Step 1: Reset the operational offline state
		SetOperationalOffline(false)

		// Step 2: Check the default state (should be false after reset)
		assert.False(GetOperationalOffline(), "Default operational offline state should be false")

		// Step 3: Set the state to true and check it
		SetOperationalOffline(true)
		assert.True(GetOperationalOffline(), "Operational offline state should be true after setting to true")

		// Step 4: Set the state back to false and check it
		SetOperationalOffline(false)
		assert.False(GetOperationalOffline(), "Operational offline state should be false after setting back to false")

		// Step 5: Toggle multiple times to verify consistency
		for i := 0; i < 5; i++ {
			SetOperationalOffline(true)
			assert.True(GetOperationalOffline(), "State should be true after setting true (iteration %d)", i)
			SetOperationalOffline(false)
			assert.False(GetOperationalOffline(), "State should be false after setting false (iteration %d)", i)
		}
	})
}

// TestUT_GR_24_01_IsOffline_OperationalStateSet_ReturnsTrue tests the IsOffline function when operational offline state is set.
//
//	Test Case ID    UT-GR-24-01
//	Title           IsOffline with Operational State
//	Description     Tests the IsOffline function when operational offline state is set
//	Preconditions   None
//	Steps           1. Set the operational offline state to true
//	                2. Call IsOffline with different errors
//	                3. Reset the operational offline state
//	                4. Call IsOffline with different errors again
//	Expected Result IsOffline returns true when operational offline is set, regardless of the error
//	Notes: This test verifies that the IsOffline function correctly handles the operational offline state.
func TestUT_GR_24_01_IsOffline_OperationalStateSet_ReturnsTrue(t *testing.T) {
	// Create a test fixture
	fixture := framework.NewUnitTestFixture("IsOfflineWithOperationalStateFixture")

	// Use the fixture to run the test
	fixture.Use(t, func(t *testing.T, fixture interface{}) {
		assert := framework.NewAssert(t)

		// Step 1: Set the operational offline state to true
		SetOperationalOffline(true)
		defer SetOperationalOffline(false) // Cleanup

		// Step 2: IsOffline should return true regardless of the error when operational offline is set
		assert.True(IsOffline(nil), "IsOffline should return true with nil error when operational offline is set")
		assert.True(IsOffline(errors.New("some random error")), "IsOffline should return true with any error when operational offline is set")
		assert.True(IsOffline(errors.New("HTTP 200 - OK")), "IsOffline should return true even with HTTP success pattern when operational offline is set")
		assert.True(IsOffline(errors.New("401 unauthorized")), "IsOffline should return true even with auth error when operational offline is set")

		// Step 3: Reset the operational offline state
		SetOperationalOffline(false)

		// Step 4: Now IsOffline should depend on the error content
		assert.False(IsOffline(nil), "IsOffline should return false with nil error when operational offline is not set")
		assert.False(IsOffline(errors.New("some random error")), "IsOffline should return false for unknown errors (non-conservative default)")
	})
}

// TestUT_GR_25_01_IsOffline_VariousErrors_IdentifiesNetworkErrors tests the IsOffline function with various error types.
//
//	Test Case ID    UT-GR-25-01
//	Title           IsOffline Error Identification
//	Description     Tests the IsOffline function with various error types
//	Preconditions   None
//	Steps           1. Reset the operational offline state
//	                2. Call IsOffline with different types of errors
//	                3. Check if the results match expectations
//	Expected Result IsOffline correctly identifies network-related errors
//	Notes: This test verifies that the IsOffline function correctly identifies network-related errors.
func TestUT_GR_25_01_IsOffline_VariousErrors_IdentifiesNetworkErrors(t *testing.T) {
	// Create a test fixture
	fixture := framework.NewUnitTestFixture("IsOfflineFixture")

	// Use the fixture to run the test
	fixture.Use(t, func(t *testing.T, fixture interface{}) {
		assert := framework.NewAssert(t)

		// Step 1: Reset the operational offline state
		SetOperationalOffline(false)

		// Step 2: Test nil error — should be online
		assert.False(IsOffline(nil), "nil error should not indicate offline")

		// Step 3: Test network errors that indicate offline state
		offlineErrors := []string{
			"dial tcp: no such host",
			"network is unreachable",
			"connection refused",
			"connection timed out",
			"dial tcp 10.0.0.1:443: connect: connection refused",
			"context deadline exceeded",
			"no route to host",
			"network is down",
			"temporary failure in name resolution",
			"operation timed out",
		}
		for _, errMsg := range offlineErrors {
			assert.True(IsOffline(errors.New(errMsg)), "Error '%s' should indicate offline", errMsg)
		}

		// Step 4: Test HTTP response errors — should be online (server responded)
		httpErrors := []string{
			"HTTP 500 - Internal Server Error",
			"HTTP 404 - Not Found",
			"HTTP 503 - Service Unavailable",
		}
		for _, errMsg := range httpErrors {
			assert.False(IsOffline(errors.New(errMsg)), "HTTP error '%s' should not indicate offline (server responded)", errMsg)
		}

		// Step 5: Test auth errors — should be online (server responded with auth error)
		authErrors := []string{
			"401 unauthorized",
			"403 forbidden",
			"invalid token",
			"permission denied",
			"access denied",
			"authentication failed",
		}
		for _, errMsg := range authErrors {
			assert.False(IsOffline(errors.New(errMsg)), "Auth error '%s' should not indicate offline", errMsg)
		}

		// Step 6: Test unknown errors — should default to online (non-conservative)
		assert.False(IsOffline(errors.New("some unknown error")), "Unknown errors should default to online")
		assert.False(IsOffline(errors.New("file not found")), "Non-network errors should not indicate offline")
	})
}
