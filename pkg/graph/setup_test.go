package graph

import (
	"fmt"
	"github.com/auriora/onemount/pkg/logging"
	"github.com/auriora/onemount/pkg/testutil"
	"os"
	"testing"
)

// TestMain is a special function recognized by the Go testing package.
// It's called before any tests in the package are run and is responsible for
// setting up the test environment and cleaning up after all tests have completed.
func TestMain(m *testing.M) {
	fmt.Println("Setting up graph package tests...")

	// Print the test directory paths for debugging
	fmt.Printf("TestSandboxDir: %s\n", testutil.TestSandboxDir)
	fmt.Printf("TestSandboxTmpDir: %s\n", testutil.TestSandboxTmpDir)
	fmt.Printf("TestLogPath: %s\n", testutil.TestLogPath)
	fmt.Printf("GraphTestDir: %s\n", testutil.GraphTestDir)

	// Check if HOME environment variable is set
	home := os.Getenv("HOME")
	fmt.Printf("HOME environment variable: %s\n", home)

	// Ensure test directories exist
	fmt.Println("Ensuring test directories exist...")
	if err := testutil.EnsureDirectoriesExist(); err != nil {
		fmt.Printf("ERROR: Failed to ensure test directories exist: %v\n", err)
		logging.Error().Err(err).Msg("Failed to ensure test directories exist")
		os.Exit(1)
	}
	fmt.Println("Test directories created successfully")

	// Set up logging
	fmt.Printf("Setting up logging to: %s\n", testutil.TestLogPath)
	logFile, err := os.OpenFile(testutil.TestLogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Printf("ERROR: Failed to open log file: %v\n", err)
		logging.Error().Err(err).Msg("Failed to open log file")
		os.Exit(1)
	}
	fmt.Println("Log file opened successfully")

	defer func() {
		if err := logFile.Close(); err != nil {
			fmt.Printf("ERROR: Failed to close log file: %v\n", err)
			logging.Error().Err(err).Msg("Failed to close log file")
		}
	}()

	// Configure logging to write to the log file
	logging.DefaultLogger = logging.New(logging.NewConsoleWriterWithOptions(logFile, "15:04:05"))
	fmt.Println("Logging configured successfully")

	// Run the tests and exit with the appropriate status code
	fmt.Println("Running graph package tests...")
	os.Exit(m.Run())
}
