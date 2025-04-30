package common

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bcherrington/onedriver/internal/testutil"
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

	if err := os.RemoveAll(testutil.TestSandboxTmpDir); err != nil {
		log.Error().Err(err).Msg("Failed to remove tmp directory")
		os.Exit(1)
	}

	// Ensure tmp directory exists
	if err := os.MkdirAll(testutil.TestSandboxTmpDir, 0755); err != nil {
		log.Error().Err(err).Msg("Failed to create tmp directory")
		os.Exit(1)
	}

	f, err := os.OpenFile(testutil.TestLogPath, os.O_TRUNC|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		log.Error().Err(err).Msg("Failed to open log file")
		os.Exit(1)
	}
	zerolog.SetGlobalLevel(zerolog.TraceLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: f, TimeFormat: "15:04:05"})
	defer f.Close()

	os.Exit(m.Run())
}
