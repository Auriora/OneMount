package graph

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestResourcePath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "path with special characters",
			path:     "/some path/here!",
			expected: "/me/drive/root:%2Fsome%20path%2Fhere%21",
		},
		{
			name:     "root path",
			path:     "/",
			expected: "/me/drive/root",
		},
		{
			name:     "simple path",
			path:     "/Onedriver-Documents",
			expected: "/me/drive/root:%2FOnedriver-Documents",
		},
		{
			name:     "nested path",
			path:     "/Onedriver-Documents/Work",
			expected: "/me/drive/root:%2FOnedriver-Documents%2FWork",
		},
		{
			name:     "path with spaces",
			path:     "/Onedriver My Documents/Work Files",
			expected: "/me/drive/root:%2FOnedriver-%20My%20Documents%2FWork%20Files",
		},
	}

	for _, tc := range tests {
		tc := tc // capture range variable
		t.Run(tc.name, func(t *testing.T) {
			result := ResourcePath(tc.path)
			assert.Equal(t, tc.expected, result, "Escaped path was wrong.")
		})
	}
}

func TestRequestUnauthenticated(t *testing.T) {
	t.Parallel()
	badAuth := &Auth{
		// Set a renewal 1 year in the future so we don't accidentally overwrite
		// our auth tokens
		ExpiresAt: time.Now().Unix() + 60*60*24*365,
	}
	_, err := Get("/me/drive/root", badAuth)
	assert.Error(t, err, "An unauthenticated request was not handled as an error")
}
