package fs

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/auriora/onemount/internal/graph"
)

// createMockAuth creates a mock authentication object for testing
func createMockAuth() *graph.Auth {
	return &graph.Auth{
		AccessToken:  "mock_access_token",
		RefreshToken: "mock_refresh_token",
		ExpiresAt:    time.Now().Add(1 * time.Hour).Unix(),
	}
}

// generateTestID generates a unique test ID
func generateTestID() string {
	return fmt.Sprintf("test-id-%d", time.Now().UnixNano())
}

// generateTestName generates a unique test name
func generateTestName() string {
	return fmt.Sprintf("test-file-%d.txt", time.Now().UnixNano())
}

// TestStatsCaching verifies that statistics are cached and reused
func TestStatsCaching(t *testing.T) {
	auth := createMockAuth()
	fs, err := NewFilesystem(auth, t.TempDir(), 14)
	if err != nil {
		t.Fatal(err)
	}
	defer fs.Stop()

	// Configure short cache TTL for testing
	fs.statsConfig = &StatsConfig{
		CacheTTL:                 2 * time.Second,
		SamplingThreshold:        10000,
		SamplingRate:             0.1,
		UseBackgroundCalculation: false,
	}

	// First call should calculate stats
	start1 := time.Now()
	stats1, err := fs.GetStats()
	duration1 := time.Since(start1)
	if err != nil {
		t.Fatalf("Failed to get stats: %v", err)
	}

	// Second call should use cache (much faster)
	start2 := time.Now()
	stats2, err := fs.GetStats()
	duration2 := time.Since(start2)
	if err != nil {
		t.Fatalf("Failed to get cached stats: %v", err)
	}

	// Verify stats are the same
	if stats1.MetadataCount != stats2.MetadataCount {
		t.Errorf("Cached stats differ: %d vs %d", stats1.MetadataCount, stats2.MetadataCount)
	}

	// Cached call should be significantly faster
	if duration2 > duration1/2 {
		t.Logf("Warning: Cached call not significantly faster (first: %v, cached: %v)", duration1, duration2)
	}

	t.Logf("First call: %v, Cached call: %v (%.2fx faster)", duration1, duration2, float64(duration1)/float64(duration2))

	// Wait for cache to expire
	time.Sleep(3 * time.Second)

	// Third call should recalculate
	start3 := time.Now()
	stats3, err := fs.GetStats()
	duration3 := time.Since(start3)
	if err != nil {
		t.Fatalf("Failed to get stats after cache expiry: %v", err)
	}

	// Should take longer than cached call
	if duration3 < duration2 {
		t.Logf("Note: Recalculation was faster than expected (cached: %v, recalc: %v)", duration2, duration3)
	}

	t.Logf("After cache expiry: %v", duration3)

	// Verify stats are still consistent
	if stats1.MetadataCount != stats3.MetadataCount {
		t.Errorf("Stats changed after recalculation: %d vs %d", stats1.MetadataCount, stats3.MetadataCount)
	}
}

// TestStatsSampling verifies that sampling works correctly for large datasets
func TestStatsSampling(t *testing.T) {
	auth := createMockAuth()
	fs, err := NewFilesystem(auth, t.TempDir(), 14)
	if err != nil {
		t.Fatal(err)
	}
	defer fs.Stop()

	// Add some test items to the metadata
	for i := 0; i < 100; i++ {
		item := &graph.DriveItem{
			ID:   generateTestID(),
			Name: generateTestName(),
			File: &graph.File{},
			Size: uint64(i * 1024),
		}
		inode := NewInodeDriveItem(item)
		fs.InsertID(item.ID, inode)
	}

	// Test with sampling enabled (low threshold)
	config := &StatsConfig{
		CacheTTL:                 0,   // Don't cache for this test
		SamplingThreshold:        50,  // Use sampling for >50 items
		SamplingRate:             0.5, // Sample 50%
		UseBackgroundCalculation: false,
	}

	stats, err := fs.GetStatsWithConfig(config)
	if err != nil {
		t.Fatalf("Failed to get stats with sampling: %v", err)
	}

	if !stats.IsSampled {
		t.Error("Expected stats to be sampled")
	}

	t.Logf("Sampled stats: MetadataCount=%d, IsSampled=%v", stats.MetadataCount, stats.IsSampled)

	// Test without sampling (high threshold)
	config2 := &StatsConfig{
		CacheTTL:                 0,
		SamplingThreshold:        1000, // Don't use sampling
		SamplingRate:             0.1,
		UseBackgroundCalculation: false,
	}

	stats2, err := fs.GetStatsWithConfig(config2)
	if err != nil {
		t.Fatalf("Failed to get stats without sampling: %v", err)
	}

	if stats2.IsSampled {
		t.Error("Expected stats to not be sampled")
	}

	t.Logf("Full stats: MetadataCount=%d, IsSampled=%v", stats2.MetadataCount, stats2.IsSampled)
}

// TestQuickStats verifies that quick stats are fast and contain basic information
func TestQuickStats(t *testing.T) {
	auth := createMockAuth()
	fs, err := NewFilesystem(auth, t.TempDir(), 14)
	if err != nil {
		t.Fatal(err)
	}
	defer fs.Stop()

	// Add some test items
	for i := 0; i < 50; i++ {
		item := &graph.DriveItem{
			ID:   generateTestID(),
			Name: generateTestName(),
			File: &graph.File{},
		}
		inode := NewInodeDriveItem(item)
		fs.InsertID(item.ID, inode)
	}

	// Get quick stats
	start := time.Now()
	stats, err := fs.GetQuickStats()
	duration := time.Since(start)
	if err != nil {
		t.Fatalf("Failed to get quick stats: %v", err)
	}

	// Quick stats should be very fast (< 100ms)
	if duration > 100*time.Millisecond {
		t.Errorf("Quick stats took too long: %v", duration)
	}

	// Verify basic stats are present
	if stats.MetadataCount == 0 {
		t.Error("Quick stats should have metadata count")
	}

	t.Logf("Quick stats completed in %v: MetadataCount=%d", duration, stats.MetadataCount)
}

// TestStatsPagination verifies that pagination works correctly
func TestStatsPagination(t *testing.T) {
	auth := createMockAuth()
	fs, err := NewFilesystem(auth, t.TempDir(), 14)
	if err != nil {
		t.Fatal(err)
	}
	defer fs.Stop()

	// Add test items with various extensions
	extensions := []string{".txt", ".pdf", ".jpg", ".png", ".doc", ".xls"}
	for i := 0; i < 60; i++ {
		ext := extensions[i%len(extensions)]
		item := &graph.DriveItem{
			ID:   generateTestID(),
			Name: "file" + ext,
			File: &graph.File{},
		}
		inode := NewInodeDriveItem(item)
		fs.InsertID(item.ID, inode)
	}

	// Force stats calculation
	_, err = fs.GetStats()
	if err != nil {
		t.Fatalf("Failed to get stats: %v", err)
	}

	// Test pagination
	page1, err := fs.GetStatsPage("extensions", 0, 3)
	if err != nil {
		t.Fatalf("Failed to get page 1: %v", err)
	}

	if len(page1) > 3 {
		t.Errorf("Page 1 has too many items: %d", len(page1))
	}

	t.Logf("Page 1 (size %d): %v", len(page1), page1)

	page2, err := fs.GetStatsPage("extensions", 1, 3)
	if err != nil {
		t.Fatalf("Failed to get page 2: %v", err)
	}

	t.Logf("Page 2 (size %d): %v", len(page2), page2)

	// Verify pages don't overlap
	for k := range page1 {
		if _, exists := page2[k]; exists {
			t.Errorf("Pages overlap on key: %s", k)
		}
	}
}

// TestBackgroundStatsUpdater verifies background statistics updates
func TestBackgroundStatsUpdater(t *testing.T) {
	auth := createMockAuth()
	fs, err := NewFilesystem(auth, t.TempDir(), 14)
	if err != nil {
		t.Fatal(err)
	}
	defer fs.Stop()

	// Configure for background updates
	fs.statsConfig = &StatsConfig{
		CacheTTL:                 1 * time.Second,
		SamplingThreshold:        10000,
		SamplingRate:             0.1,
		UseBackgroundCalculation: true,
	}

	// Start background updater with short interval
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	fs.StartBackgroundStatsUpdater(ctx, 2*time.Second)

	// Get initial stats
	stats1, err := fs.GetStats()
	if err != nil {
		t.Fatalf("Failed to get initial stats: %v", err)
	}

	t.Logf("Initial stats: MetadataCount=%d", stats1.MetadataCount)

	// Add more items
	for i := 0; i < 10; i++ {
		item := &graph.DriveItem{
			ID:   generateTestID(),
			Name: generateTestName(),
			File: &graph.File{},
		}
		inode := NewInodeDriveItem(item)
		fs.InsertID(item.ID, inode)
	}

	// Invalidate cache to force recalculation
	fs.InvalidateStatsCache()

	// Wait for background update
	time.Sleep(3 * time.Second)

	// Get updated stats
	stats2, err := fs.GetStats()
	if err != nil {
		t.Fatalf("Failed to get updated stats: %v", err)
	}

	t.Logf("Updated stats: MetadataCount=%d", stats2.MetadataCount)

	// Verify stats were updated
	if stats2.MetadataCount <= stats1.MetadataCount {
		t.Errorf("Stats were not updated: %d vs %d", stats1.MetadataCount, stats2.MetadataCount)
	}
}

// BenchmarkGetStats benchmarks the full statistics calculation
func BenchmarkGetStats(b *testing.B) {
	auth := createMockAuth()
	fs, err := NewFilesystem(auth, b.TempDir(), 14)
	if err != nil {
		b.Fatal(err)
	}
	defer fs.Stop()

	// Add test items
	for i := 0; i < 1000; i++ {
		item := &graph.DriveItem{
			ID:   generateTestID(),
			Name: generateTestName(),
			File: &graph.File{},
		}
		inode := NewInodeDriveItem(item)
		fs.InsertID(item.ID, inode)
	}

	// Disable caching for benchmark
	fs.statsConfig = &StatsConfig{
		CacheTTL:                 0,
		SamplingThreshold:        10000,
		SamplingRate:             0.1,
		UseBackgroundCalculation: false,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := fs.GetStats()
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkGetQuickStats benchmarks the quick statistics calculation
func BenchmarkGetQuickStats(b *testing.B) {
	auth := createMockAuth()
	fs, err := NewFilesystem(auth, b.TempDir(), 14)
	if err != nil {
		b.Fatal(err)
	}
	defer fs.Stop()

	// Add test items
	for i := 0; i < 1000; i++ {
		item := &graph.DriveItem{
			ID:   generateTestID(),
			Name: generateTestName(),
			File: &graph.File{},
		}
		inode := NewInodeDriveItem(item)
		fs.InsertID(item.ID, inode)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := fs.GetQuickStats()
		if err != nil {
			b.Fatal(err)
		}
	}
}
