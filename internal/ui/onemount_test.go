package ui

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/auriora/onemount/internal/graph"
	"github.com/auriora/onemount/internal/testutil/framework"
)

// TestUT_UI_01_01_Mountpoint_Validation_ReturnsCorrectResult tests that mountpoints are validated correctly.
//
//	Test Case ID    UT-UI-01-01
//	Title           Mountpoint Validation
//	Description     Tests that mountpoints are validated correctly
//	Preconditions   None
//	Steps           1. Create test directories and files
//	                2. Call MountpointIsValid with different paths
//	                3. Check if the result matches expectations
//	Expected Result Valid mountpoints return true, invalid mountpoints return false
//	Notes: This test verifies that the MountpointIsValid function correctly validates mountpoints.
func TestUT_UI_01_01_Mountpoint_Validation_ReturnsCorrectResult(t *testing.T) {
	// Create a test fixture
	fixture := framework.NewUnitTestFixture("MountpointValidationFixture")

	// Set up the fixture
	fixture.WithSetup(func(t *testing.T) (interface{}, error) {
		// Create a temporary directory for the test
		tempDir, err := os.MkdirTemp("", "onemount-test-*")
		if err != nil {
			return nil, err
		}

		// Create test directories and files
		validDir := filepath.Join(tempDir, "valid-dir")
		if err := os.Mkdir(validDir, 0755); err != nil {
			return nil, err
		}

		// Create a non-empty directory
		nonEmptyDir := filepath.Join(tempDir, "non-empty-dir")
		if err := os.Mkdir(nonEmptyDir, 0755); err != nil {
			return nil, err
		}
		if err := os.WriteFile(filepath.Join(nonEmptyDir, "file.txt"), []byte("test"), 0644); err != nil {
			return nil, err
		}

		// Create a file
		filePath := filepath.Join(tempDir, "file.txt")
		if err := os.WriteFile(filePath, []byte("test"), 0644); err != nil {
			return nil, err
		}

		return map[string]interface{}{
			"tempDir":     tempDir,
			"validDir":    validDir,
			"nonEmptyDir": nonEmptyDir,
			"filePath":    filePath,
		}, nil
	}).WithTeardown(func(t *testing.T, fixture interface{}) error {
		// Clean up the temporary directory
		data := fixture.(map[string]interface{})
		tempDir := data["tempDir"].(string)
		return os.RemoveAll(tempDir)
	})

	// Use the fixture to run the test
	fixture.Use(t, func(t *testing.T, f interface{}) {
		fixtureData := f.(*framework.UnitTestFixture)
		data := fixtureData.SetupData.(map[string]interface{})
		validDir := data["validDir"].(string)
		nonEmptyDir := data["nonEmptyDir"].(string)
		filePath := data["filePath"].(string)

		// Test valid empty directory
		if !MountpointIsValid(validDir) {
			t.Errorf("Expected valid directory to be valid, but got invalid")
		}

		// Test non-empty directory (should be invalid)
		if MountpointIsValid(nonEmptyDir) {
			t.Errorf("Expected non-empty directory to be invalid, but got valid")
		}

		// Test file path (should be invalid)
		if MountpointIsValid(filePath) {
			t.Errorf("Expected file path to be invalid, but got valid")
		}

		// Test non-existent path (should be invalid)
		if MountpointIsValid("/non/existent/path") {
			t.Errorf("Expected non-existent path to be invalid, but got valid")
		}
	})
}

// TestUT_UI_02_01_HomePath_EscapeAndUnescape_ConvertsPaths tests converting paths from ~/some_path to /home/username/some_path and back.
//
//	Test Case ID    UT-UI-02-01
//	Title           Home Path Conversion
//	Description     Tests converting paths from ~/some_path to /home/username/some_path and back
//	Preconditions   None
//	Steps           1. Call EscapeHome with different paths
//	                2. Call UnescapeHome with the escaped paths
//	                3. Check if the results match expectations
//	Expected Result Paths are correctly escaped and unescaped
//	Notes: This test verifies that the EscapeHome and UnescapeHome functions correctly convert paths.
func TestUT_UI_02_01_HomePath_EscapeAndUnescape_ConvertsPaths(t *testing.T) {
	// Create a test fixture
	fixture := framework.NewUnitTestFixture("HomePathConversionFixture")

	// Use the fixture to run the test
	fixture.Use(t, func(t *testing.T, fixture interface{}) {
		// Get the home directory for testing
		homeDir, err := os.UserHomeDir()
		if err != nil {
			t.Fatalf("Failed to get home directory: %v", err)
		}

		// Test cases for EscapeHome
		testCases := []struct {
			input    string
			expected string
		}{
			{filepath.Join(homeDir, "Documents"), "~/Documents"},
			{filepath.Join(homeDir, "Documents", "test.txt"), "~/Documents/test.txt"},
			{"/tmp/test", "/tmp/test"},
			{homeDir, "~"},
		}

		for _, tc := range testCases {
			result := EscapeHome(tc.input)
			if result != tc.expected {
				t.Errorf("EscapeHome(%q) = %q, expected %q", tc.input, result, tc.expected)
			}
		}

		// Test cases for UnescapeHome
		unescapeTestCases := []struct {
			input    string
			expected string
		}{
			{"~/Documents", filepath.Join(homeDir, "Documents")},
			{"~/Documents/test.txt", filepath.Join(homeDir, "Documents", "test.txt")},
			{"/tmp/test", "/tmp/test"},
			{"~", "~"}, // "~" alone should not be unescaped
		}

		for _, tc := range unescapeTestCases {
			result := UnescapeHome(tc.input)
			if result != tc.expected {
				t.Errorf("UnescapeHome(%q) = %q, expected %q", tc.input, result, tc.expected)
			}
		}

		// Test round-trip conversion
		originalPath := filepath.Join(homeDir, "test", "path")
		escaped := EscapeHome(originalPath)
		unescaped := UnescapeHome(escaped)
		if unescaped != originalPath {
			t.Errorf("Round-trip conversion failed: %q -> %q -> %q", originalPath, escaped, unescaped)
		}
	})
}

// TestUT_UI_03_01_AccountName_FromAuthTokenFiles_ReturnsCorrectNames tests retrieving account names from auth token files.
//
//	Test Case ID    UT-UI-03-01
//	Title           Account Name Retrieval
//	Description     Tests retrieving account names from auth token files
//	Preconditions   None
//	Steps           1. Create various auth token files (valid, invalid, empty)
//	                2. Call GetAccountName with different instances
//	                3. Check if the results match expectations
//	Expected Result Account names are correctly retrieved from valid files, errors are returned for invalid files
//	Notes: This test verifies that the GetAccountName function correctly retrieves account names from auth token files.
func TestUT_UI_03_01_AccountName_FromAuthTokenFiles_ReturnsCorrectNames(t *testing.T) {
	// Create a test fixture
	fixture := framework.NewUnitTestFixture("AccountNameRetrievalFixture")

	// Set up the fixture
	fixture.WithSetup(func(t *testing.T) (interface{}, error) {
		// Create a temporary directory for the test
		tempDir, err := os.MkdirTemp("", "onemount-test-*")
		if err != nil {
			return nil, err
		}

		// Create test auth token files
		validAuthFile := filepath.Join(tempDir, "valid-auth.json")
		if err := os.WriteFile(validAuthFile, []byte(`{"account": "test@example.com"}`), 0644); err != nil {
			return nil, err
		}

		invalidAuthFile := filepath.Join(tempDir, "invalid-auth.json")
		if err := os.WriteFile(invalidAuthFile, []byte(`{invalid json}`), 0644); err != nil {
			return nil, err
		}

		emptyAuthFile := filepath.Join(tempDir, "empty-auth.json")
		if err := os.WriteFile(emptyAuthFile, []byte(`{}`), 0644); err != nil {
			return nil, err
		}

		return map[string]interface{}{
			"tempDir":         tempDir,
			"validAuthFile":   validAuthFile,
			"invalidAuthFile": invalidAuthFile,
			"emptyAuthFile":   emptyAuthFile,
		}, nil
	}).WithTeardown(func(t *testing.T, fixture interface{}) error {
		// Clean up the temporary directory
		data := fixture.(map[string]interface{})
		tempDir := data["tempDir"].(string)
		return os.RemoveAll(tempDir)
	})

	// Use the fixture to run the test
	fixture.Use(t, func(t *testing.T, f interface{}) {
		fixtureData := f.(*framework.UnitTestFixture)
		data := fixtureData.SetupData.(map[string]interface{})
		tempDir := data["tempDir"].(string)

		// Test valid auth file
		validInstance := "valid-auth"
		validAuthPath := filepath.Join(tempDir, validInstance)
		if err := os.Mkdir(validAuthPath, 0755); err != nil {
			t.Fatalf("Failed to create valid auth directory: %v", err)
		}
		validAuthFile := filepath.Join(validAuthPath, "auth_tokens.json")
		if err := os.WriteFile(validAuthFile, []byte(`{"account": "test@example.com"}`), 0644); err != nil {
			t.Fatalf("Failed to create valid auth file: %v", err)
		}

		accountName, err := graph.GetAccountName(tempDir, validInstance)
		if err != nil {
			t.Errorf("Expected no error for valid auth file, got: %v", err)
		}
		if accountName != "test@example.com" {
			t.Errorf("Expected account name 'test@example.com', got: %q", accountName)
		}

		// Test invalid JSON auth file
		invalidInstance := "invalid-auth"
		invalidAuthPath := filepath.Join(tempDir, invalidInstance)
		if err := os.Mkdir(invalidAuthPath, 0755); err != nil {
			t.Fatalf("Failed to create invalid auth directory: %v", err)
		}
		invalidAuthFile := filepath.Join(invalidAuthPath, "auth_tokens.json")
		if err := os.WriteFile(invalidAuthFile, []byte(`{invalid json}`), 0644); err != nil {
			t.Fatalf("Failed to create invalid auth file: %v", err)
		}

		_, err = graph.GetAccountName(tempDir, invalidInstance)
		if err == nil {
			t.Errorf("Expected error for invalid JSON auth file, got nil")
		}

		// Test empty auth file
		emptyInstance := "empty-auth"
		emptyAuthPath := filepath.Join(tempDir, emptyInstance)
		if err := os.Mkdir(emptyAuthPath, 0755); err != nil {
			t.Fatalf("Failed to create empty auth directory: %v", err)
		}
		emptyAuthFile := filepath.Join(emptyAuthPath, "auth_tokens.json")
		if err := os.WriteFile(emptyAuthFile, []byte(`{}`), 0644); err != nil {
			t.Fatalf("Failed to create empty auth file: %v", err)
		}

		accountName, err = graph.GetAccountName(tempDir, emptyInstance)
		if err != nil {
			t.Errorf("Expected no error for empty auth file, got: %v", err)
		}
		if accountName != "" {
			t.Errorf("Expected empty account name, got: %q", accountName)
		}

		// Test non-existent auth file
		_, err = graph.GetAccountName(tempDir, "non-existent")
		if err == nil {
			t.Errorf("Expected error for non-existent auth file, got nil")
		}
	})
}
