package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	// RequestIDHeader is the header key for the request ID
	RequestIDHeader = "X-Request-ID"

	// RequestIDContextKey is the context key for the request ID
	RequestIDContextKey = "request_id"
)

// RequestID middleware adds a unique request ID to each request
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if request ID already exists in the request header
		requestID := c.GetHeader(RequestIDHeader)

		// If not, generate a new one
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Set request ID in the context and response header
		c.Set(RequestIDContextKey, requestID)
		c.Header(RequestIDHeader, requestID)

		c.Next()
	}
}

// GetRequestID returns the request ID from the context
func GetRequestID(c *gin.Context) string {
	if requestID, exists := c.Get(RequestIDContextKey); exists {
		if id, ok := requestID.(string); ok {
			return id
		}
	}
	return ""
}
