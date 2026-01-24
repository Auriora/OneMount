package systemd

import (
	"testing"

	"github.com/auriora/onemount/internal/testutil/framework"
)

// TestUT_UI_04_01_SystemdUnit_Template_AppliesInstanceName tests the TemplateUnit function.
//
//	Test Case ID    UT-UI-04-01
//	Title           Systemd Unit Templating
//	Description     Tests the TemplateUnit function
//	Preconditions   None
//	Steps           1. Define test cases with different unit names and instance names
//	                2. Call TemplateUnit with each test case
//	                3. Check if the result matches the expected templated unit name
//	Expected Result Unit names are correctly templated with instance names
//	Notes: This test verifies that the TemplateUnit function correctly templates unit names with instance names.
func TestUT_UI_04_01_SystemdUnit_Template_AppliesInstanceName(t *testing.T) {
	// Create a test fixture
	fixture := framework.NewUnitTestFixture("SystemdUnitTemplateFixture")

	// Use the fixture to run the test
	fixture.Use(t, func(t *testing.T, fixture interface{}) {
		// Test cases for TemplateUnit
		testCases := []struct {
			template string
			instance string
			expected string
		}{
			{"onemount@.service", "home-user-OneDrive", "onemount@home-user-OneDrive.service"},
			{"onemount@.service", "mnt-onedrive", "onemount@mnt-onedrive.service"},
			{"onemount@.service", "/home/user/OneDrive", "onemount@-home-user-OneDrive.service"}, // Forward slashes replaced with hyphens
			{"myservice@.service", "instance1", "myservice@instance1.service"},
			{"test@.timer", "test-instance", "test@test-instance.timer"},
		}

		for _, tc := range testCases {
			result := TemplateUnit(tc.template, tc.instance)
			if result != tc.expected {
				t.Errorf("TemplateUnit(%q, %q) = %q, expected %q", tc.template, tc.instance, result, tc.expected)
			}
		}
	})
}

// TestUT_UI_05_01_SystemdUnit_Untemplate_ExtractsUnitAndInstanceName tests the UntemplateUnit function.
//
//	Test Case ID    UT-UI-05-01
//	Title           Systemd Unit Untemplating
//	Description     Tests the UntemplateUnit function
//	Preconditions   None
//	Steps           1. Define test cases with different templated unit names
//	                2. Call UntemplateUnit with each test case
//	                3. Check if the result matches the expected unit name and instance name
//	Expected Result Templated unit names are correctly untemplated into unit name and instance name
//	Notes: This test verifies that the UntemplateUnit function correctly untemplates unit names into unit name and instance name.
func TestUT_UI_05_01_SystemdUnit_Untemplate_ExtractsUnitAndInstanceName(t *testing.T) {
	// Create a test fixture
	fixture := framework.NewUnitTestFixture("SystemdUnitUntemplateFixture")

	// Use the fixture to run the test
	fixture.Use(t, func(t *testing.T, fixture interface{}) {
		// Test cases for UntemplateUnit
		testCases := []struct {
			input            string
			expectedInstance string
			expectError      bool
		}{
			{"onemount@home-user-OneDrive.service", "home-user-OneDrive", false},
			{"onemount@mnt-onedrive.service", "mnt-onedrive", false},
			{"onemount@-home-user-OneDrive.service", "-home-user-OneDrive", false},
			{"myservice@instance1.service", "instance1", false},
			{"test@test-instance.timer", "test-instance", false},
			{"onemount.service", "", true}, // Not a templated unit
			{"invalid", "", true},          // Invalid format
		}

		for _, tc := range testCases {
			result, err := UntemplateUnit(tc.input)
			if tc.expectError {
				if err == nil {
					t.Errorf("UntemplateUnit(%q) expected error, got nil", tc.input)
				}
			} else {
				if err != nil {
					t.Errorf("UntemplateUnit(%q) unexpected error: %v", tc.input, err)
				}
				if result != tc.expectedInstance {
					t.Errorf("UntemplateUnit(%q) = %q, expected %q", tc.input, result, tc.expectedInstance)
				}
			}
		}
	})
}
