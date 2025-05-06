package fs

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/auriora/onemount/internal/fs/graph"
	"github.com/rs/zerolog/log"
	"github.com/yousong/socketio-go/engineio"
	"github.com/yousong/socketio-go/socketio"
)

type subscriptionResponse struct {
	Context            string    `json:"@odata.context"`
	ClientState        string    `json:"clientState"`
	ExpirationDateTime time.Time `json:"expirationDateTime"`
	Id                 string    `json:"id"`
	NotificationUrl    string    `json:"notificationUrl"`
	Resource           string    `json:"resource"`
}

func (f *Filesystem) subscribeChanges() (subscriptionResponse, error) {
	subscResp := subscriptionResponse{}

	resp, err := graph.Get(f.subscribeChangesLink, f.auth)
	if err != nil {
		return subscResp, err
	}
	if err := json.Unmarshal(resp, &subscResp); err != nil {
		return subscResp, err
	}
	return subscResp, nil
}

type subscribeFunc func() (subscriptionResponse, error)
type subscription struct {
	C <-chan struct{}

	subscribe subscribeFunc
	c         chan struct{}
	closeCh   chan struct{}
	sioErrCh  chan error
}

func newSubscription(subscribe subscribeFunc) *subscription {
	s := &subscription{
		subscribe: subscribe,
		c:         make(chan struct{}),
		closeCh:   make(chan struct{}),
		sioErrCh:  make(chan error),
	}
	s.C = s.c
	return s
}

func (s *subscription) Start() {
	const (
		errRetryInterval      = 10 * time.Second
		setupEventChanTimeout = 10 * time.Second
	)
	triggerOnErrCh := make(chan struct{}, 1)
	triggerOnErr := func() {
		select {
		case triggerOnErrCh <- struct{}{}:
		default:
		}
	}
	go func() {
		tick := time.NewTicker(30 * time.Second)
		defer tick.Stop()

		for {
			select {
			case <-tick.C:
			case <-s.closeCh:
				return
			}
			select {
			case <-triggerOnErrCh:
				s.trigger()
			default:
			}
		}
	}()

	for {
		resp, err := s.subscribe()
		if err != nil {
			log.Error().Err(err).Msg("make subscription")
			triggerOnErr()
			time.Sleep(errRetryInterval)
			continue
		}
		nextDur := resp.ExpirationDateTime.Sub(time.Now())
		ctx := context.Background()
		ctx, cancel := context.WithTimeout(ctx, setupEventChanTimeout)
		cleanup, err := s.setupEventChan(ctx, resp.NotificationUrl)
		if err != nil {
			cancel() // Cancel the context to prevent leaks
			log.Error().Err(err).Msg("subscription chan setup")
			triggerOnErr()
			time.Sleep(errRetryInterval)
			continue
		}
		cancel() // Context no longer needed after setupEventChan completes
		// Trigger once so subscribers can pick up deltas ocurred
		// between expiration of last subscription and start of this
		// subscription
		s.trigger()
		if bye := func() bool {
			defer cleanup()
			select {
			case <-time.After(nextDur):
			case err := <-s.sioErrCh:
				log.Warn().Err(err).Msg("socketio session error")
			case <-s.closeCh:
				return true
			}
			return false
		}(); bye {
			return
		}
	}
}

func (s *subscription) setupEventChan(ctx context.Context, urlstr string) (func(), error) {
	u, err := url.Parse(urlstr)
	if err != nil {
		return nil, err
	}

	// Create a ready channel to synchronize connection establishment
	readyCh := make(chan struct{})
	errCh := make(chan error, 1)

	// Create a mutex to protect access to the socketio connection
	// This helps prevent race conditions when multiple goroutines access the connection
	var siocMutex sync.Mutex
	var sioc *socketio.Conn

	// Create a connection ready handler that's thread-safe
	connectionReady := func() {
		// Only close the channel once
		select {
		case <-readyCh:
			// Already closed
		default:
			close(readyCh)
		}
	}

	// Create a thread-safe wrapper for the socketio connection
	// This ensures all access to the connection is properly synchronized
	// and prevents race conditions in the socketio-go library
	type threadSafeSocketIO struct {
		mu   sync.Mutex
		conn *socketio.Conn
	}

	safeConn := &threadSafeSocketIO{}

	// Create the socketio connection with the ready handler in a thread-safe manner
	siocMutex.Lock()
	sioc, err = socketio.DialContext(ctx, socketio.Config{
		URL:        urlstr,
		EIOVersion: engineio.EIO3,
		OnError: func(err error) {
			// Thread-safe error handling
			safeConn.mu.Lock()
			defer safeConn.mu.Unlock()
			s.socketioOnError(err)
		},
	})
	if err == nil {
		safeConn.conn = sioc
	}
	siocMutex.Unlock()

	if err != nil {
		return nil, err
	}

	// Create a namespace with the notification handler
	ns := &socketio.Namespace{
		Name: u.RequestURI(),
		PacketHandlers: map[byte]socketio.Handler{
			socketio.PacketTypeEVENT: func(msg socketio.Message) {
				// Thread-safe notification handling
				safeConn.mu.Lock()
				defer safeConn.mu.Unlock()
				s.notificationHandler(msg)
			},
		},
	}

	// Connect to the namespace in a separate goroutine to avoid blocking
	go func() {
		safeConn.mu.Lock()
		localConn := safeConn.conn // Create a local reference to avoid race conditions
		safeConn.mu.Unlock()

		if localConn == nil {
			errCh <- fmt.Errorf("socketio connection is nil")
			return
		}

		if err := localConn.Connect(ctx, ns); err != nil {
			errCh <- err
			return
		}

		// Signal that the connection is ready
		connectionReady()
	}()

	// Wait for the connection to be ready or for an error
	select {
	case <-readyCh:
		// Connection is ready
		log.Debug().Msg("socketio connection established successfully")
	case err := <-errCh:
		// Connection failed
		safeConn.mu.Lock()
		if safeConn.conn != nil {
			safeConn.conn.Close()
			safeConn.conn = nil
		}
		safeConn.mu.Unlock()
		return nil, err
	case <-ctx.Done():
		// Context timeout or cancellation
		safeConn.mu.Lock()
		if safeConn.conn != nil {
			safeConn.conn.Close()
			safeConn.conn = nil
		}
		safeConn.mu.Unlock()
		return nil, ctx.Err()
	}

	// Return a thread-safe cleanup function
	return func() {
		safeConn.mu.Lock()
		defer safeConn.mu.Unlock()
		if safeConn.conn != nil {
			safeConn.conn.Close()
			safeConn.conn = nil
		}
	}, nil
}

func (s *subscription) notificationHandler(msg socketio.Message) {
	var evt []string
	if err := json.Unmarshal(msg.DataRaw, &evt); err != nil {
		log.Warn().Err(err).Msg("unmarshal socketio event")
		return
	}
	if len(evt) < 2 || evt[0] != "notification" {
		log.Warn().Int("len", len(evt)).Str("type", evt[0]).Msg("check event type")
		return
	}
	var n struct {
		ClientState                    string `json:"clientState"`
		SubscriptionId                 string `json:"subscriptionId"`
		SubscriptionExpirationDateTime string `json:"subscriptionExpirationDateTime"`
		UserId                         string `json:"userId"`
		Resource                       string `json:"resource"`
	}
	if err := json.Unmarshal([]byte(evt[1]), &n); err != nil {
		log.Warn().Err(err).Msg("unmarshal notification content")
		return
	}
	log.Debug().Str("notification", evt[1]).Msg("notification content")
	s.trigger()
}

func (s *subscription) trigger() {
	select {
	case s.c <- struct{}{}:
	default:
	}
}

func (s *subscription) socketioOnError(err error) {
	select {
	case s.sioErrCh <- err:
	default:
	}
}

func (s *subscription) Stop() {
	close(s.closeCh)
	close(s.c)
}
