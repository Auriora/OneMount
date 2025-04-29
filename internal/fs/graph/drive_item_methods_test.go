package graph

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDriveItemIsDir(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		item     DriveItem
		expected bool
	}{
		{
			name: "folder item",
			item: DriveItem{
				Name:   "Test Folder",
				Folder: &Folder{},
			},
			expected: true,
		},
		{
			name: "file item",
			item: DriveItem{
				Name: "Test File",
				File: &File{},
			},
			expected: false,
		},
		{
			name: "empty item",
			item: DriveItem{
				Name: "Empty Item",
			},
			expected: false,
		},
	}

	for _, tc := range tests {
		tc := tc // capture range variable
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := tc.item.IsDir()
			assert.Equal(t, tc.expected, result, "IsDir returned incorrect result")
		})
	}
}

func TestDriveItemModTimeUnix(t *testing.T) {
	t.Parallel()

	// Create a fixed time for testing
	now := time.Now()
	unixTime := uint64(now.Unix())

	tests := []struct {
		name     string
		item     DriveItem
		expected uint64
	}{
		{
			name: "item with mod time",
			item: DriveItem{
				Name:    "Test Item",
				ModTime: &now,
			},
			expected: unixTime,
		},
	}

	for _, tc := range tests {
		tc := tc // capture range variable
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := tc.item.ModTimeUnix()
			assert.Equal(t, tc.expected, result, "ModTimeUnix returned incorrect result")
		})
	}
}

func TestDriveItemVerifyChecksum(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		item     DriveItem
		checksum string
		expected bool
	}{
		{
			name: "matching checksum",
			item: DriveItem{
				File: &File{
					Hashes: Hashes{
						QuickXorHash: "TestHash123",
					},
				},
			},
			checksum: "TestHash123",
			expected: true,
		},
		{
			name: "non-matching checksum",
			item: DriveItem{
				File: &File{
					Hashes: Hashes{
						QuickXorHash: "TestHash123",
					},
				},
			},
			checksum: "DifferentHash",
			expected: false,
		},
		{
			name: "case-insensitive matching",
			item: DriveItem{
				File: &File{
					Hashes: Hashes{
						QuickXorHash: "TestHash123",
					},
				},
			},
			checksum: "testhash123",
			expected: true,
		},
		{
			name:     "empty checksum",
			item:     DriveItem{},
			checksum: "",
			expected: false,
		},
		{
			name: "nil file",
			item: DriveItem{
				Name: "Test Item",
			},
			checksum: "TestHash123",
			expected: false,
		},
	}

	for _, tc := range tests {
		tc := tc // capture range variable
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := tc.item.VerifyChecksum(tc.checksum)
			assert.Equal(t, tc.expected, result, "VerifyChecksum returned incorrect result")
		})
	}
}

func TestDriveItemETagIsMatch(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		item     DriveItem
		etag     string
		expected bool
	}{
		{
			name: "matching etag",
			item: DriveItem{
				ETag: "\"12345\"",
			},
			etag:     "\"12345\"",
			expected: true,
		},
		{
			name: "non-matching etag",
			item: DriveItem{
				ETag: "\"12345\"",
			},
			etag:     "\"67890\"",
			expected: false,
		},
		{
			name:     "empty etag in item",
			item:     DriveItem{},
			etag:     "\"12345\"",
			expected: false,
		},
		{
			name: "empty etag parameter",
			item: DriveItem{
				ETag: "\"12345\"",
			},
			etag:     "",
			expected: false,
		},
	}

	for _, tc := range tests {
		tc := tc // capture range variable
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := tc.item.ETagIsMatch(tc.etag)
			assert.Equal(t, tc.expected, result, "ETagIsMatch returned incorrect result")
		})
	}
}
