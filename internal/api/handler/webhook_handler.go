package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/aigo/internal/api/response"
)

type WebhookHandler struct{}

func NewWebhookHandler() *WebhookHandler {
	return &WebhookHandler{}
}

func (h *WebhookHandler) Register(c *gin.Context) {
	response.Created(c, gin.H{"message": "webhook endpoint placeholder"})
}

func (h *WebhookHandler) List(c *gin.Context) {
	response.Success(c, []interface{}{})
}

func (h *WebhookHandler) Update(c *gin.Context) {
	response.Success(c, gin.H{"message": "updated"})
}

func (h *WebhookHandler) Delete(c *gin.Context) {
	response.Success(c, gin.H{"message": "deleted"})
}

func (h *WebhookHandler) Logs(c *gin.Context) {
	response.Success(c, []interface{}{})
}
