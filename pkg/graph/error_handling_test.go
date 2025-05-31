package graph

import (
	"net/http"
	"strings"
	"testing"

	"github.com/auriora/onemount/pkg/errors"
	"github.com/stretchr/testify/assert"
)

// TestUT_GR_ERR_01_01_NotFoundError_InvalidResource_ReturnsNotFoundError tests 404 error handling
func TestUT_GR_ERR_01_01_NotFoundError_InvalidResource_ReturnsNotFoundError(t *testing.T) {
	client := NewMockGraphClient()

	resource := "/me/drive/items/nonexistent-id"
	errorResponse := `{"error":{"code":"itemNotFound","message":"The resource could not be found."}}`
	client.AddMockResponse(resource, []byte(errorResponse), http.StatusNotFound,
		errors.NewNotFoundError("The resource could not be found", nil))

	_, err := client.Get(resource)

	assert.Error(t, err)
	assert.True(t, errors.IsNotFoundError(err))
	assert.Contains(t, err.Error(), "could not be found")
}

// TestUT_GR_ERR_01_02_AuthError_UnauthorizedRequest_ReturnsAuthError tests 401 error handling
func TestUT_GR_ERR_01_02_AuthError_UnauthorizedRequest_ReturnsAuthError(t *testing.T) {
	client := NewMockGraphClient()

	resource := "/me/drive/items/test-id"
	errorResponse := `{"error":{"code":"InvalidAuthenticationToken","message":"Access token is empty."}}`
	client.AddMockResponse(resource, []byte(errorResponse), http.StatusUnauthorized,
		errors.NewAuthError("Access token is empty", nil))

	_, err := client.Get(resource)

	assert.Error(t, err)
	assert.True(t, errors.IsAuthError(err))
	assert.Contains(t, err.Error(), "Access token")
}

// TestUT_GR_ERR_01_03_ValidationError_BadRequest_ReturnsValidationError tests 400 error handling
func TestUT_GR_ERR_01_03_ValidationError_BadRequest_ReturnsValidationError(t *testing.T) {
	client := NewMockGraphClient()

	resource := "/me/drive/items/parent-id/children"
	errorResponse := `{"error":{"code":"invalidRequest","message":"The request is malformed or incorrect."}}`
	client.AddMockResponse(resource, []byte(errorResponse), http.StatusBadRequest,
		errors.NewValidationError("The request is malformed or incorrect", nil))

	invalidData := `{"invalid": "data"}`
	content := strings.NewReader(invalidData)
	_, err := client.Post(resource, content)

	assert.Error(t, err)
	assert.True(t, errors.IsValidationError(err))
	assert.Contains(t, err.Error(), "malformed")
}

// TestUT_GR_ERR_02_01_ServerError_InternalError_ReturnsOperationError tests 500 error handling
func TestUT_GR_ERR_02_01_ServerError_InternalError_ReturnsOperationError(t *testing.T) {
	client := NewMockGraphClient()

	resource := "/me/drive/items/test-id"
	errorResponse := `{"error":{"code":"internalServerError","message":"An internal server error occurred."}}`
	client.AddMockResponse(resource, []byte(errorResponse), http.StatusInternalServerError,
		errors.NewOperationError("An internal server error occurred", nil))

	_, err := client.Get(resource)

	assert.Error(t, err)
	assert.True(t, errors.IsOperationError(err))
	assert.Contains(t, err.Error(), "internal server error")
}

// TestUT_GR_ERR_02_02_NetworkError_ConnectionFailure_ReturnsNetworkError tests network error handling
func TestUT_GR_ERR_02_02_NetworkError_ConnectionFailure_ReturnsNetworkError(t *testing.T) {
	client := NewMockGraphClient()

	resource := "/me/drive/items/test-id"
	client.AddMockResponse(resource, nil, 0,
		errors.NewNetworkError("connection refused", nil))

	_, err := client.Get(resource)

	assert.Error(t, err)
	assert.True(t, errors.IsNetworkError(err))
	assert.Contains(t, err.Error(), "connection")
}

// TestUT_GR_ERR_03_01_ErrorParsing_MalformedErrorResponse_HandlesGracefully tests malformed error response handling
func TestUT_GR_ERR_03_01_ErrorParsing_MalformedErrorResponse_HandlesGracefully(t *testing.T) {
	client := NewMockGraphClient()

	resource := "/me/drive/items/test-id"
	malformedResponse := `{"invalid": "json structure"`
	client.AddMockResponse(resource, []byte(malformedResponse), http.StatusBadRequest,
		errors.NewValidationError("Failed to parse error response", nil))

	_, err := client.Get(resource)

	assert.Error(t, err)
	assert.NotNil(t, err)
}
