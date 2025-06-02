package fs

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/auriora/onemount/pkg/graph"
	"github.com/auriora/onemount/pkg/logging"
)

// Error definitions for metadata request manager
var (
	ErrQueueFull          = errors.New("metadata request queue is full")
	ErrInvalidRequestType = errors.New("invalid metadata request type")
	ErrManagerNotStarted  = errors.New("metadata request manager not started")
)

// MetadataPriority represents the priority level for metadata requests
type MetadataPriority int

const (
	// PriorityBackground for background sync operations
	PriorityBackground MetadataPriority = iota
	// PriorityForeground for user-initiated operations (file access, directory listing)
	PriorityForeground
)

// MetadataRequest represents a queued metadata request
type MetadataRequest struct {
	ID       string
	Priority MetadataPriority
	Type     string // "children", "item", "path"
	Auth     *graph.Auth
	Callback func([]*graph.DriveItem, error)
	Context  context.Context
	Path     string // For path-based requests
}

// MetadataRequestManager manages prioritized metadata requests
type MetadataRequestManager struct {
	highPriorityQueue chan *MetadataRequest
	lowPriorityQueue  chan *MetadataRequest
	workers           int
	stopChan          chan struct{}
	wg                sync.WaitGroup
	fs                *Filesystem
}

// NewMetadataRequestManager creates a new metadata request manager
func NewMetadataRequestManager(fs *Filesystem, workers int) *MetadataRequestManager {
	return &MetadataRequestManager{
		highPriorityQueue: make(chan *MetadataRequest, 100),  // Buffer for foreground requests
		lowPriorityQueue:  make(chan *MetadataRequest, 1000), // Larger buffer for background requests
		workers:           workers,
		stopChan:          make(chan struct{}),
		fs:                fs,
	}
}

// Start begins processing metadata requests with the specified number of workers
func (m *MetadataRequestManager) Start() {
	logging.Info().Int("workers", m.workers).Msg("Starting metadata request manager")

	for i := 0; i < m.workers; i++ {
		m.wg.Add(1)
		go m.worker(i)
	}
}

// Stop gracefully stops the metadata request manager
func (m *MetadataRequestManager) Stop() {
	logging.Info().Msg("Stopping metadata request manager")
	close(m.stopChan)
	m.wg.Wait()
	logging.Info().Msg("Metadata request manager stopped")
}

// QueueChildrenRequest queues a request to fetch children of a directory
func (m *MetadataRequestManager) QueueChildrenRequest(id string, auth *graph.Auth, priority MetadataPriority, callback func([]*graph.DriveItem, error)) error {
	ctx := context.Background()
	if priority == PriorityForeground {
		// Add timeout for foreground requests to prevent hanging
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
	}

	request := &MetadataRequest{
		ID:       id,
		Priority: priority,
		Type:     "children",
		Auth:     auth,
		Callback: callback,
		Context:  ctx,
	}

	return m.queueRequest(request)
}

// QueueItemRequest queues a request to fetch a single item
func (m *MetadataRequestManager) QueueItemRequest(id string, auth *graph.Auth, priority MetadataPriority, callback func([]*graph.DriveItem, error)) error {
	ctx := context.Background()
	if priority == PriorityForeground {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
	}

	request := &MetadataRequest{
		ID:       id,
		Priority: priority,
		Type:     "item",
		Auth:     auth,
		Callback: callback,
		Context:  ctx,
	}

	return m.queueRequest(request)
}

// QueuePathRequest queues a request to fetch children by path
func (m *MetadataRequestManager) QueuePathRequest(path string, auth *graph.Auth, priority MetadataPriority, callback func([]*graph.DriveItem, error)) error {
	ctx := context.Background()
	if priority == PriorityForeground {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
	}

	request := &MetadataRequest{
		Path:     path,
		Priority: priority,
		Type:     "path",
		Auth:     auth,
		Callback: callback,
		Context:  ctx,
	}

	return m.queueRequest(request)
}

// queueRequest adds a request to the appropriate priority queue
func (m *MetadataRequestManager) queueRequest(request *MetadataRequest) error {
	var targetQueue chan *MetadataRequest
	var queueName string

	if request.Priority == PriorityForeground {
		targetQueue = m.highPriorityQueue
		queueName = "high"
	} else {
		targetQueue = m.lowPriorityQueue
		queueName = "low"
	}

	select {
	case targetQueue <- request:
		logging.Debug().
			Str("type", request.Type).
			Str("id", request.ID).
			Str("path", request.Path).
			Str("priority", queueName).
			Msg("Metadata request queued")
		return nil
	default:
		logging.Warn().
			Str("type", request.Type).
			Str("id", request.ID).
			Str("path", request.Path).
			Str("priority", queueName).
			Msg("Metadata request queue full, dropping request")
		return ErrQueueFull
	}
}

// worker processes metadata requests from the priority queues
func (m *MetadataRequestManager) worker(workerID int) {
	defer m.wg.Done()

	logging.Debug().Int("workerID", workerID).Msg("Metadata request worker started")

	for {
		select {
		case <-m.stopChan:
			logging.Debug().Int("workerID", workerID).Msg("Metadata request worker stopping")
			return

		case request := <-m.highPriorityQueue:
			// Process high priority requests immediately
			m.processRequest(workerID, request, "high")

		case request := <-m.lowPriorityQueue:
			// Process low priority requests only if no high priority requests are waiting
			select {
			case highPriorityRequest := <-m.highPriorityQueue:
				// High priority request arrived, process it first
				m.processRequest(workerID, highPriorityRequest, "high")
				// Put the low priority request back in the queue
				select {
				case m.lowPriorityQueue <- request:
				default:
					// Queue full, drop the request
					logging.Warn().Int("workerID", workerID).Msg("Low priority queue full, dropping request")
					request.Callback(nil, ErrQueueFull)
				}
			default:
				// No high priority requests, process the low priority request
				m.processRequest(workerID, request, "low")
			}

		default:
			// No requests available, wait a bit to avoid busy waiting
			time.Sleep(10 * time.Millisecond)
		}
	}
}

// processRequest executes a metadata request
func (m *MetadataRequestManager) processRequest(workerID int, request *MetadataRequest, priorityName string) {
	startTime := time.Now()

	logging.Debug().
		Int("workerID", workerID).
		Str("type", request.Type).
		Str("id", request.ID).
		Str("path", request.Path).
		Str("priority", priorityName).
		Msg("Processing metadata request")

	var result []*graph.DriveItem
	var err error

	switch request.Type {
	case "children":
		result, err = graph.GetItemChildren(request.ID, request.Auth)
	case "item":
		item, itemErr := graph.GetItem(request.ID, request.Auth)
		if itemErr != nil {
			err = itemErr
		} else {
			result = []*graph.DriveItem{item}
		}
	case "path":
		result, err = graph.GetItemChildrenPath(request.Path, request.Auth)
	default:
		err = ErrInvalidRequestType
	}

	duration := time.Since(startTime)

	if err != nil {
		logging.Debug().
			Int("workerID", workerID).
			Str("type", request.Type).
			Str("id", request.ID).
			Str("path", request.Path).
			Str("priority", priorityName).
			Dur("duration", duration).
			Err(err).
			Msg("Metadata request failed")
	} else {
		logging.Debug().
			Int("workerID", workerID).
			Str("type", request.Type).
			Str("id", request.ID).
			Str("path", request.Path).
			Str("priority", priorityName).
			Dur("duration", duration).
			Int("resultCount", len(result)).
			Msg("Metadata request completed")
	}

	// Call the callback with the result
	request.Callback(result, err)
}

// GetQueueStats returns statistics about the request queues
func (m *MetadataRequestManager) GetQueueStats() (highPriorityCount, lowPriorityCount int) {
	return len(m.highPriorityQueue), len(m.lowPriorityQueue)
}
