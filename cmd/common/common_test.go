package common

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Write a sample .xdg-volume-info file and check that it can be read.
func TestXDGVolumeInfo(t *testing.T) {
	t.Parallel()

	const expected = "some-volume name *()! $"
	content := TemplateXDGVolumeInfo(expected)

	file, err := os.CreateTemp("", "onedriver-test-*")
	require.NoError(t, err, "Failed to create temporary file")

	t.Cleanup(func() {
		if err := os.Remove(file.Name()); err != nil && !os.IsNotExist(err) {
			t.Logf("Warning: Failed to clean up test file %s: %v", file.Name(), err)
		}
	})

	require.NoError(t, os.WriteFile(file.Name(), []byte(content), 0600), "Failed to write to temporary file")

	driveName, err := GetXDGVolumeInfoName(file.Name())
	require.NoError(t, err, "Failed to get XDG volume info name")
	assert.Equal(t, expected, driveName, "Drive name did not match expected value")
}
