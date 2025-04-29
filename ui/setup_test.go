package ui

import (
	"os"
	"testing"

	"github.com/bcherrington/onedriver/testutil"
)

func TestMain(m *testing.M) {
	// Setup UI test environment
	f, err := testutil.SetupUITest("../")
	if err != nil {
		os.Exit(1)
	}
	defer f.Close()

	os.Exit(m.Run())
}
