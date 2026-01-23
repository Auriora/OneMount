package util

import (
	"context"
	"io"
	"sync"
	"time"
)

// BandwidthThrottler implements adaptive network bandwidth throttling
// to prevent network saturation and ensure fair resource usage.
type BandwidthThrottler struct {
	maxBytesPerSecond int64
	bytesTransferred  int64
	startTime         time.Time
	mutex             sync.Mutex
	enabled           bool
}

// NewBandwidthThrottler creates a new bandwidth throttler with the specified limit.
// maxBytesPerSecond: Maximum bytes per second allowed (0 = unlimited)
func NewBandwidthThrottler(maxBytesPerSecond int64) *BandwidthThrottler {
	return &BandwidthThrottler{
		maxBytesPerSecond: maxBytesPerSecond,
		bytesTransferred:  0,
		startTime:         time.Now(),
		enabled:           maxBytesPerSecond > 0,
	}
}

// Wait blocks until the specified number of bytes can be transferred
// without exceeding the bandwidth limit. Returns immediately if throttling
// is disabled or if the context is cancelled.
func (bt *BandwidthThrottler) Wait(ctx context.Context, bytes int64) error {
	if !bt.enabled || bytes <= 0 {
		return nil
	}

	bt.mutex.Lock()

	// Update bytes transferred
	bt.bytesTransferred += bytes

	// Calculate elapsed time
	elapsed := time.Since(bt.startTime).Seconds()
	if elapsed <= 0 {
		bt.mutex.Unlock()
		return nil
	}

	// Calculate current bandwidth usage
	currentBandwidth := float64(bt.bytesTransferred) / elapsed

	// If we're exceeding the limit, calculate how long to sleep
	if currentBandwidth > float64(bt.maxBytesPerSecond) {
		// Calculate the time we should have taken to transfer this many bytes
		expectedDuration := float64(bt.bytesTransferred) / float64(bt.maxBytesPerSecond)

		// Calculate how much longer we need to wait
		sleepDuration := time.Duration((expectedDuration - elapsed) * float64(time.Second))

		if sleepDuration > 0 {
			// Check context before sleeping
			select {
			case <-ctx.Done():
				bt.mutex.Unlock()
				return ctx.Err()
			default:
			}

			// Release the lock while sleeping to allow other goroutines to check
			bt.mutex.Unlock()

			// Create a timer that can be stopped to prevent goroutine leaks
			timer := time.NewTimer(sleepDuration)
			defer timer.Stop()

			// Sleep with context cancellation support
			select {
			case <-timer.C:
				// Sleep completed normally
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}

	bt.mutex.Unlock()
	return nil
}

// Reset resets the throttler's counters. This should be called periodically
// to prevent overflow and to adapt to changing network conditions.
func (bt *BandwidthThrottler) Reset() {
	if !bt.enabled {
		return
	}

	bt.mutex.Lock()
	defer bt.mutex.Unlock()

	bt.bytesTransferred = 0
	bt.startTime = time.Now()
}

// SetLimit updates the bandwidth limit. Set to 0 to disable throttling.
func (bt *BandwidthThrottler) SetLimit(maxBytesPerSecond int64) {
	bt.mutex.Lock()
	defer bt.mutex.Unlock()

	bt.maxBytesPerSecond = maxBytesPerSecond
	bt.enabled = maxBytesPerSecond > 0

	// Reset counters when limit changes
	bt.bytesTransferred = 0
	bt.startTime = time.Now()
}

// GetCurrentBandwidth returns the current bandwidth usage in bytes per second
func (bt *BandwidthThrottler) GetCurrentBandwidth() float64 {
	if !bt.enabled {
		return 0
	}

	bt.mutex.Lock()
	defer bt.mutex.Unlock()

	elapsed := time.Since(bt.startTime).Seconds()
	if elapsed <= 0 {
		return 0
	}

	return float64(bt.bytesTransferred) / elapsed
}

// IsEnabled returns whether throttling is currently enabled
func (bt *BandwidthThrottler) IsEnabled() bool {
	bt.mutex.Lock()
	defer bt.mutex.Unlock()
	return bt.enabled
}

// ThrottledWriter wraps an io.Writer and applies bandwidth throttling
type ThrottledWriter struct {
	writer    io.Writer
	throttler *BandwidthThrottler
	ctx       context.Context
}

// NewThrottledWriter creates a new throttled writer
func NewThrottledWriter(ctx context.Context, writer io.Writer, throttler *BandwidthThrottler) *ThrottledWriter {
	return &ThrottledWriter{
		writer:    writer,
		throttler: throttler,
		ctx:       ctx,
	}
}

// Write implements io.Writer with bandwidth throttling
func (tw *ThrottledWriter) Write(p []byte) (n int, err error) {
	// Write the data
	n, err = tw.writer.Write(p)
	if err != nil {
		return n, err
	}

	// Apply throttling after successful write
	if n > 0 {
		throttleErr := tw.throttler.Wait(tw.ctx, int64(n))
		if throttleErr != nil {
			return n, throttleErr
		}
	}

	return n, nil
}

// ThrottledReader wraps an io.Reader and applies bandwidth throttling
type ThrottledReader struct {
	reader    io.Reader
	throttler *BandwidthThrottler
	ctx       context.Context
}

// NewThrottledReader creates a new throttled reader
func NewThrottledReader(ctx context.Context, reader io.Reader, throttler *BandwidthThrottler) *ThrottledReader {
	return &ThrottledReader{
		reader:    reader,
		throttler: throttler,
		ctx:       ctx,
	}
}

// Read implements io.Reader with bandwidth throttling
func (tr *ThrottledReader) Read(p []byte) (n int, err error) {
	// Read the data
	n, err = tr.reader.Read(p)
	if err != nil && err != io.EOF {
		return n, err
	}

	// Apply throttling after successful read
	if n > 0 {
		throttleErr := tr.throttler.Wait(tr.ctx, int64(n))
		if throttleErr != nil {
			return n, throttleErr
		}
	}

	return n, err
}
