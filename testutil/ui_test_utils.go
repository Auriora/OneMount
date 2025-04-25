package testutil

import (
	"os"
	"path/filepath"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// SetupUITest performs common setup tasks for UI tests:
// - Changes to the project root directory
// - Sets up logging
// - Ensures the mount directory exists and is clean
// The relPath parameter should be the relative path from the test package to the project root.
func SetupUITest(relPath string) (*os.File, error) {
	// Change to the project root directory
	if err := os.Chdir(relPath); err != nil {
		log.Error().Err(err).Msg("Failed to change directory")
		return nil, err
	}

	// Setup logging
	f, err := os.OpenFile("fusefs_tests.log", os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		log.Error().Err(err).Msg("Failed to open log file")
		return nil, err
	}
	zerolog.SetGlobalLevel(zerolog.TraceLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: f, TimeFormat: "15:04:05"})

	// Remove the mount directory if it exists and is a stale mount point
	if _, err := os.Stat("mount"); err == nil {
		if _, err := os.ReadDir("mount"); err != nil {
			// If we can't read the directory, it might be a stale mount point
			log.Warn().Err(err).Msg("Mount directory exists but can't be read, attempting to remove")
			if err := os.Remove("mount"); err != nil {
				log.Error().Err(err).Msg("Failed to remove mount directory")
				f.Close()
				return nil, err
			}
		}
	}

	// Create a fresh mount directory
	if err := os.Mkdir("mount", 0700); err != nil && !os.IsExist(err) {
		log.Error().Err(err).Msg("Failed to create mount directory")
		f.Close()
		return nil, err
	}

	return f, nil
}

// EnsureMountPoint checks if the mount point exists and is accessible.
// If it doesn't exist, it creates it. If it exists but is not accessible,
// it attempts to remove it and create a new one.
func EnsureMountPoint(mountPath string) error {
	// Ensure the mount path is absolute
	mountPath, err := filepath.Abs(mountPath)
	if err != nil {
		log.Error().Err(err).Str("path", mountPath).Msg("Failed to get absolute path for mount point")
		return err
	}

	// Check if the mount point exists
	if _, err := os.Stat(mountPath); err != nil {
		if os.IsNotExist(err) {
			// Create the mount directory if it doesn't exist
			if err := os.MkdirAll(mountPath, 0700); err != nil {
				log.Error().Err(err).Str("path", mountPath).Msg("Failed to create mount directory")
				return err
			}
			log.Info().Str("path", mountPath).Msg("Created mount directory")
			return nil
		}
		log.Error().Err(err).Str("path", mountPath).Msg("Failed to check mount directory")
		return err
	}

	// Check if the mount point is accessible
	if _, err := os.ReadDir(mountPath); err != nil {
		// If we can't read the directory, it might be a stale mount point
		log.Warn().Err(err).Str("path", mountPath).Msg("Mount directory exists but can't be read, attempting to remove")
		if err := os.Remove(mountPath); err != nil {
			log.Error().Err(err).Str("path", mountPath).Msg("Failed to remove mount directory")
			return err
		}
		// Create a fresh mount directory
		if err := os.MkdirAll(mountPath, 0700); err != nil {
			log.Error().Err(err).Str("path", mountPath).Msg("Failed to create mount directory")
			return err
		}
		log.Info().Str("path", mountPath).Msg("Recreated mount directory")
	}

	return nil
}