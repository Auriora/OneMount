package graph

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIDPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		id       string
		expected string
	}{
		{
			name:     "root ID",
			id:       "root",
			expected: "/me/drive/root",
		},
		{
			name:     "regular ID",
			id:       "123456",
			expected: "/me/drive/items/123456",
		},
		{
			name:     "ID with special characters",
			id:       "abc/123",
			expected: "/me/drive/items/abc%2F123",
		},
	}

	for _, tc := range tests {
		tc := tc // capture range variable
		t.Run(tc.name, func(t *testing.T) {
			result := IDPath(tc.id)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestChildrenPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "root children",
			path:     "/",
			expected: "/me/drive/root/children",
		},
		{
			name:     "simple path children",
			path:     "/OneMount-Documents",
			expected: "/me/drive/root:%2FOneMount-Documents:/children",
		},
		{
			name:     "nested path children",
			path:     "/OneMount-Documents/Work",
			expected: "/me/drive/root:%2FOneMount-Documents%2FWork:/children",
		},
		{
			name:     "path with spaces children",
			path:     "/OneMount My Documents/Work Files",
			expected: "/me/drive/root:%2FOneMount%20My%20Documents%2FWork%20Files:/children",
		},
	}

	for _, tc := range tests {
		tc := tc // capture range variable
		t.Run(tc.name, func(t *testing.T) {
			result := childrenPath(tc.path)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestChildrenPathID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		id       string
		expected string
	}{
		{
			name:     "root ID children",
			id:       "root",
			expected: "/me/drive/items/root/children",
		},
		{
			name:     "regular ID children",
			id:       "123456",
			expected: "/me/drive/items/123456/children",
		},
		{
			name:     "ID with special characters children",
			id:       "abc/123",
			expected: "/me/drive/items/abc%2F123/children",
		},
	}

	for _, tc := range tests {
		tc := tc // capture range variable
		t.Run(tc.name, func(t *testing.T) {
			result := childrenPathID(tc.id)
			assert.Equal(t, tc.expected, result)
		})
	}
}
