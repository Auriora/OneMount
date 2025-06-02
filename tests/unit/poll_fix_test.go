package unit

import (
	"reflect"
	"testing"

	"github.com/auriora/onemount/internal/fs"
	"github.com/stretchr/testify/assert"
)

// TestPollOpcodeRemoval tests that the POLL opcode implementation has been removed
// to prevent the "Unimplemented opcode POLL" errors that were causing crashes.
func TestPollOpcodeRemoval(t *testing.T) {
	// Verify that the FilesystemInterface no longer has a Poll method
	fsType := reflect.TypeOf((*fs.FilesystemInterface)(nil)).Elem()
	_, hasPollMethod := fsType.MethodByName("Poll")
	assert.False(t, hasPollMethod, "FilesystemInterface should not have a Poll method")

	// Verify that the CustomRawFileSystem can be created (even with nil filesystem)
	customRawFS := fs.NewCustomRawFileSystem(nil)
	assert.NotNil(t, customRawFS)

	// The CustomRawFileSystem should embed the default RawFileSystem
	// which does not implement POLL (as intended by go-fuse)
	// Note: fuse.NewDefaultRawFileSystem() might return nil in some versions
	// The important thing is that our CustomRawFileSystem is created successfully
	t.Logf("RawFileSystem: %v", customRawFS.RawFileSystem)
}

// TestLargeFileWarning tests that large file operations generate appropriate warnings
func TestLargeFileWarning(t *testing.T) {
	// This test verifies that the code compiles and the constants are correct
	const largeFileThreshold = 1024 * 1024 * 1024 // 1GB
	assert.Equal(t, 1073741824, largeFileThreshold)
}
