package graph

import (
	"testing"
	"time"

	"github.com/auriora/onemount/internal/testutil/framework"
)

// TestUT_GR_03_01_DriveItem_DifferentTypes_IsDirectoryReturnsCorrectValue tests the IsDir method of DriveItem.
//
//	Test Case ID    UT-GR-03-01
//	Title           DriveItem IsDir Method
//	Description     Tests the IsDir method of DriveItem
//	Preconditions   None
//	Steps           1. Create DriveItem objects with different types (folder, file, empty)
//	                2. Call IsDir on each object
//	                3. Check if the result matches expectations
//	Expected Result IsDir returns true for folders and false for files and empty items
//	Notes: This test verifies that the IsDir method correctly identifies folders.
func TestUT_GR_03_01_DriveItem_DifferentTypes_IsDirectoryReturnsCorrectValue(t *testing.T) {
	// Create a test fixture
	fixture := framework.NewUnitTestFixture("DriveItemIsDirFixture")

	// Set up the fixture with test data
	fixture.WithSetup(func(t *testing.T) (interface{}, error) {
		// Create test DriveItems of different types
		testItems := map[string]*DriveItem{
			"folder": {
				ID:   "folder-id",
				Name: "Test Folder",
				Folder: &Folder{
					ChildCount: 5,
				},
			},
			"file": {
				ID:   "file-id",
				Name: "test.txt",
				File: &File{
					Hashes: Hashes{
						SHA1Hash:     "abc123",
						QuickXorHash: "def456",
					},
				},
			},
			"deleted": {
				ID:   "deleted-id",
				Name: "deleted-item",
				Deleted: &Deleted{
					State: "deleted",
				},
			},
			"empty": {
				ID:   "empty-id",
				Name: "empty-item",
				// No Folder, File, or Deleted fields set
			},
		}
		return testItems, nil
	})

	// Use the fixture to run the test
	fixture.Use(t, func(t *testing.T, fixtureObj interface{}) {
		fixture := fixtureObj.(*framework.UnitTestFixture)
		testItems := fixture.SetupData.(map[string]*DriveItem)
		assert := framework.NewAssert(t)

		// Test folder item - should return true
		folderItem := testItems["folder"]
		assert.True(folderItem.IsDir(), "Folder item should be identified as directory")

		// Test file item - should return false
		fileItem := testItems["file"]
		assert.False(fileItem.IsDir(), "File item should not be identified as directory")

		// Test deleted item - should return false
		deletedItem := testItems["deleted"]
		assert.False(deletedItem.IsDir(), "Deleted item should not be identified as directory")

		// Test empty item (no type fields set) - should return false
		emptyItem := testItems["empty"]
		assert.False(emptyItem.IsDir(), "Empty item should not be identified as directory")
	})
}

// TestUT_GR_04_01_DriveItem_ModificationTime_ReturnsCorrectUnixTimestamp tests the ModTimeUnix method of DriveItem.
//
//	Test Case ID    UT-GR-04-01
//	Title           DriveItem ModTimeUnix Method
//	Description     Tests the ModTimeUnix method of DriveItem
//	Preconditions   None
//	Steps           1. Create a DriveItem with a specific modification time
//	                2. Call ModTimeUnix on the item
//	                3. Check if the result matches the expected Unix timestamp
//	Expected Result ModTimeUnix returns the correct Unix timestamp
//	Notes: This test verifies that the ModTimeUnix method correctly converts modification times to Unix timestamps.
func TestUT_GR_04_01_DriveItem_ModificationTime_ReturnsCorrectUnixTimestamp(t *testing.T) {
	// Create a test fixture
	fixture := framework.NewUnitTestFixture("DriveItemModTimeUnixFixture")

	// Set up the fixture with test data
	fixture.WithSetup(func(t *testing.T) (interface{}, error) {
		// Create test DriveItems with different modification times
		testTime := time.Date(2023, 12, 25, 10, 30, 45, 0, time.UTC)
		expectedUnix := uint64(1703500245) // Unix timestamp for the test time

		testItems := map[string]*DriveItem{
			"withModTime": {
				ID:      "item-with-modtime",
				Name:    "test-file.txt",
				ModTime: &testTime,
			},
			"withoutModTime": {
				ID:      "item-without-modtime",
				Name:    "test-file2.txt",
				ModTime: nil,
			},
		}

		return map[string]interface{}{
			"items":        testItems,
			"expectedUnix": expectedUnix,
		}, nil
	})

	// Use the fixture to run the test
	fixture.Use(t, func(t *testing.T, fixtureObj interface{}) {
		fixture := fixtureObj.(*framework.UnitTestFixture)
		data := fixture.SetupData.(map[string]interface{})
		testItems := data["items"].(map[string]*DriveItem)
		expectedUnix := data["expectedUnix"].(uint64)
		assert := framework.NewAssert(t)

		// Test item with modification time
		itemWithModTime := testItems["withModTime"]
		actualUnix := itemWithModTime.ModTimeUnix()
		assert.Equal(expectedUnix, actualUnix, "ModTimeUnix should return correct Unix timestamp")

		// Test item without modification time - should return 0
		itemWithoutModTime := testItems["withoutModTime"]
		actualUnixZero := itemWithoutModTime.ModTimeUnix()
		assert.Equal(uint64(0), actualUnixZero, "ModTimeUnix should return 0 for items without modification time")
	})
}

// TestUT_GR_05_01_DriveItem_VariousChecksums_VerificationReturnsCorrectResult tests the VerifyChecksum method of DriveItem.
//
//	Test Case ID    UT-GR-05-01
//	Title           DriveItem VerifyChecksum Method
//	Description     Tests the VerifyChecksum method of DriveItem
//	Preconditions   None
//	Steps           1. Create DriveItem objects with different checksums
//	                2. Call VerifyChecksum with matching and non-matching checksums
//	                3. Check if the result matches expectations
//	Expected Result VerifyChecksum returns true for matching checksums and false for non-matching checksums
//	Notes: This test verifies that the VerifyChecksum method correctly verifies checksums.
func TestUT_GR_05_01_DriveItem_VariousChecksums_VerificationReturnsCorrectResult(t *testing.T) {
	// Create a test fixture
	fixture := framework.NewUnitTestFixture("DriveItemVerifyChecksumFixture")

	// Set up the fixture with test data
	fixture.WithSetup(func(t *testing.T) (interface{}, error) {
		// Create test DriveItems with different checksums
		testItems := map[string]*DriveItem{
			"fileWithHashes": {
				ID:   "file-with-hashes",
				Name: "test-file.txt",
				File: &File{
					Hashes: Hashes{
						SHA1Hash:     "abc123def456",
						QuickXorHash: "xyz789uvw012",
					},
				},
			},
			"fileWithoutHashes": {
				ID:   "file-without-hashes",
				Name: "test-file2.txt",
				File: &File{
					// No hashes
				},
			},
			"folder": {
				ID:   "folder-id",
				Name: "test-folder",
				Folder: &Folder{
					ChildCount: 0,
				},
			},
		}
		return testItems, nil
	})

	// Use the fixture to run the test
	fixture.Use(t, func(t *testing.T, fixtureObj interface{}) {
		fixture := fixtureObj.(*framework.UnitTestFixture)
		testItems := fixture.SetupData.(map[string]*DriveItem)
		assert := framework.NewAssert(t)

		// Test file with hashes - matching QuickXor
		fileWithHashes := testItems["fileWithHashes"]
		assert.True(fileWithHashes.VerifyChecksum("xyz789uvw012"), "Should verify matching QuickXor hash")

		// Test file with hashes - non-matching QuickXor
		assert.False(fileWithHashes.VerifyChecksum("wronghash"), "Should not verify non-matching QuickXor hash")

		// Test file with hashes - empty checksum
		assert.False(fileWithHashes.VerifyChecksum(""), "Should not verify empty checksum")

		// Test file without hashes
		fileWithoutHashes := testItems["fileWithoutHashes"]
		assert.False(fileWithoutHashes.VerifyChecksum("anyhash"), "Should not verify hash for file without hashes")

		// Test folder (should not have checksums)
		folder := testItems["folder"]
		assert.False(folder.VerifyChecksum("anyhash"), "Should not verify hash for folder")

		// Test case-insensitive comparison
		assert.True(fileWithHashes.VerifyChecksum("XYZ789UVW012"), "Should verify hash with different case")
	})
}

// TestUT_GR_06_01_DriveItem_VariousETags_MatchReturnsCorrectResult tests the ETagIsMatch method of DriveItem.
//
//	Test Case ID    UT-GR-06-01
//	Title           DriveItem ETagIsMatch Method
//	Description     Tests the ETagIsMatch method of DriveItem
//	Preconditions   None
//	Steps           1. Create DriveItem objects with different ETags
//	                2. Call ETagIsMatch with matching and non-matching ETags
//	                3. Check if the result matches expectations
//	Expected Result ETagIsMatch returns true for matching ETags and false for non-matching ETags
//	Notes: This test verifies that the ETagIsMatch method correctly matches ETags.
func TestUT_GR_06_01_DriveItem_VariousETags_MatchReturnsCorrectResult(t *testing.T) {
	// Create a test fixture
	fixture := framework.NewUnitTestFixture("DriveItemETagIsMatchFixture")

	// Set up the fixture with test data
	fixture.WithSetup(func(t *testing.T) (interface{}, error) {
		// Create test DriveItems with different ETags
		testItems := map[string]*DriveItem{
			"itemWithETag": {
				ID:   "item-with-etag",
				Name: "test-file.txt",
				ETag: "\"abc123def456\"",
			},
			"itemWithoutETag": {
				ID:   "item-without-etag",
				Name: "test-file2.txt",
				// No ETag
			},
		}
		return testItems, nil
	})

	// Use the fixture to run the test
	fixture.Use(t, func(t *testing.T, fixtureObj interface{}) {
		fixture := fixtureObj.(*framework.UnitTestFixture)
		testItems := fixture.SetupData.(map[string]*DriveItem)
		assert := framework.NewAssert(t)

		// Test item with ETag - matching ETag
		itemWithETag := testItems["itemWithETag"]
		assert.True(itemWithETag.ETagIsMatch("\"abc123def456\""), "Should match identical ETag")

		// Test item with ETag - non-matching ETag
		assert.False(itemWithETag.ETagIsMatch("\"wrongetag\""), "Should not match different ETag")

		// Test item with ETag - empty ETag comparison
		assert.False(itemWithETag.ETagIsMatch(""), "Should not match empty ETag")

		// Test item without ETag
		itemWithoutETag := testItems["itemWithoutETag"]
		assert.False(itemWithoutETag.ETagIsMatch("\"anyetag\""), "Should not match when item has no ETag")

		// Test item without ETag - empty ETag comparison
		assert.False(itemWithoutETag.ETagIsMatch(""), "Should not match empty ETag when item has no ETag")
	})
}
