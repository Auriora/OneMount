package common

import (
	"os"
	"testing"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func TestMain(m *testing.M) {
	if err := os.Chdir("../.."); err != nil {
		log.Error().Err(err).Msg("Failed to change directory")
		os.Exit(1)
	}

	if err := os.RemoveAll("tmp"); err != nil {
		log.Error().Err(err).Msg("Failed to remove tmp directory")
		os.Exit(1)
	}

	f, err := os.OpenFile("fusefs_tests.log", os.O_TRUNC|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		log.Error().Err(err).Msg("Failed to open log file")
		os.Exit(1)
	}
	zerolog.SetGlobalLevel(zerolog.TraceLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: f, TimeFormat: "15:04:05"})
	defer f.Close()

	os.Exit(m.Run())
}
