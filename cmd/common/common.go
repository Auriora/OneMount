// common functions used by both binaries
package common

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/jstaf/onedriver/fs"
	"github.com/jstaf/onedriver/fs/graph"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const version = "0.14.1"

var commit string

// Version returns the current version string
func Version() string {
	clen := 0
	if len(commit) > 7 {
		clen = 8
	}
	return fmt.Sprintf("v%s %s", version, commit[:clen])
}

// StringToLevel converts a string to a zerolog.LogLevel that can be used with zerolog
func StringToLevel(input string) zerolog.Level {
	level, err := zerolog.ParseLevel(input)
	if err != nil {
		log.Error().Err(err).Msg("Could not parse log level, defaulting to \"debug\"")
		return zerolog.DebugLevel
	}
	return level
}

// LogLevels returns the available logging levels
func LogLevels() []string {
	return []string{"trace", "debug", "info", "warn", "error", "fatal"}
}

// TemplateXDGVolumeInfo returns a formatted .xdg-volume-info file content
func TemplateXDGVolumeInfo(name string) string {
	xdgVolumeInfo := fmt.Sprintf("[Volume Info]\nName=%s\nIcon=dk-onedrive\n", name)
	if _, err := os.Stat("/usr/share/icons/onedriver/onedriver.png"); err == nil {
		xdgVolumeInfo += "IconFile=/usr/share/icons/onedriver.png\n"
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
// corner of the mountpoint and shows the account name in the nautilus sidebar
func CreateXDGVolumeInfo(filesystem *fs.Filesystem, auth *graph.Auth) {
	if child, _ := filesystem.GetPath("/.xdg-volume-info", auth); child != nil {
		return
	}
	log.Info().Msg("Creating .xdg-volume-info")
	user, err := graph.GetUser(auth)
	if err != nil {
		log.Error().Err(err).Msg("Could not create .xdg-volume-info")
		return
	}
	xdgVolumeInfo := TemplateXDGVolumeInfo(user.UserPrincipalName)

	// just upload directly and shove it in the cache
	// (since the fs isn't mounted yet)
	resp, err := graph.Put(
		graph.ResourcePath("/.xdg-volume-info")+":/content",
		auth,
		strings.NewReader(xdgVolumeInfo),
	)
	if err != nil {
		log.Error().Err(err).Msg("Failed to write .xdg-volume-info")
	}
	root, _ := filesystem.GetPath("/", auth) // cannot fail
	inode := fs.NewInode(".xdg-volume-info", 0644, root)
	if json.Unmarshal(resp, &inode) == nil {
		filesystem.InsertID(inode.ID(), inode)
	}
}

// IsUserAllowOtherEnabled checks if the 'user_allow_other' option is enabled in /etc/fuse.conf
func IsUserAllowOtherEnabled() bool {
	// Try to open /etc/fuse.conf
	file, err := os.Open("/etc/fuse.conf")
	if err != nil {
		log.Debug().Err(err).Msg("Could not open /etc/fuse.conf, assuming user_allow_other is not enabled")
		return false
	}
	defer file.Close()

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
			log.Debug().Msg("Found user_allow_other in /etc/fuse.conf")
			return true
		}
	}

	if err := scanner.Err(); err != nil {
		log.Debug().Err(err).Msg("Error reading /etc/fuse.conf, assuming user_allow_other is not enabled")
	}

	log.Debug().Msg("user_allow_other not found in /etc/fuse.conf")
	return false
}
