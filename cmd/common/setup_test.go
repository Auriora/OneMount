package common

import (
	"os"
	"testing"

	"github.com/bcherrington/onemount/internal/testutil"
	"github.com/rs/zerolog/log"
)

func TestMain(m *testing.M) {
	// Setup test environment
	f, err := testutil.SetupTestEnvironment("../..", false)
	if err != nil {
		log.Error().Err(err).Msg("Failed to setup test environment")
		os.Exit(1)
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Error().Err(err).Msg("Failed to close log file")
		}
	}()

	os.Exit(m.Run())
}
