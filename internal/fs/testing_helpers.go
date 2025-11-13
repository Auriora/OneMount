package fs

import (
	"os"
	"path/filepath"
	"time"

	"github.com/auriora/onemount/internal/logging"
)

// ConfigureTestLogging sets up logging for tests based on environment variables.
// If ONEMOUNT_LOG_TO_FILE is true, logs are written to ONEMOUNT_LOG_DIR.
// Otherwise, logs go to stdout (default behavior).
func ConfigureTestLogging() {
	logToFile := os.Getenv("ONEMOUNT_LOG_TO_FILE")
	if logToFile != "true" {
		// Default behavior: log to stdout
		return
	}

	logDir := os.Getenv("ONEMOUNT_LOG_DIR")
	if logDir == "" {
		logDir = filepath.Join(os.Getenv("HOME"), ".onemount-tests", "logs")
	}

	// Create log directory if it doesn't exist
	if err := os.MkdirAll(logDir, 0755); err != nil {
		// If we can't create the log directory, fall back to stdout
		logging.Warn().Err(err).Str("logDir", logDir).Msg("Failed to create log directory, using stdout")
		return
	}

	// Create log file with timestamp
	timestamp := time.Now().Format("20060102-150405")
	logFile := filepath.Join(logDir, "test-"+timestamp+".log")

	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		logging.Warn().Err(err).Str("logFile", logFile).Msg("Failed to open log file, using stdout")
		return
	}

	// Configure the default logger to write to the file
	logging.DefaultLogger = logging.New(file)

	// Print to stdout where logs are being written
	_, _ = os.Stdout.WriteString("Test logs redirected to: " + logFile + "\n")
}
