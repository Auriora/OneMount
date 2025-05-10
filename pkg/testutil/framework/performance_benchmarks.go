// Package testutil provides testing utilities for the OneMount project.
package framework

import (
	"context"
	"fmt"
	"github.com/auriora/onemount/pkg/testutil"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

// FileDownloadBenchmark benchmarks file download performance
func FileDownloadBenchmark(b *testing.B, framework *TestFramework, fileSize int64, thresholds PerformanceThresholds) {
	// Create a benchmark
	benchmark := NewPerformanceBenchmark(
		"FileDownloadBenchmark",
		fmt.Sprintf("Benchmark for downloading files of size %d bytes", fileSize),
		thresholds,
		framework.Config.ArtifactsDir,
	)

	// Set up the benchmark
	benchmark.SetSetupFunc(func() error {
		// Get the mock graph provider
		mockGraph, exists := framework.GetMockProvider("graph")
		if !exists {
			return fmt.Errorf("mock graph provider not found")
		}

		// Configure the mock to return a file of the specified size
		// This is a simplified example - in a real implementation, you would
		// configure the mock to return specific responses for specific API calls
		// based on the mock provider's interface
		_ = mockGraph

		return nil
	})

	// Set the benchmark function
	benchmark.SetBenchmarkFunc(func(b *testing.B) {
		// Reset the benchmark timer
		b.ResetTimer()

		// Run the benchmark
		for i := 0; i < b.N; i++ {
			// Record the start time
			start := time.Now()

			// Simulate downloading a file
			// In a real implementation, you would use the actual API client
			// to download a file from OneDrive
			err := simulateFileDownload(fileSize)

			// Record the latency
			latency := time.Since(start)
			benchmark.RecordLatency(latency)

			// Check for errors
			if err != nil {
				b.Fatalf("Error downloading file: %v", err)
			}
		}
	})

	// Run the benchmark
	benchmark.Run(b)
}

// FileUploadBenchmark benchmarks file upload performance
func FileUploadBenchmark(b *testing.B, framework *TestFramework, fileSize int64, thresholds PerformanceThresholds) {
	// Create a benchmark
	benchmark := NewPerformanceBenchmark(
		"FileUploadBenchmark",
		fmt.Sprintf("Benchmark for uploading files of size %d bytes", fileSize),
		thresholds,
		framework.Config.ArtifactsDir,
	)

	// Set up the benchmark
	benchmark.SetSetupFunc(func() error {
		// Get the mock graph provider
		mockGraph, exists := framework.GetMockProvider("graph")
		if !exists {
			return fmt.Errorf("mock graph provider not found")
		}

		// Configure the mock to handle file uploads
		// This is a simplified example - in a real implementation, you would
		// configure the mock to handle specific API calls for file uploads
		_ = mockGraph

		return nil
	})

	// Set the benchmark function
	benchmark.SetBenchmarkFunc(func(b *testing.B) {
		// Reset the benchmark timer
		b.ResetTimer()

		// Run the benchmark
		for i := 0; i < b.N; i++ {
			// Record the start time
			start := time.Now()

			// Simulate uploading a file
			// In a real implementation, you would use the actual API client
			// to upload a file to OneDrive
			err := simulateFileUpload(fileSize)

			// Record the latency
			latency := time.Since(start)
			benchmark.RecordLatency(latency)

			// Check for errors
			if err != nil {
				b.Fatalf("Error uploading file: %v", err)
			}
		}
	})

	// Run the benchmark
	benchmark.Run(b)
}

// MetadataOperationsBenchmark benchmarks metadata operations performance
func MetadataOperationsBenchmark(b *testing.B, framework *TestFramework, numItems int, thresholds PerformanceThresholds) {
	// Create a benchmark
	benchmark := NewPerformanceBenchmark(
		"MetadataOperationsBenchmark",
		fmt.Sprintf("Benchmark for metadata operations with %d items", numItems),
		thresholds,
		framework.Config.ArtifactsDir,
	)

	// Set up the benchmark
	benchmark.SetSetupFunc(func() error {
		// Get the mock graph provider
		mockGraph, exists := framework.GetMockProvider("graph")
		if !exists {
			return fmt.Errorf("mock graph provider not found")
		}

		// Configure the mock to return metadata for the specified number of items
		// This is a simplified example - in a real implementation, you would
		// configure the mock to return specific responses for specific API calls
		_ = mockGraph

		return nil
	})

	// Set the benchmark function
	benchmark.SetBenchmarkFunc(func(b *testing.B) {
		// Reset the benchmark timer
		b.ResetTimer()

		// Run the benchmark
		for i := 0; i < b.N; i++ {
			// Record the start time
			start := time.Now()

			// Simulate metadata operations
			// In a real implementation, you would use the actual API client
			// to perform metadata operations (list files, get file info, etc.)
			err := simulateMetadataOperations(context.Background(), numItems)

			// Record the latency
			latency := time.Since(start)
			benchmark.RecordLatency(latency)

			// Check for errors
			if err != nil {
				b.Fatalf("Error performing metadata operations: %v", err)
			}
		}
	})

	// Run the benchmark
	benchmark.Run(b)
}

// ConcurrentOperationsBenchmark benchmarks concurrent operations performance
func ConcurrentOperationsBenchmark(b *testing.B, framework *TestFramework, concurrency int, thresholds PerformanceThresholds) {
	// Create a benchmark
	benchmark := NewPerformanceBenchmark(
		"ConcurrentOperationsBenchmark",
		fmt.Sprintf("Benchmark for concurrent operations with %d concurrent operations", concurrency),
		thresholds,
		framework.Config.ArtifactsDir,
	)

	// Set up the benchmark
	benchmark.SetSetupFunc(func() error {
		// Get the mock graph provider
		mockGraph, exists := framework.GetMockProvider("graph")
		if !exists {
			return fmt.Errorf("mock graph provider not found")
		}

		// Configure the mock to handle concurrent operations
		// This is a simplified example - in a real implementation, you would
		// configure the mock to handle specific API calls for concurrent operations
		_ = mockGraph

		return nil
	})

	// Set the benchmark function
	benchmark.SetBenchmarkFunc(func(b *testing.B) {
		// Reset the benchmark timer
		b.ResetTimer()

		// Run the benchmark
		for i := 0; i < b.N; i++ {
			// Record the start time
			start := time.Now()

			// Simulate concurrent operations
			// In a real implementation, you would use the actual API client
			// to perform concurrent operations
			err := simulateConcurrentOperations(concurrency)

			// Record the latency
			latency := time.Since(start)
			benchmark.RecordLatency(latency)

			// Check for errors
			if err != nil {
				b.Fatalf("Error performing concurrent operations: %v", err)
			}
		}
	})

	// Run the benchmark
	benchmark.Run(b)
}

// LoadTestFileDownload performs a load test for file downloads
func LoadTestFileDownload(ctx context.Context, framework *TestFramework, fileSize int64, concurrency int, duration time.Duration, thresholds PerformanceThresholds) error {
	// Create a benchmark
	benchmark := NewPerformanceBenchmark(
		"LoadTestFileDownload",
		fmt.Sprintf("Load test for downloading files of size %d bytes with %d concurrent operations", fileSize, concurrency),
		thresholds,
		framework.Config.ArtifactsDir,
	)

	// Set up the benchmark
	benchmark.SetSetupFunc(func() error {
		// Get the mock graph provider
		mockGraph, exists := framework.GetMockProvider("graph")
		if !exists {
			return fmt.Errorf("mock graph provider not found")
		}

		// Configure the mock to return a file of the specified size
		// This is a simplified example - in a real implementation, you would
		// configure the mock to return specific responses for specific API calls
		_ = mockGraph

		return nil
	})

	// Set up the load test
	// Calculate appropriate ramp-up time based on test duration
	// For short tests, use a shorter ramp-up time to avoid appearing to hang
	rampUpTime := time.Second
	if duration > 5*time.Second {
		rampUpTime = 5 * time.Second
	}

	fmt.Printf("Starting LoadTestFileDownload with file size: %d bytes, concurrency: %d, duration: %s, ramp-up: %s\n",
		fileSize, concurrency, duration, rampUpTime)

	loadTest := &LoadTest{
		Concurrency: concurrency,
		Duration:    duration,
		RampUp:      rampUpTime,
		Scenario: func(ctx context.Context) error {
			// Simulate downloading a file
			// In a real implementation, you would use the actual API client
			// to download a file from OneDrive
			return simulateFileDownload(fileSize)
		},
	}

	// Set the load test
	benchmark.SetLoadTest(loadTest)

	// Run the load test
	return benchmark.RunLoadTest(ctx)
}

// LoadTestFileUpload performs a load test for file uploads
func LoadTestFileUpload(ctx context.Context, framework *TestFramework, fileSize int64, concurrency int, duration time.Duration, thresholds PerformanceThresholds) error {
	// Create a benchmark
	benchmark := NewPerformanceBenchmark(
		"LoadTestFileUpload",
		fmt.Sprintf("Load test for uploading files of size %d bytes with %d concurrent operations", fileSize, concurrency),
		thresholds,
		framework.Config.ArtifactsDir,
	)

	// Set up the benchmark
	benchmark.SetSetupFunc(func() error {
		// Get the mock graph provider
		mockGraph, exists := framework.GetMockProvider("graph")
		if !exists {
			return fmt.Errorf("mock graph provider not found")
		}

		// Configure the mock to handle file uploads
		// This is a simplified example - in a real implementation, you would
		// configure the mock to handle specific API calls for file uploads
		_ = mockGraph

		return nil
	})

	// Set up the load test
	// Calculate appropriate ramp-up time based on test duration
	// For short tests, use a shorter ramp-up time to avoid appearing to hang
	rampUpTime := time.Second
	if duration > 5*time.Second {
		rampUpTime = 5 * time.Second
	}

	fmt.Printf("Starting LoadTestFileUpload with file size: %d bytes, concurrency: %d, duration: %s, ramp-up: %s\n",
		fileSize, concurrency, duration, rampUpTime)

	loadTest := &LoadTest{
		Concurrency: concurrency,
		Duration:    duration,
		RampUp:      rampUpTime,
		Scenario: func(ctx context.Context) error {
			// Simulate uploading a file
			// In a real implementation, you would use the actual API client
			// to upload a file to OneDrive
			return simulateFileUpload(fileSize)
		},
	}

	// Set the load test
	benchmark.SetLoadTest(loadTest)

	// Run the load test
	return benchmark.RunLoadTest(ctx)
}

// LoadTestMetadataOperations performs a load test for metadata operations
func LoadTestMetadataOperations(ctx context.Context, framework *TestFramework, numItems int, concurrency int, duration time.Duration, thresholds PerformanceThresholds) error {
	// Create a benchmark
	benchmark := NewPerformanceBenchmark(
		"LoadTestMetadataOperations",
		fmt.Sprintf("Load test for metadata operations with %d items and %d concurrent operations", numItems, concurrency),
		thresholds,
		framework.Config.ArtifactsDir,
	)

	// Set up the benchmark
	benchmark.SetSetupFunc(func() error {
		// Get the mock graph provider
		mockGraph, exists := framework.GetMockProvider("graph")
		if !exists {
			return fmt.Errorf("mock graph provider not found")
		}

		// Configure the mock to return metadata for the specified number of items
		// This is a simplified example - in a real implementation, you would
		// configure the mock to return specific responses for specific API calls
		_ = mockGraph

		return nil
	})

	// Set up the load test
	// Calculate appropriate ramp-up time based on test duration
	// For short tests, use a shorter ramp-up time to avoid appearing to hang
	rampUpTime := time.Second
	if duration > 5*time.Second {
		rampUpTime = 5 * time.Second
	}

	fmt.Printf("Starting LoadTestMetadataOperations with items: %d, concurrency: %d, duration: %s, ramp-up: %s\n",
		numItems, concurrency, duration, rampUpTime)

	loadTest := &LoadTest{
		Concurrency: concurrency,
		Duration:    duration,
		RampUp:      rampUpTime,
		Scenario: func(ctx context.Context) error {
			// Simulate metadata operations
			// In a real implementation, you would use the actual API client
			// to perform metadata operations (list files, get file info, etc.)
			return simulateMetadataOperations(ctx, numItems)
		},
	}

	// Set the load test
	benchmark.SetLoadTest(loadTest)

	// Run the load test
	return benchmark.RunLoadTest(ctx)
}

// Helper functions for simulating operations

// simulateFileDownload simulates downloading a file of the specified size
func simulateFileDownload(fileSize int64) error {
	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "download-*.tmp")
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Create a random reader to simulate downloading data
	randomReader := io.LimitReader(rand.New(rand.NewSource(time.Now().UnixNano())), fileSize)

	// Copy the data to the file
	_, err = io.Copy(tmpFile, randomReader)
	if err != nil {
		return fmt.Errorf("failed to write to temporary file: %v", err)
	}

	return nil
}

// simulateFileUpload simulates uploading a file of the specified size
func simulateFileUpload(fileSize int64) error {
	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "upload-*.tmp")
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Write random data to the file
	randomData := make([]byte, fileSize)
	rand.Read(randomData)
	if _, err := tmpFile.Write(randomData); err != nil {
		return fmt.Errorf("failed to write to temporary file: %v", err)
	}

	// Simulate uploading the file
	// In a real implementation, you would use the actual API client
	// to upload the file to OneDrive
	return nil
}

// simulateMetadataOperations simulates performing metadata operations for the specified number of items
func simulateMetadataOperations(ctx context.Context, numItems int) error {
	fmt.Printf("Starting simulateMetadataOperations with %d items\n", numItems)

	// Check if context is already canceled
	select {
	case <-ctx.Done():
		fmt.Printf("Context already canceled at start of simulateMetadataOperations: %v\n", ctx.Err())
		return ctx.Err()
	default:
		// Continue with the operation
		fmt.Printf("Context is valid, proceeding with operation\n")
	}

	// Create a temporary directory
	fmt.Printf("Creating temporary directory\n")
	tmpDir, err := os.MkdirTemp(testutil.TestSandboxTmpDir, "metadata-")
	if err != nil {
		fmt.Printf("Failed to create temporary directory: %v\n", err)
		return fmt.Errorf("failed to create temporary directory: %v", err)
	}
	fmt.Printf("Created temporary directory: %s\n", tmpDir)
	defer os.RemoveAll(tmpDir)

	// Create files in the directory
	fmt.Printf("Creating %d files in the directory\n", numItems)
	for i := 0; i < numItems; i++ {
		// Check if context is canceled before each iteration
		select {
		case <-ctx.Done():
			fmt.Printf("Context canceled during file creation (file %d): %v\n", i, ctx.Err())
			return ctx.Err()
		default:
			// Continue with the operation
		}

		fileName := filepath.Join(tmpDir, fmt.Sprintf("file-%d.txt", i))
		fmt.Printf("Creating file: %s\n", fileName)
		if err := os.WriteFile(fileName, []byte("test"), 0644); err != nil {
			fmt.Printf("Failed to create file %s: %v\n", fileName, err)
			return fmt.Errorf("failed to create file: %v", err)
		}
	}

	// Check if context is canceled before listing files
	select {
	case <-ctx.Done():
		fmt.Printf("Context canceled before listing files: %v\n", ctx.Err())
		return ctx.Err()
	default:
		// Continue with the operation
		fmt.Printf("Listing files in directory: %s\n", tmpDir)
	}

	// Simulate listing files
	files, err := os.ReadDir(tmpDir)
	if err != nil {
		fmt.Printf("Failed to list files in directory %s: %v\n", tmpDir, err)
		return fmt.Errorf("failed to list files: %v", err)
	}
	fmt.Printf("Found %d files in directory\n", len(files))

	// Simulate getting file info
	fmt.Printf("Getting file info for each file\n")
	for i, file := range files {
		// Check if context is canceled before each iteration
		select {
		case <-ctx.Done():
			fmt.Printf("Context canceled during file info retrieval (file %d): %v\n", i, ctx.Err())
			return ctx.Err()
		default:
			// Continue with the operation
		}

		filePath := filepath.Join(tmpDir, file.Name())
		fmt.Printf("Getting info for file: %s\n", filePath)
		_, err := os.Stat(filePath)
		if err != nil {
			fmt.Printf("Failed to get file info for %s: %v\n", filePath, err)
			return fmt.Errorf("failed to get file info: %v", err)
		}
	}

	fmt.Printf("Completed simulateMetadataOperations successfully\n")
	return nil
}

// simulateConcurrentOperations simulates performing concurrent operations
func simulateConcurrentOperations(concurrency int) error {
	// Create a wait group
	var wg sync.WaitGroup
	wg.Add(concurrency)

	// Create a channel for errors
	errChan := make(chan error, concurrency)

	// Start goroutines for concurrent operations
	for i := 0; i < concurrency; i++ {
		go func(id int) {
			defer wg.Done()

			// Simulate a random operation
			var err error
			switch rand.Intn(3) {
			case 0:
				// Simulate file download
				err = simulateFileDownload(1024 * 1024) // 1MB
			case 1:
				// Simulate file upload
				err = simulateFileUpload(1024 * 1024) // 1MB
			case 2:
				// Simulate metadata operations
				err = simulateMetadataOperations(context.Background(), 10)
			}

			// Send error to channel if any
			if err != nil {
				errChan <- fmt.Errorf("operation %d failed: %v", id, err)
			}
		}(i)
	}

	// Wait for all goroutines to finish
	wg.Wait()
	close(errChan)

	// Check for errors
	for err := range errChan {
		return err
	}

	return nil
}
