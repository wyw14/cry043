package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RequestBoundary installs the invariants shared by every offline API call:
// traceability, an I/O deadline, a fixed local origin and defensive headers.
func RequestBoundary(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := newRequestID()
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("Referrer-Policy", "no-referrer")
		c.Header("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")
		c.Header("Access-Control-Allow-Origin", "http://localhost:5173")
		c.Header("Access-Control-Allow-Headers", "Content-Type,X-Actor-ID,X-Actor-Role,X-Team-ID")
		c.Header("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func RecoverToJSON(log *zap.Logger) gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered any) {
		log.Error("recovered compliance API panic", zap.Any("cause", recovered), zap.String("request_id", c.GetString("request_id")))
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"code": "UNEXPECTED_FAILURE", "message": "内部错误", "field_errors": map[string]string{}, "request_id": c.GetString("request_id"),
		})
	})
}

func newRequestID() string {
	random := make([]byte, 10)
	if _, err := rand.Read(random); err != nil {
		return "req-" + hex.EncodeToString([]byte(time.Now().UTC().Format(time.RFC3339Nano)))
	}
	return "req-" + hex.EncodeToString(random)
}
