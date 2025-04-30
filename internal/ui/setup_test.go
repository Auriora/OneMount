package ui

import (
	"os"
	"testing"

	"github.com/bcherrington/onemount/internal/testutil"
)

func TestMain(m *testing.M) {
	// Setup UI test environment
	f, err := testutil.SetupUITest("../")
	if err != nil {
		os.Exit(1)
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			panic(err)
		}
	}(f)

	os.Exit(m.Run())
}
