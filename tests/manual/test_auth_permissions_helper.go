//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/auriora/onemount/internal/graph"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <test_home>\n", os.Args[0])
		os.Exit(1)
	}

	testHome := os.Args[1]
	authFile := filepath.Join(testHome, "test_auth", "auth_tokens.json")

	// Create directory first with 0700 permissions (as the code does)
	if err := os.MkdirAll(filepath.Dir(authFile), 0700); err != nil {
		fmt.Fprintf(os.Stderr, "Could not create auth directory: %v\n", err)
		os.Exit(1)
	}

	auth := &graph.Auth{
		Account:      "test@example.com",
		ExpiresAt:    9999999999,
		AccessToken:  "test-access-token",
		RefreshToken: "test-refresh-token",
	}

	err := graph.SaveAuthTokens(auth, authFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error saving auth tokens: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Auth tokens saved successfully")
}
