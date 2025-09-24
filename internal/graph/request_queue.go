// Package graph provides the basic APIs to interact with Microsoft Graph.
package graph

import (
	"context"
	"io"
	"sync"
	"time"

	"github.com/auriora/onemount/internal/errors"
	"github.com/auriora/onemount/internal/logging"
)

// QueuedRequest represents a request that has been queued for later execution
type QueuedRequest struct {
	ctx      context.Context
	resource string
	auth     *Auth
	method   string
	content  io.Reader
	headers  []Header
	callback func([]byte, error)
}

// RequestQueue manages a queue of requests that are waiting to be executed
// due to rate limiting or other transient errors
type RequestQueue struct {
	queue     []QueuedRequest
	queueLock sync.Mutex
	running   bool
	stopChan  chan struct{}
	wg        sync.WaitGroup
}

var (
	// Global request queue instance
	globalQueue     *RequestQueue
	globalQueueOnce sync.Once
)

// getRequestQueue returns the global request queue instance
func getRequestQueue() *RequestQueue {
	globalQueueOnce.Do(func() {
		globalQueue = &RequestQueue{
			queue:    make([]QueuedRequest, 0),
			stopChan: make(chan struct{}),
		}
		globalQueue.Start()
		logging.Info().Msg("Initialized global request queue for rate-limited operations")
	})
	return globalQueue
}

// Start begins processing queued requests in the background
func (q *RequestQueue) Start() {
	q.queueLock.Lock()
	defer q.queueLock.Unlock()

	if q.running {
		return
	}

	q.running = true
	q.wg.Add(1)
	go q.processQueue()
}

// Stop stops processing queued requests
func (q *RequestQueue) Stop() {
	q.queueLock.Lock()
	defer q.queueLock.Unlock()

	if !q.running {
		return
	}

	q.running = false
	close(q.stopChan)
	q.wg.Wait()
}

// QueueRequest adds a request to the queue for later execution
func (q *RequestQueue) QueueRequest(ctx context.Context, resource string, auth *Auth, method string, content io.Reader, headers []Header, callback func([]byte, error)) {
	q.queueLock.Lock()
	defer q.queueLock.Unlock()

	// Create a new context with a timeout to prevent queued requests from hanging indefinitely
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)

	// Create a wrapper callback that will call the original callback and then cancel the context
	wrappedCallback := func(data []byte, err error) {
		defer cancel()
		callback(data, err)
	}

	q.queue = append(q.queue, QueuedRequest{
		ctx:      ctx,
		resource: resource,
		auth:     auth,
		method:   method,
		content:  content,
		headers:  headers,
		callback: wrappedCallback,
	})

	logging.Info().
		Str("resource", resource).
		Str("method", method).
		Int("queue_length", len(q.queue)).
		Msg("Request queued due to rate limiting")
}

// processQueue processes queued requests with appropriate delays
func (q *RequestQueue) processQueue() {
	defer q.wg.Done()

	// Start with a 1-second delay between requests
	delay := 1 * time.Second

	for {
		// Check if we should stop
		select {
		case <-q.stopChan:
			return
		default:
			// Continue processing
		}

		// Get the next request from the queue
		q.queueLock.Lock()
		if len(q.queue) == 0 {
			q.queueLock.Unlock()
			// No requests in the queue, wait a bit and check again
			time.Sleep(100 * time.Millisecond)
			continue
		}

		request := q.queue[0]
		q.queue = q.queue[1:]
		q.queueLock.Unlock()

		// Check if the context is still valid
		if request.ctx.Err() != nil {
			logging.Warn().
				Str("resource", request.resource).
				Str("method", request.method).
				Err(request.ctx.Err()).
				Msg("Queued request cancelled due to context expiration")
			request.callback(nil, request.ctx.Err())
			continue
		}

		// Wait for the delay before processing the next request
		select {
		case <-time.After(delay):
			// Continue with the request
		case <-q.stopChan:
			// Stop processing
			return
		case <-request.ctx.Done():
			// Request context cancelled
			logging.Warn().
				Str("resource", request.resource).
				Str("method", request.method).
				Err(request.ctx.Err()).
				Msg("Queued request cancelled while waiting for rate limit")
			request.callback(nil, request.ctx.Err())
			continue
		}

		// Execute the request
		logging.Info().
			Str("resource", request.resource).
			Str("method", request.method).
			Msg("Executing queued request after rate limit delay")

		data, err := RequestWithContext(request.ctx, request.resource, request.auth, request.method, request.content, request.headers...)

		// If we got another rate limit error, increase the delay
		if err != nil && errors.IsResourceBusyError(err) {
			delay = time.Duration(float64(delay) * 1.5)
			if delay > 60*time.Second {
				delay = 60 * time.Second
			}
			logging.Warn().
				Str("resource", request.resource).
				Str("method", request.method).
				Dur("next_delay", delay).
				Msg("Rate limit still in effect, increasing delay")
		} else {
			// If the request succeeded or failed with a non-rate-limit error, reset the delay
			delay = 1 * time.Second
		}

		// Call the callback with the result
		request.callback(data, err)
	}
}

// QueueRequestWithCallback queues a request for later execution due to rate limiting
// and calls the provided callback when the request completes
func QueueRequestWithCallback(ctx context.Context, resource string, auth *Auth, method string, content io.Reader, callback func([]byte, error), headers ...Header) {
	queue := getRequestQueue()
	queue.QueueRequest(ctx, resource, auth, method, content, headers, callback)
}

// IsRateLimited checks if the given error indicates that the request was rate limited
func IsRateLimited(err error) bool {
	return errors.IsResourceBusyError(err)
}
