package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/aigo/internal/api/response"
	"github.com/aigo/internal/eventbus"
	"github.com/aigo/internal/model"
	"github.com/aigo/internal/repository"
)

const (
	eventMessageNew = "message.new"
)

type MessageHandler struct {
	msgRepo     *repository.MessageRepo
	eventBus    *eventbus.Bus
	sseHub      *eventbus.SSEHub
	webhookH    *WebhookHandler
}

func NewMessageHandler(msgRepo *repository.MessageRepo, eventBus *eventbus.Bus, sseHub *eventbus.SSEHub, webhookH *WebhookHandler) *MessageHandler {
	return &MessageHandler{
		msgRepo:  msgRepo,
		eventBus: eventBus,
		sseHub:   sseHub,
		webhookH: webhookH,
	}
}

// Send allows an authenticated user to send a private message to another user.
// If IsEncrypted is true, Body is treated as ciphertext (RSA-OAEP encrypted with
// the receiver's public key). The server stores it blindly and never sees plaintext.
func (h *MessageHandler) Send(c *gin.Context) {
	userID := c.GetString("user_id")

	var req struct {
		ReceiverID  string `json:"receiver_id" binding:"required"`
		Subject     string `json:"subject"`
		Body        string `json:"body" binding:"required"`
		IsEncrypted bool   `json:"is_encrypted"`
		ReplyTo     *string `json:"reply_to"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.ErrInvalidRequest, "receiver_id and body are required")
		return
	}

	msg := &model.Message{
		ID:          uuid.New().String(),
		SenderID:    userID,
		ReceiverID:  req.ReceiverID,
		Subject:     req.Subject,
		Body:        req.Body,
		IsEncrypted: req.IsEncrypted,
		ReplyTo:     req.ReplyTo,
	}

	if err := h.msgRepo.Send(msg); err != nil {
		response.Error(c, http.StatusInternalServerError, response.ErrInternal, "send failed")
		return
	}

	// --- Dispatch events asynchronously ---

	// 1. Publish to NATS event bus (if available)
	if h.eventBus != nil {
		go func() {
			eventData := map[string]interface{}{
				"message_id":   msg.ID,
				"sender_id":    msg.SenderID,
				"receiver_id":  msg.ReceiverID,
				"subject":      msg.Subject,
				"body_length":  len(msg.Body),
				"is_encrypted": msg.IsEncrypted,
				"reply_to":     msg.ReplyTo,
			}
			h.eventBus.Publish(eventMessageNew, eventData)
		}()
	}

	// 2. Push to SSE hub for real-time delivery to receiver's connected clients
	go func() {
		h.sseHub.Publish(msg.ReceiverID, eventbus.SSEEvent{
			Type:      eventMessageNew,
			CreatedAt: msg.CreatedAt,
			Data: map[string]interface{}{
				"message_id":   msg.ID,
				"sender_id":    msg.SenderID,
				"subject":      msg.Subject,
				"body":         msg.Body,
				"is_encrypted": msg.IsEncrypted,
				"created_at":   msg.CreatedAt,
			},
		})
	}()

	// 3. Fire webhooks for receiver
	go func() {
		webhookPayload := map[string]interface{}{
			"message_id":   msg.ID,
			"sender_id":    msg.SenderID,
			"receiver_id":  msg.ReceiverID,
			"subject":      msg.Subject,
			"body":         msg.Body,
			"is_encrypted": msg.IsEncrypted,
			"created_at":   msg.CreatedAt,
		}
		if h.webhookH != nil {
			// Notify receiver
			h.webhookH.FireEvent(eventMessageNew, msg.ReceiverID, webhookPayload)
			// Also notify sender (so they know the message was sent)
			h.webhookH.FireEvent(eventMessageNew, msg.SenderID, map[string]interface{}{
				"message_id":   msg.ID,
				"receiver_id":  msg.ReceiverID,
				"subject":      msg.Subject,
				"status":       "sent",
				"created_at":   msg.CreatedAt,
			})
		}
	}()

	response.Created(c, gin.H{
		"message_id":   msg.ID,
		"receiver_id":  req.ReceiverID,
		"is_encrypted": req.IsEncrypted,
	})
}

// Inbox lists messages received by the authenticated user.
func (h *MessageHandler) Inbox(c *gin.Context) {
	userID := c.GetString("user_id")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	msgs, err := h.msgRepo.Inbox(userID, limit)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.ErrInternal, "query inbox failed")
		return
	}
	if msgs == nil {
		msgs = []model.Message{}
	}
	response.Success(c, msgs)
}

// Sent lists messages sent by the authenticated user.
func (h *MessageHandler) Sent(c *gin.Context) {
	userID := c.GetString("user_id")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	msgs, err := h.msgRepo.Sent(userID, limit)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.ErrInternal, "query sent failed")
		return
	}
	if msgs == nil {
		msgs = []model.Message{}
	}
	response.Success(c, msgs)
}

// UnreadCount returns the count of unread messages.
func (h *MessageHandler) UnreadCount(c *gin.Context) {
	userID := c.GetString("user_id")
	count, err := h.msgRepo.UnreadCount(userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.ErrInternal, "query failed")
		return
	}
	response.Success(c, gin.H{"unread_count": count})
}

// MarkRead marks a received message as read.
func (h *MessageHandler) MarkRead(c *gin.Context) {
	userID := c.GetString("user_id")
	msgID := c.Param("id")

	if err := h.msgRepo.MarkRead(msgID, userID); err != nil {
		response.Error(c, http.StatusInternalServerError, response.ErrInternal, "mark read failed")
		return
	}
	response.Success(c, gin.H{"status": "read"})
}

// GetByID returns a single message (only if user is sender or receiver).
func (h *MessageHandler) GetByID(c *gin.Context) {
	userID := c.GetString("user_id")
	msgID := c.Param("id")

	msg, err := h.msgRepo.FindByID(msgID)
	if err != nil {
		response.Error(c, http.StatusNotFound, response.ErrNotFound, "message not found")
		return
	}
	if msg.SenderID != userID && msg.ReceiverID != userID {
		response.Error(c, http.StatusForbidden, response.ErrForbidden, "access denied")
		return
	}
	response.Success(c, msg)
}

// Ensure json import is used
var _ = json.Marshal
