package gateway

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRouteMatcher_PathPrefixMatching(t *testing.T) {
	routes := []RouteConfig{
		{
			Description: "Health Dashboard",
			PathPrefix:  "/api/system-health",
			Backend:     BackendConfig{URL: "http://health-dashboard:8090"},
			Transform:   &TransformConfig{StripPrefix: true},
			Enabled:     true,
		},
		{
			Description: "Platform API",
			PathPrefix:  "/api/platform",
			Backend:     BackendConfig{URL: "http://platform-api:8080"},
			Enabled:     true,
		},
		{
			Description: "Admin Console",
			PathPrefix:  "/admin",
			Backend:     BackendConfig{URL: "http://admin-console:3000"},
			Enabled:     true,
		},
		{
			Description: "Disabled Route",
			PathPrefix:  "/api/disabled",
			Backend:     BackendConfig{URL: "http://disabled:8080"},
			Enabled:     false,
		},
	}

	matcher, err := NewRouteMatcher(routes)
	require.NoError(t, err)

	tests := []struct {
		name           string
		requestPath    string
		expectedRoute  string
		expectedMatch  bool
	}{
		{
			name:          "health dashboard exact match",
			requestPath:   "/api/system-health",
			expectedRoute: "Health Dashboard",
			expectedMatch: true,
		},
		{
			name:          "health dashboard with subpath",
			requestPath:   "/api/system-health/api/status",
			expectedRoute: "Health Dashboard",
			expectedMatch: true,
		},
		{
			name:          "platform API exact match",
			requestPath:   "/api/platform",
			expectedRoute: "Platform API",
			expectedMatch: true,
		},
		{
			name:          "platform API admin login",
			requestPath:   "/api/platform/admin/login",
			expectedRoute: "Platform API",
			expectedMatch: true,
		},
		{
			name:          "admin console",
			requestPath:   "/admin/dashboard",
			expectedRoute: "Admin Console",
			expectedMatch: true,
		},
		{
			name:          "disabled route should not match",
			requestPath:   "/api/disabled/test",
			expectedMatch: false,
		},
		{
			name:          "no matching route",
			requestPath:   "/api/unknown",
			expectedMatch: false,
		},
		{
			name:          "root path",
			requestPath:   "/",
			expectedMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", tt.requestPath, nil)
			require.NoError(t, err)

			route, err := matcher.Match(req)

			if tt.expectedMatch {
				assert.NoError(t, err, "Should find a matching route")
				assert.NotNil(t, route, "Route should not be nil")
				assert.Equal(t, tt.expectedRoute, route.Description, "Should match expected route")
			} else {
				assert.Error(t, err, "Should not find a matching route")
				assert.Nil(t, route, "Route should be nil")
			}
		})
	}
}

func TestRouteMatcher_RoutePriority(t *testing.T) {
	// Test that more specific routes (longer prefixes) take priority
	routes := []RouteConfig{
		{
			Description: "Generic API",
			PathPrefix:  "/api",
			Backend:     BackendConfig{URL: "http://generic:8080"},
			Enabled:     true,
		},
		{
			Description: "Platform API",
			PathPrefix:  "/api/platform",
			Backend:     BackendConfig{URL: "http://platform:8080"},
			Enabled:     true,
		},
		{
			Description: "Health API",
			PathPrefix:  "/api/system-health",
			Backend:     BackendConfig{URL: "http://health:8080"},
			Enabled:     true,
		},
	}

	matcher, err := NewRouteMatcher(routes)
	require.NoError(t, err)

	tests := []struct {
		name          string
		requestPath   string
		expectedRoute string
	}{
		{
			name:          "most specific route wins - health",
			requestPath:   "/api/system-health/status",
			expectedRoute: "Health API",
		},
		{
			name:          "most specific route wins - platform",
			requestPath:   "/api/platform/tenants",
			expectedRoute: "Platform API",
		},
		{
			name:          "falls back to generic",
			requestPath:   "/api/other/service",
			expectedRoute: "Generic API",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", tt.requestPath, nil)
			require.NoError(t, err)

			route, err := matcher.Match(req)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedRoute, route.Description)
		})
	}
}

func TestRouteMatcher_HostMatching(t *testing.T) {
	routes := []RouteConfig{
		{
			Description: "Admin Console Host",
			Host:        "admin.backsaas.dev",
			PathPrefix:  "/",
			Backend:     BackendConfig{URL: "http://admin-console:3000"},
			Enabled:     true,
		},
		{
			Description: "API Host",
			Host:        "api.backsaas.dev",
			PathPrefix:  "/",
			Backend:     BackendConfig{URL: "http://api:8080"},
			Enabled:     true,
		},
		{
			Description: "Wildcard Subdomain",
			Host:        "*.backsaas.dev",
			PathPrefix:  "/",
			Backend:     BackendConfig{URL: "http://wildcard:8080"},
			Enabled:     true,
		},
		{
			Description: "Default Route",
			PathPrefix:  "/api",
			Backend:     BackendConfig{URL: "http://default:8080"},
			Enabled:     true,
		},
	}

	matcher, err := NewRouteMatcher(routes)
	require.NoError(t, err)

	tests := []struct {
		name          string
		host          string
		path          string
		expectedRoute string
	}{
		{
			name:          "exact host match - admin",
			host:          "admin.backsaas.dev",
			path:          "/dashboard",
			expectedRoute: "Admin Console Host",
		},
		{
			name:          "exact host match - api",
			host:          "api.backsaas.dev",
			path:          "/v1/users",
			expectedRoute: "API Host",
		},
		{
			name:          "wildcard subdomain match",
			host:          "tenant1.backsaas.dev",
			path:          "/app",
			expectedRoute: "Wildcard Subdomain",
		},
		{
			name:          "host with port",
			host:          "admin.backsaas.dev:8080",
			path:          "/settings",
			expectedRoute: "Admin Console Host",
		},
		{
			name:          "no host match falls back to path",
			host:          "other.example.com",
			path:          "/api/test",
			expectedRoute: "Default Route",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", tt.path, nil)
			require.NoError(t, err)
			req.Host = tt.host

			route, err := matcher.Match(req)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedRoute, route.Description)
		})
	}
}

func TestRouteMatcher_HeaderMatching(t *testing.T) {
	routes := []RouteConfig{
		{
			Description: "API with Version Header",
			PathPrefix:  "/api",
			Headers:     map[string]string{"X-API-Version": "v2"},
			Backend:     BackendConfig{URL: "http://api-v2:8080"},
			Enabled:     true,
		},
		{
			Description: "API with Client Header",
			PathPrefix:  "/api",
			Headers:     map[string]string{"X-Client-Type": "mobile"},
			Backend:     BackendConfig{URL: "http://mobile-api:8080"},
			Enabled:     true,
		},
		{
			Description: "Default API",
			PathPrefix:  "/api",
			Backend:     BackendConfig{URL: "http://default-api:8080"},
			Enabled:     true,
		},
	}

	matcher, err := NewRouteMatcher(routes)
	require.NoError(t, err)

	tests := []struct {
		name          string
		path          string
		headers       map[string]string
		expectedRoute string
	}{
		{
			name: "version header match",
			path: "/api/users",
			headers: map[string]string{
				"X-API-Version": "v2",
			},
			expectedRoute: "API with Version Header",
		},
		{
			name: "client header match",
			path: "/api/users",
			headers: map[string]string{
				"X-Client-Type": "mobile",
			},
			expectedRoute: "API with Client Header",
		},
		{
			name:          "no header match falls back to default",
			path:          "/api/users",
			headers:       map[string]string{},
			expectedRoute: "Default API",
		},
		{
			name: "wrong header value falls back to default",
			path: "/api/users",
			headers: map[string]string{
				"X-API-Version": "v1",
			},
			expectedRoute: "Default API",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", tt.path, nil)
			require.NoError(t, err)

			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}

			route, err := matcher.Match(req)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedRoute, route.Description)
		})
	}
}

func TestRouteConfig_TransformSettings(t *testing.T) {
	tests := []struct {
		name                string
		transform           *TransformConfig
		expectedStripPrefix bool
		expectedAddHeaders  map[string]string
		expectedRemoveHeaders []string
	}{
		{
			name: "strip prefix enabled",
			transform: &TransformConfig{
				StripPrefix: true,
			},
			expectedStripPrefix: true,
		},
		{
			name: "strip prefix disabled",
			transform: &TransformConfig{
				StripPrefix: false,
			},
			expectedStripPrefix: false,
		},
		{
			name:                "nil transform config",
			transform:           nil,
			expectedStripPrefix: false,
		},
		{
			name: "add headers",
			transform: &TransformConfig{
				AddHeaders: map[string]string{
					"X-Service": "test",
					"X-Version": "1.0",
				},
			},
			expectedAddHeaders: map[string]string{
				"X-Service": "test",
				"X-Version": "1.0",
			},
		},
		{
			name: "remove headers",
			transform: &TransformConfig{
				RemoveHeaders: []string{"X-Internal", "X-Debug"},
			},
			expectedRemoveHeaders: []string{"X-Internal", "X-Debug"},
		},
		{
			name: "complex transform",
			transform: &TransformConfig{
				StripPrefix: true,
				AddHeaders: map[string]string{
					"X-Gateway": "backsaas",
				},
				RemoveHeaders: []string{"X-Internal"},
			},
			expectedStripPrefix: true,
			expectedAddHeaders: map[string]string{
				"X-Gateway": "backsaas",
			},
			expectedRemoveHeaders: []string{"X-Internal"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			route := RouteConfig{
				PathPrefix: "/api/test",
				Backend:    BackendConfig{URL: "http://test:8080"},
				Transform:  tt.transform,
			}

			// Test strip prefix setting
			if tt.transform != nil {
				assert.Equal(t, tt.expectedStripPrefix, tt.transform.StripPrefix)
			}

			// Test add headers
			if tt.expectedAddHeaders != nil {
				require.NotNil(t, route.Transform)
				assert.Equal(t, tt.expectedAddHeaders, route.Transform.AddHeaders)
			}

			// Test remove headers
			if tt.expectedRemoveHeaders != nil {
				require.NotNil(t, route.Transform)
				assert.Equal(t, tt.expectedRemoveHeaders, route.Transform.RemoveHeaders)
			}
		})
	}
}

func TestRouteMatcher_ValidationErrors(t *testing.T) {
	tests := []struct {
		name          string
		routes        []RouteConfig
		expectedError bool
	}{
		{
			name: "valid routes",
			routes: []RouteConfig{
				{
					PathPrefix: "/api/test",
					Backend:    BackendConfig{URL: "http://test:8080"},
					Enabled:    true,
				},
			},
			expectedError: false,
		},
		{
			name: "invalid regex pattern",
			routes: []RouteConfig{
				{
					PathPrefix: "/api/test/*[invalid",
					Backend:    BackendConfig{URL: "http://test:8080"},
					Enabled:    true,
				},
			},
			expectedError: true,
		},
		{
			name:          "empty routes",
			routes:        []RouteConfig{},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewRouteMatcher(tt.routes)
			
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
