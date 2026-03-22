package middleware

import (
	"crypto/rand"
	"fmt"
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		requestID := generateRequestID()
		c.Set("request_id", requestID)

		slog.Info("request incoming",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"request_id", requestID,
		)

		c.Next()

		slog.Info("request completed",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"duration_ms", time.Since(start).Milliseconds(),
			"request_id", requestID,
		)
	}
}

func generateRequestID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}
