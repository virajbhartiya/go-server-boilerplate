package middleware

import (
	"net/http"

	corspkg "github.com/rs/cors"
)

// Cors wraps rs/cors with allowed origins
func Cors(next http.Handler, allowedOrigins []string) http.Handler {
	c := corspkg.New(corspkg.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Origin", "Content-Type", "Content-Length", "Accept", "Accept-Encoding", "Authorization", "X-Request-ID"},
		ExposedHeaders:   []string{"Content-Length", "Content-Type"},
		AllowCredentials: true,
	})
	return c.Handler(next)
}
