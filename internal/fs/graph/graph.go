// Package graph provides the basic APIs to interact with Microsoft Graph. This includes
// Package graph provides the basic APIs to interact with Microsoft Graph. This includes
// the DriveItem resource and supporting resources which are the basis of working with
// files and folders through the Microsoft Graph API.
package graph

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/imdario/mergo"
	"github.com/rs/zerolog/log"
)

// GraphURL is the API endpoint of Microsoft Graph
const GraphURL = "https://graph.microsoft.com/v1.0"

// Default timeout for HTTP requests
const defaultRequestTimeout = 60 * time.Second

// responseCache is a singleton cache for API responses
var (
	httpClient    HTTPClient
	responseCache *ResponseCache
	clientOnce    sync.Once
	cacheOnce     sync.Once

	// operationalOffline is a flag that can be set to force offline mode
	operationalOffline      = false
	operationalOfflineMutex sync.RWMutex
)

// Graph represents a client for interacting with the Microsoft Graph API.
type Graph struct {
	httpClient HTTPClient
}

// NewGraph creates a new Graph client using the shared HTTP client.
func NewGraph() *Graph {
	return NewGraphWithClient(getSharedHTTPClient())
}

// NewGraphWithClient creates a new Graph client with a custom HTTP client.
func NewGraphWithClient(client HTTPClient) *Graph {
	return &Graph{
		httpClient: client,
	}
}

// getResponseCache returns the shared response cache
func getResponseCache() *ResponseCache {
	cacheOnce.Do(func() {
		// Create the response cache with a default TTL of 5 minutes
		responseCache = NewResponseCache(5 * time.Minute)
		log.Info().Msg("Initialized response cache with 5-minute TTL")
	})

	return responseCache
}

// graphError is an internal struct used when decoding Graph's error messages
type graphError struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// Header This is an additional header that can be specified to Request
type Header struct {
	key, value string
}

// Request performs an authenticated request to Microsoft Graph
func (g *Graph) Request(resource string, auth *Auth, method string, content io.Reader, headers ...Header) ([]byte, error) {
	return g.RequestWithContext(context.Background(), resource, auth, method, content, headers...)
}

// Request performs an authenticated request to Microsoft Graph
// This is a backward-compatible function that uses the default Graph client
func Request(resource string, auth *Auth, method string, content io.Reader, headers ...Header) ([]byte, error) {
	return NewGraph().Request(resource, auth, method, content, headers...)
}

// RequestWithContext performs an authenticated request to Microsoft Graph with context
func (g *Graph) RequestWithContext(ctx context.Context, resource string, auth *Auth, method string, content io.Reader, headers ...Header) ([]byte, error) {
	// Check if we're in operational offline mode
	if GetOperationalOffline() {
		log.Debug().Str("method", method).Str("resource", resource).Msg("In operational offline mode, returning network error")
		return nil, errors.New("network unavailable: operational offline mode is enabled")
	}

	if auth == nil || auth.AccessToken == "" {
		// a catch all condition to avoid wiping our auth by accident
		log.Error().Msg("Auth was empty and we attempted to make a request with it!")
		return nil, errors.New("cannot make a request with empty auth")
	}

	log.Debug().Str("method", method).Str("resource", resource).Msg("Starting auth refresh")
	if err := auth.Refresh(ctx); err != nil {
		log.Warn().Err(err).Str("method", method).Str("resource", resource).Msg("Auth refresh failed, continuing with current token")
	}
	log.Debug().Str("method", method).Str("resource", resource).Msg("Auth refresh completed")

	log.Debug().Str("method", method).Str("resource", resource).Msg("Using HTTP client")
	request, _ := http.NewRequestWithContext(ctx, method, GraphURL+resource, content)
	request.Header.Add("Authorization", "bearer "+auth.AccessToken)
	switch method { // request type-specific code here
	case "PATCH":
		request.Header.Add("If-Match", "*")
		request.Header.Add("Content-Type", "application/json")
	case "POST":
		request.Header.Add("Content-Type", "application/json")
	case "PUT":
		request.Header.Add("Content-Type", "text/plain")
	}
	for _, header := range headers {
		request.Header.Add(header.key, header.value)
	}

	log.Debug().Str("method", method).Str("resource", resource).Msg("Starting network request with context")
	log.Debug().Str("method", method).Str("resource", resource).Str("url", GraphURL+resource).Msg("About to execute HTTP request")
	response, err := g.httpClient.Do(request)
	if err != nil {
		// Check if the error was due to context cancellation
		if ctx.Err() != nil {
			log.Debug().Err(ctx.Err()).Str("method", method).Str("resource", resource).Msg("Network request cancelled by context")
			return nil, ctx.Err()
		}
		// the actual request failed for other reasons
		log.Debug().Err(err).Str("method", method).Str("resource", resource).Msg("Network request failed")
		return nil, err
	}
	log.Debug().Str("method", method).Str("resource", resource).Int("statusCode", response.StatusCode).Msg("Network request completed")

	log.Debug().Str("method", method).Str("resource", resource).Msg("Starting to read response body")
	body, err := io.ReadAll(response.Body)
	if err != nil {
		log.Error().Str("method", method).Str("resource", resource).Err(err).Msg("Error reading response body")
		return nil, fmt.Errorf("error reading response body: %v", err)
	}
	log.Debug().Str("method", method).Str("resource", resource).Int("bodySize", len(body)).Msg("Successfully read response body")

	if err := response.Body.Close(); err != nil {
		log.Warn().Err(err).Str("method", method).Str("resource", resource).Msg("Error closing response body")
	}

	if response.StatusCode == 401 {
		var err graphError
		if unmarshalErr := json.Unmarshal(body, &err); unmarshalErr != nil {
			log.Warn().Err(unmarshalErr).Str("method", method).Str("resource", resource).Msg("Failed to unmarshal error response, using default error message")
			err.Error.Code = "UnknownError"
			err.Error.Message = "Failed to parse error response"
		}
		log.Warn().
			Str("method", method).
			Str("resource", resource).
			Str("code", err.Error.Code).
			Str("message", err.Error.Message).
			Msg("Authentication token invalid or new app permissions required, " +
				"forcing reauth before retrying.")

		log.Debug().Str("method", method).Str("resource", resource).Msg("Starting reauth process")
		reauth, authErr := newAuth(ctx, auth.AuthConfig, auth.Path, false)
		if authErr != nil {
			log.Error().Err(authErr).Str("method", method).Str("resource", resource).Msg("Reauth failed")
			return nil, fmt.Errorf("reauth failed: %w", authErr)
		}
		if mergeErr := mergo.Merge(auth, reauth, mergo.WithOverride); mergeErr != nil {
			log.Error().Err(mergeErr).Str("method", method).Str("resource", resource).Msg("Failed to merge auth data")
			return nil, fmt.Errorf("failed to merge auth data: %w", mergeErr)
		}
		request.Header.Set("Authorization", "bearer "+auth.AccessToken)
		log.Debug().Str("method", method).Str("resource", resource).Msg("Reauth process completed")
	}
	if response.StatusCode >= 500 || response.StatusCode == 401 {
		// the onedrive API is having issues, retry once
		log.Debug().Str("method", method).Str("resource", resource).Int("statusCode", response.StatusCode).Msg("Server error or auth issue, retrying request")

		log.Debug().Str("method", method).Str("resource", resource).Msg("Executing retry request")
		response, err = g.httpClient.Do(request)
		if err != nil {
			log.Error().Str("method", method).Str("resource", resource).Err(err).Msg("Retry request failed")
			return nil, err
		}
		log.Debug().Str("method", method).Str("resource", resource).Int("statusCode", response.StatusCode).Msg("Retry request completed")

		log.Debug().Str("method", method).Str("resource", resource).Msg("Reading retry response body")
		body, err = io.ReadAll(response.Body)
		if err != nil {
			log.Error().Str("method", method).Str("resource", resource).Err(err).Msg("Error reading retry response body")
			return nil, fmt.Errorf("error reading retry response body: %v", err)
		}
		log.Debug().Str("method", method).Str("resource", resource).Int("bodySize", len(body)).Msg("Successfully read retry response body")

		if err := response.Body.Close(); err != nil {
			log.Warn().Err(err).Str("method", method).Str("resource", resource).Msg("Error closing retry response body")
		}
	}

	if response.StatusCode >= 400 {
		// something was wrong with the request
		log.Debug().Str("method", method).Str("resource", resource).Int("statusCode", response.StatusCode).Msg("Request failed with error status code")
		var err graphError
		unmarshalErr := json.Unmarshal(body, &err)
		if unmarshalErr != nil {
			log.Error().Str("method", method).Str("resource", resource).Err(unmarshalErr).Msg("Failed to unmarshal error response")
			return nil, fmt.Errorf("HTTP %d - failed to parse error response: %v",
				response.StatusCode, unmarshalErr)
		}
		log.Error().Str("method", method).Str("resource", resource).Int("statusCode", response.StatusCode).
			Str("errorCode", err.Error.Code).Str("errorMessage", err.Error.Message).
			Msg("Request failed with API error")
		return nil, fmt.Errorf("HTTP %d - %s: %s",
			response.StatusCode, err.Error.Code, err.Error.Message)
	}
	log.Debug().Str("method", method).Str("resource", resource).Int("bodySize", len(body)).Msg("Request completed successfully")
	return body, nil
}

// RequestWithContext performs an authenticated request to Microsoft Graph with context
// This is a backward-compatible function that uses the default Graph client
func RequestWithContext(ctx context.Context, resource string, auth *Auth, method string, content io.Reader, headers ...Header) ([]byte, error) {
	return NewGraph().RequestWithContext(ctx, resource, auth, method, content, headers...)
}

// Get is a convenience wrapper around Request
func (g *Graph) Get(resource string, auth *Auth, headers ...Header) ([]byte, error) {
	return g.Request(resource, auth, "GET", nil, headers...)
}

// Get is a convenience wrapper around Request
// This is a backward-compatible function that uses the default Graph client
func Get(resource string, auth *Auth, headers ...Header) ([]byte, error) {
	return NewGraph().Get(resource, auth, headers...)
}

// GetWithContext is a convenience wrapper around RequestWithContext with caching
func (g *Graph) GetWithContext(ctx context.Context, resource string, auth *Auth, headers ...Header) ([]byte, error) {
	// Only cache GET requests without custom headers
	if len(headers) == 0 {
		cache := getResponseCache()

		// Try to get from cache first
		if data, found := cache.Get(resource); found {
			log.Debug().Str("resource", resource).Msg("Cache hit for GET request")
			return data, nil
		}

		// Not in cache, make the request
		data, err := g.RequestWithContext(ctx, resource, auth, "GET", nil)
		if err == nil {
			// Cache the successful response
			cache.Set(resource, data)
			log.Debug().Str("resource", resource).Msg("Cached GET response")
		}
		return data, err
	}

	// For requests with custom headers, bypass cache
	return g.RequestWithContext(ctx, resource, auth, "GET", nil, headers...)
}

// GetWithContext is a convenience wrapper around RequestWithContext with caching
// This is a backward-compatible function that uses the default Graph client
func GetWithContext(ctx context.Context, resource string, auth *Auth, headers ...Header) ([]byte, error) {
	return NewGraph().GetWithContext(ctx, resource, auth, headers...)
}

// invalidateResourceCache invalidates cache entries related to the given resource
func invalidateResourceCache(resource string) {
	cache := getResponseCache()

	// Invalidate the exact resource
	cache.Invalidate(resource)

	// If this is an item, invalidate its parent's children listing
	if strings.Contains(resource, "/items/") {
		parts := strings.Split(resource, "/items/")
		if len(parts) > 1 {
			// Extract the parent path and invalidate its children listing
			parentPath := parts[0]
			cache.InvalidatePrefix(parentPath + "/children")
		}
	}

	// If this is a root operation, invalidate root children
	if resource == "/me/drive/root" || strings.HasPrefix(resource, "/me/drive/root/") {
		cache.InvalidatePrefix("/me/drive/root/children")
	}

	log.Debug().Str("resource", resource).Msg("Invalidated cache entries for modified resource")
}

// Patch is a convenience wrapper around Request
func (g *Graph) Patch(resource string, auth *Auth, content io.Reader, headers ...Header) ([]byte, error) {
	data, err := g.Request(resource, auth, "PATCH", content, headers...)
	if err == nil {
		invalidateResourceCache(resource)
	}
	return data, err
}

// Patch is a convenience wrapper around Request
// This is a backward-compatible function that uses the default Graph client
func Patch(resource string, auth *Auth, content io.Reader, headers ...Header) ([]byte, error) {
	return NewGraph().Patch(resource, auth, content, headers...)
}

// PatchWithContext is a convenience wrapper around RequestWithContext
// This function is intentionally kept for API completeness and potential future use.
// It is currently unused but maintained for API consistency.
// nolint:unused,deadcode
func (g *Graph) PatchWithContext(ctx context.Context, resource string, auth *Auth, content io.Reader, headers ...Header) ([]byte, error) {
	data, err := g.RequestWithContext(ctx, resource, auth, "PATCH", content, headers...)
	if err == nil {
		invalidateResourceCache(resource)
	}
	return data, err
}

// PatchWithContext is a convenience wrapper around RequestWithContext
// This function is intentionally kept for API completeness and potential future use.
// It is currently unused but maintained for API consistency.
// This is a backward-compatible function that uses the default Graph client
// nolint:unused,deadcode
func PatchWithContext(ctx context.Context, resource string, auth *Auth, content io.Reader, headers ...Header) ([]byte, error) {
	return NewGraph().PatchWithContext(ctx, resource, auth, content, headers...)
}

// Post is a convenience wrapper around Request
func (g *Graph) Post(resource string, auth *Auth, content io.Reader, headers ...Header) ([]byte, error) {
	data, err := g.Request(resource, auth, "POST", content, headers...)
	if err == nil {
		invalidateResourceCache(resource)
	}
	return data, err
}

// Post is a convenience wrapper around Request
// This is a backward-compatible function that uses the default Graph client
func Post(resource string, auth *Auth, content io.Reader, headers ...Header) ([]byte, error) {
	return NewGraph().Post(resource, auth, content, headers...)
}

// PostWithContext is a convenience wrapper around RequestWithContext
// This function is intentionally kept for API completeness and potential future use.
// It is currently unused but maintained for API consistency.
// nolint:unused,deadcode
func (g *Graph) PostWithContext(ctx context.Context, resource string, auth *Auth, content io.Reader, headers ...Header) ([]byte, error) {
	data, err := g.RequestWithContext(ctx, resource, auth, "POST", content, headers...)
	if err == nil {
		invalidateResourceCache(resource)
	}
	return data, err
}

// PostWithContext is a convenience wrapper around RequestWithContext
// This function is intentionally kept for API completeness and potential future use.
// It is currently unused but maintained for API consistency.
// This is a backward-compatible function that uses the default Graph client
// nolint:unused,deadcode
func PostWithContext(ctx context.Context, resource string, auth *Auth, content io.Reader, headers ...Header) ([]byte, error) {
	return NewGraph().PostWithContext(ctx, resource, auth, content, headers...)
}

// Put is a convenience wrapper around Request
func (g *Graph) Put(resource string, auth *Auth, content io.Reader, headers ...Header) ([]byte, error) {
	data, err := g.Request(resource, auth, "PUT", content, headers...)
	if err == nil {
		invalidateResourceCache(resource)
	}
	return data, err
}

// Put is a convenience wrapper around Request
// This is a backward-compatible function that uses the default Graph client
func Put(resource string, auth *Auth, content io.Reader, headers ...Header) ([]byte, error) {
	return NewGraph().Put(resource, auth, content, headers...)
}

// PutWithContext is a convenience wrapper around RequestWithContext
// This function is intentionally kept for API completeness and potential future use.
// It is currently unused but maintained for API consistency.
// nolint:unused,deadcode
func (g *Graph) PutWithContext(ctx context.Context, resource string, auth *Auth, content io.Reader, headers ...Header) ([]byte, error) {
	data, err := g.RequestWithContext(ctx, resource, auth, "PUT", content, headers...)
	if err == nil {
		invalidateResourceCache(resource)
	}
	return data, err
}

// PutWithContext is a convenience wrapper around RequestWithContext
// This function is intentionally kept for API completeness and potential future use.
// It is currently unused but maintained for API consistency.
// This is a backward-compatible function that uses the default Graph client
// nolint:unused,deadcode
func PutWithContext(ctx context.Context, resource string, auth *Auth, content io.Reader, headers ...Header) ([]byte, error) {
	return NewGraph().PutWithContext(ctx, resource, auth, content, headers...)
}

// Delete performs an HTTP delete
func (g *Graph) Delete(resource string, auth *Auth, headers ...Header) error {
	_, err := g.Request(resource, auth, "DELETE", nil, headers...)
	if err == nil {
		invalidateResourceCache(resource)
	}
	return err
}

// Delete performs an HTTP delete
// This is a backward-compatible function that uses the default Graph client
func Delete(resource string, auth *Auth, headers ...Header) error {
	return NewGraph().Delete(resource, auth, headers...)
}

// DeleteWithContext performs an HTTP delete with context
// This function is intentionally kept for API completeness and potential future use.
// It is currently unused but maintained for API consistency.
// nolint:unused,deadcode
func (g *Graph) DeleteWithContext(ctx context.Context, resource string, auth *Auth, headers ...Header) error {
	_, err := g.RequestWithContext(ctx, resource, auth, "DELETE", nil, headers...)
	if err == nil {
		invalidateResourceCache(resource)
	}
	return err
}

// DeleteWithContext performs an HTTP delete with context
// This function is intentionally kept for API completeness and potential future use.
// It is currently unused but maintained for API consistency.
// This is a backward-compatible function that uses the default Graph client
// nolint:unused,deadcode
func DeleteWithContext(ctx context.Context, resource string, auth *Auth, headers ...Header) error {
	return NewGraph().DeleteWithContext(ctx, resource, auth, headers...)
}

// IDPath computes the resource path for an item by ID
func IDPath(id string) string {
	if id == "root" {
		return "/me/drive/root"
	}
	return "/me/drive/items/" + url.PathEscape(id)
}

// ResourcePath translates an item's path to the proper path used by Graph
func ResourcePath(path string) string {
	if path == "/" {
		return "/me/drive/root"
	}
	return "/me/drive/root:" + url.PathEscape(path)
}

// ChildrenPath returns the path to an item's children
func childrenPath(path string) string {
	if path == "/" {
		return ResourcePath(path) + "/children"
	}
	return ResourcePath(path) + ":/children"
}

// ChildrenPathID returns the API resource path of an item's children
func childrenPathID(id string) string {
	return fmt.Sprintf("/me/drive/items/%s/children", url.PathEscape(id))
}

// User represents the user. Currently only used to fetch the account email so
// we can display it in file managers with .xdg-volume-info
// https://docs.microsoft.com/en-ca/graph/api/user-get
type User struct {
	UserPrincipalName string `json:"userPrincipalName"`
}

// GetUser fetches the current user details from the Graph API.
func (g *Graph) GetUser(auth *Auth) (User, error) {
	return g.GetUserWithContext(context.Background(), auth)
}

// GetUser fetches the current user details from the Graph API.
// This is a backward-compatible function that uses the default Graph client
func GetUser(auth *Auth) (User, error) {
	return NewGraph().GetUser(auth)
}

// GetUserWithContext fetches the current user details from the Graph API with context.
func (g *Graph) GetUserWithContext(ctx context.Context, auth *Auth) (User, error) {
	resp, err := g.GetWithContext(ctx, "/me", auth)
	user := User{}
	if err == nil {
		err = json.Unmarshal(resp, &user)
	}
	return user, err
}

// GetUserWithContext fetches the current user details from the Graph API with context.
// This is a backward-compatible function that uses the default Graph client
func GetUserWithContext(ctx context.Context, auth *Auth) (User, error) {
	return NewGraph().GetUserWithContext(ctx, auth)
}

// DriveQuota is used to parse the User's current storage quotas from the API
// https://docs.microsoft.com/en-us/onedrive/developer/rest-api/resources/quota
type DriveQuota struct {
	Deleted   uint64 `json:"deleted"`   // bytes in recycle bin
	FileCount uint64 `json:"fileCount"` // unavailable on personal accounts
	Remaining uint64 `json:"remaining"`
	State     string `json:"state"` // normal | nearing | critical | exceeded
	Total     uint64 `json:"total"`
	Used      uint64 `json:"used"`
}

// Drive has some general information about the user's OneDrive
// https://docs.microsoft.com/en-us/onedrive/developer/rest-api/resources/drive
type Drive struct {
	ID        string     `json:"id"`
	DriveType string     `json:"driveType"` // personal | business | documentLibrary
	Quota     DriveQuota `json:"quota,omitempty"`
}

// GetDrive is used to fetch the details of the user's OneDrive.
func (g *Graph) GetDrive(auth *Auth) (Drive, error) {
	resp, err := g.Get("/me/drive", auth)
	drive := Drive{}
	if err != nil {
		return drive, err
	}
	return drive, json.Unmarshal(resp, &drive)
}

// GetDrive is used to fetch the details of the user's OneDrive.
// This is a backward-compatible function that uses the default Graph client
func GetDrive(auth *Auth) (Drive, error) {
	return NewGraph().GetDrive(auth)
}

// SetOperationalOffline sets the operational offline state
func SetOperationalOffline(offline bool) {
	operationalOfflineMutex.Lock()
	defer operationalOfflineMutex.Unlock()
	operationalOffline = offline
	log.Info().Bool("offline", offline).Msg("Set operational offline state")
}

// GetOperationalOffline returns the current operational offline state
func GetOperationalOffline() bool {
	operationalOfflineMutex.RLock()
	defer operationalOfflineMutex.RUnlock()
	return operationalOffline
}

// IsOffline checks if the system is offline.
// It returns true if either:
// 1. The operational offline state is set to true, or
// 2. An error string from Request() is indicative of being offline.
func IsOffline(err error) bool {
	// Check operational offline state first
	if GetOperationalOffline() {
		return true
	}

	// Then check detected offline state based on error
	if err == nil {
		return false
	}
	// our error messages from Request() will be prefixed with "HTTP ### -" if we actually
	// got an HTTP response (indicating we are not offline)
	rexp := regexp.MustCompile("HTTP [0-9]+ - ")
	return !rexp.MatchString(err.Error())
}
