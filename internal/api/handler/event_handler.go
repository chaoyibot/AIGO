package handler

import (
	"github.com/gin-gonic/gin"
	"io"
)

type EventHandler struct{}

func NewEventHandler() *EventHandler {
	return &EventHandler{}
}

func (h *EventHandler) Stream(c *gin.Context) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")

	c.Stream(func(w io.Writer) bool {
		select {
		case <-c.Request.Context().Done():
			return false
		default:
			// Keep connection alive with heartbeat
			return true
		}
	})
}

// Placeholder for future SSE push logic
func (h *EventHandler) BroadcastEvent(eventType string, data interface{}) {
	// In production, use a pub/sub hub to push events to connected SSE clients
}
