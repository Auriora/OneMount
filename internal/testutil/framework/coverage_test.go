// Package testutil provides testing utilities for the OneMount project.
package framework

import (
	"github.com/auriora/onemount/internal/testutil"
	"os"
	"path/filepath"
	"testing"
)

// TestUT_CV_01_01_CoverageReporter_BasicFunctionality_WorksCorrectly tests the basic functionality of the CoverageReporter.
//
//	Test Case ID    UT-CV-01-01
//	Title           Coverage Reporter Basic Functionality
//	Description     Tests the basic functionality of the CoverageReporter
//	Preconditions   None
//	Steps           1. Create a temporary directory for test output
//	                2. Create a CoverageReporter with test thresholds
//	                3. Test CollectCoverage functionality
//	                4. Test ReportCoverage functionality
//	                5. Test CheckThresholds functionality
//	                6. Test GenerateGoLandReport functionality
//	Expected Result All coverage reporter functions work correctly
func TestUT_CV_01_01_CoverageReporter_BasicFunctionality_WorksCorrectly(t *testing.T) {
	// Create a temporary directory for test output

	tempDir, err := os.MkdirTemp(testutil.TestSandboxTmpDir, "coverage-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a CoverageReporter with test thresholds
	thresholds := CoverageThresholds{
		LineCoverage:   70.0,
		FuncCoverage:   70.0,
		BranchCoverage: 50.0,
	}
	reporter := NewCoverageReporter(thresholds, tempDir)

	// Test CollectCoverage
	t.Run("CollectCoverage", func(t *testing.T) {
		// This test is more of an integration test and might be skipped in CI
		if testing.Short() {
			t.Skip("Skipping coverage collection in short mode")
		}

		err := reporter.CollectCoverage()
		if err != nil {
			t.Fatalf("CollectCoverage failed: %v", err)
		}

		// Verify that the coverage file was created
		coverageFile := filepath.Join(tempDir, "coverage.out")
		if _, err := os.Stat(coverageFile); os.IsNotExist(err) {
			t.Errorf("Coverage file was not created: %s", coverageFile)
		}
	})

	// Test ReportCoverage
	t.Run("ReportCoverage", func(t *testing.T) {
		// Skip if we didn't collect coverage
		coverageFile := filepath.Join(tempDir, "coverage.out")
		if _, err := os.Stat(coverageFile); os.IsNotExist(err) {
			t.Skip("Coverage file not found, skipping report test")
		}

		err := reporter.ReportCoverage()
		if err != nil {
			t.Fatalf("ReportCoverage failed: %v", err)
		}

		// Verify that the HTML report was created
		htmlFile := filepath.Join(tempDir, "coverage.html")
		if _, err := os.Stat(htmlFile); os.IsNotExist(err) {
			t.Errorf("HTML report was not created: %s", htmlFile)
		}

		// Verify that the custom report was created
		customReportFile := filepath.Join(tempDir, "coverage_report.html")
		if _, err := os.Stat(customReportFile); os.IsNotExist(err) {
			t.Errorf("Custom report was not created: %s", customReportFile)
		}
	})

	// Test CheckThresholds
	t.Run("CheckThresholds", func(t *testing.T) {
		// Skip if we didn't collect coverage
		coverageFile := filepath.Join(tempDir, "coverage.out")
		if _, err := os.Stat(coverageFile); os.IsNotExist(err) {
			t.Skip("Coverage file not found, skipping threshold test")
		}

		passed, err := reporter.CheckThresholds()
		if err != nil {
			t.Fatalf("CheckThresholds failed: %v", err)
		}

		// We don't know if it will pass or fail, but we can check that the function runs
		t.Logf("Thresholds check passed: %v", passed)

		// If it failed, verify that the violations file was created
		if !passed {
			violationsFile := filepath.Join(tempDir, "threshold_violations.txt")
			if _, err := os.Stat(violationsFile); os.IsNotExist(err) {
				t.Errorf("Violations file was not created: %s", violationsFile)
			}
		}
	})

	// Test GenerateGoLandReport
	t.Run("GenerateGoLandReport", func(t *testing.T) {
		// Skip if we didn't collect coverage
		coverageFile := filepath.Join(tempDir, "coverage.out")
		if _, err := os.Stat(coverageFile); os.IsNotExist(err) {
			t.Skip("Coverage file not found, skipping GoLand report test")
		}

		err := reporter.GenerateGoLandReport()
		if err != nil {
			t.Fatalf("GenerateGoLandReport failed: %v", err)
		}

		// Verify that the coverage file was copied to the project root
		rootCoverageFile := "coverage.out"
		if _, err := os.Stat(rootCoverageFile); os.IsNotExist(err) {
			t.Errorf("Root coverage file was not created: %s", rootCoverageFile)
		} else {
			// Clean up the file after the test
			defer os.Remove(rootCoverageFile)
		}
	})
}

// TestUT_CV_02_01_HelperFunctions_UtilityFunctions_WorkCorrectly tests the helper functions in the coverage.go file.
//
//	Test Case ID    UT-CV-02-01
//	Title           Coverage Helper Functions
//	Description     Tests the utility functions used by the coverage reporter
//	Preconditions   None
//	Steps           1. Test parseFloat function with various inputs
//	                2. Test getCoverageClass function with various coverage and threshold values
//	Expected Result All helper functions work correctly with various inputs
func TestUT_CV_02_01_HelperFunctions_UtilityFunctions_WorkCorrectly(t *testing.T) {
	// Test parseFloat
	t.Run("parseFloat", func(t *testing.T) {
		tests := []struct {
			input    string
			expected float64
			hasError bool
		}{
			{"10.5", 10.5, false},
			{"0", 0.0, false},
			{"invalid", 0.0, true},
		}

		for _, test := range tests {
			result, err := parseFloat(test.input)
			if test.hasError && err == nil {
				t.Errorf("parseFloat(%s) should have returned an error", test.input)
			}
			if !test.hasError && err != nil {
				t.Errorf("parseFloat(%s) returned an error: %v", test.input, err)
			}
			if result != test.expected {
				t.Errorf("parseFloat(%s) = %f, expected %f", test.input, result, test.expected)
			}
		}
	})

	// Test getCoverageClass
	t.Run("getCoverageClass", func(t *testing.T) {
		tests := []struct {
			coverage  float64
			threshold float64
			expected  string
		}{
			{80.0, 70.0, "good"},
			{70.0, 70.0, "good"},
			{60.0, 70.0, "warning"},
			{56.0, 70.0, "warning"},
			{50.0, 70.0, "bad"},
		}

		for _, test := range tests {
			result := getCoverageClass(test.coverage, test.threshold)
			if result != test.expected {
				t.Errorf("getCoverageClass(%f, %f) = %s, expected %s",
					test.coverage, test.threshold, result, test.expected)
			}
		}
	})
}
