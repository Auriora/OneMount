package common

import (
	"github.com/auriora/onemount/internal/testutil"
	"github.com/auriora/onemount/internal/testutil/helpers"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
	"testing"
)

// TestMain is a special function recognized by the Go testing package.
// It's called before any tests in the package are run and is responsible for
// setting up the test environment and cleaning up after all tests have completed.
func TestMain(m *testing.M) {
	// Ensure test directories exist
	if err := helpers.EnsureTestDirectories(); err != nil {
		log.Error().Err(err).Msg("Failed to ensure test directories exist")
		os.Exit(1)
	}

	// Set up logging
	logFile, err := os.OpenFile(testutil.TestLogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Error().Err(err).Msg("Failed to open log file")
		os.Exit(1)
	}
	defer func() {
		if err := logFile.Close(); err != nil {
			log.Error().Err(err).Msg("Failed to close log file")
		}
	}()

	// Configure zerolog to write to the log file
	log.Logger = zerolog.New(logFile).With().Timestamp().Logger()

	// Run the tests and exit with the appropriate status code
	os.Exit(m.Run())
}
