package graph

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOperationalOfflineState(t *testing.T) {
	// Reset operational offline state before and after test
	SetOperationalOffline(false)
	defer SetOperationalOffline(false)

	// Test default state
	assert.False(t, GetOperationalOffline(), "Default operational offline state should be false")

	// Test setting to true
	SetOperationalOffline(true)
	assert.True(t, GetOperationalOffline(), "Operational offline state should be true after setting it")

	// Test setting back to false
	SetOperationalOffline(false)
	assert.False(t, GetOperationalOffline(), "Operational offline state should be false after setting it back")
}

func TestIsOfflineWithOperationalState(t *testing.T) {
	// Reset operational offline state before and after test
	SetOperationalOffline(false)
	defer SetOperationalOffline(false)

	// Test with operational offline set to true
	SetOperationalOffline(true)

	// Should return true regardless of error
	assert.True(t, IsOffline(nil), "IsOffline should return true when operational offline is set, even with nil error")
	assert.True(t, IsOffline(errors.New("HTTP 404 - Not Found")), "IsOffline should return true when operational offline is set, even with HTTP error")

	// Reset operational offline state
	SetOperationalOffline(false)

	// Now should behave normally based on error
	assert.False(t, IsOffline(nil), "IsOffline should return false with nil error when operational offline is not set")
	assert.False(t, IsOffline(errors.New("HTTP 404 - Not Found")), "IsOffline should return false with HTTP error when operational offline is not set")
	assert.True(t, IsOffline(errors.New("network error")), "IsOffline should return true with network error when operational offline is not set")
}

func TestIsOffline(t *testing.T) {
	// Reset operational offline state before test
	SetOperationalOffline(false)
	defer SetOperationalOffline(false)

	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "HTTP error",
			err:      errors.New("HTTP 404 - itemNotFound: The resource could not be found."),
			expected: false,
		},
		{
			name:     "HTTP error with different format",
			err:      errors.New("HTTP 500 - Internal Server Error"),
			expected: false,
		},
		{
			name:     "network error",
			err:      errors.New("Get \"https://graph.microsoft.com/v1.0/me/drive\": dial tcp: lookup graph.microsoft.com: no such host"),
			expected: true,
		},
		{
			name:     "timeout error",
			err:      errors.New("Get \"https://graph.microsoft.com/v1.0/me/drive\": context deadline exceeded"),
			expected: true,
		},
		{
			name:     "connection refused error",
			err:      errors.New("Get \"https://graph.microsoft.com/v1.0/me/drive\": dial tcp: connect: connection refused"),
			expected: true,
		},
	}

	for _, tc := range tests {
		tc := tc // capture range variable
		t.Run(tc.name, func(t *testing.T) {
			// Ensure we're not in operational offline mode for these tests
			require.False(t, GetOperationalOffline(), "Operational offline should be false for these tests")

			result := IsOffline(tc.err)
			assert.Equal(t, tc.expected, result, "IsOffline returned incorrect result")
		})
	}
}
