package middleware

import (
	"context"
	"go-server-boilerplate/internal/infrastructure/auth"
	"net/http"
	"strings"
)

// AuthMiddleware represents the authentication middleware
type AuthMiddleware struct {
	jwtManager *auth.JWTManager
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(jwtManager *auth.JWTManager) *AuthMiddleware {
	return &AuthMiddleware{jwtManager: jwtManager}
}

type contextKey string

const (
	contextKeyUserID contextKey = "userID"
	contextKeyRole   contextKey = "role"
)

// AuthRequiredMiddleware validates JWT and injects claims into context
func (m *AuthMiddleware) AuthRequiredMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "missing authorization header", http.StatusUnauthorized)
			return
		}
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "invalid authorization format", http.StatusUnauthorized)
			return
		}
		claims, err := m.jwtManager.ValidateToken(parts[1])
		if err != nil {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), contextKeyUserID, claims.UserID)
		ctx = context.WithValue(ctx, contextKeyRole, claims.Role)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RoleRequiredMiddleware enforces one of the given roles
func (m *AuthMiddleware) RoleRequiredMiddleware(next http.Handler, roles ...string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		roleVal := r.Context().Value(contextKeyRole)
		role, _ := roleVal.(string)
		allowed := false
		for _, rr := range roles {
			if rr == role {
				allowed = true
				break
			}
		}
		if !allowed {
			http.Error(w, "insufficient permissions", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// ExtractUserIDFromContext returns user id from context
func ExtractUserIDFromContext(ctx context.Context) (uint, bool) {
	v := ctx.Value(contextKeyUserID)
	if v == nil {
		return 0, false
	}
	id, ok := v.(uint)
	return id, ok
}

// ExtractRoleFromContext returns role from context
func ExtractRoleFromContext(ctx context.Context) (string, bool) {
	v := ctx.Value(contextKeyRole)
	if v == nil {
		return "", false
	}
	s, ok := v.(string)
	return s, ok
}
