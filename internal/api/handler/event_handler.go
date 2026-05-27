package handler

import (
	"fmt"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/aigo/internal/eventbus"
)

type EventHandler struct {
	sseHub *eventbus.SSEHub
}

func NewEventHandler(sseHub *eventbus.SSEHub) *EventHandler {
	return &EventHandler{sseHub: sseHub}
}

// Stream provides a Server-Sent Events endpoint. The authenticated user receives
// real-time events (new messages, order updates, etc.) as they happen.
func (h *EventHandler) Stream(c *gin.Context) {
	userID := c.GetString("user_id")

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")

	// Subscribe to SSE hub for this user
	eventCh := h.sseHub.Subscribe(userID)
	defer h.sseHub.Unsubscribe(userID, eventCh)

	// Send keepalive every 30 seconds
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	c.Stream(func(w io.Writer) bool {
		select {
		case event, ok := <-eventCh:
			if !ok {
				return false
			}
			msg, err := eventbus.MarshalSSE(event)
			if err != nil {
				return true // skip malformed event
			}
			_, writeErr := fmt.Fprint(w, msg)
			return writeErr == nil

		case <-ticker.C:
			fmt.Fprint(w, ": keepalive\n\n")
			return true

		case <-c.Request.Context().Done():
			return false
		}
	})
}
