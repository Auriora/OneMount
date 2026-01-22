package main

import (
	"testing"

	"github.com/auriora/onemount/cmd/common"
)

func TestUT_CMD_Main_ToRealtimeOptionsCopiesPollingOnly(t *testing.T) {
	cfg := common.RealtimeConfig{
		Enabled:     true,
		PollingOnly: true,
		Resource:    "/me/drive/root",
	}
	opts := toRealtimeOptions(cfg)
	if !opts.PollingOnly {
		t.Fatalf("expected pollingOnly to propagate to fs options")
	}
}
