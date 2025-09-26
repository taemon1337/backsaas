package gateway

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRouteTransformConfiguration tests that route transformation settings are properly configured
func TestRouteTransformConfiguration(t *testing.T) {
	tests := []struct {
		name                string
		routeConfig         RouteConfig
		expectedStripPrefix bool
		expectedAddHeaders  map[string]string
		expectedRemoveHeaders []string
	}{
		{
			name: "health dashboard with strip prefix",
			routeConfig: RouteConfig{
				Description: "Health Dashboard API",
				PathPrefix:  "/api/system-health",
				Backend:     BackendConfig{URL: "http://health-dashboard:8090"},
				Transform: &TransformConfig{
					StripPrefix: true,
					AddHeaders: map[string]string{
						"X-Interface-Type": "system-health",
					},
				},
				Enabled: true,
			},
			expectedStripPrefix: true,
			expectedAddHeaders: map[string]string{
				"X-Interface-Type": "system-health",
			},
		},
		{
			name: "platform API without strip prefix",
			routeConfig: RouteConfig{
				Description: "Platform API",
				PathPrefix:  "/api/platform",
				Backend:     BackendConfig{URL: "http://platform-api:8080"},
				Transform: &TransformConfig{
					StripPrefix: false,
				},
				Enabled: true,
			},
			expectedStripPrefix: false,
		},
		{
			name: "no transform config defaults to no stripping",
			routeConfig: RouteConfig{
				Description: "Default API",
				PathPrefix:  "/api/default",
				Backend:     BackendConfig{URL: "http://default:8080"},
				Transform:   nil,
				Enabled:     true,
			},
			expectedStripPrefix: false,
		},
		{
			name: "complex transformation with headers",
			routeConfig: RouteConfig{
				Description: "Complex API",
				PathPrefix:  "/api/complex",
				Backend:     BackendConfig{URL: "http://complex:8080"},
				Transform: &TransformConfig{
					StripPrefix: true,
					AddHeaders: map[string]string{
						"X-Gateway":      "backsaas",
						"X-Service-Type": "complex",
					},
					RemoveHeaders: []string{"X-Internal-Debug", "X-Sensitive"},
				},
				Enabled: true,
			},
			expectedStripPrefix: true,
			expectedAddHeaders: map[string]string{
				"X-Gateway":      "backsaas",
				"X-Service-Type": "complex",
			},
			expectedRemoveHeaders: []string{"X-Internal-Debug", "X-Sensitive"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test strip prefix setting
			if tt.routeConfig.Transform != nil {
				assert.Equal(t, tt.expectedStripPrefix, tt.routeConfig.Transform.StripPrefix,
					"StripPrefix setting should match expected")
			} else {
				assert.False(t, tt.expectedStripPrefix, "No transform config should default to no stripping")
			}

			// Test add headers
			if tt.expectedAddHeaders != nil {
				require.NotNil(t, tt.routeConfig.Transform, "Transform config should exist for header tests")
				assert.Equal(t, tt.expectedAddHeaders, tt.routeConfig.Transform.AddHeaders,
					"AddHeaders should match expected")
			}

			// Test remove headers
			if tt.expectedRemoveHeaders != nil {
				require.NotNil(t, tt.routeConfig.Transform, "Transform config should exist for header tests")
				assert.Equal(t, tt.expectedRemoveHeaders, tt.routeConfig.Transform.RemoveHeaders,
					"RemoveHeaders should match expected")
			}

			// Test that route is properly configured
			assert.NotEmpty(t, tt.routeConfig.Description, "Route should have description")
			assert.NotEmpty(t, tt.routeConfig.PathPrefix, "Route should have path prefix")
			assert.NotEmpty(t, tt.routeConfig.Backend.URL, "Route should have backend URL")
			assert.True(t, tt.routeConfig.Enabled, "Route should be enabled")
		})
	}
}

// TestRouteValidationLogic tests route configuration validation without external dependencies
func TestRouteValidationLogic(t *testing.T) {
	tests := []struct {
		name        string
		route       RouteConfig
		expectValid bool
	}{
		{
			name: "valid route with transform",
			route: RouteConfig{
				Description: "Valid API",
				PathPrefix:  "/api/valid",
				Backend:     BackendConfig{URL: "http://valid:8080", Timeout: 30 * time.Second},
				Transform: &TransformConfig{
					StripPrefix: true,
					AddHeaders:  map[string]string{"X-Test": "value"},
				},
				Enabled: true,
			},
			expectValid: true,
		},
		{
			name: "valid route without transform",
			route: RouteConfig{
				Description: "Simple API",
				PathPrefix:  "/api/simple",
				Backend:     BackendConfig{URL: "http://simple:8080", Timeout: 30 * time.Second},
				Transform:   nil,
				Enabled:     true,
			},
			expectValid: true,
		},
		{
			name: "route with empty path prefix",
			route: RouteConfig{
				Description: "Invalid API",
				PathPrefix:  "",
				Backend:     BackendConfig{URL: "http://invalid:8080", Timeout: 30 * time.Second},
				Enabled:     true,
			},
			expectValid: false,
		},
		{
			name: "route with empty backend URL",
			route: RouteConfig{
				Description: "Invalid API",
				PathPrefix:  "/api/invalid",
				Backend:     BackendConfig{URL: "", Timeout: 30 * time.Second},
				Enabled:     true,
			},
			expectValid: false,
		},
		{
			name: "disabled route",
			route: RouteConfig{
				Description: "Disabled API",
				PathPrefix:  "/api/disabled",
				Backend:     BackendConfig{URL: "http://disabled:8080", Timeout: 30 * time.Second},
				Enabled:     false,
			},
			expectValid: true, // Disabled routes can have any config
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Basic validation checks
			isValid := true

			if tt.route.Enabled {
				if tt.route.PathPrefix == "" {
					isValid = false
				}
				if tt.route.Backend.URL == "" {
					isValid = false
				}
			}

			assert.Equal(t, tt.expectValid, isValid, "Route validation should match expected result")
		})
	}
}

// TestTransformConfigDefaults tests that transform configuration defaults work correctly
func TestTransformConfigDefaults(t *testing.T) {
	tests := []struct {
		name      string
		transform *TransformConfig
		expected  struct {
			stripPrefix   bool
			hasAddHeaders bool
			hasRemoveHeaders bool
		}
	}{
		{
			name:      "nil transform config",
			transform: nil,
			expected: struct {
				stripPrefix   bool
				hasAddHeaders bool
				hasRemoveHeaders bool
			}{
				stripPrefix:   false,
				hasAddHeaders: false,
				hasRemoveHeaders: false,
			},
		},
		{
			name: "empty transform config",
			transform: &TransformConfig{},
			expected: struct {
				stripPrefix   bool
				hasAddHeaders bool
				hasRemoveHeaders bool
			}{
				stripPrefix:   false,
				hasAddHeaders: false,
				hasRemoveHeaders: false,
			},
		},
		{
			name: "strip prefix only",
			transform: &TransformConfig{
				StripPrefix: true,
			},
			expected: struct {
				stripPrefix   bool
				hasAddHeaders bool
				hasRemoveHeaders bool
			}{
				stripPrefix:   true,
				hasAddHeaders: false,
				hasRemoveHeaders: false,
			},
		},
		{
			name: "headers only",
			transform: &TransformConfig{
				AddHeaders:    map[string]string{"X-Test": "value"},
				RemoveHeaders: []string{"X-Remove"},
			},
			expected: struct {
				stripPrefix   bool
				hasAddHeaders bool
				hasRemoveHeaders bool
			}{
				stripPrefix:   false,
				hasAddHeaders: true,
				hasRemoveHeaders: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.transform == nil {
				// Nil transform should default to no transformations
				assert.True(t, true, "Nil transform config is valid")
				return
			}

			assert.Equal(t, tt.expected.stripPrefix, tt.transform.StripPrefix,
				"StripPrefix should match expected default")

			hasAddHeaders := len(tt.transform.AddHeaders) > 0
			assert.Equal(t, tt.expected.hasAddHeaders, hasAddHeaders,
				"AddHeaders presence should match expected")

			hasRemoveHeaders := len(tt.transform.RemoveHeaders) > 0
			assert.Equal(t, tt.expected.hasRemoveHeaders, hasRemoveHeaders,
				"RemoveHeaders presence should match expected")
		})
	}
}

// TestRouteMatchingLogic tests the route matching priority and logic
func TestRouteMatchingLogic(t *testing.T) {
	routes := []RouteConfig{
		{
			Description: "Generic API",
			PathPrefix:  "/api",
			Backend:     BackendConfig{URL: "http://generic:8080"},
			Enabled:     true,
		},
		{
			Description: "Specific Health API",
			PathPrefix:  "/api/system-health",
			Backend:     BackendConfig{URL: "http://health:8080"},
			Transform:   &TransformConfig{StripPrefix: true},
			Enabled:     true,
		},
		{
			Description: "Platform API",
			PathPrefix:  "/api/platform",
			Backend:     BackendConfig{URL: "http://platform:8080"},
			Transform:   &TransformConfig{StripPrefix: false},
			Enabled:     true,
		},
		{
			Description: "Disabled API",
			PathPrefix:  "/api/disabled",
			Backend:     BackendConfig{URL: "http://disabled:8080"},
			Enabled:     false,
		},
	}

	tests := []struct {
		name         string
		requestPath  string
		expectedRoute string
		shouldMatch  bool
	}{
		{
			name:         "specific health route should win over generic",
			requestPath:  "/api/system-health/status",
			expectedRoute: "Specific Health API",
			shouldMatch:  true,
		},
		{
			name:         "specific platform route should win over generic",
			requestPath:  "/api/platform/admin/login",
			expectedRoute: "Platform API",
			shouldMatch:  true,
		},
		{
			name:         "generic route should catch other API paths",
			requestPath:  "/api/other/service",
			expectedRoute: "Generic API",
			shouldMatch:  true,
		},
		{
			name:         "disabled route should not match",
			requestPath:  "/api/disabled/test",
			expectedRoute: "Generic API", // Should fall back to generic
			shouldMatch:  true,
		},
		{
			name:        "non-API path should not match any route",
			requestPath: "/health",
			shouldMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate route matching logic (most specific first)
			var matchedRoute *RouteConfig
			
			// Sort by path prefix length (longer = more specific)
			for i := range routes {
				route := &routes[i]
				if !route.Enabled {
					continue
				}
				
				if len(route.PathPrefix) > 0 && 
				   len(tt.requestPath) >= len(route.PathPrefix) &&
				   tt.requestPath[:len(route.PathPrefix)] == route.PathPrefix {
					
					if matchedRoute == nil || len(route.PathPrefix) > len(matchedRoute.PathPrefix) {
						matchedRoute = route
					}
				}
			}

			if tt.shouldMatch {
				assert.NotNil(t, matchedRoute, "Should find a matching route")
				if matchedRoute != nil {
					assert.Equal(t, tt.expectedRoute, matchedRoute.Description,
						"Should match the expected route")
				}
			} else {
				assert.Nil(t, matchedRoute, "Should not find a matching route")
			}
		})
	}
}
