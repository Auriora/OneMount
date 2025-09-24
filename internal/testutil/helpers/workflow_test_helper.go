// Package helpers provides testing utilities for the OneMount project.
package helpers

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/auriora/onemount/pkg/graph"
	"github.com/auriora/onemount/internal/testutil/framework"
)

// WorkflowFilesystemInterface extends FilesystemInterface with workflow-specific methods
type WorkflowFilesystemInterface interface {
	FilesystemInterface
	IsOffline() bool
	SetOfflineMode(mode int)
	GetOfflineChanges() []interface{}
	TrackOfflineChange(change interface{}) error
}

// SyncResult represents the result of a synchronization operation
type SyncResult struct {
	ProcessedChanges  int
	ConflictsFound    int
	ConflictsResolved int
	Duration          time.Duration
}

// SyncManagerInterface defines the interface for sync operations
type SyncManagerInterface interface {
	ProcessOfflineChangesWithRetry(ctx context.Context) (*SyncResult, error)
}

// WorkflowTestHelper provides utilities for testing complete user workflows
type WorkflowTestHelper struct {
	t             *testing.T
	mountHelper   *MountTestHelper
	filesystem    WorkflowFilesystemInterface
	syncManager   SyncManagerInterface
	auth          *graph.Auth
	workflowSteps []WorkflowStep
	workflowData  map[string]interface{}
	cleanup       []func() error
}

// WorkflowStep represents a single step in a user workflow
type WorkflowStep struct {
	Name        string
	Description string
	Action      func(helper *WorkflowTestHelper) error
	Verify      func(helper *WorkflowTestHelper) error
	Timeout     time.Duration
}

// WorkflowResult contains the results of a workflow execution
type WorkflowResult struct {
	Success       bool
	StepsExecuted int
	TotalSteps    int
	Duration      time.Duration
	FailedStep    string
	Error         error
}

// NewWorkflowTestHelper creates a new workflow test helper
func NewWorkflowTestHelper(t *testing.T) *WorkflowTestHelper {
	return &WorkflowTestHelper{
		t:             t,
		workflowSteps: make([]WorkflowStep, 0),
		workflowData:  make(map[string]interface{}),
		cleanup:       make([]func() error, 0),
	}
}

// WorkflowFilesystemFactory is a function that creates a workflow filesystem instance
type WorkflowFilesystemFactory func(auth *graph.Auth, mountPoint string, cacheTTL int) (WorkflowFilesystemInterface, SyncManagerInterface, error)

// SetupWorkflowWithFactory initializes the workflow test environment with a custom factory
func (h *WorkflowTestHelper) SetupWorkflowWithFactory(factory WorkflowFilesystemFactory) error {
	// Create mount helper
	h.mountHelper = NewMountTestHelper(h.t)

	// Create authentication for testing
	auth := GetTestAuth()
	h.auth = auth

	// Create filesystem and sync manager using factory
	filesystem, syncManager, err := factory(auth, h.mountHelper.GetMountPoint(), 300)
	if err != nil {
		return fmt.Errorf("failed to create filesystem: %w", err)
	}
	h.filesystem = filesystem
	h.syncManager = syncManager

	// Add cleanup for mount helper
	h.cleanup = append(h.cleanup, func() error {
		return h.mountHelper.Cleanup()
	})

	return nil
}

// AddWorkflowStep adds a step to the workflow
func (h *WorkflowTestHelper) AddWorkflowStep(step WorkflowStep) {
	if step.Timeout == 0 {
		step.Timeout = 30 * time.Second // Default timeout
	}
	h.workflowSteps = append(h.workflowSteps, step)
}

// SetWorkflowData sets data that can be shared between workflow steps
func (h *WorkflowTestHelper) SetWorkflowData(key string, value interface{}) {
	h.workflowData[key] = value
}

// GetWorkflowData gets data that was set by previous workflow steps
func (h *WorkflowTestHelper) GetWorkflowData(key string) interface{} {
	return h.workflowData[key]
}

// GetMountPoint returns the mount point path
func (h *WorkflowTestHelper) GetMountPoint() string {
	if h.mountHelper == nil {
		return ""
	}
	return h.mountHelper.GetMountPoint()
}

// GetFilesystem returns the filesystem instance
func (h *WorkflowTestHelper) GetFilesystem() WorkflowFilesystemInterface {
	return h.filesystem
}

// GetAuth returns the authentication instance
func (h *WorkflowTestHelper) GetAuth() *graph.Auth {
	return h.auth
}

// CreateFile creates a file in the mounted filesystem
func (h *WorkflowTestHelper) CreateFile(relativePath string, content []byte) error {
	if h.mountHelper == nil {
		return fmt.Errorf("workflow not setup")
	}
	return h.mountHelper.CreateTestFile(relativePath, content)
}

// ReadFile reads a file from the mounted filesystem
func (h *WorkflowTestHelper) ReadFile(relativePath string) ([]byte, error) {
	if h.mountHelper == nil {
		return nil, fmt.Errorf("workflow not setup")
	}
	return h.mountHelper.ReadTestFile(relativePath)
}

// FileExists checks if a file exists in the mounted filesystem
func (h *WorkflowTestHelper) FileExists(relativePath string) bool {
	if h.mountHelper == nil {
		return false
	}
	return h.mountHelper.VerifyFileExists(relativePath)
}

// GoOffline simulates going offline
func (h *WorkflowTestHelper) GoOffline() error {
	if h.filesystem == nil {
		return fmt.Errorf("filesystem not available")
	}

	graph.SetOperationalOffline(true)
	h.filesystem.SetOfflineMode(1) // OfflineModeEnabled = 1
	return nil
}

// GoOnline simulates going back online
func (h *WorkflowTestHelper) GoOnline() error {
	if h.filesystem == nil {
		return fmt.Errorf("filesystem not available")
	}

	graph.SetOperationalOffline(false)
	h.filesystem.SetOfflineMode(0) // OfflineModeDisabled = 0
	return nil
}

// IsOffline checks if the filesystem is in offline mode
func (h *WorkflowTestHelper) IsOffline() bool {
	if h.filesystem == nil {
		return false
	}
	return h.filesystem.IsOffline()
}

// SynchronizeChanges triggers synchronization of offline changes
func (h *WorkflowTestHelper) SynchronizeChanges() (*SyncResult, error) {
	if h.syncManager == nil {
		return nil, fmt.Errorf("sync manager not available")
	}

	ctx := context.Background()
	return h.syncManager.ProcessOfflineChangesWithRetry(ctx)
}

// ExecuteWorkflow executes all workflow steps
func (h *WorkflowTestHelper) ExecuteWorkflow() *WorkflowResult {
	startTime := time.Now()
	result := &WorkflowResult{
		Success:       true,
		StepsExecuted: 0,
		TotalSteps:    len(h.workflowSteps),
		Duration:      0,
	}

	for i, step := range h.workflowSteps {
		stepStartTime := time.Now()

		// Execute the step action with timeout
		ctx, cancel := context.WithTimeout(context.Background(), step.Timeout)

		stepDone := make(chan error, 1)
		go func() {
			if step.Action != nil {
				stepDone <- step.Action(h)
			} else {
				stepDone <- nil
			}
		}()

		var stepErr error
		select {
		case stepErr = <-stepDone:
		case <-ctx.Done():
			stepErr = fmt.Errorf("step '%s' timed out after %v", step.Name, step.Timeout)
		}
		cancel()

		if stepErr != nil {
			result.Success = false
			result.FailedStep = step.Name
			result.Error = stepErr
			result.Duration = time.Since(startTime)
			return result
		}

		// Execute the step verification if provided
		if step.Verify != nil {
			verifyCtx, verifyCancel := context.WithTimeout(context.Background(), step.Timeout)

			verifyDone := make(chan error, 1)
			go func() {
				verifyDone <- step.Verify(h)
			}()

			select {
			case stepErr = <-verifyDone:
			case <-verifyCtx.Done():
				stepErr = fmt.Errorf("verification for step '%s' timed out after %v", step.Name, step.Timeout)
			}
			verifyCancel()

			if stepErr != nil {
				result.Success = false
				result.FailedStep = step.Name + " (verification)"
				result.Error = stepErr
				result.Duration = time.Since(startTime)
				return result
			}
		}

		result.StepsExecuted = i + 1

		// Log step completion
		stepDuration := time.Since(stepStartTime)
		h.t.Logf("Workflow step '%s' completed in %v", step.Name, stepDuration)
	}

	result.Duration = time.Since(startTime)
	return result
}

// Cleanup performs cleanup operations
func (h *WorkflowTestHelper) Cleanup() error {
	var lastErr error

	// Run cleanup functions in reverse order
	for i := len(h.cleanup) - 1; i >= 0; i-- {
		if err := h.cleanup[i](); err != nil {
			lastErr = err
		}
	}

	return lastErr
}

// CreateStandardWorkflowSteps creates a set of standard workflow steps for common scenarios
func (h *WorkflowTestHelper) CreateStandardWorkflowSteps() {
	// Step 1: Verify initial mount
	h.AddWorkflowStep(WorkflowStep{
		Name:        "verify_initial_mount",
		Description: "Verify filesystem is initially mounted and accessible",
		Action: func(helper *WorkflowTestHelper) error {
			if !helper.mountHelper.IsMounted() {
				return fmt.Errorf("filesystem should be mounted")
			}
			return nil
		},
		Verify: func(helper *WorkflowTestHelper) error {
			_, err := os.ReadDir(helper.GetMountPoint())
			return err
		},
	})

	// Step 2: Create initial files
	h.AddWorkflowStep(WorkflowStep{
		Name:        "create_initial_files",
		Description: "Create initial test files",
		Action: func(helper *WorkflowTestHelper) error {
			files := map[string]string{
				"workflow_test1.txt":        "Initial content 1",
				"workflow_test2.txt":        "Initial content 2",
				"subdir/workflow_test3.txt": "Initial content 3",
			}

			for fileName, content := range files {
				if err := helper.CreateFile(fileName, []byte(content)); err != nil {
					return fmt.Errorf("failed to create file %s: %w", fileName, err)
				}
			}

			helper.SetWorkflowData("initial_files", files)
			return nil
		},
		Verify: func(helper *WorkflowTestHelper) error {
			files := helper.GetWorkflowData("initial_files").(map[string]string)
			for fileName := range files {
				if !helper.FileExists(fileName) {
					return fmt.Errorf("file %s should exist", fileName)
				}
			}
			return nil
		},
	})

	// Step 3: Go offline
	h.AddWorkflowStep(WorkflowStep{
		Name:        "go_offline",
		Description: "Simulate going offline",
		Action: func(helper *WorkflowTestHelper) error {
			return helper.GoOffline()
		},
		Verify: func(helper *WorkflowTestHelper) error {
			if !helper.IsOffline() {
				return fmt.Errorf("filesystem should be offline")
			}
			return nil
		},
	})

	// Step 4: Make offline changes
	h.AddWorkflowStep(WorkflowStep{
		Name:        "make_offline_changes",
		Description: "Make changes while offline",
		Action: func(helper *WorkflowTestHelper) error {
			offlineFiles := map[string]string{
				"offline_file1.txt": "Offline content 1",
				"offline_file2.txt": "Offline content 2",
			}

			for fileName, content := range offlineFiles {
				if err := helper.CreateFile(fileName, []byte(content)); err != nil {
					return fmt.Errorf("failed to create offline file %s: %w", fileName, err)
				}
			}

			helper.SetWorkflowData("offline_files", offlineFiles)
			return nil
		},
		Verify: func(helper *WorkflowTestHelper) error {
			files := helper.GetWorkflowData("offline_files").(map[string]string)
			for fileName := range files {
				if !helper.FileExists(fileName) {
					return fmt.Errorf("offline file %s should exist", fileName)
				}
			}
			return nil
		},
	})

	// Step 5: Go back online
	h.AddWorkflowStep(WorkflowStep{
		Name:        "go_online",
		Description: "Simulate going back online",
		Action: func(helper *WorkflowTestHelper) error {
			return helper.GoOnline()
		},
		Verify: func(helper *WorkflowTestHelper) error {
			if helper.IsOffline() {
				return fmt.Errorf("filesystem should be online")
			}
			return nil
		},
	})

	// Step 6: Synchronize changes
	h.AddWorkflowStep(WorkflowStep{
		Name:        "synchronize_changes",
		Description: "Synchronize offline changes",
		Action: func(helper *WorkflowTestHelper) error {
			result, err := helper.SynchronizeChanges()
			if err != nil {
				return err
			}
			helper.SetWorkflowData("sync_result", result)
			return nil
		},
		Verify: func(helper *WorkflowTestHelper) error {
			result := helper.GetWorkflowData("sync_result").(*SyncResult)
			if result.ProcessedChanges == 0 {
				return fmt.Errorf("should have processed some changes")
			}
			return nil
		},
	})
}

// SetupWorkflowTestFixtureWithFactory creates a test fixture for workflow testing with a custom factory
func SetupWorkflowTestFixtureWithFactory(_ *testing.T, fixtureName string, factory WorkflowFilesystemFactory) *framework.UnitTestFixture {
	return framework.NewUnitTestFixture(fixtureName).
		WithSetup(func(t *testing.T) (interface{}, error) {
			helper := NewWorkflowTestHelper(t)
			if err := helper.SetupWorkflowWithFactory(factory); err != nil {
				return nil, err
			}
			return helper, nil
		}).
		WithTeardown(func(_ *testing.T, fixture interface{}) error {
			helper := fixture.(*WorkflowTestHelper)
			return helper.Cleanup()
		})
}
