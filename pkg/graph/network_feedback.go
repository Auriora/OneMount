package graph

import (
	"sync"
	"time"

	"github.com/auriora/onemount/pkg/logging"
)

// NetworkFeedbackLevel represents the level of feedback to provide to users
type NetworkFeedbackLevel int

const (
	// FeedbackLevelNone provides no user feedback
	FeedbackLevelNone NetworkFeedbackLevel = iota
	// FeedbackLevelBasic provides basic connectivity status
	FeedbackLevelBasic
	// FeedbackLevelDetailed provides detailed network information
	FeedbackLevelDetailed
)

// NetworkFeedbackHandler handles user feedback for network state changes
type NetworkFeedbackHandler interface {
	// OnNetworkConnected is called when network connectivity is restored
	OnNetworkConnected()
	// OnNetworkDisconnected is called when network connectivity is lost
	OnNetworkDisconnected()
	// OnNetworkStatusUpdate provides periodic status updates
	OnNetworkStatusUpdate(connected bool, lastCheck time.Time)
}

// LoggingFeedbackHandler provides feedback through the logging system
type LoggingFeedbackHandler struct {
	level NetworkFeedbackLevel
}

// NewLoggingFeedbackHandler creates a new logging-based feedback handler
func NewLoggingFeedbackHandler(level NetworkFeedbackLevel) *LoggingFeedbackHandler {
	return &LoggingFeedbackHandler{
		level: level,
	}
}

// OnNetworkConnected logs when network connectivity is restored
func (h *LoggingFeedbackHandler) OnNetworkConnected() {
	if h.level >= FeedbackLevelBasic {
		logging.Info().Msg("Network connectivity restored - OneMount is back online")
	}
}

// OnNetworkDisconnected logs when network connectivity is lost
func (h *LoggingFeedbackHandler) OnNetworkDisconnected() {
	if h.level >= FeedbackLevelBasic {
		logging.Warn().Msg("Network connectivity lost - OneMount is now in offline mode")
	}
}

// OnNetworkStatusUpdate provides periodic status updates
func (h *LoggingFeedbackHandler) OnNetworkStatusUpdate(connected bool, lastCheck time.Time) {
	if h.level >= FeedbackLevelDetailed {
		logging.Debug().
			Bool("connected", connected).
			Time("lastCheck", lastCheck).
			Msg("Network status update")
	}
}

// NetworkFeedbackManager manages multiple feedback handlers
type NetworkFeedbackManager struct {
	handlers []NetworkFeedbackHandler
	mutex    sync.RWMutex
}

// NewNetworkFeedbackManager creates a new feedback manager
func NewNetworkFeedbackManager() *NetworkFeedbackManager {
	return &NetworkFeedbackManager{
		handlers: make([]NetworkFeedbackHandler, 0),
	}
}

// AddHandler adds a feedback handler
func (m *NetworkFeedbackManager) AddHandler(handler NetworkFeedbackHandler) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.handlers = append(m.handlers, handler)
}

// RemoveHandler removes a feedback handler
func (m *NetworkFeedbackManager) RemoveHandler(handler NetworkFeedbackHandler) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for i, h := range m.handlers {
		if h == handler {
			m.handlers = append(m.handlers[:i], m.handlers[i+1:]...)
			break
		}
	}
}

// NotifyConnected notifies all handlers that network is connected
func (m *NetworkFeedbackManager) NotifyConnected() {
	m.mutex.RLock()
	handlers := make([]NetworkFeedbackHandler, len(m.handlers))
	copy(handlers, m.handlers)
	m.mutex.RUnlock()

	for _, handler := range handlers {
		go func(h NetworkFeedbackHandler) {
			defer func() {
				if r := recover(); r != nil {
					logging.Error().Interface("panic", r).Msg("Network feedback handler panicked")
				}
			}()
			h.OnNetworkConnected()
		}(handler)
	}
}

// NotifyDisconnected notifies all handlers that network is disconnected
func (m *NetworkFeedbackManager) NotifyDisconnected() {
	m.mutex.RLock()
	handlers := make([]NetworkFeedbackHandler, len(m.handlers))
	copy(handlers, m.handlers)
	m.mutex.RUnlock()

	for _, handler := range handlers {
		go func(h NetworkFeedbackHandler) {
			defer func() {
				if r := recover(); r != nil {
					logging.Error().Interface("panic", r).Msg("Network feedback handler panicked")
				}
			}()
			h.OnNetworkDisconnected()
		}(handler)
	}
}

// NotifyStatusUpdate notifies all handlers of a status update
func (m *NetworkFeedbackManager) NotifyStatusUpdate(connected bool, lastCheck time.Time) {
	m.mutex.RLock()
	handlers := make([]NetworkFeedbackHandler, len(m.handlers))
	copy(handlers, m.handlers)
	m.mutex.RUnlock()

	for _, handler := range handlers {
		go func(h NetworkFeedbackHandler) {
			defer func() {
				if r := recover(); r != nil {
					logging.Error().Interface("panic", r).Msg("Network feedback handler panicked")
				}
			}()
			h.OnNetworkStatusUpdate(connected, lastCheck)
		}(handler)
	}
}

// Enhanced NetworkStateMonitor with feedback support
var (
	globalFeedbackManager *NetworkFeedbackManager
	feedbackManagerOnce   sync.Once
)

// GetGlobalFeedbackManager returns the global feedback manager
func GetGlobalFeedbackManager() *NetworkFeedbackManager {
	feedbackManagerOnce.Do(func() {
		globalFeedbackManager = NewNetworkFeedbackManager()
		// Add default logging handler
		globalFeedbackManager.AddHandler(NewLoggingFeedbackHandler(FeedbackLevelBasic))
	})
	return globalFeedbackManager
}

// Enhanced NetworkStateMonitor that integrates with feedback system
func (nsm *NetworkStateMonitor) checkAndNotifyWithFeedback() {
	currentState := nsm.checker.ForceCheck()

	nsm.mutex.RLock()
	previousState := nsm.lastKnownState
	callbacks := make([]NetworkStateCallback, len(nsm.callbacks))
	copy(callbacks, nsm.callbacks)
	nsm.mutex.RUnlock()

	if currentState != previousState {
		nsm.mutex.Lock()
		nsm.lastKnownState = currentState
		nsm.mutex.Unlock()

		// Notify callbacks
		for _, callback := range callbacks {
			go func(cb NetworkStateCallback) {
				defer func() {
					if r := recover(); r != nil {
						logging.Error().Interface("panic", r).Msg("Network state callback panicked")
					}
				}()
				cb(currentState, previousState)
			}(callback)
		}

		// Notify feedback manager
		feedbackManager := GetGlobalFeedbackManager()
		if currentState {
			feedbackManager.NotifyConnected()
		} else {
			feedbackManager.NotifyDisconnected()
		}

		// Log the state change
		if currentState {
			logging.Info().Msg("Network connectivity restored")
		} else {
			logging.Warn().Msg("Network connectivity lost")
		}
	}

	// Always provide status updates for detailed feedback
	feedbackManager := GetGlobalFeedbackManager()
	feedbackManager.NotifyStatusUpdate(currentState, time.Now())
}
