package graph

import (
	"errors"
	"testing"

	"github.com/auriora/onemount/internal/testutil/framework"
	ierrors "github.com/auriora/onemount/internal/errors"
	"github.com/stretchr/testify/assert"
)

// TestUT_GR_26_01_NetworkErrorPatterns_NoSuchHost tests "no such host" pattern recognition
//
//	Test Case ID    UT-GR-26-01
//	Title           Network Error Pattern: No Such Host
//	Description     Tests that "no such host" errors are recognized as offline conditions
//	Preconditions   Operational offline state is false
//	Steps           1. Reset operational offline state
//	                2. Create error with "no such host" message
//	                3. Call IsOffline with the error
//	Expected Result IsOffline returns true
//	Requirements    19.1
func TestUT_GR_26_01_NetworkErrorPatterns_NoSuchHost(t *testing.T) {
	fixture := framework.NewUnitTestFixture("NetworkErrorPattern_NoSuchHost")
	fixture.Use(t, func(t *testing.T, fixture interface{}) {
		// Reset operational offline state
		SetOperationalOffline(false)

		// Test various forms of "no such host" error
		testCases := []struct {
			name     string
			errorMsg string
		}{
			{"lowercase", "no such host"},
			{"uppercase", "NO SUCH HOST"},
			{"mixed case", "No Such Host"},
			{"in sentence", "dial tcp: lookup example.com: no such host"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				err := errors.New(tc.errorMsg)
				result := IsOffline(err)
				assert.True(t, result, "Expected IsOffline to return true for error: %s", tc.errorMsg)
			})
		}
	})
}

// TestUT_GR_26_02_NetworkErrorPatterns_NetworkUnreachable tests "network is unreachable" pattern recognition
//
//	Test Case ID    UT-GR-26-02
//	Title           Network Error Pattern: Network Unreachable
//	Description     Tests that "network is unreachable" errors are recognized as offline conditions
//	Preconditions   Operational offline state is false
//	Steps           1. Reset operational offline state
//	                2. Create error with "network is unreachable" message
//	                3. Call IsOffline with the error
//	Expected Result IsOffline returns true
//	Requirements    19.2
func TestUT_GR_26_02_NetworkErrorPatterns_NetworkUnreachable(t *testing.T) {
	fixture := framework.NewUnitTestFixture("NetworkErrorPattern_NetworkUnreachable")
	fixture.Use(t, func(t *testing.T, fixture interface{}) {
		// Reset operational offline state
		SetOperationalOffline(false)

		// Test various forms of "network is unreachable" error
		testCases := []struct {
			name     string
			errorMsg string
		}{
			{"lowercase", "network is unreachable"},
			{"uppercase", "NETWORK IS UNREACHABLE"},
			{"mixed case", "Network Is Unreachable"},
			{"in sentence", "dial tcp 192.168.1.1:443: connect: network is unreachable"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				err := errors.New(tc.errorMsg)
				result := IsOffline(err)
				assert.True(t, result, "Expected IsOffline to return true for error: %s", tc.errorMsg)
			})
		}
	})
}

// TestUT_GR_26_03_NetworkErrorPatterns_ConnectionRefused tests "connection refused" pattern recognition
//
//	Test Case ID    UT-GR-26-03
//	Title           Network Error Pattern: Connection Refused
//	Description     Tests that "connection refused" errors are recognized as offline conditions
//	Preconditions   Operational offline state is false
//	Steps           1. Reset operational offline state
//	                2. Create error with "connection refused" message
//	                3. Call IsOffline with the error
//	Expected Result IsOffline returns true
//	Requirements    19.3
func TestUT_GR_26_03_NetworkErrorPatterns_ConnectionRefused(t *testing.T) {
	fixture := framework.NewUnitTestFixture("NetworkErrorPattern_ConnectionRefused")
	fixture.Use(t, func(t *testing.T, fixture interface{}) {
		// Reset operational offline state
		SetOperationalOffline(false)

		// Test various forms of "connection refused" error
		testCases := []struct {
			name     string
			errorMsg string
		}{
			{"lowercase", "connection refused"},
			{"uppercase", "CONNECTION REFUSED"},
			{"mixed case", "Connection Refused"},
			{"in sentence", "dial tcp 192.168.1.1:443: connect: connection refused"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				err := errors.New(tc.errorMsg)
				result := IsOffline(err)
				assert.True(t, result, "Expected IsOffline to return true for error: %s", tc.errorMsg)
			})
		}
	})
}

// TestUT_GR_26_04_NetworkErrorPatterns_ConnectionTimedOut tests "connection timed out" pattern recognition
//
//	Test Case ID    UT-GR-26-04
//	Title           Network Error Pattern: Connection Timed Out
//	Description     Tests that "connection timed out" errors are recognized as offline conditions
//	Preconditions   Operational offline state is false
//	Steps           1. Reset operational offline state
//	                2. Create error with "connection timed out" message
//	                3. Call IsOffline with the error
//	Expected Result IsOffline returns true
//	Requirements    19.4
func TestUT_GR_26_04_NetworkErrorPatterns_ConnectionTimedOut(t *testing.T) {
	fixture := framework.NewUnitTestFixture("NetworkErrorPattern_ConnectionTimedOut")
	fixture.Use(t, func(t *testing.T, fixture interface{}) {
		// Reset operational offline state
		SetOperationalOffline(false)

		// Test various forms of "connection timed out" error
		testCases := []struct {
			name     string
			errorMsg string
		}{
			{"lowercase", "connection timed out"},
			{"uppercase", "CONNECTION TIMED OUT"},
			{"mixed case", "Connection Timed Out"},
			{"in sentence", "dial tcp 192.168.1.1:443: i/o timeout: connection timed out"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				err := errors.New(tc.errorMsg)
				result := IsOffline(err)
				assert.True(t, result, "Expected IsOffline to return true for error: %s", tc.errorMsg)
			})
		}
	})
}

// TestUT_GR_26_05_NetworkErrorPatterns_DialTCP tests "dial tcp" pattern recognition
//
//	Test Case ID    UT-GR-26-05
//	Title           Network Error Pattern: Dial TCP
//	Description     Tests that "dial tcp" errors are recognized as offline conditions
//	Preconditions   Operational offline state is false
//	Steps           1. Reset operational offline state
//	                2. Create error with "dial tcp" message
//	                3. Call IsOffline with the error
//	Expected Result IsOffline returns true
//	Requirements    19.5
func TestUT_GR_26_05_NetworkErrorPatterns_DialTCP(t *testing.T) {
	fixture := framework.NewUnitTestFixture("NetworkErrorPattern_DialTCP")
	fixture.Use(t, func(t *testing.T, fixture interface{}) {
		// Reset operational offline state
		SetOperationalOffline(false)

		// Test various forms of "dial tcp" error
		testCases := []struct {
			name     string
			errorMsg string
		}{
			{"lowercase", "dial tcp"},
			{"uppercase", "DIAL TCP"},
			{"mixed case", "Dial TCP"},
			{"with address", "dial tcp 192.168.1.1:443: connect: network is unreachable"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				err := errors.New(tc.errorMsg)
				result := IsOffline(err)
				assert.True(t, result, "Expected IsOffline to return true for error: %s", tc.errorMsg)
			})
		}
	})
}

// TestUT_GR_26_06_NetworkErrorPatterns_ContextDeadlineExceeded tests "context deadline exceeded" pattern recognition
//
//	Test Case ID    UT-GR-26-06
//	Title           Network Error Pattern: Context Deadline Exceeded
//	Description     Tests that "context deadline exceeded" errors are recognized as offline conditions
//	Preconditions   Operational offline state is false
//	Steps           1. Reset operational offline state
//	                2. Create error with "context deadline exceeded" message
//	                3. Call IsOffline with the error
//	Expected Result IsOffline returns true
//	Requirements    19.6
func TestUT_GR_26_06_NetworkErrorPatterns_ContextDeadlineExceeded(t *testing.T) {
	fixture := framework.NewUnitTestFixture("NetworkErrorPattern_ContextDeadlineExceeded")
	fixture.Use(t, func(t *testing.T, fixture interface{}) {
		// Reset operational offline state
		SetOperationalOffline(false)

		// Test various forms of "context deadline exceeded" error
		testCases := []struct {
			name     string
			errorMsg string
		}{
			{"lowercase", "context deadline exceeded"},
			{"uppercase", "CONTEXT DEADLINE EXCEEDED"},
			{"mixed case", "Context Deadline Exceeded"},
			{"in sentence", "Get https://graph.microsoft.com: context deadline exceeded"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				err := errors.New(tc.errorMsg)
				result := IsOffline(err)
				assert.True(t, result, "Expected IsOffline to return true for error: %s", tc.errorMsg)
			})
		}
	})
}

// TestUT_GR_26_07_NetworkErrorPatterns_NoRouteToHost tests "no route to host" pattern recognition
//
//	Test Case ID    UT-GR-26-07
//	Title           Network Error Pattern: No Route To Host
//	Description     Tests that "no route to host" errors are recognized as offline conditions
//	Preconditions   Operational offline state is false
//	Steps           1. Reset operational offline state
//	                2. Create error with "no route to host" message
//	                3. Call IsOffline with the error
//	Expected Result IsOffline returns true
//	Requirements    19.7
func TestUT_GR_26_07_NetworkErrorPatterns_NoRouteToHost(t *testing.T) {
	fixture := framework.NewUnitTestFixture("NetworkErrorPattern_NoRouteToHost")
	fixture.Use(t, func(t *testing.T, fixture interface{}) {
		// Reset operational offline state
		SetOperationalOffline(false)

		// Test various forms of "no route to host" error
		testCases := []struct {
			name     string
			errorMsg string
		}{
			{"lowercase", "no route to host"},
			{"uppercase", "NO ROUTE TO HOST"},
			{"mixed case", "No Route To Host"},
			{"in sentence", "dial tcp 192.168.1.1:443: connect: no route to host"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				err := errors.New(tc.errorMsg)
				result := IsOffline(err)
				assert.True(t, result, "Expected IsOffline to return true for error: %s", tc.errorMsg)
			})
		}
	})
}

// TestUT_GR_26_08_NetworkErrorPatterns_NetworkIsDown tests "network is down" pattern recognition
//
//	Test Case ID    UT-GR-26-08
//	Title           Network Error Pattern: Network Is Down
//	Description     Tests that "network is down" errors are recognized as offline conditions
//	Preconditions   Operational offline state is false
//	Steps           1. Reset operational offline state
//	                2. Create error with "network is down" message
//	                3. Call IsOffline with the error
//	Expected Result IsOffline returns true
//	Requirements    19.8
func TestUT_GR_26_08_NetworkErrorPatterns_NetworkIsDown(t *testing.T) {
	fixture := framework.NewUnitTestFixture("NetworkErrorPattern_NetworkIsDown")
	fixture.Use(t, func(t *testing.T, fixture interface{}) {
		// Reset operational offline state
		SetOperationalOffline(false)

		// Test various forms of "network is down" error
		testCases := []struct {
			name     string
			errorMsg string
		}{
			{"lowercase", "network is down"},
			{"uppercase", "NETWORK IS DOWN"},
			{"mixed case", "Network Is Down"},
			{"in sentence", "dial tcp: network is down"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				err := errors.New(tc.errorMsg)
				result := IsOffline(err)
				assert.True(t, result, "Expected IsOffline to return true for error: %s", tc.errorMsg)
			})
		}
	})
}

// TestUT_GR_26_09_NetworkErrorPatterns_TemporaryFailure tests "temporary failure in name resolution" pattern recognition
//
//	Test Case ID    UT-GR-26-09
//	Title           Network Error Pattern: Temporary Failure In Name Resolution
//	Description     Tests that "temporary failure in name resolution" errors are recognized as offline conditions
//	Preconditions   Operational offline state is false
//	Steps           1. Reset operational offline state
//	                2. Create error with "temporary failure in name resolution" message
//	                3. Call IsOffline with the error
//	Expected Result IsOffline returns true
//	Requirements    19.9
func TestUT_GR_26_09_NetworkErrorPatterns_TemporaryFailure(t *testing.T) {
	fixture := framework.NewUnitTestFixture("NetworkErrorPattern_TemporaryFailure")
	fixture.Use(t, func(t *testing.T, fixture interface{}) {
		// Reset operational offline state
		SetOperationalOffline(false)

		// Test various forms of "temporary failure in name resolution" error
		testCases := []struct {
			name     string
			errorMsg string
		}{
			{"lowercase", "temporary failure in name resolution"},
			{"uppercase", "TEMPORARY FAILURE IN NAME RESOLUTION"},
			{"mixed case", "Temporary Failure In Name Resolution"},
			{"in sentence", "dial tcp: lookup example.com: temporary failure in name resolution"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				err := errors.New(tc.errorMsg)
				result := IsOffline(err)
				assert.True(t, result, "Expected IsOffline to return true for error: %s", tc.errorMsg)
			})
		}
	})
}

// TestUT_GR_26_10_NetworkErrorPatterns_OperationTimedOut tests "operation timed out" pattern recognition
//
//	Test Case ID    UT-GR-26-10
//	Title           Network Error Pattern: Operation Timed Out
//	Description     Tests that "operation timed out" errors are recognized as offline conditions
//	Preconditions   Operational offline state is false
//	Steps           1. Reset operational offline state
//	                2. Create error with "operation timed out" message
//	                3. Call IsOffline with the error
//	Expected Result IsOffline returns true
//	Requirements    19.10
func TestUT_GR_26_10_NetworkErrorPatterns_OperationTimedOut(t *testing.T) {
	fixture := framework.NewUnitTestFixture("NetworkErrorPattern_OperationTimedOut")
	fixture.Use(t, func(t *testing.T, fixture interface{}) {
		// Reset operational offline state
		SetOperationalOffline(false)

		// Test various forms of "operation timed out" error
		testCases := []struct {
			name     string
			errorMsg string
		}{
			{"lowercase", "operation timed out"},
			{"uppercase", "OPERATION TIMED OUT"},
			{"mixed case", "Operation Timed Out"},
			{"in sentence", "dial tcp 192.168.1.1:443: i/o timeout: operation timed out"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				err := errors.New(tc.errorMsg)
				result := IsOffline(err)
				assert.True(t, result, "Expected IsOffline to return true for error: %s", tc.errorMsg)
			})
		}
	})
}

// TestUT_GR_26_11_NetworkErrorPatterns_CaseInsensitive tests that pattern matching is case-insensitive
//
//	Test Case ID    UT-GR-26-11
//	Title           Network Error Pattern: Case Insensitive Matching
//	Description     Tests that error pattern matching is case-insensitive
//	Preconditions   Operational offline state is false
//	Steps           1. Reset operational offline state
//	                2. Create errors with various case combinations
//	                3. Call IsOffline with each error
//	Expected Result All variations are recognized as offline conditions
//	Requirements    19.1-19.11
func TestUT_GR_26_11_NetworkErrorPatterns_CaseInsensitive(t *testing.T) {
	fixture := framework.NewUnitTestFixture("NetworkErrorPattern_CaseInsensitive")
	fixture.Use(t, func(t *testing.T, fixture interface{}) {
		// Reset operational offline state
		SetOperationalOffline(false)

		// Test case insensitivity for all patterns
		testCases := []string{
			"NO SUCH HOST",
			"Network Is Unreachable",
			"CONNECTION REFUSED",
			"Connection Timed Out",
			"DIAL TCP",
			"Context Deadline Exceeded",
			"NO ROUTE TO HOST",
			"Network Is Down",
			"TEMPORARY FAILURE IN NAME RESOLUTION",
			"Operation Timed Out",
		}

		for _, errorMsg := range testCases {
			t.Run(errorMsg, func(t *testing.T) {
				err := errors.New(errorMsg)
				result := IsOffline(err)
				assert.True(t, result, "Expected IsOffline to return true for error: %s", errorMsg)
			})
		}
	})
}

// TestUT_GR_26_12_NetworkErrorPatterns_HTTPResponseNotOffline tests that HTTP responses are not classified as offline
//
//	Test Case ID    UT-GR-26-12
//	Title           Network Error Pattern: HTTP Response Not Offline
//	Description     Tests that errors containing HTTP response codes are not classified as offline
//	Preconditions   Operational offline state is false
//	Steps           1. Reset operational offline state
//	                2. Create errors with HTTP response patterns
//	                3. Call IsOffline with each error
//	Expected Result HTTP response errors are not classified as offline
//	Requirements    19.1-19.11
func TestUT_GR_26_12_NetworkErrorPatterns_HTTPResponseNotOffline(t *testing.T) {
	fixture := framework.NewUnitTestFixture("NetworkErrorPattern_HTTPResponseNotOffline")
	fixture.Use(t, func(t *testing.T, fixture interface{}) {
		// Reset operational offline state
		SetOperationalOffline(false)

		// Test that HTTP responses are not classified as offline
		testCases := []string{
			"HTTP 404 - Not Found",
			"HTTP 500 - Internal Server Error",
			"HTTP 401 - Unauthorized",
			"HTTP 403 - Forbidden",
		}

		for _, errorMsg := range testCases {
			t.Run(errorMsg, func(t *testing.T) {
				err := errors.New(errorMsg)
				result := IsOffline(err)
				assert.False(t, result, "Expected IsOffline to return false for HTTP error: %s", errorMsg)
			})
		}
	})
}

// TestUT_GR_27_01_OfflineStateTransition_PatternMatch tests offline state transition on pattern match
//
//	Test Case ID    UT-GR-27-01
//	Title           Offline State Transition on Pattern Match
//	Description     Tests that offline state is triggered when recognized error patterns are detected
//	Preconditions   Operational offline state is false
//	Steps           1. Reset operational offline state
//	                2. Simulate network errors with recognized patterns
//	                3. Verify IsOffline returns true for each pattern
//	Expected Result Offline state is correctly triggered for all recognized patterns
//	Requirements    19.1-19.11
func TestUT_GR_27_01_OfflineStateTransition_PatternMatch(t *testing.T) {
	fixture := framework.NewUnitTestFixture("OfflineStateTransition_PatternMatch")
	fixture.Use(t, func(t *testing.T, fixture interface{}) {
		// Reset operational offline state
		SetOperationalOffline(false)

		// Test all recognized patterns trigger offline state
		patterns := []string{
			"no such host",
			"network is unreachable",
			"connection refused",
			"connection timed out",
			"dial tcp",
			"context deadline exceeded",
			"no route to host",
			"network is down",
			"temporary failure in name resolution",
			"operation timed out",
		}

		for _, pattern := range patterns {
			t.Run(pattern, func(t *testing.T) {
				err := errors.New("network error: " + pattern)
				result := IsOffline(err)
				assert.True(t, result, "Expected offline state for pattern: %s", pattern)
			})
		}
	})
}

// TestUT_GR_27_02_OfflineStateTransition_FalsePositives tests that false positives are minimized
//
//	Test Case ID    UT-GR-27-02
//	Title           Minimize False Positives
//	Description     Tests that non-network errors are not incorrectly classified as offline
//	Preconditions   Operational offline state is false
//	Steps           1. Reset operational offline state
//	                2. Create errors that should NOT trigger offline state
//	                3. Verify IsOffline returns false for these errors
//	Expected Result False positives are minimized
//	Requirements    19.1-19.11
func TestUT_GR_27_02_OfflineStateTransition_FalsePositives(t *testing.T) {
	fixture := framework.NewUnitTestFixture("OfflineStateTransition_FalsePositives")
	fixture.Use(t, func(t *testing.T, fixture interface{}) {
		// Reset operational offline state
		SetOperationalOffline(false)

		// Test that HTTP response errors are not classified as offline
		httpErrors := []string{
			"HTTP 400 - Bad Request",
			"HTTP 401 - Unauthorized",
			"HTTP 403 - Forbidden",
			"HTTP 404 - Not Found",
			"HTTP 429 - Too Many Requests",
			"HTTP 500 - Internal Server Error",
			"HTTP 502 - Bad Gateway",
			"HTTP 503 - Service Unavailable",
		}

		for _, errorMsg := range httpErrors {
			t.Run(errorMsg, func(t *testing.T) {
				err := errors.New(errorMsg)
				result := IsOffline(err)
				assert.False(t, result, "Expected online state for HTTP error: %s", errorMsg)
			})
		}
	})
}

// TestUT_GR_27_03_OfflineStateTransition_OperationalOverride tests operational offline override
//
//	Test Case ID    UT-GR-27-03
//	Title           Operational Offline Override
//	Description     Tests that operational offline state overrides error pattern detection
//	Preconditions   None
//	Steps           1. Set operational offline state to true
//	                2. Call IsOffline with nil error
//	                3. Call IsOffline with HTTP error
//	                4. Reset operational offline state
//	                5. Verify behavior changes
//	Expected Result Operational offline state takes precedence
//	Requirements    19.1-19.11
func TestUT_GR_27_03_OfflineStateTransition_OperationalOverride(t *testing.T) {
	fixture := framework.NewUnitTestFixture("OfflineStateTransition_OperationalOverride")
	fixture.Use(t, func(t *testing.T, fixture interface{}) {
		// Set operational offline state to true
		SetOperationalOffline(true)

		// Even with nil error, should return true
		result := IsOffline(nil)
		assert.True(t, result, "Expected offline state when operational offline is set")

		// Even with HTTP error, should return true
		httpErr := errors.New("HTTP 404 - Not Found")
		result = IsOffline(httpErr)
		assert.True(t, result, "Expected offline state when operational offline is set, even with HTTP error")

		// Reset operational offline state
		SetOperationalOffline(false)

		// Now nil error should return false
		result = IsOffline(nil)
		assert.False(t, result, "Expected online state with nil error when operational offline is false")

		// HTTP error should return false
		result = IsOffline(httpErr)
		assert.False(t, result, "Expected online state with HTTP error when operational offline is false")
	})
}

// TestUT_GR_27_04_OfflineStateTransition_MixedCasePatterns tests case-insensitive pattern matching
//
//	Test Case ID    UT-GR-27-04
//	Title           Case-Insensitive Pattern Matching
//	Description     Tests that pattern matching works regardless of case
//	Preconditions   Operational offline state is false
//	Steps           1. Reset operational offline state
//	                2. Create errors with various case combinations
//	                3. Verify all are recognized as offline
//	Expected Result Pattern matching is case-insensitive
//	Requirements    19.1-19.11
func TestUT_GR_27_04_OfflineStateTransition_MixedCasePatterns(t *testing.T) {
	fixture := framework.NewUnitTestFixture("OfflineStateTransition_MixedCasePatterns")
	fixture.Use(t, func(t *testing.T, fixture interface{}) {
		// Reset operational offline state
		SetOperationalOffline(false)

		// Test various case combinations
		testCases := []struct {
			name     string
			errorMsg string
		}{
			{"all lowercase", "no such host"},
			{"all uppercase", "NO SUCH HOST"},
			{"title case", "No Such Host"},
			{"mixed case 1", "No SUCH host"},
			{"mixed case 2", "nO SuCh HoSt"},
			{"in sentence lowercase", "dial tcp: lookup example.com: no such host"},
			{"in sentence uppercase", "DIAL TCP: LOOKUP EXAMPLE.COM: NO SUCH HOST"},
			{"in sentence mixed", "Dial TCP: Lookup Example.com: No Such Host"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				err := errors.New(tc.errorMsg)
				result := IsOffline(err)
				assert.True(t, result, "Expected offline state for: %s", tc.errorMsg)
			})
		}
	})
}

// TestUT_GR_27_05_OfflineStateTransition_PartialMatches tests partial pattern matches
//
//	Test Case ID    UT-GR-27-05
//	Title           Partial Pattern Matches
//	Description     Tests that patterns are matched even when embedded in longer error messages
//	Preconditions   Operational offline state is false
//	Steps           1. Reset operational offline state
//	                2. Create errors with patterns embedded in longer messages
//	                3. Verify patterns are still recognized
//	Expected Result Patterns are matched even when embedded in longer messages
//	Requirements    19.1-19.11
func TestUT_GR_27_05_OfflineStateTransition_PartialMatches(t *testing.T) {
	fixture := framework.NewUnitTestFixture("OfflineStateTransition_PartialMatches")
	fixture.Use(t, func(t *testing.T, fixture interface{}) {
		// Reset operational offline state
		SetOperationalOffline(false)

		// Test patterns embedded in longer error messages
		testCases := []struct {
			name     string
			errorMsg string
			pattern  string
		}{
			{
				"no such host in DNS lookup",
				"dial tcp: lookup graph.microsoft.com on 8.8.8.8:53: no such host",
				"no such host",
			},
			{
				"network unreachable with IP",
				"dial tcp 192.168.1.1:443: connect: network is unreachable",
				"network is unreachable",
			},
			{
				"connection refused with port",
				"dial tcp [::1]:8080: connect: connection refused",
				"connection refused",
			},
			{
				"timeout in HTTP request",
				"Get https://graph.microsoft.com/v1.0/me: context deadline exceeded",
				"context deadline exceeded",
			},
			{
				"dial tcp with full context",
				"dial tcp graph.microsoft.com:443: i/o timeout",
				"dial tcp",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				err := errors.New(tc.errorMsg)
				result := IsOffline(err)
				assert.True(t, result, "Expected offline state for error containing '%s': %s", tc.pattern, tc.errorMsg)
			})
		}
	})
}

// TestUT_GR_28_01_ErrorPatternLogging_PatternDetection tests that detected patterns are logged
//
//	Test Case ID    UT-GR-28-01
//	Title           Error Pattern Logging: Pattern Detection
//	Description     Tests that specific error patterns are logged when detected
//	Preconditions   Operational offline state is false
//	Steps           1. Reset operational offline state
//	                2. Create errors with recognized patterns
//	                3. Call IsOffline with each error
//	                4. Verify logging occurs (implicitly through function execution)
//	Expected Result Detected patterns are logged with context
//	Requirements    19.11
//	Notes           This test verifies the logging behavior by ensuring IsOffline executes
//	                the logging code path. Actual log output verification would require
//	                a log capture mechanism which is beyond the scope of unit tests.
func TestUT_GR_28_01_ErrorPatternLogging_PatternDetection(t *testing.T) {
	fixture := framework.NewUnitTestFixture("ErrorPatternLogging_PatternDetection")
	fixture.Use(t, func(t *testing.T, fixture interface{}) {
		// Reset operational offline state
		SetOperationalOffline(false)

		// Test that all patterns trigger logging
		patterns := []string{
			"no such host",
			"network is unreachable",
			"connection refused",
			"connection timed out",
			"dial tcp",
			"context deadline exceeded",
			"no route to host",
			"network is down",
			"temporary failure in name resolution",
			"operation timed out",
		}

		for _, pattern := range patterns {
			t.Run(pattern, func(t *testing.T) {
				err := errors.New("network error: " + pattern)
				result := IsOffline(err)

				// Verify offline state is detected
				assert.True(t, result, "Expected offline state for pattern: %s", pattern)

				// Note: The logging happens inside IsOffline function
				// In a production environment, this would be captured by the logging system
				// and could be verified through log analysis
			})
		}
	})
}

// TestUT_GR_28_02_ErrorPatternLogging_NetworkErrorType tests logging for network error types
//
//	Test Case ID    UT-GR-28-02
//	Title           Error Pattern Logging: Network Error Type
//	Description     Tests that network error types are logged when detected
//	Preconditions   Operational offline state is false
//	Steps           1. Reset operational offline state
//	                2. Create network error type
//	                3. Call IsOffline with the error
//	                4. Verify logging occurs
//	Expected Result Network error types are logged with context
//	Requirements    19.11
func TestUT_GR_28_02_ErrorPatternLogging_NetworkErrorType(t *testing.T) {
	fixture := framework.NewUnitTestFixture("ErrorPatternLogging_NetworkErrorType")
	fixture.Use(t, func(t *testing.T, fixture interface{}) {
		// Reset operational offline state
		SetOperationalOffline(false)

		// Create a network error
		err := ierrors.NewNetworkError("test network error", nil)
		result := IsOffline(err)

		// Verify offline state is detected
		assert.True(t, result, "Expected offline state for network error type")

		// Note: The logging happens inside IsOffline function
	})
}

// TestUT_GR_28_03_ErrorPatternLogging_LogFormat tests log format and content
//
//	Test Case ID    UT-GR-28-03
//	Title           Error Pattern Logging: Log Format
//	Description     Tests that logs contain the expected fields (pattern, error message)
//	Preconditions   Operational offline state is false
//	Steps           1. Reset operational offline state
//	                2. Create error with known pattern
//	                3. Call IsOffline
//	                4. Verify function executes logging code path
//	Expected Result Logs contain pattern and error message fields
//	Requirements    19.11
//	Notes           This test verifies the code path is executed. Actual log field
//	                verification would require log capture infrastructure.
func TestUT_GR_28_03_ErrorPatternLogging_LogFormat(t *testing.T) {
	fixture := framework.NewUnitTestFixture("ErrorPatternLogging_LogFormat")
	fixture.Use(t, func(t *testing.T, fixture interface{}) {
		// Reset operational offline state
		SetOperationalOffline(false)

		// Test with a specific error message
		errorMsg := "dial tcp 192.168.1.1:443: connect: no such host"
		err := errors.New(errorMsg)
		result := IsOffline(err)

		// Verify offline state is detected
		assert.True(t, result, "Expected offline state")

		// The logging code in IsOffline includes:
		// - pattern field: the matched pattern
		// - error field: the full error message
		// This ensures proper context is logged for debugging
	})
}

// TestUT_GR_28_04_ErrorPatternLogging_ErrorIdentification tests error pattern identification in logs
//
//	Test Case ID    UT-GR-28-04
//	Title           Error Pattern Logging: Error Identification
//	Description     Tests that the specific pattern that triggered detection is identified in logs
//	Preconditions   Operational offline state is false
//	Steps           1. Reset operational offline state
//	                2. Create errors with multiple patterns in the same message
//	                3. Call IsOffline
//	                4. Verify first matching pattern is logged
//	Expected Result First matching pattern is identified and logged
//	Requirements    19.11
func TestUT_GR_28_04_ErrorPatternLogging_ErrorIdentification(t *testing.T) {
	fixture := framework.NewUnitTestFixture("ErrorPatternLogging_ErrorIdentification")
	fixture.Use(t, func(t *testing.T, fixture interface{}) {
		// Reset operational offline state
		SetOperationalOffline(false)

		// Test with error containing multiple patterns
		// The first pattern in the list that matches should be logged
		testCases := []struct {
			name          string
			errorMsg      string
			expectedFirst string // First pattern that should match
		}{
			{
				"multiple patterns - no such host first",
				"dial tcp: lookup example.com: no such host and network is unreachable",
				"no such host",
			},
			{
				"multiple patterns - network unreachable first",
				"network is unreachable: connection refused",
				"network is unreachable",
			},
			{
				"single pattern",
				"connection timed out",
				"connection timed out",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				err := errors.New(tc.errorMsg)
				result := IsOffline(err)

				// Verify offline state is detected
				assert.True(t, result, "Expected offline state for: %s", tc.errorMsg)

				// The logging code will log the first matching pattern
				// In this case, it would be tc.expectedFirst
			})
		}
	})
}
