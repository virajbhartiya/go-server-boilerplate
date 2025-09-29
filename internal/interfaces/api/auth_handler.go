package api

import (
	"encoding/json"
	"net/http"
	"time"

	"errors"
	"go-server-boilerplate/internal/app/ports"
	"go-server-boilerplate/internal/infrastructure/auth"
	"go-server-boilerplate/internal/infrastructure/database/models"
	apperrs "go-server-boilerplate/internal/pkg/errors"
	"go-server-boilerplate/internal/pkg/logger"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// AuthHandler handles authentication-related HTTP requests
type AuthHandler struct {
	userService ports.Service[models.User]
	jwtManager  *auth.JWTManager
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(userService ports.Service[models.User], jwtManager *auth.JWTManager) *AuthHandler {
	return &AuthHandler{
		userService: userService,
		jwtManager:  jwtManager,
	}
}

// LoginRequest represents the login request
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// LoginResponse represents the login response
type LoginResponse struct {
	Token        string       `json:"token"`
	RefreshToken string       `json:"refresh_token,omitempty"`
	ExpiresAt    time.Time    `json:"expires_at"`
	User         UserResponse `json:"user"`
}

// RegisterRequest represents the registration request
type RegisterRequest struct {
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=8"`
	FirstName string `json:"first_name" validate:"required"`
	LastName  string `json:"last_name" validate:"required"`
}

// RegisterAuthRoutes registers authentication routes
func (h *AuthHandler) RegisterAuthRoutes(router *mux.Router) {
	api := router.PathPrefix("/api/v1/auth").Subrouter()

	api.HandleFunc("/login", h.Login).Methods(http.MethodPost)
	api.HandleFunc("/register", h.Register).Methods(http.MethodPost)
	api.HandleFunc("/refresh", h.RefreshToken).Methods(http.MethodPost)
}

// Login godoc
// @Summary User login
// @Description Authenticate user and return JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param credentials body LoginRequest true "Login credentials"
// @Success 200 {object} LoginResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Find user by email
	// Note: This is a simplified implementation. In a real app, you'd have a FindByEmail method
	users, _, err := h.userService.List(r.Context(), 1, 1)
	if err != nil {
		logger.Error("Failed to query users", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	var user *models.User
	for _, u := range users {
		if u.Email == req.Email {
			user = &u
			break
		}
	}

	if user == nil || !user.CheckPassword(req.Password) {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	if !user.Active {
		http.Error(w, "Account is deactivated", http.StatusUnauthorized)
		return
	}

	// Generate JWT token
	token, err := h.jwtManager.GenerateToken(user.ID, user.Role)
	if err != nil {
		logger.Error("Failed to generate token", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Update last login
	user.UpdateLastLogin()
	if err := h.userService.Update(r.Context(), *user); err != nil {
		logger.Warn("Failed to update last login", zap.Error(err))
	}

	// Calculate expiration time
	expiresAt := time.Now().Add(time.Duration(24) * time.Hour) // Default 24 hours

	response := LoginResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		User: UserResponse{
			ID:        user.ID,
			Email:     user.Email,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Role:      user.Role,
			Active:    user.Active,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Register godoc
// @Summary User registration
// @Description Register a new user account
// @Tags auth
// @Accept json
// @Produce json
// @Param user body RegisterRequest true "User registration information"
// @Success 201 {object} UserResponse
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/auth/register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Check if user already exists
	users, _, err := h.userService.List(r.Context(), 1, 1)
	if err != nil {
		logger.Error("Failed to query users", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	for _, u := range users {
		if u.Email == req.Email {
			http.Error(w, "User already exists", http.StatusConflict)
			return
		}
	}

	// Create new user
	user := models.User{
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Role:      "user",
		Active:    true,
	}

	// Hash password
	if err := user.SetPassword(req.Password); err != nil {
		logger.Error("Failed to hash password", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Create user
	if err := h.userService.Create(r.Context(), user); err != nil {
		logger.Error("Failed to create user", zap.Error(err))
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	response := UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Role:      user.Role,
		Active:    user.Active,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// RefreshToken godoc
// @Summary Refresh JWT token
// @Description Refresh an expired JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param token body map[string]string true "Refresh token request"
// @Success 200 {object} LoginResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/auth/refresh [post]
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req map[string]string
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	tokenStr, exists := req["token"]
	if !exists || tokenStr == "" {
		http.Error(w, "Token is required", http.StatusBadRequest)
		return
	}

	// Validate token
	claims, err := h.jwtManager.ValidateToken(tokenStr)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	// Get user
	user, err := h.userService.GetByID(r.Context(), claims.UserID)
	if err != nil {
		if errors.Is(err, apperrs.ErrNotFound) {
			http.Error(w, "User not found", http.StatusUnauthorized)
			return
		}
		logger.Error("Failed to get user", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if !user.Active {
		http.Error(w, "Account is deactivated", http.StatusUnauthorized)
		return
	}

	// Generate new token
	newToken, err := h.jwtManager.GenerateToken(user.ID, user.Role)
	if err != nil {
		logger.Error("Failed to generate token", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	expiresAt := time.Now().Add(time.Duration(24) * time.Hour)

	response := LoginResponse{
		Token:     newToken,
		ExpiresAt: expiresAt,
		User: UserResponse{
			ID:        user.ID,
			Email:     user.Email,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Role:      user.Role,
			Active:    user.Active,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
