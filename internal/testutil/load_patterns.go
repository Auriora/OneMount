// Package testutil provides testing utilities for the OneMount project.
package testutil

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"
)

// LoadPatternType defines the type of load pattern to apply
type LoadPatternType string

const (
	// ConstantLoad applies a constant number of concurrent operations
	ConstantLoad LoadPatternType = "constant"
	// RampUpLoad gradually increases the number of concurrent operations
	RampUpLoad LoadPatternType = "ramp-up"
	// SpikeLoad applies a sudden spike in concurrent operations
	SpikeLoad LoadPatternType = "spike"
	// WaveLoad applies a sinusoidal pattern of concurrent operations
	WaveLoad LoadPatternType = "wave"
	// StepLoad increases the number of concurrent operations in steps
	StepLoad LoadPatternType = "step"
)

// LoadPattern defines a pattern for applying load during a test
type LoadPattern struct {
	// Type of load pattern
	Type LoadPatternType
	// Base concurrency level
	BaseConcurrency int
	// Peak concurrency level (for non-constant patterns)
	PeakConcurrency int
	// Duration of the pattern
	Duration time.Duration
	// Additional parameters for specific patterns
	Params map[string]interface{}
}

// LoadPatternGenerator generates a load pattern over time
type LoadPatternGenerator interface {
	// GetConcurrency returns the target concurrency at the given time
	GetConcurrency(elapsed time.Duration) int
	// GetDuration returns the total duration of the pattern
	GetDuration() time.Duration
}

// ConstantLoadGenerator generates a constant load pattern
type ConstantLoadGenerator struct {
	concurrency int
	duration    time.Duration
}

// NewConstantLoadGenerator creates a new constant load generator
func NewConstantLoadGenerator(concurrency int, duration time.Duration) *ConstantLoadGenerator {
	return &ConstantLoadGenerator{
		concurrency: concurrency,
		duration:    duration,
	}
}

// GetConcurrency returns the target concurrency at the given time
func (g *ConstantLoadGenerator) GetConcurrency(elapsed time.Duration) int {
	if elapsed >= g.duration {
		return 0
	}
	return g.concurrency
}

// GetDuration returns the total duration of the pattern
func (g *ConstantLoadGenerator) GetDuration() time.Duration {
	return g.duration
}

// RampUpLoadGenerator generates a ramp-up load pattern
type RampUpLoadGenerator struct {
	baseConcurrency int
	peakConcurrency int
	duration        time.Duration
}

// NewRampUpLoadGenerator creates a new ramp-up load generator
func NewRampUpLoadGenerator(baseConcurrency, peakConcurrency int, duration time.Duration) *RampUpLoadGenerator {
	return &RampUpLoadGenerator{
		baseConcurrency: baseConcurrency,
		peakConcurrency: peakConcurrency,
		duration:        duration,
	}
}

// GetConcurrency returns the target concurrency at the given time
func (g *RampUpLoadGenerator) GetConcurrency(elapsed time.Duration) int {
	if elapsed >= g.duration {
		return 0
	}

	// Calculate the progress (0.0 to 1.0)
	progress := float64(elapsed) / float64(g.duration)

	// Calculate the target concurrency based on linear interpolation
	concurrencyRange := g.peakConcurrency - g.baseConcurrency
	targetConcurrency := g.baseConcurrency + int(float64(concurrencyRange)*progress)

	return targetConcurrency
}

// GetDuration returns the total duration of the pattern
func (g *RampUpLoadGenerator) GetDuration() time.Duration {
	return g.duration
}

// SpikeLoadGenerator generates a spike load pattern
type SpikeLoadGenerator struct {
	baseConcurrency int
	peakConcurrency int
	duration        time.Duration
	spikeStart      time.Duration
	spikeDuration   time.Duration
}

// NewSpikeLoadGenerator creates a new spike load generator
func NewSpikeLoadGenerator(baseConcurrency, peakConcurrency int, duration, spikeStart, spikeDuration time.Duration) *SpikeLoadGenerator {
	return &SpikeLoadGenerator{
		baseConcurrency: baseConcurrency,
		peakConcurrency: peakConcurrency,
		duration:        duration,
		spikeStart:      spikeStart,
		spikeDuration:   spikeDuration,
	}
}

// GetConcurrency returns the target concurrency at the given time
func (g *SpikeLoadGenerator) GetConcurrency(elapsed time.Duration) int {
	if elapsed >= g.duration {
		return 0
	}

	// Check if we're in the spike period
	if elapsed >= g.spikeStart && elapsed < g.spikeStart+g.spikeDuration {
		return g.peakConcurrency
	}

	return g.baseConcurrency
}

// GetDuration returns the total duration of the pattern
func (g *SpikeLoadGenerator) GetDuration() time.Duration {
	return g.duration
}

// WaveLoadGenerator generates a sinusoidal wave load pattern
type WaveLoadGenerator struct {
	baseConcurrency int
	peakConcurrency int
	duration        time.Duration
	frequency       float64 // Number of complete waves during the test
}

// NewWaveLoadGenerator creates a new wave load generator
func NewWaveLoadGenerator(baseConcurrency, peakConcurrency int, duration time.Duration, frequency float64) *WaveLoadGenerator {
	return &WaveLoadGenerator{
		baseConcurrency: baseConcurrency,
		peakConcurrency: peakConcurrency,
		duration:        duration,
		frequency:       frequency,
	}
}

// GetConcurrency returns the target concurrency at the given time
func (g *WaveLoadGenerator) GetConcurrency(elapsed time.Duration) int {
	if elapsed >= g.duration {
		return 0
	}

	// Calculate the progress (0.0 to 1.0)
	progress := float64(elapsed) / float64(g.duration)

	// Calculate the wave position (0.0 to 1.0)
	wave := 0.5 + 0.5*math.Sin(2*math.Pi*g.frequency*progress)

	// Calculate the target concurrency based on the wave
	concurrencyRange := g.peakConcurrency - g.baseConcurrency
	targetConcurrency := g.baseConcurrency + int(float64(concurrencyRange)*wave)

	return targetConcurrency
}

// GetDuration returns the total duration of the pattern
func (g *WaveLoadGenerator) GetDuration() time.Duration {
	return g.duration
}

// StepLoadGenerator generates a step load pattern
type StepLoadGenerator struct {
	baseConcurrency int
	peakConcurrency int
	duration        time.Duration
	steps           int
}

// NewStepLoadGenerator creates a new step load generator
func NewStepLoadGenerator(baseConcurrency, peakConcurrency int, duration time.Duration, steps int) *StepLoadGenerator {
	return &StepLoadGenerator{
		baseConcurrency: baseConcurrency,
		peakConcurrency: peakConcurrency,
		duration:        duration,
		steps:           steps,
	}
}

// GetConcurrency returns the target concurrency at the given time
func (g *StepLoadGenerator) GetConcurrency(elapsed time.Duration) int {
	if elapsed >= g.duration {
		return 0
	}

	// Calculate the progress (0.0 to 1.0)
	progress := float64(elapsed) / float64(g.duration)

	// Calculate the current step (0 to steps-1)
	step := int(progress * float64(g.steps))
	if step >= g.steps {
		step = g.steps - 1
	}

	// Calculate the target concurrency based on the step
	concurrencyRange := g.peakConcurrency - g.baseConcurrency
	stepSize := float64(concurrencyRange) / float64(g.steps)
	targetConcurrency := g.baseConcurrency + int(stepSize*float64(step+1))

	return targetConcurrency
}

// GetDuration returns the total duration of the pattern
func (g *StepLoadGenerator) GetDuration() time.Duration {
	return g.duration
}

// CreateLoadPatternGenerator creates a load pattern generator based on the pattern type
func CreateLoadPatternGenerator(pattern LoadPattern) (LoadPatternGenerator, error) {
	switch pattern.Type {
	case ConstantLoad:
		return NewConstantLoadGenerator(pattern.BaseConcurrency, pattern.Duration), nil

	case RampUpLoad:
		return NewRampUpLoadGenerator(pattern.BaseConcurrency, pattern.PeakConcurrency, pattern.Duration), nil

	case SpikeLoad:
		spikeStart, ok := pattern.Params["spikeStart"].(time.Duration)
		if !ok {
			spikeStart = pattern.Duration / 2
		}

		spikeDuration, ok := pattern.Params["spikeDuration"].(time.Duration)
		if !ok {
			spikeDuration = pattern.Duration / 10
		}

		return NewSpikeLoadGenerator(pattern.BaseConcurrency, pattern.PeakConcurrency, pattern.Duration, spikeStart, spikeDuration), nil

	case WaveLoad:
		frequency, ok := pattern.Params["frequency"].(float64)
		if !ok {
			frequency = 3.0
		}

		return NewWaveLoadGenerator(pattern.BaseConcurrency, pattern.PeakConcurrency, pattern.Duration, frequency), nil

	case StepLoad:
		steps, ok := pattern.Params["steps"].(int)
		if !ok {
			steps = 5
		}

		return NewStepLoadGenerator(pattern.BaseConcurrency, pattern.PeakConcurrency, pattern.Duration, steps), nil

	default:
		return nil, fmt.Errorf("unknown load pattern type: %s", pattern.Type)
	}
}

// RunLoadPattern runs a load test with the specified pattern
func RunLoadPattern(ctx context.Context, pattern LoadPatternGenerator, scenario func(ctx context.Context) error) ([]time.Duration, []error) {
	// Create channels for results
	latencyChan := make(chan time.Duration, 10000)
	errorChan := make(chan error, 10000)

	// Create a context with timeout that respects the parent context
	patternCtx, patternCancel := context.WithTimeout(ctx, pattern.GetDuration())
	defer patternCancel()

	// Create channels to control the number of active workers
	workerStartChan := make(chan struct{}, 100) // Buffer to prevent blocking
	workerDoneChan := make(chan struct{}, 100)  // Buffer to prevent blocking
	workerPool := make(chan int, 100)           // Worker pool to limit concurrent goroutines

	// Create a wait group for worker goroutines
	var workerWg sync.WaitGroup

	// Start the controller goroutine
	controllerDone := make(chan struct{})
	go func() {
		defer close(controllerDone)
		defer close(workerStartChan) // Signal workers to exit when controller is done

		startTime := time.Now()
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		var currentWorkers int

		for {
			select {
			case <-patternCtx.Done():
				// Context was canceled, exit the controller
				return
			case <-ticker.C:
				elapsed := time.Since(startTime)
				targetWorkers := pattern.GetConcurrency(elapsed)

				// Add workers if needed
				for i := currentWorkers; i < targetWorkers; i++ {
					select {
					case workerStartChan <- struct{}{}:
						currentWorkers++
					case <-patternCtx.Done():
						return
					default:
						// If the channel is full, try again next tick
						continue
					}
				}

				// Remove workers if needed
				for i := currentWorkers; i > targetWorkers; i-- {
					select {
					case <-workerDoneChan:
						currentWorkers--
					case <-patternCtx.Done():
						return
					case <-time.After(50 * time.Millisecond):
						// Timeout - don't wait indefinitely for workers to complete
						// This prevents the test from hanging if workers are stuck
						fmt.Printf("Warning: Timeout waiting for worker to complete\n")
						currentWorkers--
					}
				}

				if elapsed >= pattern.GetDuration() {
					return
				}
			case <-workerDoneChan:
				// Handle workers that complete on their own
				if currentWorkers > 0 {
					currentWorkers--
				}
			}
		}
	}()

	// Initialize the worker pool
	maxWorkers := 100 // Reduce the maximum number of goroutines to prevent resource contention
	for i := 0; i < maxWorkers; i++ {
		workerPool <- i
	}

	// Start a goroutine to manage the worker pool
	workerPoolDone := make(chan struct{})
	go func() {
		defer close(workerPoolDone)

		for {
			select {
			case <-patternCtx.Done():
				return
			case workerId, ok := <-workerPool:
				if !ok {
					return
				}

				// Start a worker goroutine
				workerWg.Add(1)
				go func(id int) {
					defer workerWg.Done()
					defer func() {
						// Return the worker ID to the pool when done
						select {
						case workerPool <- id:
						case <-patternCtx.Done():
							// If context is done, don't try to return to the pool
						default:
							// If the channel is closed or full, don't try to send
						}
					}()

					// Wait for work
					select {
					case <-patternCtx.Done():
						return
					case _, ok := <-workerStartChan:
						if !ok {
							return
						}

						// Run the scenario
						start := time.Now()
						err := scenario(patternCtx)
						latency := time.Since(start)

						// Record results
						select {
						case latencyChan <- latency:
						case <-patternCtx.Done():
							return
						}

						if err != nil {
							select {
							case errorChan <- err:
							case <-patternCtx.Done():
								return
							}
						}

						// Signal that the worker is done
						select {
						case workerDoneChan <- struct{}{}:
						case <-patternCtx.Done():
							return
						default:
							// If the channel is full, just continue
							// The controller will adjust the count on the next tick
						}
					}
				}(workerId)
			}
		}
	}()

	// Wait for the pattern context to be done (either timeout or parent context canceled)
	<-patternCtx.Done()

	// Wait for all workers to finish before closing the worker pool
	// This prevents the "send on closed channel" panic
	workerWg.Wait()

	// Close the worker pool to stop creating new workers
	close(workerPool)

	// Wait for the worker pool manager to finish
	<-workerPoolDone

	// Wait for the controller to finish
	<-controllerDone

	// Close result channels
	close(latencyChan)
	close(errorChan)

	// Collect results
	var latencies []time.Duration
	var errors []error

	for latency := range latencyChan {
		latencies = append(latencies, latency)
	}

	for err := range errorChan {
		errors = append(errors, err)
	}

	return latencies, errors
}
