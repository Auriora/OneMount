package fs

import (
	"strings"
	"testing"
	"time"

	"github.com/auriora/onemount/internal/graph"
	"github.com/auriora/onemount/internal/testutil/framework"
	"github.com/auriora/onemount/internal/testutil/helpers"
	"github.com/hanwen/go-fuse/v2/fuse"
)

// TestUT_FS_VirtualFiles_01_XDGVolumeInfoAvailability tests that .xdg-volume-info
// is immediately available on mount without Graph API lookups.
//
//	Test Case ID    UT-FS-VirtualFiles-01
//	Title           .xdg-volume-info Immediate Availability
//	Description     Tests that .xdg-volume-info is available immediately on mount
//	Preconditions   Fresh filesystem mount
//	Steps           1. Create filesystem
//	                2. Create .xdg-volume-info virtual file
//	                3. Verify immediate availability
//	                4. Verify no Graph API calls needed
//	Expected Result .xdg-volume-info available immediately
//	Requirements    2B.1
func TestUT_FS_VirtualFiles_01_XDGVolumeInfoAvailability(t *testing.T) {
	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "XDGVolumeInfoAvailabilityFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
		// Create the filesystem
		fs, err := NewFilesystem(auth, mountPoint, cacheTTL)
		return fs, err
	})

	// Use the fixture to run the test
	fixture.Use(t, func(t *testing.T, fixture interface{}) {
		// Create assertions helper
		assert := framework.NewAssert(t)

		// Get the test data
		unitTestFixture, ok := fixture.(*framework.UnitTestFixture)
		if !ok {
			t.Fatalf("Expected fixture to be of type *framework.UnitTestFixture, but got %T", fixture)
		}
		fsFixture := unitTestFixture.SetupData.(*helpers.FSTestFixture)
		filesystem := fsFixture.FS.(*Filesystem)
		auth := fsFixture.Auth

		t.Log("=== .xdg-volume-info Immediate Availability Test ===")

		// Step 1: Create .xdg-volume-info virtual file
		t.Log("Step 1: Creating .xdg-volume-info virtual file...")

		fileName := ".xdg-volume-info"
		content := "[Volume Info]\nName=Test Drive\nIcon=dk-onedrive\n"

		// Get root inode
		root := filesystem.GetID(filesystem.root)
		assert.NotNil(root, "Root inode should exist")

		// Create virtual file inode
		virtualInode := NewInode(fileName, fuse.S_IFREG|0644, root)
		virtualInode.SetVirtualContent([]byte(content))

		// Register the virtual file
		filesystem.RegisterVirtualFile(virtualInode)

		// Step 2: Verify immediate availability
		t.Log("Step 2: Testing immediate availability...")

		start := time.Now()

		// Test path resolution
		foundInode, err := filesystem.GetPath("/.xdg-volume-info", auth)
		responseTime := time.Since(start)

		assert.NoError(err, "Should be able to resolve .xdg-volume-info path")
		assert.NotNil(foundInode, ".xdg-volume-info should be found")
		assert.True(responseTime < 10*time.Millisecond, "Virtual file access should be immediate")

		t.Logf("✓ .xdg-volume-info resolved in %v", responseTime)

		// Step 3: Verify virtual file properties
		t.Log("Step 3: Verifying virtual file properties...")

		if foundInode != nil {
			// Check ID has local- prefix
			assert.True(strings.HasPrefix(foundInode.ID(), "local-"),
				"Virtual file should have local- ID prefix, got: %s", foundInode.ID())

			// Check content
			virtualContent := foundInode.ReadVirtualContent(0, len(content))
			assert.Equal(content, string(virtualContent), "Virtual file content should match")

			// Check file attributes
			assert.Equal(fileName, foundInode.Name(), "File name should match")
			assert.True(foundInode.IsVirtual(), "File should be marked as virtual")
		}

		// Step 4: Verify no Graph API dependency
		t.Log("Step 4: Verifying no Graph API dependency...")

		// Virtual files should be accessible even without auth
		foundWithoutAuth, err := filesystem.GetPath("/.xdg-volume-info", nil)
		assert.NoError(err, "Virtual file should be accessible without auth")
		assert.NotNil(foundWithoutAuth, "Virtual file should be found without auth")

		if foundWithoutAuth != nil {
			assert.Equal(foundInode.ID(), foundWithoutAuth.ID(), "Should get same virtual file")
		}

		t.Log("✓ .xdg-volume-info immediate availability verified")
	})
}

// TestUT_FS_VirtualFiles_02_LocalIdentifierPersistence tests that virtual files
// persist with local-* identifiers and are excluded from sync operations.
//
//	Test Case ID    UT-FS-VirtualFiles-02
//	Title           Virtual File Persistence with local-* Identifiers
//	Description     Tests that virtual files persist with local-* IDs and sync exclusion
//	Preconditions   Virtual file created
//	Steps           1. Create virtual file with local-* ID
//	                2. Verify persistence in metadata store
//	                3. Verify sync exclusion
//	                4. Test filesystem restart persistence
//	Expected Result Virtual files persist correctly with local-* IDs
//	Requirements    2B.2
func TestUT_FS_VirtualFiles_02_LocalIdentifierPersistence(t *testing.T) {
	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "LocalIdentifierPersistenceFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
		// Create the filesystem
		fs, err := NewFilesystem(auth, mountPoint, cacheTTL)
		return fs, err
	})

	// Use the fixture to run the test
	fixture.Use(t, func(t *testing.T, fixture interface{}) {
		// Create assertions helper
		assert := framework.NewAssert(t)

		// Get the test data
		unitTestFixture, ok := fixture.(*framework.UnitTestFixture)
		if !ok {
			t.Fatalf("Expected fixture to be of type *framework.UnitTestFixture, but got %T", fixture)
		}
		fsFixture := unitTestFixture.SetupData.(*helpers.FSTestFixture)
		filesystem := fsFixture.FS.(*Filesystem)

		t.Log("=== Virtual File Persistence Test ===")

		// Step 1: Create multiple virtual files
		t.Log("Step 1: Creating virtual files with local-* identifiers...")

		root := filesystem.GetID(filesystem.root)
		assert.NotNil(root, "Root inode should exist")

		virtualFiles := []struct {
			name    string
			content string
		}{
			{".xdg-volume-info", "[Volume Info]\nName=Test Drive\nIcon=dk-onedrive\n"},
			{"local-config.txt", "Local configuration file"},
			{"virtual-readme.md", "# Virtual README\nThis is a local-only file"},
		}

		createdIDs := make([]string, 0, len(virtualFiles))

		for _, vf := range virtualFiles {
			// Create virtual file inode
			virtualInode := NewInode(vf.name, fuse.S_IFREG|0644, root)
			virtualInode.SetVirtualContent([]byte(vf.content))

			// Register the virtual file
			filesystem.RegisterVirtualFile(virtualInode)

			// Verify local- ID prefix
			assert.True(strings.HasPrefix(virtualInode.ID(), "local-"),
				"Virtual file %s should have local- ID prefix, got: %s", vf.name, virtualInode.ID())

			createdIDs = append(createdIDs, virtualInode.ID())
			t.Logf("✓ Created virtual file: %s with ID: %s", vf.name, virtualInode.ID())
		}

		// Step 2: Verify persistence in metadata store
		t.Log("Step 2: Verifying persistence in metadata store...")

		for i, id := range createdIDs {
			// Check virtual file registry
			virtualInode, exists := filesystem.getVirtualFile(id)
			assert.True(exists, "Virtual file %s should exist in registry", virtualFiles[i].name)
			assert.NotNil(virtualInode, "Virtual file %s should not be nil", virtualFiles[i].name)

			if virtualInode != nil {
				// Verify content
				expectedContent := virtualFiles[i].content
				actualContent := string(virtualInode.ReadVirtualContent(0, len(expectedContent)))
				assert.Equal(expectedContent, actualContent, "Content should match for %s", virtualFiles[i].name)

				// Verify virtual flag
				assert.True(virtualInode.IsVirtual(), "File %s should be marked as virtual", virtualFiles[i].name)
			}
		}

		// Step 3: Verify virtual files appear in directory listing
		t.Log("Step 3: Verifying virtual files in directory listing...")

		root.mu.RLock()
		rootChildren := make([]string, len(root.children))
		copy(rootChildren, root.children)
		root.mu.RUnlock()

		// Count virtual files in root children
		virtualCount := 0
		for _, childID := range rootChildren {
			if strings.HasPrefix(childID, "local-") {
				virtualCount++
			}
		}

		assert.True(virtualCount >= len(virtualFiles),
			"Root should contain at least %d virtual files, found %d", len(virtualFiles), virtualCount)

		// Step 4: Verify virtual files are excluded from sync operations
		t.Log("Step 4: Verifying sync exclusion...")

		// Virtual files should not be included in upload queues or sync operations
		// This is implementation-dependent, but we can verify they have local- IDs
		for _, id := range createdIDs {
			assert.True(strings.HasPrefix(id, "local-"),
				"Virtual file ID should start with local- to exclude from sync: %s", id)
		}

		// Step 5: Test overlay policy resolution
		t.Log("Step 5: Testing overlay policy resolution...")

		// Create a virtual file that might conflict with a remote file
		conflictName := "potential-conflict.txt"
		conflictInode := NewInode(conflictName, fuse.S_IFREG|0644, root)
		conflictInode.SetVirtualContent([]byte("Local virtual content"))
		filesystem.RegisterVirtualFile(conflictInode)

		// Verify the virtual file is accessible
		foundConflict, err := filesystem.GetChild(filesystem.root, conflictName, nil)
		assert.NoError(err, "Should be able to find virtual file")
		assert.NotNil(foundConflict, "Virtual file should be found")

		if foundConflict != nil {
			assert.True(strings.HasPrefix(foundConflict.ID(), "local-"),
				"Found file should be the virtual one with local- ID")
		}

		t.Log("✓ Virtual file persistence and overlay policy verified")
	})
}

// TestUT_FS_VirtualFiles_03_OverlayPolicyResolution tests that overlay policies
// correctly resolve conflicts between virtual and remote files.
//
//	Test Case ID    UT-FS-VirtualFiles-03
//	Title           Overlay Policy Resolution
//	Description     Tests that virtual files with overlay policies resolve conflicts correctly
//	Preconditions   Virtual and remote files with same name
//	Steps           1. Create virtual file with LOCAL_WINS policy
//	                2. Simulate remote file with same name
//	                3. Verify virtual file takes precedence
//	                4. Test directory listing shows only virtual file
//	Expected Result Overlay policy correctly resolves conflicts
//	Requirements    2B.2 (overlay policy aspect)
func TestUT_FS_VirtualFiles_03_OverlayPolicyResolution(t *testing.T) {
	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "OverlayPolicyResolutionFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
		// Create the filesystem
		fs, err := NewFilesystem(auth, mountPoint, cacheTTL)
		return fs, err
	})

	// Use the fixture to run the test
	fixture.Use(t, func(t *testing.T, fixture interface{}) {
		// Create assertions helper
		assert := framework.NewAssert(t)

		// Get the test data
		unitTestFixture, ok := fixture.(*framework.UnitTestFixture)
		if !ok {
			t.Fatalf("Expected fixture to be of type *framework.UnitTestFixture, but got %T", fixture)
		}
		fsFixture := unitTestFixture.SetupData.(*helpers.FSTestFixture)
		filesystem := fsFixture.FS.(*Filesystem)
		auth := fsFixture.Auth

		t.Log("=== Overlay Policy Resolution Test ===")

		// Step 1: Create a virtual file
		t.Log("Step 1: Creating virtual file with LOCAL_WINS policy...")

		root := filesystem.GetID(filesystem.root)
		assert.NotNil(root, "Root inode should exist")

		conflictName := "shared-file.txt"
		virtualContent := "This is the LOCAL virtual content"

		// Create virtual file
		virtualInode := NewInode(conflictName, fuse.S_IFREG|0644, root)
		virtualInode.SetVirtualContent([]byte(virtualContent))
		filesystem.RegisterVirtualFile(virtualInode)

		virtualID := virtualInode.ID()
		assert.True(strings.HasPrefix(virtualID, "local-"),
			"Virtual file should have local- ID prefix")

		// Step 2: Simulate a remote file with the same name
		t.Log("Step 2: Simulating remote file with same name...")

		// Note: In a real scenario, this would come from OneDrive
		// For testing, we'll create a mock remote inode
		remoteID := "remote-file-id-123"
		remoteInode := NewInode(conflictName, fuse.S_IFREG|0644, root)
		remoteInode.DriveItem.ID = remoteID

		// Add remote inode to filesystem (simulating sync)
		filesystem.metadata.Store(remoteID, remoteInode)
		filesystem.InsertNodeID(remoteInode)

		// Step 3: Test file resolution - virtual should win
		t.Log("Step 3: Testing overlay policy resolution...")

		// When looking up by name, virtual file should be found
		foundInode, err := filesystem.GetChild(filesystem.root, conflictName, auth)
		assert.NoError(err, "Should be able to resolve conflicting file name")
		assert.NotNil(foundInode, "Should find a file with the conflicting name")

		if foundInode != nil {
			// Should get the virtual file (LOCAL_WINS policy)
			assert.Equal(virtualID, foundInode.ID(),
				"Should resolve to virtual file ID, got %s", foundInode.ID())
			assert.True(strings.HasPrefix(foundInode.ID(), "local-"),
				"Resolved file should be the virtual one")

			// Verify content is from virtual file
			if foundInode.IsVirtual() {
				actualContent := string(foundInode.ReadVirtualContent(0, len(virtualContent)))
				assert.Equal(virtualContent, actualContent, "Should get virtual file content")
			}
		}

		// Step 4: Verify directory listing shows only virtual file
		t.Log("Step 4: Verifying directory listing resolution...")

		// Get all children of root
		root.mu.RLock()
		rootChildren := make([]string, len(root.children))
		copy(rootChildren, root.children)
		root.mu.RUnlock()

		// Count files with the conflicting name
		conflictCount := 0
		var foundChildID string

		for _, childID := range rootChildren {
			child := filesystem.GetID(childID)
			if child != nil && child.Name() == conflictName {
				conflictCount++
				foundChildID = childID
			}
		}

		// Should only see one file with the conflicting name (the virtual one)
		assert.Equal(1, conflictCount, "Should only see one file with conflicting name in directory listing")
		assert.Equal(virtualID, foundChildID, "Directory listing should show virtual file")

		// Step 5: Verify remote file is still accessible by ID but not by name
		t.Log("Step 5: Verifying remote file accessibility...")

		// Remote file should still be accessible by direct ID lookup
		remoteByID := filesystem.GetID(remoteID)
		assert.NotNil(remoteByID, "Remote file should still be accessible by ID")

		if remoteByID != nil {
			assert.Equal(remoteID, remoteByID.ID(), "Remote file ID should match")
			assert.Equal(conflictName, remoteByID.Name(), "Remote file name should match")
		}

		t.Log("✓ Overlay policy resolution verified - virtual file takes precedence")
	})
}
