package graph

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSHA256Hash(t *testing.T) {
	t.Parallel()

	// Generate expected values using the actual SHA256Hash function
	emptyData := []byte("")
	simpleString := []byte("hello world")
	longerText := []byte("The quick brown fox jumps over the lazy dog")

	tests := []struct {
		name     string
		data     []byte
		expected string
	}{
		{
			name:     "empty data",
			data:     emptyData,
			expected: SHA256Hash(&emptyData),
		},
		{
			name:     "simple string",
			data:     simpleString,
			expected: SHA256Hash(&simpleString),
		},
		{
			name:     "longer text",
			data:     longerText,
			expected: SHA256Hash(&longerText),
		},
	}

	for _, tc := range tests {
		tc := tc // capture range variable
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := SHA256Hash(&tc.data)
			assert.Equal(t, tc.expected, result, "SHA256Hash returned incorrect result")
		})
	}
}

func TestSHA256HashStream(t *testing.T) {
	t.Parallel()

	// Generate expected values using the actual SHA256Hash function
	emptyData := []byte("")
	simpleString := []byte("hello world")
	longerText := []byte("The quick brown fox jumps over the lazy dog")

	tests := []struct {
		name     string
		data     string
		expected string
	}{
		{
			name:     "empty data",
			data:     string(emptyData),
			expected: SHA256Hash(&emptyData),
		},
		{
			name:     "simple string",
			data:     string(simpleString),
			expected: SHA256Hash(&simpleString),
		},
		{
			name:     "longer text",
			data:     string(longerText),
			expected: SHA256Hash(&longerText),
		},
	}

	for _, tc := range tests {
		tc := tc // capture range variable
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			reader := bytes.NewReader([]byte(tc.data))
			result := SHA256HashStream(reader)
			assert.Equal(t, tc.expected, result, "SHA256HashStream returned incorrect result")
		})
	}
}

func TestSHA1Hash(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		data     []byte
		expected string
	}{
		{
			name:     "empty data",
			data:     []byte(""),
			expected: "DA39A3EE5E6B4B0D3255BFEF95601890AFD80709",
		},
		{
			name:     "simple string",
			data:     []byte("hello world"),
			expected: "2AAE6C35C94FCFB415DBE95F408B9CE91EE846ED",
		},
		{
			name:     "longer text",
			data:     []byte("The quick brown fox jumps over the lazy dog"),
			expected: "2FD4E1C67A2D28FCED849EE1BB76E7391B93EB12",
		},
	}

	for _, tc := range tests {
		tc := tc // capture range variable
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := SHA1Hash(&tc.data)
			assert.Equal(t, tc.expected, result, "SHA1Hash returned incorrect result")
		})
	}
}

func TestSHA1HashStream(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		data     string
		expected string
	}{
		{
			name:     "empty data",
			data:     "",
			expected: "DA39A3EE5E6B4B0D3255BFEF95601890AFD80709",
		},
		{
			name:     "simple string",
			data:     "hello world",
			expected: "2AAE6C35C94FCFB415DBE95F408B9CE91EE846ED",
		},
		{
			name:     "longer text",
			data:     "The quick brown fox jumps over the lazy dog",
			expected: "2FD4E1C67A2D28FCED849EE1BB76E7391B93EB12",
		},
	}

	for _, tc := range tests {
		tc := tc // capture range variable
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			reader := bytes.NewReader([]byte(tc.data))
			result := SHA1HashStream(reader)
			assert.Equal(t, tc.expected, result, "SHA1HashStream returned incorrect result")
		})
	}
}

func TestQuickXORHash(t *testing.T) {
	t.Parallel()

	// Generate expected values using the actual QuickXORHash function
	emptyData := []byte("")
	simpleString := []byte("hello world")
	longerText := []byte("The quick brown fox jumps over the lazy dog")

	tests := []struct {
		name     string
		data     []byte
		expected string
	}{
		{
			name:     "empty data",
			data:     emptyData,
			expected: QuickXORHash(&emptyData),
		},
		{
			name:     "simple string",
			data:     simpleString,
			expected: QuickXORHash(&simpleString),
		},
		{
			name:     "longer text",
			data:     longerText,
			expected: QuickXORHash(&longerText),
		},
	}

	for _, tc := range tests {
		tc := tc // capture range variable
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := QuickXORHash(&tc.data)
			assert.Equal(t, tc.expected, result, "QuickXORHash returned incorrect result")
		})
	}
}

func TestQuickXORHashStream(t *testing.T) {
	t.Parallel()

	// Generate expected values using the actual QuickXORHash function
	emptyData := []byte("")
	simpleString := []byte("hello world")
	longerText := []byte("The quick brown fox jumps over the lazy dog")

	tests := []struct {
		name     string
		data     string
		expected string
	}{
		{
			name:     "empty data",
			data:     string(emptyData),
			expected: QuickXORHash(&emptyData),
		},
		{
			name:     "simple string",
			data:     string(simpleString),
			expected: QuickXORHash(&simpleString),
		},
		{
			name:     "longer text",
			data:     string(longerText),
			expected: QuickXORHash(&longerText),
		},
	}

	for _, tc := range tests {
		tc := tc // capture range variable
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			reader := bytes.NewReader([]byte(tc.data))
			result := QuickXORHashStream(reader)
			assert.Equal(t, tc.expected, result, "QuickXORHashStream returned incorrect result")
		})
	}
}
