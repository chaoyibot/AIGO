package eventbus

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// SSEEvent is pushed to connected SSE clients.
type SSEEvent struct {
	Type      string      `json:"type"`
	CreatedAt time.Time   `json:"created_at"`
	Data      interface{} `json:"data"`
}

// SSEHub is a per-user pub/sub hub for real-time event streaming.
type SSEHub struct {
	mu      sync.RWMutex
	clients map[string]map[chan SSEEvent]struct{}
}

func NewSSEHub() *SSEHub {
	return &SSEHub{
		clients: make(map[string]map[chan SSEEvent]struct{}),
	}
}

// Subscribe creates a buffered channel for a user's SSE events.
func (h *SSEHub) Subscribe(userID string) chan SSEEvent {
	ch := make(chan SSEEvent, 32)
	h.mu.Lock()
	if h.clients[userID] == nil {
		h.clients[userID] = make(map[chan SSEEvent]struct{})
	}
	h.clients[userID][ch] = struct{}{}
	h.mu.Unlock()
	return ch
}

// Unsubscribe removes a channel from the hub.
func (h *SSEHub) Unsubscribe(userID string, ch chan SSEEvent) {
	h.mu.Lock()
	if channels, ok := h.clients[userID]; ok {
		delete(channels, ch)
		close(ch)
		if len(channels) == 0 {
			delete(h.clients, userID)
		}
	}
	h.mu.Unlock()
}

// Publish sends an event to all connected clients for a given user.
func (h *SSEHub) Publish(userID string, event SSEEvent) {
	h.mu.RLock()
	channels := h.clients[userID]
	for ch := range channels {
		select {
		case ch <- event:
		default:
			// drop if channel full (client too slow)
		}
	}
	h.mu.RUnlock()
}

// SystemPublish sends an event to ALL connected clients.
func (h *SSEHub) SystemPublish(event SSEEvent) {
	h.mu.RLock()
	for _, channels := range h.clients {
		for ch := range channels {
			select {
			case ch <- event:
			default:
			}
		}
	}
	h.mu.RUnlock()
}

// MarshalSSE formats an SSEEvent as an SSE text/event-stream message.
func MarshalSSE(event SSEEvent) (string, error) {
	data, err := json.Marshal(event)
	if err != nil {
		return "", fmt.Errorf("marshal sse event: %w", err)
	}
	return fmt.Sprintf("data: %s\n\n", string(data)), nil
}
