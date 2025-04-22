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
		log.Fatal().Err(err).Msg("Authentication failed")
	}
	user, _ := GetUser(auth)
	drive, _ := GetDrive(auth)
	log.Info().
		Str("account", user.UserPrincipalName).
		Str("type", drive.DriveType).
		Msg("Starting tests")

	os.Exit(m.Run())
}
