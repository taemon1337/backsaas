package gateway

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGatewayIntegration tests the complete gateway functionality with real route configurations
func TestGatewayIntegration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create test backend servers
	healthBackend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Health dashboard expects stripped paths
		assert.Equal(t, "/api/status", r.URL.Path, "Health backend should receive stripped path")
		
		// Check for added headers
		assert.Equal(t, "system-health", r.Header.Get("X-Interface-Type"), "Should have interface type header")
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":    "healthy",
			"timestamp": time.Now().Format(time.RFC3339),
		})
	}))
	defer healthBackend.Close()

	platformBackend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Platform API expects full paths
		assert.Equal(t, "/api/platform/admin/login", r.URL.Path, "Platform backend should receive full path")
		
		// Check for forwarded headers
		assert.NotEmpty(t, r.Header.Get("X-Forwarded-For"), "Should have forwarded headers")
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"token": "test-jwt-token",
			"user":  map[string]string{"id": "123", "email": "admin@test.com"},
		})
	}))
	defer platformBackend.Close()

	// Create gateway configuration with test backends
	config := &Config{
		Environment: "test",
		Port:        "8080", // Add required port
		JWTSecret:   "test-secret-for-integration-test", // Add required JWT secret
		RedisURL:    "redis://localhost:6379", // Add Redis URL for test
		Routes: []RouteConfig{
			{
				Description: "Health Dashboard API",
				PathPrefix:  "/api/system-health",
				Backend: BackendConfig{
					URL:     healthBackend.URL,
					Timeout: 30 * time.Second,
				},
				Transform: &TransformConfig{
					StripPrefix: true,
					AddHeaders: map[string]string{
						"X-Interface-Type": "system-health",
					},
				},
				Enabled: true,
			},
			{
				Description: "Platform API",
				PathPrefix:  "/api/platform",
				Backend: BackendConfig{
					URL:     platformBackend.URL,
					Timeout: 30 * time.Second,
				},
				Transform: &TransformConfig{
					StripPrefix: false, // Keep full path
				},
				Enabled: true,
			},
		},
		Auth: AuthConfig{
			Enabled: false, // Disable auth for integration test
		},
		RateLimit: RateLimitConfig{
			Enabled: false, // Disable rate limiting for test
		},
		Monitoring: MonitoringConfig{
			Enabled: false, // Disable monitoring for test
		},
	}

	// Create gateway instance
	gateway, err := NewGateway(config)
	require.NoError(t, err)

	// Create test server
	testServer := httptest.NewServer(gateway.router)
	defer testServer.Close()

	tests := []struct {
		name           string
		method         string
		path           string
		body           string
		expectedStatus int
		checkResponse  func(t *testing.T, body string)
	}{
		{
			name:           "health dashboard with path stripping",
			method:         "GET",
			path:           "/api/system-health/api/status",
			expectedStatus: 200,
			checkResponse: func(t *testing.T, body string) {
				var response map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				require.NoError(t, err)
				assert.Equal(t, "healthy", response["status"])
				assert.NotEmpty(t, response["timestamp"])
			},
		},
		{
			name:           "platform API without path stripping",
			method:         "POST",
			path:           "/api/platform/admin/login",
			body:           `{"email":"admin@test.com","password":"test123"}`,
			expectedStatus: 200,
			checkResponse: func(t *testing.T, body string) {
				var response map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				require.NoError(t, err)
				assert.Equal(t, "test-jwt-token", response["token"])
				assert.NotNil(t, response["user"])
			},
		},
		{
			name:           "no matching route returns 404",
			method:         "GET",
			path:           "/api/unknown/endpoint",
			expectedStatus: 404,
			checkResponse: func(t *testing.T, body string) {
				assert.Contains(t, body, "route not found")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request
			var err error

			if tt.body != "" {
				req, err = http.NewRequest(tt.method, testServer.URL+tt.path, strings.NewReader(tt.body))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req, err = http.NewRequest(tt.method, testServer.URL+tt.path, nil)
			}
			require.NoError(t, err)

			client := &http.Client{Timeout: 10 * time.Second}
			resp, err := client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode, "Status code should match expected")

			if tt.checkResponse != nil {
				body := make([]byte, 1024)
				n, _ := resp.Body.Read(body)
				tt.checkResponse(t, string(body[:n]))
			}
		})
	}
}

// TestRouteConfigurationValidation tests that route configurations are properly validated
func TestRouteConfigurationValidation(t *testing.T) {
	tests := []struct {
		name          string
		config        *Config
		expectedError bool
		errorContains string
	}{
		{
			name: "valid configuration",
			config: &Config{
				Environment: "test",
				Port:        "8080",
				JWTSecret:   "test-secret",
				Routes: []RouteConfig{
					{
						Description: "Test API",
						PathPrefix:  "/api/test",
						Backend:     BackendConfig{URL: "http://test:8080"},
						Enabled:     true,
					},
				},
				Auth:       AuthConfig{Enabled: false},
				RateLimit:  RateLimitConfig{Enabled: false},
				Monitoring: MonitoringConfig{Enabled: false},
			},
			expectedError: false,
		},
		{
			name: "missing JWT secret when auth enabled",
			config: &Config{
				Environment: "test",
				Port:        "8080",
				JWTSecret:   "", // Empty JWT secret should cause error
				Routes: []RouteConfig{
					{
						Description: "Test API",
						PathPrefix:  "/api/test",
						Backend:     BackendConfig{URL: "http://test:8080"},
						Enabled:     true,
					},
				},
				Auth:       AuthConfig{Enabled: true},
				RateLimit:  RateLimitConfig{Enabled: false},
				Monitoring: MonitoringConfig{Enabled: false},
			},
			expectedError: true,
			errorContains: "jwt_secret is required",
		},
		{
			name: "invalid backend URL",
			config: &Config{
				Environment: "test",
				Port:        "8080",
				JWTSecret:   "test-secret",
				Routes: []RouteConfig{
					{
						Description: "Test API",
						PathPrefix:  "/api/test",
						Backend:     BackendConfig{URL: "invalid-url"},
						Enabled:     true,
					},
				},
				Auth:       AuthConfig{Enabled: false},
				RateLimit:  RateLimitConfig{Enabled: false},
				Monitoring: MonitoringConfig{Enabled: false},
			},
			expectedError: true,
			errorContains: "invalid backend URL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewGateway(tt.config)

			if tt.expectedError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestTransformationFeatures tests all transformation features work correctly
func TestTransformationFeatures(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create test backend that validates transformations
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check path transformation
		expectedPath := r.Header.Get("X-Expected-Path")
		if expectedPath != "" {
			assert.Equal(t, expectedPath, r.URL.Path, "Path should be transformed correctly")
		}

		// Check added headers
		expectedHeader := r.Header.Get("X-Expected-Header")
		if expectedHeader != "" {
			parts := strings.Split(expectedHeader, ":")
			if len(parts) == 2 {
				assert.Equal(t, parts[1], r.Header.Get(parts[0]), "Header should be added correctly")
			}
		}

		// Check removed headers
		removedHeader := r.Header.Get("X-Should-Be-Removed")
		assert.Empty(t, removedHeader, "Header should be removed")

		w.WriteHeader(200)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer backend.Close()

	config := &Config{
		Environment: "test",
		Port:        "8080",
		Routes: []RouteConfig{
			{
				Description: "Transform Test API",
				PathPrefix:  "/api/transform",
				Backend:     BackendConfig{URL: backend.URL},
				Transform: &TransformConfig{
					StripPrefix: true,
					AddHeaders: map[string]string{
						"X-Gateway":      "backsaas",
						"X-Service-Type": "test",
					},
					RemoveHeaders: []string{"X-Internal-Token", "X-Debug"},
				},
				Enabled: true,
			},
		},
		Auth:       AuthConfig{Enabled: false},
		RateLimit:  RateLimitConfig{Enabled: false},
		Monitoring: MonitoringConfig{Enabled: false},
	}

	gateway, err := NewGateway(config)
	require.NoError(t, err)

	testServer := httptest.NewServer(gateway.router)
	defer testServer.Close()

	// Test request with headers that should be transformed
	req, err := http.NewRequest("GET", testServer.URL+"/api/transform/test/endpoint", nil)
	require.NoError(t, err)

	// Add headers that should be removed
	req.Header.Set("X-Internal-Token", "secret")
	req.Header.Set("X-Debug", "true")

	// Add headers to validate transformations
	req.Header.Set("X-Expected-Path", "/test/endpoint")
	req.Header.Set("X-Expected-Header", "X-Gateway:backsaas")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 200, resp.StatusCode, "Request should succeed")
}

// TestRouteMatching tests that routes are matched in the correct priority order
func TestRouteMatching(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create backends that identify themselves
	genericBackend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"backend":"generic"}`))
	}))
	defer genericBackend.Close()

	specificBackend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"backend":"specific"}`))
	}))
	defer specificBackend.Close()

	config := &Config{
		Environment: "test",
		Port:        "8080",
		JWTSecret:   "test-secret",
		Routes: []RouteConfig{
			{
				Description: "Generic API",
				PathPrefix:  "/api",
				Backend:     BackendConfig{URL: genericBackend.URL},
				Enabled:     true,
			},
			{
				Description: "Specific API",
				PathPrefix:  "/api/specific",
				Backend:     BackendConfig{URL: specificBackend.URL},
				Enabled:     true,
			},
		},
		Auth:       AuthConfig{Enabled: false},
		RateLimit:  RateLimitConfig{Enabled: false},
		Monitoring: MonitoringConfig{Enabled: false},
	}

	gateway, err := NewGateway(config)
	require.NoError(t, err)

	testServer := httptest.NewServer(gateway.router)
	defer testServer.Close()

	tests := []struct {
		path            string
		expectedBackend string
	}{
		{"/api/specific/test", "specific"}, // Should match more specific route
		{"/api/generic/test", "generic"},   // Should match generic route
		{"/api/test", "generic"},           // Should match generic route
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			resp, err := http.Get(testServer.URL + tt.path)
			require.NoError(t, err)
			defer resp.Body.Close()

			body := make([]byte, 1024)
			n, _ := resp.Body.Read(body)

			var response map[string]string
			err = json.Unmarshal(body[:n], &response)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedBackend, response["backend"], "Should route to correct backend")
		})
	}
}
