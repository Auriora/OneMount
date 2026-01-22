package fs

import (
	"os"
	"testing"
)

// TestMain is the entry point for all tests in this package.
// It configures logging before running tests.
func TestMain(m *testing.M) {
	// Configure test logging based on environment variables
	ConfigureTestLogging()

	// Run tests
	code := m.Run()

	// Exit with the test result code
	os.Exit(code)
}
