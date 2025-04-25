package ui

import (
	"os"
	"testing"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func TestMain(m *testing.M) {
	if err := os.Chdir("../"); err != nil {
		log.Error().Err(err).Msg("Failed to change directory")
		os.Exit(1)
	}

	// Setup logging
	f, err := os.OpenFile("fusefs_tests.log", os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		log.Error().Err(err).Msg("Failed to open log file")
		os.Exit(1)
	}
	zerolog.SetGlobalLevel(zerolog.TraceLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: f, TimeFormat: "15:04:05"})
	defer f.Close()

	// Remove the mount directory if it exists and is a stale mount point
	if _, err := os.Stat("mount"); err == nil {
		if _, err := os.ReadDir("mount"); err != nil {
			// If we can't read the directory, it might be a stale mount point
			log.Warn().Err(err).Msg("Mount directory exists but can't be read, attempting to remove")
			if err := os.Remove("mount"); err != nil {
				log.Error().Err(err).Msg("Failed to remove mount directory")
				os.Exit(1)
			}
		}
	}

	// Create a fresh mount directory
	if err := os.Mkdir("mount", 0700); err != nil {
		log.Error().Err(err).Msg("Failed to create mount directory")
		os.Exit(1)
	}

	os.Exit(m.Run())
}
