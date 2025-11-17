package socketio

import (
	"fmt"
	"math"
	"net/http"
	"net/url"
	"strings"
	"sync/atomic"
	"time"

	"github.com/auriora/onemount/internal/socketio/protocol"
)

const (
	stateOpen uint32 = iota
	stateConnecting
	stateReady
	stateReconnecting
	stateClose
)

type option struct {
	AutoReconnect    bool
	MaxReconnections int32
}

var defaultOption = &option{
	AutoReconnect:    true,
	MaxReconnections: math.MaxInt32,
}

type Client struct {
	emitter
	state     uint32
	url       *url.URL
	option    *option
	transprot protocol.Transport
	outChan   chan *protocol.Packet
	closeChan chan bool
}

func Socket(urlstring string) (*Client, error) {
	u, err := url.Parse(urlstring)
	if err != nil {
		return nil, err
	}
	switch strings.ToLower(u.Scheme) {
	case "https":
		u.Scheme = "wss"
	case "http":
		u.Scheme = "ws"
	case "wss", "ws":
		// already websocket compatible
	default:
		return nil, fmt.Errorf("unsupported scheme %q", u.Scheme)
	}
	cleanPath := strings.TrimSuffix(u.Path, "/")
	u.Path = cleanPath + "/socket.io/"
	q := u.Query()
	q.Set("EIO", "4")
	q.Set("transport", "websocket")
	u.RawQuery = q.Encode()
	return &Client{
		emitter:   emitter{listeners: make(map[string][]Listener)},
		url:       u,
		option:    defaultOption,
		transprot: protocol.NewWebSocketTransport(),
		outChan:   make(chan *protocol.Packet, 64),
		closeChan: make(chan bool),
	}, nil
}

func (s *Client) Connect(requestHeader http.Header) {
	if atomic.CompareAndSwapUint32(&s.state, stateOpen, stateConnecting) {
		conn, err := s.transprot.Dial(s.url.String(), requestHeader)
		if err != nil {
			s.emit(EventError, err)
			go s.reconnect(stateConnecting, requestHeader)
			return
		}
		if atomic.CompareAndSwapUint32(&s.state, stateConnecting, stateReady) {
			go s.start(conn, requestHeader)
			s.emit(EventConnect)
		} else {
			conn.Close()
		}
	}
}

func (s *Client) Disconnect() {
	atomic.StoreUint32(&s.state, stateClose)
	close(s.outChan)
	close(s.closeChan)
}

func (s *Client) Emit(event string, args ...interface{}) {
	if atomic.LoadUint32(&s.state) == stateReady && !s.emit(event, args) {
		m := &protocol.Message{
			Type:      protocol.MessageTypeEvent,
			Namespace: "/",
			ID:        -1,
			Event:     event,
			Payloads:  args,
		}
		p, err := m.Encode()
		if err != nil {
			s.emit(EventError, err)
		} else {
			s.outChan <- p
		}
	}
}

func (s *Client) reconnect(state uint32, requestHeader http.Header) {
	time.Sleep(time.Second)
	if atomic.CompareAndSwapUint32(&s.state, state, stateReconnecting) {
		conn, err := s.transprot.Dial(s.url.String(), requestHeader)
		if err != nil {
			s.emit(EventError, err)
			go s.reconnect(stateReconnecting, requestHeader)
			return
		}
		if atomic.CompareAndSwapUint32(&s.state, stateReconnecting, stateReady) {
			go s.start(conn, requestHeader)
			s.emit(EventReconnect)
		} else {
			conn.Close()
		}
	}
}

func (s *Client) start(conn protocol.Conn, requestHeader http.Header) {
	stopper := make(chan bool)
	go s.startRead(conn, stopper)
	go s.startWrite(conn, stopper)
	select {
	case <-stopper:
		go s.reconnect(stateReady, requestHeader)
		conn.Close()
	case <-s.closeChan:
		conn.Close()
	}
}

func (s *Client) startRead(conn protocol.Conn, stopper chan bool) {
	defer func() {
		recover()
	}()
	for atomic.LoadUint32(&s.state) == stateReady {
		p, err := conn.Read()
		if err != nil {
			s.emit(EventError, err)
			close(stopper)
			return
		}
		switch p.Type {
		case protocol.PacketTypeOpen:
			h, err := p.DecodeHandshake()
			if err != nil {
				s.emit(EventError, err)
			} else {
				go s.startPing(h, stopper)
			}
		case protocol.PacketTypePing:
			s.outChan <- protocol.NewPongPacket()
		case protocol.PacketTypeMessage:
			m, err := p.DecodeMessage()
			if err != nil {
				s.emit(EventError, err)
			} else {
				s.emit(m.Event, m.Payloads...)
			}
		}
	}
}

func (s *Client) startWrite(conn protocol.Conn, stopper chan bool) {
	defer func() {
		recover()
	}()
	for atomic.LoadUint32(&s.state) == stateReady {
		select {
		case <-stopper:
			return
		case p, ok := <-s.outChan:
			if !ok {
				return
			}
			err := conn.Write(p)
			if err != nil {
				s.emit(EventError, err)
				close(stopper)
				return
			}
		}

	}
}

func (s *Client) startPing(h *protocol.Handshake, stopper chan bool) {
	defer func() {
		recover()
	}()
	for {
		time.Sleep(time.Duration(h.PingInterval) * time.Millisecond)
		select {
		case <-stopper:
			return
		case <-s.closeChan:
			return
		default:
		}
		if atomic.LoadUint32(&s.state) != stateReady {
			return
		}
		s.outChan <- protocol.NewPingPacket()
	}
}
