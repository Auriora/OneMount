package graph

import "strings"

// TrimGraphURL removes the Graph base URL prefix and normalizes leading/trailing slashes for resource paths.
func TrimGraphURL(resource string) string {
	trimmed := strings.TrimSpace(resource)
	trimmed = strings.TrimPrefix(trimmed, GraphURL)
	trimmed = strings.TrimPrefix(trimmed, GraphURL+"/")
	trimmed = strings.TrimSpace(trimmed)
	if trimmed == "" {
		return "/me/drive/root"
	}
	if !strings.HasPrefix(trimmed, "/") {
		trimmed = "/" + trimmed
	}
	trimmed = strings.TrimSuffix(trimmed, "/")
	if trimmed == "" {
		return "/me/drive/root"
	}
	return trimmed
}
