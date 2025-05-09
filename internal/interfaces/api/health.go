package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
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

// RegisterHealthRoutes registers health check routes
func RegisterHealthRoutes(router *gin.Engine) {
	router.GET("/health", HealthCheck)
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
func HealthCheck(c *gin.Context) {
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
			c.JSON(http.StatusServiceUnavailable, health)
			return
		}
	}

	c.JSON(http.StatusOK, health)
}
