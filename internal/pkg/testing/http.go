package testing

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// TestHTTPResponse represents a test HTTP response
type TestHTTPResponse struct {
	Code int
	Body map[string]interface{}
	Raw  []byte
}

// SetupRouter creates a new Gin router for testing
func SetupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

// MockJSONRequest performs a mock JSON request to a Gin handler
func MockJSONRequest(t *testing.T, router *gin.Engine, method, url string, body interface{}) *TestHTTPResponse {
	// Set up request
	var reqBody io.Reader
	if body != nil {
		jsonBytes, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("Failed to marshal request body: %v", err)
		}
		reqBody = bytes.NewBuffer(jsonBytes)
	}

	// Create request
	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	w := httptest.NewRecorder()

	// Perform request
	router.ServeHTTP(w, req)

	// Parse response
	response := &TestHTTPResponse{
		Code: w.Code,
		Raw:  w.Body.Bytes(),
	}

	// Try to parse the body as JSON
	if len(w.Body.Bytes()) > 0 {
		response.Body = make(map[string]interface{})
		err = json.Unmarshal(w.Body.Bytes(), &response.Body)
		if err != nil {
			t.Logf("Warning: Failed to unmarshal response body as JSON: %v", err)
		}
	}

	return response
}

// MockJSONRequestWithHeaders performs a mock JSON request with custom headers
func MockJSONRequestWithHeaders(t *testing.T, router *gin.Engine, method, url string, body interface{}, headers map[string]string) *TestHTTPResponse {
	// Set up request
	var reqBody io.Reader
	if body != nil {
		jsonBytes, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("Failed to marshal request body: %v", err)
		}
		reqBody = bytes.NewBuffer(jsonBytes)
	}

	// Create request
	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Add custom headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Create response recorder
	w := httptest.NewRecorder()

	// Perform request
	router.ServeHTTP(w, req)

	// Parse response
	response := &TestHTTPResponse{
		Code: w.Code,
		Raw:  w.Body.Bytes(),
	}

	// Try to parse the body as JSON
	if len(w.Body.Bytes()) > 0 {
		response.Body = make(map[string]interface{})
		err = json.Unmarshal(w.Body.Bytes(), &response.Body)
		if err != nil {
			t.Logf("Warning: Failed to unmarshal response body as JSON: %v", err)
		}
	}

	return response
}

// AssertStatusCode asserts that the status code matches the expected value
func AssertStatusCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected status code %d, got %d", expected, actual)
	}
}
