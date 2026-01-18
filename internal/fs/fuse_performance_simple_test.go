package fs

import (
	"testing"
	"time"

	"github.com/auriora/onemount/internal/graph"
	"github.com/auriora/onemount/internal/testutil/framework"
	"github.com/auriora/onemount/internal/testutil/helpers"
)

// TestUT_FS_FUSEPerformance_Simple_MetadataOperations tests that FUSE operations
// are served from local metadata/cache only without blocking on Graph API calls.
//
//	Test Case ID    UT-FS-FUSEPerformance-Simple
//	Title           FUSE Operations Performance Test
//	Description     Tests that FUSE operations are fast and served from local cache
//	Preconditions   Filesystem with cached metadata
//	Steps           1. Test basic FUSE operations
//	                2. Measure response times
//	                3. Test concurrent operations
//	                4. Verify performance requirements
//	Expected Result FUSE operations meet performance requirements
//	Requirements    2D.1
func TestUT_FS_FUSEPerformance_Simple_MetadataOperations(t *testing.T) {
	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "FUSEPerformanceSimpleFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
		// Create the filesystem
		fs, err := NewFilesystem(auth, mountPoint, cacheTTL)
		return fs, err
	})

	// Use the fixture to run the test
	fixture.Use(t, func(t *testing.T, fixture interface{}) {
		// Create assertions helper
		assert := framework.NewAssert(t)

		// Get the test data
		unitTestFixture, ok := fixture.(*framework.UnitTestFixture)
		if !ok {
			t.Fatalf("Expected fixture to be of type *framework.UnitTestFixture, but got %T", fixture)
		}
		fsFixture := unitTestFixture.SetupData.(*helpers.FSTestFixture)
		filesystem := fsFixture.FS.(*Filesystem)
		auth := fsFixture.Auth

		t.Log("=== FUSE Performance Test ===")

		// Step 1: Test basic FUSE operations performance
		t.Log("Step 1: Testing basic FUSE operations...")

		// Test getattr performance
		start := time.Now()
		rootInode := filesystem.GetID(filesystem.root)
		getattrTime := time.Since(start)

		assert.NotNil(rootInode, "Should be able to get root inode")
		assert.True(getattrTime < 10*time.Millisecond, "Getattr should be very fast from cache")

		t.Logf("✓ Getattr completed in %v", getattrTime)

		// Test readdir operation performance
		start = time.Now()
		children, err := filesystem.getChildrenID(filesystem.root, auth, false)
		readdirTime := time.Since(start)

		assert.NoError(err, "Should be able to read directory")
		assert.True(readdirTime < 100*time.Millisecond, "Readdir should be fast from cache")

		t.Logf("✓ Readdir completed in %v with %d children", readdirTime, len(children))

		// Step 2: Test multiple rapid operations
		t.Log("Step 2: Testing multiple rapid operations...")

		start = time.Now()
		operationCount := 100

		for i := 0; i < operationCount; i++ {
			// Alternate between different operations
			switch i % 3 {
			case 0:
				_, _ = filesystem.getChildrenID(filesystem.root, auth, false)
			case 1:
				_ = filesystem.GetID(filesystem.root)
			case 2:
				// Test basic path resolution
				_, _ = filesystem.GetPath("/", auth)
			}
		}

		totalTime := time.Since(start)
		avgTime := totalTime / time.Duration(operationCount)

		assert.True(totalTime < 1*time.Second, "100 operations should complete quickly")
		assert.True(avgTime < 10*time.Millisecond, "Average operation time should be very fast")

		t.Logf("✓ %d operations completed in %v (avg: %v per operation)",
			operationCount, totalTime, avgTime)

		// Step 3: Test concurrent operations
		t.Log("Step 3: Testing concurrent operations...")

		start = time.Now()
		concurrentOps := 10
		done := make(chan bool, concurrentOps)

		for i := 0; i < concurrentOps; i++ {
			go func(id int) {
				defer func() { done <- true }()

				// Perform multiple operations concurrently
				for j := 0; j < 10; j++ {
					switch j % 2 {
					case 0:
						_, _ = filesystem.getChildrenID(filesystem.root, auth, false)
					case 1:
						_ = filesystem.GetID(filesystem.root)
					}
				}
			}(i)
		}

		// Wait for all concurrent operations to complete
		for i := 0; i < concurrentOps; i++ {
			<-done
		}

		concurrentTime := time.Since(start)
		assert.True(concurrentTime < 2*time.Second, "Concurrent operations should complete quickly")

		t.Logf("✓ %d concurrent operations completed in %v", concurrentOps, concurrentTime)

		// Step 4: Verify background workers exist
		t.Log("Step 4: Verifying background workers...")

		// Check that filesystem has background components
		assert.NotNil(filesystem.metadataRequestManager, "Should have metadata request manager")
		assert.NotNil(filesystem.downloads, "Should have download manager")
		assert.NotNil(filesystem.uploads, "Should have upload manager")

		t.Log("✓ Background workers verified")

		// Step 5: Test performance consistency
		t.Log("Step 5: Testing performance consistency...")

		var times []time.Duration
		for i := 0; i < 20; i++ {
			start = time.Now()
			_ = filesystem.GetID(filesystem.root)
			times = append(times, time.Since(start))
		}

		// Calculate statistics
		var total time.Duration
		var max time.Duration
		for _, t := range times {
			total += t
			if t > max {
				max = t
			}
		}
		avg := total / time.Duration(len(times))

		assert.True(avg < 5*time.Millisecond, "Average response time should be consistent")
		assert.True(max < 20*time.Millisecond, "Maximum response time should be reasonable")

		t.Logf("✓ Consistency test - avg: %v, max: %v", avg, max)

		t.Log("✓ FUSE performance requirements verified")
	})
}
