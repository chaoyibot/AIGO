package handler

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/aigo/internal/api/response"
	"github.com/aigo/internal/model"
	"github.com/aigo/internal/repository"
)

type WebhookHandler struct {
	webhookRepo *repository.WebhookRepo
	httpClient  *http.Client
}

func NewWebhookHandler(webhookRepo *repository.WebhookRepo) *WebhookHandler {
	return &WebhookHandler{
		webhookRepo: webhookRepo,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (h *WebhookHandler) Register(c *gin.Context) {
	userID := c.GetString("user_id")
	var req struct {
		URL    string   `json:"url" binding:"required"`
		Events []string `json:"events" binding:"required"`
		Secret string   `json:"secret"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.ErrInvalidRequest, "url and events are required")
		return
	}
	if req.Secret == "" {
		req.Secret = uuid.New().String()
	}

	hook := &model.Webhook{
		ID:     uuid.New().String(),
		UserID: userID,
		URL:    req.URL,
		Secret: req.Secret,
		Events: req.Events,
		Active: true,
	}
	if err := h.webhookRepo.Create(hook); err != nil {
		response.Error(c, http.StatusInternalServerError, response.ErrInternal, "create webhook failed")
		return
	}

	response.Created(c, gin.H{
		"webhook_id": hook.ID,
		"secret":     hook.Secret,
		"url":        hook.URL,
		"events":     hook.Events,
	})
}

func (h *WebhookHandler) List(c *gin.Context) {
	userID := c.GetString("user_id")
	hooks, err := h.webhookRepo.ListByUser(userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.ErrInternal, "list failed")
		return
	}
	if hooks == nil {
		hooks = []model.Webhook{}
	}
	response.Success(c, hooks)
}

func (h *WebhookHandler) Update(c *gin.Context) {
	userID := c.GetString("user_id")
	hookID := c.Param("id")

	existing, err := h.webhookRepo.FindByID(hookID)
	if err != nil || existing.UserID != userID {
		response.Error(c, http.StatusNotFound, response.ErrNotFound, "webhook not found")
		return
	}

	var req struct {
		URL    *string  `json:"url"`
		Events []string `json:"events"`
		Secret *string  `json:"secret"`
		Active *bool    `json:"active"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.ErrInvalidRequest, "invalid request")
		return
	}

	if req.URL != nil { existing.URL = *req.URL }
	if req.Secret != nil { existing.Secret = *req.Secret }
	if req.Events != nil { existing.Events = req.Events }
	if req.Active != nil { existing.Active = *req.Active }

	if err := h.webhookRepo.Update(existing); err != nil {
		response.Error(c, http.StatusInternalServerError, response.ErrInternal, "update failed")
		return
	}
	response.Success(c, gin.H{"status": "updated"})
}

func (h *WebhookHandler) Delete(c *gin.Context) {
	userID := c.GetString("user_id")
	hookID := c.Param("id")
	if err := h.webhookRepo.Delete(hookID, userID); err != nil {
		response.Error(c, http.StatusNotFound, response.ErrNotFound, "webhook not found")
		return
	}
	response.Success(c, gin.H{"status": "deleted"})
}

func (h *WebhookHandler) Logs(c *gin.Context) {
	userID := c.GetString("user_id")
	hookID := c.Param("id")

	// Verify ownership
	existing, err := h.webhookRepo.FindByID(hookID)
	if err != nil || existing.UserID != userID {
		response.Error(c, http.StatusNotFound, response.ErrNotFound, "webhook not found")
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	logs, err := h.webhookRepo.Logs(hookID, limit)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.ErrInternal, "query logs failed")
		return
	}
	if logs == nil {
		logs = []model.WebhookLog{}
	}
	response.Success(c, logs)
}

// FireEvent dispatches an event to all webhooks registered for that event type.
func (h *WebhookHandler) FireEvent(eventType string, userID string, payload interface{}) {
	go func() {
		hooks, err := h.webhookRepo.ListByEvent(eventType)
		if err != nil {
			return
		}

		payloadJSON, _ := json.Marshal(payload)

		for _, hook := range hooks {
			// Only fire for the affected user
			if hook.UserID != userID {
				continue
			}

			go h.deliverWebhook(hook, eventType, payloadJSON)
		}
	}()
}

// FireEventToAll dispatches an event to ALL webhooks for that event type (e.g. system broadcasts).
func (h *WebhookHandler) FireEventToAll(eventType string, payload interface{}) {
	go func() {
		hooks, err := h.webhookRepo.ListByEvent(eventType)
		if err != nil {
			return
		}

		payloadJSON, _ := json.Marshal(payload)

		for _, hook := range hooks {
			go h.deliverWebhook(hook, eventType, payloadJSON)
		}
	}()
}

func (h *WebhookHandler) deliverWebhook(hook model.Webhook, eventType string, payloadJSON []byte) {
	body := map[string]interface{}{
		"event":      eventType,
		"payload":    json.RawMessage(payloadJSON),
		"timestamp":  time.Now().Unix(),
		"webhook_id": hook.ID,
	}
	bodyJSON, _ := json.Marshal(body)

	// Sign with HMAC-SHA256
	mac := hmac.New(sha256.New, []byte(hook.Secret))
	mac.Write(bodyJSON)
	signature := hex.EncodeToString(mac.Sum(nil))

	req, err := http.NewRequest("POST", hook.URL, bytes.NewReader(bodyJSON))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-AIGO-Signature", signature)
	req.Header.Set("X-AIGO-Event", eventType)

	resp, err := h.httpClient.Do(req)

	logEntry := &model.WebhookLog{
		ID:        uuid.New().String(),
		WebhookID: hook.ID,
		EventType: eventType,
		Payload:   string(payloadJSON),
	}

	if err != nil {
		logEntry.Status = "failed"
		logEntry.ResponseBody = err.Error()
	} else {
		defer resp.Body.Close()
		respBody, _ := io.ReadAll(resp.Body)
		logEntry.ResponseCode = resp.StatusCode
		logEntry.ResponseBody = string(respBody)
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			logEntry.Status = "success"
		} else if resp.StatusCode >= 500 {
			logEntry.Status = "retrying"
			nextRetry := time.Now().Add(60 * time.Second)
			logEntry.NextRetryAt = &nextRetry
		} else {
			logEntry.Status = "failed"
		}
	}

	// Best-effort log write
	h.webhookRepo.CreateLog(logEntry)
}

// Ensure interfaces
var _ = fmt.Sprintf
