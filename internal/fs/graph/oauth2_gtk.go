//go:build linux && cgo

package graph

/*
#cgo linux pkg-config: webkit2gtk-4.1
#include "stdlib.h"
#include "oauth2_gtk.h"
*/
import "C"

import (
	"fmt"
	"unsafe"
)

// Fetch the auth code required as the first part of oauth2 authentication. Uses
// webkit2gtk to create a popup browser.
func getAuthCode(a AuthConfig, accountName string) (string, error) {
	cAuthURL := C.CString(getAuthURL(a))
	cAccountName := C.CString(accountName)
	cResponse := C.webkit_auth_window(cAuthURL, cAccountName)
	response := C.GoString(cResponse)
	C.free(unsafe.Pointer(cAuthURL))
	C.free(unsafe.Pointer(cAccountName))
	C.free(unsafe.Pointer(cResponse))

	code, err := parseAuthCode(response)
	if err != nil {
		//TODO create a popup with the auth failure message here instead of a log message
		return "", fmt.Errorf("no validation code returned, or code was invalid: %w", err)
	}
	return code, nil
}

// uriGetHost is exclusively here for testing because we cannot use CGo in tests,
// but can use functions that invoke CGo in tests.
func uriGetHost(uri string) string {
	input := C.CString(uri)
	defer C.free(unsafe.Pointer(input))

	host := C.uri_get_host(input)
	defer C.free(unsafe.Pointer(host))
	if host == nil {
		return ""
	}
	return C.GoString(host)
}
