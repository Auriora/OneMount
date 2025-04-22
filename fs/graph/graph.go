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
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"time"

	"github.com/imdario/mergo"
	"github.com/rs/zerolog/log"
)

// GraphURL is the API endpoint of Microsoft Graph
const GraphURL = "https://graph.microsoft.com/v1.0"

// graphError is an internal struct used when decoding Graph's error messages
type graphError struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// This is an additional header that can be specified to Request
type Header struct {
	key, value string
}

// Request performs an authenticated request to Microsoft Graph
func Request(resource string, auth *Auth, method string, content io.Reader, headers ...Header) ([]byte, error) {
	return RequestWithContext(context.Background(), resource, auth, method, content, headers...)
}

// RequestWithContext performs an authenticated request to Microsoft Graph with context
func RequestWithContext(ctx context.Context, resource string, auth *Auth, method string, content io.Reader, headers ...Header) ([]byte, error) {
	if auth == nil || auth.AccessToken == "" {
		// a catch all condition to avoid wiping our auth by accident
		log.Error().Msg("Auth was empty and we attempted to make a request with it!")
		return nil, errors.New("cannot make a request with empty auth")
	}

	log.Debug().Str("method", method).Str("resource", resource).Msg("Starting auth refresh")
	auth.Refresh()
	log.Debug().Str("method", method).Str("resource", resource).Msg("Auth refresh completed")

	log.Debug().Str("method", method).Str("resource", resource).Msg("Creating HTTP client with 60s timeout")
	client := &http.Client{Timeout: 60 * time.Second}
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
	response, err := client.Do(request)
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
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Error().Str("method", method).Str("resource", resource).Err(err).Msg("Error reading response body")
		return nil, fmt.Errorf("error reading response body: %v", err)
	}
	log.Debug().Str("method", method).Str("resource", resource).Int("bodySize", len(body)).Msg("Successfully read response body")

	response.Body.Close()

	if response.StatusCode == 401 {
		var err graphError
		json.Unmarshal(body, &err)
		log.Warn().
			Str("method", method).
			Str("resource", resource).
			Str("code", err.Error.Code).
			Str("message", err.Error.Message).
			Msg("Authentication token invalid or new app permissions required, " +
				"forcing reauth before retrying.")

		log.Debug().Str("method", method).Str("resource", resource).Msg("Starting reauth process")
		reauth := newAuth(auth.AuthConfig, auth.path, false)
		mergo.Merge(auth, reauth, mergo.WithOverride)
		request.Header.Set("Authorization", "bearer "+auth.AccessToken)
		log.Debug().Str("method", method).Str("resource", resource).Msg("Reauth process completed")
	}
	if response.StatusCode >= 500 || response.StatusCode == 401 {
		// the onedrive API is having issues, retry once
		log.Debug().Str("method", method).Str("resource", resource).Int("statusCode", response.StatusCode).Msg("Server error or auth issue, retrying request")

		log.Debug().Str("method", method).Str("resource", resource).Msg("Executing retry request")
		response, err = client.Do(request)
		if err != nil {
			log.Error().Str("method", method).Str("resource", resource).Err(err).Msg("Retry request failed")
			return nil, err
		}
		log.Debug().Str("method", method).Str("resource", resource).Int("statusCode", response.StatusCode).Msg("Retry request completed")

		log.Debug().Str("method", method).Str("resource", resource).Msg("Reading retry response body")
		body, err = ioutil.ReadAll(response.Body)
		if err != nil {
			log.Error().Str("method", method).Str("resource", resource).Err(err).Msg("Error reading retry response body")
			return nil, fmt.Errorf("error reading retry response body: %v", err)
		}
		log.Debug().Str("method", method).Str("resource", resource).Int("bodySize", len(body)).Msg("Successfully read retry response body")

		response.Body.Close()
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

// Get is a convenience wrapper around Request
func Get(resource string, auth *Auth, headers ...Header) ([]byte, error) {
	return Request(resource, auth, "GET", nil, headers...)
}

// GetWithContext is a convenience wrapper around RequestWithContext
func GetWithContext(ctx context.Context, resource string, auth *Auth, headers ...Header) ([]byte, error) {
	return RequestWithContext(ctx, resource, auth, "GET", nil, headers...)
}

// Patch is a convenience wrapper around Request
func Patch(resource string, auth *Auth, content io.Reader, headers ...Header) ([]byte, error) {
	return Request(resource, auth, "PATCH", content, headers...)
}

// PatchWithContext is a convenience wrapper around RequestWithContext
func PatchWithContext(ctx context.Context, resource string, auth *Auth, content io.Reader, headers ...Header) ([]byte, error) {
	return RequestWithContext(ctx, resource, auth, "PATCH", content, headers...)
}

// Post is a convenience wrapper around Request
func Post(resource string, auth *Auth, content io.Reader, headers ...Header) ([]byte, error) {
	return Request(resource, auth, "POST", content, headers...)
}

// PostWithContext is a convenience wrapper around RequestWithContext
func PostWithContext(ctx context.Context, resource string, auth *Auth, content io.Reader, headers ...Header) ([]byte, error) {
	return RequestWithContext(ctx, resource, auth, "POST", content, headers...)
}

// Put is a convenience wrapper around Request
func Put(resource string, auth *Auth, content io.Reader, headers ...Header) ([]byte, error) {
	return Request(resource, auth, "PUT", content, headers...)
}

// PutWithContext is a convenience wrapper around RequestWithContext
func PutWithContext(ctx context.Context, resource string, auth *Auth, content io.Reader, headers ...Header) ([]byte, error) {
	return RequestWithContext(ctx, resource, auth, "PUT", content, headers...)
}

// Delete performs an HTTP delete
func Delete(resource string, auth *Auth, headers ...Header) error {
	_, err := Request(resource, auth, "DELETE", nil, headers...)
	return err
}

// DeleteWithContext performs an HTTP delete with context
func DeleteWithContext(ctx context.Context, resource string, auth *Auth, headers ...Header) error {
	_, err := RequestWithContext(ctx, resource, auth, "DELETE", nil, headers...)
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
	resp, err := Get("/me", auth)
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

// IsOffline checks if an error string from Request() is indicative of being offline.
func IsOffline(err error) bool {
	if err == nil {
		return false
	}
	// our error messages from Request() will be prefixed with "HTTP ### -" if we actually
	// got an HTTP response (indicating we are not offline)
	rexp := regexp.MustCompile("HTTP [0-9]+ - ")
	return !rexp.MatchString(err.Error())
}
