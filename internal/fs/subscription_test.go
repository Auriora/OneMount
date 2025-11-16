package fs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/auriora/onemount/internal/graph"
	"github.com/auriora/onemount/internal/graph/api"
)

func TestSubscriptionManagerHandlesValidationAndNotifications(t *testing.T) {
	mockClient := graph.NewMockGraphClient()
	defer mockClient.Cleanup()

	expiration := time.Now().Add(2 * time.Hour).UTC()
	createResp := graph.Subscription{
		ID:                 "sub-abc",
		Resource:           "/me/drive/root",
		ChangeType:         "updated",
		NotificationURL:    "https://example.com/onemount/webhook",
		ExpirationDateTime: expiration,
		ClientState:        "secret",
	}
	body, _ := json.Marshal(createResp)
	mockClient.AddMockResponse("/subscriptions", body, http.StatusCreated, nil)
	mockClient.AddMockResponse("/subscriptions/sub-abc", []byte("{}"), http.StatusNoContent, nil)

	opts := WebhookOptions{
		Enabled:          true,
		PublicURL:        "https://example.com",
		ListenAddress:    "127.0.0.1:0",
		Path:             "/onemount/webhook",
		ClientState:      "secret",
		Resource:         "/me/drive/root",
		ChangeType:       "updated",
		FallbackInterval: time.Minute,
	}

	manager := NewSubscriptionManager(opts, &mockClient.Auth)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := manager.Start(ctx); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	defer manager.Stop(context.Background())

	addr := manager.ListenAddress()
	for i := 0; i < 10 && addr == ""; i++ {
		time.Sleep(10 * time.Millisecond)
		addr = manager.ListenAddress()
	}
	if addr == "" {
		t.Fatalf("manager did not report listen address")
	}

	validationURL := fmt.Sprintf("http://%s%s?validationToken=abc", addr, opts.Path)
	resp, err := http.Get(validationURL)
	if err != nil {
		t.Fatalf("validation GET failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200 for validation, got %d", resp.StatusCode)
	}

	payload := map[string]any{
		"value": []map[string]any{
			{
				"subscriptionId":                 "sub-abc",
				"clientState":                    "secret",
				"subscriptionExpirationDateTime": time.Now().Add(time.Hour).Format(time.RFC3339),
			},
		},
	}
	buf := bytes.NewBuffer(nil)
	if err := json.NewEncoder(buf).Encode(payload); err != nil {
		t.Fatalf("encode payload: %v", err)
	}
	notifyURL := fmt.Sprintf("http://%s%s", addr, opts.Path)
	postResp, err := http.Post(notifyURL, "application/json", buf)
	if err != nil {
		t.Fatalf("notification POST failed: %v", err)
	}
	postResp.Body.Close()
	if postResp.StatusCode != http.StatusAccepted {
		t.Fatalf("expected status 202 for notification, got %d", postResp.StatusCode)
	}

	select {
	case <-manager.Notifications():
	case <-time.After(2 * time.Second):
		t.Fatal("expected notification trigger")
	}
}

func TestSubscriptionManagerRecoversFromRenewalFailure(t *testing.T) {
	mockClient := graph.NewMockGraphClient()
	defer mockClient.Cleanup()

	expiration := time.Now().Add(2 * time.Hour).UTC()
	createResp := graph.Subscription{
		ID:                 "sub-abc",
		Resource:           "/me/drive/root",
		ChangeType:         "updated",
		NotificationURL:    "https://example.com/onemount/webhook",
		ExpirationDateTime: expiration,
		ClientState:        "secret",
	}
	body, _ := json.Marshal(createResp)
	mockClient.AddMockResponse("/subscriptions", body, http.StatusCreated, nil)
	mockClient.AddMockResponse("/subscriptions/sub-abc", []byte("not json"), http.StatusOK, nil)

	opts := WebhookOptions{
		Enabled:          true,
		PublicURL:        "https://example.com",
		ListenAddress:    "127.0.0.1:0",
		Path:             "/onemount/webhook",
		ClientState:      "secret",
		Resource:         "/me/drive/root",
		ChangeType:       "updated",
		FallbackInterval: time.Minute,
	}

	oldInterval := subscriptionRenewalCheckInterval
	subscriptionRenewalCheckInterval = 10 * time.Millisecond
	t.Cleanup(func() {
		subscriptionRenewalCheckInterval = oldInterval
	})

	manager := NewSubscriptionManager(opts, &mockClient.Auth)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := manager.Start(ctx); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	defer manager.Stop(context.Background())

	manager.mu.Lock()
	manager.expiration = time.Now().Add(5 * time.Minute)
	manager.mu.Unlock()

	if !waitForCondition(t, 5*time.Second, func() bool {
		return countSubscriptionCreates(mockClient.Recorder.GetCalls()) >= 2
	}) {
		t.Fatalf("expected subscription recreation to issue a second POST, got %d calls", countSubscriptionCreates(mockClient.Recorder.GetCalls()))
	}
}

func waitForCondition(t *testing.T, timeout time.Duration, cond func() bool) bool {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if cond() {
			return true
		}
		time.Sleep(10 * time.Millisecond)
	}
	return false
}

func countSubscriptionCreates(calls []api.MockCall) int {
	count := 0
	for _, call := range calls {
		switch call.Method {
		case "RequestWithContext":
			if len(call.Args) < 2 {
				continue
			}
			resource, _ := call.Args[0].(string)
			method, _ := call.Args[1].(string)
			if resource == "/subscriptions" && method == http.MethodPost {
				count++
			}
		case "RoundTrip":
			if len(call.Args) < 1 {
				continue
			}
			req, _ := call.Args[0].(*http.Request)
			if req == nil {
				continue
			}
			if req.Method == http.MethodPost && strings.HasSuffix(req.URL.Path, "/subscriptions") {
				count++
			}
		}
	}
	return count
}
