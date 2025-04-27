package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"time"
)

// LogEntry represents a structured log entry
type LogEntry struct {
	Time       string  `json:"time"`
	Level      string  `json:"level"`
	Method     string  `json:"method,omitempty"`
	Phase      string  `json:"phase,omitempty"`
	Goroutine  string  `json:"goroutine,omitempty"`
	Duration   float64 `json:"duration_ms,omitempty"`
	Message    string  `json:"message"`
	Path       string  `json:"path,omitempty"`
	ID         string  `json:"id,omitempty"`
	Operation  string  `json:"operation,omitempty"`
	Component  string  `json:"component,omitempty"`
	Return1    string  `json:"return1,omitempty"`
	Param1     string  `json:"param1,omitempty"`
	Error      string  `json:"error,omitempty"`
	Status     string  `json:"status,omitempty"`
	Priority   string  `json:"priority,omitempty"`
	Name       string  `json:"name,omitempty"`
	HasChanges bool    `json:"hasChanges,omitempty"`
}

// MethodCall represents a method call with entry and exit information
type MethodCall struct {
	Method     string
	Goroutine  string
	EntryTime  time.Time
	ExitTime   time.Time
	Duration   time.Duration
	Parameters map[string]string
	Returns    map[string]string
	Caller     *MethodCall
	Callees    []*MethodCall
}

// WorkflowAnalyzer analyzes the logs to extract workflow information
type WorkflowAnalyzer struct {
	LogFile       string
	MountPoint    string
	TestDir       string
	TestFile      string
	ConflictFile  string
	MethodCalls   map[string]*MethodCall   // key is method:goroutine
	ActiveMethods map[string]*MethodCall   // key is goroutine
	CallStack     map[string][]*MethodCall // key is goroutine
	SequenceCalls []*MethodCall
	UploadCalls   []*MethodCall
	DownloadCalls []*MethodCall
	ConflictCalls []*MethodCall
}

func NewWorkflowAnalyzer(mountPoint string) *WorkflowAnalyzer {
	// Create a temporary log file
	logFile := filepath.Join(os.TempDir(), "onedriver_workflow.log")

	// Create test directory and files
	testDir := filepath.Join(mountPoint, "workflow_test_"+time.Now().Format("20060102150405"))
	testFile := filepath.Join(testDir, "test_file.txt")
	conflictFile := filepath.Join(testDir, "conflict_file.txt")

	return &WorkflowAnalyzer{
		LogFile:       logFile,
		MountPoint:    mountPoint,
		TestDir:       testDir,
		TestFile:      testFile,
		ConflictFile:  conflictFile,
		MethodCalls:   make(map[string]*MethodCall),
		ActiveMethods: make(map[string]*MethodCall),
		CallStack:     make(map[string][]*MethodCall),
		SequenceCalls: make([]*MethodCall, 0),
		UploadCalls:   make([]*MethodCall, 0),
		DownloadCalls: make([]*MethodCall, 0),
		ConflictCalls: make([]*MethodCall, 0),
	}
}

// SetupLogging configures onedriver to log at DEBUG level to our log file
func (wa *WorkflowAnalyzer) SetupLogging() error {
	// Create a config file that sets the log level to DEBUG and the log output to our log file
	configDir := filepath.Join(os.Getenv("HOME"), ".config", "onedriver")
	configFile := filepath.Join(configDir, "config.yml")

	// Backup the existing config file if it exists
	if _, err := os.Stat(configFile); err == nil {
		backupFile := configFile + ".bak"
		if err := os.Rename(configFile, backupFile); err != nil {
			return fmt.Errorf("failed to backup config file: %w", err)
		}
		fmt.Printf("Backed up existing config file to %s\n", backupFile)
	}

	// Create the config directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Write the new config file
	config := fmt.Sprintf(`log_level: debug
log_output: %s
`, wa.LogFile)

	if err := os.WriteFile(configFile, []byte(config), 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	fmt.Printf("Set up logging to %s at DEBUG level\n", wa.LogFile)
	return nil
}

// RestartOnedriver restarts the onedriver service to apply the new logging configuration
func (wa *WorkflowAnalyzer) RestartOnedriver() error {
	// Get the user's username
	user := os.Getenv("USER")
	if user == "" {
		return fmt.Errorf("failed to get username")
	}

	// Restart the onedriver service
	cmd := exec.Command("systemctl", "--user", "restart", fmt.Sprintf("onedriver@%s.service", wa.MountPoint))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to restart onedriver service: %w", err)
	}

	// Wait for the service to start
	time.Sleep(5 * time.Second)

	fmt.Println("Restarted onedriver service")
	return nil
}

// ExecuteWorkflows executes the primary workflows (file upload, download, conflict resolution)
func (wa *WorkflowAnalyzer) ExecuteWorkflows() error {
	// Create test directory
	if err := os.MkdirAll(wa.TestDir, 0755); err != nil {
		return fmt.Errorf("failed to create test directory: %w", err)
	}
	fmt.Printf("Created test directory: %s\n", wa.TestDir)

	// Wait for the directory to be uploaded
	time.Sleep(2 * time.Second)

	// File upload workflow
	if err := wa.executeUploadWorkflow(); err != nil {
		return fmt.Errorf("upload workflow failed: %w", err)
	}

	// Wait for the file to be uploaded
	time.Sleep(5 * time.Second)

	// File download workflow
	if err := wa.executeDownloadWorkflow(); err != nil {
		return fmt.Errorf("download workflow failed: %w", err)
	}

	// Wait for the file to be downloaded
	time.Sleep(2 * time.Second)

	// Conflict resolution workflow
	if err := wa.executeConflictWorkflow(); err != nil {
		return fmt.Errorf("conflict workflow failed: %w", err)
	}

	// Wait for the conflict to be resolved
	time.Sleep(5 * time.Second)

	return nil
}

// executeUploadWorkflow executes the file upload workflow
func (wa *WorkflowAnalyzer) executeUploadWorkflow() error {
	// Create a test file
	content := "This is a test file for the upload workflow.\n"
	if err := os.WriteFile(wa.TestFile, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to create test file: %w", err)
	}
	fmt.Printf("Created test file: %s\n", wa.TestFile)

	// Append to the file to trigger another upload
	additionalContent := "This is additional content to trigger another upload.\n"
	f, err := os.OpenFile(wa.TestFile, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open test file: %w", err)
	}
	if _, err := f.WriteString(additionalContent); err != nil {
		f.Close()
		return fmt.Errorf("failed to append to test file: %w", err)
	}
	f.Close()
	fmt.Printf("Appended to test file: %s\n", wa.TestFile)

	return nil
}

// executeDownloadWorkflow executes the file download workflow
func (wa *WorkflowAnalyzer) executeDownloadWorkflow() error {
	// Remove the local file to force a download
	if err := os.Remove(wa.TestFile); err != nil {
		return fmt.Errorf("failed to remove test file: %w", err)
	}
	fmt.Printf("Removed test file: %s\n", wa.TestFile)

	// Access the file to trigger a download
	if _, err := os.Stat(wa.TestFile); err != nil {
		// This is expected, as the file is not downloaded yet
		fmt.Printf("File not found as expected: %s\n", wa.TestFile)
	}

	// Read the file to trigger a download
	content, err := os.ReadFile(wa.TestFile)
	if err != nil {
		return fmt.Errorf("failed to read test file: %w", err)
	}
	fmt.Printf("Read test file (%d bytes): %s\n", len(content), wa.TestFile)

	return nil
}

// executeConflictWorkflow executes the conflict resolution workflow
func (wa *WorkflowAnalyzer) executeConflictWorkflow() error {
	// Create a conflict file
	content := "This is a test file for the conflict workflow.\n"
	if err := os.WriteFile(wa.ConflictFile, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to create conflict file: %w", err)
	}
	fmt.Printf("Created conflict file: %s\n", wa.ConflictFile)

	// Wait for the file to be uploaded
	time.Sleep(3 * time.Second)

	// Simulate a conflict by modifying the file both locally and remotely
	// For this example, we'll just modify it locally and then force a sync
	additionalContent := "This is additional content to create a conflict.\n"
	f, err := os.OpenFile(wa.ConflictFile, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open conflict file: %w", err)
	}
	if _, err := f.WriteString(additionalContent); err != nil {
		f.Close()
		return fmt.Errorf("failed to append to conflict file: %w", err)
	}
	f.Close()
	fmt.Printf("Appended to conflict file: %s\n", wa.ConflictFile)

	// Force a sync by restarting onedriver
	if err := wa.RestartOnedriver(); err != nil {
		return fmt.Errorf("failed to restart onedriver: %w", err)
	}

	return nil
}

// CleanUp cleans up the test files and directories
func (wa *WorkflowAnalyzer) CleanUp() error {
	// Remove the test directory
	if err := os.RemoveAll(wa.TestDir); err != nil {
		return fmt.Errorf("failed to remove test directory: %w", err)
	}
	fmt.Printf("Removed test directory: %s\n", wa.TestDir)

	return nil
}

// AnalyzeLogs analyzes the logs to extract the sequence of function calls
func (wa *WorkflowAnalyzer) AnalyzeLogs() error {
	// Open the log file
	file, err := os.Open(wa.LogFile)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}
	defer file.Close()

	// Read the log file line by line
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		// Parse the log entry
		entry := wa.parseLogEntry(line)
		if entry == nil {
			continue
		}

		// Process the log entry
		wa.processLogEntry(entry)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading log file: %w", err)
	}

	// Process the method calls to build the sequence diagram
	wa.processMethodCalls()

	return nil
}

// parseLogEntry parses a log entry line into a LogEntry struct
func (wa *WorkflowAnalyzer) parseLogEntry(line string) *LogEntry {
	// Check if the line is a JSON object
	if !strings.HasPrefix(line, "{") || !strings.HasSuffix(line, "}") {
		return nil
	}

	// Parse the JSON
	var entry LogEntry
	if err := json.Unmarshal([]byte(line), &entry); err != nil {
		return nil
	}

	return &entry
}

// processLogEntry processes a log entry to extract method call information
func (wa *WorkflowAnalyzer) processLogEntry(entry *LogEntry) {
	// Skip entries that are not method calls
	if entry.Method == "" || entry.Phase == "" || entry.Goroutine == "" {
		return
	}

	// Create a key for the method call
	key := fmt.Sprintf("%s:%s", entry.Method, entry.Goroutine)

	if entry.Phase == "entry" {
		// Method entry
		methodCall := &MethodCall{
			Method:     entry.Method,
			Goroutine:  entry.Goroutine,
			EntryTime:  parseTime(entry.Time),
			Parameters: make(map[string]string),
			Returns:    make(map[string]string),
			Callees:    make([]*MethodCall, 0),
		}

		// Add parameters
		for k, v := range extractFields(entry, "param") {
			methodCall.Parameters[k] = v
		}

		// Add to method calls map
		wa.MethodCalls[key] = methodCall

		// Add to active methods map
		wa.ActiveMethods[entry.Goroutine] = methodCall

		// Add to call stack
		stack, ok := wa.CallStack[entry.Goroutine]
		if !ok {
			stack = make([]*MethodCall, 0)
		}

		// If there's a parent method, set the caller and add this method as a callee
		if len(stack) > 0 {
			parent := stack[len(stack)-1]
			methodCall.Caller = parent
			parent.Callees = append(parent.Callees, methodCall)
		}

		// Push to the stack
		wa.CallStack[entry.Goroutine] = append(stack, methodCall)

	} else if entry.Phase == "exit" {
		// Method exit
		methodCall, ok := wa.MethodCalls[key]
		if !ok {
			// Method entry not found, create a new one
			methodCall = &MethodCall{
				Method:     entry.Method,
				Goroutine:  entry.Goroutine,
				ExitTime:   parseTime(entry.Time),
				Parameters: make(map[string]string),
				Returns:    make(map[string]string),
				Callees:    make([]*MethodCall, 0),
			}
			wa.MethodCalls[key] = methodCall
		} else {
			// Update the exit time
			methodCall.ExitTime = parseTime(entry.Time)
			methodCall.Duration = time.Duration(entry.Duration * float64(time.Millisecond))
		}

		// Add returns
		for k, v := range extractFields(entry, "return") {
			methodCall.Returns[k] = v
		}

		// Remove from active methods map
		delete(wa.ActiveMethods, entry.Goroutine)

		// Pop from the call stack
		stack, ok := wa.CallStack[entry.Goroutine]
		if ok && len(stack) > 0 {
			wa.CallStack[entry.Goroutine] = stack[:len(stack)-1]
		}

		// Add to sequence calls if it's a top-level method (no caller)
		if methodCall.Caller == nil {
			wa.SequenceCalls = append(wa.SequenceCalls, methodCall)
		}

		// Check if this method is part of a specific workflow
		if isUploadMethod(methodCall.Method) {
			wa.UploadCalls = append(wa.UploadCalls, methodCall)
		} else if isDownloadMethod(methodCall.Method) {
			wa.DownloadCalls = append(wa.DownloadCalls, methodCall)
		} else if isConflictMethod(methodCall.Method) {
			wa.ConflictCalls = append(wa.ConflictCalls, methodCall)
		}
	}
}

// processMethodCalls processes the method calls to build the sequence diagram
func (wa *WorkflowAnalyzer) processMethodCalls() {
	// Sort the sequence calls by entry time
	sortMethodCalls(wa.SequenceCalls)
	sortMethodCalls(wa.UploadCalls)
	sortMethodCalls(wa.DownloadCalls)
	sortMethodCalls(wa.ConflictCalls)
}

// GeneratePlantUML generates a PlantUML sequence diagram from the method calls
func (wa *WorkflowAnalyzer) GeneratePlantUML() (string, error) {
	// Generate the PlantUML for each workflow
	uploadUML, err := wa.generateWorkflowUML("File Upload", wa.UploadCalls)
	if err != nil {
		return "", fmt.Errorf("failed to generate upload workflow UML: %w", err)
	}

	downloadUML, err := wa.generateWorkflowUML("File Download", wa.DownloadCalls)
	if err != nil {
		return "", fmt.Errorf("failed to generate download workflow UML: %w", err)
	}

	conflictUML, err := wa.generateWorkflowUML("Conflict Resolution", wa.ConflictCalls)
	if err != nil {
		return "", fmt.Errorf("failed to generate conflict workflow UML: %w", err)
	}

	// Combine the UML diagrams
	uml := `@startuml
title OneDriver Function Invocation Sequence

actor User
participant "Main" as Main
participant "Filesystem" as FS
participant "DeltaLoop" as Delta
participant "ContentCache" as Cache
participant "DownloadManager" as DM
participant "UploadManager" as UM
participant "GraphAPI" as Graph

` + uploadUML + `

` + downloadUML + `

` + conflictUML + `

@enduml`

	return uml, nil
}

// generateWorkflowUML generates a PlantUML sequence diagram for a specific workflow
func (wa *WorkflowAnalyzer) generateWorkflowUML(title string, calls []*MethodCall) (string, error) {
	if len(calls) == 0 {
		return fmt.Sprintf("== %s ==\n", title), nil
	}

	var sb strings.Builder

	// Add the workflow title
	sb.WriteString(fmt.Sprintf("== %s ==\n", title))

	// Generate the sequence diagram
	for _, call := range calls {
		// Skip internal methods
		if isInternalMethod(call.Method) {
			continue
		}

		// Determine the participant
		participant := getParticipant(call.Method)

		// Add the method call
		if call.Caller != nil {
			callerParticipant := getParticipant(call.Caller.Method)
			sb.WriteString(fmt.Sprintf("%s -> %s: %s\n", callerParticipant, participant, call.Method))
			sb.WriteString(fmt.Sprintf("%s --> %s: Return\n", participant, callerParticipant))
		} else {
			// Top-level method
			sb.WriteString(fmt.Sprintf("User -> %s: %s\n", participant, call.Method))
			sb.WriteString(fmt.Sprintf("%s --> User: Return\n", participant))
		}
	}

	return sb.String(), nil
}

// SavePlantUML saves the PlantUML sequence diagram to a file
func (wa *WorkflowAnalyzer) SavePlantUML(uml string) error {
	// Save the PlantUML to a file
	umlFile := filepath.Join(filepath.Dir(wa.LogFile), "onedriver_sequence.puml")
	if err := os.WriteFile(umlFile, []byte(uml), 0644); err != nil {
		return fmt.Errorf("failed to write UML file: %w", err)
	}

	fmt.Printf("Saved PlantUML sequence diagram to %s\n", umlFile)
	return nil
}

// Helper functions

// parseTime parses a time string into a time.Time
func parseTime(timeStr string) time.Time {
	// Parse the time string
	t, err := time.Parse("15:04:05", timeStr)
	if err != nil {
		return time.Time{}
	}

	// Set the date to today
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), t.Hour(), t.Minute(), t.Second(), 0, now.Location())
}

// extractFields extracts fields with a specific prefix from a log entry
func extractFields(entry *LogEntry, prefix string) map[string]string {
	fields := make(map[string]string)

	// Use reflection to get all fields
	v := reflect.ValueOf(entry).Elem()
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("json")
		if strings.HasPrefix(tag, prefix) {
			// Extract the field value
			value := v.Field(i).Interface()
			if value != nil && value != "" {
				fields[tag] = fmt.Sprintf("%v", value)
			}
		}
	}

	return fields
}

// sortMethodCalls sorts method calls by entry time
func sortMethodCalls(calls []*MethodCall) {
	// Sort by entry time
	for i := 0; i < len(calls); i++ {
		for j := i + 1; j < len(calls); j++ {
			if calls[i].EntryTime.After(calls[j].EntryTime) {
				calls[i], calls[j] = calls[j], calls[i]
			}
		}
	}
}

// isUploadMethod checks if a method is part of the upload workflow
func isUploadMethod(method string) bool {
	uploadMethods := []string{
		"QueueUpload",
		"QueueUploadWithPriority",
		"uploadLoop",
		"UploadFile",
		"UploadSession",
	}

	for _, m := range uploadMethods {
		if strings.Contains(method, m) {
			return true
		}
	}

	return false
}

// isDownloadMethod checks if a method is part of the download workflow
func isDownloadMethod(method string) bool {
	downloadMethods := []string{
		"QueueDownload",
		"processDownload",
		"DownloadFile",
		"DownloadSession",
	}

	for _, m := range downloadMethods {
		if strings.Contains(method, m) {
			return true
		}
	}

	return false
}

// isConflictMethod checks if a method is part of the conflict resolution workflow
func isConflictMethod(method string) bool {
	conflictMethods := []string{
		"MarkFileConflict",
		"CreateConflictCopy",
		"ResolveConflict",
	}

	for _, m := range conflictMethods {
		if strings.Contains(method, m) {
			return true
		}
	}

	return false
}

// isInternalMethod checks if a method is an internal method that should be excluded from the sequence diagram
func isInternalMethod(method string) bool {
	internalMethods := []string{
		"Lock",
		"Unlock",
		"RLock",
		"RUnlock",
	}

	for _, m := range internalMethods {
		if method == m {
			return true
		}
	}

	return false
}

// getParticipant determines the participant for a method
func getParticipant(method string) string {
	if strings.Contains(method, "Upload") {
		return "UM"
	} else if strings.Contains(method, "Download") {
		return "DM"
	} else if strings.Contains(method, "Delta") {
		return "Delta"
	} else if strings.Contains(method, "Cache") {
		return "Cache"
	} else if strings.Contains(method, "Graph") {
		return "Graph"
	} else {
		return "FS"
	}
}

func main() {
	// Get the mount point from the command line arguments
	if len(os.Args) < 2 {
		fmt.Println("Usage: workflow_analyzer <mount_point>")
		os.Exit(1)
	}

	mountPoint := os.Args[1]

	// Create a new workflow analyzer
	analyzer := NewWorkflowAnalyzer(mountPoint)

	// Set up logging
	if err := analyzer.SetupLogging(); err != nil {
		fmt.Printf("Error setting up logging: %v\n", err)
		os.Exit(1)
	}

	// Restart onedriver to apply the new logging configuration
	if err := analyzer.RestartOnedriver(); err != nil {
		fmt.Printf("Error restarting onedriver: %v\n", err)
		os.Exit(1)
	}

	// Execute the workflows
	if err := analyzer.ExecuteWorkflows(); err != nil {
		fmt.Printf("Error executing workflows: %v\n", err)
		os.Exit(1)
	}

	// Clean up
	if err := analyzer.CleanUp(); err != nil {
		fmt.Printf("Error cleaning up: %v\n", err)
		// Continue anyway
	}

	// Analyze the logs
	if err := analyzer.AnalyzeLogs(); err != nil {
		fmt.Printf("Error analyzing logs: %v\n", err)
		os.Exit(1)
	}

	// Generate the PlantUML sequence diagram
	uml, err := analyzer.GeneratePlantUML()
	if err != nil {
		fmt.Printf("Error generating PlantUML: %v\n", err)
		os.Exit(1)
	}

	// Save the PlantUML sequence diagram
	if err := analyzer.SavePlantUML(uml); err != nil {
		fmt.Printf("Error saving PlantUML: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Workflow analysis completed successfully!")
}
