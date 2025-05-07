package middleware

import (
	"net/http"
	"runtime/debug"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"go-server-boilerplate/internal/pkg/logger"
)

// Recovery middleware recovers from panics and logs the error
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Clean the stack trace
				stack := strings.Split(string(debug.Stack()), "\n")
				cleanStack := []string{}
				for i := 3; i < len(stack); i++ {
					if i%2 == 1 && len(stack[i]) > 0 && !strings.HasPrefix(stack[i], "runtime/") {
						cleanStack = append(cleanStack, strings.TrimSpace(stack[i]))
					}
				}

				// Log the error
				logger.Error("Panic recovered",
					zap.Any("error", err),
					zap.String("stack", strings.Join(cleanStack, " > ")),
				)

				// Get the request ID if available
				var requestID string
				if id, exists := c.Get(RequestIDContextKey); exists {
					if reqID, ok := id.(string); ok {
						requestID = reqID
					}
				}

				// Return a 500 error
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error":      "Internal Server Error",
					"request_id": requestID,
				})
			}
		}()
		c.Next()
	}
}
