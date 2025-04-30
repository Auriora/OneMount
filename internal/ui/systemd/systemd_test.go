package systemd

import (
	"fmt"
	"testing"
	"time"

	"github.com/bcherrington/onemount/internal/testutil"
	"github.com/coreos/go-systemd/v22/unit"
	"github.com/godbus/dbus/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTemplateUnit tests that systemd unit name templating works correctly
func TestTemplateUnit(t *testing.T) {
	t.Parallel()

	// Define test cases
	testCases := []struct {
		name           string
		template       string
		input          string
		expectedOutput string
	}{
		{
			name:           "StandardTemplate_ShouldCreateCorrectUnitName",
			template:       OneMountServiceTemplate,
			input:          "this-is-a-test",
			expectedOutput: "onemount@this-is-a-test.service",
		},
		{
			name:           "PathWithSpecialChars_ShouldCreateCorrectUnitName",
			template:       OneMountServiceTemplate,
			input:          "path/with-special_chars",
			expectedOutput: "onemount@path-with-special_chars.service",
		},
	}

	// Run each test case
	for _, tc := range testCases {
		tc := tc // Capture range variable for parallel execution
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Call the function being tested
			result := TemplateUnit(tc.template, tc.input)

			// Verify the result
			require.Equal(t, tc.expectedOutput, result,
				"Templating did not work correctly for input: %s", tc.input)
		})
	}
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
			input:          "onemount@home-some-path",
			expectedOutput: "home-some-path",
			expectError:    false,
			errorMessage:   "Failed to untemplate unit",
		},
		{
			name:           "ValidUnitNameWithSuffix_ShouldUntemplate",
			input:          "onemount@opt-other.service",
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

// TestUnitEnabled tests that we can enable and disable systemd units
// and correctly check if the units are enabled/disabled
func TestUnitEnabled(t *testing.T) {
	t.Parallel()

	// Get the current directory and create a unit name for testing
	unitName := TemplateUnit(OneMountServiceTemplate, unit.UnitNamePathEscape(testutil.TestMountPoint))

	// Define test cases
	testCases := []struct {
		name          string
		setEnabled    bool
		expectedState bool
	}{
		{
			name:          "DisableUnit_ShouldBeDisabled",
			setEnabled:    false,
			expectedState: false,
		},
		{
			name:          "EnableUnit_ShouldBeEnabled",
			setEnabled:    true,
			expectedState: true,
		},
		{
			name:          "DisableAgain_ShouldBeDisabled",
			setEnabled:    false,
			expectedState: false,
		},
	}

	// Make sure everything is disabled before we start
	err := UnitSetEnabled(unitName, false)
	require.NoError(t, err, "Failed to disable unit before test")
	enabled, err := UnitIsEnabled(unitName)
	require.NoError(t, err, "Failed to check if unit is enabled")
	require.False(t, enabled, "Unit was enabled before test started and we couldn't disable it!")

	// Run each test case
	for _, tc := range testCases {
		tc := tc // Capture range variable
		t.Run(tc.name, func(t *testing.T) {
			// We don't use t.Parallel() here because we need to run these tests in sequence
			// to properly test the enable/disable functionality

			// Set the unit to the desired state
			err := UnitSetEnabled(unitName, tc.setEnabled)
			require.NoError(t, err, "Failed to set unit enabled state to %v", tc.setEnabled)

			// Check if the unit is in the expected state
			enabled, err := UnitIsEnabled(unitName)
			require.NoError(t, err, "Failed to check if unit is enabled")
			require.Equal(t, tc.expectedState, enabled,
				"Unit enabled state does not match expected state. Expected: %v, Got: %v",
				tc.expectedState, enabled)
		})
	}

	// Ensure cleanup: disable the unit after all tests
	t.Cleanup(func() {
		if err := UnitSetEnabled(unitName, false); err != nil {
			t.Logf("Warning: Failed to disable unit during cleanup: %v", err)
		}
	})
}

// TestUnitActive tests that we can start and stop systemd units
// and correctly check if the units are active/inactive
func TestUnitActive(t *testing.T) {
	t.Parallel()

	// Get the current directory and create a unit name for testing
	unitName := TemplateUnit(OneMountServiceTemplate, unit.UnitNamePathEscape(testutil.TestMountPoint))

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

	// Define test cases
	testCases := []struct {
		name          string
		setActive     bool
		expectedState bool
		waitForState  bool // Whether to wait for the state to change
	}{
		{
			name:          "StopUnit_ShouldBeInactive",
			setActive:     false,
			expectedState: false,
			waitForState:  false, // No need to wait for stopping as it's usually quick
		},
		{
			name:          "StartUnit_ShouldBeActive",
			setActive:     true,
			expectedState: true,
			waitForState:  true, // Need to wait for the unit to start
		},
		{
			name:          "StopAgain_ShouldBeInactive",
			setActive:     false,
			expectedState: false,
			waitForState:  false,
		},
	}

	// Make sure everything is stopped before we start
	err = UnitSetActive(unitName, false)
	require.NoError(t, err, "Failed to stop unit before test")
	active, err := UnitIsActive(unitName)
	require.NoError(t, err, "Failed to check if unit is active")
	require.False(t, active, "Unit was active before test started and we could not stop it!")

	// Run each test case
	for _, tc := range testCases {
		tc := tc // Capture range variable
		t.Run(tc.name, func(t *testing.T) {
			// We don't use t.Parallel() here because we need to run these tests in sequence
			// to properly test the start/stop functionality

			// Set the unit to the desired state
			err := UnitSetActive(unitName, tc.setActive)
			require.NoError(t, err, "Failed to set unit active state to %v", tc.setActive)

			// If we need to wait for the state to change (e.g., for starting the unit)
			if tc.waitForState {
				message := fmt.Sprintf("Unit did not reach expected state (%v) within timeout", tc.expectedState)
				testutil.WaitForCondition(t, func() bool {
					active, err := UnitIsActive(unitName)
					return err == nil && active == tc.expectedState
				}, 5*time.Second, 500*time.Millisecond, message)
			}

			// Check if the unit is in the expected state
			active, err := UnitIsActive(unitName)
			require.NoError(t, err, "Failed to check if unit is active")
			require.Equal(t, tc.expectedState, active,
				"Unit active state does not match expected state. Expected: %v, Got: %v",
				tc.expectedState, active)
		})
	}

	// Ensure cleanup: stop the unit after all tests
	t.Cleanup(func() {
		if err := UnitSetActive(unitName, false); err != nil {
			t.Logf("Warning: Failed to stop unit during cleanup: %v", err)
		}
	})
}
