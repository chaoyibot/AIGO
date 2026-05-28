package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/aigo/internal/api/response"
	broadcastService "github.com/aigo/internal/service/broadcast"
)

type BroadcastHandler struct {
	broadcastService *broadcastService.Service
}

func NewBroadcastHandler(broadcastService *broadcastService.Service) *BroadcastHandler {
	return &BroadcastHandler{broadcastService: broadcastService}
}

func (h *BroadcastHandler) SystemBroadcast(c *gin.Context) {
	var req struct {
		Title   string `json:"title" binding:"required"`
		Content string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.ErrInvalidRequest, "title and content required")
		return
	}

b, err := h.broadcastService.PublishSystem(req.Title, req.Content)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.ErrInternal, "publish failed")
		return
	}
	response.Created(c, gin.H{"id": b.ID})
}

func (h *BroadcastHandler) CommercialBroadcast(c *gin.Context) {
	userID := c.GetString("user_id")
	var req struct {
		Title   string `json:"title" binding:"required"`
		Content string `json:"content" binding:"required"`
		Level   string `json:"level"`    // basic | standard | premium
		LinkURL string `json:"link_url"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.ErrInvalidRequest, "invalid request")
		return
	}
	if req.Level == "" {
		req.Level = "basic"
	}

	b, err := h.broadcastService.PublishCommercial(userID, req.Title, req.Content, req.Level, req.LinkURL)
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.ErrInsufficientPoints, err.Error())
		return
	}
	response.Created(c, gin.H{
		"broadcast_id":   b.ID,
		"points_cost":    b.PointsCost,
		"is_pinned":      b.IsPinned,
		"pin_expires_at": b.PinExpiresAt,
	})
}

func (h *BroadcastHandler) List(c *gin.Context) {
	broadcasts, err := h.broadcastService.ListBroadcasts()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.ErrInternal, "list failed")
		return
	}
	response.Success(c, broadcasts)
}

func (h *BroadcastHandler) UnreadCount(c *gin.Context) {
	userID := c.GetString("user_id")
	count, err := h.broadcastService.UnreadCount(userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.ErrInternal, "query failed")
		return
	}
	response.Success(c, gin.H{"unread_count": count})
}

func (h *BroadcastHandler) MarkRead(c *gin.Context) {
	userID := c.GetString("user_id")
	broadcastID := c.Param("id")
	if err := h.broadcastService.MarkRead(broadcastID, userID); err != nil {
		response.Error(c, http.StatusInternalServerError, response.ErrInternal, "mark read failed")
		return
	}
	response.Success(c, gin.H{"status": "read"})
}

// AdminPin manually pins a broadcast until the given time.
func (h *BroadcastHandler) AdminPin(c *gin.Context) {
	broadcastID := c.Param("id")
	var req struct {
		PinExpiresAt time.Time `json:"pin_expires_at" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.ErrInvalidRequest, "pin_expires_at required")
		return
	}
	if req.PinExpiresAt.Before(time.Now()) {
		response.Error(c, http.StatusBadRequest, response.ErrInvalidRequest, "pin_expires_at must be in the future")
		return
	}
	if err := h.broadcastService.SetPin(broadcastID, req.PinExpiresAt); err != nil {
		response.Error(c, http.StatusInternalServerError, response.ErrInternal, "pin failed")
		return
	}
	response.Success(c, gin.H{"status": "pinned", "pin_expires_at": req.PinExpiresAt})
}

// AdminUnpin removes the pinned status from a broadcast.
func (h *BroadcastHandler) AdminUnpin(c *gin.Context) {
	broadcastID := c.Param("id")
	if err := h.broadcastService.Unpin(broadcastID); err != nil {
		response.Error(c, http.StatusInternalServerError, response.ErrInternal, "unpin failed")
		return
	}
	response.Success(c, gin.H{"status": "unpinned"})
}