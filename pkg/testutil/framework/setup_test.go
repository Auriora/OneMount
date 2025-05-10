package framework

import (
	"fmt"
	"github.com/auriora/onemount/pkg/testutil"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	// Setup code
	fmt.Println("Setting up tests...")

	if err := os.MkdirAll(testutil.TestSandboxTmpDir, 0755); err != nil {
		fmt.Printf("Warning: Failed to create output directory: %v\n", err)
	}
	// Run tests
	code := m.Run()

	// Teardown code
	fmt.Println("Tearing down tests...")

	// Exit with the result of m.Run()
	os.Exit(code)
}
