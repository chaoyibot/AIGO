package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/aigo/internal/api/response"
	"github.com/aigo/internal/model"
	"github.com/aigo/internal/repository"
)

type MessageHandler struct {
	msgRepo *repository.MessageRepo
}

func NewMessageHandler(msgRepo *repository.MessageRepo) *MessageHandler {
	return &MessageHandler{msgRepo: msgRepo}
}

// Send allows an authenticated user to send a private message to another user.
func (h *MessageHandler) Send(c *gin.Context) {
	userID := c.GetString("user_id")

	var req struct {
		ReceiverID string  `json:"receiver_id" binding:"required"`
		Subject    string  `json:"subject"`
		Body       string  `json:"body" binding:"required"`
		ReplyTo    *string `json:"reply_to"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.ErrInvalidRequest, "receiver_id and body are required")
		return
	}

	msg := &model.Message{
		ID:         uuid.New().String(),
		SenderID:   userID,
		ReceiverID: req.ReceiverID,
		Subject:    req.Subject,
		Body:       req.Body,
		ReplyTo:    req.ReplyTo,
	}

	if err := h.msgRepo.Send(msg); err != nil {
		response.Error(c, http.StatusInternalServerError, response.ErrInternal, "send failed")
		return
	}

	response.Created(c, gin.H{
		"message_id":  msg.ID,
		"receiver_id": req.ReceiverID,
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
