package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

const (
	RequestIDHeader = "X-Request-ID"
)

type requestIDContextKeyType string

const requestIDContextKey requestIDContextKeyType = "request_id"

// RequestIDMiddleware attaches a request ID to context and response
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get(RequestIDHeader)
		if requestID == "" {
			requestID = uuid.New().String()
		}
		w.Header().Set(RequestIDHeader, requestID)
		ctx := context.WithValue(r.Context(), requestIDContextKey, requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetRequestIDFromContext returns the request ID from context
func GetRequestIDFromContext(ctx context.Context) string {
	if v := ctx.Value(requestIDContextKey); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
