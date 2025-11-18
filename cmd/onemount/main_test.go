package main

import (
	"testing"

	"github.com/auriora/onemount/cmd/common"
)

func TestToWebhookOptionsCopiesPollingOnly(t *testing.T) {
	cfg := common.WebhookConfig{
		Enabled:     true,
		UseSocketIO: true,
		PollingOnly: true,
	}
	opts := toWebhookOptions(cfg)
	if !opts.PollingOnly {
		t.Fatalf("expected pollingOnly to propagate to fs options")
	}
}
