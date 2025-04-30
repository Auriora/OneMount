// Package testutil provides testing utilities for the OneMount project.
package testutil

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// PackageCoverage represents coverage data for a single package
type PackageCoverage struct {
	PackageName    string
	LineCoverage   float64
	FuncCoverage   float64
	BranchCoverage float64
	Files          map[string]FileCoverage
}

// FileCoverage represents coverage data for a single file
type FileCoverage struct {
	FileName       string
	LineCoverage   float64
	FuncCoverage   float64
	BranchCoverage float64
}

// HistoricalCoverage represents coverage data at a point in time
type HistoricalCoverage struct {
	Timestamp       int64
	TotalCoverage   float64
	PackageCoverage map[string]float64
}

// CoverageThresholds defines minimum acceptable coverage levels
type CoverageThresholds struct {
	LineCoverage   float64
	FuncCoverage   float64
	BranchCoverage float64
}

// CoverageReporterImpl implements the CoverageReporter interface
type CoverageReporterImpl struct {
	// Coverage data by package
	packageCoverage map[string]PackageCoverage

	// Historical coverage data
	historicalData []HistoricalCoverage

	// Coverage thresholds
	thresholds CoverageThresholds

	// Output directory for reports
	outputDir string

	// Coverage profile file path
	coverageFile string
}

// NewCoverageReporter creates a new CoverageReporter with the given thresholds
func NewCoverageReporter(thresholds CoverageThresholds, outputDir string) *CoverageReporterImpl {
	// Create output directory if it doesn't exist
	if outputDir != "" {
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			fmt.Printf("Warning: Failed to create output directory: %v\n", err)
		}
	}

	return &CoverageReporterImpl{
		packageCoverage: make(map[string]PackageCoverage),
		historicalData:  make([]HistoricalCoverage, 0),
		thresholds:      thresholds,
		outputDir:       outputDir,
		coverageFile:    filepath.Join(outputDir, "coverage.out"),
	}
}

// CollectCoverage collects coverage data using Go's built-in coverage tools
func (cr *CoverageReporterImpl) CollectCoverage() error {
	// Run go test with coverage
	cmd := exec.Command("go", "test", "-coverprofile="+cr.coverageFile, "./...")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to run tests with coverage: %v\n%s", err, output)
	}

	// Parse the coverage file
	return cr.parseCoverageFile()
}

// parseCoverageFile parses the coverage.out file and populates the packageCoverage map
func (cr *CoverageReporterImpl) parseCoverageFile() error {
	// Read the coverage file
	data, err := os.ReadFile(cr.coverageFile)
	if err != nil {
		return fmt.Errorf("failed to read coverage file: %v", err)
	}

	// Parse the coverage data
	lines := strings.Split(string(data), "\n")
	if len(lines) == 0 {
		return fmt.Errorf("empty coverage file")
	}

	// Skip the mode line
	lines = lines[1:]

	// Clear existing coverage data
	cr.packageCoverage = make(map[string]PackageCoverage)

	// Process each line
	for _, line := range lines {
		if line == "" {
			continue
		}

		// Parse the line
		parts := strings.Split(line, ":")
		if len(parts) < 2 {
			continue
		}

		filePath := parts[0]

		// Extract package name from file path
		pkgPath := filepath.Dir(filePath)
		pkgName := filepath.Base(pkgPath)

		// Get or create package coverage
		pkg, exists := cr.packageCoverage[pkgPath]
		if !exists {
			pkg = PackageCoverage{
				PackageName: pkgName,
				Files:       make(map[string]FileCoverage),
			}
		}

		// Get or create file coverage
		fileName := filepath.Base(filePath)
		file, exists := pkg.Files[fileName]
		if !exists {
			file = FileCoverage{
				FileName: fileName,
			}
		}

		// For now, we only have line coverage from the standard Go tools
		// We'll calculate it more precisely in the ReportCoverage method

		// Update the maps
		pkg.Files[fileName] = file
		cr.packageCoverage[pkgPath] = pkg
	}

	// Run go tool cover to get more detailed coverage information
	return cr.collectDetailedCoverage()
}

// collectDetailedCoverage uses go tool cover to get more detailed coverage information
func (cr *CoverageReporterImpl) collectDetailedCoverage() error {
	// Run go tool cover with -func to get function coverage
	cmd := exec.Command("go", "tool", "cover", "-func="+cr.coverageFile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to get function coverage: %v\n%s", err, output)
	}

	// Parse the output
	lines := strings.Split(string(output), "\n")

	// Process each line
	for _, line := range lines {
		if line == "" || strings.Contains(line, "total:") {
			continue
		}

		// Parse the line
		parts := strings.Fields(line)
		if len(parts) < 3 {
			continue
		}

		filePath := parts[0]
		funcName := parts[1]
		coverageStr := strings.TrimSuffix(parts[2], "%")

		// Convert coverage string to float
		coverage, err := parseFloat(coverageStr)
		if err != nil {
			continue
		}

		// Extract package name from file path
		pkgPath := filepath.Dir(filePath)
		fileName := filepath.Base(filePath)

		// Update package and file coverage
		pkg, exists := cr.packageCoverage[pkgPath]
		if !exists {
			continue
		}

		file, exists := pkg.Files[fileName]
		if !exists {
			continue
		}

		// Update function coverage (we're just averaging for now)
		// In a real implementation, we would track coverage for each function
		if funcName == "total:" {
			file.LineCoverage = coverage
		}

		// Update the maps
		pkg.Files[fileName] = file
		cr.packageCoverage[pkgPath] = pkg
	}

	// Calculate package-level coverage
	cr.calculatePackageCoverage()

	// Store historical data
	cr.storeHistoricalData()

	return nil
}

// calculatePackageCoverage calculates coverage metrics for each package
func (cr *CoverageReporterImpl) calculatePackageCoverage() {
	for pkgPath, pkg := range cr.packageCoverage {
		var totalLineCoverage float64
		var totalFuncCoverage float64
		var totalBranchCoverage float64
		var fileCount int

		for _, file := range pkg.Files {
			totalLineCoverage += file.LineCoverage
			totalFuncCoverage += file.FuncCoverage
			totalBranchCoverage += file.BranchCoverage
			fileCount++
		}

		if fileCount > 0 {
			pkg.LineCoverage = totalLineCoverage / float64(fileCount)
			pkg.FuncCoverage = totalFuncCoverage / float64(fileCount)
			pkg.BranchCoverage = totalBranchCoverage / float64(fileCount)
		}

		cr.packageCoverage[pkgPath] = pkg
	}
}

// storeHistoricalData stores the current coverage data in the historical data
func (cr *CoverageReporterImpl) storeHistoricalData() {
	// Calculate total coverage
	var totalCoverage float64
	var packageCount int
	packageCoverage := make(map[string]float64)

	for pkgPath, pkg := range cr.packageCoverage {
		totalCoverage += pkg.LineCoverage
		packageCount++
		packageCoverage[pkgPath] = pkg.LineCoverage
	}

	if packageCount > 0 {
		totalCoverage /= float64(packageCount)
	}

	// Create historical data entry
	historicalData := HistoricalCoverage{
		Timestamp:       time.Now().Unix(),
		TotalCoverage:   totalCoverage,
		PackageCoverage: packageCoverage,
	}

	// Add to historical data
	cr.historicalData = append(cr.historicalData, historicalData)

	// Save historical data to file
	if cr.outputDir != "" {
		historyFile := filepath.Join(cr.outputDir, "coverage_history.json")
		data, err := json.Marshal(cr.historicalData)
		if err == nil {
			_ = os.WriteFile(historyFile, data, 0644)
		}
	}
}

// ReportCoverage generates a coverage report
func (cr *CoverageReporterImpl) ReportCoverage() error {
	// Generate HTML report using go tool cover
	htmlFile := filepath.Join(cr.outputDir, "coverage.html")
	cmd := exec.Command("go", "tool", "cover", "-html="+cr.coverageFile, "-o="+htmlFile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to generate HTML report: %v\n%s", err, output)
	}

	// Generate our custom report
	return cr.generateCustomReport()
}

// generateCustomReport generates a custom HTML report with package-level metrics
func (cr *CoverageReporterImpl) generateCustomReport() error {
	if cr.outputDir == "" {
		return nil
	}

	// Create the report file
	reportFile := filepath.Join(cr.outputDir, "coverage_report.html")
	file, err := os.Create(reportFile)
	if err != nil {
		return fmt.Errorf("failed to create report file: %v", err)
	}
	defer file.Close()

	// Generate the report
	return cr.writeHTMLReport(file)
}

// writeHTMLReport writes the HTML report to the given writer
func (cr *CoverageReporterImpl) writeHTMLReport(w io.Writer) error {
	// Define the HTML template
	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <title>Coverage Report</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        h1 { color: #333; }
        table { border-collapse: collapse; width: 100%; }
        th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
        th { background-color: #f2f2f2; }
        tr:nth-child(even) { background-color: #f9f9f9; }
        .good { color: green; }
        .warning { color: orange; }
        .bad { color: red; }
    </style>
</head>
<body>
    <h1>Coverage Report</h1>
    
    <h2>Summary</h2>
    <p>Generated on: {{.Timestamp}}</p>
    
    <h2>Package Coverage</h2>
    <table>
        <tr>
            <th>Package</th>
            <th>Line Coverage</th>
            <th>Function Coverage</th>
            <th>Branch Coverage</th>
        </tr>
        {{range .Packages}}
        <tr>
            <td>{{.PackageName}}</td>
            <td class="{{.LineClass}}">{{.LineCoverage}}%</td>
            <td class="{{.FuncClass}}">{{.FuncCoverage}}%</td>
            <td class="{{.BranchClass}}">{{.BranchCoverage}}%</td>
        </tr>
        {{end}}
    </table>
    
    <h2>Thresholds</h2>
    <table>
        <tr>
            <th>Metric</th>
            <th>Threshold</th>
        </tr>
        <tr>
            <td>Line Coverage</td>
            <td>{{.Thresholds.LineCoverage}}%</td>
        </tr>
        <tr>
            <td>Function Coverage</td>
            <td>{{.Thresholds.FuncCoverage}}%</td>
        </tr>
        <tr>
            <td>Branch Coverage</td>
            <td>{{.Thresholds.BranchCoverage}}%</td>
        </tr>
    </table>
</body>
</html>
`

	// Create the template
	t, err := template.New("report").Parse(tmpl)
	if err != nil {
		return fmt.Errorf("failed to parse template: %v", err)
	}

	// Prepare the data for the template
	type PackageData struct {
		PackageName    string
		LineCoverage   string
		FuncCoverage   string
		BranchCoverage string
		LineClass      string
		FuncClass      string
		BranchClass    string
	}

	data := struct {
		Timestamp  string
		Packages   []PackageData
		Thresholds CoverageThresholds
	}{
		Timestamp:  time.Now().Format(time.RFC1123),
		Packages:   make([]PackageData, 0, len(cr.packageCoverage)),
		Thresholds: cr.thresholds,
	}

	// Add package data
	for _, pkg := range cr.packageCoverage {
		// Format coverage values
		lineCoverage := fmt.Sprintf("%.2f", pkg.LineCoverage)
		funcCoverage := fmt.Sprintf("%.2f", pkg.FuncCoverage)
		branchCoverage := fmt.Sprintf("%.2f", pkg.BranchCoverage)

		// Determine CSS classes based on thresholds
		lineClass := getCoverageClass(pkg.LineCoverage, cr.thresholds.LineCoverage)
		funcClass := getCoverageClass(pkg.FuncCoverage, cr.thresholds.FuncCoverage)
		branchClass := getCoverageClass(pkg.BranchCoverage, cr.thresholds.BranchCoverage)

		// Add to packages
		data.Packages = append(data.Packages, PackageData{
			PackageName:    pkg.PackageName,
			LineCoverage:   lineCoverage,
			FuncCoverage:   funcCoverage,
			BranchCoverage: branchCoverage,
			LineClass:      lineClass,
			FuncClass:      funcClass,
			BranchClass:    branchClass,
		})
	}

	// Execute the template
	return t.Execute(w, data)
}

// CheckThresholds checks if coverage meets defined thresholds
func (cr *CoverageReporterImpl) CheckThresholds() (bool, error) {
	// Calculate total coverage
	var totalLineCoverage float64
	var totalFuncCoverage float64
	var totalBranchCoverage float64
	var packageCount int

	for _, pkg := range cr.packageCoverage {
		totalLineCoverage += pkg.LineCoverage
		totalFuncCoverage += pkg.FuncCoverage
		totalBranchCoverage += pkg.BranchCoverage
		packageCount++
	}

	if packageCount > 0 {
		totalLineCoverage /= float64(packageCount)
		totalFuncCoverage /= float64(packageCount)
		totalBranchCoverage /= float64(packageCount)
	}

	// Check against thresholds
	linePassed := totalLineCoverage >= cr.thresholds.LineCoverage
	funcPassed := totalFuncCoverage >= cr.thresholds.FuncCoverage
	branchPassed := totalBranchCoverage >= cr.thresholds.BranchCoverage

	// Generate a report of threshold violations
	if !linePassed || !funcPassed || !branchPassed {
		var violations []string

		if !linePassed {
			violations = append(violations, fmt.Sprintf("Line coverage %.2f%% is below threshold %.2f%%",
				totalLineCoverage, cr.thresholds.LineCoverage))
		}

		if !funcPassed {
			violations = append(violations, fmt.Sprintf("Function coverage %.2f%% is below threshold %.2f%%",
				totalFuncCoverage, cr.thresholds.FuncCoverage))
		}

		if !branchPassed {
			violations = append(violations, fmt.Sprintf("Branch coverage %.2f%% is below threshold %.2f%%",
				totalBranchCoverage, cr.thresholds.BranchCoverage))
		}

		// Write violations to a file
		if cr.outputDir != "" {
			violationsFile := filepath.Join(cr.outputDir, "threshold_violations.txt")
			_ = os.WriteFile(violationsFile, []byte(strings.Join(violations, "\n")), 0644)
		}

		return false, nil
	}

	return true, nil
}

// GenerateGoLandReport generates a coverage report compatible with JetBrains GoLand
func (cr *CoverageReporterImpl) GenerateGoLandReport() error {
	// GoLand uses the same format as the standard Go coverage tool
	// We just need to make sure the coverage.out file is in the right format
	// and in the right location for GoLand to find it

	// Copy the coverage file to the project root if it's not already there
	if cr.outputDir != "" && cr.outputDir != "." {
		srcFile := cr.coverageFile
		dstFile := "coverage.out"

		data, err := os.ReadFile(srcFile)
		if err != nil {
			return fmt.Errorf("failed to read coverage file: %v", err)
		}

		err = os.WriteFile(dstFile, data, 0644)
		if err != nil {
			return fmt.Errorf("failed to write coverage file: %v", err)
		}
	}

	return nil
}

// Helper functions

// parseFloat parses a string to a float64
func parseFloat(s string) (float64, error) {
	var f float64
	_, err := fmt.Sscanf(s, "%f", &f)
	return f, err
}

// getCoverageClass returns a CSS class based on coverage value
func getCoverageClass(coverage, threshold float64) string {
	if coverage >= threshold {
		return "good"
	} else if coverage >= threshold*0.8 {
		return "warning"
	}
	return "bad"
}
