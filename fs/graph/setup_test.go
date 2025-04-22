package graph

import (
	"os"
	"testing"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func TestMain(m *testing.M) {
	os.Chdir("../..")
	f, _ := os.OpenFile("fusefs_tests.log", os.O_TRUNC|os.O_CREATE|os.O_RDWR, 0644)
	zerolog.SetGlobalLevel(zerolog.TraceLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: f, TimeFormat: "15:04:05"})
	defer f.Close()

	// auth and log account metadata so we're extra sure who we're testing against
	auth, err := Authenticate(AuthConfig{}, ".auth_tokens.json", false)
	if err != nil {
		log.Error().Err(err).Msg("Authentication failed")
		os.Exit(1)
	}
	user, userErr := GetUser(auth)
	if userErr != nil {
		log.Warn().Err(userErr).Msg("Failed to get user information, continuing anyway")
	}

	drive, driveErr := GetDrive(auth)
	if driveErr != nil {
		log.Warn().Err(driveErr).Msg("Failed to get drive information, continuing anyway")
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

	os.Exit(m.Run())
}
