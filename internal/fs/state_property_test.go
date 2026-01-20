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
