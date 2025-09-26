package gateway

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProxyMiddleware_PathTransformation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		route          RouteConfig
		requestPath    string
		expectedPath   string
		expectedUpstream string
	}{
		{
			name: "strip_prefix enabled - health dashboard",
			route: RouteConfig{
				PathPrefix: "/api/system-health",
				Backend: BackendConfig{
					URL: "http://health-dashboard:8090",
				},
				Transform: &TransformConfig{
					StripPrefix: true,
				},
			},
			requestPath:      "/api/system-health/api/status",
			expectedPath:     "/api/status",
			expectedUpstream: "http://health-dashboard:8090/api/status",
		},
		{
			name: "strip_prefix disabled - platform API",
			route: RouteConfig{
				PathPrefix: "/api/platform",
				Backend: BackendConfig{
					URL: "http://platform-api:8080",
				},
				Transform: &TransformConfig{
					StripPrefix: false,
				},
			},
			requestPath:      "/api/platform/admin/login",
			expectedPath:     "/api/platform/admin/login",
			expectedUpstream: "http://platform-api:8080/api/platform/admin/login",
		},
		{
			name: "no transform config - keeps full path",
			route: RouteConfig{
				PathPrefix: "/api/platform",
				Backend: BackendConfig{
					URL: "http://platform-api:8080",
				},
				Transform: nil,
			},
			requestPath:      "/api/platform/tenants",
			expectedPath:     "/api/platform/tenants",
			expectedUpstream: "http://platform-api:8080/api/platform/tenants",
		},
		{
			name: "strip_prefix with root path",
			route: RouteConfig{
				PathPrefix: "/api/external",
				Backend: BackendConfig{
					URL: "http://external-service:3000",
				},
				Transform: &TransformConfig{
					StripPrefix: true,
				},
			},
			requestPath:      "/api/external",
			expectedPath:     "/",
			expectedUpstream: "http://external-service:3000/",
		},
		{
			name: "strip_prefix with nested path",
			route: RouteConfig{
				PathPrefix: "/api/v1/service",
				Backend: BackendConfig{
					URL: "http://service:8080",
				},
				Transform: &TransformConfig{
					StripPrefix: true,
				},
			},
			requestPath:      "/api/v1/service/users/123",
			expectedPath:     "/users/123",
			expectedUpstream: "http://service:8080/users/123",
		},
		{
			name: "no path prefix match - no transformation",
			route: RouteConfig{
				PathPrefix: "/api/other",
				Backend: BackendConfig{
					URL: "http://other-service:8080",
				},
				Transform: &TransformConfig{
					StripPrefix: true,
				},
			},
			requestPath:      "/api/different/path",
			expectedPath:     "/api/different/path",
			expectedUpstream: "http://other-service:8080/api/different/path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create proxy middleware
			proxy, err := NewProxyMiddleware()
			require.NoError(t, err)

			// Create test request
			req := httptest.NewRequest("GET", tt.requestPath, nil)
			
			// Create gin context
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			// Set route in context (simulating routing middleware)
			c.Set("route", &tt.route)

			// Call modifyRequest
			proxy.modifyRequest(req, &tt.route, c)

			// Assert path transformation
			assert.Equal(t, tt.expectedPath, req.URL.Path, "Request path should be transformed correctly")

			// Check if transformed_path was set in context for logging
			if tt.route.Transform != nil && tt.route.Transform.StripPrefix && tt.expectedPath != tt.requestPath {
				transformedPath, exists := c.Get("transformed_path")
				assert.True(t, exists, "transformed_path should be set in context")
				assert.Equal(t, tt.expectedPath, transformedPath, "transformed_path in context should match expected")
			}
		})
	}
}

func TestProxyMiddleware_HeaderTransformation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name            string
		route           RouteConfig
		expectedHeaders map[string]string
		removedHeaders  []string
	}{
		{
			name: "add headers",
			route: RouteConfig{
				PathPrefix: "/api/service",
				Backend: BackendConfig{
					URL: "http://service:8080",
				},
				Transform: &TransformConfig{
					AddHeaders: map[string]string{
						"X-Service-Type": "external",
						"X-Version":      "v1",
					},
				},
			},
			expectedHeaders: map[string]string{
				"X-Service-Type": "external",
				"X-Version":      "v1",
			},
		},
		{
			name: "remove headers",
			route: RouteConfig{
				PathPrefix: "/api/service",
				Backend: BackendConfig{
					URL: "http://service:8080",
				},
				Transform: &TransformConfig{
					RemoveHeaders: []string{"X-Internal-Token", "X-Debug"},
				},
			},
			removedHeaders: []string{"X-Internal-Token", "X-Debug"},
		},
		{
			name: "add and remove headers",
			route: RouteConfig{
				PathPrefix: "/api/service",
				Backend: BackendConfig{
					URL: "http://service:8080",
				},
				Transform: &TransformConfig{
					AddHeaders: map[string]string{
						"X-Gateway": "backsaas",
					},
					RemoveHeaders: []string{"X-Internal"},
				},
			},
			expectedHeaders: map[string]string{
				"X-Gateway": "backsaas",
			},
			removedHeaders: []string{"X-Internal"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create proxy middleware
			proxy, err := NewProxyMiddleware()
			require.NoError(t, err)

			// Create test request with some initial headers
			req := httptest.NewRequest("GET", "/api/service/test", nil)
			req.Header.Set("X-Internal-Token", "secret")
			req.Header.Set("X-Debug", "true")
			req.Header.Set("X-Internal", "value")

			// Create gin context
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Set("route", &tt.route)

			// Call modifyRequest
			proxy.modifyRequest(req, &tt.route, c)

			// Check added headers
			for key, expectedValue := range tt.expectedHeaders {
				actualValue := req.Header.Get(key)
				assert.Equal(t, expectedValue, actualValue, "Header %s should be added with correct value", key)
			}

			// Check removed headers
			for _, headerName := range tt.removedHeaders {
				actualValue := req.Header.Get(headerName)
				assert.Empty(t, actualValue, "Header %s should be removed", headerName)
			}

			// Check standard forwarded headers are always added
			assert.NotEmpty(t, req.Header.Get("X-Forwarded-For"), "X-Forwarded-For should be set")
			assert.NotEmpty(t, req.Header.Get("X-Forwarded-Proto"), "X-Forwarded-Proto should be set")
			assert.NotEmpty(t, req.Header.Get("X-Forwarded-Host"), "X-Forwarded-Host should be set")
		})
	}
}

func TestProxyMiddleware_AuthHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		userID         string
		tenantID       string
		expectedHeaders map[string]string
	}{
		{
			name:     "with user and tenant context",
			userID:   "user123",
			tenantID: "tenant456",
			expectedHeaders: map[string]string{
				"X-User-ID":   "user123",
				"X-Tenant-ID": "tenant456",
			},
		},
		{
			name:   "with user context only",
			userID: "user789",
			expectedHeaders: map[string]string{
				"X-User-ID": "user789",
			},
		},
		{
			name:     "with tenant context only",
			tenantID: "tenant999",
			expectedHeaders: map[string]string{
				"X-Tenant-ID": "tenant999",
			},
		},
		{
			name:            "no auth context",
			expectedHeaders: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create proxy middleware
			proxy, err := NewProxyMiddleware()
			require.NoError(t, err)

			// Create test request
			req := httptest.NewRequest("GET", "/api/test", nil)

			// Create gin context
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			// Set auth context
			if tt.userID != "" {
				c.Set("user_id", tt.userID)
			}
			if tt.tenantID != "" {
				c.Set("tenant_id", tt.tenantID)
			}

			route := &RouteConfig{
				Backend: BackendConfig{URL: "http://test:8080"},
			}

			// Call modifyRequest
			proxy.modifyRequest(req, route, c)

			// Check auth headers
			for key, expectedValue := range tt.expectedHeaders {
				actualValue := req.Header.Get(key)
				assert.Equal(t, expectedValue, actualValue, "Auth header %s should be set correctly", key)
			}

			// Check that headers are not set when context is missing
			if tt.userID == "" {
				assert.Empty(t, req.Header.Get("X-User-ID"), "X-User-ID should not be set without user context")
			}
			if tt.tenantID == "" {
				assert.Empty(t, req.Header.Get("X-Tenant-ID"), "X-Tenant-ID should not be set without tenant context")
			}
		})
	}
}

func TestProxyMiddleware_ComplexTransformation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Test complex scenario with multiple transformations
	route := RouteConfig{
		PathPrefix: "/api/complex",
		Backend: BackendConfig{
			URL: "http://complex-service:8080",
		},
		Transform: &TransformConfig{
			StripPrefix: true,
			AddHeaders: map[string]string{
				"X-Gateway":      "backsaas",
				"X-Service-Type": "complex",
			},
			RemoveHeaders: []string{"X-Internal-Debug"},
		},
	}

	// Create proxy middleware
	proxy, err := NewProxyMiddleware()
	require.NoError(t, err)

	// Create test request
	req := httptest.NewRequest("POST", "/api/complex/users/create", nil)
	req.Header.Set("X-Internal-Debug", "true")
	req.Header.Set("Content-Type", "application/json")

	// Create gin context
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("route", &route)
	c.Set("user_id", "admin123")
	c.Set("tenant_id", "tenant456")

	// Call modifyRequest
	proxy.modifyRequest(req, &route, c)

	// Assert path transformation
	assert.Equal(t, "/users/create", req.URL.Path, "Path should be stripped correctly")

	// Assert headers added
	assert.Equal(t, "backsaas", req.Header.Get("X-Gateway"), "X-Gateway header should be added")
	assert.Equal(t, "complex", req.Header.Get("X-Service-Type"), "X-Service-Type header should be added")

	// Assert headers removed
	assert.Empty(t, req.Header.Get("X-Internal-Debug"), "X-Internal-Debug header should be removed")

	// Assert original headers preserved
	assert.Equal(t, "application/json", req.Header.Get("Content-Type"), "Original headers should be preserved")

	// Assert auth headers added
	assert.Equal(t, "admin123", req.Header.Get("X-User-ID"), "X-User-ID should be set from context")
	assert.Equal(t, "tenant456", req.Header.Get("X-Tenant-ID"), "X-Tenant-ID should be set from context")

	// Assert standard forwarded headers
	assert.NotEmpty(t, req.Header.Get("X-Forwarded-For"), "X-Forwarded-For should be set")
	assert.NotEmpty(t, req.Header.Get("X-Forwarded-Proto"), "X-Forwarded-Proto should be set")
	assert.NotEmpty(t, req.Header.Get("X-Forwarded-Host"), "X-Forwarded-Host should be set")

	// Assert transformed path is stored in context
	transformedPath, exists := c.Get("transformed_path")
	require.True(t, exists, "transformed_path should be set in context")
	assert.Equal(t, "/users/create", transformedPath, "transformed_path should match expected")
}

func TestProxyMiddleware_EdgeCases(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name        string
		route       RouteConfig
		requestPath string
		expectedPath string
	}{
		{
			name: "strip_prefix with trailing slash",
			route: RouteConfig{
				PathPrefix: "/api/service/",
				Backend: BackendConfig{URL: "http://service:8080"},
				Transform: &TransformConfig{StripPrefix: true},
			},
			requestPath:  "/api/service/test",
			expectedPath: "/test", // Should not strip trailing slash from prefix
		},
		{
			name: "strip_prefix with query parameters",
			route: RouteConfig{
				PathPrefix: "/api/service",
				Backend: BackendConfig{URL: "http://service:8080"},
				Transform: &TransformConfig{StripPrefix: true},
			},
			requestPath:  "/api/service/test?param=value",
			expectedPath: "/test", // Query params should be preserved in URL.RawQuery
		},
		{
			name: "empty path after strip becomes root",
			route: RouteConfig{
				PathPrefix: "/api/service",
				Backend: BackendConfig{URL: "http://service:8080"},
				Transform: &TransformConfig{StripPrefix: true},
			},
			requestPath:  "/api/service",
			expectedPath: "/",
		},
		{
			name: "path without leading slash after strip gets one",
			route: RouteConfig{
				PathPrefix: "/api/service/",
				Backend: BackendConfig{URL: "http://service:8080"},
				Transform: &TransformConfig{StripPrefix: true},
			},
			requestPath:  "/api/service/test",
			expectedPath: "/test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proxy, err := NewProxyMiddleware()
			require.NoError(t, err)
			req := httptest.NewRequest("GET", tt.requestPath, nil)
			
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Set("route", &tt.route)

			proxy.modifyRequest(req, &tt.route, c)

			assert.Equal(t, tt.expectedPath, req.URL.Path, "Path transformation should handle edge case correctly")
		})
	}
}
