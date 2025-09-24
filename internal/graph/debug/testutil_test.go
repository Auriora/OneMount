package debug

import (
	"fmt"
	"github.com/auriora/onemount/internal/testutil"
	"os"
	"path/filepath"
	"testing"
)

// TestTestUtilPaths tests that we can access the test directory paths from the testutil package
func TestTestUtilPaths(t *testing.T) {
	fmt.Println("=== TESTUTIL PATHS TEST STARTED ===")

	// Print the test directory paths
	fmt.Printf("TestSandboxDir: %s\n", testutil.TestSandboxDir)
	fmt.Printf("TestSandboxTmpDir: %s\n", testutil.TestSandboxTmpDir)
	fmt.Printf("TestLogPath: %s\n", testutil.TestLogPath)
	fmt.Printf("GraphTestDir: %s\n", testutil.GraphTestDir)

	// Ensure test directories exist
	fmt.Println("Ensuring test directories exist...")
	if err := testutil.EnsureDirectoriesExist(); err != nil {
		fmt.Printf("ERROR: Failed to ensure test directories exist: %v\n", err)
		t.Fatalf("Failed to ensure test directories exist: %v", err)
	}
	fmt.Println("Test directories created successfully")

	// Check if the test directories exist
	fmt.Println("Checking if test directories exist...")
	if _, err := os.Stat(testutil.TestSandboxDir); os.IsNotExist(err) {
		t.Fatalf("TestSandboxDir does not exist: %s", testutil.TestSandboxDir)
	}
	if _, err := os.Stat(testutil.TestSandboxTmpDir); os.IsNotExist(err) {
		t.Fatalf("TestSandboxTmpDir does not exist: %s", testutil.TestSandboxTmpDir)
	}
	// Check if the logs directory exists
	logsDir := filepath.Dir(testutil.TestLogPath)
	if _, err := os.Stat(logsDir); os.IsNotExist(err) {
		t.Fatalf("Logs directory does not exist: %s", logsDir)
	}
	if _, err := os.Stat(testutil.GraphTestDir); os.IsNotExist(err) {
		t.Fatalf("GraphTestDir does not exist: %s", testutil.GraphTestDir)
	}
	fmt.Println("All test directories exist")

	fmt.Println("=== TESTUTIL PATHS TEST COMPLETED ===")
}
