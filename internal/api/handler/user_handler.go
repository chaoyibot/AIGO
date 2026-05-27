package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/aigo/internal/api/response"
	"github.com/aigo/internal/repository"
)

type UserHandler struct {
	userRepo *repository.UserRepo
}

func NewUserHandler(userRepo *repository.UserRepo) *UserHandler {
	return &UserHandler{userRepo: userRepo}
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
		"role":       user.Role,
	})
}
