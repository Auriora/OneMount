// Package graph provides OneDrive API integration, authentication, and related helpers for interacting with Microsoft Graph.
// Package graph provides the basic APIs to interact with Microsoft Graph. This includes
// Package graph provides the basic APIs to interact with Microsoft Graph. This includes
// the DriveItem resource and supporting resources which are the basis of working with
// files and folders through the Microsoft Graph API.
package graph

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/auriora/onemount/pkg/errors"
	"github.com/auriora/onemount/pkg/logging"
	"github.com/auriora/onemount/pkg/retry"
	"github.com/imdario/mergo"
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

	// isMockClient is a flag that indicates whether a mock client is being used
	// This is used to bypass the offline mode check in tests
	isMockClient      = false
	isMockClientMutex sync.RWMutex
)

func init() {
	httpClient = getSharedHTTPClient()
}

func SetHTTPClient(client *http.Client) {
	if client == nil {
		httpClient = getSharedHTTPClient()
		isMockClientMutex.Lock()
		isMockClient = false
		isMockClientMutex.Unlock()
		return
	}

	// Check if the client's transport is a MockGraphClient
	_, isMock := client.Transport.(*MockGraphClient)
	isMockClientMutex.Lock()
	isMockClient = isMock
	isMockClientMutex.Unlock()

	httpClient = client
}

func getHTTPClient() HTTPClient {
	return httpClient
}

// Graph represents a client for interacting with the Microsoft Graph API.
// getResponseCache returns the shared response cache
func getResponseCache() *ResponseCache {
	cacheOnce.Do(func() {
		// Create the response cache with a default TTL of 5 minutes
		responseCache = NewResponseCache(5 * time.Minute)
		logging.Info().Msg("Initialized response cache with 5-minute TTL")
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
func Request(resource string, auth *Auth, method string, content io.Reader, headers ...Header) ([]byte, error) {
	return RequestWithContext(context.Background(), resource, auth, method, content, headers...)
}

// executeRequest executes an HTTP request and processes the response
func executeRequest(ctx context.Context, request *http.Request, auth *Auth, logCtx logging.LogContext) ([]byte, error) {
	logging.LogDebugWithContext(logCtx, "About to execute HTTP request")
	response, err := httpClient.Do(request)
	if err != nil {
		// Check if the error was due to context cancellation
		if ctx.Err() != nil {
			logging.LogDebugWithContext(logCtx, "Network request cancelled by context")
			return nil, ctx.Err()
		}
		// the actual request failed for other reasons
		// Add context to error message for better troubleshooting
		networkErr := errors.NewNetworkError("network request failed", err)
		logging.LogErrorWithContext(networkErr, logCtx, "Network request failed")
		return nil, networkErr
	}

	// Update log context with status code
	logCtx = logCtx.With("status_code", response.StatusCode)
	logging.LogDebugWithContext(logCtx, "Network request completed")

	logging.LogDebugWithContext(logCtx, "Starting to read response body")
	body, err := io.ReadAll(response.Body)
	if err != nil {
		readErr := errors.Wrap(err, "error reading response body")
		logging.LogErrorWithContext(readErr, logCtx, "Error reading response body")
		return nil, readErr
	}

	// Update log context with body size
	logCtx = logCtx.With("body_size", len(body))
	logging.LogDebugWithContext(logCtx, "Successfully read response body")

	if err := response.Body.Close(); err != nil {
		// Add context to error message for better troubleshooting
		logging.LogErrorAsWarnWithContext(err, logCtx, "Error closing response body")
	}

	// Handle authentication errors
	if response.StatusCode == 401 {
		var err graphError
		if unmarshalErr := json.Unmarshal(body, &err); unmarshalErr != nil {
			// Add context to error message for better troubleshooting
			logging.LogErrorAsWarnWithContext(unmarshalErr, logCtx, "Failed to unmarshal error response, using default error message")
			err.Error.Code = "UnknownError"
			err.Error.Message = "Failed to parse error response"
		}

		// Update log context with error details
		logCtx = logCtx.With("error_code", err.Error.Code).With("error_message", err.Error.Message)

		// Add context to error message for better troubleshooting
		logging.LogErrorAsWarnWithContext(nil, logCtx, "Authentication token invalid or new app permissions required, forcing reauth before retrying")

		logging.LogDebugWithContext(logCtx, "Starting reauth process")
		reauth, authErr := newAuth(ctx, auth.AuthConfig, auth.Path, false)
		if authErr != nil {
			reauthErr := errors.NewAuthError("reauth failed", authErr)
			logging.LogErrorWithContext(reauthErr, logCtx, "Reauth failed")
			return nil, reauthErr
		}
		if mergeErr := mergo.Merge(auth, reauth, mergo.WithOverride); mergeErr != nil {
			mergeAuthErr := errors.Wrap(mergeErr, "failed to merge auth data")
			logging.LogErrorWithContext(mergeAuthErr, logCtx, "Failed to merge auth data")
			return nil, mergeAuthErr
		}
		request.Header.Set("Authorization", "bearer "+auth.AccessToken)
		logging.LogDebugWithContext(logCtx, "Reauth process completed")

		// Return an auth error to trigger a retry
		return nil, errors.NewAuthError("authentication token refreshed, retry needed", nil)
	}

	// Handle error responses
	if response.StatusCode >= 400 {
		// something was wrong with the request
		logging.LogDebugWithContext(logCtx, "Request failed with error status code")
		var err graphError
		unmarshalErr := json.Unmarshal(body, &err)
		if unmarshalErr != nil {
			parseErr := errors.Wrap(unmarshalErr, fmt.Sprintf("HTTP %d - failed to parse error response", response.StatusCode))
			logging.LogErrorWithContext(parseErr, logCtx, "Failed to unmarshal error response")
			return nil, parseErr
		}

		// Update log context with error details
		logCtx = logCtx.With("error_code", err.Error.Code).With("error_message", err.Error.Message)
		logging.LogErrorWithContext(nil, logCtx, "Request failed with API error")

		// Create appropriate error type based on status code
		errorMsg := fmt.Sprintf("%s: %s", err.Error.Code, err.Error.Message)
		var apiErr error

		switch {
		case response.StatusCode == 404:
			apiErr = errors.NewNotFoundError(errorMsg, nil)
		case response.StatusCode == 401 || response.StatusCode == 403:
			apiErr = errors.NewAuthError(errorMsg, nil)
		case response.StatusCode == 400:
			apiErr = errors.NewValidationError(errorMsg, nil)
		case response.StatusCode == 429:
			// Create a resource busy error for rate limiting
			apiErr = errors.NewResourceBusyError(errorMsg, nil)
			// Extract retry-after header if present
			if retryAfter := response.Header.Get("Retry-After"); retryAfter != "" {
				logging.LogInfoWithContext(logCtx, "Rate limit detected with Retry-After header: "+retryAfter)
			}
		case response.StatusCode >= 500:
			apiErr = errors.NewOperationError(errorMsg, nil)
		default:
			apiErr = errors.New(fmt.Sprintf("HTTP %d - %s", response.StatusCode, errorMsg))
		}

		logging.LogErrorWithContext(apiErr, logCtx, "Returning API error")
		return nil, apiErr
	}

	logging.LogDebugWithContext(logCtx, "Request completed successfully")
	return body, nil
}

// RequestWithContextAndCallback performs an authenticated request to Microsoft Graph with context
// and calls the provided callback when the request completes
func RequestWithContextAndCallback(ctx context.Context, resource string, auth *Auth, method string, content io.Reader, callback func([]byte, error), headers ...Header) {
	// Create a log context for this request
	logCtx := logging.NewLogContext("graph_request").
		WithMethod("RequestWithContextAndCallback").
		WithPath(resource).
		With("http_method", method)

	// Check if we're in operational offline mode
	isMockClientMutex.RLock()
	mockClient := isMockClient
	isMockClientMutex.RUnlock()

	if GetOperationalOffline() && !mockClient {
		logging.LogDebugWithContext(logCtx, "In operational offline mode, returning network error")
		callback(nil, errors.NewNetworkError("operational offline mode is enabled", nil))
		return
	}

	if auth == nil || auth.AccessToken == "" {
		// a catch all condition to avoid wiping our auth by accident
		authErr := errors.NewAuthError("cannot make a request with empty auth", nil)
		logging.LogErrorWithContext(authErr, logCtx, "Auth was empty and we attempted to make a request with it!")
		callback(nil, authErr)
		return
	}

	logging.LogDebugWithContext(logCtx, "Starting auth refresh")
	if err := auth.Refresh(ctx); err != nil {
		// Add context to error message for better troubleshooting
		logging.LogErrorAsWarnWithContext(err, logCtx, "Auth refresh failed, continuing with current token")
	}
	logging.LogDebugWithContext(logCtx, "Auth refresh completed")

	logging.LogDebugWithContext(logCtx, "Using HTTP client")
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

	// Update log context with URL
	logCtx = logCtx.With("url", GraphURL+resource)

	logging.LogDebugWithContext(logCtx, "Starting network request with context")

	// Create a retry config with appropriate settings for Graph API
	retryConfig := retry.Config{
		MaxRetries:   5,                // Increase from default 3 to 5 for better handling of rate limits
		InitialDelay: 1 * time.Second,  // Start with a 1-second delay
		MaxDelay:     60 * time.Second, // Allow up to 60-second delays for severe rate limiting
		Multiplier:   2.0,              // Double the delay after each retry
		Jitter:       0.2,              // Add up to 20% random jitter to avoid thundering herd
		RetryableErrors: []retry.RetryableError{
			retry.IsRetryableNetworkError,
			retry.IsRetryableServerError,
			retry.IsRetryableRateLimitError,
		},
	}

	// Create a retryable function that captures the request and auth
	retryableFunc := func() ([]byte, error) {
		return executeRequest(ctx, request, auth, logCtx)
	}

	// Execute the request with retries
	body, err := retry.DoWithResult(ctx, retryableFunc, retryConfig)

	// If we got a rate limit error after all retries, queue the request for later execution
	if err != nil && IsRateLimited(err) {
		logging.Warn().
			Str("resource", resource).
			Str("method", method).
			Msg("Request rate limited after multiple retries, queuing for later execution")

		QueueRequestWithCallback(ctx, resource, auth, method, content, callback, headers...)
		return
	}

	// Call the callback with the result
	callback(body, err)
}

// RequestWithContext performs an authenticated request to Microsoft Graph with context
func RequestWithContext(ctx context.Context, resource string, auth *Auth, method string, content io.Reader, headers ...Header) ([]byte, error) {
	// Create a channel to receive the result
	resultChan := make(chan struct {
		data []byte
		err  error
	}, 1)

	// Create a callback that sends the result to the channel
	callback := func(data []byte, err error) {
		resultChan <- struct {
			data []byte
			err  error
		}{data, err}
	}

	// Execute the request with the callback
	RequestWithContextAndCallback(ctx, resource, auth, method, content, callback, headers...)

	// Wait for the result or context cancellation
	select {
	case result := <-resultChan:
		return result.data, result.err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// Get is a convenience wrapper around Request
func Get(resource string, auth *Auth, headers ...Header) ([]byte, error) {
	return Request(resource, auth, "GET", nil, headers...)
}

// GetWithContext is a convenience wrapper around RequestWithContext with caching
func GetWithContext(ctx context.Context, resource string, auth *Auth, headers ...Header) ([]byte, error) {
	// Only cache GET requests without custom headers
	if len(headers) == 0 {
		cache := getResponseCache()

		// Try to get from cache first
		if data, found := cache.Get(resource); found {
			logging.Debug().Str("resource", resource).Msg("Cache hit for GET request")
			return data, nil
		}

		// Not in cache, make the request
		data, err := RequestWithContext(ctx, resource, auth, "GET", nil)
		if err == nil {
			// Cache the successful response
			cache.Set(resource, data)
			logging.Debug().Str("resource", resource).Msg("Cached GET response")
		}
		return data, err
	}

	// For requests with custom headers, bypass cache
	return RequestWithContext(ctx, resource, auth, "GET", nil, headers...)
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

	logging.Debug().Str("resource", resource).Msg("Invalidated cache entries for modified resource")
}

// Patch is a convenience wrapper around Request
func Patch(resource string, auth *Auth, content io.Reader, headers ...Header) ([]byte, error) {
	data, err := Request(resource, auth, "PATCH", content, headers...)
	if err == nil {
		invalidateResourceCache(resource)
	}
	return data, err
}

// PatchWithContext is a convenience wrapper around RequestWithContext
// This function is intentionally kept for API completeness and potential future use.
// It is currently unused but maintained for API consistency.
// nolint:unused,deadcode
func PatchWithContext(ctx context.Context, resource string, auth *Auth, content io.Reader, headers ...Header) ([]byte, error) {
	data, err := RequestWithContext(ctx, resource, auth, "PATCH", content, headers...)
	if err == nil {
		invalidateResourceCache(resource)
	}
	return data, err
}

// Post is a convenience wrapper around Request
func Post(resource string, auth *Auth, content io.Reader, headers ...Header) ([]byte, error) {
	data, err := Request(resource, auth, "POST", content, headers...)
	if err == nil {
		invalidateResourceCache(resource)
	}
	return data, err
}

// PostWithContext is a convenience wrapper around RequestWithContext
// This function is intentionally kept for API completeness and potential future use.
// It is currently unused but maintained for API consistency.
// nolint:unused,deadcode
func PostWithContext(ctx context.Context, resource string, auth *Auth, content io.Reader, headers ...Header) ([]byte, error) {
	data, err := RequestWithContext(ctx, resource, auth, "POST", content, headers...)
	if err == nil {
		invalidateResourceCache(resource)
	}
	return data, err
}

// Put is a convenience wrapper around Request
func Put(resource string, auth *Auth, content io.Reader, headers ...Header) ([]byte, error) {
	data, err := Request(resource, auth, "PUT", content, headers...)
	if err == nil {
		invalidateResourceCache(resource)
	}
	return data, err
}

// PutWithContext is a convenience wrapper around RequestWithContext
// This function is intentionally kept for API completeness and potential future use.
// It is currently unused but maintained for API consistency.
// nolint:unused,deadcode
func PutWithContext(ctx context.Context, resource string, auth *Auth, content io.Reader, headers ...Header) ([]byte, error) {
	data, err := RequestWithContext(ctx, resource, auth, "PUT", content, headers...)
	if err == nil {
		invalidateResourceCache(resource)
	}
	return data, err
}

// Delete performs an HTTP delete
func Delete(resource string, auth *Auth, headers ...Header) error {
	_, err := Request(resource, auth, "DELETE", nil, headers...)
	if err == nil {
		invalidateResourceCache(resource)
	}
	return err
}

// DeleteWithContext performs an HTTP delete with context
// This function is intentionally kept for API completeness and potential future use.
// It is currently unused but maintained for API consistency.
// nolint:unused,deadcode
func DeleteWithContext(ctx context.Context, resource string, auth *Auth, headers ...Header) error {
	_, err := RequestWithContext(ctx, resource, auth, "DELETE", nil, headers...)
	if err == nil {
		invalidateResourceCache(resource)
	}
	return err
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
func GetUser(auth *Auth) (User, error) {
	return GetUserWithContext(context.Background(), auth)
}

// GetUserWithContext fetches the current user details from the Graph API with context.
func GetUserWithContext(ctx context.Context, auth *Auth) (User, error) {
	resp, err := GetWithContext(ctx, "/me", auth)
	user := User{}
	if err == nil {
		err = json.Unmarshal(resp, &user)
	}
	return user, err
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
func GetDrive(auth *Auth) (Drive, error) {
	resp, err := Get("/me/drive", auth)
	drive := Drive{}
	if err != nil {
		return drive, err
	}
	return drive, json.Unmarshal(resp, &drive)
}

// SetOperationalOffline sets the operational offline state
func SetOperationalOffline(offline bool) {
	operationalOfflineMutex.Lock()
	defer operationalOfflineMutex.Unlock()
	operationalOffline = offline
	logging.Info().Bool("offline", offline).Msg("Set operational offline state")
}

// GetOperationalOffline returns the current operational offline state
func GetOperationalOffline() bool {
	operationalOfflineMutex.RLock()
	defer operationalOfflineMutex.RUnlock()
	return operationalOffline
}

// IsMockClient returns true if a mock client is being used
func IsMockClient() bool {
	isMockClientMutex.RLock()
	defer isMockClientMutex.RUnlock()
	return isMockClient
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
