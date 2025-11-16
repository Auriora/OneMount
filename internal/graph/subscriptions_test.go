package graph

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

func TestSubscriptionsLifecycle(t *testing.T) {
	mockClient := NewMockGraphClient()
	defer mockClient.Cleanup()

	ctx := context.Background()

	createResp := Subscription{
		ID:                 "sub-123",
		Resource:           "/me/drive/root",
		ChangeType:         "updated",
		NotificationURL:    "https://example.com/webhook",
		ExpirationDateTime: time.Now().Add(2 * time.Hour).UTC(),
		ClientState:        "secret",
	}
	body, _ := json.Marshal(createResp)
	mockClient.AddMockResponse("/subscriptions", body, http.StatusCreated, nil)

	req := SubscriptionRequest{
		ChangeType:         "updated",
		NotificationURL:    "https://example.com/webhook",
		Resource:           "/me/drive/root",
		ExpirationDateTime: time.Now().Add(2 * time.Hour).UTC(),
		ClientState:        "secret",
	}

	sub, err := CreateSubscription(ctx, &mockClient.Auth, req)
	if err != nil {
		t.Fatalf("CreateSubscription failed: %v", err)
	}
	if sub.ID != createResp.ID {
		t.Fatalf("expected subscription ID %s, got %s", createResp.ID, sub.ID)
	}

	renewResp := *sub
	renewResp.ExpirationDateTime = time.Now().Add(3 * time.Hour).UTC()
	renewBody, _ := json.Marshal(renewResp)
	mockClient.AddMockResponse("/subscriptions/"+sub.ID, renewBody, http.StatusOK, nil)

	renewed, err := RenewSubscription(ctx, &mockClient.Auth, sub.ID, time.Now().Add(72*time.Hour))
	if err != nil {
		t.Fatalf("RenewSubscription failed: %v", err)
	}
	if renewed.ExpirationDateTime != renewResp.ExpirationDateTime {
		t.Fatalf("expected renewed expiration %v, got %v", renewResp.ExpirationDateTime, renewed.ExpirationDateTime)
	}

	mockClient.AddMockResponse("/subscriptions/"+sub.ID, []byte("{}"), http.StatusNoContent, nil)
	if err := DeleteSubscription(ctx, &mockClient.Auth, sub.ID); err != nil {
		t.Fatalf("DeleteSubscription failed: %v", err)
	}
}
