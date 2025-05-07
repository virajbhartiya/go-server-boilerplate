package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"go-server-boilerplate/internal/infrastructure/auth"
	"go-server-boilerplate/internal/pkg/errors"
)

// AuthMiddleware represents the authentication middleware
type AuthMiddleware struct {
	jwtManager *auth.JWTManager
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(jwtManager *auth.JWTManager) *AuthMiddleware {
	return &AuthMiddleware{
		jwtManager: jwtManager,
	}
}

// AuthRequired middleware requires valid JWT token
func (m *AuthMiddleware) AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(errors.Unauthorized("missing authorization header").StatusCode, gin.H{
				"error": "missing authorization header",
			})
			return
		}

		// Check Bearer token format
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(errors.Unauthorized("invalid authorization format").StatusCode, gin.H{
				"error": "invalid authorization format",
			})
			return
		}

		// Validate token
		claims, err := m.jwtManager.ValidateToken(parts[1])
		if err != nil {
			c.AbortWithStatusJSON(errors.Unauthorized("invalid token: "+err.Error()).StatusCode, gin.H{
				"error": "invalid token",
			})
			return
		}

		// Store claims in context for later use
		c.Set("userID", claims.UserID)
		c.Set("role", claims.Role)
		c.Set("claims", claims)

		c.Next()
	}
}

// RoleRequired middleware requires specific role
func (m *AuthMiddleware) RoleRequired(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// First apply the AuthRequired middleware
		m.AuthRequired()(c)

		// Check if the request was aborted by the AuthRequired middleware
		if c.IsAborted() {
			return
		}

		// Get the role from context
		role, exists := c.Get("role")
		if !exists {
			c.AbortWithStatusJSON(errors.Forbidden("role not found in token").StatusCode, gin.H{
				"error": "role not found in token",
			})
			return
		}

		// Check if the user has one of the required roles
		roleStr, ok := role.(string)
		if !ok {
			c.AbortWithStatusJSON(errors.Forbidden("invalid role format").StatusCode, gin.H{
				"error": "invalid role format",
			})
			return
		}

		// Check if the user has one of the required roles
		for _, r := range roles {
			if r == roleStr {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(errors.Forbidden("insufficient permissions").StatusCode, gin.H{
			"error": "insufficient permissions",
		})
	}
}

// ExtractUserID extracts the user ID from the context
func ExtractUserID(c *gin.Context) (uint, bool) {
	userID, exists := c.Get("userID")
	if !exists {
		return 0, false
	}
	id, ok := userID.(uint)
	return id, ok
}

// ExtractRole extracts the role from the context
func ExtractRole(c *gin.Context) (string, bool) {
	role, exists := c.Get("role")
	if !exists {
		return "", false
	}
	roleStr, ok := role.(string)
	return roleStr, ok
}
