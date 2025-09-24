package graph

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// ThumbnailInfo represents information about a thumbnail
type ThumbnailInfo struct {
	Height int    `json:"height"`
	Width  int    `json:"width"`
	URL    string `json:"url"`
}

// ThumbnailSet represents available thumbnails for a DriveItem
type ThumbnailSet struct {
	ID     string         `json:"id"`
	Small  *ThumbnailInfo `json:"small,omitempty"`
	Medium *ThumbnailInfo `json:"medium,omitempty"`
	Large  *ThumbnailInfo `json:"large,omitempty"`
}

// thumbnailsResponse represents the response from the thumbnails endpoint
type thumbnailsResponse struct {
	Value []ThumbnailSet `json:"value"`
}

// GetThumbnails retrieves available thumbnails for a DriveItem
// This is a convenience wrapper around GetThumbnailsWithContext for code that doesn't need
// to pass a context. It's kept for API consistency with other Graph API functions.
func GetThumbnails(itemID string, auth *Auth) (*ThumbnailSet, error) {
	return GetThumbnailsWithContext(context.Background(), itemID, auth)
}

// GetThumbnailsWithContext retrieves available thumbnails for a DriveItem with context
func GetThumbnailsWithContext(ctx context.Context, itemID string, auth *Auth) (*ThumbnailSet, error) {
	endpoint := fmt.Sprintf("/me/drive/items/%s/thumbnails", itemID)

	data, err := GetWithContext(ctx, endpoint, auth)
	if err != nil {
		return nil, fmt.Errorf("failed to get thumbnails: %w", err)
	}

	var response thumbnailsResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal thumbnails response: %w", err)
	}

	// If no thumbnails are available, return nil
	if len(response.Value) == 0 {
		return nil, nil
	}

	// Return the first thumbnail set
	return &response.Value[0], nil
}

// GetThumbnailContent retrieves the content of a thumbnail
func GetThumbnailContent(itemID string, size string, auth *Auth) ([]byte, error) {
	return GetThumbnailContentWithContext(context.Background(), itemID, size, auth)
}

// GetThumbnailContentWithContext retrieves the content of a thumbnail with context
func GetThumbnailContentWithContext(ctx context.Context, itemID string, size string, auth *Auth) ([]byte, error) {
	// Use the direct thumbnail endpoint to get the thumbnail content in a single API call
	endpoint := fmt.Sprintf("/me/drive/items/%s/thumbnails/0/%s/content", itemID, size)

	// Make the request
	req, err := http.NewRequestWithContext(ctx, "GET", GraphURL+endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add authorization header
	req.Header.Add("Authorization", "Bearer "+auth.AccessToken)

	// Send the request
	client := getHTTPClient()
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download thumbnail: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			// Just log the error since we can't return it from a defer
			fmt.Printf("Error closing response body: %v\n", err)
		}
	}()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download thumbnail: %s", resp.Status)
	}

	// Read the response body
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read thumbnail content: %w", err)
	}

	return data, nil
}

// downloadThumbnail downloads a thumbnail from the given URL
func downloadThumbnail(ctx context.Context, url string, auth *Auth) ([]byte, error) {
	// Create a new request
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add authorization header
	req.Header.Add("Authorization", "Bearer "+auth.AccessToken)

	// Send the request
	client := getHTTPClient()
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download thumbnail: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			// Just log the error since we can't return it from a defer
			fmt.Printf("Error closing response body: %v\n", err)
		}
	}()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download thumbnail: %s", resp.Status)
	}

	// Read the response body
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read thumbnail content: %w", err)
	}

	return data, nil
}

// GetThumbnailContentStream retrieves the content of a thumbnail as a stream
func GetThumbnailContentStream(itemID string, size string, auth *Auth, output io.Writer) error {
	return GetThumbnailContentStreamWithContext(context.Background(), itemID, size, auth, output)
}

// GetThumbnailContentStreamWithContext retrieves the content of a thumbnail as a stream with context
func GetThumbnailContentStreamWithContext(ctx context.Context, itemID string, size string, auth *Auth, output io.Writer) error {
	// Use the direct thumbnail endpoint to get the thumbnail content in a single API call
	endpoint := fmt.Sprintf("/me/drive/items/%s/thumbnails/0/%s/content", itemID, size)

	// Create a new request
	req, err := http.NewRequestWithContext(ctx, "GET", GraphURL+endpoint, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Add authorization header
	req.Header.Add("Authorization", "Bearer "+auth.AccessToken)

	// Send the request
	client := getHTTPClient()
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download thumbnail: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			// Just log the error since we can't return it from a defer
			fmt.Printf("Error closing response body: %v\n", err)
		}
	}()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download thumbnail: %s", resp.Status)
	}

	// Copy the response body to the output writer
	_, err = io.Copy(output, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to copy thumbnail content: %w", err)
	}

	return nil
}
