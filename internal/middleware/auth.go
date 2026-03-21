package middleware

import (
	"net/http"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/fernandovalenzuela/api-home-pay/internal/utils"
	"github.com/gin-gonic/gin"
)

func ClerkAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, ok := clerk.SessionClaimsFromContext(c.Request.Context())
		if !ok {
			utils.ErrorResponseClient(c, http.StatusUnauthorized, "Unauthorized: invalid or missing token")
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
