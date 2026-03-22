package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupResponseTest() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	return router
}

func TestResponseMiddleware_SuccessfulResponse(t *testing.T) {
	router := setupResponseTest()

	router.Use(ResponseMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "success",
			"data":   "test data",
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "success", response["status"])
	assert.Equal(t, "test data", response["data"])
}

func TestResponseMiddleware_ErrorResponse(t *testing.T) {
	router := setupResponseTest()

	router.Use(ResponseMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid request",
			"code":    400,
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "error", response["status"])
	assert.Equal(t, "Invalid request", response["message"])
	assert.Equal(t, float64(400), response["code"])
}

func TestResponseMiddleware_ResponseWithData(t *testing.T) {
	router := setupResponseTest()

	router.Use(ResponseMiddleware())
	router.GET("/test", func(c *gin.Context) {
		data := gin.H{
			"id":    1,
			"name":  "Test Item",
			"price": 99.99,
		}
		c.JSON(http.StatusOK, gin.H{
			"status": "success",
			"data":   data,
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "success", response["status"])

	// Check nested data
	data, ok := response["data"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, float64(1), data["id"])
	assert.Equal(t, "Test Item", data["name"])
	assert.Equal(t, 99.99, data["price"])
}

func TestResponseMiddleware_ResponseWithoutData(t *testing.T) {
	router := setupResponseTest()

	router.Use(ResponseMiddleware())
	router.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusCreated, gin.H{
			"status":  "success",
			"message": "Resource created successfully",
		})
	})

	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "success", response["status"])
	assert.Equal(t, "Resource created successfully", response["message"])
	// data should be nil or not present
	_, hasData := response["data"]
	assert.False(t, hasData)
}

func TestResponseMiddleware_SetHeaderOnce(t *testing.T) {
	router := setupResponseTest()

	router.Use(ResponseMiddleware())
	router.GET("/test", func(c *gin.Context) {
		// Verify the header is already set
		contentType := c.Writer.Header().Get("Content-Type")
		assert.Equal(t, "application/json", contentType)
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
}

func TestResponseMiddleware_ChainedCalls(t *testing.T) {
	router := setupResponseTest()

	router.Use(ResponseMiddleware())
	router.Use(func(c *gin.Context) {
		// Middleware in between should also work
		c.Set("test_key", "test_value")
		c.Next()
	})
	router.GET("/test", func(c *gin.Context) {
		value, exists := c.Get("test_key")
		assert.True(t, exists)
		assert.Equal(t, "test_value", value)
		c.JSON(http.StatusOK, gin.H{
			"status": "success",
			"value":  value,
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "success", response["status"])
	assert.Equal(t, "test_value", response["value"])
}

func TestResponseMiddleware_EmptyResponse(t *testing.T) {
	router := setupResponseTest()

	router.Use(ResponseMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
}

func TestResponseMiddleware_MultipleRequests(t *testing.T) {
	router := setupResponseTest()

	router.Use(ResponseMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})

	// First request
	req1 := httptest.NewRequest(http.MethodGet, "/test", nil)
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	assert.Equal(t, http.StatusOK, w1.Code)
	assert.Equal(t, "application/json", w1.Header().Get("Content-Type"))

	// Second request
	req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusOK, w2.Code)
	assert.Equal(t, "application/json", w2.Header().Get("Content-Type"))
}
