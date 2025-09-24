// Package graph provides functionality for interacting with the Microsoft Graph API.
package graph

import (
	"context"
	"io"

	"github.com/auriora/onemount/internal/graph/api"
)

// Provider implements the api.GraphProvider interface using the graph package functions.
type Provider struct {
	auth *Auth
}

// NewProvider creates a new Provider with the given Auth.
func NewProvider(auth *Auth) *Provider {
	return &Provider{
		auth: auth,
	}
}

// Setup initializes the provider.
func (p *Provider) Setup() error {
	return nil
}

// Teardown cleans up the provider.
func (p *Provider) Teardown() error {
	return nil
}

// Reset resets the provider to its initial state.
func (p *Provider) Reset() error {
	return nil
}

// RequestWithContext performs a request to the Microsoft Graph API with context.
func (p *Provider) RequestWithContext(ctx context.Context, resource string, method string, content io.Reader, headers ...api.Header) ([]byte, error) {
	// Convert api.Header to graph.Header
	graphHeaders := make([]Header, len(headers))
	for i, h := range headers {
		graphHeaders[i] = Header{
			key:   h.Key,
			value: h.Value,
		}
	}
	return RequestWithContext(ctx, resource, p.auth, method, content, graphHeaders...)
}

// Get performs a GET request to the Microsoft Graph API.
func (p *Provider) Get(resource string, headers ...api.Header) ([]byte, error) {
	// Convert api.Header to graph.Header
	graphHeaders := make([]Header, len(headers))
	for i, h := range headers {
		graphHeaders[i] = Header{
			key:   h.Key,
			value: h.Value,
		}
	}
	return Get(resource, p.auth, graphHeaders...)
}

// GetWithContext performs a GET request to the Microsoft Graph API with context.
func (p *Provider) GetWithContext(ctx context.Context, resource string, headers ...api.Header) ([]byte, error) {
	// Convert api.Header to graph.Header
	graphHeaders := make([]Header, len(headers))
	for i, h := range headers {
		graphHeaders[i] = Header{
			key:   h.Key,
			value: h.Value,
		}
	}
	return GetWithContext(ctx, resource, p.auth, graphHeaders...)
}

// GetItem fetches a DriveItem by ID.
func (p *Provider) GetItem(id string) (*api.DriveItem, error) {
	item, err := GetItem(id, p.auth)
	if err != nil {
		return nil, err
	}
	return convertDriveItem(item), nil
}

// GetItemChildren fetches the children of a DriveItem by ID.
func (p *Provider) GetItemChildren(id string) ([]*api.DriveItem, error) {
	items, err := GetItemChildren(id, p.auth)
	if err != nil {
		return nil, err
	}
	return convertDriveItems(items), nil
}

// GetItemChildrenPath fetches the children of a DriveItem by path.
func (p *Provider) GetItemChildrenPath(path string) ([]*api.DriveItem, error) {
	items, err := GetItemChildrenPath(path, p.auth)
	if err != nil {
		return nil, err
	}
	return convertDriveItems(items), nil
}

// GetItemPath fetches a DriveItem by path.
func (p *Provider) GetItemPath(path string) (*api.DriveItem, error) {
	item, err := GetItemPath(path, p.auth)
	if err != nil {
		return nil, err
	}
	return convertDriveItem(item), nil
}

// GetItemContent retrieves an item's content from the Graph endpoint.
func (p *Provider) GetItemContent(id string) ([]byte, uint64, error) {
	return GetItemContent(id, p.auth)
}

// GetItemContentStream retrieves an item's content and writes it to the provided writer.
func (p *Provider) GetItemContentStream(id string, output io.Writer) (uint64, error) {
	return GetItemContentStream(id, p.auth, output)
}

// Patch performs a PATCH request to the Microsoft Graph API.
func (p *Provider) Patch(resource string, content io.Reader, headers ...api.Header) ([]byte, error) {
	// Convert api.Header to graph.Header
	graphHeaders := make([]Header, len(headers))
	for i, h := range headers {
		graphHeaders[i] = Header{
			key:   h.Key,
			value: h.Value,
		}
	}
	return Patch(resource, p.auth, content, graphHeaders...)
}

// Post performs a POST request to the Microsoft Graph API.
func (p *Provider) Post(resource string, content io.Reader, headers ...api.Header) ([]byte, error) {
	// Convert api.Header to graph.Header
	graphHeaders := make([]Header, len(headers))
	for i, h := range headers {
		graphHeaders[i] = Header{
			key:   h.Key,
			value: h.Value,
		}
	}
	return Post(resource, p.auth, content, graphHeaders...)
}

// Put performs a PUT request to the Microsoft Graph API.
func (p *Provider) Put(resource string, content io.Reader, headers ...api.Header) ([]byte, error) {
	// Convert api.Header to graph.Header
	graphHeaders := make([]Header, len(headers))
	for i, h := range headers {
		graphHeaders[i] = Header{
			key:   h.Key,
			value: h.Value,
		}
	}
	return Put(resource, p.auth, content, graphHeaders...)
}

// Delete performs a DELETE request to the Microsoft Graph API.
func (p *Provider) Delete(resource string, headers ...api.Header) error {
	// Convert api.Header to graph.Header
	graphHeaders := make([]Header, len(headers))
	for i, h := range headers {
		graphHeaders[i] = Header{
			key:   h.Key,
			value: h.Value,
		}
	}
	return Delete(resource, p.auth, graphHeaders...)
}

// Mkdir creates a new directory.
func (p *Provider) Mkdir(name string, parentID string) (*api.DriveItem, error) {
	item, err := Mkdir(name, parentID, p.auth)
	if err != nil {
		return nil, err
	}
	return convertDriveItem(item), nil
}

// Rename renames an item.
func (p *Provider) Rename(itemID string, itemName string, parentID string) error {
	return Rename(itemID, itemName, parentID, p.auth)
}

// Remove removes an item.
func (p *Provider) Remove(id string) error {
	return Remove(id, p.auth)
}

// Helper function to convert graph.DriveItem to api.DriveItem
func convertDriveItem(item *DriveItem) *api.DriveItem {
	if item == nil {
		return nil
	}

	apiItem := &api.DriveItem{
		ID:               item.ID,
		Name:             item.Name,
		Size:             item.Size,
		ModTime:          item.ModTime,
		ConflictBehavior: item.ConflictBehavior,
		ETag:             item.ETag,
	}

	if item.Parent != nil {
		apiItem.Parent = &api.DriveItemParent{
			DriveID:   item.Parent.DriveID,
			DriveType: item.Parent.DriveType,
			ID:        item.Parent.ID,
			Path:      item.Parent.Path,
		}
	}

	if item.Folder != nil {
		apiItem.Folder = &api.Folder{
			ChildCount: item.Folder.ChildCount,
		}
	}

	if item.File != nil {
		apiItem.File = &api.File{
			Hashes: api.Hashes{
				SHA1Hash:     item.File.Hashes.SHA1Hash,
				QuickXorHash: item.File.Hashes.QuickXorHash,
			},
		}
	}

	if item.Deleted != nil {
		apiItem.Deleted = &api.Deleted{
			State: item.Deleted.State,
		}
	}

	return apiItem
}

// Helper function to convert []*graph.DriveItem to []*api.DriveItem
func convertDriveItems(items []*DriveItem) []*api.DriveItem {
	apiItems := make([]*api.DriveItem, len(items))
	for i, item := range items {
		apiItems[i] = convertDriveItem(item)
	}
	return apiItems
}
