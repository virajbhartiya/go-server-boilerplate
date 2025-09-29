package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"errors"
	"go-server-boilerplate/internal/app/ports"
	"go-server-boilerplate/internal/infrastructure/database/models"
	apperrs "go-server-boilerplate/internal/pkg/errors"
	"go-server-boilerplate/internal/pkg/logger"
	"go-server-boilerplate/internal/pkg/middleware"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// UserHandler handles user-related HTTP requests
type UserHandler struct {
	userService ports.Service[models.User]
}

// NewUserHandler creates a new user handler
func NewUserHandler(userService ports.Service[models.User]) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// CreateUserRequest represents the request to create a user
type CreateUserRequest struct {
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=8"`
	FirstName string `json:"first_name" validate:"required"`
	LastName  string `json:"last_name" validate:"required"`
}

// UpdateUserRequest represents the request to update a user
type UpdateUserRequest struct {
	FirstName *string `json:"first_name,omitempty"`
	LastName  *string `json:"last_name,omitempty"`
	Active    *bool   `json:"active,omitempty"`
}

// UserResponse represents the user response
type UserResponse struct {
	ID        uint   `json:"id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Role      string `json:"role"`
	Active    bool   `json:"active"`
}

// ListUsersResponse represents the response for listing users
type ListUsersResponse struct {
	Users      []UserResponse `json:"users"`
	Total      int64          `json:"total"`
	Page       int            `json:"page"`
	PageSize   int            `json:"page_size"`
	TotalPages int            `json:"total_pages"`
}

// RegisterUserRoutes registers user routes
func (h *UserHandler) RegisterUserRoutes(router *mux.Router, authMiddleware *middleware.AuthMiddleware) {
	api := router.PathPrefix("/api/v1/users").Subrouter()

	// Public routes
	api.HandleFunc("", h.CreateUser).Methods(http.MethodPost)

	// Protected routes
	api.Use(authMiddleware.AuthRequiredMiddleware)
	api.HandleFunc("", h.ListUsers).Methods(http.MethodGet)
	api.HandleFunc("/{id:[0-9]+}", h.GetUser).Methods(http.MethodGet)
	api.HandleFunc("/{id:[0-9]+}", h.UpdateUser).Methods(http.MethodPut, http.MethodPatch)
	api.HandleFunc("/{id:[0-9]+}", h.DeleteUser).Methods(http.MethodDelete)
}

// CreateUser godoc
// @Summary Create a new user
// @Description Create a new user with the provided information
// @Tags users
// @Accept json
// @Produce json
// @Param user body CreateUserRequest true "User information"
// @Success 201 {object} UserResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/users [post]
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Create user model
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

	// Return created user
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

// GetUser godoc
// @Summary Get user by ID
// @Description Get a user by their ID
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} UserResponse
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /api/v1/users/{id} [get]
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	user, err := h.userService.GetByID(r.Context(), uint(id))
	if err != nil {
		if errors.Is(err, apperrs.ErrNotFound) {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		logger.Error("Failed to get user", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
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
	json.NewEncoder(w).Encode(response)
}

// ListUsers godoc
// @Summary List users
// @Description Get a paginated list of users
// @Tags users
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(10)
// @Success 200 {object} ListUsersResponse
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /api/v1/users [get]
func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	page := 1
	pageSize := 10

	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if pageSizeStr := r.URL.Query().Get("page_size"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 && ps <= 100 {
			pageSize = ps
		}
	}

	users, total, err := h.userService.List(r.Context(), page, pageSize)
	if err != nil {
		logger.Error("Failed to list users", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Convert to response format
	userResponses := make([]UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = UserResponse{
			ID:        user.ID,
			Email:     user.Email,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Role:      user.Role,
			Active:    user.Active,
		}
	}

	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))

	response := ListUsersResponse{
		Users:      userResponses,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UpdateUser godoc
// @Summary Update user
// @Description Update user information
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param user body UpdateUserRequest true "User update information"
// @Success 200 {object} UserResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /api/v1/users/{id} [put]
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var req UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Get existing user
	user, err := h.userService.GetByID(r.Context(), uint(id))
	if err != nil {
		if errors.Is(err, apperrs.ErrNotFound) {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		logger.Error("Failed to get user", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Update fields if provided
	if req.FirstName != nil {
		user.FirstName = *req.FirstName
	}
	if req.LastName != nil {
		user.LastName = *req.LastName
	}
	if req.Active != nil {
		user.Active = *req.Active
	}

	// Update user
	if err := h.userService.Update(r.Context(), user); err != nil {
		logger.Error("Failed to update user", zap.Error(err))
		http.Error(w, "Failed to update user", http.StatusInternalServerError)
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
	json.NewEncoder(w).Encode(response)
}

// DeleteUser godoc
// @Summary Delete user
// @Description Delete a user by ID
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Success 204
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /api/v1/users/{id} [delete]
func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	if err := h.userService.Delete(r.Context(), uint(id)); err != nil {
		if errors.Is(err, apperrs.ErrNotFound) {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		logger.Error("Failed to delete user", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
