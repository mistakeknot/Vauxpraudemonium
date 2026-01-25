package autarch

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"sync"
	"time"

	"nhooyr.io/websocket"
)

// EventHandler is called when a domain event is received
type EventHandler func(event DomainEvent)

// Subscriber manages a WebSocket connection for real-time updates
type Subscriber struct {
	url      string
	handlers []EventHandler
	conn     *websocket.Conn
	ctx      context.Context
	cancel   context.CancelFunc
	mu       sync.Mutex
}

// NewSubscriber creates a new WebSocket subscriber
func NewSubscriber(baseURL, agent string) (*Subscriber, error) {
	// Convert http:// to ws://
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("parse url: %w", err)
	}
	if u.Scheme == "http" {
		u.Scheme = "ws"
	} else if u.Scheme == "https" {
		u.Scheme = "wss"
	}
	u.Path = "/ws/agents/" + agent

	ctx, cancel := context.WithCancel(context.Background())
	return &Subscriber{
		url:    u.String(),
		ctx:    ctx,
		cancel: cancel,
	}, nil
}

// OnEvent registers an event handler
func (s *Subscriber) OnEvent(handler EventHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.handlers = append(s.handlers, handler)
}

// Connect establishes the WebSocket connection
func (s *Subscriber) Connect() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.conn != nil {
		return nil // Already connected
	}

	ctx, cancel := context.WithTimeout(s.ctx, 10*time.Second)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, s.url, nil)
	if err != nil {
		return fmt.Errorf("dial websocket: %w", err)
	}

	s.conn = conn
	go s.readLoop()

	return nil
}

// Close closes the WebSocket connection
func (s *Subscriber) Close() error {
	s.cancel()

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.conn != nil {
		return s.conn.Close(websocket.StatusNormalClosure, "closing")
	}
	return nil
}

func (s *Subscriber) readLoop() {
	for {
		select {
		case <-s.ctx.Done():
			return
		default:
		}

		s.mu.Lock()
		conn := s.conn
		s.mu.Unlock()

		if conn == nil {
			return
		}

		_, message, err := conn.Read(s.ctx)
		if err != nil {
			// Connection closed or error
			return
		}

		var event DomainEvent
		if err := json.Unmarshal(message, &event); err != nil {
			continue // Skip malformed events
		}

		s.mu.Lock()
		handlers := make([]EventHandler, len(s.handlers))
		copy(handlers, s.handlers)
		s.mu.Unlock()

		for _, h := range handlers {
			h(event)
		}
	}
}

// ConnectedClient combines Client with real-time subscription
type ConnectedClient struct {
	*Client
	*Subscriber
}

// NewConnectedClient creates a client with WebSocket subscription
func NewConnectedClient(baseURL, agent string) (*ConnectedClient, error) {
	client := NewClient(baseURL)
	sub, err := NewSubscriber(baseURL, agent)
	if err != nil {
		return nil, err
	}
	return &ConnectedClient{
		Client:     client,
		Subscriber: sub,
	}, nil
}

// Close closes both the subscriber and any resources
func (c *ConnectedClient) Close() error {
	if c.Subscriber != nil {
		return c.Subscriber.Close()
	}
	return nil
}
