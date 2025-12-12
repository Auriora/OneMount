// Package common
// Common functions used by both binaries
package common

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/auriora/onemount/internal/fs"
	"github.com/auriora/onemount/internal/graph"
	"github.com/auriora/onemount/internal/logging"
	"github.com/hanwen/go-fuse/v2/fuse"
)

const version = "0.1.0rc1"

var commit string

// Version returns the current version string including git commit information.
// The format is "vversion commit" where commit is truncated to 8 characters.
// This is used for diagnostic output and version reporting.
func Version() string {
	clen := 0
	if len(commit) > 7 {
		clen = 8
	}
	return fmt.Sprintf("v%s %s", version, commit[:clen])
}

// StringToLevel converts a string to a logging.Level that can be used with the logging package.
// If the input string is not a valid log level, DebugLevel is returned as a safe default.
// Valid levels include: trace, debug, info, warn, error, fatal.
func StringToLevel(input string) logging.Level {
	level, err := logging.ParseLevel(input)
	if err != nil {
		logging.Error().Err(err).Msg("Could not parse log level, defaulting to \"debug\"")
		return logging.DebugLevel
	}
	return level
}

// LogLevels returns the available logging levels supported by OneMount.
// These levels can be used in configuration files and command-line arguments
// to control the verbosity of log output.
func LogLevels() []string {
	return []string{"trace", "debug", "info", "warn", "error", "fatal"}
}

// TemplateXDGVolumeInfo returns a formatted .xdg-volume-info file content
func TemplateXDGVolumeInfo(name string) string {
	xdgVolumeInfo := fmt.Sprintf("[Volume Info]\nName=%s\nIcon=dk-onedrive\n", name)
	if _, err := os.Stat("/usr/share/icons/onemount/onemount.png"); err == nil {
		xdgVolumeInfo += "IconFile=/usr/share/icons/onemount.png\n"
	} else {
		xdgVolumeInfo += "Icon=dk-onedrive\n"
	}
	// Add network mount type for Nemo to display it as a network/cloud mount
	return xdgVolumeInfo
}

// GetXDGVolumeInfoName returns the name of the drive according to whatever the
// user has named it.
func GetXDGVolumeInfoName(path string) (string, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	regex := regexp.MustCompile("Name=(.*)")
	name := regex.FindString(string(contents))
	if len(name) < 5 {
		return "", errors.New("could not find \"Name=\" key")
	}
	return name[5:], nil
}

// CreateXDGVolumeInfo creates .xdg-volume-info for a nice little onedrive logo in the
// CreateXDGVolumeInfo creates a .xdg-volume-info file that displays an icon in the
// corner of the mountpoint and shows the account name in the Nautilus sidebar.
// This file is created as a local-only virtual file and is NOT synced to OneDrive.
// The file follows the XDG Volume Info specification for removable media identification.
func CreateXDGVolumeInfo(filesystem *fs.Filesystem, auth *graph.Auth) {
	const fileName = ".xdg-volume-info"

	child, _ := filesystem.GetPath("/.xdg-volume-info", auth)
	if child != nil && !strings.HasPrefix(child.ID(), "local-") {
		logging.Info().
			Str("id", child.ID()).
			Msg("Replacing cloud-synced .xdg-volume-info with local virtual file")
		if err := graph.Remove(child.ID(), auth); err != nil {
			logging.Warn().Err(err).Str("id", child.ID()).Msg("Failed to delete remote .xdg-volume-info; continuing with local replacement")
		} else {
			logging.Info().Str("id", child.ID()).Msg("Removed remote .xdg-volume-info copy")
		}
		filesystem.DeleteID(child.ID())
		child = nil
	}

	user, err := graph.GetUser(auth)
	if err != nil {
		logging.Error().Err(err).Msg("Could not create .xdg-volume-info")
		return
	}
	content := []byte(TemplateXDGVolumeInfo(user.UserPrincipalName))
	now := time.Now()

	if child != nil {
		child.SetMode(fuse.S_IFREG | 0644)
		child.DriveItem.Size = uint64(len(content))
		child.DriveItem.ModTime = &now
		child.SetVirtualContent(content)
		filesystem.RegisterVirtualFile(child)
		logging.Debug().
			Str("id", child.ID()).
			Msg("Refreshed local .xdg-volume-info content")
		return
	}

	logging.Info().Msg("Creating .xdg-volume-info as local-only virtual file")
	root, _ := filesystem.GetPath("/", auth) // cannot fail
	inode := fs.NewInode(fileName, fuse.S_IFREG|0644, root)
	inode.DriveItem.Size = uint64(len(content))
	inode.DriveItem.ModTime = &now
	inode.SetVirtualContent(content)
	filesystem.RegisterVirtualFile(inode)

	logging.Debug().
		Str("id", inode.ID()).
		Str("name", inode.Name()).
		Msg("Created local-only .xdg-volume-info file")

}

// IsUserAllowOtherEnabled checks if the 'user_allow_other' option is enabled in /etc/fuse.conf
func IsUserAllowOtherEnabled() bool {
	// Try to open /etc/fuse.conf
	file, err := os.Open("/etc/fuse.conf")
	if err != nil {
		logging.Debug().Err(err).Msg("Could not open /etc/fuse.conf, assuming user_allow_other is not enabled")
		return false
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			logging.Error().Err(err).Msg("Error closing /etc/fuse.conf")
		}
	}(file)

	// Scan the file line by line
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		// Remove comments and trim spaces
		if idx := strings.Index(line, "#"); idx >= 0 {
			line = line[:idx]
		}
		line = strings.TrimSpace(line)

		// Check if the line contains user_allow_other
		if line == "user_allow_other" {
			logging.Debug().Msg("Found user_allow_other in /etc/fuse.conf")
			return true
		}
	}

	if err := scanner.Err(); err != nil {
		logging.Debug().Err(err).Msg("Error reading /etc/fuse.conf, assuming user_allow_other is not enabled")
	}

	logging.Debug().Msg("user_allow_other not found in /etc/fuse.conf")
	return false
}
