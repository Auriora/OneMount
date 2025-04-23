package systemd

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	os.Chdir("../..")

	// Remove the mount directory if it exists and is a stale mount point
	if _, err := os.Stat("mount"); err == nil {
		if _, err := os.ReadDir("mount"); err != nil {
			// If we can't read the directory, it might be a stale mount point
			os.Remove("mount")
		}
	}

	// Create a fresh mount directory
	os.Mkdir("mount", 0700)

	os.Exit(m.Run())
}
