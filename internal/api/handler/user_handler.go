package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/aigo/internal/api/response"
	"github.com/aigo/internal/repository"
	authService "github.com/aigo/internal/service/auth"
)

type UserHandler struct {
	userRepo    *repository.UserRepo
	authService *authService.Service
}

func NewUserHandler(userRepo *repository.UserRepo, authService *authService.Service) *UserHandler {
	return &UserHandler{userRepo: userRepo, authService: authService}
}

// GetMe returns the authenticated user's profile.
func (h *UserHandler) GetMe(c *gin.Context) {
	userID := c.GetString("user_id")
	user, err := h.userRepo.FindByID(userID)
	if err != nil {
		response.Error(c, http.StatusNotFound, response.ErrNotFound, "user not found")
		return
	}
	response.Success(c, gin.H{
		"id":         user.ID,
		"public_key": user.PublicKey,
		"nickname":   user.Nickname,
		"role":       user.Role,
		"status":     user.Status,
		"created_at": user.CreatedAt,
	})
}

// GetUserByID returns a user's public info (public key) for encryption.
func (h *UserHandler) GetUserByID(c *gin.Context) {
	id := c.Param("id")
	user, err := h.userRepo.FindByID(id)
	if err != nil {
		response.Error(c, http.StatusNotFound, response.ErrNotFound, "user not found")
		return
	}
	// Only expose safe fields
	response.Success(c, gin.H{
		"id":         user.ID,
		"public_key": user.PublicKey,
		"nickname":   user.Nickname,
		"role":       user.Role,
	})
}

// UpdateMe updates the authenticated user's profile (nickname only for now).
func (h *UserHandler) UpdateMe(c *gin.Context) {
	userID := c.GetString("user_id")
	var req struct {
		Nickname string `json:"nickname"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.ErrInvalidRequest, "invalid request")
		return
	}
	if err := h.authService.UpdateNickname(userID, req.Nickname); err != nil {
		response.Error(c, http.StatusInternalServerError, response.ErrInternal, "update failed")
		return
	}
	response.Success(c, gin.H{"updated": "nickname"})
}
