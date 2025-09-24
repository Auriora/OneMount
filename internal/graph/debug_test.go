package graph

import (
	"fmt"
	"os"
	"testing"
)

// TestDebug is a simple test to help diagnose test setup issues
func TestDebug(t *testing.T) {
	fmt.Println("=== DEBUG TEST STARTED ===")

	// Print environment information
	fmt.Println("Current working directory:")
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error getting current directory: %v\n", err)
	} else {
		fmt.Println(cwd)
	}

	// Print HOME environment variable
	fmt.Printf("HOME environment variable: %s\n", os.Getenv("HOME"))

	// Print user information
	fmt.Printf("User ID: %d\n", os.Getuid())
	fmt.Printf("Group ID: %d\n", os.Getgid())

	// Try to create a test directory in the current directory
	testDir := "test_debug_dir"
	err = os.MkdirAll(testDir, 0755)
	if err != nil {
		fmt.Printf("Error creating test directory: %v\n", err)
	} else {
		fmt.Printf("Successfully created test directory: %s\n", testDir)
		// Clean up
		err = os.RemoveAll(testDir)
		if err != nil {
			fmt.Printf("Error removing test directory: %v\n", err)
		} else {
			fmt.Printf("Successfully removed test directory: %s\n", testDir)
		}
	}

	// Try to create a test directory in the home directory
	homeTestDir := os.Getenv("HOME") + "/.onemount-tests-debug"
	err = os.MkdirAll(homeTestDir, 0755)
	if err != nil {
		fmt.Printf("Error creating test directory in home: %v\n", err)
	} else {
		fmt.Printf("Successfully created test directory in home: %s\n", homeTestDir)
		// Clean up
		err = os.RemoveAll(homeTestDir)
		if err != nil {
			fmt.Printf("Error removing test directory in home: %v\n", err)
		} else {
			fmt.Printf("Successfully removed test directory in home: %s\n", homeTestDir)
		}
	}

	fmt.Println("=== DEBUG TEST COMPLETED ===")
}
