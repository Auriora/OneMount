package systemd

import (
	"os"
	"testing"
	"time"

	"github.com/coreos/go-systemd/v22/unit"
	"github.com/godbus/dbus/v5"
	"github.com/jstaf/onedriver/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Does systemd unit name templating work correctly?
func TestTemplateUnit(t *testing.T) {
	t.Parallel()
	escaped := TemplateUnit(OnedriverServiceTemplate, "this-is-a-test")
	require.Equal(t, "onedriver@this-is-a-test.service", escaped, "Templating did not work.")
}

// TestUntemplateUnit tests that systemd unit untemplating works correctly
func TestUntemplateUnit(t *testing.T) {
	t.Parallel()

	// Define test cases
	testCases := []struct {
		name           string
		input          string
		expectedOutput string
		expectError    bool
		errorMessage   string
	}{
		{
			name:         "InvalidUnitName_ShouldReturnError",
			input:        "this-wont-work",
			expectError:  true,
			errorMessage: "Untemplating \"this-wont-work\" shouldn't have worked",
		},
		{
			name:           "ValidUnitNameWithoutSuffix_ShouldUntemplate",
			input:          "onedriver@home-some-path",
			expectedOutput: "home-some-path",
			expectError:    false,
			errorMessage:   "Failed to untemplate unit",
		},
		{
			name:           "ValidUnitNameWithSuffix_ShouldUntemplate",
			input:          "onedriver@opt-other.service",
			expectedOutput: "opt-other",
			expectError:    false,
			errorMessage:   "Failed to untemplate unit",
		},
	}

	// Run each test case
	for _, tc := range testCases {
		tc := tc // Capture range variable for parallel execution
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Call the function being tested
			output, err := UntemplateUnit(tc.input)

			// Check if error behavior matches expectations
			if tc.expectError {
				assert.Error(t, err, tc.errorMessage)
			} else {
				assert.NoError(t, err, tc.errorMessage)
				assert.Equal(t, tc.expectedOutput, output, "Did not untemplate systemd unit correctly")
			}
		})
	}
}

// can we enable and disable systemd units? (and correctly check if the units are
// enabled/disabled?)
func TestUnitEnabled(t *testing.T) {
	t.Parallel()
	testDir, _ := os.Getwd()
	unitName := TemplateUnit(OnedriverServiceTemplate, unit.UnitNamePathEscape(testDir+"/mount"))

	// make sure everything is disabled before we start
	require.NoError(t, UnitSetEnabled(unitName, false))
	enabled, err := UnitIsEnabled(unitName)
	require.NoError(t, err)
	require.False(t, enabled, "Unit was enabled before test started and we couldn't disable it!")

	// actual test content
	require.NoError(t, UnitSetEnabled(unitName, true))
	enabled, err = UnitIsEnabled(unitName)
	require.NoError(t, err)
	require.True(t, enabled, "Could not detect unit as enabled.")

	require.NoError(t, UnitSetEnabled(unitName, false))
	enabled, err = UnitIsEnabled(unitName)
	require.NoError(t, err)
	require.False(t, enabled, "Unit was still enabled after disabling it.")
}

func TestUnitActive(t *testing.T) {
	t.Parallel()
	testDir, _ := os.Getwd()
	unitName := TemplateUnit(OnedriverServiceTemplate, unit.UnitNamePathEscape(testDir+"/mount"))

	// Check if the unit exists before proceeding
	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		t.Skip("Could not connect to session bus:", err)
	}
	t.Cleanup(func() {
		if err := conn.Close(); err != nil {
			t.Logf("Warning: Failed to close dbus connection: %v", err)
		}
	})

	obj := conn.Object(SystemdBusName, SystemdObjectPath)
	call := obj.Call("org.freedesktop.systemd1.Manager.GetUnit", 0, unitName)
	if call.Err != nil {
		// Unit doesn't exist, skip the test
		t.Skipf("Unit %s not found, skipping test", unitName)
	}

	// make extra sure things are off before we start
	require.NoError(t, UnitSetActive(unitName, false))
	active, err := UnitIsActive(unitName)
	require.NoError(t, err)
	require.False(t, active, "Unit was active before job start and we could not stop it!")

	require.NoError(t, UnitSetActive(unitName, true), "Failed to start unit.")

	// Use WaitForCondition to wait for the unit to become active
	// This replaces the fixed timeout with dynamic waiting
	var isActive bool
	testutil.WaitForCondition(t, func() bool {
		var err error
		isActive, err = UnitIsActive(unitName)
		return err == nil && isActive
	}, 5*time.Second, 500*time.Millisecond, "Unit did not become active within timeout")

	require.True(t, isActive, "Could not detect unit as active following start.")

	require.NoError(t, UnitSetActive(unitName, false), "Failed to stop unit.")
	active, err = UnitIsActive(unitName)
	require.NoError(t, err, "Failed to check unit active state.")
	require.False(t, active, "Did not detect unit as stopped.")
}
