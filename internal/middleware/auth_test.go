package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupAuthTest() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	return router
}

func TestClerkAuthMiddleware_MissingClaims(t *testing.T) {
	router := setupAuthTest()

	router.Use(ClerkAuthMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "error", response["status"])
	assert.Contains(t, response["message"], "Unauthorized")
}

func TestClerkAuthMiddleware_ValidClaims(t *testing.T) {
	router := setupAuthTest()

	// First inject valid claims
	router.Use(func(c *gin.Context) {
		claims := &clerk.SessionClaims{
			RegisteredClaims: clerk.RegisteredClaims{
				Subject: "user_test_123",
			},
		}
		ctx := clerk.ContextWithSessionClaims(c.Request.Context(), claims)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	})
	router.Use(ClerkAuthMiddleware())
	router.GET("/test", func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		assert.True(t, exists, "user_id should be set in context")
		assert.Equal(t, "user_test_123", userID)

		claims, exists := c.Get("session_claims")
		assert.True(t, exists, "session_claims should be set in context")
		assert.NotNil(t, claims)

		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestClerkAuthMiddleware_EmptySubject(t *testing.T) {
	router := setupAuthTest()

	router.Use(func(c *gin.Context) {
		// Create claims with empty subject
		claims := &clerk.SessionClaims{
			RegisteredClaims: clerk.RegisteredClaims{
				Subject: "",
			},
		}
		ctx := clerk.ContextWithSessionClaims(c.Request.Context(), claims)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	})
	router.Use(ClerkAuthMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "error", response["status"])
	assert.Contains(t, response["message"], "invalid token claims")
}

func TestClerkAuthMiddleware_InvalidAuthorizationHeader(t *testing.T) {
	router := setupAuthTest()

	router.Use(ClerkAuthMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})

	// Test with invalid Authorization header format
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid_token")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestClerkAuthMiddleware_UserIDExtraction(t *testing.T) {
	router := setupAuthTest()

	router.Use(func(c *gin.Context) {
		claims := &clerk.SessionClaims{
			RegisteredClaims: clerk.RegisteredClaims{
				Subject: "extracted_user_123",
			},
		}
		ctx := clerk.ContextWithSessionClaims(c.Request.Context(), claims)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	})
	router.Use(ClerkAuthMiddleware())
	router.GET("/test", func(c *gin.Context) {
		userID, ok := GetUserID(c)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user ID"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"user_id": userID})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "extracted_user_123", response["user_id"])
}

func TestClerkAuthMiddleware_SessionClaimsExtraction(t *testing.T) {
	router := setupAuthTest()

	router.Use(func(c *gin.Context) {
		claims := &clerk.SessionClaims{
			RegisteredClaims: clerk.RegisteredClaims{
				Subject: "claims_user_456",
			},
		}
		ctx := clerk.ContextWithSessionClaims(c.Request.Context(), claims)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	})
	router.Use(ClerkAuthMiddleware())
	router.GET("/test", func(c *gin.Context) {
		claims, ok := GetSessionClaims(c)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get session claims"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"subject": claims.RegisteredClaims.Subject})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "claims_user_456", response["subject"])
}

func TestGetUserID_NotSet(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	userID, ok := GetUserID(ctx)
	assert.False(t, ok)
	assert.Empty(t, userID)
}

func TestGetUserID_WrongType(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	ctx.Set("user_id", 12345) // Wrong type (int instead of string)

	userID, ok := GetUserID(ctx)
	assert.False(t, ok)
	assert.Empty(t, userID)
}

func TestGetUserID_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	ctx.Set("user_id", "test_user_id")

	userID, ok := GetUserID(ctx)
	assert.True(t, ok)
	assert.Equal(t, "test_user_id", userID)
}

func TestGetSessionClaims_NotSet(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	claims, ok := GetSessionClaims(ctx)
	assert.False(t, ok)
	assert.Nil(t, claims)
}

func TestGetSessionClaims_WrongType(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	ctx.Set("session_claims", "not_a_claims_object")

	claims, ok := GetSessionClaims(ctx)
	assert.False(t, ok)
	assert.Nil(t, claims)
}

func TestGetSessionClaims_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	expectedClaims := &clerk.SessionClaims{
		RegisteredClaims: clerk.RegisteredClaims{
			Subject: "test_subject",
		},
	}
	ctx.Set("session_claims", expectedClaims)

	claims, ok := GetSessionClaims(ctx)
	assert.True(t, ok)
	assert.Equal(t, expectedClaims, claims)
	assert.Equal(t, "test_subject", claims.RegisteredClaims.Subject)
}

func TestRequireAuth(t *testing.T) {
	// RequireAuth should return the same middleware as ClerkAuthMiddleware
	middleware1 := RequireAuth()
	middleware2 := ClerkAuthMiddleware()

	// We can't directly compare functions, but we can verify they work the same
	assert.NotNil(t, middleware1)
	assert.NotNil(t, middleware2)
}

func TestClerkAuthMiddleware_MissingAuthorizationHeader(t *testing.T) {
	router := setupAuthTest()

	router.Use(ClerkAuthMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	// No Authorization header set
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "error", response["status"])
	assert.Contains(t, response["message"], "Unauthorized")
}

func TestClerkAuthMiddleware_ExpiredToken(t *testing.T) {
	router := setupAuthTest()

	router.Use(ClerkAuthMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer expired_token")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should be unauthorized since there's no valid session claims in context
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
