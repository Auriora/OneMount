package graph

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/imdario/mergo"
	"github.com/rs/zerolog/log"
)

// these are default values if not specified
const (
	authClientID    = "3470c3fa-bc10-45ab-a0a9-2d30836485d1"
	authCodeURL     = "https://login.microsoftonline.com/common/oauth2/v2.0/authorize"
	authTokenURL    = "https://login.microsoftonline.com/common/oauth2/v2.0/token"
	authRedirectURL = "https://login.live.com/oauth20_desktop.srf"
)

func (a *AuthConfig) applyDefaults() error {
	return mergo.Merge(a, AuthConfig{
		ClientID:    authClientID,
		CodeURL:     authCodeURL,
		TokenURL:    authTokenURL,
		RedirectURL: authRedirectURL,
	})
}

// AuthConfig configures the authentication flow
type AuthConfig struct {
	ClientID    string `json:"clientID" yaml:"clientID"`
	CodeURL     string `json:"codeURL" yaml:"codeURL"`
	TokenURL    string `json:"tokenURL" yaml:"tokenURL"`
	RedirectURL string `json:"redirectURL" yaml:"redirectURL"`
}

// Auth represents a set of oauth2 authentication tokens
type Auth struct {
	AuthConfig   `json:"config"`
	Account      string `json:"account"`
	ExpiresIn    int64  `json:"expires_in"` // only used for parsing
	ExpiresAt    int64  `json:"expires_at"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	path         string // auth tokens remember their path for use by Refresh()
}

// AuthError is an authentication error from the Microsoft API. Generally we don't see
// these unless something goes catastrophically wrong with Microsoft's authentication
// services.
type AuthError struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
	ErrorCodes       []int  `json:"error_codes"`
	ErrorURI         string `json:"error_uri"`
	Timestamp        string `json:"timestamp"` // json.Unmarshal doesn't like this timestamp format
	TraceID          string `json:"trace_id"`
	CorrelationID    string `json:"correlation_id"`
}

// ToFile writes auth tokens to a file
func (a Auth) ToFile(file string) error {
	a.path = file
	byteData, _ := json.Marshal(a)
	return os.WriteFile(file, byteData, 0600)
}

// FromFile populates an auth struct from a file
func (a *Auth) FromFile(file string) error {
	contents, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	a.path = file
	err = json.Unmarshal(contents, a)
	if err != nil {
		return err
	}
	return a.applyDefaults()
}

// createRefreshTokenRequest creates the HTTP request for refreshing tokens
func (a *Auth) createRefreshTokenRequest() *strings.Reader {
	return strings.NewReader("client_id=" + a.ClientID +
		"&redirect_uri=" + a.RedirectURL +
		"&refresh_token=" + a.RefreshToken +
		"&grant_type=refresh_token")
}

// handleRefreshResponse processes the response from the token refresh request
// Returns true if reauthorization is needed
func (a *Auth) handleRefreshResponse(resp *http.Response, err error) (bool, error) {
	if err != nil {
		if IsOffline(err) || resp == nil {
			log.Trace().Err(err).Msg("Network unreachable during token renewal, ignoring.")
			return false, err
		}
		log.Error().Err(err).Msg("Could not POST to renew tokens, forcing reauth.")
		return true, err
	}

	// put here so as to avoid spamming the log when offline
	log.Info().Msg("Auth tokens expired, attempting renewal.")
	return false, nil
}

// updateTokenExpiration updates token expiration times
func (a *Auth) updateTokenExpiration(oldTime int64) {
	if a.ExpiresAt == oldTime {
		a.ExpiresAt = time.Now().Unix() + a.ExpiresIn
	}
}

// handleFailedRefresh handles the case when token refresh fails
func (a *Auth) handleFailedRefresh(resp *http.Response, body []byte, reauth bool) {
	if reauth || a.AccessToken == "" || a.RefreshToken == "" {
		log.Error().
			Bytes("response", body).
			Int("http_code", resp.StatusCode).
			Msg("Failed to renew access tokens. Attempting to reauthenticate.")

		newAuth, err := newAuth(a.AuthConfig, a.path, false)
		if err != nil {
			log.Error().Err(err).Msg("Failed to reauthenticate. Using existing tokens.")
			return
		}
		*a = *newAuth
	} else {
		err := a.ToFile(a.path)
		if err != nil {
			log.Warn().Err(err).Msg("Failed to save auth tokens to file")
		}
	}
}

// Refresh auth tokens if expired.
func (a *Auth) Refresh() {
	if a.ExpiresAt <= time.Now().Unix() {
		oldTime := a.ExpiresAt
		postData := a.createRefreshTokenRequest()
		resp, err := http.Post(a.TokenURL,
			"application/x-www-form-urlencoded",
			postData)

		reauth, respErr := a.handleRefreshResponse(resp, err)
		if respErr != nil && (IsOffline(err) || resp == nil) {
			return
		}

		if resp != nil {
			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)
			json.Unmarshal(body, &a)
			a.updateTokenExpiration(oldTime)
			a.handleFailedRefresh(resp, body, reauth)
		}
	}
}

// Get the appropriate authentication URL for the Graph OAuth2 challenge.
func getAuthURL(a AuthConfig) string {
	return a.CodeURL +
		"?client_id=" + a.ClientID +
		"&scope=" + url.PathEscape("user.read files.readwrite.all offline_access") +
		"&response_type=code" +
		"&redirect_uri=" + a.RedirectURL
}

// getAuthCodeHeadless has the user perform authentication in their own browser
// instead of WebKit2GTK and then input the auth code in the terminal.
func getAuthCodeHeadless(a AuthConfig, accountName string) (string, error) {
	fmt.Printf("Please visit the following URL:\n%s\n\n", getAuthURL(a))
	fmt.Println("Please enter the redirect URL once you are redirected to a " +
		"blank page (after \"Let this app access your info?\"):")
	var response string
	fmt.Scanln(&response)
	code, err := parseAuthCode(response)
	if err != nil {
		return "", fmt.Errorf("no validation code returned, or code was invalid: %w", err)
	}
	return code, nil
}

// parseAuthCode is used to parse the auth code out of the redirect the server gives us
// after successful authentication
func parseAuthCode(url string) (string, error) {
	rexp := regexp.MustCompile("code=([a-zA-Z0-9-_.]+)")
	code := rexp.FindString(url)
	if len(code) == 0 {
		return "", errors.New("invalid auth code")
	}
	return code[5:], nil
}

// Exchange an auth code for a set of access tokens (returned as a new Auth struct).
func getAuthTokens(a AuthConfig, authCode string) (*Auth, error) {
	postData := strings.NewReader("client_id=" + a.ClientID +
		"&redirect_uri=" + a.RedirectURL +
		"&code=" + authCode +
		"&grant_type=authorization_code")
	resp, err := http.Post(a.TokenURL,
		"application/x-www-form-urlencoded",
		postData)
	if err != nil {
		return nil, fmt.Errorf("could not POST to obtain auth tokens: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("could not read response body: %w", err)
	}

	var auth Auth
	if err := json.Unmarshal(body, &auth); err != nil {
		return nil, fmt.Errorf("could not unmarshal auth response: %w", err)
	}

	if auth.ExpiresAt == 0 {
		auth.ExpiresAt = time.Now().Unix() + auth.ExpiresIn
	}
	auth.AuthConfig = a

	if auth.AccessToken == "" || auth.RefreshToken == "" {
		var authErr AuthError
		var errMsg string

		if err := json.Unmarshal(body, &authErr); err == nil {
			// we got a parseable error message out of microsoft's servers
			errMsg = fmt.Sprintf("Failed to retrieve access tokens: %s - %s",
				authErr.Error, authErr.ErrorDescription)
			log.Error().
				Int("status", resp.StatusCode).
				Str("error", authErr.Error).
				Str("errorDescription", authErr.ErrorDescription).
				Str("helpUrl", authErr.ErrorURI).
				Msg(errMsg)
		} else {
			// things are extra broken and this is an error type we haven't seen before
			errMsg = "Failed to retrieve access tokens with unknown error format"
			log.Error().
				Int("status", resp.StatusCode).
				Bytes("response", body).
				Err(err).
				Msg(errMsg)
		}
		return nil, errors.New(errMsg)
	}
	return &auth, nil
}

// newAuth performs initial authentication flow and saves tokens to disk. The headless
// parameter determines if we will try to auth directly in the terminal instead of
// doing it via embedded browser.
func newAuth(config AuthConfig, path string, headless bool) (*Auth, error) {
	// load the old account name
	old := Auth{}
	_ = old.FromFile(path) // Ignore error, we just want the account name if available

	config.applyDefaults()
	var code string
	var err error

	if headless {
		code, err = getAuthCodeHeadless(config, old.Account)
		if err != nil {
			return nil, fmt.Errorf("failed to get auth code: %w", err)
		}
	} else {
		// in a build without CGO, this will be the same as above
		code, err = getAuthCode(config, old.Account)
		if err != nil {
			return nil, fmt.Errorf("failed to get auth code: %w", err)
		}
	}

	auth, err := getAuthTokens(config, code)
	if err != nil {
		return nil, fmt.Errorf("failed to get auth tokens: %w", err)
	}

	if user, err := GetUser(auth); err == nil {
		auth.Account = user.UserPrincipalName
	}

	if err := auth.ToFile(path); err != nil {
		log.Warn().Err(err).Msg("Failed to save auth tokens to file")
	}

	return auth, nil
}

// Authenticate performs authentication to Graph or load auth/refreshes it
// from an existing file. If headless is true, we will authenticate in the
// terminal.
func Authenticate(config AuthConfig, path string, headless bool) (*Auth, error) {
	auth := &Auth{}
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		// no tokens found, gotta start oauth flow from beginning
		auth, err = newAuth(config, path, headless)
		if err != nil {
			return nil, fmt.Errorf("authentication failed: %w", err)
		}
	} else {
		// we already have tokens, no need to force a new auth flow
		if err := auth.FromFile(path); err != nil {
			return nil, fmt.Errorf("failed to load auth tokens: %w", err)
		}
		auth.Refresh()
	}
	return auth, nil
}
