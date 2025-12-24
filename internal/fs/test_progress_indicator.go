package fs

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// TestProgressIndicator provides real-time feedback during long-running tests
type TestProgressIndicator struct {
	mu          sync.RWMutex
	currentStep string
	startTime   time.Time
	stepTime    time.Time
	steps       []string
	ctx         context.Context
	cancel      context.CancelFunc
	done        chan struct{}
}

// NewTestProgressIndicator creates a new progress indicator
func NewTestProgressIndicator() *TestProgressIndicator {
	ctx, cancel := context.WithCancel(context.Background())
	indicator := &TestProgressIndicator{
		startTime: time.Now(),
		stepTime:  time.Now(),
		ctx:       ctx,
		cancel:    cancel,
		done:      make(chan struct{}),
	}

	// Start the progress ticker
	go indicator.progressTicker()

	return indicator
}

// Step updates the current step and logs progress
func (p *TestProgressIndicator) Step(step string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	now := time.Now()
	if p.currentStep != "" {
		stepDuration := now.Sub(p.stepTime)
		fmt.Printf("✓ Completed: %s (took %v)\n", p.currentStep, stepDuration.Round(time.Millisecond))
	}

	p.currentStep = step
	p.stepTime = now
	p.steps = append(p.steps, step)

	totalDuration := now.Sub(p.startTime)
	fmt.Printf("→ Starting: %s (total elapsed: %v)\n", step, totalDuration.Round(time.Millisecond))
}

// Substep logs a substep without changing the main step
func (p *TestProgressIndicator) Substep(substep string) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	elapsed := time.Since(p.stepTime)
	fmt.Printf("  • %s (step elapsed: %v)\n", substep, elapsed.Round(time.Millisecond))
}

// Heartbeat shows the test is still alive
func (p *TestProgressIndicator) Heartbeat(message string) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	elapsed := time.Since(p.stepTime)
	fmt.Printf("  ❤ %s (step elapsed: %v)\n", message, elapsed.Round(time.Millisecond))
}

// progressTicker shows periodic progress updates
func (p *TestProgressIndicator) progressTicker() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			p.showProgress()
		case <-p.ctx.Done():
			close(p.done)
			return
		}
	}
}

// showProgress displays current progress
func (p *TestProgressIndicator) showProgress() {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.currentStep == "" {
		return
	}

	stepElapsed := time.Since(p.stepTime)
	totalElapsed := time.Since(p.startTime)

	fmt.Printf("⏱ PROGRESS: %s | Step: %v | Total: %v | Steps completed: %d\n",
		p.currentStep,
		stepElapsed.Round(time.Second),
		totalElapsed.Round(time.Second),
		len(p.steps)-1) // -1 because current step isn't completed yet
}

// Complete finishes the progress indicator
func (p *TestProgressIndicator) Complete() {
	p.mu.Lock()
	defer p.mu.Unlock()

	now := time.Now()
	if p.currentStep != "" {
		stepDuration := now.Sub(p.stepTime)
		fmt.Printf("✓ Completed: %s (took %v)\n", p.currentStep, stepDuration.Round(time.Millisecond))
	}

	totalDuration := now.Sub(p.startTime)
	fmt.Printf("🎉 TEST COMPLETED: Total time %v, Steps: %d\n",
		totalDuration.Round(time.Millisecond), len(p.steps))

	p.cancel()
	<-p.done // Wait for ticker to stop
}

// Fail marks the test as failed and shows where it got stuck
func (p *TestProgressIndicator) Fail(reason string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	stepElapsed := time.Since(p.stepTime)
	totalElapsed := time.Since(p.startTime)

	fmt.Printf("❌ TEST FAILED: %s\n", reason)
	fmt.Printf("   Stuck on: %s (for %v)\n", p.currentStep, stepElapsed.Round(time.Second))
	fmt.Printf("   Total time: %v\n", totalElapsed.Round(time.Second))
	fmt.Printf("   Completed steps: %v\n", p.steps[:len(p.steps)-1])

	p.cancel()
	<-p.done // Wait for ticker to stop
}

// GetCurrentStep returns the current step for external monitoring
func (p *TestProgressIndicator) GetCurrentStep() (string, time.Duration) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.currentStep, time.Since(p.stepTime)
}

// IsStuck returns true if the current step has been running too long
func (p *TestProgressIndicator) IsStuck(threshold time.Duration) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return time.Since(p.stepTime) > threshold
}
