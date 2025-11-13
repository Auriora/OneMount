package fs

import (
	"testing"
	"time"
)

// TestCachedStatsExpiration verifies that cached stats expire correctly
func TestCachedStatsExpiration(t *testing.T) {
	cached := &CachedStats{}

	// Create test stats
	testStats := &Stats{
		MetadataCount: 100,
		CachedAt:      time.Now(),
	}

	// Set stats with short expiration
	cached.mu.Lock()
	cached.stats = testStats
	cached.expiresAt = time.Now().Add(100 * time.Millisecond)
	cached.mu.Unlock()

	// Verify stats are available
	cached.mu.RLock()
	if cached.stats == nil {
		t.Error("Stats should be available")
	}
	if time.Now().After(cached.expiresAt) {
		t.Error("Stats should not be expired yet")
	}
	cached.mu.RUnlock()

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Verify stats are expired
	cached.mu.RLock()
	if !time.Now().After(cached.expiresAt) {
		t.Error("Stats should be expired")
	}
	cached.mu.RUnlock()
}

// TestDefaultStatsConfig verifies default configuration values
func TestDefaultStatsConfig(t *testing.T) {
	config := DefaultStatsConfig()

	if config.CacheTTL != 5*time.Minute {
		t.Errorf("Expected CacheTTL to be 5 minutes, got %v", config.CacheTTL)
	}

	if config.SamplingThreshold != 10000 {
		t.Errorf("Expected SamplingThreshold to be 10000, got %d", config.SamplingThreshold)
	}

	if config.SamplingRate != 0.1 {
		t.Errorf("Expected SamplingRate to be 0.1, got %f", config.SamplingRate)
	}

	if !config.UseBackgroundCalculation {
		t.Error("Expected UseBackgroundCalculation to be true")
	}
}

// TestStatsIsSampled verifies the IsSampled flag
func TestStatsIsSampled(t *testing.T) {
	stats := &Stats{
		IsSampled: true,
		CachedAt:  time.Now(),
	}

	if !stats.IsSampled {
		t.Error("Expected stats to be marked as sampled")
	}

	stats2 := &Stats{
		IsSampled: false,
		CachedAt:  time.Now(),
	}

	if stats2.IsSampled {
		t.Error("Expected stats to not be marked as sampled")
	}
}

// TestFormatSize verifies size formatting
func TestFormatSize(t *testing.T) {
	tests := []struct {
		size     int64
		expected string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1024, "1.0 KiB"},
		{1536, "1.5 KiB"},
		{1048576, "1.0 MiB"},
		{1073741824, "1.0 GiB"},
		{1099511627776, "1.0 TiB"},
	}

	for _, tt := range tests {
		result := FormatSize(tt.size)
		if result != tt.expected {
			t.Errorf("FormatSize(%d) = %s, expected %s", tt.size, result, tt.expected)
		}
	}
}
