package graph

import (
	"github.com/auriora/onemount/internal/testutil/framework"
	"testing"
)

// TestUT_GR_26_01_IDPath_VariousItemIDs_FormatsCorrectly tests the IDPath function with various inputs.
//
//	Test Case ID    UT-GR-26-01
//	Title           ID Path Formatting
//	Description     Tests the IDPath function with various inputs
//	Preconditions   None
//	Steps           1. Call IDPath with different item IDs
//	                2. Check if the results match expectations
//	Expected Result IDPath correctly formats item IDs for API requests
//	Notes: This test verifies that the IDPath function correctly formats item IDs for API requests.
func TestUT_GR_26_01_IDPath_VariousItemIDs_FormatsCorrectly(t *testing.T) {
	// Create a test fixture
	fixture := framework.NewUnitTestFixture("IDPathFixture")

	// Use the fixture to run the test
	fixture.Use(t, func(t *testing.T, fixture interface{}) {
		// Test case 1: Valid OneDrive item ID
		validID := "01BYE5RZ6QN3ZWBTUQOJFZXVGS7DSFGHI"
		result := IDPath(validID)
		expected := "/me/drive/items/01BYE5RZ6QN3ZWBTUQOJFZXVGS7DSFGHI"
		if result != expected {
			t.Errorf("IDPath(%q) = %q, expected %q", validID, result, expected)
		}

		// Test case 2: Root item ID
		rootID := "root"
		result = IDPath(rootID)
		expected = "/me/drive/root"
		if result != expected {
			t.Errorf("IDPath(%q) = %q, expected %q", rootID, result, expected)
		}

		// Test case 3: Empty string (should handle gracefully)
		emptyID := ""
		result = IDPath(emptyID)
		expected = "/me/drive/items/"
		if result != expected {
			t.Errorf("IDPath(%q) = %q, expected %q", emptyID, result, expected)
		}

		// Test case 4: ID with special characters that need URL encoding
		specialID := "test/id with spaces&special=chars"
		result = IDPath(specialID)
		expected = "/me/drive/items/test%2Fid%20with%20spaces&special=chars"
		if result != expected {
			t.Errorf("IDPath(%q) = %q, expected %q", specialID, result, expected)
		}

		// Test case 5: Very long ID (realistic OneDrive ID length)
		longID := "01ABCDEFGHIJKLMNOPQRSTUVWXYZ234567890ABCDEFGHIJKLMNOPQRSTUVWXYZ"
		result = IDPath(longID)
		expected = "/me/drive/items/01ABCDEFGHIJKLMNOPQRSTUVWXYZ234567890ABCDEFGHIJKLMNOPQRSTUVWXYZ"
		if result != expected {
			t.Errorf("IDPath(%q) = %q, expected %q", longID, result, expected)
		}

		t.Log("All IDPath test cases passed successfully")
	})
}

// TestUT_GR_27_01_ChildrenPath_VariousPaths_FormatsCorrectly tests the childrenPath function with various inputs.
//
//	Test Case ID    UT-GR-27-01
//	Title           Children Path Formatting
//	Description     Tests the childrenPath function with various inputs
//	Preconditions   None
//	Steps           1. Call childrenPath with different paths
//	                2. Check if the results match expectations
//	Expected Result childrenPath correctly formats paths for retrieving children
//	Notes: This test verifies that the childrenPath function correctly formats paths for retrieving children.
func TestUT_GR_27_01_ChildrenPath_VariousPaths_FormatsCorrectly(t *testing.T) {
	// Create a test fixture
	fixture := framework.NewUnitTestFixture("ChildrenPathFixture")

	// Use the fixture to run the test
	fixture.Use(t, func(t *testing.T, fixture interface{}) {
		// Test case 1: Root path
		rootPath := "/"
		result := childrenPath(rootPath)
		expected := "/me/drive/root/children"
		if result != expected {
			t.Errorf("childrenPath(%q) = %q, expected %q", rootPath, result, expected)
		}

		// Test case 2: Simple nested path
		simplePath := "/Documents"
		result = childrenPath(simplePath)
		expected = "/me/drive/root:%2FDocuments:/children"
		if result != expected {
			t.Errorf("childrenPath(%q) = %q, expected %q", simplePath, result, expected)
		}

		// Test case 3: Nested folder path
		nestedPath := "/Documents/Folder"
		result = childrenPath(nestedPath)
		expected = "/me/drive/root:%2FDocuments%2FFolder:/children"
		if result != expected {
			t.Errorf("childrenPath(%q) = %q, expected %q", nestedPath, result, expected)
		}

		// Test case 4: Path with spaces
		spacePath := "/My Documents/New Folder"
		result = childrenPath(spacePath)
		expected = "/me/drive/root:%2FMy%20Documents%2FNew%20Folder:/children"
		if result != expected {
			t.Errorf("childrenPath(%q) = %q, expected %q", spacePath, result, expected)
		}

		// Test case 5: Path with special characters
		specialPath := "/Files & Folders/Test (1)/Data+Info"
		result = childrenPath(specialPath)
		expected = "/me/drive/root:%2FFiles%20&%20Folders%2FTest%20%281%29%2FData+Info:/children"
		if result != expected {
			t.Errorf("childrenPath(%q) = %q, expected %q", specialPath, result, expected)
		}

		// Test case 6: Very long path (testing near API limits)
		longPath := "/Very/Long/Path/With/Many/Nested/Folders/That/Goes/Deep/Into/The/Directory/Structure/To/Test/API/Limits"
		result = childrenPath(longPath)
		expectedLong := "/me/drive/root:%2FVery%2FLong%2FPath%2FWith%2FMany%2FNested%2FFolders%2FThat%2FGoes%2FDeep%2FInto%2FThe%2FDirectory%2FStructure%2FTo%2FTest%2FAPI%2FLimits:/children"
		if result != expectedLong {
			t.Errorf("childrenPath(%q) = %q, expected %q", longPath, result, expectedLong)
		}

		t.Log("All childrenPath test cases passed successfully")
	})
}

// TestUT_GR_28_01_ChildrenPathID_VariousItemIDs_FormatsCorrectly tests the childrenPathID function with various inputs.
//
//	Test Case ID    UT-GR-28-01
//	Title           Children Path ID Formatting
//	Description     Tests the childrenPathID function with various inputs
//	Preconditions   None
//	Steps           1. Call childrenPathID with different item IDs
//	                2. Check if the results match expectations
//	Expected Result childrenPathID correctly formats item IDs for retrieving children
//	Notes: This test verifies that the childrenPathID function correctly formats item IDs for retrieving children.
func TestUT_GR_28_01_ChildrenPathID_VariousItemIDs_FormatsCorrectly(t *testing.T) {
	// Create a test fixture
	fixture := framework.NewUnitTestFixture("ChildrenPathIDFixture")

	// Use the fixture to run the test
	fixture.Use(t, func(t *testing.T, fixture interface{}) {
		// Test case 1: Valid OneDrive item ID
		validID := "01BYE5RZ6QN3ZWBTUQOJFZXVGS7DSFGHI"
		result := childrenPathID(validID)
		expected := "/me/drive/items/01BYE5RZ6QN3ZWBTUQOJFZXVGS7DSFGHI/children"
		if result != expected {
			t.Errorf("childrenPathID(%q) = %q, expected %q", validID, result, expected)
		}

		// Test case 2: Root item ID
		rootID := "root"
		result = childrenPathID(rootID)
		expected = "/me/drive/items/root/children"
		if result != expected {
			t.Errorf("childrenPathID(%q) = %q, expected %q", rootID, result, expected)
		}

		// Test case 3: Folder ID (typical OneDrive folder ID format)
		folderID := "01ABCDEFGHIJKLMNOPQRSTUVWXYZ234567"
		result = childrenPathID(folderID)
		expected = "/me/drive/items/01ABCDEFGHIJKLMNOPQRSTUVWXYZ234567/children"
		if result != expected {
			t.Errorf("childrenPathID(%q) = %q, expected %q", folderID, result, expected)
		}

		// Test case 4: Empty ID (should handle gracefully)
		emptyID := ""
		result = childrenPathID(emptyID)
		expected = "/me/drive/items//children"
		if result != expected {
			t.Errorf("childrenPathID(%q) = %q, expected %q", emptyID, result, expected)
		}

		// Test case 5: ID with special characters that need URL encoding
		specialID := "test/id with spaces&special=chars"
		result = childrenPathID(specialID)
		expected = "/me/drive/items/test%2Fid%20with%20spaces&special=chars/children"
		if result != expected {
			t.Errorf("childrenPathID(%q) = %q, expected %q", specialID, result, expected)
		}

		// Test case 6: Very long ID (realistic OneDrive ID length)
		longID := "01ABCDEFGHIJKLMNOPQRSTUVWXYZ234567890ABCDEFGHIJKLMNOPQRSTUVWXYZ"
		result = childrenPathID(longID)
		expected = "/me/drive/items/01ABCDEFGHIJKLMNOPQRSTUVWXYZ234567890ABCDEFGHIJKLMNOPQRSTUVWXYZ/children"
		if result != expected {
			t.Errorf("childrenPathID(%q) = %q, expected %q", longID, result, expected)
		}

		t.Log("All childrenPathID test cases passed successfully")
	})
}
