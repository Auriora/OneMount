package graph

import (
	"github.com/bcherrington/onedriver/internal/testutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// captureFileSystemState captures the current state of the filesystem
// by listing all files and directories in the specified directory
func captureFileSystemState(dir string) (map[string]os.FileInfo, error) {
	state := make(map[string]os.FileInfo)

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Skip the directory itself
		if path == dir {
			return nil
		}
		// Store the file info in the state map
		state[path] = info
		return nil
	})

	return state, err
}

func TestMain(m *testing.M) {
	if err := os.Chdir("../.."); err != nil {
		log.Error().Err(err).Msg("Failed to change directory")
		os.Exit(1)
	}

	f, err := os.OpenFile(testutil.TestLogPath, os.O_TRUNC|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		log.Error().Err(err).Msg("Failed to open log file")
		os.Exit(1)
	}
	zerolog.SetGlobalLevel(zerolog.TraceLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: f, TimeFormat: "15:04:05"})
	defer func() {
		if err := f.Close(); err != nil {
			log.Error().Err(err).Msg("Failed to close log file")
		}
	}()

	// auth and log account metadata so we're extra sure who we're testing against
	var auth *Auth
	var user User
	var userErr error
	var drive Drive
	var driveErr error

	// Check if we should use mock authentication
	isMock := os.Getenv("ONEDRIVER_MOCK_AUTH") == "1"

	// Create authenticator based on configuration
	authenticator := NewAuthenticator(AuthConfig{}, testutil.AuthTokensPath, false, isMock)

	// Perform authentication
	var authErr error
	auth, authErr = authenticator.Authenticate()
	if authErr != nil {
		log.Error().Err(authErr).Msg("Authentication failed")
		os.Exit(1)
	}

	if isMock {
		log.Info().Msg("Using mock authentication for tests")

		// Create mock user and drive for consistent logging
		user = User{
			UserPrincipalName: "mock@example.com",
		}
		drive = Drive{
			ID:        "mock-drive-id",
			DriveType: "mock",
		}
	} else {
		// Get user and drive information
		user, userErr = GetUser(auth)
		if userErr != nil {
			log.Warn().Err(userErr).Msg("Failed to get user information, continuing anyway")
		}

		drive, driveErr = GetDrive(auth)
		if driveErr != nil {
			log.Warn().Err(driveErr).Msg("Failed to get drive information, continuing anyway")
		}
	}

	logEvent := log.Info()

	if userErr == nil {
		logEvent = logEvent.Str("account", user.UserPrincipalName)
	} else {
		logEvent = logEvent.Str("account", "unknown")
	}

	if driveErr == nil {
		logEvent = logEvent.Str("type", drive.DriveType)
	} else {
		logEvent = logEvent.Str("type", "unknown")
	}

	logEvent.Msg("Starting tests")

	// Create a test directory for capturing filesystem state under tmp/
	testDir := filepath.Join(testutil.TestSandboxTmpDir, "graph_test_dir")
	if err := os.MkdirAll(testDir, 0755); err != nil {
		log.Error().Err(err).Msg("Failed to create test directory")
		os.Exit(1)
	}

	// Ensure dmel.fa file exists for hash tests
	testutil.EnsureDmelfaExists()

	// Capture the initial state of the filesystem before running tests
	initialState, initialStateErr := captureFileSystemState(testDir)
	if initialStateErr != nil {
		log.Error().Err(initialStateErr).Msg("Failed to capture initial filesystem state")
	} else {
		log.Info().Int("files", len(initialState)).Msg("Captured initial filesystem state")
	}

	// Setup cleanup to run even if tests panic
	defer func() {
		log.Info().Msg("Running deferred cleanup...")

		// Capture the final state of the filesystem after tests
		if initialStateErr == nil {
			finalState, finalStateErr := captureFileSystemState(testDir)
			if finalStateErr != nil {
				log.Error().Err(finalStateErr).Msg("Failed to capture final filesystem state")
			} else {
				log.Info().Int("files", len(finalState)).Msg("Captured final filesystem state")

				// Check for files that exist in the final state but not in the initial state
				for path, info := range finalState {
					if _, exists := initialState[path]; !exists {
						log.Warn().Str("path", path).Bool("isDir", info.IsDir()).Msg("File created during tests but not cleaned up")

						// Attempt to clean up the file/directory
						if info.IsDir() {
							// Only remove empty directories to avoid accidentally deleting important content
							if entries, err := os.ReadDir(path); err == nil && len(entries) == 0 {
								if err := os.Remove(path); err != nil {
									log.Error().Err(err).Str("path", path).Msg("Failed to clean up directory")
								} else {
									log.Info().Str("path", path).Msg("Successfully cleaned up directory")
								}
							}
						} else {
							// Remove files
							if err := os.Remove(path); err != nil {
								log.Error().Err(err).Str("path", path).Msg("Failed to clean up file")
							} else {
								log.Info().Str("path", path).Msg("Successfully cleaned up file")
							}
						}
					}
				}
			}
		}

		// Clean up the test directory
		if err := os.RemoveAll(testDir); err != nil {
			log.Error().Err(err).Msg("Failed to remove test directory")
		} else {
			log.Info().Msg("Successfully removed test directory")
		}
	}()

	code := m.Run()
	log.Info().Msg("Tests completed")

	os.Exit(code)
}
