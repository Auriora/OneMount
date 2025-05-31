package graph

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

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

// TestUT_GR_ERR_04_01_NetworkConnectivityLoss_DuringOperation_ReturnsNetworkError tests network connectivity loss
func TestUT_GR_ERR_04_01_NetworkConnectivityLoss_DuringOperation_ReturnsNetworkError(t *testing.T) {
	client := NewMockGraphClient()

	resource := "/me/drive/items/test-id"
	client.AddMockResponse(resource, nil, 0,
		errors.NewNetworkError("network is unreachable", nil))

	_, err := client.Get(resource)

	assert.Error(t, err)
	assert.True(t, errors.IsNetworkError(err))
	assert.Contains(t, err.Error(), "network is unreachable")
}

// TestUT_GR_ERR_04_02_NetworkConnectivityLoss_ConnectionReset_ReturnsNetworkError tests connection reset scenarios
func TestUT_GR_ERR_04_02_NetworkConnectivityLoss_ConnectionReset_ReturnsNetworkError(t *testing.T) {
	client := NewMockGraphClient()

	resource := "/me/drive/items/test-id"
	client.AddMockResponse(resource, nil, 0,
		errors.NewNetworkError("connection reset by peer", nil))

	_, err := client.Get(resource)

	assert.Error(t, err)
	assert.True(t, errors.IsNetworkError(err))
	assert.Contains(t, err.Error(), "connection reset by peer")
}

// TestUT_GR_ERR_04_03_NetworkConnectivityLoss_NoRouteToHost_ReturnsNetworkError tests no route to host scenarios
func TestUT_GR_ERR_04_03_NetworkConnectivityLoss_NoRouteToHost_ReturnsNetworkError(t *testing.T) {
	client := NewMockGraphClient()

	resource := "/me/drive/items/test-id"
	client.AddMockResponse(resource, nil, 0,
		errors.NewNetworkError("no route to host", nil))

	_, err := client.Get(resource)

	assert.Error(t, err)
	assert.True(t, errors.IsNetworkError(err))
	assert.Contains(t, err.Error(), "no route to host")
}

// TestUT_GR_ERR_05_01_APITimeout_RequestTimeout_ReturnsTimeoutError tests API request timeout handling
func TestUT_GR_ERR_05_01_APITimeout_RequestTimeout_ReturnsTimeoutError(t *testing.T) {
	client := NewMockGraphClient()

	resource := "/me/drive/items/test-id"
	client.AddMockResponse(resource, nil, http.StatusRequestTimeout,
		errors.NewTimeoutError("request timeout", nil))

	_, err := client.Get(resource)

	assert.Error(t, err)
	assert.True(t, errors.IsTimeoutError(err))
	assert.Contains(t, err.Error(), "request timeout")
}

// TestUT_GR_ERR_05_02_APITimeout_ContextTimeout_ReturnsTimeoutError tests context timeout handling
func TestUT_GR_ERR_05_02_APITimeout_ContextTimeout_ReturnsTimeoutError(t *testing.T) {
	client := NewMockGraphClient()

	// Create a context with a very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	// Wait for the context to timeout
	time.Sleep(2 * time.Millisecond)

	resource := "/me/drive/items/test-id"
	_, err := client.GetWithContext(ctx, resource)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context")
}

// TestUT_GR_ERR_05_03_APITimeout_ReadTimeout_ReturnsTimeoutError tests read timeout scenarios
func TestUT_GR_ERR_05_03_APITimeout_ReadTimeout_ReturnsTimeoutError(t *testing.T) {
	client := NewMockGraphClient()

	resource := "/me/drive/items/test-id"
	client.AddMockResponse(resource, nil, 0,
		errors.NewTimeoutError("read timeout", nil))

	_, err := client.Get(resource)

	assert.Error(t, err)
	assert.True(t, errors.IsTimeoutError(err))
	assert.Contains(t, err.Error(), "read timeout")
}

// TestUT_GR_ERR_05_04_APITimeout_WriteTimeout_ReturnsTimeoutError tests write timeout scenarios
func TestUT_GR_ERR_05_04_APITimeout_WriteTimeout_ReturnsTimeoutError(t *testing.T) {
	client := NewMockGraphClient()

	resource := "/me/drive/items/parent-id/children"
	client.AddMockResponse(resource, nil, 0,
		errors.NewTimeoutError("write timeout", nil))

	content := strings.NewReader(`{"name": "test.txt"}`)
	_, err := client.Post(resource, content)

	assert.Error(t, err)
	assert.True(t, errors.IsTimeoutError(err))
	assert.Contains(t, err.Error(), "write timeout")
}

// TestUT_GR_ERR_06_01_DNSResolutionFailure_HostNotFound_ReturnsNetworkError tests DNS host not found scenarios
func TestUT_GR_ERR_06_01_DNSResolutionFailure_HostNotFound_ReturnsNetworkError(t *testing.T) {
	client := NewMockGraphClient()

	resource := "/me/drive/items/test-id"
	client.AddMockResponse(resource, nil, 0,
		errors.NewNetworkError("no such host", nil))

	_, err := client.Get(resource)

	assert.Error(t, err)
	assert.True(t, errors.IsNetworkError(err))
	assert.Contains(t, err.Error(), "no such host")
}

// TestUT_GR_ERR_06_02_DNSResolutionFailure_DNSTimeout_ReturnsNetworkError tests DNS timeout scenarios
func TestUT_GR_ERR_06_02_DNSResolutionFailure_DNSTimeout_ReturnsNetworkError(t *testing.T) {
	client := NewMockGraphClient()

	resource := "/me/drive/items/test-id"
	client.AddMockResponse(resource, nil, 0,
		errors.NewNetworkError("DNS resolution timeout", nil))

	_, err := client.Get(resource)

	assert.Error(t, err)
	assert.True(t, errors.IsNetworkError(err))
	assert.Contains(t, err.Error(), "DNS resolution timeout")
}

// TestUT_GR_ERR_06_03_DNSResolutionFailure_DNSServerUnavailable_ReturnsNetworkError tests DNS server unavailable scenarios
func TestUT_GR_ERR_06_03_DNSResolutionFailure_DNSServerUnavailable_ReturnsNetworkError(t *testing.T) {
	client := NewMockGraphClient()

	resource := "/me/drive/items/test-id"
	client.AddMockResponse(resource, nil, 0,
		errors.NewNetworkError("DNS server unavailable", nil))

	_, err := client.Get(resource)

	assert.Error(t, err)
	assert.True(t, errors.IsNetworkError(err))
	assert.Contains(t, err.Error(), "DNS server unavailable")
}

// TestUT_GR_ERR_06_04_DNSResolutionFailure_TemporaryFailure_ReturnsNetworkError tests temporary DNS failure scenarios
func TestUT_GR_ERR_06_04_DNSResolutionFailure_TemporaryFailure_ReturnsNetworkError(t *testing.T) {
	client := NewMockGraphClient()

	resource := "/me/drive/items/test-id"
	client.AddMockResponse(resource, nil, 0,
		errors.NewNetworkError("temporary DNS failure", nil))

	_, err := client.Get(resource)

	assert.Error(t, err)
	assert.True(t, errors.IsNetworkError(err))
	assert.Contains(t, err.Error(), "temporary DNS failure")
}

// TestUT_GR_ERR_07_01_SSLTLSCertificateError_CertificateExpired_ReturnsNetworkError tests expired certificate scenarios
func TestUT_GR_ERR_07_01_SSLTLSCertificateError_CertificateExpired_ReturnsNetworkError(t *testing.T) {
	client := NewMockGraphClient()

	resource := "/me/drive/items/test-id"
	client.AddMockResponse(resource, nil, 0,
		errors.NewNetworkError("certificate has expired", nil))

	_, err := client.Get(resource)

	assert.Error(t, err)
	assert.True(t, errors.IsNetworkError(err))
	assert.Contains(t, err.Error(), "certificate has expired")
}

// TestUT_GR_ERR_07_02_SSLTLSCertificateError_CertificateUntrusted_ReturnsNetworkError tests untrusted certificate scenarios
func TestUT_GR_ERR_07_02_SSLTLSCertificateError_CertificateUntrusted_ReturnsNetworkError(t *testing.T) {
	client := NewMockGraphClient()

	resource := "/me/drive/items/test-id"
	client.AddMockResponse(resource, nil, 0,
		errors.NewNetworkError("certificate signed by unknown authority", nil))

	_, err := client.Get(resource)

	assert.Error(t, err)
	assert.True(t, errors.IsNetworkError(err))
	assert.Contains(t, err.Error(), "certificate signed by unknown authority")
}

// TestUT_GR_ERR_07_03_SSLTLSCertificateError_HostnameMismatch_ReturnsNetworkError tests hostname mismatch scenarios
func TestUT_GR_ERR_07_03_SSLTLSCertificateError_HostnameMismatch_ReturnsNetworkError(t *testing.T) {
	client := NewMockGraphClient()

	resource := "/me/drive/items/test-id"
	client.AddMockResponse(resource, nil, 0,
		errors.NewNetworkError("certificate is not valid for hostname", nil))

	_, err := client.Get(resource)

	assert.Error(t, err)
	assert.True(t, errors.IsNetworkError(err))
	assert.Contains(t, err.Error(), "certificate is not valid for hostname")
}

// TestUT_GR_ERR_07_04_SSLTLSCertificateError_SSLHandshakeFailure_ReturnsNetworkError tests SSL handshake failure scenarios
func TestUT_GR_ERR_07_04_SSLTLSCertificateError_SSLHandshakeFailure_ReturnsNetworkError(t *testing.T) {
	client := NewMockGraphClient()

	resource := "/me/drive/items/test-id"
	client.AddMockResponse(resource, nil, 0,
		errors.NewNetworkError("SSL handshake failed", nil))

	_, err := client.Get(resource)

	assert.Error(t, err)
	assert.True(t, errors.IsNetworkError(err))
	assert.Contains(t, err.Error(), "SSL handshake failed")
}

// TestUT_GR_ERR_07_05_SSLTLSCertificateError_TLSVersionMismatch_ReturnsNetworkError tests TLS version mismatch scenarios
func TestUT_GR_ERR_07_05_SSLTLSCertificateError_TLSVersionMismatch_ReturnsNetworkError(t *testing.T) {
	client := NewMockGraphClient()

	resource := "/me/drive/items/test-id"
	client.AddMockResponse(resource, nil, 0,
		errors.NewNetworkError("TLS version not supported", nil))

	_, err := client.Get(resource)

	assert.Error(t, err)
	assert.True(t, errors.IsNetworkError(err))
	assert.Contains(t, err.Error(), "TLS version not supported")
}
