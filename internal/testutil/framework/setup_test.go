package framework

import (
	"fmt"
	"github.com/auriora/onemount/internal/testutil"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	// Setup code
	fmt.Println("Setting up tests...")

	// Create the parent directory first
	if err := os.MkdirAll(testutil.TestSandboxDir, 0755); err != nil {
		fmt.Printf("Error: Failed to create test sandbox directory: %v\n", err)
		os.Exit(1)
	}

	// Then create the tmp directory
	if err := os.MkdirAll(testutil.TestSandboxTmpDir, 0755); err != nil {
		fmt.Printf("Error: Failed to create output directory: %v\n", err)
		os.Exit(1)
	}
	// Run tests
	code := m.Run()

	// Teardown code
	fmt.Println("Tearing down tests...")

	// Exit with the result of m.Run()
	os.Exit(code)
}
