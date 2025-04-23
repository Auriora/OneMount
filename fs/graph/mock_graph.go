// Package graph provides mocks for testing without actual API calls
package graph

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// MockResponse represents a predefined response for a specific request
type MockResponse struct {
	Body       []byte
	StatusCode int
	Error      error
}

// MockGraphClient is a mock implementation for testing Graph API interactions
type MockGraphClient struct {
	// Auth is the authentication information
	Auth Auth
	// Mock behavior controls
	ShouldFailRefresh bool
	ShouldFailRequest bool
	RequestResponses  map[string]MockResponse
}

// NewMockGraphClient creates a new MockGraphClient with default values
func NewMockGraphClient() *MockGraphClient {
	return &MockGraphClient{
		Auth: Auth{
			AccessToken:  "mock-access-token",
			RefreshToken: "mock-refresh-token",
			ExpiresAt:    time.Now().Add(time.Hour).Unix(),
			Account:      "mock@example.com",
		},
		RequestResponses: make(map[string]MockResponse),
	}
}

// AddMockResponse adds a predefined response for a specific resource path
func (m *MockGraphClient) AddMockResponse(resource string, body []byte, statusCode int, err error) {
	m.RequestResponses[resource] = MockResponse{
		Body:       body,
		StatusCode: statusCode,
		Error:      err,
	}
}

// AddMockItem adds a predefined DriveItem response for a specific resource path
func (m *MockGraphClient) AddMockItem(resource string, item *DriveItem) {
	body, _ := json.Marshal(item)
	m.AddMockResponse(resource, body, http.StatusOK, nil)
}

// AddMockItems adds a predefined list of DriveItems for a children request
func (m *MockGraphClient) AddMockItems(resource string, items []*DriveItem) {
	response := driveChildren{
		Children: items,
	}
	body, _ := json.Marshal(response)
	m.AddMockResponse(resource, body, http.StatusOK, nil)
}

// RequestWithContext is a mock implementation of the real RequestWithContext function
func (m *MockGraphClient) RequestWithContext(ctx context.Context, resource string, method string, content io.Reader, headers ...Header) ([]byte, error) {
	if m.ShouldFailRequest {
		return nil, errors.New("mock request failure")
	}

	// Check if we have a predefined response for this resource
	if response, exists := m.RequestResponses[resource]; exists {
		if response.Error != nil {
			return nil, response.Error
		}
		return response.Body, nil
	}

	// Default response based on the resource and method
	if strings.Contains(resource, "/children") {
		// Return empty children list by default
		return []byte(`{"value":[]}`), nil
	}

	if method == "GET" && strings.Contains(resource, "/content") {
		// Return empty content by default
		return []byte{}, nil
	}

	if method == "DELETE" {
		// Return success for DELETE
		return nil, nil
	}

	// For other requests, return a generic DriveItem
	item := &DriveItem{
		ID:   "mock-id",
		Name: "mock-item",
	}
	body, _ := json.Marshal(item)
	return body, nil
}

// Get is a mock implementation of the real Get function
func (m *MockGraphClient) Get(resource string, headers ...Header) ([]byte, error) {
	return m.RequestWithContext(context.Background(), resource, "GET", nil, headers...)
}

// GetWithContext is a mock implementation of the real GetWithContext function
func (m *MockGraphClient) GetWithContext(ctx context.Context, resource string, headers ...Header) ([]byte, error) {
	return m.RequestWithContext(ctx, resource, "GET", nil, headers...)
}

// Patch is a mock implementation of the real Patch function
func (m *MockGraphClient) Patch(resource string, content io.Reader, headers ...Header) ([]byte, error) {
	return m.RequestWithContext(context.Background(), resource, "PATCH", content, headers...)
}

// Post is a mock implementation of the real Post function
func (m *MockGraphClient) Post(resource string, content io.Reader, headers ...Header) ([]byte, error) {
	return m.RequestWithContext(context.Background(), resource, "POST", content, headers...)
}

// Put is a mock implementation of the real Put function
func (m *MockGraphClient) Put(resource string, content io.Reader, headers ...Header) ([]byte, error) {
	return m.RequestWithContext(context.Background(), resource, "PUT", content, headers...)
}

// Delete is a mock implementation of the real Delete function
func (m *MockGraphClient) Delete(resource string, headers ...Header) error {
	_, err := m.RequestWithContext(context.Background(), resource, "DELETE", nil, headers...)
	return err
}

// GetItemContent is a mock implementation of the real GetItemContent function
func (m *MockGraphClient) GetItemContent(id string) ([]byte, uint64, error) {
	resource := fmt.Sprintf("/me/drive/items/%s/content", id)
	if response, exists := m.RequestResponses[resource]; exists {
		if response.Error != nil {
			return nil, 0, response.Error
		}
		return response.Body, uint64(len(response.Body)), nil
	}

	// Default empty content
	return []byte{}, 0, nil
}

// GetItemContentStream is a mock implementation of the real GetItemContentStream function
func (m *MockGraphClient) GetItemContentStream(id string, output io.Writer) (uint64, error) {
	content, size, err := m.GetItemContent(id)
	if err != nil {
		return 0, err
	}
	_, err = output.Write(content)
	if err != nil {
		return 0, err
	}
	return size, nil
}

// GetItem is a mock implementation of the real GetItem function
func (m *MockGraphClient) GetItem(id string) (*DriveItem, error) {
	resource := IDPath(id)
	body, err := m.Get(resource)
	if err != nil {
		return nil, err
	}

	item := &DriveItem{}
	err = json.Unmarshal(body, item)
	return item, err
}

// GetItemPath is a mock implementation of the real GetItemPath function
func (m *MockGraphClient) GetItemPath(path string) (*DriveItem, error) {
	resource := ResourcePath(path)
	body, err := m.Get(resource)
	if err != nil {
		return nil, err
	}

	item := &DriveItem{}
	err = json.Unmarshal(body, item)
	return item, err
}

// GetItemChildren is a mock implementation of the real GetItemChildren function
func (m *MockGraphClient) GetItemChildren(id string) ([]*DriveItem, error) {
	resource := childrenPathID(id)
	body, err := m.Get(resource)
	if err != nil {
		return nil, err
	}

	var result driveChildren
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	return result.Children, nil
}

// GetItemChildrenPath is a mock implementation of the real GetItemChildrenPath function
func (m *MockGraphClient) GetItemChildrenPath(path string) ([]*DriveItem, error) {
	resource := childrenPath(path)
	body, err := m.Get(resource)
	if err != nil {
		return nil, err
	}

	var result driveChildren
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	return result.Children, nil
}

// Mkdir is a mock implementation of the real Mkdir function
func (m *MockGraphClient) Mkdir(name string, parentID string) (*DriveItem, error) {
	newFolder := DriveItem{
		Name:   name,
		Folder: &Folder{},
	}
	bytePayload, _ := json.Marshal(newFolder)
	resp, err := m.Post(childrenPathID(parentID), strings.NewReader(string(bytePayload)))
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(resp, &newFolder)
	return &newFolder, err
}

// Rename is a mock implementation of the real Rename function
func (m *MockGraphClient) Rename(itemID string, itemName string, parentID string) error {
	patchContent := DriveItem{
		ConflictBehavior: "replace",
		Name:             itemName,
		Parent: &DriveItemParent{
			ID: parentID,
		},
	}

	jsonPatch, _ := json.Marshal(patchContent)
	_, err := m.Patch("/me/drive/items/"+itemID, strings.NewReader(string(jsonPatch)))
	return err
}

// Remove is a mock implementation of the real Remove function
func (m *MockGraphClient) Remove(id string) error {
	return m.Delete("/me/drive/items/" + id)
}
