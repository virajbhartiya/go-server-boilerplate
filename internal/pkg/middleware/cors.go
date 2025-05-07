package middleware

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CorsMiddleware configures CORS for the application
func CorsMiddleware(allowedOrigins []string) gin.HandlerFunc {
	// Default configuration
	config := cors.DefaultConfig()

	// Set allowed origins
	if len(allowedOrigins) == 1 && allowedOrigins[0] == "*" {
		config.AllowAllOrigins = true
	} else {
		config.AllowOrigins = allowedOrigins
	}

	// Allow common methods and headers
	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{
		"Origin",
		"Content-Type",
		"Content-Length",
		"Accept",
		"Accept-Encoding",
		"Authorization",
		"X-Request-ID",
	}

	// Allow credentials and expose headers
	config.AllowCredentials = true
	config.ExposeHeaders = []string{"Content-Length", "Content-Type"}

	return cors.New(config)
}
