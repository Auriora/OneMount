package fs

import (
	"context"
	"errors"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/auriora/onemount/internal/graph"
	"github.com/auriora/onemount/internal/logging"
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
	queuedAt time.Time
}

// MetadataQueueStats summarizes queue depth and latency for telemetry.
type MetadataQueueStats struct {
	HighDepth int
	LowDepth  int
	AvgWaitMs float64
}

// MetadataRequestManager manages prioritized metadata requests
type MetadataRequestManager struct {
	highPriorityQueue chan *MetadataRequest
	lowPriorityQueue  chan *MetadataRequest
	workers           int
	foregroundWorkers int
	stopChan          chan struct{}
	wg                sync.WaitGroup
	fs                *Filesystem

	inFlightMu sync.Mutex
	inFlight   map[string]*inFlightEntry

	waitTotalNs atomic.Int64
	waitCount   atomic.Int64
}

type inFlightEntry struct {
	callbacks []func([]*graph.DriveItem, error)
	priority  MetadataPriority
}

// NewMetadataRequestManager creates a new metadata request manager
func NewMetadataRequestManager(fs *Filesystem, workers, highQueueSize, lowQueueSize int) *MetadataRequestManager {
	foregroundWorkers := 0
	if workers >= 2 {
		foregroundWorkers = 1
	}
	return &MetadataRequestManager{
		highPriorityQueue: make(chan *MetadataRequest, highQueueSize), // Buffer for foreground requests
		lowPriorityQueue:  make(chan *MetadataRequest, lowQueueSize),  // Larger buffer for background requests
		workers:           workers,
		foregroundWorkers: foregroundWorkers,
		stopChan:          make(chan struct{}),
		fs:                fs,
		inFlight:          make(map[string]*inFlightEntry),
	}
}

// Start begins processing metadata requests with the specified number of workers
func (m *MetadataRequestManager) Start() {
	logging.Info().Int("workers", m.workers).Msg("Starting metadata request manager")

	for i := 0; i < m.foregroundWorkers; i++ {
		m.wg.Add(1)
		go m.foregroundWorker(i)
	}

	for i := m.foregroundWorkers; i < m.workers; i++ {
		m.wg.Add(1)
		go m.worker(i)
	}
}

// Stop gracefully stops the metadata request manager
func (m *MetadataRequestManager) Stop() {
	logging.Info().Msg("Stopping metadata request manager")
	close(m.stopChan)
	m.wg.Wait()
	m.failAllInFlight(ErrManagerNotStarted)
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
	key := request.cacheKey()
	if key != "" {
		if m.joinInFlight(key, request.Callback, request.Priority) {
			logging.Debug().
				Str("type", request.Type).
				Str("id", request.ID).
				Str("path", request.Path).
				Msg("Joined in-flight metadata request")
			return nil
		}
		request.Callback = m.dispatchInFlightCallback(key)
	}

	var targetQueue chan *MetadataRequest
	var queueName string
	request.queuedAt = time.Now()

	if request.Priority == PriorityForeground {
		if m.fs != nil {
			m.fs.RecordForegroundActivity()
		}
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
		if key != "" {
			m.dispatchInFlight(key, nil, ErrQueueFull)
		}
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

func (m *MetadataRequestManager) foregroundWorker(workerID int) {
	defer m.wg.Done()
	logging.Debug().Int("workerID", workerID).Msg("Foreground metadata request worker started")
	for {
		select {
		case <-m.stopChan:
			logging.Debug().Int("workerID", workerID).Msg("Foreground metadata request worker stopping")
			return
		case request := <-m.highPriorityQueue:
			m.processRequest(workerID, request, "high")
		case request := <-m.lowPriorityQueue:
			// Only help with low-priority work when high queue is empty.
			m.processRequest(workerID, request, "low-steal")
		default:
			time.Sleep(5 * time.Millisecond)
		}
	}
}

// processRequest executes a metadata request
func (m *MetadataRequestManager) processRequest(workerID int, request *MetadataRequest, priorityName string) {
	startTime := time.Now()
	var waitDur time.Duration
	if !request.queuedAt.IsZero() {
		waitDur = startTime.Sub(request.queuedAt)
		m.waitTotalNs.Add(waitDur.Nanoseconds())
		m.waitCount.Add(1)
	}

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

// Snapshot returns lightweight queue telemetry for stats surfaces.
func (m *MetadataRequestManager) Snapshot() MetadataQueueStats {
	stats := MetadataQueueStats{}
	if m == nil {
		return stats
	}
	stats.HighDepth = len(m.highPriorityQueue)
	stats.LowDepth = len(m.lowPriorityQueue)
	count := m.waitCount.Load()
	if count > 0 {
		total := m.waitTotalNs.Load()
		stats.AvgWaitMs = float64(total) / float64(count) / float64(time.Millisecond)
	}
	return stats
}

func (r *MetadataRequest) cacheKey() string {
	switch r.Type {
	case "children":
		if r.ID == "" {
			return ""
		}
		return "children:" + r.ID
	case "item":
		if r.ID == "" {
			return ""
		}
		return "item:" + r.ID
	case "path":
		if r.Path == "" {
			return ""
		}
		return "path:" + strings.ToLower(r.Path)
	default:
		return ""
	}
}

func (m *MetadataRequestManager) joinInFlight(key string, cb func([]*graph.DriveItem, error), priority MetadataPriority) bool {
	if key == "" {
		return false
	}
	m.inFlightMu.Lock()
	defer m.inFlightMu.Unlock()
	if entry, ok := m.inFlight[key]; ok {
		entry.callbacks = append(entry.callbacks, cb)
		if priority == PriorityForeground && entry.priority == PriorityBackground {
			entry.priority = PriorityForeground
		}
		return true
	}
	m.inFlight[key] = &inFlightEntry{
		callbacks: []func([]*graph.DriveItem, error){cb},
		priority:  priority,
	}
	return false
}

func (m *MetadataRequestManager) dispatchInFlightCallback(key string) func([]*graph.DriveItem, error) {
	return func(items []*graph.DriveItem, err error) {
		m.dispatchInFlight(key, items, err)
	}
}

func (m *MetadataRequestManager) dispatchInFlight(key string, items []*graph.DriveItem, err error) {
	if key == "" {
		return
	}
	m.inFlightMu.Lock()
	entry, ok := m.inFlight[key]
	if ok {
		delete(m.inFlight, key)
	}
	m.inFlightMu.Unlock()
	if !ok || entry == nil {
		return
	}
	for _, cb := range entry.callbacks {
		cb(items, err)
	}
}

func (m *MetadataRequestManager) failAllInFlight(err error) {
	m.inFlightMu.Lock()
	entries := m.inFlight
	m.inFlight = make(map[string]*inFlightEntry)
	m.inFlightMu.Unlock()
	for _, entry := range entries {
		for _, cb := range entry.callbacks {
			cb(nil, err)
		}
	}
}
