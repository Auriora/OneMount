// Package testutil provides utility functions and constants for testing.
package testutil

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// SetupTestEnvironment performs common setup tasks for tests:
// - Changes to the project root directory
// - Sets up logging
// - Ensures test directories exist and are clean
// - Optionally unmounts and recreates the mount point
//
// The relPath parameter should be the relative path from the test package to the project root.
// If unmountFirst is true, it will attempt to unmount the mount point before recreating it.
// Returns a file handle to the log file, which should be closed by the caller.
func SetupTestEnvironment(relPath string, unmountFirst bool) (*os.File, error) {
	// Change to the project root directory
	if err := changeToProjectRoot(relPath); err != nil {
		return nil, err
	}

	// Setup logging
	f, err := setupLogging()
	if err != nil {
		return nil, err
	}

	// Ensure test directories exist
	if err := ensureTestDirectories(unmountFirst); err != nil {
		err := f.Close()
		if err != nil {
			return nil, err
		}
		return nil, err
	}

	return f, nil
}

// changeToProjectRoot changes the current working directory to the project root.
// It uses the provided relative path to navigate to the project root.
func changeToProjectRoot(relPath string) error {
	// If relPath is provided, try to change to that directory
	if relPath != "" {
		if err := os.Chdir(relPath); err != nil {
			log.Error().Err(err).Str("path", relPath).Msg("Failed to change to project root directory using relative path")
			return err
		}

		// Log the current working directory for debugging
		cwd, err := os.Getwd()
		if err != nil {
			log.Error().Err(err).Msg("Failed to get current working directory")
		} else {
			log.Debug().Str("cwd", cwd).Msg("Changed to project root directory")
		}
	}

	return nil
}

// setupLogging sets up logging for tests.
func setupLogging() (*os.File, error) {
	// Ensure test-sandbox directory exists
	if err := os.MkdirAll(TestSandboxDir, 0755); err != nil {
		log.Error().Err(err).Msg("Failed to create test-sandbox directory")
		return nil, err
	}

	f, err := os.OpenFile(TestLogPath, os.O_TRUNC|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		log.Error().Err(err).Msg("Failed to open log file")
		return nil, err
	}
	zerolog.SetGlobalLevel(zerolog.TraceLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: f, TimeFormat: "15:04:05"})
	return f, nil
}

// ensureTestDirectories ensures that all test directories exist and are clean.
// If unmountFirst is true, it will attempt to unmount the mount point before recreating it.
func ensureTestDirectories(unmountFirst bool) error {
	// Remove and recreate the test sandbox tmp directory
	if err := os.RemoveAll(TestSandboxTmpDir); err != nil {
		log.Error().Err(err).Msg("Failed to remove test sandbox tmp directory")
		return err
	}
	if err := os.MkdirAll(TestSandboxTmpDir, 0755); err != nil {
		log.Error().Err(err).Msg("Failed to create test sandbox tmp directory")
		return err
	}

	// Check if the mount point is already in use
	if unmountFirst {
		isMounted := false
		if _, err := os.Stat(TestMountPoint); err == nil {
			// Check if it's a mount point by trying to read from it
			if _, err := os.ReadDir(TestMountPoint); err != nil {
				// If we can't read the directory, it might be a stale mount point
				log.Warn().Err(err).Msg("Mount point exists but can't be read, attempting to unmount")
				isMounted = true
			} else {
				// Check if it's a mount point using findmnt
				cmd := exec.Command("findmnt", "--noheadings", "--output", "TARGET", TestMountPoint)
				if output, err := cmd.Output(); err == nil && len(output) > 0 {
					log.Warn().Msg("Mount point is already mounted, attempting to unmount")
					isMounted = true
				}
			}
		}

		// Attempt to unmount if necessary
		if isMounted {
			log.Info().Msg("Attempting to unmount previous instance")
			// Try normal unmount first
			if unmountErr := exec.Command("fusermount3", "-u", TestMountPoint).Run(); unmountErr != nil {
				log.Warn().Err(unmountErr).Msg("Normal unmount failed, trying lazy unmount")
				// Try lazy unmount
				if lazyErr := exec.Command("fusermount3", "-uz", TestMountPoint).Run(); lazyErr != nil {
					log.Error().Err(lazyErr).Msg("Lazy unmount also failed, mount point may be in use by another process")
					// Continue anyway, but warn the user
					log.Warn().Msg("Failed to unmount existing filesystem. Tests may fail if mount point is in use")
				} else {
					log.Info().Msg("Successfully performed lazy unmount")
				}
			} else {
				log.Info().Msg("Successfully unmounted previous instance")
			}
		}
	}

	// Create the mount directory
	if err := os.MkdirAll(TestMountPoint, 0755); err != nil {
		log.Error().Err(err).Msg("Failed to create mount directory")
		return err
	}

	// Create content directory structure for tests
	contentDir := filepath.Join(TestSandboxTmpDir, "test", "content")
	if err := os.MkdirAll(contentDir, 0755); err != nil {
		log.Error().Err(err).Msg("Failed to create content directory")
		return err
	}

	// Create thumbnails directory structure for tests
	thumbnailsDir := filepath.Join(TestSandboxTmpDir, "test", "thumbnails")
	if err := os.MkdirAll(thumbnailsDir, 0755); err != nil {
		log.Error().Err(err).Msg("Failed to create thumbnails directory")
		return err
	}

	return nil
}
