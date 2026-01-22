package graph

import "testing"

func TestUT_Graph_Socket_BuildSubscriptionPath(t *testing.T) {
	tests := []struct {
		name     string
		resource string
		want     string
	}{
		{"default", "", "/me/drive/root/subscriptions/socketIo"},
		{"root", "/me/drive/root", "/me/drive/root/subscriptions/socketIo"},
		{"trailing slash", "/drives/123/root/", "/drives/123/root/subscriptions/socketIo"},
		{"with graph prefix", "https://graph.microsoft.com/v1.0/sites/foo/lists/bar/drive/root", "/sites/foo/lists/bar/drive/root/subscriptions/socketIo"},
		{"no leading slash", "drives/abc/root", "/drives/abc/root/subscriptions/socketIo"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := buildSocketSubscriptionPath(tc.resource)
			if got != tc.want {
				t.Fatalf("expected %s, got %s", tc.want, got)
			}
		})
	}
}
