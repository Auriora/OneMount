// Package testutil provides testing utilities for the OneMount project.
package framework

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"
)

// LargeFileTestConfig defines configuration for large file tests
type LargeFileTestConfig struct {
	// File size in bytes (default: 1GB + 1 byte)
	FileSize int64
	// Number of concurrent operations
	Concurrency int
	// Test timeout
	Timeout time.Duration
	// Chunk size for operations
	ChunkSize int64
}

// DefaultLargeFileTestConfig returns default configuration for large file tests
func DefaultLargeFileTestConfig() LargeFileTestConfig {
	return LargeFileTestConfig{
		FileSize:    1*1024*1024*1024 + 1, // 1GB + 1 byte
		Concurrency: 2,
		Timeout:     10 * time.Minute,
		ChunkSize:   64 * 1024 * 1024, // 64MB chunks
	}
}

// HighFileCountTestConfig defines configuration for high file count tests
type HighFileCountTestConfig struct {
	// Number of files to create (default: 10,001)
	FileCount int
	// Number of concurrent operations
	Concurrency int
	// Test timeout
	Timeout time.Duration
	// File size for each file
	FileSize int64
}

// DefaultHighFileCountTestConfig returns default configuration for high file count tests
func DefaultHighFileCountTestConfig() HighFileCountTestConfig {
	return HighFileCountTestConfig{
		FileCount:   10001, // >10k files
		Concurrency: 10,
		Timeout:     15 * time.Minute,
		FileSize:    1024, // 1KB per file
	}
}

// SustainedOperationTestConfig defines configuration for sustained operation tests
type SustainedOperationTestConfig struct {
	// Duration to run the test
	Duration time.Duration
	// Operations per second target
	OperationsPerSecond int
	// Number of concurrent workers
	Workers int
	// Memory check interval
	MemoryCheckInterval time.Duration
}

// DefaultSustainedOperationTestConfig returns default configuration for sustained operation tests
func DefaultSustainedOperationTestConfig() SustainedOperationTestConfig {
	return SustainedOperationTestConfig{
		Duration:            30 * time.Minute,
		OperationsPerSecond: 10,
		Workers:             5,
		MemoryCheckInterval: 1 * time.Minute,
	}
}

// MemoryLeakTestConfig defines configuration for memory leak detection tests
type MemoryLeakTestConfig struct {
	// Duration to run the test
	Duration time.Duration
	// Memory sampling interval
	SamplingInterval time.Duration
	// Maximum allowed memory growth (in MB)
	MaxMemoryGrowthMB int64
	// Number of operations per cycle
	OperationsPerCycle int
	// Cycle interval
	CycleInterval time.Duration
}

// DefaultMemoryLeakTestConfig returns default configuration for memory leak tests
func DefaultMemoryLeakTestConfig() MemoryLeakTestConfig {
	return MemoryLeakTestConfig{
		Duration:           20 * time.Minute,
		SamplingInterval:   30 * time.Second,
		MaxMemoryGrowthMB:  100, // 100MB max growth
		OperationsPerCycle: 100,
		CycleInterval:      1 * time.Minute,
	}
}

// MemorySample represents a memory usage sample
type MemorySample struct {
	Timestamp time.Time
	AllocMB   int64
	SysMB     int64
	HeapMB    int64
	StackMB   int64
}

// MemoryTracker tracks memory usage over time
type MemoryTracker struct {
	samples []MemorySample
	mu      sync.RWMutex
}

// NewMemoryTracker creates a new memory tracker
func NewMemoryTracker() *MemoryTracker {
	return &MemoryTracker{
		samples: make([]MemorySample, 0),
	}
}

// Sample takes a memory usage sample
func (mt *MemoryTracker) Sample() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	sample := MemorySample{
		Timestamp: time.Now(),
		AllocMB:   int64(m.Alloc) / (1024 * 1024),
		SysMB:     int64(m.Sys) / (1024 * 1024),
		HeapMB:    int64(m.HeapAlloc) / (1024 * 1024),
		StackMB:   int64(m.StackInuse) / (1024 * 1024),
	}

	mt.mu.Lock()
	mt.samples = append(mt.samples, sample)
	mt.mu.Unlock()
}

// GetSamples returns all memory samples
func (mt *MemoryTracker) GetSamples() []MemorySample {
	mt.mu.RLock()
	defer mt.mu.RUnlock()

	result := make([]MemorySample, len(mt.samples))
	copy(result, mt.samples)
	return result
}

// GetMemoryGrowth returns the memory growth from first to last sample
func (mt *MemoryTracker) GetMemoryGrowth() (int64, error) {
	mt.mu.RLock()
	defer mt.mu.RUnlock()

	if len(mt.samples) < 2 {
		return 0, fmt.Errorf("insufficient samples for growth calculation")
	}

	first := mt.samples[0]
	last := mt.samples[len(mt.samples)-1]

	return last.AllocMB - first.AllocMB, nil
}

// LargeFileHandlingTest tests handling of large files (>1GB)
func LargeFileHandlingTest(b *testing.B, framework *TestFramework, config LargeFileTestConfig, thresholds PerformanceThresholds) {
	// Create a benchmark for large file handling
	benchmark := NewPerformanceBenchmark(
		"LargeFileHandlingTest",
		fmt.Sprintf("Test for handling files of size %d bytes", config.FileSize),
		thresholds,
		framework.Config.ArtifactsDir,
	)

	// Set up the benchmark
	benchmark.SetSetupFunc(func() error {
		// Ensure we have enough disk space (rough check)
		// In a real implementation, you would check available disk space
		return nil
	})

	// Set the benchmark function
	benchmark.SetBenchmarkFunc(func(b *testing.B) {
		ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
		defer cancel()

		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			// Test large file upload
			start := time.Now()
			err := simulateLargeFileUpload(ctx, config.FileSize, config.ChunkSize)
			uploadLatency := time.Since(start)
			benchmark.RecordLatency(uploadLatency)

			if err != nil {
				b.Fatalf("Large file upload failed: %v", err)
			}

			// Test large file download
			start = time.Now()
			err = simulateLargeFileDownload(ctx, config.FileSize, config.ChunkSize)
			downloadLatency := time.Since(start)
			benchmark.RecordLatency(downloadLatency)

			if err != nil {
				b.Fatalf("Large file download failed: %v", err)
			}

			b.Logf("Large file operation completed: upload=%v, download=%v", uploadLatency, downloadLatency)
		}
	})

	// Run the benchmark
	benchmark.Run(b)
}

// HighFileCountDirectoryTest tests directories with many files (>10k files)
func HighFileCountDirectoryTest(b *testing.B, framework *TestFramework, config HighFileCountTestConfig, thresholds PerformanceThresholds) {
	// Create a benchmark for high file count directories
	benchmark := NewPerformanceBenchmark(
		"HighFileCountDirectoryTest",
		fmt.Sprintf("Test for directories with %d files", config.FileCount),
		thresholds,
		framework.Config.ArtifactsDir,
	)

	// Set up the benchmark
	benchmark.SetSetupFunc(func() error {
		// Pre-create test directory structure
		return nil
	})

	// Set the benchmark function
	benchmark.SetBenchmarkFunc(func(b *testing.B) {
		ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
		defer cancel()

		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			// Test directory creation with many files
			start := time.Now()
			err := simulateHighFileCountDirectoryCreation(ctx, config.FileCount, config.FileSize, config.Concurrency)
			creationLatency := time.Since(start)
			benchmark.RecordLatency(creationLatency)

			if err != nil {
				b.Fatalf("High file count directory creation failed: %v", err)
			}

			// Test directory listing
			start = time.Now()
			err = simulateDirectoryListing(ctx, config.FileCount)
			listingLatency := time.Since(start)
			benchmark.RecordLatency(listingLatency)

			if err != nil {
				b.Fatalf("Directory listing failed: %v", err)
			}

			// Test directory cleanup
			start = time.Now()
			err = simulateDirectoryCleanup(ctx, config.FileCount, config.Concurrency)
			cleanupLatency := time.Since(start)
			benchmark.RecordLatency(cleanupLatency)

			if err != nil {
				b.Fatalf("Directory cleanup failed: %v", err)
			}

			b.Logf("High file count operations completed: creation=%v, listing=%v, cleanup=%v",
				creationLatency, listingLatency, cleanupLatency)
		}
	})

	// Run the benchmark
	benchmark.Run(b)
}

// SustainedOperationTest tests sustained operation over time
func SustainedOperationTest(b *testing.B, framework *TestFramework, config SustainedOperationTestConfig, thresholds PerformanceThresholds) {
	// Create a benchmark for sustained operations
	benchmark := NewPerformanceBenchmark(
		"SustainedOperationTest",
		fmt.Sprintf("Test for sustained operations over %v", config.Duration),
		thresholds,
		framework.Config.ArtifactsDir,
	)

	// Set up memory tracking
	memoryTracker := NewMemoryTracker()

	// Set up the benchmark
	benchmark.SetSetupFunc(func() error {
		// Initial memory sample
		memoryTracker.Sample()
		return nil
	})

	// Set the benchmark function
	benchmark.SetBenchmarkFunc(func(b *testing.B) {
		ctx, cancel := context.WithTimeout(context.Background(), config.Duration)
		defer cancel()

		b.ResetTimer()

		// Start memory monitoring
		memoryTicker := time.NewTicker(config.MemoryCheckInterval)
		defer memoryTicker.Stop()

		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				case <-memoryTicker.C:
					memoryTracker.Sample()
				}
			}
		}()

		// Calculate operation interval
		operationInterval := time.Second / time.Duration(config.OperationsPerSecond)

		// Run sustained operations
		operationTicker := time.NewTicker(operationInterval)
		defer operationTicker.Stop()

		operationCount := 0
		for {
			select {
			case <-ctx.Done():
				b.Logf("Sustained operation test completed: %d operations over %v", operationCount, config.Duration)
				return
			case <-operationTicker.C:
				start := time.Now()
				err := simulateSustainedOperation(ctx, operationCount)
				latency := time.Since(start)
				benchmark.RecordLatency(latency)

				if err != nil {
					b.Logf("Operation %d failed: %v", operationCount, err)
				}

				operationCount++

				// Log progress every 100 operations
				if operationCount%100 == 0 {
					b.Logf("Completed %d operations", operationCount)
				}
			}
		}
	})

	// Set teardown function to check memory growth
	benchmark.SetTeardownFunc(func() error {
		memoryTracker.Sample()
		growth, err := memoryTracker.GetMemoryGrowth()
		if err != nil {
			return fmt.Errorf("failed to calculate memory growth: %v", err)
		}

		if growth > 50 { // 50MB threshold for sustained operations
			return fmt.Errorf("excessive memory growth detected: %d MB", growth)
		}

		return nil
	})

	// Run the benchmark
	benchmark.Run(b)
}

// MemoryLeakDetectionTest tests for memory leaks over time
func MemoryLeakDetectionTest(b *testing.B, framework *TestFramework, config MemoryLeakTestConfig, thresholds PerformanceThresholds) {
	// Create a benchmark for memory leak detection
	benchmark := NewPerformanceBenchmark(
		"MemoryLeakDetectionTest",
		fmt.Sprintf("Test for memory leaks over %v", config.Duration),
		thresholds,
		framework.Config.ArtifactsDir,
	)

	// Set up memory tracking
	memoryTracker := NewMemoryTracker()

	// Set up the benchmark
	benchmark.SetSetupFunc(func() error {
		// Force garbage collection and take initial sample
		runtime.GC()
		runtime.GC() // Call twice to ensure cleanup
		time.Sleep(100 * time.Millisecond)
		memoryTracker.Sample()
		return nil
	})

	// Set the benchmark function
	benchmark.SetBenchmarkFunc(func(b *testing.B) {
		ctx, cancel := context.WithTimeout(context.Background(), config.Duration)
		defer cancel()

		b.ResetTimer()

		// Start memory monitoring
		memoryTicker := time.NewTicker(config.SamplingInterval)
		defer memoryTicker.Stop()

		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				case <-memoryTicker.C:
					runtime.GC() // Force GC before sampling
					time.Sleep(50 * time.Millisecond)
					memoryTracker.Sample()
				}
			}
		}()

		// Run operation cycles
		cycleTicker := time.NewTicker(config.CycleInterval)
		defer cycleTicker.Stop()

		cycleCount := 0
		for {
			select {
			case <-ctx.Done():
				b.Logf("Memory leak detection test completed: %d cycles over %v", cycleCount, config.Duration)
				return
			case <-cycleTicker.C:
				// Run a cycle of operations
				start := time.Now()
				err := simulateMemoryLeakTestCycle(ctx, config.OperationsPerCycle)
				latency := time.Since(start)
				benchmark.RecordLatency(latency)

				if err != nil {
					b.Logf("Cycle %d failed: %v", cycleCount, err)
				}

				cycleCount++

				// Force garbage collection after each cycle
				runtime.GC()

				// Log progress every 10 cycles
				if cycleCount%10 == 0 {
					samples := memoryTracker.GetSamples()
					if len(samples) > 0 {
						currentMem := samples[len(samples)-1].AllocMB
						b.Logf("Completed %d cycles, current memory: %d MB", cycleCount, currentMem)
					}
				}
			}
		}
	})

	// Set teardown function to check for memory leaks
	benchmark.SetTeardownFunc(func() error {
		// Force final garbage collection
		runtime.GC()
		runtime.GC()
		time.Sleep(100 * time.Millisecond)
		memoryTracker.Sample()

		growth, err := memoryTracker.GetMemoryGrowth()
		if err != nil {
			return fmt.Errorf("failed to calculate memory growth: %v", err)
		}

		if growth > config.MaxMemoryGrowthMB {
			samples := memoryTracker.GetSamples()
			return fmt.Errorf("memory leak detected: growth=%d MB (max allowed: %d MB), samples=%d",
				growth, config.MaxMemoryGrowthMB, len(samples))
		}

		return nil
	})

	// Run the benchmark
	benchmark.Run(b)
}

// simulateLargeFileUpload simulates uploading a large file in chunks
func simulateLargeFileUpload(ctx context.Context, fileSize, chunkSize int64) error {
	// Create a temporary file for simulation
	tmpFile, err := os.CreateTemp("", "large-upload-*.tmp")
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Simulate writing data in chunks
	written := int64(0)
	buffer := make([]byte, chunkSize)

	for written < fileSize {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		remaining := fileSize - written
		if remaining < chunkSize {
			buffer = buffer[:remaining]
		}

		// Fill buffer with random data
		if _, err := rand.Read(buffer); err != nil {
			return fmt.Errorf("failed to generate random data: %v", err)
		}

		// Simulate network delay
		time.Sleep(10 * time.Millisecond)

		// Write chunk
		n, err := tmpFile.Write(buffer)
		if err != nil {
			return fmt.Errorf("failed to write chunk: %v", err)
		}

		written += int64(n)
	}

	return nil
}

// simulateLargeFileDownload simulates downloading a large file in chunks
func simulateLargeFileDownload(ctx context.Context, fileSize, chunkSize int64) error {
	// Create a temporary file for simulation
	tmpFile, err := os.CreateTemp("", "large-download-*.tmp")
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Pre-populate the file with data
	buffer := make([]byte, chunkSize)
	written := int64(0)

	for written < fileSize {
		remaining := fileSize - written
		if remaining < chunkSize {
			buffer = buffer[:remaining]
		}

		if _, err := rand.Read(buffer); err != nil {
			return fmt.Errorf("failed to generate random data: %v", err)
		}

		n, err := tmpFile.Write(buffer)
		if err != nil {
			return fmt.Errorf("failed to write data: %v", err)
		}

		written += int64(n)
	}

	// Reset file position
	if _, err := tmpFile.Seek(0, 0); err != nil {
		return fmt.Errorf("failed to seek to beginning: %v", err)
	}

	// Simulate reading data in chunks
	read := int64(0)
	readBuffer := make([]byte, chunkSize)

	for read < fileSize {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Simulate network delay
		time.Sleep(10 * time.Millisecond)

		// Read chunk
		n, err := tmpFile.Read(readBuffer)
		if err != nil && err != io.EOF {
			return fmt.Errorf("failed to read chunk: %v", err)
		}

		read += int64(n)

		if err == io.EOF {
			break
		}
	}

	return nil
}

// simulateHighFileCountDirectoryCreation simulates creating a directory with many files
func simulateHighFileCountDirectoryCreation(ctx context.Context, fileCount int, fileSize int64, concurrency int) error {
	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "high-file-count-*")
	if err != nil {
		return fmt.Errorf("failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create files concurrently
	semaphore := make(chan struct{}, concurrency)
	errChan := make(chan error, fileCount)
	var wg sync.WaitGroup

	for i := 0; i < fileCount; i++ {
		wg.Add(1)
		go func(fileIndex int) {
			defer wg.Done()

			select {
			case <-ctx.Done():
				errChan <- ctx.Err()
				return
			case semaphore <- struct{}{}:
			}
			defer func() { <-semaphore }()

			// Create file
			fileName := filepath.Join(tmpDir, fmt.Sprintf("file_%06d.txt", fileIndex))
			file, err := os.Create(fileName)
			if err != nil {
				errChan <- fmt.Errorf("failed to create file %s: %v", fileName, err)
				return
			}
			defer file.Close()

			// Write data to file
			data := make([]byte, fileSize)
			if _, err := rand.Read(data); err != nil {
				errChan <- fmt.Errorf("failed to generate data for file %s: %v", fileName, err)
				return
			}

			if _, err := file.Write(data); err != nil {
				errChan <- fmt.Errorf("failed to write data to file %s: %v", fileName, err)
				return
			}

			// Simulate small delay
			time.Sleep(1 * time.Millisecond)
		}(i)
	}

	// Wait for all files to be created
	wg.Wait()
	close(errChan)

	// Check for errors
	for err := range errChan {
		if err != nil {
			return err
		}
	}

	return nil
}

// simulateDirectoryListing simulates listing a directory with many files
func simulateDirectoryListing(ctx context.Context, expectedFileCount int) error {
	// Create a temporary directory with files
	tmpDir, err := os.MkdirTemp("", "dir-listing-*")
	if err != nil {
		return fmt.Errorf("failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create files for listing
	for i := 0; i < expectedFileCount; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		fileName := filepath.Join(tmpDir, fmt.Sprintf("file_%06d.txt", i))
		file, err := os.Create(fileName)
		if err != nil {
			return fmt.Errorf("failed to create file for listing: %v", err)
		}
		file.Close()

		// Add small delay every 1000 files to prevent overwhelming the system
		if i%1000 == 0 {
			time.Sleep(10 * time.Millisecond)
		}
	}

	// Perform directory listing
	start := time.Now()
	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		return fmt.Errorf("failed to list directory: %v", err)
	}

	listingDuration := time.Since(start)

	if len(entries) != expectedFileCount {
		return fmt.Errorf("expected %d files, found %d", expectedFileCount, len(entries))
	}

	// Log performance metrics
	if listingDuration > 5*time.Second {
		return fmt.Errorf("directory listing took too long: %v", listingDuration)
	}

	return nil
}

// simulateDirectoryCleanup simulates cleaning up a directory with many files
func simulateDirectoryCleanup(ctx context.Context, fileCount int, concurrency int) error {
	// Create a temporary directory with files
	tmpDir, err := os.MkdirTemp("", "dir-cleanup-*")
	if err != nil {
		return fmt.Errorf("failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tmpDir) // Ensure cleanup even if function fails

	// Create files to clean up
	fileNames := make([]string, fileCount)
	for i := 0; i < fileCount; i++ {
		fileName := filepath.Join(tmpDir, fmt.Sprintf("file_%06d.txt", i))
		file, err := os.Create(fileName)
		if err != nil {
			return fmt.Errorf("failed to create file for cleanup test: %v", err)
		}
		file.Close()
		fileNames[i] = fileName
	}

	// Delete files concurrently
	semaphore := make(chan struct{}, concurrency)
	errChan := make(chan error, fileCount)
	var wg sync.WaitGroup

	for _, fileName := range fileNames {
		wg.Add(1)
		go func(fn string) {
			defer wg.Done()

			select {
			case <-ctx.Done():
				errChan <- ctx.Err()
				return
			case semaphore <- struct{}{}:
			}
			defer func() { <-semaphore }()

			if err := os.Remove(fn); err != nil {
				errChan <- fmt.Errorf("failed to remove file %s: %v", fn, err)
				return
			}

			// Simulate small delay
			time.Sleep(1 * time.Millisecond)
		}(fileName)
	}

	// Wait for all files to be deleted
	wg.Wait()
	close(errChan)

	// Check for errors
	for err := range errChan {
		if err != nil {
			return err
		}
	}

	return nil
}

// simulateSustainedOperation simulates a sustained operation
func simulateSustainedOperation(ctx context.Context, operationIndex int) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Simulate different types of operations
	operationType := operationIndex % 4

	switch operationType {
	case 0:
		// File read operation
		return simulateFileRead(ctx)
	case 1:
		// File write operation
		return simulateFileWrite(ctx)
	case 2:
		// Directory listing operation
		return simulateDirectoryList(ctx)
	case 3:
		// Metadata operation
		return simulateMetadataOperation(ctx)
	default:
		return fmt.Errorf("unknown operation type: %d", operationType)
	}
}

// simulateMemoryLeakTestCycle simulates a cycle of operations for memory leak testing
func simulateMemoryLeakTestCycle(ctx context.Context, operationsPerCycle int) error {
	for i := 0; i < operationsPerCycle; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Perform various operations that might cause memory leaks
		if err := simulateMemoryIntensiveOperation(ctx, i); err != nil {
			return fmt.Errorf("operation %d failed: %v", i, err)
		}

		// Small delay between operations
		time.Sleep(1 * time.Millisecond)
	}

	return nil
}

// simulateFileRead simulates a file read operation
func simulateFileRead(ctx context.Context) error {
	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "read-test-*.tmp")
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Write some data
	data := make([]byte, 1024) // 1KB
	if _, err := rand.Read(data); err != nil {
		return fmt.Errorf("failed to generate data: %v", err)
	}

	if _, err := tmpFile.Write(data); err != nil {
		return fmt.Errorf("failed to write data: %v", err)
	}

	// Reset position
	if _, err := tmpFile.Seek(0, 0); err != nil {
		return fmt.Errorf("failed to seek: %v", err)
	}

	// Read the data back
	readData := make([]byte, 1024)
	if _, err := tmpFile.Read(readData); err != nil {
		return fmt.Errorf("failed to read data: %v", err)
	}

	// Simulate processing delay
	time.Sleep(1 * time.Millisecond)

	return nil
}

// simulateFileWrite simulates a file write operation
func simulateFileWrite(ctx context.Context) error {
	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "write-test-*.tmp")
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Write data
	data := make([]byte, 1024) // 1KB
	if _, err := rand.Read(data); err != nil {
		return fmt.Errorf("failed to generate data: %v", err)
	}

	if _, err := tmpFile.Write(data); err != nil {
		return fmt.Errorf("failed to write data: %v", err)
	}

	// Sync to disk
	if err := tmpFile.Sync(); err != nil {
		return fmt.Errorf("failed to sync: %v", err)
	}

	// Simulate processing delay
	time.Sleep(1 * time.Millisecond)

	return nil
}

// simulateDirectoryList simulates a directory listing operation
func simulateDirectoryList(ctx context.Context) error {
	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "list-test-*")
	if err != nil {
		return fmt.Errorf("failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create some files
	for i := 0; i < 10; i++ {
		fileName := filepath.Join(tmpDir, fmt.Sprintf("file_%d.txt", i))
		file, err := os.Create(fileName)
		if err != nil {
			return fmt.Errorf("failed to create file: %v", err)
		}
		file.Close()
	}

	// List the directory
	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		return fmt.Errorf("failed to list directory: %v", err)
	}

	if len(entries) != 10 {
		return fmt.Errorf("expected 10 entries, got %d", len(entries))
	}

	// Simulate processing delay
	time.Sleep(1 * time.Millisecond)

	return nil
}

// simulateMetadataOperation simulates a metadata operation
func simulateMetadataOperation(ctx context.Context) error {
	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "metadata-test-*.tmp")
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Get file info
	info, err := tmpFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %v", err)
	}

	// Check various metadata
	_ = info.Name()
	_ = info.Size()
	_ = info.Mode()
	_ = info.ModTime()
	_ = info.IsDir()

	// Simulate processing delay
	time.Sleep(1 * time.Millisecond)

	return nil
}

// simulateMemoryIntensiveOperation simulates an operation that might cause memory leaks
func simulateMemoryIntensiveOperation(ctx context.Context, operationIndex int) error {
	// Create temporary data structures that should be garbage collected
	data := make([]byte, 10*1024) // 10KB
	if _, err := rand.Read(data); err != nil {
		return fmt.Errorf("failed to generate data: %v", err)
	}

	// Create temporary maps and slices
	tempMap := make(map[string][]byte)
	tempSlice := make([][]byte, 100)

	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("key_%d_%d", operationIndex, i)
		value := make([]byte, 100)
		if _, err := rand.Read(value); err != nil {
			return fmt.Errorf("failed to generate value: %v", err)
		}

		tempMap[key] = value
		tempSlice[i] = value
	}

	// Simulate some processing
	for key, value := range tempMap {
		_ = key
		_ = value
	}

	for _, slice := range tempSlice {
		_ = slice
	}

	// Simulate processing delay
	time.Sleep(1 * time.Millisecond)

	// Clear references to help GC (though this should happen automatically)
	tempMap = nil
	tempSlice = nil
	data = nil

	return nil
}
