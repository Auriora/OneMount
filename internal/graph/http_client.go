// Package graph provides the basic APIs to interact with Microsoft Graph.
package graph

import "net/http"

// HTTPClient defines the interface for HTTP operations.
// This allows for dependency injection and easier testing.
type HTTPClient interface {
	// Do executes an HTTP request and returns an HTTP response.
	Do(req *http.Request) (*http.Response, error)
}
