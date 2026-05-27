package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/aigo/internal/api/response"
	authService "github.com/aigo/internal/service/auth"
)

type AuthHandler struct {
	authService *authService.Service
}

func NewAuthHandler(authService *authService.Service) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req struct {
		PublicKey string `json:"public_key" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.ErrInvalidRequest, "public_key is required")
		return
	}

	user, err := h.authService.Register(req.PublicKey)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.ErrInternal, "registration failed")
		return
	}

	response.Created(c, gin.H{"user_id": user.ID})
}

func (h *AuthHandler) GenerateAPIKey(c *gin.Context) {
	userID := c.GetString("user_id")
	var req struct {
		Name string `json:"name"`
	}
	c.ShouldBindJSON(&req)

	apiKey, rawKey, err := h.authService.GenerateAPIKey(userID, req.Name)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.ErrInternal, "key generation failed")
		return
	}

	response.Created(c, gin.H{
		"api_key_id": apiKey.ID,
		"api_key":    rawKey,
		"name":       apiKey.Name,
	})
}

func (h *AuthHandler) ExchangeToken(c *gin.Context) {
	var req struct {
		APIKey string `json:"api_key" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.ErrInvalidRequest, "api_key is required")
		return
	}

	token, claims, err := h.authService.ExchangeToken(req.APIKey)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, response.ErrUnauthorized, "invalid api key")
		return
	}

	response.Success(c, gin.H{
		"token":      token,
		"user_id":    claims.UserID,
		"expires_at": claims.ExpiresAt,
	})
}
