package graph

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestNetworkFeedbackManager_WaitGroupTracking tests that callback goroutines are tracked
func TestUT_Graph_NetworkFeedback_WaitGroupTracking(t *testing.T) {
	manager := NewNetworkFeedbackManager()

	var callbackCount atomic.Int32
	var wg sync.WaitGroup

	// Create a handler that increments a counter
	handler := &testFeedbackHandler{
		onConnected: func() {
			callbackCount.Add(1)
			wg.Done()
		},
		onDisconnected: func() {
			callbackCount.Add(1)
			wg.Done()
		},
		onStatusUpdate: func(connected bool, lastCheck time.Time) {
			callbackCount.Add(1)
			wg.Done()
		},
	}

	manager.AddHandler(handler)

	// Test NotifyConnected
	wg.Add(1)
	manager.NotifyConnected()

	// Test NotifyDisconnected
	wg.Add(1)
	manager.NotifyDisconnected()

	// Test NotifyStatusUpdate
	wg.Add(1)
	manager.NotifyStatusUpdate(true, time.Now())

	// Wait for all callbacks to complete
	wg.Wait()

	// Verify all callbacks were called
	if callbackCount.Load() != 3 {
		t.Errorf("Expected 3 callbacks, got %d", callbackCount.Load())
	}

	// Shutdown should complete immediately since all callbacks are done
	if !manager.Shutdown(1 * time.Second) {
		t.Error("Shutdown timed out when all callbacks should be complete")
	}
}

// TestNetworkFeedbackManager_ShutdownTimeout tests shutdown with active callbacks
func TestUT_Graph_NetworkFeedback_ShutdownTimeout(t *testing.T) {
	manager := NewNetworkFeedbackManager()

	// Create a handler that blocks for a long time
	handler := &testFeedbackHandler{
		onConnected: func() {
			time.Sleep(5 * time.Second) // Block longer than timeout
		},
	}

	manager.AddHandler(handler)

	// Trigger a callback that will block
	manager.NotifyConnected()

	// Give the goroutine time to start
	time.Sleep(100 * time.Millisecond)

	// Shutdown should timeout
	start := time.Now()
	if manager.Shutdown(500 * time.Millisecond) {
		t.Error("Shutdown should have timed out but returned true")
	}
	elapsed := time.Since(start)

	// Verify timeout was respected (with some tolerance)
	if elapsed < 400*time.Millisecond || elapsed > 700*time.Millisecond {
		t.Errorf("Shutdown timeout not respected: elapsed %v", elapsed)
	}
}

// TestNetworkFeedbackManager_ShutdownWithMultipleCallbacks tests shutdown with multiple active callbacks
func TestUT_Graph_NetworkFeedback_ShutdownWithMultipleCallbacks(t *testing.T) {
	manager := NewNetworkFeedbackManager()

	var completedCount atomic.Int32

	// Create handlers that complete at different times
	handler1 := &testFeedbackHandler{
		onConnected: func() {
			time.Sleep(100 * time.Millisecond)
			completedCount.Add(1)
		},
	}
	handler2 := &testFeedbackHandler{
		onConnected: func() {
			time.Sleep(200 * time.Millisecond)
			completedCount.Add(1)
		},
	}
	handler3 := &testFeedbackHandler{
		onConnected: func() {
			time.Sleep(150 * time.Millisecond)
			completedCount.Add(1)
		},
	}

	manager.AddHandler(handler1)
	manager.AddHandler(handler2)
	manager.AddHandler(handler3)

	// Trigger callbacks
	manager.NotifyConnected()

	// Shutdown should wait for all callbacks (longest is 200ms)
	if !manager.Shutdown(500 * time.Millisecond) {
		t.Error("Shutdown timed out when all callbacks should complete within timeout")
	}

	// Verify all callbacks completed
	if completedCount.Load() != 3 {
		t.Errorf("Expected 3 completed callbacks, got %d", completedCount.Load())
	}
}

// TestNetworkFeedbackManager_PanicRecovery tests that panicking handlers don't affect wait group
func TestUT_Graph_NetworkFeedback_PanicRecovery(t *testing.T) {
	manager := NewNetworkFeedbackManager()

	var normalCallbackCalled atomic.Bool

	// Create a handler that panics
	panicHandler := &testFeedbackHandler{
		onConnected: func() {
			panic("test panic")
		},
	}

	// Create a normal handler
	normalHandler := &testFeedbackHandler{
		onConnected: func() {
			normalCallbackCalled.Store(true)
		},
	}

	manager.AddHandler(panicHandler)
	manager.AddHandler(normalHandler)

	// Trigger callbacks
	manager.NotifyConnected()

	// Shutdown should complete despite the panic
	if !manager.Shutdown(1 * time.Second) {
		t.Error("Shutdown timed out - panic may have prevented Done() call")
	}

	// Verify normal callback was still called
	if !normalCallbackCalled.Load() {
		t.Error("Normal callback was not called after panic in another handler")
	}
}

// TestNetworkFeedbackManager_ConcurrentNotifications tests concurrent notifications
func TestUT_Graph_NetworkFeedback_ConcurrentNotifications(t *testing.T) {
	manager := NewNetworkFeedbackManager()

	var callbackCount atomic.Int32

	handler := &testFeedbackHandler{
		onConnected: func() {
			time.Sleep(50 * time.Millisecond)
			callbackCount.Add(1)
		},
		onDisconnected: func() {
			time.Sleep(50 * time.Millisecond)
			callbackCount.Add(1)
		},
		onStatusUpdate: func(connected bool, lastCheck time.Time) {
			time.Sleep(50 * time.Millisecond)
			callbackCount.Add(1)
		},
	}

	manager.AddHandler(handler)

	// Trigger multiple notifications concurrently
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(3)
		go func() {
			defer wg.Done()
			manager.NotifyConnected()
		}()
		go func() {
			defer wg.Done()
			manager.NotifyDisconnected()
		}()
		go func() {
			defer wg.Done()
			manager.NotifyStatusUpdate(true, time.Now())
		}()
	}

	// Wait for all notifications to be triggered
	wg.Wait()

	// Shutdown should wait for all callbacks
	if !manager.Shutdown(5 * time.Second) {
		t.Error("Shutdown timed out with concurrent notifications")
	}

	// Verify all callbacks were called (10 of each type)
	expected := int32(30)
	if callbackCount.Load() != expected {
		t.Errorf("Expected %d callbacks, got %d", expected, callbackCount.Load())
	}
}

// testFeedbackHandler is a test implementation of NetworkFeedbackHandler
type testFeedbackHandler struct {
	onConnected    func()
	onDisconnected func()
	onStatusUpdate func(bool, time.Time)
}

func (h *testFeedbackHandler) OnNetworkConnected() {
	if h.onConnected != nil {
		h.onConnected()
	}
}

func (h *testFeedbackHandler) OnNetworkDisconnected() {
	if h.onDisconnected != nil {
		h.onDisconnected()
	}
}

func (h *testFeedbackHandler) OnNetworkStatusUpdate(connected bool, lastCheck time.Time) {
	if h.onStatusUpdate != nil {
		h.onStatusUpdate(connected, lastCheck)
	}
}
