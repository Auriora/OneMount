package graph

import (
	"github.com/auriora/onemount/internal/testutil/framework"
	"testing"
)

// TestUT_GR_01_01_ResourcePath_VariousInputs_ReturnsEscapedPath tests the ResourcePath function with various inputs.
//
//	Test Case ID    UT-GR-01-01
//	Title           Resource Path Formatting
//	Description     Tests the ResourcePath function with various inputs
//	Preconditions   None
//	Steps           1. Call ResourcePath with different path inputs
//	                2. Compare the result with the expected output
//	Expected Result The escaped path matches the expected format for Microsoft Graph API
//	Notes: This test verifies that the ResourcePath function correctly formats paths for the Microsoft Graph API.
func TestUT_GR_01_01_ResourcePath_VariousInputs_ReturnsEscapedPath(t *testing.T) {
	// Create a test fixture
	fixture := framework.NewUnitTestFixture("ResourcePathFixture")

	// Use the fixture to run the test
	fixture.Use(t, func(t *testing.T, fixture interface{}) {
		// Create assertions helper
		assert := framework.NewAssert(t)

		// Define test cases with input paths and expected outputs
		testCases := []struct {
			name     string
			path     string
			expected string
		}{
			{
				name:     "Root path",
				path:     "/",
				expected: "/drive/root",
			},
			{
				name:     "Simple path",
				path:     "/documents",
				expected: "/drive/root:/documents",
			},
			{
				name:     "Nested path",
				path:     "/documents/reports",
				expected: "/drive/root:/documents/reports",
			},
			{
				name:     "Path with spaces",
				path:     "/my documents/report.docx",
				expected: "/drive/root:/my documents/report.docx",
			},
			{
				name:     "Path with special characters",
				path:     "/documents/report-2023_final.docx",
				expected: "/drive/root:/documents/report-2023_final.docx",
			},
		}

		// Run each test case
		for _, tc := range testCases {
			// Call ResourcePath with the input path
			result := ResourcePath(tc.path)

			// Compare the result with the expected output
			assert.Equal(tc.expected, result, "ResourcePath(%s) should return %s, but got %s", tc.path, tc.expected, result)
		}
	})
}

// TestUT_GR_02_01_Request_UnauthenticatedUser_ReturnsError tests the behavior of an unauthenticated request.
//
//	Test Case ID    UT-GR-02-01
//	Title           Unauthenticated Request Handling
//	Description     Tests the behavior of an unauthenticated request
//	Preconditions   None
//	Steps           1. Create an Auth object with expired token
//	                2. Attempt to make a GET request
//	                3. Check if an error is returned
//	Expected Result An error is returned for the unauthenticated request
//	Notes: This test verifies that the request functions correctly handle unauthenticated requests.
func TestUT_GR_02_01_Request_UnauthenticatedUser_ReturnsError(t *testing.T) {
	// Create a test fixture
	fixture := framework.NewUnitTestFixture("UnauthenticatedRequestFixture")

	// Set up the fixture
	fixture.WithSetup(func(t *testing.T) (interface{}, error) {
		// Create an Auth object with expired token
		auth := &Auth{
			AccessToken:  "expired-token",
			RefreshToken: "refresh-token",
			ExpiresAt:    0, // Expired
		}
		return auth, nil
	})

	// Use the fixture to run the test
	fixture.Use(t, func(t *testing.T, fixture interface{}) {
		// Create assertions helper
		assert := framework.NewAssert(t)

		// Get the Auth object from the fixture
		auth, ok := fixture.(*Auth)
		assert.True(ok, "Expected fixture to be of type *Auth, but got %T", fixture)

		// Note: We're using the Request function directly, so we don't need a mock client

		// Step 1: Verify the Auth object has an expired token
		assert.Equal("expired-token", auth.AccessToken, "Access token should be 'expired-token'")
		assert.Equal(int64(0), auth.ExpiresAt, "ExpiresAt should be 0 (expired)")

		// Step 2: Attempt to make a GET request
		// Define a test resource path
		resourcePath := "/me/drive/root"

		// Attempt to make a GET request with the expired token
		// We'll use the Request function which takes an Auth object
		_, err := Request(resourcePath, auth, "GET", nil)

		// Step 3: Check if an error is returned
		assert.Error(err, "Expected an error when making a request with an expired token")
		assert.Contains(err.Error(), "unauthorized", "Error should indicate unauthorized access")
	})
}
