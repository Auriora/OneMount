package graph

import (
	"testing"

	"github.com/auriora/onemount/internal/testutil/framework"
)

// TestUT_GR_17_02_URIGetHost_VariousURIs_ReturnsCorrectHost tests the uriGetHost function with various inputs.
//
//	Test Case ID    UT-GR-17-01
//	Title           URI Host Extraction
//	Description     Tests the uriGetHost function with various inputs
//	Preconditions   None
//	Steps           1. Call uriGetHost with an invalid URI
//	                2. Call uriGetHost with a valid HTTPS URI with a path
//	                3. Call uriGetHost with a valid HTTP URI without a path
//	                4. Check if the results match expectations
//	Expected Result uriGetHost returns the correct host for valid URIs and an empty string for invalid URIs
//	Notes: This test verifies that the uriGetHost function correctly extracts the host from URIs.
func TestUT_GR_17_02_URIGetHost_VariousURIs_ReturnsCorrectHost(t *testing.T) {
	// Create a test fixture
	fixture := framework.NewUnitTestFixture("URIGetHostFixture")

	// Use the fixture to run the test
	fixture.Use(t, func(t *testing.T, fixture interface{}) {
		assert := framework.NewAssert(t)

		// Test case 1: Invalid URI — should return empty string
		result := uriGetHost("not a valid uri")
		assert.Equal("", result, "Invalid URI should return empty string")

		// Test case 2: Valid HTTPS URI with a path
		result = uriGetHost("https://login.microsoftonline.com/common/oauth2/v2.0/authorize")
		assert.Equal("login.microsoftonline.com", result, "HTTPS URI should return correct host")

		// Test case 3: Valid HTTP URI without a path
		result = uriGetHost("http://example.com")
		assert.Equal("example.com", result, "HTTP URI without path should return correct host")

		// Test case 4: HTTPS URI with port
		result = uriGetHost("https://example.com:8443/path")
		assert.Equal("example.com", result, "URI with port should return host without port")

		// Test case 5: Empty string
		result = uriGetHost("")
		assert.Equal("", result, "Empty string should return empty string")
	})
}
