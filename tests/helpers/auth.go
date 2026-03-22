package helpers

import (
	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/gin-gonic/gin"
)

// MockAuthContext sets up a Gin context with a mock authenticated user
func MockAuthContext(c *gin.Context, userID string) {
	claims := &clerk.SessionClaims{}
	claims.Subject = userID
	c.Set("user_id", userID)
	c.Set("session_claims", claims)
}

// MockUnauthorizedContext sets up a Gin context without authentication
func MockUnauthorizedContext(c *gin.Context) {
	// Don't set any auth-related context values
}

// GetMockToken returns a mock JWT token for testing
func GetMockToken() string {
	// This is a dummy token for testing purposes
	return "Bearer mock_token_for_testing_only"
}
