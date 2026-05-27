package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/aigo/internal/api/response"
	authService "github.com/aigo/internal/service/auth"
)

func AuthRequired(authService *authService.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Error(c, http.StatusUnauthorized, response.ErrUnauthorized, "authorization header required")
			c.Abort()
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenStr == authHeader {
			response.Error(c, http.StatusUnauthorized, response.ErrUnauthorized, "bearer token required")
			c.Abort()
			return
		}

		claims, err := authService.ValidateToken(tokenStr)
		if err != nil {
			response.Error(c, http.StatusUnauthorized, response.ErrUnauthorized, "invalid or expired token")
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("public_key", claims.PublicKey)
		c.Next()
	}
}
