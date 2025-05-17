// Package api defines interfaces and types for the graph package.
package api

import (
	"context"
	"io"
	"time"
)

// GraphProvider defines the interface for interacting with the Microsoft Graph API.
type GraphProvider interface {
	// Setup initializes the provider.
	Setup() error

	// Teardown cleans up the provider.
	Teardown() error

	// Reset resets the provider to its initial state.
	Reset() error

	// RequestWithContext performs a request to the Microsoft Graph API with context.
	RequestWithContext(ctx context.Context, resource string, method string, content io.Reader, headers ...Header) ([]byte, error)

	// Get performs a GET request to the Microsoft Graph API.
	Get(resource string, headers ...Header) ([]byte, error)

	// GetWithContext performs a GET request to the Microsoft Graph API with context.
	GetWithContext(ctx context.Context, resource string, headers ...Header) ([]byte, error)

	// GetItem fetches a DriveItem by ID.
	GetItem(id string) (*DriveItem, error)

	// GetItemChildren fetches the children of a DriveItem by ID.
	GetItemChildren(id string) ([]*DriveItem, error)

	// GetItemChildrenPath fetches the children of a DriveItem by path.
	GetItemChildrenPath(path string) ([]*DriveItem, error)

	// GetItemPath fetches a DriveItem by path.
	GetItemPath(path string) (*DriveItem, error)

	// GetItemContent retrieves an item's content from the Graph endpoint.
	GetItemContent(id string) ([]byte, uint64, error)

	// GetItemContentStream retrieves an item's content and writes it to the provided writer.
	GetItemContentStream(id string, output io.Writer) (uint64, error)

	// Patch performs a PATCH request to the Microsoft Graph API.
	Patch(resource string, content io.Reader, headers ...Header) ([]byte, error)

	// Post performs a POST request to the Microsoft Graph API.
	Post(resource string, content io.Reader, headers ...Header) ([]byte, error)

	// Put performs a PUT request to the Microsoft Graph API.
	Put(resource string, content io.Reader, headers ...Header) ([]byte, error)

	// Delete performs a DELETE request to the Microsoft Graph API.
	Delete(resource string, headers ...Header) error

	// Mkdir creates a new directory.
	Mkdir(name string, parentID string) (*DriveItem, error)

	// Rename renames an item.
	Rename(itemID string, itemName string, parentID string) error

	// Remove removes an item.
	Remove(id string) error
}

// MockRecorder defines the interface for recording and verifying mock interactions.
type MockRecorder interface {
	// RecordCall records a method call.
	RecordCall(method string, args ...interface{})

	// RecordCallWithResult records a method call with a result and error.
	RecordCallWithResult(method string, result interface{}, err error, args ...interface{})

	// GetCalls returns all recorded calls.
	GetCalls() []MockCall

	// VerifyCall verifies a method was called a specific number of times.
	VerifyCall(method string, times int) bool
}

// MockCall represents a record of a method call on a mock.
type MockCall struct {
	Method string
	Args   []interface{}
	Result interface{}
	Error  error
	Time   time.Time
}
