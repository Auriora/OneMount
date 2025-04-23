package graph

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsOffline(t *testing.T) {
	t.Parallel()

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
			t.Parallel()
			result := IsOffline(tc.err)
			assert.Equal(t, tc.expected, result, "IsOffline returned incorrect result")
		})
	}
}
