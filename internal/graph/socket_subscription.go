package graph

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// SocketSubscription describes the response for the socket.io endpoint lookup.
type SocketSubscription struct {
	ID              string `json:"id"`
	NotificationURL string `json:"notificationUrl"`
}

// GetSocketSubscription retrieves the delegated Socket.IO endpoint for the specified resource.
func GetSocketSubscription(ctx context.Context, auth *Auth, resource string) (*SocketSubscription, error) {
	if auth == nil {
		return nil, fmt.Errorf("auth cannot be nil")
	}

	endpoint := buildSocketSubscriptionPath(resource)
	resp, err := RequestWithContext(ctx, endpoint, auth, http.MethodGet, nil, Header{
		key:   "Content-Type",
		value: "application/json",
	})
	if err != nil {
		return nil, err
	}

	var sub SocketSubscription
	if err := json.Unmarshal(resp, &sub); err != nil {
		return nil, fmt.Errorf("unmarshal socket subscription response: %w", err)
	}

	if sub.NotificationURL == "" {
		return nil, fmt.Errorf("socket subscription response missing notificationUrl")
	}
	return &sub, nil
}

func buildSocketSubscriptionPath(resource string) string {
	cleaned := TrimGraphURL(resource)
	if cleaned == "" {
		cleaned = "/me/drive/root"
	}
	if !strings.HasPrefix(cleaned, "/") {
		cleaned = "/" + cleaned
	}
	cleaned = strings.TrimSuffix(cleaned, "/")
	if cleaned == "" {
		cleaned = "/me/drive/root"
	}
	return fmt.Sprintf("%s/subscriptions/socketIo", cleaned)
}
