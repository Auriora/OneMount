package graph

import (
	"fmt"
	"testing"

	"github.com/auriora/onemount/pkg/graph/api"
	"github.com/auriora/onemount/pkg/graph/mock"
	"github.com/auriora/onemount/pkg/testutil/framework"
)

// TestUT_GR_07_01_GraphAPI_VariousPaths_ReturnsCorrectItems tests retrieving items from the Microsoft Graph API.
//
//	Test Case ID    UT-GR-07-01
//	Title           Graph API Item Retrieval
//	Description     Tests retrieving items from the Microsoft Graph API
//	Preconditions   1. User is authenticated with valid credentials
//	                2. Network connection is available
//	Steps           1. Load authentication tokens
//	                2. Call GetItemPath with different paths
//	                3. Check if the result matches expectations
//	Expected Result GetItemPath returns the correct item for valid paths and an error for invalid paths
//	Notes: This test verifies that the GetItemPath function correctly retrieves items from the Microsoft Graph API.
func TestUT_GR_07_01_GraphAPI_VariousPaths_ReturnsCorrectItems(t *testing.T) {
	// Create a test fixture
	fixture := framework.NewUnitTestFixture("GraphAPIItemRetrievalFixture")

	// Set up the fixture
	fixture.WithSetup(func(t *testing.T) (interface{}, error) {
		// Create a new mock graph provider
		mockProvider := mock.NewAPIGraphProvider()

		// Add some mock items
		rootItem := &api.DriveItem{
			ID:   "root",
			Name: "root",
			Folder: &api.Folder{
				ChildCount: 2,
			},
		}

		documentsItem := &api.DriveItem{
			ID:   "documents",
			Name: "Documents",
			Folder: &api.Folder{
				ChildCount: 1,
			},
			Parent: &api.DriveItemParent{
				ID:   "root",
				Path: "/drive/root:",
			},
		}

		fileItem := &api.DriveItem{
			ID:   "file1",
			Name: "file1.txt",
			File: &api.File{},
			Parent: &api.DriveItemParent{
				ID:   "documents",
				Path: "/drive/root:/Documents:",
			},
		}

		// Add the mock items to the provider
		mockProvider.AddMockItem("root", rootItem)
		mockProvider.AddMockItem("/me/drive/root", rootItem)
		mockProvider.AddMockItem("/", rootItem)
		mockProvider.AddMockItem("documents", documentsItem)
		mockProvider.AddMockItem("/Documents", documentsItem)
		mockProvider.AddMockItem("/Documents/file1.txt", fileItem)
		mockProvider.AddMockItem("/me/drive/root:/Documents:", documentsItem)
		mockProvider.AddMockItem("file1", fileItem)
		mockProvider.AddMockItem("/me/drive/root:/Documents/file1.txt:", fileItem)

		// Add mock children
		mockProvider.AddMockItems("root", []*api.DriveItem{documentsItem})
		mockProvider.AddMockItems("documents", []*api.DriveItem{fileItem})

		fmt.Printf("Created mock graph provider with items: %+v\n", mockProvider)
		return mockProvider, nil
	})

	// Use the fixture to run the test
	fixture.Use(t, func(t *testing.T, fixture interface{}) {
		unitFixture, ok := fixture.(*framework.UnitTestFixture)
		if !ok {
			t.Fatalf("Expected fixture to be *framework.UnitTestFixture, got %T", fixture)
		}
		provider, ok := unitFixture.SetupData.(*mock.APIGraphProvider)
		if !ok {
			t.Fatalf("Expected SetupData to be *mock.APIGraphProvider, got %T", unitFixture.SetupData)
		}

		// Debug prints to inspect provider and its fields
		fmt.Printf("DEBUG: provider = %+v\n", provider)
		if provider != nil {
			fmt.Printf("DEBUG: provider.Client = %+v\n", provider.Client)
		}

		// Test getting the root item
		rootItem, err := provider.GetItemPath("/")
		if err != nil {
			t.Fatalf("Failed to get root item: %v", err)
		}

		// Debug print for rootItem and err
		fmt.Printf("DEBUG: rootItem = %+v\n", rootItem)
		fmt.Printf("DEBUG: err from GetItemPath = %v\n", err)
		if rootItem == nil {
			t.Fatalf("rootItem is nil, expected a valid DriveItem")
		}
		if rootItem.ID != "root" {
			t.Errorf("Expected root item ID to be 'root', got '%s'", rootItem.ID)
		}

		// Test getting a folder
		documentsItem, err := provider.GetItemPath("/Documents")
		if err != nil {
			t.Fatalf("Failed to get documents item: %v", err)
		}
		if documentsItem.ID != "documents" {
			t.Errorf("Expected documents item ID to be 'documents', got '%s'", documentsItem.ID)
		}

		// Test getting a file
		fileItem, err := provider.GetItemPath("/Documents/file1.txt")
		if err != nil {
			t.Fatalf("Failed to get file item: %v", err)
		}
		if fileItem.ID != "file1" {
			t.Errorf("Expected file item ID to be 'file1', got '%s'", fileItem.ID)
		}

		// Test getting a non-existent item
		_, err = provider.GetItemPath("/NonExistent")
		if err == nil {
			t.Errorf("Expected error when getting non-existent item, got nil")
		}
	})
}
