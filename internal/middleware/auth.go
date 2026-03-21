package middleware

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/clerk/clerk-sdk-go/v2/jwt"
	"github.com/fernandovalenzuela/api-home-pay/internal/utils"
	"github.com/gin-gonic/gin"
)

func ClerkAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract Bearer token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			slog.Warn("auth failed: missing token", "path", c.Request.URL.Path)
			utils.ErrorResponseClient(c, http.StatusUnauthorized, "Unauthorized: missing token")
			c.Abort()
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == authHeader {
			slog.Warn("auth failed: invalid authorization format", "path", c.Request.URL.Path)
			utils.ErrorResponseClient(c, http.StatusUnauthorized, "Unauthorized: invalid authorization format")
			c.Abort()
			return
		}

		// Verify the JWT using Clerk SDK
		claims, err := jwt.Verify(c.Request.Context(), &jwt.VerifyParams{
			Token: token,
		})
		if err != nil {
			slog.Warn("auth failed: invalid token", "path", c.Request.URL.Path, "error", err.Error())
			utils.ErrorResponseClient(c, http.StatusUnauthorized, "Unauthorized: invalid token")
			c.Abort()
			return
		}

		userID := claims.Subject
		if userID == "" {
			utils.ErrorResponseClient(c, http.StatusUnauthorized, "Unauthorized: invalid token claims")
			c.Abort()
			return
		}

		c.Set("user_id", userID)
		c.Set("session_claims", claims)
		c.Next()
	}
}

func RequireAuth() gin.HandlerFunc {
	return ClerkAuthMiddleware()
}

func GetUserID(c *gin.Context) (string, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return "", false
	}
	id, ok := userID.(string)
	return id, ok
}

func GetSessionClaims(c *gin.Context) (*clerk.SessionClaims, bool) {
	claims, exists := c.Get("session_claims")
	if !exists {
		return nil, false
	}
	sessionClaims, ok := claims.(*clerk.SessionClaims)
	return sessionClaims, ok
}
