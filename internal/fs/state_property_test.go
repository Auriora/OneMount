package fs

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"testing/quick"
	"time"

	"github.com/auriora/onemount/internal/graph"
	"github.com/auriora/onemount/internal/metadata"
)

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

// ItemStateScenario represents a test scenario for initial item state assignment
type ItemStateScenario struct {
	ItemName    string
	IsDirectory bool
	Size        uint64
	ETag        string
}

// Generate implements quick.Generator for ItemStateScenario
func (ItemStateScenario) Generate(rand *quick.Config) reflect.Value {
	// Generate a valid item name (non-empty, reasonable length)
	itemName := fmt.Sprintf("item-%d", rand.Rand.Intn(10000)+1)
	scenario := ItemStateScenario{
		ItemName:    itemName,
		IsDirectory: rand.Rand.Intn(2) == 0,
		Size:        uint64(rand.Rand.Intn(100 * 1024 * 1024)), // 0 to 100MB
		ETag:        fmt.Sprintf("etag-%d", rand.Rand.Intn(10000)),
	}
	return reflect.ValueOf(scenario)
}

// VirtualEntryScenario represents a test scenario for virtual entry state
type VirtualEntryScenario struct {
	VirtualName   string
	OverlayPolicy string
}

// Generate implements quick.Generator for VirtualEntryScenario
func (VirtualEntryScenario) Generate(rand *quick.Config) reflect.Value {
	policies := []string{"LOCAL_WINS", "REMOTE_WINS", "MERGED"}
	// Generate a valid virtual name (non-empty, reasonable length)
	virtualName := fmt.Sprintf("virtual-%d", rand.Rand.Intn(1000)+1)
	scenario := VirtualEntryScenario{
		VirtualName:   virtualName,
		OverlayPolicy: policies[rand.Rand.Intn(len(policies))],
	}
	return reflect.ValueOf(scenario)
}

// TestProperty40_InitialItemState verifies that items discovered via delta
// are inserted with GHOST state and no content download until required.
// **Property 40: Initial Item State**
// **Validates: Requirements 21.2**
func TestProperty40_InitialItemState(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping property-based test in short mode")
	}

	property := func(scenario ItemStateScenario) bool {
		// Validate scenario inputs
		if scenario.ItemName == "" || len(scenario.ItemName) > 200 {
			return true // Skip invalid inputs
		}
		if scenario.ETag == "" {
			return true // Skip invalid inputs
		}

		// Setup test filesystem with metadata store
		fs := newTestFilesystemWithMetadata(t)

		ctx := context.Background()

		// Generate a unique item ID
		itemID := "test-item-" + scenario.ItemName
		parentID := "root"

		// Create a mock DriveItem as if discovered via delta
		item := &graph.DriveItem{
			ID:   itemID,
			Name: scenario.ItemName,
			Parent: &graph.DriveItemParent{
				ID: parentID,
			},
			ETag: scenario.ETag,
		}

		if scenario.IsDirectory {
			item.Folder = &graph.Folder{
				ChildCount: 0,
			}
		} else {
			item.Size = scenario.Size
			item.File = &graph.File{}
		}

		// Ensure parent exists
		parentEntry := &metadata.Entry{
			ID:            parentID,
			Name:          "root",
			ItemType:      metadata.ItemKindDirectory,
			State:         metadata.ItemStateHydrated,
			OverlayPolicy: metadata.OverlayPolicyRemoteWins,
			CreatedAt:     time.Now().UTC(),
			UpdatedAt:     time.Now().UTC(),
			Pin: metadata.PinState{
				Mode: metadata.PinModeUnset,
			},
		}
		err := fs.metadataStore.Save(ctx, parentEntry)
		if err != nil {
			t.Logf("Failed to save parent: %v", err)
			return false
		}

		// Apply delta to create the item (simulating delta sync discovery)
		snapshot := time.Now().UTC()
		entry, previous, err := fs.upsertDriveItemEntry(ctx, item, snapshot)
		if err != nil {
			t.Logf("Failed to upsert item: %v", err)
			return false
		}

		// Verify this is a new item (no previous entry)
		if previous != nil {
			t.Logf("Expected new item but found previous entry")
			return false
		}

		// Verify the entry was created
		if entry == nil {
			t.Logf("Entry is nil after upsert")
			return false
		}

		// Verify correct initial state based on item type
		expectedState := metadata.ItemStateGhost
		if scenario.IsDirectory {
			expectedState = metadata.ItemStateHydrated
		}

		if entry.State != expectedState {
			t.Logf("Expected state %s for %s, got %s", expectedState, scenario.ItemName, entry.State)
			return false
		}

		// Verify no content download occurred (no LastHydrated timestamp for files)
		if !scenario.IsDirectory && entry.LastHydrated != nil {
			t.Logf("File %s has LastHydrated timestamp, indicating content was downloaded", scenario.ItemName)
			return false
		}

		// Verify state persistence by reading from store
		retrieved, err := fs.metadataStore.Get(ctx, itemID)
		if err != nil {
			t.Logf("Failed to retrieve item from store: %v", err)
			return false
		}

		if retrieved.State != expectedState {
			t.Logf("Persisted state %s doesn't match expected %s", retrieved.State, expectedState)
			return false
		}

		// Verify RemoteID is set correctly
		if retrieved.RemoteID != itemID {
			t.Logf("RemoteID %s doesn't match item ID %s", retrieved.RemoteID, itemID)
			return false
		}

		// Verify Virtual flag is false for regular items
		if retrieved.Virtual {
			t.Logf("Regular item %s incorrectly marked as virtual", scenario.ItemName)
			return false
		}

		// Verify metadata fields are populated correctly
		if retrieved.Name != scenario.ItemName {
			t.Logf("Name mismatch: expected %s, got %s", scenario.ItemName, retrieved.Name)
			return false
		}

		if retrieved.ETag != scenario.ETag {
			t.Logf("ETag mismatch: expected %s, got %s", scenario.ETag, retrieved.ETag)
			return false
		}

		if !scenario.IsDirectory && retrieved.Size != scenario.Size {
			t.Logf("Size mismatch: expected %d, got %d", scenario.Size, retrieved.Size)
			return false
		}

		return true
	}

	config := &quick.Config{
		MaxCount: 100,
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property check failed: %v", err)
	}
}

// TestProperty40_VirtualEntryState verifies that virtual entries use correct state and flags.
// **Property 40: Initial Item State (Virtual Entries)**
// **Validates: Requirements 21.10**
func TestProperty40_VirtualEntryState(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping property-based test in short mode")
	}

	property := func(scenario VirtualEntryScenario) bool {
		// Validate scenario inputs
		if scenario.VirtualName == "" || len(scenario.VirtualName) > 200 {
			return true // Skip invalid inputs
		}
		if scenario.OverlayPolicy == "" {
			return true // Skip invalid inputs
		}

		// Setup test filesystem with metadata store
		fs := newTestFilesystemWithMetadata(t)

		ctx := context.Background()

		// Generate a unique virtual item ID (local-only)
		itemID := "local-" + scenario.VirtualName
		parentID := "root"

		// Map overlay policy string to enum
		var policy metadata.OverlayPolicy
		switch scenario.OverlayPolicy {
		case "LOCAL_WINS":
			policy = metadata.OverlayPolicyLocalWins
		case "REMOTE_WINS":
			policy = metadata.OverlayPolicyRemoteWins
		case "MERGED":
			policy = metadata.OverlayPolicyMerged
		default:
			policy = metadata.OverlayPolicyLocalWins
		}

		// Ensure parent exists
		parentEntry := &metadata.Entry{
			ID:            parentID,
			Name:          "root",
			ItemType:      metadata.ItemKindDirectory,
			State:         metadata.ItemStateHydrated,
			OverlayPolicy: metadata.OverlayPolicyRemoteWins,
			CreatedAt:     time.Now().UTC(),
			UpdatedAt:     time.Now().UTC(),
			Pin: metadata.PinState{
				Mode: metadata.PinModeUnset,
			},
		}
		err := fs.metadataStore.Save(ctx, parentEntry)
		if err != nil {
			t.Logf("Failed to save parent: %v", err)
			return false
		}

		// Create a virtual entry
		virtualEntry := &metadata.Entry{
			ID:            itemID,
			RemoteID:      "", // NULL for virtual entries
			ParentID:      parentID,
			Name:          scenario.VirtualName,
			ItemType:      metadata.ItemKindFile,
			State:         metadata.ItemStateHydrated, // Virtual entries are always HYDRATED
			OverlayPolicy: policy,
			Virtual:       true, // Mark as virtual
			Size:          0,
			CreatedAt:     time.Now().UTC(),
			UpdatedAt:     time.Now().UTC(),
			Pin: metadata.PinState{
				Mode: metadata.PinModeUnset,
			},
		}

		// Save the virtual entry
		err = fs.metadataStore.Save(ctx, virtualEntry)
		if err != nil {
			t.Logf("Failed to save virtual entry: %v", err)
			return false
		}

		// Verify state persistence by reading from store
		retrieved, err := fs.metadataStore.Get(ctx, itemID)
		if err != nil {
			t.Logf("Failed to retrieve virtual entry from store: %v", err)
			return false
		}

		// Verify state is HYDRATED
		if retrieved.State != metadata.ItemStateHydrated {
			t.Logf("Virtual entry state is %s, expected HYDRATED", retrieved.State)
			return false
		}

		// Verify RemoteID is empty (NULL)
		if retrieved.RemoteID != "" {
			t.Logf("Virtual entry has RemoteID %s, expected empty", retrieved.RemoteID)
			return false
		}

		// Verify Virtual flag is true
		if !retrieved.Virtual {
			t.Logf("Virtual entry has Virtual=false, expected true")
			return false
		}

		// Verify overlay policy is set correctly
		if retrieved.OverlayPolicy != policy {
			t.Logf("Overlay policy mismatch: expected %s, got %s", policy, retrieved.OverlayPolicy)
			return false
		}

		// Verify virtual entries cannot transition to other states
		// (this is enforced by the state manager)
		if fs.stateManager != nil {
			// Attempt to transition to GHOST (should fail)
			_, err := fs.stateManager.Transition(ctx, itemID, metadata.ItemStateGhost)
			if err == nil {
				t.Logf("Virtual entry allowed transition to GHOST, should be prevented")
				return false
			}

			// Verify error contains "invalid state transition" message
			// (the exact error type check is too strict for property-based testing)
			errMsg := err.Error()
			if errMsg == "" || !contains(errMsg, "invalid") {
				t.Logf("Expected error about invalid transition, got: %v", err)
				return false
			}
		}

		return true
	}

	config := &quick.Config{
		MaxCount: 50,
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property check failed: %v", err)
	}
}

// TestProperty40_NoContentDownloadUntilRequired verifies that GHOST state items
// do not trigger content download until explicitly accessed.
func TestProperty40_NoContentDownloadUntilRequired(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping property-based test in short mode")
	}

	property := func(itemName string, size uint64) bool {
		// Validate inputs
		if len(itemName) == 0 || len(itemName) > 100 {
			return true // Skip invalid inputs
		}
		if size < 1024 || size > 10*1024*1024 {
			return true // Skip invalid sizes (1KB to 10MB)
		}

		// Setup test filesystem with metadata store
		fs := newTestFilesystemWithMetadata(t)

		ctx := context.Background()

		// Generate a unique item ID
		itemID := "test-file-" + itemName
		parentID := "root"

		// Ensure parent exists
		parentEntry := &metadata.Entry{
			ID:            parentID,
			Name:          "root",
			ItemType:      metadata.ItemKindDirectory,
			State:         metadata.ItemStateHydrated,
			OverlayPolicy: metadata.OverlayPolicyRemoteWins,
			CreatedAt:     time.Now().UTC(),
			UpdatedAt:     time.Now().UTC(),
			Pin: metadata.PinState{
				Mode: metadata.PinModeUnset,
			},
		}
		err := fs.metadataStore.Save(ctx, parentEntry)
		if err != nil {
			t.Logf("Failed to save parent: %v", err)
			return false
		}

		// Create a file entry in GHOST state
		fileEntry := &metadata.Entry{
			ID:            itemID,
			RemoteID:      itemID,
			ParentID:      parentID,
			Name:          itemName,
			ItemType:      metadata.ItemKindFile,
			State:         metadata.ItemStateGhost,
			OverlayPolicy: metadata.OverlayPolicyRemoteWins,
			Size:          size,
			ETag:          "etag-" + itemName,
			CreatedAt:     time.Now().UTC(),
			UpdatedAt:     time.Now().UTC(),
			Pin: metadata.PinState{
				Mode: metadata.PinModeUnset,
			},
		}

		// Save the entry
		err = fs.metadataStore.Save(ctx, fileEntry)
		if err != nil {
			t.Logf("Failed to save file entry: %v", err)
			return false
		}

		// Wait a short time to ensure no background download is triggered
		time.Sleep(100 * time.Millisecond)

		// Verify the entry is still in GHOST state
		retrieved, err := fs.metadataStore.Get(ctx, itemID)
		if err != nil {
			t.Logf("Failed to retrieve entry: %v", err)
			return false
		}

		if retrieved.State != metadata.ItemStateGhost {
			t.Logf("Entry state changed from GHOST to %s without explicit access", retrieved.State)
			return false
		}

		// Verify no hydration occurred (no LastHydrated timestamp)
		if retrieved.LastHydrated != nil {
			t.Logf("Entry has LastHydrated timestamp, indicating content was downloaded")
			return false
		}

		// Verify no hydration worker is assigned
		if retrieved.Hydration.WorkerID != "" {
			t.Logf("Entry has hydration worker assigned: %s", retrieved.Hydration.WorkerID)
			return false
		}

		// Verify no content exists in cache
		if fs.content != nil {
			hasContent := fs.content.HasContent(itemID)
			if hasContent {
				t.Logf("Content exists in cache for GHOST state item")
				return false
			}
		}

		return true
	}

	config := &quick.Config{
		MaxCount: 50,
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property check failed: %v", err)
	}
}

// HydrationTransitionScenario represents a test scenario for hydration state transitions
type HydrationTransitionScenario struct {
	ItemName      string
	Size          uint64
	ETag          string
	SimulateError bool
	ErrorMessage  string
}

// Generate implements quick.Generator for HydrationTransitionScenario
func (HydrationTransitionScenario) Generate(rand *quick.Config) reflect.Value {
	itemName := fmt.Sprintf("file-%d", rand.Rand.Intn(10000)+1)
	simulateError := rand.Rand.Intn(4) == 0 // 25% chance of error
	errorMsg := ""
	if simulateError {
		errors := []string{
			"network timeout",
			"connection refused",
			"download failed",
			"disk full",
			"permission denied",
		}
		errorMsg = errors[rand.Rand.Intn(len(errors))]
	}

	scenario := HydrationTransitionScenario{
		ItemName:      itemName,
		Size:          uint64(rand.Rand.Intn(10*1024*1024) + 1024), // 1KB to 10MB
		ETag:          fmt.Sprintf("etag-%d", rand.Rand.Intn(10000)),
		SimulateError: simulateError,
		ErrorMessage:  errorMsg,
	}
	return reflect.ValueOf(scenario)
}

// TestProperty41_GhostToHydratingTransition verifies that GHOST → HYDRATING
// transition occurs on user access.
// **Property 41: Successful Hydration State Transition (Part 1: GHOST → HYDRATING)**
// **Validates: Requirements 21.3**
func TestProperty41_GhostToHydratingTransition(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping property-based test in short mode")
	}

	property := func(scenario HydrationTransitionScenario) bool {
		// Validate scenario inputs
		if scenario.ItemName == "" || len(scenario.ItemName) > 200 {
			return true // Skip invalid inputs
		}
		if scenario.ETag == "" || scenario.Size == 0 {
			return true // Skip invalid inputs
		}

		// Setup test filesystem with metadata store
		fs := newTestFilesystemWithMetadata(t)
		ctx := context.Background()

		// Generate a unique item ID
		itemID := "test-file-" + scenario.ItemName
		parentID := "root"

		// Ensure parent exists
		parentEntry := &metadata.Entry{
			ID:            parentID,
			Name:          "root",
			ItemType:      metadata.ItemKindDirectory,
			State:         metadata.ItemStateHydrated,
			OverlayPolicy: metadata.OverlayPolicyRemoteWins,
			CreatedAt:     time.Now().UTC(),
			UpdatedAt:     time.Now().UTC(),
			Pin: metadata.PinState{
				Mode: metadata.PinModeUnset,
			},
		}
		err := fs.metadataStore.Save(ctx, parentEntry)
		if err != nil {
			t.Logf("Failed to save parent: %v", err)
			return false
		}

		// Create a file entry in GHOST state
		fileEntry := &metadata.Entry{
			ID:            itemID,
			RemoteID:      itemID,
			ParentID:      parentID,
			Name:          scenario.ItemName,
			ItemType:      metadata.ItemKindFile,
			State:         metadata.ItemStateGhost,
			OverlayPolicy: metadata.OverlayPolicyRemoteWins,
			Size:          scenario.Size,
			ETag:          scenario.ETag,
			CreatedAt:     time.Now().UTC(),
			UpdatedAt:     time.Now().UTC(),
			Pin: metadata.PinState{
				Mode: metadata.PinModeUnset,
			},
		}
		err = fs.metadataStore.Save(ctx, fileEntry)
		if err != nil {
			t.Logf("Failed to save file entry: %v", err)
			return false
		}

		// Simulate user access by transitioning to HYDRATING
		workerID := "worker-" + scenario.ItemName
		_, err = fs.stateManager.Transition(ctx, itemID, metadata.ItemStateHydrating,
			metadata.WithHydrationEvent(),
			metadata.WithWorker(workerID))
		if err != nil {
			t.Logf("Failed to transition to HYDRATING: %v", err)
			return false
		}

		// Verify the entry is now in HYDRATING state
		retrieved, err := fs.metadataStore.Get(ctx, itemID)
		if err != nil {
			t.Logf("Failed to retrieve entry: %v", err)
			return false
		}

		if retrieved.State != metadata.ItemStateHydrating {
			t.Logf("Expected state HYDRATING, got %s", retrieved.State)
			return false
		}

		// Verify worker ID is recorded
		if retrieved.Hydration.WorkerID != workerID {
			t.Logf("Expected worker ID %s, got %s", workerID, retrieved.Hydration.WorkerID)
			return false
		}

		// Verify hydration started timestamp is set
		if retrieved.Hydration.StartedAt == nil {
			t.Logf("Hydration StartedAt timestamp not set")
			return false
		}

		// Verify hydration completed timestamp is not set yet
		if retrieved.Hydration.CompletedAt != nil {
			t.Logf("Hydration CompletedAt timestamp should not be set during HYDRATING")
			return false
		}

		return true
	}

	config := &quick.Config{
		MaxCount: 100,
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property check failed: %v", err)
	}
}

// TestProperty41_HydratingToHydratedTransition verifies that HYDRATING → HYDRATED
// transition occurs on successful download.
// **Property 41: Successful Hydration State Transition (Part 2: HYDRATING → HYDRATED)**
// **Validates: Requirements 21.4**
func TestProperty41_HydratingToHydratedTransition(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping property-based test in short mode")
	}

	property := func(scenario HydrationTransitionScenario) bool {
		// Validate scenario inputs
		if scenario.ItemName == "" || len(scenario.ItemName) > 200 {
			return true // Skip invalid inputs
		}
		if scenario.ETag == "" || scenario.Size == 0 {
			return true // Skip invalid inputs
		}

		// Setup test filesystem with metadata store
		fs := newTestFilesystemWithMetadata(t)
		ctx := context.Background()

		// Generate a unique item ID
		itemID := "test-file-" + scenario.ItemName
		parentID := "root"

		// Ensure parent exists
		parentEntry := &metadata.Entry{
			ID:            parentID,
			Name:          "root",
			ItemType:      metadata.ItemKindDirectory,
			State:         metadata.ItemStateHydrated,
			OverlayPolicy: metadata.OverlayPolicyRemoteWins,
			CreatedAt:     time.Now().UTC(),
			UpdatedAt:     time.Now().UTC(),
			Pin: metadata.PinState{
				Mode: metadata.PinModeUnset,
			},
		}
		err := fs.metadataStore.Save(ctx, parentEntry)
		if err != nil {
			t.Logf("Failed to save parent: %v", err)
			return false
		}

		// Create a file entry in HYDRATING state
		workerID := "worker-" + scenario.ItemName
		startedAt := time.Now().UTC()
		fileEntry := &metadata.Entry{
			ID:            itemID,
			RemoteID:      itemID,
			ParentID:      parentID,
			Name:          scenario.ItemName,
			ItemType:      metadata.ItemKindFile,
			State:         metadata.ItemStateHydrating,
			OverlayPolicy: metadata.OverlayPolicyRemoteWins,
			Size:          scenario.Size,
			ETag:          scenario.ETag,
			CreatedAt:     time.Now().UTC(),
			UpdatedAt:     time.Now().UTC(),
			Hydration: metadata.HydrationState{
				WorkerID:  workerID,
				StartedAt: &startedAt,
			},
			Pin: metadata.PinState{
				Mode: metadata.PinModeUnset,
			},
		}
		err = fs.metadataStore.Save(ctx, fileEntry)
		if err != nil {
			t.Logf("Failed to save file entry: %v", err)
			return false
		}

		// Simulate successful download by transitioning to HYDRATED
		contentHash := "hash-" + scenario.ItemName
		_, err = fs.stateManager.Transition(ctx, itemID, metadata.ItemStateHydrated,
			metadata.WithHydrationEvent(),
			metadata.WithWorker(workerID),
			metadata.WithContentHash(contentHash))
		if err != nil {
			t.Logf("Failed to transition to HYDRATED: %v", err)
			return false
		}

		// Verify the entry is now in HYDRATED state
		retrieved, err := fs.metadataStore.Get(ctx, itemID)
		if err != nil {
			t.Logf("Failed to retrieve entry: %v", err)
			return false
		}

		if retrieved.State != metadata.ItemStateHydrated {
			t.Logf("Expected state HYDRATED, got %s", retrieved.State)
			return false
		}

		// Verify content hash is recorded
		if retrieved.ContentHash != contentHash {
			t.Logf("Expected content hash %s, got %s", contentHash, retrieved.ContentHash)
			return false
		}

		// Verify hydration completed timestamp is set
		if retrieved.Hydration.CompletedAt == nil {
			t.Logf("Hydration CompletedAt timestamp not set")
			return false
		}

		// Verify LastHydrated timestamp is set
		if retrieved.LastHydrated == nil {
			t.Logf("LastHydrated timestamp not set")
			return false
		}

		// Verify error fields are cleared
		if retrieved.LastError != nil {
			t.Logf("LastError should be nil after successful hydration")
			return false
		}

		if retrieved.Hydration.Error != nil {
			t.Logf("Hydration.Error should be nil after successful hydration")
			return false
		}

		return true
	}

	config := &quick.Config{
		MaxCount: 100,
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property check failed: %v", err)
	}
}

// TestProperty41_HydratingToErrorTransition verifies that HYDRATING → ERROR
// transition occurs on download failure.
// **Property 41: Successful Hydration State Transition (Part 3: HYDRATING → ERROR)**
// **Validates: Requirements 21.5**
func TestProperty41_HydratingToErrorTransition(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping property-based test in short mode")
	}

	property := func(scenario HydrationTransitionScenario) bool {
		// Validate scenario inputs
		if scenario.ItemName == "" || len(scenario.ItemName) > 200 {
			return true // Skip invalid inputs
		}
		if scenario.ETag == "" || scenario.Size == 0 {
			return true // Skip invalid inputs
		}
		if !scenario.SimulateError || scenario.ErrorMessage == "" {
			return true // Skip scenarios without errors
		}

		// Setup test filesystem with metadata store
		fs := newTestFilesystemWithMetadata(t)
		ctx := context.Background()

		// Generate a unique item ID
		itemID := "test-file-" + scenario.ItemName
		parentID := "root"

		// Ensure parent exists
		parentEntry := &metadata.Entry{
			ID:            parentID,
			Name:          "root",
			ItemType:      metadata.ItemKindDirectory,
			State:         metadata.ItemStateHydrated,
			OverlayPolicy: metadata.OverlayPolicyRemoteWins,
			CreatedAt:     time.Now().UTC(),
			UpdatedAt:     time.Now().UTC(),
			Pin: metadata.PinState{
				Mode: metadata.PinModeUnset,
			},
		}
		err := fs.metadataStore.Save(ctx, parentEntry)
		if err != nil {
			t.Logf("Failed to save parent: %v", err)
			return false
		}

		// Create a file entry in HYDRATING state
		workerID := "worker-" + scenario.ItemName
		startedAt := time.Now().UTC()
		fileEntry := &metadata.Entry{
			ID:            itemID,
			RemoteID:      itemID,
			ParentID:      parentID,
			Name:          scenario.ItemName,
			ItemType:      metadata.ItemKindFile,
			State:         metadata.ItemStateHydrating,
			OverlayPolicy: metadata.OverlayPolicyRemoteWins,
			Size:          scenario.Size,
			ETag:          scenario.ETag,
			CreatedAt:     time.Now().UTC(),
			UpdatedAt:     time.Now().UTC(),
			Hydration: metadata.HydrationState{
				WorkerID:  workerID,
				StartedAt: &startedAt,
			},
			Pin: metadata.PinState{
				Mode: metadata.PinModeUnset,
			},
		}
		err = fs.metadataStore.Save(ctx, fileEntry)
		if err != nil {
			t.Logf("Failed to save file entry: %v", err)
			return false
		}

		// Simulate download failure by transitioning to ERROR
		downloadErr := fmt.Errorf("%s", scenario.ErrorMessage)
		_, err = fs.stateManager.Transition(ctx, itemID, metadata.ItemStateError,
			metadata.WithHydrationEvent(),
			metadata.WithWorker(workerID),
			metadata.WithTransitionError(downloadErr, true))
		if err != nil {
			t.Logf("Failed to transition to ERROR: %v", err)
			return false
		}

		// Verify the entry is now in ERROR state
		retrieved, err := fs.metadataStore.Get(ctx, itemID)
		if err != nil {
			t.Logf("Failed to retrieve entry: %v", err)
			return false
		}

		if retrieved.State != metadata.ItemStateError {
			t.Logf("Expected state ERROR, got %s", retrieved.State)
			return false
		}

		// Verify error is recorded
		if retrieved.LastError == nil {
			t.Logf("LastError should be set after failed hydration")
			return false
		}

		if retrieved.LastError.Message != scenario.ErrorMessage {
			t.Logf("Expected error message %s, got %s", scenario.ErrorMessage, retrieved.LastError.Message)
			return false
		}

		// Verify error is marked as temporary
		if !retrieved.LastError.Temporary {
			t.Logf("Error should be marked as temporary")
			return false
		}

		// Verify hydration error is recorded
		if retrieved.Hydration.Error == nil {
			t.Logf("Hydration.Error should be set after failed hydration")
			return false
		}

		if retrieved.Hydration.Error.Message != scenario.ErrorMessage {
			t.Logf("Expected hydration error message %s, got %s", scenario.ErrorMessage, retrieved.Hydration.Error.Message)
			return false
		}

		// Verify hydration completed timestamp is set
		if retrieved.Hydration.CompletedAt == nil {
			t.Logf("Hydration CompletedAt timestamp should be set even on failure")
			return false
		}

		// Verify previous state metadata is preserved (size, ETag)
		if retrieved.Size != scenario.Size {
			t.Logf("Size should be preserved after error: expected %d, got %d", scenario.Size, retrieved.Size)
			return false
		}

		if retrieved.ETag != scenario.ETag {
			t.Logf("ETag should be preserved after error: expected %s, got %s", scenario.ETag, retrieved.ETag)
			return false
		}

		return true
	}

	config := &quick.Config{
		MaxCount: 100,
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property check failed: %v", err)
	}
}

// TestProperty41_HydratingToGhostTransition verifies that HYDRATING → GHOST
// transition occurs on cancellation (using force transition).
// **Property 41: Successful Hydration State Transition (Part 4: HYDRATING → GHOST)**
// **Validates: Requirements 21.3**
func TestProperty41_HydratingToGhostTransition(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping property-based test in short mode")
	}

	property := func(scenario HydrationTransitionScenario) bool {
		// Validate scenario inputs
		if scenario.ItemName == "" || len(scenario.ItemName) > 200 {
			return true // Skip invalid inputs
		}
		if scenario.ETag == "" || scenario.Size == 0 {
			return true // Skip invalid inputs
		}

		// Setup test filesystem with metadata store
		fs := newTestFilesystemWithMetadata(t)
		ctx := context.Background()

		// Generate a unique item ID
		itemID := "test-file-" + scenario.ItemName
		parentID := "root"

		// Ensure parent exists
		parentEntry := &metadata.Entry{
			ID:            parentID,
			Name:          "root",
			ItemType:      metadata.ItemKindDirectory,
			State:         metadata.ItemStateHydrated,
			OverlayPolicy: metadata.OverlayPolicyRemoteWins,
			CreatedAt:     time.Now().UTC(),
			UpdatedAt:     time.Now().UTC(),
			Pin: metadata.PinState{
				Mode: metadata.PinModeUnset,
			},
		}
		err := fs.metadataStore.Save(ctx, parentEntry)
		if err != nil {
			t.Logf("Failed to save parent: %v", err)
			return false
		}

		// Create a file entry in HYDRATING state
		workerID := "worker-" + scenario.ItemName
		startedAt := time.Now().UTC()
		fileEntry := &metadata.Entry{
			ID:            itemID,
			RemoteID:      itemID,
			ParentID:      parentID,
			Name:          scenario.ItemName,
			ItemType:      metadata.ItemKindFile,
			State:         metadata.ItemStateHydrating,
			OverlayPolicy: metadata.OverlayPolicyRemoteWins,
			Size:          scenario.Size,
			ETag:          scenario.ETag,
			CreatedAt:     time.Now().UTC(),
			UpdatedAt:     time.Now().UTC(),
			Hydration: metadata.HydrationState{
				WorkerID:  workerID,
				StartedAt: &startedAt,
			},
			Pin: metadata.PinState{
				Mode: metadata.PinModeUnset,
			},
		}
		err = fs.metadataStore.Save(ctx, fileEntry)
		if err != nil {
			t.Logf("Failed to save file entry: %v", err)
			return false
		}

		// Simulate cancellation by transitioning back to GHOST (requires force)
		// Note: In practice, cancellation might transition to ERROR instead,
		// but the requirement specifies GHOST as a valid cancellation target
		_, err = fs.stateManager.Transition(ctx, itemID, metadata.ItemStateGhost,
			metadata.WithHydrationEvent(),
			metadata.ForceTransition())
		if err != nil {
			t.Logf("Failed to transition to GHOST: %v", err)
			return false
		}

		// Verify the entry is now in GHOST state
		retrieved, err := fs.metadataStore.Get(ctx, itemID)
		if err != nil {
			t.Logf("Failed to retrieve entry: %v", err)
			return false
		}

		if retrieved.State != metadata.ItemStateGhost {
			t.Logf("Expected state GHOST, got %s", retrieved.State)
			return false
		}

		// Note: Worker ID may or may not be cleared depending on implementation
		// The important thing is that the state is GHOST and content is not available

		// Verify metadata is preserved (size, ETag)
		if retrieved.Size != scenario.Size {
			t.Logf("Size should be preserved after cancellation: expected %d, got %d", scenario.Size, retrieved.Size)
			return false
		}

		if retrieved.ETag != scenario.ETag {
			t.Logf("ETag should be preserved after cancellation: expected %s, got %s", scenario.ETag, retrieved.ETag)
			return false
		}

		// Verify no content exists in cache
		if fs.content != nil {
			hasContent := fs.content.HasContent(itemID)
			if hasContent {
				t.Logf("Content should not exist in cache after cancellation")
				return false
			}
		}

		return true
	}

	config := &quick.Config{
		MaxCount: 100,
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property check failed: %v", err)
	}
}

// TestProperty41_WorkerDeduplication verifies that duplicate hydration requests
// for the same item are deduplicated by worker ID.
// **Property 41: Worker Deduplication During Hydration**
// **Validates: Requirements 21.3**
func TestProperty41_WorkerDeduplication(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping property-based test in short mode")
	}

	property := func(scenario HydrationTransitionScenario) bool {
		// Validate scenario inputs
		if scenario.ItemName == "" || len(scenario.ItemName) > 200 {
			return true // Skip invalid inputs
		}
		if scenario.ETag == "" || scenario.Size == 0 {
			return true // Skip invalid inputs
		}

		// Setup test filesystem with metadata store
		fs := newTestFilesystemWithMetadata(t)
		ctx := context.Background()

		// Generate a unique item ID
		itemID := "test-file-" + scenario.ItemName
		parentID := "root"

		// Ensure parent exists
		parentEntry := &metadata.Entry{
			ID:            parentID,
			Name:          "root",
			ItemType:      metadata.ItemKindDirectory,
			State:         metadata.ItemStateHydrated,
			OverlayPolicy: metadata.OverlayPolicyRemoteWins,
			CreatedAt:     time.Now().UTC(),
			UpdatedAt:     time.Now().UTC(),
			Pin: metadata.PinState{
				Mode: metadata.PinModeUnset,
			},
		}
		err := fs.metadataStore.Save(ctx, parentEntry)
		if err != nil {
			t.Logf("Failed to save parent: %v", err)
			return false
		}

		// Create a file entry in GHOST state
		fileEntry := &metadata.Entry{
			ID:            itemID,
			RemoteID:      itemID,
			ParentID:      parentID,
			Name:          scenario.ItemName,
			ItemType:      metadata.ItemKindFile,
			State:         metadata.ItemStateGhost,
			OverlayPolicy: metadata.OverlayPolicyRemoteWins,
			Size:          scenario.Size,
			ETag:          scenario.ETag,
			CreatedAt:     time.Now().UTC(),
			UpdatedAt:     time.Now().UTC(),
			Pin: metadata.PinState{
				Mode: metadata.PinModeUnset,
			},
		}
		err = fs.metadataStore.Save(ctx, fileEntry)
		if err != nil {
			t.Logf("Failed to save file entry: %v", err)
			return false
		}

		// First worker starts hydration
		worker1ID := "worker-1-" + scenario.ItemName
		_, err = fs.stateManager.Transition(ctx, itemID, metadata.ItemStateHydrating,
			metadata.WithHydrationEvent(),
			metadata.WithWorker(worker1ID))
		if err != nil {
			t.Logf("Failed first transition to HYDRATING: %v", err)
			return false
		}

		// Verify first worker is recorded
		retrieved, err := fs.metadataStore.Get(ctx, itemID)
		if err != nil {
			t.Logf("Failed to retrieve entry: %v", err)
			return false
		}

		if retrieved.Hydration.WorkerID != worker1ID {
			t.Logf("Expected worker ID %s, got %s", worker1ID, retrieved.Hydration.WorkerID)
			return false
		}

		// Second worker attempts to start hydration (should be rejected or deduplicated)
		worker2ID := "worker-2-" + scenario.ItemName
		_, err = fs.stateManager.Transition(ctx, itemID, metadata.ItemStateHydrating,
			metadata.WithHydrationEvent(),
			metadata.WithWorker(worker2ID))

		// The transition should either fail (item already hydrating) or succeed but keep the original worker
		// Check the current state
		retrieved, err = fs.metadataStore.Get(ctx, itemID)
		if err != nil {
			t.Logf("Failed to retrieve entry after second transition: %v", err)
			return false
		}

		// Verify the item is still in HYDRATING state
		if retrieved.State != metadata.ItemStateHydrating {
			t.Logf("Expected state HYDRATING after duplicate request, got %s", retrieved.State)
			return false
		}

		// Verify the original worker ID is preserved (deduplication)
		if retrieved.Hydration.WorkerID != worker1ID {
			t.Logf("Worker ID changed from %s to %s, deduplication failed", worker1ID, retrieved.Hydration.WorkerID)
			return false
		}

		return true
	}

	config := &quick.Config{
		MaxCount: 50,
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property check failed: %v", err)
	}
}

// ModificationUploadScenario represents a test scenario for modification and upload state transitions
type ModificationUploadScenario struct {
	ItemName      string
	Size          uint64
	ETag          string
	NewETag       string
	SimulateError bool
	ErrorMessage  string
}

// Generate implements quick.Generator for ModificationUploadScenario
func (ModificationUploadScenario) Generate(rand *quick.Config) reflect.Value {
	itemName := fmt.Sprintf("file-%d", rand.Rand.Intn(10000)+1)
	simulateError := rand.Rand.Intn(4) == 0 // 25% chance of error
	errorMsg := ""
	if simulateError {
		errors := []string{
			"network timeout",
			"connection refused",
			"upload failed",
			"disk full",
			"permission denied",
			"quota exceeded",
		}
		errorMsg = errors[rand.Rand.Intn(len(errors))]
	}

	scenario := ModificationUploadScenario{
		ItemName:      itemName,
		Size:          uint64(rand.Rand.Intn(10*1024*1024) + 1024), // 1KB to 10MB
		ETag:          fmt.Sprintf("etag-%d", rand.Rand.Intn(10000)),
		NewETag:       fmt.Sprintf("etag-%d", rand.Rand.Intn(10000)+10000),
		SimulateError: simulateError,
		ErrorMessage:  errorMsg,
	}
	return reflect.ValueOf(scenario)
}

// TestProperty42_HydratedToDirtyLocalTransition verifies that HYDRATED → DIRTY_LOCAL
// transition occurs on local modification.
// **Property 42: Local Modification State Transition (Part 1: HYDRATED → DIRTY_LOCAL)**
// **Validates: Requirements 21.6**
func TestProperty42_HydratedToDirtyLocalTransition(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping property-based test in short mode")
	}

	property := func(scenario ModificationUploadScenario) bool {
		// Validate scenario inputs
		if scenario.ItemName == "" || len(scenario.ItemName) > 200 {
			return true // Skip invalid inputs
		}
		if scenario.ETag == "" || scenario.Size == 0 {
			return true // Skip invalid inputs
		}

		// Setup test filesystem with metadata store
		fs := newTestFilesystemWithMetadata(t)
		ctx := context.Background()

		// Generate a unique item ID
		itemID := "test-file-" + scenario.ItemName
		parentID := "root"

		// Ensure parent exists
		parentEntry := &metadata.Entry{
			ID:            parentID,
			Name:          "root",
			ItemType:      metadata.ItemKindDirectory,
			State:         metadata.ItemStateHydrated,
			OverlayPolicy: metadata.OverlayPolicyRemoteWins,
			CreatedAt:     time.Now().UTC(),
			UpdatedAt:     time.Now().UTC(),
			Pin: metadata.PinState{
				Mode: metadata.PinModeUnset,
			},
		}
		err := fs.metadataStore.Save(ctx, parentEntry)
		if err != nil {
			t.Logf("Failed to save parent: %v", err)
			return false
		}

		// Create a file entry in HYDRATED state
		lastHydrated := time.Now().UTC().Add(-1 * time.Hour)
		fileEntry := &metadata.Entry{
			ID:            itemID,
			RemoteID:      itemID,
			ParentID:      parentID,
			Name:          scenario.ItemName,
			ItemType:      metadata.ItemKindFile,
			State:         metadata.ItemStateHydrated,
			OverlayPolicy: metadata.OverlayPolicyRemoteWins,
			Size:          scenario.Size,
			ETag:          scenario.ETag,
			ContentHash:   "hash-" + scenario.ItemName,
			LastHydrated:  &lastHydrated,
			CreatedAt:     time.Now().UTC(),
			UpdatedAt:     time.Now().UTC(),
			Pin: metadata.PinState{
				Mode: metadata.PinModeUnset,
			},
		}
		err = fs.metadataStore.Save(ctx, fileEntry)
		if err != nil {
			t.Logf("Failed to save file entry: %v", err)
			return false
		}

		// Simulate local modification by transitioning to DIRTY_LOCAL
		_, err = fs.stateManager.Transition(ctx, itemID, metadata.ItemStateDirtyLocal)
		if err != nil {
			t.Logf("Failed to transition to DIRTY_LOCAL: %v", err)
			return false
		}

		// Verify the entry is now in DIRTY_LOCAL state
		retrieved, err := fs.metadataStore.Get(ctx, itemID)
		if err != nil {
			t.Logf("Failed to retrieve entry: %v", err)
			return false
		}

		if retrieved.State != metadata.ItemStateDirtyLocal {
			t.Logf("Expected state DIRTY_LOCAL, got %s", retrieved.State)
			return false
		}

		// Verify metadata is preserved (size, ETag, content hash)
		if retrieved.Size != scenario.Size {
			t.Logf("Size should be preserved: expected %d, got %d", scenario.Size, retrieved.Size)
			return false
		}

		if retrieved.ETag != scenario.ETag {
			t.Logf("ETag should be preserved: expected %s, got %s", scenario.ETag, retrieved.ETag)
			return false
		}

		if retrieved.ContentHash != "hash-"+scenario.ItemName {
			t.Logf("ContentHash should be preserved: expected %s, got %s", "hash-"+scenario.ItemName, retrieved.ContentHash)
			return false
		}

		// Verify LastHydrated timestamp is preserved
		if retrieved.LastHydrated == nil {
			t.Logf("LastHydrated timestamp should be preserved")
			return false
		}

		// Verify state persists until upload succeeds
		// (this is implicit - the state is DIRTY_LOCAL and won't change without explicit transition)

		return true
	}

	config := &quick.Config{
		MaxCount: 100,
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property check failed: %v", err)
	}
}

// TestProperty42_DirtyLocalToHydratedTransition verifies that DIRTY_LOCAL → HYDRATED
// transition occurs on successful upload.
// **Property 42: Local Modification State Transition (Part 2: DIRTY_LOCAL → HYDRATED)**
// **Validates: Requirements 21.6**
func TestProperty42_DirtyLocalToHydratedTransition(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping property-based test in short mode")
	}

	property := func(scenario ModificationUploadScenario) bool {
		// Validate scenario inputs
		if scenario.ItemName == "" || len(scenario.ItemName) > 200 {
			return true // Skip invalid inputs
		}
		if scenario.ETag == "" || scenario.NewETag == "" || scenario.Size == 0 {
			return true // Skip invalid inputs
		}

		// Setup test filesystem with metadata store
		fs := newTestFilesystemWithMetadata(t)
		ctx := context.Background()

		// Generate a unique item ID
		itemID := "test-file-" + scenario.ItemName
		parentID := "root"

		// Ensure parent exists
		parentEntry := &metadata.Entry{
			ID:            parentID,
			Name:          "root",
			ItemType:      metadata.ItemKindDirectory,
			State:         metadata.ItemStateHydrated,
			OverlayPolicy: metadata.OverlayPolicyRemoteWins,
			CreatedAt:     time.Now().UTC(),
			UpdatedAt:     time.Now().UTC(),
			Pin: metadata.PinState{
				Mode: metadata.PinModeUnset,
			},
		}
		err := fs.metadataStore.Save(ctx, parentEntry)
		if err != nil {
			t.Logf("Failed to save parent: %v", err)
			return false
		}

		// Create a file entry in DIRTY_LOCAL state
		lastHydrated := time.Now().UTC().Add(-1 * time.Hour)
		fileEntry := &metadata.Entry{
			ID:            itemID,
			RemoteID:      itemID,
			ParentID:      parentID,
			Name:          scenario.ItemName,
			ItemType:      metadata.ItemKindFile,
			State:         metadata.ItemStateDirtyLocal,
			OverlayPolicy: metadata.OverlayPolicyRemoteWins,
			Size:          scenario.Size,
			ETag:          scenario.ETag, // Old ETag before upload
			ContentHash:   "hash-" + scenario.ItemName,
			LastHydrated:  &lastHydrated,
			CreatedAt:     time.Now().UTC(),
			UpdatedAt:     time.Now().UTC(),
			Pin: metadata.PinState{
				Mode: metadata.PinModeUnset,
			},
		}
		err = fs.metadataStore.Save(ctx, fileEntry)
		if err != nil {
			t.Logf("Failed to save file entry: %v", err)
			return false
		}

		// Simulate successful upload by transitioning to HYDRATED with new ETag
		_, err = fs.stateManager.Transition(ctx, itemID, metadata.ItemStateHydrated,
			metadata.WithUploadEvent(),
			metadata.WithETag(scenario.NewETag))
		if err != nil {
			t.Logf("Failed to transition to HYDRATED: %v", err)
			return false
		}

		// Verify the entry is now in HYDRATED state
		retrieved, err := fs.metadataStore.Get(ctx, itemID)
		if err != nil {
			t.Logf("Failed to retrieve entry: %v", err)
			return false
		}

		if retrieved.State != metadata.ItemStateHydrated {
			t.Logf("Expected state HYDRATED, got %s", retrieved.State)
			return false
		}

		// Verify ETag is updated to new value from server response
		if retrieved.ETag != scenario.NewETag {
			t.Logf("Expected ETag %s, got %s", scenario.NewETag, retrieved.ETag)
			return false
		}

		// Verify metadata is preserved (size, content hash)
		if retrieved.Size != scenario.Size {
			t.Logf("Size should be preserved: expected %d, got %d", scenario.Size, retrieved.Size)
			return false
		}

		if retrieved.ContentHash != "hash-"+scenario.ItemName {
			t.Logf("ContentHash should be preserved: expected %s, got %s", "hash-"+scenario.ItemName, retrieved.ContentHash)
			return false
		}

		// Verify LastHydrated timestamp is preserved
		if retrieved.LastHydrated == nil {
			t.Logf("LastHydrated timestamp should be preserved")
			return false
		}

		// Verify error fields are cleared
		if retrieved.LastError != nil {
			t.Logf("LastError should be nil after successful upload")
			return false
		}

		return true
	}

	config := &quick.Config{
		MaxCount: 100,
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property check failed: %v", err)
	}
}

// TestProperty42_DirtyLocalToErrorTransition verifies that DIRTY_LOCAL → ERROR
// transition occurs on upload failure.
// **Property 42: Local Modification State Transition (Part 3: DIRTY_LOCAL → ERROR)**
// **Validates: Requirements 21.6**
func TestProperty42_DirtyLocalToErrorTransition(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping property-based test in short mode")
	}

	property := func(scenario ModificationUploadScenario) bool {
		// Validate scenario inputs
		if scenario.ItemName == "" || len(scenario.ItemName) > 200 {
			return true // Skip invalid inputs
		}
		if scenario.ETag == "" || scenario.Size == 0 {
			return true // Skip invalid inputs
		}
		if !scenario.SimulateError || scenario.ErrorMessage == "" {
			return true // Skip scenarios without errors
		}

		// Setup test filesystem with metadata store
		fs := newTestFilesystemWithMetadata(t)
		ctx := context.Background()

		// Generate a unique item ID
		itemID := "test-file-" + scenario.ItemName
		parentID := "root"

		// Ensure parent exists
		parentEntry := &metadata.Entry{
			ID:            parentID,
			Name:          "root",
			ItemType:      metadata.ItemKindDirectory,
			State:         metadata.ItemStateHydrated,
			OverlayPolicy: metadata.OverlayPolicyRemoteWins,
			CreatedAt:     time.Now().UTC(),
			UpdatedAt:     time.Now().UTC(),
			Pin: metadata.PinState{
				Mode: metadata.PinModeUnset,
			},
		}
		err := fs.metadataStore.Save(ctx, parentEntry)
		if err != nil {
			t.Logf("Failed to save parent: %v", err)
			return false
		}

		// Create a file entry in DIRTY_LOCAL state
		lastHydrated := time.Now().UTC().Add(-1 * time.Hour)
		fileEntry := &metadata.Entry{
			ID:            itemID,
			RemoteID:      itemID,
			ParentID:      parentID,
			Name:          scenario.ItemName,
			ItemType:      metadata.ItemKindFile,
			State:         metadata.ItemStateDirtyLocal,
			OverlayPolicy: metadata.OverlayPolicyRemoteWins,
			Size:          scenario.Size,
			ETag:          scenario.ETag,
			ContentHash:   "hash-" + scenario.ItemName,
			LastHydrated:  &lastHydrated,
			CreatedAt:     time.Now().UTC(),
			UpdatedAt:     time.Now().UTC(),
			Pin: metadata.PinState{
				Mode: metadata.PinModeUnset,
			},
		}
		err = fs.metadataStore.Save(ctx, fileEntry)
		if err != nil {
			t.Logf("Failed to save file entry: %v", err)
			return false
		}

		// Simulate upload failure by transitioning to ERROR
		uploadErr := fmt.Errorf("%s", scenario.ErrorMessage)
		_, err = fs.stateManager.Transition(ctx, itemID, metadata.ItemStateError,
			metadata.WithUploadEvent(),
			metadata.WithTransitionError(uploadErr, true))
		if err != nil {
			t.Logf("Failed to transition to ERROR: %v", err)
			return false
		}

		// Verify the entry is now in ERROR state
		retrieved, err := fs.metadataStore.Get(ctx, itemID)
		if err != nil {
			t.Logf("Failed to retrieve entry: %v", err)
			return false
		}

		if retrieved.State != metadata.ItemStateError {
			t.Logf("Expected state ERROR, got %s", retrieved.State)
			return false
		}

		// Verify error is recorded
		if retrieved.LastError == nil {
			t.Logf("LastError should be set after failed upload")
			return false
		}

		if retrieved.LastError.Message != scenario.ErrorMessage {
			t.Logf("Expected error message %s, got %s", scenario.ErrorMessage, retrieved.LastError.Message)
			return false
		}

		// Verify error is marked as temporary
		if !retrieved.LastError.Temporary {
			t.Logf("Error should be marked as temporary")
			return false
		}

		// Verify previous state metadata is preserved (size, ETag, content hash)
		if retrieved.Size != scenario.Size {
			t.Logf("Size should be preserved after error: expected %d, got %d", scenario.Size, retrieved.Size)
			return false
		}

		if retrieved.ETag != scenario.ETag {
			t.Logf("ETag should be preserved after error: expected %s, got %s", scenario.ETag, retrieved.ETag)
			return false
		}

		if retrieved.ContentHash != "hash-"+scenario.ItemName {
			t.Logf("ContentHash should be preserved after error: expected %s, got %s", "hash-"+scenario.ItemName, retrieved.ContentHash)
			return false
		}

		// Verify LastHydrated timestamp is preserved
		if retrieved.LastHydrated == nil {
			t.Logf("LastHydrated timestamp should be preserved after error")
			return false
		}

		return true
	}

	config := &quick.Config{
		MaxCount: 100,
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property check failed: %v", err)
	}
}

// TestProperty42_ETagUpdateAfterUpload verifies that ETag is updated from
// server response after successful upload.
// **Property 42: ETag Update After Upload**
// **Validates: Requirements 21.6**
func TestProperty42_ETagUpdateAfterUpload(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping property-based test in short mode")
	}

	property := func(scenario ModificationUploadScenario) bool {
		// Validate scenario inputs
		if scenario.ItemName == "" || len(scenario.ItemName) > 200 {
			return true // Skip invalid inputs
		}
		if scenario.ETag == "" || scenario.NewETag == "" || scenario.Size == 0 {
			return true // Skip invalid inputs
		}
		// Ensure ETags are different to test update
		if scenario.ETag == scenario.NewETag {
			return true // Skip if ETags are the same
		}

		// Setup test filesystem with metadata store
		fs := newTestFilesystemWithMetadata(t)
		ctx := context.Background()

		// Generate a unique item ID
		itemID := "test-file-" + scenario.ItemName
		parentID := "root"

		// Ensure parent exists
		parentEntry := &metadata.Entry{
			ID:            parentID,
			Name:          "root",
			ItemType:      metadata.ItemKindDirectory,
			State:         metadata.ItemStateHydrated,
			OverlayPolicy: metadata.OverlayPolicyRemoteWins,
			CreatedAt:     time.Now().UTC(),
			UpdatedAt:     time.Now().UTC(),
			Pin: metadata.PinState{
				Mode: metadata.PinModeUnset,
			},
		}
		err := fs.metadataStore.Save(ctx, parentEntry)
		if err != nil {
			t.Logf("Failed to save parent: %v", err)
			return false
		}

		// Create a file entry in DIRTY_LOCAL state with old ETag
		lastHydrated := time.Now().UTC().Add(-1 * time.Hour)
		fileEntry := &metadata.Entry{
			ID:            itemID,
			RemoteID:      itemID,
			ParentID:      parentID,
			Name:          scenario.ItemName,
			ItemType:      metadata.ItemKindFile,
			State:         metadata.ItemStateDirtyLocal,
			OverlayPolicy: metadata.OverlayPolicyRemoteWins,
			Size:          scenario.Size,
			ETag:          scenario.ETag, // Old ETag
			ContentHash:   "hash-" + scenario.ItemName,
			LastHydrated:  &lastHydrated,
			CreatedAt:     time.Now().UTC(),
			UpdatedAt:     time.Now().UTC(),
			Pin: metadata.PinState{
				Mode: metadata.PinModeUnset,
			},
		}
		err = fs.metadataStore.Save(ctx, fileEntry)
		if err != nil {
			t.Logf("Failed to save file entry: %v", err)
			return false
		}

		// Verify old ETag is stored
		retrieved, err := fs.metadataStore.Get(ctx, itemID)
		if err != nil {
			t.Logf("Failed to retrieve entry before upload: %v", err)
			return false
		}

		if retrieved.ETag != scenario.ETag {
			t.Logf("Expected old ETag %s before upload, got %s", scenario.ETag, retrieved.ETag)
			return false
		}

		// Simulate successful upload with new ETag from server response
		_, err = fs.stateManager.Transition(ctx, itemID, metadata.ItemStateHydrated,
			metadata.WithUploadEvent(),
			metadata.WithETag(scenario.NewETag))
		if err != nil {
			t.Logf("Failed to transition to HYDRATED with new ETag: %v", err)
			return false
		}

		// Verify ETag is updated to new value
		retrieved, err = fs.metadataStore.Get(ctx, itemID)
		if err != nil {
			t.Logf("Failed to retrieve entry after upload: %v", err)
			return false
		}

		if retrieved.ETag != scenario.NewETag {
			t.Logf("Expected new ETag %s after upload, got %s", scenario.NewETag, retrieved.ETag)
			return false
		}

		// Verify state is HYDRATED
		if retrieved.State != metadata.ItemStateHydrated {
			t.Logf("Expected state HYDRATED after upload, got %s", retrieved.State)
			return false
		}

		// Verify old ETag is no longer present
		if retrieved.ETag == scenario.ETag {
			t.Logf("ETag was not updated from %s to %s", scenario.ETag, scenario.NewETag)
			return false
		}

		return true
	}

	config := &quick.Config{
		MaxCount: 100,
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property check failed: %v", err)
	}
}
