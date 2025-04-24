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
	// First get the thumbnail set
	thumbnailSet, err := GetThumbnailsWithContext(ctx, itemID, auth)
	if err != nil {
		return nil, err
	}

	if thumbnailSet == nil {
		return nil, fmt.Errorf("no thumbnails available for item %s", itemID)
	}

	// Get the URL for the requested size
	var url string
	switch size {
	case "small":
		if thumbnailSet.Small != nil {
			url = thumbnailSet.Small.URL
		}
	case "medium":
		if thumbnailSet.Medium != nil {
			url = thumbnailSet.Medium.URL
		}
	case "large":
		if thumbnailSet.Large != nil {
			url = thumbnailSet.Large.URL
		}
	default:
		return nil, fmt.Errorf("invalid thumbnail size: %s", size)
	}

	if url == "" {
		return nil, fmt.Errorf("no thumbnail available for size: %s", size)
	}

	// Download the thumbnail content
	return downloadThumbnail(ctx, url, auth)
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
	defer resp.Body.Close()

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
	// First get the thumbnail set
	thumbnailSet, err := GetThumbnailsWithContext(ctx, itemID, auth)
	if err != nil {
		return err
	}

	if thumbnailSet == nil {
		return fmt.Errorf("no thumbnails available for item %s", itemID)
	}

	// Get the URL for the requested size
	var url string
	switch size {
	case "small":
		if thumbnailSet.Small != nil {
			url = thumbnailSet.Small.URL
		}
	case "medium":
		if thumbnailSet.Medium != nil {
			url = thumbnailSet.Medium.URL
		}
	case "large":
		if thumbnailSet.Large != nil {
			url = thumbnailSet.Large.URL
		}
	default:
		return fmt.Errorf("invalid thumbnail size: %s", size)
	}

	if url == "" {
		return fmt.Errorf("no thumbnail available for size: %s", size)
	}

	// Create a new request
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
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
	defer resp.Body.Close()

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
