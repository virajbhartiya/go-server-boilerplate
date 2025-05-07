package api_test

import (
	"net/http"
	"testing"

	"go-server-boilerplate/internal/interfaces/api"
	testing_utils "go-server-boilerplate/internal/pkg/testing"

	"github.com/gin-gonic/gin"
)

func TestHealthCheck(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := testing_utils.SetupRouter()
	api.RegisterHealthRoutes(router)

	// Test health check endpoint
	t.Run("health check returns 200 OK", func(t *testing.T) {
		// Perform request
		response := testing_utils.MockJSONRequest(t, router, http.MethodGet, "/health", nil)

		// Assertions
		testing_utils.AssertStatusCode(t, http.StatusOK, response.Code)

		// Check body values
		if response.Body["status"] != "ok" {
			t.Errorf("Expected status 'ok', got %v", response.Body["status"])
		}

		if response.Body["version"] == nil {
			t.Error("Expected version in response")
		}

		if response.Body["timestamp"] == nil {
			t.Error("Expected timestamp in response")
		}

		// Check services array
		services, ok := response.Body["services"].([]interface{})
		if !ok {
			t.Error("Expected services array in response")
		} else {
			if len(services) == 0 {
				t.Error("Expected non-empty services array")
			}
		}
	})
}
