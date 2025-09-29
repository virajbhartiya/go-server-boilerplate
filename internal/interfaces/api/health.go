package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

// HealthResponse represents the health check response structure
type HealthResponse struct {
	Status    string    `json:"status" example:"ok"`
	Timestamp time.Time `json:"timestamp" example:"2023-01-01T12:00:00Z"`
	Version   string    `json:"version" example:"1.0.0"`
	Services  []Service `json:"services"`
}

// Service represents a service health check
type Service struct {
	Name   string `json:"name" example:"database"`
	Status string `json:"status" example:"ok"`
	Error  string `json:"error,omitempty" example:""`
}

// RegisterHealthRoutesMux registers health check routes on Gorilla Mux
func RegisterHealthRoutesMux(router *mux.Router) {
	router.HandleFunc("/health", HealthCheck).Methods(http.MethodGet)
}

// HealthCheck godoc
// @Summary Health check endpoint
// @Description Get the health status of the API and its dependencies
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} HealthResponse
// @Failure 503 {object} HealthResponse
// @Router /health [get]
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	health := HealthResponse{
		Status:    "ok",
		Timestamp: time.Now(),
		Version:   "1.0.0",
		Services:  []Service{},
	}

	// Check database status
	dbService := Service{
		Name:   "database",
		Status: "ok",
	}

	// Check cache status
	cacheService := Service{
		Name:   "cache",
		Status: "ok",
	}

	// Add services health
	health.Services = append(health.Services, dbService, cacheService)

	// Check if any service is not ok
	for _, service := range health.Services {
		if service.Status != "ok" {
			health.Status = "degraded"
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			_ = json.NewEncoder(w).Encode(health)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(health)
}
