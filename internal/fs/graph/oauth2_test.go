package graph

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthCodeFormat(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		input          string
		expectedCode   string
		shouldSucceed  bool
		errorMessage   string
		successMessage string
	}{
		{
			name:           "ArbitraryFormat_ShouldParseCorrectly",
			input:          "codecode=asd-965198osATYjfb._knlwoieurow*sdjf",
			expectedCode:   "asd-965198osATYjfb._knlwoieurow",
			shouldSucceed:  true,
			errorMessage:   "Failed to parse arbitrary auth code",
			successMessage: "Auth code parser did not succeed against arbitrary character test",
		},
		{
			name:           "PersonalAuthCode_ShouldParseCorrectly",
			input:          "https://login.live.com/oauth20_desktop.srf?code=M.R3_BAY.abcd526-817f-d8e9-590c-1227b45c7be2&lc=4105",
			expectedCode:   "M.R3_BAY.abcd526-817f-d8e9-590c-1227b45c7be2",
			shouldSucceed:  true,
			errorMessage:   "Failed to parse personal auth code",
			successMessage: "Personal auth code did not match expected result",
		},
		{
			name:           "BusinessAuthCode_ShouldParseCorrectly",
			input:          "https://login.live.com/oauth20_desktop.srf?code=0.BAAA-AeXRPDP_sEe7XktwiDweriowjeirjcDQQvKtFoKktMINkhdEzAAA.AQABAAAAARWERAB2UyzwtQEKR7-rWbgdcBZICdWKCJnfnPJurxUN_QbF3GS6OQqQiK987AbLAv2QykQMIGAz4XCvkO8kB3XC8RYV10qmnmHcMUgo7u5UubpgpR3OW3TVlMSZ-3vxjkcEHlsnVoBqfUFdcj8fYR_mP6w0xkB8MmLG3i5F-JtcaLKfQu13941lsdjkfdh0acjHBGJHVzpBbuiVfzN6vMygFiS2xAQGF668M_l69dXRmG1tq3ZwU6J0-FWYNfK_Ro4YS2m38bcNmZQ8iEolV78t34HKxCYZnl4iqeYF7b7hkTM7ZIcsDBoeZvW1Cu6dIQ7xC4NZGILltOXY5V6A-kcLCZaYuSFW_R8dEM-cqGr_5Gv1GhgfqyXd-2XYNvGda9ok20JrYEmMiezfnyRV-vc7rdtlLOVI_ubzhrjezAvtAApPEj3dJdcmW_0qns_R27pVDlU1xkDagQAquhrftE_sZHbRGvnAsdfaoim1SjcX7QosTELyoWeAczip4MPYqmJ1uVjpWb533vA5WZMyWatiDuNYhnj48SsfEP2zaUQFU55Aj90hEOhOPl77AOu0-zNfAGXeWAQhTPO2rZ0ZgHottFwLoq8aA52sTW-hf7kB0chFUaUvLkxKr1L-Zi7vyCBoArlciFV3zyMxiQ8kjR3vxfwlerjowicmcgqJD-8lxioiwerwlbrlQWyAA&session_state=3fa7b212-7dbb-44e6-bddd-812fwieojw914341",
			expectedCode:   "0.BAAA-AeXRPDP_sEe7XktwiDweriowjeirjcDQQvKtFoKktMINkhdEzAAA.AQABAAAAARWERAB2UyzwtQEKR7-rWbgdcBZICdWKCJnfnPJurxUN_QbF3GS6OQqQiK987AbLAv2QykQMIGAz4XCvkO8kB3XC8RYV10qmnmHcMUgo7u5UubpgpR3OW3TVlMSZ-3vxjkcEHlsnVoBqfUFdcj8fYR_mP6w0xkB8MmLG3i5F-JtcaLKfQu13941lsdjkfdh0acjHBGJHVzpBbuiVfzN6vMygFiS2xAQGF668M_l69dXRmG1tq3ZwU6J0-FWYNfK_Ro4YS2m38bcNmZQ8iEolV78t34HKxCYZnl4iqeYF7b7hkTM7ZIcsDBoeZvW1Cu6dIQ7xC4NZGILltOXY5V6A-kcLCZaYuSFW_R8dEM-cqGr_5Gv1GhgfqyXd-2XYNvGda9ok20JrYEmMiezfnyRV-vc7rdtlLOVI_ubzhrjezAvtAApPEj3dJdcmW_0qns_R27pVDlU1xkDagQAquhrftE_sZHbRGvnAsdfaoim1SjcX7QosTELyoWeAczip4MPYqmJ1uVjpWb533vA5WZMyWatiDuNYhnj48SsfEP2zaUQFU55Aj90hEOhOPl77AOu0-zNfAGXeWAQhTPO2rZ0ZgHottFwLoq8aA52sTW-hf7kB0chFUaUvLkxKr1L-Zi7vyCBoArlciFV3zyMxiQ8kjR3vxfwlerjowicmcgqJD-8lxioiwerwlbrlQWyAA",
			shouldSucceed:  true,
			errorMessage:   "Failed to parse business auth code",
			successMessage: "Business auth code did not match expected result",
		},
		{
			name:           "InvalidFormat_ShouldReturnError",
			input:          "invalid-format-without-code",
			expectedCode:   "",
			shouldSucceed:  false,
			errorMessage:   "Expected error for invalid auth code format",
			successMessage: "",
		},
	}

	for _, tc := range tests {
		tc := tc // capture range variable
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			code, err := parseAuthCode(tc.input)

			if tc.shouldSucceed {
				require.NoError(t, err, tc.errorMessage)
				assert.Equal(t, tc.expectedCode, code, tc.successMessage)
			} else {
				assert.Error(t, err, tc.errorMessage)
			}
		})
	}
}

func TestAuthFromfile(t *testing.T) {
	t.Parallel()
	require.FileExists(t, ".auth_tokens.json")

	var auth Auth
	auth.FromFile(".auth_tokens.json")
	assert.NotEqual(t, "", auth.AccessToken, "Could not load auth tokens from '.auth_tokens.json'!")
}

func TestAuthRefresh(t *testing.T) {
	t.Parallel()
	require.FileExists(t, ".auth_tokens.json")

	var auth Auth
	auth.FromFile(".auth_tokens.json")
	auth.ExpiresAt = 0 // force an auth refresh
	auth.Refresh(nil)  // nil context will use context.Background() internally
	require.Greater(t, auth.ExpiresAt, time.Now().Unix(), "Auth could not be refreshed successfully!")
}

func TestAuthConfigMerge(t *testing.T) {
	t.Parallel()

	testConfig := AuthConfig{RedirectURL: "test"}
	require.NoError(t, testConfig.applyDefaults(), "Failed to apply defaults to AuthConfig")
	assert.Equal(t, "test", testConfig.RedirectURL)
	assert.Equal(t, authClientID, testConfig.ClientID)
}

// TestAuthFailureWithNetworkAvailable tests the behavior when authentication fails but network is available (TC-15)
func TestAuthFailureWithNetworkAvailable(t *testing.T) {
	t.Parallel()

	// Create an Auth with invalid credentials but valid configuration
	invalidAuth := &Auth{
		AuthConfig: AuthConfig{
			ClientID:    authClientID,
			RedirectURL: authRedirectURL,
			TokenURL:    authTokenURL,
			CodeURL:     authCodeURL,
		},
		AccessToken:  "invalid_access_token",
		RefreshToken: "invalid_refresh_token",
		ExpiresAt:    0, // Force a refresh attempt
	}

	// Apply defaults to ensure the configuration is valid
	require.NoError(t, invalidAuth.AuthConfig.applyDefaults(), "Failed to apply defaults to AuthConfig")

	// Attempt to refresh the tokens, which should fail due to invalid credentials
	err := invalidAuth.Refresh(nil) // nil context will use context.Background() internally

	// Verify that an error is returned
	assert.Error(t, err, "Expected error when refreshing with invalid credentials")
	assert.Contains(t, err.Error(), "failed to refresh token", "Error message should indicate refresh failure")

	// Verify that the auth state is still invalid (tokens not updated)
	assert.Equal(t, "invalid_access_token", invalidAuth.AccessToken, "Access token should not be updated")
	assert.Equal(t, "invalid_refresh_token", invalidAuth.RefreshToken, "Refresh token should not be updated")
	assert.Equal(t, int64(0), invalidAuth.ExpiresAt, "Expiration time should not be updated")
}
