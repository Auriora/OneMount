package graph

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/auriora/onemount/internal/logging"
)

// Subscription represents a Microsoft Graph webhook subscription.
type Subscription struct {
	ID                 string    `json:"id"`
	Resource           string    `json:"resource"`
	ChangeType         string    `json:"changeType"`
	NotificationURL    string    `json:"notificationUrl"`
	ExpirationDateTime time.Time `json:"expirationDateTime"`
	ClientState        string    `json:"clientState"`
}

// SubscriptionRequest describes the payload for creating a subscription.
type SubscriptionRequest struct {
	ChangeType         string    `json:"changeType"`
	NotificationURL    string    `json:"notificationUrl"`
	Resource           string    `json:"resource"`
	ExpirationDateTime time.Time `json:"expirationDateTime"`
	ClientState        string    `json:"clientState,omitempty"`
}

// SubscriptionUpdateRequest describes the payload for renewing a subscription.
type SubscriptionUpdateRequest struct {
	ExpirationDateTime time.Time `json:"expirationDateTime"`
}

// CreateSubscription creates a webhook subscription using Microsoft Graph.
func CreateSubscription(ctx context.Context, auth *Auth, req SubscriptionRequest) (*Subscription, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal subscription request: %w", err)
	}

	resp, err := RequestWithContext(ctx, "/subscriptions", auth, http.MethodPost, bytes.NewReader(body), Header{
		key:   "Content-Type",
		value: "application/json",
	})
	if err != nil {
		return nil, err
	}

	var sub Subscription
	if err := json.Unmarshal(resp, &sub); err != nil {
		return nil, fmt.Errorf("unmarshal subscription response: %w", err)
	}
	return &sub, nil
}

// RenewSubscription renews an existing subscription before it expires.
func RenewSubscription(ctx context.Context, auth *Auth, subscriptionID string, expiration time.Time) (*Subscription, error) {
	payload := SubscriptionUpdateRequest{
		ExpirationDateTime: expiration,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal subscription renewal request: %w", err)
	}

	resource := fmt.Sprintf("/subscriptions/%s", subscriptionID)
	resp, err := RequestWithContext(ctx, resource, auth, http.MethodPatch, bytes.NewReader(body), Header{
		key:   "Content-Type",
		value: "application/json",
	})
	if err != nil {
		return nil, err
	}

	var sub Subscription
	if err := json.Unmarshal(resp, &sub); err != nil {
		return nil, fmt.Errorf("unmarshal subscription renewal response: %w", err)
	}
	return &sub, nil
}

// DeleteSubscription removes a subscription from Microsoft Graph.
func DeleteSubscription(ctx context.Context, auth *Auth, subscriptionID string) error {
	resource := fmt.Sprintf("/subscriptions/%s", subscriptionID)
	_, err := RequestWithContext(ctx, resource, auth, http.MethodDelete, nil)
	if err != nil {
		return err
	}
	return nil
}

// TrimGraphURL removes the GraphURL prefix from a subscription notification URL, if present.
func TrimGraphURL(link string) string {
	return strings.TrimPrefix(link, GraphURL)
}

// BuildExpiration clamps the requested expiration to the Graph API limit (3 days).
func BuildExpiration(preferred time.Duration) time.Time {
	maxDuration := 72 * time.Hour
	if preferred <= 0 || preferred > maxDuration {
		preferred = maxDuration
	}
	return time.Now().Add(preferred).UTC()
}

// LogSubscription writes subscription details to the log for diagnostics.
func LogSubscription(prefix string, sub *Subscription) {
	if sub == nil {
		return
	}
	logging.Info().
		Str("subscriptionID", sub.ID).
		Str("resource", sub.Resource).
		Str("changeType", sub.ChangeType).
		Str("notificationUrl", sub.NotificationURL).
		Time("expiration", sub.ExpirationDateTime).
		Msg(prefix)
}
