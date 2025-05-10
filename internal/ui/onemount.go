package ui

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/auriora/onemount/pkg/graph"
	"github.com/rs/zerolog/log"
)

// onemount specific utility functions

// PollUntilAvail will block until the mountpoint is available or a timeout is reached.
// If timeout is -1, default timeout is 120s.
func PollUntilAvail(mountpoint string, timeout int) bool {
	if timeout == -1 {
		timeout = 120
	}
	for i := 1; i < timeout*10; i++ {
		_, err := os.Stat(filepath.Join(mountpoint, ".xdg-volume-info"))
		if err == nil {
			return true
		}
		time.Sleep(100 * time.Millisecond)
	}
	return false
}

// MountpointIsValid returns if the mountpoint exists and is suitable for mounting.
// A directory is considered valid if it exists and is a directory.
func MountpointIsValid(mountpoint string) bool {
	// Check if the path exists and is a directory
	info, err := os.Stat(mountpoint)
	if err != nil {
		return false
	}

	if !info.IsDir() {
		return false
	}

	// The directory should be empty to be valid
	dirents, err := os.ReadDir(mountpoint)
	if err != nil {
		return false
	}

	return len(dirents) == 0
}

// GetKnownMounts returns the currently known mountpoints and returns their escaped name
func GetKnownMounts(cacheDir string) []string {
	mounts := make([]string, 0)

	if cacheDir == "" {
		userCacheDir, _ := os.UserCacheDir()
		cacheDir = filepath.Join(userCacheDir, "onemount")
	}
	os.MkdirAll(cacheDir, 0700)
	dirents, err := os.ReadDir(cacheDir)

	if err != nil {
		log.Error().Err(err).Msg("Could not fetch known mountpoints.")
		return mounts
	}

	for _, dirent := range dirents {
		_, err := os.Stat(graph.GetAuthTokensPath(cacheDir, dirent.Name()))
		if err == nil {
			mounts = append(mounts, dirent.Name())
		}
	}
	return mounts
}

// EscapeHome replaces the user's absolute home directory with "~"
func EscapeHome(path string) string {
	homedir, _ := os.UserHomeDir()
	if strings.HasPrefix(path, homedir) {
		return strings.Replace(path, homedir, "~", 1)
	}
	return path
}

// UnescapeHome replaces the "~" in a path with the absolute path.
func UnescapeHome(path string) string {
	if strings.HasPrefix(path, "~/") {
		homedir, _ := os.UserHomeDir()
		return filepath.Join(homedir, path[2:])
	}
	return path
}
