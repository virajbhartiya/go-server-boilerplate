package middleware

import (
	"go-server-boilerplate/internal/pkg/logger"
	"net/http"
	"runtime/debug"
	"strings"

	"go.uber.org/zap"
)

// RecoveryMiddleware recovers from panics and logs the error
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				stack := strings.Split(string(debug.Stack()), "\n")
				cleanStack := []string{}
				for i := 3; i < len(stack); i++ {
					if i%2 == 1 && len(stack[i]) > 0 && !strings.HasPrefix(stack[i], "runtime/") {
						cleanStack = append(cleanStack, strings.TrimSpace(stack[i]))
					}
				}
				logger.Error("Panic recovered",
					zap.Any("error", err),
					zap.String("stack", strings.Join(cleanStack, " > ")),
				)
				requestID := GetRequestIDFromContext(r.Context())
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte("{\"error\":\"Internal Server Error\",\"request_id\":\"" + requestID + "\"}"))
			}
		}()
		next.ServeHTTP(w, r)
	})
}
