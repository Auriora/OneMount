package util

import (
	"context"
	"testing"
	"time"
)

func TestBandwidthThrottler_Disabled(t *testing.T) {
	// Throttler with 0 limit should be disabled
	throttler := NewBandwidthThrottler(0)

	if throttler.IsEnabled() {
		t.Error("Throttler should be disabled when limit is 0")
	}

	// Should not block
	ctx := context.Background()
	start := time.Now()
	err := throttler.Wait(ctx, 1024*1024) // 1MB
	elapsed := time.Since(start)

	if err != nil {
		t.Errorf("Wait should not error when disabled: %v", err)
	}

	if elapsed > 10*time.Millisecond {
		t.Errorf("Wait should return immediately when disabled, took %v", elapsed)
	}
}

func TestBandwidthThrottler_BasicThrottling(t *testing.T) {
	// Set limit to 1MB/s
	throttler := NewBandwidthThrottler(1024 * 1024)

	if !throttler.IsEnabled() {
		t.Error("Throttler should be enabled when limit > 0")
	}

	ctx := context.Background()
	start := time.Now()

	// Transfer 2MB - should take at least 2 seconds
	err := throttler.Wait(ctx, 2*1024*1024)
	elapsed := time.Since(start)

	if err != nil {
		t.Errorf("Wait should not error: %v", err)
	}

	// Should take at least 1.5 seconds (allowing some tolerance)
	if elapsed < 1500*time.Millisecond {
		t.Errorf("Throttling should delay transfer, took only %v", elapsed)
	}

	// Should not take more than 3 seconds (allowing overhead)
	if elapsed > 3*time.Second {
		t.Errorf("Throttling took too long: %v", elapsed)
	}
}

func TestBandwidthThrottler_ContextCancellation(t *testing.T) {
	// Set limit to 100KB/s
	throttler := NewBandwidthThrottler(100 * 1024)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Try to transfer 1MB - should be cancelled
	err := throttler.Wait(ctx, 1024*1024)

	if err == nil {
		t.Error("Wait should return error when context is cancelled")
	}

	if err != context.DeadlineExceeded {
		t.Errorf("Expected context.DeadlineExceeded, got %v", err)
	}
}

func TestBandwidthThrottler_Reset(t *testing.T) {
	throttler := NewBandwidthThrottler(1024 * 1024) // 1MB/s

	ctx := context.Background()

	// Transfer some data
	throttler.Wait(ctx, 512*1024) // 512KB

	bandwidth1 := throttler.GetCurrentBandwidth()
	if bandwidth1 <= 0 {
		t.Error("Bandwidth should be > 0 after transfer")
	}

	// Reset
	throttler.Reset()

	bandwidth2 := throttler.GetCurrentBandwidth()
	if bandwidth2 != 0 {
		t.Errorf("Bandwidth should be 0 after reset, got %f", bandwidth2)
	}
}

func TestBandwidthThrottler_SetLimit(t *testing.T) {
	throttler := NewBandwidthThrottler(1024 * 1024) // 1MB/s

	if !throttler.IsEnabled() {
		t.Error("Throttler should be enabled initially")
	}

	// Disable by setting limit to 0
	throttler.SetLimit(0)

	if throttler.IsEnabled() {
		t.Error("Throttler should be disabled after setting limit to 0")
	}

	// Re-enable with new limit
	throttler.SetLimit(2 * 1024 * 1024) // 2MB/s

	if !throttler.IsEnabled() {
		t.Error("Throttler should be enabled after setting limit > 0")
	}
}

func TestBandwidthThrottler_ConcurrentAccess(t *testing.T) {
	throttler := NewBandwidthThrottler(1024 * 1024) // 1MB/s

	ctx := context.Background()
	done := make(chan bool, 10)

	// Simulate 10 concurrent transfers
	for i := 0; i < 10; i++ {
		go func() {
			err := throttler.Wait(ctx, 100*1024) // 100KB each
			if err != nil {
				t.Errorf("Concurrent wait failed: %v", err)
			}
			done <- true
		}()
	}

	// Wait for all to complete
	for i := 0; i < 10; i++ {
		select {
		case <-done:
			// Success
		case <-time.After(5 * time.Second):
			t.Fatal("Concurrent access test timed out")
		}
	}
}

func TestBandwidthThrottler_SmallTransfers(t *testing.T) {
	throttler := NewBandwidthThrottler(1024 * 1024) // 1MB/s

	ctx := context.Background()
	start := time.Now()

	// Many small transfers should not cause excessive delays
	for i := 0; i < 100; i++ {
		err := throttler.Wait(ctx, 1024) // 1KB each
		if err != nil {
			t.Errorf("Small transfer failed: %v", err)
		}
	}

	elapsed := time.Since(start)

	// 100KB should take less than 1 second at 1MB/s
	if elapsed > 2*time.Second {
		t.Errorf("Small transfers took too long: %v", elapsed)
	}
}
